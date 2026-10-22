package osv

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
