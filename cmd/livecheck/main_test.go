package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/generator"
	"github.com/saphan-ai/proven/internal/guardrail"
	"github.com/saphan-ai/proven/internal/livecheck"
	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/testrun"
)

const (
	// Fake secrets. A test fails if any of them is printed.
	fakeToken = "fake-token-9f3c7e"
	fakeKey   = "fake-guardrail-key-44aa"
	fakeZKey  = "fake-zetaris-key-88bb"
)

// TestHelpNamesTheExitCodes proves --help prints the three exit codes and
// exits 0.
func TestHelpNamesTheExitCodes(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--help"}, &stdout, &stderr, emptyEnv)
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, stderr.String())
	}
	body := stdout.String()
	for _, want := range []string{"\n  0  ", "\n  1  ", "\n  2  ", "not ready"} {
		if !strings.Contains(body, want) {
			t.Fatalf("help missing %q\n%s", want, body)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr %s", stderr.String())
	}
}

// TestUnknownArgumentIsRefused proves an unknown argument exits 2 and names
// the argument.
func TestUnknownArgumentIsRefused(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--pull"}, &stdout, &stderr, emptyEnv)
	if code != 2 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stderr.String(), "--pull") {
		t.Fatalf("stderr %s", stderr.String())
	}
}

// TestNoSettingsNamesEveryServiceNotReady proves that with no usable service
// and no runtime on PATH every service is named not ready, and that a secret
// set in the environment is absent from every printed line.
// Mutated field: the printed text. Printing one of the secrets makes this red.
// An empty environment cannot: the secret strings are not there to print.
func TestNoSettingsNamesEveryServiceNotReady(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, envWithSecrets)
	if code != 1 {
		t.Fatalf("exit %d\n%s", code, stdout.String())
	}
	body := stdout.String() + stderr.String()
	for _, name := range []string{"generator", "guardrail", "query", "container"} {
		if !strings.Contains(body, name+": not ready:") {
			t.Fatalf("missing %s\n%s", name, body)
		}
	}
	if !strings.Contains(body, "summary: 0 ready, 4 not ready") {
		t.Fatalf("summary\n%s", body)
	}
	if strings.Contains(body, fakeToken) || strings.Contains(body, fakeKey) || strings.Contains(body, fakeZKey) {
		t.Fatal("a secret was printed")
	}
}

// TestAllServicesReady proves that when every service answers, each line is
// ready, each service gets one call, the default view is queried, and the
// image is not pulled.
func TestAllServicesReady(t *testing.T) {
	w := newWorld(t, "ready")
	w.receipt = "rcpt-livecheck"
	w.model = "nvidia/livecheck-test"
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, w.env())
	body := stdout.String() + stderr.String()
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, body)
	}
	for _, want := range []string{
		"generator: ready: model nvidia/livecheck-test, receipt rcpt-livecheck",
		"guardrail: ready: verdict safe, latency ",
		"query: ready: rows 1",
		"container: ready",
		"summary: 4 ready, 0 not ready",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q\n%s", want, body)
		}
	}
	if w.genPosts != 1 || w.railPosts != 1 || w.queryPosts != 1 {
		t.Fatalf("posts gen %d rail %d query %d", w.genPosts, w.railPosts, w.queryPosts)
	}
	if !strings.Contains(w.sql, "SELECT 1 FROM px_evidence.px_kev_vulns LIMIT 1") {
		t.Fatalf("sql %q", w.sql)
	}
	assertNoPull(t, w.record)
	assertNoSecrets(t, body)
}

// TestGeneratorErrorIsNotReady proves one failing service makes the run not
// ready while the other three are still checked and named ready.
// Mutated field: the generator's HTTP status. One service answers an error.
func TestGeneratorErrorIsNotReady(t *testing.T) {
	w := newWorld(t, "ready")
	w.genStatus = http.StatusInternalServerError
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, w.env())
	body := stdout.String() + stderr.String()
	if code == 0 {
		t.Fatalf("exit 0\n%s", body)
	}
	if !strings.Contains(body, "generator: not ready:") {
		t.Fatalf("generator line\n%s", body)
	}
	if !strings.Contains(body, "summary: 3 ready, 1 not ready") {
		t.Fatalf("summary\n%s", body)
	}
	for _, name := range []string{"guardrail", "query", "container"} {
		if !strings.Contains(body, name+": ready") {
			t.Fatalf("%s was not ready\n%s", name, body)
		}
	}
}

