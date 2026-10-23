package querytpl

import (
	"errors"
	"regexp"
	"strings"
	"testing"
)

// TestLiteralByteBackstopRefusesWhatAFormatShouldNeverAdmit proves the last test inside the literal
// builder (ADR-0032 point 3) on its own: given a format that admits every text, it still refuses a
// single quote, a backslash, a control byte and a byte of 0x80 or more. Deleting the byte loop in
// literal turns this test red. The mutated field is the value; the format is held permissive.
func TestLiteralByteBackstopRefusesWhatAFormatShouldNeverAdmit(t *testing.T) {
	permissive := regexp.MustCompile(`(?s)^.*$`)
	for name, value := range map[string]string{
		"single quote": "it's",
		"backslash":    `a\b`,
		"escape byte":  "a\x1bb",
		"tab":          "a\tb",
		"line feed":    "a\nb",
		"delete":       "a\x7fb",
		"high byte":    "caf\xc3\xa9",
		"bad utf-8":    "a\xffb",
	} {
		sql, err := literal("t", "p", value, permissive)
		var ref *Refusal
		if !errors.As(err, &ref) || ref.Reason != ReasonBadValue {
			t.Errorf("%s: the backstop admitted %q as %s (err %v)", name, value, sql, err)
		}
	}
	got, err := literal("t", "p", "plain-value_1.0", permissive)
	if err != nil || got != "'plain-value_1.0'" {
		t.Fatalf("a plain value was refused or changed: %q %v", got, err)
	}
}

// TestEveryIdParameterRefusesLineBreaksNulAndAForeignId proves the id parameters refuse a carriage
// return, a line feed, a NUL, an escape byte and an id of the wrong kind, name the template and the
// parameter, and never repeat the value (ADR-0032). The mutated field is the value.
func TestEveryIdParameterRefusesLineBreaksNulAndAForeignId(t *testing.T) {
	cases := []struct {
		template, param string
		good            string
	}{
		{KEVByCVE, ParamCVEID, "CVE-2021-44228"},
		{EPSSByCVE, ParamCVEID, "CVE-2021-44228"},
		{OSVByAdvisory, ParamAdvisoryID, "GHSA-jfh8-c2jp-5v3q"},
	}
	for _, c := range cases {
		build := func(v string) error {
			_, err := Build(Request{Template: c.template, Params: map[string]string{c.param: v}}, Relations{})
			return err
		}
		if err := build(c.good); err != nil {
			t.Fatalf("%s: a well-formed id was refused: %v", c.template, err)
		}
		for name, tail := range map[string]string{
			"carriage return": "\r", "line feed": "\n", "nul": "\x00", "escape": "\x1b", "tab": "\t",
		} {
			value := c.good + tail
			err := build(value)
			var ref *Refusal
			if !errors.As(err, &ref) || ref.Reason != ReasonBadValue {
				t.Errorf("%s with a trailing %s was not refused: %v", c.template, name, err)
				continue
			}
			if !strings.Contains(ref.Detail, c.template) || !strings.Contains(ref.Detail, c.param) || strings.ContainsAny(ref.Error(), "\r\n\x00\x1b\t") {
				t.Errorf("%s: the refusal does not name the parameter alone: %q", name, ref.Error())
			}
		}
	}
	// An id of the wrong kind, and a CVE id with a query tail.
	for _, v := range []string{"GHSA-jfh8-c2jp-5v3q", "not-an-id", "cve-2021-44228", "CVE-2021-44228' OR '1'='1"} {
		if _, err := KEVSQL(DefaultKEVView, v); err == nil {
			t.Errorf("KEVSQL built a query from %q", v)
		}
		if _, err := EPSSSQL(DefaultEPSSView, v); err == nil {
			t.Errorf("EPSSSQL built a query from %q", v)
		}
	}
}

// TestPackageNamesRefuseEveryControlByte proves no package name holding a control byte, DEL or a
// byte that is not ASCII reaches a query, in any ecosystem (ADR-0032). The mutated field is the byte.
func TestPackageNamesRefuseEveryControlByte(t *testing.T) {
	for _, eco := range []string{"Go", "npm", "PyPI"} {
		for b := 0; b < 0x20; b++ {
			name := "pkg" + string(rune(b)) + "name"
			if sql, err := OSVPackageSQL(DefaultOSVTable, eco, name); err == nil {
				t.Errorf("%s: control byte %#x was built into %s", eco, b, sql)
			}
		}
		for _, name := range []string{"pkg\x7fname", "pkgéname", "pkg\xffname"} {
			if sql, err := OSVPackageSQL(DefaultOSVTable, eco, name); err == nil {
				t.Errorf("%s: %q was built into %s", eco, name, sql)
			}
		}
	}
}
