// Package report turns a completed run into the report a reader sees.
//
// The report leads with the confidence section, because a number without its
// confidence is not a measurement. The confidence section is always first and
// is never collapsible.
package report

import (
	"fmt"
	"sort"

	"github.com/Baike12/b-eval/internal/domain"
	"github.com/Baike12/b-eval/internal/metrics"
)

// PlatformFailureThreshold is the platform-failure rate above which a report
// is marked as not carrying enough confidence to be published.
const PlatformFailureThreshold = 0.05

// Report is the full single-run report.
type Report struct {
	RunID       string          `json:"run_id"`
	GeneratedAt string          `json:"generated_at"`
	Plan        PlanInfo        `json:"plan"`
	Confidence  Confidence      `json:"confidence"`
	MainScore   MainScore       `json:"main_score"`
	Metrics     []MetricRow     `json:"metrics"`
	Breakdown   []Breakdown     `json:"breakdown"`
	Failures    FailureAnalysis `json:"failures"`
	Tasks       []TaskResult    `json:"tasks"`
}

// PlanInfo identifies the run's configuration.
type PlanInfo struct {
	PlanID         string `json:"plan_id"`
	TaskSetID      string `json:"task_set_id"`
	TaskSetVersion string `json:"task_set_version"`
	AgentID        string `json:"agent_id"`
	Repetitions    int    `json:"repetitions"`
	Aggregation    string `json:"aggregation"`
	CIMethod       string `json:"ci_method"`
	Fingerprint    string `json:"fingerprint"`
}

// Confidence is the section that opens the report.
type Confidence struct {
	PlatformFailureRate    float64 `json:"platform_failure_rate"`
	JudgementFailureRate   float64 `json:"judgement_failure_rate"`
	ValidSampleSize        int     `json:"valid_sample_size"`
	PlannedSampleSize      int     `json:"planned_sample_size"`
	Coverage               float64 `json:"coverage"`
	Threshold              float64 `json:"threshold"`
	InsufficientConfidence bool    `json:"insufficient_confidence"`
	Reason                 string  `json:"reason,omitempty"`
}

// MainScore is the headline number block.
type MainScore struct {
	Aggregation        string             `json:"aggregation"`
	TotalScore         float64            `json:"total_score"`
	HasValue           bool               `json:"has_value"`
	Weights            map[string]float64 `json:"weights"`
	Coverage           float64            `json:"coverage"`
	CI                 *metrics.Interval  `json:"ci,omitempty"`
	PassAtK            float64            `json:"pass_at_k"`
	PassAtKHasValue    bool               `json:"pass_at_k_has_value"`
	PassPowerK         float64            `json:"pass_power_k"`
	PassPowerKHasValue bool               `json:"pass_power_k_has_value"`
	ContinuousMean     float64            `json:"continuous_mean"`
	PrimaryCriterion   string             `json:"primary_criterion"`
}

// MetricRow is one metric as the report shows it.
type MetricRow struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Builtin          bool              `json:"builtin"`
	Unit             string            `json:"unit"`
	Version          string            `json:"version"`
	Formula          string            `json:"formula"`
	Numerator        string            `json:"numerator"`
	Denominator      string            `json:"denominator"`
	Aggregation      string            `json:"aggregation"`
	Value            float64           `json:"value"`
	HasValue         bool              `json:"has_value"`
	Coverage         string            `json:"coverage"`
	NumeratorValue   float64           `json:"numerator_value"`
	DenominatorValue float64           `json:"denominator_value"`
	SampleSize       int               `json:"sample_size"`
	CI               *metrics.Interval `json:"ci,omitempty"`
}

// Breakdown is one decomposition dimension.
type Breakdown struct {
	Dimension string          `json:"dimension"`
	Cells     []BreakdownCell `json:"cells"`
}

// BreakdownCell is one cell of a decomposition dimension.
type BreakdownCell struct {
	Key            string            `json:"key"`
	SampleSize     int               `json:"sample_size"`
	SuccessRate    float64           `json:"success_rate"`
	HasValue       bool              `json:"has_value"`
	Coverage       string            `json:"coverage"`
	ContinuousMean float64           `json:"continuous_mean"`
	CI             *metrics.Interval `json:"ci,omitempty"`
}

// ScorerBreakdown is the PASS/FAIL/UNSCORED split of one scorer.
type ScorerBreakdown struct {
	Scorer   string         `json:"scorer"`
	Pass     int            `json:"pass"`
	Fail     int            `json:"fail"`
	Unscored int            `json:"unscored"`
	Share    map[string]int `json:"share"`
}

