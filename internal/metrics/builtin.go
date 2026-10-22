package metrics

import (
	"math"

	"github.com/Baike12/b-eval/internal/domain"
)

// taskStats holds the per-task counts the success metrics are computed from.
//
// Only trials whose failure class counts in the denominator are included, so
// a platform failure never turns into a capability deduction.
type taskStats struct {
	taskID    string
	n         int
	c         int
	reward    float64
	passNames []string
}

// groupTrials aggregates the denominator trials by task.
func groupTrials(trials []domain.TrialOutcome) []*taskStats {
	byID := map[string]*taskStats{}
	var order []string
	for _, t := range trials {
		if !t.Failure.CountsInDenominator() {
			continue
		}
		g, ok := byID[t.TaskID]
		if !ok {
			g = &taskStats{taskID: t.TaskID}
			byID[t.TaskID] = g
			order = append(order, t.TaskID)
		}
		g.n++
		if t.Passed() {
			g.c++
		}
		g.reward += t.Verdict.Score
		g.passNames = append(g.passNames, t.ActionNames...)
	}
	out := make([]*taskStats, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// builtinMetrics returns the platform's own metric definitions.
//
// The built-in set is locked: a task, a task set or a run cannot change a
// built-in metric's formula, denominator or aggregation method. Locking them
// is what makes a score comparable across platforms.
func builtinMetrics() []MetricDef {
	return []MetricDef{
		{
			ID:                 "success_rate",
			Name:               "Success rate",
			Description:        "Share of valid trials whose primary criterion was satisfied.",
			Builtin:            true,
			Unit:               UnitRatio,
			Aggregations:       []string{AggMacro, AggMicro},
			DefaultAggregation: AggMacro,
			Numerator:          "Valid trials judged as passing",
			Denominator:        "Valid trials",
			Formula:            "macro: mean over tasks of (passing trials / valid trials); micro: passing trials / valid trials",
			CIMethod:           "wilson",
			Version:            "1",
			Compute:            computeSuccessRate,
		},
		{
			ID:                 "pass_at_k",
			Name:               "pass@k",
			Description:        "Probability that a task is solved at least once within k attempts.",
			Builtin:            true,
			Unit:               UnitRatio,
			Aggregations:       []string{AggMacro},
			DefaultAggregation: AggMacro,
			Numerator:          "Tasks with at least one passing trial among those with at least k valid trials",
			Denominator:        "Tasks with at least k valid trials",
			Formula:            "mean over tasks of (1 - prod_{i=0..k-1} (n_t - c_t - i)/(n_t - i))",
			CIMethod:           "wilson",
			Version:            "1",
			Compute:            computePassAtK,
		},
		{
			ID:                 "pass_power_k",
			Name:               "pass^k",
			Description:        "Probability that a task is solved in every one of k attempts.",
			Builtin:            true,
			Unit:               UnitRatio,
			Aggregations:       []string{AggMacro},
			DefaultAggregation: AggMacro,
			Numerator:          "Tasks solved in all k trials",
			Denominator:        "Tasks with at least k valid trials",
			Formula:            "mean over tasks of C(c_t, k) / C(n_t, k)",
			CIMethod:           "wilson",
			Version:            "1",
			Compute:            computePassPowerK,
		},
		{
			ID:                 "avg_at_k",
			Name:               "avg@k",
			Description:        "Average score over all valid trials.",
			Builtin:            true,
			Unit:               UnitScore,
			Aggregations:       []string{AggMicro},
			DefaultAggregation: AggMicro,
			Numerator:          "Sum of trial scores",
			Denominator:        "Valid trials",
			Formula:            "sum(trial scores) / count(valid trials)",
			CIMethod:           "normal",
			Version:            "1",
			Compute:            computeAvgAtK,
		},
		{
			ID:                 "task_pass_rate",
			Name:               "Task pass rate",
			Description:        "Share of tasks that were solved at least once.",
			Builtin:            true,
			Unit:               UnitRatio,
			Aggregations:       []string{AggMacro},
			DefaultAggregation: AggMacro,
			Numerator:          "Tasks with at least one passing trial",
			Denominator:        "Tasks with at least one valid trial",
			Formula:            "count(tasks with c_t >= 1) / count(tasks with n_t >= 1)",
			CIMethod:           "wilson",
			Version:            "1",
			Compute:            computeTaskPassRate,
		},
		{
			ID:                 "progress_score",
			Name:               "Progress score",
			Description:        "Average continuous score over the valid trials.",
			Builtin:            true,
			Unit:               UnitScore,
			Aggregations:       []string{AggMicro},
			DefaultAggregation: AggMicro,
			Numerator:          "Sum of trial scores",
			Denominator:        "Valid trials",
			Formula:            "sum(trial scores) / count(valid trials)",
			CIMethod:           "normal",
			Version:            "1",
			Compute:            computeProgressScore,
		},
		{
			ID:                 "platform_failure_rate",
			Name:               "Platform failure rate",
			Description:        "Share of trials excluded from the denominator because the platform failed.",
			Builtin:            true,
			Unit:               UnitRatio,
			Aggregations:       []string{AggMicro},
			DefaultAggregation: AggMicro,
			Numerator:          "Trials classified as platform failures",
			Denominator:        "All trials",
			Formula:            "count(platform failures) / count(all trials)",
			CIMethod:           "wilson",
			Version:            "1",
			Compute:            computePlatformFailureRate,
		},
		{
			ID:                 "judgement_failure_rate",
			Name:               "Judgement failure rate",
			Description:        "Share of trials whose verdict is UNSCORED.",
			Builtin:            true,
			Unit:               UnitRatio,
			Aggregations:       []string{AggMicro},
			DefaultAggregation: AggMicro,
			Numerator:          "Trials whose verdict is UNSCORED",
			Denominator:        "Trials that were judged",
			Formula:            "count(unscored verdicts) / count(judged trials)",
			CIMethod:           "wilson",
			Version:            "1",
			Compute:            computeJudgementFailureRate,
		},
		{
			ID:                 "avg_steps_per_success",
			Name:               "Average steps per successful task",
			Description:        "Mean number of actions over the trials judged as passing.",
			Builtin:            true,
			Unit:               UnitScore,
			Aggregations:       []string{AggMicro},
			DefaultAggregation: AggMicro,
			Numerator:          "Sum of action counts over passing trials",
			Denominator:        "Passing trials",
			Formula:            "sum(steps of passing trials) / count(passing trials)",
			CIMethod:           "normal",
			Version:            "1",
			Compute:            computeAvgStepsPerSuccess,
		},
		{
			ID:                 "avg_tokens_per_success",
			Name:               "Average tokens per successful task",
			Description:        "Mean input plus output tokens over the trials judged as passing.",
			Builtin:            true,
			Unit:               UnitScore,
			Aggregations:       []string{AggMicro},
			DefaultAggregation: AggMicro,
			Numerator:          "Sum of tokens over passing trials",
			Denominator:        "Passing trials",
			Formula:            "sum(tokens of passing trials) / count(passing trials)",
			CIMethod:           "normal",
			Version:            "1",
			Compute:            computeAvgTokensPerSuccess,
		},
		{
			ID:                 "avg_cost_per_success",
			Name:               "Average cost per successful task",
			Description:        "Mean cost over the trials judged as passing.",
			Builtin:            true,
			Unit:               UnitScore,
			Aggregations:       []string{AggMicro},
			DefaultAggregation: AggMicro,
			Numerator:          "Sum of cost over passing trials",
			Denominator:        "Passing trials",
			Formula:            "sum(cost of passing trials) / count(passing trials)",
			CIMethod:           "normal",
			Version:            "1",
			Compute:            computeAvgCostPerSuccess,
		},
		{
			ID:                 "invalid_action_rate",
			Name:               "Invalid action rate",
			Description:        "Share of agent responses that could not be parsed into a legal action.",
			Builtin:            true,
			Unit:               UnitRatio,
			Aggregations:       []string{AggMicro},
			DefaultAggregation: AggMicro,
			Numerator:          "Responses with a format error",
			Denominator:        "All responses",
			Formula:            "count(format errors) / count(responses)",
			CIMethod:           "wilson",
			Version:            "1",
			Compute:            computeInvalidActionRate,
		},
		{
			ID:                 "repeated_action_rate",
			Name:               "Repeated action rate",
			Description:        "Share of actions that repeat the previous action.",
			Builtin:            true,
			Unit:               UnitRatio,
			Aggregations:       []string{AggMicro},
			DefaultAggregation: AggMicro,
			Numerator:          "Actions identical to the preceding action",
			Denominator:        "All actions",
			Formula:            "count(repeated actions) / count(actions)",
			CIMethod:           "wilson",
			Version:            "1",
			Compute:            computeRepeatedActionRate,
		},
		{
			ID:                 "grounding_accuracy",
			Name:               "Grounding accuracy",
			Description:        "Share of actions the environment accepted.",
			Builtin:            true,
			Unit:               UnitRatio,
			Aggregations:       []string{AggMicro},
			DefaultAggregation: AggMicro,
			Numerator:          "Actions accepted by the environment",
			Denominator:        "All actions",
			Formula:            "count(accepted actions) / count(actions)",
			CIMethod:           "wilson",
			Version:            "1",
			Compute:            computeGroundingAccuracy,
		},
	}
}

func computeSuccessRate(in MetricInput) (MetricValue, error) {
	groups := groupTrials(in.Trials)
	if len(groups) == 0 {
		return MetricValue{HasValue: false}, nil
	}
	totalN, totalC := 0, 0
	rates := make([]float64, 0, len(groups))
	for _, g := range groups {
		totalN += g.n
		totalC += g.c
		rates = append(rates, float64(g.c)/float64(g.n))
	}
	if totalN == 0 {
		return MetricValue{HasValue: false}, nil
	}
	if in.Aggregation == AggMicro {
		return MetricValue{
			Value:       float64(totalC) / float64(totalN),
			HasValue:    true,
			Numerator:   float64(totalC),
			Denominator: float64(totalN),
			SampleSize:  totalN,
			CI:          ciProportion(in.CIMethod, float64(totalC), float64(totalN)),
		}, nil
	}
	return MetricValue{
		Value:       meanOf(rates),
		HasValue:    true,
		Numerator:   float64(totalC),
		Denominator: float64(totalN),
		SampleSize:  len(groups),
		CI:          ciValues(in.CIMethod, rates),
	}, nil
}

func computePassAtK(in MetricInput) (MetricValue, error) {
	groups := groupTrials(in.Trials)
	values := make([]float64, 0, len(groups))
	totalN, totalC := 0, 0
	for _, g := range groups {
		if g.n < in.K {
			// A task with fewer valid trials than k has no pass@k value;
			// it is excluded from the average and reported as coverage
			// rather than silently counted as zero.
			continue
		}
		values = append(values, passAtKFor(g.n, g.c, in.K))
		totalN += g.n
		totalC += g.c
	}
	if len(values) == 0 {
		return MetricValue{HasValue: false}, nil
	}
	return MetricValue{
		Value:       meanOf(values),
		HasValue:    true,
		Numerator:   float64(totalC),
		Denominator: float64(totalN),
		SampleSize:  len(values),
		CI:          ciValues(in.CIMethod, values),
	}, nil
}

func computePassPowerK(in MetricInput) (MetricValue, error) {
	groups := groupTrials(in.Trials)
	values := make([]float64, 0, len(groups))
	totalN, totalC := 0, 0
	for _, g := range groups {
		if g.n < in.K {
			continue
		}
		values = append(values, passPowerKFor(g.n, g.c, in.K))
		totalN += g.n
		totalC += g.c
	}
	if len(values) == 0 {
		return MetricValue{HasValue: false}, nil
	}
	return MetricValue{
		Value:       meanOf(values),
		HasValue:    true,
		Numerator:   float64(totalC),
		Denominator: float64(totalN),
		SampleSize:  len(values),
		CI:          ciValues(in.CIMethod, values),
	}, nil
}

// passAtKFor computes the unbiased estimator of "at least one success in k
// attempts" for one task.
func passAtKFor(n, c, k int) float64 {
	if n < k || k <= 0 {
		return 0
	}
	if c == n {
		return 1
	}
	if c == 0 {
		return 0
	}
	prod := 1.0
	for i := 0; i < k; i++ {
		prod *= float64(n-c-i) / float64(n-i)
	}
	return 1 - prod
}

// passPowerKFor computes the probability that all k attempts succeed.
func passPowerKFor(n, c, k int) float64 {
	if n < k || k <= 0 {
		return 0
	}
	if c < k {
		return 0
	}
	return float64(choose(c, k)) / float64(choose(n, k))
}

// choose computes the binomial coefficient C(n, k).
func choose(n, k int) float64 {
	if k < 0 || k > n {
		return 0
	}
	result := 1.0
	for i := 0; i < k; i++ {
		result = result * float64(n-i) / float64(i+1)
	}
	return math.Round(result)
}

func computeAvgAtK(in MetricInput) (MetricValue, error) {
	groups := groupTrials(in.Trials)
	totalN, totalReward := 0, 0.0
	for _, g := range groups {
		totalN += g.n
		totalReward += g.reward
	}
	if totalN == 0 {
		return MetricValue{HasValue: false}, nil
	}
	return MetricValue{
		Value:       totalReward / float64(totalN),
		HasValue:    true,
		Numerator:   totalReward,
		Denominator: float64(totalN),
		SampleSize:  totalN,
	}, nil
}

func computeProgressScore(in MetricInput) (MetricValue, error) {
	return computeAvgAtK(in)
}

func computeTaskPassRate(in MetricInput) (MetricValue, error) {
	groups := groupTrials(in.Trials)
	if len(groups) == 0 {
		return MetricValue{HasValue: false}, nil
	}
	solved := 0
	for _, g := range groups {
		if g.c >= 1 {
			solved++
		}
	}
	return MetricValue{
		Value:       float64(solved) / float64(len(groups)),
		HasValue:    true,
		Numerator:   float64(solved),
		Denominator: float64(len(groups)),
		SampleSize:  len(groups),
		CI:          ciProportion(in.CIMethod, float64(solved), float64(len(groups))),
	}, nil
}

func computePlatformFailureRate(in MetricInput) (MetricValue, error) {
	if len(in.Trials) == 0 {
		return MetricValue{HasValue: false}, nil
	}
	failures := 0
	for _, t := range in.Trials {
		if t.Failure == domain.ClassPlatformFailure {
			failures++
		}
	}
	return MetricValue{
		Value:       float64(failures) / float64(len(in.Trials)),
		HasValue:    true,
		Numerator:   float64(failures),
		Denominator: float64(len(in.Trials)),
		SampleSize:  len(in.Trials),
		CI:          ciProportion(in.CIMethod, float64(failures), float64(len(in.Trials))),
	}, nil
}

func computeJudgementFailureRate(in MetricInput) (MetricValue, error) {
	judged, unscored := 0, 0
	for _, t := range in.Trials {
		if t.Failure != domain.ClassCompleted {
			continue
		}
		judged++
		if t.Verdict.Combined == domain.VerdictUnscored {
			unscored++
		}
	}
	if judged == 0 {
		return MetricValue{HasValue: false}, nil
	}
	return MetricValue{
		Value:       float64(unscored) / float64(judged),
		HasValue:    true,
		Numerator:   float64(unscored),
		Denominator: float64(judged),
		SampleSize:  judged,
		CI:          ciProportion(in.CIMethod, float64(unscored), float64(judged)),
	}, nil
}

func computeAvgStepsPerSuccess(in MetricInput) (MetricValue, error) {
	return averageOverPassing(in, func(t domain.TrialOutcome) float64 { return float64(t.Steps) })
}

func computeAvgTokensPerSuccess(in MetricInput) (MetricValue, error) {
	return averageOverPassing(in, func(t domain.TrialOutcome) float64 { return float64(t.Usage.TotalTokens()) })
}

func computeAvgCostPerSuccess(in MetricInput) (MetricValue, error) {
	return averageOverPassing(in, func(t domain.TrialOutcome) float64 { return t.Usage.CostUSD })
}

func averageOverPassing(in MetricInput, extract func(domain.TrialOutcome) float64) (MetricValue, error) {
	sum, count := 0.0, 0
	for _, t := range in.Trials {
		if !t.Failure.CountsInDenominator() || !t.Passed() {
			continue
		}
		sum += extract(t)
		count++
	}
	if count == 0 {
		return MetricValue{HasValue: false}, nil
	}
	return MetricValue{
		Value:       sum / float64(count),
		HasValue:    true,
		Numerator:   sum,
		Denominator: float64(count),
		SampleSize:  count,
	}, nil
}

func computeInvalidActionRate(in MetricInput) (MetricValue, error) {
	errors, responses := 0, 0
	for _, t := range in.Trials {
		if !t.Failure.CountsInDenominator() {
			continue
		}
		errors += t.FormatErrors
		responses += t.Steps
	}
	if responses == 0 {
		return MetricValue{HasValue: false}, nil
	}
	return MetricValue{
		Value:       float64(errors) / float64(responses),
		HasValue:    true,
		Numerator:   float64(errors),
		Denominator: float64(responses),
		SampleSize:  responses,
		CI:          ciProportion(in.CIMethod, float64(errors), float64(responses)),
	}, nil
}

func computeRepeatedActionRate(in MetricInput) (MetricValue, error) {
	repeated, actions := 0, 0
	for _, t := range in.Trials {
		if !t.Failure.CountsInDenominator() {
			continue
		}
		repeated += t.RepeatedActions
		actions += t.Steps
	}
	if actions == 0 {
		return MetricValue{HasValue: false}, nil
	}
	return MetricValue{
		Value:       float64(repeated) / float64(actions),
		HasValue:    true,
		Numerator:   float64(repeated),
		Denominator: float64(actions),
		SampleSize:  actions,
		CI:          ciProportion(in.CIMethod, float64(repeated), float64(actions)),
	}, nil
}

func computeGroundingAccuracy(in MetricInput) (MetricValue, error) {
	accepted, actions := 0, 0
	for _, t := range in.Trials {
		if !t.Failure.CountsInDenominator() {
			continue
		}
		accepted += t.Steps - t.FormatErrors
		actions += t.Steps
	}
	if actions == 0 {
		return MetricValue{HasValue: false}, nil
	}
	return MetricValue{
		Value:       float64(accepted) / float64(actions),
		HasValue:    true,
		Numerator:   float64(accepted),
		Denominator: float64(actions),
		SampleSize:  actions,
		CI:          ciProportion(in.CIMethod, float64(accepted), float64(actions)),
	}, nil
}
