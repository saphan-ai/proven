package cli

import (
	"archive/tar"
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saphan-ai/proven/internal/evidence"
	"github.com/saphan-ai/proven/internal/redact"
	"github.com/saphan-ai/proven/internal/review"
)

func assertNoControl(t *testing.T, s string) {
	t.Helper()
	for _, c := range strings.TrimSuffix(s, "\n") {
		if (c < 0x20 && c != '\n') || c == 0x7f || (c >= 0x80 && c <= 0x9f) || evidence.IsBidi(c) {
			t.Fatalf("the output holds the raw character %U: %q", c, s)
		}
	}
}

// A pack's own text reaches the terminal only through visible. The mutated
// field is a model name that holds an escape sequence and a bidirectional mark.
func TestVerifyShowsNoControlByteFromThePack(t *testing.T) {
	dir := t.TempDir()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pder, _ := x509.MarshalPKIXPublicKey(pub)
	pubPath := filepath.Join(dir, "pack.pub")
	if err := os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pder}), 0o644); err != nil {
		t.Fatal(err)
	}
	hostile := "evil\x1b]0;owned\x07\x1b[2J" + string(rune(0x202e)) + "model"
	pack, err := evidence.Build(evidence.Input{
		// A review with no finding still counts every level, as a real review does.
		Result:   review.Result{ProofCounts: review.CountProofs(nil), Models: []review.ModelUse{{Agent: "planner", Route: "r\x1b[31m", Model: hostile}}},
		Commits:  evidence.Commits{Target: ".", Baseline: "main\x1b[1m", Head: "HEAD"},
		Redactor: redact.New(nil),
		Key:      priv,
		Now:      time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "hostile.pack")
	if err := os.WriteFile(path, pack, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"verify", path, "--public-key", pubPath}, packEnv(nil), &stdout, &stderr, nil); code != exitOK {
		t.Fatalf("exit %d\n%s", code, stderr.String())
	}
	assertNoControl(t, stdout.String())
	if !strings.Contains(stdout.String(), `evil\x1b]0;owned\x07\x1b[2J`) {
		t.Fatalf("the model name must be shown escaped:\n%s", stdout.String())
	}

	// A refusal that names a file from an unsigned pack is escaped as well.
	var tb bytes.Buffer
	tw := tar.NewWriter(&tb)
	tw.WriteHeader(&tar.Header{Name: "a\x1b[2J", Linkname: "x", Typeflag: tar.TypeSymlink})
	tw.Close()
	bad := filepath.Join(dir, "bad.pack")
	if err := os.WriteFile(bad, tb.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"verify", bad, "--public-key", pubPath}, packEnv(nil), &stdout, &stderr, nil); code == exitOK {
		t.Fatal("a symlink pack was accepted")
	}
	assertNoControl(t, stderr.String())
	if !strings.Contains(stderr.String(), `a\x1b[2J`) {
		t.Fatalf("the refusal must name the file, escaped: %q", stderr.String())
	}
}

// The mutated field is the key file's mode: the run is refused with exit 7
// before the review, and no pack is written.
func TestAReadableKeyIsRefusedBeforeTheReview(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	priv, _ := packKeys(t)
	if err := os.Chmod(priv, 0o644); err != nil {
		t.Fatal(err)
	}
	pack := filepath.Join(t.TempDir(), "review.pack")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"scan", "--sample", "--evidence-pack", pack, target}, packEnv(map[string]string{evidence.EnvPackKey: priv}), &stdout, &stderr, nil)
	if code != exitConfig {
		t.Fatalf("exit %d, want %d\n%s", code, exitConfig, stderr.String())
	}
	if !strings.Contains(stderr.String(), evidence.EnvPackKey) || !strings.Contains(stderr.String(), "chmod 600") {
		t.Fatalf("the refusal must name the setting and the remedy: %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatal("the review ran before the key was refused")
	}
	if _, err := os.Stat(pack); err == nil {
		t.Fatal("a pack was written")
	}
}

// Both sources are set and they name different keys: the flag is the one used.
func TestThePublicKeyFlagBeatsTheSetting(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	priv, pub := packKeys(t)
	_, otherPub := packKeys(t)
	pack := filepath.Join(t.TempDir(), "review.pack")
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"scan", "--sample", "--evidence-pack", pack, target}, packEnv(map[string]string{evidence.EnvPackKey: priv}), &stdout, &stderr, nil); code != exitOK {
		t.Fatalf("scan: exit %d\n%s", code, stderr.String())
	}
	// The flag holds the right key; the setting holds the wrong one.
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"verify", pack, "--public-key", pub}, packEnv(map[string]string{evidence.EnvPackPublicKey: otherPub}), &stdout, &stderr, nil); code != exitOK {
		t.Fatalf("the flag must win: exit %d\n%s", code, stderr.String())
	}
	// The flag holds the wrong key and the setting the right one: the pack is refused.
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"verify", pack, "--public-key", otherPub}, packEnv(map[string]string{evidence.EnvPackPublicKey: pub}), &stdout, &stderr, nil); code == exitOK {
		t.Fatal("the setting won over the flag")
	}
}
