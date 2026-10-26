package hmem

import (
	"fmt"
)

// approvedRecordID is the id of the long-term record an approved proposal
// creates. It is derived from the proposal's id, not from the sequence, so the
// record can always be found again and an approval can never create it twice.
func approvedRecordID(proposal string) string { return "mem-from-" + proposal }

// ledgerHas reports whether the audit log holds a line of this action for this
// memory id. When proposal is not empty, the line must also name that proposal
// in its details. The caller holds s.mu.
//
// It answers from the ledger index, which reads only the lines added since the
// last lookup (E16). Only the actions in indexedActions can be asked about; any
// other action is a programming error and is refused by name.
func (s *Store) ledgerHas(action, id, proposal string) (bool, error) {
	if !indexedActions[action] {
		return false, fmt.Errorf("hmem: ledger action %q is not indexed for lookup", action)
	}
	if err := s.syncLedger(); err != nil {
		return false, err
	}
	return s.ledger.has[hasKey(action, id, proposal)], nil
}

// approvedLine reports whether the proposal's "dream_approved" ledger line
// exists. It reads the same index as ledgerHas. The caller holds s.mu.
func (s *Store) approvedLine(proposal string) (bool, error) {
	if err := s.syncLedger(); err != nil {
		return false, err
	}
	return s.ledger.approved[proposal], nil
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
	if p.Status == ProposalVoid {
		// Clear removed the sources; approving now would bind the old content
		// to whatever records reuse those ids.
		return nil, fmt.Errorf("hmem: proposal %s is void (%s): the records it names were removed, so it cannot be approved", id, p.Reason)
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
		// A source that is absent means no record can move. Refusing here,
		// before the approving intent is written, keeps a proposal that did
		// nothing from ever reaching "approved" or a dream_approved line.
		// The check holds in every state, "approved" too: an approval whose
		// final audit write failed must not record itself over a source that
		// is gone, or no record would have moved.
		if err := s.requireLiveSources(p); err != nil {
			return nil, err
		}
		// ADR-0018 point 3: the intent is durable before the first record
		// moves, so an interruption after the move can be told from a
		// proposal nobody has approved.
		if p.Status == ProposalProposed {
			p.Status = ProposalApproving
			// The domains the records have now go into the same line, so the
			// state before the approval survives an interruption.
			p.PreviousDomains = map[string]string{}
			for _, id := range p.DerivedFrom {
				if rec := s.records[id]; rec.Domain != p.SuggestedDomain {
					p.PreviousDomains[id] = rec.Domain
				}
			}
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
		if err := s.requireLiveSources(p); err != nil {
			return Record{}, err
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

// requireLiveSources refuses a proposal that names a source which is not a
// live record, naming the first one. The caller holds s.mu.
func (s *Store) requireLiveSources(p Proposal) error {
	for _, d := range p.DerivedFrom {
		if _, live := s.records[d]; !live {
			return fmt.Errorf("hmem: proposal %s names source %s, which is not a live record; it cannot be approved", p.ID, d)
		}
	}
	return nil
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
		if p.Status == ProposalRejected || p.Status == ProposalVoid {
			continue
		}
		if p.Type == "domain" {
			if p.Status == ProposalProposed {
				continue // nobody decided
			}
			// A source that is gone cannot be moved. Like an insight whose
			// record is missing, the proposal is left for an explicit Approve,
			// which refuses it by name; Open must not fail over it.
			if s.requireLiveSources(p) != nil {
				continue
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
// does only what is missing. The previous domain of each record comes from the
// approval's durable intent (Proposal.PreviousDomains), so a line written after
// an interruption, when the record already shows the new domain, still carries
// both states. Only a proposal written before the intent held the domains has
// no previous state to give, and its line says only the new one.
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
		m := move{id: id, before: ledgerState(rec), known: rec.Domain != p.SuggestedDomain}
		if prev, ok := p.PreviousDomains[id]; ok && !m.known {
			// A run that stopped after the move: the record already shows the new
			// domain, and the approval's durable intent still has the old one.
			m.before.Domain, m.known = prev, true
		}
		moves = append(moves, m)
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
