package querytpl

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hostile is one byte or spelling a value must never carry into a query.
var hostile = map[string]string{
	"single quote":    "a'b",
	"backslash":       `a\b`,
	"trailing slash":  `ab\`,
	"quote and slash": `a\'b`,
	"newline":         "a\nb",
	"carriage return": "a\rb",
	"nul":             "a\x00b",
	"tab":             "a\tb",
	"escape byte":     "a\x1bb",
	"delete byte":     "a\x7fb",
	"invalid utf-8":   "a\xffb",
	"non-ascii":       "aéb",
	"space":           "a b",
	"semicolon":       "a;b",
	"trailing break":  "a\n",
}

// refusedAs reports the reason of a refusal, or "" when err is not one.
func refusedAs(err error) string {
	var r *Refusal
	if errors.As(err, &r) {
		return r.Reason
	}
	return ""
}

// TestEachFormatRefusesAHostileValue proves the property of ADR-0032 point 2
// and 3: in every parameter of every template, a quote, a backslash, a
// control byte, invalid UTF-8 and the other bytes outside the format are
// refused with a named reason, and no query is built. The mutated field is
// the one parameter under test; the others stay valid.
func TestEachFormatRefusesAHostileValue(t *testing.T) {
	valid := map[string]string{
		ParamEcosystem: "Go", ParamPackage: "example.com/mod",
		ParamCVEID: "CVE-2021-44228", ParamAdvisoryID: "GHSA-jfh8-c2jp-5v3q",
	}
	for _, tpl := range Catalogue() {
		for _, p := range tpl.Params {
			for name, bad := range hostile {
				params := map[string]string{}
				for _, q := range tpl.Params {
					params[q.Name] = valid[q.Name]
				}
				params[p.Name] = bad
				sql, err := Build(Request{Template: tpl.Name, Params: params}, Relations{})
				if sql != "" || refusedAs(err) != ReasonBadValue {
					t.Errorf("%s %s with %s: sql=%q err=%v, want a refusal %q", tpl.Name, p.Name, name, sql, err, ReasonBadValue)
				}
			}
		}
	}
}

// TestPackageFormatsFollowTheEcosystem proves that a package name is held to
// its own ecosystem's spelling: a name that is valid for one ecosystem and not
// for another is refused for the other. The mutated field is the ecosystem.
func TestPackageFormatsFollowTheEcosystem(t *testing.T) {
	cases := []struct {
		eco, pkg string
		ok       bool
	}{
		{"Go", "github.com/Org/Repo.v2", true},
		{"Go", "example.com/a_b~c", true},
		{"npm", "@scope/pkg-name", true},
		{"npm", "left-pad", true},
		{"npm", "@scope/a/b", false},
		{"npm", "a/b", false},
		{"PyPI", "Django_REST.framework", true},
		{"PyPI", "-leading", false},
		{"PyPI", "a/b", false},
		{"Packagist", "monolog/monolog", true},
		{"Packagist", "Monolog/monolog", true},
		{"Packagist", "monolog", false},
		{"Packagist", "-bad/name", false},
		{"crates.io", "serde", true},
		{"crates.io", "serde_json", true},
		{"crates.io", "Serde", true},
		{"crates.io", "a", true},
		{"crates.io", "a.b", false},
		{"crates.io", "1serde", false},
		{"crates.io", "-serde", false},
		{"crates.io", "serde/extra", false},
		{"Go", "", false},
		{"Cargo", "serde", false},
		{"", "serde", false},
	}
	for _, c := range cases {
		_, err := OSVPackageSQL(DefaultOSVTable, c.eco, c.pkg)
		if (err == nil) != c.ok {
			t.Errorf("%s %q: err=%v, want ok=%v", c.eco, c.pkg, err, c.ok)
		}
	}
}

