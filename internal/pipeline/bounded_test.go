package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/saphan-ai/proven/internal/agent/planner"
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/change"
	"github.com/saphan-ai/proven/internal/fixture"
	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/markov"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/testrun"
	"github.com/saphan-ai/proven/internal/trace"
)

// These are the integration tests the owner made the condition of ADR-0048's engine landing without a separate
// test author: inside a whole review, with a scripted generator and no network, a change above the threshold is
// read one part per step with only the carried state between steps, a claim waiting for evidence is carried until
// the part that settles it, a change at or below the threshold is one whole step, and the fix loop's next attempt
// carries only what failed and why.

// The texts the scripted Planner keys on. The claim is what file 03 alone can only suspect; file 20 settles it.
const (
	sinkClaim = "f03 passes user input to sink; needs the body of sink to show whether it validates"
	sinkRisk  = "User input reaches exec through sink"
)

// unique is the marker that names file i in its content, so a prompt shows which files it carries.
func unique(i int) string {
	return fmt.Sprintf("UNIQUE-f%02d", i)
}

// boundedFixture writes a sample fixture of n changed Go files under a temporary root and returns the root. File 03
// adds a caller that passes user input to sink; file 20 adds sink, which runs it unvalidated. Neither file shows the
// risk on its own. The fixture's recorded Planner answer is never used: the scripted generator answers instead.
func boundedFixture(t *testing.T, n int) string {
	t.Helper()
	root := t.TempDir()
	filler := strings.Repeat("// a line that makes the file the size of a real one, so the threshold has something to measure\n", 30)
	for i := 1; i <= n; i++ {
		name := fmt.Sprintf("pkg/f%02d.go", i)
		base := fmt.Sprintf("package pkg\n\n// %s\nfunc F%02d() int { return %d }\n%s", unique(i), i, i, filler)
		chg := strings.Replace(base, fmt.Sprintf("return %d }", i), fmt.Sprintf("return %d + 1 }", i), 1)
		switch i {
		case 3:
			chg += "\nfunc Handle(userInput string) { sink(userInput) } // CALLS sink(userInput)\n"
		case 20:
			chg += "\nfunc sink(x string) { exec(x) } // NO-VALIDATION\n"
		}
		writeFile(t, filepath.Join(root, "base", name), base)
		writeFile(t, filepath.Join(root, "change", name), chg)
	}
	files, err := change.Read(root)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(planner.Prompt(files)))
	fx := map[string]any{
		"snapshot":              "bounded-steps-scripted",
		"planner_prompt_sha256": hex.EncodeToString(sum[:]),
		"planner_response":      json.RawMessage(`[]`),
	}
	b, err := json.Marshal(fx)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "fixture.json"), string(b))
	tables, err := os.ReadFile(filepath.Join(repoRoot(t), "testdata", "fixture", "fix", "tables.sql"))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "tables.sql"), string(tables))
	return root
}

// writeFile writes text to path, making its directory.
func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// scriptedPlanner answers the Planner as a model reading what it is shown would, and nothing else: every other
// agent is a down service, so the review ends after the Planner without any recorded answer. It keeps every
// Planner prompt and answer, for the assertions and for the measurement.
type scriptedPlanner struct {
	mu      sync.Mutex
	prompts []string
	answers []string
}

// wholeTask and stepTask tell the whole Planner prompt from a bounded step's.
const (
	wholeTask = "Task: Name the risks in this change."
	stepTask  = "Task: Name the risks in one part of a larger change."
)

func (s *scriptedPlanner) Complete(_ context.Context, agent, prompt string) (string, error) {
	if agent != "planner" {
		return "", &service.Down{Service: agent, Reason: "scripted: only the Planner answers in this test"}
	}
	var out string
	switch {
	case strings.Contains(prompt, wholeTask):
		// The whole change shows both ends of the flow at once.
		var risks []map[string]any
		if strings.Contains(prompt, unique(3)) && strings.Contains(prompt, unique(20)) {
			risks = append(risks, sinkRiskRow())
		}
		b, _ := json.Marshal(risks)
		out = string(b)
	case strings.Contains(prompt, stepTask):
		out = stepAnswer(prompt)
	default:
		return "", fmt.Errorf("scripted: an unknown Planner prompt")
	}
	s.mu.Lock()
	s.prompts = append(s.prompts, prompt)
	s.answers = append(s.answers, out)
	s.mu.Unlock()
	return out, nil
}

