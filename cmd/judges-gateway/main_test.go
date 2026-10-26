package main

import (
	"bytes"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestTokenCommandsPrintTheTokenOnceAndStoreOnlyTheHash(t *testing.T) {
	file := filepath.Join(t.TempDir(), "tokens.json")
	e := env(map[string]string{"JUDGES_GATEWAY_TOKENS_FILE": file})
	exp := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)

	var out, errb bytes.Buffer
	if code := run([]string{"token", "new", "--id", "judge-ada", "--expires", exp}, e, &out, &errb, nil); code != exitOK {
		t.Fatalf("token new: %d %s", code, errb.String())
	}
	var tok string
	for _, l := range strings.Split(out.String(), "\n") {
		if v, ok := strings.CutPrefix(l, "token: "); ok {
			tok = v
		}
	}
	if !strings.HasPrefix(tok, "pjt_") {
		t.Fatalf("no token printed: %q", out.String())
	}
	raw, _ := os.ReadFile(file)
	if strings.Contains(string(raw), tok) {
		t.Error("the file holds the token")
	}

	out.Reset()
	if code := run([]string{"token", "list"}, e, &out, &errb, nil); code != exitOK || strings.Contains(out.String(), tok) || !strings.Contains(out.String(), "judge-ada") || !strings.Contains(out.String(), "active") {
		t.Fatalf("list: %d %q", code, out.String())
	}
	out.Reset()
	if code := run([]string{"token", "revoke", "--id", "judge-ada"}, e, &out, &errb, nil); code != exitOK {
		t.Fatalf("revoke: %d", code)
	}
	out.Reset()
	run([]string{"token", "list"}, e, &out, &errb, nil)
	if !strings.Contains(out.String(), "revoked") {
		t.Errorf("list after revoke: %q", out.String())
	}
}

func TestIdentityCommandsAddRevokeAndListOneRowAtATime(t *testing.T) {
	file := filepath.Join(t.TempDir(), "tokens.json")
	e := env(map[string]string{"JUDGES_GATEWAY_TOKENS_FILE": file})
	exp := time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	var out, errb bytes.Buffer
	for _, m := range []string{"Ada@Jury.Example", "*@panel.example", "abc123.access"} {
		if code := run([]string{"identity", "add", "--match", m, "--expires", exp}, e, &out, &errb, nil); code != exitOK {
			t.Fatalf("add %s: %d %s", m, code, errb.String())
		}
	}
	if code := run([]string{"identity", "add", "--match", "not valid!", "--expires", exp}, e, &out, &errb, nil); code != exitError {
		t.Errorf("a match with no shape: %d", code)
	}
	if code := run([]string{"identity", "revoke", "--match", "abc123.access"}, e, &out, &errb, nil); code != exitOK {
		t.Fatalf("revoke: %d %s", code, errb.String())
	}
	out.Reset()
	if code := run([]string{"identity", "list"}, e, &out, &errb, nil); code != exitOK {
		t.Fatalf("list: %d", code)
	}
	rows := map[string]string{}
	for _, l := range strings.Split(strings.TrimSpace(out.String()), "\n")[1:] {
		f := strings.Fields(l)
		rows[f[0]] = f[1] + " " + strings.Join(f[5:], " ")
	}
	if rows["ada@jury.example"] != "judge active" || rows["*@panel.example"] != "judge active" || !strings.HasPrefix(rows["abc123.access"], "runner revoked") {
		t.Errorf("list: %q", out.String())
	}
}

func TestTokenNewRefusals(t *testing.T) {
	file := filepath.Join(t.TempDir(), "tokens.json")
	e := env(map[string]string{"JUDGES_GATEWAY_TOKENS_FILE": file})
	var out, errb bytes.Buffer
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	if code := run([]string{"token", "new", "--id", "judge-ada", "--expires", past}, e, &out, &errb, nil); code != exitError || out.Len() != 0 {
		t.Errorf("past expiry: %d %q", code, out.String())
	}
	if code := run([]string{"token", "new", "--id", "judge-ada"}, e, &out, &errb, nil); code != exitUsage || out.Len() != 0 {
		t.Errorf("no expiry: %d", code)
	}
	if code := run([]string{"token", "list"}, env(nil), &out, &errb, nil); code != exitError {
		t.Errorf("no settings file named: %d", code)
	}
}

