// This file is an external test. It imports review, and review imports tool,
// so an internal test would be an import cycle.
package tool_test

import (
	"context"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/judge"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/review"
	"github.com/saphan-ai/proven/internal/tool"
)

// TestSampleRaceReportsAFindingProvedByATool proves the recorded race scenario
// is a finding the Judge keeps, at the proof level of a tool (ADR-0040, ADR-0044).
func TestSampleRaceReportsAFindingProvedByATool(t *testing.T) {
	out, err := tool.NewSample(tool.Race{}, tool.RaceIntroduced()).Run(context.Background())
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
