package world_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/trace"
	"github.com/saphan-ai/proven/internal/world"
)

// TestNormalizeNameFoldsDiacriticsAndSpace proves two spellings of one name
// give one id, different names and different types give different ids.
func TestNormalizeNameFoldsDiacriticsAndSpace(t *testing.T) {
	accented := "José\u00a0\u00a0García"
	if world.NormalizeName(accented) != world.NormalizeName("Jose Garcia") {
		t.Fatalf("%q", world.NormalizeName(accented))
	}
	if world.EntityIDFromName("person", "José García") != world.EntityIDFromName("person", "jose garcia") {
		t.Fatal("same name produced two ids")
	}
	if world.EntityIDFromName("org", "Acme Corp") == world.EntityIDFromName("org", "Acme Corporation") {
		t.Fatal("different names collapsed")
	}
	if world.EntityIDFromName("person", "x") == world.EntityIDFromName("org", "x") {
		t.Fatal("type is not part of the id")
	}
}

// TestDecideBands proves the merge bands at their edges: 0.85 merges, 0.82
// queues for review, 0.81 is new.
func TestDecideBands(t *testing.T) {
	if world.Decide(0.85) != "auto-merge" || world.Decide(0.84) != "review" ||
		world.Decide(0.82) != "review" || world.Decide(0.81) != "new" {
		t.Fatalf("%s %s %s %s", world.Decide(0.85), world.Decide(0.84), world.Decide(0.82), world.Decide(0.81))
	}
}

// TestUpsertWritesAnAuditLine proves each entity write adds one audit line.
func TestUpsertWritesAnAuditLine(t *testing.T) {
	dir, logs := t.TempDir(), t.TempDir()
	s := openAt(t, dir, logs)
	one := world.Entity{
		Type: "x", ID: "ent_a", Name: "Audit One",
		Source: world.Provenance{Kind: "manual", At: "2026-01-01T00:00:00.000Z"},
	}
	two := world.Entity{
		Type: "x", ID: "ent_b", Name: "Audit Two",
		Source: world.Provenance{Kind: "manual", At: "2026-01-02T00:00:00.000Z"},
	}
	if err := s.UpsertEntity(one); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertEntity(two); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(logs, "world.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(b), "\n") != 2 {
		t.Fatalf("%s", b)
	}
}

// TestProvenanceIsRequired proves an entity write with no provenance is refused.
func TestProvenanceIsRequired(t *testing.T) {
	s := openWorld(t)
	err := s.UpsertEntity(world.Entity{Type: "person", ID: "ent_x", Name: "No Provenance"})
	if err == nil {
		t.Fatal("a write with no provenance was stored")
	}
}

// TestIngestIsIdempotent proves a second ingest of the same trace skips every
// event and leaves the state unchanged (ADR-0018 point 5).
func TestIngestIsIdempotent(t *testing.T) {
	s := openWorld(t)
	log := sampleLog(t, true)
	first, err := s.Ingest(bytes.NewReader(log), 0)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Changed || first.Applied != 4 {
		t.Fatalf("%#v", first)
	}
	shape := s.Canonical()
	second, err := s.Ingest(bytes.NewReader(log), 0)
	if err != nil {
		t.Fatal(err)
	}
	if second.Changed || second.Applied != 0 || second.Skipped != 4 {
		t.Fatalf("%#v", second)
	}
	if s.Canonical() != shape {
		t.Fatal("state changed on the second ingest")
	}
}

// TestIngestResumesAfterAStop proves an ingest stopped halfway and resumed
// from a reopened store reaches the same state as one full pass (ADR-0018 point 5).
func TestIngestResumesAfterAStop(t *testing.T) {
	full := openWorld(t)
	log := sampleLog(t, true)
	if _, err := full.Ingest(bytes.NewReader(log), 0); err != nil {
		t.Fatal(err)
	}
	dir, logs := t.TempDir(), t.TempDir()
	part := openAt(t, dir, logs)
	stopped, err := part.Ingest(bytes.NewReader(log), 2)
	if err != nil {
		t.Fatal(err)
	}
	if !stopped.Stopped || stopped.Applied != 2 {
		t.Fatalf("%#v", stopped)
	}
	resumed := openAt(t, dir, logs)
	rest, err := resumed.Ingest(bytes.NewReader(log), 0)
	if err != nil {
		t.Fatal(err)
	}
	if rest.Applied != 2 || rest.Skipped != 2 {
		t.Fatalf("%#v", rest)
	}
	if resumed.Canonical() != full.Canonical() {
		t.Fatalf("resumed\n%s\nfull\n%s", resumed.Canonical(), full.Canonical())
	}
}

