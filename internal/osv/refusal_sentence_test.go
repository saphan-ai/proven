package osv

import "testing"

// TestLoadTablePrintsAQuoteInTheNameWithoutABackslash proves a direct caller
// of LoadTable still reads a refused name the way the package printed it:
// the quote is wrapped, and a quote inside the name is not escaped.
// The shared %q sentence inserts a backslash and makes this test fail.
func TestLoadTablePrintsAQuoteInTheNameWithoutABackslash(t *testing.T) {
	_, err := LoadTable(func(k string) string {
		if k == EnvTable {
			return "bad\"name"
		}
		return ""
	})
	want := `configuration refused: PROVEN_OSV_TABLE="bad"name": not a table name; allowed: a table or schema.table of letters, digits and underscores; empty means ` + DefaultTable
	if err == nil || err.Error() != want {
		got := ""
		if err != nil {
			got = err.Error()
		}
		t.Fatalf("got %q want %q", got, want)
	}
}
