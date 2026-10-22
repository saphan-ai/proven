// Package review is the seam between Proven's shell (command line, formats, Action) and its reviewer
// (Planner, Investigator, Prover, Skeptic, Judge, Writer). Agreed by the masters of both trains on 2026-10-21.
// Neither side changes this package without the other's agreement.
package review

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Reviewer reviews one change.
type Reviewer interface {
	// Review returns a Result for every review that ran, also a degraded or partial one.
	// A non-nil error means Proven could not review at all: the shell exits 2.
	// The shell sets the deadline from --timeout on ctx; at the deadline the reviewer returns what it
	// proved so far with Partial set and a nil error.
	Review(ctx context.Context, req Request) (Result, error)
}

// ErrNotImplemented is the error Open returned before the reviewer existed.
// Open no longer returns it. The name stays so a caller can still name it.
var ErrNotImplemented = errors.New("review: not implemented")

// Open loads the reviewer's own settings and returns a Reviewer.
// A *ConfigError means the configuration was refused on load: the shell exits 7 and prints its Error().
func Open(opts Options) (Reviewer, error) { return open(opts) }

// ConfigError names a refused setting and the values that are allowed (ADR-0013).
type ConfigError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

func (e *ConfigError) Error() string {
	msg := fmt.Sprintf("configuration refused: %s=%q", e.Field, e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}

// Options are what the shell passes to Open.
type Options struct {
	ConfigPath string              // --config / PROVEN_CONFIG; "" means the defaults
	Sample     bool                // --sample / SAMPLE_MODE=true: every adapter is its sample implementation (ADR-0001)
	LogDir     string              // the run's logs/ directory; the reviewer writes its JSONL trace here (ADR-0018)
	Env        func(string) string // environment lookup; keys come from the environment only (ADR-0001)
}

// Request is one review request.
type Request struct {
	Target         string   // repository root or directory
	BaselineCommit string   // "" means the whole target; in ci, the merge base
	HeadCommit     string   // in ci, GITHUB_SHA; "" means the working tree
	Include        []string // globs over changed paths
	Exclude        []string
	MinSeverity    Severity // the shell filters the output by it
}

// Severity is INFO, WARNING or ERROR (SARIF note, warning, error).
type Severity string

const (
	SeverityInfo    Severity = "INFO"
	SeverityWarning Severity = "WARNING"
	SeverityError   Severity = "ERROR"
)

// Result is what a review produced.
type Result struct {
	Findings       []Finding        // only findings with evidence (ADR-0002); unresolved ones are marked and counted
	FailedServices []ServiceFailure // the review degraded; the shell names each and still exits 0 or 1
	Models         []ModelUse       // every model the review used, from the provider's receipt (ADR-0019)
	Partial        bool             // the deadline was reached
	TracePath      string           // the JSONL trace in LogDir
}

// Finding is one reported problem with its evidence.
type Finding struct {
	ID         string // stable within one review
	RuleID     string // the kind of finding; SARIF ruleId
	Title      string
	Message    string
	Severity   Severity
	Location   Location
	Evidence   []Evidence // at least one
	Unresolved bool       // the evidence did not settle it; counts in findings-count and in unresolved-count
}

// Location is where a finding is.
type Location struct {
	Path      string
	StartLine int
	EndLine   int
}

// Evidence kinds.
const (
	EvidenceTest  = "test"
	EvidenceQuery = "query"
)

// Evidence is a test or a query with its rows (ADR-0002).
type Evidence struct {
	Kind  string         // EvidenceTest or EvidenceQuery
	Test  *TestEvidence  // set when Kind is EvidenceTest
	Query *QueryEvidence // set when Kind is EvidenceQuery
}

// TestEvidence is a test that fails on the change and passes on the base.
type TestEvidence struct {
	Path          string
	Name          string
	Command       string
	FailsOnChange bool
	PassesOnBase  bool
	Output        string // trimmed
}

// QueryEvidence is a query and the rows it returned.
type QueryEvidence struct {
	SQL      string
	Engine   string // "zetaris" or "local-sample"
	Snapshot string // which snapshot of each source was read
	Columns  []string
	Rows     [][]string
}

// ServiceFailure names an outside service that failed and why.
type ServiceFailure struct {
	Service string // for example "nightshift", "zetaris", "guardrail", "prover"
	Reason  string
}

// ModelUse names a model one agent used, as the provider's receipt names it.
type ModelUse struct {
	Agent       string // "planner", "investigator", "prover", "skeptic", "judge" or "writer"
	Route       string // from the receipt; "" when unknown
	Model       string // from the receipt; "unknown" when the receipt had none, never a guess
	FromReceipt bool
}
