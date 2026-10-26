// This file keeps the Markdown comment under the platforms' size limit without
// dropping the proof of a finding (ADR-0020 §1, docs/OUTPUT_FORMATS.md).

package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/saphan-ai/proven/internal/review"
)

// The caps keep one comment under the 65,536 characters GitHub and Gitea accept.
// A cut is always announced in the comment, never silent.
const (
	mdMaxComment = 60000 // bytes of a whole comment, under the platforms' limit
	mdMaxMessage = 2000  // bytes of one finding's message shown
	mdMaxTitle   = 300   // bytes of one finding's title shown, in the table and the summary
	mdMaxNote    = 500   // bytes of one note line (a failed service's reason, a statement)
	mdMaxNotes   = 20    // lines of one kind of note (accepted, expired) before "N more"
)

// mdLimits is how much of one finding's proof is shown, and what was left out.
// The first level is the full proof; each next one keeps a smaller share of it, so
// a large result still shows some proof of every finding rather than none of any.
type mdLimits struct {
	rows, out         int // rows of one query and bytes of one test, tool or fix output shown
	cutRows, cutBytes int // what one finding's block left out, filled while it is written
	cutEntries        int // pieces of evidence and fix edits the least proof left out (mdLeast)
}

var mdLevels = []mdLimits{{rows: 20, out: 4000}, {rows: 10, out: 1500}, {rows: 5, out: 500}, {rows: 2, out: 200}, {rows: 1, out: 80}}

// markdownFormatter writes the review as the comment a person reads on the pull request.
type markdownFormatter struct{}

// Write prints, top to bottom: the verdict line, then either the one-sentence
// clean statement or the findings table with one <details> block per finding,
// then the statements a reader must not miss, then the one-line foot.
//
// The comment never starts with the poster's marker line; the poster
// (lab/workflow/proven.yml, action/entrypoint.sh) puts it in front, because
// only it knows the run number and the head commit.
func (markdownFormatter) Write(w io.Writer, result review.Result) error {
	_, err := io.WriteString(w, mdFit(result))
	return err
}

// mdFit renders the comment so that it always fits mdMaxComment and every finding
// it details keeps some proof. WHY levels and not a cut at a byte: a body cut
// there could leave a fence or a <details> open and swallow the rest of the pull
// request page. In order: (1) the full comment; (2) each finding's proof cut to its
// share of the budget; (3) when even the smallest proof of every finding does not
// fit, the first findings detailed and the rest only as table rows, with the count
// left out; (4) when even the table does not fit, its first rows; (5) the verdict alone.
func mdFit(result review.Result) string {
	findings := result.Findings
	blocks := make([][]string, len(findings)) // blocks[i][level]: finding i's <details> at that level
	for i, f := range findings {
		for _, lim := range mdLevels {
			blocks[i] = append(blocks[i], mdBlock(f, lim))
		}
		least, lim := mdLeast(f)
		blocks[i] = append(blocks[i], mdBlock(least, lim))
	}
	full := make([]string, len(findings))
	for i := range blocks {
		full[i] = blocks[i][0]
	}
	if body := mdBody(result, len(findings), full); len(body) <= mdMaxComment {
		return body
	}
	for tabled := len(findings); ; tabled /= 2 {
		// The most findings that can be detailed beside a table of this length. Fewer
		// detailed findings leave a bigger share to each, so a binary search holds.
		best, _ := mdShares(result, tabled, blocks, 0)
		lo, hi := 1, min(tabled, len(findings))
		for lo <= hi {
			mid := (lo + hi) / 2
			if chosen, ok := mdShares(result, tabled, blocks, mid); ok {
				best, lo = chosen, mid+1
			} else {
				hi = mid - 1
			}
		}
		// WHY len(best) > 0: a table so long that no <details> fits beside it would
		// drop all proof, while a shorter table keeps some (review r4 of round 2).
		if body := mdBody(result, tabled, best); len(body) <= mdMaxComment && (len(best) > 0 || tabled == 0) {
			return body
		}
		if tabled == 0 {
			break
		}
	}
	// WHY a last resort: the notes and the foot are bounded, so a real result does not
	// get here; it is the guarantee that a post is never refused for its size.
	return mdVerdict(result) + "\n\nThis comment would pass the size limit of the pull request page even without findings. The full review is in the JSON and SARIF files of this run.\n"
}

