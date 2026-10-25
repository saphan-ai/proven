package writer

import "github.com/saphan-ai/proven/internal/candidate"

// The run names a verdict carries. They are the words the pipeline writes on
// candidate.Verdict.Run: the default reviewer first, then the special one.
const (
	runDefault = "default"
	runSpecial = "special"
)

// agreedOpinions lifts claims both reviewers reported, at the same place, that
// the Judge dropped because nothing proved them (ADR-0041). A claim only one
// reviewer reported stays dropped. A claim the Judge kept is a finding and is
// not lifted. The order is the order of the default reviewer's claims.
func agreedOpinions(cands []candidate.Candidate) []candidate.Opinion {
	var out []candidate.Opinion
	used := make([]bool, len(cands))
	for i, c := range cands {
		if !droppedForNoEvidence(c) || !namedPlace(c.Risk) {
			continue
		}
		def, spec, ok := bothReported(c)
		if !ok {
			continue
		}
		used[i] = true
		out = append(out, opinionOf(c.Risk, c.StartLine, c.EndLine, def, spec))
	}
	// Two reviewers can name one place with different rule ids or with line
	// ranges that only overlap. The merge of the two runs does not join those,
	// so they arrive here as two dropped claims. Pair each default-only claim
	// with the first special-only claim that shares its place.
	for i, a := range cands {
		if used[i] || !droppedForNoEvidence(a) || !namedPlace(a.Risk) {
			continue
		}
		def, ok := reported(a, runDefault)
		if !ok || reportedOK(a, runSpecial) {
			continue
		}
		for j, b := range cands {
			if used[j] || i == j || !droppedForNoEvidence(b) || !namedPlace(b.Risk) {
				continue
			}
			spec, ok := reported(b, runSpecial)
			if !ok || reportedOK(b, runDefault) {
				continue
			}
			if !samePlace(a.Risk, b.Risk) {
				continue
			}
			used[i], used[j] = true, true
			start, end := overlap(a.Risk, b.Risk)
			out = append(out, opinionOf(a.Risk, start, end, def, spec))
			break
		}
	}
	return out
}

// samePlace reports whether a and b are the same place (ADR-0041).
// The place is the path and the line range. The ranges overlap when they share
// a line: the later start is at or before the earlier end, and both ends are
// inclusive, so lines 4-8 and 6-10 share 6-8, and lines 1-4 and 5-9 share none.
// A range that names no line (start below 1, or end before start) overlaps
// nothing, and an empty path is not a place. The rule id is not the place:
// each reviewer names the place in its own words.
func samePlace(a, b candidate.Risk) bool {
	if a.Path == "" || a.Path != b.Path || !namedPlace(a) || !namedPlace(b) {
		return false
	}
	return a.StartLine <= b.EndLine && b.StartLine <= a.EndLine
}

// namedPlace reports a path and a line range a reader can point at.
func namedPlace(r candidate.Risk) bool {
	return r.Path != "" && r.StartLine >= 1 && r.EndLine >= r.StartLine
}

// overlap is the lines both ranges share. samePlace has already held.
func overlap(a, b candidate.Risk) (start, end int) {
	start = a.StartLine
	if b.StartLine > start {
		start = b.StartLine
	}
	end = a.EndLine
	if b.EndLine < end {
		end = b.EndLine
	}
	return start, end
}

// droppedForNoEvidence is the Judge's own words for a claim nothing proved.
// A refutation, or a suppression from an earlier review, is a different reason
// and is not an agreement.
func droppedForNoEvidence(c candidate.Candidate) bool {
	return c.Dropped && c.DropReason == "dropped: no evidence"
}

// bothReported reports the two reviewers' own verdicts when each of them
// reported the claim. A refuted verdict, or a reviewer who did not name it,
// is not an agreement.
func bothReported(c candidate.Candidate) (def, spec candidate.Verdict, ok bool) {
	def, ok = reported(c, runDefault)
	if !ok {
		return candidate.Verdict{}, candidate.Verdict{}, false
	}
	spec, ok = reported(c, runSpecial)
	if !ok {
		return candidate.Verdict{}, candidate.Verdict{}, false
	}
	return def, spec, true
}

// reported returns run's verdict when that reviewer reported the claim.
func reported(c candidate.Candidate, run string) (candidate.Verdict, bool) {
	for _, v := range c.Verdicts {
		if v.Run == run && v.Verdict == candidate.VerdictReported {
			return v, true
		}
	}
	return candidate.Verdict{}, false
}

// reportedOK reports whether run reported the claim.
func reportedOK(c candidate.Candidate, run string) bool {
	_, ok := reported(c, run)
	return ok
}

// opinionOf builds one opinion. The title and the id are the default
// reviewer's: the special reviewer's own words are the second reason, copied
// as that reviewer wrote them.
func opinionOf(r candidate.Risk, start, end int, def, spec candidate.Verdict) candidate.Opinion {
	return candidate.Opinion{
		ID: r.ID, Title: r.Title, Path: r.Path, StartLine: start, EndLine: end,
		Reasons: []candidate.Verdict{def, spec},
	}
}
