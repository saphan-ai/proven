package querytpl_test

import (
	"testing"

	"github.com/saphan-ai/proven/internal/epss"
	"github.com/saphan-ai/proven/internal/osv"
	"github.com/saphan-ai/proven/internal/querytpl"
)

// TestOneRuleNamesAValidRelation proves the table and view validators of the OSV and EPSS settings
// give the answer of the one rule in querytpl for every name tried (ADR-0032): there is one rule, so
// the three cannot part. The mutated field is the name.
func TestOneRuleNamesAValidRelation(t *testing.T) {
	for _, name := range []string{
		"px_osvfile.px_osv_pkg", "plain", "a.b", "a.b.c", "", "1abc", "a b", "a;b", "a-b", "a.", ".b",
		"a\nb", "a\"b", "a`b", "tab\x00", "é", "a'b",
	} {
		want := querytpl.ValidRelation(name)
		if osv.ValidTable(name) != want || osv.ValidView(name) != want || epss.ValidView(name) != want {
			t.Errorf("%q: osv table %t, osv view %t, epss view %t, querytpl %t",
				name, osv.ValidTable(name), osv.ValidView(name), epss.ValidView(name), want)
		}
	}
}
