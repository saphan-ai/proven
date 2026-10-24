package hmem_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saphan-ai/proven/internal/hmem"
)

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
	_, err := hmem.Open(hmem.Options{Dir: filepath.Join(link, "hmem"), LogDir: t.TempDir(), RepoRoot: root})
	if err == nil {
		t.Fatal("a state directory that resolves inside the repository through a link was accepted")
	}
	if _, serr := os.Lstat(filepath.Join(root, "inside", "hmem")); serr == nil {
		t.Fatal("the refused state directory was created inside the repository")
	}
}

// TestOpenRefusesAStateDirectoryBehindALinkThatDoesNotResolve proves a link that
// cannot be resolved is refused rather than trusted.
func TestOpenRefusesAStateDirectoryBehindALinkThatDoesNotResolve(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(t.TempDir(), "state")
	if err := os.Symlink(filepath.Join(root, "missing"), link); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	if _, err := hmem.Open(hmem.Options{Dir: filepath.Join(link, "hmem"), LogDir: t.TempDir(), RepoRoot: root}); err == nil {
		t.Fatal("a state directory behind a dangling link was accepted")
	}
}

// TestOpenNarrowsAnExistingWideStateFile proves a state file an earlier run left
// at a wider mode is narrowed to 0600 when the store opens.
func TestOpenNarrowsAnExistingWideStateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := os.WriteFile(path, []byte(`{"version":1,"seq":0,"records":[]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir()}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("state.json mode %o after Open, want 600", got)
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
	s, err := hmem.Open(hmem.Options{Dir: dir, LogDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add(hmem.AddInput{Content: "a private fact", Source: "test", At: fixed}); err != nil {
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
