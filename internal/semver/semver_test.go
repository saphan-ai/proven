package semver

import (
	"math"
	"strconv"
	"testing"
)

// TestCompare proves SemVer precedence: an optional v, a release above its
// pre-release, numeric pre-release identifiers below words, build metadata
// ignored, Go pseudo-versions ordered by time, and short versions and
// leading zeros refused.
func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.4", -1},
		{"1.2.3", "1.2.3-rc.1", 1},
		{"1.2.3-rc.1", "1.2.3-rc.2", -1},
		{"1.2.3-rc.1", "1.2.3-rc.1.1", -1},
		{"1.2.3-alpha", "1.2.3-alpha.1", -1},
		{"1.2.3-1", "1.2.3-a", -1},
		{"v1.2.3+incompatible", "v1.2.3", 0},
		{"v0.0.0-20210101120000-abcdef", "v0.0.0-20220101120000-abcdef", -1},
	}
	for _, tc := range cases {
		got, ok := Compare(tc.a, tc.b)
		if !ok || got != tc.want {
			t.Fatalf("Compare(%s, %s) = %d, %t; want %d", tc.a, tc.b, got, ok, tc.want)
		}
	}
	if _, ok := Compare("1.2", "1.2.3"); ok {
		t.Fatal("1.2 parsed")
	}
	if _, ok := Parse("01.2.3"); ok {
		t.Fatal("leading zero parsed")
	}
	if _, ok := Parse("1.2.3-01"); ok {
		t.Fatal("pre-release leading zero parsed")
	}
}

// TestParseRefusesANumberThatDoesNotFit proves a numeric field wider than
// int is not a version. Wrapping it would order a larger field as a smaller one.
func TestParseRefusesANumberThatDoesNotFit(t *testing.T) {
	cases := []string{
		"99999999999999999999.0.0",
		"1.99999999999999999999.0",
		"1.0.99999999999999999999",
		"1.2.3-99999999999999999999",
	}
	for _, s := range cases {
		if v, ok := Parse(s); ok {
			t.Fatalf("Parse(%s) = %#v, true; want refused", s, v)
		}
	}
	// MaxInt still fits. One past it does not, on this architecture's int.
	fit := strconv.Itoa(math.MaxInt) + ".0.0"
	if _, ok := Parse(fit); !ok {
		t.Fatalf("Parse(%s) refused; MaxInt fits", fit)
	}
	over := strconv.FormatUint(uint64(math.MaxInt)+1, 10) + ".0.0"
	if v, ok := Parse(over); ok {
		t.Fatalf("Parse(%s) = %#v, true; want refused", over, v)
	}
}
