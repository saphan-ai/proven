// Package pep503 normalises a PyPI package name the way PEP 503 specifies.
// The dependency check compares requirements.txt with poetry.lock, and queries
// the OSV table, only after this. A difference of case or of separator is not
// a different package.
package pep503

import "strings"

// Normalise returns the PEP 503 form of a PyPI package name: lower case, and
// every run of `-`, `_` or `.` collapsed to one `-`.
// The OSV package table is keyed by this form, so a query for the spelling in
// the file would miss the row.
func Normalise(name string) string {
	// PEP 503 is collapse, then lower-case: re.sub(`[-_.]+`, "-", name).lower().
	// The result is the key the OSV table uses.
	var b strings.Builder
	b.Grow(len(name))
	run := false
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c == '-' || c == '_' || c == '.' {
			if run {
				continue
			}
			run = true
			b.WriteByte('-')
			continue
		}
		run = false
		b.WriteByte(c)
	}
	return strings.ToLower(b.String())
}
