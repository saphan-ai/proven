// Package candidate is the finding record passed between the six agents.
//
// The Planner fills a Risk. The Investigator and the Prover add Evidence, the
// Skeptic fills the Refutation, the Judge sets Dropped, Unresolved and
// DropReason, and the Writer hands the kept ones to the seam as a Draft. The
// Fixer, after the Judge, adds a Fix to a kept finding (ADR-0042). Each
// agent only adds its own part, so the trace can show who said what.
//
// The package has no dependency on the seam (internal/review), so the agents
// never import the shell's types; review.Open converts a Draft to a Result.
package candidate

import "github.com/saphan-ai/proven/internal/querytpl"

// Evidence kind strings. They match the seam's EvidenceTest and EvidenceQuery,
// so the conversion to a review.Result copies the kind unchanged.
const (
	KindTest  = "test"
	KindQuery = "query"
)

// Risk is one candidate the planner named.
// A Risk is a claim only: it is never reported until the Judge finds evidence
// for it (ADR-0002). Severity is INFO, WARNING or ERROR, as on the seam.
type Risk struct {
	ID        string
	RuleID    string
	Title     string
	Message   string
	Severity  string
	Path      string
	StartLine int
	EndLine   int
}

// TestEvidence is a recorded or measured test outcome.
// It counts as evidence only when FailsOnChange and PassesOnBase are both true:
// the test shows the change broke something the base did not (ADR-0002).
type TestEvidence struct {
	Path    string
	Name    string
	Command string
	Output  string
	// Content is the text of a test file the Prover wrote; empty for a test that
	// already existed. The Fixer needs it to run the same test again on a patched
	// change (ADR-0042), and it is not part of what the review shows.
	Content       string
	FailsOnChange bool
	PassesOnBase  bool
}

// Fix kinds. A patch is proved by a test; a dependency fix is named from an advisory row.
const (
	FixPatch      = "patch"
	FixDependency = "dependency"
)

// FixEdit is one change of a proposed patch: the one place where Old stands in the
// file at Path is replaced by New (ADR-0042 point 1).
type FixEdit struct {
	Path string
	Old  string
	New  string
}

// FixRun is one run of the proof of a fix, as the test runner recorded it: the
// command, how it ended and its exit code, so a reader can run it again (ADR-0042 point 7).
// Label names the tree the run was made in; ExitCode is -1 when the run had none.
type FixRun struct {
	Label    string
	Command  string
	Result   string
	ExitCode int
	Output   string
}

// DependencyFix is the smallest version of a package that leaves the advisory's
// affected range, with the advisory row that shows it (ADR-0042 point 5).
type DependencyFix struct {
	Ecosystem string
	Package   string
	From      string
	To        string
	// AdvisoryID, AdvisoryColumns and AdvisoryRow are the row of the OSV table whose
	// fixed version To is, so a reader can check the choice against the table.
	AdvisoryID      string
	AdvisoryColumns []string
	AdvisoryRow     []string
}

// Fix is what the Fixer proposes for one finding (ADR-0042).
// A Proved patch carries its Edits and the Runs that proved it. A patch that was not
// proved carries no Edits, because the review does not show it as a fix (point 3):
// Note says the attempt failed, and Runs hold the last attempt's record.
// A dependency fix carries Dependency and no Runs: no test proves a version bump.
type Fix struct {
	Kind     string
	Proved   bool
	Attempts int
	// What and Why are the model's account of the change and of why it closes the
	// problem (point 7). The validation steps are the Runs, which Proven wrote down.
	What       string
	Why        string
	Edits      []FixEdit
	Runs       []FixRun
	Note       string
	Dependency *DependencyFix
}

// QueryEvidence is a query and the rows the engine returned.
// Engine and Snapshot say where the rows came from, so a reader can run the
// same SQL again (ADR-0023).
type QueryEvidence struct {
	SQL      string
	Engine   string
	Snapshot string
	Columns  []string
	Rows     [][]string
}

// Evidence is one piece of proof.
// Kind is KindTest or KindQuery; the matching pointer is set and the other is nil.
type Evidence struct {
	Kind  string
	Test  *TestEvidence
	Query *QueryEvidence
}

// Refutation is one skeptic attempt.
// Refuted is the model's claim. It does not drop the finding.
// A drop needs TestFound with TestFailsOnChange false, or QueryRows that
// contradict the finding. TestCannotRun or QueryCannotRun means the
// instrument did not answer.
// Query is the template and values the Skeptic named (ADR-0032). It holds no
// SQL: the pipeline builds the SQL from it, after Proven has checked each value.
// A request that was refused is cleared, and Reason says why.
type Refutation struct {
	Attempted         bool
	Refuted           bool
	Reason            string
	TestPath          string
	TestName          string
	Query             querytpl.Request
	TestFound         bool
	TestFailsOnChange bool
	TestCannotRun     bool
	QueryRows         [][]string
	QueryCannotRun    bool
}

// Candidate is a risk plus what the later agents added.
// Evidence comes from the Investigator and the Prover, Refutation from the
// Skeptic. Dropped, DropReason and Unresolved are the Judge's verdict: a dropped
// candidate goes to the trace and H-MEM, not to the review; an unresolved one
// is reported and marked.
type Candidate struct {
	Risk
	Evidence   []Evidence
	Refutation Refutation
	Dropped    bool
	DropReason string
	Unresolved bool
	// Runs names the reviews that reported this risk: "default", "special", or both.
	Runs []string
	// Models names the reviewer models for those runs, in the same order.
	Models []string
	// ProverNote is why a written test is not evidence. The pipeline copies it
	// into the trace. Empty when the prover did not write a test, or when the
	// written test is evidence.
	ProverNote string
	// QueryNote is why the Investigator's answer is not query evidence: a
	// template that was refused (ADR-0032). The pipeline copies it into the
	// trace. Empty when the Investigator proposed no query, or the query ran.
	QueryNote string
	// Fix is the Fixer's proposal for this finding, or nil when it made none (ADR-0042).
	Fix *Fix
}

// Failure names a service the review degraded around.
// It becomes a review.ServiceFailure on the seam.
type Failure struct {
	Service string
	Reason  string
}

// ModelUse is one generator call, as the receipt named it.
// The model is never guessed: without a receipt it stays "unknown" (ADR-0019).
type ModelUse struct {
	Agent       string
	Route       string // winning_route.provider when the stream named it; else ""
	Model       string // routed model; "unknown" when the stream named none
	ReceiptID   string // meterless.receipt_id; "" when the stream named none
	FromReceipt bool   // true only when the stream named a routed model
}

// Draft is what the writer hands to the seam.
// Partial is set when the deadline cut the review short; TracePath is the
// JSONL trace of the run.
type Draft struct {
	Findings  []Candidate
	Failed    []Failure
	Models    []ModelUse
	Partial   bool
	TracePath string
	// SkippedPaths lists tracked paths a live review did not copy because they
	// are symlinks, slash-separated and sorted. Empty when none were skipped.
	SkippedPaths []string
}
