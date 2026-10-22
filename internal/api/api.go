// Package api exposes the platform over HTTP.
//
// The API is the seam the frontend and the end-to-end smoke run both use, so
// anything the frontend can do is reachable from a script.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Baike12/b-eval/internal/adapter"
	"github.com/Baike12/b-eval/internal/domain"
	"github.com/Baike12/b-eval/internal/metrics"
	"github.com/Baike12/b-eval/internal/report"
	"github.com/Baike12/b-eval/internal/runner"
	"github.com/Baike12/b-eval/internal/scoring"
	"github.com/Baike12/b-eval/internal/store"
)

// Server holds the platform's collaborators.
type Server struct {
	Store   *store.Store
	Agents  *adapter.Registry
	Scorers *scoring.Registry
	Metrics *metrics.Registry
	Runner  *runner.Runner
}

// NewServer builds a server over the given store.
func NewServer(st *store.Store) *Server {
	metricRegistry := metrics.NewRegistry()
	return &Server{
		Store:   st,
		Agents:  adapter.NewRegistry(),
		Scorers: scoring.NewRegistry(),
		Metrics: metricRegistry,
		Runner: &runner.Runner{
			Scorers: scoring.NewRegistry(),
			Metrics: metricRegistry,
		},
	}
}

// Handler returns the HTTP handler of the API.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", s.handleHealth)
	mux.HandleFunc("GET /api/tasksets", s.handleListTaskSets)
	mux.HandleFunc("POST /api/tasksets", s.handleCreateTaskSet)
	mux.HandleFunc("GET /api/tasksets/{id}/{version}", s.handleGetTaskSet)
	mux.HandleFunc("POST /api/tasksets/{id}/{version}/tasks", s.handleAddTask)
	mux.HandleFunc("GET /api/agents", s.handleListAgents)
	mux.HandleFunc("POST /api/agents", s.handleCreateAgent)
	mux.HandleFunc("POST /api/agents/{id}/selfcheck", s.handleSelfCheck)
	mux.HandleFunc("GET /api/metrics", s.handleListMetrics)
	mux.HandleFunc("POST /api/metrics", s.handleRegisterMetric)
	mux.HandleFunc("GET /api/scorers", s.handleListScorers)
	mux.HandleFunc("GET /api/runs", s.handleListRuns)
	mux.HandleFunc("POST /api/runs", s.handleCreateRun)
	mux.HandleFunc("GET /api/runs/{id}", s.handleGetRun)
	mux.HandleFunc("GET /api/runs/{id}/report", s.handleGetReport)
	mux.HandleFunc("GET /api/runs/{id}/trials/{trialID}", s.handleGetTrial)
	mux.HandleFunc("GET /api/compare", s.handleCompare)

	return withJSONErrors(mux)
}

