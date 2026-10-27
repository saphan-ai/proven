package pipeline

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

// A live review hands its Planner only the files that changed, and a sample review every file of its fixture, whose
// recorded answers are keyed by that prompt. Mutated field: the review's sample flag.
func TestOnlyALiveReviewReadsTheChangedFilesOnly(t *testing.T) {
	files := []change.File{
		{Path: "same", Base: []byte("x"), Change: []byte("x")},
		{Path: "edit", Base: []byte("x"), Change: []byte("y")},
	}
	if got := change.ChangedFor(readContext(context.Background(), false), files); len(got) != 1 || got[0].Path != "edit" {
		t.Fatalf("live review read %v", got)
	}
	if got := change.ChangedFor(readContext(context.Background(), true), files); len(got) != 2 {
		t.Fatalf("sample review read %v", got)
	}
}
