// Package review is the seam between Proven's shell (command line, formats, Action) and its reviewer
// (Planner, Investigator, Prover, Skeptic, Judge, Writer). Agreed by the masters of both trains on 2026-10-21.
// Neither side changes this package without the other's agreement.
//
// Besides the seam's types, the package holds what the reviewer decides before and after the six agents run:
// it loads proven.yaml (config.go), chooses the review class from the changed paths (ADR-0012, ADR-0014),
// reads the author's model from the change (ADR-0011) and names the reviewer models and their independence
// from the author (ADR-0008, ADR-0009) (decide.go). An agreed, unproven claim is an Opinion beside the
// findings, not one of them (ADR-0041). Open wires the adapters and the pipeline (open.go).
// The six agents themselves live in internal/agent and run in internal/pipeline.
package review

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/store"
	"github.com/saphan-ai/proven/internal/telemetry"
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
	// liveStore, when set, replaces the evidence store. A review leaves it unset.
	liveStore store.Store
	// dial, when set, opens every connection the live adapters make in the air-gapped mode, after the
	// address is allowed. A review leaves it unset. Tests set it to record each dial without a socket.
	dial network.DialFunc
	// testRun, when set, replaces the container runner. A review leaves it unset,
	// so a test does not start a container.
	testRun testrun.Adapter
	// tel, when set, replaces the telemetry the OTEL_* variables build (ADR-0045).
	// A test sets it to read the spans and numbers back from memory.
	tel *telemetry.T
	// startLog, when set, receives the start line of local mode (the runtime's
	// absolute path and version). A review leaves it unset and the line goes to stderr.
	startLog io.Writer
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
	Findings []Finding // only findings with evidence (ADR-0002); unresolved ones are marked and counted
	// ProofCounts counts Findings by ProofLevel (ADR-0044 point 3). Every level is a key, with 0
	// when none; the counts add up to len(Findings). Opinions are not counted (ADR-0041).
	// Findings, unresolved ones included, keep their meaning.
	ProofCounts          map[ProofLevel]int
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
	GuardrailStatement   string           // set when the guardrail was switched off for this review; empty when it was on
	// Reviewers names each reviewer that took part, the default one first, with whether it is the
	// author's own model (ADR-0037 point 5). The shell renders from it, not from the statements.
	Reviewers []ReviewerInfo
	// Disagreements counts the findings whose Disagree is true (ADR-0037 point 4).
	Disagreements int
	// Opinions lists claims both reviewers reported at the same place that no
	// evidence proves (ADR-0041). They are not Findings, they are not in
	// ProofCounts, and they do not change the exit code. Empty when none.
	Opinions []Opinion
	// SkippedPaths lists tracked paths a live review did not copy because they
	// are symlinks, slash-separated and sorted. Empty when none were skipped.
	SkippedPaths []string
}

// The roles a reviewer takes in a review. The special role is the second reviewer of ADR-0037.
const (
	RoleDefault = "default"
	RoleSpecial = "special"
)

// ReviewerInfo is one reviewer of one review (ADR-0037 point 5).
type ReviewerInfo struct {
	Role   string // RoleDefault or RoleSpecial
	Model  string // the name the review gives it
	Family string // a configured family, or "family unknown"
	// IsAuthorModel is "yes" when the change named this model or a model of its family as its author,
	// "no" when it named authors and none match, and "unknown" when the change named none or a
	// family is unknown. It is never a guess: unknown stays unknown.
	IsAuthorModel string
}

// The verdicts one reviewer can give on a finding (ADR-0037 point 4). The strings are carried by
// JSON and SARIF, so a value never changes once released.
const (
	VerdictReported    = "reported"
	VerdictNotReported = "not reported"
	VerdictRefuted     = "refuted"
	VerdictNoAnswer    = "no answer"
)

