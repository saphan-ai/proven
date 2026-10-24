package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/testrun"
)

// fixSpy is the sample test runner that counts the proofs of fixes it is asked for.
type fixSpy struct {
	*testrun.Sample
	mu       sync.Mutex
	requests []testrun.FixRequest
	// refutingTest, when set, is a test name that passes on the change and on the base.
	refutingTest string
}

func (s *fixSpy) RunFix(ctx context.Context, req testrun.FixRequest) (testrun.FixOutcome, error) {
	s.mu.Lock()
	s.requests = append(s.requests, req)
	s.mu.Unlock()
	return s.Sample.RunFix(ctx, req)
}

// fixerCounter counts the calls the Fixer makes to the generator. A non-empty refuteF1 replaces the
// Skeptic's answer for F1, so a test can make the Judge drop a finding that a test proved.
type fixerCounter struct {
	inner     generator.Generator
	calls     int
	refuteF1  string
	passingOn string
}

func (c *fixerCounter) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if agent == "fixer" {
		c.calls++
	}
	if agent == "skeptic" && c.refuteF1 != "" && strings.Contains(prompt, `"id":"F1"`) {
		return c.refuteF1, nil
	}
	return c.inner.Complete(ctx, agent, prompt)
}

// Run answers the refuting test the Skeptic named as one that passes on the change, and
// asks the sample for every other test.
func (s *fixSpy) Run(ctx context.Context, req testrun.Request) (testrun.Outcome, error) {
	if s.refutingTest != "" && req.Name == s.refutingTest {
		return testrun.Outcome{Found: true, Path: req.Path, Name: req.Name, PassesOnBase: true}, nil
	}
	return s.Sample.Run(ctx, req)
}

// reviewFixture reviews a sample fixture with the fix settings given and returns the draft.
func reviewFixture(t *testing.T, name string, spy *fixSpy, gen *fixerCounter, req Request, redactor *redact.Redactor) candidate.Draft {
	t.Helper()
	return reviewFixtureGuarded(t, name, spy, gen, req, redactor, nil)
}

