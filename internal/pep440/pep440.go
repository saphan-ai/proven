// Package pep440 compares the PEP 440 forms this check accepts:
// a release, a pre-release and a post-release.
// An epoch, a dev release, a local version or any other string is refused.
// The refusal is a sentence. The version is not rewritten.
package pep440

import "strings"

// Version is one release with an optional pre-release and post-release.
// PreKind is 0 when there is no pre-release, then 1 for a, 2 for b, 3 for rc.
type Version struct {
	Release []int
	PreKind int
	PreNum  int
	HasPost bool
	Post    int
}

type kind int

const (
	kindOK kind = iota
	kindSyntax
	kindEpoch
	kindLocal
	kindWild
	kindDev
)

// Parse reads a release, pre-release or post-release. The bool is false
// when s is not one of those forms.
func Parse(s string) (Version, bool) {
	v, k := classify(s)
	return v, k == kindOK
}

// Compare reports whether a is less than, equal to, or greater than b.
// The bool is false when either string is not a form this package accepts.
func Compare(a, b string) (int, bool) {
	av, ak := classify(a)
	bv, bk := classify(b)
	if ak != kindOK || bk != kindOK {
		return 0, false
	}
	return av.cmp(bv), true
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
	case kindEpoch:
		return shown + " has an epoch; this check compares release, pre-release and post-release only"
	case kindLocal:
		return shown + " has a local version; this check does not compare it"
	case kindDev:
		return shown + " has a dev release; this check does not compare it"
	case kindWild:
		return shown + " is not a final release, pre-release or post-release"
	default:
		return shown + " is not a release, pre-release or post-release"
	}
}

func classify(s string) (Version, kind) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Version{}, kindSyntax
	}
	low := strings.ToLower(s)
	if strings.Contains(low, "!") {
		return Version{}, kindEpoch
	}
	if strings.Contains(low, "+") {
		return Version{}, kindLocal
	}
	if strings.Contains(low, "*") {
		return Version{}, kindWild
	}
	if strings.Contains(low, "dev") {
		return Version{}, kindDev
	}
	if low[0] == 'v' {
		low = low[1:]
	}
	rel, rest, ok := parseRelease(low)
	if !ok {
		return Version{}, kindSyntax
	}
	v := Version{Release: rel}
	if rest == "" {
		return v, kindOK
	}
	if !looksLikePost(rest) {
		kindN, num, next, took := takePre(rest)
		if !took {
			return Version{}, kindSyntax
		}
		v.PreKind = kindN
		v.PreNum = num
		rest = next
		if rest == "" {
			return v, kindOK
		}
	}
	n, next, ok := parsePost(rest)
	if !ok || next != "" {
		return Version{}, kindSyntax
	}
	v.HasPost = true
	v.Post = n
	return v, kindOK
}

func (a Version) cmp(b Version) int {
	n := len(a.Release)
	if len(b.Release) > n {
		n = len(b.Release)
	}
	for i := 0; i < n; i++ {
		av, bv := 0, 0
		if i < len(a.Release) {
			av = a.Release[i]
		}
		if i < len(b.Release) {
			bv = b.Release[i]
		}
		if av != bv {
			return sign(av - bv)
		}
	}
	if c := cmpPre(a, b); c != 0 {
		return c
	}
	return cmpPost(a, b)
}

func cmpPre(a, b Version) int {
	if a.PreKind == 0 && b.PreKind == 0 {
		return 0
	}
	if a.PreKind == 0 {
		return 1
	}
	if b.PreKind == 0 {
		return -1
	}
	if a.PreKind != b.PreKind {
		return sign(a.PreKind - b.PreKind)
	}
	return sign(a.PreNum - b.PreNum)
}

func cmpPost(a, b Version) int {
	if !a.HasPost && !b.HasPost {
		return 0
	}
	if !a.HasPost {
		return -1
	}
	if !b.HasPost {
		return 1
	}
	return sign(a.Post - b.Post)
}

