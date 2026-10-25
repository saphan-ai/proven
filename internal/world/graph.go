package world

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/saphan-ai/proven/internal/trace"
)

// ErrEdgeConflict is returned, wrapped, when an edge would start at the very
// instant another target's edge already starts for the same from, type and
// context. Neither could be closed at that instant without leaving an empty
// interval, so the write is refused and the caller decides which one is true.
var ErrEdgeConflict = errors.New("world: another target already starts at that instant")

// Edge is one bitemporal relationship between two entities. ValidTo empty
// means the edge is open. A superseding edge closes the old one by setting
// ValidTo and SupersededBy; an edge is never deleted (ADR-0018 point 1,
// Meterless World Model specification). ValidFrom and ValidTo are held as
// canonical UTC instants (see ParseInstant).
type Edge struct {
	ID           string            `json:"id"`
	From         string            `json:"from"`
	Type         string            `json:"type"`
	To           string            `json:"to"`
	Context      string            `json:"context,omitempty"`
	Attrs        map[string]string `json:"attrs,omitempty"`
	ValidFrom    string            `json:"valid_from"`
	ValidTo      string            `json:"valid_to,omitempty"`
	SupersededBy string            `json:"superseded_by,omitempty"`
	Source       Provenance        `json:"source"`
}

// edgeID is the content address of an edge: the same from, type, to, context
// and start time always name the same edge, so repeating a write is a no-op.
// validFrom is the canonical instant, so two spellings of one instant agree.
func edgeID(from, typ, to, context, validFrom string) string {
	sum := sha256.Sum256([]byte(from + "\x00" + typ + "\x00" + to + "\x00" + context + "\x00" + validFrom))
	return "rel_" + hex.EncodeToString(sum[:])[:16]
}

// sameChain reports whether two edges share the from, type and context whose
// targets supersede one another. An edge that was closed by the caller with
// an explicit ValidTo (no SupersededBy) is not part of the chain: it keeps
// the history it recorded.
func sameChain(a, b Edge) bool {
	if a.From != b.From || a.Type != b.Type || a.Context != b.Context {
		return false
	}
	return a.ValidTo == "" || a.SupersededBy != ""
}

// Relate writes one edge. Timestamps are parsed to instants and stored in
// canonical UTC form; one that is not an RFC 3339 instant is refused with
// ErrTimestamp. An empty ValidFrom takes the provenance time.
//
// An open edge joins the history of its from, type and context in time
// order, whatever order the writes arrive in:
//   - If the edge valid at its start already has the same target, the write
//     renews it: the existing edge is returned and nothing is added, so
//     history never holds two edges for one stretch of one target.
//   - If the edge valid at its start has another target, that edge is closed
//     at the new start, and the new edge keeps the end the old one had.
//   - If nothing is valid at its start but a later edge exists, the new edge
//     ends where that later edge begins; it never covers it.
//
// An edge already stored under the same content address is returned as it is.
// A failed audit or persist puts the edge map back as it was.
func (s *Store) Relate(e Edge) (Edge, error) {
	if err := checkProvenance(e.Source); err != nil {
		return Edge{}, err
	}
	if e.From == "" || e.Type == "" || e.To == "" {
		return Edge{}, fmt.Errorf("world: an edge needs a from, a type and a to")
	}
	// ValidFrom defaults to the provenance time, as the specification says.
	if e.ValidFrom == "" {
		e.ValidFrom = e.Source.At
	}
	from, err := ParseInstant(e.ValidFrom)
	if err != nil {
		return Edge{}, fmt.Errorf("world: valid_from: %w", err)
	}
	e.ValidFrom = from
	if e.ValidTo != "" {
		to, err := ParseInstant(e.ValidTo)
		if err != nil {
			return Edge{}, fmt.Errorf("world: valid_to: %w", err)
		}
		// An interval that ends before it starts would be invisible to every
		// point-in-time query, so it is refused rather than stored.
		if tf, tt := mustInstant(from), mustInstant(to); !tt.After(tf) {
			return Edge{}, fmt.Errorf("world: valid_to %s is not after valid_from %s", to, from)
		}
		e.ValidTo = to
	}
	e.ID = edgeID(e.From, e.Type, e.To, e.Context, e.ValidFrom)
	if prev, ok := s.edges[e.ID]; ok {
		return prev, nil
	}
	snap := cloneEdges(s.edges)
	if e.ValidTo == "" {
		renewed, found, err := s.place(&e)
		if err != nil {
			return Edge{}, err
		}
		if found {
			return renewed, nil
		}
	}
	s.edges[e.ID] = e
	actor := e.Source.By
	if actor == "" {
		actor = "system"
	}
	err = s.audit(trace.Event{
		EventID:   "relate:" + e.ID,
		Kind:      "relate",
		Agent:     actor,
		FindingID: e.From,
		Text:      e.From + " -[" + e.Type + "]-> " + e.To,
	})
	if err == nil {
		err = s.persist()
	}
	if err != nil {
		s.edges = snap
		return Edge{}, err
	}
	return e, nil
}

