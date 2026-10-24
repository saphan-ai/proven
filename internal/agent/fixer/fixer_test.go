package fixer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

// scriptedModel answers each call with the next reply and records every prompt and agent name.
type scriptedModel struct {
	replies []string
	prompts []string
	agents  []string
}

func (m *scriptedModel) Complete(_ context.Context, agent, prompt string) (string, error) {
	m.agents = append(m.agents, agent)
	m.prompts = append(m.prompts, prompt)
	if len(m.prompts) > len(m.replies) {
		return "", &generator.Missing{Agent: agent}
	}
	return m.replies[len(m.prompts)-1], nil
}

// scriptedRunner returns the next recorded attempt and keeps every request it got.
type scriptedRunner struct {
	outcomes []testrun.FixOutcome
	requests []testrun.FixRequest
	err      error
}

func (r *scriptedRunner) RunFix(_ context.Context, req testrun.FixRequest) (testrun.FixOutcome, error) {
	r.requests = append(r.requests, req)
	if r.err != nil {
		return testrun.FixOutcome{}, r.err
	}
	return r.outcomes[len(r.requests)-1], nil
}

// run builds one FixRun the way the test runner reports it.
func run(label, result string, code int, output string) testrun.FixRun {
	cmd := "go test -json -run '^TestPrice$' ./billing"
	if label == testrun.RunPackage {
		cmd = "go test -json ./billing"
	}
	return testrun.FixRun{Label: label, Command: cmd, Result: result, ExitCode: code, Output: output}
}

// good is the record of a proved fix: every run ends the way ADR-0042 point 2 asks.
func good() testrun.FixOutcome {
	return testrun.FixOutcome{Runs: []testrun.FixRun{
		run(testrun.RunBase, testrun.ResultPassed, 0, "ok"),
		run(testrun.RunChange, testrun.ResultFailed, 1, "FAIL got 60 want 80"),
		run(testrun.RunPatched, testrun.ResultPassed, 0, "ok"),
		run(testrun.RunPackage, testrun.ResultPassed, 0, "ok"),
	}}
}

// withRun returns o with the run named label replaced, so a test mutates exactly one run.
func withRun(o testrun.FixOutcome, r testrun.FixRun) testrun.FixOutcome {
	runs := append([]testrun.FixRun(nil), o.Runs...)
	for i := range runs {
		if runs[i].Label == r.Label {
			runs[i] = r
		}
	}
	return testrun.FixOutcome{Runs: runs}
}

const answerOne = `{"what":"Subtract once.","why":"The percent was taken off twice.","edits":[{"path":"billing/price.go","old":"return 60","new":"return 80"}]}`
const answerTwo = `{"what":"Subtract once, in the helper.","why":"Same cause.","edits":[{"path":"billing/price.go","old":"return 60","new":"return 80 // fixed"}]}`

// finding is a finding proved by a test, as the Prover leaves it.
func finding() candidate.Candidate {
	return candidate.Candidate{
		Risk: candidate.Risk{ID: "F1", Title: "Price is wrong", Message: "Price returns 60.", Path: "billing/price.go", StartLine: 3, EndLine: 3},
		Evidence: []candidate.Evidence{{Kind: candidate.KindTest, Test: &candidate.TestEvidence{
			Path: "billing/price_test.go", Name: "TestPrice", Command: "go test -json -run '^TestPrice$' ./billing",
			Output: "change: failed", FailsOnChange: true, PassesOnBase: true,
		}}},
	}
}

// TestAProvedPatchIsShownAsAFix proves ADR-0042 point 2: when every run ends the way the
// proof asks, the fix is proved, keeps its edits and its runs, and the runner got the
// proving test's own path, name and the edits the model wrote.
func TestAProvedPatchIsShownAsAFix(t *testing.T) {
	m := &scriptedModel{replies: []string{answerOne}}
	r := &scriptedRunner{outcomes: []testrun.FixOutcome{good()}}
	got, err := Propose(context.Background(), finding(), m, r, Options{BaseDir: "b", ChangeDir: "c"})
	if err != nil {
		t.Fatal(err)
	}
	f := got.Fix
	if f == nil || !f.Proved || f.Attempts != 1 || len(f.Edits) != 1 || len(f.Runs) != 4 || f.Note != "" {
		t.Fatalf("%#v", f)
	}
	if f.What != "Subtract once." || f.Why == "" {
		t.Fatalf("the explanation was lost: %#v", f)
	}
	req := r.requests[0]
	if req.Path != "billing/price_test.go" || req.Name != "TestPrice" || req.Attempt != 1 || req.Edits[0].New != "return 80" {
		t.Fatalf("%#v", req)
	}
	if m.agents[0] != "fixer" {
		t.Fatalf("the model was asked as %q", m.agents[0])
	}
}

