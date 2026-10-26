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

// ErrEdgeOverlap is returned, wrapped, when a write would leave two edges of
// the same from, type and context valid at one instant, and no renewal or
// supersession can resolve it. Rule: a chain holds at most one valid edge at
// any instant. An open write resolves an overlap with an edge in the chain
// (renewal, or closing the old edge); an edge the caller bounded with an
// explicit valid_to is a recorded fact and is never rewritten, so a write
// that overlaps it is refused and the caller decides which one is true.
var ErrEdgeOverlap = errors.New("world: the edge overlaps another valid edge of the same from, type and context")

// ErrEdgeInterval is returned, wrapped, for an edge whose valid_to is not
// after its valid_from, or whose valid_from or valid_to cannot be read. Rule: an interval is non-empty, so
// it can be seen by a point-in-time query. Relate refuses such a write and
// Open refuses such a stored edge, from one check (checkInterval).
var ErrEdgeInterval = errors.New("world: the edge interval ends before it starts")

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
// A chain never holds two edges valid at one instant. A write, open or
// bounded, that still overlaps an edge after the steps above (typically one
// the caller bounded with valid_to) is refused with ErrEdgeOverlap and
// changes nothing.
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
		if err := checkInterval(from, to); err != nil {
			return Edge{}, err
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
	// A bounded write skips place, and place skips edges the caller bounded,
	// so the invariant is checked here, after place has made its closings: no
	// other edge of the chain may be valid at any instant e is. A refusal puts
	// back what place changed.
	if clash, ok := s.overlapping(e); ok {
		s.edges = snap
		return Edge{}, fmt.Errorf("%w: %s [%s, %s) meets %s [%s, %s)", ErrEdgeOverlap,
			e.To, e.ValidFrom, openEnd(e.ValidTo), clash.To, clash.ValidFrom, openEnd(clash.ValidTo))
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

// overlapping returns a stored edge of e's from, type and context whose
// interval [ValidFrom, ValidTo) shares an instant with e's, an empty ValidTo
// being open-ended. Edges that only touch (one ends where the other starts)
// do not overlap. An edge whose time cannot be read is skipped, as Neighbors
// skips it.
func (s *Store) overlapping(e Edge) (Edge, bool) {
	for _, o := range s.edges {
		if o.From != e.From || o.Type != e.Type || o.Context != e.Context {
			continue
		}
		if intervalsOverlap(e, o) {
			return o, true
		}
	}
	return Edge{}, false
}

// intervalsOverlap reports whether the intervals [ValidFrom, ValidTo) of a and
// b share an instant, an empty ValidTo being open-ended. It does not compare
// from, type or context; the caller picks the chain. One rule serves both the
// write check and the check Open makes on a stored file, so the two cannot
// drift. An edge whose start cannot be read never overlaps, as Neighbors
// skips it; an unreadable end is read as open.
func intervalsOverlap(a, b Edge) bool {
	af, ok := instantOf(a.ValidFrom)
	if !ok {
		return false
	}
	bf, ok := instantOf(b.ValidFrom)
	if !ok {
		return false
	}
	// Each must start before the other ends.
	if a.ValidTo != "" {
		if end, ok := instantOf(a.ValidTo); ok && !bf.Before(end) {
			return false
		}
	}
	if b.ValidTo != "" {
		if end, ok := instantOf(b.ValidTo); ok && !af.Before(end) {
			return false
		}
	}
	return true
}

// checkInterval refuses an interval [from, to) that is empty or reversed, or
// whose start or end cannot be read, with ErrEdgeInterval. An empty to is
// open and accepted. Both Relate and Open call it, so a write and a stored
// edge are held to one rule. A start that cannot be read is refused for open
// edges too: an open edge's query path never parses its start, so such an
// edge would load and be returned beside another open edge of its chain.
func checkInterval(from, to string) error {
	start, ok := instantOf(from)
	if !ok {
		return fmt.Errorf("%w: valid_from %q cannot be read", ErrEdgeInterval, from)
	}
	if to == "" {
		return nil
	}
	end, ok := instantOf(to)
	if !ok {
		return fmt.Errorf("%w: valid_to %q cannot be read", ErrEdgeInterval, to)
	}
	if !end.After(start) {
		return fmt.Errorf("%w: valid_to %s is not after valid_from %s", ErrEdgeInterval, to, from)
	}
	return nil
}

// endsAfter reports whether a ends strictly later than b, an empty ValidTo
// being open-ended. Both intervals have passed checkInterval.
func endsAfter(a, b Edge) bool {
	if a.ValidTo == "" {
		return b.ValidTo != ""
	}
	if b.ValidTo == "" {
		return false
	}
	return mustInstant(a.ValidTo).After(mustInstant(b.ValidTo))
}

// checkChains refuses a set of stored edges in which two edges of one from,
// type and context are valid at one instant. Rule (ADR-0018 point 1): the
// store refuses such a state and does not repair it. Choosing which of two
// recorded facts is true is the caller's decision, as it is for a write, so
// Open names the pair and stops before any query can see it. The error wraps
// ErrEdgeOverlap and names both edge ids. Every stored interval is checked
// first (checkInterval, ErrEdgeInterval), so the scan below only sees
// non-empty intervals. Within a chain the edges are sorted by start and the
// scan keeps the edge that ends furthest so far: the next edge overlaps some
// earlier one exactly when it starts before that furthest end. It does not
// rely on neighbours alone, so no edge between two clashing ones can hide
// the clash. The chain with the smallest key is reported, so the message is
// the same for any map order.
func checkChains(edges map[string]Edge) error {
	ids := make([]string, 0, len(edges))
	for id := range edges {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		e := edges[id]
		if err := checkInterval(e.ValidFrom, e.ValidTo); err != nil {
			return fmt.Errorf("edge %s: %w", e.ID, err)
		}
	}
	type chainKey struct{ from, typ, context string }
	chains := map[chainKey][]Edge{}
	for _, e := range edges {
		k := chainKey{e.From, e.Type, e.Context}
		chains[k] = append(chains[k], e)
	}
	var first error
	var firstKey string
	for k, list := range chains {
		sort.Slice(list, func(i, j int) bool {
			fi, fj := mustInstant(list[i].ValidFrom), mustInstant(list[j].ValidFrom)
			if !fi.Equal(fj) {
				return fi.Before(fj)
			}
			return list[i].ID < list[j].ID
		})
		far := list[0]
		for i := 1; i < len(list); i++ {
			a, b := far, list[i]
			if endsAfter(b, far) {
				far = b
			}
			if !intervalsOverlap(a, b) {
				continue
			}
			key := k.from + "\x00" + k.typ + "\x00" + k.context
			if first == nil || key < firstKey {
				firstKey = key
				first = fmt.Errorf("%w: stored edges %s (%s [%s, %s)) and %s (%s [%s, %s)) of %s -[%s]->",
					ErrEdgeOverlap, a.ID, a.To, a.ValidFrom, openEnd(a.ValidTo),
					b.ID, b.To, b.ValidFrom, openEnd(b.ValidTo), k.from, k.typ)
			}
			break
		}
	}
	return first
}

// openEnd names an empty ValidTo for an error message.
func openEnd(validTo string) string {
	if validTo == "" {
		return "open"
	}
	return validTo
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