// FailedTrial is one failed trial in the failure list.
type FailedTrial struct {
	TrialID    string `json:"trial_id"`
	TaskID     string `json:"task_id"`
	Domain     string `json:"domain"`
	Difficulty string `json:"difficulty"`
	Reason     string `json:"termination_reason"`
	Class      string `json:"failure_class"`
	Detail     string `json:"detail,omitempty"`
}

// FailureAnalysis is the failure section of the report.
type FailureAnalysis struct {
	TerminationReasons map[string]int    `json:"termination_reasons"`
	FailureClasses     map[string]int    `json:"failure_classes"`
	Scorers            []ScorerBreakdown `json:"scorers"`
	FailedTrials       []FailedTrial     `json:"failed_trials"`
}

// TaskResult is the per-task row of the report.
type TaskResult struct {
	TaskID      string  `json:"task_id"`
	DisplayName string  `json:"display_name"`
	Form        string  `json:"form"`
	Domain      string  `json:"domain"`
	Difficulty  string  `json:"difficulty"`
	Trials      int     `json:"trials"`
	Passed      int     `json:"passed"`
	SuccessRate float64 `json:"success_rate"`
	HasValue    bool    `json:"has_value"`
	MeanScore   float64 `json:"mean_score"`
}

// Options configures the report.
type Options struct {
	// Weights are the domain weights of the total score. When empty, every
	// domain present in the run weighs the same.
	Weights map[string]float64
	// K is the repetition count used by pass@k and pass^k.
	K int
	// Aggregation is the aggregation method of the report.
	Aggregation string
	// CIMethod is the confidence-interval method of the report.
	CIMethod string
	// PrimaryCriterion is the sentence describing what "success" means.
	PrimaryCriterion string
}

// Build assembles the report of a run.
func Build(run domain.Run, reg *metrics.Registry, taskNames map[string]string, opts Options) Report {
	if opts.Aggregation == "" {
		opts.Aggregation = metrics.AggMacro
	}
	if opts.CIMethod == "" {
		opts.CIMethod = "wilson"
	}
	if opts.K <= 0 {
		opts.K = run.Plan.Repetitions
	}
	if opts.K <= 0 {
		opts.K = 1
	}

	r := Report{
		RunID:       run.RunID,
		GeneratedAt: domain.Now(),
		Plan: PlanInfo{
			PlanID:         run.Plan.PlanID,
			TaskSetID:      run.Plan.TaskSetID,
			TaskSetVersion: run.Plan.TaskSetVersion,
			AgentID:        run.Plan.AgentID,
			Repetitions:    run.Plan.Repetitions,
			Aggregation:    opts.Aggregation,
			CIMethod:       opts.CIMethod,
			Fingerprint:    run.Plan.Fingerprint(),
		},
	}

	in := metrics.MetricInput{
		Trials:      run.Trials,
		K:           opts.K,
		Aggregation: opts.Aggregation,
		CIMethod:    opts.CIMethod,
	}

	// The metric rows are resolved from the registry, so a report never
	// recomputes a formula at the call site.
	values := map[string]metrics.MetricValue{}
	for _, def := range reg.List() {
		def, v, err := reg.Compute(def.ID, in)
		if err != nil {
			continue
		}
		values[def.ID] = v
		r.Metrics = append(r.Metrics, metricRow(def, v, opts.Aggregation))
	}

	r.Confidence = buildConfidence(run, values, opts)
	r.Breakdown = buildBreakdown(run, reg, in, opts)
	r.MainScore = buildMainScore(run, values, opts, domainCells(r.Breakdown))
	r.Failures = buildFailures(run)
	r.Tasks = buildTaskResults(run, taskNames, reg, in)
	return r
}

// buildConfidence computes the section that opens the report.
func buildConfidence(run domain.Run, values map[string]metrics.MetricValue, opts Options) Confidence {
	c := Confidence{
		Threshold:         PlatformFailureThreshold,
		PlannedSampleSize: len(run.Trials),
	}
	if v, ok := values["platform_failure_rate"]; ok && v.HasValue {
		c.PlatformFailureRate = v.Value
	}
	if v, ok := values["judgement_failure_rate"]; ok && v.HasValue {
		c.JudgementFailureRate = v.Value
	}
	valid := 0
	for _, t := range run.Trials {
		if t.Failure.CountsInDenominator() {
			valid++
		}
	}
	c.ValidSampleSize = valid
	if len(run.Trials) > 0 {
		c.Coverage = float64(valid) / float64(len(run.Trials))
	}
	if c.PlatformFailureRate > c.Threshold {
		c.InsufficientConfidence = true
		c.Reason = fmt.Sprintf("the platform failure rate %.4f exceeds the %.2f threshold", c.PlatformFailureRate, c.Threshold)
	}
	return c
}

