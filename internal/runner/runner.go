// Package runner executes an evaluation plan.
//
// A run is a set of trials, one per (task, repetition) pair. The runner
// produces the trajectory, the verdict and the flat outcome the metric layer
// consumes, and it classifies every trial that did not complete.
package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Baike12/b-eval/internal/adapter"
	"github.com/Baike12/b-eval/internal/domain"
	"github.com/Baike12/b-eval/internal/metrics"
	"github.com/Baike12/b-eval/internal/scoring"
)

// Runner executes plans.
type Runner struct {
	Scorers *scoring.Registry
	Metrics *metrics.Registry
	// Concurrency is the number of trials that run at the same time.
	Concurrency int
	// MaxConsecutiveViolations is the number of protocol violations in a row
	// that terminates a trial.
	MaxConsecutiveViolations int
}

// Input is everything one run needs.
type Input struct {
	Plan    domain.Plan
	TaskSet domain.TaskSet
	Adapter adapter.Adapter
	AgentID string
}

// Run executes the plan and returns the completed run.
func (r *Runner) Run(ctx context.Context, in Input) (domain.Run, error) {
	run := domain.Run{
		RunID:     newRunID(in.Plan),
		Plan:      in.Plan,
		Status:    "running",
		StartedAt: domain.Now(),
	}
	if in.Plan.Concurrency <= 0 {
		in.Plan.Concurrency = 1
	}

	type job struct {
		task       domain.Task
		trialIndex int
		seed       int64
	}
	var jobs []job
	for _, task := range in.TaskSet.Tasks {
		for i := 0; i < in.Plan.Repetitions; i++ {
			jobs = append(jobs, job{task: task, trialIndex: i, seed: seedFor(in.Plan, i)})
		}
	}

	results := make([]trialResult, len(jobs))
	concurrency := in.Plan.Concurrency
	if concurrency > len(jobs) {
		concurrency = len(jobs)
	}
	if concurrency < 1 {
		concurrency = 1
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	for i := range jobs {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = r.runTrial(ctx, in, jobs[idx].task, jobs[idx].trialIndex, jobs[idx].seed)
		}(i)
	}
	wg.Wait()

	for _, res := range results {
		run.Trials = append(run.Trials, res.Outcome)
		run.Trajectories = append(run.Trajectories, res.Trajectory)
		run.Verdicts = append(run.Verdicts, res.Verdict)
	}
	run.Status = "completed"
	run.FinishedAt = domain.Now()
	return run, nil
}

// trialResult bundles the three records one trial produces.
type trialResult struct {
	Outcome    domain.TrialOutcome
	Trajectory domain.Trajectory
	Verdict    domain.Verdict
}

