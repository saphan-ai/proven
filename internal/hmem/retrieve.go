package hmem

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

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

// QueryResult is the ranked list and the reinjection block.
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
	if len(qTokens) > 8 {
		qTokens = qTokens[:8]
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
		return scored[i].Record.ID < scored[j].Record.ID
	})
	if len(scored) > topN {
		scored = scored[:topN]
	}
	now := s.now()
	for i := range scored {
		rec := s.records[scored[i].Record.ID]
		rec.LastAccessed = now
		rec.AccessCount++
		s.records[rec.ID] = rec
		scored[i].Record = rec
		if err := s.audit("read", rec.ID, "retrieval"); err != nil {
			return QueryResult{}, err
		}
	}
	if len(scored) > 0 {
		if err := s.persist(); err != nil {
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
func Score(rec Record, query string, now int64) float64 {
	tokens := Tokenize(query)
	if len(tokens) > 8 {
		tokens = tokens[:8]
	}
	return score(rec, Embed(query), tokens, extractEntities(query), classifyDomain(query, nil, TypeGeneral), now)
}

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
	if kind == "helpful" {
		rec.AccessCount++
	}
	rec.Confidence = clamp01(rec.Confidence + delta)
	if kind == "wrong" && !containsTag(rec.Tags, "review") {
		rec.Tags = append(rec.Tags, "review")
	}
	s.records[id] = rec
	if err := s.audit("feedback", id, "retrieval"); err != nil {
		return err
	}
	return s.persist()
}

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

func containsTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

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

var techTerms = []string{
	"react", "python", "typescript", "javascript", "rust", "sqlite", "postgres",
	"node", "docker", "graphql", "redis", "llm", "embedding",
}

var domainKeywords = []struct {
	name  string
	words []string
}{
	{"tech", append([]string{"code", "bug", "api", "database", "deploy", "framework", "library", "server"}, techTerms...)},
	{"work", []string{"meeting", "deadline", "client", "project", "sprint", "review", "report"}},
	{"personal", []string{"family", "hobby", "prefer", "like", "dislike", "morning", "evening"}},
}

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

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

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
