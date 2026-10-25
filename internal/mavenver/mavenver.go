// Package mavenver compares a subset of Maven versions for the Maven
// dependency check (ADR-0038).
//
// The subset is a numbered version and an optional qualifier. Dot and hyphen
// both separate tokens, and a letter after a digit starts a new token, so
// 1.0-rc1 and 1.0-rc-1 are one version. Numeric tokens compare as integers.
// A leading zero does not change the number, and a missing numeric token
// equals zero, so 1, 1.0 and 1.0.0 are one version. A hyphen between numbers
// is the same separator as a dot in this subset, so 1.0-1 and 1.0.1 are one
// version.
//
// Qualifiers, from lowest to highest, are the Maven order: alpha, beta,
// milestone, rc, snapshot, release, sp. The spelling is not case-sensitive.
// Aliases: a = alpha, b = beta, m = milestone, cr = rc, and ga, final and
// release are the release itself, so 1.0-ga and 1.0 are one version. A
// qualifier below release is below the same numbers with no qualifier. sp
// is above that release. A number on the qualifier orders the same
// qualifier, and a missing number equals zero, so 1.0-alpha and
// 1.0-alpha-0 are one version and both are below 1.0-alpha-1.
//
// A qualifier that is not in that list is above sp. Two such qualifiers
// compare as lowercase text, so 1.2-jre and 1.3-jre still order by the
// numbers. This placement was not checked against one named Maven release
// (UNVERIFIED). The known order above is the order the Maven version
// specification publishes.
//
// A constraint ("[", "]", "(", ")", ",", "+", "*", "^", "~", ">", "<", "=",
// a space) is refused. A leading v is refused. The refusal is a sentence.
// The version string stored from the manifest is not rewritten.
package mavenver

import "strings"

// Version is one Maven version this package compares.
// Nums is the numeric tokens, with no trailing zeros removed: comparison
// treats a missing token as zero, which is the same result.
// Rank is the qualifier order. Word is the canonical qualifier, empty for
// a release, and the lowercase text of an unknown qualifier.
// QualN is the number written on the qualifier. A qualifier with no number
// has zero.
type Version struct {
	Nums  []int
	Rank  int
	Word  string
	QualN int
}

// Qualifier ranks. Numeric order is the comparison order.
const (
	rankAlpha = iota
	rankBeta
	rankMilestone
	rankRC
	rankSnapshot
	rankRelease
	rankSP
	rankUnknown
)

// kind is why classify accepted or refused a string.
type kind int

const (
	kindOK kind = iota
	kindSyntax
	kindConstraint
	kindLeadingV
)

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
// bound names the bound the value came from, and is empty for the package's
// own version. The bool is true when Parse accepts value.
func Explain(bound, value string) (string, bool) {
	if _, ok := Parse(value); ok {
		return "", true
	}
	why := Reject(value)
	if bound == "" {
		return why, false
	}
	return bound + " " + why, false
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
		return shown + " is a constraint; this check compares one Maven version"
	case kindLeadingV:
		return shown + " is not a Maven version; a Maven version has no leading v"
	default:
		return shown + " is not a Maven version this check compares"
	}
}

// classify parses s and says which form it is. A constraint and a leading v
// are named before a generic refusal, so the trace can say which.
func classify(s string) (Version, kind) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return Version{}, kindSyntax
	}
	// A leading v is how some tags are written. A Maven version is not.
	if raw[0] == 'v' || raw[0] == 'V' {
		return Version{}, kindLeadingV
	}
	if isConstraint(raw) {
		return Version{}, kindConstraint
	}
	toks, ok := tokenize(strings.ToLower(raw))
	if !ok {
		return Version{}, kindSyntax
	}
	v, ok := fromTokens(toks)
	if !ok {
		return Version{}, kindSyntax
	}
	return v, kindOK
}

// isConstraint reports a range or a dynamic version rather than one version.
// Gradle writes "+" in a dynamic version. The check does not guess which
// version that would select.
func isConstraint(s string) bool {
	return strings.ContainsAny(s, "^~><=*,[](){}+| \t")
}

// token is one numeric or textual piece of a version.
type token struct {
	num  bool
	n    int
	word string
}

// tokenize splits s on '.', '-' and a change between a digit and a letter.
// An empty piece between two separators is zero, which is how "1..0" is read.
// A leading separator is not a version.
func tokenize(s string) ([]token, bool) {
	if s == "" || s[0] == '.' || s[0] == '-' {
		return nil, false
	}
	var out []token
	i := 0
	for i < len(s) {
		c := s[i]
		if c == '.' || c == '-' {
			i++
			if i >= len(s) {
				break
			}
			if s[i] == '.' || s[i] == '-' {
				out = append(out, token{num: true, n: 0})
			}
			continue
		}
		if c >= '0' && c <= '9' {
			j := i
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}
			n, ok := atoi(s[i:j])
			if !ok {
				return nil, false
			}
			out = append(out, token{num: true, n: n})
			i = j
			continue
		}
		if c >= 'a' && c <= 'z' {
			j := i
			for j < len(s) && s[j] >= 'a' && s[j] <= 'z' {
				j++
			}
			out = append(out, token{word: s[i:j]})
			i = j
			continue
		}
		return nil, false
	}
	return out, len(out) > 0
}

// fromTokens takes the leading numbers as the version and at most one
// qualifier, with its number, after them. A second qualifier is not in
// the subset: guessing an order for it would hide a version the table
// cannot place.
func fromTokens(toks []token) (Version, bool) {
	i := 0
	var nums []int
	for i < len(toks) && toks[i].num {
		nums = append(nums, toks[i].n)
		i++
	}
	if len(nums) == 0 {
		return Version{}, false
	}
	v := Version{Nums: nums, Rank: rankRelease}
	if i >= len(toks) {
		return v, true
	}
	word, rank := canon(toks[i].word)
	i++
	qn := 0
	if i < len(toks) {
		if !toks[i].num {
			return Version{}, false
		}
		qn = toks[i].n
		i++
	}
	if i != len(toks) {
		return Version{}, false
	}
	v.Rank = rank
	v.Word = word
	v.QualN = qn
	return v, true
}

// canon returns the qualifier's canonical spelling and its rank.
// An unknown word keeps its letters and ranks above sp.
func canon(word string) (string, int) {
	switch word {
	case "a", "alpha":
		return "alpha", rankAlpha
	case "b", "beta":
		return "beta", rankBeta
	case "m", "milestone":
		return "milestone", rankMilestone
	case "cr", "rc":
		return "rc", rankRC
	case "snapshot":
		return "snapshot", rankSnapshot
	case "ga", "final", "release":
		return "", rankRelease
	case "sp":
		return "sp", rankSP
	default:
		return word, rankUnknown
	}
}

// cmp orders two parsed versions: numbers, then qualifier rank, then the
// qualifier's number, then the unknown qualifier's letters.
func (v Version) cmp(o Version) int {
	n := len(v.Nums)
	if len(o.Nums) > n {
		n = len(o.Nums)
	}
	for i := 0; i < n; i++ {
		a, b := 0, 0
		if i < len(v.Nums) {
			a = v.Nums[i]
		}
		if i < len(o.Nums) {
			b = o.Nums[i]
		}
		if a != b {
			if a < b {
				return -1
			}
			return 1
		}
	}
	if v.Rank != o.Rank {
		if v.Rank < o.Rank {
			return -1
		}
		return 1
	}
	if v.QualN != o.QualN {
		if v.QualN < o.QualN {
			return -1
		}
		return 1
	}
	if v.Word != o.Word {
		if v.Word < o.Word {
			return -1
		}
		return 1
	}
	return 0
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