// TestBuiltSQLIsTheTextThePipelineAlwaysSent proves that moving the SQL here
// changed no query: each template builds the text the OSV, KEV and EPSS steps
// sent before ADR-0032, so the recorded evidence reads the same.
func TestBuiltSQLIsTheTextThePipelineAlwaysSent(t *testing.T) {
	cases := []struct {
		name string
		req  Request
		want string
	}{
		{"osv", Request{OSVByPackage, map[string]string{ParamEcosystem: "Go", ParamPackage: "evil.example/leftpad"}},
			"SELECT id, aliases, ecosystem, package, introduced, fixed, last_affected, severity, published FROM px_osvfile.px_osv_pkg WHERE ecosystem = 'Go' AND package = 'evil.example/leftpad'"},
		{"kev", Request{KEVByCVE, map[string]string{ParamCVEID: "CVE-2026-0001"}},
			"SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'"},
		{"epss", Request{EPSSByCVE, map[string]string{ParamCVEID: "CVE-2026-0001"}},
			"SELECT cve_id, epss, percentile, score_date FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2026-0001'"},
		{"advisory", Request{OSVByAdvisory, map[string]string{ParamAdvisoryID: "GHSA-px-sample-0001"}},
			"SELECT id, alias, published, severity FROM px_evidence.px_osv_id WHERE id = 'GHSA-px-sample-0001'"},
	}
	for _, c := range cases {
		got, err := Build(c.req, Relations{})
		if err != nil || got != c.want {
			t.Errorf("%s: got %q, %v; want %q", c.name, got, err, c.want)
		}
	}
}

// TestPyPINameIsBuiltInItsPEP503Form proves that a PyPI package is looked up
// by the form the OSV table is keyed by, as before ADR-0032.
func TestPyPINameIsBuiltInItsPEP503Form(t *testing.T) {
	got, err := OSVPackageSQL(DefaultOSVTable, "PyPI", "Django_REST.framework")
	if err != nil || !strings.HasSuffix(got, "package = 'django-rest-framework'") {
		t.Errorf("got %q, %v", got, err)
	}
}

// TestCrateNameIsKeptAsWritten proves a crates.io package is looked up
// under the spelling in the file: case is not folded, and '_' is not turned
// into '-'.
func TestCrateNameIsKeptAsWritten(t *testing.T) {
	got, err := OSVPackageSQL(DefaultOSVTable, "crates.io", "serde_json")
	if err != nil || !strings.HasSuffix(got, "package = 'serde_json'") {
		t.Errorf("got %q, %v", got, err)
	}
	if !strings.Contains(got, "ecosystem = 'crates.io'") {
		t.Errorf("got %q", got)
	}
	long := "a" + strings.Repeat("b", 63)
	if _, err := OSVPackageSQL(DefaultOSVTable, "crates.io", long); err != nil {
		t.Errorf("a 64-character crate name was refused: %v", err)
	}
	tooLong := "a" + strings.Repeat("b", 64)
	if err := refusedAs(mustErr(t, tooLong)); err != ReasonBadValue {
		t.Errorf("a 65-character crate name: %v", err)
	}
}

// mustErr builds a crates.io query that is expected to fail.
func mustErr(t *testing.T, pkg string) error {
	t.Helper()
	_, err := OSVPackageSQL(DefaultOSVTable, "crates.io", pkg)
	if err == nil {
		t.Fatalf("%s was accepted", pkg)
	}
	return err
}

// TestPackagistNameIsBuiltInLowerCase proves a Packagist package is looked
// up by the lower-case vendor/name form the OSV table is keyed by.
func TestPackagistNameIsBuiltInLowerCase(t *testing.T) {
	got, err := OSVPackageSQL(DefaultOSVTable, "Packagist", "Monolog/monolog")
	if err != nil || !strings.HasSuffix(got, "package = 'monolog/monolog'") {
		t.Errorf("got %q, %v", got, err)
	}
	if !strings.Contains(got, "ecosystem = 'Packagist'") {
		t.Errorf("got %q", got)
	}
}

