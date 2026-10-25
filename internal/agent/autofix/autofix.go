// Package autofix makes the one automated fix attempt of a review and judges it by the same evidence.
//
// ADR-0050: after the review has its proved findings, the generator gets all of them at once, each
// with the evidence that proved it, and writes one patch for the whole change. The patch is judged
// by no model. Proven runs each finding's own evidence again on the patched change, in the places
// where the change's code may run (ADR-0025): a finding proved by a test closes when that test
// passes and the package's other tests still pass; a finding proved by data closes when the same
// query no longer returns the row; a finding proved by a tool closes when the same tool run no
// longer reports it. Then nothing else may have broken: every package a patched file belongs to is
// run on the patched change, and a failure, or a run that did not happen, is not "fixed". If every
// proved finding closed and those runs passed, the verdict is "fixed automatically, waiting
// for a person"; otherwise it is "rejected", with each open finding and its re-run evidence. There
// is no second attempt (point 3), and nothing here writes to the change or to the author's branch
// (point 4): the patch is applied to copies only, and it is a suggestion a person decides on.
package autofix

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/saphan-ai/proven/internal/agent/answer"
	"github.com/saphan-ai/proven/internal/agent/fixer"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

// Agent is the name the generator is asked under, so a recorded sample answer and a trace line
// name the whole-change attempt apart from the Fixer's single-finding suggestions (ADR-0042).
const Agent = "autofix"

// maxFileBytes bounds the text of each finding's file in the prompt, and maxOutputBytes bounds each
// piece of evidence output, so a prompt about many findings stays small.
const (
	maxFileBytes   = 16 << 10
	maxOutputBytes = 2 << 10
	maxProseBytes  = 2 << 10
	maxRows        = 20
)

// spec is the JSON the attempt asks the generator for. It asks for edits and never for a
// command: Proven builds every command that runs here (ADR-0050 point 5, ADR-0025).
var spec = answer.Spec{
	Task: "Write one patch for the whole change that closes every finding in the data. " +
		"Each finding carries the evidence that proved it: a failing test and its output, a query and its rows, or a tool and its report. " +
		"Proven will run exactly that evidence again on the patched change, and the patch is accepted only if every finding closes and no other test of the package breaks. " +
		"Each edit names a file, the exact old text, which must stand in the file once, and the new text. " +
		"Never edit a test file and never weaken, skip or delete a test. " +
		"Say in what what the patch changes and in why why it closes the findings.",
	Shape:   "one JSON object. It has what (string), why (string) and edits (a list of objects with path, old and new, all strings). Do not add a command: Proven runs the evidence again.",
	Example: `{"what":"Subtract the discount once and raise the dependency.","why":"The price took the percent off twice, and the dependency version was inside the advisory range.","edits":[{"path":"pkg/price.go","old":"return a - b - b","new":"return a - b"}]}`,
}

// Reading is what one run of a finding's evidence on the patched change saw. Ran is false when the
// instrument did not answer (a service down, a tool that could not start): that is never a closed
// finding, because absence of evidence is not evidence that the problem is gone. Present is true
// when the evidence still shows the problem.
type Reading struct {
	Ran     bool
	Present bool
	Command string
	Output  string
	Reason  string
}

// DataRerun runs a data-proved finding's own query again on the change with the patch applied
// (ADR-0050 point 2). The implementation builds the query from the template request Proven already
// holds; it never takes a query from the generator.
type DataRerun interface {
	Rerun(ctx context.Context, edits []testrun.Edit, c candidate.Candidate, q *candidate.QueryEvidence) (Reading, error)
}

// ToolRerun runs a tool-proved finding's own tool again on the change with the patch applied, in
// the container with no network and no secret (ADR-0050 points 2 and 5).
type ToolRerun interface {
	Rerun(ctx context.Context, edits []testrun.Edit, c candidate.Candidate, t *candidate.ToolEvidence) (Reading, error)
}

// Judges are the three instruments that run evidence again, one per kind of proof. A nil one means
// this review cannot run that kind of evidence again, and every finding that needs it stays open.
type Judges struct {
	Test testrun.FixRunner
	Data DataRerun
	Tool ToolRerun
	// Breakage runs every package the patch touched on the patched change. A nil one means the
	// attempt cannot show nothing broke, and the verdict is "rejected" (ADR-0050 point 3).
	Breakage testrun.BreakageRunner
}

