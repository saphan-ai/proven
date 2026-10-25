package testrun

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Refusal reasons for the test a model names. The trace prefixes each one with "refused: ".
const (
	ReasonModelCommand = "the answer carries a command; Proven builds it"
	ReasonBadName      = "not a test name"
	ReasonBadPackage   = "not a package path inside the checkout"
)

// namePattern is the whole of what a test name may be: Go's own shape for a
// test function, and nothing a shell or a regular expression would read as more.
var namePattern = regexp.MustCompile(`^Test([A-Z0-9_][A-Za-z0-9_]*)?$`)

// segmentPattern is one directory or file name of a test's path. It starts with a
// letter, a digit or an underscore, so it can be neither a flag nor "." nor "..".
var segmentPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*$`)

// RefuseTest reports why a named test must not run, or "" when Proven can build
// its command. A model names a test by its file and its name and nothing else
// that is executed (ADR-0025): the command comes from TestArgv, never from text
// a model wrote. It does not look at any checkout and starts nothing.
func RefuseTest(testPath, name string) string {
	if !namePattern.MatchString(name) {
		return ReasonBadName
	}
	if filepath.IsAbs(testPath) {
		return ReasonAbsolutePath
	}
	if pathLeaves(testPath) {
		return ReasonLeavesRepo
	}
	if !isGoTestFile(testPath) {
		return ReasonNotATestFile
	}
	return refusePackage(testPath)
}

// refusePackage checks every segment of a test's path against segmentPattern.
// The directory becomes an argument of the build tool, and a segment such as
// "..." or "-x" would mean something to it that a directory name does not.
func refusePackage(testPath string) string {
	for _, seg := range strings.Split(filepath.ToSlash(testPath), "/") {
		if !segmentPattern.MatchString(seg) || strings.Contains(seg, "..") {
			return ReasonBadPackage
		}
	}
	return ""
}

// TestArgv is the command that runs one Go test: the words after the image of
// the container run line, with no shell between them (ADR-0025). It is a fixed
// template — go test -run '^<name>$' ./<package> — filled from a name and a
// path that RefuseTest accepted. A non-empty reason means there is no command.
func TestArgv(testPath, name string) ([]string, string) {
	if reason := RefuseTest(testPath, name); reason != "" {
		return nil, reason
	}
	pkg := "."
	if dir := path.Dir(filepath.ToSlash(testPath)); dir != "." {
		pkg = "./" + dir
	}
	// ADR-0042: -json makes the output carry a pass event per test, which is how a
	// run shows the test ran; an exit code alone cannot show that.
	return []string{"go", "test", jsonFlag, "-run", "^" + name + "$", pkg}, ""
}

// displayCommand is the line the trace and the review show for an argv built by
// TestArgv. It is only text for a reader; nothing runs it.
func displayCommand(argv []string) string {
	shown := append([]string(nil), argv...)
	for i, word := range shown {
		if i > 0 && shown[i-1] == "-run" {
			shown[i] = "'" + word + "'"
		}
	}
	return strings.Join(shown, " ")
}
