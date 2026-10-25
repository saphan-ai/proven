package review

import (
	"errors"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/querytpl"
)

// TestEvidenceScopeKeysLoad proves the two scope keys read a template list
// into each role's scope, and that a file that names neither leaves both
// scopes nil, which is every template. Mutated field: the keys in the file.
func TestEvidenceScopeKeysLoad(t *testing.T) {
	cfg := mustParse(t, "investigator_templates: kev-by-cve, osv-by-package, epss-by-cve\nskeptic_templates: kev-by-cve\n")
	if len(cfg.InvestigatorScope) != 3 || len(cfg.SkepticScope) != 1 || cfg.SkepticScope[0] != querytpl.KEVByCVE {
		t.Fatalf("investigator %v, skeptic %v", cfg.InvestigatorScope, cfg.SkepticScope)
	}
	none := mustParse(t, "fixes: on\n")
	if none.InvestigatorScope != nil || none.SkepticScope != nil {
		t.Fatalf("an unset key narrowed a role: %v %v", none.InvestigatorScope, none.SkepticScope)
	}
}

// TestEvidenceScopeIsRefusedOnLoadByName proves a template that is not on the
// list, a template twice, an empty list and a role that has no query path are
// each refused on load, naming the field and the value. Mutated field: the
// line of the file.
func TestEvidenceScopeIsRefusedOnLoadByName(t *testing.T) {
	cases := []struct{ file, field, value, reason string }{
		{"skeptic_templates: kev-by-cve, nope\n", "skeptic_templates", "nope", "unknown template"},
		{"investigator_templates: kev-by-cve, kev-by-cve\n", "investigator_templates", "kev-by-cve", "duplicate template"},
		{"skeptic_templates:\n", "skeptic_templates", "", "empty entry"},
		{"planner_templates: kev-by-cve\n", "planner_templates", "kev-by-cve", "unknown key"},
	}
	for _, c := range cases {
		_, err := parseConfig([]byte(c.file))
		var ce *ConfigError
		if !errors.As(err, &ce) || ce.Field != c.field || ce.Value != c.value || ce.Reason != c.reason {
			t.Errorf("%q: %v", c.file, err)
			continue
		}
		if c.reason != "unknown key" && !strings.Contains(err.Error(), "epss-by-cve") {
			t.Errorf("%q: the refusal does not list the templates allowed: %v", c.file, err)
		}
	}
}
