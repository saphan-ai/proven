package tool

import (
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// raceBanner is the line the Go race detector prints when it reports a race.
// Measured in the race runtime of go1.26.8, the toolchain this module's go
// line selects. A test that fails without this line is not a race.
const raceBanner = "WARNING: DATA RACE"

// segmentPattern is one directory name that may become an argument of go test.
// It cannot be a flag, and it cannot be "." or ".." or "...".
var segmentPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*$`)

// Race is the Go race detector, the first tool (ADR-0040).
// Its command is go test -race on the packages of the changed Go files.
type Race struct{}

// ID is the tool's stable name.
func (Race) ID() string { return "go-race" }

// Version is empty: a live run does not start a second container to ask,
// and the runner records unknown. A sample scenario carries its own version.
func (Race) Version() string { return "" }

// Applies reports whether any changed path is a Go file inside the checkout.
// A change with no Go file is not a race the detector can see, so the tool
// is not run.
func (Race) Applies(paths []string) bool {
	for _, p := range paths {
		if isGoFile(p) {
			return true
		}
	}
	return false
}

// Argv is go test -race on each changed package, with no shell and no -run
// filter: the detector reports a race in any test of the package (ADR-0040).
// A path that could mean more than a directory to go test is refused.
func (Race) Argv(paths []string) ([]string, string) {
	pkgs, reason := racePackages(paths)
	if reason != "" {
		return nil, reason
	}
	if len(pkgs) == 0 {
		return nil, "no Go package"
	}
	argv := []string{"go", "test", "-race"}
	return append(argv, pkgs...), ""
}

// Read reports a race when the output carries the detector's banner.
// The comparison is exact: a different spelling is not the detector's line.
func (Race) Read(output string) Reading {
	return Reading{Present: strings.Contains(output, raceBanner), Output: output}
}

// isGoFile reports a relative Go path. An absolute path is not offered to
// Argv. A path that leaves the checkout still applies, so Argv can refuse it
// by name instead of the runner treating it as "does not apply".
func isGoFile(p string) bool {
	return strings.HasSuffix(p, ".go") && !filepath.IsAbs(p)
}

// racePackages lists the packages of the Go files, sorted, or a reason.
// The directory becomes an argument of go test, so a segment such as "..."
// or "-x" is refused: it would mean something a directory name does not.
func racePackages(paths []string) ([]string, string) {
	seen := map[string]bool{}
	var pkgs []string
	for _, p := range paths {
		if !strings.HasSuffix(p, ".go") {
			continue
		}
		if filepath.IsAbs(p) || strings.Contains(p, "..") {
			return nil, "not a package path inside the checkout"
		}
		slash := filepath.ToSlash(p)
		dir := path.Dir(slash)
		pkg := "."
		if dir != "." {
			for _, seg := range strings.Split(dir, "/") {
				if seg == "..." || strings.Contains(seg, "..") || !segmentPattern.MatchString(seg) {
					return nil, "not a package path inside the checkout"
				}
			}
			pkg = "./" + dir
		}
		if !seen[pkg] {
			seen[pkg] = true
			pkgs = append(pkgs, pkg)
		}
	}
	sort.Strings(pkgs)
	return pkgs, ""
}
