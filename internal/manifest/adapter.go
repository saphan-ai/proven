package manifest

// Adapter is one language's dependency reader (ADR-0038).
// It is the same type as Reader. A new language fills one Adapter for its
// manifest and one for its lock file, both with the same ecosystem name and
// the same Compare and Explain, and adds them to Readers. The rest of the
// review (OSV ranges, KEV, EPSS, the report) is shared and does not change.
type Adapter = Reader

// Readers is every adapter the dependency check runs, manifest and lock
// file of each language. The order is fixed so a review of the same change
// always asks the same questions. Two adapters may share an ecosystem name:
// the lock file then wins for a package both of them name (preferLock).
func Readers() []Reader {
	return []Reader{
		Go(), GoSum(),
		NPM(), NPMLock(),
		PyPI(), Poetry(),
		Composer(), ComposerLock(),
		Cargo(), CargoLock(),
	}
}

// ByEcosystem returns one adapter for an OSV ecosystem name.
// The manifest and the lock file of one language share Compare and Explain,
// so the first adapter that names the ecosystem and can compare is enough.
// The bool is false when no adapter does, and the check then does not guess
// an order.
func ByEcosystem(name string) (Reader, bool) {
	for _, r := range Readers() {
		if r.Ecosystem == name && r.Compare != nil && r.Explain != nil {
			return r, true
		}
	}
	return Reader{}, false
}
