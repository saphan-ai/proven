package usl_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
)

// update rewrites the recorded sample-mode answers from sample mode.
//
//	go test ./internal/usl -run TestParityRecordedAnswersMatchSampleMode -update
var update = flag.Bool("update", false, "rewrite testdata/zetaris-usl/answers.json from sample mode")

// The sample tables the answers are recorded from: the same tables a sample
// review runs over (testdata/fixture/discount).
var tablesPath = filepath.Join(root, "testdata", "fixture", "discount", "tables.sql")

var answersPath = filepath.Join(root, "testdata", "zetaris-usl", "answers.json")

// The names the layer's views take in Proven's settings. They are the values
// docs/ZETARIS_USL.md tells an administrator to set.
var layerRelations = querytpl.Relations{
	OSVTable:  "px_usl.px_package",
	KEVView:   "px_usl.px_kev_cve",
	EPSSView:  "px_usl.px_epss_score",
	OSVIDView: "px_usl.px_advisory",
}

// answer is what one request returned: the columns and the rows, as text.
type answer struct {
	Name     string            `json:"name"`
	Template string            `json:"template"`
	Params   map[string]string `json:"params"`
	Columns  []string          `json:"columns"`
	Rows     [][]string        `json:"rows"`
}

func (a answer) request() querytpl.Request {
	return querytpl.Request{Template: a.Template, Params: a.Params}
}

// requests is the fixed set: one per template the layer serves, and two that
// find nothing. The values are in the sample tables.
var requests = []answer{
	{Name: "osv-by-package finds two ranges", Template: querytpl.OSVByPackage, Params: map[string]string{"ecosystem": "Go", "package": "evil.example/leftpad"}},
	{Name: "osv-by-package finds nothing", Template: querytpl.OSVByPackage, Params: map[string]string{"ecosystem": "Go", "package": "absent.example/nothing"}},
	{Name: "osv-by-advisory finds one", Template: querytpl.OSVByAdvisory, Params: map[string]string{"advisory_id": "GHSA-px-sample-0001"}},
	{Name: "kev-by-cve finds one", Template: querytpl.KEVByCVE, Params: map[string]string{"cve_id": "CVE-2026-0001"}},
	{Name: "kev-by-cve finds nothing", Template: querytpl.KEVByCVE, Params: map[string]string{"cve_id": "CVE-2026-9999"}},
	{Name: "epss-by-cve finds one", Template: querytpl.EPSSByCVE, Params: map[string]string{"cve_id": "CVE-2021-44228"}},
}

// sampleAnswers asks sample mode each request, with the default names.
func sampleAnswers(t *testing.T) []answer {
	t.Helper()
	store, err := query.OpenSample(tablesPath)
	if err != nil {
		t.Fatal(err)
	}
	var out []answer
	for _, r := range requests {
		res, err := store.Run(context.Background(), r.request(), querytpl.Relations{})
		if err != nil {
			t.Fatalf("%s: %v", r.Name, err)
		}
		r.Columns, r.Rows = res.Columns, res.Rows
		out = append(out, r)
	}
	return out
}