// sinkRiskRow is the one risk the change holds.
func sinkRiskRow() map[string]any {
	return map[string]any{
		"id": "R1", "rule_id": "unvalidated-exec", "title": sinkRisk,
		"message":  "Handle passes user input to sink, and sink runs it with exec without validating it.",
		"severity": "ERROR", "path": "pkg/f20.go", "start_line": 36, "end_line": 36,
	}
}

// stepAnswer is a model at one bounded step: it reads only the part and the carried state in the prompt. File 03
// raises the claim; the claim is carried while it waits; file 20 settles it, but only if the claim reached it.
func stepAnswer(prompt string) string {
	carried := strings.Contains(prompt, "awaiting evidence: "+sinkClaim)
	ans := map[string]any{"risks": []any{}, "established": []string{}, "suspected": []string{}, "awaiting_evidence": []string{}}
	switch {
	case isLastStep(prompt):
		// No part is left to supply evidence, so a model that is done leaves nothing open (the terminal check of
		// ADR-0048 stops a review that ends with an open claim).
	case strings.Contains(prompt, unique(3)):
		ans["awaiting_evidence"] = []string{sinkClaim}
	case strings.Contains(prompt, unique(20)) && carried:
		ans["risks"] = []any{sinkRiskRow()}
		ans["established"] = []string{"sink runs exec on its input unvalidated; reported as R1"}
	case strings.Contains(prompt, unique(20)):
		// Without the claim, sink alone is a function that runs what it is given: suspected, not a risk.
		ans["suspected"] = []string{"sink runs exec on its argument; no caller seen"}
	case carried:
		ans["awaiting_evidence"] = []string{sinkClaim}
	}
	b, _ := json.Marshal(ans)
	return string(b)
}

// lastStepRe matches the carried block's line for the last part.
var lastStepRe = regexp.MustCompile(`REMAINING: 1 of \d+ parts?, this one included`)

// isLastStep reports whether the prompt is the last bounded step's.
func isLastStep(prompt string) bool { return lastStepRe.MatchString(prompt) }

// boundedReview reviews the fixture at root in sample mode with the bounded-steps setting s and returns the draft,
// the scripted Planner and the trace's lines.
func boundedReview(t *testing.T, root string, s markov.Settings) (candidate.Draft, *scriptedPlanner, []trace.Event) {
	t.Helper()
	gen := &scriptedPlanner{}
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
	d, err := p.Review(context.Background(), Request{Target: root, FixesOff: true})
	if err != nil {
		t.Fatal(err)
	}
	return d, gen, readTrace(t, d.TracePath)
}

