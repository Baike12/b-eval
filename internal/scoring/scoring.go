// Package scoring turns a trajectory into a verdict.
//
// A scorer returns exactly one of PASS, FAIL or UNSCORED. UNSCORED is a
// first-class state: a scorer that could not run is never reported as a
// failure, because a judging failure must not become a capability deduction
// for the agent.
package scoring

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Baike12/b-eval/internal/domain"
)

// Scorer judges one trial against one task.
type Scorer interface {
	// Type is the scorer's registered type name, matching the task's
	// scorer reference.
	Type() string
	// Judge evaluates the trial and returns the outcome with its evidence.
	Judge(task domain.Task, spec domain.ScorerSpec, traj domain.Trajectory) (domain.ScorerResult, error)
}

// Registry resolves a scorer reference to an implementation.
type Registry struct {
	byType map[string]Scorer
}

// NewRegistry returns a registry holding the scorers the platform ships.
func NewRegistry() *Registry {
	r := &Registry{byType: map[string]Scorer{}}
	for _, s := range builtinScorers() {
		r.byType[s.Type()] = s
	}
	return r
}

// Register adds a scorer implementation.
func (r *Registry) Register(s Scorer) { r.byType[s.Type()] = s }

// Get resolves a scorer reference.
func (r *Registry) Get(t string) (Scorer, bool) {
	s, ok := r.byType[t]
	return s, ok
}

