package evidence

// Round 2 of verify-rederives: the tests the two round-1 reviews asked for. Each one changes the
// one recorded fact its name says.

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/review"
)

// A version-1 pack written before Bound and Refutation were recorded holds too little to re-derive
// whether a data finding is reported. Verify must not refuse it as a forged verdict.
func TestAPackThatDoesNotRecordBoundIsNotRefusedAsAForgedVerdict(t *testing.T) {
	in, pub := forgeryInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{FindingsName, QueriesName} {
		dec := json.NewDecoder(bytes.NewReader(files[name]))
		dec.UseNumber()
		var v any
		if err := dec.Decode(&v); err != nil {
			t.Fatal(err)
		}
		var strip func(any)
		strip = func(x any) {
			switch y := x.(type) {
			case map[string]any:
				delete(y, "Bound")
				delete(y, "bound")
				delete(y, "Refutation")
				for _, z := range y {
					strip(z)
				}
			case []any:
				for _, z := range y {
					strip(z)
				}
			}
		}
		strip(v)
		b, err := marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		files[name] = b
	}
	rep, err := Verify(sealed(t, files, in.Key, false), pub)
	if err != nil {
		r := refusalOf(err)
		if r != nil && r.Check == CheckFormat {
			return // refused as an older format, by name
		}
		t.Fatalf("a pack that does not record Bound is refused as a forged verdict: %v", err)
	}
	named := false
	for _, n := range rep.NotRederived {
		named = named || strings.Contains(n, "F3")
	}
	if !named {
		t.Fatalf("F3 (data) must be named not re-derived, got %v", rep.NotRederived)
	}
}

// docs/EVIDENCE_PACK.md step 6: the level is the one the evidence that counts gives. F4 records a
// test that did not fail on the change and a bound query, so a pack that says proved-by-test is
// refused and the refusal names proved-by-data.
func TestALevelNeedsEvidenceOfThatLevelThatCounts(t *testing.T) {
	in, pub := forgeryInput(t)
	in.Result.Findings[3].Evidence[0].Test.FailsOnChange = false
	pack, err := build(in)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Verify(pack, pub)
	r := refusalOf(err)
	if r == nil || r.Check != CheckOutcome || !strings.Contains(err.Error(), "F4") || !strings.Contains(err.Error(), string(review.ProvedByData)) {
		t.Fatalf("want an outcome refusal naming F4 and %s, got %v", review.ProvedByData, err)
	}
}

// A genuine review whose Judge kept F3 must verify after the pack's redaction. The refutation row
// holds one configured value and the finding's message another; before redaction the row does not
// name the finding, so the Judge keeps it. Redaction writes the same word into both.
func TestRedactionDoesNotTurnAKeptFindingIntoARefutedOne(t *testing.T) {
	in, pub := forgeryInput(t)
	secretA := strings.Repeat("A", 32)
	secretB := strings.Repeat("B", 32)
	f := &in.Result.Findings[2]
	f.Message = "pkg-a 1.2.0 is in an advisory; the lock file pins " + secretA
	f.Refutation = &review.Refutation{QueryRows: [][]string{{secretB, "no"}}}
	before, err := review.Rederive(*f)
	if err != nil || !before.Reported {
		t.Fatalf("precondition: the Judge keeps F3 before redaction: %+v %v", before, err)
	}
	in.Redactor = redact.New(map[string]string{"TOKEN_A": secretA, "TOKEN_B": secretB})
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(pack, pub); err != nil {
		t.Fatalf("a genuine pack is refused after redaction: %v", err)
	}
}

// queries.json records whether the rows are bound to the change.
func TestQueriesJSONRecordsBound(t *testing.T) {
	in, _ := forgeryInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	var qs []map[string]any
	if err := json.Unmarshal(files[QueriesName], &qs); err != nil {
		t.Fatal(err)
	}
	seen := false
	for _, q := range qs {
		if q["finding_id"] == "F3" {
			seen = true
			if q["bound"] != true {
				t.Fatalf("queries.json does not record F3's rows as bound: %v", q)
			}
		}
	}
	if !seen {
		t.Fatal("queries.json holds no query for F3")
	}
}

// Probe 1 of the codex review: a non-counting test beside a counting query is not proved-by-test.
func TestVerifyRejectsProofFromNonCountingEvidence(t *testing.T) {
	in, pub := sampleInput(t)
	f := &in.Result.Findings[0]
	f.Evidence = []review.Evidence{
		{Kind: review.EvidenceTest, Test: &review.TestEvidence{FailsOnChange: false, PassesOnBase: true}},
		{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{Bound: true, Rows: [][]string{{"pkg"}}}},
	}
	f.Refutation = &review.Refutation{}
	f.ProofLevel = review.ProvedByTest
	in.Result.ProofCounts = review.CountProofs(in.Result.Findings)
	// build, not Build: Build refuses a pack its own verification refuses, and this pack is one
	// that Verify must refuse.
	pack, err := build(in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(pack, pub); err == nil {
		t.Fatal("verify accepted a proved-by-test outcome although only the query counts")
	}
}

// Probe 2 of the codex review: an older pack with no bound field is named, not refused as forged.
func TestOldPackWithoutBoundIsNamedNotRederived(t *testing.T) {
	in, pub := forgeryInput(t)
	in.Result.Findings[2].Refutation = nil
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{FindingsName, QueriesName} {
		before := files[name]
		marker := regexp.MustCompile(`,\s*"bound": true`)
		if name == FindingsName {
			marker = regexp.MustCompile(`,\s*"Bound": true`)
		}
		files[name] = marker.ReplaceAll(before, nil)
		if bytes.Equal(before, files[name]) {
			t.Fatalf("%s had no bound field to remove", name)
		}
	}
	old := sealed(t, files, in.Key, false)
	rep, err := Verify(old, pub)
	if err != nil {
		t.Fatalf("old pack should verify with F3 named not re-derived: %v", err)
	}
	if len(rep.NotRederived) == 0 {
		t.Fatal("old pack did not name its incomplete finding")
	}
}

// A forger who writes a denial word into F3's refutation row, hashes recomputed and signed again,
// is refused by the outcome check: the one-value change moves F3 through the Judge's contradiction rule.
func TestADenialWrittenIntoARefutationRowIsRefused(t *testing.T) {
	in, pub := forgeryInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(files[FindingsName], []byte(`"affected"`)) {
		t.Fatal("findings.json holds no refutation row to change")
	}
	files[FindingsName] = bytes.Replace(files[FindingsName], []byte(`"affected"`), []byte(`"no"`), 1)
	_, err = Verify(sealed(t, files, in.Key, false), pub)
	r := refusalOf(err)
	if r == nil || r.Check != CheckOutcome || !strings.Contains(err.Error(), "F3") {
		t.Fatalf("want an outcome refusal naming F3, got %v", err)
	}
}

// Probe 3 of the codex review: the bench changes a refutation row into a denial word.
func TestForgeryBenchChangesQueryRowToDenial(t *testing.T) {
	for _, change := range variants("affected") {
		if change() == "no" {
			return
		}
	}
	t.Fatal("the bench never changes a refutation row into a denial, so that Judge branch is unmeasured")
}
