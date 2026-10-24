package osv

import (
	"testing"

	"github.com/saphan-ai/proven/internal/manifest"
)

// goHit is a Go package at one version, as the manifest reader reports it.
func goHit(version string) manifest.Hit {
	return manifest.Hit{Ecosystem: "Go", File: "go.mod", Change: manifest.Change{Package: "example.com/lib", Version: version, Line: 5}}
}

// row is one advisory row for example.com/lib.
func row(id, introduced, fixed, last string) Row {
	return Row{ID: id, Ecosystem: "Go", Package: "example.com/lib", Introduced: introduced, Fixed: fixed, LastAffected: last}
}

// TestSmallestFixIsTheLowestFixedVersionNoRangeHolds proves ADR-0042 point 5: of the fixed
// bounds above the version in use, the lowest one that no row's range holds is named, with the row
// whose bound it is.
func TestSmallestFixIsTheLowestFixedVersionNoRangeHolds(t *testing.T) {
	rows := []Row{
		row("GO-2026-0002", "1.3.0", "1.4.0", ""),
		row("GO-2026-0001", "0", "1.2.5", ""),
	}
	got := SmallestFix(goHit("v1.2.3"), rows)
	if got == nil || got.To != "1.2.5" || got.AdvisoryID != "GO-2026-0001" || got.From != "v1.2.3" {
		t.Fatalf("%#v", got)
	}
	if len(got.AdvisoryRow) != len(got.AdvisoryColumns) || got.AdvisoryRow[0] != "GO-2026-0001" {
		t.Fatalf("the advisory row does not follow its columns: %v %v", got.AdvisoryColumns, got.AdvisoryRow)
	}
}

// TestSmallestFixSkipsABoundAnotherAdvisoryStillHolds proves a fixed bound inside another
// advisory's range is not a way out: the next bound is named instead.
func TestSmallestFixSkipsABoundAnotherAdvisoryStillHolds(t *testing.T) {
	rows := []Row{
		row("GO-2026-0001", "0", "1.2.5", ""),
		// A second advisory begins at 1.2.5 and runs to 1.3.0, so 1.2.5 is still affected.
		row("GO-2026-0003", "1.2.5", "1.3.0", ""),
	}
	got := SmallestFix(goHit("1.2.3"), rows)
	if got == nil || got.To != "1.3.0" || got.AdvisoryID != "GO-2026-0003" {
		t.Fatalf("%#v", got)
	}
}

// TestSmallestFixNamesNothingWithoutAFixedVersion proves a table that only says "last affected"
// gives no fix: the next version is not a fact in the table, and Proven does not guess it.
func TestSmallestFixNamesNothingWithoutAFixedVersion(t *testing.T) {
	cases := map[string][]Row{
		"last affected only":                {row("GO-2026-0004", "0", "", "1.9.9")},
		"no upper bound":                    {row("GO-2026-0005", "0", "", "")},
		"version outside every range":       {row("GO-2026-0006", "0", "1.0.0", "")},
		"fixed bound below the version":     {row("GO-2026-0007", "0", "1.2.0", "")},
		"no rows":                           nil,
		"a fixed bound that does not parse": {row("GO-2026-0008", "0", "next", "")},
	}
	for name, rows := range cases {
		if got := SmallestFix(goHit("1.2.3"), rows); got != nil {
			t.Fatalf("%s: %#v", name, got)
		}
	}
}

// TestSmallestFixComparesPyPIAndNpmByTheirOwnOrder proves the choice follows each ecosystem's
// version order: PEP 440 for PyPI, semantic versions for npm. 1.10 is above 1.9 in both, so a text sort would name the wrong one.
func TestSmallestFixComparesPyPIAndNpmByTheirOwnOrder(t *testing.T) {
	py := manifest.Hit{Ecosystem: "PyPI", File: "requirements.txt", Change: manifest.Change{Package: "lib", Version: "1.8", Line: 1}}
	rows := []Row{
		{ID: "PYSEC-2", Ecosystem: "PyPI", Package: "lib", Introduced: "1.9.5", Fixed: "1.10"},
		{ID: "PYSEC-1", Ecosystem: "PyPI", Package: "lib", Introduced: "0", Fixed: "1.9"},
	}
	if got := SmallestFix(py, rows); got == nil || got.To != "1.9" {
		t.Fatalf("pypi: %#v", got)
	}
	npm := manifest.Hit{Ecosystem: "npm", File: "package.json", Change: manifest.Change{Package: "lib", Version: "1.8.0", Line: 1}}
	rows = []Row{
		{ID: "GHSA-a", Ecosystem: "npm", Package: "lib", Introduced: "1.9.5", Fixed: "1.10.0"},
		{ID: "GHSA-b", Ecosystem: "npm", Package: "lib", Introduced: "0", Fixed: "1.9.0"},
	}
	if got := SmallestFix(npm, rows); got == nil || got.To != "1.9.0" {
		t.Fatalf("npm: %#v", got)
	}
}

// TestSmallestFixUsesTheCargoOrder proves crates.io uses the adapter's order.
// 1.10.0 is above 1.9.1 and outside the later range, so it is the smallest way out.
func TestSmallestFixUsesTheCargoOrder(t *testing.T) {
	hit := manifest.Hit{Ecosystem: "crates.io", File: "Cargo.lock", Change: manifest.Change{Package: "regex", Version: "1.9.1", Line: 15}}
	rows := []Row{
		{ID: "GHSA-rs-sample-0002", Ecosystem: "crates.io", Package: "regex", Introduced: "1.11.0", Fixed: "1.12.0"},
		{ID: "GHSA-rs-sample-0001", Ecosystem: "crates.io", Package: "regex", Introduced: "1.7.0", Fixed: "1.10.0"},
	}
	got := SmallestFix(hit, rows)
	if got == nil || got.To != "1.10.0" || got.AdvisoryID != "GHSA-rs-sample-0001" || got.From != "1.9.1" {
		t.Fatalf("%#v", got)
	}
}

// TestSmallestFixRefusesAnEcosystemItCannotOrder proves a version is never ranked in an ecosystem
// the check has no order for.
func TestSmallestFixRefusesAnEcosystemItCannotOrder(t *testing.T) {
	hit := manifest.Hit{Ecosystem: "Maven", Change: manifest.Change{Package: "lib", Version: "1.0"}}
	if got := SmallestFix(hit, []Row{{ID: "X", Ecosystem: "Maven", Introduced: "0", Fixed: "2.0"}}); got != nil {
		t.Fatalf("%#v", got)
	}
}

// TestADependencyFindingCarriesItsFix proves the check attaches the fix to the finding it
// raises, never marks it proved, and says in a note when no row names a way out.
func TestADependencyFindingCarriesItsFix(t *testing.T) {
	hit := goHit("1.2.3")
	fix := dependencyFix(hit, []Row{row("GO-2026-0001", "0", "1.2.5", "")})
	if fix.Proved || fix.Dependency == nil || fix.Dependency.To != "1.2.5" || fix.Kind != "dependency" || fix.What == "" || fix.Why == "" {
		t.Fatalf("%#v", fix)
	}
	none := dependencyFix(hit, []Row{row("GO-2026-0004", "0", "", "1.9.9")})
	if none.Dependency != nil || none.Note == "" {
		t.Fatalf("%#v", none)
	}
}
