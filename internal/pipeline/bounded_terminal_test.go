package pipeline

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/markov"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
)

// A final open planning claim is a visible coverage gap. It must not discard
// already identified risks or be presented as a proved finding.

// lastClaim is the claim the scripted model leaves open at the last file.
const lastClaim = "f02 passes its input to a function that may live in a file this change does not show"

// terminalPlanner answers the Planner only. At the last file (UNIQUE-f02) it returns no risk and, when open is
// true, one claim awaiting evidence; at every other file it returns an empty state.
type terminalPlanner struct {
	open    bool
	risk    bool
	checked []string
	mu      sync.Mutex
	prompts []string
}

func (g *terminalPlanner) Complete(_ context.Context, agent, prompt string) (string, error) {
	if agent != "planner" {
		g.checked = append(g.checked, agent)
		if g.risk {
			if agent == "prover" {
				return `{"path":"pkg/f01_test.go","name":"TestF01","content":""}`, nil
			}
			return `{}`, nil
		}
		return "", &service.Down{Service: agent, Reason: "scripted: only the Planner answers in this test"}
	}
	g.mu.Lock()
	g.prompts = append(g.prompts, prompt)
	g.mu.Unlock()
	ans := map[string]any{"risks": []any{}, "established": []string{}, "suspected": []string{}, "awaiting_evidence": []string{}}
	if g.risk && strings.Contains(prompt, unique(1)) {
		ans["risks"] = []any{map[string]any{"id": "R1", "rule_id": "wrong-result", "title": "F01 returns the wrong value", "message": "the changed value is wrong", "severity": "ERROR", "path": "pkg/f01.go", "start_line": 4, "end_line": 4}}
	}
	if g.open && strings.Contains(prompt, unique(2)) {
		ans["awaiting_evidence"] = []string{lastClaim}
	}
	b, _ := json.Marshal(ans)
	return string(b), nil
}

// reviewTwoFiles runs a two-file change, above a low threshold so it is cut into two steps, through a whole review.
func reviewTwoFiles(t *testing.T, gen *terminalPlanner) (candidate.Draft, error, *terminalPlanner, []string) {
	t.Helper()
	logs := t.TempDir()
	s := on(markov.UnitFile)
	s.Threshold = 1000
	p, err := Build(Options{
		Sample: true, LogDir: logs, Generator: gen, Markov: s,
		TestRun: terminalTestRun{},
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	d, rerr := p.Review(context.Background(), Request{Target: boundedFixture(t, 2), FixesOff: true})
	var texts []string
	for _, e := range linesOf(readTrace(t, filepath.Join(logs, "trace.jsonl")), traceKindBounded) {
		texts = append(texts, e.Text)
	}
	return d, rerr, gen, texts
}

type terminalTestRun struct{}

func (terminalTestRun) Run(_ context.Context, req testrun.Request) (testrun.Outcome, error) {
	return testrun.Outcome{Path: req.Path, Name: req.Name, Found: true, FailsOnChange: true, PassesOnBase: true, Output: "recorded differential outcome"}, nil
}

// A proved risk survives an unrelated open claim; every evidence stage runs.
func TestFinalOpenClaimPreservesTheReportAndKnownEvidence(t *testing.T) {
	d, err, gen, lines := reviewTwoFiles(t, &terminalPlanner{open: true, risk: true})
	if err != nil {
		t.Fatal(err)
	}
	if d.Partial || len(d.UnresolvedHypotheses) != 1 || d.UnresolvedHypotheses[0].Description != lastClaim || d.UnresolvedHypotheses[0].Reviewer != runDefault || d.UnresolvedHypotheses[0].AfterStep != 2 {
		t.Fatalf("draft: %#v", d)
	}
	if len(d.Findings) != 1 || d.Findings[0].ID != "R1" {
		t.Fatalf("the known finding was lost: %#v", d.Findings)
	}
	for _, agent := range []string{"investigator", "prover", "skeptic"} {
		if !strings.Contains(strings.Join(gen.checked, ","), agent) {
			t.Fatalf("%s was skipped: %v", agent, gen.checked)
		}
	}
	if len(gen.prompts) != 2 {
		t.Fatalf("%d steps ran, want 2", len(gen.prompts))
	}
	var traced bool
	for _, line := range lines {
		traced = traced || strings.HasPrefix(line, "bounded planning finished with unresolved hypotheses;")
	}
	if !traced {
		t.Fatal("the trace does not say the review is incomplete")
	}
}

func TestFinalOpenClaimWithoutRisksStillProducesAReport(t *testing.T) {
	d, err, _, _ := reviewTwoFiles(t, &terminalPlanner{open: true})
	if err != nil || len(d.UnresolvedHypotheses) != 1 || len(d.Findings) != 0 {
		t.Fatalf("%#v %v", d, err)
	}
	var found bool
	for _, e := range readTrace(t, d.TracePath) {
		if e.Kind == "unresolved-hypothesis" && e.Text == lastClaim {
			found = true
		}
	}
	if !found {
		t.Fatal("the trace lost the hypothesis description")
	}
}

// TestALastStepWithNoOpenClaimStillSucceeds proves the check is not over-eager: the same change, with a last answer
// that holds no claim awaiting evidence, finishes with no error.
func TestALastStepWithNoOpenClaimStillSucceeds(t *testing.T) {
	d, err, gen, _ := reviewTwoFiles(t, &terminalPlanner{})
	if err != nil {
		t.Fatalf("the review ended with %v, want no error", err)
	}
	if len(gen.prompts) != 2 {
		t.Fatalf("%d steps ran, want 2", len(gen.prompts))
	}
	if len(d.UnresolvedHypotheses) != 0 {
		t.Fatal("a closed plan was marked unresolved")
	}
}

func TestFinalHypothesesAreRedactedAndAttributedToEachReviewer(t *testing.T) {
	p, err := Build(Options{Sample: true, LogDir: t.TempDir(), Generator: &terminalPlanner{open: true}, Special: &terminalPlanner{open: true}, Markov: on(markov.UnitFile), Redactor: redact.New(map[string]string{"TEST_SECRET": lastClaim})})
	if err != nil {
		t.Fatal(err)
	}
	p.markov.Threshold = 1000
	d, err := p.Review(context.Background(), Request{Target: boundedFixture(t, 2), DoubleRun: true, DefaultModel: "model-a", SpecialModel: "model-b", FixesOff: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.UnresolvedHypotheses) != 2 {
		t.Fatalf("%#v", d.UnresolvedHypotheses)
	}
	for i, h := range d.UnresolvedHypotheses {
		if h.Description != "[redacted: TEST_SECRET]" || h.Reviewer != []string{runDefault, runSpecial}[i] {
			t.Fatalf("%#v", h)
		}
	}
	for _, e := range readTrace(t, d.TracePath) {
		if strings.Contains(e.Text, lastClaim) {
			t.Fatal("secret in trace")
		}
	}
}
