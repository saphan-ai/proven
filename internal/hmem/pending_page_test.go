package hmem_test

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// waitingProposals is the number of undecided proposals the page work-bound
// test puts in the store: far past what a review run makes.
const waitingProposals = 50000

// TestPendingPageWorkFollowsTheLimit proves R16 for a page: with 50,000
// proposals waiting, a one-item page, at the start and deep in the list, visits
// at most two proposals and allocates a few kilobytes, where a full read of the
// waiting set visits all of them and allocates megabytes. The full read is the
// negative control: the counter and the allocation must see it, or the bounds
// above prove nothing.
func TestPendingPageWorkFollowsTheLimit(t *testing.T) {
	s := openApprove(t, t.TempDir(), t.TempDir())
	if err := hmem.AppendProposalLines(s, waitingProposals); err != nil {
		t.Fatal(err)
	}
	// The first call builds the index from the log; the pages below are measured
	// after it, because the index build is a read of the log, not of a page.
	if _, _, err := s.PendingPage(0, 1); err != nil {
		t.Fatal(err)
	}
	visited := func(fn func()) int64 {
		before := hmem.ProposalsVisited(s)
		fn()
		return hmem.ProposalsVisited(s) - before
	}
	for _, cursor := range []int{0, waitingProposals / 2, waitingProposals - 2} {
		var got int
		v := visited(func() {
			page, _, err := s.PendingPage(cursor, 1)
			if err != nil {
				t.Fatal(err)
			}
			got = len(page)
		})
		if got != 1 || v > 2 {
			t.Fatalf("a one-item page at cursor %d returned %d and visited %d proposals; want 1 and at most 2", cursor, got, v)
		}
	}
	pageBytes := allocated(func() {
		if _, _, err := s.PendingPage(waitingProposals/2, 1); err != nil {
			t.Fatal(err)
		}
	})
	wholeBytes := allocated(func() { hmem.ScanAllWaiting(s) })
	if pageBytes > 16<<10 || pageBytes > wholeBytes/100 {
		t.Fatalf("a one-item page allocated %d bytes; a full read of the waiting set takes %d", pageBytes, wholeBytes)
	}

	// Negative control: the full read is visible to the counter.
	if v := visited(func() { hmem.ScanAllWaiting(s) }); v < waitingProposals {
		t.Fatalf("the control scan visited %d proposals, want at least %d: the counter does not see a scan", v, waitingProposals)
	}
}

// TestPendingCursorSurvivesADecision proves the page cursor is stable: a
// proposal decided between two page calls, one already seen and one not yet
// seen, does not make any other waiting proposal vanish from the pages. Every
// proposal that still waits is returned exactly once.
func TestPendingCursorSurvivesADecision(t *testing.T) {
	s := openApprove(t, t.TempDir(), t.TempDir())
	if err := hmem.AppendProposalLines(s, 6); err != nil {
		t.Fatal(err)
	}
	first, next, err := s.PendingPage(0, 1)
	if err != nil || len(first) != 1 || next < 0 {
		t.Fatalf("page one = %d proposals, next %d, err %v", len(first), next, err)
	}
	// Decide the page-one item, and one the caller has not reached yet.
	if err := s.Reject(first[0].ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Reject("dream-3"); err != nil {
		t.Fatal(err)
	}
	// The page-one item was returned before it was decided, so it counts as seen.
	seen := []string{first[0].ID}
	for cursor := next; cursor >= 0; {
		page, nxt, err := s.PendingPage(cursor, 1)
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range page {
			seen = append(seen, p.ID)
		}
		cursor = nxt
	}
	want := []string{"dream-1", "dream-2", "dream-4", "dream-5", "dream-6"}
	if strings.Join(seen, ",") != strings.Join(want, ",") {
		t.Fatalf("pages returned %v; every proposal still waiting must appear once, want %v", seen, want)
	}
}
