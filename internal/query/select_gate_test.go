package query

import "testing"

// TestSelectGateSeesASecondStatementHoweverItIsSpelled proves that the live
// client's gate counts statements however the text before the semicolon is
// spelled: a quote inside a comment, or after a backslash, must not hide a
// second statement. The mutated field is the bytes before the semicolon.
func TestSelectGateSeesASecondStatementHoweverItIsSpelled(t *testing.T) {
	texts := map[string]string{
		"line comment holding a quote":  "SELECT 1 FROM t -- '\n; DELETE FROM t",
		"block comment holding a quote": "SELECT 1 FROM t /* ' */; DELETE FROM t; SELECT '",
		"backslash before a quote":      `SELECT 'a\''; DELETE FROM t; SELECT ''`,
	}
	for name, sql := range texts {
		if oneSelect(sql) {
			t.Errorf("%s: oneSelect(%q) = true; the text holds a second statement and must not be sent", name, sql)
		}
	}
}

// TestSelectGateAdmitsTheQueriesProvenBuilds proves the gate is not so strict
// that it refuses Proven's own queries: a literal holding "--", "/*" or a
// semicolon-free value with a hyphen is plain text inside quotes. The mutated
// field is the literal's content.
func TestSelectGateAdmitsTheQueriesProvenBuilds(t *testing.T) {
	for _, sql := range []string{
		"SELECT cve_id FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'",
		"SELECT id FROM t WHERE ecosystem = 'npm' AND package = 'a--b'",
		"SELECT id FROM t WHERE package = '@scope/a/*b'",
		"select 1 from t\n where a = 'x'",
	} {
		if !oneSelect(sql) {
			t.Errorf("oneSelect(%q) = false; Proven's own query was refused", sql)
		}
	}
}

// TestSelectGateRefusesWhatIsNotOneSelect proves the gate refuses a text that
// is not one SELECT: another verb first, a statement separator, a control
// byte, invalid UTF-8 and an unclosed quote. The mutated field is the text.
func TestSelectGateRefusesWhatIsNotOneSelect(t *testing.T) {
	for name, sql := range map[string]string{
		"delete":         "DELETE FROM t",
		"two statements": "SELECT 1; SELECT 2",
		"escape byte":    "SELECT '\x1b' FROM t",
		"nul":            "SELECT 1\x00",
		"invalid utf-8":  "SELECT '\xff' FROM t",
		"unclosed quote": "SELECT 'a FROM t",
		"empty":          "",
	} {
		if oneSelect(sql) {
			t.Errorf("%s: oneSelect(%q) = true", name, sql)
		}
	}
}

// TestLocalEngineRefusesTheSameTexts is the paired path: the sample engine
// refuses each text the live gate must refuse.
func TestLocalEngineRefusesTheSameTexts(t *testing.T) {
	db, err := LoadSQL("CREATE TABLE t (a text); INSERT INTO t VALUES ('x');")
	if err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		"SELECT 1 FROM t -- '\n; DELETE FROM t",
		"SELECT 1 FROM t /* ' */; DELETE FROM t; SELECT '",
		`SELECT 'a\''; DELETE FROM t; SELECT ''`,
	} {
		if _, _, err := db.Query(sql); err == nil {
			t.Errorf("local engine ran %q", sql)
		}
	}
}
