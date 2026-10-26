package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/saphan-ai/proven/internal/markov"
	"github.com/saphan-ai/proven/internal/service"
)

// These tests answer the review of ADR-0048's last step: a claim still awaiting evidence when no part is left can
// never be settled, so the bounded review must not return it as a clean success.

// lastClaim is the claim the scripted model leaves open at the last file.
const lastClaim = "f02 passes its input to a function that may live in a file this change does not show"

// terminalPlanner answers the Planner only. At the last file (UNIQUE-f02) it returns no risk and, when open is
// true, one claim awaiting evidence; at every other file it returns an empty state.
type terminalPlanner struct {
	open    bool
	mu      sync.Mutex
	prompts []string
}

func (g *terminalPlanner) Complete(_ context.Context, agent, prompt string) (string, error) {
	if agent != "planner" {
		return "", &service.Down{Service: agent, Reason: "scripted: only the Planner answers in this test"}
	}
	g.mu.Lock()
	g.prompts = append(g.prompts, prompt)
	g.mu.Unlock()
	ans := map[string]any{"risks": []any{}, "established": []string{}, "suspected": []string{}, "awaiting_evidence": []string{}}
	if g.open && strings.Contains(prompt, unique(2)) {
		ans["awaiting_evidence"] = []string{lastClaim}
	}
	b, _ := json.Marshal(ans)
	return string(b), nil
}

// reviewTwoFiles runs a two-file change, above a low threshold so it is cut into two steps, through a whole review.
func reviewTwoFiles(t *testing.T, gen *terminalPlanner) (error, *terminalPlanner, []string) {
	t.Helper()
	logs := t.TempDir()
	s := on(markov.UnitFile)
	s.Threshold = 1000
	p, err := Build(Options{
		Sample: true, LogDir: logs, Generator: gen, Markov: s,
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, rerr := p.Review(context.Background(), Request{Target: boundedFixture(t, 2), FixesOff: true})
	var texts []string
	for _, e := range linesOf(readTrace(t, filepath.Join(logs, "trace.jsonl")), traceKindBounded) {
		texts = append(texts, e.Text)
	}
	return rerr, gen, texts
}

// TestAClaimStillOpenAfterTheLastStepStopsTheReviewByName proves the property of ADR-0048's last step: when the
// final answer holds a claim awaiting evidence and no part is left, the review ends with *markov.OpenClaimsError,
// the error says the review is incomplete and why, and the trace says so too. Before the fix this ended with no
// error and no risks. Unplug control: with the terminal check removed from planner.PlanBounded, the review ends
// without an error and this test is red.
func TestAClaimStillOpenAfterTheLastStepStopsTheReviewByName(t *testing.T) {
	err, gen, trace := reviewTwoFiles(t, &terminalPlanner{open: true})
	var open *markov.OpenClaimsError
	if !errors.As(err, &open) {
		t.Fatalf("the review ended with %v, want *markov.OpenClaimsError", err)
	}
	if open.Steps != 2 || open.Open != 1 || !strings.Contains(err.Error(), "the bounded review is incomplete") {
		t.Fatalf("%+v: %v", open, err)
	}
	if strings.Contains(err.Error(), lastClaim) {
		t.Fatal("the error carries model text")
	}
	if len(gen.prompts) != 2 {
		t.Fatalf("%d steps ran, want 2", len(gen.prompts))
	}
	var traced bool
	for _, line := range trace {
		traced = traced || strings.HasPrefix(line, "bounded steps stopped, the review is incomplete: ")
	}
	if !traced {
		t.Fatal("the trace does not say the review is incomplete")
	}
}

// TestALastStepWithNoOpenClaimStillSucceeds proves the check is not over-eager: the same change, with a last answer
// that holds no claim awaiting evidence, finishes with no error.
func TestALastStepWithNoOpenClaimStillSucceeds(t *testing.T) {
	err, gen, _ := reviewTwoFiles(t, &terminalPlanner{})
	if err != nil {
		t.Fatalf("the review ended with %v, want no error", err)
	}
	if len(gen.prompts) != 2 {
		t.Fatalf("%d steps ran, want 2", len(gen.prompts))
	}
}
