package hmem

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// approvedRecordID is the id of the long-term record an approved proposal
// creates. It is derived from the proposal's id, not from the sequence, so the
// record can always be found again and an approval can never create it twice.
func approvedRecordID(proposal string) string { return "mem-from-" + proposal }

// proposalState returns the newest state of one proposal in any status. The
// caller holds s.mu.
func (s *Store) proposalState(id string) (Proposal, bool, error) {
	all, err := s.proposals()
	if err != nil {
		return Proposal{}, false, err
	}
	for _, p := range all {
		if p.ID == id {
			return p, true, nil
		}
	}
	return Proposal{}, false, nil
}

// ledgerHas reports whether the audit log holds a line of this action for this
// memory id. When proposal is not empty, the line must also name that proposal
// in its details. The caller holds s.mu.
func (s *Store) ledgerHas(action, id, proposal string) (bool, error) {
	f, err := os.Open(s.auditPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("hmem: ledger: %w", err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var e LedgerEntry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			return false, fmt.Errorf("hmem: ledger line is not valid JSON: %w", err)
		}
		if e.MemoryID != id || e.Action != action {
			continue
		}
		if proposal == "" || e.Details["proposal"] == proposal {
			return true, nil
		}
	}
	if err := sc.Err(); err != nil {
		return false, fmt.Errorf("hmem: ledger: %w", err)
	}
	return false, nil
}

// Approve is the human's decision to keep a proposal. An insight becomes a
// long-term factual record and an invariant a long-term preference, each with
// derivedFrom lineage; a domain proposal moves the named records to the
// suggested domain and returns nil. A rejected or unknown id is refused.
//
// Approve is idempotent and recovers forward (ADR-0018 point 3: no
// half-applied approval, and no compensating write that could itself fail).
// It does four steps in this order, and each step checks whether it is already
// done: the record (id from approvedRecordID, written to state.json), the
// ledger "create" line, the approved proposal state, the "dream_approved"
// ledger line. A failure at any step returns its error and leaves the earlier
// steps in place; calling Approve again, or reopening the store, does only the
// steps still missing. So the ledger never names a record that does not exist,
// and a retry ends with exactly one record and an approved proposal.
//
// A proposal marked approved whose "dream_approved" ledger line is missing is
// an unfinished approval, and Approve completes it, creating the record if an
// older log lost it. A proposal that is approved and has that line is finished:
// Approve refuses it, so an approval that was cleared away later is not undone.
func (s *Store) Approve(id string) (*Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, found, err := s.proposalState(id)
	if err != nil {
		return nil, err
	}
	if !found || p.Status == ProposalRejected {
		return nil, fmt.Errorf("hmem: no pending proposal %s", id)
	}
	if p.Status == ProposalApproved {
		done, err := s.ledgerHas("dream_approved", p.ID, "")
		if err != nil {
			return nil, err
		}
		if done {
			return nil, fmt.Errorf("hmem: proposal %s is already approved", id)
		}
	}
	return s.completeApproval(p)
}

// completeApproval does the steps of an approval that are not yet done. The
// caller holds s.mu. It is the only place an approval is carried out, for
// Approve and for the recovery at Open alike.
func (s *Store) completeApproval(p Proposal) (*Record, error) {
	var created *Record
	if p.Type == "domain" {
		// ADR-0018 point 3: the intent is durable before the first record
		// moves, so an interruption after the move can be told from a
		// proposal nobody has approved.
		if p.Status == ProposalProposed {
			p.Status = ProposalApproving
			if err := s.appendDream(p); err != nil {
				return nil, err
			}
		}
		if err := s.reclassify(p); err != nil {
			return nil, err
		}
	} else {
		rec, err := s.ensureApprovedRecord(p)
		if err != nil {
			return nil, err
		}
		created = &rec
	}
	if p.Status != ProposalApproved {
		p.Status = ProposalApproved
		if err := s.appendDream(p); err != nil {
			return nil, err
		}
	}
	done, err := s.ledgerHas("dream_approved", p.ID, "")
	if err != nil {
		return nil, err
	}
	if !done {
		if err := s.audit("dream_approved", p.ID, "dreaming"); err != nil {
			return nil, err
		}
	}
	return created, nil
}

