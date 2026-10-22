package metrics

import (
	"strings"
	"testing"
)

func customDef(id string) MetricDef {
	return MetricDef{
		ID:                 id,
		Name:               "Custom",
		Description:        "A registered metric.",
		Builtin:            false,
		Unit:               UnitRatio,
		Aggregations:       []string{AggMacro, AggMicro},
		DefaultAggregation: AggMacro,
		Numerator:          "Trials",
		Denominator:        "Valid trials",
		Formula:            "count / count",
		CIMethod:           "wilson",
		Version:            "1",
		Compute: func(in MetricInput) (MetricValue, error) {
			return MetricValue{Value: 1, HasValue: true}, nil
		},
	}
}

func TestRegisterRejectsAnIncompleteRegistration(t *testing.T) {
	reg := NewRegistry()
	def := customDef("incomplete")
	def.Unit = ""
	def.Numerator = ""
	if err := reg.Register(def); err == nil {
		t.Fatal("expected the incomplete registration to be refused")
	} else {
		msg := err.Error()
		for _, field := range []string{"unit", "numerator"} {
			if !strings.Contains(msg, field) {
				t.Errorf("the rejection does not name the missing field %q: %s", field, msg)
			}
		}
	}
}

func TestRegisterRejectsARatioMetricWithoutACIMethod(t *testing.T) {
	reg := NewRegistry()
	def := customDef("no-ci")
	def.CIMethod = ""
	if err := reg.Register(def); err == nil {
		t.Fatal("expected a ratio metric without a CI method to be refused")
	}
}

func TestRegisterRejectsARemoteBuiltinID(t *testing.T) {
	reg := NewRegistry()
	def := customDef("success_rate")
	err := reg.Register(def)
	if err == nil {
		t.Fatal("expected the redefinition of a built-in metric to be refused")
	}
	if !strings.Contains(err.Error(), "built-in metric ID is reserved") {
		t.Errorf("unexpected rejection message: %v", err)
	}
}

func TestRegisterRejectsADuplicateID(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(customDef("twice")); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(customDef("twice")); err == nil {
		t.Fatal("expected the duplicate registration to be refused")
	}
}

func TestBuiltinMetricsAreRegisteredAndLocked(t *testing.T) {
	reg := NewRegistry()
	expected := []string{
		"success_rate", "pass_at_k", "pass_power_k", "avg_at_k", "task_pass_rate",
		"progress_score", "platform_failure_rate", "judgement_failure_rate",
		"avg_steps_per_success", "avg_tokens_per_success", "avg_cost_per_success",
		"invalid_action_rate", "repeated_action_rate", "grounding_accuracy",
	}
	for _, id := range expected {
		def, ok := reg.Get(id)
		if !ok {
			t.Errorf("the built-in metric %q is not registered", id)
			continue
		}
		if !def.Builtin {
			t.Errorf("the metric %q is not marked as built-in", id)
		}
		if def.Numerator == "" || def.Denominator == "" || def.Formula == "" {
			t.Errorf("the built-in metric %q does not declare its numerator, denominator and formula", id)
		}
	}
}

func TestRegisterAcceptsACompleteCustomMetric(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(customDef("keyword_coverage")); err != nil {
		t.Fatalf("a complete registration was refused: %v", err)
	}
	def, ok := reg.Get("keyword_coverage")
	if !ok {
		t.Fatal("the registered metric is not in the registry")
	}
	if def.Builtin {
		t.Error("a registered metric must not be marked as built-in")
	}
}

func TestComputeFallsBackToTheDefaultAggregation(t *testing.T) {
	reg := NewRegistry()
	// avg_at_k is a micro-only metric. A run that asks for macro must still
	// get a defined number rather than an error.
	_, _, err := reg.Compute("avg_at_k", MetricInput{Aggregation: AggMacro})
	if err != nil {
		t.Fatalf("expected a fallback, got %v", err)
	}
}

func TestUnknownMetricIsAnError(t *testing.T) {
	reg := NewRegistry()
	if _, _, err := reg.Compute("does_not_exist", MetricInput{}); err == nil {
		t.Fatal("expected an unknown metric to be an error")
	}
}

func TestCustomKindMustExist(t *testing.T) {
	if _, err := BuildCustom(CustomRequest{ID: "x", Kind: "not_a_kind"}); err == nil {
		t.Fatal("expected an unknown custom kind to be refused")
	}
}

func TestCustomKindRequiresItsParameters(t *testing.T) {
	_, err := BuildCustom(CustomRequest{ID: "x", Kind: "keyword_coverage", Params: map[string]any{}})
	if err == nil {
		t.Fatal("expected the missing parameter to be refused")
	}
	if !strings.Contains(err.Error(), "keywords") {
		t.Errorf("the rejection does not name the missing parameter: %v", err)
	}
}
