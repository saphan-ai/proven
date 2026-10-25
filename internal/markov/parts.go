package markov

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/change"
)

// Part is what one review step reads: a whole file (its base and its change), or one hunk of it.
type Part struct {
	Path string
	// File is set for a whole-file part.
	File *change.File
	// Hunk is set for a hunk part: the changed region with its context lines, in unified-diff form.
	Hunk string
}

// Name is the part as a trace line names it.
func (p Part) Name() string {
	if p.File != nil {
		return p.Path
	}
	return p.Path + " " + strings.SplitN(p.Hunk, "\n", 2)[0]
}

// hunkContext is how many unchanged lines stand on each side of a change, as in a unified diff.
const hunkContext = 3

// maxDiffCells bounds the line-diff table. Above it a file is one hunk from its first to its last differing line:
// the review still reads the whole change of that file, in one step, rather than running a diff without bound.
const maxDiffCells = 4_000_000

// Split cuts a change into parts (ADR-0048 point 1). With UnitFile every file is a part, in the change's order, so a
// step sees exactly what the whole review would see of that file. With UnitHunk every changed region of a text file
// is a part; a file with no line changed has no part, and a file that is not valid UTF-8 stays one whole-file part,
// because its lines cannot be compared as text.
func Split(files []change.File, unit string) []Part {
	var out []Part
	for i := range files {
		f := files[i]
		if unit != UnitHunk || !utf8.Valid(f.Base) || !utf8.Valid(f.Change) {
			out = append(out, Part{Path: f.Path, File: &files[i]})
			continue
		}
		for _, h := range Hunks(string(f.Base), string(f.Change)) {
			out = append(out, Part{Path: f.Path, Hunk: h})
		}
	}
	return out
}

// Hunks is the line diff of base and change as unified-diff hunks with hunkContext lines of context. Equal texts
// have no hunk.
func Hunks(base, chg string) []string {
	a, b := lines(base), lines(chg)
	// The common head and tail are trimmed first, so the table covers only the region that differs.
	pre := 0
	for pre < len(a) && pre < len(b) && a[pre] == b[pre] {
		pre++
	}
	suf := 0
	for suf < len(a)-pre && suf < len(b)-pre && a[len(a)-1-suf] == b[len(b)-1-suf] {
		suf++
	}
	if pre == len(a) && pre == len(b) {
		return nil
	}
	ops := diffOps(a[pre:len(a)-suf], b[pre:len(b)-suf])
	// The full script: the common head, the middle, the common tail.
	var script []op
	for i := 0; i < pre; i++ {
		script = append(script, op{kind: ' ', text: a[i]})
	}
	script = append(script, ops...)
	for i := len(a) - suf; i < len(a); i++ {
		script = append(script, op{kind: ' ', text: a[i]})
	}
	return group(script)
}

// op is one line of an edit script: ' ' kept, '-' removed from base, '+' added in the change.
type op struct {
	kind byte
	text string
}

// lines splits text into lines without their newline. An empty text has no line.
func lines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// diffOps is the longest-common-subsequence edit script of a and b. A pair too large for the table is one removal
// of a and one addition of b.
func diffOps(a, b []string) []op {
	if len(a)*len(b) > maxDiffCells {
		var out []op
		for _, l := range a {
			out = append(out, op{'-', l})
		}
		for _, l := range b {
			out = append(out, op{'+', l})
		}
		return out
	}
	// lcs[i][j] is the common-subsequence length of a[i:] and b[j:].
	lcs := make([][]int, len(a)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var out []op
	i, j := 0, 0
	for i < len(a) && j < len(b) {
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
	for ; i < len(a); i++ {
		out = append(out, op{'-', a[i]})
	}
	for ; j < len(b); j++ {
		out = append(out, op{'+', b[j]})
	}
	return out
}

// group turns an edit script into hunks: each run of changes with hunkContext kept lines on each side, and two
// runs whose context would touch joined into one hunk.
func group(script []op) []string {
	var changed []int
	for i, o := range script {
		if o.kind != ' ' {
			changed = append(changed, i)
		}
	}
	var hunks []string
	for k := 0; k < len(changed); {
		start := max(changed[k]-hunkContext, 0)
		end := changed[k]
		// Extend over every change whose context overlaps this hunk's.
		for k < len(changed) && changed[k] <= end+2*hunkContext {
			end = changed[k]
			k++
		}
		end = min(end+hunkContext, len(script)-1)
		hunks = append(hunks, render(script, start, end))
	}
	return hunks
}

// render writes script[start..end] as one hunk with its unified-diff header; line numbers are 1-based.
func render(script []op, start, end int) string {
	// Count the base and change lines before the hunk, for the header.
	aLine, bLine := 1, 1
	for _, o := range script[:start] {
		if o.kind != '+' {
			aLine++
		}
		if o.kind != '-' {
			bLine++
		}
	}
	aN, bN := 0, 0
	var body strings.Builder
	for _, o := range script[start : end+1] {
		if o.kind != '+' {
			aN++
		}
		if o.kind != '-' {
			bN++
		}
		body.WriteByte(o.kind)
		body.WriteString(o.text)
		body.WriteByte('\n')
	}
	return fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", aLine, aN, bLine, bN) + body.String()
}
