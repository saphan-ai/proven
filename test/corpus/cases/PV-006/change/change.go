// Package change is the version of the controls under review.
//
// The text control refuses any token in braces, with the same pattern the
// product uses to fill. The duration control passes when the duration shown
// equals the retry-after value or is up to two seconds short of it.
package change

import "regexp"

var token = regexp.MustCompile(`\{[^{}]+\}`)

// CheckText answers pass or fail for a rendered notice and names every raw
// token in braces found.
func CheckText(notice string) (pass bool, raw []string) {
	raw = token.FindAllString(notice, -1)
	return len(raw) == 0, raw
}

// CheckDuration passes when shown equals retryAfter or is up to two seconds
// short of it.
func CheckDuration(shown, retryAfter int) bool {
	return shown <= retryAfter && retryAfter-shown <= 2
}
