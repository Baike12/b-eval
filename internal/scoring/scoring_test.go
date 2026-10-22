package scoring

import (
	"testing"

	"github.com/Baike12/b-eval/internal/domain"
)

func taskWithScorers(scorers []domain.ScorerSpec, reference string) domain.Task {
	return domain.Task{
		TaskID: "T-1",
		Form:   domain.FormStaticQA,
		Judgement: domain.Judgement{
			Combination:     "and",
			Primary:         0,
			ReferenceAnswer: reference,
			Scorers:         scorers,
		},
	}
}

func containsScorer(needles []string) domain.ScorerSpec {
	params := map[string]any{}
	list := make([]any, 0, len(needles))
	for _, n := range needles {
		list = append(list, n)
	}
	params["needles"] = list
	return domain.ScorerSpec{Type: "contains", Params: params, Weight: 1, Primary: true}
}

func TestContainsScorerPassesWhenEveryNeedleIsPresent(t *testing.T) {
	reg := NewRegistry()
	task := taskWithScorers([]domain.ScorerSpec{containsScorer([]string{"30 days", "receipt"})}, "ref")
	traj := domain.Trajectory{Answer: "The window is 30 days, keep the receipt."}

	verdict, err := Judge(task, traj, reg)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Combined != domain.VerdictPass {
		t.Errorf("combined verdict = %q, want PASS", verdict.Combined)
	}
	if verdict.Score != 1 {
		t.Errorf("score = %v, want 1", verdict.Score)
	}
}

func TestContainsScorerFailsWhenANeedleIsMissing(t *testing.T) {
	reg := NewRegistry()
	task := taskWithScorers([]domain.ScorerSpec{containsScorer([]string{"30 days", "receipt"})}, "ref")
	traj := domain.Trajectory{Answer: "The window is 30 days."}

	verdict, err := Judge(task, traj, reg)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Combined != domain.VerdictFail {
		t.Errorf("combined verdict = %q, want FAIL", verdict.Combined)
	}
}

func TestAnUnknownScorerIsUnscoredRatherThanFailed(t *testing.T) {
	reg := NewRegistry()
	task := taskWithScorers([]domain.ScorerSpec{{Type: "not_registered", Weight: 1, Primary: true}}, "ref")

	verdict, err := Judge(task, domain.Trajectory{Answer: "anything"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Combined != domain.VerdictUnscored {
		t.Errorf("combined verdict = %q, want UNSCORED", verdict.Combined)
	}
}

func TestAnUnscoredPrimaryCriterionMakesTheWholeTrialUnscored(t *testing.T) {
	reg := NewRegistry()
	// The primary criterion cannot be evaluated; the auxiliary one can. The
	// trial must be unscored, because a task cannot be judged on an
	// auxiliary criterion alone.
	task := taskWithScorers([]domain.ScorerSpec{
		{Type: "not_registered", Weight: 1, Primary: true},
		containsScorer([]string{"present"}),
	}, "ref")
	task.Judgement.Primary = 0

	verdict, err := Judge(task, domain.Trajectory{Answer: "present"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Combined != domain.VerdictUnscored {
		t.Errorf("combined verdict = %q, want UNSCORED", verdict.Combined)
	}
}

func TestTheContinuousScoreExcludesUnscoredScorers(t *testing.T) {
	reg := NewRegistry()
	task := taskWithScorers([]domain.ScorerSpec{
		containsScorer([]string{"present"}),
		{Type: "not_registered", Weight: 1},
	}, "ref")

	verdict, err := Judge(task, domain.Trajectory{Answer: "present"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Score != 1 {
		t.Errorf("score = %v, want 1 (the unscored scorer must not drag the score down)", verdict.Score)
	}
}

func TestKeywordSetScorerGivesPartialCredit(t *testing.T) {
	reg := NewRegistry()
	params := map[string]any{"keywords": []any{"alpha", "beta", "gamma", "delta"}}
	task := taskWithScorers([]domain.ScorerSpec{{Type: "keyword_set", Params: params, Weight: 1, Primary: true}}, "ref")

	verdict, err := Judge(task, domain.Trajectory{Answer: "alpha and beta"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Score != 0.5 {
		t.Errorf("score = %v, want 0.5", verdict.Score)
	}
	if verdict.Combined != domain.VerdictFail {
		t.Errorf("combined verdict = %q, want FAIL (the criterion is not fully satisfied)", verdict.Combined)
	}
}

func TestTheVerdictCarriesTheAnswerAndTheReference(t *testing.T) {
	reg := NewRegistry()
	task := taskWithScorers([]domain.ScorerSpec{containsScorer([]string{"x"})}, "the reference")
	verdict, err := Judge(task, domain.Trajectory{TrialID: "T-1#0#1", Answer: "x"}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if verdict.TrialID != "T-1#0#1" {
		t.Errorf("trial id = %q, want T-1#0#1", verdict.TrialID)
	}
	if verdict.Evidence["answer"] != "x" {
		t.Errorf("the verdict does not carry the answer: %v", verdict.Evidence)
	}
	if verdict.Evidence["reference"] != "the reference" {
		t.Errorf("the verdict does not carry the reference: %v", verdict.Evidence)
	}
}
