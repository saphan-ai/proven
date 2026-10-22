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
			if p == "os/exec" && rel == "internal/testrun/container.go" {
				continue
			}
			if banned[p] {
				t.Errorf("%s imports %s", rel, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
