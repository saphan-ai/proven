package change

import (
	"go/parser"
	"go/token"
	"testing"
)

// TestCensusSeesAnAliasOfTheProcessDefault: a Policy method that takes a process-default function as
// a value (`run := GitCommand`) runs git under the process default as surely as one that calls it,
// so the in-package census names it. The census reads files from disk; this test hands it the same
// AST check on a source string. Mutated field: the alias in WorkTree's body.
func TestCensusSeesAnAliasOfTheProcessDefault(t *testing.T) {
	for _, src := range []string{
		"package change\nfunc (p Policy) WorkTree(t string) { run := GitCommand; _ = run(\"-C\", t) }\n",
		"package change\nvar run = GitCommand\n",
	} {
		if got := censusOf(t, src); len(got) == 0 {
			t.Errorf("the census accepted an alias of the process default:\n%s", src)
		}
	}
	if got := censusOf(t, "package change\nfunc (p Policy) WorkTree(t string) { _ = p.GitCommand(\"-C\", t) }\n"); len(got) != 0 {
		t.Errorf("control: a call through the receiver was named: %v", got)
	}
}

// censusOf runs the in-package census over one source file and returns what it names.
func censusOf(t *testing.T, src string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	return censusFile(fset, f, "x.go")
}
