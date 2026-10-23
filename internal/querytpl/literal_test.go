package querytpl

import "testing"

// readBackslashLiteral reads one SQL string literal at the start of s the way a
// dialect with backslash escapes does: a backslash takes the next byte as
// itself, and a doubled quote is one quote. It returns the decoded value and
// the number of bytes the literal used, or ok=false when the literal never
// closes.
func readBackslashLiteral(s string) (val string, n int, ok bool) {
	if s == "" || s[0] != '\'' {
		return "", 0, false
	}
	var b []byte
	for i := 1; i < len(s); i++ {
		switch {
		case s[i] == '\\' && i+1 < len(s):
			b = append(b, s[i+1])
			i++
		case s[i] == '\'' && i+1 < len(s) && s[i+1] == '\'':
			b = append(b, '\'')
			i++
		case s[i] == '\'':
			return string(b), i + 1, true
		default:
			b = append(b, s[i])
		}
	}
	return "", 0, false
}

// readStandardLiteral reads one literal the way standard SQL and the local
// sample engine do: a doubled quote is one quote, a backslash is itself.
func readStandardLiteral(s string) (val string, n int, ok bool) {
	if s == "" || s[0] != '\'' {
		return "", 0, false
	}
	var b []byte
	for i := 1; i < len(s); i++ {
		if s[i] == '\'' {
			if i+1 < len(s) && s[i+1] == '\'' {
				b = append(b, '\'')
				i++
				continue
			}
			return string(b), i + 1, true
		}
		b = append(b, s[i])
	}
	return "", 0, false
}

// lastLiteral returns the text after the final " = " of a built query, which
// every template ends with: the one literal that holds a value.
func lastLiteral(sql string) string {
	const marker = " = "
	for i := len(sql) - len(marker); i >= 0; i-- {
		if sql[i:i+len(marker)] == marker {
			return sql[i+len(marker):]
		}
	}
	return ""
}

// TestEveryBuiltLiteralIsOneLiteralInEitherDialect proves the property of
// ADR-0032 for every template: the value that enters a query becomes exactly
// one string literal holding exactly that value, whether the engine reads a
// backslash as an escape or as a plain byte, or the value is refused. The
// mutated field is the value: each spelling that once reached a query is tried
// in the package, CVE and advisory parameters.
func TestEveryBuiltLiteralIsOneLiteralInEitherDialect(t *testing.T) {
	values := []string{
		`plain/name`,
		`CVE-2021-44228`,
		`GHSA-jfh8-c2jp-5v3q`,
		`it's`,
		`trailing\`,
		`mid\'quote`,
		`two\\'`,
		"tab\tbyte",
		"esc\x1bbyte",
		"del\x7fbyte",
		"bad\xffutf8",
		"CVE-2026-0001' OR cve_id = 'CVE-2026-0002",
	}
	builders := map[string]func(v string) (string, error){
		"go package":  func(v string) (string, error) { return OSVPackageSQL(DefaultOSVTable, "Go", v) },
		"npm package": func(v string) (string, error) { return OSVPackageSQL(DefaultOSVTable, "npm", v) },
		"cve in kev":  func(v string) (string, error) { return KEVSQL(DefaultKEVView, v) },
		"cve in epss": func(v string) (string, error) { return EPSSSQL(DefaultEPSSView, v) },
		"advisory": func(v string) (string, error) {
			return Build(Request{Template: OSVByAdvisory, Params: map[string]string{ParamAdvisoryID: v}}, Relations{})
		},
	}
	for name, build := range builders {
		for _, v := range values {
			sql, err := build(v)
			if err != nil {
				continue // a refusal is an allowed answer
			}
			lit := lastLiteral(sql)
			for dialect, read := range map[string]func(string) (string, int, bool){
				"standard":  readStandardLiteral,
				"backslash": readBackslashLiteral,
			} {
				got, n, ok := read(lit)
				if !ok || n != len(lit) || got != v {
					t.Errorf("%s, %s reader: value %q built %s, which is not one literal holding the value (ok=%v used=%d of %d, read %q)",
						name, dialect, v, sql, ok, n, len(lit), got)
				}
			}
		}
	}
}

// TestOnlyFormedValuesAreBuilt is the unplug control of the test above: the
// values that hold a quote, a backslash, a control byte or bad UTF-8 are
// refused, not merely harmless, so the test above is not passing on an empty
// set. The mutated field is the value.
func TestOnlyFormedValuesAreBuilt(t *testing.T) {
	for _, v := range []string{`it's`, `trailing\`, `mid\'quote`, "tab\tbyte", "esc\x1bbyte", "del\x7fbyte", "bad\xffutf8"} {
		if sql, err := OSVPackageSQL(DefaultOSVTable, "npm", v); err == nil {
			t.Errorf("%q was built into %s", v, sql)
		}
	}
	if _, err := OSVPackageSQL(DefaultOSVTable, "npm", "plain-name"); err != nil {
		t.Errorf("a plain name was refused: %v", err)
	}
}
