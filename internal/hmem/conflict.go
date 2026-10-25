package hmem

import (
	"fmt"
	"math"
	"regexp"
)

// conflictAutoResolveGate is the decision confidence at or above which a
// conflict is resolved without a person (Meterless H-MEM section 10.2). Below
// it the conflict is queued.
const conflictAutoResolveGate = 0.7

// opposingPairs are the phrase pairs that make two similar memories disagree.
var opposingPairs = [][2]*regexp.Regexp{
	{regexp.MustCompile(`(?i)\balways\b`), regexp.MustCompile(`(?i)\bnever\b`)},
	{regexp.MustCompile(`(?i)\benable[ds]?\b`), regexp.MustCompile(`(?i)\bdisable[ds]?\b`)},
	{regexp.MustCompile(`(?i)\bprefer[s]?\b`), regexp.MustCompile(`(?i)\bavoid[s]?|dislike[s]?\b`)},
	{regexp.MustCompile(`(?i)\bon\b`), regexp.MustCompile(`(?i)\boff\b`)},
	{regexp.MustCompile(`(?i)\byes\b`), regexp.MustCompile(`(?i)\bno\b`)},
}

var numberToken = regexp.MustCompile(`\d+(\.\d+)?`)

// Conflict is two memories that disagree. Resolved and Resolution are set only
// on the copy AutoResolve returns.
type Conflict struct {
	ID         string  `json:"id"`
	MemoryA    string  `json:"memoryA"`
	MemoryB    string  `json:"memoryB"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
	Resolved   bool    `json:"resolved"`
	Resolution string  `json:"resolution,omitempty"`
}

// Outcome of AutoResolve.
const (
	ConflictResolved = "resolved"
	ConflictQueued   = "queued"
)

// ScanConflicts compares every pair of records and returns the conflicts found,
// writing one conflict_detected ledger line for each. A pair that is already
// a replacement (one supersedes the other) is not a conflict. It changes no record.
func (s *Store) ScanConflicts() ([]Conflict, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := s.snapshot()
	found := []Conflict{}
	for i := range all {
		for j := i + 1; j < len(all); j++ {
			c, ok := checkConflict(all[i], all[j])
			if !ok {
				continue
			}
			if err := s.auditWith("conflict_detected", c.MemoryA, "conflicts", map[string]any{
				"details": map[string]any{"with": c.MemoryB, "reason": c.Reason, "confidence": c.Confidence},
			}); err != nil {
				return nil, err
			}
			found = append(found, c)
		}
	}
	return found, nil
}

// checkConflict decides whether a and b disagree, and how sure that is.
func checkConflict(a, b Record) (Conflict, bool) {
	// A supersedes link is an intended replacement, so it is not a conflict.
	if a.Supersedes == b.ID || b.Supersedes == a.ID || a.SupersededBy == b.ID || b.SupersededBy == a.ID {
		return Conflict{}, false
	}
	sim := Cosine(a.Embedding, b.Embedding)
	reason := ""
	conf := 0.0
	for _, p := range opposingPairs {
		if (p[0].MatchString(a.Content) && p[1].MatchString(b.Content)) || (p[1].MatchString(a.Content) && p[0].MatchString(b.Content)) {
			if sim > 0.5 {
				reason = "opposing phrase pair on similar content"
				conf = 0.5 + 0.3*sim
			}
		}
	}
	if reason == "" && sim > 0.8 {
		na := numberToken.FindAllString(a.Content, -1)
		nb := numberToken.FindAllString(b.Content, -1)
		if len(na) > 0 && len(nb) > 0 && fmt.Sprint(na) != fmt.Sprint(nb) {
			reason = "high-similarity conflicting numeric values"
			conf = 0.55 + 0.25*sim
		}
	}
	if reason == "" {
		return Conflict{}, false
	}
	// Confidence modifiers (section 10.1).
	for _, e := range a.Entities {
		if containsTag(b.Entities, e) {
			conf += 0.1
			break
		}
	}
	if a.Domain == b.Domain {
		conf += 0.05
	}
	if containsTag(a.RelatedTo, b.ID) {
		// An explicit relation lowers false positives.
		conf -= 0.1
	}
	return Conflict{
		ID:         "conflict:" + a.ID + ":" + b.ID,
		MemoryA:    a.ID,
		MemoryB:    b.ID,
		Reason:     reason,
		Confidence: clamp01(conf),
	}, true
}

// AutoResolve decides a conflict by weighted heuristics (section 10.2). At a
// decision confidence of 0.70 or more the loser is marked supersededBy the
// winner, the winner is marked supersedes the loser and both records stay, because ranking penalizes a superseded record
// and a deletion could not be taken back. Below the gate it returns "queued"
// and changes nothing. It returns the conflict as resolved or as given.
func (s *Store) AutoResolve(c Conflict) (string, Conflict, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, okA := s.records[c.MemoryA]
	b, okB := s.records[c.MemoryB]
	if !okA || !okB {
		return "", c, fmt.Errorf("hmem: conflict %s names a memory that is not in the store", c.ID)
	}
	now := s.now()
	scoreOf := func(m Record) float64 {
		corr := 0.5
		if m.Provenance.Origin == "user_correction" {
			corr = 1
		}
		return 0.3*(float64(m.LastAccessed)/math.Max(1, float64(now))) +
			0.2*math.Min(1, float64(m.AccessCount)/10) +
			0.25*m.Confidence +
			0.1*math.Min(1, float64(len([]rune(m.Content)))/200) +
			0.15*corr
	}
	sa, sb := scoreOf(a), scoreOf(b)
	margin := math.Abs(sa-sb) / math.Max(sa, sb)
	decision := clamp01(c.Confidence*0.6 + margin*0.8)
	if decision < conflictAutoResolveGate {
		return ConflictQueued, c, nil
	}
	winner, loser, resolution := a, b, "keep_a"
	if sa < sb {
		winner, loser, resolution = b, a, "keep_b"
	}
	before := ledgerState(loser)
	prev := loser
	// Both links, as MarkSuperseded sets them: a later sleep then protects
	// the winner and the loser, so the conflict history stays readable.
	prevWinner := winner
	loser.SupersededBy = winner.ID
	winner.Supersedes = loser.ID
	s.records[loser.ID] = loser
	s.records[winner.ID] = winner
	extra := map[string]any{
		"previousState": before,
		"newState":      ledgerState(loser),
		"details": map[string]any{
			"conflictId": c.ID, "winner": winner.ID,
			"decisionConfidence": math.Round(decision*1000) / 1000,
		},
	}
	if err := s.publishWith("conflict_resolved", loser.ID, "conflicts", extra, s.seq, map[string]Record{loser.ID: prev, winner.ID: prevWinner}, nil); err != nil {
		return "", c, err
	}
	c.Resolved = true
	c.Resolution = resolution
	return ConflictResolved, c, nil
}
