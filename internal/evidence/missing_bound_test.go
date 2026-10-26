package evidence

// A missing Bound makes only the components that depend on it undecidable.
// Two report lines were left untested.

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestMissingIrrelevantBoundDoesNotHideWrongOutcome(t *testing.T) {
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
		if name == FindingsName {
			pos := bytes.Index(files[name], []byte(`"ID": "F4"`))
			if pos < 0 {
				t.Fatal("F4 missing")
			}
			suffix := regexp.MustCompile(`,\s*"Bound": true`).ReplaceAll(files[name][pos:], nil)
			files[name] = append(append([]byte(nil), files[name][:pos]...), suffix...)
			pos = bytes.Index(files[name], []byte(`"ID": "F4"`))
			suffix = bytes.Replace(files[name][pos:], []byte(`"Unresolved": true`), []byte(`"Unresolved": false`), 1)
			files[name] = append(append([]byte(nil), files[name][:pos]...), suffix...)
		} else {
			pos := bytes.Index(files[name], []byte(`"finding_id": "F4"`))
			if pos < 0 {
				t.Fatal("F4 query missing")
			}
			suffix := regexp.MustCompile(`,\s*"bound": true`).ReplaceAll(files[name][pos:], nil)
			files[name] = append(append([]byte(nil), files[name][:pos]...), suffix...)
		}
	}
	_, err = Verify(sealed(t, files, in.Key, false), pub)
	if err == nil {
		t.Fatal("verify accepted F4 as closed although its counting test and recorded refutation make it unresolved regardless of the query's missing Bound")
	}
}

// editJSON decodes one pack file, lets edit change it, and writes it back the way the pack does.
func editJSON(t *testing.T, files map[string][]byte, name string, edit func(any)) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(files[name]))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		t.Fatal(err)
	}
	edit(v)
	b, err := marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	files[name] = b
}

// The forgery bench's part (c): a forger removes the Bound field of a finding's queries, so that
// the finding is at least partly undecidable, and in the same pack changes a result the rest of
// the recorded evidence still decides. Hashes are recomputed and the manifest is signed again.
// Part (b) of the bench changes one recorded value at a time and never removes a field.
func TestForgeryBenchBoundRemoved(t *testing.T) {
	in, pub := forgeryInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	type change struct {
		id, desc string
		edit     func(f map[string]any, counts map[string]any)
	}
	flip := func(level string) func(map[string]any, map[string]any) {
		return func(f, counts map[string]any) {
			old := f["ProofLevel"].(string)
			f["ProofLevel"] = level
			one := func(k string, d int64) {
				n, _ := counts[k].(json.Number).Int64()
				counts[k] = json.Number(strconv.FormatInt(n+d, 10))
			}
			one(old, -1)
			one(level, 1)
		}
	}
	changes := []change{
		{"F4", "open state flipped (a counting test decides it)", func(f, _ map[string]any) { f["Unresolved"] = false }},
		{"F4", "level moved to proved-by-data (a counting test decides it)", flip("proved-by-data")},
		{"F3", "open state flipped (no refutation can cannot-run it)", func(f, _ map[string]any) { f["Unresolved"] = true }},
		{"F3", "level moved to proved-by-test (only a query is recorded)", flip("proved-by-test")},
	}
	refused := 0
	for _, c := range changes {
		edited := map[string][]byte{}
		for k, v := range files {
			edited[k] = v
		}
		editJSON(t, edited, FindingsName, func(v any) {
			root := v.(map[string]any)
			for _, x := range root["Findings"].([]any) {
				f := x.(map[string]any)
				if f["ID"] != c.id {
					continue
				}
				for _, e := range f["Evidence"].([]any) {
					if q, ok := e.(map[string]any)["Query"].(map[string]any); ok {
						delete(q, "Bound")
					}
				}
				c.edit(f, root["ProofCounts"].(map[string]any))
			}
		})
		editJSON(t, edited, QueriesName, func(v any) {
			for _, x := range v.([]any) {
				if q := x.(map[string]any); q["finding_id"] == c.id {
					delete(q, "bound")
				}
			}
		})
		_, err := Verify(sealed(t, edited, in.Key, false), pub)
		if r := refusalOf(err); r == nil || r.Check != CheckOutcome {
			t.Errorf("%s, Bound removed and %s: not refused by the outcome check (%v)", c.id, c.desc, err)
			continue
		}
		refused++
	}
	t.Logf("forgery bench (c) Bound removed with a decidable result changed: %d changes, %d refused by the outcome check, %d verified", len(changes), refused, len(changes)-refused)
}

// "every" is said only when every finding was re-derived.
func TestTheEverySentenceIsNotSaidWhenAFindingIsNotRederived(t *testing.T) {
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
	named := false
	for _, c := range rep.ChecksPassed {
		if strings.Contains(c, "outcome of every finding") {
			t.Fatalf("verify says every outcome was re-derived although F1 was not: %q", c)
		}
		named = named || strings.Contains(c, "(1 not re-derived)")
	}
	if !named {
		t.Fatalf("the passed checks do not say how many were not re-derived: %v", rep.ChecksPassed)
	}
}
