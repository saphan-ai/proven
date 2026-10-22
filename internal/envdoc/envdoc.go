// Package envdoc turns setting declarations into the environment tables
// of docs/CONFIGURATION.md.
package envdoc

import (
	"fmt"
	"strings"
	"time"
)

const (
	// MarkerBegin opens the generated environment tables.
	MarkerBegin = "<!-- generated: settings -->"
	// MarkerEnd closes the generated environment tables.
	MarkerEnd = "<!-- end generated -->"
)

const (
	startHeading = "## The model generator: Meterless Nightshift"
	endHeading   = "## The Prover's container"
)

// Setting is one environment variable, declared beside the constant that
// names it. Default and Description are markdown for one table cell each.
type Setting struct {
	Name        string
	Default     string
	Secret      bool
	Description string
}

// Section is one heading, its table, and optional prose after the table.
type Section struct {
	Heading string
	After   string
	Rows    []Setting
}

// Tick wraps a literal so the table shows it as code.
func Tick(s string) string {
	return "`" + s + "`"
}

// SecondSpec is the environment spelling of a whole number of seconds.
// time.Duration.String writes 60s as 1m0s, which is not the form the
// settings accept in their examples.
func SecondSpec(d time.Duration) string {
	if d < 0 || d%time.Second != 0 {
		return d.String()
	}
	return fmt.Sprintf("%ds", int64(d/time.Second))
}

// Check refuses a table the page cannot render: an empty or repeated name,
// or a cell that contains a pipe or a newline.
func Check(sections []Section) error {
	seen := map[string]bool{}
	for _, sec := range sections {
		if sec.Heading == "" {
			return fmt.Errorf("a section has no heading")
		}
		if strings.Contains(sec.Heading, "\n") || strings.Contains(sec.After, "\n") {
			return fmt.Errorf("section %q has a newline in its heading or its trailing prose", sec.Heading)
		}
		for _, row := range sec.Rows {
			if row.Name == "" || seen[row.Name] {
				return fmt.Errorf("duplicate or empty setting %q", row.Name)
			}
			seen[row.Name] = true
			for _, part := range []string{row.Name, row.Default, row.Description} {
				if strings.ContainsAny(part, "|\n") {
					return fmt.Errorf("setting %s has a pipe or a newline in a cell", row.Name)
				}
			}
		}
	}
	return nil
}

// Render writes the environment sections. The result ends with a newline.
func Render(sections []Section) string {
	var b strings.Builder
	for i, sec := range sections {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("## ")
		b.WriteString(sec.Heading)
		b.WriteString("\n\n")
		b.WriteString("| Environment variable | Default | What it changes |\n")
		b.WriteString("|---|---|---|\n")
		for _, row := range sec.Rows {
			b.WriteString("| `")
			b.WriteString(row.Name)
			b.WriteString("` | ")
			b.WriteString(row.Default)
			b.WriteString(" | ")
			b.WriteString(row.Description)
			b.WriteString(" |\n")
		}
		if sec.After != "" {
			b.WriteByte('\n')
			b.WriteString(sec.After)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// Extract returns the generated block, including its trailing newline.
// A missing marker is an error that names the marker.
func Extract(file string) (string, error) {
	i := strings.Index(file, MarkerBegin)
	if i < 0 {
		return "", fmt.Errorf("missing %s", MarkerBegin)
	}
	rest := file[i+len(MarkerBegin):]
	if !strings.HasPrefix(rest, "\n") {
		return "", fmt.Errorf("%s is not on its own line", MarkerBegin)
	}
	rest = rest[1:]
	j := strings.Index(rest, MarkerEnd)
	if j < 0 {
		return "", fmt.Errorf("missing %s", MarkerEnd)
	}
	return rest[:j], nil
}

// Apply writes rendered between the markers. When the markers are absent it
// replaces the hand-written environment sections, from the model-generator
// heading up to the Prover heading, which stays outside the markers.
func Apply(file, rendered string) (string, error) {
	if !strings.HasSuffix(rendered, "\n") {
		rendered += "\n"
	}
	block := MarkerBegin + "\n" + rendered + MarkerEnd + "\n"
	if i := strings.Index(file, MarkerBegin); i >= 0 {
		j := strings.Index(file, MarkerEnd)
		if j < 0 || j < i {
			return "", fmt.Errorf("has %s without %s", MarkerBegin, MarkerEnd)
		}
		after := file[j+len(MarkerEnd):]
		if strings.HasPrefix(after, "\n") {
			after = after[1:]
		}
		return file[:i] + block + after, nil
	}
	start := strings.Index(file, startHeading)
	end := strings.Index(file, endHeading)
	if start < 0 || end < 0 || end < start {
		return "", fmt.Errorf("neither generated markers nor the environment sections")
	}
	return file[:start] + block + "\n" + file[end:], nil
}

// DriftSetting names the environment variable on the first line where the
// page and the render disagree. An empty name means the difference is not
// a setting row.
func DriftSetting(page, rendered string) string {
	pl := strings.Split(page, "\n")
	rl := strings.Split(rendered, "\n")
	n := len(pl)
	if len(rl) < n {
		n = len(rl)
	}
	for i := 0; i < n; i++ {
		if pl[i] == rl[i] {
			continue
		}
		if name := rowName(rl[i]); name != "" {
			return name
		}
		if name := rowName(pl[i]); name != "" {
			return name
		}
		return ""
	}
	extra := rl[n:]
	if len(pl) > len(rl) {
		extra = pl[n:]
	}
	for _, line := range extra {
		if name := rowName(line); name != "" {
			return name
		}
	}
	return ""
}

func rowName(line string) string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "| `") {
		return ""
	}
	rest := line[len("| `"):]
	i := strings.Index(rest, "`")
	if i <= 0 {
		return ""
	}
	return rest[:i]
}
