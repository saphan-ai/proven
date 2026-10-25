package manifest

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/saphan-ai/proven/internal/mavenver"
)

// GradleCatalog reads gradle/libs.versions.toml, the Gradle version catalog.
// [libraries] is the dependency list. A version.ref is the string named in
// [versions]; a ref that is not there is a note, not a guessed version.
// [plugins] is not a Maven coordinate: Gradle plugin resolution is out of
// scope, and each plugin is a note. [bundles] names library aliases, not
// versions, and is not read. The package name is group:artifact, as written.
// The ecosystem is Maven.
func GradleCatalog() Reader {
	return Reader{
		Ecosystem: "Maven",
		Match: func(path string) bool {
			return path == "libs.versions.toml" || strings.HasSuffix(path, "/libs.versions.toml")
		},
		Diff:    diffGradleCatalog,
		Notes:   gradleCatalogNotes,
		Compare: mavenver.Compare,
		Explain: mavenver.Explain,
	}
}

// diffGradleCatalog compares the libraries of two version catalogs.
func diffGradleCatalog(base, change []byte) ([]Change, error) {
	before, _, err := parseGradleCatalog(base)
	if err != nil {
		return nil, err
	}
	after, _, err := parseGradleCatalog(change)
	if err != nil {
		return nil, err
	}
	return diffNamed(before, after, mavenver.Compare), nil
}

// gradleCatalogNotes names catalog entries on the change side that were
// not Maven coordinates.
func gradleCatalogNotes(change []byte) []string {
	_, notes, err := parseGradleCatalog(change)
	if err != nil {
		return nil
	}
	return notes
}

// catalogVer is one [versions] entry. ok is false when the value was not
// a string: a ref to it is a note, not a version we invent.
type catalogVer struct {
	value string
	ok    bool
}

// catalogLib is one [libraries] entry before version.ref is resolved.
type catalogLib struct {
	alias     string
	line      int
	coord     string // the "group:artifact:version" string form
	module    string
	group     string
	name      string
	version   string
	ref       string
	literal   bool
	hasRef    bool
	badVer    bool
	notString bool
}

// parseGradleCatalog reads [versions] and [libraries]. It is not a TOML
// parser: one assignment is one line, and a table that opens on a line may
// close on a later line. The later line wins when a name is repeated.
// A line in [versions], [libraries] or [plugins] that is not an assignment
// refuses the file, so a file of another kind is not taken for an empty
// catalog. An empty file has no dependencies.
func parseGradleCatalog(b []byte) (map[string]req, []string, error) {
	out := map[string]req{}
	if len(bytes.TrimSpace(b)) == 0 {
		return out, nil, nil
	}
	versions := map[string]catalogVer{}
	var libs []catalogLib
	var notes []string
	section := ""
	sc := bufio.NewScanner(bytes.NewReader(b))
	line := 0
	var pending string
	pendingLine := 0
	for sc.Scan() {
		line++
		trim := strings.TrimSpace(sc.Text())
		if pending != "" {
			// A comment between the fields of one table is not a field.
			if trim == "" || strings.HasPrefix(trim, "#") {
				continue
			}
			pending = pending + " " + trim
			if braceDepth(pending) > 0 {
				continue
			}
			if err := catalogLine(pending, pendingLine, &section, versions, &libs, &notes); err != nil {
				return nil, nil, err
			}
			pending = ""
			continue
		}
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		if !strings.HasPrefix(trim, "[") && braceDepth(trim) > 0 {
			pending = trim
			pendingLine = line
			continue
		}
		if err := catalogLine(trim, line, &section, versions, &libs, &notes); err != nil {
			return nil, nil, err
		}
	}
	if err := sc.Err(); err != nil {
		return nil, nil, err
	}
	if pending != "" {
		return nil, nil, fmt.Errorf("libs.versions.toml: line %d: unclosed table", pendingLine)
	}
	for _, lib := range libs {
		notes = append(notes, resolveCatalogLib(lib, versions, out)...)
	}
	return out, notes, nil
}

