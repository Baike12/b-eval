package metrics

import (
	"math"
	"testing"

	"github.com/Baike12/b-eval/internal/domain"
)

// buildWorkedExample builds the trial set of the worked example in
// FEAT-004 §4.6.
//
// The example is the convention consistency check: a second implementation
// of the same formulas must reproduce the same numbers, so the example is
// asserted here rather than restated in prose.
func buildWorkedExample() []domain.TrialOutcome {
	trials := []domain.TrialOutcome{}
	add := func(taskID string, index int, passed bool, score float64, failure domain.FailureClass) {
		verdict := domain.Verdict{Combined: domain.VerdictFail, Score: score}
		if passed {
			verdict.Combined = domain.VerdictPass
		}
		if failure == domain.ClassPlatformFailure {
			verdict.Combined = domain.VerdictUnscored
		}
		trials = append(trials, domain.TrialOutcome{
			TrialID:    taskID + "-" + string(rune('0'+index)),
			TaskID:     taskID,
			TrialIndex: index,
			Failure:    failure,
			Verdict:    verdict,
		})
	}
	// t1: 3 trials, all passing, progress 1.00 each.
	for i := 0; i < 3; i++ {
		add("t1", i, true, 1.00, domain.ClassCompleted)
	}
	// t2: 3 trials, 1 passing, mean progress 0.55.
	add("t2", 0, true, 1.00, domain.ClassCompleted)
	add("t2", 1, false, 0.50, domain.ClassCompleted)
	add("t2", 2, false, 0.15, domain.ClassCompleted)
	// t3: 2 valid trials, none passing, progress 0.10 each, plus one
	// platform failure that is excluded from the denominator.
	add("t3", 0, false, 0.10, domain.ClassCompleted)
	add("t3", 1, false, 0.10, domain.ClassCompleted)
	add("t3", 2, false, 0.00, domain.ClassPlatformFailure)
	return trials
}

func TestWorkedExampleReproducesTheDocumentedNumbers(t *testing.T) {
	reg := NewRegistry()
	trials := buildWorkedExample()

	cases := []struct {
		id      string
		agg     string
		want    float64
		tol     float64
		comment string
	}{
		{"success_rate", AggMicro, 0.500, 1e-9, "SR = 4/8"},
		{"success_rate", AggMacro, 0.444, 0.001, "macro success rate = (1.000+0.333+0.000)/3"},
		{"pass_at_k", AggMacro, 1.000, 1e-9, "pass@3 over the two tasks with n_t >= 3"},
		{"pass_power_k", AggMacro, 0.500, 1e-9, "pass^3 = (1+0)/2"},
		{"avg_at_k", AggMicro, 0.606, 0.001, "avg@3 = 4.85/8"},
		{"platform_failure_rate", AggMicro, 0.111, 0.001, "platform failure rate = 1/9"},
	}

	for _, c := range cases {
		_, v, err := reg.Compute(c.id, MetricInput{Trials: trials, K: 3, Aggregation: c.agg, CIMethod: "wilson"})
		if err != nil {
			t.Fatalf("%s: %v", c.id, err)
		}
		if !v.HasValue {
			t.Fatalf("%s: expected a value, got no coverage", c.id)
		}
		if math.Abs(v.Value-c.want) > c.tol {
			t.Errorf("%s (%s): got %v, want %v — %s", c.id, c.agg, v.Value, c.want, c.comment)
		}
	}
}

func TestPlatformFailuresAreExcludedFromTheDenominator(t *testing.T) {
	reg := NewRegistry()
	trials := buildWorkedExample()

	_, v, err := reg.Compute("success_rate", MetricInput{Trials: trials, Aggregation: AggMicro})
	if err != nil {
		t.Fatal(err)
	}
	// The denominator is the 8 valid trials, not the 9 trials that ran.
	if v.Denominator != 8 {
		t.Errorf("denominator = %v, want 8 (the platform failure must be excluded)", v.Denominator)
	}
	if v.Numerator != 4 {
		t.Errorf("numerator = %v, want 4", v.Numerator)
	}
}

