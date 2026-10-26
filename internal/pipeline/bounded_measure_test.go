package pipeline

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/markov"
	"github.com/saphan-ai/proven/internal/trace"
)

// TestBoundedStepsMeasurement is ADR-0048 point 4's measurement on the sample scenarios. It runs each scenario
// whole and in bounded steps and logs, per review, the Planner's model tokens per step (prompt plus answer) and in
// total, and whether the two reviews named the same risks. Every figure is the scripted generator's, estimated as
// characters / 4: it measures the size of what Proven sends and receives, never a live model's behaviour.
//
// The property it holds: with the engine on, no step of a change above the threshold is as large as the whole
// review's one step, and the findings are the same. It does not hold, and does not claim, that the total is
// smaller: each step repeats the instruction and the carried state.
func TestBoundedStepsMeasurement(t *testing.T) {
	scenarios := []struct {
		name  string
		files int
		unit  string
	}{
		{"24 files, file steps", 24, markov.UnitFile},
		{"24 files, hunk steps", 24, markov.UnitHunk},
		{"60 files, file steps", 60, markov.UnitFile},
		{"3 files, below the threshold", 3, markov.UnitFile},
	}
	var rows []string
	for _, sc := range scenarios {
		root := boundedFixture(t, sc.files)
		_, whole, wholeEvents := boundedReview(t, root, markov.Settings{})
		_, cut, cutEvents := boundedReview(t, root, on(sc.unit))
		w, c := tokensPerStep(whole), tokensPerStep(cut)
		same := sameStrings(claims(wholeEvents), claims(cutEvents))
		rows = append(rows, fmt.Sprintf("| %s | %d | %d | %d | %d | %d | %d | %d | %t |",
			sc.name, len(w), maxOf(w), sumOf(w), len(c), maxOf(c), sumOf(c)/len(c), sumOf(c), same))
		if !same {
			t.Errorf("%s: the bounded review named %v, the whole one %v", sc.name, claims(cutEvents), claims(wholeEvents))
		}
		if len(c) > 1 && maxOf(c) >= maxOf(w) {
			t.Errorf("%s: a bounded step (%d) is not smaller than the whole step (%d)", sc.name, maxOf(c), maxOf(w))
		}
	}
	t.Logf("\n| scenario (scripted generator) | whole: steps | whole: tokens per step | whole: tokens per review | bounded: steps | bounded: largest step | bounded: mean step | bounded: tokens per review | same findings |\n|---|---|---|---|---|---|---|---|---|\n%s",
		strings.Join(rows, "\n"))
}

// tokensPerStep is each Planner call's prompt and answer, in estimated tokens.
func tokensPerStep(g *scriptedPlanner) []int {
	out := make([]int, len(g.prompts))
	for i := range g.prompts {
		out[i] = markov.EstimateTokens(g.prompts[i]) + markov.EstimateTokens(g.answers[i])
	}
	return out
}

// claims is the sorted titles of the Planner's claims in a trace.
func claims(events []trace.Event) []string {
	var out []string
	for _, e := range linesOf(events, "claim") {
		out = append(out, e.Text)
	}
	sort.Strings(out)
	return out
}

func sameStrings(a, b []string) bool {
	return strings.Join(a, "\x00") == strings.Join(b, "\x00")
}

func maxOf(v []int) int {
	m := 0
	for _, x := range v {
		m = max(m, x)
	}
	return m
}

func sumOf(v []int) int {
	s := 0
	for _, x := range v {
		s += x
	}
	return s
}
