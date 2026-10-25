package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/autofix"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

// osvRoot is the sample change whose go.mod raises example.com/widget to v1.4.0, inside the advisory
// range 1.3.0 up to 1.5.0 of the sample table, with the query runner over that table.
func osvRoot(t *testing.T) (string, query.Runner) {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(dir, "..", "..", "testdata", "fixture", "osv-go-module")
	q, err := query.OpenSample(filepath.Join(root, "tables.sql"))
	if err != nil {
		t.Fatal(err)
	}
	return root, q
}

// provedByData runs the dependency check on the sample change and returns its data-proved finding.
func provedByData(t *testing.T, root string, q query.Runner) candidate.Candidate {
	t.Helper()
	out, err := osv.Check(context.Background(), q, root, "", "sample")
	if err != nil || len(out.Candidates) != 1 {
		t.Fatalf("%v %#v", err, out)
	}
	proved := autofix.ProvedFindings(out.Candidates)
	if len(proved) != 1 {
		t.Fatalf("the dependency finding is not proved by data: %#v", out.Candidates[0].Evidence)
	}
	return proved[0]
}

func widgetEdit(to string) []testrun.Edit {
	return []testrun.Edit{{Path: "go.mod", Old: "example.com/widget v1.4.0", New: "example.com/widget " + to}}
}

// TestDataJudgeClosesAFindingOnlyWhenTheSameQueryNoLongerReturnsTheRow proves ADR-0050 point 2 for a
// finding proved by data: with the dependency raised past the advisory's fixed version the same check no
// longer names it, and with a raise that stays inside the range the row still comes back with its rows.
func TestDataJudgeClosesAFindingOnlyWhenTheSameQueryNoLongerReturnsTheRow(t *testing.T) {
	root, q := osvRoot(t)
	c := provedByData(t, root, q)
	judge := dataJudge{q: q, root: root, snapshot: "sample", maxBytes: 0}
	qe := c.Evidence[0].Query

	fixed, err := judge.Rerun(context.Background(), widgetEdit("v1.5.0"), c, qe)
	if err != nil || !fixed.Ran || fixed.Present {
		t.Fatalf("the fixed version still returns the row: %v %#v", err, fixed)
	}
	still, err := judge.Rerun(context.Background(), widgetEdit("v1.4.1"), c, qe)
	if err != nil || !still.Ran || !still.Present || !strings.Contains(still.Output, "GHSA-w19-sample-0001") {
		t.Fatalf("a version inside the range does not return the row: %v %#v", err, still)
	}
	if !strings.Contains(strings.ToLower(still.Command), "px_osv_pkg") {
		t.Fatalf("the re-run does not name its query: %q", still.Command)
	}
}

// TestDataJudgeNeverClosesAFindingItCouldNotRerun proves a store that is down is an open finding with the
// service named, and a patch that does not apply is an error and not a closed finding.
func TestDataJudgeNeverClosesAFindingItCouldNotRerun(t *testing.T) {
	root, q := osvRoot(t)
	c := provedByData(t, root, q)
	down := dataJudge{q: failingQuery{&service.Down{Service: "zetaris", Reason: "unreachable"}}, root: root, snapshot: "sample"}
	got, err := down.Rerun(context.Background(), widgetEdit("v1.5.0"), c, c.Evidence[0].Query)
	if err != nil || got.Ran || got.Present || !strings.Contains(got.Reason, "zetaris") {
		t.Fatalf("%v %#v", err, got)
	}
	judge := dataJudge{q: q, root: root, snapshot: "sample"}
	if _, err := judge.Rerun(context.Background(), []testrun.Edit{{Path: "go.mod", Old: "absent text", New: "x"}}, c, c.Evidence[0].Query); err == nil {
		t.Fatal("a patch that does not apply was judged")
	}
}

type failingQuery struct{ err error }

func (f failingQuery) Run(context.Context, querytpl.Request, querytpl.Relations) (query.Result, error) {
	return query.Result{}, f.err
}

