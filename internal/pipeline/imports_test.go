package pipeline

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSamplePackagesDoNotImportNetOrExec(t *testing.T) {
	root := repoRoot(t)
	banned := map[string]bool{"net": true, "net/http": true, "os/exec": true}
	dirs := []string{
		"cmd/proven",
		"internal/agent",
		"internal/candidate",
		"internal/change",
		"internal/fixture",
		"internal/generator",
		"internal/limit",
		"internal/pipeline",
		"internal/query",
		"internal/review",
		"internal/service",
		"internal/testrun",
		"internal/trace",
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "internal/netguard/") {
			return nil
		}
		keep := false
		for _, d := range dirs {
			if rel == d || strings.HasPrefix(rel, d+"/") {
				keep = true
			}
		}
		if !keep {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("parse %s: %v", rel, err)
			return nil
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if !banned[p] {
				continue
			}
			// container.go starts the container runtime that runs the change's
			// tests (ADR-0025); it is the one file allowed to run a program.
			if rel == "internal/testrun/container.go" && p == "os/exec" {
				continue
			}
			// between.go runs git read commands (rev-parse, diff, ls-files, log) to
			// name changed files and read commit messages. It does not run the change.
			// Any other file that imports os/exec or net still fails this test.
			if rel == "internal/change/between.go" && p == "os/exec" {
				continue
			}
			// live.go is the Nightshift client. sample.go still must not
			// import net or net/http; this exemption is that one file.
			if rel == "internal/generator/live.go" && p == "net/http" {
				continue
			}
			t.Errorf("%s imports %s", rel, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
