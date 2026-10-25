package tool_test

import (
	"context"
	"strings"
	"testing"

	eg "github.com/saphan-ai/proven/examples/tool"
	itt "github.com/saphan-ai/proven/internal/testrun"
	it "github.com/saphan-ai/proven/internal/tool"
)

// The assignment proves Detector compiles against the real interface:
// examples/tool.Detector and internal/tool.Adapter agree on one method set.
var _ it.Adapter = eg.Detector{}

// cannedHost returns the change side on the first call and the base side on
// the second, exactly as the runner runs them.
type cannedHost struct {
	calls  int
	change it.Side
	base   it.Side
}

func (h *cannedHost) Run(ctx context.Context, dir string, argv []string) (it.Side, error) {
	h.calls++
	if h.calls == 1 {
		return h.change, nil
	}
	return h.base, nil
}

// TestRunProvesTheChange proves one answer: a banner only on the change is
// evidence, so Proves is true.
func TestRunProvesTheChange(t *testing.T) {
	host := &cannedHost{
		change: it.Side{Output: "WARNING: DATA RACE\nFAIL\n", Status: ""},
		base:   it.Side{Output: "ok   example/counter\t0.012s\n", Status: ""},
	}
	runner := it.NewRunner(eg.Detector{}, host)
	out, err := runner.Run(context.Background(), it.Request{
		Paths:     []string{"counter/counter.go"},
		BaseDir:   "/base",
		ChangeDir: "/change",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !out.Answered || out.Tool != "example-detector" {
		t.Fatalf("got %+v, want an answered run of example-detector", out)
	}
	if !out.PresentOnChange || out.PresentOnBase {
		t.Fatalf("got change=%v base=%v, want change only", out.PresentOnChange, out.PresentOnBase)
	}
	if !out.Proves() {
		t.Fatal("got no evidence, want a result introduced by the change")
	}
	if out.Command != "go test -race ./counter" {
		t.Fatalf("got command %q, want go test -race ./counter", out.Command)
	}
	if host.calls != 2 {
		t.Fatalf("got %d host calls, want two (change then base)", host.calls)
	}
}

// TestRunDoesNotProveWhatTheBaseAlreadyHad proves the refusal of evidence:
// the same banner on the base is not a finding of the change.
func TestRunDoesNotProveWhatTheBaseAlreadyHad(t *testing.T) {
	host := &cannedHost{
		change: it.Side{Output: "WARNING: DATA RACE\nFAIL\n", Status: ""},
		base:   it.Side{Output: "WARNING: DATA RACE\nFAIL\n", Status: ""},
	}
	runner := it.NewRunner(eg.Detector{}, host)
	out, err := runner.Run(context.Background(), it.Request{
		Paths:     []string{"counter/counter.go"},
		BaseDir:   "/base",
		ChangeDir: "/change",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.Proves() {
		t.Fatal("a result on both sides is not evidence of the change")
	}
}

// TestRunNamesANonAnswer proves the interface's error path: a side that did
// not answer is not a clean run, and the outcome names why.
func TestRunNamesANonAnswer(t *testing.T) {
	host := &cannedHost{
		change: it.Side{Output: "WARNING: DATA RACE\nFAIL\n", Status: ""},
		base:   it.Side{Output: "", Status: itt.CommandTimeout},
	}
	runner := it.NewRunner(eg.Detector{}, host)
	out, err := runner.Run(context.Background(), it.Request{
		Paths:     []string{"counter/counter.go"},
		BaseDir:   "/base",
		ChangeDir: "/change",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.Answered {
		t.Fatal("a timed-out side must not count as answered")
	}
	if out.Reason == "" || !strings.Contains(out.Reason, "timeout") {
		t.Fatalf("got reason %q, want a named timeout", out.Reason)
	}
}

// TestRunSkipsNonGoFiles proves a change with no Go file is not run at all.
func TestRunSkipsNonGoFiles(t *testing.T) {
	host := &cannedHost{}
	runner := it.NewRunner(eg.Detector{}, host)
	out, err := runner.Run(context.Background(), it.Request{
		Paths:     []string{"README.md"},
		BaseDir:   "/base",
		ChangeDir: "/change",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !out.Skipped || out.Reason == "" {
		t.Fatalf("got %+v, want skipped with a reason", out)
	}
	if host.calls != 0 {
		t.Fatalf("got %d host calls, want none", host.calls)
	}
}
