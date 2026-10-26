package hmem

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Dreaming thresholds, from the Meterless H-MEM specification (section 7).
const (
	dreamClusterMinSize = 2
	dreamRelatedness    = 0.35
	maxDreamPerType     = 8
	// dreamConfidence is the confidence of a record an approved proposal creates.
	dreamConfidence = 0.75
)

// Proposal states.
const (
	ProposalProposed = "proposed"
	// ProposalApproving marks a domain approval that has begun: the person
	// decided, and the move of the records may be half done. It is written
	// before the first record moves, so a reopen can finish the decision.
	ProposalApproving = "approving"
	ProposalApproved  = "approved"
	ProposalRejected  = "rejected"
	// ProposalVoid marks a proposal whose source records were removed by Clear.
	// It is final: a proposal never outlives the records it names, so a later
	// record that reuses an id cannot be bound to it.
	ProposalVoid = "void"
)

// VoidReasonCleared is the reason a proposal reads with after a Clear ended it.
const VoidReasonCleared = "cleared"

// VoidReasonUnaudited is the reason written on a proposal whose dream_proposed
// ledger line could not be written, so no proposal exists without its line.
const VoidReasonUnaudited = "unaudited"

// Proposal is one thing dreaming suggests: an insight, a domain change or an
// invariant. It is a suggestion only; no record exists for it until a person
// approves it.
type Proposal struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	Content         string   `json:"content"`
	DerivedFrom     []string `json:"derivedFrom"`
	SuggestedDomain string   `json:"suggestedDomain,omitempty"`
	Status          string   `json:"status"`
	// PreviousDomains is part of a domain approval's durable intent: the domain
	// each source record had before the approval moved it, written with the
	// "approving" state and before the first record moves. A run that stops
	// after the move and before its ledger line still has it on reopen, so the
	// update line can carry the previous state, not only the new one.
	PreviousDomains map[string]string `json:"previousDomains,omitempty"`
	// Reason says why a proposal is void; it is empty in every other state.
	Reason string `json:"reason,omitempty"`
	// Epoch is the store's Clear count when the proposal was made. A proposal
	// of an older epoch names records a Clear removed (see Store.Clear).
	Epoch int `json:"epoch,omitempty"`
}

// dreamLog is the file proposals live in, beside the audit log. Each line is
// the whole proposal at one moment, so the newest line per id is its state and
// nothing is rewritten (ADR-0018 point 3: writes are append-only JSONL). No
// model is called: ADR-0018 point 2, and the specification's dreaming is
// deterministic clustering of what the store holds.
func (s *Store) dreamLog() string {
	return filepath.Join(filepath.Dir(s.auditPath), "dreams.jsonl")
}

// dreamIndex is the dream log read into the newest state of each proposal. Like
// the ledger index it is extended by the lines added since the last read, so a
// lookup never rereads the file (E16). Its memory grows with the number of
// proposals, which dreaming caps per call, not with the number of log lines.
//
// The rule (R16): no call of the store does work in proportion to the proposals
// ever made. A single proposal is found in byID; the proposals that still wait
// are kept apart in pending, by the epoch they were made in, so Pending, the
// sleep guards and Approve never walk the rejected, approved and void ones.
// order lists every id for the one caller that must see them all, the recovery
// at Open.
type dreamIndex struct {
	cur   logCursor
	order []string
	byID  map[string]Proposal
	// pending holds, per epoch, the ids whose newest state is proposed or
	// approving. A Clear raises the epoch, so the old epochs' sets are never
	// read again and cost no work.
	pending map[int]map[string]bool
	// visited counts the proposals the lookups examined. The growth tests read
	// it to prove the work does not follow the history.
	visited int64
}

func newDreamIndex(path string) *dreamIndex {
	ix := &dreamIndex{cur: logCursor{path: path, label: "dream log"}}
	ix.reset()
	return ix
}

func (ix *dreamIndex) reset() {
	ix.order = nil
	ix.byID = map[string]Proposal{}
	ix.pending = map[int]map[string]bool{}
}

