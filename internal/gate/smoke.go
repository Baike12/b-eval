package gate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/Baike12/b-eval/internal/api"
	"github.com/Baike12/b-eval/internal/domain"
	"github.com/Baike12/b-eval/internal/store"
)

// SmokeCheck runs the whole vertical slice through the HTTP API.
//
// The slice is the one the platform promises: a batch of task data, an agent
// under test, a run, a verdict and a report. Running it through the real
// HTTP surface means the gate exercises the same code path a user does, not
// a shortcut around it.
type SmokeCheck struct {
	// GateID is the stable identifier of the gate.
	GateID string
	// Description states what the gate protects.
	Description string
	// Root is the repository root.
	Root string
}

// ID returns the gate identifier.
func (c SmokeCheck) ID() string { return c.GateID }

// Describe returns what the gate protects.
func (c SmokeCheck) Describe() string { return c.Description }

// Run executes the vertical slice and asserts its outcome.
func (c SmokeCheck) Run(ctx context.Context) Result {
	res := Result{GateID: c.GateID, Status: StatusPassed}

	dir, err := os.MkdirTemp("", "b-eval-smoke-*")
	if err != nil {
		res.Status = StatusFailed
		res.Detail = err.Error()
		return res
	}
	defer os.RemoveAll(dir)

	srv := api.NewServer(store.New(filepath.Join(dir, "data")))
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// 1. A batch of task data enters the platform.
	taskset := map[string]any{
		"task_set_id": "smoke",
		"version":     "v1",
		"benchmark":   "smoke-benchmark",
		"tasks": []map[string]any{
			smokeTask("T-1", "orders over $100 qualify for free returns", "30 days"),
			smokeTask("T-2", "the return window is 30 days from delivery", "receipt"),
		},
	}
	code, body, err := post(ts.URL+"/api/tasksets", taskset)
	if err != nil {
		return smokeFailed(res, "create task set: %v", err)
	}
	if code != http.StatusCreated {
		return smokeFailed(res, "create task set: status %d: %s", code, body)
	}

	// 2. A static-QA task that carries an environment field is refused.
	// A gate that never sees the rejection is not evidence that the rule
	// works.
	invalid := map[string]any{
		"task_set_id": "invalid",
		"version":     "v1",
		"tasks": []map[string]any{
			withInitialState(smokeTask("T-9", "anything", "keyword")),
		},
	}
	code, body, err = post(ts.URL+"/api/tasksets", invalid)
	if err != nil {
		return smokeFailed(res, "create invalid task set: %v", err)
	}
	if code != http.StatusUnprocessableEntity {
		return smokeFailed(res, "a static-QA task carrying an environment field was accepted (status %d): %s", code, body)
	}

	// 3. An agent registers and passes the self-check.
	code, body, err = post(ts.URL+"/api/agents", map[string]any{
		"agent_id": "mock-good",
		"name":     "Mock agent",
		"shape":    "mock",
		"answers": map[string]string{
			"T-1": "Orders over $100 qualify for free returns within 30 days; keep the receipt.",
			"T-2": "The return window is 30 days from delivery and a receipt is required.",
		},
	})
	if err != nil {
		return smokeFailed(res, "register agent: %v", err)
	}
	if code != http.StatusCreated {
		return smokeFailed(res, "register agent: status %d: %s", code, body)
	}
	code, body, err = post(ts.URL+"/api/agents/mock-good/selfcheck", map[string]any{})
	if err != nil {
		return smokeFailed(res, "self-check: %v", err)
	}
	if code != http.StatusOK {
		return smokeFailed(res, "self-check: status %d: %s", code, body)
	}
	var selfCheck domain.SelfCheckResult
	if err := json.Unmarshal(body, &selfCheck); err != nil {
		return smokeFailed(res, "self-check: decode: %v", err)
	}
	if selfCheck.Status != "passed" {
		return smokeFailed(res, "self-check did not pass: %v", selfCheck.InvalidReasons)
	}

	// 4. A run executes.
	code, body, err = post(ts.URL+"/api/runs", map[string]any{
		"task_set_id": "smoke",
		"agent_id":    "mock-good",
		"repetitions": 2,
		"aggregation": "macro",
	})
	if err != nil {
		return smokeFailed(res, "create run: %v", err)
	}
	if code != http.StatusCreated {
		return smokeFailed(res, "create run: status %d: %s", code, body)
	}
	var run domain.Run
	if err := json.Unmarshal(body, &run); err != nil {
		return smokeFailed(res, "create run: decode: %v", err)
	}
	if len(run.Trials) != 4 {
		return smokeFailed(res, "expected 4 trials (2 tasks x 2 repetitions), got %d", len(run.Trials))
	}

	// 5. The report carries the confidence section first and a score.
	code, body, err = get(ts.URL + "/api/runs/" + run.RunID + "/report")
	if err != nil {
		return smokeFailed(res, "get report: %v", err)
	}
	if code != http.StatusOK {
		return smokeFailed(res, "get report: status %d: %s", code, body)
	}
	var rep map[string]any
	if err := json.Unmarshal(body, &rep); err != nil {
		return smokeFailed(res, "get report: decode: %v", err)
	}
	if _, ok := rep["confidence"]; !ok {
		return smokeFailed(res, "the report has no confidence section")
	}
	main, _ := rep["main_score"].(map[string]any)
	if main == nil {
		return smokeFailed(res, "the report has no main score section")
	}
	if v, _ := main["total_score"].(float64); v < 0.999 {
		return smokeFailed(res, "the mock agent should pass every task, got total score %v", v)
	}

	// 6. A registered custom metric appears in the report.
	code, body, err = post(ts.URL+"/api/metrics", map[string]any{
		"id":          "keyword_coverage_smoke",
		"name":        "Keyword coverage",
		"kind":        "keyword_coverage",
		"unit":        "ratio",
		"numerator":   "Trials whose answer contains every keyword",
		"denominator": "Valid trials",
		"params":      map[string]any{"keywords": []string{"30 days"}},
	})
	if err != nil {
		return smokeFailed(res, "register metric: %v", err)
	}
	if code != http.StatusCreated {
		return smokeFailed(res, "register metric: status %d: %s", code, body)
	}
	code, body, err = get(ts.URL + "/api/metrics")
	if err != nil {
		return smokeFailed(res, "list metrics: %v", err)
	}
	if code != http.StatusOK || !bytes.Contains(body, []byte("keyword_coverage_smoke")) {
		return smokeFailed(res, "the registered custom metric is not in the registry")
	}
	code, body, err = get(ts.URL + "/api/runs/" + run.RunID + "/report")
	if err != nil {
		return smokeFailed(res, "get report: %v", err)
	}
	if code != http.StatusOK || !bytes.Contains(body, []byte("keyword_coverage_smoke")) {
		return smokeFailed(res, "the registered custom metric is not in the report")
	}

	res.Evidence = map[string]any{
		"run_id":  run.RunID,
		"trials":  len(run.Trials),
		"tasks":   len(run.Trials) / 2,
		"metrics": len(rep["metrics"].([]any)),
	}
	return res
}

