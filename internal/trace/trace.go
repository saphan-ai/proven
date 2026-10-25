// Package trace is the World Model log: append-only JSONL in the run's logs directory.
//
// Every agent of the review writes here through the pipeline: the Planner's
// claims, the Investigator's and the Prover's evidence, the Skeptic's
// refutations, the Judge's verdicts, each guardrail check and each model call.
// ADR-0003 and ADR-0018 point 3: the trace the judges read is this stream, not
// text written beside it, and the world package folds the same lines into
// the review's state. A line is never rewritten, so the log shows what the
// agents did in the order they did it.
package trace

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/saphan-ai/proven/internal/redact"
)

// Event is one claim, piece of evidence, refutation, verdict or guardrail check.
// RunID and Time are stamped by Append when the caller leaves them empty.
// EventID, the change key and the finding key are optional. Older lines omit them.
// HasEvidence is set on a refutation the skeptic made of a finding that has proof.
// ChangeKey, RuleID, Path and Title name the finding across reviews; the
// world package uses them to remember a refutation in H-MEM (ADR-0006 point 4).
type Event struct {
	EventID     string      `json:"event_id,omitempty"`
	RunID       string      `json:"run_id"`
	Time        string      `json:"time"`
	FindingID   string      `json:"finding_id"`
	Kind        string      `json:"kind"`
	Agent       string      `json:"agent"`
	Text        string      `json:"text"`
	Evidence    []Evidence  `json:"evidence,omitempty"`
	Refutation  *Refutation `json:"refutation,omitempty"`
	Guardrail   *Check      `json:"guardrail,omitempty"`
	Model       *Model      `json:"model,omitempty"`
	ChangeKey   string      `json:"change_key,omitempty"`
	RuleID      string      `json:"rule_id,omitempty"`
	Path        string      `json:"path,omitempty"`
	Title       string      `json:"title,omitempty"`
	HasEvidence bool        `json:"has_evidence,omitempty"`
	// Run is "default" or "special": which review a claim, or a refutation, came from.
	Run string `json:"run,omitempty"`
}

// Model is one generator call as the stream named it: the route, the routed
// model and the Nightshift receipt id. FromReceipt is true only when the
// stream named a routed model.
type Model struct {
	Route       string `json:"route,omitempty"`
	Model       string `json:"model"`
	ReceiptID   string `json:"receipt_id,omitempty"`
	FromReceipt bool   `json:"from_receipt"`
}

// Check is one guardrail check: the side, the verdict, the categories and the
// latency, and the backend and model that answered. A check that could not be
// made has no verdict; Down then names why, as one value of the guardrail's
// closed set, so a count never has to read free text. Older lines omit
// Backend, Model and Down, and older down lines have no Check at all.
type Check struct {
	Side       string   `json:"side,omitempty"`
	Verdict    string   `json:"verdict,omitempty"`
	Categories []string `json:"categories,omitempty"`
	LatencyMS  int64    `json:"latency_ms"`
	// Backend is the adapter that answered: "server", "nvidia", or the name
	// of a sample or refusing adapter.
	Backend string `json:"backend,omitempty"`
	// Model is the configured model id; empty for an adapter that calls no model.
	Model string `json:"model,omitempty"`
	// Down is set only when the check could not be made.
	Down string `json:"down,omitempty"`
}

// Evidence is one piece of proof stored with its finding. Kind says which of
// Test, Query or Tool is set (ADR-0002: a test or a query; ADR-0040: a tool's run).
type Evidence struct {
	Kind  string         `json:"kind"`
	Test  *TestEvidence  `json:"test,omitempty"`
	Query *QueryEvidence `json:"query,omitempty"`
	Tool  *ToolEvidence  `json:"tool,omitempty"`
}

// ToolEvidence is a deterministic tool's run on both sides (ADR-0040).
// The bools are stored even when false, so a result the base already had
// is visible and not mistaken for a missing field.
type ToolEvidence struct {
	Tool            string `json:"tool"`
	Version         string `json:"version"`
	Command         string `json:"command"`
	Output          string `json:"output"`
	BaseOutput      string `json:"base_output"`
	PresentOnChange bool   `json:"present_on_change"`
	PresentOnBase   bool   `json:"present_on_base"`
}

// TestEvidence is the prover outcome. Bools are stored even when false, so a
// reader sees that a test did not fail rather than a missing field.
type TestEvidence struct {
	Path          string `json:"path"`
	Name          string `json:"name"`
	Command       string `json:"command"`
	FailsOnChange bool   `json:"fails_on_change"`
	PassesOnBase  bool   `json:"passes_on_base"`
	Output        string `json:"output"`
}

