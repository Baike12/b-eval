package runner

import (
	"context"
	"testing"

	"github.com/Baike12/b-eval/internal/adapter"
	"github.com/Baike12/b-eval/internal/domain"
	"github.com/Baike12/b-eval/internal/metrics"
	"github.com/Baike12/b-eval/internal/scoring"
)

func staticTask(id, reference, needle string) domain.Task {
	multi := false
	return domain.Task{
		TaskID:         id,
		Form:           domain.FormStaticQA,
		DisplayName:    id,
		Benchmark:      "b",
		TaskSetVersion: "v1",
		Capability:     domain.Capability{Level1: []string{"C4"}, Difficulty: "L1", Domains: []string{"D7"}},
		Description:    domain.Description{Instruction: "Answer."},
		Budget:         domain.Budget{Steps: 1, Tokens: 1000},
		InteractionContract: &domain.InteractionContract{
			TerminationConditions: []string{"answer"},
			MultiTurn:             &multi,
		},
		Judgement: domain.Judgement{
			Combination:     "and",
			Primary:         0,
			ReferenceAnswer: reference,
			Scorers: []domain.ScorerSpec{{
				Type:    "contains",
				Params:  map[string]any{"needles": []any{needle}},
				Weight:  1,
				Primary: true,
			}},
		},
		Meta: domain.TaskMeta{Author: "test", CreatedAt: "2026-10-22T00:00:00Z"},
	}
}

func newRunner() *Runner {
	metricRegistry := metrics.NewRegistry()
	return &Runner{Scorers: scoring.NewRegistry(), Metrics: metricRegistry}
}

