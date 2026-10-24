package pipeline

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// exemptBanned is the only allow-list. Every other non-test Go file under
// internal/ and cmd/ is checked. A new package is checked because it is absent
// here. Each reason is one line.
var exemptBanned = map[string]map[string]string{
	"internal/generator/live.go": {
		"net/http": "the live generator calls the OpenAI-compatible endpoint",
	},
	"internal/testrun/container.go": {
		"os/exec": "the container runner starts the runtime that runs the change",
	},
	"internal/change/between.go": {
		"os/exec": "git read commands (rev-parse, diff, ls-files, log, archive) name changed files, read commit messages and copy tracked trees; they do not run the change",
	},
	"internal/cli/run.go": {
		"os/exec": "git merge-base names the fork point and does not run the change",
	},
	"internal/cli/output_open_unix.go": {
		"syscall": "O_NOFOLLOW on the output open, so a final symbolic link is not followed",
	},
	"internal/guardrail/live.go": {
		"net/http": "the live guardrail calls the content-safety endpoint",
	},
	"internal/query/live.go": {
		"net/http": "the live query adapter calls the Zetaris REST API",
	},
}

// No production file under internal/ or cmd/ imports a package that dials the
// network, starts a process or loads code, except the files named in
// exemptBanned, each with its reason. ADR-0025: only the container runner
// starts the change's code.
func TestInternalAndCmdDoNotImportBannedPackages(t *testing.T) {
	banned := map[string]bool{
		"net":      true,
		"net/http": true,
		"os/exec":  true,
		"syscall":  true,
		"plugin":   true,
	}
	for file, allowed := range exemptBanned {
		for path, reason := range allowed {
			if strings.TrimSpace(reason) == "" || strings.Contains(reason, "\n") {
				t.Errorf("exempt %s %s needs a one-line reason", file, path)
			}
			if !banned[path] {
				t.Errorf("exempt %s %s is not a banned import", file, path)
			}
		}
	}

	root := repoRoot(t)
	checked := 0
	used := map[string]bool{}
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
		if !strings.HasPrefix(rel, "internal/") && !strings.HasPrefix(rel, "cmd/") {
			return nil
		}
		checked++
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Errorf("parse %s: %v", rel, err)
			return nil
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if !banned[p] {
				continue
			}
			if reason := exemptBanned[rel][p]; reason != "" {
				used[rel+"\x00"+p] = true
				continue
			}
			// query/live.go is the Zetaris REST client. sample.go and the
			// local engine still must not import net or net/http.
			if rel == "internal/query/live.go" && p == "net/http" {
				continue
			}
			t.Errorf("%s imports %s", rel, p)
		}
		if name := osImportName(f); name != "" {
			ast.Inspect(f, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				id, ok := sel.X.(*ast.Ident)
				if ok && id.Name == name && sel.Sel != nil && sel.Sel.Name == "StartProcess" {
					t.Errorf("%s calls os.StartProcess", rel)
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatal("no non-test Go file under internal/ or cmd/ was checked")
	}
	for file, allowed := range exemptBanned {
		for path := range allowed {
			if !used[file+"\x00"+path] {
				t.Errorf("exempt %s import %s is unused", file, path)
			}
		}
	}
}

// osImportName is the local name of package os. A blank or dot import has
// no os.StartProcess selector, so it returns empty.
func osImportName(f *ast.File) string {
	for _, imp := range f.Imports {
		if strings.Trim(imp.Path.Value, `"`) != "os" {
			continue
		}
		if imp.Name == nil {
			return "os"
		}
		if imp.Name.Name == "_" || imp.Name.Name == "." {
			return ""
		}
		return imp.Name.Name
	}
	return ""
}