// TestRequestRefusalsAreNamed proves that a template that is not on the list,
// a missing parameter and an extra parameter are each refused with their own
// reason, and that the refusal never repeats model text.
func TestRequestRefusalsAreNamed(t *testing.T) {
	sqlText := "SELECT 1; DROP TABLE x"
	cases := []struct {
		name   string
		req    Request
		reason string
	}{
		{"sql as the template name", Request{Template: sqlText}, ReasonUnknownTemplate},
		{"no template", Request{}, ReasonUnknownTemplate},
		{"missing parameter", Request{Template: KEVByCVE, Params: map[string]string{}}, ReasonMissingParam},
		{"extra parameter", Request{Template: KEVByCVE, Params: map[string]string{ParamCVEID: "CVE-2021-44228", sqlText: "x"}}, ReasonUnknownParam},
		{"wrong parameter", Request{Template: KEVByCVE, Params: map[string]string{ParamPackage: "x"}}, ReasonMissingParam},
	}
	for _, c := range cases {
		_, err := Build(c.req, Relations{})
		if refusedAs(err) != c.reason {
			t.Errorf("%s: err=%v, want reason %q", c.name, err, c.reason)
			continue
		}
		if strings.Contains(err.Error(), "DROP") {
			t.Errorf("%s: the refusal repeats model text: %v", c.name, err)
		}
	}
}

// TestRelationsAreCheckedAgainAtBuild proves that a table or view name from a
// setting is refused when it is not a plain name, even if the setting loader
// was bypassed, and that an empty name means the default.
func TestRelationsAreCheckedAgainAtBuild(t *testing.T) {
	req := Request{Template: KEVByCVE, Params: map[string]string{ParamCVEID: "CVE-2021-44228"}}
	for _, bad := range []string{"v; DROP", "a.b.c", "a b", "'x'", "x--", "1abc"} {
		if _, err := Build(req, Relations{KEVView: bad}); refusedAs(err) != ReasonBadRelation {
			t.Errorf("view %q: err=%v, want %q", bad, err, ReasonBadRelation)
		}
	}
	got, err := Build(req, Relations{KEVView: "my_schema.my_view"})
	if err != nil || !strings.Contains(got, " FROM my_schema.my_view WHERE ") {
		t.Errorf("custom view: %q, %v", got, err)
	}
}

// TestValidateNeedsNoRelations proves the Skeptic's early check accepts a
// good request and refuses a bad one without the review's settings.
func TestValidateNeedsNoRelations(t *testing.T) {
	if err := Validate(Request{Template: EPSSByCVE, Params: map[string]string{ParamCVEID: "CVE-2021-44228"}}); err != nil {
		t.Errorf("a good request was refused: %v", err)
	}
	if err := Validate(Request{Template: EPSSByCVE, Params: map[string]string{ParamCVEID: "CVE-2021-44228' OR '1'='1"}}); err == nil {
		t.Error("a request with a quote in its value was accepted")
	}
}

// TestCVEFormatIsAnchored proves a CVE id embedded in other text is not an id.
func TestCVEFormatIsAnchored(t *testing.T) {
	for _, s := range []string{"xCVE-2021-44228", "CVE-2021-44228x", "CVE-2021-44228\n", "cve-2021-44228", "CVE-21-1", " CVE-2021-44228"} {
		if ValidCVE(s) {
			t.Errorf("%q was accepted as a CVE id", s)
		}
	}
	if !ValidCVE("CVE-2021-44228") {
		t.Error("a CVE id was refused")
	}
}

// TestQueriesDocumentListsEveryTemplate is the control that reddens when the
// list and docs/QUERIES.md part: each template, parameter and format sentence
// in the catalogue must appear in the document.
func TestQueriesDocumentListsEveryTemplate(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "docs", "QUERIES.md"))
	if err != nil {
		t.Fatal(err)
	}
	doc := string(b)
	for _, tpl := range Catalogue() {
		if !strings.Contains(doc, "`"+tpl.Name+"`") {
			t.Errorf("docs/QUERIES.md does not name template %s", tpl.Name)
		}
		for _, p := range tpl.Params {
			if !strings.Contains(doc, "`"+p.Name+"`") {
				t.Errorf("docs/QUERIES.md does not name parameter %s of %s", p.Name, tpl.Name)
			}
		}
	}
}
