package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/planner"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/markov"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/trace"
	"path/filepath"
)

// These tests answer the review of ADR-0048's carried-state bound: a claim that awaits evidence must never be lost
// when the bound is reached. Either it reaches its evidence, or the review stops and says it is incomplete.

// awaitingRe reads the claims a prompt's carried state lists as awaiting evidence. Claim texts hold no quote or
// backslash, so the JSON escaping of the prompt does not touch them.
var awaitingRe = regexp.MustCompile(`awaiting evidence: ([^\\"]+)`)

// minCarry is the smallest carry bound the settings allow (the spec bounds 128 to 32768).
const minCarry = 128

// openClaim is the claim file n raises: long enough that a handful of them fill the smallest allowed carry bound.
func openClaim(n int) string {
	return fmt.Sprintf("f%02d needs a later file: %s", n, strings.Repeat("this claim is long on purpose to use the bound; ", 3))
}

// hoardingPlanner is a model that raises one long open claim in each of files 03 to 12, carries every claim it was
// handed, and settles the earliest claim, the one file 03 raised, at file 20 if that claim reached it. It answers
// the Planner only and keeps its prompts.
type hoardingPlanner struct {
	mu      sync.Mutex
	prompts []string
}

func (h *hoardingPlanner) Complete(_ context.Context, agent, prompt string) (string, error) {
	if agent != "planner" {
		return "", &service.Down{Service: agent, Reason: "scripted: only the Planner answers in this test"}
	}
	h.mu.Lock()
	h.prompts = append(h.prompts, prompt)
	h.mu.Unlock()
	if !strings.Contains(prompt, stepTask) {
		return "", fmt.Errorf("scripted: a whole prompt; this test is about bounded steps")
	}
	var carried []string
	for _, m := range awaitingRe.FindAllStringSubmatch(prompt, -1) {
		carried = append(carried, strings.TrimSpace(m[1]))
	}
	n := 0
	if f := filesIn(prompt); len(f) == 1 {
		fmt.Sscanf(f[0], "UNIQUE-f%d", &n)
	}
	ans := map[string]any{"risks": []any{}, "established": []string{}, "suspected": []string{}}
	switch {
	case n == 20 && hasClaim(carried, openClaim(3)):
		ans["risks"] = []any{sinkRiskRow()}
		carried = without(carried, openClaim(3))
	case n >= 3 && n <= 12:
		carried = append(carried, openClaim(n))
	}
	if isLastStep(prompt) {
		// Nothing is left to read, so the model gives up every claim; an open claim here would stop the review.
		carried = nil
	}
	ans["awaiting_evidence"] = carried
	b, _ := json.Marshal(ans)
	return string(b), nil
}

// hasClaim reports whether claims holds want.
func hasClaim(claims []string, want string) bool {
	for _, c := range claims {
		if strings.TrimSpace(c) == strings.TrimSpace(want) {
			return true
		}
	}
	return false
}

// without returns claims minus want.
func without(claims []string, want string) []string {
	var out []string
	for _, c := range claims {
		if strings.TrimSpace(c) != strings.TrimSpace(want) {
			out = append(out, c)
		}
	}
	return out
}

// reviewHoarding runs the 24-file fixture through a whole review with gen and settings s and returns the review's
// error and the trace.
func reviewHoarding(t *testing.T, gen *hoardingPlanner, s markov.Settings) (error, []trace.Event) {
	t.Helper()
	logs := t.TempDir()
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
	_, rerr := p.Review(context.Background(), Request{Target: boundedFixture(t, 24), FixesOff: true})
	return rerr, readTrace(t, filepath.Join(logs, "trace.jsonl"))
}

