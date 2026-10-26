package review

import "testing"

// A finding whose message does not name the rows of its refutation, which hold a denial. The
// unrelated reading holds no rows, so the rule keeps the finding open. The related reading places
// every cell in the message, so the rows are about the finding and the rule drops it.
func TestTextReadingsGiveTheTwoEndsOfWhatRedactionHides(t *testing.T) {
	f := Finding{
		ID: "F1", RuleID: "r", Title: "t", Message: "a message",
		Evidence:   []Evidence{testEv(true, true)},
		Refutation: &Refutation{TestCannotRun: true, QueryRows: [][]string{{"zzzz", "no"}}},
	}
	unrelated, related := TextReadings(f)
	u, err := Rederive(unrelated)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Rederive(related)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Reported || !u.Unresolved {
		t.Errorf("the unrelated reading should keep the finding open, got %+v", u)
	}
	if r.Reported {
		t.Errorf("the related reading should drop the finding, got %+v", r)
	}
	if len(f.Refutation.QueryRows) != 1 || f.Message != "a message" {
		t.Errorf("TextReadings changed its input: %+v", f)
	}
}

// With no refutation there is nothing to read two ways.
func TestTextReadingsOfAFindingWithNoRefutationAreTheFinding(t *testing.T) {
	f := Finding{Message: "m", Evidence: []Evidence{testEv(true, true)}}
	a, b := TextReadings(f)
	if a.Message != "m" || b.Message != "m" || a.Refutation != nil || b.Refutation != nil {
		t.Errorf("got %+v and %+v", a, b)
	}
}
