package markov

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
)

// TestModelTotalsReproducesTheDocumentedTable proves ModelTotals is docs/efficiency-model.md's formula: every
// row of its worked-example table, naive and Markovian totals exactly.
func TestModelTotalsReproducesTheDocumentedTable(t *testing.T) {
	rows := []struct{ n, naive, markovian int }{
		{5, 16500, 8500}, {10, 63000, 17000}, {20, 246000, 34000}, {50, 1515000, 85000}, {100, 6030000, 170000},
	}
	for _, r := range rows {
		got := ModelTotals(r.n, DocDefaults)
		if got.NaiveTotal != r.naive || got.MarkovianTotal != r.markovian {
			t.Errorf("N=%d: naive %d markovian %d, want %d %d", r.n, got.NaiveTotal, got.MarkovianTotal, r.naive, r.markovian)
		}
	}
	if e := ModelTotals(20, DocDefaults).Efficiency * 100; math.Abs(e-86.18) > 0.01 {
		t.Fatalf("N=20 efficiency %.2f%%, want 86.18%%", e)
	}
}

// TestRunEfficiencyUsesTheSnapshotItIsGiven proves the historical efficiency is the prefix-sum formula of
// AGENTS.md section 6.3 and moves with the snapshot's carryover, not with any current setting.
func TestRunEfficiencyUsesTheSnapshotItIsGiven(t *testing.T) {
	chunks := []Chunk{{ID: 1, Tokens: 1000}, {ID: 2, Tokens: 1000}, {ID: 3, Tokens: 1000}, {ID: 4, Tokens: 1000}}
	a := RunEfficiency(chunks, Config{CarryoverTokens: 512})
	if a.StandardCost != 10000 || a.MarkovianCost != 6048 || a.TokensSaved != 3952 {
		t.Fatalf("%+v", a)
	}
	b := RunEfficiency(chunks, Config{CarryoverTokens: 2048})
	if b.TokensSaved == a.TokensSaved {
		t.Fatal("two snapshots gave the same savings")
	}
	if z := RunEfficiency(nil, DefaultConfig()); z.EfficiencyPercent != 0 || z.StandardCost != 0 {
		t.Fatalf("empty run %+v", z)
	}
}

// TestValidateRefusesEachBoundByName proves a value outside the spec's bounds is refused with the field named and
// never clamped, and that a step budget too small for its framing, state and answer is refused.
func TestValidateRefusesEachBoundByName(t *testing.T) {
	if err := DefaultConfig().Validate(); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		field string
		edit  func(*Config)
	}{
		{"chunkSize", func(c *Config) { c.ChunkSize = 128 }},
		{"chunkSize", func(c *Config) { c.ChunkSize = 128001 }},
		{"maxChunks", func(c *Config) { c.MaxChunks = 0 }},
		{"maxChunks", func(c *Config) { c.MaxChunks = 33 }},
		{"carryoverTokens", func(c *Config) { c.CarryoverTokens = 127 }},
		{"overlapTokens", func(c *Config) { c.OverlapTokens = 4097 }},
		{"chunkSize", func(c *Config) { c.ChunkSize = 2000; c.CarryoverTokens = 800 }},
	}
	for _, tc := range cases {
		c := DefaultConfig()
		tc.edit(&c)
		err := c.Validate()
		var ce *ConfigError
		if !errors.As(err, &ce) || ce.Field != tc.field || len(ce.Allowed) == 0 {
			t.Errorf("%+v: %v", c, err)
		}
	}
}

// TestMarkersAreParsedAndCleaned proves the state is read after the state marker, completion is read from the
// final marker, neither reaches display text, and a bare [STATE] after an underscore is not a marker.
func TestMarkersAreParsedAndCleaned(t *testing.T) {
	m := ParseMarkers("work done\n[STATE_CHECKPOINT]\nCOMPLETED: a\nREMAINING: b")
	if !m.HasState || m.State != "COMPLETED: a\nREMAINING: b" || m.Done || m.Cleaned != "work done" {
		t.Fatalf("%+v", m)
	}
	f := ParseMarkers("all done [TASK_COMPLETE]")
	if !f.Done || f.HasState || f.Cleaned != "all done" {
		t.Fatalf("%+v", f)
	}
	if p := ParseMarkers("a FOO_[STATE] b"); p.HasState {
		t.Fatalf("%+v", p)
	}
	if old := ParseMarkers("x\n@@@STATE@@@ carried on in the older spelling"); !old.HasState || old.Cleaned != "x" {
		t.Fatalf("%+v", old)
	}
}

// TestCascadeFallsThroughInOrder proves the four levels of AGENTS.md section 4.3 run in order: a long enough
// marker state wins, a short one falls to the compressor, a failed compressor to key phrases, and no phrase to the
// word tail within the budget.
func TestCascadeFallsThroughInOrder(t *testing.T) {
	long := "COMPLETED: step one of the plan"
	if s, l := NextCarryover(long, true, "", "", 800, nil); l != LevelMarker || s != long {
		t.Fatalf("%s %q", l, s)
	}
	comp := func(string, string) (string, error) { return "COMPLETED: compressed by the model", nil }
	if _, l := NextCarryover("short", true, "", "body", 800, comp); l != LevelModel {
		t.Fatal(l)
	}
	broken := func(string, string) (string, error) { return "", errors.New("down") }
	if s, l := NextCarryover("", false, "", "Intro. Status: half done. More.", 800, broken); l != LevelHeuristic || !strings.Contains(s, "Status: half done") {
		t.Fatalf("%s %q", l, s)
	}
	body := strings.Repeat("word ", 2000)
	s, l := NextCarryover("", false, "", body, 128, nil)
	if l != LevelTruncation || EstimateTokens(s) > 128 || s == "" {
		t.Fatalf("%s %d", l, EstimateTokens(s))
	}
}

