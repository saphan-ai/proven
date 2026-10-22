package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/saphan-ai/proven/tests/corpus/cases/PV-009/base"
	"github.com/saphan-ai/proven/tests/corpus/cases/PV-009/change"
)

var referenceArrow = regexp.MustCompile(`->\s*([0-9]+)`)

func reference(notes []string, pinned int64) []string {
	var stale []string
	for _, note := range notes {
		_, after, found := strings.Cut(note, "LANGUAGES")
		if !found {
			continue
		}
		for _, pair := range referenceArrow.FindAllStringSubmatch(after, -1) {
			if pair[1] != strconv.FormatInt(pinned, 10) {
				stale = append(stale, note)
				break
			}
		}
	}
	return stale
}

func result(stale []string) string {
	if len(stale) == 0 {
		return "pass"
	}
	return "fail:" + strings.Join(stale, "|")
}

func main() {
	inputs := []struct {
		name  string
		notes []string
	}{
		{"one-current", []string{"LANGUAGES 41 -> 42"}},
		{"one-stale", []string{"LANGUAGES 41 -> 43"}},
		{"stale-after-current", []string{"LANGUAGES 41 -> 42; 42 -> 43"}},
		{"current-after-stale", []string{"LANGUAGES 41 -> 43; 43 -> 42"}},
		{"three-pairs", []string{"LANGUAGES 40 -> 42; 41 -> 42; 42 -> 43"}},
		{"no-arrow", []string{"LANGUAGES are available"}},
		{"max-int64", []string{"LANGUAGES 41 -> 9223372036854775807"}},
		{"overflow-int64", []string{"LANGUAGES 41 -> 9223372036854775808"}},
		{"twenty-digits", []string{"LANGUAGES 41 -> 10000000000000000000"}},
	}
	for _, input := range inputs {
		fmt.Printf("%s base=%s change=%s reference=%s\n", input.name,
			result(base.Stale(input.notes, 42)), result(change.Stale(input.notes, 42)), result(reference(input.notes, 42)))
	}
}
