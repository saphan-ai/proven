package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/query"
	"github.com/saphan-ai/proven/internal/querytpl"
)

// The test secrets. A test that greps for them proves nothing leaked.
const (
	nightshiftSecret = "NIGHTSHIFT-secret-0001"
	zetarisSecret    = "ZETARIS-secret-0002"
	codeMarker       = "JUDGE-CODE-MARKER-0003"
	sqlMarker        = "SELECT-MARKER-0004"
)

// upstream is a fake service that records every request it receives.
type upstream struct {
	mu      sync.Mutex
	calls   int
	headers []http.Header
	bodies  []string
	status  int
	reply   string
	srv     *httptest.Server
}

func newUpstream(t *testing.T, status int, reply string) *upstream {
	t.Helper()
	u := &upstream{status: status, reply: reply}
	u.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		u.mu.Lock()
		u.calls++
		u.headers = append(u.headers, r.Header.Clone())
		u.bodies = append(u.bodies, string(b))
		u.mu.Unlock()
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(u.status)
		_, _ = io.WriteString(w, u.reply)
	}))
	t.Cleanup(u.srv.Close)
	return u
}

func (u *upstream) count() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.calls
}

// rig is a gateway wired to two fake upstreams, a token file and a clock.
type rig struct {
	t      *testing.T
	gw     *httptest.Server
	ns, zt *upstream
	path   string
	log    *failWriter
	alarm  *failWriter
	mu     sync.Mutex
	now    time.Time
}

var t0 = time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)

func newRig(t *testing.T) *rig { return newRigWith(t, nil) }

// newRigWith builds the rig, then lets tweak change the gateway's
// configuration before the gateway is created.
func newRigWith(t *testing.T, tweak func(*Config)) *rig {
	t.Helper()
	r := &rig{t: t, now: t0, log: &failWriter{}, alarm: &failWriter{}}
	r.ns = newUpstream(t, 200, "data: {\"model\":\"nvidia/x\"}\n\ndata: [DONE]\n\n")
	r.zt = newUpstream(t, 200, `{"headers":["id"],"data":[["GHSA-1"]]}`)
	r.zt.srv.Config.Handler = wrapJSON(r.zt)
	r.path = filepath.Join(t.TempDir(), "tokens.json")
	store := func() *Store {
		// An empty file is a valid settings file with no token.
		if err := os.WriteFile(r.path, []byte(`{"version":1,"tokens":[]}`), 0o600); err != nil {
			t.Fatal(err)
		}
		s, err := NewStore(r.path)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}()
	live := query.NewLive(query.Settings{BaseURL: r.zt.srv.URL, APIKey: zetarisSecret, OrgID: "42", RowLimit: 10, Timeout: 5 * time.Second}, nil)
	cfg := Config{
		Tokens: store, Limiter: NewLimiter(),
		NightshiftURL: r.ns.srv.URL, NightshiftToken: nightshiftSecret, NightshiftApp: "proven",
		Models:   []string{"meterless/auto"},
		Upstream: &http.Client{Timeout: 5 * time.Second},
		Zetaris:  live, MaxModelBody: 1 << 16, MaxEvidenceBody: 1 << 10,
		Log: r.log, Alarm: r.alarm,
		Now: func() time.Time { r.mu.Lock(); defer r.mu.Unlock(); return r.now },
	}
	if tweak != nil {
		tweak(&cfg)
	}
	r.gw = httptest.NewServer(New(cfg))
	t.Cleanup(r.gw.Close)
	return r
}

// wrapJSON makes the fake Zetaris answer with JSON, not an event stream.
func wrapJSON(u *upstream) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		u.mu.Lock()
		u.calls++
		u.headers = append(u.headers, r.Header.Clone())
		u.bodies = append(u.bodies, string(b))
		u.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, u.reply)
	})
}

func (r *rig) advance(d time.Duration) {
	r.mu.Lock()
	r.now = r.now.Add(d)
	r.mu.Unlock()
}

// issue makes a token valid for a day.
func (r *rig) issue(id string, perMin, perDay int) string {
	r.t.Helper()
	tok, err := Add(r.path, id, t0.Add(24*time.Hour), perMin, perDay, t0)
	if err != nil {
		r.t.Fatal(err)
	}
	return tok
}

type reply struct {
	status int
	body   string
	header http.Header
}

