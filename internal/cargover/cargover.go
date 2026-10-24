// Package cargover compares Cargo versions for the crates.io dependency check
// (ADR-0038).
//
// A Cargo version is a semantic version: MAJOR.MINOR.PATCH, an optional
// pre-release, and optional build metadata. Build metadata does not change
// order, so 1.2.3+build and 1.2.3 are one version. A pre-release is lower
// than the release it precedes. Precedence is the semantic order in
// internal/semver. Cargo versions are not Go versions: a leading v is
// refused, because a Cargo version is not written that way.
//
// A constraint (^, ~, >, <, =, *, a comma, ||, or a space) is refused. A
// string that is not three numeric components, such as 1.2, is refused. The
// refusal is a sentence. The version string stored from the manifest is not
// rewritten.
package cargover

import (
	"strings"

	"github.com/saphan-ai/proven/internal/semver"
)

// kind is why classify accepted or refused a string.
type kind int

const (
	kindOK kind = iota
	kindConstraint
	kindLeadingV
	kindSyntax
)

// Compare reports whether a is less than, equal to, or greater than b.
// The bool is false when either string is outside the subset.
func Compare(a, b string) (int, bool) {
	if classify(a) != kindOK || classify(b) != kindOK {
		return 0, false
	}
	return semver.Compare(strings.TrimSpace(a), strings.TrimSpace(b))
}

// Explain says why value is outside the subset.
// kind names the bound the value came from, and is empty for the package's
// own version. The bool is true when the value is a Cargo version.
func Explain(bound, value string) (string, bool) {
	if classify(value) == kindOK {
		return "", true
	}
	why := reject(value)
	if bound == "" {
		return why, false
	}
	return bound + " " + why, false
}

// reject says why s is not compared. It is empty when s is a Cargo version.
func reject(s string) string {
	shown := strings.TrimSpace(s)
	if shown == "" {
		shown = "empty"
	}
	switch classify(s) {
	case kindOK:
		return ""
	case kindConstraint:
		return shown + " is a constraint; this check compares one Cargo version"
	case kindLeadingV:
		return shown + " is not a Cargo version; a Cargo version has no leading v"
	default:
		return shown + " is not a Cargo version; this check compares one semantic version"
	}
}

// classify parses s and says which form it is. A constraint is named before
// a generic refusal, so the trace can say which.
func classify(s string) kind {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return kindSyntax
	}
	if isConstraint(raw) {
		return kindConstraint
	}
	// semver.Parse accepts a leading v for Go modules. Cargo does not write
	// one, and accepting it would order a different spelling as this version.
	if raw[0] == 'v' || raw[0] == 'V' {
		return kindLeadingV
	}
	if _, ok := semver.Parse(raw); !ok {
		return kindSyntax
	}
	return kindOK
}

// isConstraint reports a Cargo requirement rather than one version.
// The characters are the ones Cargo writes in a requirement and never in
// MAJOR.MINOR.PATCH.
func isConstraint(s string) bool {
	if strings.Contains(s, "||") {
		return true
	}
	return strings.ContainsAny(s, "^~><=*, ")
}
