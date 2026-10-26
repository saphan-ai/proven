package hmem_test

import (
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

// TestSleepExecuteRefusesAProtectedIdInASuppliedPlan proves the archive guard
// holds at execution, not only in the preview: a caller-supplied plan that
// names a derived record, or a record that became superseded after the
// preview, archives neither (H-MEM section 8.2).
func TestSleepExecuteRefusesAProtectedIdInASuppliedPlan(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	derived, err := s.Add(hmem.AddInput{Content: "synthesized insight about the quarterly roadmap", Source: "dreaming", DerivedFrom: []string{"mem-x"}})
	if err != nil {
		t.Fatal(err)
	}
	late, _ := s.Add(hmem.AddInput{Content: "random one-off note about parking near the office", Source: "chat"})
	now += 40 * day
	preview := s.SleepPreview()
	if !contains(preview.ToArchive, late.ID) {
		t.Fatalf("setup: preview %v should plan the plain record", preview.ToArchive)
	}
	// State change after the preview: another record now supersedes late.
	newer, _ := s.Add(hmem.AddInput{Content: "replacement note about transit passes", Source: "chat"})
	if err := s.MarkSuperseded(late.ID, newer.ID); err != nil {
		t.Fatal(err)
	}
	rep, err := s.SleepExecute(&hmem.SleepPreview{ToArchive: []string{derived.ID, late.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Archived != 0 {
		t.Fatalf("report %+v: a protected record was archived", rep)
	}
	for _, id := range []string{derived.ID, late.ID} {
		if _, ok := s.Get(id); !ok {
			t.Fatalf("%s was archived by a supplied plan", id)
		}
	}
}

// TestSleepSynthesisKeepsItsSources proves an ordinary sleep cycle never
// deletes the memories it synthesizes from: they stay readable, the derived
// record names them, and a second cycle does not synthesize them again.
func TestSleepSynthesisKeepsItsSources(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	a, _ := s.Add(hmem.AddInput{Content: "the release train leaves every friday afternoon", Source: "chat"})
	b, _ := s.Add(hmem.AddInput{Content: "the release train leaves every friday afternoon at five", Source: "chat"})
	rep, err := s.SleepExecute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Synthesized != 1 {
		t.Fatalf("setup: report %+v, want one synthesis", rep)
	}
	var derived hmem.Record
	for _, r := range s.All() {
		if len(r.DerivedFrom) > 0 {
			derived = r
		}
	}
	if len(derived.DerivedFrom) != 2 {
		t.Fatalf("derived record %+v names no sources", derived)
	}
	for _, id := range []string{a.ID, b.ID} {
		if _, ok := s.Get(id); !ok {
			t.Fatalf("source %s was deleted by synthesis", id)
		}
		if !contains(derived.DerivedFrom, id) {
			t.Fatalf("lineage %v misses %s", derived.DerivedFrom, id)
		}
	}
	rep, err = s.SleepExecute(nil)
	if err != nil || rep.Synthesized != 0 {
		t.Fatalf("second cycle %+v %v: sources were synthesized again", rep, err)
	}
	// The derived record references its sources, so a late cycle keeps them.
	now += 40 * day
	if _, err := s.SleepExecute(nil); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{a.ID, b.ID, derived.ID} {
		if _, ok := s.Get(id); !ok {
			t.Fatalf("%s archived after 40 days", id)
		}
	}
}

// TestSleepKeepsBothSidesOfAResolvedConflict proves a resolved conflict is
// protected through sleep: the winner carries Supersedes, the loser
// SupersededBy, and a cycle 40 days later archives neither.
func TestSleepKeepsBothSidesOfAResolvedConflict(t *testing.T) {
	now := fixed
	s := clocked(t, &now)
	hi, lo := 0.95, 0.2
	a, _ := s.Add(hmem.AddInput{Content: "always enable telemetry uploads for diagnostics", Confidence: &hi, Source: "t"})
	b, _ := s.Add(hmem.AddInput{Content: "never enable telemetry uploads for diagnostics", Confidence: &lo, Source: "t"})
	found, _ := s.ScanConflicts()
	if len(found) != 1 {
		t.Fatalf("setup: %v", found)
	}
	if outcome, _, err := s.AutoResolve(found[0]); err != nil || outcome != hmem.ConflictResolved {
		t.Fatalf("setup: %q %v", outcome, err)
	}
	winner, _ := s.Get(a.ID)
	if winner.Supersedes != b.ID {
		t.Fatalf("winner %+v does not name the loser", winner)
	}
	now += 40 * day
	plan := s.SleepPreview()
	if contains(plan.ToArchive, a.ID) || contains(plan.ToArchive, b.ID) {
		t.Fatalf("plan %v archives a side of a resolved conflict", plan.ToArchive)
	}
	if _, err := s.SleepExecute(nil); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{a.ID, b.ID} {
		if _, ok := s.Get(id); !ok {
			t.Fatalf("%s archived after resolution", id)
		}
	}
}
