package source_test

import (
	"testing"

	"github.com/saphan-ai/proven/internal/source"
)

// tableAdapter is a source whose only unusual field is its load table.
type tableAdapter struct{ table string }

func (tableAdapter) ID() string { return "table-probe" }
func (a tableAdapter) Load() source.Load {
	return source.Load{Origin: "test", Refresh: "never", Table: a.table}
}
func (tableAdapter) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{Name: "table-probe-by-name", Answer: "a row"},
		Build:    func(map[string]string, source.Rel) (string, error) { return "SELECT 1", nil },
	}}
}
func (tableAdapter) Rule(in source.Evidence) (source.Evidence, bool) { return in, len(in.Rows) > 0 }

// TestRegisterRefusesALoadTableThatIsNotAPlainName proves Register refuses a
// load table that is not one identifier or schema.name, because a builder may
// place that name in a query unquoted. The mutated field is Load.Table.
func TestRegisterRefusesALoadTableThatIsNotAPlainName(t *testing.T) {
	for name, table := range map[string]string{
		"statement": "t; DROP TABLE x",
		"quote":     "t'x",
		"space":     "a b",
		"three":     "a.b.c",
		"dash":      "a-b",
	} {
		if err := source.Register(tableAdapter{table: table}); err == nil {
			source.Unregister("table-probe")
			t.Errorf("%s: Register admitted Load.Table %q", name, table)
		}
	}
	if err := source.Register(tableAdapter{table: "px_schema.banned_pkg"}); err != nil {
		t.Fatalf("a plain schema.name was refused: %v", err)
	}
	source.Unregister("table-probe")
}
