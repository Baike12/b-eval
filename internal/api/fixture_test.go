package api

import "github.com/Baike12/b-eval/internal/domain"

// taskSetFixture builds a minimal stored task set for the store tests.
func taskSetFixture() domain.TaskSet {
	multi := false
	return domain.TaskSet{
		TaskSetID: "fixture",
		Version:   "v1",
		Benchmark: "b",
		CreatedAt: "2026-10-22T00:00:00Z",
		Tasks: []domain.Task{{
			TaskID:         "T-1",
			Form:           domain.FormStaticQA,
			DisplayName:    "A question",
			Benchmark:      "b",
			TaskSetVersion: "v1",
			Capability: domain.Capability{
				Level1:     []string{"C4"},
				Difficulty: "L1",
				Domains:    []string{"D7"},
			},
			Description:         domain.Description{Instruction: "Answer."},
			Budget:              domain.Budget{Steps: 1},
			InteractionContract: &domain.InteractionContract{TerminationConditions: []string{"answer"}, MultiTurn: &multi},
			Judgement: domain.Judgement{
				Combination:     "and",
				Primary:         0,
				ReferenceAnswer: "ref",
				Scorers:         []domain.ScorerSpec{{Type: "contains", Weight: 1, Primary: true}},
			},
			Meta: domain.TaskMeta{Author: "test", CreatedAt: "2026-10-22T00:00:00Z"},
		}},
	}
}