// TestAFailedPatchIsRetriedWithTheFailureInHand proves ADR-0042 point 6: the second prompt
// carries the first patch and the output of the run that failed, and a second patch that
// proves the fix is shown with two attempts.
func TestAFailedPatchIsRetriedWithTheFailureInHand(t *testing.T) {
	still := withRun(good(), run(testrun.RunPatched, testrun.ResultFailed, 1, "FAIL got 70 want 80"))
	still.Runs = still.Runs[:3]
	m := &scriptedModel{replies: []string{answerOne, answerTwo}}
	r := &scriptedRunner{outcomes: []testrun.FixOutcome{still, good()}}
	got, err := Propose(context.Background(), finding(), m, r, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Fix == nil || !got.Fix.Proved || got.Fix.Attempts != 2 {
		t.Fatalf("%#v", got.Fix)
	}
	second := m.prompts[1]
	for _, want := range []string{"attempt 1 did not prove the fix", "Subtract once.", "return 80", "FAIL got 70 want 80", "exit code 1", "change+patch"} {
		if !strings.Contains(second, want) {
			t.Fatalf("the retry prompt does not carry %q:\n%s", want, second)
		}
	}
	if strings.Contains(m.prompts[0], "did not prove") {
		t.Fatal("the first prompt already carries a failure")
	}
}

// TestAPatchThatNeverPassesIsNotShownAsAFix proves ADR-0042 points 3 and 6: after the attempts the
// settings allow, the review says the fix is not proved and keeps no patch, and the last
// attempt's runs stay as the record of what failed.
func TestAPatchThatNeverPassesIsNotShownAsAFix(t *testing.T) {
	still := withRun(good(), run(testrun.RunPatched, testrun.ResultFailed, 1, "FAIL got 70 want 80"))
	still.Runs = still.Runs[:3]
	m := &scriptedModel{replies: []string{answerOne, answerTwo}}
	r := &scriptedRunner{outcomes: []testrun.FixOutcome{still, still}}
	got, err := Propose(context.Background(), finding(), m, r, Options{Attempts: 2})
	if err != nil {
		t.Fatal(err)
	}
	f := got.Fix
	if f == nil || f.Proved || len(f.Edits) != 0 || f.Attempts != 2 || len(f.Runs) != 3 {
		t.Fatalf("%#v", f)
	}
	if !strings.HasPrefix(f.Note, "fix not proved after 2 attempts:") || !strings.Contains(f.Note, "does not pass on the change with the patch applied") {
		t.Fatal(f.Note)
	}
}

// TestAPatchThatBreaksAnotherTestIsNotShownAsAFix proves ADR-0042 point 3 and the negative control
// of law #27: the only difference between the proved and the unproved record is the Result of
// the package run, and that one field turns the fix from proved to not proved.
func TestAPatchThatBreaksAnotherTestIsNotShownAsAFix(t *testing.T) {
	propose := func(o testrun.FixOutcome) *candidate.Fix {
		t.Helper()
		m := &scriptedModel{replies: []string{answerOne}}
		r := &scriptedRunner{outcomes: []testrun.FixOutcome{o}}
		got, err := Propose(context.Background(), finding(), m, r, Options{Attempts: 1})
		if err != nil {
			t.Fatal(err)
		}
		return got.Fix
	}
	if f := propose(good()); f == nil || !f.Proved {
		t.Fatalf("the control is not proved: %#v", f)
	}
	broken := withRun(good(), run(testrun.RunPackage, testrun.ResultFailed, 1, "FAIL TestNeighbour"))
	f := propose(broken)
	if f == nil || f.Proved || len(f.Edits) != 0 || !strings.Contains(f.Note, "breaks another test of the package") {
		t.Fatalf("%#v", f)
	}
}

// TestEveryOtherRunOfTheProofIsNeeded proves each run is required, one field at a time: a
// patched test that could not run, a base that failed and a change that passed each stop the fix.
func TestEveryOtherRunOfTheProofIsNeeded(t *testing.T) {
	cases := []struct {
		name   string
		mutate testrun.FixRun
		note   string
	}{
		{"patched run could not run", run(testrun.RunPatched, testrun.ResultNotRun, 125, "docker: error"), "does not pass on the change with the patch applied"},
		{"patched run timed out", run(testrun.RunPatched, testrun.ResultTimedOut, -1, ""), "does not pass on the change with the patch applied"},
		{"base fails when run again", run(testrun.RunBase, testrun.ResultFailed, 1, "FAIL"), "did not pass on the base"},
		{"change passes when run again", run(testrun.RunChange, testrun.ResultPassed, 0, "ok"), "did not fail on the change"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &scriptedModel{replies: []string{answerOne}}
			r := &scriptedRunner{outcomes: []testrun.FixOutcome{withRun(good(), tc.mutate)}}
			got, err := Propose(context.Background(), finding(), m, r, Options{Attempts: 1})
			if err != nil {
				t.Fatal(err)
			}
			if got.Fix == nil || got.Fix.Proved || !strings.Contains(got.Fix.Note, tc.note) {
				t.Fatalf("%#v", got.Fix)
			}
		})
	}
}