// reviewFixtureGuarded is reviewFixture with a guardrail adapter standing at every model call.
// A nil adapter leaves the sample's own, which answers safe.
func reviewFixtureGuarded(t *testing.T, name string, spy *fixSpy, gen *fixerCounter, req Request, redactor *redact.Redactor, rail guardrail.Adapter) candidate.Draft {
	t.Helper()
	fx, err := fixture.Load(filepath.Join(repoRoot(t), "testdata", "fixture", name))
	if err != nil {
		t.Fatal(err)
	}
	spy.Sample = testrun.NewSample(fx.Tests, fx.Fixes...)
	inner, err := generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	gen.inner = inner
	p, err := Build(Options{
		Sample: true, LogDir: t.TempDir(), TestRun: spy, Generator: gen, Redactor: redactor, Guardrail: rail,
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req.Target = fx.Root
	d, err := p.Review(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// TestEachProvedFindingGetsItsFixAttemptsInOrder proves the fixer step: F1 is proved on the first
// attempt, F2 on neither, and the runner was asked for exactly those three proofs with the attempt
// number of each. This is the control for the settings tests below, which change one field.
func TestEachProvedFindingGetsItsFixAttemptsInOrder(t *testing.T) {
	spy, gen := &fixSpy{}, &fixerCounter{}
	d := reviewFixture(t, "fix", spy, gen, Request{}, nil)
	f1, ok1 := kept(d, "F1")
	f2, ok2 := kept(d, "F2")
	if !ok1 || !ok2 || f1.Fix == nil || f2.Fix == nil {
		t.Fatalf("%#v %#v", f1.Fix, f2.Fix)
	}
	if !f1.Fix.Proved || f1.Fix.Attempts != 1 || f2.Fix.Proved || f2.Fix.Attempts != 2 || len(f2.Fix.Edits) != 0 {
		t.Fatalf("F1 %#v F2 %#v", f1.Fix, f2.Fix)
	}
	var got []string
	for _, r := range spy.requests {
		got = append(got, r.FindingID+"/"+string(rune('0'+r.Attempt)))
	}
	if strings.Join(got, " ") != "F1/1 F2/1 F2/2" {
		t.Fatalf("proofs asked for: %v", got)
	}
}

// TestFixesOffAsksNeitherTheModelNorTheRunner proves the setting that turns fixes off (ADR-0042): no
// model call for a fix, no proof run, no Fix on any finding, with the findings themselves unchanged.
func TestFixesOffAsksNeitherTheModelNorTheRunner(t *testing.T) {
	spy, gen := &fixSpy{}, &fixerCounter{}
	d := reviewFixture(t, "fix", spy, gen, Request{FixesOff: true}, nil)
	if gen.calls != 0 || len(spy.requests) != 0 {
		t.Fatalf("model calls %d, proofs %d", gen.calls, len(spy.requests))
	}
	for _, c := range d.Findings {
		if c.Fix != nil {
			t.Fatalf("%s has a fix with fixes off", c.ID)
		}
	}
	if _, ok := kept(d, "F1"); !ok {
		t.Fatal("F1 is gone")
	}
}

// TestFixAttemptsBoundsTheModelCalls proves the attempts setting reaches the Fixer: with one attempt
// F2 is asked once.
func TestFixAttemptsBoundsTheModelCalls(t *testing.T) {
	spy, gen := &fixSpy{}, &fixerCounter{}
	reviewFixture(t, "fix", spy, gen, Request{FixAttempts: 1}, nil)
	if gen.calls != 2 || len(spy.requests) != 2 {
		t.Fatalf("model calls %d, proofs %d; want one for F1 and one for F2", gen.calls, len(spy.requests))
	}
}

// TestADroppedFindingGetsNoFix proves a finding the Judge dropped is not fixed. F1 of the fix scenario
// has a proving test; the Skeptic's answer is replaced by a refutation with a test that passes on the
// change, so the Judge drops F1, and the Fixer is asked about F2 only. The control is the first test above,
// where F1 is kept and fixed.
func TestADroppedFindingGetsNoFix(t *testing.T) {
	refute := `{"refuted":true,"reason":"the test passes","path":"billing/discount_test.go","name":"TestRefute","template":"","params":{}}`
	spy, gen := &fixSpy{refutingTest: "TestRefute"}, &fixerCounter{refuteF1: refute}
	d := reviewFixture(t, "fix", spy, gen, Request{}, nil)
	if _, ok := kept(d, "F1"); ok {
		t.Fatal("F1 was not dropped, so this test does not test what it says")
	}
	for _, r := range spy.requests {
		if r.FindingID == "F1" {
			t.Fatalf("a dropped finding was sent to the proof: %#v", r)
		}
	}
	if _, ok := kept(d, "F2"); !ok || gen.calls != 2 {
		t.Fatalf("F2 should be fixed alone: %d model calls", gen.calls)
	}
}

// TestAFixIsNeverWrittenToTheChange proves ADR-0042 point 4: the fix is a suggestion. Every file of the
// change and base trees is the same after the review as before it.
func TestAFixIsNeverWrittenToTheChange(t *testing.T) {
	root := filepath.Join(repoRoot(t), "testdata", "fixture", "fix")
	before := snapshotTree(t, root)
	spy, gen := &fixSpy{}, &fixerCounter{}
	reviewFixture(t, "fix", spy, gen, Request{}, nil)
	after := snapshotTree(t, root)
	if len(before) == 0 || len(before) != len(after) {
		t.Fatalf("%d files before, %d after", len(before), len(after))
	}
	for path, text := range before {
		if after[path] != text {
			t.Fatalf("%s was changed by the review", path)
		}
	}
}

// TestAFixIsRedactedLikeTheRestOfTheFinding proves the last place model text and test output reach the
// result is redacted: a configured value found in the model's explanation, in a patch and in a run's
// output does not reach the draft.
func TestAFixIsRedactedLikeTheRestOfTheFinding(t *testing.T) {
	secret := "percent off twice"
	rd := redact.New(map[string]string{"PROVEN_TEST_SECRET": secret})
	spy, gen := &fixSpy{}, &fixerCounter{}
	d := reviewFixture(t, "fix", spy, gen, Request{}, rd)
	f1, _ := kept(d, "F1")
	if f1.Fix == nil || strings.Contains(f1.Fix.Why, secret) || !strings.Contains(f1.Fix.Why, "[redacted: PROVEN_TEST_SECRET]") {
		t.Fatalf("%#v", f1.Fix)
	}
}

// snapshotTree reads every regular file under root into a map by relative path.
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.Mode().IsRegular() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[path] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// fixerRail is a guardrail that judges only the Fixer unsafe: the prompt side when onPrompt is set,
// the answer side otherwise. It tells the Fixer from the other agents by the text of its prompt.
type fixerRail struct {
	onPrompt bool
	off      bool
	checks   int
}

func (r *fixerRail) Check(_ context.Context, prompt, answer string) (guardrail.Verdict, error) {
	r.checks++
	if r.off || !strings.Contains(prompt, "Task: Write a fix for the change") {
		return guardrail.Verdict{User: "safe"}, nil
	}
	if r.onPrompt && answer == "" {
		return guardrail.Verdict{User: "unsafe", Categories: []string{"Other"}}, nil
	}
	if !r.onPrompt && answer != "" {
		return guardrail.Verdict{User: "safe", Response: "unsafe", Categories: []string{"Other"}}, nil
	}
	return guardrail.Verdict{User: "safe"}, nil
}

// TestTheGuardrailCoversTheFixer proves ADR-0005 holds for the Fixer: every model call of the fix step
// goes through the guardrail. A prompt it refuses never reaches the model, an answer it refuses is
// not used, and in both cases the finding shows no fix and no proof is run. Re-wiring the fix step to
// the unguarded generator makes both cases red. The control is the same review with a guardrail that
// judges nothing unsafe, which shows the proved fix.
func TestTheGuardrailCoversTheFixer(t *testing.T) {
	spy, gen := &fixSpy{}, &fixerCounter{}
	d := reviewFixtureGuarded(t, "fix", spy, gen, Request{}, nil, &fixerRail{off: true})
	if f1, ok := kept(d, "F1"); !ok || f1.Fix == nil || !f1.Fix.Proved {
		t.Fatalf("control: a guardrail that refuses nothing hides the fix: %#v", f1.Fix)
	}
	for _, tc := range []struct {
		name       string
		onPrompt   bool
		modelCalls int
	}{
		{"prompt refused", true, 0},
		{"answer refused", false, -1},
	} {
		spy, gen := &fixSpy{}, &fixerCounter{}
		d := reviewFixtureGuarded(t, "fix", spy, gen, Request{}, nil, &fixerRail{onPrompt: tc.onPrompt})
		for _, id := range []string{"F1", "F2"} {
			c, ok := kept(d, id)
			if !ok {
				t.Fatalf("%s: %s is missing", tc.name, id)
			}
			if c.Fix != nil && (c.Fix.Proved || len(c.Fix.Edits) != 0) {
				t.Errorf("%s: %s shows a fix the guardrail refused: %#v", tc.name, id, c.Fix)
			}
		}
		if len(spy.requests) != 0 {
			t.Errorf("%s: %d proofs ran for a fix the guardrail refused", tc.name, len(spy.requests))
		}
		if tc.modelCalls == 0 && gen.calls != 0 {
			t.Errorf("%s: a refused prompt reached the model %d times", tc.name, gen.calls)
		}
	}
}
