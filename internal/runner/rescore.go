package runner

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Baike12/b-eval/internal/domain"
	"github.com/Baike12/b-eval/internal/scoring"
)

// Rejudge re-scores already persisted trajectories without calling the agent.
//
// The trajectories are the platform's durable record, so judgement can be
// repeated with a different scorer set, a different judge model or a changed
// rubric while the original run stays untouched. The caller stores the
// result as a new run whose RescoredFrom names its source.
func Rejudge(taskSet domain.TaskSet, agentID string, trajectories []domain.Trajectory, reg *scoring.Registry) ([]domain.TrialOutcome, []domain.Verdict, error) {
	byID := make(map[string]domain.Task, len(taskSet.Tasks))
	for _, task := range taskSet.Tasks {
		byID[task.TaskID] = task
	}

	outcomes := make([]domain.TrialOutcome, 0, len(trajectories))
	verdicts := make([]domain.Verdict, 0, len(trajectories))
	for _, original := range trajectories {
		task, ok := byID[original.TaskID]
		if !ok {
			// A trajectory whose task is no longer in the task set cannot be
			// judged. Skipping it would silently shrink the denominator, so
			// the whole re-judgement is refused instead.
			return nil, nil, fmt.Errorf("task %q of trial %q is not in task set %q", original.TaskID, original.TrialID, taskSet.TaskSetID)
		}
		traj := original
		failure := domain.Classify(traj.Termination)
		verdict, err := scoring.Judge(task, traj, reg)
		if err != nil {
			verdict = domain.Verdict{
				TrialID:  traj.TrialID,
				TaskID:   task.TaskID,
				Combined: domain.VerdictUnscored,
				Evidence: map[string]any{"error": err.Error()},
			}
		} else {
			absorbJudgeUsage(&traj.Usage, verdict.Scorers)
		}
		if failure != domain.ClassCompleted {
			// The same rule as a fresh execution: a trial that did not
			// finish is a failure of the agent, not of the judgement, unless
			// the platform itself failed.
			if failure == domain.ClassPlatformFailure {
				verdict.Combined = domain.VerdictUnscored
				verdict.PrimaryPassed = false
			} else {
				verdict.Combined = domain.VerdictFail
				verdict.PrimaryPassed = false
				verdict.Score = 0
			}
		}

		trialIndex := traj.TrialIndex
		if trialIndex == 0 {
			// Runs recorded before TrialIndex was persisted carry the index
			// only in the trial ID.
			trialIndex = trialIndexFromID(traj.TrialID)
		}
		outcome := buildOutcome(task, trialIndex, traj.Seed, traj, failure, "", verdict)
		outcomes = append(outcomes, outcome)
		verdicts = append(verdicts, outcome.Verdict)
	}
	return outcomes, verdicts, nil
}

// trialIndexFromID recovers the trial index from the "<task>#<index>#<seed>"
// trial identifier.
func trialIndexFromID(trialID string) int {
	parts := strings.Split(trialID, "#")
	if len(parts) < 3 {
		return 0
	}
	index, err := strconv.Atoi(parts[len(parts)-2])
	if err != nil || index < 0 {
		return 0
	}
	return index
}
