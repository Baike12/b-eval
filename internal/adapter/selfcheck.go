package adapter

import (
	"context"
	"fmt"

	"github.com/Baike12/b-eval/internal/domain"
)

// SelfCheck runs a registered adapter against a small task set and reports
// whether the integration is sound.
//
// The readiness gate refuses evaluation for an adapter that has not passed,
// so that an integration bug is never recorded as weak agent capability.
type SelfCheck struct {
	// MaxSteps bounds the check so a broken adapter cannot hang it.
	MaxSteps int
}

// Run executes the self-check against the given tasks.
func (sc SelfCheck) Run(ctx context.Context, impl Adapter, tasks []domain.Task) domain.SelfCheckResult {
	result := domain.SelfCheckResult{
		Status:    "passed",
		Tasks:     len(tasks),
		CheckedAt: domain.Now(),
	}
	if len(tasks) == 0 {
		result.Status = "failed"
		result.InvalidReasons = []string{"the self-check needs at least one task"}
		return result
	}

	maxSteps := sc.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 4
	}

	responses, formatErrors, violations, usageOK := 0, 0, 0, 0
	for _, task := range tasks {
		contract := Contract{
			TaskID:                task.TaskID,
			Form:                  string(task.Form),
			Instruction:           task.Description.Instruction,
			TerminationConditions: terminationConditions(task),
			MaxSteps:              maxSteps,
			Seed:                  task.Seed(),
		}
		session, action, err := impl.StartSession(ctx, contract)
		if err != nil {
			formatErrors++
			violations++
			result.InvalidReasons = append(result.InvalidReasons,
				fmt.Sprintf("task %s: start_session failed: %v", task.TaskID, err))
			continue
		}
		responses++
		if action.Name == "" {
			formatErrors++
			violations++
			result.InvalidReasons = append(result.InvalidReasons,
				fmt.Sprintf("task %s: the adapter returned an empty action", task.TaskID))
		}

		stop, err := impl.ShouldStop(ctx, session)
		if err != nil {
			violations++
			result.InvalidReasons = append(result.InvalidReasons,
				fmt.Sprintf("task %s: should_stop failed: %v", task.TaskID, err))
		} else if stop.Stop && stop.Reason == "" {
			violations++
			result.InvalidReasons = append(result.InvalidReasons,
				fmt.Sprintf("task %s: the adapter stopped without a termination reason", task.TaskID))
		}

		if _, err := impl.ReportUsage(ctx, session); err != nil {
			violations++
			result.InvalidReasons = append(result.InvalidReasons,
				fmt.Sprintf("task %s: report_usage failed: %v", task.TaskID, err))
		} else {
			usageOK++
		}
		_ = impl.Close(ctx, session)
	}

	if responses > 0 {
		result.ActionFormatErrorRate = float64(formatErrors) / float64(responses)
	}
	result.ContractViolations = violations
	result.UsageReportingOK = len(tasks) > 0 && usageOK == len(tasks)
	if len(result.InvalidReasons) > 0 {
		result.Status = "failed"
	}
	return result
}

func terminationConditions(task domain.Task) []string {
	if task.InteractionContract == nil {
		return nil
	}
	return task.InteractionContract.TerminationConditions
}
