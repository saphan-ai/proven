// Package composerver compares a subset of Composer versions for the
// Packagist dependency check (ADR-0038).
//
// The subset is a numbered version, not a constraint and not a branch name.
// One to four numeric components compare, and a missing component equals
// zero, so 1.2, 1.2.0 and 1.2.0.0 are one version. A leading v is ignored.
// Text after + is build metadata and is ignored. A leading zero on a
// component is still that number, so 1.02 and 1.2 are one version.
//
// Stability, from lowest to highest: dev, alpha (a), beta (b), RC (rc),
// stable, patch (p, pl, patch). The spelling is not case-sensitive. One
// separator before the suffix may be "-", "." or "_". The number may be
// glued on (RC1) or written after a separator (RC.1, RC-1). A suffix with
// no number is below the same suffix with number 0. An explicit "-stable"
// is the same as no suffix. A number written on stable is not compared.
//
// A constraint ("^", "~", ">", "<", "=", "*", ",", "||", a space, or the
// word " as ") and a branch name ("dev-*", or a component "x", including
// "1.0.x-dev") are refused. The refusal is a sentence. The version string
// stored from the manifest is not rewritten.
package composerver

import "strings"

// Version is one Composer version this package compares.
// Parts holds four numeric components. A version written with fewer is
// padded with zeros.
// Stab is the stability rank, from dev up to patch.
// StabN is the stability number, or -1 when the suffix names no number.
// A bare RC is below RC1.
type Version struct {
	Parts [4]int
	Stab  int
	StabN int
}

// Stability ranks. Numeric order is the comparison order.
const (
	stabDev = iota
	stabAlpha
	stabBeta
	stabRC
	stabStable
	stabPatch
)

// kind is why classify accepted or refused a string.
type kind int

const (
	kindOK kind = iota
	kindSyntax
	kindConstraint
	kindBranch
)

// stabWord is one stability spelling, longest first so "patch" wins over "p".
var stabWords = []struct {
	word string
	stab int
}{
	{"stable", stabStable},
	{"alpha", stabAlpha},
	{"patch", stabPatch},
	{"beta", stabBeta},
	{"dev", stabDev},
	{"rc", stabRC},
	{"pl", stabPatch},
	{"a", stabAlpha},
	{"b", stabBeta},
	{"p", stabPatch},
}

// Parse reads one version in the subset above. The bool is false when s
// is not in that subset.
func Parse(s string) (Version, bool) {
	v, k := classify(s)
	return v, k == kindOK
}

// Compare reports whether a is less than, equal to, or greater than b.
// The bool is false when either string is outside the subset.
func Compare(a, b string) (int, bool) {
	av, ak := classify(a)
	bv, bk := classify(b)
	if ak != kindOK || bk != kindOK {
		return 0, false
	}
	return av.cmp(bv), true
}

// Explain says why value is outside the subset.
// kind names the bound the value came from, and is empty for the package's
// own version. The bool is true when Parse accepts value.
func Explain(kind, value string) (string, bool) {
	if _, ok := Parse(value); ok {
		return "", true
	}
	why := Reject(value)
	if kind == "" {
		return why, false
	}
	return kind + " " + why, false
}

// Reject says why s is not compared. It is empty when Parse accepts s.
func Reject(s string) string {
	shown := strings.TrimSpace(s)
	if shown == "" {
		shown = "empty"
	}
	switch _, k := classify(s); k {
	case kindOK:
		return ""
	case kindConstraint:
		return shown + " is a constraint; this check compares one Composer version"
	case kindBranch:
		return shown + " is a branch name; this check compares numbered Composer versions only"
	default:
		return shown + " is not a Composer version this check compares"
	}
}