// domainCells returns the cells of the domain breakdown.
//
// The total score is the weighted view over the domain axis, so the domain
// cells are what the total is computed from.
func domainCells(breakdown []Breakdown) []BreakdownCell {
	for _, b := range breakdown {
		if b.Dimension == "domain" {
			return b.Cells
		}
	}
	return nil
}

// buildMainScore computes the headline numbers.
//
// The total score is the weighted view of FEAT-002 §2.1.4 over the domain
// axis: every domain that has coverage contributes its success rate,
// weighted, and a domain without coverage is excluded from the weighting
// and reported as a coverage gap rather than as a zero.
//
// The weights are shown next to the total. A total without its weights, its
// aggregation method and its coverage is not a measurement.
func buildMainScore(run domain.Run, values map[string]metrics.MetricValue, opts Options, cells []BreakdownCell) MainScore {
	m := MainScore{
		Aggregation:      opts.Aggregation,
		PrimaryCriterion: opts.PrimaryCriterion,
		Weights:          map[string]float64{},
	}
	for _, cell := range cells {
		weight := 1.0
		if w, ok := opts.Weights[cell.Key]; ok {
			weight = w
		}
		m.Weights[cell.Key] = weight
	}

	coveredWeight, totalWeight := 0.0, 0.0
	weighted := 0.0
	coveredValues := make([]float64, 0, len(cells))
	for _, cell := range cells {
		weight := m.Weights[cell.Key]
		totalWeight += weight
		if !cell.HasValue {
			continue
		}
		coveredWeight += weight
		weighted += weight * cell.SuccessRate
		coveredValues = append(coveredValues, cell.SuccessRate)
	}
	if coveredWeight > 0 {
		m.TotalScore = weighted / coveredWeight
		m.HasValue = true
	}
	if totalWeight > 0 {
		m.Coverage = coveredWeight / totalWeight
	}

	if len(coveredValues) == 1 {
		// A single covered domain is the whole score, so the metric's own
		// interval is the interval of the total.
		if v, ok := values["success_rate"]; ok && v.HasValue {
			m.CI = v.CI
		}
	} else if len(coveredValues) > 1 {
		// Several domains are averaged, so the interval describes the
		// spread across them.
		iv := metrics.NormalInterval(coveredValues)
		m.CI = &iv
	}
	if v, ok := values["pass_at_k"]; ok && v.HasValue {
		m.PassAtK = v.Value
		m.PassAtKHasValue = true
	}
	if v, ok := values["pass_power_k"]; ok && v.HasValue {
		m.PassPowerK = v.Value
		m.PassPowerKHasValue = true
	}
	if v, ok := values["progress_score"]; ok && v.HasValue {
		m.ContinuousMean = v.Value
	}
	return m
}