func (r *rig) post(path, token, body string) reply {
	r.t.Helper()
	req, _ := http.NewRequest(http.MethodPost, r.gw.URL+path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		r.t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return reply{resp.StatusCode, string(b), resp.Header}
}

func refusedCode(t *testing.T, rep reply) string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal([]byte(rep.body), &m); err != nil {
		t.Fatalf("not a refusal body: %q", rep.body)
	}
	return m["refused"]
}

func chat(model string) string {
	return `{"model":"` + model + `","stream":true,"messages":[{"role":"user","content":"` + codeMarker + `"}]}`
}

func evidence(template, params string) string {
	return `{"template":"` + template + `","params":` + params + `}`
}

const kevParams = `{"cve_id":"CVE-2021-44228"}`

func TestModelRouteInjectsOurTokenAndNeverTheJudges(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 10, 100)
	rep := r.post(PathModel, judge, chat("meterless/auto"))
	if rep.status != 200 || !strings.Contains(rep.body, "[DONE]") {
		t.Fatalf("model call: %d %q", rep.status, rep.body)
	}
	if r.ns.count() != 1 {
		t.Fatalf("upstream calls: %d", r.ns.count())
	}
	h := r.ns.headers[0]
	if got := h.Get("Authorization"); got != "Bearer "+nightshiftSecret {
		t.Errorf("upstream Authorization = %q, want our token", got)
	}
	for name, vals := range h {
		for _, v := range vals {
			if strings.Contains(v, judge) {
				t.Errorf("the judge's token went upstream in header %s", name)
			}
		}
	}
	if strings.Contains(r.ns.bodies[0], judge) {
		t.Error("the judge's token went upstream in the body")
	}
	if !strings.Contains(r.ns.bodies[0], codeMarker) {
		t.Error("the chat body was not forwarded")
	}
	// The judge never receives our credential.
	all := rep.body
	for k, v := range rep.header {
		all += k + strings.Join(v, "")
	}
	for _, s := range []string{nightshiftSecret, zetarisSecret} {
		if strings.Contains(all, s) {
			t.Errorf("the judge received a credential of ours")
		}
	}
}

func TestExpiredTokenIsRefusedByName(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 10, 100)
	if rep := r.post(PathModel, judge, chat("meterless/auto")); rep.status != 200 {
		t.Fatalf("before expiry: %d", rep.status)
	}
	// The mutated field is the clock against the token's expires.
	r.advance(24 * time.Hour)
	rep := r.post(PathModel, judge, chat("meterless/auto"))
	if rep.status != 401 || refusedCode(t, rep) != CodeTokenExpired {
		t.Fatalf("after expiry: %d %s", rep.status, rep.body)
	}
	if r.ns.count() != 1 {
		t.Errorf("an expired token reached the upstream: %d calls", r.ns.count())
	}
}

func TestUnknownAndMissingTokensAreRefused(t *testing.T) {
	r := newRig(t)
	r.issue("judge-ada", 10, 100)
	other := t.TempDir() + "/other.json"
	foreign, err := Add(other, "judge-bob", t0.Add(time.Hour), 5, 5, t0)
	if err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct{ token, code string }{
		"foreign": {foreign, CodeUnknownToken},
		"garbage": {"pjt_nothing", CodeUnknownToken},
		"missing": {"", CodeMissingToken},
	} {
		for _, path := range []string{PathModel, PathEvidence} {
			rep := r.post(path, tc.token, chat("meterless/auto"))
			if rep.status != 401 || refusedCode(t, rep) != tc.code {
				t.Errorf("%s on %s: %d %s", name, path, rep.status, rep.body)
			}
		}
	}
	if r.ns.count()+r.zt.count() != 0 {
		t.Error("a refused token reached an upstream")
	}
}

func TestRevokingOneTokenLeavesTheOthersWorking(t *testing.T) {
	r := newRig(t)
	ada := r.issue("judge-ada", 10, 100)
	bob := r.issue("judge-bob", 10, 100)
	for _, tok := range []string{ada, bob} {
		if rep := r.post(PathModel, tok, chat("meterless/auto")); rep.status != 200 {
			t.Fatalf("before revoke: %d", rep.status)
		}
	}
	if err := Revoke(r.path, "judge-ada", t0); err != nil {
		t.Fatal(err)
	}
	rep := r.post(PathModel, ada, chat("meterless/auto"))
	if rep.status != 401 || refusedCode(t, rep) != CodeTokenRevoked {
		t.Fatalf("revoked token: %d %s", rep.status, rep.body)
	}
	if rep := r.post(PathModel, bob, chat("meterless/auto")); rep.status != 200 {
		t.Fatalf("the other token stopped working: %d %s", rep.status, rep.body)
	}
}

