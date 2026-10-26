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
type dreamIndex struct {
	cur   logCursor
	order []string
	byID  map[string]Proposal
}

func newDreamIndex(path string) *dreamIndex {
	ix := &dreamIndex{cur: logCursor{path: path, label: "dream log"}}
	ix.reset()
	return ix
}

func (ix *dreamIndex) reset() {
	ix.order = nil
	ix.byID = map[string]Proposal{}
}

// proposals returns the newest state of each proposal, in the order they were
// first proposed, as the store sees them now: a proposal made before the last
// Clear and not finished reads as void with the reason "cleared". Finished is
// rejected, or approved with its dream_approved ledger line; an approved
// proposal without that line is an unfinished approval and names sources the
// Clear removed. The caller holds s.mu.
func (s *Store) proposals() ([]Proposal, error) {
	ix := s.dreams
	err := ix.cur.advance(ix.reset, func(line []byte, _ int64) error {
		var p Proposal
		if err := json.Unmarshal(line, &p); err != nil {
			return fmt.Errorf("hmem: dream log line is not valid JSON: %w", err)
		}
		if _, ok := ix.byID[p.ID]; !ok {
			ix.order = append(ix.order, p.ID)
		}
		ix.byID[p.ID] = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]Proposal, 0, len(ix.order))
	for _, id := range ix.order {
		p := ix.byID[id]
		if p.Epoch < s.epoch {
			switch p.Status {
			case ProposalProposed, ProposalApproving:
				p.Status, p.Reason = ProposalVoid, VoidReasonCleared
			case ProposalApproved:
				done, err := s.approvedLine(p.ID)
				if err != nil {
					return nil, err
				}
				if !done {
					p.Status, p.Reason = ProposalVoid, VoidReasonCleared
				}
			}
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
	if err := appendSynced(s.dreamLog(), buf.Bytes()); err != nil {
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
	known, err := s.proposals()
	if err != nil {
		return nil, err
	}
	seq := len(known)
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

// Pending returns the proposals nobody has approved or rejected.
func (s *Store) Pending() ([]Proposal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.proposals()
	if err != nil {
		return nil, err
	}
	out := []Proposal{}
	for _, p := range all {
		if p.Status == ProposalProposed {
			out = append(out, p)
		}
	}
	return out, nil
}

// pendingOne finds a proposal that is still waiting. The caller holds s.mu.
func (s *Store) pendingOne(id string) (Proposal, error) {
	all, err := s.proposals()
	if err != nil {
		return Proposal{}, err
	}
	for _, p := range all {
		if p.ID == id && p.Status == ProposalProposed {
			return p, nil
		}
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
