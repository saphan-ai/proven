package guardrail

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Property: the down set is closed over this package's own refusals. Every
// Refuse call with a literal reason in the package's code maps to a named down
// value, not to DownOther, so a reason added later without its value turns
// this test red instead of landing on the trace as "other".
func TestEveryLiteralRefusalHasADownValue(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	seen := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				return true
			}
			if id, ok := call.Fun.(*ast.Ident); !ok || id.Name != "Refuse" {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			reason, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatal(err)
			}
			seen++
			if downCode(reason) == DownOther {
				t.Errorf("%s: Refuse(%q) has no down value", fset.Position(call.Pos()), reason)
			}
			return true
		})
	}
	// The census must have found the refusals it guards, or it proves nothing.
	if seen < 10 {
		t.Fatalf("found %d literal refusals; the census did not read the package", seen)
	}
}

// Property: the refusals not made from a literal also have their values: the
// empty key Load reports, and each HTTP status httpReason names; a status
// outside 100 to 599, or any other text, is DownOther.
func TestComputedRefusalsHaveDownValues(t *testing.T) {
	cases := map[string]string{
		EnvKey + " is empty": DownNoKey,
		httpReason(503):      "http-503",
		httpReason(404):      "http-404",
		httpReason(600):      DownOther,
		"http abc":           DownOther,
		"unsafe":             DownOther,
	}
	for reason, want := range cases {
		if got := downCode(reason); got != want {
			t.Errorf("downCode(%q) = %q, want %q", reason, got, want)
		}
	}
}
