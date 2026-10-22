// Package trace is the World Model log: append-only JSONL in the run's logs directory.
package trace

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Event is one claim, piece of evidence, refutation or verdict.
// EventID, the change key and the finding key are optional. Older lines omit them.
// HasEvidence is set on a refutation the skeptic made of a finding that has proof.
type Event struct {
	EventID     string      `json:"event_id,omitempty"`
	FindingID   string      `json:"finding_id"`
	Kind        string      `json:"kind"`
	Agent       string      `json:"agent"`
	Text        string      `json:"text"`
	Evidence    []Evidence  `json:"evidence,omitempty"`
	Refutation  *Refutation `json:"refutation,omitempty"`
	ChangeKey   string      `json:"change_key,omitempty"`
	RuleID      string      `json:"rule_id,omitempty"`
	Path        string      `json:"path,omitempty"`
	Title       string      `json:"title,omitempty"`
	HasEvidence bool        `json:"has_evidence,omitempty"`
}

// Evidence is one piece of proof stored with its finding.
type Evidence struct {
	Kind  string         `json:"kind"`
	Test  *TestEvidence  `json:"test,omitempty"`
	Query *QueryEvidence `json:"query,omitempty"`
}

// TestEvidence is the prover outcome. Bools are stored even when false.
type TestEvidence struct {
	Path          string `json:"path"`
	Name          string `json:"name"`
	Command       string `json:"command"`
	FailsOnChange bool   `json:"fails_on_change"`
	PassesOnBase  bool   `json:"passes_on_base"`
	Output        string `json:"output"`
}

// QueryEvidence is the investigator result, including every column and row.
type QueryEvidence struct {
	SQL      string     `json:"sql"`
	Engine   string     `json:"engine"`
	Snapshot string     `json:"snapshot"`
	Columns  []string   `json:"columns"`
	Rows     [][]string `json:"rows"`
}

// Refutation is the skeptic attempt. Bools are stored even when false.
type Refutation struct {
	Attempted bool   `json:"attempted"`
	Refuted   bool   `json:"refuted"`
	Reason    string `json:"reason"`
}

// Writer appends events to logs/trace.jsonl.
type Writer struct {
	path string
	f    *os.File
	mu   sync.Mutex
}

// Open creates the logs directory and opens the trace for appending.
func Open(dir string) (*Writer, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "trace.jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &Writer{path: path, f: f}, nil
}

// Path is the JSONL file.
func (w *Writer) Path() string { return w.path }

// Append writes one event and flushes it. It does not rewrite earlier lines.
func (w *Writer) Append(e Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if e.EventID == "" {
		e.EventID = eventID(e)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(e); err != nil {
		return err
	}
	if _, err := w.f.Write(buf.Bytes()); err != nil {
		return err
	}
	return w.f.Sync()
}

// eventID is the sha256 of the event with an empty id, so a replay of the
// same line names the same event. The id is not part of its own hash.
func eventID(e Event) string {
	e.EventID = ""
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(e); err != nil {
		return ""
	}
	sum := sha256.Sum256(bytes.TrimSpace(buf.Bytes()))
	return hex.EncodeToString(sum[:])[:16]
}
