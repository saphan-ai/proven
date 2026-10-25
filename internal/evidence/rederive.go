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
	unbound, err := findingsWithoutBound(files[FindingsName])
	if err != nil {
		return refuse(CheckFormat, FindingsName, "not readable: %v", err)
	}
	for i, f := range res.Findings {
		if slices.Contains(redactionMoved, f.ID) {
			rep.NotRederived = append(rep.NotRederived, fmt.Sprintf("finding %s: redaction changed the text the rule reads for it, so its outcome is not re-derived", Visible(f.ID)))
			continue
		}
		// A pack written before Bound was recorded holds rows whose tie to the change is unknown.
		// Absent is not false. The rule is run with each absent Bound read both ways, and a
		// component is checked whenever the two readings agree on it: the rest of the recorded
		// evidence decides it.
		lo, hi, err := boundedOutcomes(f, unbound[i])
		if err != nil {
			return refuse(CheckOutcome, FindingsName, "finding %s: %v", Visible(f.ID), err)
		}
		var undecided []string
		if lo.Level != hi.Level {
			undecided = append(undecided, "its proof level")
		} else if lo.Level != f.ProofLevel {
			return refuse(CheckOutcome, FindingsName, "finding %s records proof level %q, the rule gives %q for the evidence recorded", Visible(f.ID), Visible(string(f.ProofLevel)), string(lo.Level))
		}
		if lo.Reported != hi.Reported {
			undecided = append(undecided, "whether it is reported")
		} else if !lo.Reported {
			return refuse(CheckOutcome, FindingsName, "finding %s is recorded as reported, the rule drops it for the evidence recorded", Visible(f.ID))
		}
		switch {
		case !lo.Complete:
			undecided = append(undecided, "whether it was refuted or left open")
		case lo.Unresolved != hi.Unresolved:
			undecided = append(undecided, "whether it is open")
		case lo.Unresolved != f.Unresolved:
			return refuse(CheckOutcome, FindingsName, "finding %s records unresolved=%t, the rule gives unresolved=%t for the evidence recorded", Visible(f.ID), f.Unresolved, lo.Unresolved)
		}
		if len(undecided) > 0 {
			var why []string
			if len(unbound[i]) > 0 {
				why = append(why, "does not record whether its query rows are bound to the change")
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

// boundedOutcomes gives the rule's outcome for f with every query listed in unbound read as not
// bound (lo) and as bound (hi). With none listed, the two are the same.
func boundedOutcomes(f review.Finding, unbound map[int]bool) (lo, hi review.Outcome, err error) {
	if lo, err = review.Rederive(withBound(f, unbound, false)); err != nil {
		return lo, hi, err
	}
	hi, err = review.Rederive(withBound(f, unbound, true))
	return lo, hi, err
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
// whose Bound field is absent from findings.json, as opposed to recorded false.
func findingsWithoutBound(findings []byte) ([]map[int]bool, error) {
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
	out := make([]map[int]bool, len(raw.Findings))
	for i, f := range raw.Findings {
		for j, e := range f.Evidence {
			if e.Query != nil && e.Query.Bound == nil {
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