// TestAProblemThatDoesNotShowAgainIsNotRetried proves a failed premise stops the attempts: when
// the base does not pass on the second look, another patch cannot help, and the model is asked once.
func TestAProblemThatDoesNotShowAgainIsNotRetried(t *testing.T) {
	m := &scriptedModel{replies: []string{answerOne, answerTwo}}
	r := &scriptedRunner{outcomes: []testrun.FixOutcome{withRun(good(), run(testrun.RunBase, testrun.ResultFailed, 1, "FAIL")), good()}}
	got, err := Propose(context.Background(), finding(), m, r, Options{Attempts: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.prompts) != 1 || got.Fix == nil || got.Fix.Proved || got.Fix.Attempts != 1 {
		t.Fatalf("asked %d times: %#v", len(m.prompts), got.Fix)
	}
}

// TestTheAttemptsSettingBoundsTheModelCalls proves ADR-0042 point 6: the number of attempts is the
// setting's, and zero means the default of two.
func TestTheAttemptsSettingBoundsTheModelCalls(t *testing.T) {
	still := withRun(good(), run(testrun.RunPatched, testrun.ResultFailed, 1, "FAIL"))
	for attempts, want := range map[int]int{1: 1, 3: 3, 0: DefaultAttempts} {
		m := &scriptedModel{replies: []string{answerOne, answerOne, answerOne, answerOne}}
		r := &scriptedRunner{outcomes: []testrun.FixOutcome{still, still, still, still}}
		if _, err := Propose(context.Background(), finding(), m, r, Options{Attempts: attempts}); err != nil {
			t.Fatal(err)
		}
		if len(m.prompts) != want {
			t.Fatalf("attempts %d: the model was asked %d times, want %d", attempts, len(m.prompts), want)
		}
	}
}

// TestAFindingWithoutAProvingTestGetsNoFix proves a finding that no test proved is left as it was and
// the model is not asked: the fix is proved by the test, so without one there is nothing to prove it by.
func TestAFindingWithoutAProvingTestGetsNoFix(t *testing.T) {
	for name, mutate := range map[string]func(*candidate.Candidate){
		"query evidence only": func(c *candidate.Candidate) {
			c.Evidence = []candidate.Evidence{{Kind: candidate.KindQuery, Query: &candidate.QueryEvidence{}}}
		},
		"test that passed on both": func(c *candidate.Candidate) { c.Evidence[0].Test.FailsOnChange = false },
		"test that failed on base": func(c *candidate.Candidate) { c.Evidence[0].Test.PassesOnBase = false },
		"no evidence":              func(c *candidate.Candidate) { c.Evidence = nil },
	} {
		t.Run(name, func(t *testing.T) {
			c := finding()
			mutate(&c)
			m := &scriptedModel{}
			got, err := Propose(context.Background(), c, m, &scriptedRunner{}, Options{})
			if err != nil || got.Fix != nil || len(m.prompts) != 0 {
				t.Fatalf("%#v %v asked %d", got.Fix, err, len(m.prompts))
			}
		})
	}
}

// TestAnAnswerWithACommandIsRefusedWhole proves ADR-0025 holds for a fix: a model's command is never
// handed to the runner, and the attempt counts as failed with the refusal as its reason.
func TestAnAnswerWithACommandIsRefusedWhole(t *testing.T) {
	m := &scriptedModel{replies: []string{`{"what":"x","why":"y","command":"rm -rf /","edits":[{"path":"billing/price.go","old":"a","new":"b"}]}`}}
	r := &scriptedRunner{}
	got, err := Propose(context.Background(), finding(), m, r, Options{Attempts: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.requests) != 0 || got.Fix == nil || got.Fix.Proved || !strings.Contains(got.Fix.Note, testrun.ReasonModelCommand) {
		t.Fatalf("%#v requests %d", got.Fix, len(r.requests))
	}
}

// TestARefusedPatchIsRetriedWithTheReason proves a patch the runner refuses (an edit that does not
// apply) is an attempt like any other: the next prompt says why it was refused.
func TestARefusedPatchIsRetriedWithTheReason(t *testing.T) {
	m := &scriptedModel{replies: []string{answerOne, answerTwo}}
	r := &scriptedRunner{outcomes: []testrun.FixOutcome{{Refused: testrun.ReasonEditNotFound}, good()}}
	got, err := Propose(context.Background(), finding(), m, r, Options{Attempts: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got.Fix == nil || !got.Fix.Proved || !strings.Contains(m.prompts[1], testrun.ReasonEditNotFound) {
		t.Fatalf("%#v\n%s", got.Fix, m.prompts[1])
	}
}

// TestASampleWithNoRecordedAnswerMakesNoClaim proves a generator that has no answer leaves the finding
// as it was on the first attempt, and ends a retry as not proved on a later one.
func TestASampleWithNoRecordedAnswerMakesNoClaim(t *testing.T) {
	got, err := Propose(context.Background(), finding(), &scriptedModel{}, &scriptedRunner{}, Options{})
	if err != nil || got.Fix != nil {
		t.Fatalf("%#v %v", got.Fix, err)
	}
	still := withRun(good(), run(testrun.RunPatched, testrun.ResultFailed, 1, "FAIL"))
	m := &scriptedModel{replies: []string{answerOne}}
	r := &scriptedRunner{outcomes: []testrun.FixOutcome{still}}
	got, err = Propose(context.Background(), finding(), m, r, Options{Attempts: 2})
	if err != nil || got.Fix == nil || got.Fix.Proved || !strings.Contains(got.Fix.Note, "no answer for attempt 2") {
		t.Fatalf("%#v %v", got.Fix, err)
	}
}

// TestADownRunnerIsReturnedAndLeavesTheFindingAsItWas proves the Prover being off is not a failed
// fix: the error comes back for the review to name the service, and no fix is claimed.
func TestADownRunnerIsReturnedAndLeavesTheFindingAsItWas(t *testing.T) {
	down := &service.Down{Service: "prover", Reason: "Prover off"}
	m := &scriptedModel{replies: []string{answerOne}}
	got, err := Propose(context.Background(), finding(), m, &scriptedRunner{err: down}, Options{})
	var d *service.Down
	if !errors.As(err, &d) || got.Fix != nil {
		t.Fatalf("%v %#v", err, got.Fix)
	}
}

// TestARunThatShowedNoPassIsNotAProvedFix proves the property of ADR-0042 that a fix is shown as
// proved only when the proving test ran and passed on every run, end to end through the
// sample runner: runs recorded as passed (exit code 0) whose output has no pass line do not
// prove the fix. The mutated field is the output of the recorded runs, and nothing else.
func TestARunThatShowedNoPassIsNotAProvedFix(t *testing.T) {
	change := t.TempDir()
	if err := os.MkdirAll(filepath.Join(change, "billing"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(change, "billing", "price.go"), []byte("package billing\n\nfunc Price() int { return 60 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	recorded := func(passLine string) *testrun.Sample {
		runs := good().Runs
		for i := range runs {
			if runs[i].Result == testrun.ResultPassed {
				runs[i].Output = passLine
			}
		}
		return testrun.NewSample(nil, testrun.RecordedFix{FindingID: "F1", Attempt: 1, Runs: runs})
	}
	for _, tc := range []struct {
		name, output string
		proved       bool
	}{
		{"pass lines", "--- PASS: TestPrice (0.00s)\nPASS\nok  \texample.com/p\t0.001s", true},
		{"exit 0 and an ok line only", "ok  \texample.com/p\t0.001s", false},
		{"exit 0 and no output", "", false},
	} {
		m := &scriptedModel{replies: []string{answerOne}}
		got, err := Propose(context.Background(), finding(), m, recorded(tc.output), Options{BaseDir: "b", ChangeDir: change, Attempts: 1})
		if err != nil || got.Fix == nil {
			t.Fatalf("%s: %#v %v", tc.name, got.Fix, err)
		}
		if got.Fix.Proved != tc.proved {
			t.Errorf("%s: proved=%v, want %v", tc.name, got.Fix.Proved, tc.proved)
		}
		if !tc.proved && len(got.Fix.Edits) != 0 {
			t.Errorf("%s: a fix that was not proved keeps its edits", tc.name)
		}
	}
}
