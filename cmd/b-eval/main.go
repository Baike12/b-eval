// Command b-eval runs the agent evaluation platform.
//
// Subcommands:
//
//	serve    start the HTTP server
//	gate     run the repository's quality gates
//	seed     write a demonstration task set and agent into the data directory
//	version  print the build version
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Baike12/b-eval/internal/adapter"
	"github.com/Baike12/b-eval/internal/api"
	"github.com/Baike12/b-eval/internal/domain"
	"github.com/Baike12/b-eval/internal/gate"
	"github.com/Baike12/b-eval/internal/store"
	"github.com/Baike12/b-eval/internal/web"
)

// version is injected at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "serve":
		os.Exit(runServe(os.Args[2:]))
	case "gate":
		os.Exit(runGate(os.Args[2:]))
	case "seed":
		os.Exit(runSeed(os.Args[2:]))
	case "version", "-v", "--version":
		fmt.Println(version)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `b-eval — agent evaluation platform

Subcommands:
  serve    start the HTTP server
  gate     run the repository's quality gates
  seed     write a demonstration task set and agent into the data directory
  version  print the build version`)
}

func runServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:8091", "listen address")
	dataDir := fs.String("data", "data", "data directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	st := store.New(*dataDir)
	srv := api.NewServer(st)
	if err := loadAgents(st, srv); err != nil {
		log.Printf("warn: could not restore the registered agents: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", srv.Handler())
	mux.Handle("/", web.Handler())

	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Printf("b-eval %s listening on %s (data: %s)", version, *addr, *dataDir)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("error: %v", err)
			stop <- syscall.SIGTERM
		}
	}()
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("warn: shutdown: %v", err)
		return 1
	}
	return 0
}

// loadAgents restores the registered agents from the data directory.
//
// An agent is restored without its self-check result when it has none, so
// that a restored agent must pass the self-check before it can be used in a
// run.
func loadAgents(st *store.Store, srv *api.Server) error {
	agents, err := st.Agents()
	if err != nil {
		return err
	}
	for _, a := range agents {
		impl := adapter.NewMockAdapter("")
		// The mock adapter's answers are part of the agent's registration.
		// Without them a restored mock would answer every task with an
		// empty string, which looks like a total capability failure rather
		// than like a missing configuration.
		if raw, ok := a.Options["answers"]; ok {
			if answers, ok := raw.(map[string]any); ok {
				for taskID, value := range answers {
					if text, ok := value.(string); ok {
						impl.Answers[taskID] = text
					}
				}
			}
		}
		srv.Agents.RegisterAgent(a, impl)
	}
	return nil
}

