package hmem_test

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// dreamedPair adds two related plain records and returns the first insight or
// invariant proposal dreaming makes from them, with the two record ids.
func dreamedPair(t *testing.T, s *hmem.Store) (hmem.Proposal, []string) {
	t.Helper()
	a, err := s.Add(hmem.AddInput{Content: "prefer short bullet answers", Type: hmem.TypePreference, Source: "t"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Add(hmem.AddInput{Content: "prefer python for scripting", Type: hmem.TypePreference, Source: "t"})
	if err != nil {
		t.Fatal(err)
	}
	ps, err := s.Dream()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range ps {
		if p.Type != "domain" {
			return p, []string{a.ID, b.ID}
		}
	}
	t.Fatalf("setup: dreaming made no insight or invariant from %v", ps)
	return hmem.Proposal{}, nil
}

// TestSleepKeepsTheSourcesOfAPendingProposalUntilApproval proves F1 of the
// review of 9154bbe: two related plain records are dreamed into a proposal,
// 40 days pass, sleep runs, and the proposal is approved. Sleep must not
// archive the sources while the proposal is pending, and the approved record's
// derivedFrom must name records that are live (ADR-0018 point 3: no lineage
// that points at nothing).
func TestSleepKeepsTheSourcesOfAPendingProposalUntilApproval(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	p, ids := dreamedPair(t, s)
	now += 40 * day
	for _, id := range ids {
		if contains(s.SleepPreview().ToArchive, id) {
			t.Fatalf("preview plans %s, a source of pending proposal %s", id, p.ID)
		}
	}
	// A supplied plan gets no way around the guard either.
	if _, err := s.SleepExecute(&hmem.SleepPreview{ToArchive: ids}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SleepExecute(nil); err != nil {
		t.Fatal(err)
	}
	rec, err := s.Approve(p.ID)
	if err != nil || rec == nil {
		t.Fatalf("approve: %v %v", rec, err)
	}
	for _, id := range ids {
		if _, ok := s.Get(id); !ok {
			t.Fatalf("source %s is not live after approval", id)
		}
	}
	for _, d := range rec.DerivedFrom {
		if _, ok := s.Get(d); !ok {
			t.Fatalf("approved record %s names %s, which is absent", rec.ID, d)
		}
	}
	if len(rec.DerivedFrom) != len(ids) {
		t.Fatalf("lineage %v, want the sources %v", rec.DerivedFrom, ids)
	}
}

// TestSleepArchivesTheSourcesOfARejectedProposal proves the protection ends
// with the proposal: after Reject nothing names the records, so sleep may
// archive them as it would any plain old record.
func TestSleepArchivesTheSourcesOfARejectedProposal(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	p, ids := dreamedPair(t, s)
	if err := s.Reject(p.ID); err != nil {
		t.Fatal(err)
	}
	now += 40 * day
	plan := s.SleepPreview()
	for _, id := range ids {
		if !contains(plan.ToArchive, id) {
			t.Fatalf("plan %v keeps %s after its proposal was rejected", plan.ToArchive, id)
		}
	}
}

// TestApproveRefusesAnAbsentSource proves approval never writes a lineage
// reference to a record that is gone (an older log, or records cleared since
// the dream): it refuses by name and creates nothing.
func TestApproveRefusesAnAbsentSource(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	p, ids := dreamedPair(t, s)
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	rec, err := s.Approve(p.ID)
	if err == nil {
		t.Fatalf("approved %+v with sources %v absent", rec, ids)
	}
	if !strings.Contains(err.Error(), ids[0]) {
		t.Fatalf("refusal %q does not name the absent source %s", err, ids[0])
	}
	if len(s.All()) != 0 {
		t.Fatalf("a refused approval left %d records", len(s.All()))
	}
}
