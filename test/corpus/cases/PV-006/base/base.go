// Package base is the earlier version of the controls (a stand-in; see CASE.md §9).
//
// The text control refuses only the placeholder names it lists, each written
// in braces. There is no duration control: it always passes.
package base

import "strings"

var listed = []string{"{wait}", "{seconds}", "{limit}"}

// CheckText answers pass or fail for a rendered notice and names every listed
// raw placeholder found.
func CheckText(notice string) (pass bool, raw []string) {
	for _, p := range listed {
		if strings.Contains(notice, p) {
			raw = append(raw, p)
		}
	}
	return len(raw) == 0, raw
}

// CheckDuration always passes.
func CheckDuration(shown, retryAfter int) bool {
	return true
}
