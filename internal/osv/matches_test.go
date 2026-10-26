package osv

import (
	"context"
	"testing"
)

// TestCheckReportsEveryRowItComparedAsAMatch proves that Outcome.Matches
// keeps, for a package whose query returned rows, each row with the verdict of
// its comparison, the KEV aliases, and the query as the store ran it, and that
// a package with no row has no entry.
func TestCheckReportsEveryRowItComparedAsAMatch(t *testing.T) {
	root := t.TempDir()
	writeMods(t, root,
		"module m\n\ngo 1.26\n",
		"module m\n\ngo 1.26\n\nrequire (\n\texample.com/mod v1.2.3\n\texample.com/none v1.0.0\n)\n",
	)
	q := loadSQL(t, `
CREATE TABLE px_osvfile.px_osv_pkg (id TEXT, aliases TEXT, ecosystem TEXT, package TEXT, introduced TEXT, fixed TEXT, last_affected TEXT, severity TEXT, published TEXT);
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-1', 'CVE-2026-0001', 'Go', 'example.com/mod', '1.0.0', '1.2.4', '', 'HIGH', '2026-01-15');
INSERT INTO px_osvfile.px_osv_pkg VALUES ('GHSA-2', '', 'Go', 'example.com/mod', '2.0.0', '2.1.0', '', 'LOW', '2026-03-01');
CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-0001', 'example', 'mod', '2026-01-15', 'Known');
`)
	out, err := Check(context.Background(), q, root, "", "snap-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Matches) != 1 {
		t.Fatalf("matches %#v, want one entry: the package with rows", out.Matches)
	}
	m := out.Matches[0]
	if m.Package != "example.com/mod" || m.Version != "v1.2.3" || m.File != "go.mod" || m.Ecosystem != "Go" {
		t.Fatalf("match %#v", m)
	}
	if len(m.Advisories) != 2 || !m.Advisories[0].Affected || m.Advisories[1].Affected {
		t.Fatalf("advisories %#v, want GHSA-1 affected and GHSA-2 not", m.Advisories)
	}
	if m.Advisories[0].Row.ID != "GHSA-1" || m.Advisories[1].Row.ID != "GHSA-2" {
		t.Fatalf("advisories %#v", m.Advisories)
	}
	if len(m.KEV) != 1 || m.KEV[0] != "CVE-2026-0001" {
		t.Fatalf("kev %#v", m.KEV)
	}
	if m.Query == "" || m.Form != "sql" || m.Snapshot != "snap-1" {
		t.Fatalf("query %q form %q snapshot %q", m.Query, m.Form, m.Snapshot)
	}
}
