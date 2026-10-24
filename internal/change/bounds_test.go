package change

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// commitFile makes a repository in a temporary directory with one committed
// file of size bytes and returns the directory and the commit.
func commitFile(t *testing.T, name string, size int) (dir, commit string) {
	t.Helper()
	dir = t.TempDir()
	git(t, dir, "init")
	write(t, filepath.Join(dir, name), strings.Repeat("x", size))
	git(t, dir, "add", name)
	git(t, dir, "commit", "-m", "[change] git: first")
	return dir, rev(t, dir)
}

// writeTree writes body at path and creates the parent directories.
func writeTree(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, path, body)
}

// Property: a revision that begins with a dash is refused as a revision by
// every reader that takes one, and is never read as a git option.
func TestDashLeadingRevisionIsRefusedByEveryReader(t *testing.T) {
	dir, base := commitFile(t, "a.txt", 1)
	probe := filepath.Join(t.TempDir(), "written-if-read-as-option")
	hostile := "--output=" + probe

	// Mutated field: only the revision. The refusal must be the "does not
	// exist" sentence of the commit check, so the guard before git archive is
	// observed and not only git's own refusal.
	_, err := ExportCommit(dir, hostile, t.TempDir())
	if err == nil {
		t.Fatal("ExportCommit accepted a revision that begins with a dash")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("ExportCommit refused the revision without the commit check: %v", err)
	}
	if _, err := Between(dir, hostile, ""); err == nil {
		t.Fatal("Between accepted a baseline that begins with a dash")
	}
	if _, err := Between(dir, base, hostile); err == nil {
		t.Fatal("Between accepted a head that begins with a dash")
	}
	if _, err := Messages(dir, hostile, ""); err == nil {
		t.Fatal("Messages accepted a baseline that begins with a dash")
	}
	if _, _, err := Blob(dir, hostile, "a.txt"); err == nil {
		t.Fatal("Blob accepted a revision that begins with a dash")
	}
	if _, err := os.Stat(probe); err == nil {
		t.Fatal("a revision that begins with a dash was read as a git option")
	}
}

// Property: a prepared tree file larger than the file bound is refused, and the
// sentence names the file and the bound, never the content.
func TestReadRefusesAFileOverTheFileBound(t *testing.T) {
	root := t.TempDir()
	writeTree(t, filepath.Join(root, "base", "a.txt"), "a")
	writeTree(t, filepath.Join(root, "change", "a.txt"), strings.Repeat("secret-body", (9<<20)/11+1))
	_, err := Read(root)
	if err == nil {
		t.Fatal("a file over the file bound was read")
	}
	if !strings.Contains(err.Error(), "a.txt") || strings.Contains(err.Error(), "secret-body") {
		t.Fatalf("the refusal must name the file and not quote it: %v", err)
	}
}

// Property: a committed file larger than the file bound is refused when the
// commit is exported, whichever way it is exported.
func TestExportRefusesAFileOverTheFileBound(t *testing.T) {
	dir, commit := commitFile(t, "big.txt", 9<<20)
	if _, err := ExportCommit(dir, commit, t.TempDir()); err == nil {
		t.Fatal("ExportCommit wrote a committed file over the file bound")
	}
	if _, err := ExportWorktree(dir, t.TempDir()); err == nil {
		t.Fatal("ExportWorktree copied a tracked file over the file bound")
	}
}

// Property: a file at the file bound is read, so the bound is a ceiling and not
// a refusal of every large file.
func TestReadAcceptsAFileAtTheFileBound(t *testing.T) {
	root := t.TempDir()
	writeTree(t, filepath.Join(root, "base", "a.txt"), "a")
	writeTree(t, filepath.Join(root, "change", "a.txt"), strings.Repeat("x", int(DefaultLimits().File)))
	if _, err := Read(root); err != nil {
		t.Fatalf("a file at the bound was refused: %v", err)
	}
}

// Property: the archive stream has its own bound, so many small files cannot
// add up to an unbounded read.
func TestExportRefusesAnArchiveOverTheArchiveBound(t *testing.T) {
	defer UseLimits(UseLimits(Limits{Archive: 4 << 10, File: 1 << 20, Config: 1 << 20, Record: 1 << 20}))
	dir, commit := commitFile(t, "a.txt", 8<<10)
	_, err := ExportCommit(dir, commit, t.TempDir())
	if err == nil {
		t.Fatal("ExportCommit read an archive over the archive bound")
	}
	if !strings.Contains(err.Error(), "4096") {
		t.Fatalf("the refusal must name the bound: %v", err)
	}
}

// Property: a tracked file read from the working tree counts against the
// archive bound as well, so the two exports are bounded the same way.
func TestExportWorktreeRefusesATreeOverTheArchiveBound(t *testing.T) {
	defer UseLimits(UseLimits(Limits{Archive: 4 << 10, File: 1 << 20, Config: 1 << 20, Record: 1 << 20}))
	dir, _ := commitFile(t, "a.txt", 8<<10)
	if _, err := ExportWorktree(dir, t.TempDir()); err == nil {
		t.Fatal("ExportWorktree copied a tree over the archive bound")
	}
}

// Property: a blob and a commit log over the file bound are refused.
func TestBlobAndMessagesRefuseOutputOverTheFileBound(t *testing.T) {
	defer UseLimits(UseLimits(Limits{Archive: 1 << 20, File: 1 << 10, Config: 1 << 20, Record: 1 << 20}))
	dir, commit := commitFile(t, "proven.yaml", 4<<10)
	if _, _, err := Blob(dir, commit, "proven.yaml"); err == nil {
		t.Fatal("Blob returned a file over the file bound")
	}
	write(t, filepath.Join(dir, "b.txt"), "b")
	git(t, dir, "add", "b.txt")
	// A valid subject, so the commit hook accepts it; the size is in the body.
	git(t, dir, "commit", "-m", "[change] git: second", "-m", strings.Repeat("m", 4<<10))
	if _, err := Messages(dir, commit, ""); err == nil {
		t.Fatal("Messages returned a log over the file bound")
	}
}

// TestLimitAboveTheLargestBoundIsRefused proves a size bound cannot overflow the
// readers' limit+1: the largest accepted bound still reads a file whole, and a
// bound above it (the int64 maximum included) is refused, never read as empty.
func TestLimitAboveTheLargestBoundIsRefused(t *testing.T) {
	for _, raw := range []string{"9223372036854775807", "1099511627777"} {
		if _, err := LoadLimits(func(k string) string {
			if k == EnvMaxFileBytes {
				return raw
			}
			return ""
		}); err == nil {
			t.Fatalf("bound %s was accepted", raw)
		}
	}
	l, err := LoadLimits(func(k string) string {
		if k == EnvMaxFileBytes {
			return "1099511627776"
		}
		return ""
	})
	if err != nil || l.File != 1<<40 {
		t.Fatalf("largest bound refused: %v %d", err, l.File)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := ReadBounded(path, l.File, "the file f.txt", EnvMaxFileBytes)
	if err != nil || string(b) != "content" {
		t.Fatalf("read at the largest bound: %q %v", b, err)
	}
}
