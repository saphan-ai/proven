package hmem_test

import (
	"runtime"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// historyProposals is the number of finished (rejected) proposals the growth
// tests put behind one waiting proposal: far past what a review run makes.
const historyProposals = 20000

// growthEnv is a store with one waiting proposal and a long history of rejected
// ones, reopened so the dream index is built from the log.
func growthEnv(t *testing.T, history int) (*hmem.Store, string) {
	t.Helper()
	dir, logs := t.TempDir(), t.TempDir()
	s := openApprove(t, dir, logs)
	s.Add(hmem.AddInput{Content: "prefer short bullet answers", Type: hmem.TypePreference, Source: "t"})
	s.Add(hmem.AddInput{Content: "prefer python for scripting", Type: hmem.TypePreference, Source: "t"})
	ps, err := s.Dream()
	if err != nil || len(ps) == 0 {
		t.Fatalf("Dream = %d proposals, err %v", len(ps), err)
	}
	if err := hmem.AppendRejectedProposalLines(s, history); err != nil {
		t.Fatal(err)
	}
	s = openApprove(t, dir, logs)
	return s, ps[0].ID
}

// allocated is the bytes the Go runtime allocated while fn ran.
func allocated(fn func()) uint64 {
	var a, b runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&a)
	fn()
	runtime.ReadMemStats(&b)
	return b.TotalAlloc - a.TotalAlloc
}

// TestProposalLookupsDoNotFollowHistory proves R16 for proposals: with 20,000
// rejected proposals in the dream log, Pending, a single-proposal Approve and a
// sleep examine only the proposals that wait, return only those, and allocate
// far less than one list of the whole history would. The work done (proposals
// visited) is the same as with no history at all. A lookup that builds the
// whole history (the old shape) visits 20,000 and allocates megabytes.
func TestProposalLookupsDoNotFollowHistory(t *testing.T) {
	if raceBuild {
		t.Skip("race instrumentation inflates allocation counts; the normal build measures the allocation budget")
	}
	small, _ := growthEnv(t, 0)
	big, id := growthEnv(t, historyProposals)

	visitedBy := func(s *hmem.Store, fn func()) int64 {
		before := hmem.ProposalsVisited(s)
		fn()
		return hmem.ProposalsVisited(s) - before
	}
	pendingWork := func(s *hmem.Store) (int, int64) {
		var n int
		v := visitedBy(s, func() {
			p, err := s.Pending()
			if err != nil {
				t.Fatal(err)
			}
			n = len(p)
		})
		return n, v
	}
	nSmall, vSmall := pendingWork(small)
	nBig, vBig := pendingWork(big)
	if nBig != nSmall || vBig != vSmall {
		t.Fatalf("Pending returned %d and visited %d proposals with history, %d and %d without", nBig, vBig, nSmall, vSmall)
	}

	if v := visitedBy(big, func() { big.SleepPreview() }); v > vSmall {
		t.Fatalf("a sleep preview visited %d proposals behind %d waiting", v, vSmall)
	}
	if v := visitedBy(big, func() {
		if _, err := big.SleepExecute(nil); err != nil {
			t.Fatal(err)
		}
	}); v > 2*vSmall {
		t.Fatalf("a sleep visited %d proposals behind %d waiting", v, vSmall)
	}
	if v := visitedBy(big, func() {
		if _, err := big.Approve(id); err != nil {
			t.Fatal(err)
		}
	}); v > 4 {
		t.Fatalf("Approve visited %d proposals to find one", v)
	}

	// Allocation: the history must not be materialized by a lookup.
	whole := allocated(func() { hmem.ScanAllProposals(big) })
	lookup := allocated(func() {
		if _, err := big.Pending(); err != nil {
			t.Fatal(err)
		}
		big.SleepPreview()
	})
	if lookup > whole/20 {
		t.Fatalf("Pending and a sleep preview allocated %d bytes; one whole-history list takes %d", lookup, whole)
	}

	// Negative control: the whole-history scan is visible to the counter. If it
	// were not, the assertions above would prove nothing.
	if v := visitedBy(big, func() { hmem.ScanAllProposals(big) }); v < historyProposals {
		t.Fatalf("the control scan visited %d proposals, want at least %d: the counter does not see a scan", v, historyProposals)
	}
}

// TestPendingIsPaged proves an external list of proposals is capped: more than
// one page waiting returns one page and a cursor, and the cursor reaches every
// proposal once.
func TestPendingIsPaged(t *testing.T) {
	s := openApprove(t, t.TempDir(), t.TempDir())
	if err := hmem.AppendProposalLines(s, 2500); err != nil {
		t.Fatal(err)
	}
	if p, err := s.Pending(); err != nil || len(p) != 1000 {
		t.Fatalf("Pending = %d, err %v, want one page of 1000", len(p), err)
	}
	seen := map[string]bool{}
	for cursor := 0; cursor >= 0; {
		page, next, err := s.PendingPage(cursor, 700)
		if err != nil {
			t.Fatal(err)
		}
		if len(page) > 700 {
			t.Fatalf("a page of %d, limit 700", len(page))
		}
		for _, p := range page {
			if seen[p.ID] {
				t.Fatalf("proposal %s returned twice", p.ID)
			}
			seen[p.ID] = true
		}
		cursor = next
	}
	if len(seen) != 2500 {
		t.Fatalf("pages held %d proposals, want 2500", len(seen))
	}
	if _, _, err := s.PendingPage(0, 0); err == nil {
		t.Fatal("a page limit of 0 was accepted")
	}
}