func TestSettingsFileThatCannotBeReadClosesTheGateway(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 10, 100)
	good, _ := os.ReadFile(r.path)
	if err := os.WriteFile(r.path, []byte(`{"version":1,"tokens":[{"id":"judge-ada"`), 0o600); err != nil {
		t.Fatal(err)
	}
	rep := r.post(PathModel, judge, chat("meterless/auto"))
	if rep.status != 503 || refusedCode(t, rep) != CodeSettingsDown {
		t.Fatalf("broken file: %d %s", rep.status, rep.body)
	}
	if err := os.WriteFile(r.path, append(good, ' '), 0o600); err != nil {
		t.Fatal(err)
	}
	if rep := r.post(PathModel, judge, chat("meterless/auto")); rep.status != 200 {
		t.Fatalf("repaired file: %d %s", rep.status, rep.body)
	}
}

func TestModelOutsideTheListIsRefused(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 10, 100)
	rep := r.post(PathModel, judge, chat("some/other-model"))
	if rep.status != 400 || refusedCode(t, rep) != CodeModelNotAllowed {
		t.Fatalf("%d %s", rep.status, rep.body)
	}
	if r.ns.count() != 0 {
		t.Error("a refused model reached the upstream")
	}
}

func TestModelBodyIsCheckedAsNightshiftReadsIt(t *testing.T) {
	// The model allow-list read the body case-insensitively, last key winning, and forwarded the original bytes: a
	// second "MODEL" key passed the check while Nightshift read the first. A field that could pick another model is
	// refused too. Each body below reaches no upstream.
	r := newRig(t)
	judge := r.issue("judge-ada", 10, 100)
	msgs := `"messages":[{"role":"user","content":"` + codeMarker + `"}]`
	for name, body := range map[string]string{
		"case variant":    `{"model":"evil/x","MODEL":"meterless/auto",` + msgs + `}`,
		"repeated key":    `{"model":"evil/x","model":"meterless/auto",` + msgs + `}`,
		"fallback models": `{"model":"meterless/auto","models":["evil/x"],` + msgs + `}`,
		"provider route":  `{"model":"meterless/auto","provider":{"order":["x"]},` + msgs + `}`,
		"trailing data":   `{"model":"meterless/auto",` + msgs + `}{"model":"evil/x"}`,
	} {
		rep := r.post(PathModel, judge, body)
		if rep.status != 400 || refusedCode(t, rep) != CodeBadRequest {
			t.Errorf("%s: %d %s", name, rep.status, rep.body)
		}
	}
	if r.ns.count() != 0 {
		t.Fatalf("a refused body reached the upstream %d times", r.ns.count())
	}
	// What is forwarded is the body that was checked: the allowed model, re-encoded, nothing else.
	rep := r.post(PathModel, judge, `{"model":"meterless/auto","stream":true,"thinking_budget":512,`+msgs+`}`)
	if rep.status != 200 || r.ns.count() != 1 {
		t.Fatalf("an allowed body: %d %s", rep.status, rep.body)
	}
	var sent map[string]json.RawMessage
	if err := json.Unmarshal([]byte(r.ns.bodies[0]), &sent); err != nil || string(sent["model"]) != `"meterless/auto"` || string(sent["thinking_budget"]) != "512" || len(sent) != 4 {
		t.Errorf("the forwarded body is not the checked one: %s", r.ns.bodies[0])
	}
}

func TestUpstreamErrorBodyIsNotPassedOn(t *testing.T) {
	r := newRig(t)
	r.ns.status, r.ns.reply = 500, "boom "+nightshiftSecret
	judge := r.issue("judge-ada", 10, 100)
	rep := r.post(PathModel, judge, chat("meterless/auto"))
	if rep.status != 502 || refusedCode(t, rep) != CodeUpstreamError {
		t.Fatalf("%d %s", rep.status, rep.body)
	}
	if strings.Contains(rep.body, nightshiftSecret) {
		t.Error("an upstream error body reached the judge")
	}
}

