package osv

import (
	"strings"

	"github.com/saphan-ai/proven/internal/pep440"
	"github.com/saphan-ai/proven/internal/semver"
)

// InRange reports whether version is inside one affected range.
// An empty introduced, or the OSV value "0", means there is no lower bound.
// A non-empty fixed is an exclusive upper bound and wins over last_affected.
// When fixed is empty, last_affected is an inclusive upper bound.
// When both ends are empty, every version this comparer accepts at or after
// introduced matches. Go and npm use semantic versions.
// The string is the comparison, whether or not the version is inside the range.
func InRange(version, introduced, fixed, last string) (bool, string) {
	return inRange(version, introduced, fixed, last, explainSemver, semver.Compare)
}

// InRangeEcosystem is InRange for one package ecosystem.
// Go and npm use semantic versions. PyPI uses PEP 440 release, pre-release
// and post-release forms. Any other ecosystem is not compared.
func InRangeEcosystem(ecosystem, version, introduced, fixed, last string) (bool, string) {
	switch ecosystem {
	case "Go", "npm":
		return InRange(version, introduced, fixed, last)
	case "PyPI":
		return inRange(version, introduced, fixed, last, explainPyPI, pep440.Compare)
	default:
		return false, "ecosystem " + ecosystem + " has no version comparison"
	}
}

// Uncompared says why version cannot be ordered in ecosystem.
// It is empty when the version can be compared.
func Uncompared(ecosystem, version string) string {
	switch ecosystem {
	case "Go", "npm":
		why, ok := explainSemver("", version)
		if ok {
			return ""
		}
		return why
	case "PyPI":
		why, ok := explainPyPI("", version)
		if ok {
			return ""
		}
		return why
	default:
		return "ecosystem " + ecosystem + " has no version comparison"
	}
}

// explainSemver says why value is not a semantic version. kind names the
// bound the value came from, and is empty for the package's own version.
func explainSemver(kind, value string) (string, bool) {
	if _, ok := semver.Parse(value); ok {
		return "", true
	}
	if kind == "" {
		return value + " is not a semantic version", false
	}
	return kind + " " + value + " is not a semantic version", false
}

// explainPyPI says why value is not a PEP 440 form this check compares.
func explainPyPI(kind, value string) (string, bool) {
	if _, ok := pep440.Parse(value); ok {
		return "", true
	}
	why := pep440.Reject(value)
	if kind == "" {
		return why, false
	}
	return kind + " " + why, false
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
