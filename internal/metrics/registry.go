// Package metrics resolves every metric the platform reports.
//
// A metric is a registry entry, never an ad-hoc computation at a call site.
// The registry holds a locked built-in set and accepts registered custom
// metrics that satisfy the same completeness rules.
//
// See Docs/Tech Design/TD-003-metric-registry-and-scoring-contract.md.
package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/Baike12/b-eval/internal/domain"
)

// Unit is the measurement unit of a metric. A ratio metric always carries a
// confidence interval.
type Unit string

const (
	UnitRatio Unit = "ratio"
	UnitCount Unit = "count"
	UnitScore Unit = "score"
)

// Aggregation names the way a metric is aggregated over tasks and trials.
const (
	AggMacro = "macro"
	AggMicro = "micro"
)

// SupportedCIMethods lists the confidence-interval methods the platform
// implements.
var SupportedCIMethods = []string{"wilson", "normal", "bootstrap"}

// Interval is a confidence interval with the method that produced it.
//
// Method is the method that was actually applied; Requested is the method the
// run asked for. The two differ when the requested method does not apply to
// the metric's unit of observation, and the report shows both so a reader is
// never misled about which interval they are looking at.
type Interval struct {
	Lower     float64 `json:"lower"`
	Upper     float64 `json:"upper"`
	Method    string  `json:"method"`
	Requested string  `json:"requested"`
	Basis     string  `json:"basis"`
	Level     float64 `json:"level"`
}

// MetricInput is what a metric computation receives.
type MetricInput struct {
	Trials      []domain.TrialOutcome
	K           int
	Aggregation string
	CIMethod    string
}

// MetricValue is the result of one metric computation.
//
// HasValue == false means "no coverage". A metric with no coverage is
// excluded from aggregation and rendered as "no coverage"; it is never
// coerced to zero, because a metric of zero and a metric with no data are
// different facts.
type MetricValue struct {
	Value       float64   `json:"value"`
	HasValue    bool      `json:"has_value"`
	Numerator   float64   `json:"numerator"`
	Denominator float64   `json:"denominator"`
	SampleSize  int       `json:"sample_size"`
	CI          *Interval `json:"ci,omitempty"`
}

// MetricDef is one registry entry.
//
// The Compute function is excluded from the JSON form: a registry entry is
// shown to readers of the API, and a function value cannot be serialised.
type MetricDef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Builtin     bool   `json:"builtin"`
	Unit        Unit   `json:"unit"`
	// Aggregations lists the aggregation methods the metric supports.
	Aggregations []string `json:"aggregations"`
	// DefaultAggregation is the method used when a run does not choose one.
	DefaultAggregation string `json:"default_aggregation"`
	Numerator          string `json:"numerator"`
	Denominator        string `json:"denominator"`
	Formula            string `json:"formula"`
	CIMethod           string `json:"ci_method"`
	Version            string `json:"version"`
	// Compute evaluates the metric over a set of trials.
	Compute func(in MetricInput) (MetricValue, error) `json:"-"`
}

// RegistryError is a refusal to register or resolve a metric.
type RegistryError struct {
	MetricID string
	Reason   string
	Fields   []string
}

func (e *RegistryError) Error() string {
	if len(e.Fields) > 0 {
		return fmt.Sprintf("metric %q: %s (missing: %s)", e.MetricID, e.Reason, strings.Join(e.Fields, ", "))
	}
	return fmt.Sprintf("metric %q: %s", e.MetricID, e.Reason)
}

// Registry holds the metric definitions.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]MetricDef
	order   []string
}

// NewRegistry returns a registry holding the built-in metrics.
//
// The built-in entries are registered through Register itself rather than
// through a private path, so the completeness gate that protects custom
// registrations also protects the built-in set.
func NewRegistry() *Registry {
	r := &Registry{entries: map[string]MetricDef{}}
	for _, def := range builtinMetrics() {
		// The built-in set is known-good, so a rejection here is a
		// programming error rather than a user error.
		if err := r.Register(def); err != nil {
			panic(fmt.Sprintf("built-in metric %q is invalid: %v", def.ID, err))
		}
	}
	return r
}

