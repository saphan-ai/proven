package manifest

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/pep440"
)

// Poetry reads poetry.lock. Each [[package]] table contributes its name
// and its version. The ecosystem is PyPI. The name and the version are
// the assignments written before the next table header.
func Poetry() Reader {
	return Reader{
		Ecosystem: "PyPI",
		Lock:      true,
		NoteErr:   true,
		Match: func(path string) bool {
			return path == "poetry.lock" || strings.HasSuffix(path, "/poetry.lock")
		},
		Diff: diffPoetry,
	}
}

// diffPoetry compares two poetry.lock files with PEP 440 order. The error
// names the side that did not parse.
func diffPoetry(base, change []byte) ([]Change, error) {
	before, err := parsePoetry(base)
	if err != nil {
		return nil, fmt.Errorf("base: %w", err)
	}
	after, err := parsePoetry(change)
	if err != nil {
		return nil, fmt.Errorf("change: %w", err)
	}
	return diffVersions(before, after, pep440.Compare), nil
}

// parsePoetry reads poetry.lock line by line. It is not a TOML parser: it
// reads only what a [[package]] table needs, and refuses a line it does
// not understand outside one, so a file of another kind is not taken for
// an empty lock.
func parsePoetry(b []byte) (versions, error) {
	out := versions{}
	if len(bytes.TrimSpace(b)) == 0 {
		return out, nil
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	line := 0
	seen := false
	inHeader := false
	start := 0
	name, version := "", ""
	verLine := 0
	// finish closes the open [[package]] table. A table with no name or no
	// version refuses the file rather than dropping one package unseen.
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
		// Any table header ends the open package, including its own
		// sub-tables such as [package.dependencies]. A key written there
		// names a dependency, not this package's name or version.
		if strings.HasPrefix(trim, "[") {
			if err := finish(); err != nil {
				return nil, err
			}
			if trim == "[[package]]" {
				seen = true
				inHeader = true
				start = line
			}
			continue
		}
		key, val, ok, err := poetryAssign(trim, line)
		if err != nil {
			return nil, err
		}
		// Outside a [[package]] table only key = value lines are expected.
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

// poetryAssign reads a line of the form key = "string". The bool is false
// for any other line, which is not an error: a package table also holds
// arrays and inline tables this reader does not need.
func poetryAssign(trim string, line int) (string, string, bool, error) {
	i := strings.IndexByte(trim, '=')
	if i < 0 {
		return "", "", false, nil
	}
	key := strings.TrimSpace(trim[:i])
	rest := strings.TrimSpace(trim[i+1:])
	if key == "" || strings.ContainsAny(key, " \t[]{}") {
		return "", "", false, nil
	}
	if rest == "" || (rest[0] != '"' && rest[0] != '\'') {
		return "", "", false, nil
	}
	val, err := poetryUnquote(rest, line)
	if err != nil {
		return "", "", false, err
	}
	return key, val, true, nil
}

// poetryUnquote reads one quoted TOML string and allows only a comment
// after it. It knows the escapes a package name or version can need.
func poetryUnquote(s string, line int) (string, error) {
	if len(s) < 2 || (s[0] != '"' && s[0] != '\'') {
		return "", fmt.Errorf("line %d: unclosed string", line)
	}
	q := s[0]
	buf := make([]byte, 0, len(s)-2)
	for i := 1; i < len(s); i++ {
		c := s[i]
		if c == '\\' {
			if i+1 >= len(s) {
				return "", fmt.Errorf("line %d: broken escape", line)
			}
			i++
			switch s[i] {
			case '\\', '\'', '"':
				buf = append(buf, s[i])
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			default:
				return "", fmt.Errorf("line %d: broken escape", line)
			}
			continue
		}
		if c == q {
			rest := strings.TrimSpace(s[i+1:])
			if rest != "" && !strings.HasPrefix(rest, "#") {
				return "", fmt.Errorf("line %d: trailing %q", line, rest)
			}
			return string(buf), nil
		}
		buf = append(buf, c)
	}
	return "", fmt.Errorf("line %d: unclosed string", line)
}
