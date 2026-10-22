package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Baike12/b-eval/internal/domain"
)

// HTTPAdapter drives an agent that is exposed as an HTTP service.
//
// Two shapes are supported. The "single" shape is one request per trial and
// is what a static-QA agent normally needs. The "step" shape is the full
// four-capability protocol and is what an environment agent needs.
type HTTPAdapter struct {
	Endpoint string
	Mode     string
	Client   *http.Client
	Headers  map[string]string
}

// NewHTTPAdapter builds an adapter for an endpoint.
func NewHTTPAdapter(endpoint, mode string, headers map[string]string) *HTTPAdapter {
	return &HTTPAdapter{
		Endpoint: endpoint,
		Mode:     mode,
		Headers:  headers,
		Client:   &http.Client{Timeout: 120 * time.Second},
	}
}

// httpSession is the wire form of the adapter's session state.
type httpSession struct {
	ID string `json:"session_id"`
}

type startRequest struct {
	TaskID      string   `json:"task_id"`
	Form        string   `json:"form"`
	Instruction string   `json:"instruction"`
	ActionSpace []string `json:"action_space,omitempty"`
	Termination []string `json:"termination_conditions,omitempty"`
	MultiTurn   bool     `json:"multi_turn"`
	MaxSteps    int      `json:"max_steps"`
	MaxTokens   int      `json:"max_tokens"`
	Seed        int64    `json:"seed"`
}

type stepRequest struct {
	SessionID   string `json:"session_id"`
	Observation string `json:"observation,omitempty"`
}

type answerRequest struct {
	TaskID      string `json:"task_id"`
	Instruction string `json:"instruction"`
	Seed        int64  `json:"seed"`
}

type agentResponse struct {
	SessionID string `json:"session_id"`
	Action    struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
		Content   string         `json:"content"`
		Bundled   []string       `json:"bundled"`
	} `json:"action"`
	Answer string `json:"answer"`
	Stop   bool   `json:"stop"`
	Reason string `json:"reason"`
	Usage  struct {
		InputTokens  int     `json:"input_tokens"`
		OutputTokens int     `json:"output_tokens"`
		ModelCalls   int     `json:"model_calls"`
		ToolCalls    int     `json:"tool_calls"`
		ElapsedMS    int64   `json:"elapsed_ms"`
		CostUSD      float64 `json:"cost_usd"`
	} `json:"usage"`
}

// StartSession opens a session and returns the agent's first action.
func (a *HTTPAdapter) StartSession(ctx context.Context, c Contract) (Session, Action, error) {
	if a.Mode == "single" {
		return a.singleShot(ctx, c)
	}
	var resp agentResponse
	body := startRequest{
		TaskID:      c.TaskID,
		Form:        c.Form,
		Instruction: c.Instruction,
		ActionSpace: c.ActionSpace,
		Termination: c.TerminationConditions,
		MultiTurn:   c.MultiTurn,
		MaxSteps:    c.MaxSteps,
		MaxTokens:   c.MaxTokens,
		Seed:        c.Seed,
	}
	if err := a.post(ctx, "/session", body, &resp); err != nil {
		return Session{}, Action{}, err
	}
	return Session{ID: resp.SessionID, Payload: map[string]any{"usage": resp.Usage}}, actionOf(resp), nil
}

// Advance hands the latest observation to the agent.
func (a *HTTPAdapter) Advance(ctx context.Context, s Session, obs Observation) (Session, Action, error) {
	var resp agentResponse
	if err := a.post(ctx, "/step", stepRequest{SessionID: s.ID, Observation: obs.Text}, &resp); err != nil {
		return s, Action{}, err
	}
	return Session{ID: resp.SessionID, Payload: map[string]any{"usage": resp.Usage}}, actionOf(resp), nil
}

// ShouldStop asks the agent whether the run may end.
func (a *HTTPAdapter) ShouldStop(ctx context.Context, s Session) (StopDecision, error) {
	var resp agentResponse
	if err := a.post(ctx, "/stop", map[string]string{"session_id": s.ID}, &resp); err != nil {
		return StopDecision{}, err
	}
	decision := StopDecision{Stop: resp.Stop}
	if resp.Reason != "" {
		decision.Reason = domain.TerminationReason(resp.Reason)
	}
	return decision, nil
}

// ReportUsage returns the cumulative consumption of the session.
func (a *HTTPAdapter) ReportUsage(ctx context.Context, s Session) (domain.Usage, error) {
	var resp agentResponse
	if err := a.post(ctx, "/usage", map[string]string{"session_id": s.ID}, &resp); err != nil {
		return domain.Usage{}, err
	}
	return usageOf(resp), nil
}

// Close releases the session.
func (a *HTTPAdapter) Close(ctx context.Context, s Session) error {
	// The session is server-side state; a failure to release it must not
	// fail the trial, because the trial's outcome has already been decided.
	_ = a.post(ctx, "/close", map[string]string{"session_id": s.ID}, &struct{}{})
	return nil
}

// singleShot performs the one-request-per-trial flow.
func (a *HTTPAdapter) singleShot(ctx context.Context, c Contract) (Session, Action, error) {
	var resp agentResponse
	body := answerRequest{TaskID: c.TaskID, Instruction: c.Instruction, Seed: c.Seed}
	if err := a.post(ctx, "/answer", body, &resp); err != nil {
		return Session{}, Action{}, err
	}
	action := Action{Name: "answer", Content: resp.Answer}
	if action.Content == "" {
		action.Content = resp.Answer
	}
	return Session{ID: resp.SessionID, Payload: map[string]any{"usage": resp.Usage, "answer": resp.Answer}}, action, nil
}

// post sends a JSON request and decodes a JSON response.
func (a *HTTPAdapter) post(ctx context.Context, path string, in any, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshal request for %s: %w", path, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.Endpoint+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request for %s: %w", path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range a.Headers {
		req.Header.Set(k, v)
	}
	resp, err := a.Client.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", path, err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response from %s: %w", path, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("agent returned %d from %s: %s", resp.StatusCode, path, truncate(string(payload), 200))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("decode response from %s: %w", path, err)
	}
	return nil
}

func actionOf(r agentResponse) Action {
	return Action{
		Name:      r.Action.Name,
		Arguments: r.Action.Arguments,
		Content:   r.Action.Content,
		Bundled:   r.Action.Bundled,
	}
}

func usageOf(r agentResponse) domain.Usage {
	return domain.Usage{
		InputTokens:  r.Usage.InputTokens,
		OutputTokens: r.Usage.OutputTokens,
		ModelCalls:   r.Usage.ModelCalls,
		ToolCalls:    r.Usage.ToolCalls,
		ElapsedMS:    r.Usage.ElapsedMS,
		CostUSD:      r.Usage.CostUSD,
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
