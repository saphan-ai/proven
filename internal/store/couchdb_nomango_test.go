package store

import (
	"context"
	"regexp"
	"testing"

	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
	"github.com/saphan-ai/proven/internal/source"
)

// sqlOnly is a source added through the adapter interface whose template has
// the sql form only, as every registered template does today (ADR-0046).
type sqlOnly struct{}

func (sqlOnly) ID() string { return "sql-only" }

func (sqlOnly) Load() source.Load {
	return source.Load{Origin: "a test list", Refresh: "never", Table: "sql_only", Snapshot: "2026-10-25"}
}

func (sqlOnly) Templates() []source.Spec {
	return []source.Spec{{
		Template: source.Template{
			Name:   "sql-only-by-name",
			Answer: "the row of one name",
			Params: []source.Param{{Name: "name", Format: "letters", Match: regexp.MustCompile(`^[a-z]{1,20}$`)}},
		},
		Build: func(params map[string]string, _ source.Rel) (string, error) {
			return "SELECT name FROM sql_only WHERE name = '" + params["name"] + "'", nil
		},
	}}
}

func (sqlOnly) Rule(in source.Evidence) (source.Evidence, bool) { return in, len(in.Rows) > 0 }

// TestCouchDBNamesItselfDownForATemplateWithNoMangoForm proves a template with
// no Mango form is never sent to CouchDB: the store is named down with that
// reason. The mutated line is the empty-form check in CouchDB.Run; without it
// the request goes to the server and the reason changes.
func TestCouchDBNamesItselfDownForATemplateWithNoMangoForm(t *testing.T) {
	if err := source.Register(sqlOnly{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { source.Unregister(sqlOnly{}.ID()) })
	_, srv := newFakeCouch(t)
	req := querytpl.Request{Template: "sql-only-by-name", Params: map[string]string{"name": "left"}}
	_, err := couchStore(srv).Run(context.Background(), req, querytpl.Relations{})
	d, ok := service.As(err)
	if !ok || d.Service != "couchdb" || d.Reason != "no mango form for sql-only-by-name" {
		t.Fatalf("want couchdb named down for a template with no mango form, got %v", err)
	}
}
