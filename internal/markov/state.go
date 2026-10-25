package markov

import (
	"fmt"
	"strings"
)

// Block is the spec's canonical carryover shape (AGENTS.md section 3): COMPLETED, REMAINING, DECISIONS and
// CONTEXT. Proven's two users fill it: the review with what is established, suspected and awaiting evidence, the
// fix loop with what failed and why.
type Block struct {
	Completed string
	Remaining string
	Decisions []string
	Context   []string
}

// maxItem bounds one carried item, so a single long sentence cannot take the whole state.
const maxItem = 400

// Render writes the block as the labelled text a prompt carries.
func (b Block) Render() string {
	var s strings.Builder
	fmt.Fprintf(&s, "COMPLETED: %s\nREMAINING: %s\nDECISIONS:", b.Completed, b.Remaining)
	writeItems(&s, b.Decisions)
	s.WriteString("\nCONTEXT:")
	writeItems(&s, b.Context)
	return s.String()
}

// writeItems writes one item per line, or "none".
func writeItems(s *strings.Builder, items []string) {
	if len(items) == 0 {
		s.WriteString(" none")
		return
	}
	for _, it := range items {
		s.WriteString("\n- " + it)
	}
}

// Tokens is the block's estimated size as a prompt carries it.
func (b Block) Tokens() int {
	return EstimateTokens(b.Render())
}

// Fit bounds the block to budget estimated tokens and returns how many items it dropped. Each item is first cut
// to maxItem bytes. Then the oldest decision goes first, and the oldest context item only when no decision is left:
// what a later step still has to act on is kept longest. A block that cannot fit even with every item gone keeps
// its two summary lines, which are the step count and nothing a model wrote.
func (b Block) Fit(budget int) (Block, int) {
	out := Block{Completed: b.Completed, Remaining: b.Remaining}
	for _, it := range b.Decisions {
		out.Decisions = append(out.Decisions, clipItem(it))
	}
	for _, it := range b.Context {
		out.Context = append(out.Context, clipItem(it))
	}
	dropped := 0
	for out.Tokens() > budget {
		switch {
		case len(out.Decisions) > 0:
			out.Decisions = out.Decisions[1:]
		case len(out.Context) > 0:
			out.Context = out.Context[1:]
		default:
			return out, dropped
		}
		dropped++
	}
	return out, dropped
}

// clipItem trims an item, makes it one line and bounds it at maxItem bytes on a rune boundary.
func clipItem(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= maxItem {
		return s
	}
	return clipRunes(s, maxItem) + " [cut]"
}

// The labels that mark which list a review item belongs to, so the model reads them back unambiguously.
const (
	labelEstablished = "established: "
	labelSuspected   = "suspected: "
	labelAwaiting    = "awaiting evidence: "
)

// ReviewState is the review's carried state (ADR-0048 point 1): what is already established, what is suspected,
// and which claims wait for evidence from a part not yet read.
type ReviewState struct {
	Established []string `json:"established"`
	Suspected   []string `json:"suspected"`
	Awaiting    []string `json:"awaiting_evidence"`
}

// Block maps the review's state onto the spec's shape for the step about to run: COMPLETED and REMAINING count the
// parts; what is established is a DECISION; what is suspected and what awaits evidence is CONTEXT, the awaiting
// claims last, so the bound drops them last.
func (r ReviewState) Block(step, steps int) Block {
	b := Block{
		Completed: completedText(step-1, "part"),
		Remaining: fmt.Sprintf("%d of %d %s, this one included", steps-step+1, steps, plural(steps, "part")),
	}
	for _, e := range r.Established {
		b.Decisions = append(b.Decisions, labelEstablished+e)
	}
	for _, s := range r.Suspected {
		b.Context = append(b.Context, labelSuspected+s)
	}
	for _, a := range r.Awaiting {
		b.Context = append(b.Context, labelAwaiting+a)
	}
	return b
}

// completedText says how many units are done.
func completedText(n int, unit string) string {
	if n == 0 {
		return "none yet"
	}
	return fmt.Sprintf("%d %s", n, plural(n, unit))
}

// plural adds an s to word unless n is 1.
func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

// FixFailure is one failed attempt as the fix loop carries it: why it failed and, when a run failed, which run.
type FixFailure struct {
	Attempt int
	Reason  string
	// Run is the failing run's label, command and exit code with its output, already clipped; empty when no run
	// failed (the patch was refused before it ran).
	Run string
}

// FixBlock maps the fix loop's failures onto the spec's shape (ADR-0048 point 2): the earlier failures are
// DECISIONS, the latest one and its failing run are CONTEXT, so the bound drops the oldest first. Nothing an
// attempt proposed is carried, only what failed and why.
func FixBlock(failures []FixFailure, attempts int) Block {
	b := Block{
		Completed: completedText(len(failures), "attempt") + ", none proved",
		Remaining: fmt.Sprintf("%d of %d %s, this one included", attempts-len(failures), attempts, plural(attempts, "attempt")),
	}
	for i, f := range failures {
		line := fmt.Sprintf("attempt %d failed: %s", f.Attempt, f.Reason)
		if i < len(failures)-1 {
			b.Decisions = append(b.Decisions, line)
			continue
		}
		b.Context = append(b.Context, line)
		if f.Run != "" {
			b.Context = append(b.Context, "failing run of attempt "+fmt.Sprint(f.Attempt)+": "+f.Run)
		}
	}
	return b
}