func TestEvidenceRouteSendsOurKeyAndTheBuiltSQL(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 10, 100)
	rep := r.post(PathEvidence, judge, evidence(querytpl.KEVByCVE, kevParams))
	if rep.status != 200 {
		t.Fatalf("%d %s", rep.status, rep.body)
	}
	var got struct {
		Rows [][]string `json:"rows"`
	}
	if err := json.Unmarshal([]byte(rep.body), &got); err != nil || len(got.Rows) != 1 || got.Rows[0][0] != "GHSA-1" {
		t.Fatalf("rows: %s", rep.body)
	}
	if r.zt.count() != 1 {
		t.Fatalf("zetaris calls: %d", r.zt.count())
	}
	if a := r.zt.headers[0].Get("Authorization"); a != "Bearer "+zetarisSecret {
		t.Errorf("zetaris Authorization = %q, want our key", a)
	}
	for name, vals := range r.zt.headers[0] {
		for _, v := range vals {
			if strings.Contains(v, judge) {
				t.Errorf("the judge's token went to zetaris in header %s", name)
			}
		}
	}
	var sent struct {
		SQL string `json:"sql"`
	}
	_ = json.Unmarshal([]byte(r.zt.bodies[0]), &sent)
	want, _ := querytpl.Build(querytpl.Request{Template: querytpl.KEVByCVE, Params: map[string]string{"cve_id": "CVE-2021-44228"}}, querytpl.Relations{})
	if sent.SQL != want {
		t.Errorf("zetaris got %q, want the SQL Proven builds: %q", sent.SQL, want)
	}
	if strings.Contains(rep.body, zetarisSecret) {
		t.Error("the judge received our Zetaris key")
	}
}

func TestEvidenceRouteRefusesEverythingButATemplate(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 100, 1000)
	cases := []struct{ name, body, code string }{
		{"raw sql", `{"sql":"` + sqlMarker + ` 1"}`, CodeNotATemplate},
		{"raw sql with a template name", `{"template":"` + querytpl.KEVByCVE + `","params":` + kevParams + `,"sql":"` + sqlMarker + `"}`, CodeNotATemplate},
		{"sql as the template name", evidence(sqlMarker+" * FROM users", `{}`), CodeUnknownTemplate},
		{"a bare statement", sqlMarker + " 1", CodeNotATemplate},
		{"an array", `[]`, CodeNotATemplate},
		{"empty object", `{}`, CodeNotATemplate},
		{"two documents", evidence(querytpl.KEVByCVE, kevParams) + evidence(querytpl.KEVByCVE, kevParams), CodeNotATemplate},
		{"unknown template", evidence("no-such-template", `{}`), CodeUnknownTemplate},
		{"missing parameter", evidence(querytpl.KEVByCVE, `{}`), CodeMissingParameter},
		{"extra parameter", evidence(querytpl.KEVByCVE, `{"cve_id":"CVE-2021-44228","x":"1"}`), CodeUnknownParameter},
		{"bad parameter", evidence(querytpl.KEVByCVE, `{"cve_id":"CVE-1'; DROP TABLE t; --"}`), CodeBadParameter},
	}
	for _, tc := range cases {
		rep := r.post(PathEvidence, judge, tc.body)
		if rep.status != 400 || refusedCode(t, rep) != tc.code {
			t.Errorf("%s: %d %s, want %s", tc.name, rep.status, rep.body, tc.code)
		}
		if strings.Contains(rep.body, sqlMarker) || strings.Contains(rep.body, "DROP") {
			t.Errorf("%s: the refusal repeats the caller's text: %s", tc.name, rep.body)
		}
	}
	if r.zt.count() != 0 {
		t.Errorf("a refused request reached zetaris: %d calls", r.zt.count())
	}
}

func TestEvidenceBodyOverTheLimitIsRefused(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 10, 100)
	rep := r.post(PathEvidence, judge, evidence(querytpl.KEVByCVE, `{"cve_id":"`+strings.Repeat("A", 2000)+`"}`))
	if rep.status != 413 || refusedCode(t, rep) != CodeBodyTooLarge {
		t.Fatalf("%d %s", rep.status, rep.body)
	}
}

func TestUnknownRouteAndMethodAreRefused(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 10, 100)
	if rep := r.post("/v1/models", judge, ""); rep.status != 404 || refusedCode(t, rep) != CodeNoRoute {
		t.Errorf("route: %d %s", rep.status, rep.body)
	}
	req, _ := http.NewRequest(http.MethodGet, r.gw.URL+PathModel, nil)
	req.Header.Set("Authorization", "Bearer "+judge)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 405 {
		t.Errorf("method: %d", resp.StatusCode)
	}
}

