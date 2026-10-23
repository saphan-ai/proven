package prover

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

// writtenBody is a small Go test file the model might write.
const writtenBody = "package billing\n\nimport \"testing\"\n\nfunc TestExtra(t *testing.T) {\n\tif Discounted(100, 20) != 80 {\n\t\tt.Fatal(Discounted(100, 20))\n\t}\n}\n"

// A written test that fails on the change and passes on the base becomes test evidence.
func TestWrittenTestThatFailsOnChangeIsEvidence(t *testing.T) {
	run := &fakeRun{}
	run.when = func(req testrun.Request) testrun.Outcome {
		if req.Content == "" {
			return testrun.Outcome{Found: false}
		}
		return testrun.Outcome{
			Found: true, FailsOnChange: true, PassesOnBase: true,
			Path: req.Path, Name: req.Name, Command: "go test -run '^TestExtra$' ./billing", Output: "change: FAIL\nbase: ok",
		}
	}
	got := proveWrittenCase(t, run, "billing/extra_test.go", writtenBody, "TestExtra", "", Options{})
	if run.calls != 1 {
		t.Fatalf("calls %d", run.calls)
	}
	if run.last.Content != writtenBody || run.last.Path != "billing/extra_test.go" {
		t.Fatalf("request %#v", run.last)
	}
	if len(got.Evidence) != 1 || got.Evidence[0].Test == nil {
		t.Fatalf("evidence %#v", got.Evidence)
	}
	te := got.Evidence[0].Test
	if !te.FailsOnChange || !te.PassesOnBase || te.Name != "TestExtra" {
		t.Fatalf("%#v", te)
	}
	if got.ProverNote != "" {
		t.Fatal(got.ProverNote)
	}
}

// A written test that passes on both checkouts is not evidence and leaves a note.
func TestPassesOnBothWritesNoEvidence(t *testing.T) {
	run := &fakeRun{out: testrun.Outcome{Found: true, FailsOnChange: false, PassesOnBase: true}}
	got := proveWrittenCase(t, run, "billing/extra_test.go", writtenBody, "TestExtra", "", Options{})
	if len(got.Evidence) != 0 {
		t.Fatalf("evidence %#v", got.Evidence)
	}
	if got.ProverNote != notePassedOnChange {
		t.Fatalf("note %q", got.ProverNote)
	}
	if run.calls != 1 {
		t.Fatalf("calls %d", run.calls)
	}
}

// A written test that fails on the base is not evidence and leaves a note.
func TestWrittenTestThatFailsOnTheBaseWritesNoEvidence(t *testing.T) {
	run := &fakeRun{out: testrun.Outcome{Found: true, FailsOnChange: true, PassesOnBase: false}}
	got := proveWrittenCase(t, run, "billing/extra_test.go", writtenBody, "TestExtra", "", Options{})
	if len(got.Evidence) != 0 {
		t.Fatalf("evidence %#v", got.Evidence)
	}
	if got.ProverNote != noteFailedOnBase {
		t.Fatalf("note %q", got.ProverNote)
	}
}

