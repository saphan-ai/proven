package source_test

import (
	"testing"

	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/source"
)

// TestRegistryOrderIsOSVThenKEVThenEPSS proves the built-in registry is a
// fixed order and that Catalogue follows it. The mutated field is the order
// of the three adapters. An added test adapter is not part of this list.
func TestRegistryOrderIsOSVThenKEVThenEPSS(t *testing.T) {
	got := source.All()
	want := []string{"osv", "kev", "epss"}
	if len(got) != len(want) {
		t.Fatalf("registry has %d adapters, want %d", len(got), len(want))
	}
	for i, id := range want {
		if got[i].ID() != id {
			t.Errorf("registry[%d] = %s, want %s", i, got[i].ID(), id)
		}
		load := got[i].Load()
		if load.Origin == "" || load.Refresh == "" || load.Table == "" {
			t.Errorf("%s: load is missing origin, refresh or table: %+v", id, load)
		}
	}
	cat := querytpl.Catalogue()
	wantNames := []string{querytpl.OSVByPackage, querytpl.OSVByAdvisory, querytpl.KEVByCVE, querytpl.EPSSByCVE}
	if len(cat) != len(wantNames) {
		t.Fatalf("catalogue has %d templates, want %d", len(cat), len(wantNames))
	}
	owners := []string{"osv", "osv", "kev", "epss"}
	for i, name := range wantNames {
		if cat[i].Name != name {
			t.Errorf("catalogue[%d] = %s, want %s", i, cat[i].Name, name)
		}
		ad, spec, ok := source.Lookup(name)
		if !ok || ad.ID() != owners[i] || spec.Template.Name != name || spec.Build == nil {
			t.Errorf("template %s: owner %v spec %s ok %v", name, ad, spec.Template.Name, ok)
		}
	}
}

// TestBuiltinRulesKeepRowsAndDropNone proves each built-in rule: no row is
// nothing, and a row is evidence that still carries the query. The mutated
// field is the row list.
func TestBuiltinRulesKeepRowsAndDropNone(t *testing.T) {
	for _, id := range []string{"osv", "kev", "epss"} {
		ad, ok := source.ByID(id)
		if !ok {
			t.Fatalf("adapter %s is not registered", id)
		}
		if _, ruled := ad.Rule(source.Evidence{SQL: "SELECT 1", Rows: nil}); ruled {
			t.Errorf("%s: no rows were treated as evidence", id)
		}
		got, ruled := ad.Rule(source.Evidence{
			SQL: "SELECT 1", Form: "sql", Engine: "local", Snapshot: "2026-10-24T00:00:00Z",
			Columns: []string{"id"}, Rows: [][]string{{"row-1"}},
		})
		if !ruled || got.SQL != "SELECT 1" || len(got.Rows) != 1 || got.Rows[0][0] != "row-1" || got.Snapshot != "2026-10-24T00:00:00Z" {
			t.Errorf("%s: ruled %v evidence %+v", id, ruled, got)
		}
	}
}
