package querytpl

import (
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/source"
)

// TestADR0039SQLIsBuiltFromTheTemplates proves each new template names its
// own table and the value it was given. The mutated field is the package
// or the CVE id.
func TestADR0039SQLIsBuiltFromTheTemplates(t *testing.T) {
	cases := []struct {
		name string
		req  Request
		want string
	}{
		{"mal", Request{MalByPackage, map[string]string{ParamEcosystem: "Go", ParamPackage: "evil.example/drop"}},
			"SELECT id, ecosystem, package, summary FROM px_evidence.px_mal_pkg WHERE ecosystem = 'Go' AND package = 'evil.example/drop'"},
		{"scorecard", Request{ScorecardByPackage, map[string]string{ParamEcosystem: "npm", ParamPackage: "left-pad"}},
			"SELECT ecosystem, package, maintained, signed, release_age_days FROM px_evidence.px_scorecard WHERE ecosystem = 'npm' AND package = 'left-pad'"},
		{"cvss", Request{CVSSByCVE, map[string]string{ParamCVEID: "CVE-2026-7777"}},
			"SELECT cve_id, cvss_score, severity, source FROM px_evidence.px_cvss_cve WHERE cve_id = 'CVE-2026-7777'"},
		{"spdx", Request{SPDXByPackage, map[string]string{ParamEcosystem: "PyPI", ParamPackage: "Django_REST"}},
			"SELECT ecosystem, package, licence FROM px_evidence.px_spdx_lic WHERE ecosystem = 'PyPI' AND package = 'django-rest'"},
	}
	for _, c := range cases {
		got, err := Build(c.req, Relations{})
		if err != nil || got != c.want {
			t.Errorf("%s: got %q, %v; want %q", c.name, got, err, c.want)
		}
	}
}

// TestADR0039Rules proves what each rule keeps. An empty row list is not
// evidence. A malicious id must start with MAL-. A healthy project and an
// allowed licence are not evidence. A CVSS row is. The mutated field is the
// row list of the case.
func TestADR0039Rules(t *testing.T) {
	mal, _ := source.ByID("mal")
	score, _ := source.ByID("scorecard")
	cvss, _ := source.ByID("cvss")
	spdx, _ := source.ByID("spdx")
	if mal == nil || score == nil || cvss == nil || spdx == nil {
		t.Fatal("an ADR-0039 adapter is not registered")
	}
	for _, ad := range []source.Adapter{mal, score, cvss, spdx} {
		if _, ok := ad.Rule(source.Evidence{SQL: "SELECT 1"}); ok {
			t.Errorf("%s: an empty row list was evidence", ad.ID())
		}
	}
	kept, ok := mal.Rule(source.Evidence{
		SQL: "SELECT 1", Columns: malColumns,
		Rows: [][]string{{"MAL-2024-1001", "Go", "evil.example/drop", "published to take credentials"}},
	})
	if !ok || len(kept.Rows) != 1 || !strings.HasPrefix(kept.Rows[0][0], "MAL-") {
		t.Fatalf("malicious row: ok=%v %+v", ok, kept.Rows)
	}
	if _, ok := mal.Rule(source.Evidence{
		SQL: "SELECT 1", Columns: malColumns,
		Rows: [][]string{{"GHSA-not-malicious", "Go", "evil.example/drop", "ordinary advisory"}},
	}); ok {
		t.Fatal("a row whose id does not start with MAL- was evidence")
	}
	if _, ok := score.Rule(source.Evidence{
		SQL: "SELECT 1", Columns: scorecardColumns,
		Rows: [][]string{{"Go", "example.com/ok", "10", "true", "400"}},
	}); ok {
		t.Fatal("a maintained, signed, old release was evidence")
	}
	for _, row := range [][]string{
		{"Go", "evil.example/young", "0", "true", "400"},
		{"Go", "evil.example/young", "10", "false", "400"},
		{"Go", "evil.example/young", "10", "true", "2"},
	} {
		got, ok := score.Rule(source.Evidence{SQL: "SELECT 1", Columns: scorecardColumns, Rows: [][]string{row}})
		if !ok || len(got.Rows) != 1 {
			t.Fatalf("unhealthy %v: ok=%v", row, ok)
		}
	}
	if _, ok := score.Rule(source.Evidence{
		SQL: "SELECT 1", Columns: scorecardColumns,
		Rows: [][]string{{"Go", "evil.example/young", "10", "true", "14"}},
	}); ok {
		t.Fatal("a release of 14 days was treated as only days old")
	}
	got, ok := cvss.Rule(source.Evidence{
		SQL: "SELECT 1", Columns: cvssColumns,
		Rows: [][]string{{"CVE-2026-7777", "9.8", "CRITICAL", "nvd"}},
	})
	if !ok || got.Rows[0][1] != "9.8" {
		t.Fatalf("cvss: ok=%v %+v", ok, got.Rows)
	}
	if _, ok := spdx.Rule(source.Evidence{
		SQL: "SELECT 1", Columns: spdxColumns,
		Rows: [][]string{{"Go", "example.com/ok", "MIT"}},
	}); ok {
		t.Fatal("MIT was evidence")
	}
	got, ok = spdx.Rule(source.Evidence{
		SQL: "SELECT 1", Columns: spdxColumns,
		Rows: [][]string{{"Go", "evil.example/copyleft", "GPL-3.0-only"}},
	})
	if !ok || got.Rows[0][2] != "GPL-3.0-only" {
		t.Fatalf("spdx: ok=%v %+v", ok, got.Rows)
	}
}
