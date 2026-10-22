package domain

import (
	"fmt"
	"sort"
	"strings"
)

// Requirement states how strongly a field group applies to a task form.
type Requirement string

const (
	ReqRequired  Requirement = "required"
	ReqOptional  Requirement = "optional"
	ReqForbidden Requirement = "forbidden"
	ReqPartial   Requirement = "partial"
)

// formContract is the single source of truth for the metadata contract.
//
// The validator and the table live together on purpose: a contract declared
// in one place and enforced in another is a contract that drifts, and the
// drift shows up as a task that passed validation but cannot be run.
//
// See Docs/Tech Design/TD-002-task-forms-and-metadata-contract.md §5.
var formContract = map[TaskForm]map[string]Requirement{
	FormEnvAgentic: {
		"identity":             ReqRequired,
		"capability":           ReqRequired,
		"description":          ReqRequired,
		"budget":               ReqRequired,
		"initial_state":        ReqRequired,
		"interaction_contract": ReqRequired,
		"judgement":            ReqRequired,
		"reference_material":   ReqOptional,
		"meta":                 ReqRequired,
	},
	FormStaticQA: {
		"identity":             ReqRequired,
		"capability":           ReqRequired,
		"description":          ReqRequired,
		"budget":               ReqRequired,
		"initial_state":        ReqForbidden,
		"interaction_contract": ReqPartial,
		"judgement":            ReqRequired,
		"reference_material":   ReqForbidden,
		"meta":                 ReqRequired,
	},
}

// interactionPartial lists the fields a static-QA task may declare inside
// the interaction contract, and the fields it may not.
var interactionPartial = struct {
	Allowed   []string
	Forbidden []string
}{
	Allowed:   []string{"termination_conditions", "multi_turn"},
	Forbidden: []string{"action_space", "observation_shape"},
}

// ValidationError is one reason a task was refused. The message names the
// offending field so the author can fix it without reading the spec.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateTask checks a task against the contract of its own form.
//
// It returns every violation it finds rather than the first one, so an
// author sees the whole list in one pass instead of fixing one field per
// round trip.
func ValidateTask(t Task) []ValidationError {
	var errs []ValidationError
	add := func(field, format string, args ...any) {
		errs = append(errs, ValidationError{Field: field, Message: fmt.Sprintf(format, args...)})
	}

	if !t.Form.Valid() {
		add("form", "unknown task form %q; expected %q or %q", t.Form, FormEnvAgentic, FormStaticQA)
		return errs
	}
	if strings.TrimSpace(t.TaskID) == "" {
		add("task_id", "task ID is required")
	}
	if strings.TrimSpace(t.DisplayName) == "" {
		add("display_name", "display name is required")
	}
	if strings.TrimSpace(t.Description.Instruction) == "" {
		add("description.instruction", "instruction is required")
	}
	if len(t.Capability.Level1) == 0 {
		add("capability.level1_dimensions", "at least one level-1 capability dimension is required")
	}
	if len(t.Capability.Domains) == 0 {
		add("capability.domains", "at least one domain is required")
	}
	if strings.TrimSpace(t.Capability.Difficulty) == "" {
		add("capability.difficulty", "difficulty tier is required")
	}
	if len(t.Judgement.Scorers) == 0 {
		add("judgement.scorers", "at least one scorer reference is required")
	}
	if t.Judgement.Primary < 0 || t.Judgement.Primary >= len(t.Judgement.Scorers) {
		add("judgement.primary", "primary criterion index %d is out of range", t.Judgement.Primary)
	}
	if t.Judgement.Combination != "and" && t.Judgement.Combination != "or" {
		add("judgement.combination", "combination must be %q or %q", "and", "or")
	}

	contract := formContract[t.Form]
	for _, group := range sortedGroups(contract) {
		requirement := contract[group]
		present, fields := groupState(t, group)
		switch requirement {
		case ReqForbidden:
			if present {
				add(group, "field group %q is not used by form %q and must be absent", group, t.Form)
			}
		case ReqRequired:
			if !present {
				add(group, "field group %q is required by form %q", group, t.Form)
			}
		case ReqPartial:
			if !present {
				add(group, "field group %q is required by form %q", group, t.Form)
				continue
			}
			for _, f := range interactionPartial.Allowed {
				if strings.TrimSpace(fields[f]) == "" {
					add(group+"."+f, "field %q is required by form %q", f, t.Form)
				}
			}
			for _, f := range interactionPartial.Forbidden {
				if strings.TrimSpace(fields[f]) != "" {
					add(group+"."+f, "field %q is not used by form %q and must be absent", f, t.Form)
				}
			}
		}
	}

	if t.Form == FormStaticQA && strings.TrimSpace(t.Judgement.ReferenceAnswer) == "" {
		add("judgement.reference_answer", "a static-QA task requires a reference answer")
	}

	return errs
}

// groupState reports whether a field group is populated, and returns the
// group's scalar fields so the partial requirement can be checked field by
// field.
func groupState(t Task, group string) (bool, map[string]string) {
	fields := map[string]string{}
	switch group {
	case "identity":
		return t.TaskID != "" && t.DisplayName != "" && t.Benchmark != "" && t.TaskSetVersion != "", fields
	case "capability":
		return len(t.Capability.Level1) > 0 && len(t.Capability.Domains) > 0 && t.Capability.Difficulty != "", fields
	case "description":
		return t.Description.Instruction != "", fields
	case "budget":
		return true, fields
	case "initial_state":
		return t.InitialState != nil, fields
	case "interaction_contract":
		if t.InteractionContract == nil {
			return false, fields
		}
		fields["action_space"] = strings.Join(t.InteractionContract.ActionSpace, ",")
		fields["observation_shape"] = t.InteractionContract.ObservationShape
		fields["termination_conditions"] = strings.Join(t.InteractionContract.TerminationConditions, ",")
		if t.InteractionContract.MultiTurn != nil {
			fields["multi_turn"] = "declared"
		}
		return true, fields
	case "judgement":
		return len(t.Judgement.Scorers) > 0, fields
	case "reference_material":
		return t.ReferenceMaterial != nil, fields
	case "meta":
		return t.Meta.Author != "" && t.Meta.CreatedAt != "", fields
	}
	return false, fields
}

func sortedGroups(m map[string]Requirement) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ValidateForm returns the contract rows for a form, for display in the API
// and in the task editor.
func ValidateForm(form TaskForm) map[string]Requirement {
	out := map[string]Requirement{}
	for k, v := range formContract[form] {
		out[k] = v
	}
	return out
}
