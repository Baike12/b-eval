// Package store persists task sets, agents and runs.
//
// The store keeps one JSON file per record under a data directory. The
// format is deliberately inspectable: a run must be readable by a third
// party without the platform, because the audit package promises exactly
// that.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Baike12/b-eval/internal/domain"
)

// Store is a file-backed record store.
type Store struct {
	root string
	mu   sync.RWMutex
}

// New returns a store rooted at the given directory.
func New(root string) *Store {
	return &Store{root: root}
}

// Root returns the store's data directory.
func (s *Store) Root() string { return s.root }

// SaveTaskSet writes a task set.
func (s *Store) SaveTaskSet(ts domain.TaskSet) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeJSON(filepath.Join(s.root, "tasksets"), taskSetFile(ts), ts)
}

// TaskSet loads a task set by ID and version.
func (s *Store) TaskSet(id, version string) (domain.TaskSet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var ts domain.TaskSet
	if err := s.readJSON(filepath.Join(s.root, "tasksets"), taskSetFileName(id, version), &ts); err != nil {
		return ts, err
	}
	return ts, nil
}

// TaskSets lists every stored task set, newest first.
func (s *Store) TaskSets() ([]domain.TaskSet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.TaskSet
	err := s.each(filepath.Join(s.root, "tasksets"), ".json", func(path string) error {
		var ts domain.TaskSet
		if err := readJSONFile(path, &ts); err != nil {
			return err
		}
		out = append(out, ts)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt > out[j].CreatedAt
		}
		return out[i].TaskSetID < out[j].TaskSetID
	})
	return out, nil
}

// SaveAgent writes an agent registration.
func (s *Store) SaveAgent(a domain.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeJSON(filepath.Join(s.root, "agents"), a.AgentID+".json", a)
}

// Agents lists the registered agents.
func (s *Store) Agents() ([]domain.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Agent
	err := s.each(filepath.Join(s.root, "agents"), ".json", func(path string) error {
		var a domain.Agent
		if err := readJSONFile(path, &a); err != nil {
			return err
		}
		out = append(out, a)
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].AgentID < out[j].AgentID })
	return out, err
}

// Agent loads one agent registration.
func (s *Store) Agent(id string) (domain.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var a domain.Agent
	if err := s.readJSON(filepath.Join(s.root, "agents"), id+".json", &a); err != nil {
		return a, err
	}
	return a, nil
}

// SaveRun writes a run.
func (s *Store) SaveRun(run domain.Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeJSON(filepath.Join(s.root, "runs"), run.RunID+".json", run)
}

// Run loads one run.
func (s *Store) Run(id string) (domain.Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var run domain.Run
	if err := s.readJSON(filepath.Join(s.root, "runs"), id+".json", &run); err != nil {
		return run, err
	}
	return run, nil
}

// Runs lists the stored runs, newest first.
func (s *Store) Runs() ([]domain.Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Run
	err := s.each(filepath.Join(s.root, "runs"), ".json", func(path string) error {
		var run domain.Run
		if err := readJSONFile(path, &run); err != nil {
			return err
		}
		out = append(out, run)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].StartedAt != out[j].StartedAt {
			return out[i].StartedAt > out[j].StartedAt
		}
		return out[i].RunID < out[j].RunID
	})
	return out, nil
}

// writeJSON writes a record atomically.
//
// The write goes to a temporary file and is renamed into place, so a crash
// mid-write cannot leave a half-written record that fails to parse on the
// next start.
func (s *Store) writeJSON(dir, name string, value any) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", name, err)
	}
	target := filepath.Join(dir, name)
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, target); err != nil {
		return fmt.Errorf("rename %s: %w", tmp, err)
	}
	return nil
}

func (s *Store) readJSON(dir, name string, out any) error {
	return readJSONFile(filepath.Join(dir, name), out)
}

// each calls fn for every file in dir with the given suffix.
func (s *Store) each(dir, suffix string, fn func(path string) error) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), suffix) {
			continue
		}
		if err := fn(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func readJSONFile(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

// taskSetFile returns the file name of a task set.
func taskSetFile(ts domain.TaskSet) string { return taskSetFileName(ts.TaskSetID, ts.Version) }

func taskSetFileName(id, version string) string {
	return id + "__" + version + ".json"
}
