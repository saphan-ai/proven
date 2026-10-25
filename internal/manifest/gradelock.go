package manifest

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/mavenver"
)

// GradleLock reads gradle.lockfile. Each dependency line is
// group:artifact:version and the configurations that locked it, separated
// by '='. The ecosystem is Maven. The package name is group:artifact, as
// written. A line "empty=" names configurations that locked no dependency
// and is not a package. A line this reader cannot read refuses the file.
func GradleLock() Reader {
	return Reader{
		Ecosystem: "Maven",
		Lock:      true,
		NoteErr:   true,
		Match: func(path string) bool {
			return path == "gradle.lockfile" || strings.HasSuffix(path, "/gradle.lockfile")
		},
		Diff:    diffGradleLock,
		Compare: mavenver.Compare,
		Explain: mavenver.Explain,
	}
}

// diffGradleLock compares two Gradle lock files. The error names the side
// that did not parse, because Collect writes it into the trace as it is.
func diffGradleLock(base, change []byte) ([]Change, error) {
	before, err := parseGradleLock(base)
	if err != nil {
		return nil, fmt.Errorf("base: %w", err)
	}
	after, err := parseGradleLock(change)
	if err != nil {
		return nil, fmt.Errorf("change: %w", err)
	}
	return diffVersions(before, after, mavenver.Compare), nil
}

// parseGradleLock reads one pinned coordinate per dependency line.
// A comment starts with '#'. An empty file has no pins. A non-empty file
// with no dependency line and no "empty=" line is refused, so a file of
// another kind is not taken for an empty lock.
func parseGradleLock(b []byte) (versions, error) {
	out := versions{}
	if len(bytes.TrimSpace(b)) == 0 {
		return out, nil
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	line := 0
	seen := false
	for sc.Scan() {
		line++
		trim := strings.TrimSpace(sc.Text())
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		if strings.HasPrefix(trim, "empty=") {
			if !lockConfigs(trim[len("empty="):]) {
				return nil, fmt.Errorf("line %d: bad empty configurations", line)
			}
			seen = true
			continue
		}
		pkg, ver, ok := lockCoordinate(trim)
		if !ok {
			return nil, fmt.Errorf("line %d: not a locked coordinate", line)
		}
		addVersion(out, pkg, ver, line)
		seen = true
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if !seen {
		return nil, fmt.Errorf("no locked dependency")
	}
	return out, nil
}

// lockCoordinate reads group:artifact:version=configurations.
// The configurations are not part of the pin. The version is the third part.
func lockCoordinate(trim string) (string, string, bool) {
	i := strings.IndexByte(trim, '=')
	if i < 0 {
		return "", "", false
	}
	if !lockConfigs(trim[i+1:]) {
		return "", "", false
	}
	group, artifact, ver, ok := splitMavenCoordinate(trim[:i])
	if !ok || strings.Count(trim[:i], ":") != 2 {
		// A classifier is not how the lock file writes a pin.
		return "", "", false
	}
	return group + ":" + artifact, ver, true
}

// lockConfigs reports a comma-separated list of configuration names.
// An empty list is the form of a pin whose configurations were not written.
func lockConfigs(s string) bool {
	if s == "" {
		return true
	}
	for _, part := range strings.Split(s, ",") {
		if part == "" || strings.ContainsAny(part, " \t:#") {
			return false
		}
		for i := 0; i < len(part); i++ {
			c := part[i]
			if c < '0' || (c > '9' && c < 'A') || (c > 'Z' && c < 'a' && c != '_') || c > 'z' {
				return false
			}
		}
	}
	return true
}
