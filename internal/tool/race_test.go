package tool

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/judge"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/review"
	"github.com/saphan-ai/proven/internal/testrun"
)

// TestSampleRaceReportsAFindingProvedByATool proves the recorded race scenario
// is a finding the Judge keeps, at the proof level of a tool (ADR-0040, ADR-0044).
func TestSampleRaceReportsAFindingProvedByATool(t *testing.T) {
	out, err := NewSample((Race{}), RaceIntroduced()).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !out.Proves() || out.Version != "go1.26" || out.Command != "go test -race ./counter" {
		t.Fatalf("%#v", out)
	}
	if !strings.Contains(out.ChangeOutput, "WARNING: DATA RACE") || strings.Contains(out.BaseOutput, "WARNING: DATA RACE") {
		t.Fatalf("outputs: change %q base %q", out.ChangeOutput, out.BaseOutput)
	}
	c, ok := out.Candidate("counter/counter.go")
	if !ok {
		t.Fatal("the scenario reported no finding")
	}
	kept := judge.Decide(c)
	if kept.Dropped || kept.Evidence[0].Kind != candidate.KindTool {
		t.Fatalf("%#v", kept)
	}
	level, err := review.ProofLevelOf([]review.Evidence{{Kind: review.EvidenceTool}})
	if err != nil || level != review.ProvedByTool {
		t.Fatalf("%s %v", level, err)
	}
}

// TestRaceOnTheBaseIsNotEvidence proves a race that is also on the base is not
// evidence of the change (ADR-0040).
// Mutated field: PresentOnBase inside Introduced.
func TestRaceOnTheBaseIsNotEvidence(t *testing.T) {
	out, err := NewSample((Race{}), RaceOnBoth()).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !out.Answered || !out.PresentOnChange || !out.PresentOnBase {
		t.Fatalf("the recording was not a race on both sides: %#v", out)
	}
	if out.Proves() {
		t.Fatal("a race on the base was evidence")
	}
	if _, ok := out.Candidate("counter/counter.go"); ok {
		t.Fatal("a finding was reported for a race the base already had")
	}
	kept := judge.Decide(candidate.Candidate{
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindTool,
			Tool: &candidate.ToolEvidence{Tool: "go-race", PresentOnChange: true, PresentOnBase: true},
		}},
	})
	if !kept.Dropped {
		t.Fatal("the judge kept a race the base already had")
	}
}

// TestUnrunToolIsNamed proves a tool that does not answer is named and is not
// counted as a clean run (ADR-0040).
func TestUnrunToolIsNamed(t *testing.T) {
	out, err := NewSample((Race{}), RaceTimeout()).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if out.Answered || out.Proves() || out.Reason == "" {
		t.Fatalf("%#v", out)
	}
	if !strings.Contains(out.Reason, "go-race") || !strings.Contains(out.Reason, "timeout") {
		t.Fatalf("reason %q does not name the tool and the timeout", out.Reason)
	}
}

// TestAFailingTestWithoutTheBannerIsNotARace proves a test failure is not a
// data race. The banner is the detector's own line.
func TestAFailingTestWithoutTheBannerIsNotARace(t *testing.T) {
	if (Race{}).Read("FAIL\n--- FAIL: TestSum\n").Present {
		t.Fatal("a failing test was read as a race")
	}
	if (Race{}).Read("warning: data race").Present {
		t.Fatal("a different spelling was read as the detector's banner")
	}
	if !(Race{}).Read("WARNING: DATA RACE\n").Present {
		t.Fatal("the banner was missed")
	}
}

// TestRaceArgvIsTheDetectorCommand proves the command is go test -race on the
// changed packages, with no shell, and that a path which leaves the checkout
// is refused.
func TestRaceArgvIsTheDetectorCommand(t *testing.T) {
	argv, reason := (Race{}).Argv([]string{"counter/race_test.go", "other/a.go", "README.md"})
	if reason != "" {
		t.Fatal(reason)
	}
	if strings.Join(argv, " ") != "go test -race ./counter ./other" {
		t.Fatalf("%q", argv)
	}
	for _, word := range argv {
		if word == "sh" || word == "-c" {
			t.Fatalf("a shell word in %q", argv)
		}
	}
	if _, reason := (Race{}).Argv([]string{"../counter/counter.go"}); reason == "" {
		t.Fatal("a path that leaves the checkout was accepted")
	}
	if _, reason := (Race{}).Argv([]string{"counter/.../x.go"}); reason == "" {
		t.Fatal("a ... segment was accepted")
	}
	root, reason := (Race{}).Argv([]string{"main.go"})
	if reason != "" || strings.Join(root, " ") != "go test -race ." {
		t.Fatalf("%q %s", root, reason)
	}
}

// TestRaceDoesNotApplyWhenNoGoFileChanged proves a change with no Go file
// is skipped by name, not reported as a clean race run.
func TestRaceDoesNotApplyWhenNoGoFileChanged(t *testing.T) {
	if (Race{}).Applies([]string{"go.mod", "README.md"}) {
		t.Fatal("applies")
	}
	out, err := NewRunner((Race{}), nil).Run(context.Background(), Request{Paths: []string{"go.mod"}})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Skipped || out.Proves() || !strings.Contains(out.Reason, "go-race") {
		t.Fatalf("%#v", out)
	}
}

