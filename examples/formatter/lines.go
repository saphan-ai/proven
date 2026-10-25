// Package formatter is a worked example of an output format.
//
// Lines writes one finding per line. It implements cli.Formatter.
// It is not registered on the command. text, json and sarif stay the
// names FormatterFor accepts. How to register a format is in
// docs/ADAPTERS.md, under "Adding an output format".
package formatter

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/cli"
	"github.com/saphan-ai/proven/internal/review"
)

// The assignment fails the build when cli.Formatter changes.
// The example lives outside package cli, so this line is the failure
// a reader sees. A test that only compares bytes would still pass.
var _ cli.Formatter = Lines{}

// Lines writes one finding per line, as path:line: title.
// A review with no findings writes "findings: none".
// Opinions are not findings and are not written.
type Lines struct{}

// Write writes result to w as one line per finding.
// The path and the title are escaped before they are written.
// The line number is Location.StartLine.
// Write returns the error from w and does not wrap it.
// The body is built first and handed over in one write, so a failure
// of that write is the destination's own error. The shell's render
// passes a buffer and drops it when Write returns an error.
func (Lines) Write(w io.Writer, result review.Result) error {
	var body strings.Builder
	if len(result.Findings) == 0 {
		body.WriteString("findings: none\n")
	} else {
		for _, finding := range result.Findings {
			// WHY the line is one Fprintf, after escaping: a newline in
			// the path or the title must not become a second line. The
			// shape matches a compiler's "file:line: message". A path
			// may itself contain a colon, so a program must not split
			// the line on the first colon. The golden file is the contract.
			fmt.Fprintf(&body, "%s:%d: %s\n",
				Escaped(finding.Location.Path),
				finding.Location.StartLine,
				Escaped(finding.Title),
			)
		}
	}
	_, err := io.WriteString(w, body.String())
	return err
}

// Escaped returns s with every byte that could act on a terminal written
// as \xNN. C0, DEL and C1 are escaped. A byte that is not valid UTF-8 is
// escaped as that byte, not replaced with U+FFFD.
//
// This is the same byte contract as visible in internal/cli. visible is
// not exported, so a formatter outside that package keeps its own copy.
// A terminal writer that skips this step lets a string from the change
// move the cursor.
func Escaped(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			fmt.Fprintf(&b, `\x%02x`, s[i])
		case isTerminalControl(r):
			fmt.Fprintf(&b, `\x%02x`, r)
		default:
			b.WriteString(s[i : i+size])
		}
		i += size
	}
	return b.String()
}

// isTerminalControl reports a rune visible would hide: C0, DEL or C1.
func isTerminalControl(r rune) bool {
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f)
}