// scripted answers step n with text that ends in a state marker, and the last step with the final marker. It
// records every prompt it was given.
type scripted struct {
	total   int
	prompts []string
	failAt  int
}

func (s *scripted) gen(_ context.Context, prompt, _ string) (string, error) {
	s.prompts = append(s.prompts, prompt)
	n := len(s.prompts)
	if n == s.failAt {
		return "", errors.New("provider down")
	}
	body := strings.Repeat(fmt.Sprintf("output of step %d. ", n), 10)
	if n >= s.total {
		return body + MarkerFinal, nil
	}
	return body + "\n" + MarkerState + fmt.Sprintf("\nCOMPLETED: steps 1-%d\nREMAINING: %d", n, s.total-n), nil
}

// clock is a fixed clock that advances one millisecond per read.
func clock() func() int64 {
	var t int64 = 1_750_000_000_000
	return func() int64 { t++; return t }
}

// TestExecuteCarriesOnlyTheStateBetweenSteps proves the orchestrator loop: each continuation prompt carries the
// previous step's state and none of the earlier steps' output, the final marker ends the run, and the record
// carries its snapshot and its label.
func TestExecuteCarriesOnlyTheStateBetweenSteps(t *testing.T) {
	s := &scripted{total: 3}
	run, err := Execute(context.Background(), DefaultConfig(), Options{Goal: "write the thing"}, s.gen, clock())
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != StatusCompleted || len(run.Chunks) != 3 || run.TokenLabel != LabelEstimated || run.ChunkConfig != DefaultConfig() {
		t.Fatalf("%+v", run)
	}
	// Three steps and no reflection: Options.Reflect is false.
	if len(s.prompts) != 3 {
		t.Fatalf("%d prompts", len(s.prompts))
	}
	if !strings.Contains(s.prompts[2], "COMPLETED: steps 1-2") || strings.Contains(s.prompts[2], "output of step 1") {
		t.Fatalf("step 3 prompt:\n%s", s.prompts[2])
	}
	if strings.Contains(run.Output, MarkerState) || strings.Contains(run.Output, MarkerFinal) {
		t.Fatal("markers in the output")
	}
}

// TestExecuteStopsAtMaxChunksAndKeepsAFailedStep proves the recursion cap ends a run that never completes as
// max-chunks, and a step that fails ends it as errored with the earlier steps kept.
func TestExecuteStopsAtMaxChunksAndKeepsAFailedStep(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxChunks = 2
	run, _ := Execute(context.Background(), cfg, Options{Goal: "g"}, (&scripted{total: 9}).gen, clock())
	if run.Status != StatusMaxChunks || len(run.Chunks) != 2 {
		t.Fatalf("%s %d", run.Status, len(run.Chunks))
	}
	run, _ = Execute(context.Background(), DefaultConfig(), Options{Goal: "g"}, (&scripted{total: 9, failAt: 3}).gen, clock())
	if run.Status != StatusErrored || len(run.Chunks) != 2 || run.Error == "" {
		t.Fatalf("%s %d %q", run.Status, len(run.Chunks), run.Error)
	}
}

// TestReflectionFailureKeepsTheRunCompleted proves AGENTS.md section 7: the reflection runs after a completed
// run, and its failure is recorded without failing the run.
func TestReflectionFailureKeepsTheRunCompleted(t *testing.T) {
	s := &scripted{total: 1, failAt: 2}
	run, _ := Execute(context.Background(), DefaultConfig(), Options{Goal: "g", Reflect: true}, s.gen, clock())
	if run.Status != StatusCompleted || !strings.Contains(run.Error, "reflection failed") || len(s.prompts) != 2 {
		t.Fatalf("%s %q %d", run.Status, run.Error, len(s.prompts))
	}
}

// TestOverlapWindowIsCarriedWhenSet proves the overlap contract is implemented, not ignored: with overlapTokens
// set, the next prompt carries the [OVERLAP] block with the tail of the previous output, and with zero it does not.
func TestOverlapWindowIsCarriedWhenSet(t *testing.T) {
	cfg := DefaultConfig()
	cfg.OverlapTokens = 8
	s := &scripted{total: 2}
	if _, err := Execute(context.Background(), cfg, Options{Goal: "g"}, s.gen, clock()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s.prompts[1], "[OVERLAP]\n") || !strings.Contains(s.prompts[1], "step 1.") {
		t.Fatalf("%s", s.prompts[1])
	}
	off := &scripted{total: 2}
	if _, err := Execute(context.Background(), DefaultConfig(), Options{Goal: "g"}, off.gen, clock()); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(off.prompts[1], "[OVERLAP]") {
		t.Fatal("overlap block with zero overlap")
	}
}

// TestCancelledContextAbortsTheRun proves the abort is checked before each step: a context already ended
// produces an aborted run with no step.
func TestCancelledContextAbortsTheRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	run, err := Execute(ctx, DefaultConfig(), Options{Goal: "g", Reflect: true}, (&scripted{total: 2}).gen, clock())
	if err != nil || run.Status != StatusAborted || len(run.Chunks) != 0 {
		t.Fatalf("%v %s %d", err, run.Status, len(run.Chunks))
	}
}
