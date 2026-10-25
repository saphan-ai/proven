package evidence

// The forgery bench (ADR-0035). It takes a sample pack, changes one value at a time, and counts
// what Verify refuses.
//
//	(a) the archive is left as it is: every change must be refused by the hash or the signature check;
//	(b) the manifest's hashes are recomputed and the manifest is signed again with the test key, as a
//	    careful forger would: every change that moves a finding's outcome must be refused by the
//	    re-derivation.
//
// Run it with `go test -count=1 -v -run TestForgeryBench ./internal/evidence` to see both counts.

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// forgeryInput is a pack with every level of proof, a finding left open and a finding whose refutation
// carries rows, so each recorded field has a job in the rule.
func forgeryInput(t *testing.T) (Input, ed25519.PublicKey) {
	t.Helper()
	in, pub := sampleInput(t)
	test := func(name string) review.Evidence {
		return review.Evidence{Kind: review.EvidenceTest, Test: &review.TestEvidence{Path: "a_test.go", Name: name, Command: "go test ./...", FailsOnChange: true, PassesOnBase: true, Output: "FAIL: " + name}}
	}
	in.Result.Findings = []review.Finding{
		{
			ID: "F1", RuleID: "discount.negative", Title: "negative total", Message: "total < 0 for a 150% discount",
			Severity: review.SeverityError, Location: review.Location{Path: "discount.go", StartLine: 4, EndLine: 9},
			ProofLevel: review.ProvedByTest, Evidence: []review.Evidence{test("TestNeg")},
			Refutation: &review.Refutation{},
		},
		{
			ID: "F2", RuleID: "tool.vet", Title: "lost error", Message: "an error is dropped",
			Severity: review.SeverityWarning, Location: review.Location{Path: "io.go", StartLine: 12, EndLine: 12},
			ProofLevel: review.ProvedByTool,
			Evidence: []review.Evidence{{Kind: review.EvidenceTool, Tool: &review.ToolEvidence{
				Tool: "vet", Version: "1", Command: "go vet ./...", Output: "io.go:12: unchecked", BaseOutput: "", PresentOnChange: true, PresentOnBase: false}}},
			Refutation: &review.Refutation{},
		},
		{
			ID: "F3", RuleID: "dep.advisory", Title: "vulnerable pkg-a", Message: "pkg-a 1.2.0 is in an advisory",
			Severity: review.SeverityError, Location: review.Location{Path: "go.mod", StartLine: 3, EndLine: 3},
			ProofLevel: review.ProvedByData,
			Evidence: []review.Evidence{{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{
				SQL: "select * from osv", Engine: "local-sample", Snapshot: "s1", Columns: []string{"package", "fixed"}, Rows: [][]string{{"pkg-a", "1.2.1"}}, Bound: true}}},
			Refutation: &review.Refutation{QueryRows: [][]string{{"pkg-a", "affected"}}},
		},
		{
			ID: "F4", RuleID: "discount.cap", Title: "no cap", Message: "discount has no cap",
			Severity: review.SeverityWarning, Location: review.Location{Path: "discount.go", StartLine: 20, EndLine: 22},
			ProofLevel: review.ProvedByTest, Unresolved: true,
			Evidence: []review.Evidence{test("TestCap"),
				{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{SQL: "select 1", Engine: "local-sample", Snapshot: "s1", Columns: []string{"n"}, Rows: [][]string{{"1"}}, Bound: true}}},
			Refutation: &review.Refutation{TestCannotRun: true},
		},
	}
	in.Result.ProofCounts = review.CountProofs(in.Result.Findings)
	return in, pub
}

// outcomeKey is what the bench calls a finding's outcome.
type outcomeKey struct {
	Level      review.ProofLevel
	Reported   bool
	Unresolved bool
}

