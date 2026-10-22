package gate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// StandardChecks returns the gate set of the repository.
//
// The order is the order the pipeline runs in: the cheap checks first, so a
// formatting mistake is reported before a test suite is spent on it.
func StandardChecks(env Env) []Check {
	return []Check{
		GoFmtCheck{Root: env.Root},
		CommandCheck{
			GateID:      "go-vet",
			Description: "go vet reports no suspicious constructs.",
			Dir:         env.Root,
			Name:        "go",
			Args:        []string{"vet", "./..."},
		},
		CommandCheck{
			GateID:      "go-test",
			Description: "Every Go package passes its unit tests.",
			Dir:         env.Root,
			Name:        "go",
			Args:        []string{"test", "./..."},
		},
		NPMCheck{
			GateID:      "web-typecheck",
			Description: "The frontend type-checks with no error.",
			Dir:         filepath.Join(env.Root, "web"),
			Script:      "typecheck",
		},
		NPMCheck{
			GateID:      "web-test",
			Description: "Every frontend test passes.",
			Dir:         filepath.Join(env.Root, "web"),
			Script:      "test",
		},
		NPMCheck{
			GateID:      "web-build",
			Description: "The frontend builds.",
			Dir:         filepath.Join(env.Root, "web"),
			Script:      "build",
		},
		CommandCheck{
			GateID:      "go-build",
			Description: "The platform compiles with the frontend bundle embedded.",
			Dir:         env.Root,
			Name:        "go",
			Args:        []string{"build", "./..."},
		},
		EnglishCheck(env.Root),
		LinksCheck(env.Root),
		IndexCheck(env.Root),
		SmokeCheck{GateID: "e2e-smoke", Description: "The vertical slice runs end to end over the HTTP API.", Root: env.Root},
	}
}

// GoFmtCheck verifies that every Go file is gofmt-clean.
//
// It is a dedicated check rather than a command check because gofmt exits
// zero when it finds unformatted files: it prints them instead. A gate that
// only looked at the exit code would pass on unformatted code.
type GoFmtCheck struct {
	Root string
}

// ID returns the gate identifier.
func (c GoFmtCheck) ID() string { return "go-fmt" }

// Describe returns what the gate protects.
func (c GoFmtCheck) Describe() string {
	return "Every Go file is formatted with gofmt."
}

// Run lists the unformatted files and fails when there are any.
func (c GoFmtCheck) Run(ctx context.Context) Result {
	res := Result{GateID: "go-fmt", Status: StatusPassed}
	cmd := exec.CommandContext(ctx, "gofmt", "-l", ".")
	cmd.Dir = c.Root
	out, err := cmd.Output()
	if err != nil {
		res.Status = StatusFailed
		res.Detail = err.Error()
		return res
	}
	files := strings.Fields(string(out))
	if len(files) > 0 {
		res.Status = StatusFailed
		res.Detail = fmt.Sprintf("%d file(s) are not gofmt-clean:\n%s", len(files), strings.Join(files, "\n"))
	}
	res.Evidence = map[string]any{"unformatted": len(files)}
	return res
}

// NPMCheck runs an npm script as a gate.
//
// The check is skipped when the frontend dependencies are not installed,
// because a missing dependency directory is not a verdict on the code. A
// skipped gate produces the insufficient-evidence classification, which is
// what keeps a skipped gate from being mistaken for a passing one.
type NPMCheck struct {
	GateID      string
	Description string
	Dir         string
	Script      string
}

// ID returns the gate identifier.
func (c NPMCheck) ID() string { return c.GateID }

// Describe returns what the gate protects.
func (c NPMCheck) Describe() string { return c.Description }

// Run executes the npm script.
func (c NPMCheck) Run(ctx context.Context) Result {
	res := Result{GateID: c.GateID, Status: StatusPassed}
	if _, err := os.Stat(filepath.Join(c.Dir, "node_modules")); err != nil {
		res.Status = StatusSkipped
		res.Detail = "the frontend dependencies are not installed; run npm install in web/"
		res.Evidence = map[string]any{"reason": "node_modules is missing"}
		return res
	}
	cmd := exec.CommandContext(ctx, "npm", "run", "--silent", c.Script)
	cmd.Dir = c.Dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		res.Status = StatusFailed
		res.Detail = truncate(strings.TrimSpace(string(out)), 4000)
		res.Evidence = map[string]any{
			"exit_code": exitCode(err),
			"command":   "npm run " + c.Script,
		}
		return res
	}
	res.Detail = truncate(strings.TrimSpace(string(out)), 2000)
	res.Evidence = map[string]any{
		"exit_code": 0,
		"command":   "npm run " + c.Script,
	}
	return res
}
