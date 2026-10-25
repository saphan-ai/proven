package manifest

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"

	"github.com/saphan-ai/proven/internal/mavenver"
)

// GradleKts reads build.gradle.kts. A double-quoted string that is a literal
// group:artifact:version is a Maven coordinate, and so is one line that
// names group, name and version as literal strings. A version that contains
// "$", or a version written as a name, is a variable: it is a note, not a
// guessed value. A plugin id and a libs. catalog reference are notes.
// Plugin resolution is out of scope. Groovy build.gradle is not this file.
// The package name is group:artifact, as written. The ecosystem is Maven.
func GradleKts() Reader {
	return Reader{
		Ecosystem: "Maven",
		Match: func(path string) bool {
			return path == "build.gradle.kts" || strings.HasSuffix(path, "/build.gradle.kts")
		},
		Diff:    diffGradleKts,
		Notes:   gradleKtsNotes,
		Compare: mavenver.Compare,
		Explain: mavenver.Explain,
	}
}

// diffGradleKts compares the literal coordinates of two Kotlin build files.
func diffGradleKts(base, change []byte) ([]Change, error) {
	before, _, err := parseGradleKts(base)
	if err != nil {
		return nil, err
	}
	after, _, err := parseGradleKts(change)
	if err != nil {
		return nil, err
	}
	return diffNamed(before, after, mavenver.Compare), nil
}

// gradleKtsNotes names coordinates on the change side that were not literals.
func gradleKtsNotes(change []byte) []string {
	_, notes, err := parseGradleKts(change)
	if err != nil {
		return nil
	}
	return notes
}

var (
	ktsPlugin = regexp.MustCompile(`\b(?:id|kotlin)\s*\(\s*"([^"]+)"\s*\)`)
	ktsLibs   = regexp.MustCompile(`\blibs\.([A-Za-z_][A-Za-z0-9_.]*)`)
	ktsGroup  = regexp.MustCompile(`\bgroup\s*=\s*"([^"]+)"`)
	ktsName   = regexp.MustCompile(`\bname\s*=\s*"([^"]+)"`)
	ktsVerQ   = regexp.MustCompile(`\bversion\s*=\s*"([^"]*)"`)
	ktsVerID  = regexp.MustCompile(`\bversion\s*=\s*([A-Za-z_][A-Za-z0-9_.]*)`)
)

// parseGradleKts reads literal coordinates. It is not a Kotlin parser.
// One coordinate is one double-quoted string, or group, name and version
// on one line. Other lines are not dependencies. An empty file has none.
func parseGradleKts(b []byte) (map[string]req, []string, error) {
	out := map[string]req{}
	if len(bytes.TrimSpace(b)) == 0 {
		return out, nil, nil
	}
	var notes []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	line := 0
	inBlock := false
	for sc.Scan() {
		line++
		text := sc.Text()
		var kept string
		kept, inBlock = kotlinLine(text, inBlock)
		if strings.TrimSpace(kept) == "" {
			continue
		}
		ktsPlugins(kept, &notes)
		ktsCatalogRefs(kept, &notes)
		ktsStrings(kept, line, out, &notes)
		ktsNamed(kept, line, out, &notes)
	}
	if err := sc.Err(); err != nil {
		return nil, nil, err
	}
	return out, notes, nil
}