// ReviewerVerdict is one reviewer's word on one finding: the verdict and the reviewer's own reason.
type ReviewerVerdict struct {
	Role    string // RoleDefault or RoleSpecial
	Model   string
	Verdict string // one of the Verdict constants
	Reason  string
}

// ProofLevel names how far a reader may trust one finding (ADR-0044). The string is what the
// JSON and SARIF formats carry, so a value never changes once released.
type ProofLevel string

// The four levels of proof, strongest first (ADR-0044 point 1).
const (
	// ProvedByTest: a test fails on the change and passes on the base (ADR-0002).
	ProvedByTest ProofLevel = "proved-by-test"
	// ProvedByTool: a deterministic tool's result differs between base and change (ADR-0040).
	// EvidenceTool is the kind that sets it, in ProofLevelOf.
	ProvedByTool ProofLevel = "proved-by-tool"
	// ProvedByData: a query and the rows it returned (ADR-0002).
	ProvedByData ProofLevel = "proved-by-data"
	// AgreedNotProven is the level name ADR-0044 reserved for an opinion.
	// ADR-0041 does not put that opinion on a finding: it is Result.Opinions,
	// and ProofCounts keeps this key at 0 so the totals stay the findings.
	AgreedNotProven ProofLevel = "agreed-by-two-models-not-proven"
)

// OpinionLevel is the level an opinion carries for JSON and SARIF (ADR-0041).
// It is note, so a machine does not read the opinion as a proven finding.
// Those writers are not in this build. The text section is what a reader gets,
// and a writer that lands later copies this value and does not put the opinion
// among the findings.
const OpinionLevel = "note"

// Opinion is one claim both reviewers reported at the same place that no evidence
// proves (ADR-0041). It is not a Finding. ProofCounts does not count it, and the
// exit code does not see it. Reasons holds both reviewers' own words, the default
// reviewer first.
type Opinion struct {
	ID       string
	Title    string
	Location Location
	// Level is OpinionLevel. It is not a proof level and not a severity.
	Level   string
	Reasons []ReviewerVerdict
}

// ProofLevels lists every level, strongest first. Counts and renderers walk this order.
var ProofLevels = []ProofLevel{ProvedByTest, ProvedByTool, ProvedByData, AgreedNotProven}

// Valid reports whether l is one of the four named levels.
func (l ProofLevel) Valid() bool {
	for _, known := range ProofLevels {
		if l == known {
			return true
		}
	}
	return false
}

// ProofLevelOf returns the level of the strongest evidence in ev, or an error when ev holds no
// evidence of a kind that has a level. ADR-0002: a finding without evidence cannot exist, so
// there is no level to fall back to, and a guess would show a reader proof that is not there.
func ProofLevelOf(ev []Evidence) (ProofLevel, error) {
	level := ProofLevel("")
	rank := len(ProofLevels)
	for _, e := range ev {
		var got ProofLevel
		switch e.Kind {
		case EvidenceTest:
			got = ProvedByTest
		case EvidenceTool:
			// ADR-0040: a tool's run sits between a test and a query in strength (ADR-0044).
			got = ProvedByTool
		case EvidenceQuery:
			got = ProvedByData
		default:
			continue
		}
		for i, l := range ProofLevels {
			if l == got && i < rank {
				rank, level = i, got
			}
		}
	}
	if level == "" {
		return "", errors.New("review: a finding has no evidence of a known kind, so it has no level of proof")
	}
	return level, nil
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
	ID       string // stable within one review
	RuleID   string // the kind of finding; SARIF ruleId
	Title    string
	Message  string
	Severity Severity
	Location Location
	Evidence []Evidence // at least one
	// ProofLevel is the finding's one named level of proof (ADR-0044). It is never empty in a Result.
	ProofLevel ProofLevel
	Unresolved bool // the evidence did not settle it; counts in findings-count and in unresolved-count
	// Models names the reviewer models that reported this finding.
	// One name when one run found it. Both names, the default reviewer first, when both did.
	Models []string
	// Verdicts holds each reviewer's verdict and reason, the default reviewer first. It is empty when
	// one reviewer ran. The verdicts are never combined (ADR-0037 point 4).
	Verdicts []ReviewerVerdict
	// Disagree is true when the reviewers who answered gave different verdicts. A reviewer that did not
	// answer (VerdictNoAnswer) is not a disagreement.
	Disagree bool
	// Fix is the fix Proven proposes for the finding, or nil when it proposed none (ADR-0042).
	// It is a suggestion for a person to apply; Proven never writes it to the change.
	Fix *Fix
}