// recordedOutcomes reads the outcome the pack's files give, by a path that does not go through
// Verify: from findings.json, and again from the evidence tests.json and queries.json record. It
// returns false when a file cannot be read.
func recordedOutcomes(files map[string][]byte) (map[string][2]outcomeKey, map[review.ProofLevel]int, bool) {
	var res review.Result
	var tr []TestRun
	var qr []QueryRun
	if json.Unmarshal(files[FindingsName], &res) != nil || json.Unmarshal(files[TestsName], &tr) != nil || json.Unmarshal(files[QueriesName], &qr) != nil {
		return nil, nil, false
	}
	out := map[string][2]outcomeKey{}
	for _, f := range res.Findings {
		own, err := review.Rederive(f)
		if err != nil {
			own = review.Outcome{}
		}
		// The same finding with its evidence read from the two evidence files.
		g := f
		g.Evidence = nil
		for _, r := range tr {
			if r.FindingID != f.ID {
				continue
			}
			if r.Kind == review.EvidenceTool {
				g.Evidence = append(g.Evidence, review.Evidence{Kind: r.Kind, Tool: &review.ToolEvidence{PresentOnChange: r.FailsOnChange, PresentOnBase: !r.PassesOnBase}})
			} else {
				g.Evidence = append(g.Evidence, review.Evidence{Kind: r.Kind, Test: &review.TestEvidence{FailsOnChange: r.FailsOnChange, PassesOnBase: r.PassesOnBase}})
			}
		}
		for _, q := range qr {
			if q.FindingID == f.ID {
				g.Evidence = append(g.Evidence, review.Evidence{Kind: review.EvidenceQuery, Query: &review.QueryEvidence{Rows: q.Rows, Bound: q.Bound}})
			}
		}
		again, err := review.Rederive(g)
		if err != nil {
			again = review.Outcome{}
		}
		out[f.ID] = [2]outcomeKey{
			{f.ProofLevel, own.Reported, f.Unresolved},
			{again.Level, again.Reported, again.Unresolved},
		}
	}
	return out, res.ProofCounts, true
}

// movesOutcome reports whether the pack's files state a different outcome for any finding than the
// original did, or cannot be read at all.
func movesOutcome(orig, files map[string][]byte) bool {
	want, wantCounts, _ := recordedOutcomes(orig)
	got, gotCounts, ok := recordedOutcomes(files)
	if !ok {
		return true
	}
	return !reflect.DeepEqual(want, got) || !reflect.DeepEqual(wantCounts, gotCounts)
}

// A change is one value of one file, replaced.
type change struct {
	file   string
	line   int // the trace line, or 0
	leaf   int
	name   string
	mutate func(any) any
}

// variants lists the values one leaf is replaced by, one change each.
func variants(v any) []func() any {
	switch x := v.(type) {
	case bool:
		return []func() any{func() any { return !x }}
	case json.Number:
		return []func() any{func() any {
			n, _ := x.Int64()
			return json.Number(strconv.FormatInt(n+1, 10))
		}}
	case string:
		out := []func() any{func() any { return x + "x" }}
		// A text is also replaced by a word the Judge reads as a denial. In a refutation row it
		// moves an outcome through judge.QueryContradicts; elsewhere it is one more value.
		if x != "no" {
			out = append(out, func() any { return "no" })
		}
		// A proof level is also replaced by each other level, which is what a forger who wants a
		// stronger or weaker level writes.
		if review.ProofLevel(x).Valid() {
			for _, l := range review.ProofLevels {
				if l := l; string(l) != x {
					out = append(out, func() any { return string(l) })
				}
			}
		}
		return out
	case nil:
		return []func() any{func() any { return "x" }}
	}
	return nil
}

// leaves reads the JSON value in b and returns each leaf, in document order.
func leaves(t *testing.T, b []byte) []any {
	t.Helper()
	var out []any
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var walk func()
	walk = func() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatal(err)
		}
		if d, ok := tok.(json.Delim); ok {
			closeWith := json.Delim(']')
			obj := d == '{'
			if obj {
				closeWith = '}'
			}
			for dec.More() {
				if obj {
					if _, err := dec.Token(); err != nil { // the key
						t.Fatal(err)
					}
				}
				walk()
			}
			if _, err := dec.Token(); err != nil {
				t.Fatal(err)
			}
			_ = closeWith
			return
		}
		out = append(out, tok)
	}
	walk()
	return out
}

// replaceLeaf writes b again with leaf n replaced by v, in the same order, without other changes
// to the values.
func replaceLeaf(t *testing.T, b []byte, n int, v any) []byte {
	t.Helper()
	var w bytes.Buffer
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	count := 0
	var walk func()
	walk = func() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatal(err)
		}
		if d, ok := tok.(json.Delim); ok {
			obj := d == '{'
			w.WriteString(string(d))
			first := true
			for dec.More() {
				if !first {
					w.WriteByte(',')
				}
				first = false
				if obj {
					k, err := dec.Token()
					if err != nil {
						t.Fatal(err)
					}
					kb, _ := json.Marshal(k)
					w.Write(kb)
					w.WriteByte(':')
				}
				walk()
			}
			end, err := dec.Token()
			if err != nil {
				t.Fatal(err)
			}
			w.WriteString(string(end.(json.Delim)))
			return
		}
		use := tok
		if count == n {
			use = v
		}
		count++
		var enc bytes.Buffer
		e := json.NewEncoder(&enc)
		e.SetEscapeHTML(false)
		if err := e.Encode(use); err != nil {
			t.Fatal(err)
		}
		w.Write(bytes.TrimSuffix(enc.Bytes(), []byte("\n")))
	}
	walk()
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		t.Fatalf("trailing data after the value: %v", err)
	}
	return w.Bytes()
}