func runGate(args []string) int {
	fs := flag.NewFlagSet("gate", flag.ContinueOnError)
	root := fs.String("root", ".", "repository root")
	dataDir := fs.String("data", "", "data directory for the end-to-end smoke run")
	only := fs.String("only", "", "run only the gate with this identifier")
	asJSON := fs.Bool("json", false, "print the results as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	env := gate.Env{
		Root:    absRoot,
		DataDir: *dataDir,
		GateDir: filepath.Join(absRoot, ".gates"),
	}

	checks := gate.StandardChecks(env)
	if *only != "" {
		// The identifier list is comma separated, so a staged script can
		// run one stage's gates without naming each one on its own command
		// line.
		wanted := map[string]bool{}
		for _, id := range strings.Split(*only, ",") {
			wanted[strings.TrimSpace(id)] = true
		}
		filtered := make([]gate.Check, 0, len(checks))
		for _, c := range checks {
			if wanted[c.ID()] {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) == 0 {
			fmt.Fprintf(os.Stderr, "error: no gate matches %q\n", *only)
			return 2
		}
		checks = filtered
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	results := gate.Run(ctx, env, checks)
	evidencePath, err := gate.AppendEvidence(env.GateDir, results)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	summaryPath, err := gate.WriteSummary(env.GateDir, results)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
	} else {
		for _, r := range results {
			fmt.Printf("%-16s %-8s %-22s %s\n", r.GateID, r.Status, r.Classification, r.RecommendedAction)
			if r.Status == gate.StatusFailed && r.Detail != "" {
				fmt.Printf("                 %s\n", firstLine(r.Detail))
			}
		}
		fmt.Printf("\nevidence: %s\nsummary:  %s\n", evidencePath, summaryPath)
	}

	if gate.Failed(results) {
		return 1
	}
	return 0
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}

func runSeed(args []string) int {
	fs := flag.NewFlagSet("seed", flag.ContinueOnError)
	dataDir := fs.String("data", "data", "data directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	st := store.New(*dataDir)
	ts, agents := demonstrationRecords()

	if err := st.SaveTaskSet(ts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	for _, a := range agents {
		if err := st.SaveAgent(a); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
	}
	fmt.Printf("seeded task set %q (%d tasks) and %d agents into %s\n", ts.TaskSetID, len(ts.Tasks), len(agents), *dataDir)
	return 0
}

// demonstrationRecords builds the sample task set and agents.
//
// The task set is static-QA on purpose: it is the form that needs no
// environment, so a reader can start a run without provisioning anything.
func demonstrationRecords() (domain.TaskSet, []domain.Agent) {
	now := domain.Now()
	tasks := []domain.Task{
		demoTask("T-001", "Refund eligibility", "A customer asks whether their order qualifies for a free return. The order is $120 and they have the receipt.", "orders over $100 qualify for a free return", []string{"30 days", "receipt"}, "C4", "D7", now),
		demoTask("T-002", "Return window", "A customer asks how long they have to return an item bought last week.", "the return window is 30 days from delivery", []string{"30 days"}, "C4", "D7", now),
		demoTask("T-003", "Shipping cost", "A customer asks how much shipping costs for a small order.", "shipping is free above $50", []string{"free", "$50"}, "C1", "D7", now),
		demoTask("T-004", "Invoice copy", "A customer asks where to download a copy of their invoice.", "the invoice is available in the billing section of the account page", []string{"billing"}, "C2", "D7", now),
	}
	ts := domain.TaskSet{
		TaskSetID:   "demo",
		Version:     "v1",
		Benchmark:   "demo-faq",
		Description: "A small static-QA task set used by the demonstration seed.",
		CreatedAt:   now,
		Tasks:       tasks,
	}

	agents := []domain.Agent{
		{
			AgentID:   "demo-good",
			Name:      "Demo agent (answers correctly)",
			Version:   "1",
			Shape:     "mock",
			CreatedAt: now,
			Options: map[string]any{
				"answers": map[string]string{
					"T-001": "Orders over $100 qualify for a free return within 30 days, and you will need the receipt.",
					"T-002": "The return window is 30 days from delivery.",
					"T-003": "Shipping is free on orders above $50.",
					"T-004": "You can download the invoice from the billing section of your account page.",
				},
			},
		},
		{
			AgentID:   "demo-weak",
			Name:      "Demo agent (answers partially)",
			Version:   "1",
			Shape:     "mock",
			CreatedAt: now,
			Options: map[string]any{
				"answers": map[string]string{
					"T-001": "I am not sure about the return policy.",
					"T-002": "The return window is 30 days from delivery.",
					"T-003": "I am not sure about the shipping cost.",
					"T-004": "You can download the invoice from the billing section of your account page.",
				},
			},
		},
	}
	return ts, agents
}

// boolPtr returns a pointer to a bool value.
func boolPtr(v bool) *bool { return &v }

// demoTask builds one demonstration task.
func demoTask(id, name, instruction, reference string, needles []string, capability, domainCode, now string) domain.Task {
	needleList := make([]any, 0, len(needles))
	for _, n := range needles {
		needleList = append(needleList, n)
	}
	return domain.Task{
		TaskID:         id,
		Form:           domain.FormStaticQA,
		DisplayName:    name,
		Benchmark:      "demo-faq",
		TaskSetVersion: "v1",
		Capability: domain.Capability{
			Level1:     []string{capability},
			Level2:     []string{"requirement-clarification"},
			Difficulty: "L1",
			Domains:    []string{domainCode},
		},
		Description: domain.Description{Instruction: instruction},
		Budget: domain.Budget{
			Steps:           1,
			Tokens:          8000,
			DurationSeconds: 60,
		},
		InteractionContract: &domain.InteractionContract{
			TerminationConditions: []string{"answer"},
			MultiTurn:             boolPtr(false),
		},
		Judgement: domain.Judgement{
			Combination:     "and",
			Primary:         0,
			ReferenceAnswer: reference,
			Scorers: []domain.ScorerSpec{
				{
					Type:    "contains",
					Params:  map[string]any{"needles": needleList},
					Weight:  1,
					Primary: true,
				},
			},
		},
		Meta: domain.TaskMeta{Author: "seed", CreatedAt: now, UpdatedAt: now},
	}
}
