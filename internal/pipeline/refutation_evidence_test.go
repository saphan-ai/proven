package pipeline

import (
	"path/filepath"
	"testing"
)

// Property: a model's word alone is not a refutation. The skeptic answers
// refuted=true with no refuting test and no refuting query; the finding
// stays on the next review of the same change.
// Mutated field: the skeptic's answer for F1.
func TestSkepticWordAloneDoesNotSuppressTheNextReview(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	gen := refuteGenerator(t, "F1")
	first := reviewWith(t, filepath.Join(root, "logs1"), state, gen)
	if _, ok := kept(first, "F1"); !ok {
		t.Fatal("precondition: F1 must be kept on the first review")
	}
	second := reviewWith(t, filepath.Join(root, "logs2"), state, gen)
	if _, ok := kept(second, "F1"); !ok {
		t.Fatalf("F1 was dropped on the second review by a refutation with no evidence (H-MEM record %q)",
			memoryID(t, second.TracePath, "F1"))
	}
}