// Types lists the registered scorer type names.
func (r *Registry) Types() []string {
	out := make([]string, 0, len(r.byType))
	for t := range r.byType {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// Judge evaluates a trial against the task's declared scorers and composes
// the results into one verdict.
func Judge(task domain.Task, traj domain.Trajectory, reg *Registry) (domain.Verdict, error) {
	results := make([]domain.ScorerResult, 0, len(task.Judgement.Scorers))
	for i, spec := range task.Judgement.Scorers {
		s, ok := reg.Get(spec.Type)
		if !ok {
			// An unknown scorer reference is an unscored condition, not a
			// failure: the platform cannot judge the criterion, so it must
			// not charge the agent for it.
			results = append(results, domain.ScorerResult{
				Index:  i,
				Type:   spec.Type,
				Weight: spec.Weight,
				Value:  domain.VerdictUnscored,
				Detail: fmt.Sprintf("scorer type %q is not registered", spec.Type),
			})
			continue
		}
		res, err := s.Judge(task, spec, traj)
		if err != nil {
			res = domain.ScorerResult{
				Index:  i,
				Type:   spec.Type,
				Weight: spec.Weight,
				Value:  domain.VerdictUnscored,
				Detail: err.Error(),
			}
		}
		res.Index = i
		res.Type = spec.Type
		res.Weight = spec.Weight
		results = append(results, res)
	}
	verdict := Compose(task, results)
	// The verdict identifies its trial. Without it the caller cannot tell a
	// computed verdict from an absent one, and an absent verdict reads as a
	// failure rather than as a missing measurement.
	verdict.TrialID = traj.TrialID
	verdict.TaskID = task.TaskID
	// The verdict carries the answer and the reference so that a report, an
	// export and a custom metric can all inspect what was judged without
	// re-reading the trajectory.
	verdict.Evidence = map[string]any{
		"answer":    traj.Answer,
		"reference": task.Judgement.ReferenceAnswer,
	}
	return verdict, nil
}

// Compose combines scorer results into a verdict.
//
// The primary criterion decides the binary outcome. An unscored primary
// criterion makes the whole trial unscored, because a task cannot be judged
// on an auxiliary criterion alone.
func Compose(task domain.Task, results []domain.ScorerResult) domain.Verdict {
	v := domain.Verdict{
		TrialID:  "",
		TaskID:   task.TaskID,
		Scorers:  results,
		Combined: domain.VerdictFail,
		JudgedAt: domain.Now(),
	}
	if len(results) == 0 {
		v.Combined = domain.VerdictUnscored
		return v
	}

	primaryIdx := task.Judgement.Primary
	if primaryIdx < 0 || primaryIdx >= len(results) {
		v.Combined = domain.VerdictUnscored
		v.Evidence = map[string]any{"reason": "primary criterion index is out of range"}
		return v
	}
	primary := results[primaryIdx]
	v.PrimaryPassed = primary.Value == domain.VerdictPass
	switch primary.Value {
	case domain.VerdictUnscored:
		v.Combined = domain.VerdictUnscored
	case domain.VerdictPass:
		v.Combined = domain.VerdictPass
	default:
		v.Combined = domain.VerdictFail
	}

	// The continuous score is a weighted average over the scorers that
	// actually produced a value. Unscored scorers are excluded from both
	// the numerator and the denominator, so a judge that could not run
	// neither inflates nor deflates the score.
	var weighted, totalWeight float64
	for _, r := range results {
		if r.Value == domain.VerdictUnscored {
			continue
		}
		w := r.Weight
		if w <= 0 {
			w = 1
		}
		weighted += w * r.Score
		totalWeight += w
	}
	if totalWeight > 0 {
		v.Score = weighted / totalWeight
	}
	return v
}

// builtinScorers returns the scorers the platform ships in M0.
func builtinScorers() []Scorer {
	return []Scorer{
		&ExactMatchScorer{},
		&ContainsScorer{},
		&RegexScorer{},
		&KeywordSetScorer{},
		&RubricScorer{},
		&ActionSequenceScorer{},
	}
}

// ExactMatchScorer passes when the answer equals the reference answer after
// trimming surrounding whitespace.
type ExactMatchScorer struct{}

// Type identifies the scorer.
func (s *ExactMatchScorer) Type() string { return "exact" }

// Judge evaluates the answer against the reference.
func (s *ExactMatchScorer) Judge(task domain.Task, spec domain.ScorerSpec, traj domain.Trajectory) (domain.ScorerResult, error) {
	got := strings.TrimSpace(traj.Answer)
	want := strings.TrimSpace(task.Judgement.ReferenceAnswer)
	if want == "" {
		return domain.ScorerResult{Value: domain.VerdictUnscored, Detail: "task declares no reference answer"}, nil
	}
	pass := got == want
	return domain.ScorerResult{
		Value:    verdictOf(pass),
		Score:    scoreOf(pass),
		Detail:   fmt.Sprintf("answer %q vs reference %q", got, want),
		Evidence: map[string]any{"answer": got, "reference": want},
	}, nil
}

// ContainsScorer passes when the answer contains every configured needle.
type ContainsScorer struct{}

// Type identifies the scorer.
func (s *ContainsScorer) Type() string { return "contains" }

// Judge evaluates the answer against the configured needles.
func (s *ContainsScorer) Judge(task domain.Task, spec domain.ScorerSpec, traj domain.Trajectory) (domain.ScorerResult, error) {
	needles := stringList(spec.Params, "needles")
	if len(needles) == 0 {
		return domain.ScorerResult{Value: domain.VerdictUnscored, Detail: "contains scorer has no needles"}, nil
	}
	answer := strings.ToLower(traj.Answer)
	var missing []string
	for _, n := range needles {
		if !strings.Contains(answer, strings.ToLower(n)) {
			missing = append(missing, n)
		}
	}
	pass := len(missing) == 0
	return domain.ScorerResult{
		Value:    verdictOf(pass),
		Score:    scoreOf(pass),
		Detail:   fmt.Sprintf("missing needles: %v", missing),
		Evidence: map[string]any{"answer": traj.Answer, "missing": missing, "needles": needles},
	}, nil
}

// RegexScorer passes when the answer matches the configured pattern.
type RegexScorer struct{}

// Type identifies the scorer.
func (s *RegexScorer) Type() string { return "regex" }

// Judge evaluates the answer against the configured pattern.
func (s *RegexScorer) Judge(task domain.Task, spec domain.ScorerSpec, traj domain.Trajectory) (domain.ScorerResult, error) {
	pattern, _ := spec.Params["pattern"].(string)
	if pattern == "" {
		return domain.ScorerResult{Value: domain.VerdictUnscored, Detail: "regex scorer has no pattern"}, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return domain.ScorerResult{Value: domain.VerdictUnscored}, fmt.Errorf("invalid regex %q: %w", pattern, err)
	}
	pass := re.MatchString(traj.Answer)
	return domain.ScorerResult{
		Value:    verdictOf(pass),
		Score:    scoreOf(pass),
		Detail:   fmt.Sprintf("pattern %q matched=%v", pattern, pass),
		Evidence: map[string]any{"pattern": pattern, "answer": traj.Answer},
	}, nil
}

// KeywordSetScorer scores the share of required keywords the answer
// contains, rather than a binary outcome.
type KeywordSetScorer struct{}

// Type identifies the scorer.
func (s *KeywordSetScorer) Type() string { return "keyword_set" }

// Judge scores the keyword coverage of the answer.
func (s *KeywordSetScorer) Judge(task domain.Task, spec domain.ScorerSpec, traj domain.Trajectory) (domain.ScorerResult, error) {
	keywords := stringList(spec.Params, "keywords")
	if len(keywords) == 0 {
		return domain.ScorerResult{Value: domain.VerdictUnscored, Detail: "keyword_set scorer has no keywords"}, nil
	}
	answer := strings.ToLower(traj.Answer)
	found := 0
	var missing []string
	for _, k := range keywords {
		if strings.Contains(answer, strings.ToLower(k)) {
			found++
		} else {
			missing = append(missing, k)
		}
	}
	ratio := float64(found) / float64(len(keywords))
	return domain.ScorerResult{
		Value:    verdictOf(ratio >= 1),
		Score:    ratio,
		Detail:   fmt.Sprintf("covered %d/%d keywords", found, len(keywords)),
		Evidence: map[string]any{"covered": found, "total": len(keywords), "missing": missing},
	}, nil
}

// RubricScorer judges a rubric item declared on the task. In M0 the rubric
// item is evaluated by a deterministic rule; a model judge plugs into the
// same interface later.
type RubricScorer struct{}

// Type identifies the scorer.
func (s *RubricScorer) Type() string { return "rubric" }

// Judge evaluates the rubric items of the task.
//
// A rubric item carries its own expectation in the params. Items are judged
// independently and the score is the share of satisfied items, so a partially
// correct answer earns partial credit.
func (s *RubricScorer) Judge(task domain.Task, spec domain.ScorerSpec, traj domain.Trajectory) (domain.ScorerResult, error) {
	items := rubricItems(spec)
	if len(items) == 0 {
		return domain.ScorerResult{Value: domain.VerdictUnscored, Detail: "rubric scorer has no items"}, nil
	}
	answer := strings.ToLower(traj.Answer)
	satisfied := 0
	var unmet []string
	for _, item := range items {
		needle := strings.ToLower(item.Needle)
		if needle != "" && strings.Contains(answer, needle) {
			satisfied++
			continue
		}
		unmet = append(unmet, item.ID)
	}
	ratio := float64(satisfied) / float64(len(items))
	return domain.ScorerResult{
		Value:    verdictOf(ratio >= 1),
		Score:    ratio,
		Detail:   fmt.Sprintf("satisfied %d/%d rubric items", satisfied, len(items)),
		Evidence: map[string]any{"satisfied": satisfied, "total": len(items), "unmet": unmet},
	}, nil
}

// ActionSequenceScorer checks the recorded action sequence against the
// declared one. Action judgement is opt-in: a task that does not declare an
// expected sequence is judged on outcome only.
type ActionSequenceScorer struct{}

// Type identifies the scorer.
func (s *ActionSequenceScorer) Type() string { return "action_sequence" }

// Judge compares the recorded action names with the declared sequence.
func (s *ActionSequenceScorer) Judge(task domain.Task, spec domain.ScorerSpec, traj domain.Trajectory) (domain.ScorerResult, error) {
	expected := stringList(spec.Params, "actions")
	if len(expected) == 0 {
		return domain.ScorerResult{Value: domain.VerdictUnscored, Detail: "no action sequence declared"}, nil
	}
	var actual []string
	for _, st := range traj.Steps {
		actual = append(actual, st.Action.Name)
	}
	pass := strings.Join(expected, ">") == strings.Join(actual, ">")
	return domain.ScorerResult{
		Value:    verdictOf(pass),
		Score:    scoreOf(pass),
		Detail:   fmt.Sprintf("expected %v, actual %v", expected, actual),
		Evidence: map[string]any{"expected": expected, "actual": actual},
	}, nil
}

// RubricItem is one atomic requirement of a rubric.
type RubricItem struct {
	ID     string  `json:"id"`
	Needle string  `json:"needle,omitempty"`
	Weight float64 `json:"weight,omitempty"`
}

// rubricItems reads the rubric items declared on one scorer spec.
func rubricItems(spec domain.ScorerSpec) []RubricItem {
	raw, ok := spec.Params["items"]
	if !ok {
		return nil
	}
	return parseRubricItems(raw)
}

// parseRubricItems accepts the two shapes a task file may use: an array of
// objects, or an array of plain strings.
func parseRubricItems(raw any) []RubricItem {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	items := make([]RubricItem, 0, len(list))
	for i, v := range list {
		switch item := v.(type) {
		case map[string]any:
			id, _ := item["id"].(string)
			needle, _ := item["needle"].(string)
			weight := 1.0
			if w, ok := item["weight"].(float64); ok {
				weight = w
			}
			items = append(items, RubricItem{ID: id, Needle: needle, Weight: weight})
		case string:
			items = append(items, RubricItem{ID: fmt.Sprintf("item-%d", i+1), Needle: item, Weight: 1})
		}
	}
	return items
}

// stringList reads a string list parameter out of a scorer's params.
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

func verdictOf(pass bool) domain.VerdictValue {
	if pass {
		return domain.VerdictPass
	}
	return domain.VerdictFail
}

func scoreOf(pass bool) float64 {
	if pass {
		return 1
	}
	return 0
}
