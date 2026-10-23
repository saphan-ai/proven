package manifest

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/semver"
)

// GoSum reads go.sum. A line is a module path, a version and an h1 hash.
// A version that ends in /go.mod is the checksum of that module's go.mod
// file, so it is not a module version. One module may appear at more than
// one version. Each of those versions is kept.
func GoSum() Reader {
	return Reader{
		Ecosystem: "Go",
		Lock:      true,
		NoteErr:   true,
		Match: func(path string) bool {
			return path == "go.sum" || strings.HasSuffix(path, "/go.sum")
		},
		Diff: diffGoSum,
	}
}

// diffGoSum compares two go.sum files. The error names the side that did
// not parse, because Collect writes it into the trace as it is.
func diffGoSum(base, change []byte) ([]Change, error) {
	before, err := parseGoSum(base)
	if err != nil {
		return nil, fmt.Errorf("base: %w", err)
	}
	after, err := parseGoSum(change)
	if err != nil {
		return nil, fmt.Errorf("change: %w", err)
	}
	return diffVersions(before, after, semver.Compare), nil
}

// parseGoSum reads every module version a go.sum pins.
func parseGoSum(b []byte) (versions, error) {
	out := versions{}
	if len(bytes.TrimSpace(b)) == 0 {
		return out, nil
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" {
			continue
		}
		fields := strings.Fields(text)
		// One bad line refuses the whole file. A half-read go.sum could hide
		// a module and look like a clean check.
		if len(fields) != 3 || !strings.HasPrefix(fields[2], "h1:") {
			return nil, fmt.Errorf("line %d: not a module version hash", line)
		}
		// "v1.2.3/go.mod" is the hash of that module's go.mod file. The
		// module itself has its own line, so this one is not a version.
		if strings.HasSuffix(fields[1], "/go.mod") {
			continue
		}
		addVersion(out, fields[0], fields[1], line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