// buildBreakdown decomposes the score along the fixed dimensions.
//
// A cell with no coverage is reported as "no coverage" and is never filled
// with zero, because a metric of zero and a metric with no data are
// different facts.
func buildBreakdown(run domain.Run, reg *metrics.Registry, in metrics.MetricInput, opts Options) []Breakdown {
	dimensions := []struct {
		name  string
		keyOf func(domain.TrialOutcome) []string
	}{
		{"domain", func(t domain.TrialOutcome) []string { return []string{t.Domain} }},
		{"capability", func(t domain.TrialOutcome) []string { return t.Level1 }},
		{"difficulty", func(t domain.TrialOutcome) []string { return []string{t.Difficulty} }},
		{"form", func(t domain.TrialOutcome) []string { return []string{string(t.Form)} }},
	}

	out := make([]Breakdown, 0, len(dimensions))
	for _, d := range dimensions {
		cells := map[string][]domain.TrialOutcome{}
		for _, t := range run.Trials {
			for _, key := range d.keyOf(t) {
				cells[key] = append(cells[key], t)
			}
		}
		keys := make([]string, 0, len(cells))
		for k := range cells {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		b := Breakdown{Dimension: d.name}
		for _, k := range keys {
			subset := cells[k]
			cellIn := in
			cellIn.Trials = subset
			_, v, err := reg.Compute("success_rate", cellIn)
			if err != nil || !v.HasValue {
				b.Cells = append(b.Cells, BreakdownCell{
					Key:        k,
					SampleSize: len(subset),
					HasValue:   false,
					Coverage:   "no coverage",
				})
				continue
			}
			cell := BreakdownCell{
				Key:         k,
				SampleSize:  v.SampleSize,
				SuccessRate: v.Value,
				HasValue:    true,
				Coverage:    "covered",
				CI:          v.CI,
			}
			if _, cont, err := reg.Compute("progress_score", cellIn); err == nil && cont.HasValue {
				cell.ContinuousMean = cont.Value
			}
			b.Cells = append(b.Cells, cell)
		}
		out = append(out, b)
	}
	return out
}

// buildFailures computes the failure section.
func buildFailures(run domain.Run) FailureAnalysis {
	fa := FailureAnalysis{
		TerminationReasons: map[string]int{},
		FailureClasses:     map[string]int{},
	}
	scorerStats := map[string]*ScorerBreakdown{}
	for _, t := range run.Trials {
		fa.FailureClasses[string(t.Failure)]++
		for _, tr := range run.Trajectories {
			if tr.TrialID != t.TrialID {
				continue
			}
			fa.TerminationReasons[string(tr.Termination)]++
			break
		}
		for _, s := range t.Verdict.Scorers {
			st, ok := scorerStats[s.Type]
			if !ok {
				st = &ScorerBreakdown{Scorer: s.Type, Share: map[string]int{}}
				scorerStats[s.Type] = st
			}
			switch s.Value {
			case domain.VerdictPass:
				st.Pass++
			case domain.VerdictFail:
				st.Fail++
			default:
				st.Unscored++
			}
		}
		if !t.Passed() && t.Failure != domain.ClassCancelled {
			reason := ""
			detail := ""
			for _, tr := range run.Trajectories {
				if tr.TrialID == t.TrialID {
					reason = string(tr.Termination)
					detail = tr.TerminationDetail
					break
				}
			}
			fa.FailedTrials = append(fa.FailedTrials, FailedTrial{
				TrialID:    t.TrialID,
				TaskID:     t.TaskID,
				Domain:     t.Domain,
				Difficulty: t.Difficulty,
				Reason:     reason,
				Class:      string(t.Failure),
				Detail:     detail,
			})
		}
	}
	names := make([]string, 0, len(scorerStats))
	for name := range scorerStats {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fa.Scorers = append(fa.Scorers, *scorerStats[name])
	}
	return fa
}

// buildTaskResults computes the per-task rows.
func buildTaskResults(run domain.Run, taskNames map[string]string, reg *metrics.Registry, in metrics.MetricInput) []TaskResult {
	byTask := map[string][]domain.TrialOutcome{}
	var order []string
	for _, t := range run.Trials {
		if _, ok := byTask[t.TaskID]; !ok {
			order = append(order, t.TaskID)
		}
		byTask[t.TaskID] = append(byTask[t.TaskID], t)
	}
	out := make([]TaskResult, 0, len(order))
	for _, id := range order {
		trials := byTask[id]
		cellIn := in
		cellIn.Trials = trials
		res := TaskResult{
			TaskID:      id,
			DisplayName: taskNames[id],
			Trials:      len(trials),
		}
		if len(trials) > 0 {
			res.Domain = trials[0].Domain
			res.Difficulty = trials[0].Difficulty
			res.Form = string(trials[0].Form)
		}
		passed := 0
		scoreSum := 0.0
		for _, t := range trials {
			if t.Passed() {
				passed++
			}
			scoreSum += t.Verdict.Score
		}
		res.Passed = passed
		if _, v, err := reg.Compute("success_rate", cellIn); err == nil && v.HasValue {
			res.SuccessRate = v.Value
			res.HasValue = true
		}
		if len(trials) > 0 {
			res.MeanScore = scoreSum / float64(len(trials))
		}
		out = append(out, res)
	}
	return out
}

func metricRow(def metrics.MetricDef, v metrics.MetricValue, aggregation string) MetricRow {
	coverage := "covered"
	if !v.HasValue {
		coverage = "no coverage"
	}
	return MetricRow{
		ID:               def.ID,
		Name:             def.Name,
		Description:      def.Description,
		Builtin:          def.Builtin,
		Unit:             string(def.Unit),
		Version:          def.Version,
		Formula:          def.Formula,
		Numerator:        def.Numerator,
		Denominator:      def.Denominator,
		Aggregation:      aggregation,
		Value:            v.Value,
		HasValue:         v.HasValue,
		Coverage:         coverage,
		NumeratorValue:   v.Numerator,
		DenominatorValue: v.Denominator,
		SampleSize:       v.SampleSize,
		CI:               v.CI,
	}
}
