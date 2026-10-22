// Package trace is the World Model log: append-only JSONL in the run's logs directory.
package trace

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Event is one claim, piece of evidence, refutation, verdict or guardrail check.
// RunID and Time are stamped by Append when the caller leaves them empty.
type Event struct {
	RunID      string      `json:"run_id"`
	Time       string      `json:"time"`
	FindingID  string      `json:"finding_id"`
	Kind       string      `json:"kind"`
	Agent      string      `json:"agent"`
	Text       string      `json:"text"`
	Evidence   []Evidence  `json:"evidence,omitempty"`
	Refutation *Refutation `json:"refutation,omitempty"`
	Guardrail  *Check      `json:"guardrail,omitempty"`
}

// Check is one guardrail verdict: the side, the verdict, the categories and the latency.
type Check struct {
	Side       string   `json:"side,omitempty"`
	Verdict    string   `json:"verdict,omitempty"`
	Categories []string `json:"categories,omitempty"`
	LatencyMS  int64    `json:"latency_ms"`
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
// One Open is one run: every event carries that run id.
type Writer struct {
	path  string
	f     *os.File
	runID string
	mu    sync.Mutex
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
	id, err := newRunID()
	if err != nil {
		f.Close()
		return nil, err
	}
	return &Writer{path: path, f: f, runID: id}, nil
}

func newRunID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// Path is the JSONL file.
func (w *Writer) Path() string { return w.path }

// Append writes one event and flushes it. It does not rewrite earlier lines.
// An empty run id and an empty time are filled: the run id from Open, the
// time as UTC in RFC 3339.
func (w *Writer) Append(e Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return fmt.Errorf("trace: closed")
	}
	if e.RunID == "" {
		e.RunID = w.runID
	}
	if e.Time == "" {
		e.Time = time.Now().UTC().Format(time.RFC3339)
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

// Close closes the trace file. A second call does nothing.
// Events already written stay on disk.
func (w *Writer) Close() error {
	if w == nil {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	return err
}
