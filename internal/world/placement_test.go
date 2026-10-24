package world_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
	"github.com/saphan-ai/proven/internal/world"
)

func openMemory(t *testing.T) *hmem.Store {
	t.Helper()
	mem, err := hmem.Open(hmem.Options{Dir: t.TempDir(), LogDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	return mem
}

// TestOpenRefusesAStateDirectoryReachedThroughALinkIntoTheRepository proves the
// inside-the-repository refusal compares resolved paths: a state path that is a
// link to a directory inside the repository is refused.
func TestOpenRefusesAStateDirectoryReachedThroughALinkIntoTheRepository(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "inside"), 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "state")
	if err := os.Symlink(filepath.Join(root, "inside"), link); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	_, err := world.Open(world.Options{Dir: filepath.Join(link, "world"), LogDir: t.TempDir(), RepoRoot: root, Memory: openMemory(t)})
	if err == nil {
		t.Fatal("a state directory that resolves inside the repository through a link was accepted")
	}
	if _, serr := os.Lstat(filepath.Join(root, "inside", "world")); serr == nil {
		t.Fatal("the refused state directory was created inside the repository")
	}
}

// TestOpenNarrowsAnExistingWideStateFile proves a state file an earlier run left
// at a wider mode is narrowed to 0600 when the store opens.
func TestOpenNarrowsAnExistingWideStateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := os.WriteFile(path, []byte(`{"version":1,"applied":[],"findings":[],"entities":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := world.Open(world.Options{Dir: dir, LogDir: t.TempDir(), Memory: openMemory(t)}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("world state mode %o after Open, want 600", got)
	}
}

// TestAtomicWriteDoesNotCarryAWiderTemporaryMode proves the state file written
// through the temporary file is 0600 even when a wider temporary file already
// exists: the mode of an old temporary file never passes through the rename.
func TestAtomicWriteDoesNotCarryAWiderTemporaryMode(t *testing.T) {
	dir := t.TempDir()
	tmp := filepath.Join(dir, "state.json.tmp")
	if err := os.WriteFile(tmp, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(tmp, 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := world.Open(world.Options{Dir: dir, LogDir: t.TempDir(), Memory: openMemory(t)})
	if err != nil {
		t.Fatal(err)
	}
	err = s.UpsertEntity(world.Entity{
		Type: "x", ID: "ent_p", Name: "Private",
		Source: world.Provenance{Kind: "manual", At: "2026-01-01T00:00:00.000Z"},
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("state.json mode %o after a write over a wider temporary file, want 600", got)
	}
}