func parseRelease(s string) ([]int, string, bool) {
	if s == "" || !isDigit(s[0]) {
		return nil, s, false
	}
	var rel []int
	for {
		n, next, ok := readInt(s)
		if !ok {
			return nil, s, false
		}
		rel = append(rel, n)
		s = next
		if s == "" || s[0] != '.' || len(s) == 1 || !isDigit(s[1]) {
			break
		}
		s = s[1:]
	}
	return rel, s, true
}

func looksLikePost(rest string) bool {
	if rest == "" {
		return false
	}
	if rest[0] == '-' && len(rest) > 1 && isDigit(rest[1]) {
		return true
	}
	t := rest
	if t[0] == '.' || t[0] == '-' || t[0] == '_' {
		t = t[1:]
	}
	return postLabelLen(t) > 0
}

func takePre(rest string) (int, int, string, bool) {
	if rest == "" {
		return 0, 0, rest, false
	}
	t := rest
	if t[0] == '.' || t[0] == '-' || t[0] == '_' {
		t = t[1:]
	}
	label, n := preLabel(t)
	if n == 0 {
		return 0, 0, rest, false
	}
	num, next, ok := takeNum(t[n:])
	if !ok {
		return 0, 0, rest, false
	}
	return label, num, next, true
}

func takeNum(s string) (int, string, bool) {
	if s == "" {
		return 0, "", true
	}
	if s[0] == '.' || s[0] == '-' || s[0] == '_' {
		if len(s) > 1 && isDigit(s[1]) {
			return readInt(s[1:])
		}
		return 0, s, true
	}
	if isDigit(s[0]) {
		return readInt(s)
	}
	return 0, s, false
}

func parsePost(rest string) (int, string, bool) {
	if rest == "" {
		return 0, "", false
	}
	if rest[0] == '-' && len(rest) > 1 && isDigit(rest[1]) {
		return readInt(rest[1:])
	}
	t := rest
	if t[0] == '.' || t[0] == '-' || t[0] == '_' {
		t = t[1:]
		if t == "" {
			return 0, rest, false
		}
	}
	n := postLabelLen(t)
	if n == 0 {
		return 0, rest, false
	}
	t = t[n:]
	if t == "" {
		return 0, "", true
	}
	if t[0] == '.' || t[0] == '-' || t[0] == '_' {
		if len(t) == 1 || !isDigit(t[1]) {
			return 0, t, false
		}
		return readInt(t[1:])
	}
	if isDigit(t[0]) {
		return readInt(t)
	}
	return 0, t, false
}

func preLabel(t string) (int, int) {
	labels := []struct {
		text string
		kind int
	}{
		{"preview", 3},
		{"alpha", 1},
		{"beta", 2},
		{"rc", 3},
		{"pre", 3},
		{"c", 3},
		{"a", 1},
		{"b", 2},
	}
	for _, l := range labels {
		if strings.HasPrefix(t, l.text) && boundary(t[len(l.text):]) {
			return l.kind, len(l.text)
		}
	}
	return 0, 0
}

func postLabelLen(t string) int {
	switch {
	case strings.HasPrefix(t, "post") && boundary(t[4:]):
		return 4
	case strings.HasPrefix(t, "rev") && boundary(t[3:]):
		return 3
	case strings.HasPrefix(t, "rc"):
		return 0
	case strings.HasPrefix(t, "r") && boundary(t[1:]):
		return 1
	default:
		return 0
	}
}

func boundary(rest string) bool {
	if rest == "" {
		return true
	}
	c := rest[0]
	return isDigit(c) || c == '.' || c == '-' || c == '_'
}

func readInt(s string) (int, string, bool) {
	if s == "" || !isDigit(s[0]) {
		return 0, s, false
	}
	n := 0
	i := 0
	for i < len(s) && isDigit(s[i]) {
		d := int(s[i] - '0')
		if n > (int(^uint(0)>>1)-d)/10 {
			return 0, s, false
		}
		n = n*10 + d
		i++
	}
	return n, s[i:], true
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

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