// TestSecretFromTheEnvironmentIsNotPrinted proves a secret echoed back by a
// service is redacted from the printed line.
// The routed model contains the runtime token. Mutated field: redaction of
// the printed line. Removing it makes this test red.
func TestSecretFromTheEnvironmentIsNotPrinted(t *testing.T) {
	w := newWorld(t, "ready")
	w.model = "nvidia/" + fakeToken
	w.receipt = "rcpt-livecheck"
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, w.env())
	body := stdout.String() + stderr.String()
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, body)
	}
	if strings.Contains(body, fakeToken) || strings.Contains(body, fakeKey) || strings.Contains(body, fakeZKey) {
		t.Fatalf("secret leaked\n%s", body)
	}
	if !strings.Contains(body, "model nvidia/[redacted: METERLESS_RUNTIME_TOKEN], receipt rcpt-livecheck") {
		t.Fatalf("model line\n%s", body)
	}
}

// TestGuardrailOffBySettingDoesNotCall proves guardrail mode off is ready and
// sends no call.
func TestGuardrailOffBySettingDoesNotCall(t *testing.T) {
	w := newWorld(t, "ready")
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, w.envWith(map[string]string{guardrail.EnvMode: guardrail.ModeOff}))
	body := stdout.String() + stderr.String()
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, body)
	}
	if !strings.Contains(body, "guardrail: ready: off by setting") {
		t.Fatalf("guardrail line\n%s", body)
	}
	if w.railPosts != 0 {
		t.Fatalf("guardrail posts %d", w.railPosts)
	}
	assertNoSecrets(t, body)
}

// TestBadEvidenceViewIsNotSent proves a bad evidence view is refused before
// any query is sent.
// Mutated field: PROVEN_EVIDENCE_VIEW. A value that is not one view name is
// not sent.
func TestBadEvidenceViewIsNotSent(t *testing.T) {
	w := newWorld(t, "ready")
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, w.envWith(map[string]string{livecheck.EnvEvidenceView: "px_evidence.bad;drop"}))
	body := stdout.String() + stderr.String()
	if code == 0 {
		t.Fatalf("exit 0\n%s", body)
	}
	if !strings.Contains(body, "query: not ready:") {
		t.Fatalf("query line\n%s", body)
	}
	if w.queryPosts != 0 {
		t.Fatalf("query was sent: %q", w.sql)
	}
}

// TestQueryUsesTheConfiguredView proves the query check uses the configured
// evidence view in place of the default.
func TestQueryUsesTheConfiguredView(t *testing.T) {
	w := newWorld(t, "ready")
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, w.envWith(map[string]string{livecheck.EnvEvidenceView: "px_evidence.px_epss_cve_rows"}))
	body := stdout.String() + stderr.String()
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, body)
	}
	if !strings.Contains(w.sql, "SELECT 1 FROM px_evidence.px_epss_cve_rows LIMIT 1") {
		t.Fatalf("sql %q", w.sql)
	}
	if strings.Contains(w.sql, "px_kev_vulns") {
		t.Fatalf("default view was used: %q", w.sql)
	}
}

// TestMissingImageIsNotReadyAndDoesNotPull proves a missing image is named
// not ready, and the check never pulls it.
// Mutated field: image inspect's exit status, through a runtime binary on PATH.
func TestMissingImageIsNotReadyAndDoesNotPull(t *testing.T) {
	w := newWorld(t, "missing-image")
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, w.env())
	body := stdout.String() + stderr.String()
	if code == 0 {
		t.Fatalf("exit 0\n%s", body)
	}
	if !strings.Contains(body, "container: not ready:") || !strings.Contains(body, "golang:1.26") || !strings.Contains(body, "not present") {
		t.Fatalf("container line\n%s", body)
	}
	if strings.Contains(body, "pull") {
		t.Fatalf("pull was named\n%s", body)
	}
	assertNoPull(t, w.record)
	if !strings.Contains(readRecord(t, w.record), "image inspect golang:1.26") {
		t.Fatalf("record %s", readRecord(t, w.record))
	}
}

