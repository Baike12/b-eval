package gate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EvidenceSchemaVersion is the version of the evidence record format.
//
// It is written into every record so a reader can tell which shape a record
// has without guessing.
const EvidenceSchemaVersion = "gate_evidence_v1"

// EvidenceRecord is one appended gate record.
//
// The record is append-only: a gate run never rewrites an earlier record,
// because a rewritten history cannot prove what was known when a decision
// was made.
type EvidenceRecord struct {
	SchemaVersion     string         `json:"schema_version"`
	GeneratedAt       string         `json:"generated_at"`
	RunID             string         `json:"run_id"`
	GateID            string         `json:"gate_id"`
	Status            Status         `json:"status"`
	Classification    Classification `json:"classification"`
	RecommendedAction string         `json:"recommended_action"`
	Description       string         `json:"description"`
	DurationMS        int64          `json:"duration_ms"`
	Detail            string         `json:"detail,omitempty"`
	Evidence          map[string]any `json:"evidence,omitempty"`
}

// Summary is the latest state of the gate, written for readers that do not
// want to replay the whole evidence file.
type Summary struct {
	SchemaVersion string         `json:"schema_version"`
	GeneratedAt   string         `json:"generated_at"`
	RunID         string         `json:"run_id"`
	Overall       Classification `json:"overall"`
	Passed        int            `json:"passed"`
	Failed        int            `json:"failed"`
	Skipped       int            `json:"skipped"`
	Results       []Result       `json:"results"`
}

// AppendEvidence appends the results of a gate run to the evidence file.
//
// The file is opened in append mode and each record is one line of JSON, so
// a concurrent reader always sees whole records.
func AppendEvidence(gateDir string, results []Result) (string, error) {
	if err := os.MkdirAll(gateDir, 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", gateDir, err)
	}
	path := filepath.Join(gateDir, "evidence.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	runID := time.Now().UTC().Format("20060102T150405Z")
	for _, r := range results {
		rec := EvidenceRecord{
			SchemaVersion:     EvidenceSchemaVersion,
			GeneratedAt:       time.Now().UTC().Format(time.RFC3339),
			RunID:             runID,
			GateID:            r.GateID,
			Status:            r.Status,
			Classification:    r.Classification,
			RecommendedAction: r.RecommendedAction,
			Description:       r.Description,
			DurationMS:        r.DurationMS,
			Detail:            r.Detail,
			Evidence:          r.Evidence,
		}
		line, err := json.Marshal(rec)
		if err != nil {
			return "", fmt.Errorf("marshal evidence: %w", err)
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			return "", fmt.Errorf("append evidence: %w", err)
		}
	}
	return path, nil
}

// WriteSummary writes the latest gate state.
func WriteSummary(gateDir string, results []Result) (string, error) {
	if err := os.MkdirAll(gateDir, 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", gateDir, err)
	}
	passed, failed, skipped := 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case StatusPassed:
			passed++
		case StatusFailed:
			failed++
		default:
			skipped++
		}
	}
	overall := ClassificationReady
	if failed > 0 {
		overall = ClassificationBlocked
	} else if skipped > 0 {
		overall = ClassificationInsufficientEvidence
	}
	summary := Summary{
		SchemaVersion: EvidenceSchemaVersion,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Overall:       overall,
		Passed:        passed,
		Failed:        failed,
		Skipped:       skipped,
		Results:       results,
	}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal summary: %w", err)
	}
	path := filepath.Join(gateDir, "latest.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	return path, nil
}
