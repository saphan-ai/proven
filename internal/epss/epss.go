// Package epss attaches one EPSS score to each kept finding that names a CVE.
// The query is deterministic. It does not ask a model.
//
// It runs after the Judge, on the findings it kept, and before the Writer,
// which orders findings by their best score (see Best). FIRST EPSS is one of
// the first evidence sources (ADR-0004), read through a Zetaris view
// (ADR-0023). A score is query evidence like any other: the SQL and its row
// go into the review (ADR-0002).
package epss

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
)

const (
	// EnvView is the environment variable that names the EPSS view.
	// Sample mode does not read it.
	EnvView = "PROVEN_EPSS_VIEW"
	// DefaultView is the view a review queries when EnvView is empty.
	DefaultView = "px_evidence.px_epss_cve_rows"
)

// cvePattern is a CVE id and nothing else. A cell that merely contains one
// is not a match, so a title or a sentence cannot name a score.
var cvePattern = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)

// ConfigError names a refused EPSS setting and the values that are allowed.
type ConfigError struct {
	Field   string
	Value   string
	Allowed []string
	Reason  string
}

// Error names the setting, the refused value, why, and what is allowed
// (ADR-0013).
func (e *ConfigError) Error() string {
	if e == nil {
		return "configuration refused"
	}
	msg := "configuration refused: " + e.Field + "=" + quote(e.Value)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if len(e.Allowed) > 0 {
		msg += "; allowed: " + strings.Join(e.Allowed, ", ")
	}
	return msg
}

// quote wraps the refused value in double quotes. It does not escape it.
func quote(s string) string {
	// %q escapes a quote or a control character, so a refusal prints the value as it is.
	return fmt.Sprintf("%q", s)
}

// LoadView reads EnvView. An empty value is DefaultView.
// ADR-0013: a wrong name is refused before the review starts, never replaced
// by the default.
func LoadView(env func(string) string) (string, error) {
	if env == nil {
		env = func(string) string { return "" }
	}
	raw := strings.TrimSpace(env(EnvView))
	if raw == "" {
		return DefaultView, nil
	}
	if !ValidView(raw) {
		return "", &ConfigError{
			Field: EnvView, Value: raw, Reason: "not a view name",
			Allowed: []string{"a view or schema.view of letters, digits and underscores; empty means " + DefaultView},
		}
	}
	return raw, nil
}

// ValidView reports whether name can be placed in a FROM clause.
// One identifier, or schema.view. Nothing else.
// The view name is the one part of the score query that is neither a fixed
// word nor a checked CVE id, so this check is what keeps that query a single
// SELECT.
func ValidView(name string) bool {
	parts := strings.Split(name, ".")
	if len(parts) < 1 || len(parts) > 2 {
		return false
	}
	for _, p := range parts {
		if !ident(p) {
			return false
		}
	}
	return true
}

// ident reports whether s is a plain identifier: a letter or an underscore,
// then letters, digits or underscores.
func ident(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c == '_':
		case i > 0 && c >= '0' && c <= '9':
		default:
			return false
		}
	}
	return true
}

// Environment lists the EPSS setting for the configuration page.
func Environment() []envdoc.Setting {
	return []envdoc.Setting{
		{
			Name:        EnvView,
			Default:     envdoc.Tick(DefaultView),
			Description: "the EPSS view a finding's CVE is scored from. Sample mode uses the default and does not read this variable",
		},
	}
}

// Note is one trace line for a CVE that was looked up.
// HasEvidence is false when the view returned no row. Text then contains
// "no EPSS row".
type Note struct {
	FindingID   string
	Text        string
	HasEvidence bool
	Evidence    candidate.Evidence
}