func TestPassAtKIsNoCoverageWhenNoTaskHasEnoughTrials(t *testing.T) {
	reg := NewRegistry()
	trials := []domain.TrialOutcome{
		{TrialID: "a", TaskID: "t1", Failure: domain.ClassCompleted, Verdict: domain.Verdict{Combined: domain.VerdictPass}},
	}
	_, v, err := reg.Compute("pass_at_k", MetricInput{Trials: trials, K: 3, Aggregation: AggMacro})
	if err != nil {
		t.Fatal(err)
	}
	if v.HasValue {
		t.Errorf("expected no coverage when no task has k valid trials, got %v", v.Value)
	}
}

func TestCancelledTrialsAreExcludedFromTheDenominator(t *testing.T) {
	reg := NewRegistry()
	trials := []domain.TrialOutcome{
		{TrialID: "a", TaskID: "t1", Failure: domain.ClassCompleted, Verdict: domain.Verdict{Combined: domain.VerdictPass}},
		{TrialID: "b", TaskID: "t1", Failure: domain.ClassCancelled, Verdict: domain.Verdict{Combined: domain.VerdictFail}},
	}
	_, v, err := reg.Compute("success_rate", MetricInput{Trials: trials, Aggregation: AggMicro})
	if err != nil {
		t.Fatal(err)
	}
	if v.Denominator != 1 {
		t.Errorf("denominator = %v, want 1 (a cancelled trial is excluded)", v.Denominator)
	}
}

func TestConfidenceIntervalIsPresentForRatioMetrics(t *testing.T) {
	reg := NewRegistry()
	trials := buildWorkedExample()
	_, v, err := reg.Compute("success_rate", MetricInput{Trials: trials, Aggregation: AggMicro, CIMethod: "wilson"})
	if err != nil {
		t.Fatal(err)
	}
	if v.CI == nil {
		t.Fatal("a ratio metric must carry a confidence interval")
	}
	if v.CI.Lower > v.Value || v.CI.Upper < v.Value {
		t.Errorf("the interval [%v, %v] does not contain the point estimate %v", v.CI.Lower, v.CI.Upper, v.Value)
	}
	if v.CI.Method != "wilson" {
		t.Errorf("interval method = %q, want wilson", v.CI.Method)
	}
}

func TestBootstrapIntervalIsReproducible(t *testing.T) {
	reg := NewRegistry()
	trials := buildWorkedExample()
	_, first, err := reg.Compute("success_rate", MetricInput{Trials: trials, Aggregation: AggMacro, CIMethod: "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	_, second, err := reg.Compute("success_rate", MetricInput{Trials: trials, Aggregation: AggMacro, CIMethod: "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	if first.CI == nil || second.CI == nil {
		t.Fatal("expected an interval")
	}
	if first.CI.Lower != second.CI.Lower || first.CI.Upper != second.CI.Upper {
		t.Errorf("the bootstrap interval is not reproducible: %v vs %v", first.CI, second.CI)
	}
}

func TestARatioIntervalStaysInsideTheUnitRange(t *testing.T) {
	reg := NewRegistry()
	trials := buildWorkedExample()
	// The normal approximation over per-unit values does not know that a
	// proportion cannot leave [0, 1], so the bound is applied by the
	// registry. An interval of [-0.07, 1.07] on a success rate is not a
	// measurement.
	_, v, err := reg.Compute("success_rate", MetricInput{Trials: trials, Aggregation: AggMacro, CIMethod: "normal"})
	if err != nil {
		t.Fatal(err)
	}
	if v.CI == nil {
		t.Fatal("expected an interval")
	}
	if v.CI.Lower < 0 || v.CI.Upper > 1 {
		t.Errorf("the interval [%v, %v] leaves the unit range of a ratio", v.CI.Lower, v.CI.Upper)
	}
}