// errorResponse is the shape of every error the API returns.
type errorResponse struct {
	Error string `json:"error"`
	// Field carries the offending field when the error is a validation
	// failure, so a client can point at it.
	Field string `json:"field,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

// writeRecordError answers a request for a record the store could not give.
//
// The store's error names the file, and the file's path is the deployment's
// own layout rather than the reader's business, so the response states what
// was looked for and the internal detail goes to the server's log.
func writeRecordError(w http.ResponseWriter, err error, subject string) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, fmt.Errorf("%s does not exist", subject))
		return
	}
	log.Printf("reading %s: %v", subject, err)
	writeError(w, http.StatusInternalServerError, fmt.Errorf("%s could not be read", subject))
}

// taskSetSubject names the task set a request asked for.
func taskSetSubject(r *http.Request) string {
	return fmt.Sprintf("task set %q version %q", r.PathValue("id"), r.PathValue("version"))
}

// runSubject names the run a request asked for.
func runSubject(r *http.Request) string {
	return fmt.Sprintf("run %q", r.PathValue("id"))
}

// withJSONErrors turns an unhandled panic into a JSON error rather than an
// empty connection, so a client never has to guess what went wrong.
func withJSONErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				writeError(w, http.StatusInternalServerError, fmt.Errorf("internal error: %v", rec))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"time":      domain.Now(),
		"task_sets": s.countTaskSets(),
	})
}

func (s *Server) countTaskSets() int {
	sets, err := s.Store.TaskSets()
	if err != nil {
		return 0
	}
	return len(sets)
}

// --- task sets ---

func (s *Server) handleListTaskSets(w http.ResponseWriter, r *http.Request) {
	sets, err := s.Store.TaskSets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, sets)
}

// createTaskSetRequest is the body of a task-set creation.
type createTaskSetRequest struct {
	TaskSetID   string        `json:"task_set_id"`
	Version     string        `json:"version"`
	Benchmark   string        `json:"benchmark"`
	Description string        `json:"description"`
	Tasks       []domain.Task `json:"tasks"`
}

func (s *Server) handleCreateTaskSet(w http.ResponseWriter, r *http.Request) {
	var req createTaskSetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(req.TaskSetID) == "" {
		writeError(w, http.StatusBadRequest, errors.New("task_set_id is required"))
		return
	}
	if strings.TrimSpace(req.Version) == "" {
		req.Version = "v1"
	}
	if strings.TrimSpace(req.Benchmark) == "" {
		// The benchmark is part of the task's identity. A task set that
		// does not name one belongs to a benchmark of its own name, which is
		// what the task set was called in the first place.
		req.Benchmark = req.TaskSetID
	}
	now := domain.Now()
	ts := domain.TaskSet{
		TaskSetID:   req.TaskSetID,
		Version:     req.Version,
		Benchmark:   req.Benchmark,
		Description: req.Description,
		CreatedAt:   now,
	}
	for i := range req.Tasks {
		task := req.Tasks[i]
		task.Benchmark = req.Benchmark
		if task.TaskSetVersion == "" {
			task.TaskSetVersion = req.Version
		}
		if task.Meta.CreatedAt == "" {
			task.Meta.CreatedAt = now
		}
		task.Meta.UpdatedAt = now
		if errs := domain.ValidateTask(task); len(errs) > 0 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"error":      fmt.Sprintf("task %q does not satisfy its form contract", task.TaskID),
				"violations": violationsOf(errs),
			})
			return
		}
		ts.Tasks = append(ts.Tasks, task)
	}
	if err := s.Store.SaveTaskSet(ts); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, ts)
}

func violationsOf(errs []domain.ValidationError) []map[string]string {
	out := make([]map[string]string, 0, len(errs))
	for _, e := range errs {
		out = append(out, map[string]string{"field": e.Field, "message": e.Message})
	}
	return out
}

func (s *Server) handleGetTaskSet(w http.ResponseWriter, r *http.Request) {
	ts, err := s.Store.TaskSet(r.PathValue("id"), r.PathValue("version"))
	if err != nil {
		writeRecordError(w, err, taskSetSubject(r))
		return
	}
	writeJSON(w, http.StatusOK, ts)
}

func (s *Server) handleAddTask(w http.ResponseWriter, r *http.Request) {
	ts, err := s.Store.TaskSet(r.PathValue("id"), r.PathValue("version"))
	if err != nil {
		writeRecordError(w, err, taskSetSubject(r))
		return
	}
	var task domain.Task
	if err := decodeJSON(r, &task); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	task.Benchmark = ts.Benchmark
	if task.TaskSetVersion == "" {
		task.TaskSetVersion = ts.Version
	}
	now := domain.Now()
	if task.Meta.CreatedAt == "" {
		task.Meta.CreatedAt = now
	}
	task.Meta.UpdatedAt = now
	if errs := domain.ValidateTask(task); len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":      fmt.Sprintf("task %q does not satisfy its form contract", task.TaskID),
			"violations": violationsOf(errs),
		})
		return
	}
	for _, existing := range ts.Tasks {
		if existing.TaskID == task.TaskID {
			writeError(w, http.StatusConflict, fmt.Errorf("task %q already exists in the task set", task.TaskID))
			return
		}
	}
	ts.Tasks = append(ts.Tasks, task)
	if err := s.Store.SaveTaskSet(ts); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

// --- agents ---

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := s.Store.Agents()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, agents)
}

// createAgentRequest is the body of an agent registration.
type createAgentRequest struct {
	AgentID  string            `json:"agent_id"`
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Shape    string            `json:"shape"`
	Mode     string            `json:"mode"`
	Endpoint string            `json:"endpoint"`
	Options  map[string]any    `json:"options"`
	Answers  map[string]string `json:"answers"`
}

func (s *Server) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	var req createAgentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(req.AgentID) == "" {
		writeError(w, http.StatusBadRequest, errors.New("agent_id is required"))
		return
	}
	if req.Shape == "" {
		req.Shape = "mock"
	}
	agent := domain.Agent{
		AgentID:   req.AgentID,
		Name:      req.Name,
		Version:   req.Version,
		Shape:     req.Shape,
		Endpoint:  req.Endpoint,
		Mode:      req.Mode,
		Options:   req.Options,
		CreatedAt: domain.Now(),
	}
	impl, err := s.buildAdapter(agent, req.Answers)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	s.Agents.RegisterAgent(agent, impl)
	if err := s.Store.SaveAgent(agent); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, agent)
}

// buildAdapter creates the adapter implementation of a registered agent.
func (s *Server) buildAdapter(agent domain.Agent, answers map[string]string) (adapter.Adapter, error) {
	switch agent.Shape {
	case "mock":
		mock := adapter.NewMockAdapter("")
		mock.Answers = answers
		return mock, nil
	case "http":
		if agent.Endpoint == "" {
			return nil, errors.New("an http agent requires an endpoint")
		}
		mode := agent.Mode
		if mode == "" {
			mode = "single"
		}
		return adapter.NewHTTPAdapter(agent.Endpoint, mode, nil), nil
	}
	return nil, fmt.Errorf("unknown agent shape %q; expected %q or %q", agent.Shape, "mock", "http")
}

func (s *Server) handleSelfCheck(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	impl, ok := s.Agents.Adapter(id)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Errorf("agent %q is not registered", id))
		return
	}
	tasks, err := s.smokeTasks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	result := adapter.SelfCheck{}.Run(r.Context(), impl, tasks)
	agent, _ := s.Agents.Agent(id)
	agent.SelfCheck = &result
	s.Agents.RegisterAgent(agent, impl)
	_ = s.Store.SaveAgent(agent)
	writeJSON(w, http.StatusOK, result)
}

// smokeTasks returns the task set the self-check runs against.
//
// The self-check runs against the smallest available task set, because its
// job is to detect integration bugs, not to measure capability.
func (s *Server) smokeTasks() ([]domain.Task, error) {
	sets, err := s.Store.TaskSets()
	if err != nil {
		return nil, err
	}
	if len(sets) == 0 {
		return nil, errors.New("no task set is stored")
	}
	sort.Slice(sets, func(i, j int) bool {
		if sets[i].TaskSetID == "smoke" {
			return true
		}
		if sets[j].TaskSetID == "smoke" {
			return false
		}
		return len(sets[i].Tasks) < len(sets[j].Tasks)
	})
	tasks := sets[0].Tasks
	if len(tasks) > 10 {
		tasks = tasks[:10]
	}
	return tasks, nil
}

// --- metrics ---

func (s *Server) handleListMetrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"metrics":      s.Metrics.List(),
		"custom_kinds": metrics.CustomKinds(),
	})
}

// registerMetricRequest is the body of a metric registration.
type registerMetricRequest struct {
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	Kind               string         `json:"kind"`
	Unit               string         `json:"unit"`
	Aggregations       []string       `json:"aggregations"`
	DefaultAggregation string         `json:"default_aggregation"`
	Numerator          string         `json:"numerator"`
	Denominator        string         `json:"denominator"`
	Formula            string         `json:"formula"`
	CIMethod           string         `json:"ci_method"`
	Params             map[string]any `json:"params"`
}

func (s *Server) handleRegisterMetric(w http.ResponseWriter, r *http.Request) {
	var req registerMetricRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Unit == "" {
		req.Unit = string(metrics.UnitRatio)
	}
	if len(req.Aggregations) == 0 {
		req.Aggregations = []string{metrics.AggMacro, metrics.AggMicro}
	}
	if req.CIMethod == "" {
		req.CIMethod = "wilson"
	}
	def, err := metrics.BuildCustom(metrics.CustomRequest{
		ID:                 req.ID,
		Name:               req.Name,
		Description:        req.Description,
		Kind:               req.Kind,
		Unit:               metrics.Unit(req.Unit),
		Aggregations:       req.Aggregations,
		DefaultAggregation: req.DefaultAggregation,
		Numerator:          req.Numerator,
		Denominator:        req.Denominator,
		Formula:            req.Formula,
		CIMethod:           req.CIMethod,
		Params:             req.Params,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.Metrics.Register(def); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":  err.Error(),
			"metric": def.ID,
		})
		return
	}
	writeJSON(w, http.StatusCreated, def)
}

func (s *Server) handleListScorers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Scorers.Types())
}

// --- runs ---

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.Store.Runs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

// createRunRequest is the body of a run creation.
type createRunRequest struct {
	PlanID         string  `json:"plan_id"`
	TaskSetID      string  `json:"task_set_id"`
	TaskSetVersion string  `json:"task_set_version"`
	AgentID        string  `json:"agent_id"`
	Repetitions    int     `json:"repetitions"`
	Seeds          []int64 `json:"seeds"`
	Aggregation    string  `json:"aggregation"`
	CIMethod       string  `json:"ci_method"`
	Concurrency    int     `json:"concurrency"`
}

func (s *Server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	var req createRunRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.TaskSetID == "" || req.AgentID == "" {
		writeError(w, http.StatusBadRequest, errors.New("task_set_id and agent_id are required"))
		return
	}
	if req.TaskSetVersion == "" {
		req.TaskSetVersion = "v1"
	}
	if req.Repetitions <= 0 {
		req.Repetitions = 1
	}
	if req.Aggregation == "" {
		req.Aggregation = metrics.AggMacro
	}
	if req.CIMethod == "" {
		req.CIMethod = "wilson"
	}
	if req.PlanID == "" {
		req.PlanID = fmt.Sprintf("plan-%s-%s", req.TaskSetID, req.AgentID)
	}

	ts, err := s.Store.TaskSet(req.TaskSetID, req.TaskSetVersion)
	if err != nil {
		writeRecordError(w, err, fmt.Sprintf("task set %q version %q", req.TaskSetID, req.TaskSetVersion))
		return
	}
	impl, ok := s.Agents.Adapter(req.AgentID)
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Errorf("agent %q is not registered", req.AgentID))
		return
	}
	agent, _ := s.Agents.Agent(req.AgentID)
	if agent.SelfCheck == nil {
		writeError(w, http.StatusConflict, fmt.Errorf("agent %q has not passed the self-check; run the self-check first", req.AgentID))
		return
	}
	if agent.SelfCheck.Status != "passed" {
		writeError(w, http.StatusConflict, fmt.Errorf("agent %q did not pass the self-check: %s", req.AgentID, strings.Join(agent.SelfCheck.InvalidReasons, "; ")))
		return
	}

	plan := domain.Plan{
		PlanID:         req.PlanID,
		TaskSetID:      req.TaskSetID,
		TaskSetVersion: req.TaskSetVersion,
		AgentID:        req.AgentID,
		Repetitions:    req.Repetitions,
		Seeds:          req.Seeds,
		Aggregation:    req.Aggregation,
		CIMethod:       req.CIMethod,
		Concurrency:    req.Concurrency,
		CreatedAt:      domain.Now(),
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()
	run, err := s.Runner.Run(ctx, runner.Input{
		Plan:    plan,
		TaskSet: ts,
		Adapter: impl,
		AgentID: req.AgentID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.Store.SaveRun(run); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, run)
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	run, err := s.Store.Run(r.PathValue("id"))
	if err != nil {
		writeRecordError(w, err, runSubject(r))
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	run, err := s.Store.Run(r.PathValue("id"))
	if err != nil {
		writeRecordError(w, err, runSubject(r))
		return
	}
	rep := s.buildReport(run)
	writeJSON(w, http.StatusOK, rep)
}

func (s *Server) handleGetTrial(w http.ResponseWriter, r *http.Request) {
	run, err := s.Store.Run(r.PathValue("id"))
	if err != nil {
		writeRecordError(w, err, runSubject(r))
		return
	}
	trialID := r.PathValue("trialID")
	for _, tr := range run.Trajectories {
		if tr.TrialID == trialID {
			writeJSON(w, http.StatusOK, map[string]any{"trajectory": tr})
			return
		}
	}
	writeError(w, http.StatusNotFound, fmt.Errorf("trial %q is not in run %q", trialID, run.RunID))
}

func (s *Server) handleCompare(w http.ResponseWriter, r *http.Request) {
	aID := r.URL.Query().Get("a")
	bID := r.URL.Query().Get("b")
	if aID == "" || bID == "" {
		writeError(w, http.StatusBadRequest, errors.New("the a and b query parameters are required"))
		return
	}
	a, err := s.Store.Run(aID)
	if err != nil {
		writeRecordError(w, err, fmt.Sprintf("run %q", aID))
		return
	}
	b, err := s.Store.Run(bID)
	if err != nil {
		writeRecordError(w, err, fmt.Sprintf("run %q", bID))
		return
	}
	repA := s.buildReport(a)
	repB := s.buildReport(b)
	writeJSON(w, http.StatusOK, map[string]any{
		"a": repA,
		"b": repB,
		"difference": map[string]any{
			"total_score": repB.MainScore.TotalScore - repA.MainScore.TotalScore,
			"a_total":     repA.MainScore.TotalScore,
			"b_total":     repB.MainScore.TotalScore,
		},
	})
}

// buildReport assembles the report of a run.
func (s *Server) buildReport(run domain.Run) report.Report {
	names := map[string]string{}
	if ts, err := s.Store.TaskSet(run.Plan.TaskSetID, run.Plan.TaskSetVersion); err == nil {
		for _, t := range ts.Tasks {
			names[t.TaskID] = t.DisplayName
		}
	}
	return report.Build(run, s.Metrics, names, report.Options{
		K:                run.Plan.Repetitions,
		Aggregation:      run.Plan.Aggregation,
		CIMethod:         run.Plan.CIMethod,
		PrimaryCriterion: "The task's primary criterion is satisfied.",
	})
}

// decodeJSON reads a JSON request body with a size limit.
func decodeJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 8<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	return nil
}