// TestAnOverflowingCarryStopsTheReviewByName proves the property of ADR-0048's carried-state bound: when the open
// claims no longer fit the carry bound, the earliest claim is not dropped in silence. The review stops with
// *markov.CarryFullError, the error says the review is incomplete and why, and the trace says so too. Unplug
// control: with the check removed from ReviewState.Carry, the bound drops the earliest claim, the review ends
// without an error and without the risk, and this test is red.
func TestAnOverflowingCarryStopsTheReviewByName(t *testing.T) {
	gen := &hoardingPlanner{}
	s := on(markov.UnitFile)
	s.Carry = minCarry
	err, events := reviewHoarding(t, gen, s)
	var full *markov.CarryFullError
	if !errors.As(err, &full) {
		t.Fatalf("the review ended with %v, want *markov.CarryFullError", err)
	}
	if full.Held >= full.Open || full.Budget != minCarry || !strings.Contains(err.Error(), "the bounded review is incomplete") {
		t.Fatalf("%+v: %v", full, err)
	}
	if len(gen.prompts) >= 20 {
		t.Fatalf("%d steps ran: the review went on past the overflow", len(gen.prompts))
	}
	var traced bool
	for _, e := range linesOf(events, traceKindBounded) {
		if strings.HasPrefix(e.Text, "bounded steps stopped, the review is incomplete: ") {
			traced = true
		}
	}
	if !traced {
		t.Fatal("the trace does not say the review is incomplete")
	}
}

// TestTheEarliestOpenClaimReachesItsEvidenceWhenTheBoundHoldsThem proves the other side: with a carry bound that
// holds every open claim, the claim file 03 raised is carried to file 20, which never sees file 03, and becomes a
// risk. The stop above is therefore caused by the bound and not by the scripted model.
func TestTheEarliestOpenClaimReachesItsEvidenceWhenTheBoundHoldsThem(t *testing.T) {
	gen := &hoardingPlanner{}
	s := on(markov.UnitFile)
	s.Carry = markov.DefaultCarry
	err, events := reviewHoarding(t, gen, s)
	if err != nil {
		t.Fatal(err)
	}
	if len(gen.prompts) != 24 || strings.Contains(gen.prompts[19], unique(3)) {
		t.Fatalf("%d steps; step 20 must not see file 03", len(gen.prompts))
	}
	var claimed bool
	for _, e := range linesOf(events, "claim") {
		claimed = claimed || (e.FindingID == "R1" && e.Text == sinkRisk)
	}
	if !claimed {
		t.Fatal("the earliest claim did not become a risk at its evidence")
	}
}

// longEvidence names the evidence the long claim waits for. It is the claim's last words, beyond byte 400.
const longEvidence = "evidence-in-file-20"

// longClaim is an open claim longer than the 400 bytes a decision or context item may take, whose required evidence
// is named only at its end.
func longClaim() string {
	return "f03 needs a later file: " + strings.Repeat("this claim is long on purpose and names its evidence last; ", 8) + "it needs " + longEvidence
}

// longClaimPlanner raises longClaim at file 03, carries what it is handed, and settles the claim at file 20 only if
// the claim reached it whole, that is with its evidence name intact. It answers the Planner only and keeps its prompts.
type longClaimPlanner struct {
	mu      sync.Mutex
	prompts []string
}

func (l *longClaimPlanner) Complete(_ context.Context, agent, prompt string) (string, error) {
	if agent != "planner" {
		return "", &service.Down{Service: agent, Reason: "scripted: only the Planner answers in this test"}
	}
	l.mu.Lock()
	l.prompts = append(l.prompts, prompt)
	l.mu.Unlock()
	var carried []string
	for _, m := range awaitingRe.FindAllStringSubmatch(prompt, -1) {
		carried = append(carried, strings.TrimSpace(m[1]))
	}
	n := 0
	if f := filesIn(prompt); len(f) == 1 {
		fmt.Sscanf(f[0], "UNIQUE-f%d", &n)
	}
	ans := map[string]any{"risks": []any{}, "established": []string{}, "suspected": []string{}}
	switch {
	case n == 3:
		carried = append(carried, longClaim())
	case n == 20 && hasClaim(carried, longClaim()):
		ans["risks"] = []any{sinkRiskRow()}
		carried = without(carried, longClaim())
	}
	ans["awaiting_evidence"] = carried
	b, _ := json.Marshal(ans)
	return string(b), nil
}

// reviewLong runs the 24-file fixture through gen with carry bound carry and returns the review's error and the trace.
func reviewLong(t *testing.T, gen *longClaimPlanner, carry int) (error, []trace.Event) {
	t.Helper()
	logs := t.TempDir()
	s := on(markov.UnitFile)
	s.Carry = carry
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
	_, rerr := p.Review(context.Background(), Request{Target: boundedFixture(t, 24), FixesOff: true})
	return rerr, readTrace(t, filepath.Join(logs, "trace.jsonl"))
}

