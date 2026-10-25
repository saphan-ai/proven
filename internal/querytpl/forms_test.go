package querytpl

import "testing"

// TestBuildFormsNamesTheSQLFormAndLeavesMangoEmpty proves a built template
// carries the sql text Build returns, and the mango form is empty because
// CouchDB is not built (ADR-0047). The mutated field is Forms.Mango: a
// non-empty value would mean a selector was built for a store that does
// not exist yet.
func TestBuildFormsNamesTheSQLFormAndLeavesMangoEmpty(t *testing.T) {
	req := Request{Template: KEVByCVE, Params: map[string]string{ParamCVEID: "CVE-2021-44228"}}
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
	if forms.Mango != "" {
		t.Fatalf("mango form %q, want empty until CouchDB", forms.Mango)
	}
	// A refused request fills no form. Mutated field: the template name.
	forms, err = BuildForms(Request{Template: "not-a-template"}, Relations{})
	if forms.SQL != "" || forms.Mango != "" || refusedAs(err) != ReasonUnknownTemplate {
		t.Fatalf("forms %#v err %v", forms, err)
	}
}