// classify parses s and says which form it is. Constraints and branch names
// are named before a generic refusal, so the trace can say which.
func classify(s string) (Version, kind) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return Version{}, kindSyntax
	}
	low := strings.ToLower(raw)
	if isConstraint(low) {
		return Version{}, kindConstraint
	}
	if isBranch(low) {
		return Version{}, kindBranch
	}
	body := low
	// A leading v is how a tag is often written. It is not part of the number.
	if len(body) >= 2 && body[0] == 'v' && body[1] >= '0' && body[1] <= '9' {
		body = body[1:]
	}
	// Build metadata does not change order. An empty suffix is not metadata.
	if i := strings.IndexByte(body, '+'); i >= 0 {
		if i == 0 || body[i+1:] == "" {
			return Version{}, kindSyntax
		}
		body = body[:i]
	}
	parts, rest, ok := splitNumeric(body)
	if !ok {
		return Version{}, kindSyntax
	}
	stab, stabN, ok := splitStability(rest)
	if !ok {
		return Version{}, kindSyntax
	}
	return Version{Parts: parts, Stab: stab, StabN: stabN}, kindOK
}

// isConstraint reports a Composer constraint rather than one version.
// The check does not guess which version a range would install.
func isConstraint(low string) bool {
	if strings.ContainsAny(low, "^~><=,*| \t") {
		return true
	}
	return strings.Contains(low, " as ")
}

// isBranch reports a branch name. dev-* and a component x are branches.
// 1.0.x-dev is the branch form of the 1.0 line, not version 1.0.
func isBranch(low string) bool {
	body := low
	if len(body) >= 2 && body[0] == 'v' && body[1] >= '0' && body[1] <= '9' {
		body = body[1:]
	}
	if strings.HasPrefix(body, "dev-") {
		return true
	}
	if body == "x" || strings.HasPrefix(body, "x.") || strings.HasSuffix(body, "x-dev") {
		return true
	}
	return strings.Contains(body, ".x.") || strings.Contains(body, ".x-") || strings.Contains(body, ".x+")
}

// splitNumeric reads one to four components. The rest is the stability
// suffix, including its separator. A fifth component is not in the subset.
func splitNumeric(s string) (parts [4]int, rest string, ok bool) {
	i := 0
	n := 0
	for n < 4 {
		if i >= len(s) || s[i] < '0' || s[i] > '9' {
			break
		}
		j := i
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		v, good := atoi(s[i:j])
		if !good {
			return parts, "", false
		}
		parts[n] = v
		n++
		i = j
		// Another component only when a dot is followed by a digit.
		// "1.2.3-rc1" stops here, and the hyphen stays in the suffix.
		if n < 4 && i+1 < len(s) && s[i] == '.' && s[i+1] >= '0' && s[i+1] <= '9' {
			i++
			continue
		}
		break
	}
	if n == 0 {
		return parts, "", false
	}
	return parts, s[i:], true
}

// splitStability reads the suffix. An empty suffix is stable.
func splitStability(rest string) (int, int, bool) {
	if rest == "" {
		return stabStable, -1, true
	}
	if rest[0] == '.' || rest[0] == '-' || rest[0] == '_' {
		rest = rest[1:]
		if rest == "" {
			return 0, 0, false
		}
	}
	word, stab, ok := stabilityWord(rest)
	if !ok {
		return 0, 0, false
	}
	rest = rest[len(word):]
	// "-stable1" is not a form this subset orders.
	if stab == stabStable {
		return stabStable, -1, rest == ""
	}
	if rest == "" {
		return stab, -1, true
	}
	if rest[0] == '.' || rest[0] == '-' || rest[0] == '_' {
		rest = rest[1:]
		if rest == "" {
			return 0, 0, false
		}
	}
	n, ok := atoi(rest)
	if !ok {
		return 0, 0, false
	}
	return stab, n, true
}

// stabilityWord returns the longest spelling that starts s.
func stabilityWord(s string) (string, int, bool) {
	for _, st := range stabWords {
		if strings.HasPrefix(s, st.word) {
			return st.word, st.stab, true
		}
	}
	return "", 0, false
}

// atoi reads a decimal integer. A leading zero is kept as a number.
// A value that does not fit in int is refused, not wrapped.
func atoi(s string) (int, bool) {
	if s == "" {
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

// cmp orders numeric components first, then stability, then the stability
// number. An absent number (-1) is below 0.
func (a Version) cmp(b Version) int {
	for i := 0; i < 4; i++ {
		if a.Parts[i] != b.Parts[i] {
			return sign(a.Parts[i] - b.Parts[i])
		}
	}
	if a.Stab != b.Stab {
		return sign(a.Stab - b.Stab)
	}
	return sign(a.StabN - b.StabN)
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
