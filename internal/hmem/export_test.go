package hmem

import "os"

// FailWrites installs a hook that runs before every durable write of the
// store (op is "dream", "state" or "audit:<action>"). A non-nil error refuses
// that write. It lets the tests prove a failed write leaves no half-applied
// change.
func (s *Store) FailWrites(hook func(op string) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writeHook = hook
}

// StateFile is what the state read needs from an open file. A test supplies one
// whose size is wrong or unknown and counts the bytes read from it.
type StateFile interface {
	Read(p []byte) (int, error)
	Close() error
	Stat() (os.FileInfo, error)
}

// OpenWithStateFile opens a store like Open, but reads state.json through the
// file the test supplies. It is the only way a test sees how many bytes the
// state read consumes.
func OpenWithStateFile(opts Options, f StateFile) (*Store, error) {
	return open(opts, func(string) (stateFile, error) { return f, nil })
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
