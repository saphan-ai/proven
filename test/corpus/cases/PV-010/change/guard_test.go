package main

import (
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	cases := []struct {
		name string
		note string
		pass bool
	}{
		{"current", "LANGUAGES 41 -> 42", true},
		{"stale", "LANGUAGES 41 -> 40", false},
		{"overflow", "LANGUAGES 41 -> 99999999999999999999", false},
		{"no arrow", "LANGUAGES 42", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ok, stale := Check([]string{c.note}, 42)
			if ok != c.pass {
				t.Fatalf("Check(%q) pass = %v, want %v (stale %v)", c.note, ok, c.pass, stale)
			}
			if !ok && (len(stale) != 1 || stale[0] != c.note) {
				t.Fatalf("stale = %v, want [%q]", stale, c.note)
			}
		})
	}
}

// TestDescriptionFollowsMatcher alters the matcher's gap characters and checks
// that the description still agrees with the matcher.
func TestDescriptionFollowsMatcher(t *testing.T) {
	saved := gapClass
	defer func() { gapClass = saved }()

	gapClass = "0-9 x"
	if !strings.Contains(Describe(), "[0-9 x]") {
		t.Fatalf("description %q does not follow the altered matcher", Describe())
	}
	if ok, _ := Check([]string{"LANGUAGESx-> 41"}, 42); ok {
		t.Fatal("altered matcher did not accept the new gap character")
	}
}
