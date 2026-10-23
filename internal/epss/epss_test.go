package epss

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/query"
)

// TestPhraseUsesTheStoredScoreAndAnOrdinalPercentile proves the phrase keeps
// the score as stored, writes the percentile as an English ordinal, and
// falls back to the raw text when the percentile is not a number.
func TestPhraseUsesTheStoredScoreAndAnOrdinalPercentile(t *testing.T) {
	if got := Phrase("0.97", "0.99", "2026-10-20"); got != "EPSS 0.97 (99th percentile, 2026-10-20)" {
		t.Fatalf("%s", got)
	}
	if got := Phrase("0.50000", "0.90000", "2026-10-20"); got != "EPSS 0.50000 (90th percentile, 2026-10-20)" {
		t.Fatalf("%s", got)
	}
	if got := Phrase("0.99999", "1.0", "2026-10-20"); got != "EPSS 0.99999 (100th percentile, 2026-10-20)" {
		t.Fatalf("%s", got)
	}
	for _, tc := range []struct{ raw, want string }{
		{"0.01", "1st"}, {"0.02", "2nd"}, {"0.03", "3rd"}, {"0.04", "4th"},
		{"0.11", "11th"}, {"0.12", "12th"}, {"0.13", "13th"},
		{"0.21", "21st"}, {"0.22", "22nd"}, {"0.23", "23rd"},
	} {
		got := Phrase("0.1", tc.raw, "2026-10-20")
		if !strings.Contains(got, tc.want+" percentile") {
			t.Fatalf("%s → %s", tc.raw, got)
		}
	}
	if got := Phrase("0.5", "not-a-number", "2026-10-20"); got != "EPSS 0.5 (percentile not-a-number, 2026-10-20)" {
		t.Fatalf("%s", got)
	}
}

// TestLoadViewEmptyIsTheDefaultAndABadNameIsRefused proves an empty or blank
// value takes the default, and a name that is not a view is refused with
// what is allowed.
func TestLoadViewEmptyIsTheDefaultAndABadNameIsRefused(t *testing.T) {
	got, err := LoadView(nil)
	if err != nil || got != DefaultView {
		t.Fatalf("%s %v", got, err)
	}
	got, err = LoadView(func(string) string { return "  " })
	if err != nil || got != DefaultView {
		t.Fatalf("blank %s %v", got, err)
	}
	got, err = LoadView(func(string) string { return DefaultView })
	if err != nil || got != DefaultView {
		t.Fatalf("set %s %v", got, err)
	}
	// Mutated field: PROVEN_EPSS_VIEW.
	_, err = LoadView(func(k string) string {
		if k == EnvView {
			return "px_evidence.bad;drop"
		}
		return ""
	})
	var cfg *ConfigError
	if err == nil || !asConfig(err, &cfg) || cfg.Field != EnvView || cfg.Value != "px_evidence.bad;drop" {
		t.Fatal(err)
	}
	if !strings.Contains(err.Error(), "allowed") {
		t.Fatal(err)
	}
}

// asConfig reports whether err is a *ConfigError and stores it in dst.
func asConfig(err error, dst **ConfigError) bool {
	c, ok := err.(*ConfigError)
	if ok {
		*dst = c
	}
	return ok
}

// TestAttachReadsTheSampleRowAndSkipsAMissingCVE proves, on the sample
// tables, that a CVE with a row gains the score as query evidence with the
// exact SQL, a CVE with no row gains only a trace note, and a finding with
// no CVE gains nothing.
func TestAttachReadsTheSampleRowAndSkipsAMissingCVE(t *testing.T) {
	q, err := query.OpenSample(filepath.Join(repoRoot(t), "testdata", "fixture", "discount", "tables.sql"))
	if err != nil {
		t.Fatal(err)
	}
	in := []candidate.Candidate{
		kev("F4", "CVE-2026-0001"),
		kev("F9", "CVE-2026-0002"),
		kev("F8", "CVE-2021-44228"),
		{Risk: candidate.Risk{ID: "F1", Severity: "ERROR"}},
	}
	out, notes, err := Attach(context.Background(), q, "", "sample-2026-10-22", in)
	if err != nil {
		t.Fatal(err)
	}
	f4 := out[0]
	if len(f4.Evidence) != 2 || f4.Evidence[1].Query == nil {
		t.Fatalf("F4 %#v", f4.Evidence)
	}
	row := f4.Evidence[1].Query.Rows
	if len(row) != 1 || row[0][0] != "CVE-2026-0001" || row[0][1] != "0.50000" || row[0][2] != "0.90000" || row[0][3] != "2026-10-20" {
		t.Fatalf("row %#v", row)
	}
	if f4.Evidence[1].Query.SQL != "SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-0001'" {
		t.Fatalf("sql %s", f4.Evidence[1].Query.SQL)
	}
	if f4.Evidence[1].Query.Snapshot != "sample-2026-10-22" || f4.Evidence[1].Query.Engine != "local-sample" {
		t.Fatalf("query %#v", f4.Evidence[1].Query)
	}
	f9 := out[1]
	if len(f9.Evidence) != 1 {
		t.Fatalf("missing CVE gained evidence %#v", f9.Evidence)
	}
	f8 := out[2]
	if len(f8.Evidence) != 2 || f8.Evidence[1].Query.Rows[0][1] != "0.99999" {
		t.Fatalf("log4j %#v", f8.Evidence)
	}
	if len(out[3].Evidence) != 0 {
		t.Fatalf("F1 %#v", out[3].Evidence)
	}
	var missing bool
	for _, n := range notes {
		if n.FindingID == "F9" && n.Text == "no EPSS row for CVE-2026-0002" && !n.HasEvidence {
			missing = true
		}
		if n.FindingID == "F4" && !strings.Contains(n.Text, "EPSS 0.50000 (90th percentile, 2026-10-20)") {
			t.Fatalf("F4 note %q", n.Text)
		}
	}
	if !missing {
		t.Fatalf("notes %#v", notes)
	}
}

