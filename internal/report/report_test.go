package report

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/Baike12/b-eval/internal/domain"
	"github.com/Baike12/b-eval/internal/metrics"
)

// The report is the only thing the web client reads, and it reads the
// collections as lists. A nil slice marshals to null, and a client that
// reads a length off null throws, so the report must never contain one.

func TestReportContainsNoNullValue(t *testing.T) {
	cases := []struct {
		name string
		run  domain.Run
	}{
		{
			name: "a passing trial with a scorer",
			run: domain.Run{
				RunID: "run-1",
				Plan:  domain.Plan{PlanID: "plan-1", Repetitions: 1},
				Trials: []domain.TrialOutcome{{
					TrialID:    "T-001#0#1",
					TaskID:     "T-001",
					Form:       domain.FormStaticQA,
					Domain:     "D1",
					Level1:     []string{"C1"},
					Difficulty: "L1",
					Failure:    domain.ClassCompleted,
					Verdict: domain.Verdict{
						TrialID:  "T-001#0#1",
						Scorers:  []domain.ScorerResult{{Type: "contains", Value: domain.VerdictPass, Score: 1}},
						Combined: domain.VerdictPass,
						Score:    1,
					},
				}},
			},
		},
		{
			name: "a trial with no scorer and no capability dimension",
			run: domain.Run{
				RunID: "run-2",
				Plan:  domain.Plan{PlanID: "plan-2", Repetitions: 1},
				Trials: []domain.TrialOutcome{{
					TrialID: "T-001#0#1",
					TaskID:  "T-001",
					Form:    domain.FormStaticQA,
					Domain:  "D1",
					Failure: domain.ClassCompleted,
					Verdict: domain.Verdict{
						TrialID:  "T-001#0#1",
						Combined: domain.VerdictPass,
						Score:    1,
					},
				}},
			},
		},
		{
			name: "a run with no trial at all",
			run: domain.Run{
				RunID: "run-3",
				Plan:  domain.Plan{PlanID: "plan-3"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := Build(tc.run, metrics.NewRegistry(), map[string]string{}, Options{})
			body, err := json.Marshal(rep)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			for _, path := range nullPaths(t, body) {
				t.Errorf("report contains a null at %s", path)
			}
		})
	}
}

// nullPaths returns the JSON paths that hold a null value.
func nullPaths(t *testing.T, body []byte) []string {
	t.Helper()
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	var found []string
	var walk func(value any, path string)
	walk = func(value any, path string) {
		switch v := value.(type) {
		case map[string]any:
			for key, child := range v {
				walk(child, path+"."+key)
			}
		case []any:
			for i, child := range v {
				walk(child, path+"["+strconv.Itoa(i)+"]")
			}
		case nil:
			found = append(found, path)
		}
	}
	walk(decoded, "$")
	return found
}