// A reply with no content runs the named existing test, and writes no file.
func TestNamedTestWithoutContentStaysTheFallback(t *testing.T) {
	run := &fakeRun{out: testrun.Outcome{
		Found: true, FailsOnChange: true, PassesOnBase: true,
		Path: "billing/discount_test.go", Name: "TestDiscounted", Command: "go test -run '^TestDiscounted$' ./billing",
	}}
	got, err := Prove(context.Background(), candidate.Candidate{
		Risk: candidate.Risk{ID: "F1", Title: "Discount is applied twice"},
	}, scriptGen{text: `{"path":"billing/discount_test.go","name":"TestDiscounted"}`}, run, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if run.calls != 1 || run.last.Content != "" {
		t.Fatalf("calls %d content %q", run.calls, run.last.Content)
	}
	if len(got.Evidence) != 1 || got.ProverNote != "" {
		t.Fatalf("evidence %#v note %q", got.Evidence, got.ProverNote)
	}
}

// An empty reply calls no runner and leaves no evidence and no note.
func TestEmptyAnswerProposesNoTest(t *testing.T) {
	run := &fakeRun{}
	got, err := Prove(context.Background(), candidate.Candidate{
		Risk: candidate.Risk{ID: "F2", Title: "no test"},
	}, scriptGen{text: `{"path":"","content":"","name":"","command":""}`}, run, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if run.calls != 0 || len(got.Evidence) != 0 || got.ProverNote != "" {
		t.Fatalf("calls %d evidence %#v note %q", run.calls, got.Evidence, got.ProverNote)
	}
}

// A written test at an absolute path is refused before the runner.
func TestRefusesAbsolutePath(t *testing.T) {
	assertRefused(t, "/tmp/extra_test.go", writtenBody, "", Options{}, "refused: "+testrun.ReasonAbsolutePath)
}

// A written test whose path leaves the repository is refused.
func TestRefusesPathThatLeavesTheRepository(t *testing.T) {
	assertRefused(t, "../extra_test.go", writtenBody, "", Options{}, "refused: "+testrun.ReasonLeavesRepo)
}

// A written file that is not a test file is refused.
func TestRefusesAPathThatIsNotATestFile(t *testing.T) {
	assertRefused(t, "billing/discount.go", writtenBody, "", Options{}, "refused: "+testrun.ReasonNotATestFile)
}

// A written test may not overwrite a non-test file of the checkout.
func TestRefusesOverwritingANonTestFile(t *testing.T) {
	base := t.TempDir()
	change := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "billing"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "billing", "discount.go"), []byte("package billing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	assertRefused(t, "billing/discount.go", writtenBody, "", Options{BaseDir: base, ChangeDir: change}, "refused: "+testrun.ReasonOverwrites)
}

// Content over the size limit is refused.
func TestRefusesContentOverTheSizeLimit(t *testing.T) {
	opts := Options{MaxWrittenBytes: 8}
	assertRefused(t, "billing/extra_test.go", "package p\n", "", opts, "refused: "+testrun.ReasonTooLarge)
}

// Content exactly at the size limit runs: the limit is inclusive.
func TestContentAtTheSizeLimitRuns(t *testing.T) {
	body := "12345678"
	run := &fakeRun{out: testrun.Outcome{Found: true, FailsOnChange: true, PassesOnBase: true, Name: "TestExtra"}}
	got := proveWrittenCase(t, run, "billing/extra_test.go", body, "TestExtra", "", Options{MaxWrittenBytes: len(body)})
	if run.calls != 1 || len(got.Evidence) != 1 {
		t.Fatalf("calls %d evidence %#v note %q", run.calls, got.Evidence, got.ProverNote)
	}
}

// A name that is not a test name is refused before the runner is called.
func TestRefusesAWrittenNameThatIsNotATestName(t *testing.T) {
	run := &fakeRun{out: testrun.Outcome{Found: true, FailsOnChange: true, PassesOnBase: true}}
	got := proveWrittenCase(t, run, "billing/extra_test.go", writtenBody, "TestExtra; id", "", Options{})
	if run.calls != 0 || len(got.Evidence) != 0 || got.ProverNote != "refused: "+testrun.ReasonBadName {
		t.Fatalf("calls %d evidence %#v note %q", run.calls, got.Evidence, got.ProverNote)
	}
}

// A written test may replace an existing test file.
func TestAllowsReplacingATestFile(t *testing.T) {
	base := t.TempDir()
	change := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "extra_test.go"), []byte("package billing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run := &fakeRun{out: testrun.Outcome{Found: true, FailsOnChange: true, PassesOnBase: true, Name: "TestExtra"}}
	got := proveWrittenCase(t, run, "extra_test.go", writtenBody, "TestExtra", "", Options{BaseDir: base, ChangeDir: change})
	if run.calls != 1 || len(got.Evidence) != 1 || got.ProverNote != "" {
		t.Fatalf("calls %d evidence %#v note %q", run.calls, got.Evidence, got.ProverNote)
	}
}

// assertRefused checks that a written test is refused with note, before any run and with no evidence.
func assertRefused(t *testing.T, path, content, command string, opts Options, note string) {
	t.Helper()
	run := &fakeRun{out: testrun.Outcome{Found: true, FailsOnChange: true, PassesOnBase: true}}
	got := proveWrittenCase(t, run, path, content, "TestExtra", command, opts)
	if run.calls != 0 {
		t.Fatalf("runner was called for %s", note)
	}
	if len(got.Evidence) != 0 {
		t.Fatalf("evidence %#v", got.Evidence)
	}
	if got.ProverNote != note {
		t.Fatalf("note %q want %q", got.ProverNote, note)
	}
}

// proveWrittenCase runs Prove with a model reply that writes this test.
// A non-empty command is added to the reply, as a model that disobeys the prompt would.
func proveWrittenCase(t *testing.T, run *fakeRun, path, content, name, command string, opts Options) candidate.Candidate {
	t.Helper()
	fields := map[string]string{"path": path, "content": content, "name": name}
	if command != "" {
		fields["command"] = command
	}
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Prove(context.Background(), candidate.Candidate{
		Risk: candidate.Risk{ID: "F9", Title: "Discount is applied twice"},
	}, scriptGen{text: string(raw)}, run, opts)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// scriptGen is a generator that always returns text.
type scriptGen struct{ text string }

func (s scriptGen) Complete(context.Context, string, string) (string, error) {
	return s.text, nil
}

// fakeRun is a test runner that records its requests and returns a fixed or computed outcome.
type fakeRun struct {
	calls int
	last  testrun.Request
	out   testrun.Outcome
	when  func(testrun.Request) testrun.Outcome
}

func (f *fakeRun) Run(_ context.Context, req testrun.Request) (testrun.Outcome, error) {
	f.calls++
	f.last = req
	if f.when != nil {
		return f.when(req), nil
	}
	out := f.out
	if out.Path == "" {
		out.Path = req.Path
	}
	if out.Name == "" {
		out.Name = req.Name
	}
	return out, nil
}

// TestNilRunnerIsANamedRefusal proves a nil test runner is the same named
// refusal Execute returns, for a named test and for a written test (F3).
func TestNilRunnerIsANamedRefusal(t *testing.T) {
	t.Run("named", func(t *testing.T) {
		_, err := Prove(context.Background(), candidate.Candidate{
			Risk: candidate.Risk{ID: "F1", Title: "Discount is applied twice"},
		}, scriptGen{text: `{"path":"billing/discount_test.go","name":"TestDiscounted"}`}, nil, Options{})
		assertProverDown(t, err)
	})
	t.Run("written", func(t *testing.T) {
		raw, err := json.Marshal(map[string]string{
			"path": "billing/extra_test.go", "content": writtenBody, "name": "TestExtra",
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = Prove(context.Background(), candidate.Candidate{
			Risk: candidate.Risk{ID: "F9", Title: "Discount is applied twice"},
		}, scriptGen{text: string(raw)}, nil, Options{})
		assertProverDown(t, err)
	})
}

// assertProverDown fails the test unless err is the named missing-runner refusal.
func assertProverDown(t *testing.T, err error) {
	t.Helper()
	d, ok := service.As(err)
	if !ok || d.Service != "prover" || d.Reason != "no test runner" {
		t.Fatalf("err %v", err)
	}
}

// TestWrittenTestNotFoundSaysNotFound proves a written test the runner did
// not find is noted as not found, even when the outcome also says it passed
// on the base (F4).
func TestWrittenTestNotFoundSaysNotFound(t *testing.T) {
	run := &fakeRun{out: testrun.Outcome{Found: false, PassesOnBase: true}}
	got := proveWrittenCase(t, run, "billing/extra_test.go", writtenBody, "TestExtra", "", Options{})
	if len(got.Evidence) != 0 {
		t.Fatalf("evidence %#v", got.Evidence)
	}
	if !strings.Contains(got.ProverNote, "not found") {
		t.Fatalf("note %q", got.ProverNote)
	}
}

// The prompt asks for the file's content, forbids an attack (ADR-0007), and allows naming an existing test.
func TestPromptAsksForTheFile(t *testing.T) {
	p := Prompt("F1", "Discount is applied twice")
	for _, want := range []string{"content (string)", "does not attack anything", "Leave content empty"} {
		if !strings.Contains(p, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}
