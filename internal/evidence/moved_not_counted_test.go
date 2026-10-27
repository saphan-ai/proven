package evidence

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

// A finding listed in redaction_moved whose readings disagree is named and NOT counted: the
// re-derived count drops by exactly one against the same pack with the list empty. Mutated field:
// the count of a named finding (rep.Rederived++ for a moved finding makes the two counts equal).
func TestMovedFindingIsNotCounted(t *testing.T) {
	in, pub := forgeryInput(t)
	in.Result.Findings[3].Refutation = &review.Refutation{TestCannotRun: true, QueryRows: [][]string{{"zzzz", "no"}}}
	pack, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	base, err := Verify(pack, pub)
	if err != nil {
		t.Fatalf("the genuine pack does not verify: %v", err)
	}
	files, err := readArchive(pack)
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err := json.Unmarshal(files[ManifestName], &m); err != nil {
		t.Fatal(err)
	}
	m.RedactionMoved = []string{"F4"}
	if files[ManifestName], err = marshal(m); err != nil {
		t.Fatal(err)
	}
	rep, err := Verify(sealed(t, files, in.Key, false), pub)
	if err != nil {
		t.Fatalf("a pack that one reading of the redacted text matches is refused: %v", err)
	}
	named := false
	for _, s := range rep.NotRederived {
		named = named || strings.Contains(s, "finding F4:")
	}
	if !named {
		t.Fatalf("F4 is listed in redaction_moved and its readings disagree, yet it is not named: %q", rep.NotRederived)
	}
	if rep.Rederived != base.Rederived-1 {
		t.Fatalf("F4 is named not re-derived, yet the count is %d (the same pack with no list counts %d): a named finding was counted", rep.Rederived, base.Rederived)
	}
}
