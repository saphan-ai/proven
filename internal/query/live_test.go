package query_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/service"
)

// TestLoadRefusesABadSettingAndDoesNotPrintTheKey proves a bad organisation
// id and a missing key are refused by name, and neither refusal prints the
// key.
func TestLoadRefusesABadSettingAndDoesNotPrintTheKey(t *testing.T) {
	const secret = "super-secret-zetaris-key"
	// Mutated field: ZETARIS_ORG_ID.
	_, err := query.Load(func(k string) string {
		switch k {
		case query.EnvRESTURL:
			return "http://127.0.0.1:9/api/v1.0"
		case query.EnvAPIKey:
			return secret
		case query.EnvOrgID:
			return "nope"
		default:
			return ""
		}
	})
	var cfg *query.ConfigError
	if err == nil || !asConfig(err, &cfg) || cfg.Field != query.EnvOrgID || cfg.Value != "nope" {
		t.Fatal(err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("the refusal printed the API key")
	}
	// Mutated field: ZETARIS_API_KEY, set to empty. The service is down, and the
	// error does not invent a value.
	_, err = query.Load(func(k string) string {
		switch k {
		case query.EnvRESTURL:
			return "http://127.0.0.1:9/api/v1.0"
		case query.EnvOrgID:
			return "1"
		default:
			return ""
		}
	})
	d, ok := service.As(err)
	if !ok || d.Service != "zetaris" || !strings.Contains(d.Reason, query.EnvAPIKey) {
		t.Fatal(err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("the error printed the API key")
	}
	// Mutated field: ZETARIS_API_KEY, set to a non-empty value with a line break.
	_, err = query.Load(func(k string) string {
		switch k {
		case query.EnvRESTURL:
			return "http://127.0.0.1:9/api/v1.0"
		case query.EnvAPIKey:
			return "not-a-secret\n"
		case query.EnvOrgID:
			return "1"
		default:
			return ""
		}
	})
	if err == nil || !asConfig(err, &cfg) || cfg.Field != query.EnvAPIKey || cfg.Value != "" {
		t.Fatal(err)
	}
	if strings.Contains(err.Error(), "not-a-secret") || strings.Contains(err.Error(), secret) {
		t.Fatal("the refusal printed a key")
	}
}

// TestLiveSelectReadsRowsFromTheSQLEditor proves the client posts to the
// SQL-editor route with the bearer key, the organisation id, and a new
// request id and query id for each call, and turns JSON cells into text.
func TestLiveSelectReadsRowsFromTheSQLEditor(t *testing.T) {
	const secret = "live-test-key"
	var ids []string
	var bodies []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sql-editor/sqls/run" {
			t.Errorf("path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer "+secret {
			t.Errorf("authorization missing the bearer token")
		}
		if r.Header.Get("X-Org-ID") != "1" {
			t.Errorf("org %q", r.Header.Get("X-Org-ID"))
		}
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			t.Error("missing request id")
		}
		ids = append(ids, id)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		bodies = append(bodies, body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"headers":["cve_id","epss"],"data":[["CVE-2021-44228",0.99999]],"total":1,"timeUsed":10}`)
	}))
	defer srv.Close()
	cfg := loadLive(t, srv.URL, secret)
	live := query.NewLive(cfg, nil)
	sql := "SELECT cve_id, epss FROM px_evidence.px_epss_cve_rows WHERE cve_id = 'CVE-2021-44228'"
	res, err := live.Query(context.Background(), sql)
	if err != nil {
		t.Fatal(err)
	}
	if res.Engine != "zetaris" || res.SQL != sql || len(res.Columns) != 2 || res.Columns[0] != "cve_id" {
		t.Fatalf("%#v", res)
	}
	if len(res.Rows) != 1 || res.Rows[0][0] != "CVE-2021-44228" || res.Rows[0][1] != "0.99999" {
		t.Fatalf("rows %#v", res.Rows)
	}
	if _, err := time.Parse(time.RFC3339, res.Snapshot); err != nil {
		t.Fatalf("snapshot %q", res.Snapshot)
	}
	if _, err := live.Query(context.Background(), sql); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] == ids[1] {
		t.Fatalf("request ids %v", ids)
	}
	if len(bodies) != 2 {
		t.Fatalf("bodies %d", len(bodies))
	}
	if bodies[0]["source"] != "SqlEditor" || bodies[0]["sql"] != sql || bodies[0]["limit"] != float64(100) {
		t.Fatalf("body %#v", bodies[0])
	}
	q1, _ := bodies[0]["queryId"].(string)
	q2, _ := bodies[1]["queryId"].(string)
	if q1 == "" || q1 == q2 || q1 == ids[0] {
		t.Fatalf("query ids %q %q request %q", q1, q2, ids[0])
	}
	if strings.Contains(res.SQL, secret) {
		t.Fatal("the result carried the API key")
	}
}

// TestUnauthorizedNamesZetarisAndHidesTheKey proves a 401 is a failed zetaris
// service, and a body that echoes the key is not printed.
// Mutated field: the HTTP status, set to 401.
func TestUnauthorizedNamesZetarisAndHidesTheKey(t *testing.T) {
	const secret = "super-secret-zetaris-key"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"msg":"`+secret+`"}`)
	}))
	defer srv.Close()
	live := query.NewLive(loadLive(t, srv.URL, secret), nil)
	_, err := live.Query(context.Background(), "SELECT cve_id FROM px_evidence.px_kev_vulns")
	d, ok := service.As(err)
	if !ok || d.Service != "zetaris" || d.Reason != "unauthorized" {
		t.Fatal(err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("the refusal printed the API key")
	}
}

// TestClosedEndpointNamesZetarisUnreachable proves a refused connection is
// named zetaris, unreachable.
// Mutated field: the listening socket. It is closed before the query.
func TestClosedEndpointNamesZetarisUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the closed endpoint accepted a call")
	}))
	url := srv.URL
	srv.Close()
	cfg := loadLive(t, url, "k")
	cfg.Timeout = time.Second
	live := query.NewLive(cfg, nil)
	_, err := live.Query(context.Background(), "SELECT cve_id FROM px_evidence.px_kev_vulns")
	d, ok := service.As(err)
	if !ok || d.Service != "zetaris" || d.Reason != "unreachable" {
		t.Fatal(err)
	}
}

// TestNonSelectIsNotSent proves a statement that is not a SELECT is refused
// before any request.
// Mutated field: the SQL verb. It is not a SELECT, so the client does not send it.
func TestNonSelectIsNotSent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("a non-SELECT was sent")
	}))
	defer srv.Close()
	live := query.NewLive(loadLive(t, srv.URL, "k"), nil)
	_, err := live.Query(context.Background(), "DELETE FROM px_evidence.px_kev_vulns")
	d, ok := service.As(err)
	if !ok || d.Service != "zetaris" || d.Reason != "not a select" {
		t.Fatal(err)
	}
}

// loadLive builds valid Settings that point at a test server.
func loadLive(t *testing.T, base, key string) query.Settings {
	t.Helper()
	cfg, err := query.Load(func(k string) string {
		switch k {
		case query.EnvRESTURL:
			return base
		case query.EnvAPIKey:
			return key
		case query.EnvOrgID:
			return "1"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

// asConfig reports whether err is a *query.ConfigError and stores it in
// target.
func asConfig(err error, target **query.ConfigError) bool {
	cfg, ok := err.(*query.ConfigError)
	if ok {
		*target = cfg
	}
	return ok
}
