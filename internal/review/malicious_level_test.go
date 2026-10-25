package review_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// TestMaliciousPackageFixtureIsProvedByData proves the malicious-package
// sample review keeps one finding, names it proved-by-data, and counts that
// level once. The mutated field is the EvidenceQuery arm of ProofLevelOf:
// mapping that arm to ProvedByTool turns this test red.
func TestMaliciousPackageFixtureIsProvedByData(t *testing.T) {
	target := filepath.Join(repoRoot(t), "testdata", "fixture", "malicious-package")
	res, err := openSample(t, t.TempDir()).Review(context.Background(), review.Request{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 1 {
		t.Fatalf("%d findings, want 1", len(res.Findings))
	}
	if res.Findings[0].ProofLevel != review.ProvedByData {
		t.Fatalf("level %q, want %q", res.Findings[0].ProofLevel, review.ProvedByData)
	}
	if res.ProofCounts[review.ProvedByData] != 1 {
		t.Fatalf("proved-by-data count %d, want 1", res.ProofCounts[review.ProvedByData])
	}
}
