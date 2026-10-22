package domain

import "testing"

func staticQATask() Task {
	multi := false
	return Task{
		TaskID:         "T-1",
		Form:           FormStaticQA,
		DisplayName:    "A question",
		Benchmark:      "b",
		TaskSetVersion: "v1",
		Capability: Capability{
			Level1:     []string{"C4"},
			Difficulty: "L1",
			Domains:    []string{"D7"},
		},
		Description: Description{Instruction: "Answer the question."},
		Budget:      Budget{Steps: 1, Tokens: 1000},
		InteractionContract: &InteractionContract{
			TerminationConditions: []string{"answer"},
			MultiTurn:             &multi,
		},
		Judgement: Judgement{
			Combination:     "and",
			Primary:         0,
			ReferenceAnswer: "the reference answer",
			Scorers:         []ScorerSpec{{Type: "contains", Weight: 1, Primary: true}},
		},
		Meta: TaskMeta{Author: "test", CreatedAt: "2026-10-22T00:00:00Z"},
	}
}

func envAgenticTask() Task {
	multi := false
	return Task{
		TaskID:         "T-2",
		Form:           FormEnvAgentic,
		DisplayName:    "An environment task",
		Benchmark:      "b",
		TaskSetVersion: "v1",
		Capability: Capability{
			Level1:     []string{"C2"},
			Difficulty: "L2",
			Domains:    []string{"D1"},
		},
		Description:  Description{Instruction: "Do the thing."},
		Budget:       Budget{Steps: 5},
		InitialState: &InitialState{SnapshotID: "snap-1", Seed: 7},
		InteractionContract: &InteractionContract{
			ActionSpace:           []string{"click"},
			ObservationShape:      "text",
			TerminationConditions: []string{"done"},
			MultiTurn:             &multi,
		},
		Judgement: Judgement{
			Combination: "and",
			Primary:     0,
			Scorers:     []ScorerSpec{{Type: "state", Weight: 1, Primary: true}},
		},
		Meta: TaskMeta{Author: "test", CreatedAt: "2026-10-22T00:00:00Z"},
	}
}

func TestStaticQATaskSatisfiesItsContract(t *testing.T) {
	if errs := ValidateTask(staticQATask()); len(errs) != 0 {
		t.Fatalf("a valid static-QA task was refused: %v", errs)
	}
}

func TestEnvAgenticTaskSatisfiesItsContract(t *testing.T) {
	if errs := ValidateTask(envAgenticTask()); len(errs) != 0 {
		t.Fatalf("a valid environment-agentic task was refused: %v", errs)
	}
}

func TestStaticQATaskWithAnEnvironmentFieldIsRefused(t *testing.T) {
	task := staticQATask()
	task.InitialState = &InitialState{SnapshotID: "snap-1"}
	errs := ValidateTask(task)
	if len(errs) == 0 {
		t.Fatal("expected the environment field to be refused")
	}
	found := false
	for _, e := range errs {
		if e.Field == "initial_state" {
			found = true
		}
	}
	if !found {
		t.Errorf("the refusal does not name the offending group: %v", errs)
	}
}

func TestStaticQATaskWithAnActionSpaceIsRefused(t *testing.T) {
	task := staticQATask()
	task.InteractionContract.ActionSpace = []string{"click"}
	errs := ValidateTask(task)
	if len(errs) == 0 {
		t.Fatal("expected the action space to be refused")
	}
	found := false
	for _, e := range errs {
		if e.Field == "interaction_contract.action_space" {
			found = true
		}
	}
	if !found {
		t.Errorf("the refusal does not name the offending field: %v", errs)
	}
}

func TestStaticQATaskWithoutAMultiTurnDeclarationIsRefused(t *testing.T) {
	task := staticQATask()
	task.InteractionContract.MultiTurn = nil
	errs := ValidateTask(task)
	if len(errs) == 0 {
		t.Fatal("expected the missing multi_turn declaration to be refused")
	}
}

func TestEnvAgenticTaskWithoutAnInitialStateIsRefused(t *testing.T) {
	task := envAgenticTask()
	task.InitialState = nil
	errs := ValidateTask(task)
	if len(errs) == 0 {
		t.Fatal("expected the missing initial state to be refused")
	}
}

func TestStaticQATaskWithoutAReferenceAnswerIsRefused(t *testing.T) {
	task := staticQATask()
	task.Judgement.ReferenceAnswer = ""
	errs := ValidateTask(task)
	if len(errs) == 0 {
		t.Fatal("expected the missing reference answer to be refused")
	}
}

func TestTaskWithoutAFormIsRefused(t *testing.T) {
	task := staticQATask()
	task.Form = ""
	errs := ValidateTask(task)
	if len(errs) == 0 {
		t.Fatal("expected the missing form to be refused")
	}
}

func TestFailureClassDenominatorRules(t *testing.T) {
	excluded := []FailureClass{ClassPlatformFailure, ClassCancelled}
	for _, c := range excluded {
		if c.CountsInDenominator() {
			t.Errorf("%q must not count in the denominator", c)
		}
	}
	included := []FailureClass{ClassCompleted, ClassModelFailure, ClassBudgetExhausted, ClassProtocolViolation}
	for _, c := range included {
		if !c.CountsInDenominator() {
			t.Errorf("%q must count in the denominator", c)
		}
	}
}

func TestPlanFingerprintChangesWithTheScoringConfiguration(t *testing.T) {
	base := Plan{PlanID: "p", TaskSetID: "ts", TaskSetVersion: "v1", AgentID: "a", Repetitions: 2, Aggregation: "macro", CIMethod: "wilson"}
	other := base
	other.Aggregation = "micro"
	if base.Fingerprint() == other.Fingerprint() {
		t.Error("the fingerprint must change when the aggregation changes")
	}
	same := base
	if base.Fingerprint() != same.Fingerprint() {
		t.Error("the fingerprint must be stable for an unchanged plan")
	}
}
