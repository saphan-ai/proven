package query_test

import (
	"context"
	"testing"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
)

// sampleSQL is one advisory row, in the form a fixture's tables.sql uses.
const sampleSQL = `
CREATE TABLE advisories (package TEXT, version TEXT, cve TEXT, kev TEXT);
INSERT INTO advisories VALUES ('evil.example/leftpad', '1.2.3', 'CVE-2026-0001', 'true');
`

// TestLocalEngineReturnsTheSampleRow proves the local engine returns the
// inserted row for a matching WHERE, and no row for another package.
func TestLocalEngineReturnsTheSampleRow(t *testing.T) {
	db, err := query.LoadSQL(sampleSQL)
	if err != nil {
		t.Fatal(err)
	}
	cols, rows, err := db.Query("SELECT package, version, cve, kev FROM advisories WHERE package = 'evil.example/leftpad' AND version = '1.2.3'")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows %d", len(rows))
	}
	if rows[0][0] != "evil.example/leftpad" || rows[0][2] != "CVE-2026-0001" || cols[3] != "kev" {
		t.Fatalf("got %#v %#v", cols, rows)
	}
	_, empty, err := db.Query("SELECT package FROM advisories WHERE package = 'billing/note' AND version = 'change'")
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("want no rows, got %#v", empty)
	}
}

// TestQualifiedViewAndInnerJoin proves schema.table names and an aliased
// inner join work, and that the join drops a CVE missing from one side.
func TestQualifiedViewAndInnerJoin(t *testing.T) {
	const sql = `
CREATE TABLE px_evidence.px_kev_vulns (cve_id TEXT, vendor_project TEXT, product TEXT, date_added TEXT, known_ransomware TEXT);
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-0001', 'evil.example', 'leftpad', '2026-01-15', 'Known');
INSERT INTO px_evidence.px_kev_vulns VALUES ('CVE-2026-0002', 'other.example', 'widget', '2026-02-01', 'Unknown');
CREATE TABLE px_evidence.px_epss_cve_rows (cve_id TEXT, epss TEXT, percentile TEXT, score_date TEXT);
INSERT INTO px_evidence.px_epss_cve_rows VALUES ('CVE-2026-0001', '0.50000', '0.90000', '2026-10-20');
`
	db, err := query.LoadSQL(sql)
	if err != nil {
		t.Fatal(err)
	}
	cols, rows, err := db.Query("SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0][0] != "CVE-2026-0001" || cols[1] != "vendor_project" {
		t.Fatalf("kev %#v %#v", cols, rows)
	}
	join := "SELECT k.cve_id, k.vendor_project, e.epss, e.percentile FROM px_evidence.px_kev_vulns k JOIN px_evidence.px_epss_cve_rows e ON k.cve_id = e.cve_id WHERE k.cve_id = 'CVE-2026-0001'"
	_, hit, err := db.Query(join)
	if err != nil {
		t.Fatal(err)
	}
	if len(hit) != 1 || hit[0][0] != "CVE-2026-0001" || hit[0][2] != "0.50000" {
		t.Fatalf("join hit %#v", hit)
	}
	// Mutated field: the CVE in the WHERE clause. CVE-2026-0002 is on the KEV
	// sample and absent from the EPSS sample, so the inner join drops it.
	missSQL := "SELECT k.cve_id, e.epss FROM px_evidence.px_kev_vulns k JOIN px_evidence.px_epss_cve_rows e ON k.cve_id = e.cve_id WHERE k.cve_id = 'CVE-2026-0002'"
	_, miss, err := db.Query(missSQL)
	if err != nil {
		t.Fatal(err)
	}
	if len(miss) != 0 {
		t.Fatalf("absent CVE returned %#v", miss)
	}
}

// TestLiveQueryUnconfiguredNamesZetaris proves a Live with no settings fails
// as the zetaris service, with no panic.
func TestLiveQueryUnconfiguredNamesZetaris(t *testing.T) {
	_, err := (&query.Live{}).Query(context.Background(), "SELECT package FROM advisories")
	d, ok := service.As(err)
	if !ok || d.Service != "zetaris" || d.Reason != "not configured" {
		t.Fatalf("got %v", err)
	}
}
