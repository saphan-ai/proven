package osv

import (
	"strings"
	"testing"
)

// TestInRange proves the OSV range rules for semantic versions: introduced
// and last_affected are inclusive, fixed is exclusive and wins over
// last_affected, "0" means no lower bound, a pre-release is below its
// release, and every outcome comes with its sentence.
func TestInRange(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		introduced string
		fixed      string
		last       string
		want       bool
		sentence   string
	}{
		{name: "introduced only", version: "1.2.3", introduced: "1.0.0", want: true, sentence: "introduced 1.0.0 <= 1.2.3"},
		{name: "introduced only below", version: "0.9.0", introduced: "1.0.0", want: false, sentence: "0.9.0 < introduced 1.0.0"},
		{name: "fixed", version: "1.2.3", introduced: "1.0.0", fixed: "1.2.4", want: true, sentence: "introduced 1.0.0 <= 1.2.3 < fixed 1.2.4"},
		{name: "equal to fixed", version: "1.2.4", introduced: "1.0.0", fixed: "1.2.4", want: false, sentence: "fixed 1.2.4 <= 1.2.4"},
		{name: "last_affected", version: "1.2.3", introduced: "1.0.0", last: "1.2.3", want: true, sentence: "introduced 1.0.0 <= 1.2.3 <= last_affected 1.2.3"},
		{name: "above last_affected", version: "1.2.4", introduced: "1.0.0", last: "1.2.3", want: false, sentence: "last_affected 1.2.3 < 1.2.4"},
		{name: "pre-release", version: "1.2.3-rc.1", introduced: "1.2.0", fixed: "1.2.3", want: true, sentence: "introduced 1.2.0 <= 1.2.3-rc.1 < fixed 1.2.3"},
		{name: "pre-release before introduced", version: "1.2.3-rc.1", introduced: "1.2.3", want: false, sentence: "1.2.3-rc.1 < introduced 1.2.3"},
		{name: "outside every range", version: "3.0.0", introduced: "1.0.0", fixed: "2.0.0", want: false, sentence: "fixed 2.0.0 <= 3.0.0"},
		{name: "introduced zero", version: "v1.2.3", introduced: "0", fixed: "1.2.4", want: true, sentence: "v1.2.3 < fixed 1.2.4"},
		{name: "v prefix", version: "v1.2.3", introduced: "1.0.0", fixed: "1.2.4", want: true, sentence: "introduced 1.0.0 <= v1.2.3 < fixed 1.2.4"},
		{name: "fixed wins over last_affected", version: "1.3.0", introduced: "1.0.0", fixed: "1.2.0", last: "1.5.0", want: false, sentence: "fixed 1.2.0 <= 1.3.0"},
		{name: "not semantic", version: "latest", introduced: "1.0.0", want: false, sentence: "latest is not a semantic version"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, sentence := InRange(tc.version, tc.introduced, tc.fixed, tc.last)
			if got != tc.want || sentence != tc.sentence {
				t.Fatalf("got %t %q", got, sentence)
			}
		})
	}
}

