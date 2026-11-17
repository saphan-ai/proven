package feeds

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/network"
	"github.com/saphan-ai/proven/internal/querytpl"
	"github.com/saphan-ai/proven/internal/store"
)

func TestRefreshUsesModifiedAdvisoryDeltaAndConditionalCatalogues(t *testing.T) {
	ctx := context.Background()
	modified := "2026-11-17T00:00:00Z"
	records, conditionals := 0, 0
	broken := false
	withdrawn, noMatches, wrongPackage := false, false, false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/osv/querybatch":
			if noMatches {
				fmt.Fprint(w, `{"results":[{}]}`)
				return
			}
			fmt.Fprintf(w, `{"results":[{"vulns":[{"id":"GHSA-example-one","modified":%q},{"id":"MAL-example-two","modified":%q}]}]}`, modified, modified)
		case strings.HasPrefix(r.URL.Path, "/osv/vulns/"):
			records++
			if broken {
				http.Error(w, "unavailable", 503)
				return
			}
			id := strings.TrimPrefix(r.URL.Path, "/osv/vulns/")
			if withdrawn {
				fmt.Fprintf(w, `{"id":%q,"modified":%q,"withdrawn":"2026-11-17T02:00:00Z"}`, id, modified)
				return
			}
			fmt.Fprintf(w, `{"id":%q,"modified":%q,"published":"2026-01-01T00:00:00Z","aliases":["CVE-2026-4242"],"summary":"Synthetic fixture","affected":[{"package":{"ecosystem":"npm","name":"example-package"},"ranges":[{"type":"SEMVER","events":[{"introduced":"1.0.0"},{"fixed":"2.0.0"}]}]}]}`, id, modified)
		case r.URL.Path == "/kev":
			if r.Header.Get("If-None-Match") == `"fixture-kev"` {
				conditionals++
				w.WriteHeader(304)
				return
			}
			w.Header().Set("ETag", `"fixture-kev"`)
			fmt.Fprint(w, `{"dateReleased":"2026-11-16T00:00:00Z","count":1,"vulnerabilities":[{"cveID":"CVE-2026-4242","vendorProject":"Example","product":"Example","dateAdded":"2026-11-01","knownRansomwareCampaignUse":"Unknown"}]}`)
		case r.URL.Path == "/epss/epss":
			fmt.Fprint(w, `{"status":"OK","total":1,"data":[{"cve":"CVE-2026-4242","epss":"0.42","percentile":"0.91","date":"2026-11-16"}]}`)
		case strings.Contains(r.URL.Path, "/versions/"):
			if wrongPackage {
				fmt.Fprint(w, `{"versionKey":{"system":"NPM","name":"another-package","version":"1.5.0"},"licenses":["MIT"]}`)
				return
			}
			fmt.Fprint(w, `{"versionKey":{"system":"NPM","name":"example-package","version":"1.5.0"},"licenses":["MIT"],"publishedAt":"2026-01-01T00:00:00Z","relatedProjects":[{"projectKey":{"id":"github.com/example/project"},"relationType":"SOURCE_REPO","relationProvenance":"UNVERIFIED_METADATA"}]}`)
		case strings.HasPrefix(r.URL.Path, "/deps/projects/"):
			fmt.Fprint(w, `{"projectKey":{"id":"github.com/example/project"},"scorecard":{"date":"2026-11-10T00:00:00Z","checks":[{"name":"Maintained","score":10},{"name":"Signed-Releases","score":10}]}}`)
		case strings.HasPrefix(r.URL.Path, "/deps/advisories/"):
			fmt.Fprint(w, `{"advisoryKey":{"id":"GHSA-example-one"},"aliases":["CVE-2026-4242"],"cvss3Score":8.1,"cvss3Vector":"CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:U/C:H/I:H/A:H"}`)
		default:
			t.Errorf("unexpected request %s", r.URL.String())
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	policy, err := network.Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{OSV: server.URL + "/osv", KEV: server.URL + "/kev", EPSS: server.URL + "/epss", Deps: server.URL + "/deps", Cache: t.TempDir(), Network: policy}
	db, err := store.OpenSQLite(filepath.Join(t.TempDir(), "evidence.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	packages := []Package{{"npm", "example-package", "1.5.0"}}
	for i := 0; i < 2; i++ {
		result, err := Sync(ctx, cfg, packages, nil, db.Replace)
		if err != nil || len(result.Sources) != 7 {
			t.Fatalf("%+v %v", result, err)
		}
		for _, source := range result.Sources {
			if source.Status == "error" || source.CheckedAt == "" || source.Origin == "" {
				t.Fatalf("%+v", source)
			}
		}
		res, err := db.Run(ctx, querytpl.Request{Template: querytpl.OSVByPackage, Params: map[string]string{"ecosystem": "npm", "package": "example-package"}}, querytpl.Relations{})
		if err != nil || len(res.Rows) != 2 {
			t.Fatalf("OSV table: %+v %v", res, err)
		}
		res, err = db.Run(ctx, querytpl.Request{Template: querytpl.OSVByAdvisory, Params: map[string]string{"advisory_id": "GHSA-example-one"}}, querytpl.Relations{})
		if err != nil || len(res.Rows) != 1 {
			t.Fatalf("advisory index: %+v %v", res, err)
		}
	}
	if records != 2 || conditionals != 1 {
		t.Fatalf("unchanged data downloaded again: records=%d conditional=%d", records, conditionals)
	}
	// A changed ID whose body cannot be fetched cannot reuse yesterday's
	// record as today's success. The prior SQL source is retained as-is.
	modified = "2026-11-17T01:00:00Z"
	broken = true
	result, err := Sync(ctx, cfg, packages, nil, db.Replace)
	if err == nil || result.Sources[0].Status != "error" {
		t.Fatalf("%+v %v", result, err)
	}
	res, err := db.Run(ctx, querytpl.Request{Template: querytpl.OSVByPackage, Params: map[string]string{"ecosystem": "npm", "package": "example-package"}}, querytpl.Relations{})
	if err != nil || len(res.Rows) != 2 {
		t.Fatal("failed refresh overwrote the retained table")
	}
	// Withdrawal and a subsequent empty query both reconcile the active
	// tables; a previously cached advisory must not stay active forever.
	broken, withdrawn = false, true
	_, err = Sync(ctx, cfg, packages, nil, db.Replace)
	if err != nil {
		t.Fatal(err)
	}
	res, err = db.Run(ctx, querytpl.Request{Template: querytpl.OSVByPackage, Params: map[string]string{"ecosystem": "npm", "package": "example-package"}}, querytpl.Relations{})
	if err != nil || len(res.Rows) != 0 {
		t.Fatal("withdrawn advisory remained active")
	}
	noMatches, wrongPackage = true, true
	result, err = Sync(ctx, cfg, packages, nil, db.Replace)
	if err == nil {
		t.Fatal("accepted metadata for another package")
	}
	for _, status := range result.Sources {
		if (status.ID == "spdx" || status.ID == "scorecard" || status.ID == "cvss") && status.Status != "error" {
			t.Fatalf("identity mismatch was hidden: %+v", status)
		}
	}
	res, err = db.Run(ctx, querytpl.Request{Template: querytpl.OSVByPackage, Params: map[string]string{"ecosystem": "npm", "package": "example-package"}}, querytpl.Relations{})
	if err != nil || len(res.Rows) != 0 {
		t.Fatal("cached advisory reappeared without a current match")
	}
}

func TestSourceClientBoundsAndRefusesRedirectsWithoutStaleFallback(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/secret", 302)
			return
		}
		fmt.Fprint(w, `not valid json`)
	}))
	defer server.Close()
	policy, _ := network.Load(nil)
	client := NewClient(policy, t.TempDir())
	if _, err := client.JSON(context.Background(), "GET", server.URL+"/redirect", nil); err == nil || calls != 1 {
		t.Fatal("followed redirect")
	}
	if _, err := client.JSON(context.Background(), "GET", server.URL+"/invalid", nil); err == nil {
		t.Fatal("accepted malformed source")
	}
	if _, err := Load(func(k string) string {
		if k == "PROVEN_OSV_URL" {
			return "https://user:credential@example.com/"
		}
		return ""
	}); err == nil {
		t.Fatal("accepted credentials in a reported URL")
	}
	if _, err := Load(func(k string) string {
		if k == "PROVEN_NETWORK" {
			return "local"
		}
		return ""
	}); err == nil {
		t.Fatal("local policy accepted the public source URLs")
	}
}

func TestOSVBatchRevisionAllowsOnlyItsMeasuredPrecisionLoss(t *testing.T) {
	if !sameOSVRevision("2026-10-17T02:55:38.085813Z", "2026-10-17T02:55:38.085813876Z") {
		t.Fatal("rejected the same microsecond")
	}
	if sameOSVRevision("2026-10-17T02:55:38.085813Z", "2026-10-17T02:55:38.085814Z") || sameOSVRevision("not-a-date", "not-a-date") {
		t.Fatal("accepted an unverified revision")
	}
}
