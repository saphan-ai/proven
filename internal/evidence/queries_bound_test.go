package evidence

import "testing"

// queries.json at this build always writes "bound". A pack whose queries.json records F3's query as
// not bound, while findings.json leaves Bound out, records the fact: the query is not bound. Then
// the rule drops F3 (its only evidence does not count), so a pack recording F3 reported must be
// refused, not verified with F3 named not re-derived.
func TestBoundInQueriesJSONIsRead(t *testing.T) {
	in, pub := forgeryInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
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
		}
	})
	editJSON(t, files, QueriesName, func(v any) {
		for _, x := range v.([]any) {
			if q := x.(map[string]any); q["finding_id"] == "F3" {
				q["bound"] = false
			}
		}
	})
	rep, err := Verify(sealed(t, files, in.Key, false), pub)
	if r := refusalOf(err); r == nil || r.Check != CheckOutcome {
		var named []string
		if rep != nil {
			named = rep.NotRederived
		}
		t.Fatalf("queries.json records F3's rows as not bound, yet F3 recorded reported verified (named %q, err %v)", named, err)
	}
}
