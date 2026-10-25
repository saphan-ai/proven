package writer

import (
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
)

// TestSamePlaceIsThePathAndAnOverlappingRange proves the place rule of ADR-0041.
// Lines 4-8 and 6-10 share 6-8. Lines that only meet at a boundary do share that
// line. Lines 1-4 and 5-9 share none. A different path, a missing line, or an
// empty path is not the same place. The rule id is not the place.
func TestSamePlaceIsThePathAndAnOverlappingRange(t *testing.T) {
	a := candidate.Risk{RuleID: "one", Path: "a.go", StartLine: 4, EndLine: 8}
	b := candidate.Risk{RuleID: "two", Path: "a.go", StartLine: 6, EndLine: 10}
	if !samePlace(a, b) {
		t.Fatal("overlapping ranges of different rules are one place")
	}
	touch := candidate.Risk{Path: "a.go", StartLine: 8, EndLine: 9}
	if !samePlace(a, touch) {
		t.Fatal("a shared boundary line is an overlap")
	}
	apart := candidate.Risk{Path: "a.go", StartLine: 9, EndLine: 12}
	if samePlace(a, apart) {
		t.Fatal("ranges that do not share a line were called one place")
	}
	other := candidate.Risk{Path: "b.go", StartLine: 4, EndLine: 8}
	if samePlace(a, other) {
		t.Fatal("the same lines in another file were called one place")
	}
	for _, bad := range []candidate.Risk{
		{Path: "a.go", StartLine: 0, EndLine: 8},
		{Path: "a.go", StartLine: 4, EndLine: 3},
		{Path: "", StartLine: 4, EndLine: 8},
	} {
		if samePlace(a, bad) {
			t.Fatalf("a range that names no place overlapped: %+v", bad)
		}
	}
}

// TestAgreedOpinionsLiftOnlyWhatBothReviewersReported proves the lift.
// Both reported on one candidate is one opinion. Two dropped claims whose
// ranges only overlap are one opinion, and the place shown is the shared lines.
// A claim only one reviewer reported, a refuted claim, and a kept finding are
// not opinions.
func TestAgreedOpinionsLiftOnlyWhatBothReviewersReported(t *testing.T) {
	both := dropped("B1", "a.go", 1, 1)
	both.Verdicts = []candidate.Verdict{
		{Run: runDefault, Model: "NVIDIA", Verdict: candidate.VerdictReported, Reason: "default words"},
		{Run: runSpecial, Model: "other", Verdict: candidate.VerdictReported, Reason: "special words"},
	}
	left := dropped("L1", "a.go", 4, 8)
	left.Verdicts = []candidate.Verdict{
		{Run: runDefault, Model: "NVIDIA", Verdict: candidate.VerdictReported, Reason: "only default"},
		{Run: runSpecial, Model: "other", Verdict: candidate.VerdictNotReported, Reason: "not named"},
	}
	right := dropped("R1", "a.go", 6, 10)
	right.Title = "special title"
	right.Verdicts = []candidate.Verdict{
		{Run: runDefault, Model: "NVIDIA", Verdict: candidate.VerdictNotReported, Reason: "not named"},
		{Run: runSpecial, Model: "other", Verdict: candidate.VerdictReported, Reason: "only special"},
	}
	lone := dropped("N1", "a.go", 30, 32)
	lone.Verdicts = []candidate.Verdict{
		{Run: runDefault, Model: "NVIDIA", Verdict: candidate.VerdictReported, Reason: "alone"},
		{Run: runSpecial, Model: "other", Verdict: candidate.VerdictNotReported, Reason: "not named"},
	}
	refuted := dropped("F1", "a.go", 1, 2)
	refuted.Verdicts = []candidate.Verdict{
		{Run: runDefault, Model: "NVIDIA", Verdict: candidate.VerdictReported, Reason: "keeps"},
		{Run: runSpecial, Model: "other", Verdict: candidate.VerdictRefuted, Reason: "refutes"},
	}
	kept := candidate.Candidate{Risk: candidate.Risk{ID: "K1", Path: "a.go", StartLine: 1, EndLine: 1}}
	kept.Verdicts = both.Verdicts

	got := agreedOpinions([]candidate.Candidate{both, left, right, lone, refuted, kept})
	if len(got) != 2 {
		t.Fatalf("opinions %d, want 2: %+v", len(got), got)
	}
	if got[0].ID != "B1" || got[0].Reasons[0].Reason != "default words" || got[0].Reasons[1].Reason != "special words" {
		t.Fatalf("merged opinion %+v", got[0])
	}
	if got[1].ID != "L1" || got[1].StartLine != 6 || got[1].EndLine != 8 || got[1].Title != left.Title {
		t.Fatalf("overlap %+v", got[1])
	}
	if got[1].Reasons[0].Reason != "only default" || got[1].Reasons[1].Reason != "only special" {
		t.Fatalf("overlap reasons %+v", got[1].Reasons)
	}
}

// dropped is a claim the Judge dropped for no evidence.
func dropped(id, path string, start, end int) candidate.Candidate {
	return candidate.Candidate{
		Risk:       candidate.Risk{ID: id, Title: id, Path: path, StartLine: start, EndLine: end},
		Dropped:    true,
		DropReason: "dropped: no evidence",
	}
}
