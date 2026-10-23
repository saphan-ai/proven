package osv

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/saphan-ai/proven/internal/query"
)

// recorder is a query adapter that keeps every statement it is asked to run and
// answers each with no rows.
type recorder struct {
	sql []string
}

func (r *recorder) Query(_ context.Context, sql string) (query.Result, error) {
	r.sql = append(r.sql, sql)
	return query.Result{Engine: "mem", Columns: Columns, Rows: [][]string{}}, nil
}

// TestCheckSendsNoQueryForAPackageOutsideItsFormat proves ADR-0032 point 2 and
// 3 for the dependency check: a package name taken from the change that holds
// a quote, a backslash, a control byte or invalid UTF-8 is refused before any
// query is built, the refusal is a trace note with no raw control byte in it,
// and the packages beside it are still queried. The mutated field is the
// package name in the change's package.json.
func TestCheckSendsNoQueryForAPackageOutsideItsFormat(t *testing.T) {
	for name, bad := range map[string]string{
		"quote":         `bad'name`,
		"backslash":     `bad\\name`,
		"escape byte":   `bad\u001bname`,
		"newline":       `bad\nname`,
		"invalid utf-8": "bad\xffname",
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			writePair(t, root, "package.json", `{"dependencies":{}}`,
				"{\"dependencies\":{\""+bad+"\":\"1.0.0\",\"good-name\":\"1.0.0\"}}")
			rec := &recorder{}
			out, err := Check(context.Background(), rec, root, "", "snap")
			if err != nil {
				t.Fatal(err)
			}
			if len(rec.sql) != 1 || !strings.HasSuffix(rec.sql[0], "package = 'good-name'") {
				t.Fatalf("queries sent: %q", rec.sql)
			}
			refused := false
			for _, note := range out.Notes {
				if !utf8.ValidString(note) || strings.ContainsAny(note, "\x00\x1b\n\r\t") {
					t.Fatalf("a note carries a raw control byte or invalid UTF-8: %q", note)
				}
				if strings.Contains(note, "osv query refused: query parameter value is outside its format") {
					refused = true
				}
			}
			if !refused {
				t.Fatalf("no refusal note in:\n%s", strings.Join(out.Notes, "\n"))
			}
		})
	}
}
