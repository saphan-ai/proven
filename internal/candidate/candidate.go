// Package candidate is the finding record passed between the six agents.
//
// The Planner fills a Risk. The Investigator and the Prover add Evidence, the
// Skeptic fills the Refutation, the Judge sets Dropped, Unresolved and
// DropReason, and the Writer hands the kept ones to the seam as a Draft.
// A claim both reviewers reported that the Judge dropped for no evidence is
// an Opinion on that draft, not a finding (ADR-0041). The Fixer, after the
// Judge, adds a Fix to a kept finding (ADR-0042). With the whole-change attempt on, one Autofix on the
// draft records the single patch for every proved finding and the verdict on it (ADR-0050). A dependency finding also
// carries its ecosystem and, after govulncheck, a reachability label
// (ADR-0043). Each agent only adds its own part, so the trace can show who
// said what.
//
// The package has no dependency on the seam (internal/review), so the agents
// never import the shell's types; review.Open converts a Draft to a Result.
package candidate

import "github.com/saphan-ai/proven/internal/querytpl"

// Evidence kind strings. They match the seam's EvidenceTest, EvidenceQuery
// and EvidenceTool, so the conversion to a review.Result copies the kind unchanged.
const (
	KindTest  = "test"
	KindQuery = "query"
	// KindTool is a deterministic tool's run compared between base and change (ADR-0040).
	KindTool = "tool"
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

// Verdicts of the whole-change fix attempt (ADR-0050 point 3). They are the words a person reads.
const (
	// AutofixFixed means every proved finding closed on the patched change and every package
	// the patch touched still passes its tests. A person still decides: nothing was committed.
	AutofixFixed = "fixed automatically, waiting for a person"
	// AutofixRejected means the one attempt left a proved finding open, broke something, or
	// was refused before it ran. There is no second attempt.
	AutofixRejected = "rejected"
)

// The kinds of proof a finding is judged by again (ADR-0050 point 2).
const (
	AutofixTest = "test"
	AutofixData = "data"
	AutofixTool = "tool"
	// AutofixBreakage is the proof that nothing the patch touched broke: the packages it edited, run again.
	AutofixBreakage = "breakage"
)

// AutofixProof is one piece of a finding's own evidence, run again on the patched change.
// Closed is true only when the run happened and the evidence no longer shows the problem;
// a run that could not happen is never closed (ADR-0050 point 2: no opinion, no silence).
// Command and Output are what Proven ran and what it printed, so a reader can run it again.
type AutofixProof struct {
	Kind    string
	Closed  bool
	Reason  string
	Command string
	Output  string
	Runs    []FixRun
}

// AutofixFinding is the verdict on one proved finding: closed when every one of its proofs closed.
type AutofixFinding struct {
	ID     string
	Closed bool
	Proofs []AutofixProof
}

// Autofix is the record of the one whole-change fix attempt (ADR-0050). Edits and Diff are
// what the generator wrote, kept for the record whatever the verdict. Only a Fixed verdict
// is a suggestion; a rejected patch is never shown as a fix (ADR-0042 point 3).
type Autofix struct {
	Verdict  string
	What     string
	Why      string
	Edits    []FixEdit
	Diff     string
	Findings []AutofixFinding
	// Breakage is the run of every package the patch touched, on the patched change; it is nil when
	// the attempt never reached the judges. The verdict is Fixed only when it is closed.
	Breakage *AutofixProof
	// Refused names why the patch was refused before any code ran, or is empty.
	Refused string
}

// Fixed reports whether the attempt ended in the verdict "fixed automatically, waiting for a person".
func (a *Autofix) Fixed() bool { return a != nil && a.Verdict == AutofixFixed }

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
// Form names the query language that ran. Engine and Snapshot say where the
// rows came from, so a reader can run the same query again (ADR-0023, ADR-0047).
//
// Bound is set by Proven, never by a model: the rows are about the change's own
// package, version or advisory (the dependency check derives those from the
// change's manifests and compares versions in Go). The Judge counts query rows
// as evidence only when Bound is set; a word that a row shares with the
// finding's text binds nothing. Request is the template and values the
// Investigator named, kept so the pipeline can decide Bound; it holds no SQL.
type QueryEvidence struct {
	SQL      string
	Form     string
	Engine   string
	Snapshot string
	Columns  []string
	Rows     [][]string
	Request  querytpl.Request
	Bound    bool
}

// ToolEvidence is a deterministic tool's run on the base and the change (ADR-0040).
// It counts as evidence only when PresentOnChange is true and PresentOnBase is
// false: the change produced a result the base did not. A tool that did not
// run is not this struct; the runner names that tool instead.
type ToolEvidence struct {
	Tool            string
	Version         string
	Command         string
	Output          string // the change side, the output that differs
	BaseOutput      string
	PresentOnChange bool
	PresentOnBase   bool
}

// Evidence is one piece of proof.
// Kind is KindTest, KindQuery or KindTool; the matching pointer is set and the others are nil.
type Evidence struct {
	Kind  string
	Test  *TestEvidence
	Query *QueryEvidence
	Tool  *ToolEvidence
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
	// Verdicts holds what each reviewer said about this risk, one per run (ADR-0037 point 4). It is
	// empty when only one reviewer ran. The verdicts are kept side by side and never combined.
	Verdicts []Verdict
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
	// Ecosystem is the OSV ecosystem of a dependency finding, for example "Go".
	// Empty when the finding did not come from the dependency check.
	Ecosystem string
	// ReachLabel is how far the vulnerable function was measured (ADR-0043).
	// It is ReachReached, ReachImported or ReachNotMeasured, or empty when
	// the finding is not a dependency finding.
	ReachLabel string
	// CallPath is the call path govulncheck printed. It is set only when
	// ReachLabel is ReachReached.
	CallPath string
}

// How far a dependency finding's vulnerable function was measured (ADR-0043).
// A reached function ranks above one that is only present. A finding is never
// dropped because it was not reached.
const (
	// ReachReached means the vulnerable function is called from the change.
	ReachReached = "reached"
	// ReachImported means the dependency is present and the vulnerable function is not called.
	ReachImported = "imported but not reached"
	// ReachNotMeasured means no tool measured this finding. Other ecosystems
	// have no such tool. A Go finding is also not measured when govulncheck
	// did not answer: that is never reported as not reached.
	ReachNotMeasured = "not measured"
)

// The four things one reviewer can say about a risk (ADR-0037 point 4).
const (
	// VerdictReported: the reviewer named the risk and its own Skeptic did not claim to refute it.
	VerdictReported = "reported"
	// VerdictNotReported: the reviewer's own review did not name the risk.
	VerdictNotReported = "not reported"
	// VerdictRefuted: the reviewer named the risk and its own Skeptic claimed to refute it.
	VerdictRefuted = "refuted"
	// VerdictNoAnswer: the reviewer's service failed, so it has no word on the risk. It is not a
	// disagreement: a missing answer is not a different answer.
	VerdictNoAnswer = "no answer"
)

// Verdict is one reviewer's word on one risk. Run is "default" or "special", Model the reviewer's
// name, and Reason the reviewer's own words, or the plain reason it has none.
type Verdict struct {
	Run     string
	Model   string
	Verdict string
	Reason  string
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
	Requested   string // the model field the request sent; "" for a call that sent none
	ResponseID  string // the completion response id, which the receipt names as its request id; "" when none
}

// Opinion is a claim both reviewers reported at one place that no evidence
// proves (ADR-0041). It is not a finding: the Writer lifts it out of the
// candidates the Judge dropped, and the totals do not count it.
type Opinion struct {
	ID        string
	Title     string
	Path      string
	StartLine int // the lines both ranges share; equal to EndLine when the place is one line
	EndLine   int
	// Reasons holds each reviewer's own words, the default reviewer first.
	Reasons []Verdict
}

// Draft is what the writer hands to the seam.
// Partial is set when the deadline cut the review short; TracePath is the
// JSONL trace of the run.
type Draft struct {
	Findings  []Candidate
	Opinions  []Opinion
	Failed    []Failure
	Models    []ModelUse
	Partial   bool
	TracePath string
	// SkippedPaths lists tracked paths a live review did not copy because they
	// are symlinks, slash-separated and sorted. Empty when none were skipped.
	SkippedPaths []string
	// GuardrailStatement is set when the review ran with the guardrail
	// switched off, so the result says so and not only the trace
	// (ADR-0005). Empty when the guardrail was on.
	GuardrailStatement string
	// Guardrail counts the guardrail's checks in this review (ADR-0005). The
	// zero value means no check was made.
	Guardrail GuardrailCount
	// Autofix is the record of the whole-change fix attempt, or nil when none was made:
	// the setting is off, or no finding was proved (ADR-0050).
	Autofix *Autofix
}

// GuardrailCount is what the guardrail did in one review, measured at the one
// place every check passes through and never typed by hand.
// Prompts and Answers are the checks that got a verdict, Blocked the ones whose
// verdict was unsafe, and Unanswered the checks the guardrail could not make.
// Backend and Model name the guardrail that answered, or failed to.
type GuardrailCount struct {
	Prompts, Answers, Blocked, Unanswered int
	Backend, Model                        string
}

// GuardrailOffStatement is the sentence a review carries when its guardrail
// was switched off. The trace line, the draft and every format use this one text.
const GuardrailOffStatement = "no guardrail: this review ran with the guardrail switched off"
