package hmem_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/hmem"
)

// failOnCalls refuses the durable writes whose position in the run (counted
// from 1, over every kind of write) is in the set, and passes the others. It
// lets a test fail one write point, or two in a row, whatever order the store
// writes in.
func failOnCalls(calls ...int) func(string) error {
	want := map[int]bool{}
	for _, c := range calls {
		want[c] = true
	}
	n := 0
	return func(op string) error {
		n++
		if want[n] {
			return fmt.Errorf("injected write failure at call %d (%s)", n, op)
		}
		return nil
	}
}

// openClocked opens the store at dir and logs with a clock the test moves
// through *now, so a test can reopen the same files.
func openClocked(t *testing.T, dir, logs string, now *int64) *hmem.Store {
	t.Helper()
	s, err := hmem.Open(hmem.Options{Dir: dir, LogDir: logs, Clock: func() time.Time { return time.UnixMilli(*now) }})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// failureSets lists every way to fail one or two of the first maxCall writes.
func failureSets(maxCall int) [][]int {
	var sets [][]int
	for a := 1; a <= maxCall; a++ {
		sets = append(sets, []int{a})
		for b := a + 1; b <= maxCall; b++ {
			sets = append(sets, []int{a, b})
		}
	}
	return sets
}

// recordIDs is the sorted ids of the records a store holds.
func recordIDs(s *hmem.Store) string {
	var ids []string
	for _, r := range s.All() {
		ids = append(ids, r.ID)
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

// ledgerCount counts the ledger lines of one action for one memory id.
func ledgerCount(t *testing.T, s *hmem.Store, id, action string) int {
	t.Helper()
	h, err := s.History(id)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range h {
		if e.Action == action {
			n++
		}
	}
	return n
}

// TestClearAgreesWithItsAuditAtEveryWritePoint proves the property that the
// audit and the state never disagree about a Clear (ADR-0018 point 3: no
// half-applied change). Every one- and two-write failure of a Clear is
// injected, then the store is reopened. The state is either untouched (records
// live, no clear line, the proposal still pending) or cleared (no records, the
// proposal void, and a clear line unless the error says the line is missing).
// In every case memory equals what a new process loads.
func TestClearAgreesWithItsAuditAtEveryWritePoint(t *testing.T) {
	for _, fail := range failureSets(5) {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			dir, logs := t.TempDir(), t.TempDir()
			s := openApprove(t, dir, logs)
			p, _ := dreamedPair(t, s)
			before := recordIDs(s)
			s.FailWrites(failOnCalls(fail...))
			err := s.Clear()
			s.FailWrites(nil)

			re := openApprove(t, dir, logs)
			if got := recordIDs(s); got != recordIDs(re) {
				t.Fatalf("memory holds %q but a new process loads %q (err %v)", got, recordIDs(re), err)
			}
			clears := ledgerCount(t, re, "all", "clear")
			switch recordIDs(re) {
			case before:
				if clears != 0 {
					t.Fatalf("the records are live but the ledger reports %d clear lines (err %v)", clears, err)
				}
				if got := hmem.ProposalStatus(re, p.ID); got != hmem.ProposalProposed {
					t.Fatalf("the records are live but the proposal is %q (err %v)", got, err)
				}
				if err == nil {
					t.Fatal("Clear reported success but the records are live")
				}
			case "":
				if got := hmem.ProposalStatus(re, p.ID); got != hmem.ProposalVoid {
					t.Fatalf("the records are gone but the proposal is %q (err %v)", got, err)
				}
				if clears == 0 && (err == nil || !strings.Contains(err.Error(), "audit")) {
					t.Fatalf("no clear line and the error does not name the missing audit line: %v", err)
				}
				if clears > 1 {
					t.Fatalf("%d clear lines for one Clear", clears)
				}
			default:
				t.Fatalf("a new process loads %q, neither the records before (%q) nor none", recordIDs(re), before)
			}
		})
	}
}

// TestFailedClearAuditLeavesNoFalseClearEvent proves the reviewer's overlay
// reproduction: the state write of a Clear fails, and a new process still
// holds the records and History("all") reports no clear.
func TestFailedClearAuditLeavesNoFalseClearEvent(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	s := openApprove(t, dir, logs)
	dreamedPair(t, s)
	hook, _ := failing(func(op string) bool { return op == "state" })
	s.FailWrites(hook)
	if err := s.Clear(); err == nil {
		t.Fatal("Clear succeeded with a refused state write")
	}
	s.FailWrites(nil)
	re := openApprove(t, dir, logs)
	if len(re.All()) != 2 {
		t.Fatalf("a new process loads %d records, want 2", len(re.All()))
	}
	if n := ledgerCount(t, re, "all", "clear"); n != 0 {
		t.Fatalf("failed Clear left %d durable clear events", n)
	}
}

// sleepFixture is a store with one plain record old enough to archive, a
// short-term record old enough to consolidate, and two near-identical records
// that synthesis joins, so one cycle makes every kind of change.
func sleepFixture(t *testing.T, dir, logs string, now *int64) *hmem.Store {
	t.Helper()
	s := openClocked(t, dir, logs, now)
	for _, in := range []hmem.AddInput{
		{Content: "invoice 4471 was settled with the vendor in March", Source: "chat", At: fixed - 40*day},
		{Content: "the standing desk arrives next week", Source: "chat", Layer: hmem.LayerShort, At: fixed - 8*day},
		{Content: "the release train leaves every friday afternoon", Source: "chat", At: fixed},
		{Content: "the release train leaves every friday afternoon at five", Source: "chat", At: fixed},
	} {
		if _, err := s.Add(in); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

// TestSleepAgreesWithItsLedgerAtEveryWritePoint proves a failed sleep cycle
// leaves the durable state, the ledger and the returned error telling the same
// story. Every one- and two-write failure is injected, including a failed audit
// followed by a failed rollback write (the reviewer's overlay reproduction).
// After each, memory equals what a new process loads; a sleep ledger line
// exists only for a change the state file holds; and when the state file holds
// a change the call reported as failed, the error says so by naming the state
// file.
func TestSleepAgreesWithItsLedgerAtEveryWritePoint(t *testing.T) {
	for _, fail := range failureSets(7) {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			dir, logs := t.TempDir(), t.TempDir()
			now := fixed
			s := sleepFixture(t, dir, logs, &now)
			base := len(s.All())
			s.FailWrites(failOnCalls(fail...))
			rep, err := s.SleepExecute(nil)
			s.FailWrites(nil)

			re := openClocked(t, dir, logs, &now)
			if got := recordIDs(s); got != recordIDs(re) {
				t.Fatalf("memory holds %q but a new process loads %q (err %v)", got, recordIDs(re), err)
			}
			// What the state file holds, read back as the ledger lines it supports.
			derived := 0
			for _, r := range re.All() {
				if len(r.DerivedFrom) > 0 {
					derived++
					if ledgerCount(t, re, r.ID, "sleep_synthesize") == 0 && err == nil {
						t.Fatalf("synthesized record %s has no sleep_synthesize line", r.ID)
					}
				}
			}
			for _, id := range []string{"mem-1", "mem-2", "mem-3", "mem-4", "mem-5"} {
				_, live := re.Get(id)
				if n := ledgerCount(t, re, id, "sleep_archive"); n > 0 && live {
					t.Fatalf("the ledger reports %s archived but a new process still loads it (err %v)", id, err)
				}
				if n := ledgerCount(t, re, id, "sleep_synthesize"); n > 0 && !live {
					t.Fatalf("the ledger reports %s synthesized but a new process does not load it (err %v)", id, err)
				}
			}
			if err != nil {
				if rep.Archived+rep.Consolidated+rep.Synthesized != 0 {
					t.Fatalf("failed sleep reports work done: %+v", rep)
				}
				if len(re.All()) != base && !strings.Contains(err.Error(), "state file") {
					t.Fatalf("the state file holds a change the call reported failed, and the error does not name the state file: %v", err)
				}
			} else if derived != 1 {
				t.Fatalf("sleep succeeded with %d derived records, want 1", derived)
			}
		})
	}
}
