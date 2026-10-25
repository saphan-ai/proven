package hmem

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
	ProposalApproved = "approved"
	ProposalRejected = "rejected"
)

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
}

// dreamLog is the file proposals live in, beside the audit log. Each line is
// the whole proposal at one moment, so the newest line per id is its state and
// nothing is rewritten (ADR-0018 point 3: writes are append-only JSONL). No
// model is called: ADR-0018 point 2, and the specification's dreaming is
// deterministic clustering of what the store holds.
func (s *Store) dreamLog() string {
	return filepath.Join(filepath.Dir(s.auditPath), "dreams.jsonl")
}

// proposals replays the dream log into the newest state of each proposal, in
// the order they were first proposed.
func (s *Store) proposals() ([]Proposal, error) {
	f, err := os.Open(s.dreamLog())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("hmem: dream log: %w", err)
	}
	defer f.Close()
	var order []string
	byID := map[string]Proposal{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var p Proposal
		if err := json.Unmarshal(sc.Bytes(), &p); err != nil {
			return nil, fmt.Errorf("hmem: dream log line is not valid JSON: %w", err)
		}
		if _, ok := byID[p.ID]; !ok {
			order = append(order, p.ID)
		}
		byID[p.ID] = p
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("hmem: dream log: %w", err)
	}
	out := make([]Proposal, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
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
	f, err := os.OpenFile(s.dreamLog(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("hmem: dream log: %w", err)
	}
	defer f.Close()
	if err := f.Chmod(0o600); err != nil {
		return fmt.Errorf("hmem: dream log: %w", err)
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("hmem: dream log: %w", err)
	}
	return f.Sync()
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
	for _, p := range created {
		if err := s.auditWith("dream_proposed", p.ID, "dreaming", map[string]any{
			"details": map[string]any{"type": p.Type, "derivedFrom": p.DerivedFrom},
		}); err != nil {
			return nil, err
		}
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

// Approve is the human's decision to keep a proposal. An insight becomes a
// long-term factual record and an invariant a long-term preference, each with
// derivedFrom lineage; a domain proposal moves the named records to the
// suggested domain and returns nil. An id that is not pending is refused.
//
// Approve is one recoverable transition: the record (or reclassification), the
// approved proposal state and the audit entry all happen, or none does. If the
// proposal-state write or the audit write fails after the record was created,
// the store returns to the memory and state file it had before the call, the
// proposal stays pending, and the error is returned, so a retry creates exactly
// one record (ADR-0018 point 3: no half-applied approval).
func (s *Store) Approve(id string) (*Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, err := s.pendingOne(id)
	if err != nil {
		return nil, err
	}
	// The snapshot is what a failed later step returns the store to.
	before := make(map[string]Record, len(s.records))
	for k, v := range s.records {
		before[k] = v
	}
	prevSeq := s.seq
	var created *Record
	if p.Type == "domain" {
		if err := s.reclassify(p); err != nil {
			return nil, s.rollbackApprove(before, prevSeq, p.ID, err)
		}
	} else {
		typ := TypePreference
		if p.Type == "insight" {
			typ = TypeFactual
		}
		conf := dreamConfidence
		rec, err := s.addLocked(AddInput{
			Content:     p.Content,
			Type:        typ,
			Layer:       LayerLong,
			Tags:        []string{"dream:" + p.Type},
			Confidence:  &conf,
			Source:      "dreaming",
			Origin:      "dream",
			Label:       "approved " + p.Type + " proposal",
			DerivedFrom: p.DerivedFrom,
		})
		if err != nil {
			return nil, err
		}
		created = &rec
	}
	p.Status = ProposalApproved
	if err := s.appendDream(p); err != nil {
		return nil, s.rollbackApprove(before, prevSeq, p.ID, err)
	}
	if err := s.audit("dream_approved", p.ID, "dreaming"); err != nil {
		// The approved state is already in the dream log. The proposal is
		// appended again as pending, so the log's last state wins and it is
		// pending once more.
		p.Status = ProposalProposed
		if rerr := s.appendDream(p); rerr != nil {
			err = errors.Join(err, rerr)
		}
		return nil, s.rollbackApprove(before, prevSeq, p.ID, err)
	}
	return created, nil
}

// rollbackApprove puts the records and the sequence number back to what they
// were before an approval, rewrites state.json to match, and notes the undo in
// the audit trail. The caller holds s.mu. It returns cause joined with any
// failure of the undo itself, so a failure that left the store changed is
// never reported as a clean one.
func (s *Store) rollbackApprove(before map[string]Record, seq int, proposal string, cause error) error {
	s.records = before
	s.seq = seq
	err := cause
	if perr := s.persist(); perr != nil {
		err = errors.Join(err, perr)
	}
	if aerr := s.audit("dream_approval_rolled_back", proposal, "dreaming"); aerr != nil {
		err = errors.Join(err, aerr)
	}
	return err
}

// Reject records a decision to drop a proposal. Nothing is created.
func (s *Store) Reject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, err := s.pendingOne(id)
	if err != nil {
		return err
	}
	p.Status = ProposalRejected
	if err := s.appendDream(p); err != nil {
		return err
	}
	return s.audit("dream_rejected", p.ID, "dreaming")
}

// reclassify moves the records of a domain proposal to its suggested domain.
// The caller holds s.mu. A failed audit or write puts the records back.
func (s *Store) reclassify(p Proposal) error {
	saved := map[string]Record{}
	for _, id := range p.DerivedFrom {
		rec, ok := s.records[id]
		if !ok || p.SuggestedDomain == "" {
			continue
		}
		saved[id] = rec
		rec.Domain = p.SuggestedDomain
		rec.Namespace = namespace(rec.Domain, rec.Type)
		s.records[id] = rec
	}
	for id, before := range saved {
		extra := map[string]any{"previousState": ledgerState(before), "newState": ledgerState(s.records[id])}
		if err := s.publishWith("update", id, "dreaming", extra, s.seq, saved, nil); err != nil {
			return err
		}
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
