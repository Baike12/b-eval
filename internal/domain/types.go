// Package domain holds the platform's data model: tasks, task sets, plans,
// trials, trajectories and verdicts.
//
// Nothing in this package imports another b-eval package. Every other package
// depends on it, so keeping it dependency-free keeps the model stable while
// the layers above it change.
package domain

import "time"

// TaskForm selects the contract a task satisfies. The form decides which
// metadata groups the task carries and which admission gates apply to it.
//
// See Docs/Tech Design/TD-002-task-forms-and-metadata-contract.md.
type TaskForm string

const (
	// FormEnvAgentic runs the agent against a live environment and reads the
	// verdict from the environment's final state.
	FormEnvAgentic TaskForm = "environment-agentic"

	// FormStaticQA gives the agent a prompt and reads the verdict from the
	// answer text. No environment is brought up.
	FormStaticQA TaskForm = "static-qa"
)

// Valid reports whether the form is one the platform knows.
func (f TaskForm) Valid() bool {
	switch f {
	case FormEnvAgentic, FormStaticQA:
		return true
	}
	return false
}

// Budget caps what a single Trial may consume. A zero cap means "no cap for
// this dimension", so an unset field never silently blocks a run.
type Budget struct {
	Steps           int     `json:"steps"`
	Tokens          int     `json:"tokens"`
	DurationSeconds int     `json:"duration_seconds"`
	CostUSD         float64 `json:"cost_usd"`
	Errors          int     `json:"errors"`
}

// Capability is the task's label set: where it sits on the capability axis,
// the domain axis, and the difficulty tier.
type Capability struct {
	Level1     []string `json:"level1_dimensions"`
	Level2     []string `json:"level2_labels"`
	Difficulty string   `json:"difficulty"`
	Domains    []string `json:"domains"`
}

// Description carries the agent-facing instruction and the optional
// evaluator-facing note. The note never reaches the agent.
type Description struct {
	Instruction   string `json:"instruction"`
	EvaluatorNote string `json:"evaluator_note,omitempty"`
}

// InteractionContract is the shape of the interaction the task expects.
// For a static-QA task only TerminationConditions and MultiTurn are used.
type InteractionContract struct {
	ActionSpace           []string `json:"action_space,omitempty"`
	ObservationShape      string   `json:"observation_shape,omitempty"`
	TerminationConditions []string `json:"termination_conditions"`
	// MultiTurn is a pointer so that "declared as false" is distinguishable
	// from "not declared". The form contract requires the field to be
	// declared, and a plain bool cannot express that.
	MultiTurn *bool `json:"multi_turn,omitempty"`
}

// AllowsMultiTurn reports whether the task permits more than one turn.
func (c InteractionContract) AllowsMultiTurn() bool {
	return c.MultiTurn != nil && *c.MultiTurn
}

// InitialState is the environment the agent starts from. It exists only for
// an environment-agentic task.
type InitialState struct {
	SnapshotID         string         `json:"snapshot_id"`
	InitialData        map[string]any `json:"initial_data,omitempty"`
	InitialFiles       []string       `json:"initial_files,omitempty"`
	InitialCredentials []string       `json:"initial_credentials,omitempty"`
	Seed               int64          `json:"seed"`
}

// ScorerSpec is one scorer reference plus its combination weight. The
// platform resolves the reference to an implementation; the task never names
// code.
type ScorerSpec struct {
	Type    string         `json:"type"`
	Params  map[string]any `json:"params,omitempty"`
	Weight  float64        `json:"weight"`
	Primary bool           `json:"primary"`
}

// Judgement is the task's scoring declaration: which scorers, how they
// combine, and what the expected outcome is.
type Judgement struct {
	Combination     string         `json:"combination"`
	Scorers         []ScorerSpec   `json:"scorers"`
	Weights         []float64      `json:"weights,omitempty"`
	Primary         int            `json:"primary"`
	ReferenceAnswer string         `json:"reference_answer,omitempty"`
	TargetState     map[string]any `json:"target_state,omitempty"`
}

