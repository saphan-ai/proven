package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sample writes a receipts file with one receipt for chatcmpl-1 signed by a
// fixed test key, and returns the file and the key, base64. The model that
// answered is nvidia/test.
func sample(t *testing.T, dir string) (string, string) {
	t.Helper()
	key := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{3}, ed25519.SeedSize))
	pub := base64.StdEncoding.EncodeToString(key.Public().(ed25519.PublicKey))
	body, err := json.Marshal(map[string]any{
		"ts_ms": 1792921200000,
		"receipt": map[string]any{
			"request_id": "chatcmpl-1", "status": "ok",
			"winning_route": map[string]any{"provider": "nvidia", "model": "nvidia/test"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	env, err := json.Marshal(map[string]any{
		"body": json.RawMessage(body), "pubkey_b64": pub,
		"sig_b64": base64.StdEncoding.EncodeToString(ed25519.Sign(key, body)),
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "receipts.jsonl")
	if err := os.WriteFile(path, append(env, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return path, pub
}

func traceFile(t *testing.T, dir, response, requested string) string {
	t.Helper()
	line, err := json.Marshal(map[string]any{
		"time": "2026-10-25T09:00:00Z", "kind": "model", "agent": "judge",
		"model": map[string]any{"model": "nvidia/test", "receipt_id": "rcpt_1", "requested": requested, "response_id": response},
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "trace.jsonl")
	if err := os.WriteFile(path, append(line, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// Property: the command exits 0 when every answer is in custody, 1 when one
// is not, and 2 when it cannot run. Mutated input: the trace's response id,
// then the requested model, then the pinned key.
func TestExitCodesFollowCustody(t *testing.T) {
	dir := t.TempDir()
	receipts, pub := sample(t, dir)

	var out, errb bytes.Buffer
	if code := run([]string{"-receipts", receipts, "-pubkey", pub, traceFile(t, dir, "chatcmpl-1", "nvidia/test")}, &out, &errb); code != 0 {
		t.Fatalf("custody holds: exit %d\n%s%s", code, out.String(), errb.String())
	}
	if !strings.Contains(out.String(), "custody holds for every answer") {
		t.Fatalf("output:\n%s", out.String())
	}

	out.Reset()
	if code := run([]string{"-receipts", receipts, "-pubkey", pub, traceFile(t, dir, "chatcmpl-9", "nvidia/test")}, &out, &errb); code != 1 {
		t.Fatalf("no receipt: exit %d, want 1\n%s", code, out.String())
	}
	if !strings.Contains(out.String(), "no receipt matches response id chatcmpl-9") {
		t.Fatalf("output:\n%s", out.String())
	}

	out.Reset()
	if code := run([]string{"-receipts", receipts, "-pubkey", pub, traceFile(t, dir, "chatcmpl-1", "nvidia/asked-for")}, &out, &errb); code != 1 {
		t.Fatalf("model mismatch: exit %d, want 1\n%s", code, out.String())
	}

	if code := run([]string{"-receipts", receipts, "-pubkey", "AAAA", traceFile(t, dir, "chatcmpl-1", "nvidia/test")}, &out, &errb); code != 2 {
		t.Fatalf("bad key: exit %d, want 2", code)
	}
	if code := run([]string{"-receipts", receipts, "-pubkey", pub}, &out, &errb); code != 2 {
		t.Fatalf("no trace: exit %d, want 2", code)
	}
}