func TestServeListensOnLoopbackOnly(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "tokens.json")
	if err := os.WriteFile(file, []byte(`{"version":1,"tokens":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	e := env(map[string]string{
		"JUDGES_GATEWAY_TOKENS_FILE": file,
		"JUDGES_GATEWAY_PORT":        freePort(t),
		"JUDGES_GATEWAY_LOG":         filepath.Join(dir, "log.jsonl"),
		"METERLESS_RUNTIME_TOKEN":    "test-token-0001",
		"ZETARIS_REST_URL":           "http://127.0.0.1:1",
		"ZETARIS_API_KEY":            "test-key-00002",
		"ZETARIS_ORG_ID":             "42",
		"JUDGES_GATEWAY_ACCESS_TEAM": "jury.cloudflareaccess.com",
		"JUDGES_GATEWAY_ACCESS_AUD":  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	addr := make(chan net.Addr, 1)
	done := make(chan int, 1)
	var errb bytes.Buffer
	go func() { done <- run(nil, e, nil, &errb, func(a net.Addr) { addr <- a }) }()
	select {
	case a := <-addr:
		tcp := a.(*net.TCPAddr)
		if !tcp.IP.IsLoopback() {
			t.Errorf("listens on %v, not loopback", tcp.IP)
		}
		resp, err := http.Get("http://" + a.String() + "/v1/chat/completions")
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 401 {
			t.Errorf("a call with no assertion got %d, want 401", resp.StatusCode)
		}
	case code := <-done:
		t.Fatalf("serve exited %d: %s", code, errb.String())
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not start")
	}
	_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	select {
	case code := <-done:
		if code != exitOK {
			t.Errorf("exit %d", code)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("serve did not stop")
	}
}

func TestServeRefusesBadSettingsByNameWithoutTheSecret(t *testing.T) {
	file := filepath.Join(t.TempDir(), "tokens.json")
	_ = os.WriteFile(file, []byte(`{"version":1,"tokens":[]}`), 0o600)
	base := map[string]string{
		"JUDGES_GATEWAY_TOKENS_FILE": file,
		"METERLESS_RUNTIME_TOKEN":    "test-token-0001",
		"ZETARIS_REST_URL":           "http://127.0.0.1:1",
		"ZETARIS_API_KEY":            "test-key-00002",
		"ZETARIS_ORG_ID":             "42",
		"JUDGES_GATEWAY_ACCESS_TEAM": "jury.cloudflareaccess.com",
		"JUDGES_GATEWAY_ACCESS_AUD":  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	for name, tc := range map[string]struct{ key, val, names string }{
		"bad port":         {"JUDGES_GATEWAY_PORT", "99999", "JUDGES_GATEWAY_PORT"},
		"short token":      {"METERLESS_RUNTIME_TOKEN", "x", "METERLESS_RUNTIME_TOKEN"},
		"origin with path": {"PROVEN_GENERATOR_BASE_URL", "http://127.0.0.1:8088/other", "PROVEN_GENERATOR_BASE_URL"},
		"no tokens file":   {"JUDGES_GATEWAY_TOKENS_FILE", "", "JUDGES_GATEWAY_TOKENS_FILE"},
	} {
		m := map[string]string{}
		for k, v := range base {
			m[k] = v
		}
		m[tc.key] = tc.val
		var errb bytes.Buffer
		if code := run(nil, env(m), nil, &errb, nil); code != exitError || !strings.Contains(errb.String(), tc.names) {
			t.Errorf("%s: %d %q", name, code, errb.String())
		}
		if strings.Contains(errb.String(), "test-token-0001") || strings.Contains(errb.String(), "test-key-00002") {
			t.Errorf("%s: a secret is in the refusal", name)
		}
	}
}

func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
}
