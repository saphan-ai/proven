// Package candidate is the finding record passed between the six agents.
package candidate

// Evidence kind strings. They match the seam's EvidenceTest and EvidenceQuery.
const (
	KindTest  = "test"
	KindQuery = "query"
)

// Risk is one candidate the planner named.
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
type TestEvidence struct {
	Path          string
	Name          string
	Command       string
	Output        string
	FailsOnChange bool
	PassesOnBase  bool
}

// QueryEvidence is a query and the rows the engine returned.
type QueryEvidence struct {
	SQL      string
	Engine   string
	Snapshot string
	Columns  []string
	Rows     [][]string
}

// Evidence is one piece of proof.
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
type Refutation struct {
	Attempted         bool
	Refuted           bool
	Reason            string
	TestPath          string
	TestName          string
	Command           string
	SQL               string
	TestFound         bool
	TestFailsOnChange bool
	TestCannotRun     bool
	QueryRows         [][]string
	QueryCannotRun    bool
}

// Candidate is a risk plus what the later agents added.
type Candidate struct {
	Risk
	Evidence   []Evidence
	Refutation Refutation
	Dropped    bool
	DropReason string
	Unresolved bool
}

// Failure names a service the review degraded around.
type Failure struct {
	Service string
	Reason  string
}

// ModelUse is one generator call, as the receipt named it.
type ModelUse struct {
	Agent       string
	Route       string // winning_route.provider when the stream named it; else ""
	Model       string // routed model; "unknown" when the stream named none
	ReceiptID   string // meterless.receipt_id; "" when the stream named none
	FromReceipt bool   // true only when the stream named a routed model
}

// Draft is what the writer hands to the seam.
type Draft struct {
	Findings  []Candidate
	Failed    []Failure
	Models    []ModelUse
	Partial   bool
	TracePath string
}
