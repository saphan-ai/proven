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
	_, err = local.Run(context.Background(), querytpl.Request{
		Template: querytpl.KEVByCVE,
		Params:   map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"},
	}, querytpl.Relations{})
	d, ok := service.As(err)
	if !ok {
		t.Fatalf("want a down source, got %v", err)
	}
	if d.Service != "local" || !strings.HasPrefix(d.Reason, "no table ") {
		t.Fatalf("%#v", d)
	}
}

// TestStoresAgreeOnAMissingTable proves a table the source never registered is
// a named down source on every store, never an empty answer and never an
// error that ends the review (ADR-0047). Zetaris is a recorded server that
// answers 404 for the missing view. The mutated field is the table: neither
// store holds it. Negative control: the local down conversion removed turns
// the local half red.
func TestStoresAgreeOnAMissingTable(t *testing.T) {
	req := querytpl.Request{
		Template: querytpl.KEVByCVE,
		Params:   map[string]string{querytpl.ParamCVEID: "CVE-2021-44228"},
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
	stores := map[string]Store{"local": local, "zetaris": z}
	for name, s := range stores {
		_, err := s.Run(ctx, req, querytpl.Relations{})
		d, ok := service.As(err)
		if !ok || d.Service != name {
			t.Fatalf("%s: want a down source naming it, got %v", name, err)
		}
	}
}
