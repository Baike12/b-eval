package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Baike12/b-eval/internal/domain"
	"github.com/Baike12/b-eval/internal/store"
)

// prepareCompletedRun registers the fixture task set and agent, runs the
// self-check, and returns the completed run.
func prepareCompletedRun(t *testing.T, tsURL string) map[string]any {
	t.Helper()
	if code, body := post(t, tsURL+"/api/tasksets", map[string]any{
		"task_set_id": "demo", "version": "v1", "tasks": []any{taskPayload()},
	}); code != http.StatusCreated {
		t.Fatalf("create task set: %d: %s", code, body)
	}
	if code, body := post(t, tsURL+"/api/agents", map[string]any{
		"agent_id": "a", "shape": "mock", "answers": map[string]string{"T-1": "within 30 days"},
	}); code != http.StatusCreated {
		t.Fatalf("register agent: %d: %s", code, body)
	}
	if code, body := post(t, tsURL+"/api/agents/a/selfcheck", map[string]any{}); code != http.StatusOK {
		t.Fatalf("self-check: %d: %s", code, body)
	}
	code, body := post(t, tsURL+"/api/runs", map[string]any{
		"task_set_id": "demo", "agent_id": "a", "repetitions": 1,
	})
	if code != http.StatusCreated {
		t.Fatalf("create run: %d: %s", code, body)
	}
	var run map[string]any
	if err := json.Unmarshal(body, &run); err != nil {
		t.Fatal(err)
	}
	return run
}

func TestRescoreProducesANewRunAndKeepsTheOriginal(t *testing.T) {
	ts := newTestServer(t)
	original := prepareCompletedRun(t, ts.URL)
	originalID := original["run_id"].(string)

	code, body := post(t, ts.URL+"/api/runs/"+originalID+"/rescore", map[string]any{
		"reason": "the rubric changed after review",
	})
	if code != http.StatusCreated {
		t.Fatalf("rescore: status %d: %s", code, body)
	}
	var rescored map[string]any
	if err := json.Unmarshal(body, &rescored); err != nil {
		t.Fatal(err)
	}
	rescoredID := rescored["run_id"].(string)
	if rescoredID == originalID {
		t.Fatal("the re-judgement overwrote the original run")
	}
	if rescored["rescored_from"] != originalID {
		t.Errorf("rescored_from = %v, want %s", rescored["rescored_from"], originalID)
	}
	if rescored["rescore_reason"] != "the rubric changed after review" {
		t.Errorf("rescore_reason = %v", rescored["rescore_reason"])
	}
	if len(rescored["trials"].([]any)) != len(original["trials"].([]any)) {
		t.Errorf("trial count changed: %d vs %d", len(rescored["trials"].([]any)), len(original["trials"].([]any)))
	}
	if len(rescored["trajectories"].([]any)) != len(original["trajectories"].([]any)) {
		t.Errorf("the trajectories were not carried over")
	}

	// The source run is untouched and still readable.
	code, body = get(t, ts.URL+"/api/runs/"+originalID)
	if code != http.StatusOK {
		t.Fatalf("original run: %d: %s", code, body)
	}
	var reloaded map[string]any
	if err := json.Unmarshal(body, &reloaded); err != nil {
		t.Fatal(err)
	}
	if reloaded["rescored_from"] != nil {
		t.Errorf("the original run was modified: %v", reloaded["rescored_from"])
	}
}

func TestRescoreRequiresAReason(t *testing.T) {
	ts := newTestServer(t)
	run := prepareCompletedRun(t, ts.URL)
	code, body := post(t, ts.URL+"/api/runs/"+run["run_id"].(string)+"/rescore", map[string]any{})
	if code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", code, body)
	}
	if !bytes.Contains(body, []byte("reason")) {
		t.Errorf("the refusal does not name the reason: %s", body)
	}
}

func TestRescoreCanApplyANewTaskSetVersion(t *testing.T) {
	ts := newTestServer(t)
	run := prepareCompletedRun(t, ts.URL)

	// A corrected task set applies a stricter scorer. The trajectories are
	// unchanged, but the verdict must follow the new criterion.
	strict := taskPayload()
	strict["judgement"] = map[string]any{
		"combination":      "and",
		"primary":          0,
		"reference_answer": "the reference",
		"scorers": []map[string]any{{
			"type": "exact", "weight": 1, "primary": true,
		}},
	}
	if code, body := post(t, ts.URL+"/api/tasksets", map[string]any{
		"task_set_id": "demo", "version": "v2", "tasks": []any{strict},
	}); code != http.StatusCreated {
		t.Fatalf("create v2: %d: %s", code, body)
	}

	code, body := post(t, ts.URL+"/api/runs/"+run["run_id"].(string)+"/rescore", map[string]any{
		"reason":           "the v2 criterion requires an exact answer",
		"task_set_id":      "demo",
		"task_set_version": "v2",
	})
	if code != http.StatusCreated {
		t.Fatalf("rescore against v2: %d: %s", code, body)
	}
	var rescored map[string]any
	if err := json.Unmarshal(body, &rescored); err != nil {
		t.Fatal(err)
	}
	plan := rescored["plan"].(map[string]any)
	if plan["task_set_version"] != "v2" {
		t.Errorf("the re-scored plan still names %v", plan["task_set_version"])
	}
	trial := rescored["trials"].([]any)[0].(map[string]any)
	verdict := trial["verdict"].(map[string]any)
	if verdict["combined"] != "FAIL" {
		t.Errorf("verdict = %v, want FAIL under the exact scorer", verdict["combined"])
	}
}

func TestRunExportServesCSVAndJSON(t *testing.T) {
	ts := newTestServer(t)
	run := prepareCompletedRun(t, ts.URL)
	id := run["run_id"].(string)

	code, body := get(t, ts.URL+"/api/runs/"+id+"/export?format=csv")
	if code != http.StatusOK {
		t.Fatalf("csv export: %d: %s", code, body)
	}
	if !bytes.HasPrefix(body, []byte("trial_id,")) {
		t.Errorf("the CSV does not lead with its header: %s", body)
	}
	if !bytes.Contains(body, []byte("T-1#0#1")) {
		t.Errorf("the CSV does not carry the trial row: %s", body)
	}

	code, body = get(t, ts.URL+"/api/runs/"+id+"/export?format=json")
	if code != http.StatusOK {
		t.Fatalf("json export: %d: %s", code, body)
	}
	var exported map[string]any
	if err := json.Unmarshal(body, &exported); err != nil {
		t.Fatalf("the JSON export is not the run record: %v", err)
	}
	if exported["run_id"] != id {
		t.Errorf("run_id = %v, want %s", exported["run_id"], id)
	}

	code, body = get(t, ts.URL+"/api/runs/"+id+"/export?format=xlsx")
	if code != http.StatusBadRequest {
		t.Fatalf("unknown format: %d, want 400: %s", code, body)
	}
}

func TestUniqueRunIDDoesNotOverwriteAnExistingRun(t *testing.T) {
	dir := t.TempDir()
	srv := NewServer(store.New(dir))
	if err := srv.Store.SaveRun(domain.Run{RunID: "run-x"}); err != nil {
		t.Fatal(err)
	}
	id, err := srv.uniqueRunID("run-x")
	if err != nil {
		t.Fatal(err)
	}
	if id != "run-x-2" {
		t.Errorf("id = %q, want run-x-2", id)
	}
}
