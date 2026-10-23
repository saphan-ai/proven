// Package review is the seam between Proven's shell (command line, formats, Action) and its reviewer
// (Planner, Investigator, Prover, Skeptic, Judge, Writer). Agreed by the masters of both trains on 2026-10-21.
// Neither side changes this package without the other's agreement.
//
// Besides the seam's types, the package holds what the reviewer decides before and after the six agents run:
// it loads proven.yaml (config.go), chooses the review class from the changed paths (ADR-0012, ADR-0014),
// reads the author's model from the change (ADR-0011) and names the reviewer models and their independence
// from the author (ADR-0008, ADR-0009) (decide.go). Open wires the adapters and the pipeline (open.go).
// The six agents themselves live in internal/agent and run in internal/pipeline.
package review

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/testrun"
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

// Error returns one line the shell prints as it is: the refused field and value, the reason,
// and the allowed values, so a user can fix the setting without reading the code.
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
	// gen, when set, replaces the live generator. A review leaves it unset.
	// Tests set it so a live review makes no network call.
	gen generator.Generator
	// liveQuery, when set, replaces the live query adapter. A review leaves it unset.
	liveQuery query.Adapter
	// testRun, when set, replaces the container runner. A review leaves it unset,
	// so a test does not start a container.
	testRun testrun.Adapter
}

// Request is one review request.
type Request struct {
	Target         string   // repository root or directory
	BaselineCommit string   // "" means the whole target; in ci, the merge base
	HeadCommit     string   // in ci, GITHUB_SHA; "" means the working tree
	Include        []string // globs over changed paths
	Exclude        []string
	MinSeverity    Severity // the shell filters the output by it
	Description    string   // pull-request description; empty means the request carries none
}

// Severity is INFO, WARNING or ERROR (SARIF note, warning, error).
type Severity string

// The three severities, lowest first. The planner's severity string is copied into a Finding as one of these.
const (
	SeverityInfo    Severity = "INFO"
	SeverityWarning Severity = "WARNING"
	SeverityError   Severity = "ERROR"
)

// Result is what a review produced.
type Result struct {
	Findings             []Finding        // only findings with evidence (ADR-0002); unresolved ones are marked and counted
	FailedServices       []ServiceFailure // the review degraded; the shell names each and still exits 0 or 1
	Models               []ModelUse       // every model the review used, from the provider's receipt (ADR-0019)
	Partial              bool             // the deadline was reached
	TracePath            string           // the JSONL trace in LogDir
	Class                string           // review class the path rules chose (ADR-0012, ADR-0014)
	ClassRule            string           // path pattern that chose Class, or "fallback"
	Doubled              bool             // Class is one of the doubled classes
	DefaultReviewer      string           // configured default reviewer
	SpecialReviewer      string           // configured special-review model; empty means not configured
	ReviewerModels       []string         // models this review uses, the default reviewer first
	ReviewerStatements   []string         // one line per reviewer, plus a line when the special model is not configured
	AuthorModels         []AuthorModel    // models the change named; empty means none were stated
	AuthorModelStatement string           // "author model not stated" when the change names no model
	SettingsStatement    string           // set when the change edits proven.yaml; empty otherwise
	// SkippedPaths lists tracked paths a live review did not copy because they
	// are symlinks, slash-separated and sorted. Empty when none were skipped.
	SkippedPaths []string
}

// AuthorModel is one model the change named (ADR-0011).
type AuthorModel struct {
	Agent  string // the part before the first colon
	Model  string // the part after the first colon
	Family string // a configured family, or "family unknown"
	Source string // "commit", "description" or "record"
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
	// Models names the reviewer models that reported this finding.
	// One name when one run found it. Both names, the default reviewer first, when both did.
	Models []string
}

// Location is where a finding is.
type Location struct {
	Path      string
	StartLine int
	EndLine   int
}

// Evidence kinds. candidate.KindTest and candidate.KindQuery carry the same strings.
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
	Route       string // winning_route.provider when the stream named it; "" when unknown
	Model       string // routed model; "unknown" when the stream named none, never a guess
	ReceiptID   string // meterless.receipt_id when the stream named it; "" otherwise
	FromReceipt bool   // true only when the stream named a routed model
}