// forgeries lists every one-value change of the pack's files. The signature file, which holds one
// value that is not JSON, is the last of them.
func forgeries(t *testing.T, files map[string][]byte) []forgery {
	t.Helper()
	var out []forgery
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		body := files[name]
		switch name {
		case SignatureName:
			continue
		case TraceName:
			lines := strings.Split(strings.TrimSuffix(string(body), "\n"), "\n")
			for li, line := range lines {
				for i, v := range leaves(t, []byte(line)) {
					for vi, make := range variants(v) {
						li, i, make := li, i, make
						out = append(out, forgery{file: name, desc: fmt.Sprintf("line %d value %d variant %d", li+1, i, vi), apply: func() []byte {
							edited := append([]string(nil), lines...)
							edited[li] = string(replaceLeaf(t, []byte(line), i, make()))
							return []byte(strings.Join(edited, "\n") + "\n")
						}})
					}
				}
			}
		default:
			for i, v := range leaves(t, body) {
				for vi, make := range variants(v) {
					i, make := i, make
					out = append(out, forgery{file: name, desc: fmt.Sprintf("value %d variant %d", i, vi), apply: func() []byte {
						return replaceLeaf(t, body, i, make())
					}})
				}
			}
		}
	}
	out = append(out, forgery{file: SignatureName, desc: "one character", apply: func() []byte {
		sig := []byte(string(files[SignatureName]))
		if sig[0] == 'A' {
			sig[0] = 'B'
		} else {
			sig[0] = 'A'
		}
		return sig
	}})
	return out
}

type forgery struct {
	file  string
	desc  string
	apply func() []byte
}

// sealed returns the archive of files with the manifest's hashes recomputed for every file the
// manifest lists, and the manifest signed again by key. A forger who changed a manifest value keeps it:
// only the signature is made again.
func sealed(t *testing.T, files map[string][]byte, key ed25519.PrivateKey, manifestEdited bool) []byte {
	t.Helper()
	out := map[string][]byte{}
	for k, v := range files {
		out[k] = v
	}
	if !manifestEdited {
		var m Manifest
		if err := json.Unmarshal(out[ManifestName], &m); err != nil {
			t.Fatal(err)
		}
		for i, ref := range m.Files {
			sum := sha256.Sum256(out[ref.Path])
			m.Files[i].SHA256, m.Files[i].Size = hex.EncodeToString(sum[:]), len(out[ref.Path])
		}
		b, err := marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		out[ManifestName] = b
	}
	out[SignatureName] = []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(key, out[ManifestName])) + "\n")
	b, err := archive(out)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func refusalOf(err error) *Refusal {
	var r *Refusal
	if errors.As(err, &r) {
		return r
	}
	return nil
}

