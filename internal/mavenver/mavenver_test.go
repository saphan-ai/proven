package mavenver

import "testing"

// TestMavenRules names each ordering rule of the subset: numeric tokens,
// padding, the qualifier order alpha < beta < milestone < rc < snapshot <
// release < sp, the aliases, a qualifier number, and an unknown qualifier.
func TestMavenRules(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want int
	}{
		{name: "numeric segments", a: "1.2", b: "1.10", want: -1},
		{name: "leading zeros", a: "1.01", b: "1.1", want: 0},
		{name: "missing segment is zero", a: "1", b: "1.0.0", want: 0},
		{name: "hyphen and dot separate numbers", a: "1.0-1", b: "1.0.1", want: 0},
		{name: "alpha below beta", a: "1.0-alpha", b: "1.0-beta", want: -1},
		{name: "beta below milestone", a: "1.0-beta", b: "1.0-milestone", want: -1},
		{name: "milestone below rc", a: "1.0-milestone", b: "1.0-rc", want: -1},
		{name: "rc below snapshot", a: "1.0-rc", b: "1.0-snapshot", want: -1},
		{name: "snapshot below release", a: "1.0-snapshot", b: "1.0", want: -1},
		{name: "release below sp", a: "1.0", b: "1.0-sp", want: -1},
		{name: "alias a", a: "1.0-a", b: "1.0-alpha", want: 0},
		{name: "alias b", a: "1.0-b", b: "1.0-beta", want: 0},
		{name: "alias m", a: "1.0-m", b: "1.0-milestone", want: 0},
		{name: "alias cr", a: "1.0-cr", b: "1.0-rc", want: 0},
		{name: "alias ga", a: "1.0-ga", b: "1.0", want: 0},
		{name: "alias final", a: "1.0-final", b: "1.0", want: 0},
		{name: "alias release", a: "1.0-release", b: "1.0", want: 0},
		{name: "case is ignored", a: "1.0-SNAPSHOT", b: "1.0-snapshot", want: 0},
		{name: "qualifier number", a: "1.0-alpha", b: "1.0-alpha-1", want: -1},
		{name: "qualifier number zero", a: "1.0-alpha", b: "1.0-alpha-0", want: 0},
		{name: "glued qualifier number", a: "1.0-rc1", b: "1.0-rc-1", want: 0},
		{name: "dot before qualifier", a: "1.0.0.Final", b: "1.0.0", want: 0},
		{name: "glued dot qualifier", a: "1.0.0.Beta1", b: "1.0.0-beta-1", want: 0},
		{name: "qualifier below the same numbers", a: "1.0-rc99", b: "1.0", want: -1},
		{name: "sp number", a: "1.0-sp", b: "1.0-sp1", want: -1},
		{name: "unknown qualifier above sp", a: "1.0-sp", b: "1.0-jre", want: -1},
		{name: "unknown qualifiers are lexical", a: "1.0-android", b: "1.0-jre", want: -1},
		{name: "numbers before qualifier", a: "1.2-jre", b: "1.3-jre", want: -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Compare(tc.a, tc.b)
			if !ok || got != tc.want {
				t.Fatalf("Compare(%s, %s) = %d, %t, want %d", tc.a, tc.b, got, ok, tc.want)
			}
		})
	}
}

// TestMavenRefusals proves a constraint, a leading v and a string that is
// neither are each named, and are not ordered.
func TestMavenRefusals(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"[1.0,2.0)", "[1.0,2.0) is a constraint; this check compares one Maven version"},
		{"1.2.+", "1.2.+ is a constraint; this check compares one Maven version"},
		{"^1.2", "^1.2 is a constraint; this check compares one Maven version"},
		{"v1.2.3", "v1.2.3 is not a Maven version; a Maven version has no leading v"},
		{"latest", "latest is not a Maven version this check compares"},
		{"", "empty is not a Maven version this check compares"},
		{"1.0-alpha-beta", "1.0-alpha-beta is not a Maven version this check compares"},
	}
	for _, tc := range cases {
		if _, ok := Compare(tc.in, "1.0.0"); ok {
			t.Fatalf("%q was ordered", tc.in)
		}
		if got := Reject(tc.in); got != tc.want {
			t.Fatalf("%q: got %q", tc.in, got)
		}
		why, ok := Explain("introduced", tc.in)
		if ok || why != "introduced "+tc.want {
			t.Fatalf("Explain %q: %q %t", tc.in, why, ok)
		}
	}
}