// layerDB builds the layer over the sample tables on the local engine: each
// entity view of layer.sql becomes a table of the projected rows of its source.
// The view definitions are read from layer.sql, so an edit there changes the
// expected answers.
func layerDB(t *testing.T) *query.DB {
	t.Helper()
	tables := read(t, tablesPath)
	src, err := query.LoadSQL(tables)
	if err != nil {
		t.Fatal(err)
	}
	ents, _ := entities(t)
	var b strings.Builder
	b.WriteString(tables)
	b.WriteString("\n")
	for _, e := range ents {
		_, rows, err := src.Query("SELECT " + strings.Join(e.Columns, ", ") + " FROM " + e.Source)
		if err != nil {
			t.Fatalf("view %s over %s: %v", e.Name, e.Source, err)
		}
		typed := make([]string, len(e.Columns))
		for i, c := range e.Columns {
			typed[i] = c + " TEXT"
		}
		fmt.Fprintf(&b, "CREATE TABLE %s (%s);\n", e.Name, strings.Join(typed, ", "))
		for _, row := range rows {
			cells := make([]string, len(row))
			for i, c := range row {
				cells[i] = "'" + strings.ReplaceAll(c, "'", "''") + "'"
			}
			fmt.Fprintf(&b, "INSERT INTO %s VALUES (%s);\n", e.Name, strings.Join(cells, ", "))
		}
	}
	db, err := query.LoadSQL(b.String())
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// layerAnswers asks the layer each request, with the layer's names.
func layerAnswers(t *testing.T, db *query.DB) []answer {
	t.Helper()
	var out []answer
	for _, r := range requests {
		sql, err := querytpl.Build(r.request(), layerRelations)
		if err != nil {
			t.Fatalf("%s: %v", r.Name, err)
		}
		cols, rows, err := db.Query(sql)
		if err != nil {
			t.Fatalf("%s: %v", r.Name, err)
		}
		r.Columns, r.Rows = cols, rows
		out = append(out, r)
	}
	return out
}

// diff compares two answers to one request. Rows are compared as a multiset: a
// SELECT with no ORDER BY has no order, on Zetaris or here. It returns "" when
// the answers are equal, and otherwise says what is different.
func diff(want, got answer) string {
	if strings.Join(want.Columns, ",") != strings.Join(got.Columns, ",") {
		return fmt.Sprintf("columns: want %v, got %v", want.Columns, got.Columns)
	}
	w, g := keys(want.Rows), keys(got.Rows)
	if strings.Join(w, "\n") == strings.Join(g, "\n") {
		return ""
	}
	return fmt.Sprintf("rows: want %d %q, got %d %q", len(w), w, len(g), g)
}

func keys(rows [][]string) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		b, _ := json.Marshal(r)
		out[i] = string(b)
	}
	sort.Strings(out)
	return out
}

