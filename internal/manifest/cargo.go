package manifest

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/cargover"
)

// Cargo reads Cargo.toml dependencies and dev-dependencies.
// The version string is kept as written. Comparison is the Cargo subset in
// cargover. A dependency with no version, such as a path or git crate, is
// named in Notes and not checked. The package name is stored as written.
// A table other than those two, including build-dependencies and a
// target-specific table, is not read.
func Cargo() Reader {
	return Reader{
		Ecosystem: "crates.io",
		Match: func(path string) bool {
			return path == "Cargo.toml" || strings.HasSuffix(path, "/Cargo.toml")
		},
		Diff:    diffCargo,
		Notes:   cargoNotes,
		Compare: cargover.Compare,
		Explain: cargover.Explain,
	}
}

// diffCargo compares the dependencies of two Cargo.toml files.
func diffCargo(base, change []byte) ([]Change, error) {
	before, _, err := parseCargo(base)
	if err != nil {
		return nil, err
	}
	after, _, err := parseCargo(change)
	if err != nil {
		return nil, err
	}
	return diffNamed(before, after, cargover.Compare), nil
}

// cargoNotes names the entries of the change side that were not crates.io pins.
func cargoNotes(change []byte) []string {
	_, notes, err := parseCargo(change)
	if err != nil {
		return nil
	}
	return notes
}

// cargoTable is one open dotted dependency table, such as [dependencies.tokio].
type cargoTable struct {
	section   string // "dependencies" or "dev-dependencies"
	pkg       string // the crate the header names
	version   string
	rename    string
	line      int
	nonString bool // a version key was present and was not a string
}

