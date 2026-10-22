package prover

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/testrun"
)

const writtenBody = "package billing\n\nimport \"testing\"\n\nfunc TestExtra(t *testing.T) {\n\tif Discounted(100, 20) != 80 {\n\t\tt.Fatal(Discounted(100, 20))\n\t}\n}\n"

func TestWrittenTestThatFailsOnChangeIsEvidence(t *testing.T) {
	run := &fakeRun{}
	run.when = func(req testrun.Request) testrun.Outcome {
		if req.Content == "" {
			return testrun.Outcome{Found: false}
		}
		return testrun.Outcome{
			Found: true, FailsOnChange: true, PassesOnBase: true,
			Path: req.Path, Name: req.Name, Command: req.Command, Output: "change: FAIL\nbase: ok",
		}
	}
	got := proveWrittenCase(t, run, "billing/extra_test.go", writtenBody, "TestExtra", "go test ./billing -run TestExtra", Options{})
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

func TestPassesOnBothWritesNoEvidence(t *testing.T) {
	run := &fakeRun{out: testrun.Outcome{Found: true, FailsOnChange: false, PassesOnBase: true}}
	got := proveWrittenCase(t, run, "billing/extra_test.go", writtenBody, "TestExtra", "go test ./billing -run TestExtra", Options{})
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

func TestWrittenTestThatFailsOnTheBaseWritesNoEvidence(t *testing.T) {
	run := &fakeRun{out: testrun.Outcome{Found: true, FailsOnChange: true, PassesOnBase: false}}
	got := proveWrittenCase(t, run, "billing/extra_test.go", writtenBody, "TestExtra", "go test ./billing -run TestExtra", Options{})
	if len(got.Evidence) != 0 {
		t.Fatalf("evidence %#v", got.Evidence)
	}
	if got.ProverNote != noteFailedOnBase {
		t.Fatalf("note %q", got.ProverNote)
	}
}

func TestNamedTestWithoutContentStaysTheFallback(t *testing.T) {
	run := &fakeRun{out: testrun.Outcome{
		Found: true, FailsOnChange: true, PassesOnBase: true,
		Path: "billing/discount_test.go", Name: "TestDiscounted", Command: "go test ./billing -run TestDiscounted",
	}}
	got, err := Prove(context.Background(), candidate.Candidate{
		Risk: candidate.Risk{ID: "F1", Title: "Discount is applied twice"},
	}, scriptGen{text: `{"path":"billing/discount_test.go","name":"TestDiscounted","command":"go test ./billing -run TestDiscounted"}`}, run, Options{})
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

func TestRefusesAbsolutePath(t *testing.T) {
	assertRefused(t, "/tmp/extra_test.go", writtenBody, "go test", Options{}, "refused: "+testrun.ReasonAbsolutePath)
}

func TestRefusesPathThatLeavesTheRepository(t *testing.T) {
	assertRefused(t, "../extra_test.go", writtenBody, "go test", Options{}, "refused: "+testrun.ReasonLeavesRepo)
}

func TestRefusesAPathThatIsNotATestFile(t *testing.T) {
	assertRefused(t, "billing/discount.go", writtenBody, "go test", Options{}, "refused: "+testrun.ReasonNotATestFile)
}

func TestRefusesOverwritingANonTestFile(t *testing.T) {
	base := t.TempDir()
	change := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "billing"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "billing", "discount.go"), []byte("package billing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	assertRefused(t, "billing/discount.go", writtenBody, "go test", Options{BaseDir: base, ChangeDir: change}, "refused: "+testrun.ReasonOverwrites)
}

func TestRefusesContentOverTheSizeLimit(t *testing.T) {
	opts := Options{MaxWrittenBytes: 8}
	assertRefused(t, "billing/extra_test.go", "package p\n", "go test", opts, "refused: "+testrun.ReasonTooLarge)
}

func TestContentAtTheSizeLimitRuns(t *testing.T) {
	body := "12345678"
	run := &fakeRun{out: testrun.Outcome{Found: true, FailsOnChange: true, PassesOnBase: true, Name: "TestExtra"}}
	got := proveWrittenCase(t, run, "billing/extra_test.go", body, "TestExtra", "go test", Options{MaxWrittenBytes: len(body)})
	if run.calls != 1 || len(got.Evidence) != 1 {
		t.Fatalf("calls %d evidence %#v note %q", run.calls, got.Evidence, got.ProverNote)
	}
}

func TestRefusesEmptyCommand(t *testing.T) {
	assertRefused(t, "billing/extra_test.go", writtenBody, "   ", Options{}, "refused: "+testrun.ReasonEmptyCommand)
}

func TestAllowsReplacingATestFile(t *testing.T) {
	base := t.TempDir()
	change := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "extra_test.go"), []byte("package billing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run := &fakeRun{out: testrun.Outcome{Found: true, FailsOnChange: true, PassesOnBase: true, Name: "TestExtra"}}
	got := proveWrittenCase(t, run, "extra_test.go", writtenBody, "TestExtra", "go test", Options{BaseDir: base, ChangeDir: change})
	if run.calls != 1 || len(got.Evidence) != 1 || got.ProverNote != "" {
		t.Fatalf("calls %d evidence %#v note %q", run.calls, got.Evidence, got.ProverNote)
	}
}

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

func proveWrittenCase(t *testing.T, run *fakeRun, path, content, name, command string, opts Options) candidate.Candidate {
	t.Helper()
	raw, err := json.Marshal(map[string]string{
		"path": path, "content": content, "name": name, "command": command,
	})
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

type scriptGen struct{ text string }

func (s scriptGen) Complete(context.Context, string, string) (string, error) {
	return s.text, nil
}

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
	if out.Command == "" {
		out.Command = req.Command
	}
	return out, nil
}

func TestPromptAsksForTheFile(t *testing.T) {
	p := Prompt("F1", "Discount is applied twice")
	for _, want := range []string{"content (string)", "does not attack anything", "Leave content empty"} {
		if !strings.Contains(p, want) {
			t.Fatalf("prompt missing %q", want)
		}
	}
}