func TestLogLineHasNoBodyAndNoCredential(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 100, 1000)
	r.post(PathModel, judge, chat("meterless/auto"))
	r.post(PathModel, judge, chat("not/allowed"))
	r.post(PathEvidence, judge, evidence(querytpl.KEVByCVE, kevParams))
	r.post(PathEvidence, judge, `{"sql":"`+sqlMarker+`"}`)
	r.post(PathEvidence, judge, evidence(sqlMarker, `{}`))
	r.post(PathModel, "pjt_wrong-"+codeMarker, chat("meterless/auto"))
	r.post(PathModel, "", chat("meterless/auto"))

	lines := strings.Split(strings.TrimSpace(r.log.String()), "\n")
	if len(lines) != 7 {
		t.Fatalf("%d log lines for 7 calls:\n%s", len(lines), r.log)
	}
	for _, line := range lines {
		var e map[string]any
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("not JSON: %q", line)
		}
		for _, k := range []string{"time", "token", "route", "status", "bytes"} {
			if _, ok := e[k]; !ok {
				t.Errorf("log line lacks %q: %s", k, line)
			}
		}
	}
	everything := r.log.String() + r.alarm.String()
	for _, secret := range []string{judge, nightshiftSecret, zetarisSecret, codeMarker, sqlMarker, "DROP", "Bearer"} {
		if strings.Contains(everything, secret) {
			t.Errorf("the log holds %q", secret)
		}
	}
	for _, want := range []string{`"token":"judge-ada"`, `"model":"meterless/auto"`, `"template":"` + querytpl.KEVByCVE + `"`} {
		if !strings.Contains(r.log.String(), want) {
			t.Errorf("the log lacks %s", want)
		}
	}
}

func TestMinuteLimitRefusesPastNAndAlarmsOnce(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 5, 100)
	var codes []int
	for i := 0; i < 8; i++ {
		codes = append(codes, r.post(PathModel, judge, chat("meterless/auto")).status)
	}
	want := []int{200, 200, 200, 200, 200, 429, 429, 429}
	for i := range want {
		if codes[i] != want[i] {
			t.Fatalf("statuses %v, want %v", codes, want)
		}
	}
	if n := r.ns.count(); n != 5 {
		t.Errorf("upstream calls = %d, want 5; a refused call must not reach it", n)
	}
	// One alarm, for the minute window; the day window (100) is far from 80%.
	if n := strings.Count(r.alarm.String(), AlarmPrefix); n != 1 {
		t.Errorf("alarms = %d:\n%s", n, r.alarm)
	}
	if got := r.alarm.String(); !strings.Contains(got, AlarmPrefix+" token=judge-ada window=minute used=4 limit=5") {
		t.Errorf("alarm line: %q", got)
	}
	// The next minute starts again, and the alarm may fire once more.
	r.advance(61 * time.Second)
	if rep := r.post(PathModel, judge, chat("meterless/auto")); rep.status != 200 {
		t.Errorf("next minute: %d", rep.status)
	}
}

func TestDayAllowanceRefusesPastN(t *testing.T) {
	r := newRig(t)
	judge := r.issue("judge-ada", 50, 3)
	for i := 0; i < 3; i++ {
		if rep := r.post(PathModel, judge, chat("meterless/auto")); rep.status != 200 {
			t.Fatalf("call %d: %d", i, rep.status)
		}
	}
	rep := r.post(PathModel, judge, chat("meterless/auto"))
	if rep.status != 429 || refusedCode(t, rep) != CodeAllowanceDay {
		t.Fatalf("%d %s", rep.status, rep.body)
	}
	if !strings.Contains(r.alarm.String(), "window=day used=3 limit=3") {
		t.Errorf("no day alarm: %q", r.alarm)
	}
}

func TestLimitsAreOnePerToken(t *testing.T) {
	r := newRig(t)
	ada := r.issue("judge-ada", 1, 100)
	bob := r.issue("judge-bob", 1, 100)
	r.post(PathModel, ada, chat("meterless/auto"))
	if rep := r.post(PathModel, ada, chat("meterless/auto")); rep.status != 429 {
		t.Fatalf("ada past her limit: %d", rep.status)
	}
	if rep := r.post(PathModel, bob, chat("meterless/auto")); rep.status != 200 {
		t.Fatalf("bob was held by ada's limit: %d", rep.status)
	}
}
