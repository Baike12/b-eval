// Package gate is the repository's quality gate system.
//
// The gate runs a fixed set of checks, refuses to pass when any of them
// fails, and records what it did as append-only evidence. The evidence is
// never rewritten: a gate run that changed its own history would make the
// history worthless.
//
// The model follows two references. The skeleton is bblog's: a check is an
// executable command, the script runs numbered stages, and a failing stage
// blocks the next one. The interior is loopx's: every gate produces a
// classification and a recommended action, and every gate appends an
// evidence record that is only ever added to.
package gate

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Status is the outcome of one gate.
type Status string

const (
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
)

// Classification is the judgement a gate produces.
//
// The three values answer one question: can the next stage run on the
// strength of this gate?
type Classification string

const (
	// ClassificationReady means the gate passed and the pipeline may proceed.
	ClassificationReady Classification = "ready"
	// ClassificationBlocked means the gate failed and the pipeline must stop.
	ClassificationBlocked Classification = "blocked"
	// ClassificationInsufficientEvidence means the gate could not run, so
	// nothing is known about the property it guards.
	ClassificationInsufficientEvidence Classification = "insufficient_evidence"
)

// Recommended actions attached to a classification.
const (
	ActionProceed     = "proceed"
	ActionFix         = "fix the failing check and re-run the gate"
	ActionInvestigate = "run the check manually; the gate produced no verdict"
)

// Result is the outcome of one gate.
type Result struct {
	GateID            string         `json:"gate_id"`
	Status            Status         `json:"status"`
	Classification    Classification `json:"classification"`
	RecommendedAction string         `json:"recommended_action"`
	Description       string         `json:"description"`
	DurationMS        int64          `json:"duration_ms"`
	Detail            string         `json:"detail,omitempty"`
	Evidence          map[string]any `json:"evidence,omitempty"`
	CheckedAt         string         `json:"checked_at"`
}

// Check is one gate.
type Check interface {
	// ID is the stable identifier of the gate.
	ID() string
	// Describe states what the gate protects.
	Describe() string
	// Run executes the gate.
	Run(ctx context.Context) Result
}

// Env is the environment the checks run in.
type Env struct {
	// Root is the repository root.
	Root string
	// DataDir is where the platform stores its records.
	DataDir string
	// GateDir is where the gate evidence is written.
	GateDir string
	// SkipSlow marks the checks that are excluded from the fast subset.
	SkipSlow bool
}

// Run executes the given checks in order and returns their results.
//
// The checks run in order and a failing check does not stop the run: the
// point of the evidence file is that a reader can see every gate's verdict,
// not only the first failure. The overall verdict is decided by the caller.
func Run(ctx context.Context, env Env, checks []Check) []Result {
	results := make([]Result, 0, len(checks))
	for _, c := range checks {
		start := time.Now()
		res := c.Run(ctx)
		if res.GateID == "" {
			res.GateID = c.ID()
		}
		if res.Description == "" {
			res.Description = c.Describe()
		}
		if res.Status == "" {
			res.Status = StatusSkipped
		}
		if res.Classification == "" {
			res.Classification = Classify(res.Status)
		}
		if res.RecommendedAction == "" {
			res.RecommendedAction = ActionFor(res.Classification)
		}
		res.DurationMS = time.Since(start).Milliseconds()
		if res.CheckedAt == "" {
			res.CheckedAt = time.Now().UTC().Format(time.RFC3339)
		}
		results = append(results, res)
	}
	return results
}

// Classify maps a gate status onto a classification.
func Classify(s Status) Classification {
	switch s {
	case StatusPassed:
		return ClassificationReady
	case StatusFailed:
		return ClassificationBlocked
	}
	return ClassificationInsufficientEvidence
}

// ActionFor returns the recommended action of a classification.
func ActionFor(c Classification) string {
	switch c {
	case ClassificationReady:
		return ActionProceed
	case ClassificationBlocked:
		return ActionFix
	}
	return ActionInvestigate
}

// Failed reports whether any result blocked the pipeline.
func Failed(results []Result) bool {
	for _, r := range results {
		if r.Classification == ClassificationBlocked {
			return true
		}
	}
	return false
}

// CommandCheck runs an external command as a gate.
type CommandCheck struct {
	// GateID is the stable identifier of the gate.
	GateID string
	// Description states what the gate protects.
	Description string
	// Dir is the working directory of the command.
	Dir string
	// Name is the executable.
	Name string
	// Args are the arguments.
	Args []string
	// Env holds extra environment variables for the command.
	Env []string
	// Slow marks the check as excluded from the fast subset.
	Slow bool
}

// ID returns the gate identifier.
func (c CommandCheck) ID() string { return c.GateID }

// Describe returns what the gate protects.
func (c CommandCheck) Describe() string { return c.Description }

// Run executes the command and turns its exit code into a verdict.
func (c CommandCheck) Run(ctx context.Context) Result {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	cmd.Env = append(os.Environ(), c.Env...)
	out, err := cmd.CombinedOutput()
	res := Result{GateID: c.GateID, Status: StatusPassed}
	text := strings.TrimSpace(string(out))
	if err != nil {
		res.Status = StatusFailed
		res.Detail = truncate(text, 4000)
		res.Evidence = map[string]any{
			"exit_code": exitCode(err),
			"command":   strings.Join(append([]string{c.Name}, c.Args...), " "),
		}
		return res
	}
	res.Detail = truncate(text, 2000)
	res.Evidence = map[string]any{
		"exit_code": 0,
		"command":   strings.Join(append([]string{c.Name}, c.Args...), " "),
	}
	return res
}

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if ok := asExitError(err, &exitErr); ok {
		return exitErr.ExitCode()
	}
	return -1
}

func asExitError(err error, target **exec.ExitError) bool {
	if e, ok := err.(*exec.ExitError); ok {
		*target = e
		return true
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n... (truncated)"
}

// SortResults orders the results by gate identifier so the evidence file is
// diffable between runs.
func SortResults(results []Result) {
	sort.Slice(results, func(i, j int) bool { return results[i].GateID < results[j].GateID })
}

// JoinPath joins a path relative to the repository root.
func (e Env) JoinPath(parts ...string) string {
	return filepath.Join(append([]string{e.Root}, parts...)...)
}
