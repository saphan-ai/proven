package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// aliasRow is one pair of spellings that reach one file. setup builds the
// disk under dir (the working directory of the run) and returns the two paths
// exactly as they are typed on the command line: raw, never cleaned by
// filepath.Join, because the defect lived in what happens to the raw string.
type aliasRow struct {
	name  string
	setup func(t *testing.T, dir string) (first, second string)
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func symlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks are not available here: %v", err)
	}
}

// aliasTable is the class of aliases the reviews found one at a time. Each row
// names the alias it covers; none of the files exists unless the row says so.
var aliasTable = []aliasRow{
	{"identical spelling", func(t *testing.T, dir string) (string, string) {
		return "report", "report"
	}},
	{"dot segment and a/../ segment", func(t *testing.T, dir string) (string, string) {
		mkdir(t, filepath.Join(dir, "a"))
		return "report", "./a/../report"
	}},
	{"relative versus absolute", func(t *testing.T, dir string) (string, string) {
		return "report", dir + string(filepath.Separator) + "report"
	}},
	{"symlinked directory", func(t *testing.T, dir string) (string, string) {
		mkdir(t, filepath.Join(dir, "real"))
		symlink(t, "real", filepath.Join(dir, "link"))
		return "real/report", "link/report"
	}},
	{"symlinked directory then ..", func(t *testing.T, dir string) (string, string) {
		// link -> elsewhere/dir: the kernel resolves link first, so
		// link/.. is elsewhere, and link/../report is elsewhere/report.
		mkdir(t, filepath.Join(dir, "elsewhere", "dir"))
		symlink(t, "elsewhere/dir", filepath.Join(dir, "link"))
		return "elsewhere/report", "link/../report"
	}},
	{"chain of two directory links", func(t *testing.T, dir string) (string, string) {
		mkdir(t, filepath.Join(dir, "real"))
		symlink(t, "real", filepath.Join(dir, "mid"))
		symlink(t, "mid", filepath.Join(dir, "link"))
		return "real/report", "link/report"
	}},
	{"dangling final link", func(t *testing.T, dir string) (string, string) {
		symlink(t, "actual", filepath.Join(dir, "alias"))
		return "actual", "alias"
	}},
	{"final link to a link", func(t *testing.T, dir string) (string, string) {
		symlink(t, "actual", filepath.Join(dir, "inner"))
		symlink(t, "inner", filepath.Join(dir, "outer"))
		return "actual", "outer"
	}},
	{"hard link to an existing file", func(t *testing.T, dir string) (string, string) {
		if err := os.WriteFile(filepath.Join(dir, "first"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Link(filepath.Join(dir, "first"), filepath.Join(dir, "second")); err != nil {
			t.Skipf("hard links are not available here: %v", err)
		}
		return "first", "second"
	}},
	{"trailing slash on the directory", func(t *testing.T, dir string) (string, string) {
		mkdir(t, filepath.Join(dir, "real"))
		symlink(t, "real/", filepath.Join(dir, "link"))
		return "real/report", "link//report"
	}},
}

// TestAliasTableIsRefusedAtLoad proves that every alias in the table is
// refused at load: exit 7, both flags named, the review never runs and no file
// is written. The mutated field is the second spelling of each row. With
// outputIdentity going through filepath.Abs before resolving links (the old
// path) the "symlinked directory then .." row returns exit 2 after -o was
// written, because Abs cleans link/../report to report before link is read.
func TestAliasTableIsRefusedAtLoad(t *testing.T) {
	for _, row := range aliasTable {
		t.Run(row.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			first, second := row.setup(t, dir)
			rev := &captureReviewer{}
			var stdout, stderr bytes.Buffer
			code := Run([]string{"scan", "--json", "-o", first, "--sarif-output", second}, emptyEnv, &stdout, &stderr, openCapture(rev))
			if code != exitConfig {
				t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
			}
			for _, want := range []string{"-o", "--sarif-output"} {
				if !strings.Contains(stderr.String(), want) {
					t.Errorf("refusal does not name %q: %s", want, stderr.String())
				}
			}
			if rev.req.Target != "" {
				t.Error("the review ran despite the refusal")
			}
			if _, err := os.Stat(filepath.Join(dir, first)); err == nil && row.name != "hard link to an existing file" {
				t.Error("a file was written despite the refusal")
			}
		})
	}
}

// TestAliasTableIsCaughtByTheLoadCheckNotTheWriteGuard proves agreement: for
// every row, refuseCollidingOutputs itself returns the refusal, so the write
// guard in run.go is never the one that fires for an alias in the table. The
// mutated field is the same second spelling; with the load check emptied, this
// test is red on every row while TestAliasTableIsRefusedAtLoad would still see
// exit 2 from the write guard on some rows.
func TestAliasTableIsCaughtByTheLoadCheckNotTheWriteGuard(t *testing.T) {
	for _, row := range aliasTable {
		t.Run(row.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			first, second := row.setup(t, dir)
			err := refuseCollidingOutputs(config{output: first, sarifOutput: second})
			if err == nil {
				t.Fatalf("the load check let %q and %q through", first, second)
			}
			if !strings.Contains(err.Error(), "-o") || !strings.Contains(err.Error(), "--sarif-output") {
				t.Errorf("refusal does not name both flags: %v", err)
			}
		})
	}
}

