package metrics

import (
	"math"
	"testing"

	"github.com/Baike12/b-eval/internal/domain"
)

func trialWithAnswer(taskID, answer string, elapsed int64) domain.TrialOutcome {
	return domain.TrialOutcome{
		TrialID: taskID + "-" + answer,
		TaskID:  taskID,
		Failure: domain.ClassCompleted,
		Usage:   domain.Usage{ElapsedMS: elapsed},
		Verdict: domain.Verdict{Combined: domain.VerdictPass, Evidence: map[string]any{"answer": answer}},
	}
}

func TestAnswerConsistencyIsTheMeanPairwiseSimilarityAcrossTasks(t *testing.T) {
	reg := NewRegistry()
	trials := []domain.TrialOutcome{
		// The two answers of t1 are identical: similarity 1.
		trialWithAnswer("t1", "alpha beta gamma", 0),
		trialWithAnswer("t1", "alpha beta gamma", 0),
		// The two answers of t2 share no token: similarity 0.
		trialWithAnswer("t2", "alpha", 0),
		trialWithAnswer("t2", "gamma", 0),
	}
	_, v, err := reg.Compute("answer_consistency", MetricInput{Trials: trials, Aggregation: AggMacro})
	if err != nil {
		t.Fatal(err)
	}
	if !v.HasValue {
		t.Fatal("expected a value")
	}
	if math.Abs(v.Value-0.5) > 1e-9 {
		t.Errorf("consistency = %v, want 0.5", v.Value)
	}
}

func TestAnswerConsistencyHasNoCoverageWithoutTwoTrialsOfATask(t *testing.T) {
	reg := NewRegistry()
	trials := []domain.TrialOutcome{
		trialWithAnswer("t1", "alpha", 0),
		trialWithAnswer("t2", "beta", 0),
	}
	_, v, err := reg.Compute("answer_consistency", MetricInput{Trials: trials, Aggregation: AggMacro})
	if err != nil {
		t.Fatal(err)
	}
	if v.HasValue {
		t.Errorf("consistency = %v, want no coverage", v.Value)
	}
}

func labelledTrial(taskID, expected, predicted string) domain.TrialOutcome {
	return domain.TrialOutcome{
		TrialID: taskID + "-" + expected + "-" + predicted,
		TaskID:  taskID,
		Failure: domain.ClassCompleted,
		Verdict: domain.Verdict{
			Combined: domain.VerdictPass,
			Evidence: map[string]any{"expected_label": expected, "predicted_label": predicted},
		},
	}
}

func TestClassificationAccuracyCountsOnlyLabelledTrials(t *testing.T) {
	reg := NewRegistry()
	trials := []domain.TrialOutcome{
		labelledTrial("t1", "pass", "pass"),
		labelledTrial("t2", "pass", "fail"),
		labelledTrial("t3", "fail", "fail"),
		// A trial without labels is not a classification observation.
		trialWithAnswer("t4", "unlabelled", 0),
	}
	_, v, err := reg.Compute("classification_accuracy", MetricInput{Trials: trials, Aggregation: AggMicro})
	if err != nil {
		t.Fatal(err)
	}
	if !v.HasValue {
		t.Fatal("expected a value")
	}
	if v.Denominator != 3 {
		t.Errorf("denominator = %v, want 3 labelled trials", v.Denominator)
	}
	if math.Abs(v.Value-2.0/3.0) > 1e-9 {
		t.Errorf("accuracy = %v, want 2/3", v.Value)
	}
}

func TestLatencyAndJudgeTokenMetrics(t *testing.T) {
	reg := NewRegistry()
	trials := []domain.TrialOutcome{
		{TrialID: "a", TaskID: "t1", Failure: domain.ClassCompleted, Usage: domain.Usage{ElapsedMS: 100, JudgeInputTokens: 20, JudgeOutputTokens: 10}},
		{TrialID: "b", TaskID: "t1", Failure: domain.ClassCompleted, Usage: domain.Usage{ElapsedMS: 300, JudgeInputTokens: 30}},
	}
	_, latency, err := reg.Compute("avg_latency_ms", MetricInput{Trials: trials, Aggregation: AggMicro})
	if err != nil {
		t.Fatal(err)
	}
	if latency.Value != 200 {
		t.Errorf("latency = %v, want 200", latency.Value)
	}
	_, judge, err := reg.Compute("avg_judge_tokens_per_trial", MetricInput{Trials: trials, Aggregation: AggMicro})
	if err != nil {
		t.Fatal(err)
	}
	if judge.Value != 30 {
		t.Errorf("judge tokens per trial = %v, want 30", judge.Value)
	}
}

func TestLabelPrecisionRecallAndF1(t *testing.T) {
	trials := []domain.TrialOutcome{
		labelledTrial("t1", "pass", "pass"),
		labelledTrial("t2", "pass", "pass"),
		labelledTrial("t3", "fail", "pass"),
		labelledTrial("t4", "pass", "fail"),
	}
	cases := []struct {
		kind string
		want float64
	}{
		{"label_precision", 2.0 / 3.0},
		{"label_recall", 2.0 / 3.0},
		{"label_f1", 2.0 / 3.0},
	}
	for _, c := range cases {
		def, err := BuildCustom(CustomRequest{
			ID:           "metric-" + c.kind,
			Name:         c.kind,
			Kind:         c.kind,
			Unit:         UnitRatio,
			Aggregations: []string{AggMicro},
			Numerator:    "positive predictions that are correct",
			Denominator:  "positive predictions",
			Formula:      "tp / (tp + fp)",
			CIMethod:     "wilson",
			Params:       map[string]any{"positive_label": "pass"},
		})
		if err != nil {
			t.Fatalf("%s: %v", c.kind, err)
		}
		v, err := def.Compute(MetricInput{Trials: trials, Aggregation: AggMicro, CIMethod: "wilson"})
		if err != nil {
			t.Fatalf("%s: %v", c.kind, err)
		}
		if math.Abs(v.Value-c.want) > 1e-9 {
			t.Errorf("%s = %v, want %v", c.kind, v.Value, c.want)
		}
	}
}
