package hmem_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// TestFailedClearLeavesRecordsAndProposalsAsTheyWere proves the reviewer's
// first reproduction: a state-file write that fails inside Clear, after the
// proposals were voided, must leave every record and every pending proposal as
// it was, in this process and after a reopen (ADR-0018 point 3: no half-applied
// change). Clear is one transition across proposal and record state.
func TestFailedClearLeavesRecordsAndProposalsAsTheyWere(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	s := openApprove(t, dir, logs)
	p, ids := dreamedPair(t, s)

	hook, boom := failing(func(op string) bool { return op == "state" })
	s.FailWrites(hook)
	if err := s.Clear(); !errors.Is(err, boom) {
		t.Fatalf("Clear error = %v, want the injected failure", err)
	}
	s.FailWrites(nil)

	check := func(label string, s *hmem.Store) {
		t.Helper()
		if n := len(s.All()); n != len(ids) {
			t.Fatalf("%s: %d records, want %d", label, n, len(ids))
		}
		if got := hmem.ProposalStatus(s, p.ID); got != hmem.ProposalProposed {
			t.Fatalf("%s: proposal %s is %q, want proposed", label, p.ID, got)
		}
		if pend, _ := s.Pending(); len(pend) != 1 {
			t.Fatalf("%s: %d pending proposals, want 1", label, len(pend))
		}
	}
	check("same process", s)
	check("after reopen", openApprove(t, dir, logs))

	// The proposal is still approvable: nothing was lost by the failed Clear.
	if _, err := openApprove(t, dir, logs).Approve(p.ID); err != nil {
		t.Fatalf("approve after a failed Clear: %v", err)
	}
}

// TestFailedClearReportsAProposalThatCouldNotBeRestored proves the guard of the
// restore itself: when the state write and the restore of the proposals both
// fail, Clear says so by name instead of reporting only the first error, and
// the proposal stays void (the safe side: it cannot bind reused ids).
func TestFailedClearReportsAProposalThatCouldNotBeRestored(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	s := openApprove(t, dir, logs)
	p, _ := dreamedPair(t, s)
	voids := 0
	hook, _ := failing(func(op string) bool {
		if op == "dream" {
			voids++
			return voids > 1 // the void write passes, the restore write fails
		}
		return op == "state"
	})
	s.FailWrites(hook)
	err := s.Clear()
	if err == nil || !strings.Contains(err.Error(), p.ID) {
		t.Fatalf("Clear error = %v, want one naming the unrestored proposal %s", err, p.ID)
	}
	if got := hmem.ProposalStatus(s, p.ID); got != hmem.ProposalVoid {
		t.Fatalf("proposal %q, want void", got)
	}
}

// TestClearVoidsAnApprovedDomainProposalWhoseFinalAuditIsMissing proves the
// reviewer's second reproduction: a domain approval whose final dream_approved
// line failed leaves an approved proposal that is not finished. Clear removes
// its source, so Clear ends it as void like any live proposal. After a reopen
// recovery writes no dream_approved line and Approve refuses by name.
func TestClearVoidsAnApprovedDomainProposalWhoseFinalAuditIsMissing(t *testing.T) {
	e := newDomainEnv(t)
	s := openApprove(t, e.dir, e.logs)
	hook, boom := failing(func(op string) bool { return op == "audit:dream_approved" })
	s.FailWrites(hook)
	if _, err := s.Approve(e.proposal); !errors.Is(err, boom) {
		t.Fatalf("approve err = %v, want the injected failure", err)
	}
	if got := hmem.ProposalStatus(s, e.proposal); got != hmem.ProposalApproved {
		t.Fatalf("setup: state %q, want approved", got)
	}
	s.FailWrites(nil)
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}

	s = openApprove(t, e.dir, e.logs)
	if done, _ := hmem.LedgerHas(s, "dream_approved", e.proposal); done {
		t.Fatal("reopen wrote a dream_approved line for a cleared approval")
	}
	if got := hmem.ProposalStatus(s, e.proposal); got != hmem.ProposalVoid {
		t.Fatalf("state %q after Clear and reopen, want void", got)
	}
	if _, err := s.Approve(e.proposal); err == nil || !strings.Contains(err.Error(), e.proposal) {
		t.Fatalf("approve err = %v, want a refusal naming %s", err, e.proposal)
	}
	if done, _ := hmem.LedgerHas(s, "dream_approved", e.proposal); done {
		t.Fatal("a dream_approved line was written after the refusal")
	}
}

// TestReopenRefusesAnApprovedDomainProposalWithAnAbsentSource proves recovery
// of an approved proposal with no final audit line checks its sources exactly as
// the approving state does, for a source lost some way other than Clear: Open
// does not fail, no dream_approved line is written, and an explicit Approve
// refuses by naming the absent source.
func TestReopenRefusesAnApprovedDomainProposalWithAnAbsentSource(t *testing.T) {
	e := newDomainEnv(t)
	s := openApprove(t, e.dir, e.logs)
	hook, boom := failing(func(op string) bool { return op == "audit:dream_approved" })
	s.FailWrites(hook)
	if _, err := s.Approve(e.proposal); !errors.Is(err, boom) {
		t.Fatalf("approve err = %v, want the injected failure", err)
	}
	emptyStateFile(t, e.dir)

	s = openApprove(t, e.dir, e.logs)
	if done, _ := hmem.LedgerHas(s, "dream_approved", e.proposal); done {
		t.Fatal("recovery wrote a dream_approved line with no source")
	}
	if _, err := s.Approve(e.proposal); err == nil || !strings.Contains(err.Error(), e.record) {
		t.Fatalf("approve err = %v, want a refusal naming the absent source %s", err, e.record)
	}
	if done, _ := hmem.LedgerHas(s, "dream_approved", e.proposal); done {
		t.Fatal("a dream_approved line was written after the refusal")
	}
}
