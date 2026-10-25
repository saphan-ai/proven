package store

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/service"
)

// TestLocalMissingTableIsADownSource proves a query on an empty local store
// is a named down source and not a plain error, so a review degrades and
// names `local` instead of ending (ADR-0047, ADR-0005). The mutated field is
// the table: the store holds none. Negative control: removing the down
// conversion in the file's query makes service.As fail.
func TestLocalMissingTableIsADownSource(t *testing.T) {
	local, err := OpenLocal(filepath.Join(t.TempDir(), "evidence.sql"))
	if err != nil {
		t.Fatal(err)
	}
	req := querytpl.Request{
		Template: querytpl.KEVByCVE,
		Params:   map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"},
	}
	for name, run := range map[string]func() error{
		"Run": func() error {
			_, err := local.Run(context.Background(), req, querytpl.Relations{})
			return err
		},
		"Query": func() error {
			forms, err := querytpl.BuildForms(req, querytpl.Relations{})
			if err != nil {
				return err
			}
			_, err = local.Query(context.Background(), forms.SQL)
			return err
		},
	} {
		err := run()
		d, ok := service.As(err)
		if !ok {
			t.Fatalf("%s: want a down source, got %v", name, err)
		}
		if d.Service != "local" || !strings.HasPrefix(d.Reason, "no table ") {
			t.Fatalf("%s: %#v", name, d)
		}
	}
}

// TestStoresAgreeOnAMissingTable proves a table the source never registered is
// a named down source on every store, never an empty answer and never an
// error that ends the review (ADR-0047). Zetaris is a recorded server that
// answers 404 for the missing view, and the down source is named zetaris.
// CouchDB is a recorded server with no database, and the reason starts with
// "no table", as it does on the local store. The requests are kev-by-cve and
// mal-by-package, so a new source table is the same kind of miss. The mutated
// field is the table: no store holds it. Negative control: the local down
// conversion removed turns the local half red.
func TestStoresAgreeOnAMissingTable(t *testing.T) {
	reqs := []querytpl.Request{
		{Template: querytpl.KEVByCVE, Params: map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"}},
		{Template: querytpl.MalByPackage, Params: map[string]string{querytpl.ParamEcosystem: "Go", querytpl.ParamPackage: "evil.example/drop"}},
	}
	ctx := context.Background()
	local, err := OpenLocal(filepath.Join(t.TempDir(), "evidence.sql"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()
	z := NewZetaris(query.Settings{
		BaseURL: srv.URL, APIKey: "conform-key", OrgID: "1",
		RowLimit: 100, Timeout: time.Second,
	})
	_, couchSrv := newFakeCouch(t)
	stores := map[string]Store{"local": local, "zetaris": z, "couchdb": couchStore(couchSrv)}
	for name, s := range stores {
		for _, req := range reqs {
			_, err := s.Run(ctx, req, querytpl.Relations{})
			d, ok := service.As(err)
			if !ok || d.Service != name {
				t.Fatalf("%s %s: want a down source naming it, got %v", name, req.Template, err)
			}
			// Local and CouchDB say which table is missing. Zetaris answers
			// 404 for the view and the reason is that status; the service
			// name is still zetaris.
			if name != "zetaris" && !strings.HasPrefix(d.Reason, "no table ") {
				t.Fatalf("%s %s: reason %q", name, req.Template, d.Reason)
			}
		}
	}
}
