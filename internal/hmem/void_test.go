package hmem_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// emptyStateFile removes every record from state.json behind the store's back.
// It stands for records lost some way other than Clear (an older log), the
// only way left to meet an absent source of an approving proposal.
func emptyStateFile(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, "state.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	raw["records"] = []any{}
	out, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestClearVoidsAPendingProposalSoNoIDReuseBindsIt proves the reviewer's first
// reproduction: two records, Dream, Clear, two unrelated records that reuse the
// ids, then Approve of the old proposal. Approval must be refused by name and
// no record may be created, because a proposal never outlives the records it
// names (ADR-0018 point 3: lineage names the real sources).
func TestClearVoidsAPendingProposalSoNoIDReuseBindsIt(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	p, ids := dreamedPair(t, s)
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	if got := hmem.ProposalStatus(s, p.ID); got != hmem.ProposalVoid {
		t.Fatalf("proposal state %q after Clear, want void", got)
	}
	for _, c := range []string{"unrelated apples are red", "unrelated rivers are long"} {
		if _, err := s.Add(hmem.AddInput{Content: c, Source: "t"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, ok := s.Get(ids[0]); !ok {
		t.Fatalf("setup: id %s was not reused, the test no longer proves the case", ids[0])
	}
	rec, err := s.Approve(p.ID)
	if err == nil {
		t.Fatalf("approved %+v: the old proposal bound itself to unrelated records", rec)
	}
	if !strings.Contains(err.Error(), p.ID) || !strings.Contains(err.Error(), "void") {
		t.Fatalf("refusal %q does not name the voided proposal %s", err, p.ID)
	}
	if len(s.All()) != 2 {
		t.Fatalf("%d records, want the 2 unrelated ones only", len(s.All()))
	}
	if pend, _ := s.Pending(); len(pend) != 0 {
		t.Fatalf("%d proposals still pending after Clear", len(pend))
	}
}

// TestClearVoidsAPendingDomainProposal proves the reviewer's second
// reproduction: a pending domain proposal, Clear, Approve. Approval is refused
// by name and the proposal never reaches an approved state.
func TestClearVoidsAPendingDomainProposal(t *testing.T) {
	e := newDomainEnv(t)
	s := openApprove(t, e.dir, e.logs)
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Approve(e.proposal); err == nil || !strings.Contains(err.Error(), e.proposal) {
		t.Fatalf("approve after Clear: err = %v, want a refusal naming %s", err, e.proposal)
	}
	if got := hmem.ProposalStatus(s, e.proposal); got != hmem.ProposalVoid {
		t.Fatalf("proposal state %q, want void", got)
	}
	if done, _ := hmem.LedgerHas(s, "dream_approved", e.proposal); done {
		t.Fatal("a dream_approved line was written for a voided proposal")
	}
}

// TestDomainApprovalWithAnAbsentSourceIsRefusedBeforeIntent proves that a
// domain proposal whose named source is gone is refused before the approving
// state is written, so no approved state or dream_approved line follows. The
// source is lost behind the store's back, so Clear's voiding is not what
// refuses it.
func TestDomainApprovalWithAnAbsentSourceIsRefusedBeforeIntent(t *testing.T) {
	e := newDomainEnv(t)
	emptyStateFile(t, e.dir)
	s := openApprove(t, e.dir, e.logs)
	_, err := s.Approve(e.proposal)
	if err == nil || !strings.Contains(err.Error(), e.record) {
		t.Fatalf("approve err = %v, want a refusal naming the absent source %s", err, e.record)
	}
	if got := hmem.ProposalStatus(s, e.proposal); got != hmem.ProposalProposed {
		t.Fatalf("proposal state %q, want proposed: the intent was written for a refused approval", got)
	}
	if done, _ := hmem.LedgerHas(s, "dream_approved", e.proposal); done {
		t.Fatal("a dream_approved line was written for a refused approval")
	}
}

// TestReopenRefusesAnApprovingDomainProposalWithAnAbsentSource proves recovery
// of an approving domain proposal checks its sources too: Open does not fail,
// the proposal is not approved, and an explicit Approve refuses by name.
func TestReopenRefusesAnApprovingDomainProposalWithAnAbsentSource(t *testing.T) {
	e := newDomainEnv(t)
	s := openApprove(t, e.dir, e.logs)
	hook, boom := failing(func(op string) bool { return op == "state" })
	s.FailWrites(hook)
	if _, err := s.Approve(e.proposal); !errors.Is(err, boom) {
		t.Fatalf("approve err = %v, want the injected failure", err)
	}
	if got := hmem.ProposalStatus(s, e.proposal); got != hmem.ProposalApproving {
		t.Fatalf("setup: state %q, want approving", got)
	}
	emptyStateFile(t, e.dir)
	s = openApprove(t, e.dir, e.logs)
	if got := hmem.ProposalStatus(s, e.proposal); got != hmem.ProposalApproving {
		t.Fatalf("state %q after reopen, want approving: recovery approved a proposal with no source", got)
	}
	if done, _ := hmem.LedgerHas(s, "dream_approved", e.proposal); done {
		t.Fatal("recovery wrote a dream_approved line with no source")
	}
	if _, err := s.Approve(e.proposal); err == nil || !strings.Contains(err.Error(), e.record) {
		t.Fatalf("approve err = %v, want a refusal naming %s", err, e.record)
	}
	if got := hmem.ProposalStatus(s, e.proposal); got != hmem.ProposalApproving {
		t.Fatalf("state %q after the refused approve, want approving", got)
	}
}
