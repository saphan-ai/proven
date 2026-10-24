package osv

import (
	"strings"

	"github.com/saphan-ai/proven/internal/manifest"
	"github.com/saphan-ai/proven/internal/semver"
)

// InRange reports whether version is inside one affected range.
// An empty introduced, or the OSV value "0", means there is no lower bound.
// A non-empty fixed is an exclusive upper bound and wins over last_affected.
// When fixed is empty, last_affected is an inclusive upper bound.
// When both ends are empty, every version this comparer accepts at or after
// introduced matches. This function compares semantic versions. An ecosystem
// uses its own adapter (InRangeEcosystem).
// The string is the comparison, whether or not the version is inside the range.
func InRange(version, introduced, fixed, last string) (bool, string) {
	return inRange(version, introduced, fixed, last, semver.Explain, semver.Compare)
}

// InRangeEcosystem is InRange for one package ecosystem.
// The order is the adapter's, not a switch kept beside it (ADR-0038): Go and
// npm compare semantic versions, PyPI compares PEP 440, and a later language
// brings its own Compare and Explain. An ecosystem with no adapter is not
// compared.
func InRangeEcosystem(ecosystem, version, introduced, fixed, last string) (bool, string) {
	r, ok := manifest.ByEcosystem(ecosystem)
	if !ok {
		return false, "ecosystem " + ecosystem + " has no version comparison"
	}
	return inRange(version, introduced, fixed, last, r.Explain, r.Compare)
}

// Uncompared says why version cannot be ordered in ecosystem.
// It is empty when the version can be compared.
func Uncompared(ecosystem, version string) string {
	r, ok := manifest.ByEcosystem(ecosystem)
	if !ok {
		return "ecosystem " + ecosystem + " has no version comparison"
	}
	why, parsed := r.Explain("", version)
	if parsed {
		return ""
	}
	return why
}

// inRange is the comparison behind InRange and InRangeEcosystem. Each value
// is checked before it is compared, so a value that does not parse gives a
// sentence and no match, never a guess.
func inRange(version, introduced, fixed, last string, explain func(kind, value string) (string, bool), cmp func(a, b string) (int, bool)) (bool, string) {
	version = strings.TrimSpace(version)
	introduced = strings.TrimSpace(introduced)
	fixed = strings.TrimSpace(fixed)
	last = strings.TrimSpace(last)
	if why, ok := explain("", version); !ok {
		return false, why
	}
	// OSV writes "0" for a range that starts at the first version.
	unbounded := introduced == "" || introduced == "0"
	if !unbounded {
		if why, ok := explain("introduced", introduced); !ok {
			return false, why
		}
		// Both values passed explain, so the pair orders.
		c, _ := cmp(version, introduced)
		if c < 0 {
			return false, version + " < introduced " + introduced
		}
	}
	// fixed is exclusive and wins over last_affected (docs/OSV_TABLE.md).
	if fixed != "" {
		if why, ok := explain("fixed", fixed); !ok {
			return false, why
		}
		c, _ := cmp(version, fixed)
		if c >= 0 {
			return false, "fixed " + fixed + " <= " + version
		}
		if unbounded {
			return true, version + " < fixed " + fixed
		}
		return true, "introduced " + introduced + " <= " + version + " < fixed " + fixed
	}
	if last != "" {
		if why, ok := explain("last_affected", last); !ok {
			return false, why
		}
		c, _ := cmp(version, last)
		if c > 0 {
			return false, "last_affected " + last + " < " + version
		}
		if unbounded {
			return true, version + " <= last_affected " + last
		}
		return true, "introduced " + introduced + " <= " + version + " <= last_affected " + last
	}
	if unbounded {
		return true, version + " has no upper bound"
	}
	return true, "introduced " + introduced + " <= " + version
}
