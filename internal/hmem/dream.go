package hmem

import (
	"bufio"
	"bytes"
	"encoding/json"
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
	// ProposalApproving marks a domain approval that has begun: the person
	// decided, and the move of the records may be half done. It is written
	// before the first record moves, so a reopen can finish the decision.
	ProposalApproving = "approving"
	ProposalApproved  = "approved"
	ProposalRejected  = "rejected"
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
//
// When a domain proposal can arise: only for a stored record whose domain is
// "general" while classifyDomain now returns something else. Add stores the
// result of that same pure function on the same content, tags and type, and no
// public call changes those three afterwards, so a store filled through Add,
// Query, Feedback, SleepExecute and Approve never yields one. It arises when
// the stored domain is stale: state written by a build with different
// classifier rules, or restored from such a backup. That is by design: the
// proposal is the way a person reviews a classifier change, not a second
// opinion on the same rules. The classifier's rules are not changed here.
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

// truncateRunes cuts text to at most n runes.
func truncateRunes(text string, n int) string {
	r := []rune(text)
	if len(r) <= n {
		return text
	}
	return string(r[:n])
}
