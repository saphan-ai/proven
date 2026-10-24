package query_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
)

// TestLocalFileKeepsARegisteredTable proves the local store writes a table
// into one file and returns the same rows after the file is opened again,
// including a cell that contains a quote. The mutated field is the cell
// text: a quote that was not doubled would fail to load.
func TestLocalFileKeepsARegisteredTable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evidence.sql")
	f, err := query.CreateFile(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("mode %o is readable by group or others", info.Mode().Perm())
	}
	when := time.Date(2026, 10, 24, 7, 0, 0, 0, time.UTC)
	f.SetClock(func() time.Time { return when })
	cols := []string{"cve_id", "note"}
	rows := [][]string{{"CVE-2021-44228", "it's a quote"}}
	if err := f.Register("px_evidence.px_kev_vulns", cols, rows); err != nil {
		t.Fatal(err)
	}
	if err := f.Register("px_evidence.px_kev_vulns", cols, rows); err == nil {
		t.Fatal("a second register of the same table was accepted")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(body), "CREATE TABLE") != 1 {
		t.Fatalf("the refused register changed the file:\n%s", body)
	}
	got, err := f.Query(context.Background(), "SELECT cve_id, note FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2021-44228'")
	if err != nil {
		t.Fatal(err)
	}
	if got.Form != querytpl.FormSQL || got.Engine != "local" || got.Snapshot != "2026-10-24T07:00:00Z" {
		t.Fatalf("%#v", got)
	}
	if len(got.Rows) != 1 || got.Rows[0][1] != "it's a quote" {
		t.Fatalf("rows %#v", got.Rows)
	}
	again, err := query.CreateFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err = again.Query(context.Background(), "SELECT note FROM px_evidence.px_kev_vulns")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) != 1 || got.Rows[0][0] != "it's a quote" {
		t.Fatalf("reopened rows %#v", got.Rows)
	}
}

// TestOpenSampleIsTheLocalFile proves a fixture file opened for sample mode
// is the local store: the same rows, the sql form, the engine name
// local-sample, and an empty snapshot so the caller can fill the fixture
// date. The mutated field is the engine name.
func TestOpenSampleIsTheLocalFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tables.sql")
	text := "CREATE TABLE advisories (package TEXT);\nINSERT INTO advisories VALUES ('evil.example/leftpad');\n"
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := query.OpenSample(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Query(context.Background(), "SELECT package FROM advisories")
	if err != nil {
		t.Fatal(err)
	}
	if got.Form != querytpl.FormSQL || got.Engine != "local-sample" || got.Snapshot != "" {
		t.Fatalf("%#v", got)
	}
	if len(got.Rows) != 1 || got.Rows[0][0] != "evil.example/leftpad" {
		t.Fatalf("rows %#v", got.Rows)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != text {
		t.Fatal("sample mode rewrote the fixture file")
	}
}
