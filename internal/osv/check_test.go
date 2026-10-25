package osv

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

// TestCheckKeepsAMatchAndLooksUpTheCVE proves that a Go module inside a range
// becomes one candidate on its go.mod line, that the message names the row
// that matched and the row that did not, and that the evidence is the OSV
// query with only this package's rows plus the KEV query for its CVE alias.
func TestCheckKeepsAMatchAndLooksUpTheCVE(t *testing.T) {
	root := t.TempDir()
	writeMods(t, root,
		"module m\n\ngo 1.26\n",
		"module m\n\ngo 1.26\n\nrequire example.com/mod v1.2.3\n",
	)
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-1', 'CVE-2026-0001', 'Go', 'example.com/mod', '1.0.0', '1.2.4', '', 'HIGH', '2026-01-15');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-2', '', 'Go', 'example.com/mod', '2.0.0', '2.1.0', '', 'LOW', '2026-03-01');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-other', '', 'Go', 'example.com/other', '0', '9.0.0', '', 'LOW', '2026-04-01');
CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-0001', 'example', 'mod', '2026-01-15', 'Known');
`)
	out, err := Check(context.Background(), q, root, "", "snap-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 1 {
		t.Fatalf("%#v", out)
	}
	c := out.Candidates[0]
	if c.ID != "osv:example.com/mod" || c.Path != "go.mod" || c.StartLine != 5 || c.Severity != "ERROR" {
		t.Fatalf("%#v", c.Risk)
	}
	if !strings.Contains(c.Message, "introduced 1.0.0 <= v1.2.3 < fixed 1.2.4") {
		t.Fatalf("message %s", c.Message)
	}
	if !strings.Contains(c.Message, "GHSA-2 does not match") || !strings.Contains(c.Message, "alias CVE-2026-0001 is on the KEV list") {
		t.Fatalf("message %s", c.Message)
	}
	if len(c.Evidence) != 2 || c.Evidence[0].Query == nil || c.Evidence[1].Query == nil {
		t.Fatalf("evidence %#v", c.Evidence)
	}
	if c.Evidence[0].Query.Snapshot != "snap-1" || len(c.Evidence[0].Query.Rows) != 2 {
		t.Fatalf("osv query %#v", c.Evidence[0].Query)
	}
	if c.Evidence[0].Query.Rows[0][0] != "GHSA-1" || c.Evidence[1].Query.Rows[0][0] != "CVE-2026-0001" {
		t.Fatalf("rows %#v / %#v", c.Evidence[0].Query.Rows, c.Evidence[1].Query.Rows)
	}
	for _, row := range c.Evidence[0].Query.Rows {
		if row[3] == "example.com/other" {
			t.Fatal("the other package was returned")
		}
	}
}

// TestCheckTracesAnUnaffectedVersion proves that a version past every fixed
// bound is no finding, and that the trace note gives the comparison.
func TestCheckTracesAnUnaffectedVersion(t *testing.T) {
	root := t.TempDir()
	writeMods(t, root, "module m\n\ngo 1.26\n", "module m\n\ngo 1.26\n\nrequire example.com/mod v9.0.0\n")
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-1', '', 'Go', 'example.com/mod', '1.0.0', '2.0.0', '', 'LOW', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 0 {
		t.Fatalf("finding %#v", out.Candidates)
	}
	if len(out.Notes) != 1 || !strings.Contains(out.Notes[0], "no affected row") || !strings.Contains(out.Notes[0], "fixed 2.0.0 <= v9.0.0") {
		t.Fatalf("notes %#v", out.Notes)
	}
}

// TestCheckNamesAMissingPackage proves that a package with no row is no
// finding, and that the note says the query returned no rows.
func TestCheckNamesAMissingPackage(t *testing.T) {
	root := t.TempDir()
	writeMods(t, root, "module m\n\ngo 1.26\n", "module m\n\ngo 1.26\n\nrequire example.com/mod v1.0.0\n")
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-1', '', 'Go', 'example.com/other', '0', '9.0.0', '', 'LOW', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil || len(out.Candidates) != 0 || len(out.Notes) != 1 || !strings.Contains(out.Notes[0], "no rows") {
		t.Fatalf("err %v notes %#v cands %d", err, out.Notes, len(out.Candidates))
	}
}

// TestCheckSkipsALoweredModule proves that a lowered module is not checked:
// no query runs, so even a missing query engine is not needed.
func TestCheckSkipsALoweredModule(t *testing.T) {
	root := t.TempDir()
	writeMods(t, root,
		"module m\n\ngo 1.26\n\nrequire example.com/mod v1.2.3\n",
		"module m\n\ngo 1.26\n\nrequire example.com/mod v1.0.0\n",
	)
	out, err := Check(context.Background(), nil, root, "", "snap")
	if err != nil || len(out.Candidates) != 0 || len(out.Notes) != 1 || out.Notes[0] != "no dependency added or raised" {
		t.Fatalf("err %v %#v", err, out)
	}
}

// TestCheckNamesADownQuery proves that an unreachable query engine is named
// as a failed service with a trace note, not returned as an error, and
// gives no finding.
func TestCheckNamesADownQuery(t *testing.T) {
	root := t.TempDir()
	writeMods(t, root, "module m\n\ngo 1.26\n", "module m\n\ngo 1.26\n\nrequire example.com/mod v1.0.0\n")
	out, err := Check(context.Background(), downQuery{}, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if out.Down == nil || out.Down.Service != "query" || len(out.Candidates) != 0 {
		t.Fatalf("%#v", out)
	}
	if len(out.Notes) != 1 || !strings.Contains(out.Notes[0], "did not run") {
		t.Fatalf("notes %#v", out.Notes)
	}
}

// TestCheckNPMMatchesASemver proves that a pinned npm version in package.json
// is compared as a semantic version and that the query filters on the npm
// ecosystem and the package name.
func TestCheckNPMMatchesASemver(t *testing.T) {
	root := t.TempDir()
	writePair(t, root, "package.json", "{}\n", "{\n  \"dependencies\": {\n    \"left-pad\": \"1.2.3\"\n  }\n}\n")
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-npm', '', 'npm', 'left-pad', '1.0.0', '1.2.4', '', 'HIGH', '2026-01-15');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-other', '', 'npm', 'right-pad', '0', '9.0.0', '', 'LOW', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 1 {
		t.Fatalf("%#v", out)
	}
	c := out.Candidates[0]
	if c.Path != "package.json" || c.StartLine != 3 || !strings.Contains(c.Message, "introduced 1.0.0 <= 1.2.3 < fixed 1.2.4") {
		t.Fatalf("%s line %d", c.Message, c.StartLine)
	}
	if c.Evidence[0].Query == nil || !strings.Contains(c.Evidence[0].Query.SQL, "ecosystem = 'npm' AND package = 'left-pad'") {
		t.Fatalf("sql %#v", c.Evidence[0].Query)
	}
}

// TestCheckPyPIMatchesARelease proves that a pin in requirements.txt is
// compared with PEP 440 and queried under the PyPI ecosystem.
func TestCheckPyPIMatchesARelease(t *testing.T) {
	root := t.TempDir()
	writePair(t, root, "requirements.txt", "", "leftpad==1.2.3\n")
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('PYSEC-1', '', 'PyPI', 'leftpad', '1.0', '1.2.4', '', 'HIGH', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 1 {
		t.Fatalf("%#v", out)
	}
	c := out.Candidates[0]
	if c.Path != "requirements.txt" || c.StartLine != 1 || !strings.Contains(c.Message, "introduced 1.0 <= 1.2.3 < fixed 1.2.4") {
		t.Fatalf("%s line %d", c.Message, c.StartLine)
	}
	if !strings.Contains(c.Evidence[0].Query.SQL, "ecosystem = 'PyPI' AND package = 'leftpad'") {
		t.Fatal(c.Evidence[0].Query.SQL)
	}
}

// TestCheckUnparsedVersionIsNotAFinding proves that a version that is not
// PEP 440 gives no finding even against a range that covers everything,
// and that the trace says why, with the SQL and the skipped line.
func TestCheckUnparsedVersionIsNotAFinding(t *testing.T) {
	root := t.TempDir()
	writePair(t, root, "requirements.txt", "", "leftpad==latest\nDjango>=1\n")
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('PYSEC-1', '', 'PyPI', 'leftpad', '0', '99', '', 'HIGH', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 0 {
		t.Fatalf("finding %#v", out.Candidates)
	}
	body := strings.Join(out.Notes, "\n")
	for _, want := range []string{
		"latest is not a release, pre-release or post-release",
		"ecosystem = 'PyPI'",
		`line 2 skipped ("Django>=1"): not a pinned name==version`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
}

// TestCheckNPMRangeIsNotAFinding proves that an npm range is not rewritten
// into a version: it gives no finding, and the note says it is not a
// semantic version.
func TestCheckNPMRangeIsNotAFinding(t *testing.T) {
	root := t.TempDir()
	writePair(t, root, "package.json", "{}\n", "{\"dependencies\":{\"left-pad\":\"^1.2.3\"}}\n")
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-npm', '', 'npm', 'left-pad', '0', '99.0.0', '', 'HIGH', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil || len(out.Candidates) != 0 {
		t.Fatalf("err %v cands %#v", err, out.Candidates)
	}
	if len(out.Notes) != 1 || !strings.Contains(out.Notes[0], "^1.2.3 is not a semantic version") {
		t.Fatalf("%#v", out.Notes)
	}
}

// TestCheckEmptyTableStillSaysWhyAVersionWasNotCompared proves that a note
// for a query with no rows also says when the version could not have been
// compared.
func TestCheckEmptyTableStillSaysWhyAVersionWasNotCompared(t *testing.T) {
	root := t.TempDir()
	writePair(t, root, "requirements.txt", "", "leftpad==1.0.dev1\n")
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil || len(out.Candidates) != 0 || len(out.Notes) != 1 {
		t.Fatalf("err %v %#v", err, out)
	}
	if !strings.Contains(out.Notes[0], "no rows") || !strings.Contains(out.Notes[0], "has a dev release") {
		t.Fatal(out.Notes[0])
	}
}

// TestCheckGoSumFindsATransitive proves that a module only go.sum names, a
// transitive one, is checked and its finding points at go.sum.
func TestCheckGoSumFindsATransitive(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"base/go.mod":   "module m\n\ngo 1.26\n",
		"change/go.mod": "module m\n\ngo 1.26\n",
		"change/go.sum": "example.com/trans v1.2.3 h1:abc\nexample.com/trans v1.2.3/go.mod h1:def\n",
	})
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-t', '', 'Go', 'example.com/trans', '1.0.0', '1.2.4', '', 'HIGH', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 1 {
		t.Fatalf("%#v", out)
	}
	c := out.Candidates[0]
	if c.Path != "go.sum" || !strings.Contains(c.Message, "example.com/trans v1.2.3") {
		t.Fatalf("%s %s", c.Path, c.Message)
	}
	if !strings.Contains(c.Evidence[0].Query.SQL, "ecosystem = 'Go' AND package = 'example.com/trans'") {
		t.Fatal(c.Evidence[0].Query.SQL)
	}
}

// TestCheckLockVersionWins proves that the lock file's exact version is the
// one compared, and that the manifest's range for the same package is not
// checked at all.
func TestCheckLockVersionWins(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"base/package.json":      "{}\n",
		"change/package.json":    "{\"dependencies\":{\"left-pad\":\"^9.9.9\"}}\n",
		"base/package-lock.json": "{\"lockfileVersion\":3,\"packages\":{}}\n",
		"change/package-lock.json": `{
  "lockfileVersion": 3,
  "packages": {
    "node_modules/left-pad": {"version": "1.2.3"}
  }
}
`,
	})
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-npm', '', 'npm', 'left-pad', '1.0.0', '1.2.4', '', 'HIGH', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 1 {
		t.Fatalf("%#v", out)
	}
	c := out.Candidates[0]
	if c.Path != "package-lock.json" || !strings.Contains(c.Message, "left-pad 1.2.3") || strings.Contains(c.Message, "^9.9.9") {
		t.Fatalf("%s %s", c.Path, c.Message)
	}
	if strings.Contains(strings.Join(out.Notes, "\n"), "not a semantic version") {
		t.Fatalf("manifest range was checked: %#v", out.Notes)
	}
}

// TestCheckPoetryLockMatches proves that a poetry.lock pin is checked under
// the PyPI ecosystem and its finding points at poetry.lock.
func TestCheckPoetryLockMatches(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"base/keep.txt":      "ok\n",
		"change/poetry.lock": "[[package]]\nname = \"leftpad\"\nversion = \"1.2.3\"\n",
	})
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('PYSEC-1', '', 'PyPI', 'leftpad', '1.0', '1.2.4', '', 'HIGH', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 1 {
		t.Fatalf("%#v", out)
	}
	c := out.Candidates[0]
	if c.Path != "poetry.lock" || !strings.Contains(c.Message, "leftpad 1.2.3") {
		t.Fatalf("%s %s", c.Path, c.Message)
	}
	if !strings.Contains(c.Evidence[0].Query.SQL, "ecosystem = 'PyPI' AND package = 'leftpad'") {
		t.Fatal(c.Evidence[0].Query.SQL)
	}
}

// TestCheckPyPIQueriesTheNormalisedName proves that a pin written as Requests
// is queried as requests. The OSV table keys PyPI packages by the PEP 503 name,
// so the spelling in the file would miss the row.
func TestCheckPyPIQueriesTheNormalisedName(t *testing.T) {
	root := t.TempDir()
	writePair(t, root, "requirements.txt", "", "Requests==1.2.3\n")
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('PYSEC-1', '', 'PyPI', 'requests', '1.0', '1.2.4', '', 'HIGH', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 1 {
		t.Fatalf("%#v", out)
	}
	c := out.Candidates[0]
	if c.Path != "requirements.txt" || !strings.Contains(c.Message, "requests 1.2.3") {
		t.Fatalf("%s %s", c.Path, c.Message)
	}
	if c.Evidence[0].Query == nil || !strings.Contains(c.Evidence[0].Query.SQL, "ecosystem = 'PyPI' AND package = 'requests'") {
		t.Fatalf("sql %#v", c.Evidence[0].Query)
	}
	if strings.Contains(c.Evidence[0].Query.SQL, "Requests") {
		t.Fatal(c.Evidence[0].Query.SQL)
	}
}

// TestCheckMalformedLockIsANote proves that lock files that do not parse are
// trace notes, and that go.mod beside them is still checked and found.
func TestCheckMalformedLockIsANote(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"base/go.mod":              "module m\n\ngo 1.26\n",
		"change/go.mod":            "module m\n\ngo 1.26\n\nrequire example.com/mod v1.2.3\n",
		"change/go.sum":            "not-a-sum\n",
		"change/package-lock.json": "{",
	})
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-1', '', 'Go', 'example.com/mod', '1.0.0', '1.2.4', '', 'HIGH', '2026-01-15');
`)
	out, err := Check(context.Background(), q, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 1 || out.Candidates[0].Path != "go.mod" {
		t.Fatalf("%#v", out.Candidates)
	}
	body := strings.Join(out.Notes, "\n")
	for _, want := range []string{
		"go.sum: change: line 1: not a module version hash",
		"package-lock.json: change: invalid JSON:",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
}

// TestLoadTableRefusesABadName proves that an empty setting is the default,
// that a valid name is kept, and that a name carrying SQL is refused with a
// ConfigError that names the variable on one line.
func TestLoadTableRefusesABadName(t *testing.T) {
	if _, err := LoadTable(func(string) string { return "" }); err != nil {
		t.Fatal(err)
	}
	got, err := LoadTable(func(k string) string {
		if k == EnvTable {
			return "px_osvfile.px_osv_pkg"
		}
		return ""
	})
	if err != nil || got != DefaultTable {
		t.Fatalf("%s %v", got, err)
	}
	_, err = LoadTable(func(string) string { return "px_osv.pkg;drop" })
	var cfg *ConfigError
	if err == nil || !asConfig(err, &cfg) || cfg.Field != EnvTable || strings.Contains(err.Error(), "\n") {
		t.Fatal(err)
	}
}

// TestLoadViewEmptyIsTheDefaultAndABadNameIsRefused proves an empty setting
// is KEVView, a real view name is kept, and a name carrying SQL is refused
// with the variable on one line.
// Mutated field: PROVEN_KEV_VIEW.
func TestLoadViewEmptyIsTheDefaultAndABadNameIsRefused(t *testing.T) {
	got, err := LoadView(nil)
	if err != nil || got != KEVView {
		t.Fatalf("nil %s %v", got, err)
	}
	got, err = LoadView(func(string) string { return "  " })
	if err != nil || got != KEVView {
		t.Fatalf("blank %s %v", got, err)
	}
	got, err = LoadView(func(k string) string {
		if k == EnvKEVView {
			return "px_evidence.px_kev_alt"
		}
		return ""
	})
	if err != nil || got != "px_evidence.px_kev_alt" {
		t.Fatalf("set %s %v", got, err)
	}
	_, err = LoadView(func(string) string { return "px_evidence.bad;drop" })
	var cfg *ConfigError
	if err == nil || !asConfig(err, &cfg) || cfg.Field != EnvKEVView || cfg.Value != "px_evidence.bad;drop" || strings.Contains(err.Error(), "\n") {
		t.Fatal(err)
	}
	if !strings.Contains(err.Error(), "allowed") {
		t.Fatal(err)
	}
}

// TestCheckUsesTheKEVViewOnTheContext proves the lookup queries the view the
// review loaded, not the default table beside it.
// Mutated field: the view on the context. Leaving it unset queries
// px_evidence.px_kev_vulns, which has no row in this fixture.
func TestCheckUsesTheKEVViewOnTheContext(t *testing.T) {
	root := t.TempDir()
	writeMods(t, root,
		"module m\n\ngo 1.26\n",
		"module m\n\ngo 1.26\n\nrequire example.com/mod v1.2.3\n",
	)
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-1', 'CVE-2026-0001', 'Go', 'example.com/mod', '1.0.0', '1.2.4', '', 'HIGH', '2026-01-15');
CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
CREATE TABLE px_evidence.px_kev_alt (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
INSERT INTO px_evidence.px_kev_alt VALUES ('CVE-2026-0001', 'example', 'mod', '2026-01-15', 'Known');
`)
	ctx := WithKEVView(context.Background(), "px_evidence.px_kev_alt")
	out, err := Check(ctx, q, root, "", "snap-kev")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 1 {
		t.Fatalf("%#v", out)
	}
	c := out.Candidates[0]
	if !strings.Contains(c.Message, "alias CVE-2026-0001 is on the KEV list") {
		t.Fatalf("message %s", c.Message)
	}
	if len(c.Evidence) != 2 || c.Evidence[1].Query == nil {
		t.Fatalf("evidence %#v", c.Evidence)
	}
	if !strings.Contains(c.Evidence[1].Query.SQL, "FROM px_evidence.px_kev_alt ") {
		t.Fatalf("sql %s", c.Evidence[1].Query.SQL)
	}
}

// downQuery is a query adapter whose service is always down.
type downQuery struct{}

func (downQuery) Run(context.Context, querytpl.Request, querytpl.Relations) (query.Result, error) {
	return query.Result{}, &service.Down{Service: "query", Reason: "unreachable"}
}

// asConfig reports whether err is a *ConfigError and stores it in target.
func asConfig(err error, target **ConfigError) bool {
	cfg, ok := err.(*ConfigError)
	if ok {
		*target = cfg
	}
	return ok
}

// writePair writes one file on the base and the change side. An empty body
// leaves the file out of that side.
func writePair(t *testing.T, root, name, base, change string) {
	t.Helper()
	for _, side := range []struct{ dir, body string }{{"base", base}, {"change", change}} {
		dir := filepath.Join(root, side.dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if side.body == "" {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(side.body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// writeTree writes files by slash path under root; each path starts with
// base/ or change/.
func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// writeMods writes go.mod on both sides.
func writeMods(t *testing.T, root, base, change string) {
	t.Helper()
	for _, side := range []struct{ name, body string }{{"base", base}, {"change", change}} {
		dir := filepath.Join(root, side.name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(side.body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestRaisedPackageBindsTheADR0039Templates proves a package the change
// added binds the malicious, scorecard and licence templates, and does not
// bind osv-by-package until a range matches. CVSS binds only a CVE that
// match recorded. The mutated field is the raised set.
func TestRaisedPackageBindsTheADR0039Templates(t *testing.T) {
	var f Facts
	f.noteRaised("Go", "evil.example/drop")
	pkg := map[string]string{querytpl.ParamEcosystem: "Go", querytpl.ParamPackage: "evil.example/drop"}
	for _, name := range []string{querytpl.MalByPackage, querytpl.ScorecardByPackage, querytpl.SPDXByPackage} {
		if !f.Binds(querytpl.Request{Template: name, Params: pkg}) {
			t.Errorf("%s did not bind a raised package", name)
		}
	}
	if f.Binds(querytpl.Request{Template: querytpl.OSVByPackage, Params: pkg}) {
		t.Fatal("a raised package bound osv-by-package with no affected row")
	}
	cve := map[string]string{querytpl.ParamCVEID: "CVE-2026-7777"}
	if f.Binds(querytpl.Request{Template: querytpl.CVSSByCVE, Params: cve}) {
		t.Fatal("cvss bound a CVE the check had not recorded")
	}
	f.add("Go", "evil.example/drop", Row{Aliases: "CVE-2026-7777"})
	if !f.Binds(querytpl.Request{Template: querytpl.CVSSByCVE, Params: cve}) {
		t.Fatal("cvss did not bind a CVE of an affected advisory")
	}
	if !f.Binds(querytpl.Request{Template: querytpl.EPSSByCVE, Params: cve}) {
		t.Fatal("epss did not bind the same CVE")
	}
}

// TestBindsKeysBothSidesOnTheNormalisedPyPIName proves a raised PyPI package
// binds the malicious, scorecard and licence templates when the manifest
// spelling and the model's spelling differ and PEP 503 makes them one name.
// Go is not folded. The mutated field is the package key: comparing either
// side as written turns this red.
func TestBindsKeysBothSidesOnTheNormalisedPyPIName(t *testing.T) {
	var f Facts
	f.noteRaised("PyPI", "Django_REST.framework")
	model := map[string]string{querytpl.ParamEcosystem: "PyPI", querytpl.ParamPackage: "django.rest-framework"}
	for _, name := range []string{querytpl.MalByPackage, querytpl.ScorecardByPackage, querytpl.SPDXByPackage} {
		if !f.Binds(querytpl.Request{Template: name, Params: model}) {
			t.Errorf("%s did not bind the normalised PyPI name", name)
		}
	}
	other := map[string]string{querytpl.ParamEcosystem: "PyPI", querytpl.ParamPackage: "flask"}
	if f.Binds(querytpl.Request{Template: querytpl.MalByPackage, Params: other}) {
		t.Fatal("flask bound a different package")
	}
	f.noteRaised("Go", "Example.com/Mod")
	folded := map[string]string{querytpl.ParamEcosystem: "Go", querytpl.ParamPackage: "example.com/mod"}
	if f.Binds(querytpl.Request{Template: querytpl.MalByPackage, Params: folded}) {
		t.Fatal("a Go module path was folded")
	}
	same := map[string]string{querytpl.ParamEcosystem: "Go", querytpl.ParamPackage: "Example.com/Mod"}
	if !f.Binds(querytpl.Request{Template: querytpl.MalByPackage, Params: same}) {
		t.Fatal("the Go path as written did not bind")
	}
}

// labelledQuery answers every request with one form and one text.
type labelledQuery struct {
	form string
	text string
	cols []string
	rows [][]string
}

func (q labelledQuery) Run(context.Context, querytpl.Request, querytpl.Relations) (query.Result, error) {
	rows := q.rows
	if rows == nil {
		rows = [][]string{}
	}
	cols := q.cols
	if cols == nil {
		cols = Columns
	}
	return query.Result{SQL: q.text, Form: q.form, Engine: "mem", Columns: cols, Rows: rows}, nil
}

// TestMissNoteNamesTheQueryForm proves a package with no row, and a package
// whose rows all miss, name the form of the text in the trace. Mango is not
// labelled SQL. The mutated field is the label in queryText. Restoring
// "SQL:" turns both cases red.
func TestMissNoteNamesTheQueryForm(t *testing.T) {
	root := t.TempDir()
	writeMods(t, root, "module m\n\ngo 1.26\n", "module m\n\ngo 1.26\n\nrequire example.com/mod v1.0.0\n")
	mango := `{"table":"px_osvfile.px_osv_pkg","selector":{"package":"example.com/mod"}}`
	out, err := Check(context.Background(), labelledQuery{form: querytpl.FormMango, text: mango}, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Notes) != 1 || !strings.Contains(out.Notes[0], "query (mango): "+mango) || strings.Contains(out.Notes[0], "SQL:") {
		t.Fatalf("no-rows note %#v", out.Notes)
	}
	miss := []string{"GHSA-1", "", "Go", "example.com/mod", "9.0.0", "", "", "LOW", "2026-01-15"}
	out, err = Check(context.Background(), labelledQuery{
		form: querytpl.FormSQL, text: "SELECT 1", cols: Columns, rows: [][]string{miss},
	}, root, "", "snap")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Candidates) != 0 || len(out.Notes) != 1 || !strings.Contains(out.Notes[0], "query (sql): SELECT 1") || strings.Contains(out.Notes[0], "SQL:") {
		t.Fatalf("miss note cands %d notes %#v", len(out.Candidates), out.Notes)
	}
}

// loadSQL loads the tables into the sample query engine, the one sample
// mode uses, so the check runs its real SQL with no network.
func loadSQL(t *testing.T, body string) *query.Sample {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tables.sql")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	q, err := query.OpenSample(path)
	if err != nil {
		t.Fatal(err)
	}
	return q
}
