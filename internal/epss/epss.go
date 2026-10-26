// Package epss attaches one EPSS score to each kept finding that names a CVE.
// The query is deterministic. It does not ask a model. The seam takes the
// epss-by-cve request, and the store builds the form (ADR-0032, ADR-0046, ADR-0047).
//
// It runs after the Judge, on the findings it kept, and before the Writer,
// which orders findings by their best score (see Best). FIRST EPSS is one of
// the first evidence sources (ADR-0004), read through the evidence store
// (ADR-0023, ADR-0047). A score is query evidence like any other: the form
// that ran and its row go into the review (ADR-0002).
package epss

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/envdoc"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/source"
)

const (
	// EnvView is the environment variable that names the EPSS view.
	// Sample mode does not read it.
	EnvView = "PROVEN_EPSS_VIEW"
	// DefaultView is the view a review queries when EnvView is empty.
	DefaultView = querytpl.DefaultEPSSView
)

// validCVE reports whether s is a CVE id and nothing else. A cell that merely
// contains one is not a match, so a title or a sentence cannot name a score.
func validCVE(s string) bool { return querytpl.ValidCVE(s) }

// ConfigError names a refused EPSS setting and the values that are allowed
// (ADR-0013). A quote or a control character in the value is escaped, so it
// cannot break or forge the refusal line.
type ConfigError = service.ConfigError

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
// SELECT. ADR-0032: the one rule lives in querytpl, which builds the query.
func ValidView(name string) bool {
	return querytpl.ValidRelation(name)
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
func Attach(ctx context.Context, q query.Runner, view, snapshot string, cands []candidate.Candidate) ([]candidate.Candidate, []Note, error) {
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
			// ADR-0032, ADR-0046, ADR-0047: the seam takes the epss-by-cve
			// request. The store builds the form it speaks.
			res, err := q.Run(ctx, querytpl.Request{
				Template: querytpl.EPSSByCVE,
				Params:   map[string]string{querytpl.ParamCVEID: id},
			}, querytpl.Relations{EPSSView: view})
			if err != nil {
				var refusal *querytpl.Refusal
				if errors.As(err, &refusal) && refusal.Reason == querytpl.ReasonUnknownTemplate {
					return cands, notes, fmt.Errorf("epss: template %s is not registered", querytpl.EPSSByCVE)
				}
				return cands, notes, err
			}
			// No row is a trace note, never evidence: a missing score says nothing about
			// the CVE, so it neither keeps nor drops the finding. The adapter's
			// rule is that decision (ADR-0046).
			adapter, ok := source.ByID("epss")
			if !ok {
				return cands, notes, fmt.Errorf("epss: adapter is not registered")
			}
			if _, ruled := adapter.Rule(source.Evidence{Rows: res.Rows}); !ruled {
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
			SQL: res.SQL, Form: res.Form, Engine: res.Engine, Snapshot: snap,
			Columns: append([]string(nil), res.Columns...),
			// Nil becomes an empty list, so the evidence prints [] and not null.
			Rows: service.ListRows(res.Rows),
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
			if validCVE(id) {
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
		if validCVE(id) {
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
