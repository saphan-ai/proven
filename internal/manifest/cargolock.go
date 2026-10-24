package manifest

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/cargover"
)

// CargoLock reads Cargo.lock. Each [[package]] table contributes its name
// and its version. The ecosystem is crates.io. The name is stored as written.
// A table with no name or no version refuses the file. The package being
// built is a [[package]] table too.
func CargoLock() Reader {
	return Reader{
		Ecosystem: "crates.io",
		Lock:      true,
		NoteErr:   true,
		Match: func(path string) bool {
			return path == "Cargo.lock" || strings.HasSuffix(path, "/Cargo.lock")
		},
		Diff:    diffCargoLock,
		Compare: cargover.Compare,
		Explain: cargover.Explain,
	}
}

// diffCargoLock compares two Cargo.lock files. The error names the side
// that did not parse, because Collect writes it into the trace as it is.
func diffCargoLock(base, change []byte) ([]Change, error) {
	before, err := parseCargoLock(base)
	if err != nil {
		return nil, fmt.Errorf("base: %w", err)
	}
	after, err := parseCargoLock(change)
	if err != nil {
		return nil, fmt.Errorf("change: %w", err)
	}
	return diffVersions(before, after, cargover.Compare), nil
}

// parseCargoLock reads the name and version of each [[package]] table.
// It is not a TOML parser. Lines in any other table, including a multi-line
// array, are skipped. A line outside every table that is not an assignment
// is refused, and a file with no [[package]] table is refused, so a file of
// another kind is not taken for an empty lock.
func parseCargoLock(b []byte) (versions, error) {
	out := versions{}
	if len(bytes.TrimSpace(b)) == 0 {
		return out, nil
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	line := 0
	seen := false
	inHeader := false
	inOther := false
	start := 0
	name, version := "", ""
	verLine := 0
	finish := func() error {
		if !inHeader {
			return nil
		}
		inHeader = false
		if name == "" {
			return fmt.Errorf("[[package]] at line %d has no name", start)
		}
		if version == "" {
			return fmt.Errorf("[[package]] at line %d has no version", start)
		}
		addVersion(out, name, version, verLine)
		name, version = "", ""
		return nil
	}
	for sc.Scan() {
		line++
		trim := strings.TrimSpace(sc.Text())
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		if strings.HasPrefix(trim, "[") {
			if err := finish(); err != nil {
				return nil, err
			}
			header := stripHashComment(trim)
			if header == "[[package]]" {
				seen = true
				inHeader = true
				inOther = false
				start = line
			} else {
				inHeader = false
				inOther = true
			}
			continue
		}
		if inOther {
			continue
		}
		key, val, ok, err := poetryAssign(trim, line)
		if err != nil {
			return nil, err
		}
		if !inHeader {
			if !ok && !strings.Contains(trim, "=") {
				return nil, fmt.Errorf("line %d: not a package table", line)
			}
			continue
		}
		if !ok {
			continue
		}
		switch key {
		case "name":
			name = val
		case "version":
			version = val
			verLine = line
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if err := finish(); err != nil {
		return nil, err
	}
	if !seen {
		return nil, fmt.Errorf("no [[package]] table")
	}
	return out, nil
}