// kotlinLine drops a line comment and a block comment. inBlock is true
// while a "/*" on an earlier line has not closed. Newlines stay in the
// caller, because this function sees one line.
func kotlinLine(text string, inBlock bool) (string, bool) {
	var out strings.Builder
	inStr := false
	for i := 0; i < len(text); i++ {
		c := text[i]
		if inBlock {
			if c == '*' && i+1 < len(text) && text[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if inStr {
			out.WriteByte(c)
			if c == '\\' && i+1 < len(text) {
				i++
				out.WriteByte(text[i])
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}
		if c == '"' {
			inStr = true
			out.WriteByte(c)
			continue
		}
		if c == '/' && i+1 < len(text) && text[i+1] == '/' {
			break
		}
		if c == '/' && i+1 < len(text) && text[i+1] == '*' {
			inBlock = true
			i++
			continue
		}
		out.WriteByte(c)
	}
	return out.String(), inBlock
}

// ktsPlugins notes a plugin id. The id is not turned into a Maven coordinate.
func ktsPlugins(line string, notes *[]string) {
	for _, m := range ktsPlugin.FindAllStringSubmatch(line, -1) {
		id := m[1]
		if id == "" {
			*notes = append(*notes, "a plugin id is empty")
			continue
		}
		*notes = append(*notes, "plugin "+id+" is not resolved to a Maven coordinate")
	}
}

// ktsCatalogRefs notes a libs. reference. The catalog file is a separate
// reader; this file does not guess which coordinate the alias names.
func ktsCatalogRefs(line string, notes *[]string) {
	spans := stringSpans(line)
	for _, m := range ktsLibs.FindAllStringSubmatchIndex(line, -1) {
		if insideSpan(m[0], spans) {
			continue
		}
		*notes = append(*notes, "libs."+line[m[2]:m[3]]+" is a catalog reference, not a literal coordinate")
	}
}

// ktsStrings records every double-quoted coordinate on the line.
// A "$" in the version is a variable.
func ktsStrings(line string, n int, out map[string]req, notes *[]string) {
	for _, s := range quotedStrings(line) {
		group, artifact, ver, ok := splitMavenCoordinate(s)
		if !ok {
			continue
		}
		if strings.Contains(ver, "$") {
			*notes = append(*notes, group+":"+artifact+" version is a variable")
			continue
		}
		out[group+":"+artifact] = req{version: ver, line: n}
	}
}

// ktsNamed records group, name and version written as named arguments on
// one line. An unquoted version is a variable.
func ktsNamed(line string, n int, out map[string]req, notes *[]string) {
	g := ktsGroup.FindStringSubmatch(line)
	name := ktsName.FindStringSubmatch(line)
	if g == nil || name == nil {
		return
	}
	if v := ktsVerQ.FindStringSubmatch(line); v != nil {
		if strings.Contains(v[1], "$") {
			*notes = append(*notes, g[1]+":"+name[1]+" version is a variable")
			return
		}
		if g[1] == "" || name[1] == "" || v[1] == "" {
			*notes = append(*notes, "a named coordinate is empty")
			return
		}
		out[g[1]+":"+name[1]] = req{version: v[1], line: n}
		return
	}
	if ktsVerID.FindStringSubmatch(line) != nil {
		*notes = append(*notes, g[1]+":"+name[1]+" version is a variable")
	}
}

// quotedStrings returns the bodies of double-quoted strings on one line.
// An unclosed string is skipped: the rest of the line is Kotlin, and one
// unclosed quote is not a coordinate.
func quotedStrings(line string) []string {
	var out []string
	for i := 0; i < len(line); i++ {
		if line[i] != '"' {
			continue
		}
		var buf strings.Builder
		i++
		closed := false
		for ; i < len(line); i++ {
			if line[i] == '\\' && i+1 < len(line) {
				buf.WriteByte(line[i+1])
				i++
				continue
			}
			if line[i] == '"' {
				closed = true
				break
			}
			buf.WriteByte(line[i])
		}
		if closed {
			out = append(out, buf.String())
		}
	}
	return out
}

// span is the byte range of one quoted string, including its quotes.
type span struct{ start, end int }

// stringSpans locates quoted strings so a libs. match inside one is not a
// catalog reference. The reference form is an identifier, not a string.
func stringSpans(line string) []span {
	var out []span
	for i := 0; i < len(line); i++ {
		if line[i] != '"' {
			continue
		}
		start := i
		i++
		for ; i < len(line); i++ {
			if line[i] == '\\' && i+1 < len(line) {
				i++
				continue
			}
			if line[i] == '"' {
				out = append(out, span{start, i + 1})
				break
			}
		}
	}
	return out
}

// insideSpan reports whether i falls inside a quoted string.
func insideSpan(i int, spans []span) bool {
	for _, s := range spans {
		if i >= s.start && i < s.end {
			return true
		}
	}
	return false
}