// parseCargo reads dependencies and dev-dependencies. It is not a TOML
// parser: one assignment is one line, and an inline table is one line.
// The later line wins when a crate is named twice. A line in those tables
// that is not an assignment refuses the file, so a file of another kind is
// not taken for an empty manifest. An empty file has no dependencies.
func parseCargo(b []byte) (map[string]req, []string, error) {
	out := map[string]req{}
	if len(bytes.TrimSpace(b)) == 0 {
		return out, nil, nil
	}
	var notes []string
	var cur *cargoTable
	flush := func() {
		if cur == nil || cur.pkg == "" {
			cur = nil
			return
		}
		name := cur.pkg
		if cur.rename != "" {
			name = cur.rename
		}
		switch {
		case cur.nonString:
			notes = append(notes, cur.section+"."+cur.pkg+" is not a string")
		case strings.TrimSpace(name) == "":
			notes = append(notes, cur.section+" has an empty package name")
		case cur.version == "":
			notes = append(notes, cur.section+"."+cur.pkg+" has no version")
		default:
			out[name] = req{version: cur.version, line: cur.line}
		}
		cur = nil
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	line := 0
	inFlat := false
	section := ""
	for sc.Scan() {
		line++
		trim := strings.TrimSpace(sc.Text())
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		if strings.HasPrefix(trim, "[") {
			flush()
			inFlat = false
			header, ok := cargoHeader(trim)
			if !ok {
				return nil, nil, fmt.Errorf("Cargo.toml: line %d: bad table header", line)
			}
			kind, pkg, dep := cargoSection(header)
			if !dep {
				section = ""
				continue
			}
			section = kind
			if pkg == "" {
				inFlat = true
				continue
			}
			name, err := cargoKey(pkg, line)
			if err != nil {
				return nil, nil, err
			}
			cur = &cargoTable{section: kind, pkg: name}
			continue
		}
		if cur != nil {
			if err := cargoDottedLine(trim, line, cur); err != nil {
				return nil, nil, err
			}
			continue
		}
		if !inFlat {
			continue
		}
		if err := cargoFlatLine(trim, line, section, out, &notes); err != nil {
			return nil, nil, err
		}
	}
	if err := sc.Err(); err != nil {
		return nil, nil, err
	}
	flush()
	return out, notes, nil
}

// cargoHeader returns the table name inside the brackets, without a comment.
func cargoHeader(trim string) (string, bool) {
	body := stripHashComment(trim)
	if len(body) < 2 || body[0] != '[' || body[len(body)-1] != ']' {
		return "", false
	}
	return strings.TrimSpace(body[1 : len(body)-1]), true
}

// cargoSection reports whether header is dependencies or dev-dependencies,
// and the crate name when the header is a dotted package table.
func cargoSection(header string) (section, pkg string, ok bool) {
	switch {
	case header == "dependencies" || header == "dev-dependencies":
		return header, "", true
	case strings.HasPrefix(header, "dependencies."):
		return "dependencies", header[len("dependencies."):], true
	case strings.HasPrefix(header, "dev-dependencies."):
		return "dev-dependencies", header[len("dev-dependencies."):], true
	default:
		return "", "", false
	}
}

// cargoFlatLine reads one dependency written as a key on its own line.
func cargoFlatLine(trim string, line int, section string, out map[string]req, notes *[]string) error {
	key, rest, err := cargoSplit(trim, line)
	if err != nil {
		return err
	}
	if key == "" {
		*notes = append(*notes, section+" has an empty package name")
		return nil
	}
	if rest == "" {
		return fmt.Errorf("Cargo.toml: line %d: empty value", line)
	}
	switch rest[0] {
	case '"', '\'':
		ver, err := poetryUnquote(rest, line)
		if err != nil {
			return fmt.Errorf("Cargo.toml: %w", err)
		}
		out[key] = req{version: ver, line: line}
		return nil
	case '{':
		ver, rename, hasVer, stringVer, err := cargoInline(rest, line)
		if err != nil {
			return err
		}
		name := key
		if rename != "" {
			name = rename
		}
		if !hasVer {
			*notes = append(*notes, section+"."+key+" has no version")
			return nil
		}
		if !stringVer {
			*notes = append(*notes, section+"."+key+" is not a string")
			return nil
		}
		out[name] = req{version: ver, line: line}
		return nil
	default:
		*notes = append(*notes, section+"."+key+" is not a string")
		return nil
	}
}

// cargoDottedLine reads version and package inside [dependencies.name].
// Any other line is skipped, including a multi-line features array.
func cargoDottedLine(trim string, line int, cur *cargoTable) error {
	if !strings.Contains(trim, "=") {
		return nil
	}
	key, rest, err := cargoSplit(trim, line)
	if err != nil {
		return err
	}
	if key != "version" && key != "package" {
		return nil
	}
	if rest == "" || (rest[0] != '"' && rest[0] != '\'') {
		if key == "version" {
			cur.nonString = true
			cur.version = ""
		}
		return nil
	}
	val, err := poetryUnquote(rest, line)
	if err != nil {
		return fmt.Errorf("Cargo.toml: %w", err)
	}
	if key == "package" {
		cur.rename = val
		return nil
	}
	cur.version = val
	cur.line = line
	cur.nonString = false
	return nil
}

// cargoSplit splits one assignment into a key and the raw value.
// The key is unquoted when Cargo quotes it.
func cargoSplit(trim string, line int) (string, string, error) {
	i := strings.IndexByte(trim, '=')
	if i < 0 {
		return "", "", fmt.Errorf("Cargo.toml: line %d: not a dependency", line)
	}
	key, err := cargoKey(strings.TrimSpace(trim[:i]), line)
	if err != nil {
		return "", "", err
	}
	return key, strings.TrimSpace(trim[i+1:]), nil
}

// cargoKey reads a dependency key. A quoted key is the name inside the quotes.
func cargoKey(key string, line int) (string, error) {
	if key == "" {
		return "", nil
	}
	if key[0] == '"' || key[0] == '\'' {
		name, err := poetryUnquote(key, line)
		if err != nil {
			return "", fmt.Errorf("Cargo.toml: %w", err)
		}
		return name, nil
	}
	if strings.ContainsAny(key, " \t[]{}") {
		return "", fmt.Errorf("Cargo.toml: line %d: bad package name", line)
	}
	return key, nil
}

// cargoInline reads version and package from one inline table.
// hasVer is false when the table names no version. stringVer is false when
// a version key is present and is not a string.
func cargoInline(rest string, line int) (version, rename string, hasVer, stringVer bool, err error) {
	end, ok := closeBrace(rest)
	if !ok {
		return "", "", false, false, fmt.Errorf("Cargo.toml: line %d: unclosed table", line)
	}
	after := strings.TrimSpace(rest[end+1:])
	if after != "" && !strings.HasPrefix(after, "#") {
		return "", "", false, false, fmt.Errorf("Cargo.toml: line %d: trailing %q", line, after)
	}
	for _, part := range splitTop(rest[1:end]) {
		part = strings.TrimSpace(part)
		if part == "" || !strings.Contains(part, "=") {
			continue
		}
		k, v, err := cargoSplit(part, line)
		if err != nil {
			return "", "", false, false, err
		}
		if k != "version" && k != "package" {
			continue
		}
		if v == "" || (v[0] != '"' && v[0] != '\'') {
			if k == "version" {
				hasVer = true
				stringVer = false
			}
			continue
		}
		s, err := poetryUnquote(v, line)
		if err != nil {
			return "", "", false, false, fmt.Errorf("Cargo.toml: %w", err)
		}
		if k == "package" {
			rename = s
			continue
		}
		version = s
		hasVer = true
		stringVer = true
	}
	return version, rename, hasVer, stringVer, nil
}

// closeBrace returns the index of the brace that closes the table at s[0].
func closeBrace(s string) (int, bool) {
	if s == "" || s[0] != '{' {
		return 0, false
	}
	depth := 0
	in := false
	var q byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if in {
			if c == '\\' && i+1 < len(s) {
				i++
				continue
			}
			if c == q {
				in = false
			}
			continue
		}
		switch c {
		case '"', '\'':
			in = true
			q = c
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth == 0 && c == '}' {
				return i, true
			}
		}
	}
	return 0, false
}

// splitTop splits s on commas that are not inside a string, table or array.
func splitTop(s string) []string {
	var parts []string
	depth := 0
	in := false
	var q byte
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if in {
			if c == '\\' && i+1 < len(s) {
				i++
				continue
			}
			if c == q {
				in = false
			}
			continue
		}
		switch c {
		case '"', '\'':
			in = true
			q = c
		case '{', '[':
			depth++
		case '}', ']':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// stripHashComment drops a comment that is not inside a quoted string.
func stripHashComment(s string) string {
	in := false
	var q byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if in {
			if c == '\\' && i+1 < len(s) {
				i++
				continue
			}
			if c == q {
				in = false
			}
			continue
		}
		switch c {
		case '"', '\'':
			in = true
			q = c
		case '#':
			return strings.TrimSpace(s[:i])
		}
	}
	return strings.TrimSpace(s)
}
