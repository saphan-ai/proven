// Package writer fills the review draft from the judge's candidates.
//
// The Writer is the last agent of the review and calls no model. It adds no
// finding of its own: it keeps what the Judge kept, orders it by EPSS inside
// each severity (ADR-0004), and carries the failures, model uses and trace
// path into the draft.
package writer

import (
	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/epss"
)

// Input is everything the writer needs.
type Input struct {
	Candidates []candidate.Candidate
	Failed     []candidate.Failure
	Models     []candidate.ModelUse
	Partial    bool
	TracePath  string
}

// Write keeps the findings the judge did not drop and copies the rest of the draft.
func Write(in Input) candidate.Draft {
	d := candidate.Draft{
		Failed: in.Failed, Models: in.Models, Partial: in.Partial, TracePath: in.TracePath,
	}
	for _, c := range in.Candidates {
		if c.Dropped {
			continue
		}
		d.Findings = append(d.Findings, c)
	}
	// The likelihood of use in attacks is what a reader acts on first, so it orders the list.
	d.Findings = orderByEPSS(d.Findings)
	return d
}

// orderByEPSS orders each severity by exploitation likelihood.
// Severities stay in the order they first appear. Inside one severity the
// highest EPSS score comes first. Findings with no score keep their order
// and follow the scored ones. Equal scores keep their order.
func orderByEPSS(in []candidate.Candidate) []candidate.Candidate {
	if len(in) < 2 {
		return in
	}
	type group struct {
		items []candidate.Candidate
	}
	var seen []string
	groups := map[string]*group{}
	for _, c := range in {
		g := groups[c.Severity]
		if g == nil {
			g = &group{}
			groups[c.Severity] = g
			seen = append(seen, c.Severity)
		}
		g.items = append(g.items, c)
	}
	out := make([]candidate.Candidate, 0, len(in))
	for _, sev := range seen {
		items := groups[sev].items
		sortStable(items)
		out = append(out, items...)
	}
	return out
}

// sortStable sorts one severity's findings by EPSS, keeping equal ones in order.
func sortStable(items []candidate.Candidate) {
	// Insertion keeps equal elements in place. A score sorts before no
	// score. A higher score sorts before a lower one.
	for i := 1; i < len(items); i++ {
		j := i
		for j > 0 && before(items[j], items[j-1]) {
			items[j], items[j-1] = items[j-1], items[j]
			j--
		}
	}
}

// before reports whether a sorts ahead of b: a score beats no score, a higher score a lower one.
func before(a, b candidate.Candidate) bool {
	as, aok := epss.Best(a)
	bs, bok := epss.Best(b)
	if aok && bok {
		return as > bs
	}
	return aok && !bok
}