// TestInRangeEcosystem proves that npm compares semantic versions, that PyPI
// compares PEP 440 and refuses dev, epoch and local versions with a
// sentence, that Packagist compares the Composer subset and names a
// constraint, that crates.io compares a Cargo version and names a constraint,
// that a bound which does not parse is named, and that an
// unknown ecosystem is not compared.
func TestInRangeEcosystem(t *testing.T) {
	cases := []struct {
		name       string
		ecosystem  string
		version    string
		introduced string
		fixed      string
		last       string
		want       bool
		sentence   string
	}{
		{name: "npm semver", ecosystem: "npm", version: "1.2.3", introduced: "1.0.0", fixed: "1.2.4", want: true, sentence: "introduced 1.0.0 <= 1.2.3 < fixed 1.2.4"},
		{name: "npm pre-release", ecosystem: "npm", version: "1.2.3-rc.1", introduced: "1.2.3", want: false, sentence: "1.2.3-rc.1 < introduced 1.2.3"},
		{name: "npm range is not a version", ecosystem: "npm", version: "^1.2.3", introduced: "1.0.0", want: false, sentence: "^1.2.3 is not a semantic version"},
		{name: "pypi release", ecosystem: "PyPI", version: "1.2.3", introduced: "1.0", fixed: "1.2.4", want: true, sentence: "introduced 1.0 <= 1.2.3 < fixed 1.2.4"},
		{name: "pypi pre below release", ecosystem: "PyPI", version: "1.0a1", introduced: "1.0a1", fixed: "1.0", want: true, sentence: "introduced 1.0a1 <= 1.0a1 < fixed 1.0"},
		{name: "pypi post above release", ecosystem: "PyPI", version: "1.0.post1", introduced: "1.0", fixed: "1.1", want: true, sentence: "introduced 1.0 <= 1.0.post1 < fixed 1.1"},
		{name: "pypi dev is not compared", ecosystem: "PyPI", version: "1.0.dev1", introduced: "0", fixed: "9", want: false, sentence: "1.0.dev1 has a dev release; this check does not compare it"},
		{name: "pypi epoch is not compared", ecosystem: "PyPI", version: "1!1.0", introduced: "0", want: false, sentence: "1!1.0 has an epoch; this check compares release, pre-release and post-release only"},
		{name: "pypi local is not compared", ecosystem: "PyPI", version: "1.0+local", introduced: "0", want: false, sentence: "1.0+local has a local version; this check does not compare it"},
		{name: "pypi unparsed", ecosystem: "PyPI", version: "latest", introduced: "1.0", want: false, sentence: "latest is not a release, pre-release or post-release"},
		{name: "pypi bad introduced", ecosystem: "PyPI", version: "1.0", introduced: "1.0.dev1", want: false, sentence: "introduced 1.0.dev1 has a dev release; this check does not compare it"},
		{name: "packagist rc below the release", ecosystem: "Packagist", version: "1.0.0-RC1", introduced: "1.0.0-RC1", fixed: "1.0.0", want: true, sentence: "introduced 1.0.0-RC1 <= 1.0.0-RC1 < fixed 1.0.0"},
		{name: "packagist constraint is not a version", ecosystem: "Packagist", version: "^1.2", introduced: "1.0.0", want: false, sentence: "^1.2 is a constraint; this check compares one Composer version"},
		{name: "crates.io release", ecosystem: "crates.io", version: "1.9.1", introduced: "1.7.0", fixed: "1.10.0", want: true, sentence: "introduced 1.7.0 <= 1.9.1 < fixed 1.10.0"},
		{name: "crates.io pre-release", ecosystem: "crates.io", version: "1.0.0-alpha.1", introduced: "1.0.0", want: false, sentence: "1.0.0-alpha.1 < introduced 1.0.0"},
		{name: "crates.io constraint is not a version", ecosystem: "crates.io", version: "^1.2.3", introduced: "1.0.0", want: false, sentence: "^1.2.3 is a constraint; this check compares one Cargo version"},
		{name: "crates.io leading v", ecosystem: "crates.io", version: "v1.2.3", introduced: "1.0.0", want: false, sentence: "v1.2.3 is not a Cargo version; a Cargo version has no leading v"},
		{name: "unknown ecosystem", ecosystem: "Maven", version: "1.0", introduced: "0", want: false, sentence: "ecosystem Maven has no version comparison"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, sentence := InRangeEcosystem(tc.ecosystem, tc.version, tc.introduced, tc.fixed, tc.last)
			if got != tc.want || sentence != tc.sentence {
				t.Fatalf("got %t %q", got, sentence)
			}
		})
	}
	if why := Uncompared("npm", "1.2.3"); why != "" {
		t.Fatal(why)
	}
	if !strings.Contains(Uncompared("PyPI", "latest"), "not a release") {
		t.Fatal(Uncompared("PyPI", "latest"))
	}
}
