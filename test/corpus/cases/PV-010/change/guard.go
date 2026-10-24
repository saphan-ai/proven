package main

import (
	"fmt"
	"regexp"
	"strconv"
)

// CHANGE: a closing number that differs is refused, an equal one is accepted,
// one that cannot be converted is refused.

// matcherMaxGap is the widest gap the matcher accepts.
const matcherMaxGap = 48

// statedMaxGap is the width the help page states. It is written here by hand;
// the matcher does not read it and nothing reads it back.
const statedMaxGap = 48

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
		if err != nil {
			stale = append(stale, n)
			continue
		}
		if v != pinned {
			stale = append(stale, n)
		}
	}
	return len(stale) == 0, stale
}

var classInPattern = regexp.MustCompile(`LANGUAGES\[([^\]]*)\]`)

// Describe is the help page's description of what the guard accepts between the
// name and the arrow. The characters are read from the matcher; the width is
// the stated constant.
func Describe() string {
	chars := classInPattern.FindStringSubmatch(matcher().String())[1]
	return fmt.Sprintf("Between the name and the arrow the guard accepts the characters [%s], up to %d characters wide.", chars, statedMaxGap)
}
