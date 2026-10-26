package hmem_test

import (
	"strconv"
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

// approvingProposals is the number of unfinished approvals the approving-state
// test puts before the one proposal a page should return.
const approvingProposals = 20000

// TestPendingPageSkipsApprovingWithoutVisiting proves R16 for a store that holds
// many approving proposals (a persisted state a failed write leaves behind)
// before one proposed one: a one-item page returns that proposal and visits at
// most two proposals, however many approving ones precede it. The negative
// control is the full read of the waiting set, which does visit them all; the
// counter must see it, or the bound proves nothing.
func TestPendingPageSkipsApprovingWithoutVisiting(t *testing.T) {
	s := openApprove(t, t.TempDir(), t.TempDir())
	if err := hmem.AppendProposalLinesWith(s, approvingProposals, hmem.ProposalApproving); err != nil {
		t.Fatal(err)
	}
	if err := hmem.AppendProposalLines(s, 1); err != nil {
		t.Fatal(err)
	}
	// Build the index first: reading the log is not the work of a page.
	if _, _, err := s.PendingPage(0, 1); err != nil {
		t.Fatal(err)
	}
	before := hmem.ProposalsVisited(s)
	page, next, err := s.PendingPage(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	v := hmem.ProposalsVisited(s) - before
	want := "dream-" + strconv.Itoa(approvingProposals+1)
	if len(page) != 1 || page[0].ID != want {
		t.Fatalf("the page returned %v, want only %s", page, want)
	}
	if v > 2 {
		t.Fatalf("a one-item page visited %d proposals with %d approving ahead of it; want at most 2", v, approvingProposals)
	}
	// An approving proposal is never offered: the page that filled at the end
	// of the list, if it gave a cursor, is followed by an empty one.
	if next >= 0 {
		if rest, _, err := s.PendingPage(next, 1); err != nil || len(rest) != 0 {
			t.Fatalf("the page after the last proposal returned %d proposals, err %v", len(rest), err)
		}
	}

	// Negative control: the full read visits the approving proposals too.
	before = hmem.ProposalsVisited(s)
	hmem.ScanAllWaiting(s)
	if v := hmem.ProposalsVisited(s) - before; v < approvingProposals {
		t.Fatalf("the control scan visited %d proposals, want at least %d: the counter does not see a scan", v, approvingProposals)
	}
}
