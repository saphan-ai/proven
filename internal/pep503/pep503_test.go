package pep503

import "testing"

// TestNormalise proves PEP 503: lower case, and a run of `-`, `_` or `.`
// becomes one `-`. The same call again changes nothing, so a name that is
// already normalised stays the query key.
func TestNormalise(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{in: "Requests", want: "requests"},
		{in: "zope_interface", want: "zope-interface"},
		{in: "zope.interface", want: "zope-interface"},
		{in: "Foo--Bar__baz..Qux", want: "foo-bar-baz-qux"},
		{in: "leftpad", want: "leftpad"},
	}
	for _, tc := range cases {
		got := Normalise(tc.in)
		if got != tc.want {
			t.Fatalf("Normalise(%q) = %q, want %q", tc.in, got, tc.want)
		}
		if again := Normalise(got); again != got {
			t.Fatalf("Normalise(%q) = %q", got, again)
		}
	}
}
