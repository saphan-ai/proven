package world_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/world"
)

// graphFixture opens a store holding one player and two teams, the shape the
// Meterless bitemp-001 check builds.
func graphFixture(t *testing.T) (*world.Store, string, string) {
	t.Helper()
	dir, logs := t.TempDir(), t.TempDir()
	s := openAt(t, dir, logs)
	src := world.Provenance{Kind: "manual", At: "2087-01-01T00:00:00.000Z"}
	for _, id := range []string{"player:g:p1", "team:g:a", "team:g:b", "team:g:c"} {
		typ := strings.SplitN(id, ":", 2)[0]
		if err := s.UpsertEntity(world.Entity{Type: typ, ID: id, Source: src}); err != nil {
			t.Fatal(err)
		}
	}
	return s, dir, logs
}

func relate(t *testing.T, s *world.Store, to, from string) world.Edge {
	t.Helper()
	e, err := s.Relate(world.Edge{
		From: "player:g:p1", Type: "plays-for", To: to, ValidFrom: from,
		Source: world.Provenance{Kind: "manual", At: from},
	})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func ids(es []world.Entity) []string {
	out := []string{}
	for _, e := range es {
		out = append(out, e.ID)
	}
	return out
}

// TestSupersedingEdgeClosesTheOldOne proves a new open edge with the same
// from, type and context but another target closes the old edge at the new
// edge's start and never deletes it, and that "now" sees only the new one.
func TestSupersedingEdgeClosesTheOldOne(t *testing.T) {
	s, _, _ := graphFixture(t)
	relate(t, s, "team:g:a", "2087-01-01T00:00:00.000Z")
	b := relate(t, s, "team:g:b", "2088-03-01T00:00:00.000Z")
	if got := ids(s.Neighbors("player:g:p1", "plays-for", "")); len(got) != 1 || got[0] != "team:g:b" {
		t.Fatalf("now: %v", got)
	}
	var closed int
	for _, r := range edgesOf(t, s) {
		if r.To == "team:g:a" {
			closed++
			if r.ValidTo != "2088-03-01T00:00:00.000Z" || r.SupersededBy != b.ID {
				t.Fatalf("old edge not closed at the new start: %+v", r)
			}
		}
	}
	if closed != 1 {
		t.Fatalf("the old edge must be kept, found %d", closed)
	}
}

// TestPointInTimeQueryRespectsValidity proves an at-query returns the edge
// valid then (ValidFrom <= at < ValidTo), including one since closed, and
// nothing before the first edge began or exactly at a closing instant.
func TestPointInTimeQueryRespectsValidity(t *testing.T) {
	s, _, _ := graphFixture(t)
	relate(t, s, "team:g:a", "2087-01-01T00:00:00.000Z")
	relate(t, s, "team:g:b", "2088-03-01T00:00:00.000Z")
	cases := []struct{ at, want string }{
		{"2086-12-31T00:00:00.000Z", ""},
		{"2087-06-01T00:00:00.000Z", "team:g:a"},
		{"2088-03-01T00:00:00.000Z", "team:g:b"},
	}
	for _, c := range cases {
		got := ids(s.Neighbors("player:g:p1", "plays-for", c.at))
		if c.want == "" && len(got) != 0 || c.want != "" && (len(got) != 1 || got[0] != c.want) {
			t.Fatalf("at %s: got %v want %q", c.at, got, c.want)
		}
	}
}

// TestRelateIsIdempotentAndAudited proves the same edge written twice is one
// edge and one audit line, and that a write with no provenance is refused.
func TestRelateIsIdempotentAndAudited(t *testing.T) {
	s, _, logs := graphFixture(t)
	relate(t, s, "team:g:a", "2087-01-01T00:00:00.000Z")
	relate(t, s, "team:g:a", "2087-01-01T00:00:00.000Z")
	if n := len(edgesOf(t, s)); n != 1 {
		t.Fatalf("edges %d", n)
	}
	b, err := os.ReadFile(filepath.Join(logs, "world.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(b), `"kind":"relate"`); n != 1 {
		t.Fatalf("relate audit lines %d", n)
	}
	if _, err := s.Relate(world.Edge{From: "player:g:p1", Type: "plays-for", To: "team:g:a"}); err == nil {
		t.Fatal("an edge with no provenance was accepted")
	}
}

// TestEdgesSurviveReopen proves closed and open edges persist in state.json,
// so a point-in-time query gives the same answer after a restart.
func TestEdgesSurviveReopen(t *testing.T) {
	s, dir, logs := graphFixture(t)
	relate(t, s, "team:g:a", "2087-01-01T00:00:00.000Z")
	relate(t, s, "team:g:b", "2088-03-01T00:00:00.000Z")
	again := openAt(t, dir, logs)
	if got := ids(again.Neighbors("player:g:p1", "plays-for", "2087-06-01T00:00:00.000Z")); len(got) != 1 || got[0] != "team:g:a" {
		t.Fatalf("after reopen: %v", got)
	}
}

// wantNeighbors fails unless the neighbours of the player at that instant are
// exactly want, in id order. An empty at asks for "now".
func wantNeighbors(t *testing.T, s *world.Store, at string, want ...string) {
	t.Helper()
	got := ids(s.Neighbors("player:g:p1", "plays-for", at))
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("at %q: got %v want %v", at, got, want)
	}
}

// TestRepeatedTargetRenewsOneEdge proves writing the same target again at a
// later time keeps one edge for that stretch: A, A again, then B leaves A
// valid once (not twice) and closed exactly where B starts. The rule is
// renewal: the open edge already says A, so the repeat adds nothing.
func TestRepeatedTargetRenewsOneEdge(t *testing.T) {
	s, _, _ := graphFixture(t)
	relate(t, s, "team:g:a", "2087-01-01T00:00:00.000Z")
	relate(t, s, "team:g:a", "2088-01-01T00:00:00.000Z")
	wantNeighbors(t, s, "2088-06-01T00:00:00.000Z", "team:g:a")
	relate(t, s, "team:g:b", "2089-01-01T00:00:00.000Z")
	wantNeighbors(t, s, "2088-06-01T00:00:00.000Z", "team:g:a")
	wantNeighbors(t, s, "2087-06-01T00:00:00.000Z", "team:g:a")
	wantNeighbors(t, s, "2089-06-01T00:00:00.000Z", "team:g:b")
	wantNeighbors(t, s, "", "team:g:b")
	if n := len(edgesOf(t, s)); n != 2 {
		t.Fatalf("a renewed target must stay one edge, found %d edges in all", n)
	}
}

// TestReverseTimeInsertionKeepsIntervalsOrdered proves a write that starts
// before an open edge goes into history instead of over it: B from 2089 is
// written first, then C from 2088. C is valid until B starts, B stays open,
// and no edge ends before it begins.
func TestReverseTimeInsertionKeepsIntervalsOrdered(t *testing.T) {
	s, _, _ := graphFixture(t)
	relate(t, s, "team:g:b", "2089-01-01T00:00:00.000Z")
	relate(t, s, "team:g:a", "2088-01-01T00:00:00.000Z")
	wantNeighbors(t, s, "2087-06-01T00:00:00.000Z")
	wantNeighbors(t, s, "2088-06-01T00:00:00.000Z", "team:g:a")
	wantNeighbors(t, s, "2089-06-01T00:00:00.000Z", "team:g:b")
	wantNeighbors(t, s, "", "team:g:b")
	for _, r := range edgesOf(t, s) {
		if r.ValidTo != "" && r.ValidTo <= r.ValidFrom {
			t.Fatalf("edge ends before it starts: %+v", r)
		}
	}
	// A write in the middle of a closed stretch splits it, never overlaps it.
	relate(t, s, "team:g:c", "2088-07-01T00:00:00.000Z")
	wantNeighbors(t, s, "2088-06-01T00:00:00.000Z", "team:g:a")
	wantNeighbors(t, s, "2088-08-01T00:00:00.000Z", "team:g:c")
	wantNeighbors(t, s, "2089-06-01T00:00:00.000Z", "team:g:b")
}

// TestEquivalentOffsetsAreTheSameInstant proves timestamps are compared as
// instants: an edge starting 01:00+01:00 (00:00Z) has started at 00:30Z, a
// boundary written with one offset closes an edge written with another at
// the same instant, and one instant in two spellings is one edge.
func TestEquivalentOffsetsAreTheSameInstant(t *testing.T) {
	s, _, _ := graphFixture(t)
	relate(t, s, "team:g:a", "2087-01-01T01:00:00+01:00")
	wantNeighbors(t, s, "2087-01-01T00:30:00Z", "team:g:a")
	wantNeighbors(t, s, "2086-12-31T23:59:00Z")
	relate(t, s, "team:g:b", "2088-01-01T05:00:00+05:00")
	wantNeighbors(t, s, "2087-12-31T23:59:00Z", "team:g:a")
	wantNeighbors(t, s, "2088-01-01T00:00:00Z", "team:g:b")
	wantNeighbors(t, s, "2088-01-01T09:00:00+09:00", "team:g:b")
	// The same instant written in another offset is the same edge.
	relate(t, s, "team:g:b", "2088-01-01T00:00:00Z")
	if n := len(edgesOf(t, s)); n != 2 {
		t.Fatalf("equivalent instants must name one edge, found %d edges", n)
	}
}

// TestUnparsableTimestampIsRefusedByName proves a timestamp that is not an
// RFC 3339 instant is refused on write and on query, with an error naming it.
func TestUnparsableTimestampIsRefusedByName(t *testing.T) {
	s, _, _ := graphFixture(t)
	_, err := s.Relate(world.Edge{
		From: "player:g:p1", Type: "plays-for", To: "team:g:a", ValidFrom: "last spring",
		Source: world.Provenance{Kind: "manual", At: "2087-01-01T00:00:00.000Z"},
	})
	if err == nil || !errors.Is(err, world.ErrTimestamp) {
		t.Fatalf("an unparsable valid_from was not refused by name: %v", err)
	}
	if n := len(edgesOf(t, s)); n != 0 {
		t.Fatalf("a refused write stored %d edges", n)
	}
	if _, err := world.ParseInstant("2087-01-01"); !errors.Is(err, world.ErrTimestamp) {
		t.Fatalf("a date without a time was accepted: %v", err)
	}
}

// edgesOf reads the edges back through the canonical state, the store's one
// public view of what it holds.
func edgesOf(t *testing.T, s *world.Store) []world.Edge {
	t.Helper()
	var d struct {
		Edges []world.Edge `json:"edges"`
	}
	if err := json.Unmarshal([]byte(s.Canonical()), &d); err != nil {
		t.Fatal(err)
	}
	return d.Edges
}
