package usl_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The local engine of internal/query reads one inner join, so it cannot run the
// relation view. python3's sqlite3 runs the view's own SELECT text, with the
// container as an attached database, over a fixture built to break the grain.
// What Zetaris's dialect makes of the same text is not measured here.

const sqliteRunner = `import json, sqlite3, sys
setup, select = open(sys.argv[1]).read(), open(sys.argv[2]).read()
db = sqlite3.connect(":memory:")
db.execute("ATTACH DATABASE ':memory:' AS px_usl")
db.executescript(setup)
print(json.dumps(db.execute(select).fetchall()))
`

// grainFixture: one package range whose advisory GHSA-x has two aliases, one of
// them a CVE with two EPSS score dates and a KEV row; one package range with no
// advisory row at all.
const grainFixture = `
CREATE TABLE px_usl.px_package (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
CREATE TABLE px_usl.px_advisory (id TEXT, alias TEXT, published TEXT, severity TEXT);
CREATE TABLE px_usl.px_kev_cve (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
CREATE TABLE px_usl.px_epss_score (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
INSERT INTO px_usl.px_package VALUES ('GHSA-x', '', 'Go', 'example.com/a', '0', '1.2.3', '', '', '');
INSERT INTO px_usl.px_package VALUES ('GHSA-none', '', 'Go', 'example.com/b', '0', '2.0.0', '', '', '');
INSERT INTO px_usl.px_advisory VALUES ('GHSA-x', 'CVE-2026-0001', '', '');
INSERT INTO px_usl.px_advisory VALUES ('GHSA-x', 'GO-2026-0001', '', '');
INSERT INTO px_usl.px_kev_cve VALUES ('CVE-2026-0001', 'v', 'p', '2026-10-01', 'Unknown');
INSERT INTO px_usl.px_epss_score VALUES ('CVE-2026-0001', '0.5', '0.9', '2026-10-19');
INSERT INTO px_usl.px_epss_score VALUES ('CVE-2026-0001', '0.6', '0.95', '2026-10-20');
`

const secondScoreDate = "INSERT INTO px_usl.px_epss_score VALUES ('CVE-2026-0001', '0.6', '0.95', '2026-10-20');\n"

// relationRows runs the SELECT of the relation view over the fixture setup and
// returns each row as "package|advisory|cve|kev|epss|score_date".
func relationRows(t *testing.T, setup string) []string {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("no python3 on the path")
	}
	var sel string
	for _, st := range parseStatements(t, read(t, uslPath("layer.sql"))) {
		if strings.HasPrefix(st.SQL, "CREATE SCHEMASTORE VIEW px_package_evidence ") {
			sel = st.SQL[strings.Index(st.SQL, " AS SELECT ")+len(" AS "):]
		}
	}
	if sel == "" {
		t.Fatal("no px_package_evidence view")
	}
	sel = "SELECT package, advisory_id, cve_id, kev_date_added, epss, epss_score_date FROM (" + sel + ") ORDER BY 1, 2, 3, 6"
	dir := t.TempDir()
	for name, text := range map[string]string{"run.py": sqliteRunner, "setup.sql": setup, "select.sql": sel} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	out, err := exec.Command("python3", "-I", filepath.Join(dir, "run.py"), filepath.Join(dir, "setup.sql"), filepath.Join(dir, "select.sql")).CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	var rows [][]any
	if err := json.Unmarshal(out, &rows); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, r := range rows {
		parts := make([]string, len(r))
		for i, c := range r {
			if s, ok := c.(string); ok {
				parts[i] = s
			} else {
				parts[i] = "NULL"
			}
		}
		got = append(got, strings.Join(parts, "|"))
	}
	sort.Strings(got)
	return got
}

// The grain the runbook states: one row per package range, per alias of its
// advisory, per EPSS score date. Two aliases and two score dates make three
// rows for the first range, and the range with no advisory is one row of NULLs.
func TestRelationViewGrainIsPerAliasPerScoreDate(t *testing.T) {
	want := []string{
		"example.com/a|GHSA-x|CVE-2026-0001|2026-10-01|0.5|2026-10-19",
		"example.com/a|GHSA-x|CVE-2026-0001|2026-10-01|0.6|2026-10-20",
		"example.com/a|GHSA-x|GO-2026-0001|NULL|NULL|NULL",
		"example.com/b|GHSA-none|NULL|NULL|NULL|NULL",
	}
	got := relationRows(t, grainFixture)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("relation rows:\n got %q\nwant %q", got, want)
	}
}

// Negative control, law #27: the mutated field is the number of EPSS dates of
// CVE-2026-0001. With the second date removed that CVE's rows drop from two to
// one, so the grain test is red for the grain and not for the fixture.
func TestRelationViewGrainFollowsTheScoreDates(t *testing.T) {
	one := strings.Replace(grainFixture, secondScoreDate, "", 1)
	if one == grainFixture {
		t.Fatal("the fixture lost its second score date")
	}
	if got := relationRows(t, one); len(got) != 3 {
		t.Fatalf("one score date: %d rows, want 3: %q", len(got), got)
	}
}
