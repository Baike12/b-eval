package metrics

import (
	"fmt"
	"strings"

	"github.com/Baike12/b-eval/internal/domain"
)

// CustomKind is a computation the platform offers for registration as a
// custom metric.
//
// The registry accepts a metric registration, not arbitrary code: a
// registration names one of these kinds and supplies its parameters. That
// keeps the second definition path reviewable and keeps an expression
// language out of the report path.
type CustomKind struct {
	// Name is the kind identifier a registration uses.
	Name string `json:"name"`
	// Description explains what the kind computes.
	Description string `json:"description"`
	// RequiredParams lists the parameters a registration must supply.
	RequiredParams []string `json:"required_params"`
}

// CustomKinds lists the computations available for registration.
func CustomKinds() []CustomKind {
	return []CustomKind{
		{
			Name:           "keyword_coverage",
			Description:    "Share of trials whose answer contains every required keyword.",
			RequiredParams: []string{"keywords"},
		},
		{
			Name:           "answer_length_band",
			Description:    "Share of trials whose answer length falls inside the configured band.",
			RequiredParams: []string{"min_chars", "max_chars"},
		},
		{
			Name:           "exact_reference_match",
			Description:    "Share of trials whose answer equals the task's reference answer.",
			RequiredParams: []string{},
		},
		{
			Name:           "step_budget_respected",
			Description:    "Share of trials that finished within the task's declared step budget.",
			RequiredParams: []string{},
		},
	}
}

// CustomKindByName resolves a custom kind.
func CustomKindByName(name string) (CustomKind, bool) {
	for _, k := range CustomKinds() {
		if k.Name == name {
			return k, true
		}
	}
	return CustomKind{}, false
}

// BuildCustom assembles a registry entry for a custom metric registration.
//
// The entry goes through the same completeness gate as every other entry,
// so a custom metric can never be less well specified than a built-in one.
func BuildCustom(reg CustomRequest) (MetricDef, error) {
	kind, ok := CustomKindByName(reg.Kind)
	if !ok {
		return MetricDef{}, fmt.Errorf("unknown custom metric kind %q", reg.Kind)
	}
	for _, p := range kind.RequiredParams {
		if _, present := reg.Params[p]; !present {
			return MetricDef{}, fmt.Errorf("custom metric kind %q requires the parameter %q", reg.Kind, p)
		}
	}

	def := MetricDef{
		ID:                 reg.ID,
		Name:               reg.Name,
		Description:        reg.Description,
		Builtin:            false,
		Unit:               reg.Unit,
		Aggregations:       reg.Aggregations,
		DefaultAggregation: reg.DefaultAggregation,
		Numerator:          reg.Numerator,
		Denominator:        reg.Denominator,
		Formula:            reg.Formula,
		CIMethod:           reg.CIMethod,
		Version:            "1",
	}
	if def.Description == "" {
		def.Description = kind.Description
	}
	if def.Formula == "" {
		def.Formula = kind.Description
	}

	switch reg.Kind {
	case "keyword_coverage":
		keywords := stringList(reg.Params, "keywords")
		def.Compute = func(in MetricInput) (MetricValue, error) {
			return ratioOverTrials(in, func(t domain.TrialOutcome) (bool, float64) {
				answer := strings.ToLower(evidenceString(t, "answer"))
				return containsAll(answer, keywords), 0
			})
		}
	case "answer_length_band":
		minChars, maxChars := intParam(reg.Params, "min_chars"), intParam(reg.Params, "max_chars")
		def.Compute = func(in MetricInput) (MetricValue, error) {
			return ratioOverTrials(in, func(t domain.TrialOutcome) (bool, float64) {
				n := len(evidenceString(t, "answer"))
				return n >= minChars && n <= maxChars, 0
			})
		}
	case "exact_reference_match":
		def.Compute = func(in MetricInput) (MetricValue, error) {
			return ratioOverTrials(in, func(t domain.TrialOutcome) (bool, float64) {
				return strings.TrimSpace(evidenceString(t, "answer")) == strings.TrimSpace(evidenceString(t, "reference")), 0
			})
		}
	case "step_budget_respected":
		def.Compute = func(in MetricInput) (MetricValue, error) {
			return ratioOverTrials(in, func(t domain.TrialOutcome) (bool, float64) {
				return t.BudgetSteps == 0 || t.Steps <= t.BudgetSteps, 0
			})
		}
	}
	return def, nil
}

// CustomRequest is the registration a caller submits.
type CustomRequest struct {
	ID                 string
	Name               string
	Description        string
	Kind               string
	Unit               Unit
	Aggregations       []string
	DefaultAggregation string
	Numerator          string
	Denominator        string
	Formula            string
	CIMethod           string
	Params             map[string]any
}

// ratioOverTrials computes a ratio metric over the denominator trials.
//
// The success function decides the numerator; the score function is used for
// the per-trial continuous value when the caller supplies one.
func ratioOverTrials(in MetricInput, success func(domain.TrialOutcome) (bool, float64)) (MetricValue, error) {
	var successes, total int
	for _, t := range in.Trials {
		if !t.Failure.CountsInDenominator() {
			continue
		}
		total++
		if ok, _ := success(t); ok {
			successes++
		}
	}
	if total == 0 {
		return MetricValue{HasValue: false}, nil
	}
	if in.Aggregation == AggMicro {
		return MetricValue{
			Value:       float64(successes) / float64(total),
			HasValue:    true,
			Numerator:   float64(successes),
			Denominator: float64(total),
			SampleSize:  total,
			CI:          ciProportion(in.CIMethod, float64(successes), float64(total)),
		}, nil
	}
	return MetricValue{
		Value:       float64(successes) / float64(total),
		HasValue:    true,
		Numerator:   float64(successes),
		Denominator: float64(total),
		SampleSize:  total,
		CI:          ciProportion(in.CIMethod, float64(successes), float64(total)),
	}, nil
}

func containsAll(answer string, needles []string) bool {
	for _, n := range needles {
		if !strings.Contains(answer, strings.ToLower(n)) {
			return false
		}
	}
	return len(needles) > 0
}

func coverageRatio(answer string, needles []string) float64 {
	if len(needles) == 0 {
		return 0
	}
	found := 0
	for _, n := range needles {
		if strings.Contains(answer, strings.ToLower(n)) {
			found++
		}
	}
	return float64(found) / float64(len(needles))
}

// evidenceString reads a string out of a trial's verdict evidence.
func evidenceString(t domain.TrialOutcome, key string) string {
	if t.Verdict.Evidence == nil {
		return ""
	}
	s, _ := t.Verdict.Evidence[key].(string)
	return s
}

// intParam reads an integer parameter.
func intParam(params map[string]any, key string) int {
	switch v := params[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		var n int
		fmt.Sscanf(v, "%d", &n)
		return n
	}
	return 0
}

// stringList reads a string list parameter out of a params map.
func stringList(params map[string]any, key string) []string {
	raw, ok := params[key]
	if !ok {
		return nil
	}
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(list))
	for _, v := range list {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
