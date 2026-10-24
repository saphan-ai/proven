package osv

import (
	"sort"

	"github.com/saphan-ai/proven/internal/candidate"
	"github.com/saphan-ai/proven/internal/manifest"
)

// compareFor is the version order of one ecosystem. It is the adapter's
// Compare (ADR-0038), so a language added to Readers is ordered here too.
// It is nil when no adapter compares that ecosystem.
func compareFor(ecosystem string) func(a, b string) (int, bool) {
	r, ok := manifest.ByEcosystem(ecosystem)
	if !ok || r.Compare == nil {
		return nil
	}
	return r.Compare
}

// SmallestFix names the smallest version of the package that leaves the affected range of
// every advisory row the table returned, with the row whose fixed version it is
// (ADR-0042 point 5). It reads only rows the osv-by-package template returned (ADR-0032),
// and it compares in Go, never in SQL and never by a model.
//
// A candidate version is the fixed bound of some row. It is chosen when it is above the
// version in use and no row's range holds it: a fixed bound that another advisory's range
// still contains is not a way out. The result is nil when no row names a usable fixed
// version, for example when an advisory has only a last-affected bound, because the next
// version is then not a fact in the table and Proven does not guess it.
func SmallestFix(hit manifest.Hit, rows []Row) *candidate.DependencyFix {
	cmp := compareFor(hit.Ecosystem)
	if cmp == nil {
		return nil
	}
	// The version in use must lie in some range, or there is nothing to leave.
	var matched []Row
	for _, r := range rows {
		if ok, _ := InRangeEcosystem(hit.Ecosystem, hit.Version, r.Introduced, r.Fixed, r.LastAffected); ok {
			matched = append(matched, r)
		}
	}
	if len(matched) == 0 {
		return nil
	}
	type option struct {
		version string
		row     Row
	}
	var options []option
	for _, r := range rows {
		if r.Fixed == "" {
			continue
		}
		// A bound that does not parse, or one at or below the version in use, is no way out.
		above, ok := cmp(r.Fixed, hit.Version)
		if !ok || above <= 0 {
			continue
		}
		options = append(options, option{version: r.Fixed, row: r})
	}
	sort.SliceStable(options, func(i, j int) bool {
		c, _ := cmp(options[i].version, options[j].version)
		return c < 0
	})
	for _, o := range options {
		if stillAffected(hit.Ecosystem, o.version, rows) {
			continue
		}
		return &candidate.DependencyFix{
			Ecosystem: hit.Ecosystem, Package: hit.Package, From: hit.Version, To: o.version,
			AdvisoryID:      o.row.ID,
			AdvisoryColumns: append([]string(nil), Columns...),
			AdvisoryRow:     o.row.Fields(),
		}
	}
	return nil
}

// dependencyFix is the Fix of a dependency finding: the version SmallestFix names with its
// reason, or a note saying no row names a way out (ADR-0042 point 5).
func dependencyFix(hit manifest.Hit, rows []Row) *candidate.Fix {
	dep := SmallestFix(hit, rows)
	if dep == nil {
		return &candidate.Fix{
			Kind: candidate.FixDependency,
			Note: "no fix named: no advisory row names a fixed version that leaves every affected range of " + hit.Package,
		}
	}
	return &candidate.Fix{
		Kind: candidate.FixDependency,
		What: "Raise " + hit.Package + " from " + hit.Version + " to " + dep.To + " in " + hit.File + ".",
		Why: dep.To + " is the fixed version of advisory " + dep.AdvisoryID + ", and it is the smallest fixed version of the rows " +
			"returned that no affected range of those rows holds. No test proves a version change, so this fix is named from the table, not run.",
		Dependency: dep,
	}
}

// stillAffected reports whether version lies in the range of any row.
func stillAffected(ecosystem, version string, rows []Row) bool {
	for _, r := range rows {
		if ok, _ := InRangeEcosystem(ecosystem, version, r.Introduced, r.Fixed, r.LastAffected); ok {
			return true
		}
	}
	return false
}
