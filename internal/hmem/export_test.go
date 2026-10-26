package hmem

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