// TestDataJudgeLeavesTheReviewedTreesAsTheyWere proves ADR-0050 point 4: the patch is applied to copies,
// so after the re-run the change's go.mod is byte for byte what it was.
func TestDataJudgeLeavesTheReviewedTreesAsTheyWere(t *testing.T) {
	root, q := osvRoot(t)
	c := provedByData(t, root, q)
	path := filepath.Join(root, "change", "go.mod")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (dataJudge{q: q, root: root, snapshot: "sample"}).Rerun(context.Background(), widgetEdit("v1.5.0"), c, c.Evidence[0].Query); err != nil {
		t.Fatal(err)
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatalf("the change's go.mod was written:\n%s", after)
	}
}

// passingPackages is a breakage runner under which every package the patch touched passes; it keeps the requests.
type passingPackages struct{ got []testrun.BreakageRequest }

func (p *passingPackages) RunBreakage(_ context.Context, req testrun.BreakageRequest) (testrun.BreakageOutcome, error) {
	p.got = append(p.got, req)
	var out testrun.BreakageOutcome
	for _, argv := range testrun.BreakageArgvs(req.Edits) {
		out.Runs = append(out.Runs, testrun.FixRun{Label: testrun.RunBreakage + " " + argv[3], Command: strings.Join(argv, " "), Result: testrun.ResultPassed})
	}
	return out, nil
}

// TestACheckoutRunnerThatCannotRunPackagesIsTheProverOff proves the wrapper never turns a runner that
// cannot run a patched package into an empty, passing outcome: it reports the Prover off, and fills the
// change directory from the review's own tree when the caller left it empty.
func TestACheckoutRunnerThatCannotRunPackagesIsTheProverOff(t *testing.T) {
	edits := []testrun.Edit{{Path: "a.go", Old: "a", New: "b"}}
	out, err := (checkoutRun{inner: noPackages{}, change: "/c"}).RunBreakage(context.Background(), testrun.BreakageRequest{Edits: edits})
	if err == nil || len(out.Runs) != 0 {
		t.Fatalf("%v %#v", err, out)
	}
	pkgs := &passingPackages{}
	if _, err := (checkoutRun{inner: pkgsAdapter{pkgs}, change: "/c"}).RunBreakage(context.Background(), testrun.BreakageRequest{Edits: edits}); err != nil || len(pkgs.got) != 1 || pkgs.got[0].ChangeDir != "/c" {
		t.Fatalf("%v %#v", err, pkgs.got)
	}
}

// noPackages is an adapter that cannot run a patched package: it is not a testrun.BreakageRunner.
type noPackages struct{}

func (noPackages) Run(context.Context, testrun.Request) (testrun.Outcome, error) {
	return testrun.Outcome{}, nil
}

// pkgsAdapter lets passingPackages stand behind a checkoutRun, which wraps a testrun.Adapter.
type pkgsAdapter struct{ *passingPackages }

func (pkgsAdapter) Run(context.Context, testrun.Request) (testrun.Outcome, error) {
	return testrun.Outcome{}, nil
}

// scripted answers the attempt's prompt with one reply.
type scripted struct{ reply string }

func (s scripted) Complete(context.Context, string, string) (string, error) { return s.reply, nil }

// TestAnAttemptOnADependencyFindingIsJudgedByTheCheckAgain proves ADR-0050 points 2 and 3 end to end for
// data: the generator's raise to 1.5.0 ends "fixed automatically, waiting for a person", and its raise to
// 1.4.1, which stays inside the range, ends "rejected" with the finding open and its rows shown.
func TestAnAttemptOnADependencyFindingIsJudgedByTheCheckAgain(t *testing.T) {
	root, q := osvRoot(t)
	c := provedByData(t, root, q)
	pkgs := &passingPackages{}
	judges := autofix.Judges{Data: dataJudge{q: q, root: root, snapshot: "sample"}, Breakage: pkgs}
	opts := autofix.Options{BaseDir: filepath.Join(root, "base"), ChangeDir: filepath.Join(root, "change")}
	answer := func(to string) scripted {
		return scripted{`{"what":"Raise the widget.","why":"The version leaves the range.","edits":[{"path":"go.mod","old":"example.com/widget v1.4.0","new":"example.com/widget ` + to + `"}]}`}
	}
	fixed, err := autofix.Attempt(context.Background(), []candidate.Candidate{c}, answer("v1.5.0"), judges, opts)
	if err != nil || !fixed.Fixed() || !fixed.Findings[0].Closed {
		t.Fatalf("%v %#v", err, fixed)
	}
	// A manifest edit asks for the whole module, with Proven's own words, on the change's own tree.
	if len(pkgs.got) != 1 || pkgs.got[0].ChangeDir != filepath.Join(root, "change") || fixed.Breakage == nil || !fixed.Breakage.Closed || fixed.Breakage.Command != "go test -json ./..." {
		t.Fatalf("the module was not run for a manifest edit: %#v %#v", pkgs.got, fixed.Breakage)
	}
	if !strings.Contains(fixed.Diff, "-require example.com/widget v1.4.0") || !strings.Contains(fixed.Diff, "+require example.com/widget v1.5.0") {
		t.Fatalf("no diff of the raise:\n%s", fixed.Diff)
	}
	rejected, err := autofix.Attempt(context.Background(), []candidate.Candidate{c}, answer("v1.4.1"), judges, opts)
	if err != nil || rejected.Verdict != candidate.AutofixRejected || rejected.Findings[0].Closed {
		t.Fatalf("%v %#v", err, rejected)
	}
	if p := rejected.Findings[0].Proofs[0]; p.Kind != candidate.AutofixData || !strings.Contains(p.Output, "GHSA-w19-sample-0001") {
		t.Fatalf("the open finding does not carry the row: %#v", p)
	}
}