// place fits the open edge e into the history of its chain, in time order. It
// closes the edge valid at e's start when that edge has another target, and
// gives e the end of the stretch it lands in. When the edge valid at e's
// start already has the same target it returns that edge with found true, and
// changes nothing: the write is a renewal.
func (s *Store) place(e *Edge) (Edge, bool, error) {
	start := mustInstant(e.ValidFrom)
	var covering, next *Edge
	var coveringID string
	var coveringStart, nextStart time.Time
	for id, o := range s.edges {
		o := o
		if !sameChain(o, *e) {
			continue
		}
		of, ok := instantOf(o.ValidFrom)
		if !ok {
			continue
		}
		if of.After(start) {
			// The earliest later edge bounds the new one.
			if next == nil || of.Before(nextStart) {
				next, nextStart = &o, of
			}
			continue
		}
		if o.ValidTo != "" {
			if end, ok := instantOf(o.ValidTo); !ok || !start.Before(end) {
				continue
			}
		}
		// The latest start at or before ours is the edge valid at our start.
		if covering == nil || of.After(coveringStart) {
			covering, coveringID, coveringStart = &o, id, of
		}
	}
	if covering != nil {
		if covering.To == e.To {
			return *covering, true, nil
		}
		if covering.ValidFrom == e.ValidFrom {
			return Edge{}, false, fmt.Errorf("%w: %s at %s", ErrEdgeConflict, e.From, e.ValidFrom)
		}
		e.ValidTo, e.SupersededBy = covering.ValidTo, covering.SupersededBy
		covering.ValidTo, covering.SupersededBy = e.ValidFrom, e.ID
		s.edges[coveringID] = *covering
		return Edge{}, false, nil
	}
	if next != nil {
		e.ValidTo, e.SupersededBy = next.ValidFrom, next.ID
	}
	return Edge{}, false, nil
}

// mustInstant parses a timestamp that Relate has already canonicalised.
func mustInstant(s string) time.Time {
	t, _ := instantOf(s)
	return t
}

// Neighbors returns the entities joined to id by an edge, sorted by id. A
// non-empty via keeps one edge type. With at empty only open edges count
// ("now"); with at set, an edge counts when ValidFrom <= at < ValidTo,
// compared as instants, so a closed edge still answers for the time it was
// valid and equivalent offsets agree. An at that is not an RFC 3339 instant
// matches nothing: callers that must refuse it check ParseInstant first. An
// edge whose other end is not a stored entity is skipped.
func (s *Store) Neighbors(id, via, at string) []Entity {
	out := []Entity{}
	var when time.Time
	if at != "" {
		t, ok := instantOf(at)
		if !ok {
			return out
		}
		when = t
	}
	for _, r := range s.edges {
		if via != "" && r.Type != via {
			continue
		}
		if at != "" {
			from, ok := instantOf(r.ValidFrom)
			if !ok || when.Before(from) {
				continue
			}
			if r.ValidTo != "" {
				if to, ok := instantOf(r.ValidTo); !ok || !when.Before(to) {
					continue
				}
			}
		} else if r.ValidTo != "" {
			continue
		}
		other := ""
		switch id {
		case r.From:
			other = r.To
		case r.To:
			other = r.From
		default:
			continue
		}
		if ent, ok := s.entities[other]; ok {
			out = append(out, ent)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func cloneEdges(in map[string]Edge) map[string]Edge {
	out := make(map[string]Edge, len(in))
	for id, e := range in {
		e.Attrs = cloneAttrs(e.Attrs)
		out[id] = e
	}
	return out
}
