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
// state, as a call or as a value, and only NewPolicy and ProcessPolicy build a Policy literal.
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
		found = append(found, censusFile(fset, f, name)...)
	}
	if seen < 5 {
		t.Fatalf("the census read %d production files in internal/change; it is not seeing the package", seen)
	}
	if len(found) > 0 {
		t.Fatalf("code in internal/change reaches the process-wide git policy (take the policy from the receiver):\n%s", strings.Join(found, "\n"))
	}
}

// censusFile is the in-package census over one parsed file.
func censusFile(fset *token.FileSet, f *ast.File, name string) []string {
	var found []string
	// inspect lists, in one scope, every name of the process default used as a call or as a
	// value (an alias such as `run := GitCommand` included), every read of the process-wide
	// state and every Policy literal. A name after a dot (`p.GitCommand`, a method or a field)
	// is not the package-level function and is skipped.
	inspect := func(scope string, root ast.Node, mayBuild bool) {
		// skip holds the names that are not a package-level name: a method or field after a dot,
		// and a key of a composite literal (`offline: offline` keys the Policy field).
		skip := map[*ast.Ident]bool{}
		ast.Inspect(root, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.SelectorExpr:
				skip[x.Sel] = true
			case *ast.KeyValueExpr:
				if id, ok := x.Key.(*ast.Ident); ok {
					skip[id] = true
				}
			}
			return true
		})
		ast.Inspect(root, func(n ast.Node) bool {
			at := func(what string) {
				found = append(found, filepath.Join("internal/change", name)+":"+strconv.Itoa(fset.Position(n.Pos()).Line)+": "+scope+" "+what)
			}
			switch x := n.(type) {
			case *ast.Ident:
				if skip[x] {
					return true
				}
				// The state is named wherever it is used: read (offline.Load()), or taken as a
				// value (&offline) that is read later. No parameter or local of this package
				// carries its name.
				if x.Name == "offline" || x.Name == "limitsInUse" {
					at("uses the process-wide state " + x.Name)
				} else if slices.Contains(processDefault, x.Name) || x.Name == "processBounds" {
					at("names " + x.Name)
				}
			case *ast.CompositeLit:
				if id, ok := x.Type.(*ast.Ident); ok && id.Name == "Policy" && !mayBuild {
					at("builds a Policy literal")
				}
			}
			return true
		})
	}
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Body == nil || processOwner(d) {
				continue
			}
			inspect(d.Name.Name, d.Body, d.Name.Name == "NewPolicy" || d.Name.Name == "ProcessPolicy")
		case *ast.GenDecl:
			// A package-level declaration may hold an alias too (`var run = GitCommand`,
			// `var o = &offline`); only its values are read, so the declarations of the state
			// itself (`var offline atomic.Bool`) are not named.
			for _, spec := range d.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, v := range vs.Values {
					inspect("package-level value", v, false)
				}
			}
		}
	}
	return found
}
