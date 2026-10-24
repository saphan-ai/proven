package review

import (
	"strings"
	"testing"
)

// riskPlaced is one planner risk at a path and a line range, with the reviewer's own message.
// The message is the reason the opinion must carry word for word (ADR-0041).
func riskPlaced(id, title, message, path string, start, end int) riskRow {
	return riskRow{
		ID: id, RuleID: id, Title: title, Message: message, Severity: "WARNING",
		Path: path, StartLine: start, EndLine: end,
	}
}

// TestAgreedOpinionIsShownApartAndNotCounted proves the recorded scenario of ADR-0041.
// Both reviewers name one claim at overlapping lines of the same file, and nothing proves it:
// it is an opinion, with both messages word for word, and it is not a finding.
// A claim only the default reviewer names, and one only the special reviewer names, stay dropped.
// The mutated field of the negative control is Result.ProofCounts: adding the opinion to that
// map turns this test red. The exit code is covered beside this, in the shell, because it is
// decided there from len(Findings) alone.
func TestAgreedOpinionIsShownApartAndNotCounted(t *testing.T) {
	const (
		defaultReason = "default: the bounds check is missing"
		specialReason = "special: the bounds check is missing"
		onlyDefault   = "only the default model names this"
		onlySpecial   = "only the special model names this"
	)
	def := newScript("NVIDIA", risks(
		riskPlaced("O1", "bounds check", defaultReason, "billing/discount.go", 4, 8),
		riskPlaced("L1", "lone default", onlyDefault, "billing/discount.go", 30, 32),
	), `{"refuted":false,"reason":"no refutation"}`)
	spec := newScript("qwen-special", risks(
		riskPlaced("O9", "bounds check too", specialReason, "billing/discount.go", 6, 10),
		riskPlaced("S1", "lone special", onlySpecial, "billing/discount.go", 40, 42),
	), `{"refuted":false,"reason":"no refutation"}`)
	cfg := defaultSettings()
	cfg.SpecialReviewer = "qwen-special"
	res := reviewScripted(t, cfg, def, spec)

	if len(res.Opinions) != 1 {
		t.Fatalf("opinions %d; want the one claim both reviewers named", len(res.Opinions))
	}
	op := res.Opinions[0]
	if op.Location.Path != "billing/discount.go" || op.Location.StartLine != 6 || op.Location.EndLine != 8 {
		t.Fatalf("place %+v; want billing/discount.go:6-8, the lines both ranges share", op.Location)
	}
	if op.Level != OpinionLevel {
		t.Fatalf("level %q; want %q", op.Level, OpinionLevel)
	}
	if len(op.Reasons) != 2 || op.Reasons[0].Role != RoleDefault || op.Reasons[1].Role != RoleSpecial {
		t.Fatalf("reasons %+v", op.Reasons)
	}
	if op.Reasons[0].Reason != defaultReason || op.Reasons[0].Model != "NVIDIA" {
		t.Fatalf("default reason %+v", op.Reasons[0])
	}
	if op.Reasons[1].Reason != specialReason || op.Reasons[1].Model != "qwen-special" {
		t.Fatalf("special reason %+v", op.Reasons[1])
	}
	for _, f := range res.Findings {
		if f.ID == "O1" || f.ID == "O9" || f.ID == "L1" || f.ID == "S1" {
			t.Fatalf("an unproven claim is a finding: %+v", f)
		}
		if strings.Contains(f.Message, defaultReason) || strings.Contains(f.Message, onlyDefault) || strings.Contains(f.Message, onlySpecial) {
			t.Fatalf("an unproven claim is mixed into findings: %+v", f)
		}
	}
	for _, other := range res.Opinions {
		for _, r := range other.Reasons {
			if r.Reason == onlyDefault || r.Reason == onlySpecial {
				t.Fatalf("a single-reviewer claim was shown: %+v", r)
			}
		}
	}
	if res.ProofCounts[AgreedNotProven] != 0 {
		t.Fatalf("agreed opinions counted in the totals: %d", res.ProofCounts[AgreedNotProven])
	}
	sum := 0
	for _, n := range res.ProofCounts {
		sum += n
	}
	if sum != len(res.Findings) {
		t.Fatalf("totals %d, findings %d; an opinion was counted", sum, len(res.Findings))
	}
}
