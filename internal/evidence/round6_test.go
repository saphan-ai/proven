package evidence

import (
	"encoding/json"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// F3 records a bound query and a refuting test that could not run, so the rule keeps it open.
// A forger removes the query's Bound and records F3 closed. Read as not bound, the rule drops F3;
// read as bound, the rule keeps it open. No reading keeps it and closes it, so the pack must be
// refused.
func TestReviewR8MissingBoundCannotHideAnOutcomeNoReadingGives(t *testing.T) {
	in, pub := forgeryInput(t)
	f := &in.Result.Findings[2]
	f.Refutation.TestCannotRun = true
	f.Unresolved = true
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(pack, pub); err != nil {
		t.Fatalf("the genuine pack does not verify: %v", err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	editJSON(t, files, FindingsName, func(v any) {
		for _, x := range v.(map[string]any)["Findings"].([]any) {
			g := x.(map[string]any)
			if g["ID"] != "F3" {
				continue
			}
			for _, e := range g["Evidence"].([]any) {
				if q, ok := e.(map[string]any)["Query"].(map[string]any); ok {
					delete(q, "Bound")
				}
			}
			g["Unresolved"] = false
		}
	})
	editJSON(t, files, QueriesName, func(v any) {
		for _, x := range v.([]any) {
			if q := x.(map[string]any); q["finding_id"] == "F3" {
				delete(q, "bound")
			}
		}
	})
	rep, err := Verify(sealed(t, files, in.Key, false), pub)
	if r := refusalOf(err); r == nil || r.Check != CheckOutcome {
		var named []string
		if rep != nil {
			named = rep.NotRederived
		}
		t.Fatalf("F3 recorded kept and closed verified, although no reading of the missing Bound gives that outcome (not re-derived: %q, err: %v)", named, err)
	}
}

// F4 records a counting test and a refuting test that could not run, so the rule keeps it open.
// Its refutation also holds a row with a denial that the pack's text does not name. A forger marks
// F4 redaction-moved and records it closed. Read as not about F4, the rows leave it open; read as
// about F4, they refute it and the rule drops it. No reading keeps it and closes it, so the pack
// must be refused.
func TestReviewR8RedactionCannotHideAnOutcomeNoReadingGives(t *testing.T) {
	in, pub := forgeryInput(t)
	f := &in.Result.Findings[3]
	f.Refutation = &review.Refutation{TestCannotRun: true, QueryRows: [][]string{{"zzzz", "no"}}}
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(pack, pub); err != nil {
		t.Fatalf("the genuine pack does not verify: %v", err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	editJSON(t, files, FindingsName, func(v any) {
		for _, x := range v.(map[string]any)["Findings"].([]any) {
			if g := x.(map[string]any); g["ID"] == "F4" {
				g["Unresolved"] = false
			}
		}
	})
	var m Manifest
	if err := json.Unmarshal(files[ManifestName], &m); err != nil {
		t.Fatal(err)
	}
	m.RedactionMoved = []string{"F4"}
	if files[ManifestName], err = marshal(m); err != nil {
		t.Fatal(err)
	}
	rep, err := Verify(sealed(t, files, in.Key, false), pub)
	if r := refusalOf(err); r == nil || r.Check != CheckOutcome {
		var named []string
		if rep != nil {
			named = rep.NotRederived
		}
		t.Fatalf("F4 recorded kept and closed verified, although no reading of the redacted text gives that outcome (not re-derived: %q, err: %v)", named, err)
	}
}

