// Package semver compares semantic versions for the dependency check.
// The manifest readers use it to tell a raised Go or npm version from a
// lowered one, and package osv uses it to place a version inside or
// outside an OSV range.
//
// A version is MAJOR.MINOR.PATCH, an optional leading v, an optional
// pre-release, and optional build metadata that is ignored. A numeric
// field that does not fit in int is not a version. Anything else,
// an npm range such as ^1.2.3 included, is not a version: Compare says so
// and the caller writes the reason into the trace instead of guessing.
package semver

import "strings"

// Version is one semantic version. Pre holds the dot-separated pre-release
// identifiers. An empty Pre means a release.
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   []string
}

// Parse reads a semantic version. The bool is false when s is not one.
func Parse(s string) (Version, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Version{}, false
	}
	// Go writes a leading v and npm does not; both forms are one version.
	if s[0] == 'v' || s[0] == 'V' {
		s = s[1:]
	}
	// Build metadata has no precedence, so v1.2.3+incompatible equals v1.2.3.
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}
	pre := ""
	if i := strings.IndexByte(s, '-'); i >= 0 {
		pre = s[i+1:]
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, false
	}
	nums := [3]int{}
	for i, p := range parts {
		n, ok := number(p)
		if !ok {
			return Version{}, false
		}
		nums[i] = n
	}
	var ids []string
	if pre != "" {
		ids = strings.Split(pre, ".")
		for _, id := range ids {
			if !identifier(id) {
				return Version{}, false
			}
		}
	}
	return Version{Major: nums[0], Minor: nums[1], Patch: nums[2], Pre: ids}, true
}

// Compare reports whether a is less than, equal to, or greater than b.
// The bool is false when either string is not a semantic version.
func Compare(a, b string) (int, bool) {
	av, aok := Parse(a)
	bv, bok := Parse(b)
	if !aok || !bok {
		return 0, false
	}
	return av.cmp(bv), true
}

// Explain says why value is not a semantic version.
// kind names the bound the value came from, and is empty for the package's
// own version. The bool is true when Parse accepts value.
func Explain(kind, value string) (string, bool) {
	if _, ok := Parse(value); ok {
		return "", true
	}
	if kind == "" {
		return value + " is not a semantic version", false
	}
	return kind + " " + value + " is not a semantic version", false
}

// cmp orders two versions by SemVer precedence: the numbers first, then a
// release above any of its pre-releases, then the pre-release identifiers
// one by one. A numeric identifier is below a word; numbers compare as
// numbers and words as text. When every shared identifier is equal, the
// longer list is higher.
func (a Version) cmp(b Version) int {
	if a.Major != b.Major {
		return sign(a.Major - b.Major)
	}
	if a.Minor != b.Minor {
		return sign(a.Minor - b.Minor)
	}
	if a.Patch != b.Patch {
		return sign(a.Patch - b.Patch)
	}
	switch {
	case len(a.Pre) == 0 && len(b.Pre) == 0:
		return 0
	case len(a.Pre) == 0:
		return 1
	case len(b.Pre) == 0:
		return -1
	}
	n := len(a.Pre)
	if len(b.Pre) < n {
		n = len(b.Pre)
	}
	for i := 0; i < n; i++ {
		an, aNum := number(a.Pre[i])
		bn, bNum := number(b.Pre[i])
		switch {
		case aNum && bNum && an != bn:
			return sign(an - bn)
		case aNum && bNum:
			continue
		case aNum != bNum:
			if aNum {
				return -1
			}
			return 1
		case a.Pre[i] != b.Pre[i]:
			if a.Pre[i] < b.Pre[i] {
				return -1
			}
			return 1
		}
	}
	return sign(len(a.Pre) - len(b.Pre))
}

// number reads a numeric field. A leading zero is refused, as SemVer
// refuses it. A value past MaxInt is refused, not wrapped: wrapping it
// would compare a larger field as a smaller one. The bound check is the
// one pep440's readInt uses, so both parsers stop at the same limit.
func number(s string) (int, bool) {
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return 0, false
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		d := int(c - '0')
		if n > (int(^uint(0)>>1)-d)/10 {
			return 0, false
		}
		n = n*10 + d
	}
	return n, true
}

// identifier reports whether s is a valid pre-release identifier: ASCII
// letters, digits and hyphens. An all-digit identifier is a numeric field,
// so number must accept it: no leading zero, and it must fit in int.
func identifier(s string) bool {
	if s == "" {
		return false
	}
	digits := true
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c == '-':
			digits = false
		default:
			return false
		}
	}
	if digits {
		if _, ok := number(s); !ok {
			return false
		}
	}
	return true
}

// sign maps a difference to -1, 0 or 1.
func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
