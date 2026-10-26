package autofix

import (
	"fmt"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/testrun"
)

// contextLines is the number of unchanged lines a hunk shows around a change, as `diff -u` does.
const contextLines = 3

// Diff renders the patch as a unified diff against the files of the change in changeDir, the
// form a review comment can publish as suggested changes (ADR-0050 point 4). It reads the files
// by the same checked path as a patch edit and writes nothing. It returns "" when the edits do
// not apply, which Attempt has already refused before it calls Diff.
//
// The diff is derived from the edits, never from the generator's text: what a person is shown is
// what Proven applied.
func Diff(changeDir string, edits []testrun.Edit) string {
	original := map[string]string{}
	current := map[string]string{}
	var order []string
	for _, e := range edits {
		text, ok := current[e.Path]
		if !ok {
			t, reason := testrun.ReadEditable(changeDir, e.Path, 0)
			if reason != "" {
				return ""
			}
			original[e.Path], text = t, t
			order = append(order, e.Path)
		}
		if strings.Count(text, e.Old) != 1 {
			return ""
		}
		current[e.Path] = strings.Replace(text, e.Old, e.New, 1)
	}
	sort.Strings(order)
	var b strings.Builder
	for _, p := range order {
		if original[p] == current[p] {
			continue
		}
		fmt.Fprintf(&b, "--- a/%s\n+++ b/%s\n", p, p)
		b.WriteString(fileDiff(splitLines(original[p]), splitLines(current[p])))
	}
	return b.String()
}

// splitLines splits text into lines, each keeping its newline. A last line with no newline is kept
// as it is, so a change to the end of a file is shown.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.SplitAfter(s, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// op is one line of an edit script: ' ' keeps a line of both files, '-' removes one of the old,
// '+' adds one of the new.
type op struct {
	kind byte
	text string
}

// script computes a line edit script from a to b with the longest common subsequence.
func script(a, b []string) []op {
	n, m := len(a), len(b)
	// lcs[i][j] is the length of the longest common subsequence of a[i:] and b[j:].
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}
	var out []op
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			out = append(out, op{' ', a[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, op{'-', a[i]})
			i++
		default:
			out = append(out, op{'+', b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, op{'-', a[i]})
	}
	for ; j < m; j++ {
		out = append(out, op{'+', b[j]})
	}
	return out
}

// fileDiff renders the hunks of one file: each change with contextLines of unchanged lines on both
// sides, and changes whose context touches merged into one hunk.
func fileDiff(a, b []string) string {
	ops := script(a, b)
	var changed []int
	for i, o := range ops {
		if o.kind != ' ' {
			changed = append(changed, i)
		}
	}
	var out strings.Builder
	for len(changed) > 0 {
		start := max(changed[0]-contextLines, 0)
		end := min(changed[0]+contextLines, len(ops)-1)
		k := 1
		for k < len(changed) && changed[k]-contextLines <= end+1 {
			end = min(changed[k]+contextLines, len(ops)-1)
			k++
		}
		changed = changed[k:]
		// The line numbers of the hunk's first line in the old and the new file.
		oldLine, newLine := 1, 1
		for _, o := range ops[:start] {
			if o.kind != '+' {
				oldLine++
			}
			if o.kind != '-' {
				newLine++
			}
		}
		var body strings.Builder
		oldCount, newCount := 0, 0
		for _, o := range ops[start : end+1] {
			if o.kind != '+' {
				oldCount++
			}
			if o.kind != '-' {
				newCount++
			}
			body.WriteByte(o.kind)
			body.WriteString(o.text)
			if !strings.HasSuffix(o.text, "\n") {
				body.WriteString("\n\\ No newline at end of file\n")
			}
		}
		fmt.Fprintf(&out, "@@ -%s +%s @@\n", rangeOf(oldLine, oldCount), rangeOf(newLine, newCount))
		out.WriteString(body.String())
	}
	return out.String()
}

// rangeOf is the "start,count" of a hunk header; a count of one is written as the start alone,
// and an empty range names the line before it, as `diff -u` does.
func rangeOf(start, count int) string {
	switch count {
	case 1:
		return fmt.Sprint(start)
	case 0:
		return fmt.Sprintf("%d,0", start-1)
	}
	return fmt.Sprintf("%d,%d", start, count)
}