// TestMissingDirectoryThenDotDotIsUnwritable proves that report.json and
// missing/../report.json are not one file while the directory missing is
// absent. The kernel cannot traverse that directory, so the second path is
// unwritable: exit 2, and the error names the path. The mutated field is
// that directory. The absent subtest leaves it absent. Collapsing ".."
// lexically after the missing component makes this subtest exit 7.
//
// The present subtest changes only that directory into one that exists.
// missing/../report.json is then report.json, so the load check refuses
// with exit 7 and writes nothing.
func TestMissingDirectoryThenDotDotIsUnwritable(t *testing.T) {
	const (
		jsonPath  = "report.json"
		sarifPath = "missing/../report.json"
	)
	args := []string{"scan", "--json-output", jsonPath, "--sarif-output", sarifPath}

	t.Run("absent", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		if _, err := os.Stat(filepath.Join(dir, "missing")); !os.IsNotExist(err) {
			t.Fatalf("missing is present: %v", err)
		}
		rev := &captureReviewer{}
		var stdout, stderr bytes.Buffer
		code := Run(args, emptyEnv, &stdout, &stderr, openCapture(rev))
		if code != exitFatal {
			t.Fatalf("exit %d, want %d; stderr %s", code, exitFatal, stderr.String())
		}
		if strings.Contains(stderr.String(), "same file") {
			t.Fatalf("the load check refused a path the filesystem cannot reach: %s", stderr.String())
		}
		if !strings.Contains(stderr.String(), sarifPath) {
			t.Fatalf("error does not name the unwritable path: %s", stderr.String())
		}
		if rev.req.Target == "" {
			t.Fatal("the review did not run")
		}
		if _, err := os.Stat(filepath.Join(dir, jsonPath)); err != nil {
			t.Fatalf("the reachable output was not written: %v", err)
		}
	})

	t.Run("present", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		mkdir(t, filepath.Join(dir, "missing"))
		rev := &captureReviewer{}
		var stdout, stderr bytes.Buffer
		code := Run(args, emptyEnv, &stdout, &stderr, openCapture(rev))
		if code != exitConfig {
			t.Fatalf("exit %d, want %d; stderr %s", code, exitConfig, stderr.String())
		}
		for _, want := range []string{"--json-output", "--sarif-output"} {
			if !strings.Contains(stderr.String(), want) {
				t.Errorf("refusal does not name %q: %s", want, stderr.String())
			}
		}
		if rev.req.Target != "" {
			t.Error("the review ran despite the refusal")
		}
		if _, err := os.Stat(filepath.Join(dir, jsonPath)); err == nil {
			t.Error("a file was written despite the refusal")
		}
	})
}