func TestForgeryBench(t *testing.T) {
	in, pub := forgeryInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(pack, pub); err != nil {
		t.Fatalf("the sample pack does not verify before any change: %v", err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	all := forgeries(t, files)
	if len(all) < 100 {
		t.Fatalf("the bench enumerated %d changes; a sample pack of this size has many more", len(all))
	}

	t.Run("a-hashes-left", func(t *testing.T) {
		refused := map[string]int{}
		for _, f := range all {
			edited := map[string][]byte{}
			for k, v := range files {
				edited[k] = v
			}
			edited[f.file] = f.apply()
			if bytes.Equal(edited[f.file], files[f.file]) {
				t.Fatalf("%s %s: the change changed nothing", f.file, f.desc)
			}
			b, err := archive(edited)
			if err != nil {
				t.Fatal(err)
			}
			_, err = Verify(b, pub)
			r := refusalOf(err)
			if r == nil || (r.Check != CheckHash && r.Check != CheckSignature) {
				t.Errorf("%s %s: verified or refused by the wrong check (%v)", f.file, f.desc, err)
				continue
			}
			refused[r.Check]++
		}
		t.Logf("forgery bench (a) hashes left: %d changes, %d refused by the hash check, %d refused by the signature check, %d verified",
			len(all), refused[CheckHash], refused[CheckSignature], len(all)-refused[CheckHash]-refused[CheckSignature])
	})

	t.Run("b-hashes-recomputed", func(t *testing.T) {
		var changes, moving, movingRefused, quiet, quietRefused int
		// movingBy and quietBy tally the refusals by the check that made them.
		movingBy, quietBy := map[string]int{}, map[string]int{}
		for _, f := range all {
			if f.file == SignatureName {
				continue // signing again makes the signature file the original's
			}
			edited := map[string][]byte{}
			for k, v := range files {
				edited[k] = v
			}
			edited[f.file] = f.apply()
			b := sealed(t, edited, in.Key, f.file == ManifestName)
			sealedFiles, err := readArchive(b)
			if err != nil {
				t.Fatal(err)
			}
			moves := movesOutcome(files, sealedFiles)
			_, verr := Verify(b, pub)
			changes++
			if moves {
				moving++
				r := refusalOf(verr)
				if r == nil || (r.Check != CheckOutcome && r.Check != CheckFormat) {
					t.Errorf("%s %s: moves an outcome and was not refused by the re-derivation (%v)", f.file, f.desc, verr)
					continue
				}
				movingBy[r.Check]++
				movingRefused++
				continue
			}
			quiet++
			if r := refusalOf(verr); r != nil {
				quietBy[r.Check]++
				quietRefused++
			}
		}
		// The oracle that says which changes move an outcome is review.Rederive, the rule Verify
		// applies. So this count shows that Verify applies the rule to every recorded value; it
		// does not show that the rule is right. The rule is pinned by its own tests.
		t.Logf("forgery bench (b) hashes recomputed and signed again: %d changes, %d move an outcome, %d of those refused (%d by the outcome check, %d by the format check); %d leave every outcome as it was, %d of those refused by another check (%v) and %d verify",
			changes, moving, movingRefused, movingBy[CheckOutcome], movingBy[CheckFormat], quiet, quietRefused, quietBy, quiet-quietRefused)
		if moving == 0 {
			t.Fatal("no change moved an outcome, so the bench measures nothing")
		}
	})
}

// A forger who keeps findings.json and tests.json agreeing still cannot make a test that does not
// fail on the change prove a finding.
func TestAConsistentForgeryOfATestResultIsRefused(t *testing.T) {
	in, pub := forgeryInput(t)
	in.Result.Findings[0].Evidence[0].Test.FailsOnChange = false
	pack, err := build(in)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Verify(pack, pub)
	r := refusalOf(err)
	if r == nil || r.Check != CheckOutcome || r.File != FindingsName {
		t.Fatalf("want an outcome refusal on findings.json, got %v", err)
	}
	for _, want := range []string{"F1", "recorded as reported", "drops it"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q: %v", want, err)
		}
	}
}

// The refusal names the finding, the recorded outcome and the re-derived one.
func TestTheRefusalNamesTheRecordedAndTheRederivedOutcome(t *testing.T) {
	in, pub := forgeryInput(t)
	in.Result.Findings[2].ProofLevel = review.ProvedByTest
	in.Result.ProofCounts = review.CountProofs(in.Result.Findings)
	pack, err := build(in)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Verify(pack, pub)
	r := refusalOf(err)
	if r == nil || r.Check != CheckOutcome {
		t.Fatalf("want an outcome refusal, got %v", err)
	}
	for _, want := range []string{"F3", string(review.ProvedByTest), string(review.ProvedByData)} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q: %v", want, err)
		}
	}
}

// A finding the pack records no refutation for verifies, and is named as not re-derived.
func TestAFindingWithNoRecordedRefutationIsNamedNotRederived(t *testing.T) {
	in, pub := forgeryInput(t)
	in.Result.Findings[0].Refutation = nil
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Verify(pack, pub)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Rederived != 3 || len(rep.NotRederived) != 1 || !strings.Contains(rep.NotRederived[0], "F1") {
		t.Fatalf("rederived %d, not rederived %v", rep.Rederived, rep.NotRederived)
	}
}

func TestEveryRecordedFindingOfTheSampleIsRederived(t *testing.T) {
	in, pub := forgeryInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Verify(pack, pub)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Rederived != len(in.Result.Findings) || len(rep.NotRederived) != 0 {
		t.Fatalf("rederived %d of %d, not rederived %v", rep.Rederived, len(in.Result.Findings), rep.NotRederived)
	}
}
