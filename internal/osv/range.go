package osv

import (
	"strings"

	"github.com/saphan-ai/proven/internal/semver"
)

// InRange reports whether version is inside one affected range.
// An empty introduced, or the OSV value "0", means there is no lower bound.
// A non-empty fixed is an exclusive upper bound and wins over last_affected.
// When fixed is empty, last_affected is an inclusive upper bound.
// When both ends are empty, every semantic version at or after introduced matches.
// The string is the comparison, whether or not the version is inside the range.
func InRange(version, introduced, fixed, last string) (bool, string) {
	version = strings.TrimSpace(version)
	introduced = strings.TrimSpace(introduced)
	fixed = strings.TrimSpace(fixed)
	last = strings.TrimSpace(last)
	if _, ok := semver.Parse(version); !ok {
		return false, version + " is not a semantic version"
	}
	unbounded := introduced == "" || introduced == "0"
	if !unbounded {
		if _, ok := semver.Parse(introduced); !ok {
			return false, "introduced " + introduced + " is not a semantic version"
		}
		cmp, _ := semver.Compare(version, introduced)
		if cmp < 0 {
			return false, version + " < introduced " + introduced
		}
	}
	if fixed != "" {
		if _, ok := semver.Parse(fixed); !ok {
			return false, "fixed " + fixed + " is not a semantic version"
		}
		cmp, _ := semver.Compare(version, fixed)
		if cmp >= 0 {
			return false, "fixed " + fixed + " <= " + version
		}
		if unbounded {
			return true, version + " < fixed " + fixed
		}
		return true, "introduced " + introduced + " <= " + version + " < fixed " + fixed
	}
	if last != "" {
		if _, ok := semver.Parse(last); !ok {
			return false, "last_affected " + last + " is not a semantic version"
		}
		cmp, _ := semver.Compare(version, last)
		if cmp > 0 {
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