func smokeFailed(res Result, format string, args ...any) Result {
	res.Status = StatusFailed
	res.Detail = fmt.Sprintf(format, args...)
	return res
}

// smokeTask builds one static-QA task for the smoke run.
func smokeTask(id, reference, keyword string) map[string]any {
	return map[string]any{
		"task_id":      id,
		"form":         "static-qa",
		"display_name": "Smoke task " + id,
		"capability": map[string]any{
			"level1_dimensions": []string{"C4"},
			"level2_labels":     []string{"requirement-clarification"},
			"difficulty":        "L1",
			"domains":           []string{"D7"},
		},
		"description": map[string]any{
			"instruction": "Answer the customer question about the return policy.",
		},
		"budget": map[string]any{
			"steps":            1,
			"tokens":           8000,
			"duration_seconds": 30,
		},
		"interaction_contract": map[string]any{
			"termination_conditions": []string{"answer"},
			"multi_turn":             false,
		},
		"judgement": map[string]any{
			"combination":      "and",
			"primary":          0,
			"reference_answer": reference,
			"scorers": []map[string]any{
				{
					"type":    "contains",
					"params":  map[string]any{"needles": []string{keyword}},
					"weight":  1,
					"primary": true,
				},
			},
		},
		"meta": map[string]any{"author": "gate"},
	}
}

// withInitialState adds an environment field to a static-QA task, which the
// form contract must refuse.
func withInitialState(task map[string]any) map[string]any {
	task["initial_state"] = map[string]any{"snapshot_id": "snap-1", "seed": 1}
	return task
}

func post(url string, payload any) (int, []byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, buf.Bytes(), nil
}

func get(url string) (int, []byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, buf.Bytes(), nil
}
