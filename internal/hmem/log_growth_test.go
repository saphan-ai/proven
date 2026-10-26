package hmem_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// Growth sizes: a log of 20,000 ledger lines (about 2 MB) is far past what a
// review run writes, and 200 lookups over it make an unbounded reader obvious.
const (
	growthLines   = 20000
	growthSpread  = 2000 // ids, so one memory has ten lines
	growthLookups = 200
)

// fileSize is the size of a log file in bytes.
func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
}

// TestLedgerLookupsReadEachByteOnce proves the ledger readers (History,
// ledgerHas, the approved-line check) are bounded by the bytes added since the
// last lookup, not by the log (E16): after 200 rounds of lookups over a
// 20,000-line log, the bytes read from disk are within twice the file size, and
// one new line costs about one line. A reader that rereads the log per call
// reads 200 times the file and turns this red.
func TestLedgerLookupsReadEachByteOnce(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	s := openApprove(t, dir, logs)
	if err := hmem.AppendLedgerLines(s, growthLines, growthSpread); err != nil {
		t.Fatal(err)
	}
	size := fileSize(t, filepath.Join(logs, "hmem.jsonl"))
	for i := 0; i < growthLookups; i++ {
		h, err := s.History("mem-3")
		if err != nil || len(h) != growthLines/growthSpread {
			t.Fatalf("History = %d entries, err %v, want %d", len(h), err, growthLines/growthSpread)
		}
		if _, err := hmem.LedgerHas(s, "create", "mem-3"); err != nil {
			t.Fatal(err)
		}
		if _, err := hmem.LedgerHas(s, "dream_approved", "dream-1"); err != nil {
			t.Fatal(err)
		}
	}
	if got := hmem.BytesRead(s); got > 2*size {
		t.Fatalf("%d lookups read %d bytes of a %d-byte log: the readers are not bounded", growthLookups, got, size)
	}

	// One more line costs about one line, not the log.
	before := hmem.BytesRead(s)
	if err := hmem.AppendLedgerLines(s, 1, growthSpread); err != nil {
		t.Fatal(err)
	}
	if _, err := hmem.LedgerHas(s, "create", "mem-3"); err != nil {
		t.Fatal(err)
	}
	if delta := hmem.BytesRead(s) - before; delta > 1024 {
		t.Fatalf("one new line made the reader read %d bytes", delta)
	}
}

// TestHistoryIsPagedAndCapped proves one History call returns at most one page
// of a long trail, and HistoryPage reads the rest by cursor with nothing lost
// or repeated.
func TestHistoryIsPagedAndCapped(t *testing.T) {
	s := openApprove(t, t.TempDir(), t.TempDir())
	const total = 2500
	if err := hmem.AppendLedgerLines(s, total, 1); err != nil {
		t.Fatal(err)
	}
	h, err := s.History("mem-0")
	if err != nil || len(h) != 1000 {
		t.Fatalf("History = %d entries, err %v, want the 1000 of one page", len(h), err)
	}
	seen, cursor := 0, 0
	for cursor >= 0 {
		page, next, err := s.HistoryPage("mem-0", cursor, 700)
		if err != nil {
			t.Fatal(err)
		}
		if len(page) > 700 {
			t.Fatalf("a page of %d entries, limit 700", len(page))
		}
		seen += len(page)
		cursor = next
	}
	if seen != total {
		t.Fatalf("pages held %d entries, want %d", seen, total)
	}
	if _, _, err := s.HistoryPage("mem-0", 0, 0); err == nil {
		t.Fatal("a page limit of 0 was accepted")
	}
}

// TestDreamLogLookupsReadEachByteOnce is the growth test of the dream log: the
// proposal lookups every sleep, approval and Clear make read only the lines
// added since the last one.
func TestDreamLogLookupsReadEachByteOnce(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	s := openApprove(t, dir, logs)
	if err := hmem.AppendProposalLines(s, 5000); err != nil {
		t.Fatal(err)
	}
	size := fileSize(t, filepath.Join(logs, "dreams.jsonl"))
	for i := 0; i < growthLookups; i++ {
		if pend, err := s.Pending(); err != nil || len(pend) != 5000 {
			t.Fatalf("Pending = %d, err %v, want 5000", len(pend), err)
		}
		s.SleepPreview()
	}
	if got := hmem.BytesRead(s); got > 2*size {
		t.Fatalf("%d lookups read %d bytes of a %d-byte dream log: the reader is not bounded", growthLookups, got, size)
	}
}

// TestLedgerIndexFollowsAShrunkLog proves the index does not outlive its file:
// when the audit log is emptied behind the store's back, the next lookup reads
// the empty log and answers from it, not from lines it remembers.
func TestLedgerIndexFollowsAShrunkLog(t *testing.T) {
	logs := t.TempDir()
	s := openApprove(t, t.TempDir(), logs)
	if err := hmem.AppendLedgerLines(s, 50, 5); err != nil {
		t.Fatal(err)
	}
	if h, _ := s.History("mem-1"); len(h) != 10 {
		t.Fatalf("setup: %d entries, want 10", len(h))
	}
	if err := os.Truncate(filepath.Join(logs, "hmem.jsonl"), 0); err != nil {
		t.Fatal(err)
	}
	if h, err := s.History("mem-1"); err != nil || len(h) != 0 {
		t.Fatalf("after the log was emptied History = %d entries, err %v, want none", len(h), err)
	}
}
