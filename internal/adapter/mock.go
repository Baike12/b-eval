package adapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/Baike12/b-eval/internal/domain"
)

// MockAdapter is a deterministic agent used by the self-check, the end-to-end
// smoke run and the tests.
//
// It is deterministic on purpose: a gate that depends on a real model would
// be a gate that fails for reasons unrelated to the code it guards.
type MockAdapter struct {
	// Answers maps a task ID to the answer the mock produces. A task that is
	// not listed gets DefaultAnswer.
	Answers map[string]string
	// DefaultAnswer is used for any task not listed in Answers.
	DefaultAnswer string
	// Usage is the consumption the mock reports.
	Usage domain.Usage
	// FailWith, when set, makes the mock fail the trial with this error.
	FailWith error
}

// NewMockAdapter returns a mock that answers every task with the same text.
func NewMockAdapter(answer string) *MockAdapter {
	return &MockAdapter{Answers: map[string]string{}, DefaultAnswer: answer}
}

// StartSession returns the first action.
func (m *MockAdapter) StartSession(ctx context.Context, c Contract) (Session, Action, error) {
	if m.FailWith != nil {
		return Session{}, Action{}, m.FailWith
	}
	answer, ok := m.Answers[c.TaskID]
	if !ok {
		answer = m.DefaultAnswer
	}
	return Session{ID: "mock-" + c.TaskID}, Action{Name: "answer", Content: answer}, nil
}

// Advance returns the same answer; the mock is single-turn.
func (m *MockAdapter) Advance(ctx context.Context, s Session, obs Observation) (Session, Action, error) {
	return s, Action{Name: "answer"}, nil
}

// ShouldStop reports that the run is complete.
func (m *MockAdapter) ShouldStop(ctx context.Context, s Session) (StopDecision, error) {
	return StopDecision{Stop: true, Reason: domain.TermCompleted}, nil
}

// ReportUsage returns the configured usage.
func (m *MockAdapter) ReportUsage(ctx context.Context, s Session) (domain.Usage, error) {
	return m.Usage, nil
}

// Close releases nothing.
func (m *MockAdapter) Close(ctx context.Context, s Session) error { return nil }

// ScriptedAdapter replays a fixed action sequence.
//
// It exists so the action-sequence scorer and the protocol-violation
// accounting can be exercised without a real agent.
type ScriptedAdapter struct {
	// Actions is the sequence the adapter emits, one per advance call.
	Actions []Action
	// StopAfter is the number of actions before the adapter asks to stop.
	StopAfter int
	// Usage is the consumption the adapter reports.
	Usage domain.Usage
	index int
}

// StartSession returns the first scripted action.
func (s *ScriptedAdapter) StartSession(ctx context.Context, c Contract) (Session, Action, error) {
	return Session{ID: "scripted"}, s.next(), nil
}

// Advance returns the next scripted action.
func (s *ScriptedAdapter) Advance(ctx context.Context, sess Session, obs Observation) (Session, Action, error) {
	return sess, s.next(), nil
}

// ShouldStop reports whether the scripted sequence is exhausted.
func (s *ScriptedAdapter) ShouldStop(ctx context.Context, sess Session) (StopDecision, error) {
	if s.StopAfter > 0 && s.index >= s.StopAfter {
		return StopDecision{Stop: true, Reason: domain.TermCompleted}, nil
	}
	return StopDecision{Stop: false}, nil
}

// ReportUsage returns the configured usage.
func (s *ScriptedAdapter) ReportUsage(ctx context.Context, sess Session) (domain.Usage, error) {
	return s.Usage, nil
}

// Close releases nothing.
func (s *ScriptedAdapter) Close(ctx context.Context, sess Session) error { return nil }

func (s *ScriptedAdapter) next() Action {
	if s.index >= len(s.Actions) {
		return Action{Name: "done"}
	}
	a := s.Actions[s.index]
	s.index++
	return a
}

// ParseTextAction parses the text-action shape.
//
// The format is `Action: <name>` optionally followed by `Action Input: <json>`
// on the next line, which is the shape several upstream agents emit. A parse
// failure is returned as an error so the runner can count it as an
// action-format error instead of guessing.
func ParseTextAction(text string) (Action, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return Action{}, fmt.Errorf("empty response")
	}
	name := ""
	input := ""
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "Action:"):
			name = strings.TrimSpace(strings.TrimPrefix(line, "Action:"))
		case strings.HasPrefix(line, "Action Input:"):
			input = strings.TrimSpace(strings.TrimPrefix(line, "Action Input:"))
		}
	}
	if name == "" {
		return Action{}, fmt.Errorf("response does not contain an action")
	}
	action := Action{Name: name, Content: trimmed}
	if input != "" {
		action.Arguments = map[string]any{"input": input}
	}
	return action, nil
}
