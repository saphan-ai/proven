package world_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/world"
)

var storedSrc = world.Provenance{Kind: "manual", At: "2087-01-01T00:00:00.000Z"}

// writeEdges replaces the edges in the state file of dir, the way an earlier
// writer could have left them. Ids are fixed so a refusal can be checked for
// the names it gives.
func writeEdges(t *testing.T, dir string, edges []world.Edge) {
	t.Helper()
	path := filepath.Join(dir, "state.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var d map[string]json.RawMessage
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatal(err)
	}
	if d["edges"], err = json.Marshal(edges); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatal(err)
	}
}

// reopen opens the state in dir the way openAt does, and returns the error
// instead of failing, so a test can look at a refusal.
func reopen(t *testing.T, dir, logs string) (*world.Store, error) {
	t.Helper()
	mem, err := hmem.Open(hmem.Options{Dir: dir + "-hmem", LogDir: logs})
	if err != nil {
		t.Fatal(err)
	}
	return world.Open(world.Options{Dir: dir, LogDir: logs, Memory: mem})
}

func plays(id, to, from, until string) world.Edge {
	return world.Edge{ID: id, From: "player:g:p1", Type: "plays-for", To: to,
		ValidFrom: from, ValidTo: until, Source: storedSrc}
}

// TestOpenRefusesStoredEdgesThatOverlap proves a state file written by the
// former writer, a bounded edge [2087, 2089) and an open renewal from 2088 for
// the same team, is refused at Open with ErrEdgeOverlap naming both edge ids,
// so Neighbors never returns two edges for one instant.
func TestOpenRefusesStoredEdgesThatOverlap(t *testing.T) {
	_, dir, logs := graphFixture(t)
	writeEdges(t, dir, []world.Edge{
		plays("rel_bounded", "team:g:a", "2087-01-01T00:00:00.000Z", "2089-01-01T00:00:00.000Z"),
		plays("rel_renewal", "team:g:a", "2088-01-01T00:00:00.000Z", ""),
	})
	s, err := reopen(t, dir, logs)
	if !errors.Is(err, world.ErrEdgeOverlap) {
		t.Fatalf("a state with overlapping edges was opened: %v", err)
	}
	if s != nil {
		t.Fatal("a refused open returned a store")
	}
	for _, id := range []string{"rel_bounded", "rel_renewal"} {
		if !strings.Contains(err.Error(), id) {
			t.Fatalf("the refusal does not name %s: %v", id, err)
		}
	}
}

// TestOpenAcceptsTouchingAndSeparateChains proves the check is no wider than
// the write rule: edges that only touch, and overlapping times in another
// context or another type, open fine.
func TestOpenAcceptsTouchingAndSeparateChains(t *testing.T) {
	_, dir, logs := graphFixture(t)
	loan := plays("rel_3", "team:g:c", "2088-01-01T00:00:00.000Z", "")
	loan.Context = "loan"
	coach := plays("rel_4", "team:g:c", "2088-01-01T00:00:00.000Z", "")
	coach.Type = "coached-by"
	writeEdges(t, dir, []world.Edge{
		plays("rel_1", "team:g:a", "2087-01-01T00:00:00.000Z", "2089-01-01T00:00:00.000Z"),
		plays("rel_2", "team:g:b", "2089-01-01T00:00:00.000Z", ""),
		loan, coach,
	})
	if _, err := reopen(t, dir, logs); err != nil {
		t.Fatalf("a valid state was refused: %v", err)
	}
}

// TestOpenRefusesAnOverlapOfAnotherTarget proves the check is on time, not on
// target: a long edge for one team that covers a later bounded edge for
// another team is refused too, whatever order the file lists them in.
func TestOpenRefusesAnOverlapOfAnotherTarget(t *testing.T) {
	_, dir, logs := graphFixture(t)
	writeEdges(t, dir, []world.Edge{
		plays("rel_late", "team:g:b", "2090-01-01T00:00:00.000Z", "2091-01-01T00:00:00.000Z"),
		plays("rel_long", "team:g:a", "2087-01-01T00:00:00.000Z", "2095-01-01T00:00:00.000Z"),
	})
	if _, err := reopen(t, dir, logs); !errors.Is(err, world.ErrEdgeOverlap) {
		t.Fatalf("a long edge over a later one was opened: %v", err)
	}
}

// TestOpenRefusesAnOverlapHiddenByAReversedEdge proves the reviewer's case: two
// edges valid at one instant, 2089, with a stored reversed interval between
// them in start order. The file is refused, and the reversed edge is named, so
// Neighbors never returns two edges for 2089.
func TestOpenRefusesAnOverlapHiddenByAReversedEdge(t *testing.T) {
	_, dir, logs := graphFixture(t)
	writeEdges(t, dir, []world.Edge{
		plays("rel_long", "team:g:a", "2087-01-01T00:00:00.000Z", "2097-01-01T00:00:00.000Z"),
		plays("rel_reversed", "team:g:b", "2088-01-01T00:00:00.000Z", "2087-01-01T00:00:00.000Z"),
		plays("rel_short", "team:g:c", "2089-01-01T00:00:00.000Z", "2090-01-01T00:00:00.000Z"),
	})
	s, err := reopen(t, dir, logs)
	if err == nil || s != nil {
		t.Fatalf("a state with a reversed edge was opened: %v", err)
	}
	if !errors.Is(err, world.ErrEdgeInterval) || !strings.Contains(err.Error(), "rel_reversed") {
		t.Fatalf("the refusal does not name the reversed edge: %v", err)
	}
}