// The kinds of Fix.
const (
	// FixPatch is a patch to the change, proved by the test that proved the finding.
	FixPatch = "patch"
	// FixDependency is a version of a package that leaves the advisory's affected range.
	FixDependency = "dependency"
)

// Fix is a proposed fix for one finding (ADR-0042).
// A patch is shown (Edits) only when Proved is true: it passed the proving test on the change with
// the patch applied, the test failed on the change and passed on the base, and the other tests of the
// package still passed. A patch that was not proved carries no Edits, and Note says the attempt failed.
// Runs is the record of the proof, with the command and exit code of each run, so a reader can run it
// again (point 7). A dependency fix has Dependency set and no Runs: no test proves a version change.
type Fix struct {
	Kind     string
	Proved   bool
	Attempts int
	// What the fix changes and Why that closes the problem, as the model explained it.
	What       string
	Why        string
	Edits      []FixEdit
	Runs       []FixRun
	Note       string
	Dependency *DependencyFix
}

// FixEdit is one edit of a proposed patch: the one place where Old stands in the file at Path is
// replaced by New. It is the whole patch in a form a person can apply by hand.
type FixEdit struct {
	Path string
	Old  string
	New  string
}

// FixRun is one run of the proof of a fix. Label names the tree it ran in (base, change,
// change+patch, or the package on change+patch) and Result how it ended: passed, failed,
// could not run or timed out. ExitCode is -1 when the run ended without one.
type FixRun struct {
	Label    string
	Command  string
	Result   string
	ExitCode int
	Output   string
}

// DependencyFix is the smallest version of Package that leaves the advisory's affected range (point 5).
// AdvisoryColumns and AdvisoryRow are the row of the OSV table whose fixed version To is, so a reader
// can check the choice against the table.
type DependencyFix struct {
	Ecosystem       string
	Package         string
	From            string
	To              string
	AdvisoryID      string
	AdvisoryColumns []string
	AdvisoryRow     []string
}

// Location is where a finding is.
type Location struct {
	Path      string
	StartLine int
	EndLine   int
}

// Evidence kinds. candidate.KindTest, candidate.KindQuery and candidate.KindTool
// carry the same strings.
const (
	EvidenceTest  = "test"
	EvidenceQuery = "query"
	// EvidenceTool is a deterministic tool's run compared between base and change (ADR-0040).
	EvidenceTool = "tool"
)

// Evidence is a test, a query with its rows (ADR-0002), or a tool's run (ADR-0040).
type Evidence struct {
	Kind  string         // EvidenceTest, EvidenceQuery or EvidenceTool
	Test  *TestEvidence  // set when Kind is EvidenceTest
	Query *QueryEvidence // set when Kind is EvidenceQuery
	Tool  *ToolEvidence  // set when Kind is EvidenceTool
}

// ToolEvidence is a deterministic tool's run on the base and the change (ADR-0040).
// PresentOnChange and PresentOnBase are both stored, so a reader can see a
// result the base already had and not mistake it for proof of the change.
type ToolEvidence struct {
	Tool            string
	Version         string
	Command         string
	Output          string // the change side, the output that differs
	BaseOutput      string
	PresentOnChange bool
	PresentOnBase   bool
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
	Form     string // "sql", or "mango" for CouchDB (ADR-0047)
	Engine   string // "zetaris", "local", "couchdb", or "local-sample"
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
