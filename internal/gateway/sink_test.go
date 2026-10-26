package gateway

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// failWriter is a sink that can be broken and mended.
type failWriter struct {
	mu   sync.Mutex
	fail bool
	buf  bytes.Buffer
}

func (f *failWriter) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail {
		return 0, errors.New("no space left on device")
	}
	return f.buf.Write(p)
}

func (f *failWriter) set(fail bool) {
	f.mu.Lock()
	f.fail = fail
	f.mu.Unlock()
}

func (f *failWriter) String() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.buf.String()
}

func (r *rig) health() int {
	r.t.Helper()
	resp, err := http.Get(r.gw.URL + PathHealth)
	if err != nil {
		r.t.Fatal(err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

// A log write that fails is named on the fault channel, the health probe goes
// red, the next call is refused before it reaches the upstream, and a mended
// sink receives the line that was held back.
func TestFailedLogWriteIsNeverSilent(t *testing.T) {
	log, fault := &failWriter{}, &bytes.Buffer{}
	r := newRigWith(t, func(c *Config) { c.Log, c.Fault = log, fault })
	judge := r.issue("judge-ada", 100, 1000)
	if r.health() != 200 {
		t.Fatal("a fresh gateway is not healthy")
	}

	log.set(true)
	first := r.post(PathModel, judge, chat("meterless/auto"))
	// The call that finds the sink broken was already being served. It may
	// end 200, but then the fault line and the health probe say what was lost.
	if first.status == 200 {
		if !strings.Contains(fault.String(), FaultPrefix+" sink=log") {
			t.Fatalf("a 200 with a lost log line and no fault line; fault channel: %q", fault)
		}
		if r.health() != 503 {
			t.Error("health still says ok after a lost log line")
		}
	}
	callsBefore := r.ns.count()
	second := r.post(PathModel, judge, chat("meterless/auto"))
	if second.status != 503 || refusedCode(t, second) != CodeLogUnavailable {
		t.Fatalf("with the log down: %d %s", second.status, second.body)
	}
	if r.ns.count() != callsBefore {
		t.Error("a call that cannot be logged reached the upstream")
	}
	if n := strings.Count(fault.String(), FaultPrefix); n != 1 {
		t.Errorf("fault lines = %d, want one per failure:\n%s", n, fault)
	}

	log.set(false)
	third := r.post(PathModel, judge, chat("meterless/auto"))
	if third.status != 200 {
		t.Fatalf("after the sink recovers: %d %s", third.status, third.body)
	}
	if r.health() != 200 {
		t.Error("health stays red after recovery")
	}
	if !strings.Contains(fault.String(), RecoveredPrefix+" sink=log lost=0 unlogged_refusals=1") {
		t.Errorf("no recovered line: %q", fault)
	}
	// The held-back line is delivered first: no served call is missing.
	if n := strings.Count(log.String(), `"token":"judge-ada"`); n != 2 {
		t.Errorf("served calls in the log = %d, want 2 (the first and the third):\n%s", n, log)
	}
}

// The same holds for the allowance alarm.
func TestFailedAlarmWriteIsNeverSilent(t *testing.T) {
	alarm, fault := &failWriter{}, &bytes.Buffer{}
	r := newRigWith(t, func(c *Config) { c.Alarm, c.Fault = alarm, fault })
	judge := r.issue("judge-ada", 5, 1000)
	alarm.set(true)
	var last reply
	for i := 0; i < 4; i++ { // the fourth call is at 80% of 5
		last = r.post(PathModel, judge, chat("meterless/auto"))
	}
	if last.status == 200 && !strings.Contains(fault.String(), FaultPrefix+" sink=alarm") {
		t.Fatalf("an alarm was lost and nothing says so: %q", fault)
	}
	if r.health() != 503 {
		t.Error("health is ok while the alarm cannot be written")
	}
	calls := r.ns.count()
	if rep := r.post(PathModel, judge, chat("meterless/auto")); rep.status != 503 || refusedCode(t, rep) != CodeLogUnavailable {
		t.Fatalf("with the alarm down: %d %s", rep.status, rep.body)
	}
	if r.ns.count() != calls {
		t.Error("a call reached the upstream while the alarm was down")
	}
	alarm.set(false)
	if rep := r.post(PathModel, judge, chat("meterless/auto")); rep.status != 200 {
		t.Fatalf("after the alarm sink recovers: %d", rep.status)
	}
	if n := strings.Count(alarm.String(), AlarmPrefix); n != 1 {
		t.Errorf("the held alarm is delivered once, got %d:\n%s", n, alarm)
	}
}

// bigUpstream answers a model call with n bytes. declared sets Content-Length;
// otherwise the answer is chunked.
func bigUpstream(t *testing.T, n int, declared bool) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = io.Copy(io.Discard, req.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		if declared {
			w.Header().Set("Content-Length", strconv.Itoa(n))
		}
		w.WriteHeader(200)
		chunk := bytes.Repeat([]byte("x"), 1000)
		for sent := 0; sent < n; {
			m := min(len(chunk), n-sent)
			if _, err := w.Write(chunk[:m]); err != nil {
				return
			}
			sent += m
			if f, ok := w.(http.Flusher); ok && !declared {
				f.Flush()
			}
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

const replyCap = 4096

func capRig(t *testing.T, n int, declared bool) *rig {
	up := bigUpstream(t, n, declared)
	return newRigWith(t, func(c *Config) { c.NightshiftURL, c.MaxModelReply = up.URL, replyCap })
}

// lastLine returns the last call-log line, waiting briefly: a cut stream is
// logged while the connection is closing.
func lastLine(t *testing.T, r *rig) string {
	t.Helper()
	for i := 0; i < 100; i++ {
		if s := strings.TrimSpace(r.log.String()); s != "" {
			lines := strings.Split(s, "\n")
			return lines[len(lines)-1]
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("no log line")
	return ""
}

func TestModelAnswerOverTheCapIsCutAndNamed(t *testing.T) {
	r := capRig(t, 5*replyCap, false)
	judge := r.issue("judge-ada", 10, 100)
	req, _ := http.NewRequest(http.MethodPost, r.gw.URL+PathModel, strings.NewReader(chat("meterless/auto")))
	req.Header.Set("Authorization", "Bearer "+judge)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	got, rerr := io.ReadAll(resp.Body)
	if len(got) > replyCap {
		t.Fatalf("the judge received %d bytes, the cap is %d", len(got), replyCap)
	}
	if rerr == nil {
		t.Error("the cut stream ended cleanly: the judge cannot tell it was cut")
	}
	line := lastLine(t, r)
	for _, want := range []string{`"refused":"` + CodeResponseTooLarge + `"`, `"bytes":` + strconv.Itoa(replyCap), `"status":200`} {
		if !strings.Contains(line, want) {
			t.Errorf("log line lacks %s: %s", want, line)
		}
	}
}

func TestModelAnswerDeclaredOverTheCapIsRefusedBeforeAnyByte(t *testing.T) {
	r := capRig(t, 5*replyCap, true)
	judge := r.issue("judge-ada", 10, 100)
	rep := r.post(PathModel, judge, chat("meterless/auto"))
	if rep.status != 502 || refusedCode(t, rep) != CodeResponseTooLarge {
		t.Fatalf("%d %s", rep.status, rep.body)
	}
	if strings.Contains(rep.body, "xxxx") {
		t.Error("upstream bytes reached the judge")
	}
	line := lastLine(t, r)
	if !strings.Contains(line, `"refused":"`+CodeResponseTooLarge+`"`) || !strings.Contains(line, `"status":502`) {
		t.Errorf("log line: %s", line)
	}
}

func TestModelAnswerUnderTheCapIsRelayedWhole(t *testing.T) {
	for _, declared := range []bool{false, true} {
		r := capRig(t, replyCap, declared) // exactly the cap
		judge := r.issue("judge-ada", 10, 100)
		rep := r.post(PathModel, judge, chat("meterless/auto"))
		if rep.status != 200 || len(rep.body) != replyCap {
			t.Errorf("declared=%v: %d, %d bytes, want 200, %d", declared, rep.status, len(rep.body), replyCap)
		}
		if line := lastLine(t, r); strings.Contains(line, `"refused"`) || !strings.Contains(line, `"bytes":`+strconv.Itoa(replyCap)) {
			t.Errorf("declared=%v: log line: %s", declared, line)
		}
	}
}

func TestModelReplySettingIsRefusedByNameWhenBad(t *testing.T) {
	for _, bad := range []string{"abc", "0", "100", "-5", "99999999999"} {
		_, err := Load(func(k string) string {
			switch k {
			case EnvMaxModelReply:
				return bad
			case EnvTokensFile:
				return t.TempDir() + "/none.json"
			}
			return ""
		})
		var ce *ConfigError
		if !errors.As(err, &ce) {
			continue // the token file or another setting is refused first
		}
		if ce.Field == EnvMaxModelReply && !strings.Contains(ce.Allowed, "from 1024 to 67108864") {
			t.Errorf("%q: %v", bad, err)
		}
	}
}