// Attach queries the EPSS view once for each CVE that a kept finding's query
// evidence names and that the finding has no score for yet. A row is
// appended as query evidence. No row appends nothing and returns a note
// whose text says "no EPSS row". A dropped finding is left as it is.
// view empty means DefaultView.
func Attach(ctx context.Context, q query.Adapter, view, snapshot string, cands []candidate.Candidate) ([]candidate.Candidate, []Note, error) {
	if view == "" {
		view = DefaultView
	}
	if !ValidView(view) {
		return cands, nil, fmt.Errorf("epss: view name %q is not allowed", view)
	}
	var notes []Note
	for i := range cands {
		if cands[i].Dropped {
			continue
		}
		if err := ctx.Err(); err != nil {
			return cands, notes, err
		}
		ids := cvesToScore(cands[i])
		if len(ids) == 0 {
			continue
		}
		// The engine is needed only when a CVE needs a score, so a review with no
		// CVE does not fail for want of one.
		if q == nil {
			return cands, notes, &service.Down{Service: "query", Reason: "no query engine"}
		}
		for _, id := range ids {
			if err := ctx.Err(); err != nil {
				return cands, notes, err
			}
			sqlText, err := scoreSQL(view, id)
			if err != nil {
				return cands, notes, err
			}
			res, err := q.Query(ctx, sqlText)
			if err != nil {
				return cands, notes, err
			}
			// No row is a trace note, never evidence: a missing score says nothing about
			// the CVE, so it neither keeps nor drops the finding.
			if len(res.Rows) == 0 {
				notes = append(notes, Note{FindingID: cands[i].ID, Text: "no EPSS row for " + id})
				continue
			}
			ev := toEvidence(res, snapshot)
			cands[i].Evidence = append(cands[i].Evidence, ev)
			text := strings.Join(Phrases(ev.Query.Columns, ev.Query.Rows), "; ")
			if text == "" {
				text = "EPSS row for " + id
			}
			notes = append(notes, Note{FindingID: cands[i].ID, Text: text, HasEvidence: true, Evidence: ev})
		}
	}
	return cands, notes, nil
}

// scoreSQL builds the score query. The view and the id are checked again
// here because they are pasted into the SQL text: an Adapter takes plain
// SQL, with no bound parameters.
func scoreSQL(view, id string) (string, error) {
	if !ValidView(view) {
		return "", fmt.Errorf("epss: view name %q is not allowed", view)
	}
	if !cvePattern.MatchString(id) {
		return "", fmt.Errorf("epss: %q is not a CVE id", id)
	}
	return "SELECT cve_id, epss, percentile, score_date FROM " + view + " WHERE cve_id = '" + id + "'", nil
}

// toEvidence turns a score answer into query evidence. The engine's snapshot
// wins; the review's snapshot fills in for an engine that has none, such as
// the local sample engine.
func toEvidence(res query.Result, snapshot string) candidate.Evidence {
	snap := res.Snapshot
	if snap == "" {
		snap = snapshot
	}
	return candidate.Evidence{
		Kind: candidate.KindQuery,
		Query: &candidate.QueryEvidence{
			SQL: res.SQL, Engine: res.Engine, Snapshot: snap,
			Columns: append([]string(nil), res.Columns...),
			Rows:    copyRows(res.Rows),
		},
	}
}

// Best is the highest EPSS score on the finding.
// The second result is false when no attached row has a numeric score.
func Best(c candidate.Candidate) (float64, bool) {
	var best float64
	found := false
	for _, e := range c.Evidence {
		if e.Query == nil || !isScore(e.Query.Columns) {
			continue
		}
		ix, ok := colIndex(e.Query.Columns, "epss")
		if !ok {
			continue
		}
		for _, row := range e.Query.Rows {
			f, err := strconv.ParseFloat(strings.TrimSpace(cellAt(row, ix)), 64)
			if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
				continue
			}
			if !found || f > best {
				best = f
				found = true
			}
		}
	}
	return best, found
}

// Phrases names each score row. One line is
// "EPSS 0.97 (99th percentile, 2026-10-20)".
// A value of percentile from 0 to 1 is a fraction of the catalog.
func Phrases(columns []string, rows [][]string) []string {
	if !isScore(columns) {
		return nil
	}
	si, _ := colIndex(columns, "epss")
	pi, _ := colIndex(columns, "percentile")
	di, _ := colIndex(columns, "score_date")
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, Phrase(cellAt(row, si), cellAt(row, pi), cellAt(row, di)))
	}
	return out
}