// QueryEvidence is the investigator result, including every column and row,
// so a judge can read the proof without rerunning the query.
// Form is the query language that ran (ADR-0047).
type QueryEvidence struct {
	SQL      string     `json:"sql"`
	Form     string     `json:"form,omitempty"`
	Engine   string     `json:"engine"`
	Snapshot string     `json:"snapshot"`
	Columns  []string   `json:"columns"`
	Rows     [][]string `json:"rows"`
}

// Refutation is the skeptic attempt. Bools are stored even when false, so a
// failed attempt is visible and not mistaken for no attempt.
type Refutation struct {
	Attempted bool   `json:"attempted"`
	Refuted   bool   `json:"refuted"`
	Reason    string `json:"reason"`
	// Evidenced is set only when the refutation carries evidence of its own: a
	// refuting test or contradicting query rows. Refuted alone is the model's
	// word, and a store never learns from it.
	Evidenced bool `json:"evidenced,omitempty"`
}

// Writer appends events to logs/trace.jsonl.
// One Open is one run: every event carries that run id. A mutex serialises
// Append, so concurrent agents never interleave two lines.
// A Writer with a Redactor replaces every configured secret in an event
// before the event is hashed and written.
type Writer struct {
	path     string
	f        *os.File
	runID    string
	redactor *redact.Redactor
	observer func(Event)
	mu       sync.Mutex
}

// SetObserver sets a function Append calls with each event after the event is
// redacted and written (ADR-0045: the events ride along as span events). The
// observer sees what the file holds and never a secret. It must not call
// Append, and it cannot fail the write.
func (w *Writer) SetObserver(f func(Event)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.observer = f
}

// Open creates the logs directory and opens the trace for appending.
func Open(dir string) (*Writer, error) {
	if err := privateDir(dir); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "trace.jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	// OpenFile sets the mode only when it creates; an existing file is narrowed too.
	if err := f.Chmod(0o600); err != nil {
		f.Close()
		return nil, err
	}
	id, err := newRunID()
	if err != nil {
		f.Close()
		return nil, err
	}
	return &Writer{path: path, f: f, runID: id}, nil
}

// newRunID returns 128 random bits in hex, so two runs never share an id.
func newRunID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// Path is the JSONL file.
func (w *Writer) Path() string { return w.path }

// SetRedactor sets the Redactor Append applies to every event from now on.
// The pipeline sets it once, before any agent writes, from the values the
// review's settings loaded.
func (w *Writer) SetRedactor(r *redact.Redactor) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.redactor = r
}

// Append writes one event and flushes it. It does not rewrite earlier lines.
// An empty run id and an empty time are filled: the run id from Open, the
// time as UTC in RFC 3339.
func (w *Writer) Append(e Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return fmt.Errorf("trace: closed")
	}
	// ADR-0005, ADR-0013 point 4: a secret never reaches the trace. The event is redacted before its id is
	// computed, so the id names the text that is written and not the text that was refused.
	e = redactEvent(e, w.redactor)
	if e.RunID == "" {
		e.RunID = w.runID
	}
	if e.Time == "" {
		e.Time = time.Now().UTC().Format(time.RFC3339)
	}
	if e.EventID == "" {
		e.EventID = eventID(e)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// Code and SQL in an event keep their <, > and & as written.
	enc.SetEscapeHTML(false)
	if err := enc.Encode(e); err != nil {
		return err
	}
	if _, err := w.f.Write(buf.Bytes()); err != nil {
		return err
	}
	// Sync after each line: a review that dies still leaves every event it wrote.
	if err := w.f.Sync(); err != nil {
		return err
	}
	if w.observer != nil {
		w.observer(e)
	}
	return nil
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

// eventID is the first 16 hex digits of the sha256 of the event with an empty
// id, run id and time, so the same claim names the same event in every run and
// on every replay. The id is not part of its own hash. The world package skips
// an id it has already applied, which makes its ingest idempotent (ADR-0018 point 5).
func eventID(e Event) string {
	e.EventID = ""
	e.RunID = ""
	e.Time = ""
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(e); err != nil {
		return ""
	}
	sum := sha256.Sum256(bytes.TrimSpace(buf.Bytes()))
	return hex.EncodeToString(sum[:])[:16]
}

// privateDir creates dir, or narrows an existing one, to 0700: MkdirAll sets a
// mode only when it creates, and a directory an earlier run left at 0755 must
// not stay readable by others (Proven's own state is readable only by its user).
func privateDir(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.Chmod(dir, 0o700)
}
