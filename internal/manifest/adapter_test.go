package manifest

import "testing"

// TestReadersCarryTheComparison proves ADR-0038's contract for the languages
// already shipped: each ecosystem name orders two versions, and a name with
// no adapter does not. The mutated field of the last case is the ecosystem
// name.
func TestReadersCarryTheComparison(t *testing.T) {
	cases := []struct {
		eco string
		a   string
		b   string
		cmp int
	}{
		{"Go", "v1.2.3", "v1.2.0", 1},
		{"npm", "1.2.3", "1.2.3", 0},
		{"PyPI", "1.2.0", "1.2", 0},
		{"Packagist", "1.0.0-RC1", "1.0.0", -1},
		{"crates.io", "1.2.4", "1.2.3", 1},
	}
	for _, tc := range cases {
		r, ok := ByEcosystem(tc.eco)
		if !ok {
			t.Fatalf("%s has no adapter", tc.eco)
		}
		got, ordered := r.Compare(tc.a, tc.b)
		if !ordered || got != tc.cmp {
			t.Fatalf("%s Compare(%s, %s) = %d, %t", tc.eco, tc.a, tc.b, got, ordered)
		}
		if why, parsed := r.Explain("", tc.a); !parsed || why != "" {
			t.Fatalf("%s Explain %q %t", tc.eco, why, parsed)
		}
	}
	if _, ok := ByEcosystem("Maven"); ok {
		t.Fatal("Maven was compared")
	}
}
