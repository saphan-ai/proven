package world

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/saphan-ai/proven/internal/trace"
)

// Edge is one bitemporal relationship between two entities. ValidTo empty
// means the edge is open. A superseding edge closes the old one by setting
// ValidTo and SupersededBy; an edge is never deleted (ADR-0018 point 1,
// Meterless World Model specification).
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
func edgeID(from, typ, to, context, validFrom string) string {
	sum := sha256.Sum256([]byte(from + "\x00" + typ + "\x00" + to + "\x00" + context + "\x00" + validFrom))
	return "rel_" + hex.EncodeToString(sum[:])[:16]
}

// Relate writes one edge. An empty ValidFrom takes the provenance time. A new
// open edge with the same from, type and context but another target closes
// the previous open edge at the new edge's ValidFrom (it never deletes it).
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
	e.ID = edgeID(e.From, e.Type, e.To, e.Context, e.ValidFrom)
	if prev, ok := s.edges[e.ID]; ok {
		return prev, nil
	}
	snap := cloneEdges(s.edges)
	// Only an open edge is closed: one already closed or superseded keeps the
	// history it recorded.
	if e.ValidTo == "" {
		for id, o := range s.edges {
			if o.From == e.From && o.Type == e.Type && o.Context == e.Context &&
				o.To != e.To && o.ValidTo == "" && o.SupersededBy == "" {
				o.ValidTo = e.ValidFrom
				o.SupersededBy = e.ID
				s.edges[id] = o
			}
		}
	}
	s.edges[e.ID] = e
	actor := e.Source.By
	if actor == "" {
		actor = "system"
	}
	err := s.audit(trace.Event{
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

// Neighbors returns the entities joined to id by an edge, sorted by id. A
// non-empty via keeps one edge type. With at empty only open edges count
// ("now"); with at set, an edge counts when ValidFrom <= at < ValidTo, so a
// closed edge still answers for the time it was valid. An edge whose other
// end is not a stored entity is skipped.
func (s *Store) Neighbors(id, via, at string) []Entity {
	out := []Entity{}
	for _, r := range s.edges {
		if via != "" && r.Type != via {
			continue
		}
		if at != "" {
			if at < r.ValidFrom || (r.ValidTo != "" && at >= r.ValidTo) {
				continue
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
