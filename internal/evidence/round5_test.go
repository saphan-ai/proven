package evidence

// Round 4 of verify-rederives: a redaction marker must not skip the outcome checks the remaining
// evidence still decides.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestReviewRedactionMarkerCannotHideDecidableOpenState(t *testing.T) {
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
			f := x.(map[string]any)
			if f["ID"] == "F4" {
				f["Unresolved"] = false
			}
		}
	})
	var m Manifest
	if err := json.Unmarshal(files[ManifestName], &m); err != nil {
		t.Fatal(err)
	}
	m.RedactionMoved = []string{"F4"}
	files[ManifestName], err = marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var updated Manifest
	if err := json.Unmarshal(files[ManifestName], &updated); err != nil {
		t.Fatal(err)
	}
	for i, ref := range updated.Files {
		if ref.Path == FindingsName {
			updated.Files[i].Size = len(files[FindingsName])
			sum := sha256.Sum256(files[FindingsName])
			updated.Files[i].SHA256 = hex.EncodeToString(sum[:])
		}
	}
	files[ManifestName], err = marshal(updated)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Verify(sealed(t, files, in.Key, true), pub)
	if r := refusalOf(err); r == nil || r.Check != CheckOutcome {
		t.Fatalf("F4's recorded refutation makes it open regardless of any text redaction: %v", err)
	}
}
