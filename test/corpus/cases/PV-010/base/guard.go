package main

import (
	"fmt"
	"regexp"
	"strconv"
)

// BASE: the guard before the change. A closing number it cannot convert is
// treated as current.

const matcherMaxGap = 48

var gapClass = "0-9 "

// matcher finds an arrow pair: the name, a gap, the arrow, then a run of ASCII
// decimal digits. A sign or digits of another script find no pair.
func matcher() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`LANGUAGES[%s]{0,%d}-> ?([0-9]+)`, gapClass, matcherMaxGap))
}

// Check answers pass or fail for the notes against the pinned count and names
// the stale notes.
func Check(notes []string, pinned int64) (bool, []string) {
	m := matcher()
	var stale []string
	for _, n := range notes {
		sub := m.FindStringSubmatch(n)
		if sub == nil {
			continue
		}
		v, err := strconv.ParseInt(sub[1], 10, 64)
		if err != nil || v == pinned {
			continue
		}
		stale = append(stale, n)
	}
	return len(stale) == 0, stale
}

var classInPattern = regexp.MustCompile(`LANGUAGES\[([^\]]*)\]`)

// Describe is the help page's description of what the guard accepts between the
// name and the arrow. It names the accepted characters only.
func Describe() string {
	chars := classInPattern.FindStringSubmatch(matcher().String())[1]
	return fmt.Sprintf("Between the name and the arrow the guard accepts the characters [%s].", chars)
}
