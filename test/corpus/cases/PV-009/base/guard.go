package base

import (
	"regexp"
	"strconv"
	"strings"
)

var arrow = regexp.MustCompile(`[0-9]+\s*->\s*([0-9]+)`)

func Stale(notes []string, pinned int64) []string {
	var stale []string
	for _, note := range notes {
		_, after, found := strings.Cut(note, "LANGUAGES")
		if !found {
			continue
		}
		pair := arrow.FindStringSubmatch(after)
		if pair == nil {
			continue
		}
		closing, err := strconv.ParseInt(pair[1], 10, 64)
		if err != nil || closing == pinned {
			continue
		}
		stale = append(stale, note)
	}
	return stale
}
