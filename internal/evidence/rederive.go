package evidence

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/saphan-ai/proven/internal/review"
)

// checkOutcomes re-derives what the pack states about each finding and refuses a pack whose
// stored verdict does not follow from its stored evidence. It reads only the pack's own files: it
// runs no test, no query and no model, and opens no connection. The rule is review.Rederive, the
// Judge's own, so a change to the rule changes the review and the check together.
//
// What it proves is that the verdict follows from the evidence the pack records. It does not
// prove that the recorded evidence is true.
func checkOutcomes(files map[string][]byte, rep *Report) error {
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
	for _, f := range res.Findings {
		out, err := review.Rederive(f)
		if err != nil {
			return refuse(CheckOutcome, FindingsName, "finding %s: %v", Visible(f.ID), err)
		}
		if out.Level != f.ProofLevel {
			return refuse(CheckOutcome, FindingsName, "finding %s records proof level %q, the rule gives %q for the evidence recorded", Visible(f.ID), Visible(string(f.ProofLevel)), string(out.Level))
		}
		if !out.Reported {
			return refuse(CheckOutcome, FindingsName, "finding %s is recorded as reported, the rule drops it for the evidence recorded", Visible(f.ID))
		}
		if !out.Complete {
			rep.NotRederived = append(rep.NotRederived, fmt.Sprintf("finding %s: the pack records no refutation, so whether it was refuted or left open is not re-derived", Visible(f.ID)))
			continue
		}
		if out.Unresolved != f.Unresolved {
			return refuse(CheckOutcome, FindingsName, "finding %s records unresolved=%t, the rule gives unresolved=%t for the evidence recorded", Visible(f.ID), f.Unresolved, out.Unresolved)
		}
		rep.Rederived++
	}
	if want := review.CountProofs(res.Findings); !reflect.DeepEqual(res.ProofCounts, want) {
		return refuse(CheckOutcome, FindingsName, "the recorded proof counts %v are not the counts %v of the findings recorded", res.ProofCounts, want)
	}
	return nil
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
