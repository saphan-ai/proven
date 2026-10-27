package change

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// treeOf lists every file under dir with its mode and content, so two exports can be compared whole.
func treeOf(t *testing.T, dir string) map[string]string {
	t.Helper()
	got := map[string]string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		if info.IsDir() {
			got[rel+"/"] = "dir"
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		got[rel] = info.Mode().String() + " " + string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// TestLocalExportIsTheSameTreeAsTheArchive: reading the blobs as stored gives the files git archive gives
// for a commit that no attribute converts: the same paths, modes and bytes, a link left out and named,
// and a repository whose target is a subdirectory. Mutated field: the export path (local mode or not).
func TestLocalExportIsTheSameTreeAsTheArchive(t *testing.T) {
	repo := t.TempDir()
	git(t, repo, "init")
	writeTree(t, filepath.Join(repo, "a", "b", "f.txt"), "deep\n")
	writeTree(t, filepath.Join(repo, "a", "run.sh"), "#!/bin/sh\n")
	writeTree(t, filepath.Join(repo, "top.txt"), "top\n")
	writeTree(t, filepath.Join(repo, "empty.txt"), "")
	writeTree(t, filepath.Join(repo, "sp ace", "tab\tname.txt"), "odd name\n")
	if err := os.Chmod(filepath.Join(repo, "a", "run.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("top.txt", filepath.Join(repo, "link")); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-m", "[change] export: first")

	for _, target := range []string{repo, filepath.Join(repo, "a")} {
		online, onlineSkipped := t.TempDir(), []string(nil)
		var err error
		if onlineSkipped, err = ExportCommit(target, "HEAD", online); err != nil {
			t.Fatal(err)
		}
		local := t.TempDir()
		prev := UseOffline(true)
		localSkipped, err := ExportCommit(target, "HEAD", local)
		UseOffline(prev)
		if err != nil {
			t.Fatalf("local export of %s: %v", target, err)
		}
		if !reflect.DeepEqual(treeOf(t, online), treeOf(t, local)) {
			t.Errorf("target %s: local export %v differs from the archive %v", target, treeOf(t, local), treeOf(t, online))
		}
		if !reflect.DeepEqual(onlineSkipped, localSkipped) {
			t.Errorf("target %s: skipped %v, the archive skipped %v", target, localSkipped, onlineSkipped)
		}
	}
}

// TestLocalExportRefusesAFileOverTheBoundWithoutReadingIt: the bounds of the archive hold for the blob
// reader: a file over the file limit, and files whose sizes together pass the archive limit, are refused
// by the limit's own name. Mutated field: the limits.
func TestLocalExportRefusesAFileOverTheBound(t *testing.T) {
	repo, _ := commitFile(t, "big.txt", 100)
	writeTree(t, filepath.Join(repo, "b.txt"), strings.Repeat("y", 100))
	git(t, repo, "add", "b.txt")
	git(t, repo, "commit", "-m", "[change] export: second")
	defer UseOffline(UseOffline(true))
	defer UseLimits(UseLimits(Limits{File: 50}))
	if _, err := ExportCommit(repo, "HEAD", t.TempDir()); err == nil || !strings.Contains(err.Error(), EnvMaxFileBytes) {
		t.Fatalf("a file over the file limit: %v", err)
	}
	UseLimits(Limits{File: 1000, Archive: 150})
	if _, err := ExportCommit(repo, "HEAD", t.TempDir()); err == nil || !strings.Contains(err.Error(), EnvMaxArchiveBytes) {
		t.Fatalf("files over the archive limit: %v", err)
	}
	UseLimits(Limits{File: 1000, Archive: 1000})
	if _, err := ExportCommit(repo, "HEAD", t.TempDir()); err != nil {
		t.Fatalf("files under both limits: %v", err)
	}
}

// TestLocalExportNamesAMissingObjectOfAPartialClone: the blob reader fetches nothing. In a partial clone
// the missing blob is named with the setting and the clone, and is still missing afterwards.
// Mutated field: the guard of GitCommand; with it removed git fetches the blob and the export succeeds.
func TestLocalExportNamesAMissingObjectOfAPartialClone(t *testing.T) {
	clone := partialClone(t)
	defer UseOffline(UseOffline(true))
	_, err := ExportCommit(clone, "HEAD", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "PROVEN_NETWORK=local") || !strings.Contains(err.Error(), "partial clone") {
		t.Fatalf("local export of a partial clone: %v", err)
	}
}
