package evidence

import (
	"os"
	"path/filepath"
	"testing"
)

// The mutated field is the whole-file bound VerifyFile holds a pack to
// (maxPack). VerifyFile calls the seam with maxPack; the seam takes the bound,
// so the guard can be unplugged and seen to go red: a pack one byte over the
// bound is refused by the format check naming the bound, and a pack at the
// bound is verified.
func TestAPackOverTheWholeFileBoundIsRefused(t *testing.T) {
	pack, _, pub := builtPack(t)
	path := filepath.Join(t.TempDir(), "pack.tar")
	if err := os.WriteFile(path, pack, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := verifyFileWithin(path, pub, int64(len(pack))-1)
	wantCheck(t, err, CheckFormat, "byte bound")
	if _, err := verifyFileWithin(path, pub, int64(len(pack))); err != nil {
		t.Fatalf("a pack at the bound must verify: %v", err)
	}
}
