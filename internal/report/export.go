package report

import (
	"bytes"
	"encoding/csv"
	"strconv"

	"github.com/Baike12/b-eval/internal/domain"
)

// TrialCSVHeader is the column order of the trial export.
//
// The export is the drill-down artefact: one row per trial, carrying the
// verdict, the failure class, the termination reason and the full usage. A
// reader can recompute every aggregate in the report from these rows.
var TrialCSVHeader = []string{
	"trial_id",
	"task_id",
	"form",
	"domain",
	"difficulty",
	"failure_class",
	"verdict",
	"score",
	"primary_passed",
	"steps",
	"budget_steps",
	"input_tokens",
	"output_tokens",
	"judge_input_tokens",
	"judge_output_tokens",
	"model_calls",
	"tool_calls",
	"cost_usd",
	"judge_cost_usd",
	"elapsed_ms",
	"termination_reason",
}

// TrialsCSV renders the trial rows of a run as CSV.
//
// A platform without a spreadsheet dependency still has to hand a third party
// the rows behind its numbers, so the export is plain RFC 4180 CSV.
func TrialsCSV(run domain.Run) ([]byte, error) {
	terminations := make(map[string]string, len(run.Trajectories))
	for _, tr := range run.Trajectories {
		terminations[tr.TrialID] = string(tr.Termination)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(TrialCSVHeader); err != nil {
		return nil, err
	}
	for _, t := range run.Trials {
		row := []string{
			t.TrialID,
			t.TaskID,
			string(t.Form),
			t.Domain,
			t.Difficulty,
			string(t.Failure),
			string(t.Verdict.Combined),
			formatFloat(t.Verdict.Score, 6),
			strconv.FormatBool(t.Verdict.PrimaryPassed),
			strconv.Itoa(t.Steps),
			strconv.Itoa(t.BudgetSteps),
			strconv.Itoa(t.Usage.InputTokens),
			strconv.Itoa(t.Usage.OutputTokens),
			strconv.Itoa(t.Usage.JudgeInputTokens),
			strconv.Itoa(t.Usage.JudgeOutputTokens),
			strconv.Itoa(t.Usage.ModelCalls),
			strconv.Itoa(t.Usage.ToolCalls),
			formatFloat(t.Usage.CostUSD, 6),
			formatFloat(t.Usage.JudgeCostUSD, 6),
			strconv.FormatInt(t.Usage.ElapsedMS, 10),
			terminations[t.TrialID],
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func formatFloat(value float64, digits int) string {
	return strconv.FormatFloat(value, 'f', digits, 64)
}
