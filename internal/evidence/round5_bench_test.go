package evidence

// Bench part (d): the manifest marks a finding redaction-moved, and the same pack changes a
// component the recorded evidence decides.

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/review"
)

// Where the redacted text really decides a component, only that component is named not
// re-derived; the open state, which both readings give as false, is still checked.
func TestRedactionNamesOnlyTheComponentsTheTextDecides(t *testing.T) {
	in, pub := forgeryInput(t)
	secretA, secretB := strings.Repeat("A", 32), strings.Repeat("B", 32)
	f := &in.Result.Findings[2]
	f.Message = "pkg-a 1.2.0 is in an advisory; the lock file pins " + secretA
	f.Refutation = &review.Refutation{QueryRows: [][]string{{secretB, "no"}}}
	in.Redactor = redact.New(map[string]string{"TOKEN_A": secretA, "TOKEN_B": secretB})
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Verify(pack, pub)
	if err != nil {
		t.Fatalf("a genuine pack is refused after redaction: %v", err)
	}
	var line string
	for _, s := range rep.NotRederived {
		if strings.Contains(s, "finding F3:") {
			line = s
		}
	}
	if !strings.Contains(line, "whether it is reported") || strings.Contains(line, "whether it is open") || strings.Contains(line, "proof level") {
		t.Fatalf("F3 should be named not re-derived for whether it is reported only (open is false in both readings), got %q", line)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	editJSON(t, files, FindingsName, func(v any) {
		root := v.(map[string]any)
		for _, x := range root["Findings"].([]any) {
			if g := x.(map[string]any); g["ID"] == "F3" {
				g["Unresolved"] = true
			}
		}
	})
	if _, err := Verify(sealed(t, files, in.Key, false), pub); refusalOf(err) == nil {
		t.Fatalf("a changed open state of a redaction-moved finding verified: %v", err)
	}
}

func TestForgeryBenchRedactionMarked(t *testing.T) {
	in, pub := forgeryInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	type change struct{ id, desc, component string }
	var changes []change
	byID := map[string]review.Finding{}
	for _, f := range in.Result.Findings {
		byID[f.ID] = f
		changes = append(changes, change{f.ID, "open state flipped", "open"})
		for _, l := range review.ProofLevels {
			if l != f.ProofLevel {
				changes = append(changes, change{f.ID, "level moved to " + string(l), string(l)})
			}
		}
	}
	var refused, notRederived, wrong int
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
				if c.component == "open" {
					f["Unresolved"] = !f["Unresolved"].(bool)
					continue
				}
				counts := root["ProofCounts"].(map[string]any)
				one := func(k string, d int64) {
					n, _ := counts[k].(json.Number).Int64()
					counts[k] = json.Number(strconv.FormatInt(n+d, 10))
				}
				one(f["ProofLevel"].(string), -1)
				one(c.component, 1)
				f["ProofLevel"] = c.component
			}
		})
		var m Manifest
		if err := json.Unmarshal(edited[ManifestName], &m); err != nil {
			t.Fatal(err)
		}
		m.RedactionMoved = []string{c.id}
		edited[ManifestName], err = marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		rep, err := Verify(sealed(t, edited, in.Key, false), pub)
		// The text the rule reads decides the open state only for a finding whose refuting query
		// rows the text could name; every other component is decided by the rest of the evidence.
		textDecides := c.component == "open" && textMatters(byID[c.id])
		if r := refusalOf(err); r != nil && r.Check == CheckOutcome {
			refused++
			continue
		}
		if err == nil && textDecides && slices.ContainsFunc(rep.NotRederived, func(s string) bool {
			return strings.Contains(s, "finding "+c.id+":") && strings.Contains(s, "whether it is open")
		}) {
			notRederived++
			continue
		}
		wrong++
		t.Errorf("%s marked redaction-moved, %s: not refused by the outcome check and not named not re-derived (%v)", c.id, c.desc, err)
	}
	t.Logf("forgery bench (d) redaction-moved with a decidable component changed: %d changes, %d refused by the outcome check, %d verified and named not re-derived (text decides), %d verified unnamed", len(changes), refused, notRederived, wrong)
}
