package hmem_test

import (
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/hmem"
)

// day is one day in milliseconds, the unit of the store's clock.
const day = int64(24 * 60 * 60 * 1000)

// clocked opens a store whose clock the test moves through *now.
func clocked(t *testing.T, now *int64) *hmem.Store {
	t.Helper()
	s, err := hmem.Open(hmem.Options{
		Dir: t.TempDir(), LogDir: t.TempDir(),
		Clock: func() time.Time { return time.UnixMilli(*now) },
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func contains(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// TestSleepArchivesAPlainOldRecordAndProtectsADerivedOne proves the sleep
// guardrail: at the same age a plain low-access record is planned for archive
// and a record with derivedFrom lineage is not, and preview changes nothing.
func TestSleepArchivesAPlainOldRecordAndProtectsADerivedOne(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	plain, _ := s.Add(hmem.AddInput{Content: "random one-off note about parking near the office", Source: "chat"})
	derived, err := s.Add(hmem.AddInput{Content: "synthesized insight about the quarterly roadmap", Source: "dreaming", DerivedFrom: []string{"mem-x"}})
	if err != nil {
		t.Fatal(err)
	}
	now += 40 * day
	plan := s.SleepPreview()
	if len(s.All()) != 2 {
		t.Fatal("preview changed the store")
	}
	if !contains(plan.ToArchive, plain.ID) || contains(plan.ToArchive, derived.ID) {
		t.Fatalf("plan %v: plain must archive, derived must not", plan.ToArchive)
	}
	rep, err := s.SleepExecute(&plan)
	if err != nil {
		t.Fatal(err)
	}
	if rep.BackupID == "" || rep.Archived != 1 {
		t.Fatalf("report %+v", rep)
	}
	if _, ok := s.Get(derived.ID); !ok {
		t.Fatal("the derived record was archived")
	}
	if _, ok := s.Get(plain.ID); ok {
		t.Fatal("the plain record was not archived")
	}
}

// TestSleepBackupRestoresTheArchivedRecord proves the way back: Execute's
// backup brings an archived record back, and an id that is not a backup is refused.
func TestSleepBackupRestoresTheArchivedRecord(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	plain, _ := s.Add(hmem.AddInput{Content: "random one-off note about parking", Source: "chat"})
	now += 40 * day
	rep, err := s.SleepExecute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Get(plain.ID); ok {
		t.Fatal("not archived")
	}
	if err := s.Restore(rep.BackupID); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Get(plain.ID); !ok {
		t.Fatal("restore did not bring the record back")
	}
	if err := s.Restore("../state"); err == nil {
		t.Fatal("a path was accepted as a backup id")
	}
}

// TestAutoResolveMarksTheLoserAndDeletesNothing proves a confident conflict is
// resolved by marking the loser supersededBy the winner, with both kept.
func TestAutoResolveMarksTheLoserAndDeletesNothing(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	hi, lo := 0.95, 0.2
	a, _ := s.Add(hmem.AddInput{Content: "always enable telemetry uploads for diagnostics", Confidence: &hi, Source: "t"})
	b, _ := s.Add(hmem.AddInput{Content: "never enable telemetry uploads for diagnostics", Confidence: &lo, Source: "t"})
	found, err := s.ScanConflicts()
	if err != nil || len(found) != 1 {
		t.Fatalf("scan: %v %v", found, err)
	}
	outcome, c, err := s.AutoResolve(found[0])
	if err != nil || outcome != hmem.ConflictResolved || !c.Resolved {
		t.Fatalf("outcome %q %+v %v", outcome, c, err)
	}
	loser, _ := s.Get(b.ID)
	winner, _ := s.Get(a.ID)
	if loser.SupersededBy != a.ID || winner.SupersededBy != "" {
		t.Fatalf("loser %+v winner %+v", loser, winner)
	}
}

// TestAutoResolveQueuesBelowTheGate proves a close, weak conflict is queued and
// changes no record: auto-resolve is gated on decision confidence.
func TestAutoResolveQueuesBelowTheGate(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	a, _ := s.Add(hmem.AddInput{Content: "always enable telemetry uploads", Source: "t"})
	b, _ := s.Add(hmem.AddInput{Content: "never enable telemetry uploads", Source: "t"})
	found, _ := s.ScanConflicts()
	if len(found) != 1 {
		t.Fatalf("scan: %v", found)
	}
	found[0].Confidence = 0.1 // a weak detection with equal records
	outcome, _, err := s.AutoResolve(found[0])
	if err != nil || outcome != hmem.ConflictQueued {
		t.Fatalf("outcome %q %v", outcome, err)
	}
	for _, id := range []string{a.ID, b.ID} {
		if r, _ := s.Get(id); r.SupersededBy != "" {
			t.Fatalf("%s was marked while queued", id)
		}
	}
}

// TestLedgerReconstructsCreateThenFeedback proves history lists the actions in
// order with the confidence before and after, and survives a reopen.
func TestLedgerReconstructsCreateThenFeedback(t *testing.T) {
	now := fixed
	dir, logs := t.TempDir(), t.TempDir()
	open := func() *hmem.Store {
		s, err := hmem.Open(hmem.Options{Dir: dir, LogDir: logs, Clock: func() time.Time { return time.UnixMilli(now) }})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	s := open()
	start := 0.7
	m, _ := s.Add(hmem.AddInput{Content: "audited record", Confidence: &start, Source: "t"})
	if err := s.Feedback(m.ID, "wrong"); err != nil {
		t.Fatal(err)
	}
	h, err := open().History(m.ID)
	if err != nil || len(h) != 2 || h[0].Action != "create" || h[1].Action != "feedback" {
		t.Fatalf("history %+v %v", h, err)
	}
	if h[1].PreviousState == nil || h[1].NewState == nil || h[1].PreviousState.Confidence != 0.7 || h[1].NewState.Confidence < 0.499 || h[1].NewState.Confidence > 0.501 {
		t.Fatalf("states %+v %+v", h[1].PreviousState, h[1].NewState)
	}
}

// TestDreamProposesWithoutMaterializingAndApproveCreatesLineage proves the
// approval boundary: dreaming adds no record, approving one creates a
// long-term record with derivedFrom, and approving it twice is refused.
func TestDreamProposesWithoutMaterializingAndApproveCreatesLineage(t *testing.T) {
	now := fixed
	dir, logs := t.TempDir(), t.TempDir()
	open := func() *hmem.Store {
		s, err := hmem.Open(hmem.Options{Dir: dir, LogDir: logs, Clock: func() time.Time { return time.UnixMilli(now) }})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	s := open()
	s.Add(hmem.AddInput{Content: "prefer short bullet answers", Type: hmem.TypePreference, Source: "t"})
	s.Add(hmem.AddInput{Content: "prefer python for scripting", Type: hmem.TypePreference, Source: "t"})
	ps, err := s.Dream()
	if err != nil || len(ps) == 0 {
		t.Fatalf("proposals %v %v", ps, err)
	}
	if len(s.All()) != 2 {
		t.Fatal("dreaming materialized a record")
	}
	// A new process finds the proposal: approval does not depend on memory of the dream.
	var pick string
	for _, p := range ps {
		if p.Type != "domain" {
			pick = p.ID
			break
		}
	}
	rec, err := open().Approve(pick)
	if err != nil || rec == nil {
		t.Fatalf("approve: %v %v", rec, err)
	}
	if rec.Layer != hmem.LayerLong || len(rec.DerivedFrom) == 0 {
		t.Fatalf("approved record %+v", rec)
	}
	if _, err := open().Approve(pick); err == nil {
		t.Fatal("a proposal was approved twice")
	}
}