// Phrase formats one score, one percentile and one date.
func Phrase(score, percentile, date string) string {
	score = strings.TrimSpace(score)
	date = strings.TrimSpace(date)
	if n, ok := percentRank(percentile); ok {
		return fmt.Sprintf("EPSS %s (%s percentile, %s)", score, ordinal(n), date)
	}
	return fmt.Sprintf("EPSS %s (percentile %s, %s)", score, strings.TrimSpace(percentile), date)
}

// percentRank reads a percentile as a whole percent. EPSS publishes it as a
// fraction from 0 to 1; a value above 1 is taken as a percent already.
func percentRank(raw string) (int, bool) {
	f, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) || f < 0 {
		return 0, false
	}
	if f <= 1 {
		f *= 100
	}
	return int(math.Round(f)), true
}

// ordinal writes 1st, 2nd, 3rd, 4th, and so on. 11th, 12th and 13th are the
// English exceptions.
func ordinal(n int) string {
	s := strconv.Itoa(n)
	if n%100 >= 11 && n%100 <= 13 {
		return s + "th"
	}
	switch n % 10 {
	case 1:
		return s + "st"
	case 2:
		return s + "nd"
	case 3:
		return s + "rd"
	default:
		return s + "th"
	}
}

// cvesToScore lists the CVE ids in a finding's query evidence that have no
// score yet, each once, in the order found. Score rows themselves are
// skipped, so a score never asks for another score.
func cvesToScore(c candidate.Candidate) []string {
	have := alreadyScored(c)
	seen := map[string]struct{}{}
	var out []string
	for _, e := range c.Evidence {
		if e.Query == nil || isScore(e.Query.Columns) {
			continue
		}
		for _, row := range e.Query.Rows {
			for _, cell := range row {
				for _, id := range cveIDs(cell) {
					if _, ok := have[id]; ok {
						continue
					}
					if _, ok := seen[id]; ok {
						continue
					}
					seen[id] = struct{}{}
					out = append(out, id)
				}
			}
		}
	}
	return out
}

// alreadyScored collects the CVE ids that already have a score row on the
// finding, so a second pass does not query them again.
func alreadyScored(c candidate.Candidate) map[string]struct{} {
	out := map[string]struct{}{}
	for _, e := range c.Evidence {
		if e.Query == nil || !isScore(e.Query.Columns) {
			continue
		}
		ix, ok := colIndex(e.Query.Columns, "cve_id")
		if !ok {
			continue
		}
		for _, row := range e.Query.Rows {
			id := strings.ToUpper(strings.TrimSpace(cellAt(row, ix)))
			if cvePattern.MatchString(id) {
				out[id] = struct{}{}
			}
		}
	}
	return out
}

// cveIDs reads CVE ids from a cell that is one id or a list separated by
// spaces, commas or semicolons, such as an OSV alias list. An id inside other
// text yields nothing.
func cveIDs(cell string) []string {
	var out []string
	for _, part := range strings.FieldsFunc(cell, func(r rune) bool {
		switch r {
		case ' ', '\t', '\n', '\r', ',', ';':
			return true
		default:
			return false
		}
	}) {
		id := strings.ToUpper(strings.TrimSpace(part))
		if cvePattern.MatchString(id) {
			out = append(out, id)
		}
	}
	return out
}

// isScore reports whether a result has all four EPSS columns. That is how a
// score is told apart from other query evidence.
func isScore(columns []string) bool {
	for _, name := range []string{"cve_id", "epss", "percentile", "score_date"} {
		if _, ok := colIndex(columns, name); !ok {
			return false
		}
	}
	return true
}

// colIndex finds a column by name, ignoring case and surrounding spaces.
func colIndex(columns []string, name string) (int, bool) {
	for i, c := range columns {
		if strings.EqualFold(strings.TrimSpace(c), name) {
			return i, true
		}
	}
	return 0, false
}

// cellAt reads one cell, or an empty string when the row is short.
func cellAt(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}

// copyRows copies the rows, so the evidence shares no memory with the
// adapter's answer. Nil becomes empty, so the evidence prints [] and not
// null.
func copyRows(rows [][]string) [][]string {
	if rows == nil {
		return [][]string{}
	}
	out := make([][]string, len(rows))
	for i, row := range rows {
		out[i] = append([]string(nil), row...)
	}
	return out
}