func loadAnswers(t *testing.T) []answer {
	t.Helper()
	var out []answer
	if err := json.Unmarshal([]byte(read(t, answersPath)), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// The recorded answers are sample mode's, not a copy that drifted.
func TestParityRecordedAnswersMatchSampleMode(t *testing.T) {
	got := sampleAnswers(t)
	if *update {
		b, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(answersPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(answersPath, append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want := loadAnswers(t)
	if len(want) != len(got) {
		t.Fatalf("recorded %d answers, sample mode gives %d", len(want), len(got))
	}
	for i := range got {
		if want[i].Name != got[i].Name {
			t.Fatalf("answer %d: recorded %q, request is %q", i, want[i].Name, got[i].Name)
		}
		if d := diff(want[i], got[i]); d != "" {
			t.Errorf("%s: the recorded answer is not sample mode's: %s", got[i].Name, d)
		}
	}
}

// The request set covers every template the layer serves and a request that
// finds nothing, and the recorded answers are not all empty.
func TestParityRequestSetCoversTheServedTemplates(t *testing.T) {
	served := map[string]bool{
		querytpl.OSVByPackage: false, querytpl.OSVByAdvisory: false, querytpl.KEVByCVE: false, querytpl.EPSSByCVE: false,
	}
	empty, full := 0, 0
	for _, a := range loadAnswers(t) {
		if _, ok := served[a.Template]; !ok {
			t.Errorf("%s: template %s is not one the layer serves", a.Name, a.Template)
		}
		served[a.Template] = true
		if len(a.Rows) == 0 {
			empty++
		} else {
			full++
		}
	}
	for tpl, ok := range served {
		if !ok {
			t.Errorf("no request for template %s", tpl)
		}
	}
	if empty == 0 || full == 0 {
		t.Errorf("want requests that find rows and requests that find nothing: %d with rows, %d empty", full, empty)
	}
}

// The layer, built from layer.sql over the sample tables, returns the rows sample
// mode returned.
func TestParityLayerReturnsTheRecordedRows(t *testing.T) {
	want := loadAnswers(t)
	got := layerAnswers(t, layerDB(t))
	for i := range got {
		if d := diff(want[i], got[i]); d != "" {
			t.Errorf("%s: %s", got[i].Name, d)
		}
	}
}

// Negative control (law #27). The comparer is red when one expected row is
// changed, and when one row is missing or extra. Mutated field: one cell of one
// recorded row.
func TestParityComparerIsRedWhenAnExpectedRowChanges(t *testing.T) {
	got := layerAnswers(t, layerDB(t))
	for i, a := range loadAnswers(t) {
		if len(a.Rows) == 0 {
			continue
		}
		mutated := a
		mutated.Rows = make([][]string, len(a.Rows))
		for j := range a.Rows {
			mutated.Rows[j] = append([]string(nil), a.Rows[j]...)
		}
		mutated.Rows[0][0] += "-changed"
		if diff(mutated, got[i]) == "" {
			t.Errorf("%s: a changed cell was not seen", a.Name)
		}
		dropped := a
		dropped.Rows = a.Rows[1:]
		if diff(dropped, got[i]) == "" {
			t.Errorf("%s: a missing row was not seen", a.Name)
		}
		extra := a
		extra.Rows = append(append([][]string(nil), a.Rows...), a.Rows[0])
		if diff(extra, got[i]) == "" {
			t.Errorf("%s: an extra row was not seen", a.Name)
		}
		renamed := a
		renamed.Columns = append([]string(nil), a.Columns...)
		renamed.Columns[0] += "_x"
		if diff(renamed, got[i]) == "" {
			t.Errorf("%s: a renamed column was not seen", a.Name)
		}
	}
}

// Negative control on the recording. Mutated field: the cve_id of one row of the
// KEV source table. Sample mode then answers differently from the recorded
// answer, and the comparer must see it.
func TestParityIsRedWhenASourceRowChanges(t *testing.T) {
	bad := strings.Replace(read(t, tablesPath), "INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-0001'", "INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-0003'", 1)
	if bad == read(t, tablesPath) {
		t.Fatal("the mutation did not change the tables")
	}
	// The recorded answers were made from the unmutated tables.
	dir := t.TempDir()
	mutated := filepath.Join(dir, "tables.sql")
	if err := os.WriteFile(mutated, []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := query.OpenSample(mutated)
	if err != nil {
		t.Fatal(err)
	}
	r := requests[3]
	res, err := store.Run(context.Background(), r.request(), querytpl.Relations{})
	if err != nil {
		t.Fatal(err)
	}
	r.Columns, r.Rows = res.Columns, res.Rows
	if diff(loadAnswers(t)[3], r) == "" {
		t.Fatal("a changed source row was not seen")
	}
}

// The live comparison, run by `make zetaris-usl-parity`. It needs a Zetaris the
// layer is applied to. Without the three Zetaris variables it is skipped, by name.
//
// The data on Zetaris is not the sample data, so the live parity compares the
// layer to its own sources: every request is asked twice through the same
// adapter, once with the default names and once with the layer's, and the rows
// must be equal.
func TestParityLive(t *testing.T) {
	var missing []string
	for _, name := range []string{query.EnvRESTURL, query.EnvOrgID, query.EnvAPIKey} {
		if os.Getenv(name) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Skipf("zetaris-usl-parity skipped: %s not set; this check runs against a Zetaris the layer is applied to", strings.Join(missing, ", "))
	}
	cfg, err := query.Load(os.Getenv)
	if err != nil {
		t.Fatal(err)
	}
	live := query.NewLive(cfg, nil)
	ctx := context.Background()
	ask := func(r answer, rel querytpl.Relations) answer {
		sql, err := querytpl.Build(r.request(), rel)
		if err != nil {
			t.Fatalf("%s: %v", r.Name, err)
		}
		res, err := live.Query(ctx, sql)
		if err != nil {
			t.Fatalf("%s: %v", r.Name, err)
		}
		r.Columns, r.Rows = res.Columns, res.Rows
		return r
	}
	for _, r := range requests {
		if d := diff(ask(r, querytpl.Relations{}), ask(r, layerRelations)); d != "" {
			t.Errorf("%s: the layer and its source differ: %s", r.Name, d)
		}
	}
}
