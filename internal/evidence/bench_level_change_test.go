package evidence

import (
	"encoding/json"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// The bench counts a change as moving an outcome when the rule's level changes, even though the
// level the pack records does not.
func TestBenchCountsRuleLevelChange(t *testing.T) {
	in, _ := forgeryInput(t)
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	orig, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"FailsOnChange", "PassesOnBase"} {
		t.Run(field, func(t *testing.T) {
			edited := make(map[string][]byte, len(orig))
			for k, v := range orig {
				edited[k] = v
			}
			editJSON(t, edited, FindingsName, func(v any) {
				for _, raw := range v.(map[string]any)["Findings"].([]any) {
					f := raw.(map[string]any)
					if f["ID"] != "F4" {
						continue
					}
					f["Evidence"].([]any)[0].(map[string]any)["Test"].(map[string]any)[field] = false
				}
			})
			var result review.Result
			if err := json.Unmarshal(edited[FindingsName], &result); err != nil {
				t.Fatal(err)
			}
			got, err := review.Rederive(result.Findings[3])
			if err != nil {
				t.Fatal(err)
			}
			if got.Level != review.ProvedByData {
				t.Fatalf("setup: got rule level %q", got.Level)
			}
			if !movesOutcome(orig, edited) {
				t.Fatal("bench classified a rule-level change as non-moving; count every rule outcome change")
			}
		})
	}
}