// TestContainerChecksTheConfiguredImage proves the container check inspects
// the configured image, not the default.
// Mutated field: PROVEN_TEST_IMAGE. The inspect argument is that image.
func TestContainerChecksTheConfiguredImage(t *testing.T) {
	w := newWorld(t, "missing-image")
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, w.envWith(map[string]string{testrun.EnvImage: "alpine:3"}))
	body := stdout.String() + stderr.String()
	if code == 0 {
		t.Fatalf("exit 0\n%s", body)
	}
	rec := readRecord(t, w.record)
	if !strings.Contains(rec, "image inspect alpine:3") {
		t.Fatalf("record %s", rec)
	}
	if strings.Contains(rec, "golang:1.26") {
		t.Fatalf("default image was checked\n%s", rec)
	}
	if strings.Contains(rec, "pull") {
		t.Fatalf("pull was named\n%s", rec)
	}
}

// TestBadContainerMemoryIsRefusedByName proves a bad memory setting is
// refused by name, with what is allowed, before any runtime starts.
// Mutated field: PROVEN_TEST_MEMORY. The container line names it and does not start a runtime.
func TestBadContainerMemoryIsRefusedByName(t *testing.T) {
	w := newWorld(t, "ready")
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, w.envWith(map[string]string{testrun.EnvMemory: "lots"}))
	body := stdout.String() + stderr.String()
	if code == 0 {
		t.Fatalf("exit 0\n%s", body)
	}
	if !strings.Contains(body, "container: not ready:") || !strings.Contains(body, testrun.EnvMemory) || !strings.Contains(body, "allowed") || !strings.Contains(body, "512m") {
		t.Fatalf("container line\n%s", body)
	}
	if b, err := os.ReadFile(w.record); err == nil && len(b) != 0 {
		t.Fatalf("runtime was started\n%s", b)
	}
}

// emptyEnv is an environment with nothing set.
func emptyEnv(string) string { return "" }

// envWithSecrets sets the three secrets and base URLs that Load refuses, so
// each service stops before a dial and a printed secret can fail the test.
func envWithSecrets(k string) string {
	switch k {
	case generator.EnvToken:
		return fakeToken
	case generator.EnvBaseURL:
		return "not-a-url"
	case guardrail.EnvKey:
		return fakeKey
	case guardrail.EnvBaseURL:
		return "not-a-url"
	case query.EnvAPIKey:
		return fakeZKey
	default:
		return ""
	}
}

// world is a set of fake services: the generator, the guardrail, Zetaris and
// a container runtime script. It counts the calls each one gets.
type world struct {
	genURL     string
	railURL    string
	zetURL     string
	record     string
	genStatus  int
	model      string
	receipt    string
	genPosts   int
	railPosts  int
	queryPosts int
	sql        string
}

