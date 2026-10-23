package guardrail

import (
	"errors"
	"strings"
)

// errVerdict means the completion did not carry the measured verdict lines.
var errVerdict = errors.New("guardrail: verdict lines are missing or not safe or unsafe")

// Parse reads the measured verdict lines.
// A </think> block may come first; those lines are not the verdict.
// wantResponse is set when the check sent an answer, and then Response Safety is required.
func Parse(text string, wantResponse bool) (Verdict, error) {
	// The verdict follows the last </think>; a reasoning block before it may quote the verdict words.
	if i := strings.LastIndex(text, "</think>"); i >= 0 {
		text = text[i+len("</think>"):]
	}
	var user, response string
	var cats []string
	sawUser := false
	sawResponse := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(line, "User Safety:"); ok {
			user = strings.TrimSpace(rest)
			sawUser = true
			continue
		}
		if rest, ok := strings.CutPrefix(line, "Response Safety:"); ok {
			response = strings.TrimSpace(rest)
			sawResponse = true
			continue
		}
		if rest, ok := strings.CutPrefix(line, "Safety Categories:"); ok {
			cats = splitCategories(rest)
		}
	}
	// Anything other than the exact words "safe" or "unsafe" is refused, so an
	// unexpected answer is never read as safe.
	if !sawUser || !safeWord(user) {
		return Verdict{}, errVerdict
	}
	if sawResponse && !safeWord(response) {
		return Verdict{}, errVerdict
	}
	if wantResponse && !sawResponse {
		return Verdict{}, errVerdict
	}
	v := Verdict{User: user, Categories: cats}
	if sawResponse {
		v.Response = response
	}
	// Categories only explain an unsafe verdict; a safe one carries none.
	if !v.Unsafe() {
		v.Categories = nil
	}
	return v, nil
}

// safeWord reports whether s is one of the two measured verdict words.
func safeWord(s string) bool {
	return s == "safe" || s == "unsafe"
}

// splitCategories reads a comma-separated category line, dropping empty entries.
func splitCategories(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