// Register adds a metric definition.
//
// The rejection rules, in order: empty or malformed ID, an ID that is already
// taken, missing mandatory fields, a ratio metric without a CI method, and
// an unsupported CI method. Every missing field is reported at once so the
// author fixes the registration in a single pass.
func (r *Registry) Register(def MetricDef) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.TrimSpace(def.ID) == "" {
		return &RegistryError{MetricID: def.ID, Reason: "metric ID is required"}
	}
	if strings.ContainsAny(def.ID, " \t") {
		return &RegistryError{MetricID: def.ID, Reason: "metric ID must not contain whitespace"}
	}
	if existing, ok := r.entries[def.ID]; ok {
		if existing.Builtin {
			return &RegistryError{MetricID: def.ID, Reason: "built-in metric ID is reserved"}
		}
		return &RegistryError{MetricID: def.ID, Reason: "metric ID is already registered"}
	}

	var missing []string
	if strings.TrimSpace(def.Name) == "" {
		missing = append(missing, "name")
	}
	if strings.TrimSpace(def.Description) == "" {
		missing = append(missing, "description")
	}
	if def.Unit == "" {
		missing = append(missing, "unit")
	}
	if len(def.Aggregations) == 0 {
		missing = append(missing, "aggregations")
	}
	if strings.TrimSpace(def.Numerator) == "" {
		missing = append(missing, "numerator")
	}
	if strings.TrimSpace(def.Denominator) == "" {
		missing = append(missing, "denominator")
	}
	if strings.TrimSpace(def.Formula) == "" {
		missing = append(missing, "formula")
	}
	if def.Compute == nil {
		missing = append(missing, "compute")
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return &RegistryError{MetricID: def.ID, Reason: "registration is incomplete", Fields: missing}
	}

	if def.Unit == UnitRatio {
		if strings.TrimSpace(def.CIMethod) == "" {
			return &RegistryError{MetricID: def.ID, Reason: "a ratio metric must declare a CI method", Fields: []string{"ci_method"}}
		}
		if !isSupportedCIMethod(def.CIMethod) {
			return &RegistryError{
				MetricID: def.ID,
				Reason:   fmt.Sprintf("CI method %q is not supported; use one of %s", def.CIMethod, strings.Join(SupportedCIMethods, ", ")),
			}
		}
	}

	for _, a := range def.Aggregations {
		if a != AggMacro && a != AggMicro {
			return &RegistryError{MetricID: def.ID, Reason: fmt.Sprintf("aggregation %q is not supported", a)}
		}
	}
	if def.DefaultAggregation == "" {
		def.DefaultAggregation = def.Aggregations[0]
	}
	if !contains(def.Aggregations, def.DefaultAggregation) {
		return &RegistryError{
			MetricID: def.ID,
			Reason:   fmt.Sprintf("default aggregation %q is not in the supported list %v", def.DefaultAggregation, def.Aggregations),
		}
	}
	if def.Version == "" {
		def.Version = "1"
	}

	r.entries[def.ID] = def
	r.order = append(r.order, def.ID)
	return nil
}

// Get resolves a metric definition.
func (r *Registry) Get(id string) (MetricDef, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	def, ok := r.entries[id]
	return def, ok
}

// List returns every metric definition, built-in first, in registration
// order within each group.
func (r *Registry) List() []MetricDef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]MetricDef, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.entries[id])
	}
	return out
}

// Compute evaluates a metric over a set of trials.
//
// The requested aggregation is honoured when the metric supports it and
// falls back to the metric's default otherwise, so a run that selects a
// single aggregation for the whole report still gets a defined number.
func (r *Registry) Compute(id string, in MetricInput) (MetricDef, MetricValue, error) {
	def, ok := r.Get(id)
	if !ok {
		return MetricDef{}, MetricValue{}, fmt.Errorf("metric %q is not registered", id)
	}
	if !contains(def.Aggregations, in.Aggregation) {
		in.Aggregation = def.DefaultAggregation
	}
	if in.CIMethod == "" {
		in.CIMethod = def.CIMethod
	}
	if in.CIMethod == "" {
		in.CIMethod = "wilson"
	}
	if in.K <= 0 {
		in.K = 1
	}
	v, err := def.Compute(in)
	if err != nil {
		return def, MetricValue{}, err
	}
	if def.Unit == UnitRatio && v.CI != nil {
		// A proportion cannot leave [0, 1]. The normal approximation over
		// per-unit values does not know that, so the bound is applied here,
		// in the one place every ratio metric passes through.
		v.CI.Lower = clamp(v.CI.Lower, 0, 1)
		v.CI.Upper = clamp(v.CI.Upper, 0, 1)
		if v.CI.Lower > v.CI.Upper {
			v.CI.Lower, v.CI.Upper = v.CI.Upper, v.CI.Lower
		}
	}
	return def, v, nil
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func isSupportedCIMethod(m string) bool {
	return contains(SupportedCIMethods, m)
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