// TestToolJudgeNeedsAContainerAndAKnownTool proves a tool finding is open, with the reason, when the review
// has no container (a sample review starts none) or when the tool is not one whose result is evidence.
func TestToolJudgeNeedsAContainerAndAKnownTool(t *testing.T) {
	c := candidate.Candidate{Risk: candidate.Risk{ID: "T1", Path: "counter/counter.go"}}
	race := &candidate.ToolEvidence{Tool: "go-race", PresentOnChange: true}
	for name, tc := range map[string]struct {
		judge toolJudge
		ev    *candidate.ToolEvidence
		want  string
	}{
		"no container": {toolJudge{timeout: 1}, race, "no container"},
		"unknown tool": {toolJudge{timeout: 1}, &candidate.ToolEvidence{Tool: "rm"}, "no command is known"},
	} {
		got, err := tc.judge.Rerun(context.Background(), nil, c, tc.ev)
		if err != nil || got.Ran || !strings.Contains(got.Reason, tc.want) {
			t.Fatalf("%s: %v %#v", name, err, got)
		}
	}
}

// TestAutofixTextIsRedacted proves ADR-0005 for the attempt: a configured secret the generator echoed in
// its explanation, its patch or a re-run's output is replaced before the result is built.
func TestAutofixTextIsRedacted(t *testing.T) {
	rd := redact.New(map[string]string{"SECRET_TOKEN": "s3cret-value"})
	a := &candidate.Autofix{
		Verdict: candidate.AutofixFixed, What: "uses s3cret-value", Why: "s3cret-value", Diff: "+s3cret-value", Refused: "s3cret-value",
		Edits: []candidate.FixEdit{{Path: "a.go", Old: "x", New: "s3cret-value"}},
		Findings: []candidate.AutofixFinding{{ID: "F1", Proofs: []candidate.AutofixProof{{
			Reason: "s3cret-value", Output: "s3cret-value", Command: "echo s3cret-value",
			Runs: []candidate.FixRun{{Label: "base", Output: "s3cret-value", Command: "s3cret-value"}},
		}}}},
		Breakage: &candidate.AutofixProof{
			Reason: "s3cret-value", Output: "s3cret-value", Command: "go test s3cret-value",
			Runs: []candidate.FixRun{{Label: "breakage ./a", Output: "s3cret-value", Command: "s3cret-value"}},
		},
	}
	got := redactAutofix(rd, a)
	text := got.What + got.Why + got.Diff + got.Refused + got.Edits[0].New
	p := got.Findings[0].Proofs[0]
	text += p.Reason + p.Output + p.Command + p.Runs[0].Output + p.Runs[0].Command
	b := got.Breakage
	text += b.Reason + b.Output + b.Command + b.Runs[0].Output + b.Runs[0].Command
	if strings.Contains(text, "s3cret-value") {
		t.Fatalf("a secret reached the result: %s", text)
	}
	if redactAutofix(rd, nil) != nil || a.What != "uses s3cret-value" {
		t.Fatal("a nil attempt did not stay nil, or the original was changed")
	}
}
