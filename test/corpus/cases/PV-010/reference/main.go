// Reference reading for PV-010, independent of the guard. Used only by check.sh.
//
//	reference note <pinned> <note>     prints current or stale
//	reference widths <binary> <pinned> prints the gap widths 0..256 the guard's matcher accepts
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) == 4 && os.Args[1] == "note" {
		fmt.Println(readNote(os.Args[2], os.Args[3]))
		return
	}
	if len(os.Args) == 4 && os.Args[1] == "widths" {
		fmt.Println(readWidths(os.Args[2], os.Args[3]))
		return
	}
	fmt.Fprintln(os.Stderr, "usage: reference note <pinned> <note> | reference widths <binary> <pinned>")
	os.Exit(2)
}

// readNote: the closing token is the text after the arrow up to the next space,
// without trailing sentence punctuation; the note is current only if that token
// is exactly the pinned count in ASCII digits.
func readNote(pinned, note string) string {
	_, after, ok := strings.Cut(note, "->")
	if !ok {
		return "current"
	}
	tok := strings.TrimLeft(after, " ")
	if i := strings.IndexByte(tok, ' '); i >= 0 {
		tok = tok[:i]
	}
	tok = strings.TrimRight(tok, ".,;:!?")
	if tok == pinned {
		return "current"
	}
	return "stale"
}

// readWidths probes the matcher with a gap of every width from 0 to 256: a note
// whose closing number differs from the pinned count is refused only if the
// matcher found its pair.
func readWidths(binary, pinned string) string {
	var accepted []int
	for w := 0; w <= 256; w++ {
		note := "LANGUAGES" + strings.Repeat(" ", w) + "-> 41"
		out, err := exec.Command(binary, "check", pinned, note).Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "probe failed:", err)
			os.Exit(1)
		}
		if strings.HasPrefix(string(out), "FAIL") {
			accepted = append(accepted, w)
		}
	}
	if len(accepted) == 0 {
		return "none"
	}
	return fmt.Sprintf("%d..%d (%d widths)", accepted[0], accepted[len(accepted)-1], len(accepted))
}
