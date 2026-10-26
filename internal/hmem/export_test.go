package hmem

import "fmt"

// BytesRead is how many bytes the ledger and dream-log readers have read from
// disk since Open. The growth tests assert it stays near the log size however
// many lookups run.
func BytesRead(s *Store) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ledger.cur.read + s.dreams.cur.read
}

// AppendLedgerLines writes n synthetic "read" lines to the audit log, spread
// over ids mem-0 to mem-(spread-1), in batches, without touching state. It
// stands for a log grown by long use.
func AppendLedgerLines(s *Store, n, spread int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for done := 0; done < n; {
		batch := make([]auditEntry, 0, 1000)
		for ; len(batch) < 1000 && done < n; done++ {
			batch = append(batch, auditEntry{action: "read", id: fmt.Sprintf("mem-%d", done%spread), actor: "growth"})
		}
		if err := s.auditBatch(batch); err != nil {
			return err
		}
	}
	return nil
}

// AppendProposalLines writes n pending proposals to the dream log, without
// touching state.
func AppendProposalLines(s *Store, n int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	ps := make([]Proposal, 0, n)
	for i := 0; i < n; i++ {
		ps = append(ps, Proposal{ID: fmt.Sprintf("dream-grown-%d", i), Type: "insight", Content: "grown", DerivedFrom: []string{"mem-1"}, Status: ProposalProposed})
	}
	return s.appendDream(ps...)
}

// FailWrites installs a hook that runs before every durable write of the
// store (op is "dream", "state" or "audit:<action>"). A non-nil error refuses
// that write. It lets the tests prove a failed write leaves no half-applied
// change.
func (s *Store) FailWrites(hook func(op string) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writeHook = hook
}

// LedgerHas reports whether the audit log holds a line of this action for this
// memory id, so a test can prove a line was or was not written.
func LedgerHas(s *Store, action, id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ledgerHas(action, id, "")
}

// ProposalStatus returns the newest status of one proposal, or "" when the
// store has no such proposal. It lets a test read the dream log's last state.
func ProposalStatus(s *Store, id string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ps, _ := s.proposals()
	for _, p := range ps {
		if p.ID == id {
			return p.Status
		}
	}
	return ""
}
