package world_test

import (
	"errors"
	"testing"

	"github.com/saphan-ai/proven/internal/world"
)

// relateUntil writes an edge of the player with an explicit end and returns
// the store's answer, so a test can look at a refusal as well as a success.
func relateUntil(s *world.Store, to, from, until string) (world.Edge, error) {
	return s.Relate(world.Edge{
		From: "player:g:p1", Type: "plays-for", To: to, ValidFrom: from, ValidTo: until,
		Source: world.Provenance{Kind: "manual", At: from},
	})
}

// TestBoundedEdgeThenOpenRenewalNeverOverlaps proves a same-target open write
// that starts inside an edge with an explicit end is refused by name, so
// the instant inside the overlap has one edge and one neighbour, not two.
func TestBoundedEdgeThenOpenRenewalNeverOverlaps(t *testing.T) {
	s, _, _ := graphFixture(t)
	if _, err := relateUntil(s, "team:g:a", "2087-01-01T00:00:00.000Z", "2089-01-01T00:00:00.000Z"); err != nil {
		t.Fatal(err)
	}
	_, err := relateUntil(s, "team:g:a", "2088-01-01T00:00:00.000Z", "")
	if !errors.Is(err, world.ErrEdgeOverlap) {
		t.Fatalf("an open renewal inside a bounded edge was not refused by name: %v", err)
	}
	wantNeighbors(t, s, "2088-06-01T00:00:00.000Z", "team:g:a")
	es := edgesOf(t, s)
	if len(es) != 1 || es[0].ValidTo != "2089-01-01T00:00:00.000Z" {
		t.Fatalf("the first interval must stay as written: %+v", es)
	}
	// A write that starts where the bounded edge ends does not overlap it.
	if _, err := relateUntil(s, "team:g:a", "2089-01-01T00:00:00.000Z", ""); err != nil {
		t.Fatalf("an edge starting at the end of another was refused: %v", err)
	}
	wantNeighbors(t, s, "2089-06-01T00:00:00.000Z", "team:g:a")
}

// TestBoundedEdgeThenBoundedRenewalNeverOverlaps proves a same-target bounded
// write that overlaps an earlier bounded edge is refused by name, and so is a
// bounded write over an edge of another target in the same chain.
func TestBoundedEdgeThenBoundedRenewalNeverOverlaps(t *testing.T) {
	s, _, _ := graphFixture(t)
	if _, err := relateUntil(s, "team:g:a", "2087-01-01T00:00:00.000Z", "2089-01-01T00:00:00.000Z"); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ to, from, until string }{
		{"team:g:a", "2088-01-01T00:00:00.000Z", "2090-01-01T00:00:00.000Z"},
		{"team:g:a", "2086-01-01T00:00:00.000Z", "2087-06-01T00:00:00.000Z"},
		{"team:g:b", "2088-01-01T00:00:00.000Z", "2088-09-01T00:00:00.000Z"},
	} {
		if _, err := relateUntil(s, c.to, c.from, c.until); !errors.Is(err, world.ErrEdgeOverlap) {
			t.Fatalf("%+v: an overlapping bounded write was not refused by name: %v", c, err)
		}
	}
	wantNeighbors(t, s, "2088-06-01T00:00:00.000Z", "team:g:a")
	if n := len(edgesOf(t, s)); n != 1 {
		t.Fatalf("refused writes stored edges: %d in all", n)
	}
	// Touching intervals are fine: [2089, 2090) follows [2087, 2089).
	if _, err := relateUntil(s, "team:g:a", "2089-01-01T00:00:00.000Z", "2090-01-01T00:00:00.000Z"); err != nil {
		t.Fatal(err)
	}
	wantNeighbors(t, s, "2089-06-01T00:00:00.000Z", "team:g:a")
}

// TestOpenWriteOverEarlierBoundedEdgeOfAnotherTarget proves the overlap rule
// reaches an edge closed by the caller: an open write for another target that
// starts inside it is refused, not left valid beside it.
func TestOpenWriteOverEarlierBoundedEdgeOfAnotherTarget(t *testing.T) {
	s, _, _ := graphFixture(t)
	if _, err := relateUntil(s, "team:g:a", "2087-01-01T00:00:00.000Z", "2089-01-01T00:00:00.000Z"); err != nil {
		t.Fatal(err)
	}
	if _, err := relateUntil(s, "team:g:b", "2088-01-01T00:00:00.000Z", ""); !errors.Is(err, world.ErrEdgeOverlap) {
		t.Fatalf("not refused by name: %v", err)
	}
	wantNeighbors(t, s, "2088-06-01T00:00:00.000Z", "team:g:a")
}
