// Package adapter is the boundary between the platform and the agent under
// test.
//
// The protocol fixes four capabilities — start a session, advance one step,
// decide whether to stop, report usage — and nothing else. It deliberately
// does not constrain how an agent plans, remembers or reflects, because those
// are the capabilities under test.
//
// See Docs/Tech Design/TD-001-unified-agent-adapter-protocol.md.
package adapter

import (
	"context"
	"fmt"

	"github.com/Baike12/b-eval/internal/domain"
)

// Contract is the task-side interaction contract handed to the adapter.
type Contract struct {
	TaskID                string   `json:"task_id"`
	Form                  string   `json:"form"`
	Instruction           string   `json:"instruction"`
	ActionSpace           []string `json:"action_space,omitempty"`
	ObservationShape      string   `json:"observation_shape,omitempty"`
	TerminationConditions []string `json:"termination_conditions,omitempty"`
	MultiTurn             bool     `json:"multi_turn"`
	MaxSteps              int      `json:"max_steps"`
	MaxTokens             int      `json:"max_tokens"`
	Seed                  int64    `json:"seed"`
}

// Session is the adapter's live state for one trial.
type Session struct {
	ID      string
	Payload map[string]any
}

// Action is what the adapter returns for one step.
//
// Bundled carries the names of any further actions that arrived in the same
// response. A terminating action must own its turn, so a bundle that
// contains one is rejected as a whole and counted as a protocol violation.
type Action struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
	Content   string         `json:"content,omitempty"`
	Bundled   []string       `json:"bundled,omitempty"`
}

// Observation is what the platform hands back to the adapter.
type Observation struct {
	Text string         `json:"text,omitempty"`
	Data map[string]any `json:"data,omitempty"`
}

// StopDecision is the adapter's answer to the stop query.
type StopDecision struct {
	Stop   bool                     `json:"stop"`
	Reason domain.TerminationReason `json:"reason,omitempty"`
	Detail string                   `json:"detail,omitempty"`
}

// Adapter is the contract every system under test must satisfy.
type Adapter interface {
	// StartSession opens a session for the task and returns the adapter's
	// first action.
	StartSession(ctx context.Context, c Contract) (Session, Action, error)
	// Advance hands the latest observation to the agent and returns the next
	// action.
	Advance(ctx context.Context, s Session, obs Observation) (Session, Action, error)
	// ShouldStop asks the agent whether the run may end.
	ShouldStop(ctx context.Context, s Session) (StopDecision, error)
	// ReportUsage returns the cumulative consumption of the session.
	ReportUsage(ctx context.Context, s Session) (domain.Usage, error)
	// Close releases whatever the adapter holds for the session.
	Close(ctx context.Context, s Session) error
}

// Registry resolves a registered agent to an adapter implementation.
type Registry struct {
	agents map[string]domain.Agent
	byID   map[string]Adapter
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{agents: map[string]domain.Agent{}, byID: map[string]Adapter{}}
}

// RegisterAgent records an agent and the adapter that serves it.
func (r *Registry) RegisterAgent(a domain.Agent, impl Adapter) {
	r.agents[a.AgentID] = a
	r.byID[a.AgentID] = impl
}

// Agents lists the registered agents.
func (r *Registry) Agents() []domain.Agent {
	out := make([]domain.Agent, 0, len(r.agents))
	for _, a := range r.agents {
		out = append(out, a)
	}
	return out
}

// Agent resolves a registered agent.
func (r *Registry) Agent(id string) (domain.Agent, bool) {
	a, ok := r.agents[id]
	return a, ok
}

// Adapter resolves the adapter implementation of a registered agent.
func (r *Registry) Adapter(id string) (Adapter, bool) {
	a, ok := r.byID[id]
	return a, ok
}

// ErrNoAdapter is returned when an agent has no adapter implementation.
var ErrNoAdapter = fmt.Errorf("agent has no adapter implementation")

// ViolationType is the closed set of protocol violations.
type ViolationType string

const (
	ViolationMalformedAction    ViolationType = "malformed_action"
	ViolationActionOutOfSpace   ViolationType = "action_outside_declared_space"
	ViolationBundledTermination ViolationType = "bundled_termination_action"
	ViolationEmptyResponse      ViolationType = "empty_response"
	ViolationMissingUsage       ViolationType = "missing_usage_report"
	ViolationRepeatedInvalid    ViolationType = "repeated_invalid_action"
)