// isWaiting reports whether a proposal in this status still needs a decision.
func isWaiting(status string) bool {
	return status == ProposalProposed || status == ProposalApproving
}

// note folds one dream log line into the index: byID takes the newest state and
// the pending sets follow the status change.
func (ix *dreamIndex) note(p Proposal) {
	old, seen := ix.byID[p.ID]
	if !seen {
		ix.order = append(ix.order, p.ID)
	} else if isWaiting(old.Status) {
		delete(ix.pending[old.Epoch], p.ID)
	}
	ix.byID[p.ID] = p
	if isWaiting(p.Status) {
		if ix.pending[p.Epoch] == nil {
			ix.pending[p.Epoch] = map[string]bool{}
		}
		ix.pending[p.Epoch][p.ID] = true
	}
}

// syncDreams brings the dream index up to the end of the dream log. The caller
// holds s.mu.
func (s *Store) syncDreams() error {
	ix := s.dreams
	return ix.cur.advance(ix.reset, func(line []byte, _ int64) error {
		var p Proposal
		if err := json.Unmarshal(line, &p); err != nil {
			return fmt.Errorf("hmem: dream log line is not valid JSON: %w", err)
		}
		ix.note(p)
		return nil
	})
}

// view is a proposal as the store sees it now: made before the last Clear and
// not finished, it reads as void with the reason "cleared". Finished is
// rejected, or approved with its dream_approved ledger line; an approved
// proposal without that line is an unfinished approval and names sources the
// Clear removed. The caller holds s.mu.
func (s *Store) view(p Proposal) (Proposal, error) {
	if p.Epoch >= s.epoch {
		return p, nil
	}
	switch p.Status {
	case ProposalProposed, ProposalApproving:
		p.Status, p.Reason = ProposalVoid, VoidReasonCleared
	case ProposalApproved:
		done, err := s.approvedLine(p.ID)
		if err != nil {
			return Proposal{}, err
		}
		if !done {
			p.Status, p.Reason = ProposalVoid, VoidReasonCleared
		}
	}
	return p, nil
}

// proposalState returns the newest state of one proposal in any status, from
// the by-id map: one lookup, whatever the history. The caller holds s.mu.
func (s *Store) proposalState(id string) (Proposal, bool, error) {
	if err := s.syncDreams(); err != nil {
		return Proposal{}, false, err
	}
	p, ok := s.dreams.byID[id]
	if !ok {
		return Proposal{}, false, nil
	}
	s.dreams.visited++
	p, err := s.view(p)
	return p, true, err
}

// waiting returns the proposals of the current epoch that still need a
// decision, in the order they were made. Only those are visited: the work
// follows the number waiting, not the number ever made. The caller holds s.mu.
func (s *Store) waiting() ([]Proposal, error) {
	if err := s.syncDreams(); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(s.dreams.pending[s.epoch]))
	for id := range s.dreams.pending[s.epoch] {
		ids = append(ids, id)
	}
	// Ids are dream-N, so the shorter id is the older one within a prefix.
	sort.Slice(ids, func(i, j int) bool {
		if len(ids[i]) != len(ids[j]) {
			return len(ids[i]) < len(ids[j])
		}
		return ids[i] < ids[j]
	})
	out := make([]Proposal, 0, len(ids))
	for _, id := range ids {
		s.dreams.visited++
		out = append(out, s.dreams.byID[id])
	}
	return out, nil
}