// ReferenceMaterial holds the optional artefacts of an environment-agentic
// task. It does not exist for a static-QA task.
type ReferenceMaterial struct {
	ReferenceSolution   string   `json:"reference_solution,omitempty"`
	ReferenceTrajectory []string `json:"reference_trajectory,omitempty"`
	PolicyDocument      string   `json:"policy_document,omitempty"`
}

// TaskMeta is the provenance of a task.
type TaskMeta struct {
	Author                string `json:"author"`
	Source                string `json:"source,omitempty"`
	DifficultyBasis       string `json:"difficulty_basis,omitempty"`
	EstimatedHumanMinutes int    `json:"estimated_human_minutes,omitempty"`
	EstimatedAgentMinutes int    `json:"estimated_agent_minutes,omitempty"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
	Deprecated            bool   `json:"deprecated"`
}

// Task is one evaluable item. The form decides which of the optional groups
// are populated; the validator refuses a task whose groups do not match its
// form.
type Task struct {
	TaskID              string               `json:"task_id"`
	Form                TaskForm             `json:"form"`
	DisplayName         string               `json:"display_name"`
	Benchmark           string               `json:"benchmark"`
	TaskSetVersion      string               `json:"task_set_version"`
	Capability          Capability           `json:"capability"`
	Description         Description          `json:"description"`
	Budget              Budget               `json:"budget"`
	InitialState        *InitialState        `json:"initial_state,omitempty"`
	InteractionContract *InteractionContract `json:"interaction_contract,omitempty"`
	Judgement           Judgement            `json:"judgement"`
	ReferenceMaterial   *ReferenceMaterial   `json:"reference_material,omitempty"`
	Meta                TaskMeta             `json:"meta"`
}

// Seed returns the task's seed, or zero when the task carries no initial
// state. A static-QA task has no environment and therefore no seed of its
// own; the run supplies the seed.
func (t Task) Seed() int64 {
	if t.InitialState == nil {
		return 0
	}
	return t.InitialState.Seed
}

// TaskSet is a versioned, frozen collection of tasks.
type TaskSet struct {
	TaskSetID   string `json:"task_set_id"`
	Version     string `json:"version"`
	Benchmark   string `json:"benchmark"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	Frozen      bool   `json:"frozen"`

	Tasks []Task `json:"tasks"`
}

// TerminationReason is the closed vocabulary of trial endings. Every trial
// ends with exactly one reason, and the reason is the primary input to
// failure attribution.
type TerminationReason string

const (
	TermCompleted          TerminationReason = "completed"
	TermDeclaredInfeasible TerminationReason = "declared_infeasible"
	TermStepLimit          TerminationReason = "step_limit"
	TermTurnLimit          TerminationReason = "turn_limit"
	TermBudgetExhausted    TerminationReason = "budget_exhausted"
	TermProtocolViolation  TerminationReason = "protocol_violation"
	TermAgentError         TerminationReason = "agent_error"
	TermEnvironmentError   TerminationReason = "environment_error"
	TermPlatformError      TerminationReason = "platform_error"
	TermCancelled          TerminationReason = "cancelled"
)

// FailureClass groups termination reasons by how they enter the statistics.
// Platform failures never enter a denominator; cancelled trials never do
// either.
type FailureClass string

const (
	ClassCompleted         FailureClass = "completed"
	ClassModelFailure      FailureClass = "model_failure"
	ClassPlatformFailure   FailureClass = "platform_failure"
	ClassBudgetExhausted   FailureClass = "budget_exhausted"
	ClassProtocolViolation FailureClass = "protocol_violation"
	ClassCancelled         FailureClass = "cancelled"
)

