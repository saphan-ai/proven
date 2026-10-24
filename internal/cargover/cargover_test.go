package cargover

import "testing"

// TestCompareOrdersEqualLowerAndHigher proves Cargo precedence on three
// semantic versions: the same triple is equal, build metadata does not
// change that, a lower patch is lower, a pre-release is lower than its
// release, and a higher patch is higher.
func TestCompareOrdersEqualLowerAndHigher(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{a: "1.2.3", b: "1.2.3", want: 0},
		{a: "1.2.3+build.1", b: "1.2.3", want: 0},
		{a: "1.2.3", b: "1.2.4", want: -1},
		{a: "1.2.3-rc.1", b: "1.2.3", want: -1},
		{a: "1.2.4", b: "1.2.3", want: 1},
	}
	for _, tc := range cases {
		got, ok := Compare(tc.a, tc.b)
		if !ok || got != tc.want {
			t.Fatalf("Compare(%s, %s) = %d, %t; want %d", tc.a, tc.b, got, ok, tc.want)
		}
	}
}

// TestCompareRefusesAStringThatIsNotAVersion proves a constraint, a short
// requirement and a leading v are each refused with a sentence, and that a
// pair that includes one of them is not ordered.
func TestCompareRefusesAStringThatIsNotAVersion(t *testing.T) {
	cases := []struct {
		value string
		want  string
	}{
		{value: "^1.2.3", want: "^1.2.3 is a constraint; this check compares one Cargo version"},
		{value: "1.2", want: "1.2 is not a Cargo version; this check compares one semantic version"},
		{value: "v1.2.3", want: "v1.2.3 is not a Cargo version; a Cargo version has no leading v"},
		{value: "", want: "empty is not a Cargo version; this check compares one semantic version"},
	}
	for _, tc := range cases {
		why, parsed := Explain("", tc.value)
		if parsed || why != tc.want {
			t.Fatalf("Explain(%q) = %q, %t", tc.value, why, parsed)
		}
		if _, ok := Compare("1.2.3", tc.value); ok {
			t.Fatalf("Compare ordered %q", tc.value)
		}
	}
	why, parsed := Explain("introduced", "^1.2.3")
	want := "introduced ^1.2.3 is a constraint; this check compares one Cargo version"
	if parsed || why != want {
		t.Fatalf("Explain bound = %q, %t", why, parsed)
	}
}