// proposals returns the newest state of every proposal ever made, in the order
// they were first proposed, as view shows them. It is the whole-history scan, so
// only the recovery at Open may call it (once, before the store is shared); a
// lookup on a live store uses proposalState or waiting, whose work does not
// follow the history. The caller holds s.mu.
func (s *Store) proposals() ([]Proposal, error) {
	if err := s.syncDreams(); err != nil {
		return nil, err
	}
	ix := s.dreams
	out := make([]Proposal, 0, len(ix.order))
	for _, id := range ix.order {
		ix.visited++
		p, err := s.view(ix.byID[id])
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// appendDream appends one proposal state to the dream log and syncs it.
func (s *Store) appendDream(ps ...Proposal) error {
	if s.writeHook != nil {
		if err := s.writeHook("dream"); err != nil {
			return err
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	for _, p := range ps {
		if err := enc.Encode(p); err != nil {
			return err
		}
	}
	if err := appendSynced(s.dreamLog(), buf.Bytes(), s.appendFault); err != nil {
		return fmt.Errorf("hmem: dream log: %w", err)
	}
	return nil
}

// relatedness is the multi-signal closeness of two records: content, tags,
// shared domain and shared entities.
func relatedness(a, b Record) float64 {
	same := 0.0
	if a.Domain == b.Domain {
		same = 1
	}
	return 0.5*max(0, Cosine(a.Embedding, b.Embedding)) + 0.2*Jaccard(a.Tags, b.Tags) + 0.1*same + 0.2*Jaccard(a.Entities, b.Entities)
}

// Dream looks at the store and returns new proposals: insights from clusters of
// related records, domain suggestions for uncategorized records, and
// invariants from preference and correction records. It creates no record and
// changes no record: only Approve does (the approval boundary, ADR-0018 point 2).
func (s *Store) Dream() ([]Proposal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// The count of every proposal ever made names the next id; it is a length,
	// not a scan.
	if err := s.syncDreams(); err != nil {
		return nil, err
	}
	seq := len(s.dreams.order)
	all := s.snapshot()
	created := []Proposal{}
	propose := func(p Proposal) {
		seq++
		p.ID = fmt.Sprintf("dream-%d", seq)
		p.Status = ProposalProposed
		p.Epoch = s.epoch
		created = append(created, p)
	}

	used := map[string]bool{}
	insights := 0
	for _, seed := range all {
		if used[seed.ID] || insights >= maxDreamPerType {
			continue
		}
		group := []Record{seed}
		for _, other := range all {
			if other.ID == seed.ID || used[other.ID] {
				continue
			}
			if relatedness(seed, other) >= dreamRelatedness {
				group = append(group, other)
			}
		}
		if len(group) < dreamClusterMinSize {
			continue
		}
		var contents, ids, domains []string
		seenDomain := map[string]bool{}
		for _, m := range group {
			used[m.ID] = true
			contents = append(contents, m.Content)
			ids = append(ids, m.ID)
			if !seenDomain[m.Domain] {
				seenDomain[m.Domain] = true
				domains = append(domains, m.Domain)
			}
		}
		propose(Proposal{
			Type:        "insight",
			Content:     fmt.Sprintf("Across %d related memories (%s): %s", len(group), strings.Join(domains, "/"), strings.Join(contents, " + ")),
			DerivedFrom: ids,
		})
		insights++
	}

	domainCount := 0
	for _, m := range all {
		if domainCount >= maxDreamPerType {
			break
		}
		if m.Domain != "general" {
			continue
		}
		if suggested := classifyDomain(m.Content, m.Tags, m.Type); suggested != "general" {
			propose(Proposal{
				Type:            "domain",
				Content:         fmt.Sprintf("Reclassify %q as %s", truncateRunes(m.Content, 60), suggested),
				DerivedFrom:     []string{m.ID},
				SuggestedDomain: suggested,
			})
			domainCount++
		}
	}

	byDomain := map[string][]Record{}
	var domainOrder []string
	for _, m := range all {
		if m.Type == TypePreference || m.Type == TypePersonal || containsTag(m.Tags, "event:user_correction") {
			if _, ok := byDomain[m.Domain]; !ok {
				domainOrder = append(domainOrder, m.Domain)
			}
			byDomain[m.Domain] = append(byDomain[m.Domain], m)
		}
	}
	sort.Strings(domainOrder)
	invariants := 0
	for _, domain := range domainOrder {
		list := byDomain[domain]
		if invariants >= maxDreamPerType {
			break
		}
		if len(list) < 2 {
			continue
		}
		var contents, ids []string
		for _, m := range list {
			contents = append(contents, m.Content)
			ids = append(ids, m.ID)
		}
		propose(Proposal{
			Type:        "invariant",
			Content:     fmt.Sprintf("Stable pattern in %s: %s", domain, strings.Join(contents, "; ")),
			DerivedFrom: ids,
		})
		invariants++
	}

	if err := s.appendDream(created...); err != nil {
		return nil, err
	}
	entries := make([]auditEntry, 0, len(created))
	for _, p := range created {
		entries = append(entries, auditEntry{action: "dream_proposed", id: p.ID, actor: "dreaming", extra: map[string]any{
			"details": map[string]any{"type": p.Type, "derivedFrom": p.DerivedFrom},
		}})
	}
	if err := s.auditBatch(entries); err != nil {
		// The lines are written together, so none was. The proposals are void
		// rather than pending: a proposal exists only with its ledger line.
		voided := make([]Proposal, 0, len(created))
		ids := make([]string, 0, len(created))
		for _, p := range created {
			ids = append(ids, p.ID)
			p.Status, p.Reason = ProposalVoid, VoidReasonUnaudited
			voided = append(voided, p)
		}
		if verr := s.appendDream(voided...); verr != nil {
			return nil, fmt.Errorf("%w; the proposals %s stay pending without a ledger line, because voiding them failed: %v", err, strings.Join(ids, ", "), verr)
		}
		return nil, err
	}
	return created, nil
}

// pendingPageMax is the most proposals one Pending call returns. A longer list
// is read page by page with PendingPage, so a call never returns an unbounded
// result (R16).
const pendingPageMax = 1000

// Pending returns the proposals nobody has approved or rejected, oldest first,
// at most pendingPageMax of them. The rest are read with PendingPage.
func (s *Store) Pending() ([]Proposal, error) {
	out, _, err := s.PendingPage(0, pendingPageMax)
	return out, err
}

// PendingPage returns up to limit pending proposals, oldest first, starting at
// the one numbered cursor (zero is the first). The second result is the cursor
// of the next page, or -1 when this page reached the end. A limit above
// pendingPageMax is cut to it; a limit below one is refused. The work is the
// proposals that wait, never the rejected, approved or void ones.
func (s *Store) PendingPage(cursor, limit int) ([]Proposal, int, error) {
	if limit < 1 || cursor < 0 {
		return nil, -1, fmt.Errorf("hmem: pending page needs a cursor of at least 0 and a limit of at least 1")
	}
	limit = min(limit, pendingPageMax)
	s.mu.Lock()
	defer s.mu.Unlock()
	w, err := s.waiting()
	if err != nil {
		return nil, -1, err
	}
	// An approving proposal is a decision already taken, so it is not offered.
	queue := w[:0]
	for _, p := range w {
		if p.Status == ProposalProposed {
			queue = append(queue, p)
		}
	}
	if cursor >= len(queue) {
		return []Proposal{}, -1, nil
	}
	end := min(cursor+limit, len(queue))
	out := append([]Proposal{}, queue[cursor:end]...)
	if end >= len(queue) {
		return out, -1, nil
	}
	return out, end, nil
}

// pendingOne finds a proposal that is still waiting, by its id, in one lookup.
// The caller holds s.mu.
func (s *Store) pendingOne(id string) (Proposal, error) {
	p, found, err := s.proposalState(id)
	if err != nil {
		return Proposal{}, err
	}
	if found && p.Status == ProposalProposed {
		return p, nil
	}
	return Proposal{}, fmt.Errorf("hmem: no pending proposal %s", id)
}

// Reject records a decision to drop a proposal. Nothing is created.
func (s *Store) Reject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, err := s.pendingOne(id)
	if err != nil {
		return err
	}
	pending := p
	p.Status = ProposalRejected
	if err := s.appendDream(p); err != nil {
		return err
	}
	if err := s.audit("dream_rejected", p.ID, "dreaming"); err != nil {
		// No ledger line, so no decision: the proposal is pending again.
		if rerr := s.appendDream(pending); rerr != nil {
			return fmt.Errorf("%w; proposal %s stays rejected without a ledger line, because putting it back failed: %v", err, p.ID, rerr)
		}
		return err
	}
	return nil
}

// truncateRunes cuts text to at most n runes.
func truncateRunes(text string, n int) string {
	r := []rune(text)
	if len(r) <= n {
		return text
	}
	return string(r[:n])
}