// mdLeast is a finding's last level of proof, bounded by construction whatever the
// finding holds: its first piece of evidence only, with every name, column and cell
// shortened, one row, and no fix edits. WHY: the levels above cut the length of
// each piece, not their number, so a finding with many pieces of evidence or many
// edits could exceed any share and lose all its proof (review r5 of round 2).
// What it leaves out is counted in the returned limits and announced by the block.
func mdLeast(f review.Finding) (review.Finding, mdLimits) {
	const name, cols = 80, 4
	lim := mdLevels[len(mdLevels)-1]
	least := review.Finding{ID: f.ID, RuleID: f.RuleID, Severity: f.Severity, Title: f.Title, Message: f.Message}
	if len(f.Evidence) > 0 {
		e := f.Evidence[0]
		switch {
		case e.Query != nil:
			q := *e.Query
			q.Engine, q.Snapshot, q.Form = mdShort(q.Engine, name), mdShort(q.Snapshot, name), mdShort(q.Form, name)
			// New slices: the finding's own columns and rows stay as they are for the
			// levels above and for the JSON and SARIF writers.
			columns := q.Columns
			if len(columns) > cols {
				columns = columns[:cols]
			}
			q.Columns = make([]string, len(columns))
			for i, c := range columns {
				q.Columns[i] = mdShort(c, name)
			}
			if len(q.Rows) > 0 {
				lim.cutRows += len(q.Rows) - 1
				row := q.Rows[0]
				if len(row) > len(q.Columns) {
					row = row[:len(q.Columns)]
				}
				short := make([]string, len(row))
				for i, c := range row {
					short[i] = mdShort(c, name)
				}
				q.Rows = [][]string{short}
			}
			e.Query = &q
		case e.Test != nil:
			t := *e.Test
			t.Name, t.Path = mdShort(t.Name, name), mdShort(t.Path, name)
			e.Test = &t
		case e.Tool != nil:
			t := *e.Tool
			t.Tool, t.Version = mdShort(t.Tool, name), mdShort(t.Version, name)
			e.Tool = &t
		}
		least.Evidence = []review.Evidence{e}
		lim.cutEntries += len(f.Evidence) - 1
	}
	if f.Fix != nil {
		lim.cutEntries++ // the fix itself, its edits and runs with it
	}
	return least, lim
}

// mdShares chooses a level for each of the first m findings so that, beside a table
// of `tabled` rows, the comment fits. A finding with a small proof leaves its unused
// share to the larger ones (smallest first). ok is false when some finding does not
// fit even at the smallest level.
func mdShares(result review.Result, tabled int, blocks [][]string, m int) (chosen []string, ok bool) {
	budget := mdMaxComment - len(mdBody(result, tabled, nil))
	order := make([]int, m)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return len(blocks[order[a]][0]) < len(blocks[order[b]][0]) })
	chosen = make([]string, m)
	for k, idx := range order {
		share := budget / (m - k)
		level := 0
		for level < len(blocks[idx]) && len(blocks[idx][level]) > share {
			level++
		}
		if level == len(blocks[idx]) {
			return nil, false
		}
		chosen[idx] = blocks[idx][level]
		budget -= len(chosen[idx])
	}
	return chosen, true
}

// mdBody renders the whole comment: the table lists the first `tabled` findings and
// `details` holds the <details> blocks of the first findings. What a shorter table
// or fewer blocks leave out is announced, with the count.
func mdBody(result review.Result, tabled int, details []string) string {
	var b strings.Builder
	b.WriteString(mdVerdict(result))
	b.WriteString("\n\n")
	if len(result.Findings) == 0 {
		b.WriteString(mdChecked(result))
		b.WriteString("\n\n")
	} else {
		mdTable(&b, result.Findings[:tabled])
		for _, d := range details {
			b.WriteString(d)
		}
		if n := len(result.Findings) - tabled; n > 0 {
			fmt.Fprintf(&b, "%d more %s not listed here: the size limit of the pull request page. They are in the JSON and SARIF files of this run.\n\n", n, pluralWord(n, "finding is", "findings are"))
		} else if n := tabled - len(details); n > 0 {
			fmt.Fprintf(&b, "%d more %s only in the table above: the message and the evidence do not fit the size limit of the pull request page. They are in the JSON and SARIF files of this run.\n\n", n, pluralWord(n, "finding is", "findings are"))
		}
	}
	mdNotes(&b, result)
	b.WriteString(mdFoot(result))
	b.WriteString("\n")
	return b.String()
}

// cutRaw bounds a raw value at max bytes on a rune boundary, before it is escaped.
// It returns what was cut so the caller can announce it.
func cutRaw(s string, max int) (kept string, cut int) {
	if len(s) <= max {
		return s, 0
	}
	n := max
	for n > 0 && !runeStart(s[n]) {
		n--
	}
	return s[:n], len(s) - n
}

func runeStart(b byte) bool { return b&0xC0 != 0x80 }