// catalogLine reads one logical line. section is the open table, or empty
// before the first header. Lines outside the four tables are not dependencies.
func catalogLine(trim string, line int, section *string, versions map[string]catalogVer, libs *[]catalogLib, notes *[]string) error {
	trim = stripHashComment(trim)
	if trim == "" {
		return nil
	}
	if strings.HasPrefix(trim, "[") {
		header, ok := cargoHeader(trim)
		if !ok {
			return fmt.Errorf("libs.versions.toml: line %d: bad table header", line)
		}
		switch header {
		case "versions", "libraries", "bundles", "plugins":
			*section = header
			return nil
		default:
			// An unknown table is refused. Skipping it would hide a library
			// written under a name this reader does not know.
			return fmt.Errorf("libs.versions.toml: line %d: unsupported table %q", line, header)
		}
	}
	switch *section {
	case "", "bundles":
		// Bundles name library aliases. They are not coordinates.
		return nil
	case "plugins":
		key, _, err := catalogAssign(trim, line)
		if err != nil {
			return err
		}
		if key == "" {
			*notes = append(*notes, "plugins has an empty name")
			return nil
		}
		// A plugin id is not resolved to a Maven coordinate. Resolution
		// of Gradle plugins is out of scope.
		*notes = append(*notes, "plugins."+key+" is a Gradle plugin, not a Maven coordinate")
		return nil
	case "versions":
		key, raw, err := catalogAssign(trim, line)
		if err != nil {
			return err
		}
		if key == "" {
			*notes = append(*notes, "versions has an empty name")
			return nil
		}
		if raw == "" || (raw[0] != '"' && raw[0] != '\'') {
			versions[key] = catalogVer{}
			*notes = append(*notes, "versions."+key+" is not a string")
			return nil
		}
		val, err := poetryUnquote(raw, line)
		if err != nil {
			return fmt.Errorf("libs.versions.toml: %w", err)
		}
		versions[key] = catalogVer{value: val, ok: true}
		return nil
	default:
		key, raw, err := catalogAssign(trim, line)
		if err != nil {
			return err
		}
		if key == "" {
			*notes = append(*notes, "libraries has an empty name")
			return nil
		}
		lib, err := catalogLibrary(key, raw, line)
		if err != nil {
			return err
		}
		*libs = append(*libs, lib)
		return nil
	}
}

// catalogAssign splits one assignment. The key may be quoted.
func catalogAssign(trim string, line int) (string, string, error) {
	i := strings.IndexByte(trim, '=')
	if i < 0 {
		return "", "", fmt.Errorf("libs.versions.toml: line %d: not an assignment", line)
	}
	key, err := catalogKey(strings.TrimSpace(trim[:i]), line)
	if err != nil {
		return "", "", err
	}
	return key, strings.TrimSpace(trim[i+1:]), nil
}

// catalogKey reads a table key. A quoted key is the name inside the quotes.
func catalogKey(key string, line int) (string, error) {
	if key == "" {
		return "", nil
	}
	if key[0] == '"' || key[0] == '\'' {
		name, err := poetryUnquote(key, line)
		if err != nil {
			return "", fmt.Errorf("libs.versions.toml: %w", err)
		}
		return name, nil
	}
	if strings.ContainsAny(key, " \t[]{}") {
		return "", fmt.Errorf("libs.versions.toml: line %d: bad name", line)
	}
	return key, nil
}

// catalogLibrary reads one library value: a coordinate string, or an inline
// table with module or group and name, and version or version.ref.
func catalogLibrary(alias, raw string, line int) (catalogLib, error) {
	lib := catalogLib{alias: alias, line: line}
	if raw == "" {
		return lib, fmt.Errorf("libs.versions.toml: line %d: empty value", line)
	}
	if raw[0] == '"' || raw[0] == '\'' {
		s, err := poetryUnquote(raw, line)
		if err != nil {
			return lib, fmt.Errorf("libs.versions.toml: %w", err)
		}
		lib.coord = s
		return lib, nil
	}
	if raw[0] != '{' {
		lib.notString = true
		return lib, nil
	}
	fields, err := catalogFields(raw, line)
	if err != nil {
		return lib, err
	}
	if f, ok := fields["module"]; ok && f.isStr {
		lib.module = f.str
	}
	if f, ok := fields["group"]; ok && f.isStr {
		lib.group = f.str
	}
	if f, ok := fields["name"]; ok && f.isStr {
		lib.name = f.str
	}
	if f, ok := fields["version"]; ok {
		if f.isStr {
			lib.version = f.str
			lib.literal = true
		} else {
			lib.badVer = true
		}
	}
	if f, ok := fields["version.ref"]; ok && f.isStr {
		lib.ref = f.str
		lib.hasRef = true
	}
	return lib, nil
}

// catalogField is one key of an inline table.
type catalogField struct {
	str   string
	isStr bool
}