// ensureApprovedRecord makes sure the record of an insight or invariant exists
// in memory, in state.json and in the ledger, doing only what is missing. The
// state file is written before the ledger line, so the ledger can never say a
// record was created that is not there. The caller holds s.mu.
func (s *Store) ensureApprovedRecord(p Proposal) (Record, error) {
	rid := approvedRecordID(p.ID)
	rec, ok := s.records[rid]
	if !ok {
		// The rule: approval never writes a lineage reference to a record that
		// is absent or archived. A source that is gone (cleared, or lost from an
		// older log) makes approval refuse by name; it never restores a record,
		// because a record that is absent may have been removed on purpose.
		// Sleep keeps the sources of a pending proposal, so this only meets
		// records removed some other way.
		for _, d := range p.DerivedFrom {
			if _, live := s.records[d]; !live {
				return Record{}, fmt.Errorf("hmem: proposal %s names source %s, which is not a live record; it cannot be approved", p.ID, d)
			}
		}
		typ := TypePreference
		if p.Type == "insight" {
			typ = TypeFactual
		}
		conf := dreamConfidence
		var err error
		rec, err = s.addLocked(AddInput{
			Content:     p.Content,
			Type:        typ,
			Layer:       LayerLong,
			Tags:        []string{"dream:" + p.Type},
			Confidence:  &conf,
			Source:      "dreaming",
			Origin:      "dream",
			Label:       "approved " + p.Type + " proposal",
			DerivedFrom: p.DerivedFrom,
			fixedID:     rid,
		})
		if err != nil {
			return Record{}, err
		}
	}
	done, err := s.ledgerHas("create", rid, "")
	if err != nil {
		return Record{}, err
	}
	if !done {
		extra := map[string]any{"details": map[string]any{"proposal": p.ID}}
		if err := s.auditWith("create", rid, "dreaming", extra); err != nil {
			return Record{}, err
		}
	}
	return rec, nil
}

// recoverApprovals finishes the approvals an earlier run left half done: a
// proposal whose record exists but whose approved state or ledger line is
// missing. It needs no person. A proposal whose record is missing is left for
// an explicit Approve: a record that is absent may have been cleared on
// purpose, and recovery must not bring it back. A domain proposal is
// finished only when its state is "approving" or "approved": the "approving"
// state is written before any record moves, so it proves the person decided
// (a reclassified record alone cannot be told from one that was always in the
// new domain). The rule is the same for every type: finish what was decided,
// never decide. A pending domain proposal is left alone. Open calls it before
// the store is shared.
func (s *Store) recoverApprovals() error {
	all, err := s.proposals()
	if err != nil {
		return err
	}
	for _, p := range all {
		if p.Status == ProposalRejected {
			continue
		}
		if p.Type == "domain" {
			if p.Status == ProposalProposed {
				continue // nobody decided
			}
		} else if _, ok := s.records[approvedRecordID(p.ID)]; !ok {
			continue
		}
		if p.Status == ProposalApproved {
			done, err := s.ledgerHas("dream_approved", p.ID, "")
			if err != nil {
				return err
			}
			if done {
				continue
			}
		}
		if _, err := s.completeApproval(p); err != nil {
			return fmt.Errorf("hmem: finishing the approval of %s: %w", p.ID, err)
		}
	}
	return nil
}

// reclassify moves the records of a domain proposal to its suggested domain.
// The caller holds s.mu. It is idempotent: a record already in the domain is
// not moved again. The state file is written first, then one "update" ledger
// line per record that has none for this proposal, so a retry after a failure
// does only what is missing. A line written for a record that was already in
// the new domain carries no previous state, because that state is gone.
func (s *Store) reclassify(p Proposal) error {
	type move struct {
		id     string
		before LedgerState
		known  bool
	}
	var moves []move
	saved := map[string]Record{}
	for _, id := range p.DerivedFrom {
		rec, ok := s.records[id]
		if !ok || p.SuggestedDomain == "" {
			continue
		}
		moves = append(moves, move{id: id, before: ledgerState(rec), known: rec.Domain != p.SuggestedDomain})
		if rec.Domain == p.SuggestedDomain {
			continue
		}
		saved[id] = rec
		rec.Domain = p.SuggestedDomain
		rec.Namespace = namespace(rec.Domain, rec.Type)
		s.records[id] = rec
	}
	if len(saved) > 0 {
		if err := s.persist(); err != nil {
			// Nothing durable was written: only memory is put back.
			for id, rec := range saved {
				s.records[id] = rec
			}
			return err
		}
	}
	for _, m := range moves {
		done, err := s.ledgerHas("update", m.id, p.ID)
		if err != nil {
			return err
		}
		if done {
			continue
		}
		extra := map[string]any{"newState": ledgerState(s.records[m.id]), "details": map[string]any{"proposal": p.ID}}
		if m.known {
			extra["previousState"] = m.before
		}
		if err := s.auditWith("update", m.id, "dreaming", extra); err != nil {
			return err
		}
	}
	return nil
}