// scriptHost returns canned sides. It records the argv it was given.
type scriptHost struct {
	change Side
	base   Side
	err    error
	argv   [][]string
	dirs   []string
}

func (h *scriptHost) Run(_ context.Context, dir string, argv []string) (Side, error) {
	h.argv = append(h.argv, append([]string(nil), argv...))
	h.dirs = append(h.dirs, dir)
	if h.err != nil {
		return Side{}, h.err
	}
	if len(h.dirs) == 1 {
		return h.change, nil
	}
	return h.base, nil
}

// TestTimeoutIsNamedEvenWhenTheBannerWasPrinted proves a side that ran past
// its limit is not evidence, and the base is not started after it.
func TestTimeoutIsNamedEvenWhenTheBannerWasPrinted(t *testing.T) {
	h := &scriptHost{change: Side{Output: "WARNING: DATA RACE", Status: testrun.CommandTimeout}}
	out, err := NewRunner((Race{}), h).Run(context.Background(), Request{
		Paths: []string{"counter/counter.go"}, BaseDir: "/base", ChangeDir: "/change",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Answered || out.Proves() || !strings.Contains(out.Reason, "go-race") || !strings.Contains(out.Reason, "timeout") {
		t.Fatalf("%#v", out)
	}
	if len(h.dirs) != 1 || h.dirs[0] != "/change" {
		t.Fatalf("dirs %v", h.dirs)
	}
	if strings.Join(h.argv[0], " ") != "go test -race ./counter" {
		t.Fatalf("%q", h.argv[0])
	}
}

// TestARaceOnlyOnTheChangeIsEvidence proves the live comparison, with a fake
// host and no container: a banner on the change and a pass on the base.
func TestARaceOnlyOnTheChangeIsEvidence(t *testing.T) {
	h := &scriptHost{
		change: Side{Output: "WARNING: DATA RACE\nFAIL\n", Status: testrun.CommandFail},
		base:   Side{Output: "ok\n", Status: testrun.CommandPass},
	}
	out, err := NewRunner((Race{}), h).Run(context.Background(), Request{
		Paths: []string{"counter/counter.go"}, BaseDir: "/base", ChangeDir: "/change",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Proves() || out.Version != "unknown" {
		t.Fatalf("%#v", out)
	}
	if len(h.dirs) != 2 || h.dirs[0] != "/change" || h.dirs[1] != "/base" {
		t.Fatalf("dirs %v", h.dirs)
	}
}

// TestAMissingContainerIsNamed proves the runner does not stay silent when
// there is no container to run the tool in.
func TestAMissingContainerIsNamed(t *testing.T) {
	out, err := NewRunner((Race{}), nil).Run(context.Background(), Request{
		Paths: []string{"counter/counter.go"}, BaseDir: "/base", ChangeDir: "/change",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Proves() || !strings.Contains(out.Reason, "go-race") || !strings.Contains(out.Reason, "no container") {
		t.Fatalf("%#v", out)
	}
}

// TestAHostErrorIsNamed proves an error from the host names the tool and is
// not returned as a silent empty outcome. A cancelled context still stops the run.
func TestAHostErrorIsNamed(t *testing.T) {
	h := &scriptHost{err: errors.New("runtime missing")}
	out, err := NewRunner((Race{}), h).Run(context.Background(), Request{
		Paths: []string{"counter/counter.go"}, BaseDir: "/base", ChangeDir: "/change",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.Reason, "go-race") || !strings.Contains(out.Reason, "runtime missing") || out.Proves() {
		t.Fatalf("%#v", out)
	}
}

// TestRaceCommandRidesTheLockedDownContainer proves the race command is the
// tail of the Prover's run line: no network, no shell (ADR-0025, ADR-0040).
// This test starts no container.
func TestRaceCommandRidesTheLockedDownContainer(t *testing.T) {
	box, err := testrun.NewContainer(testrun.DefaultContainerConfig(), func(string) (string, error) {
		return "/bin/docker", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	argv, reason := (Race{}).Argv([]string{"counter/counter.go"})
	if reason != "" {
		t.Fatal(reason)
	}
	line := box.Argv(t.TempDir(), argv)
	if !containsPair(line, "--network", "none") {
		t.Fatalf("no --network none in %q", line)
	}
	tail := line[len(line)-len(argv):]
	if strings.Join(tail, "\x00") != strings.Join(argv, "\x00") {
		t.Fatalf("tail %q", tail)
	}
	for _, word := range line {
		if word == "sh" || word == "-c" {
			t.Fatalf("a shell in %q", line)
		}
	}
}

func containsPair(argv []string, flag, value string) bool {
	for i := 0; i+1 < len(argv); i++ {
		if argv[i] == flag && argv[i+1] == value {
			return true
		}
	}
	return false
}
