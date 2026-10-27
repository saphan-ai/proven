package evidence

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/saphan-ai/proven/internal/review"
)

// checkOutcomes re-derives what the pack states about each finding and refuses a pack whose
// stored verdict does not follow from its stored evidence. It reads only the pack's own files: it
// runs no test, no query and no model, and opens no connection. The rule is review.Rederive, the
// Judge's own, so a change to the rule changes the review and the check together.
//
// What it proves is that the verdict follows from the evidence the pack records. It does not
// prove that the recorded evidence is true.
func checkOutcomes(files map[string][]byte, redactionMoved []string, rep *Report) error {
	var res review.Result
	if err := json.Unmarshal(files[FindingsName], &res); err != nil {
		return refuse(CheckFormat, FindingsName, "not readable: %v", err)
	}
	// tests.json and queries.json are the evidence records an auditor reads. They must be what
	// findings.json holds, or one of them was changed alone.
	var gotTests []TestRun
	if err := json.Unmarshal(files[TestsName], &gotTests); err != nil {
		return refuse(CheckFormat, TestsName, "not readable: %v", err)
	}
	if id, same := firstTestDifference(gotTests, tests(res)); !same {
		return refuse(CheckOutcome, TestsName, "the test and tool runs recorded for finding %s are not the ones findings.json holds", Visible(id))
	}
	var gotQueries []QueryRun
	if err := json.Unmarshal(files[QueriesName], &gotQueries); err != nil {
		return refuse(CheckFormat, QueriesName, "not readable: %v", err)
	}
	if id, same := firstQueryDifference(gotQueries, queries(res)); !same {
		return refuse(CheckOutcome, QueriesName, "the queries recorded for finding %s are not the ones findings.json holds", Visible(id))
	}
	unbound, err := findingsWithoutBound(files[FindingsName], files[QueriesName])
	if err != nil {
		return refuse(CheckFormat, FindingsName, "not readable: %v", err)
	}
	for i, f := range res.Findings {
		// A pack written before Bound was recorded holds rows whose tie to the change is unknown.
		// Absent is not false. A finding named by redaction_moved holds text that is not the text
		// the review had; the rule reads that text for one thing, whether a refuting query's rows
		// are about the finding. The rule is run with each unknown read both ways, and a
		// component is checked whenever every reading agrees on it: the rest of the recorded
		// evidence decides it.
		moved := slices.Contains(redactionMoved, f.ID)
		got, err := readings(f, unbound[i], moved)
		if err != nil {
			return refuse(CheckOutcome, FindingsName, "finding %s: %v", Visible(f.ID), err)
		}
		lo := got[0]
		if err := checkRecorded(f, got); err != nil {
			return err
		}
		var undecided []string
		if !agree(got, func(o review.Outcome) any { return o.Level }) {
			undecided = append(undecided, "its proof level")
		}
		if !agree(got, func(o review.Outcome) any { return o.Reported }) {
			undecided = append(undecided, "whether it is reported")
		}
		switch {
		case !lo.Complete:
			undecided = append(undecided, "whether it was refuted or left open")
		case !agree(got, func(o review.Outcome) any { return o.Unresolved }):
			undecided = append(undecided, "whether it is open")
		}
		if len(undecided) > 0 {
			var why []string
			if len(unbound[i]) > 0 {
				why = append(why, "does not record whether its query rows are bound to the change")
			}
			if moved && textMatters(f) {
				why = append(why, "no longer holds the text the rule matched its refuting query rows against")
			}
			if !lo.Complete {
				why = append(why, "records no refutation")
			}
			rep.NotRederived = append(rep.NotRederived, fmt.Sprintf("finding %s: the pack %s, so %s is not re-derived", Visible(f.ID), strings.Join(why, " and "), strings.Join(undecided, ", ")))
			continue
		}
		rep.Rederived++
	}
	if want := review.CountProofs(res.Findings); !reflect.DeepEqual(res.ProofCounts, want) {
		return refuse(CheckOutcome, FindingsName, "the recorded proof counts %v are not the counts %v of the findings recorded", res.ProofCounts, want)
	}
	return nil
}

// checkRecorded refuses a finding when no reading of what the pack leaves open gives the outcome
// the pack records for it as a whole: reported, its proof level and its open state together. Each
// component alone is not enough, because a reading that matches one component can differ on
// another. The pack records every finding as reported, so a reading that drops it cannot match. A
// finding with no recorded refutation has no derived open state, and only its level is matched.
func checkRecorded(f review.Finding, got []review.Outcome) error {
	for _, o := range got {
		if o.Reported && o.Level == f.ProofLevel && (!o.Complete || o.Unresolved == f.Unresolved) {
			return nil
		}
	}
	lo := got[0]
	switch {
	case agree(got, func(o review.Outcome) any { return o.Level }) && lo.Level != f.ProofLevel:
		return refuse(CheckOutcome, FindingsName, "finding %s records proof level %q, the rule gives %q for the evidence recorded", Visible(f.ID), Visible(string(f.ProofLevel)), string(lo.Level))
	case agree(got, func(o review.Outcome) any { return o.Reported }) && !lo.Reported:
		return refuse(CheckOutcome, FindingsName, "finding %s is recorded as reported, the rule drops it for the evidence recorded", Visible(f.ID))
	case lo.Complete && agree(got, func(o review.Outcome) any { return o.Unresolved }) && lo.Unresolved != f.Unresolved:
		return refuse(CheckOutcome, FindingsName, "finding %s records unresolved=%t, the rule gives unresolved=%t for the evidence recorded", Visible(f.ID), f.Unresolved, lo.Unresolved)
	}
	return refuse(CheckOutcome, FindingsName, "finding %s records proof level %q, reported, unresolved=%t, and no reading of what the pack leaves open gives that outcome as a whole", Visible(f.ID), Visible(string(f.ProofLevel)), f.Unresolved)
}

