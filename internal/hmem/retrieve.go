package hmem

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// The defaults and constants of the H-MEM hybrid ranking. They are part of the
// Meterless specification the store is built to (ADR-0018), not settings.
// recencyDays is the decay constant of the recency term.
const (
	dayMS            = 24 * 60 * 60 * 1000
	recencyDays      = 14
	defaultTopN      = 5
	defaultThreshold = 0.35
)

// Hit is one memory the ranking kept, with its final score.
type Hit struct {
	Record    Record
	Relevance float64
}

// QueryResult is the ranked list and the reinjection block. Context is the
// text a caller can give back to a model; Reason and Strategy say how the
// list was chosen.
type QueryResult struct {
	Hits     []Hit
	Context  string
	Reason   string
	Strategy string
}

// Query ranks memories with the hybrid formula and returns at most topN
// at or above the threshold. topN or threshold of zero selects the default.
// A negative threshold keeps every positive score (the zero value of the
// threshold option is the default 0.35; pass a negative number to disable it).
func (s *Store) Query(query string, topN int, threshold float64) (QueryResult, error) {
	if topN <= 0 {
		topN = defaultTopN
	}
	if threshold == 0 {
		threshold = defaultThreshold
	}
	if threshold < 0 {
		threshold = 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	qEmb := Embed(query)
	qTokens := Tokenize(query)
	// Only the first QueryTokens query tokens count (default eight, ADR-0013),
	// so a long query cannot drown the keyword and tag terms.
	if n := s.limits.QueryTokens; len(qTokens) > n {
		qTokens = qTokens[:n]
	}
	qEntities := extractEntities(query)
	qDomain := classifyDomain(query, nil, TypeGeneral)
	var scored []Hit
	for _, rec := range s.records {
		rel := score(rec, qEmb, qTokens, qEntities, qDomain, s.now())
		if rel >= threshold {
			scored = append(scored, Hit{Record: rec, Relevance: rel})
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Relevance != scored[j].Relevance {
			return scored[i].Relevance > scored[j].Relevance
		}
		// Equal scores fall back to the id, so the order is the same every run.
		return scored[i].Record.ID < scored[j].Record.ID
	})
	if len(scored) > topN {
		scored = scored[:topN]
	}
	// Retrieval is an access: each returned record is touched, audited and
	// saved, which feeds its recency on the next query.
	now := s.now()
	saved := make(map[string]Record, len(scored))
	for i := range scored {
		id := scored[i].Record.ID
		if _, ok := saved[id]; !ok {
			saved[id] = s.records[id]
		}
		rec := s.records[id]
		rec.LastAccessed = now
		rec.AccessCount++
		s.records[rec.ID] = rec
		scored[i].Record = rec
		if err := s.audit("read", rec.ID, "retrieval"); err != nil {
			// A partial touch must not stay: the state file was not written.
			for id, prev := range saved {
				s.records[id] = prev
			}
			return QueryResult{}, err
		}
	}
	if len(scored) > 0 {
		if err := s.persist(); err != nil {
			for id, prev := range saved {
				s.records[id] = prev
			}
			return QueryResult{}, err
		}
	}
	strategy := "minimal"
	if len(scored) > 1 {
		strategy = "comprehensive"
		for _, h := range scored {
			if h.Record.Type == TypePersonal || h.Record.Type == TypePreference {
				strategy = "personal"
				break
			}
		}
	}
	return QueryResult{
		Hits:     scored,
		Context:  formatReinjection(scored),
		Reason:   fmt.Sprintf("hybrid ranking for %q (domain: %s)", query, qDomain),
		Strategy: strategy,
	}, nil
}

// Score is the final ranking value for one record. It does not touch the store.
// It applies the same query trimming as Query, so a test can check one score by hand.
func Score(rec Record, query string, now int64) float64 {
	tokens := Tokenize(query)
	if len(tokens) > 8 {
		tokens = tokens[:8]
	}
	return score(rec, Embed(query), tokens, extractEntities(query), classifyDomain(query, nil, TypeGeneral), now)
}

// score is the hybrid formula: a weighted sum of semantic, keyword, tag,
// domain, entity, layer and recency terms, less a penalty for a superseded
// record, clamped to [0, 1] and scaled by the record's confidence.
func score(rec Record, qEmb []float64, qTokens, qEntities []string, qDomain string, now int64) float64 {
	semantic := clamp01(Cosine(qEmb, rec.Embedding))
	keyword := Jaccard(qTokens, Tokenize(rec.Content))
	tags := make([]string, len(rec.Tags))
	for i, t := range rec.Tags {
		tags[i] = strings.ToLower(t)
	}
	tag := Jaccard(qTokens, tags)
	domain := 0.0
	if rec.Domain == qDomain {
		domain = 1
	}
	entity := Jaccard(qEntities, rec.Entities)
	layer := layerWeight(rec.Layer)
	ageDays := float64(now-rec.LastAccessed) / dayMS
	recency := math.Exp(-ageDays / recencyDays)
	superseded := 0.0
	if rec.SupersededBy != "" {
		superseded = 1
	}
	raw := 0.35*semantic + 0.20*keyword + 0.10*tag + 0.10*domain +
		0.15*entity + 0.05*layer + 0.05*recency - 0.20*superseded
	return clamp01(raw) * rec.Confidence
}

// layerWeight ranks long-term memory above working, and working above short-term.
func layerWeight(layer string) float64 {
	switch layer {
	case LayerLong:
		return 1
	case LayerWork:
		return 0.8
	default:
		return 0.6
	}
}

// Feedback applies the fixed confidence deltas. helpful also counts one access.
func (s *Store) Feedback(id, kind string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[id]
	if !ok {
		return fmt.Errorf("hmem: unknown memory %s", id)
	}
	delta, ok := feedbackDelta(kind)
	if !ok {
		return fmt.Errorf("hmem: unknown feedback %q", kind)
	}
	prev := rec
	if kind == "helpful" {
		rec.AccessCount++
	}
	rec.Confidence = clamp01(rec.Confidence + delta)
	// A memory marked wrong is tagged for a person to look at, not deleted.
	// Copy the tags first so a rollback does not share the appended array.
	if kind == "wrong" && !containsTag(rec.Tags, "review") {
		rec.Tags = append(append([]string(nil), rec.Tags...), "review")
	}
	s.records[id] = rec
	// The ledger keeps the state before and after, so a history can show what
	// the feedback changed (the trust ledger, Meterless H-MEM section 9.1).
	states := map[string]any{"previousState": ledgerState(prev), "newState": ledgerState(rec)}
	if err := s.publishWith("feedback", id, "retrieval", states, s.seq, map[string]Record{id: prev}, nil); err != nil {
		return err
	}
	return nil
}

// feedbackDelta is the fixed confidence change for each feedback kind.
// An unknown kind is refused, not ignored.
func feedbackDelta(kind string) (float64, bool) {
	switch kind {
	case "helpful":
		return 0.05, true
	case "not_helpful":
		return -0.03, true
	case "wrong":
		return -0.20, true
	default:
		return 0, false
	}
}

// containsTag reports whether want is one of tags.
func containsTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

// formatReinjection writes the hits as a memory block grouped by domain, in
// sorted domain order, each line marked with its layer and ending in its id
// so a reader can trace a line back to its record.
func formatReinjection(hits []Hit) string {
	if len(hits) == 0 {
		return ""
	}
	by := map[string][]Hit{}
	var domains []string
	for _, h := range hits {
		d := h.Record.Domain
		if _, ok := by[d]; !ok {
			domains = append(domains, d)
		}
		by[d] = append(by[d], h)
	}
	sort.Strings(domains)
	var b strings.Builder
	b.WriteString("[MEMORY CONTEXT]\n")
	for _, d := range domains {
		fmt.Fprintf(&b, "## %s\n", d)
		for _, h := range by[d] {
			marker := "ST"
			switch h.Record.Layer {
			case LayerLong:
				marker = "LT"
			case LayerWork:
				marker = "WK"
			}
			ents := ""
			if len(h.Record.Entities) > 0 {
				ents = " (entities: " + strings.Join(h.Record.Entities, ", ") + ")"
			}
			fmt.Fprintf(&b, "- [%s] %s%s <%s>\n", marker, h.Record.Content, ents, h.Record.ID)
		}
	}
	b.WriteString("[/MEMORY CONTEXT]")
	return b.String()
}

// techTerms are the technology names that count as entities and as tech domain words.
var techTerms = []string{
	"react", "python", "typescript", "javascript", "rust", "sqlite", "postgres",
	"node", "docker", "graphql", "redis", "llm", "embedding",
}

// domainKeywords are the words that place a memory in a domain. The order
// matters: on a tie the earlier domain wins.
var domainKeywords = []struct {
	name  string
	words []string
}{
	{"tech", append([]string{"code", "bug", "api", "database", "deploy", "framework", "library", "server"}, techTerms...)},
	{"work", []string{"meeting", "deadline", "client", "project", "sprint", "review", "report"}},
	{"personal", []string{"family", "hobby", "prefer", "like", "dislike", "morning", "evening"}},
}

// classifyDomain picks a memory's domain. An explicit project: or tech: tag
// wins over the words, because a tag is a decision and the words are a guess.
func classifyDomain(content string, tags []string, typ string) string {
	for _, t := range tags {
		if strings.HasPrefix(t, "project:") {
			return strings.TrimPrefix(t, "project:")
		}
		if strings.HasPrefix(t, "tech:") {
			return "tech"
		}
	}
	tokens := map[string]struct{}{}
	for _, t := range Tokenize(content) {
		tokens[t] = struct{}{}
	}
	best := "general"
	bestScore := 0
	for _, d := range domainKeywords {
		score := 0
		for _, w := range d.words {
			if _, ok := tokens[w]; ok {
				score++
			}
		}
		if d.name == "personal" && (typ == TypePersonal || typ == TypePreference) {
			score++
		}
		if score > bestScore {
			best = d.name
			bestScore = score
		}
	}
	if bestScore == 0 {
		return "general"
	}
	return best
}

// extractEntities finds known technology names, then capitalised words of
// three or more letters, at most ten in all and each once.
func extractEntities(content string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(s string) {
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	toks := Tokenize(content)
	tech := map[string]struct{}{}
	for _, t := range techTerms {
		tech[t] = struct{}{}
	}
	for _, t := range toks {
		if _, ok := tech[t]; ok {
			add(t)
		}
	}
	// A capitalised word of three or more letters, as written.
	raw := content
	i := 0
	for i < len(raw) {
		for i < len(raw) && !isLetter(rune(raw[i])) {
			i++
		}
		j := i
		for j < len(raw) && isLetter(rune(raw[j])) {
			j++
		}
		if j-i >= 3 && raw[i] >= 'A' && raw[i] <= 'Z' {
			word := strings.ToLower(raw[i:j])
			if word != "the" && word != "this" && word != "that" && word != "with" {
				add(word)
			}
		}
		i = j
	}
	if len(out) > 10 {
		out = out[:10]
	}
	return out
}

// isLetter reports whether r is an ASCII letter.
func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// namespace files a memory under its domain and a bucket chosen by its type.
func namespace(domain, typ string) string {
	bucket := "notes"
	switch typ {
	case TypePreference, TypePersonal:
		bucket = "profile"
	case TypeFactual:
		bucket = "facts"
	}
	return domain + "/" + bucket
}
