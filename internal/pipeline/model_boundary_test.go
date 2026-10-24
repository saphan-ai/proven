package pipeline

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
)

// unboundRisk answers the planner with one risk the discount change does not
// contain, and has the investigator look up a known-exploited row for a CVE
// that is in the sample tables but not in the change. The prover and the
// skeptic propose nothing.
type unboundRisk struct{ investigatorAnswer, title string }

func (g unboundRisk) Complete(_ context.Context, agent, _ string) (string, error) {
	switch agent {
	case "planner":
		title := g.title
		if title == "" {
			title = "CVE-2021-44228 remote code execution in the discount helper"
		}
		return `[{"id":"F9","rule_id":"remote-code-execution","title":"` + title + `","message":"The change is affected by CVE-2021-44228.","severity":"ERROR","path":"billing/discount.go","start_line":5,"end_line":5}]`, nil
	case "investigator":
		if g.investigatorAnswer != "" {
			return g.investigatorAnswer, nil
		}
		return `{"template":"kev-by-cve","params":{"cve_id":"CVE-2021-44228"}}`, nil
	case "prover":
		return `{"path":"","name":""}`, nil
	case "skeptic":
		return `{"refuted":false,"reason":"no refutation"}`, nil
	}
	return "", nil
}

func reviewUnbound(t *testing.T, gen generator.Generator) (kept9 bool, reason string) {
	t.Helper()
	p, err := Build(Options{
		Sample:    true,
		LogDir:    t.TempDir(),
		Generator: gen,
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Review(context.Background(), Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	c, ok := kept(d, "F9")
	return ok, c.DropReason
}

// Property: query rows are evidence for a finding only when they are about
// the change's own packages, versions or advisories. Rows that merely repeat
// a word of the model's own finding text are not evidence.
// Mutated field: the planner's answer names a CVE the change does not carry.
func TestQueryRowsUnrelatedToTheChangeAreNotEvidence(t *testing.T) {
	if ok, reason := reviewUnbound(t, unboundRisk{}); ok {
		t.Fatalf("a finding was kept on rows that say nothing about the change: reason %q", reason)
	}
}

// Negative control for the test above: a lookup that is bound to the change
// (the package the change adds, found in the advisory table) keeps its finding.
// Mutated field: the investigator's template and values.
func TestQueryRowsBoundToTheChangesPackageAreEvidence(t *testing.T) {
	gen := unboundRisk{
		investigatorAnswer: `{"template":"osv-by-package","params":{"ecosystem":"Go","package":"evil.example/leftpad"}}`,
		title:              "evil.example/leftpad v1.2.3 is affected",
	}
	if ok, _ := reviewUnbound(t, gen); !ok {
		t.Fatal("a lookup of the package the change adds, with rows for it, must remain evidence")
	}
}
