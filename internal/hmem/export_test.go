package hmem

import (
	"fmt"
	"os"
)

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
	if err := s.syncDreams(); err != nil {
		return err
	}
	base := len(s.dreams.order)
	ps := make([]Proposal, 0, n)
	for i := 0; i < n; i++ {
		ps = append(ps, Proposal{ID: fmt.Sprintf("dream-%d", base+i+1), Type: "insight", Content: "grown", DerivedFrom: []string{"mem-1"}, Status: ProposalProposed})
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

// Append stages the hook of FailAppends is asked about, after the step each
// names. They stand for the writes that place bytes and then fail.
const (
	StagePartial  = stagePartial
	StageWritten  = stageWritten
	StageSynced   = stageSynced
	StageClosed   = stageClosed
	StageRollback = stageRollback
)

// FailAppends installs a hook that runs after each step of every log append
// (path is the log, stage one of the Stage constants). A non-nil error fails
// the append at that step, after the step's bytes were placed.
func (s *Store) FailAppends(hook func(path, stage string) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.appendFault = hook
}

// CountDirListings makes the store count the directory listings the backup
// allocator makes, and returns the counter. A sleep must not list the backups
// directory once a sequence file exists.
func (s *Store) CountDirListings() func() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	s.readDir = func(dir string) ([]os.DirEntry, error) {
		n++
		return os.ReadDir(dir)
	}
	return func() int {
		s.mu.Lock()
		defer s.mu.Unlock()
		return n
	}
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
	p, found, _ := s.proposalState(id)
	if !found {
		return ""
	}
	return p.Status
}

// ProposalsVisited is how many proposals the lookups have examined since Open.
// The growth tests assert it follows the proposals that wait, not the history.
func ProposalsVisited(s *Store) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dreams.visited
}

// ScanAllProposals is the whole-history scan the lookups must not use. The
// growth tests call it as the negative control: it must make the counter move
// with the history, so a lookup that fell back to it would turn the tests red.
func ScanAllProposals(s *Store) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	ps, _ := s.proposals()
	return len(ps)
}

// ScanAllWaiting is the full read of the waiting proposals that a page must not
// make. The work-bound test calls it as the negative control: it moves the
// visited counter and the allocation with the number waiting.
func ScanAllWaiting(s *Store) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	ps, _ := s.waiting()
	return len(ps)
}

// AppendRejectedProposalLines writes n rejected proposals to the dream log,
// without touching state: a history of finished proposals.
func AppendRejectedProposalLines(s *Store, n int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.syncDreams(); err != nil {
		return err
	}
	base := len(s.dreams.order)
	ps := make([]Proposal, 0, n)
	for i := 0; i < n; i++ {
		ps = append(ps, Proposal{ID: fmt.Sprintf("dream-%d", base+i+1), Type: "insight", Content: "old", DerivedFrom: []string{"mem-1"}, Status: ProposalRejected})
	}
	return s.appendDream(ps...)
}