// runTrial executes one task against the agent once.
func (r *Runner) runTrial(ctx context.Context, in Input, task domain.Task, trialIndex int, seed int64) trialResult {
	trialID := fmt.Sprintf("%s#%d#%d", task.TaskID, trialIndex, seed)
	traj := domain.Trajectory{
		TrialID: trialID,
		TaskID:  task.TaskID,
		AgentID: in.AgentID,
		Form:    task.Form,
		Seed:    seed,
	}
	started := time.Now()

	contract := adapter.Contract{
		TaskID:                task.TaskID,
		Form:                  string(task.Form),
		Instruction:           task.Description.Instruction,
		TerminationConditions: terminationConditions(task),
		MultiTurn:             multiTurn(task),
		MaxSteps:              maxSteps(task),
		MaxTokens:             task.Budget.Tokens,
		Seed:                  seed,
	}
	if task.InteractionContract != nil {
		contract.ActionSpace = task.InteractionContract.ActionSpace
		contract.ObservationShape = task.InteractionContract.ObservationShape
	}

	session, action, err := in.Adapter.StartSession(ctx, contract)
	if err != nil {
		traj.Termination = domain.TermAgentError
		traj.TerminationDetail = err.Error()
		return r.finish(task, trialIndex, seed, traj, domain.ClassModelFailure, err.Error(), domain.Verdict{})
	}
	defer func() { _ = in.Adapter.Close(ctx, session) }()

	maxSteps := contract.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 1
	}
	consecutiveViolations := 0
	answer := ""

	for step := 0; ; step++ {
		violation, formatError := r.inspectAction(task, action)
		if violation != "" {
			traj.Violations = append(traj.Violations, domain.Violation{
				Type:      violation,
				StepIndex: step,
				Detail:    action.Name,
			})
			consecutiveViolations++
			if r.maxViolations() > 0 && consecutiveViolations >= r.maxViolations() {
				traj.Termination = domain.TermProtocolViolation
				traj.TerminationDetail = fmt.Sprintf("%d consecutive protocol violations", consecutiveViolations)
				break
			}
		} else {
			consecutiveViolations = 0
		}

		if task.Form == domain.FormStaticQA && action.Content != "" {
			answer = action.Content
		}
		traj.Steps = append(traj.Steps, domain.Step{
			Index:       step,
			Action:      domain.Action{Index: step, Name: action.Name, Arguments: action.Arguments, Content: action.Content, At: time.Now()},
			DurationMS:  0,
			FormatError: formatError,
		})

		// The consumption is polled before the stop query so that a trial
		// which ran out of tokens is terminated by the budget rather than
		// being reported as the agent's own decision to finish.
		if usage, err := in.Adapter.ReportUsage(ctx, session); err == nil {
			traj.Usage = usage
		}

		if reason, exceeded := budgetExceeded(task, traj, time.Since(started)); exceeded {
			traj.Termination = reason
			traj.TerminationDetail = "a declared budget was reached"
			break
		}

		decision, err := in.Adapter.ShouldStop(ctx, session)
		if err != nil {
			traj.Termination = domain.TermAgentError
			traj.TerminationDetail = err.Error()
			break
		}
		if decision.Stop {
			traj.Termination = decision.Reason
			if traj.Termination == "" {
				traj.Termination = domain.TermCompleted
			}
			traj.TerminationDetail = decision.Detail
			break
		}
		if step+1 >= maxSteps {
			// The agent has used its whole step budget and has not declared
			// completion, so the trial ends on the step limit rather than on
			// the agent's own signal.
			traj.Termination = domain.TermStepLimit
			traj.TerminationDetail = fmt.Sprintf("step limit %d reached", maxSteps)
			break
		}

		session, action, err = in.Adapter.Advance(ctx, session, adapter.Observation{})
		if err != nil {
			traj.Termination = domain.TermAgentError
			traj.TerminationDetail = err.Error()
			break
		}
	}

	if usage, err := in.Adapter.ReportUsage(ctx, session); err == nil {
		traj.Usage = usage
	} else {
		traj.Violations = append(traj.Violations, domain.Violation{
			Type:   string(adapter.ViolationMissingUsage),
			Detail: err.Error(),
		})
	}
	traj.Answer = answer

	failure := domain.Classify(traj.Termination)
	verdict, judgeErr := scoring.Judge(task, traj, r.Scorers)
	if judgeErr != nil {
		return r.finish(task, trialIndex, seed, traj, failure, judgeErr.Error(), domain.Verdict{})
	}
	if failure != domain.ClassCompleted {
		// A trial that did not complete is a failure of the agent, not of the
		// judgement, so the verdict is forced to FAIL. An exception is the
		// platform's own problem and stays out of the denominator.
		if failure == domain.ClassPlatformFailure {
			verdict.Combined = domain.VerdictUnscored
			verdict.PrimaryPassed = false
		} else {
			verdict.Combined = domain.VerdictFail
			verdict.PrimaryPassed = false
			verdict.Score = 0
		}
	}
	return r.finish(task, trialIndex, seed, traj, failure, "", verdict)
}

