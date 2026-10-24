package review

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
)

// TestProofLevelOfEachKind proves that each evidence kind that exists today maps to its named level
// (ADR-0044 point 1), including a tool's run (ADR-0040), and that the strongest evidence decides
// when a finding has more than one kind.
func TestProofLevelOfEachKind(t *testing.T) {
	test := Evidence{Kind: EvidenceTest}
	query := Evidence{Kind: EvidenceQuery}
	tool := Evidence{Kind: EvidenceTool}
	cases := []struct {
		name string
		ev   []Evidence
		want ProofLevel
	}{
		{"a test is proved by test", []Evidence{test}, ProvedByTest},
		{"a query is proved by data", []Evidence{query}, ProvedByData},
		{"a tool is proved by a tool", []Evidence{tool}, ProvedByTool},
		{"query then test: the test is stronger", []Evidence{query, test}, ProvedByTest},
		{"test then query: order does not matter", []Evidence{test, query}, ProvedByTest},
		{"query then tool: the tool is stronger", []Evidence{query, tool}, ProvedByTool},
		{"tool then test: the test is stronger", []Evidence{tool, test}, ProvedByTest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ProofLevelOf(c.ev)
			if err != nil || got != c.want {
				t.Fatalf("got %q, %v; want %q", got, err, c.want)
			}
		})
	}
}

// TestProofLevelOfRefusesNoEvidence proves that evidence of no known kind gives no level and an error,
// rather than a default level that would show proof that is not there.
func TestProofLevelOfRefusesNoEvidence(t *testing.T) {
	for _, ev := range [][]Evidence{nil, {}, {{Kind: "guess"}}} {
		if got, err := ProofLevelOf(ev); err == nil || got != "" {
			t.Fatalf("evidence %#v: got %q, %v; want an error and no level", ev, got, err)
		}
	}
}

// TestFillRefusesAFindingWithNoLevel proves that a finding without evidence cannot reach the Result:
// fill returns an error naming the finding (ADR-0044; ADR-0002).
func TestFillRefusesAFindingWithNoLevel(t *testing.T) {
	_, err := fill(candidate.Draft{Findings: []candidate.Candidate{{Risk: candidate.Risk{ID: "X1"}}}})
	if err == nil || !strings.Contains(err.Error(), "X1") {
		t.Fatalf("error %v; want one naming X1", err)
	}
}

// TestFillCountsByLevel proves that every level is a key of ProofCounts, that each finding is counted
// once under its own level, and that the counts add up to the findings.
func TestFillCountsByLevel(t *testing.T) {
	d := candidate.Draft{Findings: []candidate.Candidate{
		{Risk: candidate.Risk{ID: "A"}, Evidence: []candidate.Evidence{{Kind: candidate.KindTest}}},
		{Risk: candidate.Risk{ID: "B"}, Evidence: []candidate.Evidence{{Kind: candidate.KindQuery}}, Unresolved: true},
		{Risk: candidate.Risk{ID: "C"}, Evidence: []candidate.Evidence{{Kind: candidate.KindQuery}}},
		{Risk: candidate.Risk{ID: "D"}, Evidence: []candidate.Evidence{{
			Kind: candidate.KindTool,
			Tool: &candidate.ToolEvidence{Tool: "go-race", Command: "go test -race ./counter", PresentOnChange: true},
		}}},
	}}
	res, err := fill(d)
	if err != nil {
		t.Fatal(err)
	}
	want := map[ProofLevel]int{ProvedByTest: 1, ProvedByTool: 1, ProvedByData: 2, AgreedNotProven: 0}
	if len(res.ProofCounts) != len(want) {
		t.Fatalf("counts %v; want every level as a key", res.ProofCounts)
	}
	for l, n := range want {
		if res.ProofCounts[l] != n {
			t.Fatalf("level %s: %d; want %d", l, res.ProofCounts[l], n)
		}
	}
	for _, f := range res.Findings {
		if !f.ProofLevel.Valid() {
			t.Fatalf("finding %s has level %q", f.ID, f.ProofLevel)
		}
	}
	if len(res.Findings) != 4 {
		t.Fatalf("%d findings; an unresolved one keeps its place", len(res.Findings))
	}
	var tool Finding
	for _, f := range res.Findings {
		if f.ID == "D" {
			tool = f
		}
	}
	if tool.ProofLevel != ProvedByTool || len(tool.Evidence) != 1 || tool.Evidence[0].Tool == nil || tool.Evidence[0].Tool.Command != "go test -race ./counter" {
		t.Fatalf("tool finding %#v", tool)
	}
}
