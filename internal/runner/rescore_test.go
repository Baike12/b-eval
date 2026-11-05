package runner

import (
	"testing"

	"github.com/Baike12/b-eval/internal/domain"
	"github.com/Baike12/b-eval/internal/scoring"
)

func rejudgeTaskSet() domain.TaskSet {
	return domain.TaskSet{
		TaskSetID: "demo",
		Version:   "v1",
		Benchmark: "b",
		Tasks: []domain.Task{{
			TaskID:         "T-1",
			Form:           domain.FormStaticQA,
			DisplayName:    "Return window",
			Benchmark:      "b",
			TaskSetVersion: "v1",
			Capability: domain.Capability{
				Level1:     []string{"C4"},
				Difficulty: "L1",
				Domains:    []string{"D7"},
			},
			Description: domain.Description{Instruction: "Answer."},
			Budget:      domain.Budget{Steps: 1},
			Judgement: domain.Judgement{
				Combination:     "and",
				Primary:         0,
				ReferenceAnswer: "30 days",
				Scorers: []domain.ScorerSpec{{
					Type:   "contains",
					Params: map[string]any{"needles": []any{"30 days"}},
					Weight: 1, Primary: true,
				}},
			},
		}},
	}
}

func TestRejudgeScoresPersistedTrajectories(t *testing.T) {
	trajectory := domain.Trajectory{
		TrialID:     "T-1#0#7",
		TaskID:      "T-1",
		TrialIndex:  0,
		Seed:        7,
		Answer:      "The window is 30 days.",
		Termination: domain.TermCompleted,
	}

	outcomes, verdicts, err := Rejudge(rejudgeTaskSet(), "agent", []domain.Trajectory{trajectory}, scoring.NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if len(outcomes) != 1 || len(verdicts) != 1 {
		t.Fatalf("outcomes = %d, verdicts = %d, want 1 and 1", len(outcomes), len(verdicts))
	}
	if outcomes[0].Verdict.Combined != domain.VerdictPass {
		t.Errorf("verdict = %q, want PASS", outcomes[0].Verdict.Combined)
	}
	if outcomes[0].TrialIndex != 0 || outcomes[0].Seed != 7 {
		t.Errorf("trial identity = (%d, %d), want (0, 7)", outcomes[0].TrialIndex, outcomes[0].Seed)
	}
}

// usageScorer reports judge consumption so the test can prove it reaches the
// trial row.
type usageScorer struct{}

func (usageScorer) Type() string { return "usage_stub" }

func (usageScorer) Judge(task domain.Task, spec domain.ScorerSpec, traj domain.Trajectory) (domain.ScorerResult, error) {
	return domain.ScorerResult{
		Value: domain.VerdictPass,
		Score: 1,
		Usage: &domain.Usage{InputTokens: 40, OutputTokens: 10, ModelCalls: 1},
	}, nil
}

func TestRejudgeFoldsTheJudgeUsageIntoTheTrial(t *testing.T) {
	reg := scoring.NewRegistry()
	reg.Register(usageScorer{})
	ts := rejudgeTaskSet()
	ts.Tasks[0].Judgement.Scorers[0].Type = "usage_stub"

	trajectory := domain.Trajectory{
		TrialID:     "T-1#0#1",
		TaskID:      "T-1",
		Termination: domain.TermCompleted,
		Answer:      "x",
	}
	outcomes, _, err := Rejudge(ts, "agent", []domain.Trajectory{trajectory}, reg)
	if err != nil {
		t.Fatal(err)
	}
	if outcomes[0].Usage.JudgeInputTokens != 40 || outcomes[0].Usage.JudgeOutputTokens != 10 {
		t.Errorf("judge usage = %+v, want 40 input and 10 output tokens", outcomes[0].Usage)
	}
	if outcomes[0].Usage.TotalTokens() != 50 {
		t.Errorf("total tokens = %d, want 50 (the judge is part of the cost)", outcomes[0].Usage.TotalTokens())
	}
}

func TestRejudgeRefusesATrajectoryWhoseTaskIsGone(t *testing.T) {
	trajectory := domain.Trajectory{TrialID: "T-9#0#1", TaskID: "T-9", Termination: domain.TermCompleted}
	if _, _, err := Rejudge(rejudgeTaskSet(), "agent", []domain.Trajectory{trajectory}, scoring.NewRegistry()); err == nil {
		t.Fatal("expected a refusal when the task set no longer holds the task")
	}
}

func TestRejudgeKeepsTheFailureOfAnUnfinishedTrial(t *testing.T) {
	trajectory := domain.Trajectory{
		TrialID:     "T-1#0#1",
		TaskID:      "T-1",
		Termination: domain.TermStepLimit,
		Answer:      "30 days",
	}
	outcomes, _, err := Rejudge(rejudgeTaskSet(), "agent", []domain.Trajectory{trajectory}, scoring.NewRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if outcomes[0].Failure != domain.ClassBudgetExhausted {
		t.Errorf("failure class = %q, want budget_exhausted", outcomes[0].Failure)
	}
	if outcomes[0].Verdict.Combined != domain.VerdictFail || outcomes[0].Verdict.Score != 0 {
		t.Errorf("verdict = %q score %v, want FAIL 0", outcomes[0].Verdict.Combined, outcomes[0].Verdict.Score)
	}
}
