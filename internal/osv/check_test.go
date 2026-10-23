package osv

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
)

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

type downQuery struct{}

func (downQuery) Query(context.Context, string) (query.Result, error) {
	return query.Result{}, &service.Down{Service: "query", Reason: "unreachable"}
}

func asConfig(err error, target **ConfigError) bool {
	cfg, ok := err.(*ConfigError)
	if ok {
		*target = cfg
	}
	return ok
}

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
