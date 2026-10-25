package hmem

// FailWrites installs a hook that runs before every durable write of the
// store (op is "dream" or "audit:<action>"). A non-nil error refuses that
// write. It lets the tests prove a failed write leaves no half-applied change.
func (s *Store) FailWrites(hook func(op string) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writeHook = hook
}