// TestRefutationWithoutEvidenceIsNotRemembered proves a refutation of a
// finding with no evidence is folded but not stored in H-MEM.
func TestRefutationWithoutEvidenceIsNotRemembered(t *testing.T) {
	s := openWorld(t)
	log := sampleLog(t, false)
	if _, err := s.Ingest(bytes.NewReader(log), 0); err != nil {
		t.Fatal(err)
	}
	ctx := s.Context("F1", "chg", "discount-applied-twice", "billing/discount.go", "Discount is applied twice")
	if ctx.MemoryID != "" || len(ctx.Memories) != 0 {
		t.Fatalf("%#v", ctx)
	}
	if len(ctx.Claims) != 1 || len(ctx.Evidence) != 1 || len(ctx.Refutations) != 1 || ctx.Verdict == "" {
		t.Fatalf("%#v", ctx)
	}
}

// TestContextNamesTheMemory proves a refutation with evidence is stored in
// H-MEM and the finding's context names that memory.
func TestContextNamesTheMemory(t *testing.T) {
	s := openWorld(t)
	if _, err := s.Ingest(bytes.NewReader(sampleLog(t, true)), 0); err != nil {
		t.Fatal(err)
	}
	ctx := s.Context("F1", "chg", "discount-applied-twice", "billing/discount.go", "Discount is applied twice")
	if ctx.MemoryID == "" || len(ctx.Memories) != 1 || ctx.Memories[0].ID != ctx.MemoryID {
		t.Fatalf("%#v", ctx)
	}
	if !strings.HasPrefix(ctx.MemoryID, "mem_") {
		t.Fatal(ctx.MemoryID)
	}
}

// TestGeneratorIsNotCalled proves ingest never calls the injected model.
func TestGeneratorIsNotCalled(t *testing.T) {
	s := openWorld(t)
	g := &boom{}
	s.SetGenerator(g)
	if _, err := s.Ingest(bytes.NewReader(sampleLog(t, true)), 0); err != nil {
		t.Fatal(err)
	}
	if g.called {
		t.Fatal("ingest called the generator")
	}
}

// boom is a generator that records whether it was called.
type boom struct{ called bool }

func (b *boom) Complete(context.Context, string, string) (string, error) {
	b.called = true
	return "", nil
}

// sampleLog is one finding's trace: claim, evidence, refutation and verdict.
// withEvidence sets whether the refutation was of a finding with proof.
func sampleLog(t *testing.T, withEvidence bool) []byte {
	t.Helper()
	events := []trace.Event{
		{EventID: "e1", FindingID: "F1", Kind: "claim", Agent: "planner", Text: "Discount is applied twice",
			ChangeKey: "chg", RuleID: "discount-applied-twice", Path: "billing/discount.go", Title: "Discount is applied twice"},
		{EventID: "e2", FindingID: "F1", Kind: "evidence", Agent: "prover", Text: "test billing/discount_test.go TestDiscounted",
			ChangeKey: "chg", RuleID: "discount-applied-twice", Path: "billing/discount.go", Title: "Discount is applied twice"},
		{EventID: "e3", FindingID: "F1", Kind: "refutation", Agent: "skeptic", Text: "the helper applies it once",
			ChangeKey: "chg", RuleID: "discount-applied-twice", Path: "billing/discount.go", Title: "Discount is applied twice",
			HasEvidence: withEvidence,
			Refutation:  &trace.Refutation{Attempted: true, Refuted: true, Reason: "the helper applies it once"}},
		{EventID: "e4", FindingID: "F1", Kind: "verdict", Agent: "judge", Text: "kept",
			ChangeKey: "chg", RuleID: "discount-applied-twice", Path: "billing/discount.go", Title: "Discount is applied twice"},
	}
	var buf bytes.Buffer
	for _, ev := range events {
		b, err := json.Marshal(ev)
		if err != nil {
			t.Fatal(err)
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

func openWorld(t *testing.T) *world.Store {
	t.Helper()
	return openAt(t, t.TempDir(), t.TempDir())
}

func openAt(t *testing.T, dir, logs string) *world.Store {
	t.Helper()
	// Memory lives beside the world directory so a reopen sees the same records.
	mem, err := hmem.Open(hmem.Options{Dir: dir + "-hmem", LogDir: logs})
	if err != nil {
		t.Fatal(err)
	}
	s, err := world.Open(world.Options{Dir: dir, LogDir: logs, Memory: mem})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
