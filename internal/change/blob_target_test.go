package change

import (
	"os"
	"path/filepath"
	"testing"
)

// A subdirectory target reads its own proven.yaml at the baseline, not the root's.
func TestBlobReadsTheTargetsOwnFileInASubdirectory(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	sub := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "proven.yaml"), "timeout: 1m\n")
	write(t, filepath.Join(sub, "proven.yaml"), "timeout: 5m\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "[change] git: base")
	base := rev(t, dir)

	got, ok, err := Blob(sub, base, "proven.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || string(got) != "timeout: 5m\n" {
		t.Fatalf("ok %v, got %q; want the subdirectory's own file", ok, got)
	}
}

// A root proven.yaml is not taken for a subdirectory target that has none.
func TestBlobDoesNotTakeTheRootFileForASubdirectory(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init")
	sub := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "proven.yaml"), "timeout: 1m\n")
	write(t, filepath.Join(sub, "a.txt"), "a")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-m", "[change] git: base")
	base := rev(t, dir)

	got, ok, err := Blob(sub, base, "proven.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("took %q for the subdirectory; want absent", got)
	}
}
