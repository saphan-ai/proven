package change

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// processDefault names every exported function of this package that reads the process-wide git
// policy (ProcessPolicy and the setting it reads, UseOffline and UseLimits, Bounds, and the
// package-level wrappers that call ProcessPolicy). The process policy is always public: a
// production caller of any of them would run git under it, whatever the review it belongs to was
// opened with.
var processDefault = []string{
	"ProcessPolicy", "UseOffline", "UseLimits", "Bounds", "GitCommand",
	"Read", "WorkTree", "TopLevel", "Between", "StatusBetween", "Blob", "Messages",
	"ResolveCommit", "ExportCommit", "ExportWorktree", "FactsBetween",
}

// TestNoProductionCodeReachesTheProcessGitPolicy is the census of the process-wide git policy:
// no non-test Go file outside this package may name a function of processDefault. A production
// caller holds a Policy (NewPolicy, or the one a review put in its context) and calls its method.
// Tests may use the process default; they are the callers it exists for.
func TestNoProductionCodeReachesTheProcessGitPolicy(t *testing.T) {
	root := moduleRoot(t)
	own := filepath.Join(root, "internal", "change")
	var seen int
	var found []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path == own || (path != root && (strings.HasPrefix(name, ".") || name == "testdata" || name == "vendor")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		seen++
		rel, _ := filepath.Rel(root, path)
		found = append(found, reaches(t, path, rel)...)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if seen < 100 {
		t.Fatalf("the census read %d production files under %s; the walk is not seeing the module", seen, root)
	}
	if len(found) > 0 {
		t.Fatalf("production code reaches the process-wide git policy (use a change.Policy):\n%s", strings.Join(found, "\n"))
	}
}

// reaches lists the places in one file that name a function of processDefault through an import of
// this package. A dot import or a blank name for the package is listed too, because the census
// could not read what it calls.
func reaches(t *testing.T, path, rel string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("%s: %v", rel, err)
	}
	var out []string
	local := map[string]bool{}
	for _, imp := range f.Imports {
		p, _ := strconv.Unquote(imp.Path.Value)
		if p != "github.com/saphan-ai/proven/internal/change" {
			continue
		}
		name := "change"
		if imp.Name != nil {
			name = imp.Name.Name
		}
		if name == "." {
			out = append(out, rel+": dot-imports internal/change, so the census cannot read what it calls")
			continue
		}
		local[name] = true
	}
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && local[id.Name] && slices.Contains(processDefault, sel.Sel.Name) {
			out = append(out, rel+":"+strconv.Itoa(fset.Position(sel.Pos()).Line)+": "+id.Name+"."+sel.Sel.Name)
		}
		return true
	})
	return out
}

// moduleRoot is the directory that holds go.mod, found upward from this package.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		up := filepath.Dir(dir)
		if up == dir {
			t.Fatal("no go.mod above internal/change")
		}
		dir = up
	}
}

// A context that carries no policy gets the closed one: offline with the default limits, whatever
// UseOffline and UseLimits hold. This is the other way production code could reach the process
// default.
func TestPolicyFromAContextWithNoPolicyIsClosed(t *testing.T) {
	defer UseOffline(UseOffline(false))
	defer UseLimits(UseLimits(Limits{File: 7}))
	got := PolicyFrom(context.Background())
	if !got.Offline() {
		t.Fatal("a context with no policy gets a policy that can reach a remote")
	}
	if got.Bounds() != DefaultLimits() {
		t.Fatalf("a context with no policy reads the process-wide limits: %+v", got.Bounds())
	}
}
