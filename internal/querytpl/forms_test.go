package querytpl

import (
	"reflect"
	"strings"
	"testing"
)

// formRequests is one valid request for each template.
var formRequests = []Request{
	{Template: OSVByPackage, Params: map[string]string{ParamEcosystem: "PyPI", ParamPackage: "Django_Rest"}},
	{Template: OSVByAdvisory, Params: map[string]string{ParamAdvisoryID: "GHSA-aaaa-bbbb-cccc"}},
	{Template: KEVByCVE, Params: map[string]string{ParamCVEID: "CVE-2021-44228"}},
	{Template: EPSSByCVE, Params: map[string]string{ParamCVEID: "CVE-2021-44228"}},
	{Template: MalByPackage, Params: map[string]string{ParamEcosystem: "Go", ParamPackage: "evil.example/drop"}},
	{Template: ScorecardByPackage, Params: map[string]string{ParamEcosystem: "Go", ParamPackage: "evil.example/young"}},
	{Template: CVSSByCVE, Params: map[string]string{ParamCVEID: "CVE-2026-7777"}},
	{Template: SPDXByPackage, Params: map[string]string{ParamEcosystem: "Go", ParamPackage: "evil.example/copyleft"}},
}

// selectList returns the column names of a SELECT built by this package.
func selectList(sql string) []string {
	head := strings.TrimPrefix(sql, "SELECT ")
	head = head[:strings.Index(head, " FROM ")]
	return strings.Split(head, ", ")
}

// TestBuildFormsFillsBothForms proves every template carries the sql text
// Build returns and a mango form whose fields are the columns of that sql
// text, in the same order (ADR-0047). The mutated field is the field list of
// one template: reordering it makes the comparison fail.
func TestBuildFormsFillsBothForms(t *testing.T) {
	for _, req := range formRequests {
		t.Run(req.Template, func(t *testing.T) {
			sql, err := Build(req, Relations{})
			if err != nil {
				t.Fatal(err)
			}
			forms, err := BuildForms(req, Relations{})
			if err != nil {
				t.Fatal(err)
			}
			if forms.SQL != sql || forms.SQL == "" {
				t.Fatalf("sql form %q, build %q", forms.SQL, sql)
			}
			if forms.Mango == "" || !strings.HasPrefix(forms.Mango, `{"table":`) {
				t.Fatalf("mango form %q", forms.Mango)
			}
			if !reflect.DeepEqual(forms.Find.Fields, selectList(forms.SQL)) {
				t.Fatalf("mango fields %v, sql columns %v", forms.Find.Fields, selectList(forms.SQL))
			}
			if forms.Find.Table == "" || len(forms.Find.Selector) == 0 {
				t.Fatalf("find %#v", forms.Find)
			}
		})
	}
}

// TestMangoSelectorHoldsTheValuesTheSQLFormMatches proves the selector carries
// the validated values, in the spelling the sql form uses: a PyPI name is in
// its PEP 503 form, and the relation names come from the settings. The
// mutated fields are the package name and the KEV view.
func TestMangoSelectorHoldsTheValuesTheSQLFormMatches(t *testing.T) {
	forms, err := BuildForms(formRequests[0], Relations{})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"ecosystem": "PyPI", "package": "django-rest"}
	if !reflect.DeepEqual(forms.Find.Selector, want) || !strings.Contains(forms.SQL, "package = 'django-rest'") {
		t.Fatalf("selector %v, sql %q", forms.Find.Selector, forms.SQL)
	}
	if forms.Mango != `{"table":"px_osvfile.px_osv_pkg","selector":{"ecosystem":"PyPI","package":"django-rest"},"fields":["id","aliases","ecosystem","package","introduced","fixed","last_affected","severity","published"]}` {
		t.Fatalf("mango form %q", forms.Mango)
	}
	forms, err = BuildForms(formRequests[2], Relations{KEVView: "my_kev"})
	if err != nil || forms.Find.Table != "my_kev" {
		t.Fatalf("%#v %v", forms, err)
	}
}

// TestBuildFormsRefusesWhatBuildRefuses proves a refused request fills no
// form, and a value that is outside its format reaches neither language
// (ADR-0032, ADR-0047). The mutated fields are the template name, a CVE id
// that carries a quote and a brace, and a relation name that is not plain.
func TestBuildFormsRefusesWhatBuildRefuses(t *testing.T) {
	forms, err := BuildForms(Request{Template: "not-a-template"}, Relations{})
	if forms.SQL != "" || forms.Mango != "" || refusedAs(err) != ReasonUnknownTemplate {
		t.Fatalf("forms %#v err %v", forms, err)
	}
	bad := Request{Template: KEVByCVE, Params: map[string]string{ParamCVEID: `CVE-2021-1"} {"$or":[`}}
	forms, err = BuildForms(bad, Relations{})
	if forms.SQL != "" || forms.Mango != "" || refusedAs(err) != ReasonBadValue {
		t.Fatalf("forms %#v err %v", forms, err)
	}
	// The mango builder is reachable on its own: it must refuse by itself,
	// not lean on the sql check having run first.
	if _, err := buildFind(KEVByCVE, bad.Params, Relations{}.resolved()); refusedAs(err) != ReasonBadValue {
		t.Fatalf("mango builder accepted a bad value: %v", err)
	}
	_, err = BuildForms(formRequests[2], Relations{KEVView: "a.b.c"})
	if refusedAs(err) != ReasonBadRelation {
		t.Fatalf("relation: %v", err)
	}
	if _, err := buildFind(KEVByCVE, formRequests[2].Params, Relations{KEVView: "a b"}); refusedAs(err) != ReasonBadRelation {
		t.Fatalf("mango builder accepted a bad relation: %v", err)
	}
}