func TestRunScoresTheTrials(t *testing.T) {
	mock := adapter.NewMockAdapter("the window is 30 days")
	task := staticTask("T-1", "reference", "30 days")

	run, err := newRunner().Run(context.Background(), Input{
		Plan:    domain.Plan{PlanID: "p", AgentID: "a", Repetitions: 2, Aggregation: "macro"},
		TaskSet: domain.TaskSet{TaskSetID: "ts", Version: "v1", Tasks: []domain.Task{task}},
		Adapter: mock,
		AgentID: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(run.Trials) != 2 {
		t.Fatalf("trials = %d, want 2", len(run.Trials))
	}
	for _, trial := range run.Trials {
		if trial.Failure != domain.ClassCompleted {
			t.Errorf("failure class = %q, want completed", trial.Failure)
		}
		if trial.Verdict.Combined != domain.VerdictPass {
			t.Errorf("verdict = %q, want PASS", trial.Verdict.Combined)
		}
	}
}

func TestRunRecordsTheTerminationReason(t *testing.T) {
	mock := adapter.NewMockAdapter("the window is 30 days")
	task := staticTask("T-1", "reference", "30 days")

	run, err := newRunner().Run(context.Background(), Input{
		Plan:    domain.Plan{PlanID: "p", AgentID: "a", Repetitions: 1},
		TaskSet: domain.TaskSet{TaskSetID: "ts", Version: "v1", Tasks: []domain.Task{task}},
		Adapter: mock,
		AgentID: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(run.Trajectories) != 1 {
		t.Fatalf("trajectories = %d, want 1", len(run.Trajectories))
	}
	if run.Trajectories[0].Termination != domain.TermCompleted {
		t.Errorf("termination = %q, want completed", run.Trajectories[0].Termination)
	}
}

func TestAFailingAgentIsJudgedAsAFailure(t *testing.T) {
	mock := adapter.NewMockAdapter("I do not know")
	task := staticTask("T-1", "reference", "30 days")

	run, err := newRunner().Run(context.Background(), Input{
		Plan:    domain.Plan{PlanID: "p", AgentID: "a", Repetitions: 1},
		TaskSet: domain.TaskSet{TaskSetID: "ts", Version: "v1", Tasks: []domain.Task{task}},
		Adapter: mock,
		AgentID: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.Trials[0].Verdict.Combined != domain.VerdictFail {
		t.Errorf("verdict = %q, want FAIL", run.Trials[0].Verdict.Combined)
	}
	if run.Trials[0].Failure != domain.ClassCompleted {
		t.Errorf("failure class = %q, want completed", run.Trials[0].Failure)
	}
}

func TestAnAgentCrashIsAModelFailure(t *testing.T) {
	mock := adapter.NewMockAdapter("unused")
	mock.FailWith = errCrash{}
	task := staticTask("T-1", "reference", "30 days")

	run, err := newRunner().Run(context.Background(), Input{
		Plan:    domain.Plan{PlanID: "p", AgentID: "a", Repetitions: 1},
		TaskSet: domain.TaskSet{TaskSetID: "ts", Version: "v1", Tasks: []domain.Task{task}},
		Adapter: mock,
		AgentID: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.Trials[0].Failure != domain.ClassModelFailure {
		t.Errorf("failure class = %q, want model_failure", run.Trials[0].Failure)
	}
	if run.Trials[0].Verdict.Combined != domain.VerdictFail {
		t.Errorf("verdict = %q, want FAIL", run.Trials[0].Verdict.Combined)
	}
}

func TestTheTokenBudgetTerminatesTheTrial(t *testing.T) {
	mock := adapter.NewMockAdapter("the window is 30 days")
	mock.Usage = domain.Usage{InputTokens: 5000}
	task := staticTask("T-1", "reference", "30 days")
	task.Budget.Tokens = 1000

	run, err := newRunner().Run(context.Background(), Input{
		Plan:    domain.Plan{PlanID: "p", AgentID: "a", Repetitions: 1},
		TaskSet: domain.TaskSet{TaskSetID: "ts", Version: "v1", Tasks: []domain.Task{task}},
		Adapter: mock,
		AgentID: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.Trials[0].Failure != domain.ClassBudgetExhausted {
		t.Errorf("failure class = %q, want budget_exhausted", run.Trials[0].Failure)
	}
}

func TestAnActionOutsideTheDeclaredSpaceIsAViolation(t *testing.T) {
	task := staticTask("T-1", "reference", "30 days")
	task.InteractionContract.ActionSpace = []string{"answer"}
	// The scripted adapter emits an action the task never declared.
	scripted := &adapter.ScriptedAdapter{
		Actions:   []adapter.Action{{Name: "click", Content: "the window is 30 days"}},
		StopAfter: 1,
	}

	run, err := newRunner().Run(context.Background(), Input{
		Plan:    domain.Plan{PlanID: "p", AgentID: "a", Repetitions: 1},
		TaskSet: domain.TaskSet{TaskSetID: "ts", Version: "v1", Tasks: []domain.Task{task}},
		Adapter: scripted,
		AgentID: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(run.Trajectories[0].Violations) == 0 {
		t.Fatal("expected a protocol violation for the action outside the declared space")
	}
	if run.Trajectories[0].Violations[0].Type != string(adapter.ViolationActionOutOfSpace) {
		t.Errorf("violation type = %q, want %q", run.Trajectories[0].Violations[0].Type, adapter.ViolationActionOutOfSpace)
	}
}

func TestTheResumeUnitIsStableAcrossRuns(t *testing.T) {
	task := staticTask("T-1", "reference", "30 days")
	plan := domain.Plan{PlanID: "p", AgentID: "a", Repetitions: 3}

	first, err := newRunner().Run(context.Background(), Input{
		Plan:    plan,
		TaskSet: domain.TaskSet{TaskSetID: "ts", Version: "v1", Tasks: []domain.Task{task}},
		Adapter: adapter.NewMockAdapter("the window is 30 days"),
		AgentID: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := newRunner().Run(context.Background(), Input{
		Plan:    plan,
		TaskSet: domain.TaskSet{TaskSetID: "ts", Version: "v1", Tasks: []domain.Task{task}},
		Adapter: adapter.NewMockAdapter("the window is 30 days"),
		AgentID: "a",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Trials) != len(second.Trials) {
		t.Fatalf("trial counts differ: %d vs %d", len(first.Trials), len(second.Trials))
	}
	for i := range first.Trials {
		if first.Trials[i].TrialID != second.Trials[i].TrialID {
			t.Errorf("trial %d: id %q vs %q", i, first.Trials[i].TrialID, second.Trials[i].TrialID)
		}
	}
}

type errCrash struct{}

func (errCrash) Error() string { return "the agent crashed" }
