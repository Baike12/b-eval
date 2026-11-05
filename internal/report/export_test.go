package report

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/Baike12/b-eval/internal/domain"
)

func TestTrialsCSVCarriesTheDrillDownColumns(t *testing.T) {
	run := domain.Run{
		RunID: "run-1",
		Trajectories: []domain.Trajectory{{
			TrialID:     "T-1#0#1",
			TaskID:      "T-1",
			Termination: domain.TermCompleted,
		}},
		Trials: []domain.TrialOutcome{{
			TrialID:    "T-1#0#1",
			TaskID:     "T-1",
			Form:       domain.FormStaticQA,
			Domain:     "D7",
			Difficulty: "L1",
			Failure:    domain.ClassCompleted,
			Steps:      1,
			Usage: domain.Usage{
				InputTokens:       100,
				OutputTokens:      20,
				JudgeInputTokens:  30,
				JudgeOutputTokens: 5,
				ElapsedMS:         250,
			},
			Verdict: domain.Verdict{Combined: domain.VerdictPass, Score: 1, PrimaryPassed: true},
		}},
	}

	data, err := TrialsCSV(run)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want a header and one trial", len(rows))
	}
	if rows[0][0] != "trial_id" || rows[0][len(rows[0])-1] != "termination_reason" {
		t.Errorf("unexpected header: %v", rows[0])
	}
	record := map[string]string{}
	for i, column := range rows[0] {
		record[column] = rows[1][i]
	}
	if record["trial_id"] != "T-1#0#1" {
		t.Errorf("trial_id = %q", record["trial_id"])
	}
	if record["verdict"] != "PASS" || record["termination_reason"] != "completed" {
		t.Errorf("verdict/termination = %q/%q", record["verdict"], record["termination_reason"])
	}
	if record["judge_input_tokens"] != "30" || record["elapsed_ms"] != "250" {
		t.Errorf("usage columns = %q/%q", record["judge_input_tokens"], record["elapsed_ms"])
	}
}
