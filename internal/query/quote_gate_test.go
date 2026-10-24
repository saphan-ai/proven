package query

import "testing"

// TestSelectGateSeesAStatementHiddenBehindAnotherQuoteCharacter proves the live client's gate does
// not let a double quote or a backtick hide a second statement. A dialect that reads one of them as
// a quote sees the semicolon below as a separator, while a gate that tracks only the single quote
// reads it as text inside a literal. The mutated field is the quote character around the single
// quote that would flip the gate's state.
func TestSelectGateSeesAStatementHiddenBehindAnotherQuoteCharacter(t *testing.T) {
	for name, sql := range map[string]string{
		"double quote": `SELECT 1 FROM t WHERE a = "'" ; DELETE FROM t WHERE b = "'"`,
		"backtick":     "SELECT 1 FROM t WHERE a = `'` ; DELETE FROM t WHERE b = `'`",
		"identifier":   `SELECT "a'" FROM t; DELETE FROM t WHERE x = 'y`,
	} {
		if oneSelect(sql) {
			t.Errorf("%s: oneSelect(%q) = true; a second statement hides behind the quote", name, sql)
		}
	}
}