// Classify maps a termination reason onto its failure class.
func Classify(reason TerminationReason) FailureClass {
	switch reason {
	case TermCompleted, TermDeclaredInfeasible:
		return ClassCompleted
	case TermAgentError, TermEnvironmentError:
		return ClassModelFailure
	case TermBudgetExhausted, TermStepLimit, TermTurnLimit:
		return ClassBudgetExhausted
	case TermProtocolViolation:
		return ClassProtocolViolation
	case TermPlatformError:
		return ClassPlatformFailure
	case TermCancelled:
		return ClassCancelled
	}
	return ClassModelFailure
}

// CountsInDenominator reports whether a failure class contributes to the
// success denominator.
//
// Platform failures are excluded because they are the platform's fault, not
// the agent's; cancelled trials are excluded because the operator stopped
// them. Everything else counts, because an agent that crashes or runs out of
// budget has failed the task.
func (c FailureClass) CountsInDenominator() bool {
	switch c {
	case ClassPlatformFailure, ClassCancelled:
		return false
	}
	return true
}

// Action is one agent action as recorded in the trajectory.
type Action struct {
	Index     int            `json:"index"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
	Content   string         `json:"content,omitempty"`
	At        time.Time      `json:"at"`
}

// Observation is what the agent saw after an action.
type Observation struct {
	Index int            `json:"index"`
	Text  string         `json:"text,omitempty"`
	Data  map[string]any `json:"data,omitempty"`
}

// Usage is the resource consumption of a trial.
type Usage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	ModelCalls   int     `json:"model_calls"`
	ToolCalls    int     `json:"tool_calls"`
	ElapsedMS    int64   `json:"elapsed_ms"`
	CostUSD      float64 `json:"cost_usd"`
}

// TotalTokens is the token count the cost metric uses.
func (u Usage) TotalTokens() int { return u.InputTokens + u.OutputTokens }

// Violation is one protocol violation, counted separately from the task
// outcome so an integration bug is never scored as weak capability.
type Violation struct {
	Type      string `json:"type"`
	StepIndex int    `json:"step_index"`
	Detail    string `json:"detail,omitempty"`
}

// Step is one action/observation pair.
type Step struct {
	Index       int         `json:"index"`
	Action      Action      `json:"action"`
	Observation Observation `json:"observation"`
	DurationMS  int64       `json:"duration_ms"`
	FormatError bool        `json:"format_error"`
}

// Trajectory is the record of a trial's execution. It carries everything
// needed to re-judge the trial offline.
type Trajectory struct {
	TrialID           string            `json:"trial_id"`
	TaskID            string            `json:"task_id"`
	AgentID           string            `json:"agent_id"`
	Form              TaskForm          `json:"form"`
	Seed              int64             `json:"seed"`
	Steps             []Step            `json:"steps"`
	Termination       TerminationReason `json:"termination"`
	TerminationDetail string            `json:"termination_detail,omitempty"`
	Usage             Usage             `json:"usage"`
	Violations        []Violation       `json:"violations,omitempty"`
	Answer            string            `json:"answer,omitempty"`
}

// TerminationSet reports whether the trajectory already carries a
// termination reason.
func (t Trajectory) TerminationSet() bool { return t.Termination != "" }

// VerdictValue is the three-state result of one scorer call.
type VerdictValue string

const (
	VerdictPass     VerdictValue = "PASS"
	VerdictFail     VerdictValue = "FAIL"
	VerdictUnscored VerdictValue = "UNSCORED"
)

// ScorerResult is one scorer's outcome with its evidence.
type ScorerResult struct {
	Index    int            `json:"index"`
	Type     string         `json:"type"`
	Weight   float64        `json:"weight"`
	Value    VerdictValue   `json:"value"`
	Score    float64        `json:"score"`
	Detail   string         `json:"detail,omitempty"`
	Evidence map[string]any `json:"evidence,omitempty"`
}

// Verdict is the combined judgement of a trial.
type Verdict struct {
	TrialID       string         `json:"trial_id"`
	TaskID        string         `json:"task_id"`
	Scorers       []ScorerResult `json:"scorers"`
	Combined      VerdictValue   `json:"combined"`
	Score         float64        `json:"score"`
	PrimaryPassed bool           `json:"primary_passed"`
	Evidence      map[string]any `json:"evidence,omitempty"`
	JudgedAt      string         `json:"judged_at"`
}

// TrialOutcome is the flat view of a trial that the metric layer consumes.
type TrialOutcome struct {
	TrialID         string       `json:"trial_id"`
	TaskID          string       `json:"task_id"`
	TrialIndex      int          `json:"trial_index"`
	Seed            int64        `json:"seed"`
	Form            TaskForm     `json:"form"`
	Domain          string       `json:"domain"`
	Level1          []string     `json:"level1_dimensions"`
	Difficulty      string       `json:"difficulty"`
	Failure         FailureClass `json:"failure_class"`
	Verdict         Verdict      `json:"verdict"`
	Steps           int          `json:"steps"`
	BudgetSteps     int          `json:"budget_steps"`
	Usage           Usage        `json:"usage"`
	ActionNames     []string     `json:"action_names,omitempty"`
	FormatErrors    int          `json:"format_errors"`
	RepeatedActions int          `json:"repeated_actions"`
}

// Passed reports whether the trial's primary criterion was satisfied.
func (t TrialOutcome) Passed() bool {
	return t.Failure.CountsInDenominator() && t.Verdict.Combined == VerdictPass
}

// Plan is the description of one evaluation.
type Plan struct {
	PlanID         string  `json:"plan_id"`
	TaskSetID      string  `json:"task_set_id"`
	TaskSetVersion string  `json:"task_set_version"`
	AgentID        string  `json:"agent_id"`
	Repetitions    int     `json:"repetitions"`
	Seeds          []int64 `json:"seeds"`
	Aggregation    string  `json:"aggregation"`
	CIMethod       string  `json:"ci_method"`
	Concurrency    int     `json:"concurrency"`
	CreatedAt      string  `json:"created_at"`
}

// Fingerprint is a stable hash of the plan's scoring-relevant configuration.
// Two runs with the same fingerprint may be resumed; two without may not.
func (p Plan) Fingerprint() string {
	parts := []string{
		p.TaskSetID, p.TaskSetVersion, p.AgentID,
		itoa(p.Repetitions), p.Aggregation, p.CIMethod,
	}
	for _, s := range p.Seeds {
		parts = append(parts, itoa64(s))
	}
	return hashStrings(parts)
}

// Run is one execution of a plan.
type Run struct {
	RunID        string         `json:"run_id"`
	Plan         Plan           `json:"plan"`
	Status       string         `json:"status"`
	StartedAt    string         `json:"started_at"`
	FinishedAt   string         `json:"finished_at,omitempty"`
	Error        string         `json:"error,omitempty"`
	Trials       []TrialOutcome `json:"trials"`
	Trajectories []Trajectory   `json:"trajectories"`
	Verdicts     []Verdict      `json:"verdicts"`
}

// Agent is a registered system under test.
type Agent struct {
	AgentID   string           `json:"agent_id"`
	Name      string           `json:"name"`
	Version   string           `json:"version"`
	Shape     string           `json:"shape"`
	Endpoint  string           `json:"endpoint,omitempty"`
	Mode      string           `json:"mode,omitempty"`
	Options   map[string]any   `json:"options,omitempty"`
	SelfCheck *SelfCheckResult `json:"self_check,omitempty"`
	CreatedAt string           `json:"created_at"`
}

// SelfCheckResult is the outcome of the adapter self-check.
type SelfCheckResult struct {
	Status                string   `json:"status"`
	Tasks                 int      `json:"tasks"`
	ActionFormatErrorRate float64  `json:"action_format_error_rate"`
	ContractViolations    int      `json:"contract_violations"`
	UsageReportingOK      bool     `json:"usage_reporting_ok"`
	InvalidReasons        []string `json:"invalid_reasons,omitempty"`
	CheckedAt             string   `json:"checked_at"`
}