// readings gives the rule's outcome for f under every reading of what the pack leaves open: each
// query listed in unbound read as not bound and as bound, and, when redaction moved the finding's
// text, the refutation's query rows read as not about the finding and as about it. The first
// reading is the least one: nothing bound, no query refutation.
func readings(f review.Finding, unbound map[int]bool, moved bool) ([]review.Outcome, error) {
	texts := []review.Finding{f}
	if moved {
		unrelated, related := review.TextReadings(f)
		texts = []review.Finding{unrelated, related}
	}
	var out []review.Outcome
	for _, t := range texts {
		for _, bound := range []bool{false, true} {
			o, err := review.Rederive(withBound(t, unbound, bound))
			if err != nil {
				return nil, err
			}
			out = append(out, o)
			if len(unbound) == 0 {
				break
			}
		}
	}
	return out, nil
}

// agree reports whether every reading gives the same value of one outcome component.
func agree(got []review.Outcome, component func(review.Outcome) any) bool {
	for _, o := range got[1:] {
		if component(o) != component(got[0]) {
			return false
		}
	}
	return true
}

// textMatters reports whether the two readings of f's text give different outcomes.
func textMatters(f review.Finding) bool {
	a, b := review.TextReadings(f)
	oa, errA := review.Rederive(a)
	ob, errB := review.Rederive(b)
	return errA != nil || errB != nil || oa != ob
}

// withBound copies f with the listed queries' Bound set to v.
func withBound(f review.Finding, unbound map[int]bool, v bool) review.Finding {
	if len(unbound) == 0 {
		return f
	}
	f.Evidence = slices.Clone(f.Evidence)
	for i := range f.Evidence {
		if unbound[i] && f.Evidence[i].Query != nil {
			q := *f.Evidence[i].Query
			q.Bound = v
			f.Evidence[i].Query = &q
		}
	}
	return f
}

// findingsWithoutBound reports, per finding in file order, the evidence indexes of the queries
// whose Bound is recorded nowhere: absent from findings.json and absent from queries.json. A
// query whose queries.json entry carries "bound" has it recorded, whatever findings.json holds
// (the two decode to the same value, or the cross-file check has refused the pack). Only a pack
// written before bound existed lacks the key in both files. The queries are matched by position,
// which is how queries.json lists them: the finding order, then the evidence order.
func findingsWithoutBound(findings, queries []byte) ([]map[int]bool, error) {
	var raw struct {
		Findings []struct {
			Evidence []struct {
				Kind  string
				Query *struct{ Bound *bool }
			}
		}
	}
	if err := json.Unmarshal(findings, &raw); err != nil {
		return nil, err
	}
	var recorded []struct{ Bound *bool }
	if err := json.Unmarshal(queries, &recorded); err != nil {
		return nil, err
	}
	out := make([]map[int]bool, len(raw.Findings))
	n := 0
	for i, f := range raw.Findings {
		for j, e := range f.Evidence {
			if e.Query == nil {
				continue
			}
			inQueries := n < len(recorded) && recorded[n].Bound != nil
			n++
			if e.Query.Bound == nil && !inQueries {
				if out[i] == nil {
					out[i] = map[int]bool{}
				}
				out[i][j] = true
			}
		}
	}
	return out, nil
}

// firstTestDifference returns the finding id of the first run where the two lists differ.
func firstTestDifference(got, want []TestRun) (string, bool) {
	for i := 0; i < len(got) || i < len(want); i++ {
		switch {
		case i >= len(got):
			return want[i].FindingID, false
		case i >= len(want):
			return got[i].FindingID, false
		case !reflect.DeepEqual(got[i], want[i]):
			return want[i].FindingID, false
		}
	}
	return "", true
}

// firstQueryDifference returns the finding id of the first query where the two lists differ.
func firstQueryDifference(got, want []QueryRun) (string, bool) {
	for i := 0; i < len(got) || i < len(want); i++ {
		switch {
		case i >= len(got):
			return want[i].FindingID, false
		case i >= len(want):
			return got[i].FindingID, false
		case !reflect.DeepEqual(got[i], want[i]):
			return want[i].FindingID, false
		}
	}
	return "", true
}
