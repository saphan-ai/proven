package cli

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/evidence"
)

const packSecret = "TOKEN-MARKER-5c1d9e47aa"

// packKeys writes a key pair as PEM files and returns their paths.
func packKeys(t *testing.T) (priv, pub string) {
	t.Helper()
	dir := t.TempDir()
	pk, sk, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sder, _ := x509.MarshalPKCS8PrivateKey(sk)
	pder, _ := x509.MarshalPKIXPublicKey(pk)
	priv = filepath.Join(dir, "pack.key")
	pub = filepath.Join(dir, "pack.pub")
	if err := os.WriteFile(priv, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: sder}), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pub, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pder}), 0o644); err != nil {
		t.Fatal(err)
	}
	return priv, pub
}

func packEnv(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestSampleReviewWritesAPackThatVerifies(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	priv, pub := packKeys(t)
	env := packEnv(map[string]string{
		evidence.EnvPackKey:       priv,
		"METERLESS_RUNTIME_TOKEN": packSecret,
	})
	pack := filepath.Join(t.TempDir(), "review.pack")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", "--evidence-pack", pack, target}, env, &stdout, &stderr, nil)
	if code != exitOK {
		t.Fatalf("scan: exit %d\n%s", code, stderr.String())
	}
	assertDiscountText(t, stdout.String())
	body, err := os.ReadFile(pack)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(body, []byte(packSecret)) {
		t.Fatal("the pack holds the configured token")
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"verify", pack, "--public-key", pub}, packEnv(nil), &stdout, &stderr, nil)
	if code != exitOK {
		t.Fatalf("verify: exit %d\n%s", code, stderr.String())
	}
	for _, want := range []string{"evidence pack verified", "reviewed:", "models:", "evidence:", "checked:", "not checked: the provider's own signature"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("verify output has no %q:\n%s", want, stdout.String())
		}
	}

	// The public key may come from the setting as well as the flag.
	stdout.Reset()
	code = Run([]string{"verify", pack}, packEnv(map[string]string{evidence.EnvPackPublicKey: pub}), &stdout, &stderr, nil)
	if code != exitOK {
		t.Fatalf("verify by setting: exit %d\n%s", code, stderr.String())
	}

	// One byte changed anywhere in the pack is refused, with a non-zero exit.
	body[len(body)/2] ^= 0x01
	bad := filepath.Join(t.TempDir(), "bad.pack")
	if err := os.WriteFile(bad, body, 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"verify", bad, "--public-key", pub}, packEnv(nil), &stdout, &stderr, nil)
	if code == exitOK || !strings.Contains(stderr.String(), "evidence pack refused") {
		t.Fatalf("a changed pack: exit %d, stderr %q", code, stderr.String())
	}
}

func TestAMissingKeyIsNamedAndNotPrinted(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	pack := filepath.Join(t.TempDir(), "review.pack")
	for name, env := range map[string]map[string]string{
		"unset":      {},
		"unreadable": {evidence.EnvPackKey: filepath.Join(t.TempDir(), "absent.pem"), "METERLESS_RUNTIME_TOKEN": packSecret},
	} {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run([]string{"scan", "--sample", "--evidence-pack", pack, target}, packEnv(env), &stdout, &stderr, nil)
			if code != exitConfig {
				t.Fatalf("exit %d, want %d\n%s", code, exitConfig, stderr.String())
			}
			if !strings.Contains(stderr.String(), evidence.EnvPackKey) || strings.Contains(stderr.String(), packSecret) {
				t.Fatalf("the refusal must name %s and print no secret: %q", evidence.EnvPackKey, stderr.String())
			}
			if _, err := os.Stat(pack); err == nil {
				t.Fatal("a pack was written without a key")
			}
			if stdout.Len() != 0 {
				t.Fatal("the review ran before the key was refused")
			}
		})
	}
}

func TestVerifyWithoutAPublicKeyNamesTheSetting(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"verify", "x.pack"}, packEnv(nil), &stdout, &stderr, nil)
	if code != exitConfig || !strings.Contains(stderr.String(), evidence.EnvPackPublicKey) {
		t.Fatalf("exit %d: %q", code, stderr.String())
	}
}