// TestALongOpenClaimIsCarriedWholeToItsEvidence proves the property of ADR-0048's carried state: a claim awaiting
// evidence is never shortened. The claim is over 400 bytes and names its evidence beyond byte 400; the bound holds it
// whole; it reaches file 20 intact and becomes a finding. Unplug control: clip awaiting claims again, as decisions
// and context items are clipped, and file 20 receives the claim cut to "[cut]", the scripted model cannot settle it,
// no risk is reported, and this test is red.
func TestALongOpenClaimIsCarriedWholeToItsEvidence(t *testing.T) {
	if i := strings.Index(longClaim(), longEvidence); i <= 400 {
		t.Fatalf("the evidence name is at byte %d; the test needs it beyond 400", i)
	}
	gen := &longClaimPlanner{}
	err, events := reviewLong(t, gen, markov.DefaultCarry)
	if err != nil {
		t.Fatal(err)
	}
	if len(gen.prompts) != 24 || strings.Contains(gen.prompts[19], unique(3)) || !strings.Contains(gen.prompts[19], longEvidence) || strings.Contains(gen.prompts[19], "[cut]") {
		t.Fatalf("%d steps; step 20 must carry the whole claim and not file 03", len(gen.prompts))
	}
	var claimed bool
	for _, e := range linesOf(events, "claim") {
		claimed = claimed || (e.FindingID == "R1" && e.Text == sinkRisk)
	}
	if !claimed {
		t.Fatal("the long claim did not become a risk at its evidence")
	}
}

// TestALongOpenClaimThatCannotBeHeldWholeStopsTheReviewBeforeAnyCut proves the other side: when the bound cannot hold
// the claim whole, the review stops with *markov.CarryFullError naming step 3, the step that raised it, and no later
// step runs on a shortened claim. Unplug control: with the bound check removed, the claim is carried over budget or
// cut, the review goes on without an error, and this test is red.
func TestALongOpenClaimThatCannotBeHeldWholeStopsTheReviewBeforeAnyCut(t *testing.T) {
	gen := &longClaimPlanner{}
	err, _ := reviewLong(t, gen, minCarry)
	var full *markov.CarryFullError
	if !errors.As(err, &full) {
		t.Fatalf("the review ended with %v, want *markov.CarryFullError", err)
	}
	if full.Step != 3 || full.Open != 1 || full.Held != 0 {
		t.Fatalf("%+v", full)
	}
	if len(gen.prompts) != 3 {
		t.Fatalf("%d steps ran, want 3: the review went on past the overflow", len(gen.prompts))
	}
	for _, p := range gen.prompts {
		if strings.Contains(p, "[cut]") {
			t.Fatal("a prompt carried a shortened claim")
		}
	}
}

// TestTheThresholdSettingDecidesWhetherAChangeIsCut is the negative control for the threshold setting (acceptance
// 3(c)). The same 3-file change, about 4,600 estimated tokens, is reviewed with thresholds on both sides of it, and
// of the default 8000. A review that ignores the setting, whether it always cuts, never cuts, or reads the default
// constant, gives the wrong number of Planner calls for at least one row and this test is red.
func TestTheThresholdSettingDecidesWhetherAChangeIsCut(t *testing.T) {
	root := boundedFixture(t, 3)
	files, err := change.Read(root)
	if err != nil {
		t.Fatal(err)
	}
	size := planner.Size(files)
	for _, c := range []struct {
		threshold, calls int
	}{{1000, 3}, {size - 1, 3}, {size, 1}, {size + 1, 1}, {markov.DefaultThreshold * 2, 1}} {
		s := on(markov.UnitFile)
		s.Threshold = c.threshold
		_, gen, _ := boundedReview(t, root, s)
		if len(gen.prompts) != c.calls {
			t.Errorf("threshold %d (change %d): %d Planner calls, want %d", c.threshold, size, len(gen.prompts), c.calls)
		}
	}
}
