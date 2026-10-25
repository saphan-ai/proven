package scout

// This file is the Sense stage (the spec's docs/scoring.md): the Stage-1 deterministic scan over the
// whole registry, for a text prompt or for a list of changed files, and the confidence bands.
// Stage 2 (a local embedding model) is not built: the spec makes it optional, and with it absent the
// Stage-1 result stands.

import (
	"math"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Candidate is one scored intent. Span is the part of the prompt it was scored on when the prompt was
// split into parts (multi-intent); it is empty when the whole prompt was scored.
type Candidate struct {
	Intent string  `json:"intent"`
	Score  float64 `json:"score"`
	Span   string  `json:"span,omitempty"`
}

// Signal is one piece of input that counted for an intent: a word or phrase of a prompt, or a changed
// file and the pattern it matched. It is what a reader is shown to see why a class was chosen.
type Signal struct {
	Intent string  `json:"intent"`
	Input  string  `json:"input"`
	Match  string  `json:"match"`
	Weight float64 `json:"weight"`
}

// sensed is what one Stage-1 scan returns.
type sensed struct {
	candidates []Candidate
	signals    []Signal
	// unexplained is the input no intent accounts for: words no intent knows, or files no pattern matches.
	unexplained []string
}

// The canonical band names (spec: confidence bands).
const (
	BandHigh   = "high"
	BandMedium = "medium"
	BandLow    = "low"
	BandReject = "reject"
)

// Bands are the lower edges of the High, Medium and Low bands. The spec makes them configurable and
// names these defaults: 0.80, 0.55, 0.40.
type Bands struct {
	High   float64 `json:"high"`
	Medium float64 `json:"medium"`
	Low    float64 `json:"low"`
}

// DefaultBands are the spec's canonical band edges.
func DefaultBands() Bands { return Bands{High: 0.80, Medium: 0.55, Low: 0.40} }

// Of names the band a confidence falls in.
func (b Bands) Of(score float64) string {
	switch {
	case score >= b.High:
		return BandHigh
	case score >= b.Medium:
		return BandMedium
	case score >= b.Low:
		return BandLow
	default:
		return BandReject
	}
}

// ---------------------------------------------------------------------------
// Text prompts
// ---------------------------------------------------------------------------

// stopwords carry no intent. They are dropped before scoring, so "the" in two intents' examples does
// not make them look alike.
var stopwords = map[string]bool{
	"a": true, "an": true, "the": true, "to": true, "of": true, "and": true, "or": true, "for": true,
	"in": true, "on": true, "at": true, "by": true, "with": true, "from": true, "about": true, "me": true,
	"my": true, "our": true, "i": true, "we": true, "you": true, "it": true, "is": true, "are": true,
	"be": true, "do": true, "does": true, "this": true, "that": true, "these": true, "those": true,
	"what": true, "how": true, "all": true, "up": true, "out": true, "into": true, "should": true,
	"can": true, "could": true, "would": true, "please": true, "also": true, "now": true, "there": true,
	"us": true, "your": true, "its": true, "as": true, "so": true, "if": true, "then": true, "any": true,
}

// wordRE picks the word-like runs of a prompt: letters, digits and inner hyphens or apostrophes.
var wordRE = regexp.MustCompile(`[\p{L}\p{N}][\p{L}\p{N}'’\-]*`)

// words splits text into content tokens: lower case, possessives and plurals folded, stopwords and bare
// numbers dropped. Numbers are parameters (a count, a value), not signs of an intent.
func words(text string) []string {
	var out []string
	for _, w := range wordRE.FindAllString(strings.ToLower(text), -1) {
		w = strings.TrimRight(w, "-'’")
		w = strings.TrimSuffix(strings.TrimSuffix(w, "'s"), "’s")
		w = strings.ReplaceAll(strings.ReplaceAll(w, "'", ""), "’", "")
		if w == "" || stopwords[w] || isNumber(w) {
			continue
		}
		out = append(out, stem(w))
	}
	return out
}

// isNumber reports whether w is a number, with an optional k or m suffix ("280k").
func isNumber(w string) bool {
	w = strings.TrimRight(w, "km")
	if w == "" {
		return false
	}
	for _, r := range w {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// stem folds the plural endings that separate a prompt from an example ("deals", "deal"). It is a
// deliberately small rule set: a stemmer that also folds verbs would merge intents the registry keeps apart.
func stem(w string) string {
	switch {
	case len(w) > 4 && strings.HasSuffix(w, "ies"):
		return w[:len(w)-3] + "y"
	case len(w) > 4 && strings.HasSuffix(w, "sses"):
		return w[:len(w)-2]
	case len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") && !strings.HasSuffix(w, "us"):
		return w[:len(w)-1]
	}
	return w
}

// textIndex is the Stage-1 trigger table of a text registry: for every token and token pair, the
// intents that carry it and with what weight.
type textIndex struct {
	reg   *Registry
	uni   map[string]map[int]float64
	bi    map[string]map[int]float64
	maxID float64
}

// The weight of each source of a token. Examples are the spec's classifier seeds, so they count most;
// the description is prose and counts half; the id's own words ("deal", "recover") count as examples do.
const (
	weightExample     = 1.0
	weightDescription = 0.5
	weightID          = 1.0
	// A pair of adjacent content words in an example is a stronger sign than either word alone.
	weightPair = 2.0
)

// newTextIndex builds the trigger table from every intent's examples, description and id.
func newTextIndex(reg *Registry) *textIndex {
	ix := &textIndex{reg: reg, uni: map[string]map[int]float64{}, bi: map[string]map[int]float64{}}
	add := func(m map[string]map[int]float64, key string, i int, w float64) {
		if m[key] == nil {
			m[key] = map[int]float64{}
		}
		if w > m[key][i] {
			m[key][i] = w
		}
	}
	for i, in := range reg.Intents {
		for _, ex := range in.Examples {
			ws := words(ex)
			for j, w := range ws {
				add(ix.uni, w, i, weightExample)
				if j > 0 {
					add(ix.bi, ws[j-1]+" "+w, i, weightExample)
				}
			}
		}
		for _, w := range words(in.Description) {
			add(ix.uni, w, i, weightDescription)
		}
		if in.ID != GeneralIntent {
			for _, w := range words(strings.NewReplacer(".", " ", "-", " ").Replace(in.ID)) {
				add(ix.uni, w, i, weightID)
			}
		}
	}
	ix.maxID = math.Log(1 + float64(len(reg.Intents)))
	return ix
}

// idf is the inverse document frequency of a token over the registry's intents: a word every intent
// uses tells the scan nothing, a word one intent uses tells it a lot.
func (ix *textIndex) idf(df int) float64 {
	return math.Log(1 + float64(len(ix.reg.Intents))/float64(df))
}

// scan scores one prompt. Every intent's lexical score is its weighted match sum (spec, Stage 1). The
// confidence of an intent is the square root of two fractions: its share of all matched weight (the
// spec's "winner's score as a fraction of matched weight") and the part of the prompt's own weight it
// explains. The second fraction keeps a prompt of mostly unknown words from scoring high on the one
// word it shares with an intent.
func (ix *textIndex) scan(prompt string) sensed {
	ws := words(prompt)
	n := len(ix.reg.Intents)
	score := make([]float64, n)
	explained := make([]float64, n)
	var total float64
	var out sensed
	seen := map[string]bool{}
	for j, w := range ws {
		if seen[w] {
			continue
		}
		seen[w] = true
		hits := ix.uni[w]
		if len(hits) == 0 {
			// A word no intent knows weighs as much as the rarest word, so it dilutes every intent equally.
			total += ix.maxID
			out.unexplained = append(out.unexplained, w)
			continue
		}
		idf := ix.idf(len(hits))
		total += idf
		for i, wt := range hits {
			score[i] += wt * idf
			explained[i] += idf
			out.signals = append(out.signals, Signal{Intent: ix.reg.Intents[i].ID, Input: w, Match: "word", Weight: wt * idf})
		}
		if j > 0 {
			pair := ws[j-1] + " " + w
			if ph := ix.bi[pair]; len(ph) > 0 {
				pidf := ix.idf(len(ph))
				for i, wt := range ph {
					score[i] += weightPair * wt * pidf
					out.signals = append(out.signals, Signal{Intent: ix.reg.Intents[i].ID, Input: pair, Match: "pair", Weight: weightPair * wt * pidf})
				}
			}
		}
	}
	// Phrase triggers match the lower-cased prompt as written, so a phrase can carry stopwords.
	low := strings.ToLower(prompt)
	for i, in := range ix.reg.Intents {
		for _, t := range in.Triggers {
			if t.Phrase != "" && strings.Contains(low, strings.ToLower(t.Phrase)) {
				score[i] += t.Weight
				out.signals = append(out.signals, Signal{Intent: in.ID, Input: t.Phrase, Match: "phrase", Weight: t.Weight})
			}
		}
	}
	var matched float64
	for _, s := range score {
		matched += s
	}
	for i, s := range score {
		if s <= 0 {
			continue
		}
		share := s / matched
		cover := 1.0
		if total > 0 {
			cover = math.Min(1, explained[i]/total)
		}
		out.candidates = append(out.candidates, Candidate{Intent: ix.reg.Intents[i].ID, Score: round3(math.Sqrt(share * cover))})
	}
	sortCandidates(out.candidates)
	sortSignals(out.signals)
	return out
}

// ---------------------------------------------------------------------------
// Changed files
// ---------------------------------------------------------------------------

// ChangedFile is one file of a change: its slash path and its status, A added, M modified, D deleted
// or R renamed. From is the old path of a renamed file.
type ChangedFile struct {
	Status string
	Path   string
	From   string
}

// ParseChange reads a change prompt: one file per line, "<status> <path>", and for a rename
// "R <old> -> <new>". A line with no status reads as modified. Blank lines are skipped.
func ParseChange(prompt string) []ChangedFile {
	var out []ChangedFile
	for _, line := range strings.Split(prompt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		f := ChangedFile{Status: "M", Path: line}
		if st, rest, ok := strings.Cut(line, " "); ok && isStatus(st) {
			f.Status = st[:1]
			f.Path = strings.TrimSpace(rest)
		}
		if f.Status == "R" {
			if from, to, ok := strings.Cut(f.Path, " -> "); ok {
				f.From, f.Path = strings.TrimSpace(from), strings.TrimSpace(to)
			}
		}
		out = append(out, f)
	}
	return out
}

// isStatus reports whether s is a git status letter, with git's optional rename score ("R087").
func isStatus(s string) bool {
	if s == "" || !strings.Contains("AMDR", s[:1]) {
		return false
	}
	for _, r := range s[1:] {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// FormatChange writes files as a change prompt, the inverse of ParseChange.
func FormatChange(files []ChangedFile) string {
	var b strings.Builder
	for _, f := range files {
		b.WriteString(f.Status)
		b.WriteByte(' ')
		if f.Status == "R" && f.From != "" {
			b.WriteString(f.From + " -> ")
		}
		b.WriteString(f.Path)
		b.WriteByte('\n')
	}
	return b.String()
}

// scanChange scores a change. Each file counts once, for the intent (or the tied intents) whose
// heaviest pattern it matches; a file no pattern matches is unexplained. An intent's confidence is its
// files' weight over the weight of every file, unexplained files counted at weight one, so a change
// is confidently one kind only when its files are that kind and no other.
func scanChange(reg *Registry, prompt string) sensed {
	var out sensed
	files := ParseChange(prompt)
	score := make([]float64, len(reg.Intents))
	var total float64
	for _, f := range files {
		best := 0.0
		var hits []int
		var pats []string
		for i, in := range reg.Intents {
			for _, t := range in.Triggers {
				if !t.matches(f) {
					continue
				}
				switch {
				case t.Weight > best:
					best, hits, pats = t.Weight, []int{i}, []string{t.Path}
				case t.Weight == best && !containsInt(hits, i):
					hits, pats = append(hits, i), append(pats, t.Path)
				}
			}
		}
		if len(hits) == 0 {
			total++
			out.unexplained = append(out.unexplained, f.Status+" "+f.Path)
			continue
		}
		total += best
		for k, i := range hits {
			score[i] += best
			out.signals = append(out.signals, Signal{Intent: reg.Intents[i].ID, Input: f.Status + " " + f.Path, Match: pats[k], Weight: best})
		}
		// A file shared by tied intents is weighed once per intent, so the tie lowers every share.
		total += best * float64(len(hits)-1)
	}
	for i, s := range score {
		if s <= 0 {
			continue
		}
		out.candidates = append(out.candidates, Candidate{Intent: reg.Intents[i].ID, Score: round3(s / total)})
	}
	sortCandidates(out.candidates)
	sortSignals(out.signals)
	return out
}

// matches reports whether a path trigger matches one changed file. A renamed file matches on its new
// path and on its old one, because moving a file out of a sensitive directory is a change to that directory.
func (t Trigger) matches(f ChangedFile) bool {
	if t.Path == "" {
		return false
	}
	if t.Status != "" && !strings.Contains(t.Status, f.Status) {
		return false
	}
	if globMatch(t.Path, f.Path) {
		return true
	}
	return f.From != "" && globMatch(t.Path, f.From)
}

// globMatch matches a pattern against a slash path, segment by segment, with ** for any number of
// directories. A pattern with no slash also matches the base name, so go.mod matches a nested go.mod.
// These are the semantics of Proven's own path rules (ADR-0014), so a pattern reads the same in both.
func globMatch(pattern, file string) bool {
	if matchSegs(strings.Split(pattern, "/"), strings.Split(file, "/")) {
		return true
	}
	if !strings.Contains(pattern, "/") {
		return matchSegs([]string{pattern}, []string{path.Base(file)})
	}
	return false
}

// matchSegs matches pattern segments against path segments; a bad segment fails closed.
func matchSegs(pat, name []string) bool {
	if len(pat) == 0 {
		return len(name) == 0
	}
	if pat[0] == "**" {
		rest := pat
		for len(rest) > 0 && rest[0] == "**" {
			rest = rest[1:]
		}
		if matchSegs(rest, name) {
			return true
		}
		return len(name) > 0 && matchSegs(pat, name[1:])
	}
	if len(name) == 0 {
		return false
	}
	ok, err := path.Match(pat[0], name[0])
	if err != nil || !ok {
		return false
	}
	return matchSegs(pat[1:], name[1:])
}

// validGlob checks every segment of a pattern on load.
func validGlob(pattern string) error {
	for _, seg := range strings.Split(pattern, "/") {
		if seg == "" || seg == "**" {
			continue
		}
		if _, err := path.Match(seg, "x"); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Shared
// ---------------------------------------------------------------------------

// sortCandidates orders candidates by score, highest first, and by id among equals, so the order is the
// same on every run.
func sortCandidates(c []Candidate) {
	sort.SliceStable(c, func(i, j int) bool {
		if c[i].Score != c[j].Score {
			return c[i].Score > c[j].Score
		}
		return c[i].Intent < c[j].Intent
	})
}

// sortSignals orders signals by intent, then by weight, heaviest first.
func sortSignals(s []Signal) {
	sort.SliceStable(s, func(i, j int) bool {
		if s[i].Intent != s[j].Intent {
			return s[i].Intent < s[j].Intent
		}
		if s[i].Weight != s[j].Weight {
			return s[i].Weight > s[j].Weight
		}
		return s[i].Input < s[j].Input
	})
}

// round3 rounds to three decimals, so a printed score is the score that was compared.
func round3(x float64) float64 { return math.Round(x*1000) / 1000 }

func containsInt(xs []int, x int) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