// Options are the checkouts the attempt runs on and its limits. Timeout is the time limit of the
// whole attempt, the generator's call and every re-run (ADR-0050 point 6); zero is no limit.
// MaxWrittenBytes is the size limit of a patch; zero is the test runner's default.
type Options struct {
	BaseDir         string
	ChangeDir       string
	Timeout         time.Duration
	MaxWrittenBytes int
}

// row is the generator's answer. Command is read only so that an answer that carries one is refused.
type row struct {
	What    string `json:"what"`
	Why     string `json:"why"`
	Edits   []edit `json:"edits"`
	Command string `json:"command"`
}

type edit struct {
	Path string `json:"path"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

// proof is one piece of evidence that proves a finding, as the judge will run it again.
type proof struct {
	kind  string
	test  *candidate.TestEvidence
	query *candidate.QueryEvidence
	tool  *candidate.ToolEvidence
}

// ProvedFindings returns the findings the attempt is about: kept findings with at least one piece
// of deterministic proof (ADR-0050 point 2). A dropped finding, and a finding that is only an
// opinion, are not part of the attempt.
func ProvedFindings(cands []candidate.Candidate) []candidate.Candidate {
	var out []candidate.Candidate
	for _, c := range cands {
		if !c.Dropped && len(proofsOf(c)) > 0 {
			out = append(out, c)
		}
	}
	return out
}

// proofsOf lists the evidence of c that proves it, in the order the Judge keeps it:
// a test that fails on the change and passes on the base, a query whose rows Proven bound to the
// change, a tool result the change introduced (ADR-0002, ADR-0040).
func proofsOf(c candidate.Candidate) []proof {
	var out []proof
	for _, e := range c.Evidence {
		switch {
		case e.Kind == candidate.KindTest && e.Test != nil && e.Test.FailsOnChange && e.Test.PassesOnBase:
			out = append(out, proof{kind: candidate.AutofixTest, test: e.Test})
		case e.Kind == candidate.KindQuery && e.Query != nil && e.Query.Bound:
			out = append(out, proof{kind: candidate.AutofixData, query: e.Query})
		case e.Kind == candidate.KindTool && e.Tool != nil && e.Tool.PresentOnChange && !e.Tool.PresentOnBase:
			out = append(out, proof{kind: candidate.AutofixTool, tool: e.Tool})
		}
	}
	return out
}

// Attempt makes the one whole-change attempt for the proved findings among cands and returns its
// record. It returns nil, nil when there is nothing to attempt: no proved finding, or a sample
// that recorded no answer and so made no claim. A generator that did not answer returns the error,
// and no verdict, because an attempt that was never made is not a rejected one. A patch that is
// refused, that does not follow the schema, or that leaves a finding open is a "rejected" verdict.
// ADR-0050 points 1 to 3.
func Attempt(ctx context.Context, cands []candidate.Candidate, gen generator.Generator, j Judges, opts Options) (*candidate.Autofix, error) {
	proved := ProvedFindings(cands)
	if len(proved) == 0 {
		return nil, nil
	}
	parent := ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	var got row
	err := answer.Read(ctx, gen, Agent, Prompt(proved, opts.ChangeDir), spec, &got)
	if err != nil {
		var missing *generator.Missing
		if errors.As(err, &missing) {
			return nil, nil
		}
		if timedOut(parent, ctx) {
			return rejected(proved, candidate.Autofix{}, "the time limit of the attempt ended before the generator answered"), nil
		}
		if d := schemaFailure(err); d != "" {
			return rejected(proved, candidate.Autofix{}, "no patch: "+d), nil
		}
		return nil, err
	}
	a := candidate.Autofix{What: clip(got.What, maxProseBytes), Why: clip(got.Why, maxProseBytes)}
	edits := make([]testrun.Edit, len(got.Edits))
	for i, e := range got.Edits {
		edits[i] = testrun.Edit{Path: e.Path, Old: e.Old, New: e.New}
		a.Edits = append(a.Edits, candidate.FixEdit{Path: e.Path, Old: e.Old, New: e.New})
	}
	// ADR-0050 point 5: the answer is a patch and nothing else. A command in it refuses the patch whole.
	if strings.TrimSpace(got.Command) != "" {
		return rejected(proved, a, "the patch was refused: "+testrun.ReasonModelCommand), nil
	}
	if reason := testrun.RefuseEdits(opts.ChangeDir, edits, opts.MaxWrittenBytes); reason != "" {
		return rejected(proved, a, "the patch was refused: "+reason), nil
	}
	a.Diff = Diff(opts.ChangeDir, edits)
	for _, c := range proved {
		f, err := judgeFinding(ctx, c, edits, j, opts)
		if err != nil {
			if parent.Err() != nil {
				return nil, parent.Err()
			}
			if !timedOut(parent, ctx) {
				return nil, err
			}
			f = candidate.AutofixFinding{ID: c.ID, Proofs: []candidate.AutofixProof{{
				Kind: kindOf(c), Reason: "the time limit of the attempt ended before this proof ran",
			}}}
		}
		a.Findings = append(a.Findings, f)
	}
	// ADR-0050 point 3: "nothing else broke" is its own condition. It is asked only of a patch whose
	// findings all closed, since any open finding already makes the verdict "rejected".
	if allClosed(a.Findings) {
		b, err := judgeBreakage(ctx, edits, j.Breakage, opts)
		if err != nil {
			if parent.Err() != nil {
				return nil, parent.Err()
			}
			if !timedOut(parent, ctx) {
				return nil, err
			}
			b = candidate.AutofixProof{Kind: candidate.AutofixBreakage, Reason: "the time limit of the attempt ended before the touched packages were run"}
		}
		a.Breakage = &b
	}
	a.Verdict = verdictOf(a.Findings, a.Breakage)
	return &a, nil
}

// timedOut reports whether the attempt's own time limit ended while the review's context did not.
func timedOut(parent, ctx context.Context) bool {
	return parent.Err() == nil && ctx.Err() != nil
}

// schemaFailure returns the plain reason when err is the answer package's "the model did not follow
// the schema" outcome, which is an attempt made and failed, or "" for any other error.
func schemaFailure(err error) string {
	var d *service.Down
	if errors.As(err, &d) && d.Service == Agent && d.Reason == answer.FailReason {
		return answer.FailReason
	}
	return ""
}

// rejected is the record of an attempt that did not reach the judge: every proved finding is open
// and says why, and no code ran. a carries whatever the generator wrote before the refusal.
func rejected(proved []candidate.Candidate, a candidate.Autofix, why string) *candidate.Autofix {
	a.Verdict = candidate.AutofixRejected
	a.Refused = why
	for _, c := range proved {
		a.Findings = append(a.Findings, candidate.AutofixFinding{ID: c.ID, Proofs: []candidate.AutofixProof{{
			Kind: kindOf(c), Reason: why,
		}}})
	}
	return &a
}

// kindOf names the first kind of proof of c, for a record that holds no run.
func kindOf(c candidate.Candidate) string {
	if p := proofsOf(c); len(p) > 0 {
		return p[0].kind
	}
	return ""
}

// verdictOf is the whole attempt's verdict from the findings' closures and the breakage run
// (ADR-0050 point 3). It is the only place a verdict is decided, and it reads nothing but the
// closures: a patch is "fixed" because every finding's evidence ran again and closed and every
// package the patch touched still passed, never because it was written confidently. No finding,
// or no breakage run, is never "fixed".
func verdictOf(findings []candidate.AutofixFinding, breakage *candidate.AutofixProof) string {
	if !allClosed(findings) || breakage == nil || !breakage.Closed {
		return candidate.AutofixRejected
	}
	return candidate.AutofixFixed
}

// allClosed reports whether there is at least one finding and every one of them closed.
func allClosed(findings []candidate.AutofixFinding) bool {
	for _, f := range findings {
		if !f.Closed {
			return false
		}
	}
	return len(findings) > 0
}

// judgeBreakage runs every package the patch touched on the patched change and closes only when
// each run ended passed (ADR-0050 point 3). A run that did not happen, a refused request and a
// missing runner leave it open with the reason: a package that was not run is not a package that
// did not break. An error that is not the context ending leaves it open too.
func judgeBreakage(ctx context.Context, edits []testrun.Edit, run testrun.BreakageRunner, opts Options) (candidate.AutofixProof, error) {
	p := candidate.AutofixProof{Kind: candidate.AutofixBreakage}
	if run == nil {
		p.Reason = "this review cannot run the packages the patch touched"
		return p, nil
	}
	out, err := run.RunBreakage(ctx, testrun.BreakageRequest{ChangeDir: opts.ChangeDir, Edits: edits})
	if err != nil {
		if ctx.Err() != nil {
			return p, ctx.Err()
		}
		p.Reason = "the packages the patch touched could not be run: " + err.Error()
		return p, nil
	}
	if out.Refused != "" {
		p.Reason = "the patch was refused: " + out.Refused
		return p, nil
	}
	for _, r := range out.Runs {
		p.Runs = append(p.Runs, candidate.FixRun{Label: r.Label, Command: r.Command, Result: r.Result, ExitCode: r.ExitCode, Output: r.Output})
	}
	if len(p.Runs) == 0 {
		p.Reason = "no package of the patch was run"
		return p, nil
	}
	p.Command, p.Output = p.Runs[0].Command, p.Runs[0].Output
	for _, r := range p.Runs {
		if r.Result != testrun.ResultPassed {
			p.Command, p.Output = r.Command, r.Output
			p.Reason = fmt.Sprintf("the run %q ended %s (exit code %d): the patch is not shown to leave this package working", r.Label, r.Result, r.ExitCode)
			return p, nil
		}
	}
	p.Closed = true
	return p, nil
}

// judgeFinding runs every proof of c again on the patched change. The finding closes only when
// every one of its proofs closed and it has at least one (ADR-0050 point 2).
func judgeFinding(ctx context.Context, c candidate.Candidate, edits []testrun.Edit, j Judges, opts Options) (candidate.AutofixFinding, error) {
	f := candidate.AutofixFinding{ID: c.ID, Closed: true}
	for _, p := range proofsOf(c) {
		var r candidate.AutofixProof
		var err error
		switch p.kind {
		case candidate.AutofixTest:
			r, err = judgeTest(ctx, c, p.test, edits, j.Test, opts)
		case candidate.AutofixData:
			r, err = judgeData(ctx, c, p.query, edits, j.Data)
		case candidate.AutofixTool:
			r, err = judgeTool(ctx, c, p.tool, edits, j.Tool)
		}
		if err != nil {
			return candidate.AutofixFinding{}, err
		}
		if !r.Closed {
			f.Closed = false
		}
		f.Proofs = append(f.Proofs, r)
	}
	if len(f.Proofs) == 0 {
		f.Closed = false
	}
	return f, nil
}

// judgeTest runs a test-proved finding's test and its package again (ADR-0050 point 2). An error
// that is not the context ending leaves the finding open, with the reason: a Prover that is off
// cannot show the test passes, and the verdict must not say it does.
func judgeTest(ctx context.Context, c candidate.Candidate, t *candidate.TestEvidence, edits []testrun.Edit, run testrun.FixRunner, opts Options) (candidate.AutofixProof, error) {
	p := candidate.AutofixProof{Kind: candidate.AutofixTest}
	if run == nil {
		p.Reason = "this review cannot run the test again"
		return p, nil
	}
	out, err := run.RunFix(ctx, testrun.FixRequest{
		FindingID: c.ID, Path: t.Path, Name: t.Name, Content: t.Content,
		BaseDir: opts.BaseDir, ChangeDir: opts.ChangeDir, Edits: edits, Attempt: 1,
	})
	if err != nil {
		if ctx.Err() != nil {
			return p, ctx.Err()
		}
		p.Reason = "the test could not be run again: " + err.Error()
		return p, nil
	}
	if out.Refused != "" {
		p.Reason = "the patch was refused: " + out.Refused
		return p, nil
	}
	for _, r := range out.Runs {
		p.Runs = append(p.Runs, candidate.FixRun{Label: r.Label, Command: r.Command, Result: r.Result, ExitCode: r.ExitCode, Output: r.Output})
	}
	if n := len(p.Runs); n > 0 {
		last := p.Runs[n-1]
		p.Command, p.Output = last.Command, last.Output
	}
	ok, reason := fixer.Judge(p.Runs)
	p.Closed, p.Reason = ok, reason
	return p, nil
}

// judgeData runs a data-proved finding's own query again on the patched change (ADR-0050 point 2).
func judgeData(ctx context.Context, c candidate.Candidate, q *candidate.QueryEvidence, edits []testrun.Edit, rerun DataRerun) (candidate.AutofixProof, error) {
	p := candidate.AutofixProof{Kind: candidate.AutofixData}
	if rerun == nil {
		p.Reason = "this review cannot run the query again"
		return p, nil
	}
	r, err := rerun.Rerun(ctx, edits, c, q)
	if err != nil {
		if ctx.Err() != nil {
			return p, ctx.Err()
		}
		p.Reason = "the query could not be run again: " + err.Error()
		return p, nil
	}
	return fromReading(p, r, "the same query on the patched change still returns the row"), nil
}

// judgeTool runs a tool-proved finding's own tool again on the patched change (ADR-0050 point 2).
func judgeTool(ctx context.Context, c candidate.Candidate, t *candidate.ToolEvidence, edits []testrun.Edit, rerun ToolRerun) (candidate.AutofixProof, error) {
	p := candidate.AutofixProof{Kind: candidate.AutofixTool}
	if rerun == nil {
		p.Reason = "this review cannot run the tool again"
		return p, nil
	}
	r, err := rerun.Rerun(ctx, edits, c, t)
	if err != nil {
		if ctx.Err() != nil {
			return p, ctx.Err()
		}
		p.Reason = "the tool could not be run again: " + err.Error()
		return p, nil
	}
	return fromReading(p, r, "the same tool run on the patched change still reports it"), nil
}

// fromReading turns a re-run into a proof record. A run that did not happen is open with the
// instrument's own reason; a run that happened closes the finding only when the evidence is gone.
func fromReading(p candidate.AutofixProof, r Reading, still string) candidate.AutofixProof {
	p.Command, p.Output = r.Command, clip(r.Output, maxOutputBytes)
	switch {
	case !r.Ran:
		p.Reason = r.Reason
		if p.Reason == "" {
			p.Reason = "the evidence could not be run again"
		}
	case r.Present:
		p.Reason = still
	default:
		p.Closed = true
	}
	return p
}

// Prompt asks for one patch for every proved finding. Everything the change and the evidence wrote
// enters the prompt as data, through answer.Encode, so no byte of it can start an instruction
// (ADR-0050 point 1).
func Prompt(proved []candidate.Candidate, changeDir string) string {
	type proofView struct {
		Kind    string            `json:"kind"`
		Test    string            `json:"test,omitempty"`
		Path    string            `json:"test_path,omitempty"`
		Output  string            `json:"output,omitempty"`
		Query   string            `json:"query,omitempty"`
		Params  map[string]string `json:"params,omitempty"`
		Columns []string          `json:"columns,omitempty"`
		Rows    [][]string        `json:"rows,omitempty"`
		Tool    string            `json:"tool,omitempty"`
		Command string            `json:"command,omitempty"`
	}
	type findingView struct {
		ID       string      `json:"id"`
		Title    string      `json:"title"`
		Message  string      `json:"message"`
		Location string      `json:"location"`
		Proofs   []proofView `json:"proofs"`
	}
	type fileView struct {
		Path string `json:"path"`
		Text string `json:"text"`
	}
	var data struct {
		Findings []findingView `json:"findings"`
		Files    []fileView    `json:"files"`
	}
	seen := map[string]bool{}
	for _, c := range proved {
		fv := findingView{ID: c.ID, Title: c.Title, Message: c.Message, Location: fmt.Sprintf("%s:%d-%d", c.Path, c.StartLine, c.EndLine)}
		for _, p := range proofsOf(c) {
			v := proofView{Kind: p.kind}
			switch p.kind {
			case candidate.AutofixTest:
				v.Test, v.Path, v.Output = p.test.Name, p.test.Path, clip(p.test.Output, maxOutputBytes)
			case candidate.AutofixData:
				v.Query, v.Params = p.query.Request.Template, p.query.Request.Params
				v.Columns = p.query.Columns
				v.Rows = p.query.Rows
				if len(v.Rows) > maxRows {
					v.Rows = v.Rows[:maxRows]
				}
			case candidate.AutofixTool:
				v.Tool, v.Command, v.Output = p.tool.Tool, p.tool.Command, clip(p.tool.Output, maxOutputBytes)
			}
			fv.Proofs = append(fv.Proofs, v)
		}
		data.Findings = append(data.Findings, fv)
		// The file is read from the change by the checked path, never from a path the generator chose.
		if !seen[c.Path] {
			seen[c.Path] = true
			if text, _ := testrun.ReadEditable(changeDir, c.Path, maxFileBytes); text != "" {
				data.Files = append(data.Files, fileView{Path: c.Path, Text: text})
			}
		}
	}
	return answer.Instruct(spec, answer.Encode(data))
}

// clip bounds s at max bytes, on a rune boundary, and says so when it cut.
func clip(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && s[cut]&0xC0 == 0x80 {
		cut--
	}
	return s[:cut] + "\n[truncated]"
}
