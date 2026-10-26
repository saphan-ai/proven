package source_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/source"
)

// concatBuilder is a registered source whose builder places the parameter
// in the query with no check of its own. The format admits every text, so
// the only refusal is the one source.Build applies before it calls the
// builder (ADR-0032).
type concatBuilder struct {
	calls int
}

func (concatBuilder) ID() string { return "concat-build" }

func (concatBuilder) Load() source.Load {
	return source.Load{Origin: "test", Refresh: "never", Table: "t"}
}

func (b *concatBuilder) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{
			Name:   "concat-build-by-name",
			Answer: "a row",
			Params: []source.Param{{
				Name:   "n",
				Format: "anything",
				Match:  regexp.MustCompile(`(?s)^.*$`),
			}},
		},
		Build: func(p map[string]string, _ source.Rel) (string, error) {
			b.calls++
			return "SELECT '" + p["n"] + "'", nil
		},
	}}
}

func (concatBuilder) Rule(in source.Evidence) (source.Evidence, bool) {
	return in, len(in.Rows) > 0
}

// TestBuildRefusesAValueBeforeTheBuilderRuns proves source.Build returns
// *source.ValueError for a parameter value the byte floor rejects, and does
// not call a builder that would have concatenated that value into the query.
// The mutated field is the parameter value. Negative control: deleting the
// CheckValue loop in source.Build lets the quote through and this test goes red.
func TestBuildRefusesAValueBeforeTheBuilderRuns(t *testing.T) {
	b := &concatBuilder{}
	if err := source.Register(b); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { source.Unregister("concat-build") })

	_, ok, err := source.Build("concat-build-by-name", map[string]string{"n": "x' OR '1'='1"}, source.Rel{})
	var refused *source.ValueError
	if !ok || !errors.As(err, &refused) || refused.Template != "concat-build-by-name" || refused.Param != "n" {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if b.calls != 0 {
		t.Fatalf("the builder ran %d times with a refused value", b.calls)
	}
	if strings.Contains(err.Error(), "OR") {
		t.Fatalf("the refusal repeated the value: %v", err)
	}

	sql, ok, err := source.Build("concat-build-by-name", map[string]string{"n": "left-pad"}, source.Rel{})
	if !ok || err != nil || sql != "SELECT 'left-pad'" || b.calls != 1 {
		t.Fatalf("a plain value: ok=%v sql=%q err=%v calls=%d", ok, sql, err, b.calls)
	}
}
