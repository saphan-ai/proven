// Package tool is a worked example of a tool adapter.
//
// Detector answers the four questions from a fixed rule: it applies to
// changed Go files, Argv builds `go test -race ./<package>` for each package
// of a changed file, and Read reports the race banner. A reader copies this
// package and replaces the rule with their own tool. What to add is in
// docs/ADAPTERS.md under "Adding a tool".
package tool

import (
	"path/filepath"
	"strings"

	it "github.com/saphan-ai/proven/internal/tool"
)

var _ it.Adapter = Detector{}

// Detector is a worked tool: a fixed detector over changed Go files.
type Detector struct{}

// ID is the tool's stable name.
func (Detector) ID() string { return "example-detector" }

// Version is the detector's fixed version; the runner never guesses it.
func (Detector) Version() string { return "1.0" }

// Applies reports true when one of the changed paths is a Go file.
func (Detector) Applies(paths []string) bool {
	for _, path := range paths {
		if strings.HasSuffix(path, ".go") {
			return true
		}
	}
	return false
}

// Argv builds `go test -race ./<package>` for each package of a changed Go
// file, with no shell. An absolute path or a path that names a parent
// directory is refused and nothing is started.
func (Detector) Argv(paths []string) ([]string, string) {
	var argv []string
	seen := map[string]bool{}
	for _, path := range paths {
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		if !validPath(path) {
			return nil, "example-detector: " + path + " is not a plain checkout path"
		}
		dir := filepath.Dir(path)
		if dir == "." {
			dir = "."
		}
		if seen[dir] {
			continue
		}
		seen[dir] = true
		argv = append(argv, "go", "test", "-race", "./"+dir)
	}
	if len(argv) == 0 {
		return nil, "example-detector: no Go file to test"
	}
	return argv, ""
}

// validPath reports that the path is relative, has no parent segment, and
// names a regular file name segment by segment.
func validPath(path string) bool {
	if path == "" || filepath.IsAbs(path) || strings.Contains(path, "..") {
		return false
	}
	for _, seg := range strings.Split(filepath.ToSlash(path), "/") {
		if seg == "" || seg == "." || seg == ".." {
			return false
		}
	}
	return true
}

// Read reports one side: the change races when its output carries the race
// banner. A test that fails without that line is not a race.
func (Detector) Read(output string) it.Reading {
	return it.Reading{Present: strings.Contains(output, "WARNING: DATA RACE"), Output: output}
}