// finish assembles the three records of a trial.
func (r *Runner) finish(task domain.Task, trialIndex int, seed int64, traj domain.Trajectory, failure domain.FailureClass, errDetail string, verdict domain.Verdict) trialResult {
	if verdict.TrialID == "" {
		verdict.TrialID = traj.TrialID
		verdict.TaskID = task.TaskID
		verdict.Combined = domain.VerdictFail
		verdict.JudgedAt = domain.Now()
	}
	outcome := domain.TrialOutcome{
		TrialID:     traj.TrialID,
		TaskID:      task.TaskID,
		TrialIndex:  trialIndex,
		Seed:        seed,
		Form:        task.Form,
		Domain:      firstDomain(task),
		Level1:      task.Capability.Level1,
		Difficulty:  task.Capability.Difficulty,
		Failure:     failure,
		Verdict:     verdict,
		Steps:       len(traj.Steps),
		BudgetSteps: task.Budget.Steps,
		Usage:       traj.Usage,
	}
	names := make([]string, 0, len(traj.Steps))
	repeated := 0
	for i, s := range traj.Steps {
		names = append(names, s.Action.Name)
		if s.FormatError {
			outcome.FormatErrors++
		}
		if i > 0 && s.Action.Name == traj.Steps[i-1].Action.Name {
			repeated++
		}
	}
	outcome.ActionNames = names
	outcome.RepeatedActions = repeated
	if errDetail != "" {
		outcome.Verdict.Evidence = map[string]any{"error": errDetail}
	}
	return trialResult{Outcome: outcome, Trajectory: traj, Verdict: outcome.Verdict}
}

// inspectAction checks one action against the task's contract.
//
// It returns the violation type and whether the action was unparseable. A
// violation is counted separately from the task outcome so an integration bug
// is never scored as weak agent capability.
func (r *Runner) inspectAction(task domain.Task, action adapter.Action) (string, bool) {
	if action.Name == "" && action.Content == "" {
		return string(adapter.ViolationEmptyResponse), true
	}
	if task.InteractionContract != nil && len(task.InteractionContract.ActionSpace) > 0 {
		if action.Name != "" && !contains(task.InteractionContract.ActionSpace, action.Name) {
			return string(adapter.ViolationActionOutOfSpace), true
		}
	}
	if len(action.Bundled) > 0 && isTerminating(task, action.Name) {
		return string(adapter.ViolationBundledTermination), true
	}
	return "", false
}

// budgetExceeded reports the first budget the trial has hit.
//
// The step budget is enforced by the loop guard instead, because an agent
// that used its last allowed step and then declared completion has finished
// the task rather than run out of room.
func budgetExceeded(task domain.Task, traj domain.Trajectory, elapsed time.Duration) (domain.TerminationReason, bool) {
	b := task.Budget
	if b.Tokens > 0 && traj.Usage.TotalTokens() >= b.Tokens {
		return domain.TermBudgetExhausted, true
	}
	if b.DurationSeconds > 0 && elapsed >= time.Duration(b.DurationSeconds)*time.Second {
		return domain.TermBudgetExhausted, true
	}
	if b.CostUSD > 0 && traj.Usage.CostUSD >= b.CostUSD {
		return domain.TermBudgetExhausted, true
	}
	return "", false
}

func (r *Runner) maxViolations() int {
	if r.MaxConsecutiveViolations > 0 {
		return r.MaxConsecutiveViolations
	}
	return 3
}

func terminationConditions(task domain.Task) []string {
	if task.InteractionContract == nil {
		return nil
	}
	return task.InteractionContract.TerminationConditions
}

func multiTurn(task domain.Task) bool {
	return task.InteractionContract != nil && task.InteractionContract.AllowsMultiTurn()
}

func maxSteps(task domain.Task) int {
	if task.Budget.Steps > 0 {
		return task.Budget.Steps
	}
	return 1
}

func firstDomain(task domain.Task) string {
	if len(task.Capability.Domains) == 0 {
		return "unassigned"
	}
	return task.Capability.Domains[0]
}

func isTerminating(task domain.Task, name string) bool {
	for _, condition := range terminationConditions(task) {
		if condition == name {
			return true
		}
	}
	return name == "terminate"
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// seedFor returns the seed of a repetition.
//
// The resume unit is (task ID, trial index, seed), so the seed of an
// index must be stable across restarts of the same plan.
func seedFor(plan domain.Plan, index int) int64 {
	if len(plan.Seeds) > 0 {
		return plan.Seeds[index%len(plan.Seeds)]
	}
	return int64(index + 1)
}

// newRunID derives a run identifier from the plan.
func newRunID(plan domain.Plan) string {
	return fmt.Sprintf("run-%s-%s", plan.PlanID, plan.Fingerprint()[:8])
}