// newWorld starts the fake services and puts a fake runtime on PATH in the
// given mode.
func newWorld(t *testing.T, runtime string) *world {
	t.Helper()
	w := &world{model: "nvidia/livecheck-test", receipt: "rcpt-livecheck"}
	gen := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		w.genPosts++
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("generator path %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer "+fakeToken {
			t.Errorf("generator authorization")
		}
		if w.genStatus != 0 {
			rw.WriteHeader(w.genStatus)
			return
		}
		rw.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(rw, sse(w.model, w.receipt))
	}))
	t.Cleanup(gen.Close)
	rail := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		w.railPosts++
		if r.Header.Get("Authorization") != "Bearer "+fakeKey {
			t.Errorf("guardrail authorization")
		}
		_, _ = io.WriteString(rw, verdictJSON("User Safety: safe\n"))
	}))
	t.Cleanup(rail.Close)
	zet := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		w.queryPosts++
		if r.URL.Path != "/sql-editor/sqls/run" {
			t.Errorf("query path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer "+fakeZKey {
			t.Errorf("query authorization")
		}
		var body struct {
			SQL string `json:"sql"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		w.sql = body.SQL
		if strings.Contains(body.SQL, fakeZKey) {
			t.Error("the SQL carried the API key")
		}
		_, _ = io.WriteString(rw, `{"headers":["one"],"data":[["1"]]}`)
	}))
	t.Cleanup(zet.Close)
	w.genURL, w.railURL, w.zetURL = gen.URL, rail.URL, zet.URL
	w.record = installRuntime(t, runtime)
	return w
}

// env is the environment that points every setting at the fakes.
func (w *world) env() func(string) string {
	return w.envWith(nil)
}

// envWith is env with some settings replaced.
func (w *world) envWith(extra map[string]string) func(string) string {
	m := map[string]string{
		generator.EnvToken:   fakeToken,
		generator.EnvBaseURL: w.genURL,
		generator.EnvTimeout: "2s",
		guardrail.EnvMode:    guardrail.ModeServer,
		guardrail.EnvBaseURL: w.railURL,
		guardrail.EnvKey:     fakeKey,
		guardrail.EnvTimeout: "2s",
		query.EnvRESTURL:     w.zetURL,
		query.EnvAPIKey:      fakeZKey,
		query.EnvOrgID:       "1",
		query.EnvTimeout:     "2s",
	}
	for k, v := range extra {
		m[k] = v
	}
	return func(k string) string { return m[k] }
}

// sse is a streamed generator answer that carries the routed model and the
// receipt id.
func sse(model, receipt string) string {
	return "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{}}],\"meterless\":{\"routed_model\":\"" + model + "\",\"receipt_id\":\"" + receipt + "\"}}\n\n" +
		"data: [DONE]\n"
}

// verdictJSON is a guardrail answer whose message is text.
func verdictJSON(text string) string {
	raw, err := json.Marshal(map[string]any{
		"choices": []any{
			map[string]any{"message": map[string]any{"content": text}},
		},
	})
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// installRuntime writes a docker script that records its arguments and puts
// it alone on PATH. Mode ready answers every call; mode missing-image
// answers only the version call.
func installRuntime(t *testing.T, mode string) string {
	t.Helper()
	dir := t.TempDir()
	record := filepath.Join(dir, "record")
	// The path is written into the shell script unquoted, so it must hold no
	// space and no quote.
	if strings.ContainsAny(record, " \t\n'\"\\") {
		t.Fatalf("record path %q", record)
	}
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + record + "\n"
	switch mode {
	case "ready":
		script += "exit 0\n"
	case "missing-image":
		script += "if [ \"$1\" = version ]; then exit 0; fi\nexit 1\n"
	default:
		t.Fatalf("runtime mode %s", mode)
	}
	if err := os.WriteFile(filepath.Join(dir, "docker"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	return record
}

// assertNoPull fails if the runtime was asked to pull, and requires the
// version and inspect calls.
func assertNoPull(t *testing.T, record string) {
	t.Helper()
	body := readRecord(t, record)
	if strings.Contains(body, "pull") {
		t.Fatalf("runtime record pulled\n%s", body)
	}
	if !strings.Contains(body, "version") || !strings.Contains(body, "image inspect golang:1.26") {
		t.Fatalf("runtime record\n%s", body)
	}
}

// readRecord reads the runtime's argument record.
func readRecord(t *testing.T, record string) string {
	t.Helper()
	b, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// assertNoSecrets fails if any fake secret is in body.
func assertNoSecrets(t *testing.T, body string) {
	t.Helper()
	for _, secret := range []string{fakeToken, fakeKey, fakeZKey} {
		if strings.Contains(body, secret) {
			t.Fatalf("secret %q leaked\n%s", secret, body)
		}
	}
}
