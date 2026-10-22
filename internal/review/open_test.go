package review_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/review"
)

func TestOpenSampleReviewKeepsF1AndF2(t *testing.T) {
	rev := openSample(t, t.TempDir())
	res, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Partial {
		t.Fatal("partial")
	}
	if len(res.FailedServices) != 0 {
		t.Fatalf("failed services %#v", res.FailedServices)
	}
	f1, ok := find(res, "F1")
	if !ok {
		t.Fatal("F1 missing")
	}
	if len(f1.Evidence) != 1 || f1.Evidence[0].Kind != review.EvidenceTest || f1.Evidence[0].Test == nil {
		t.Fatalf("F1 evidence %#v", f1.Evidence)
	}
	if !f1.Evidence[0].Test.FailsOnChange || !f1.Evidence[0].Test.PassesOnBase {
		t.Fatalf("F1 test %#v", f1.Evidence[0].Test)
	}
	if f1.Evidence[0].Test.Name != "TestDiscounted" {
		t.Fatalf("test %s", f1.Evidence[0].Test.Name)
	}
	f2, ok := find(res, "F2")
	if !ok {
		t.Fatal("F2 missing")
	}
	if len(f2.Evidence) != 1 || f2.Evidence[0].Kind != review.EvidenceQuery || f2.Evidence[0].Query == nil {
		t.Fatalf("F2 evidence %#v", f2.Evidence)
	}
	q := f2.Evidence[0].Query
	if q.Engine != "local-sample" || len(q.Rows) != 1 || q.Rows[0][0] != "evil.example/leftpad" {
		t.Fatalf("F2 query %#v", q)
	}
	f4, ok := find(res, "F4")
	if !ok {
		t.Fatal("F4 missing")
	}
	if len(f4.Evidence) != 1 || f4.Evidence[0].Query == nil {
		t.Fatalf("F4 evidence %#v", f4.Evidence)
	}
	kq := f4.Evidence[0].Query
	if kq.Engine != "local-sample" || kq.SQL != "SELECT cve_id, vendor_project, product, date_added, known_ransomware FROM px_evidence.px_kev_vulns WHERE cve_id = 'CVE-2026-0001'" ||
		len(kq.Rows) != 1 || kq.Rows[0][0] != "CVE-2026-0001" {
		t.Fatalf("F4 query %#v", kq)
	}
	if _, ok := find(res, "F3"); ok {
		t.Fatal("F3 must be dropped")
	}
	for _, m := range res.Models {
		if m.Model != "unknown" || m.FromReceipt {
			t.Fatalf("model %#v", m)
		}
	}
	f3 := traceEvents(t, res.TracePath, "F3")
	kinds := make([]string, 0, len(f3))
	for _, e := range f3 {
		kinds = append(kinds, e.Kind)
	}
	if len(kinds) != 4 || kinds[0] != "claim" || kinds[1] != "evidence" || kinds[2] != "refutation" || kinds[3] != "verdict" {
		t.Fatalf("F3 events %#v", f3)
	}
	if f3[3].Text != "dropped: no evidence" {
		t.Fatalf("verdict %q", f3[3].Text)
	}
}

func TestSampleReviewDoesNotReadTheEnvironment(t *testing.T) {
	rev, err := review.Open(review.Options{
		Sample: true,
		LogDir: t.TempDir(),
		Env: func(k string) string {
			t.Fatalf("sample mode read %s", k)
			return ""
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)}); err != nil {
		t.Fatal(err)
	}
}

func TestCancelledContextIsPartial(t *testing.T) {
	rev := openSample(t, t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res, err := rev.Review(ctx, review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Partial {
		t.Fatal("want partial")
	}
}

func TestLiveModeWithoutRuntimeNamesProverOff(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	// The endpoint is closed so this test cannot reach a real Nightshift.
	// Mutated field: the listening socket. It is closed before Open returns.
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the closed endpoint accepted a call")
	}))
	srv.Close()
	rev, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env:    liveEnv(srv.URL, map[string]string{"PROVEN_GENERATOR_TIMEOUT": "1s"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := rev.Review(context.Background(), review.Request{Target: fixturePath(t)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Partial {
		t.Fatal("partial")
	}
	found := false
	for _, s := range res.FailedServices {
		if s.Service == "prover" && strings.Contains(s.Reason, "Prover off") {
			found = true
		}
	}
	if !found {
		t.Fatalf("failed %#v", res.FailedServices)
	}
	for _, f := range res.Findings {
		for _, e := range f.Evidence {
			if e.Kind == review.EvidenceTest && e.Test != nil && e.Test.FailsOnChange && e.Test.PassesOnBase {
				t.Fatalf("finding %s proved by a test %#v", f.ID, e.Test)
			}
		}
	}
}

// Mutated field: PROVEN_TEST_MEMORY. A value Validate refuses stops Open.
func TestLiveModeRefusesABadContainerSetting(t *testing.T) {
	_, err := review.Open(review.Options{
		Sample: false,
		LogDir: t.TempDir(),
		Env:    liveEnv("http://127.0.0.1:9", map[string]string{"PROVEN_TEST_MEMORY": "lots"}),
	})
	var cfg *review.ConfigError
	if !errorAs(err, &cfg) || cfg.Field != "PROVEN_TEST_MEMORY" {
		t.Fatal(err)
	}
	if !strings.Contains(err.Error(), "allowed") || !strings.Contains(err.Error(), "512m") {
		t.Fatal(err)
	}
}

func TestEmptyLogDirIsRefused(t *testing.T) {
	_, err := review.Open(review.Options{Sample: true, Env: os.Getenv})
	var cfg *review.ConfigError
	if err == nil || !errorAs(err, &cfg) || cfg.Field != "LogDir" {
		t.Fatal(err)
	}
}

func openSample(t *testing.T, logDir string) review.Reviewer {
	t.Helper()
	rev, err := review.Open(review.Options{Sample: true, LogDir: logDir, Env: os.Getenv})
	if err != nil {
		t.Fatal(err)
	}
	return rev
}

func find(res review.Result, id string) (review.Finding, bool) {
	for _, f := range res.Findings {
		if f.ID == id {
			return f, true
		}
	}
	return review.Finding{}, false
}

type traceEvent struct {
	FindingID string `json:"finding_id"`
	Kind      string `json:"kind"`
	Text      string `json:"text"`
}

func traceEvents(t *testing.T, path, id string) []traceEvent {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out []traceEvent
	for _, line := range bytes.Split(b, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e traceEvent
		if err := json.Unmarshal(line, &e); err != nil {
			t.Fatal(err)
		}
		if e.FindingID == id {
			out = append(out, e)
		}
	}
	return out
}

func fixturePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "testdata", "fixture", "discount")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && bytes.Contains(b, []byte("module github.com/saphan-ai/proven")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}

func errorAs(err error, target **review.ConfigError) bool {
	if err == nil {
		return false
	}
	type unwrapper interface{ Unwrap() error }
	for err != nil {
		if cfg, ok := err.(*review.ConfigError); ok {
			*target = cfg
			return true
		}
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