// catalogFields reads an inline table. version = { ref = "name" } is the
// same fact as version.ref = "name". A rich version (strictly, prefer,
// require) is not a string.
func catalogFields(rest string, line int) (map[string]catalogField, error) {
	end, ok := closeBrace(rest)
	if !ok {
		return nil, fmt.Errorf("libs.versions.toml: line %d: unclosed table", line)
	}
	after := strings.TrimSpace(rest[end+1:])
	if after != "" && !strings.HasPrefix(after, "#") {
		return nil, fmt.Errorf("libs.versions.toml: line %d: trailing %q", line, after)
	}
	out := map[string]catalogField{}
	for _, part := range splitTop(rest[1:end]) {
		part = strings.TrimSpace(part)
		if part == "" || !strings.Contains(part, "=") {
			continue
		}
		key, raw, err := catalogAssign(part, line)
		if err != nil {
			return nil, err
		}
		if raw == "" {
			continue
		}
		if raw[0] == '{' && key == "version" {
			inner, err := catalogFields(raw, line)
			if err != nil {
				return nil, err
			}
			if ref, ok := inner["ref"]; ok && ref.isStr && !richVersion(inner) {
				out["version.ref"] = ref
				continue
			}
			out["version"] = catalogField{}
			continue
		}
		if raw[0] != '"' && raw[0] != '\'' {
			out[key] = catalogField{}
			continue
		}
		s, err := poetryUnquote(raw, line)
		if err != nil {
			return nil, fmt.Errorf("libs.versions.toml: %w", err)
		}
		out[key] = catalogField{str: s, isStr: true}
	}
	return out, nil
}

// richVersion reports a Gradle rich version. It is a range, not one version.
func richVersion(inner map[string]catalogField) bool {
	_, strict := inner["strictly"]
	_, prefer := inner["prefer"]
	_, require := inner["require"]
	return strict || prefer || require
}

// resolveCatalogLib turns one library into a coordinate, or into notes when
// the version cannot be known without guessing.
func resolveCatalogLib(lib catalogLib, versions map[string]catalogVer, out map[string]req) []string {
	if lib.notString {
		return []string{"libraries." + lib.alias + " is not a string"}
	}
	if lib.coord != "" {
		group, artifact, ver, ok := splitMavenCoordinate(lib.coord)
		if !ok {
			return []string{"libraries." + lib.alias + " is not a group:artifact:version coordinate"}
		}
		if strings.Contains(ver, "$") {
			return []string{"libraries." + lib.alias + " version is a variable"}
		}
		out[group+":"+artifact] = req{version: ver, line: lib.line}
		return nil
	}
	var notes []string
	pkg := ""
	switch {
	case lib.module != "":
		parts := strings.Split(lib.module, ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			notes = append(notes, "libraries."+lib.alias+" module is not a group:artifact coordinate")
		} else {
			pkg = parts[0] + ":" + parts[1]
		}
	case lib.group != "" && lib.name != "":
		pkg = lib.group + ":" + lib.name
	default:
		notes = append(notes, "libraries."+lib.alias+" has no module")
	}
	ver, ok, why := catalogVersion(lib, versions)
	if !ok {
		notes = append(notes, "libraries."+lib.alias+" "+why)
	}
	if pkg != "" && ok {
		out[pkg] = req{version: ver, line: lib.line}
		return nil
	}
	return notes
}

// catalogVersion is the literal version, or the [versions] string a ref
// names. A variable and a missing ref are not versions.
func catalogVersion(lib catalogLib, versions map[string]catalogVer) (string, bool, string) {
	if lib.literal {
		if lib.version == "" {
			return "", false, "has no version"
		}
		if strings.Contains(lib.version, "$") {
			return "", false, "version is a variable"
		}
		return lib.version, true, ""
	}
	if lib.badVer {
		return "", false, "version is not a string"
	}
	if !lib.hasRef {
		return "", false, "has no version"
	}
	cv, ok := versions[lib.ref]
	if !ok {
		return "", false, "version.ref " + lib.ref + " is not in [versions]"
	}
	if !cv.ok {
		return "", false, "version.ref " + lib.ref + " is not a string"
	}
	if strings.Contains(cv.value, "$") {
		return "", false, "version is a variable"
	}
	return cv.value, true, ""
}

// splitMavenCoordinate splits group:artifact:version. A fourth part is a
// classifier and is not part of the version. Anything else is not a coordinate.
func splitMavenCoordinate(s string) (group, artifact, version string, ok bool) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 && len(parts) != 4 {
		return "", "", "", false
	}
	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", false
	}
	if strings.ContainsAny(parts[0]+parts[1]+parts[2], " \t") {
		return "", "", "", false
	}
	return parts[0], parts[1], parts[2], true
}

// braceDepth is the number of '{' not yet closed, ignoring braces in quotes.
// A line whose depth is above zero continues on the next line.
func braceDepth(s string) int {
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
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
			}
		}
	}
	return depth
}
