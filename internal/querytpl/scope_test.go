package querytpl

import (
	"strings"
	"testing"
)

// TestNilScopeIsEveryTemplate proves that a review with no setting is the
// behaviour before scopes existed: the nil scope allows and describes every
// template. Mutated field: the scope, nil against a narrowed one.
func TestNilScopeIsEveryTemplate(t *testing.T) {
	var s Scope
	for _, tpl := range Catalogue() {
		if !s.Allows(tpl.Name) {
			t.Errorf("the nil scope refuses %s", tpl.Name)
		}
	}
	if s.Describe() != Describe() {
		t.Fatal("the nil scope describes something other than the whole list")
	}
	if err := s.Check(RoleSkeptic, Request{Template: EPSSByCVE}); err != nil {
		t.Fatalf("the nil scope refused a request: %v", err)
	}
}

// TestScopeDescribeListsOnlyItsTemplates proves a role's prompt text names the
// templates its scope allows and no other. Mutated field: the scope's names.
func TestScopeDescribeListsOnlyItsTemplates(t *testing.T) {
	got := Scope{KEVByCVE}.Describe()
	if !strings.Contains(got, KEVByCVE+" (") {
		t.Fatalf("the allowed template is missing: %s", got)
	}
	for _, tpl := range Catalogue() {
		if tpl.Name != KEVByCVE && strings.Contains(got, tpl.Name) {
			t.Errorf("the description names %s, which the scope does not allow", tpl.Name)
		}
	}
}

// TestScopeCheckRefusesByName proves the three answers of Check: a template in
// scope passes, a listed template outside scope is refused as out of scope
// with its name and the role, and a name that is not on the list is refused as
// unknown without repeating the model's text. Mutated field: the request's
// template name.
func TestScopeCheckRefusesByName(t *testing.T) {
	s := Scope{KEVByCVE}
	if err := s.Check(RoleSkeptic, Request{Template: KEVByCVE}); err != nil {
		t.Fatalf("a template in scope was refused: %v", err)
	}
	err := s.Check(RoleSkeptic, Request{Template: EPSSByCVE})
	ref, ok := err.(*Refusal)
	if !ok || ref.Reason != ReasonOutOfScope || ref.Detail != EPSSByCVE+" for the "+RoleSkeptic {
		t.Fatalf("out of scope: %#v", err)
	}
	err = s.Check(RoleInvestigator, Request{Template: "drop-everything"})
	ref, ok = err.(*Refusal)
	if !ok || ref.Reason != ReasonUnknownTemplate || strings.Contains(ref.Error(), "drop-everything") {
		t.Fatalf("unknown: %#v", err)
	}
}

// TestParseScope proves the loader's three refusals and its one success.
// Mutated field: the list text.
func TestParseScope(t *testing.T) {
	got, bad, why := ParseScope(" kev-by-cve , osv-by-package ")
	if why != "" || len(got) != 2 || got[0] != KEVByCVE || got[1] != OSVByPackage {
		t.Fatalf("%v %q %q", got, bad, why)
	}
	cases := []struct{ list, bad, why string }{
		{"kev-by-cve, nope", "nope", "unknown template"},
		{"kev-by-cve, kev-by-cve", "kev-by-cve", "duplicate template"},
		{"kev-by-cve,", "", "empty entry"},
		{"", "", "empty entry"},
	}
	for _, c := range cases {
		if _, bad, why := ParseScope(c.list); bad != c.bad || why != c.why {
			t.Errorf("%q: got %q %q, want %q %q", c.list, bad, why, c.bad, c.why)
		}
	}
}