// linesOf returns the trace lines of one kind.
func linesOf(events []trace.Event, kind string) []trace.Event {
	var out []trace.Event
	for _, e := range events {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// uniqueRe matches a file marker.
var uniqueRe = regexp.MustCompile(`UNIQUE-f\d\d`)

// filesIn lists, once each and in order of first appearance, the files a prompt carries; a file's marker stands in
// its base and in its change.
func filesIn(prompt string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range uniqueRe.FindAllString(prompt, -1) {
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	return out
}

// on is the engine switched on with a threshold the 24-file fixture is far above (about 36,000 estimated tokens),
// and the 3-file fixture (about 4,600) far below, so neither sits near the edge.
func on(unit string) markov.Settings {
	return markov.Settings{On: true, Threshold: 6000, Unit: unit, Carry: markov.DefaultCarry}
}

// TestALargeChangeIsReviewedOnePartPerStep proves acceptance 3(a): a 24-file change above the threshold is read in
// 24 Planner steps, each prompt carrying exactly one file and the carried state and never another file; the trace
// says bounded steps ran and gives each step's carried-state size. With the engine off it is one prompt with every
// file, and this test fails.
func TestALargeChangeIsReviewedOnePartPerStep(t *testing.T) {
	const n = 24
	_, gen, events := boundedReview(t, boundedFixture(t, n), on(markov.UnitFile))
	if len(gen.prompts) != n {
		t.Fatalf("%d Planner calls, want %d (one per file)", len(gen.prompts), n)
	}
	for i, p := range gen.prompts {
		got := filesIn(p)
		if len(got) != 1 || got[0] != unique(i+1) {
			t.Fatalf("step %d carries %v, want only %s", i+1, got, unique(i+1))
		}
		if !strings.Contains(p, `"carried_state":"COMPLETED: `) {
			t.Fatalf("step %d has no carried state", i+1)
		}
	}
	ran := linesOf(events, traceKindBounded)
	if len(ran) != 1 || !strings.HasPrefix(ran[0].Text, fmt.Sprintf("bounded steps ran: %d of %d steps, unit file", n, n)) {
		t.Fatalf("%#v", ran)
	}
	steps := linesOf(events, traceKindStep)
	if len(steps) != n || !strings.Contains(steps[4].Text, "carried state") || steps[4].Path != "pkg/f05.go" {
		t.Fatalf("%d step lines; fifth %#v", len(steps), steps[4])
	}
}

// TestAClaimAwaitingEvidenceIsReportedWhenItsEvidenceIsRead proves acceptance 3(b): file 03 can only raise a claim
// that waits for evidence; it is carried, as "awaiting evidence", through every step up to file 20, which never sees
// file 03, and the risk is named there and reaches the review. Dropping the awaiting list from the carried state
// loses the risk.
func TestAClaimAwaitingEvidenceIsReportedWhenItsEvidenceIsRead(t *testing.T) {
	d, gen, events := boundedReview(t, boundedFixture(t, 24), on(markov.UnitFile))
	if len(gen.prompts) < 20 {
		t.Fatalf("%d Planner calls: bounded steps did not run", len(gen.prompts))
	}
	for i := 3; i < 20; i++ {
		if !strings.Contains(gen.prompts[i], "awaiting evidence: "+sinkClaim) {
			t.Fatalf("step %d does not carry the claim", i+1)
		}
	}
	if strings.Contains(gen.prompts[19], unique(3)) {
		t.Fatal("the step of file 20 saw file 03: this test would not prove the state carried the claim")
	}
	var claimed bool
	for _, e := range linesOf(events, "claim") {
		if e.FindingID == "R1" && e.Text == sinkRisk {
			claimed = true
		}
	}
	if !claimed {
		t.Fatalf("the risk the carried claim settles was not reported; findings %d", len(d.Findings))
	}
}

// TestASmallChangeIsOneWholeStep proves acceptance 3(c): a change at or below the threshold is one Planner call
// with every file, and the trace says why bounded steps did not run. Ignoring the threshold makes it 3 steps and
// this test red.
func TestASmallChangeIsOneWholeStep(t *testing.T) {
	_, gen, events := boundedReview(t, boundedFixture(t, 3), on(markov.UnitFile))
	if len(gen.prompts) != 1 || len(filesIn(gen.prompts[0])) != 3 || !strings.Contains(gen.prompts[0], wholeTask) {
		t.Fatalf("%d Planner calls; want one whole prompt with all 3 files", len(gen.prompts))
	}
	ran := linesOf(events, traceKindBounded)
	if len(ran) != 1 || !strings.Contains(ran[0].Text, "bounded steps did not run: the change is") || !strings.Contains(ran[0].Text, "not above the threshold 6000") {
		t.Fatalf("%#v", ran)
	}
	if len(linesOf(events, traceKindStep)) != 0 {
		t.Fatal("step lines for a whole review")
	}
}

// TestTheThresholdIsStrictlyAbove proves the edge of ADR-0048 point 1: a change exactly at the threshold is whole,
// one token less of threshold cuts it.
func TestTheThresholdIsStrictlyAbove(t *testing.T) {
	root := boundedFixture(t, 3)
	files, err := change.Read(root)
	if err != nil {
		t.Fatal(err)
	}
	size := planner.Size(files)
	at := on(markov.UnitFile)
	at.Threshold = size
	if _, gen, _ := boundedReview(t, root, at); len(gen.prompts) != 1 {
		t.Fatalf("at the threshold: %d calls", len(gen.prompts))
	}
	at.Threshold = size - 1
	if _, gen, _ := boundedReview(t, root, at); len(gen.prompts) != 3 {
		t.Fatalf("one below: %d calls", len(gen.prompts))
	}
}

// TestWithTheSwitchOffTheReviewIsWholeAndTheTraceSaysSo proves the unplug (law #41, acceptance 5): the same 24-file
// change with the engine off is one whole Planner prompt, as before ADR-0048, and the trace names that bounded
// steps did not run and why.
func TestWithTheSwitchOffTheReviewIsWholeAndTheTraceSaysSo(t *testing.T) {
	d, gen, events := boundedReview(t, boundedFixture(t, 24), markov.Settings{})
	if len(gen.prompts) != 1 || len(filesIn(gen.prompts[0])) != 24 {
		t.Fatalf("%d Planner calls", len(gen.prompts))
	}
	ran := linesOf(events, traceKindBounded)
	if len(ran) != 1 || ran[0].Text != "bounded steps did not run: the engine is off (PROVEN_MARKOVIAN is off); the Planner read the change whole" {
		t.Fatalf("%#v", ran)
	}
	// The whole review still finds the risk: off is the review as it was, not a weaker one.
	if len(linesOf(events, "claim")) != 1 || len(d.Findings) > 1 {
		t.Fatalf("claims %d", len(linesOf(events, "claim")))
	}
}

// TestHunkStepsReadOneChangedRegionEach proves the hunk unit: one file with two changed regions far apart is two
// steps, each prompt carrying one region as a hunk and not the other.
func TestHunkStepsReadOneChangedRegionEach(t *testing.T) {
	root := t.TempDir()
	var base strings.Builder
	for i := 1; i <= 200; i++ {
		fmt.Fprintf(&base, "// line %03d of a long file that exists to push the change over the threshold\n", i)
	}
	chg := strings.Replace(strings.Replace(base.String(), "line 010 ", "line 010 CHANGED-TOP ", 1), "line 190 ", "line 190 CHANGED-BOTTOM ", 1)
	writeFile(t, filepath.Join(root, "base", "big.go"), base.String())
	writeFile(t, filepath.Join(root, "change", "big.go"), chg)
	files, err := change.Read(root)
	if err != nil {
		t.Fatal(err)
	}
	s := on(markov.UnitHunk)
	s.Threshold = 100
	gen := &scriptedPlanner{}
	if _, _, err := planner.PlanBounded(context.Background(), files, gen, s); err != nil {
		t.Fatal(err)
	}
	if len(gen.prompts) != 2 {
		t.Fatalf("%d steps, want 2 hunks", len(gen.prompts))
	}
	if !strings.Contains(gen.prompts[0], "CHANGED-TOP") || strings.Contains(gen.prompts[0], "CHANGED-BOTTOM") ||
		!strings.Contains(gen.prompts[1], "CHANGED-BOTTOM") || !strings.Contains(gen.prompts[1], `"hunk":"@@ -187,7 +187,7 @@`) {
		t.Fatalf("hunk prompts:\n%s\n---\n%s", gen.prompts[0], gen.prompts[1])
	}
}

// fixPrompts wraps the fix scenario's recorded generator and keeps every Fixer prompt.
type fixPrompts struct {
	fixerCounter
	prompts []string
}

func (f *fixPrompts) Complete(ctx context.Context, agent, prompt string) (string, error) {
	if agent == "fixer" {
		f.prompts = append(f.prompts, prompt)
	}
	return f.fixerCounter.Complete(ctx, agent, prompt)
}

// TestTheFixLoopCarriesOnlyWhatFailedAndWhy proves acceptance 3(d), ADR-0048 point 2: in the fix scenario F2's
// first attempt fails; with the engine on, the second attempt's prompt carries that failure and its reason in the
// carried state and nothing the first attempt proposed (neither what it changed nor its edits). With the engine
// off the second prompt carries the first attempt whole, and this test fails.
func TestTheFixLoopCarriesOnlyWhatFailedAndWhy(t *testing.T) {
	fx := filepath.Join(repoRoot(t), "testdata", "fixture", "fix")
	gen := &fixPrompts{}
	spy := &fixSpy{}
	reviewFixtureWith(t, fx, spy, gen, on(markov.UnitFile))
	var f2 []string
	for _, p := range gen.prompts {
		if strings.Contains(p, "id: F2\n") {
			f2 = append(f2, p)
		}
	}
	if len(f2) != 2 {
		t.Fatalf("%d Fixer prompts for F2, want 2", len(f2))
	}
	second := f2[1]
	if !strings.Contains(second, "carried state:\nCOMPLETED: 1 attempt, none proved") ||
		!strings.Contains(second, "- attempt 1 failed: the proving test does not pass on the change with the patch applied") {
		t.Fatalf("the second attempt does not carry the failure and its reason:\n%s", second)
	}
	if strings.Contains(second, "it changed:") || strings.Contains(second, "\nold: ") || strings.Contains(second, "\nnew: ") {
		t.Fatalf("the second attempt carries what the first proposed:\n%s", second)
	}
}

// reviewFixtureWith reviews a sample fixture with the bounded-steps setting s; it is reviewFixture with Markov set.
func reviewFixtureWith(t *testing.T, root string, spy *fixSpy, gen *fixPrompts, s markov.Settings) candidate.Draft {
	t.Helper()
	fx, err := fixture.LoadUnder(change.ProcessPolicy(), root)
	if err != nil {
		t.Fatal(err)
	}
	spy.Sample = testrun.NewSample(fx.Tests, fx.Fixes...)
	gen.inner, err = generator.NewSample(fx.Answers)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Build(Options{Sample: true, LogDir: t.TempDir(), TestRun: spy, Generator: gen, Markov: s})
	if err != nil {
		t.Fatal(err)
	}
	d, err := p.Review(context.Background(), Request{Target: root})
	if err != nil {
		t.Fatal(err)
	}
	return d
}
