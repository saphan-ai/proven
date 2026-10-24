package source_test

import (
	"context"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/source"
	"github.com/saphan-ai/proven/internal/store"
)

// bannedPackage is a source added only through the adapter interface.
// It is not one of the built-in sources. The sample table is banned_pkg.
type bannedPackage struct{}

func (bannedPackage) ID() string { return "banned-package" }

func (bannedPackage) Load() source.Load {
	return source.Load{
		Origin:   "the organisation's own list",
		Refresh:  "when the organisation updates the list",
		Table:    "banned_pkg",
		Snapshot: "2026-10-24",
	}
}

func (bannedPackage) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{
			Name:   "banned-by-package",
			Answer: "the banned-package row of one package",
			Params: []source.Param{{
				Name:   "package",
				Format: "a package name of letters, digits, dots, underscores and hyphens",
				Match:  regexp.MustCompile(`^[A-Za-z0-9._-]{1,200}$`),
			}},
		},
		// The value sits between quotes. Proven has already held it to the
		// format above and to the byte floor, so it cannot end the literal
		// (ADR-0032). The builder may still refuse for its own reasons, and
		// it refuses with a *querytpl.Refusal. The table name is the
		// adapter's own, not a review setting.
		Build: func(params map[string]string, _ source.Rel) (string, error) {
			pkg := params["package"]
			if pkg == "." || pkg == ".." {
				return "", &querytpl.Refusal{Reason: querytpl.ReasonBadValue, Detail: "banned-by-package package"}
			}
			return "SELECT name, reason FROM banned_pkg WHERE name = '" + pkg + "'", nil
		},
	}}
}

func (bannedPackage) Rule(in source.Evidence) (source.Evidence, bool) {
	// A row means the package is on the banned list. The evidence is the
	// query and that row. No row means the package is not on the list, and
	// that is not a finding (ADR-0046).
	if len(in.Rows) == 0 {
		return source.Evidence{}, false
	}
	return in, true
}

// TestAddedSourceShowsItsTemplateItsQueryAndItsRule proves a source added
// only through the interface: its template is in the catalogue, its query
// runs on the local store, and its rule returns the row as evidence and
// returns nothing when the query has no row. The mutated field is the
// adapter in the added list below. Remove bannedPackage{} from that line
// and this test goes red.
func TestAddedSourceShowsItsTemplateItsQueryAndItsRule(t *testing.T) {
	// mutated line: the bannedPackage element of this list.
	added := []source.Adapter{bannedPackage{}}
	for _, a := range added {
		if err := source.Register(a); err != nil {
			t.Fatal(err)
		}
		id := a.ID()
		t.Cleanup(func() { source.Unregister(id) })
	}

	var listed bool
	for _, tpl := range querytpl.Catalogue() {
		if tpl.Name == "banned-by-package" {
			listed = true
		}
	}
	if !listed {
		t.Fatal("catalogue has no banned-by-package template; the adapter is not in the registry")
	}

	sample := bannedPackage{}
	load := sample.Load()
	if load.Table != "banned_pkg" || load.Snapshot != "2026-10-24" || load.Origin == "" {
		t.Fatalf("load: %+v", load)
	}

	local, err := store.OpenLocal(filepath.Join(t.TempDir(), "evidence.sql"))
	if err != nil {
		t.Fatal(err)
	}
	local.SetClock(func() time.Time {
		return time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)
	})
	err = local.Register(context.Background(), store.Table{
		Name:     "banned_pkg",
		Columns:  []string{"name", "reason"},
		Rows:     [][]string{{"left-pad", "not allowed"}},
		Snapshot: load.Snapshot,
	})
	if err != nil {
		t.Fatal(err)
	}

	hit, err := local.Run(context.Background(), querytpl.Request{
		Template: "banned-by-package",
		Params:   map[string]string{"package": "left-pad"},
	}, querytpl.Relations{})
	if err != nil {
		t.Fatal(err)
	}
	ev, ok := sample.Rule(source.Evidence{
		SQL: hit.Query, Form: hit.Form, Engine: hit.Engine, Snapshot: hit.Snapshot,
		Columns: hit.Columns, Rows: hit.Rows,
	})
	if !ok || ev.SQL != "SELECT name, reason FROM banned_pkg WHERE name = 'left-pad'" ||
		ev.Snapshot != "2026-10-24T00:00:00Z" || len(ev.Rows) != 1 || ev.Rows[0][0] != "left-pad" || ev.Rows[0][1] != "not allowed" {
		t.Fatalf("evidence ok=%v %+v", ok, ev)
	}

	miss, err := local.Run(context.Background(), querytpl.Request{
		Template: "banned-by-package",
		Params:   map[string]string{"package": "right-pad"},
	}, querytpl.Relations{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := sample.Rule(source.Evidence{SQL: miss.Query, Rows: miss.Rows}); ok || len(miss.Rows) != 0 {
		t.Fatalf("a package that is not banned produced evidence: %+v", miss.Rows)
	}
}
