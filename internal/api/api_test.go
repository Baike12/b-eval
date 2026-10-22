package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Baike12/b-eval/internal/store"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	srv := NewServer(store.New(filepath.Join(dir, "data")))
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func post(t *testing.T, url string, payload any) (int, []byte) {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, buf.Bytes()
}

func get(t *testing.T, url string) (int, []byte) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, buf.Bytes()
}

func taskPayload() map[string]any {
	return map[string]any{
		"task_id":      "T-1",
		"form":         "static-qa",
		"display_name": "A question",
		"capability": map[string]any{
			"level1_dimensions": []string{"C4"},
			"difficulty":        "L1",
			"domains":           []string{"D7"},
		},
		"description": map[string]any{"instruction": "Answer the question."},
		"budget":      map[string]any{"steps": 1, "tokens": 1000},
		"interaction_contract": map[string]any{
			"termination_conditions": []string{"answer"},
			"multi_turn":             false,
		},
		"judgement": map[string]any{
			"combination":      "and",
			"primary":          0,
			"reference_answer": "the reference",
			"scorers": []map[string]any{{
				"type":    "contains",
				"params":  map[string]any{"needles": []string{"30 days"}},
				"weight":  1,
				"primary": true,
			}},
		},
		"meta": map[string]any{"author": "test"},
	}
}

func TestHealthEndpoint(t *testing.T) {
	ts := newTestServer(t)
	code, body := get(t, ts.URL+"/api/healthz")
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["status"] != "ok" {
		t.Errorf("status = %v, want ok", payload["status"])
	}
}

func TestTaskSetLifecycle(t *testing.T) {
	ts := newTestServer(t)

	code, body := post(t, ts.URL+"/api/tasksets", map[string]any{
		"task_set_id": "demo",
		"version":     "v1",
		"benchmark":   "b",
		"tasks":       []any{taskPayload()},
	})
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", code, body)
	}

	code, body = get(t, ts.URL+"/api/tasksets/demo/v1")
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	var set map[string]any
	if err := json.Unmarshal(body, &set); err != nil {
		t.Fatal(err)
	}
	tasks := set["tasks"].([]any)
	if len(tasks) != 1 {
		t.Fatalf("tasks = %d, want 1", len(tasks))
	}
}

func TestAStaticQATaskCarryingAnEnvironmentFieldIsRefused(t *testing.T) {
	ts := newTestServer(t)
	task := taskPayload()
	task["initial_state"] = map[string]any{"snapshot_id": "snap-1"}

	code, body := post(t, ts.URL+"/api/tasksets", map[string]any{
		"task_set_id": "bad",
		"version":     "v1",
		"tasks":       []any{task},
	})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", code, body)
	}
	if !bytes.Contains(body, []byte("initial_state")) {
		t.Errorf("the refusal does not name the offending field: %s", body)
	}
}

func TestARunIsRefusedBeforeTheSelfCheck(t *testing.T) {
	ts := newTestServer(t)
	post(t, ts.URL+"/api/tasksets", map[string]any{
		"task_set_id": "demo", "version": "v1", "tasks": []any{taskPayload()},
	})
	post(t, ts.URL+"/api/agents", map[string]any{
		"agent_id": "a", "shape": "mock", "answers": map[string]string{"T-1": "within 30 days"},
	})

	code, body := post(t, ts.URL+"/api/runs", map[string]any{
		"task_set_id": "demo", "agent_id": "a", "repetitions": 1,
	})
	if code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", code, body)
	}
	if !bytes.Contains(body, []byte("self-check")) {
		t.Errorf("the refusal does not mention the self-check: %s", body)
	}
}