// TestAttachReadsAnAliasListOnceEach proves a cell that lists several CVE
// ids is split, and each id is queried once.
func TestAttachReadsAnAliasListOnceEach(t *testing.T) {
	q := &memQuery{rows: map[string][]string{
		"CVE-2026-0001":  {"CVE-2026-0001", "0.50000", "0.90000", "2026-10-20"},
		"CVE-2021-44228": {"CVE-2021-44228", "0.99999", "1.0", "2026-10-20"},
	}}
	c := candidate.Candidate{
		Risk: candidate.Risk{ID: "osv:leftpad"},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindQuery,
			Query: &candidate.QueryEvidence{
				Columns: []string{"aliases"},
				Rows:    [][]string{{"CVE-2026-0001, CVE-2021-44228"}},
			},
		}},
	}
	out, notes, err := Attach(context.Background(), q, "", "snap", []candidate.Candidate{c})
	if err != nil {
		t.Fatal(err)
	}
	if len(q.calls) != 2 || len(out[0].Evidence) != 3 || len(notes) != 2 {
		t.Fatalf("calls %v evidence %d notes %d", q.calls, len(out[0].Evidence), len(notes))
	}
}

// TestAttachDoesNotQueryADroppedFinding proves a dropped finding is left as
// it is. The engine is nil, so any query would fail the test.
func TestAttachDoesNotQueryADroppedFinding(t *testing.T) {
	c := kev("F3", "CVE-2026-0001")
	c.Dropped = true
	out, notes, err := Attach(context.Background(), nil, "", "", []candidate.Candidate{c})
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 || len(out[0].Evidence) != 1 {
		t.Fatalf("notes %#v evidence %d", notes, len(out[0].Evidence))
	}
}

// TestAttachIgnoresACellThatIsNotACVE proves a cell that is not exactly a
// CVE id, such as CVE-2026-1, sends no query.
func TestAttachIgnoresACellThatIsNotACVE(t *testing.T) {
	q := &memQuery{}
	c := candidate.Candidate{
		Risk: candidate.Risk{ID: "F2"},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindQuery,
			Query: &candidate.QueryEvidence{
				Columns: []string{"package", "cve"},
				Rows:    [][]string{{"evil.example/leftpad", "not-a-cve"}, {"CVE-2026-1", "true"}},
			},
		}},
	}
	out, notes, err := Attach(context.Background(), q, "", "", []candidate.Candidate{c})
	if err != nil {
		t.Fatal(err)
	}
	if len(q.calls) != 0 || len(notes) != 0 || len(out[0].Evidence) != 1 {
		t.Fatalf("calls %v notes %#v", q.calls, notes)
	}
}

// kev is a kept finding whose KEV evidence names one CVE.
func kev(id, cve string) candidate.Candidate {
	return candidate.Candidate{
		Risk: candidate.Risk{ID: id, Severity: "ERROR", Title: cve},
		Evidence: []candidate.Evidence{{
			Kind: candidate.KindQuery,
			Query: &candidate.QueryEvidence{
				Columns: []string{"cve_id", "vendor_project"},
				Rows:    [][]string{{cve, "sample"}},
			},
		}},
	}
}

// memQuery is an Adapter that records each SQL it is sent and answers a
// score row for a CVE it knows.
type memQuery struct {
	calls []string
	rows  map[string][]string
}

func (m *memQuery) Query(_ context.Context, sql string) (query.Result, error) {
	m.calls = append(m.calls, sql)
	for id, row := range m.rows {
		if strings.Contains(sql, "'"+id+"'") {
			return query.Result{
				SQL: sql, Engine: "mem",
				Columns: []string{"cve_id", "epss", "percentile", "score_date"},
				Rows:    [][]string{row},
			}, nil
		}
	}
	return query.Result{
		SQL: sql, Engine: "mem",
		Columns: []string{"cve_id", "epss", "percentile", "score_date"},
		Rows:    [][]string{},
	}, nil
}

// repoRoot finds the module root, where the sample fixtures live.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.Contains(string(b), "module github.com/saphan-ai/proven") {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
