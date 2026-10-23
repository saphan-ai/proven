package osv

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadDirWritesOneRowPerRange proves one row per affected range in file
// order: a fixed range, an open range, a last_affected range, two ranges in
// one entry, a versions list as one-version ranges, severity from the most
// specific place, sorted aliases, a non-JSON file skipped, and the CSV header.
func TestReadDirWritesOneRowPerRange(t *testing.T) {
	rows, err := ReadDir(filepath.Join(repoRoot(t), "testdata", "osv"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 7 {
		t.Fatalf("rows %d", len(rows))
	}
	want := []Row{
		{ID: "GO-2026-1001", Aliases: "CVE-2024-1111 GHSA-zz", Ecosystem: "Go", Package: "example.com/widget", Introduced: "1.0.0", Fixed: "1.2.3", Severity: "HIGH", Published: "2026-01-02T03:04:05Z"},
		{ID: "GO-2026-1002", Ecosystem: "Go", Package: "example.com/only", Introduced: "0.1.0", Published: "2026-02-01T00:00:00Z"},
		{ID: "GO-2026-1003", Ecosystem: "Go", Package: "example.com/last", Introduced: "1.0.0", LastAffected: "1.4.2", Severity: "MODERATE", Published: "2026-03-01T00:00:00Z"},
		{ID: "GO-2026-1004", Ecosystem: "Go", Package: "example.com/two", Introduced: "0", Fixed: "1.2.3", Published: "2026-04-01T00:00:00Z"},
		{ID: "GO-2026-1004", Ecosystem: "Go", Package: "example.com/two", Introduced: "1.4.0", LastAffected: "1.4.2", Published: "2026-04-01T00:00:00Z"},
		{ID: "GO-2026-1005", Ecosystem: "npm", Package: "left-pad", Introduced: "1.0.0", LastAffected: "1.0.0", Severity: "CVSS:3.1/AV:N", Published: "2026-05-01T00:00:00Z"},
		{ID: "GO-2026-1005", Ecosystem: "npm", Package: "left-pad", Introduced: "1.0.1", LastAffected: "1.0.1", Severity: "CVSS:3.1/AV:N", Published: "2026-05-01T00:00:00Z"},
	}
	for i, row := range want {
		if rows[i] != row {
			t.Fatalf("row %d\n got %#v\nwant %#v", i, rows[i], row)
		}
	}
	var buf bytes.Buffer
	if err := WriteCSV(&buf, rows); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(buf.String(), strings.Join(Columns, ",")+"\n") {
		t.Fatalf("header %q", buf.String())
	}
	if !strings.Contains(buf.String(), "CVE-2024-1111 GHSA-zz") {
		t.Fatal(buf.String())
	}
}

// TestReadDirSkipsGitRangeBesideSemver proves a GIT range is not a version
// row when a SEMVER range is also present. Commit hashes must not appear.
func TestReadDirSkipsGitRangeBesideSemver(t *testing.T) {
	dir := t.TempDir()
	body := `{
		"id": "GO-2026-2001",
		"published": "2026-06-01T00:00:00Z",
		"affected": [{
			"package": {"ecosystem": "Go", "name": "example.com/mixed"},
			"ranges": [
				{"type": "GIT", "events": [{"introduced": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, {"fixed": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}]},
				{"type": "SEMVER", "events": [{"introduced": "1.0.0"}, {"fixed": "1.2.0"}]}
			]
		}]
	}`
	if err := os.WriteFile(filepath.Join(dir, "mixed.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows %d, want 1 SEMVER row: %#v", len(rows), rows)
	}
	want := Row{ID: "GO-2026-2001", Ecosystem: "Go", Package: "example.com/mixed", Introduced: "1.0.0", Fixed: "1.2.0", Published: "2026-06-01T00:00:00Z"}
	if rows[0] != want {
		t.Fatalf("got %#v\nwant %#v", rows[0], want)
	}
}

// TestReadDirKeepsVersionsWhenRangesAreOnlyGit proves an entry whose ranges
// are only GIT still yields one row per versions entry. The versions list
// is the affected set when no version range remains.
func TestReadDirKeepsVersionsWhenRangesAreOnlyGit(t *testing.T) {
	dir := t.TempDir()
	body := `{
		"id": "GO-2026-2002",
		"published": "2026-06-02T00:00:00Z",
		"affected": [{
			"package": {"ecosystem": "Go", "name": "example.com/gitonly"},
			"ranges": [
				{"type": "GIT", "events": [{"introduced": "cccccccccccccccccccccccccccccccccccccccc"}, {"fixed": "dddddddddddddddddddddddddddddddddddddddd"}]}
			],
			"versions": ["2.0.0", "2.1.0"]
		}]
	}`
	if err := os.WriteFile(filepath.Join(dir, "gitonly.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []Row{
		{ID: "GO-2026-2002", Ecosystem: "Go", Package: "example.com/gitonly", Introduced: "2.0.0", LastAffected: "2.0.0", Published: "2026-06-02T00:00:00Z"},
		{ID: "GO-2026-2002", Ecosystem: "Go", Package: "example.com/gitonly", Introduced: "2.1.0", LastAffected: "2.1.0", Published: "2026-06-02T00:00:00Z"},
	}
	if len(rows) != len(want) {
		t.Fatalf("rows %d, want %d: %#v", len(rows), len(want), rows)
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Fatalf("row %d\n got %#v\nwant %#v", i, rows[i], want[i])
		}
	}
}

// TestReadDirKeepsEcosystemRangeBesideGit proves an ECOSYSTEM range is a
// version row and a GIT range beside it is not.
func TestReadDirKeepsEcosystemRangeBesideGit(t *testing.T) {
	dir := t.TempDir()
	body := `{
		"id": "GO-2026-2003",
		"published": "2026-06-03T00:00:00Z",
		"affected": [{
			"package": {"ecosystem": "Go", "name": "example.com/eco"},
			"ranges": [
				{"type": "GIT", "events": [{"introduced": "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"}]},
				{"type": "ECOSYSTEM", "events": [{"introduced": "1.5.0"}, {"last_affected": "1.5.2"}]}
			]
		}]
	}`
	if err := os.WriteFile(filepath.Join(dir, "eco.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := Row{ID: "GO-2026-2003", Ecosystem: "Go", Package: "example.com/eco", Introduced: "1.5.0", LastAffected: "1.5.2", Published: "2026-06-03T00:00:00Z"}
	if len(rows) != 1 || rows[0] != want {
		t.Fatalf("got %#v\nwant %#v", rows, want)
	}
}

// TestReadDirRefusesAnAdvisoryWithNoID proves that an advisory with no id
// refuses the run and the error names the file.
func TestReadDirRefusesAnAdvisoryWithNoID(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{"affected":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ReadDir(dir)
	if err == nil || !strings.Contains(err.Error(), "bad.json") || !strings.Contains(err.Error(), "no id") {
		t.Fatal(err)
	}
}

// repoRoot finds the module root, where testdata/ lives.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
