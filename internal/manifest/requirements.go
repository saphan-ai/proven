package manifest

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/pep440"
)

// PyPI reads requirements.txt. A pinned line is name==version.
// Extras in brackets are not part of the package name. A blank line or a
// comment is ignored. Every other line is skipped, and Notes names it.
func PyPI() Reader {
	return Reader{
		Ecosystem: "PyPI",
		Match: func(path string) bool {
			return path == "requirements.txt" || strings.HasSuffix(path, "/requirements.txt")
		},
		Diff:  diffRequirements,
		Notes: requirementsNotes,
	}
}

func diffRequirements(base, change []byte) ([]Change, error) {
	before, _, err := parseRequirements(base)
	if err != nil {
		return nil, err
	}
	after, _, err := parseRequirements(change)
	if err != nil {
		return nil, err
	}
	return diffNamed(before, after, pep440.Compare), nil
}

func requirementsNotes(change []byte) []string {
	_, notes, err := parseRequirements(change)
	if err != nil {
		return nil
	}
	return notes
}

func parseRequirements(b []byte) (map[string]req, []string, error) {
	out := map[string]req{}
	var notes []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	line := 0
	for sc.Scan() {
		line++
		raw := strings.TrimSpace(sc.Text())
		name, version, ok, why := parsePin(raw)
		if why != "" {
			notes = append(notes, fmt.Sprintf("line %d skipped (%q): %s", line, raw, why))
			continue
		}
		if !ok {
			continue
		}
		out[name] = req{version: version, line: line}
	}
	if err := sc.Err(); err != nil {
		return nil, nil, fmt.Errorf("requirements.txt: %w", err)
	}
	return out, notes, nil
}

func parsePin(raw string) (string, string, bool, string) {
	if raw == "" || strings.HasPrefix(raw, "#") {
		return "", "", false, ""
	}
	body := strings.TrimSpace(stripInlineComment(raw))
	if body == "" {
		return "", "", false, ""
	}
	spec := cutMarker(body)
	spec, ok := splitOptions(spec)
	if !ok || spec == "" {
		return "", "", false, "not a pinned name==version"
	}
	if strings.HasPrefix(spec, "-") {
		return "", "", false, "not a pinned name==version"
	}
	left, right, ok := splitPin(spec)
	if !ok {
		return "", "", false, "not a pinned name==version"
	}
	name, ok := packageName(left)
	if !ok {
		return "", "", false, "not a pinned name==version"
	}
	version := strings.TrimSpace(right)
	if version == "" || strings.ContainsAny(version, " \t,") {
		return "", "", false, "not a pinned name==version"
	}
	return name, version, true, ""
}

func stripInlineComment(s string) string {
	in := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if in != 0 {
			if c == in {
				in = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			in = c
		case '#':
			return s[:i]
		}
	}
	return s
}

func cutMarker(s string) string {
	in := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if in != 0 {
			if c == in {
				in = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			in = c
		case ';':
			return strings.TrimSpace(s[:i])
		}
	}
	return strings.TrimSpace(s)
}

func splitOptions(spec string) (string, bool) {
	i := strings.Index(spec, " --")
	if i < 0 {
		return strings.TrimSpace(spec), true
	}
	rest := strings.TrimSpace(spec[i:])
	spec = strings.TrimSpace(spec[:i])
	for _, f := range strings.Fields(rest) {
		if !strings.HasPrefix(f, "--") {
			return "", false
		}
	}
	return spec, true
}

func splitPin(spec string) (string, string, bool) {
	if strings.Contains(spec, "===") {
		return "", "", false
	}
	i := strings.Index(spec, "==")
	if i < 0 {
		return "", "", false
	}
	if strings.Contains(spec[i+2:], "==") {
		return "", "", false
	}
	return spec[:i], spec[i+2:], true
}

func packageName(s string) (string, bool) {
	s = strings.TrimSpace(s)
	name := s
	if i := strings.IndexByte(s, '['); i >= 0 {
		if !strings.HasSuffix(s, "]") || strings.Count(s, "[") != 1 {
			return "", false
		}
		if !extrasOK(s[i+1 : len(s)-1]) {
			return "", false
		}
		name = s[:i]
	}
	if !identOK(name) {
		return "", false
	}
	return name, true
}

func extrasOK(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	for _, part := range strings.Split(s, ",") {
		if !identOK(strings.TrimSpace(part)) {
			return false
		}
	}
	return true
}

func identOK(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case (c == '.' || c == '_' || c == '-') && i > 0 && i < len(s)-1:
		default:
			return false
		}
	}
	return true
}
