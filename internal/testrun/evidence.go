package testrun

import (
	"encoding/json"
	"strings"
)

// A run counts as passed only when the output shows the test ran and passed, and
// not merely because the process ended with exit code 0 (ADR-0042). Code in the
// package can end the process early with exit 0 before any test starts, so an
// exit code alone cannot say that the named test ran. Every proof run is built
// by TestArgv or PackageArgv with -json, and the pass events that go test prints
// are the evidence. The program under test cannot forge them: the go tool
// encodes everything the program prints into an output event.

// jsonFlag is the argument that makes go test print one JSON event per line.
const jsonFlag = "-json"

// testEvent is the part of a go test -json event this judgement reads.
type testEvent struct {
	Action  string
	Package string
	Test    string
	Output  string
}

// isProofArgv reports whether argv is a go test run built with -json, whose
// output the judgement below can read. Other commands keep their exit-code result.
func isProofArgv(argv []string) bool {
	if len(argv) < 3 || argv[0] != "go" || argv[1] != "test" {
		return false
	}
	for _, w := range argv[2:] {
		if w == jsonFlag {
			return true
		}
	}
	return false
}

// provingName is the test a proof argv selects with -run '^Name$', or "" for a
// run of the whole package. The words are the ones TestArgv built, so the name
// holds only the characters RefuseTest allowed.
func provingName(words []string) string {
	for i, w := range words {
		if w == "-run" && i+1 < len(words) {
			return strings.Trim(words[i+1], "'^$")
		}
	}
	return ""
}

// parseEvents reads the events of a go test -json output and the text a person
// reads, which is what the events carry in their Output fields. A line that is
// not an event (a build error, a runtime message) stays as it is.
func parseEvents(out string) ([]testEvent, string) {
	var events []testEvent
	var text strings.Builder
	for _, line := range strings.SplitAfter(out, "\n") {
		trimmed := strings.TrimSpace(line)
		var ev testEvent
		if strings.HasPrefix(trimmed, "{") && json.Unmarshal([]byte(trimmed), &ev) == nil && ev.Action != "" {
			events = append(events, ev)
			text.WriteString(ev.Output)
			continue
		}
		text.WriteString(line)
	}
	return events, text.String()
}

// ranAndPassed reports whether the events show a pass for the run argv asked for.
// A named test needs a pass event of that test. A package run needs the package's
// pass event and at least one test that passed, so a package with no test to run
// is not a pass. A skipped test has a skip event, never a pass event.
func ranAndPassed(name string, events []testEvent) bool {
	pkgPassed, testPassed := false, false
	for _, ev := range events {
		if ev.Action != "pass" {
			continue
		}
		switch {
		case ev.Test == "":
			pkgPassed = true
		case name == "" || ev.Test == name:
			testPassed = true
		}
	}
	if name != "" {
		return testPassed
	}
	return pkgPassed && testPassed
}

// judgeProof turns the output of a proof run that ended with exit code 0 into a
// result. It returns the text to show and whether the run counts as passed. When
// it does not, the text says why, so a reader sees a run that proves nothing.
func judgeProof(argv []string, out string) (string, bool) {
	events, text := parseEvents(out)
	if ranAndPassed(provingName(argv), events) {
		return text, true
	}
	what := "the package"
	if name := provingName(argv); name != "" {
		what = "the test " + name
	}
	note := "the run ended with exit code 0, but its output has no pass event for " + what + ": it did not run, and a run that did not run is not a pass"
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return text + note, false
}

// RecordedPass reports whether the output of a recorded proof run, in the text a
// person reads (--- PASS: lines and an ok line, as go test prints them), shows
// the run passed. Sample replays recorded runs, and it applies the judgement the
// container applies, so a recording that never showed a pass is not replayed as one.
func RecordedPass(command, output string) bool {
	words := strings.Fields(command)
	name := provingName(words)
	passed, pkgOK := false, false
	for _, line := range strings.Split(output, "\n") {
		switch {
		case strings.HasPrefix(line, "--- PASS: "):
			if name == "" || strings.HasPrefix(line, "--- PASS: "+name+" ") {
				passed = true
			}
		case strings.HasPrefix(line, "ok ") || strings.HasPrefix(line, "ok\t"):
			pkgOK = true
		}
	}
	if name != "" {
		return passed
	}
	return passed && pkgOK
}
