package review_test

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestSampleReviewNamesEachFindingsLevel proves that on the sample review the test-proved finding
// (F1) is named proved-by-test, the three data-proved findings (F2, F4 and the OSV match) are named
// proved-by-data, and the totals by level are 1, 0, 3, 0 in the order of strength (ADR-0044).
// Changing a finding's level in the reviewer turns this test red.
func TestSampleReviewNamesEachFindingsLevel(t *testing.T) {
	res, err := openSample(t, t.TempDir()).Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]review.ProofLevel{
		"F1":                       review.ProvedByTest,
		"F2":                       review.ProvedByData,
		"F4":                       review.ProvedByData,
		"osv:evil.example/leftpad": review.ProvedByData,
	}
	if len(res.Findings) != len(want) {
		t.Fatalf("%d findings; want %d", len(res.Findings), len(want))
	}
	for id, level := range want {
		f, ok := find(res, id)
		if !ok {
			t.Fatalf("%s missing", id)
		}
		if f.ProofLevel != level {
			t.Errorf("%s: level %q; want %q", id, f.ProofLevel, level)
		}
	}
	totals := map[review.ProofLevel]int{
		review.ProvedByTest: 1, review.ProvedByTool: 0, review.ProvedByData: 3, review.AgreedNotProven: 0,
	}
	for _, l := range review.ProofLevels {
		if res.ProofCounts[l] != totals[l] {
			t.Errorf("total for %s: %d; want %d", l, res.ProofCounts[l], totals[l])
		}
	}
}