func TestTheFullVerticalSlice(t *testing.T) {
	ts := newTestServer(t)

	post(t, ts.URL+"/api/tasksets", map[string]any{
		"task_set_id": "demo", "version": "v1", "tasks": []any{taskPayload()},
	})
	code, body := post(t, ts.URL+"/api/agents", map[string]any{
		"agent_id": "a", "shape": "mock", "answers": map[string]string{"T-1": "the window is 30 days"},
	})
	if code != http.StatusCreated {
		t.Fatalf("register agent: status %d: %s", code, body)
	}
	code, body = post(t, ts.URL+"/api/agents/a/selfcheck", map[string]any{})
	if code != http.StatusOK {
		t.Fatalf("self-check: status %d: %s", code, body)
	}
	code, body = post(t, ts.URL+"/api/runs", map[string]any{
		"task_set_id": "demo", "agent_id": "a", "repetitions": 2,
	})
	if code != http.StatusCreated {
		t.Fatalf("create run: status %d: %s", code, body)
	}
	var run map[string]any
	if err := json.Unmarshal(body, &run); err != nil {
		t.Fatal(err)
	}
	runID := run["run_id"].(string)

	code, body = get(t, ts.URL+"/api/runs/"+runID+"/report")
	if code != http.StatusOK {
		t.Fatalf("report: status %d: %s", code, body)
	}
	var report map[string]any
	if err := json.Unmarshal(body, &report); err != nil {
		t.Fatal(err)
	}
	confidence, ok := report["confidence"].(map[string]any)
	if !ok {
		t.Fatal("the report has no confidence section")
	}
	main := report["main_score"].(map[string]any)
	if v, _ := main["total_score"].(float64); v != 1 {
		t.Errorf("total score = %v, want 1", v)
	}
	if v, _ := confidence["valid_sample_size"].(float64); v != 2 {
		t.Errorf("valid sample size = %v, want 2", v)
	}
}

func TestACustomMetricIsRefusedWhenItIsIncomplete(t *testing.T) {
	ts := newTestServer(t)
	code, body := post(t, ts.URL+"/api/metrics", map[string]any{
		"id": "incomplete", "name": "Incomplete", "kind": "keyword_coverage",
		"unit": "ratio", "params": map[string]any{"keywords": []string{"a"}},
	})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", code, body)
	}
	if !bytes.Contains(body, []byte("numerator")) || !bytes.Contains(body, []byte("denominator")) {
		t.Errorf("the refusal does not name the missing fields: %s", body)
	}
}

func TestTheMetricRegistryRefusesARemoteBuiltinID(t *testing.T) {
	ts := newTestServer(t)
	code, body := post(t, ts.URL+"/api/metrics", map[string]any{
		"id": "success_rate", "name": "Hijack", "kind": "keyword_coverage",
		"unit": "ratio", "numerator": "a", "denominator": "b", "formula": "a/b",
		"params": map[string]any{"keywords": []string{"a"}},
	})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", code, body)
	}
	if !bytes.Contains(body, []byte("built-in metric ID is reserved")) {
		t.Errorf("the refusal does not name the reason: %s", body)
	}
}

func TestTheStorePersistsAcrossServers(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")

	first := httptest.NewServer(NewServer(store.New(dataDir)).Handler())
	code, body := post(t, first.URL+"/api/tasksets", map[string]any{
		"task_set_id": "demo", "version": "v1", "tasks": []any{taskPayload()},
	})
	if code != http.StatusCreated {
		t.Fatalf("status = %d: %s", code, body)
	}
	first.Close()

	second := httptest.NewServer(NewServer(store.New(dataDir)).Handler())
	defer second.Close()
	code, body = get(t, second.URL+"/api/tasksets")
	if code != http.StatusOK {
		t.Fatalf("status = %d: %s", code, body)
	}
	if !bytes.Contains(body, []byte("demo")) {
		t.Errorf("the task set did not survive the restart: %s", body)
	}
}

func TestTheDataDirectoryIsCreatedOnDemand(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "data")
	st := store.New(dir)
	if err := st.SaveTaskSet(taskSetFixture()); err != nil {
		t.Fatalf("the store did not create its directory: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("the data directory does not exist: %v", err)
	}
}
