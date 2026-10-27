package change

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// processState names this package's unexported process-wide state and its reader: what the
// package-level functions read and a Policy method must not.
var processState = []string{"offline", "limitsInUse", "processBounds"}

// processOwners are the functions allowed to read the process-wide policy: the package-level
// functions of processDefault themselves (each is the process default, or its setter, or a
// wrapper over ProcessPolicy) and processBounds. Every other function of this package, every
// Policy method included, takes its policy from its receiver or its arguments.
func processOwner(fd *ast.FuncDecl) bool {
	return fd.Recv == nil && (slices.Contains(processDefault, fd.Name.Name) || fd.Name.Name == "processBounds")
}

// TestNoCodeInThisPackageReachesTheProcessGitPolicy is the census inside internal/change, which the
// census over the module skips because the process default is defined here: no non-test function
// of this package other than a processOwner names a function of processDefault or the process-wide
// state, and only NewPolicy and ProcessPolicy build a Policy literal.
func TestNoCodeInThisPackageReachesTheProcessGitPolicy(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var seen int
	var found []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		seen++
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil || processOwner(fd) {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				at := func(what string) {
					found = append(found, filepath.Join("internal/change", name)+":"+strconv.Itoa(fset.Position(n.Pos()).Line)+": "+fd.Name.Name+" "+what)
				}
				switch x := n.(type) {
				case *ast.CallExpr:
					if id, ok := x.Fun.(*ast.Ident); ok && (slices.Contains(processDefault, id.Name) || id.Name == "processBounds") {
						at("calls " + id.Name)
					}
				case *ast.SelectorExpr:
					if id, ok := x.X.(*ast.Ident); ok && (id.Name == "offline" || id.Name == "limitsInUse") {
						at("reads " + id.Name)
					}
				case *ast.CompositeLit:
					if id, ok := x.Type.(*ast.Ident); ok && id.Name == "Policy" && fd.Name.Name != "NewPolicy" && fd.Name.Name != "ProcessPolicy" {
						at("builds a Policy literal")
					}
				}
				return true
			})
		}
	}
	if seen < 5 {
		t.Fatalf("the census read %d production files in internal/change; it is not seeing the package", seen)
	}
	if len(found) > 0 {
		t.Fatalf("code in internal/change reaches the process-wide git policy (take the policy from the receiver):\n%s", strings.Join(found, "\n"))
	}
}
