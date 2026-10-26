package tool

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/testrun"
)

// fixedHost answers every command with one side, or one error, and keeps what it was asked.
type fixedHost struct {
	side Side
	err  error
	dirs []string
	argv [][]string
}

func (h *fixedHost) Run(_ context.Context, dir string, argv []string) (Side, error) {
	h.dirs = append(h.dirs, dir)
	h.argv = append(h.argv, argv)
	return h.side, h.err
}

// TestRerunReadsOnlyAnAnswerAsClean proves ADR-0050 point 2 for a tool finding: the result is "present"
// when the output carries the detector's banner and "gone" only when the tool answered without it; a tool
// that timed out, errored or has no host has not answered, so it is never read as clean.
func TestRerunReadsOnlyAnAnswerAsClean(t *testing.T) {
	paths := []string{"counter/counter.go"}
	cases := []struct {
		name     string
		host     Host
		answered bool
		present  bool
	}{
		{"still races", &fixedHost{side: Side{Output: raceChangeOutput, Status: testrun.CommandFail}}, true, true},
		{"no longer races", &fixedHost{side: Side{Output: raceBaseOutput, Status: testrun.CommandPass}}, true, false},
		{"timed out", &fixedHost{side: Side{Output: raceBaseOutput, Status: testrun.CommandTimeout}}, false, false},
		{"host error", &fixedHost{err: errors.New("no runtime")}, false, false},
		{"no host", nil, false, false},
	}
	for _, tc := range cases {
		got, err := Rerun(context.Background(), Race{}, tc.host, "/patched", paths)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got.Answered != tc.answered || got.Present != tc.present || (!got.Answered && got.Reason == "") {
			t.Fatalf("%s: %#v", tc.name, got)
		}
		if got.Command != "go test -race ./counter" {
			t.Fatalf("%s: the command was not the one Proven builds: %q", tc.name, got.Command)
		}
	}
}

// TestRerunRunsInTheDirectoryItWasGiven proves the tool runs on the patched copy and nowhere else.
func TestRerunRunsInTheDirectoryItWasGiven(t *testing.T) {
	h := &fixedHost{side: Side{Output: raceBaseOutput, Status: testrun.CommandPass}}
	if _, err := Rerun(context.Background(), Race{}, h, "/patched-copy", []string{"counter/counter.go"}); err != nil {
		t.Fatal(err)
	}
	if len(h.dirs) != 1 || h.dirs[0] != "/patched-copy" || strings.Join(h.argv[0], " ") != "go test -race ./counter" {
		t.Fatalf("%v %v", h.dirs, h.argv)
	}
}

// TestRerunDoesNotRunATestThatDoesNotApply proves a change with no Go file is not run and not read as clean.
func TestRerunDoesNotRunATestThatDoesNotApply(t *testing.T) {
	h := &fixedHost{}
	got, err := Rerun(context.Background(), Race{}, h, "/p", []string{"README.md"})
	if err != nil || got.Answered || len(h.dirs) != 0 || !strings.Contains(got.Reason, "does not apply") {
		t.Fatalf("%v %#v %v", err, got, h.dirs)
	}
}

// TestRerunReturnsACancelledContext proves the review's own cancellation is an error and not an open finding.
func TestRerunReturnsACancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Rerun(ctx, Race{}, &fixedHost{}, "/p", []string{"a/a.go"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

// TestByIDKnowsOnlyTheToolsWhoseResultIsEvidence proves a tool name from evidence maps to its adapter,
// and an unknown name has none, so no command is ever guessed.
func TestByIDKnowsOnlyTheToolsWhoseResultIsEvidence(t *testing.T) {
	if a, ok := ByID("go-race"); !ok || a.ID() != "go-race" {
		t.Fatal("go-race has no adapter")
	}
	for _, id := range []string{"", "govulncheck", "rm"} {
		if _, ok := ByID(id); ok {
			t.Fatalf("%q has an adapter", id)
		}
	}
}