// TestOpenRefusesAnOverlapOverAMiddleEdge proves the overlap scan does not
// rest on neighbours: a long edge and a late edge overlap with a short valid
// edge between them in start order, and the refusal names the long edge.
func TestOpenRefusesAnOverlapOverAMiddleEdge(t *testing.T) {
	_, dir, logs := graphFixture(t)
	writeEdges(t, dir, []world.Edge{
		plays("rel_long", "team:g:a", "2087-01-01T00:00:00.000Z", "2097-01-01T00:00:00.000Z"),
		plays("rel_mid", "team:g:b", "2087-06-01T00:00:00.000Z", "2087-09-01T00:00:00.000Z"),
		plays("rel_late", "team:g:c", "2089-01-01T00:00:00.000Z", "2090-01-01T00:00:00.000Z"),
	})
	_, err := reopen(t, dir, logs)
	if !errors.Is(err, world.ErrEdgeOverlap) {
		t.Fatalf("an overlap over a middle edge was opened: %v", err)
	}
	if !strings.Contains(err.Error(), "rel_long") {
		t.Fatalf("the refusal does not name the long edge: %v", err)
	}
}

// TestOpenRefusesOverlapWithFurthestMiddleEdge checks the running far end of a chain.
func TestOpenRefusesOverlapWithFurthestMiddleEdge(t *testing.T) {
	_, dir, logs := graphFixture(t)
	writeEdges(t, dir, []world.Edge{
		plays("rel_first", "team:g:a", "2087-01-01T00:00:00.000Z", "2088-01-01T00:00:00.000Z"),
		plays("rel_middle", "team:g:b", "2089-01-01T00:00:00.000Z", "2095-01-01T00:00:00.000Z"),
		plays("rel_last", "team:g:c", "2091-01-01T00:00:00.000Z", "2092-01-01T00:00:00.000Z"),
	})
	_, err := reopen(t, dir, logs)
	if !errors.Is(err, world.ErrEdgeOverlap) {
		t.Fatalf("middle edge overlap was opened: %v", err)
	}
	if !strings.Contains(err.Error(), "rel_middle") || !strings.Contains(err.Error(), "rel_last") {
		t.Fatalf("refusal does not name the overlapping edges: %v", err)
	}
}

// TestOpenRefusesAnUnreadableStoredStart proves a stored edge whose valid_from
// cannot be read is refused at Open by name, open edges included. Without the
// check, two open edges of one chain, one with a start "invalid", load, and
// Neighbors with no time returns both targets.
func TestOpenRefusesAnUnreadableStoredStart(t *testing.T) {
	for name, edges := range map[string][]world.Edge{
		"open pair": {
			plays("rel_bad", "team:g:a", "invalid", ""),
			plays("rel_ok", "team:g:b", "2087-01-01T00:00:00.000Z", ""),
		},
		"bounded": {
			plays("rel_bad", "team:g:a", "invalid", "2089-01-01T00:00:00.000Z"),
		},
		"empty start": {
			plays("rel_bad", "team:g:a", "", ""),
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, dir, logs := graphFixture(t)
			writeEdges(t, dir, edges)
			s, err := reopen(t, dir, logs)
			if err == nil || s != nil {
				t.Fatalf("a state with an unreadable start was opened: %v", err)
			}
			if !errors.Is(err, world.ErrEdgeInterval) || !strings.Contains(err.Error(), "rel_bad") {
				t.Fatalf("the refusal does not name the edge: %v", err)
			}
		})
	}
}

// TestOpenRefusesAnEmptyOrUnreadableInterval proves each bad stored bound is
// refused by name: an empty interval (valid_to equal to valid_from) and an end
// that is not a time.
func TestOpenRefusesAnEmptyOrUnreadableInterval(t *testing.T) {
	for name, until := range map[string]string{
		"empty":      "2087-01-01T00:00:00.000Z",
		"unreadable": "not-a-time",
	} {
		t.Run(name, func(t *testing.T) {
			_, dir, logs := graphFixture(t)
			writeEdges(t, dir, []world.Edge{
				plays("rel_bad", "team:g:a", "2087-01-01T00:00:00.000Z", until),
			})
			if _, err := reopen(t, dir, logs); !errors.Is(err, world.ErrEdgeInterval) {
				t.Fatalf("a bad interval was opened: %v", err)
			}
		})
	}
}
