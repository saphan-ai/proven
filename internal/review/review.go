// Package review loads configuration and wires the change reviewer.
// Model hypotheses, controlled evidence, deterministic adjudication and report
// assembly remain separate packages. Open selects the configured adapters.
package review

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/report"
	"github.com/saphan-ai/proven/internal/service"
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

// ConfigError names a refused setting and the values that are allowed .
// Error returns one line the shell prints as it is: the refused field and value, the reason,
// and the allowed values, so a user can fix the setting without reading the code.
type ConfigError = service.ConfigError

// Options are what the shell passes to Open.
type Options struct {
	ConfigPath string              // --config / PROVEN_CONFIG; "" means the defaults
	PolicyPath string              // PROVEN_POLICY, read by the shell; "" means no organisation policy (ADR-0036)
	Sample     bool                // --sample / SAMPLE_MODE=true: every adapter is its sample implementation (ADR-0001)
	LogDir     string              // the run's logs/ directory; the reviewer writes its JSONL trace here (ADR-0018)
	Env        func(string) string // environment lookup; keys come from the environment only (ADR-0001)
	// Now is the review's clock, compared with an exception's expiry date . nil means time.Now;
	// a test sets it so an exception expires on a day the test chose.
	Now func() time.Time
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
	// tel, when set, replaces the telemetry the OTEL_* variables build .
	// A test sets it to read the spans and numbers back from memory.
	tel *telemetry.T
	// recordsClient, when set, is the HTTP client of the evidence records' writer
	// and of the viewer check. A review leaves it unset. Tests set it to a
	// recorded CouchDB, so no test dials one.
	recordsClient *http.Client
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

	// read is how this review reads the repository: its git policy and size bounds. The runner sets
	// it from the policy fixed at Open, so no later Open in the process can change it .
	read change.Policy
}

type Severity = report.Severity
type Result = report.Result
type ReviewerInfo = report.ReviewerInfo
type ReviewerVerdict = report.ReviewerVerdict
type ProofLevel = report.ProofLevel
type Opinion = report.Opinion
type AcceptedFinding = report.AcceptedFinding
type ExpiredException = report.ExpiredException
type AuthorModel = report.AuthorModel
type Finding = report.Finding
type Fix = report.Fix
type FixEdit = report.FixEdit
type FixRun = report.FixRun
type DependencyFix = report.DependencyFix
type Autofix = report.Autofix
type AutofixFinding = report.AutofixFinding
type AutofixProof = report.AutofixProof
type Location = report.Location
type Evidence = report.Evidence
type ToolEvidence = report.ToolEvidence
type TestEvidence = report.TestEvidence
type QueryEvidence = report.QueryEvidence
type Refutation = report.Refutation
type ServiceFailure = report.ServiceFailure
type ModelUse = report.ModelUse

const (
	SeverityInfo         = report.SeverityInfo
	SeverityWarning      = report.SeverityWarning
	SeverityError        = report.SeverityError
	RoleDefault          = report.RoleDefault
	RoleSpecial          = report.RoleSpecial
	VerdictReported      = report.VerdictReported
	VerdictNotReported   = report.VerdictNotReported
	VerdictRefuted       = report.VerdictRefuted
	VerdictNoAnswer      = report.VerdictNoAnswer
	ProvedByTest         = report.ProvedByTest
	ProvedByTool         = report.ProvedByTool
	ProvedByData         = report.ProvedByData
	AgreedNotProven      = report.AgreedNotProven
	FixPatch             = report.FixPatch
	FixDependency        = report.FixDependency
	AutofixFixed         = report.AutofixFixed
	AutofixRejected      = report.AutofixRejected
	AutofixProofTest     = report.AutofixProofTest
	AutofixProofData     = report.AutofixProofData
	AutofixProofTool     = report.AutofixProofTool
	AutofixProofBreakage = report.AutofixProofBreakage
	EvidenceTest         = report.EvidenceTest
	EvidenceQuery        = report.EvidenceQuery
	EvidenceTool         = report.EvidenceTool
	OpinionLevel         = report.OpinionLevel
)

var ProofLevels = report.ProofLevels

func ProofLevelOf(ev []Evidence) (ProofLevel, error) { return report.ProofLevelOf(ev) }
