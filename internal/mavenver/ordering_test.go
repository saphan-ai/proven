package mavenver

import "testing"

// Compare every pair across ordered equivalence classes, including aliases,
// qualifiers and numeric padding. Both argument directions must agree with the
// same ordering, and every spelling within one class must compare equal.
func TestCompareOrderedEquivalenceClasses(t *testing.T) {
	classes := [][]string{
		{"1.0-alpha", "1.0-a", "1.0-alpha0"},
		{"1.0-beta1", "1.0-b1"},
		{"1.0-rc1", "1.0-cr1"},
		{"1.0-SNAPSHOT", "1.0-snapshot"},
		{"1", "1.0", "1.0.0", "1.0-final", "1.0-ga"},
		{"1.0-sp", "1.0-sp0"},
		{"1.0-sp1"},
		{"1.0.1", "1.0-1"},
		{"1.1", "1.01.0"},
		{"2"},
	}
	for i, left := range classes {
		for j, right := range classes {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			for _, a := range left {
				for _, b := range right {
					if got, ok := Compare(a, b); !ok || got != want {
						t.Fatalf("Compare(%q, %q) = %d, %t; want %d, true", a, b, got, ok, want)
					}
				}
			}
		}
	}
}
