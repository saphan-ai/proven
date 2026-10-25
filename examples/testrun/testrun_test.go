package testrun_test

import (
	"context"
	"testing"

	eg "github.com/saphan-ai/proven/examples/testrun"
	it "github.com/saphan-ai/proven/internal/testrun"
)

// The assignment proves Runner compiles against both real interfaces:
// examples/testrun.Runner and internal/testrun agree on both method sets.
var _ it.Adapter = eg.Runner{}
var _ it.FixRunner = eg.Runner{}

// provingOutcome is a test that fails on the change and passes on the base:
// the only shape that proves a finding (ADR-0002).
func provingOutcome() it.Outcome {
	return it.Outcome{
		Found:         true,
		Path:          "cap_test.go",
		Name:          "TestCap",
		Command:       "go test -run TestCap ./...",
		Output:        "cap_test.go:12: cap exceeded\n",
		FailsOnChange: true,
		PassesOnBase:  true,
	}
}

// TestRunnerProvesAFinding proves one answer: a recorded outcome that fails
// on the change and passes on the base proves the finding.
func TestRunnerProvesAFinding(t *testing.T) {
	runner, err := eg.New([]eg.Recorded{
		{FindingID: "f1", Outcome: provingOutcome()},
	}, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	out, err := runner.Run(context.Background(), it.Request{FindingID: "f1", Name: "TestCap"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !out.Found {
		t.Fatal("got no outcome, want the recorded one")
	}
	if !out.FailsOnChange || !out.PassesOnBase {
		t.Fatalf("got %+v, want both fails-on-change and passes-on-base", out)
	}
}

// TestRunnerFindsNothing proves the refusal: a finding the runner did not
// record is not found, and that is not an error.
func TestRunnerFindsNothing(t *testing.T) {
	runner, err := eg.New([]eg.Recorded{
		{FindingID: "f1", Outcome: provingOutcome()},
	}, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	out, err := runner.Run(context.Background(), it.Request{FindingID: "f2", Name: "TestCap"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.Found {
		t.Fatalf("got an outcome for f2, want none: %+v", out)
	}
}

// TestRunnerRefusesACanceledContext proves the interface's error path: a
// canceled context returns an error and no outcome.
func TestRunnerRefusesACanceledContext(t *testing.T) {
	runner, err := eg.New([]eg.Recorded{
		{FindingID: "f1", Outcome: provingOutcome()},
	}, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = runner.Run(ctx, it.Request{FindingID: "f1", Name: "TestCap"})
	if err == nil {
		t.Fatal("a canceled context must be refused")
	}
}

// provingFix is one attempt whose four runs all ran and the patched test passed.
func provingFix() it.FixOutcome {
	return it.FixOutcome{
		Runs: []it.FixRun{
			{Label: it.RunBase, Command: "go test -run TestCap ./...", Result: it.ResultPassed, ExitCode: 0},
			{Label: it.RunChange, Command: "go test -run TestCap ./...", Result: it.ResultFailed, ExitCode: 1},
			{Label: it.RunPatched, Command: "go test -run TestCap ./...", Result: it.ResultPassed, ExitCode: 0},
			{Label: it.RunPackage, Command: "go test ./...", Result: it.ResultPassed, ExitCode: 0},
		},
	}
}

// TestRunnerProvesAFix proves one answer for the fix runner: the recorded
// runs show the patch proving the fix.
func TestRunnerProvesAFix(t *testing.T) {
	runner, err := eg.New(nil, []eg.RecordedFix{
		{FindingID: "f1", Attempt: 1, Outcome: provingFix()},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	out, err := runner.RunFix(context.Background(), it.FixRequest{FindingID: "f1", Attempt: 1})
	if err != nil {
		t.Fatalf("RunFix: %v", err)
	}
	if len(out.Runs) != 4 {
		t.Fatalf("got %d runs, want four: %+v", len(out.Runs), out.Runs)
	}
	if out.Refused != "" {
		t.Fatalf("got refused %q, want no refusal", out.Refused)
	}
	if out.Runs[3].Result != it.ResultPassed {
		t.Fatalf("package run is %s, want passed", out.Runs[3].Result)
	}
}

// TestRunnerRefusesAnUnknownAttempt proves the refusal for the fix runner: a
// missing attempt is refused by name and no code ran.
func TestRunnerRefusesAnUnknownAttempt(t *testing.T) {
	runner, err := eg.New(nil, []eg.RecordedFix{
		{FindingID: "f1", Attempt: 1, Outcome: provingFix()},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	out, err := runner.RunFix(context.Background(), it.FixRequest{FindingID: "f1", Attempt: 2})
	if err != nil {
		t.Fatalf("RunFix: %v", err)
	}
	if out.Refused == "" {
		t.Fatal("a missing attempt must be refused by name")
	}
	if len(out.Runs) != 0 {
		t.Fatalf("got %d runs, want none", len(out.Runs))
	}
}

// TestRunnerRefusesADuplicateFinding proves the builder refuses two outcomes for one finding.
func TestRunnerRefusesADuplicateFinding(t *testing.T) {
	_, err := eg.New([]eg.Recorded{
		{FindingID: "f1", Outcome: provingOutcome()},
		{FindingID: "f1", Outcome: provingOutcome()},
	}, nil)
	if err == nil {
		t.Fatal("a duplicate finding must be refused")
	}
}
