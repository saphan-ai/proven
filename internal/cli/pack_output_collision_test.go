package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/evidence"
	"github.com/saphan-ai/proven/internal/review"
)

// assertPackIntact fails when a file at packPath is not a pack that verifies.
func assertPackIntact(t *testing.T, packPath, publicKey string) {
	t.Helper()
	body, err := os.ReadFile(packPath)
	if err != nil {
		return // no pack is left; nothing unverified stands in its place
	}
	public, err := evidence.LoadPublicKey("test", publicKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := evidence.Verify(body, public); err != nil {
		t.Fatalf("a written pack must remain verifiable: %v", err)
	}
}

func TestEvidencePackOutputCannotReplacePack(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Chdir(t.TempDir())
	privateKey, publicKey := packKeys(t)
	packPath := filepath.Join(t.TempDir(), "review.pack")
	var stdout, stderr bytes.Buffer
	exit := Run([]string{"scan", "--sample", "--evidence-pack", packPath, "-o", packPath, target}, packEnv(map[string]string{evidence.EnvPackKey: privateKey}), &stdout, &stderr, nil)
	if exit == exitOK {
		t.Fatalf("colliding output was accepted: %q", stderr.String())
	}
	assertPackIntact(t, packPath, publicKey)
}

// The mutated field is the pack's place in the collision check: each output
// flag, named with the pack's own path, must be refused before the review.
func TestEvidencePackCollidesWithEveryOutputFlag(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	for _, flag := range []string{"-o", "--output", "--text-output", "--json-output", "--sarif-output"} {
		t.Run(flag, func(t *testing.T) {
			t.Chdir(t.TempDir())
			privateKey, _ := packKeys(t)
			packPath := filepath.Join(t.TempDir(), "review.pack")
			var stdout, stderr bytes.Buffer
			exit := Run([]string{"scan", "--sample", "--evidence-pack", packPath, flag, packPath, target}, packEnv(map[string]string{evidence.EnvPackKey: privateKey}), &stdout, &stderr, nil)
			if exit == exitOK {
				t.Fatalf("%s on the pack's path was accepted", flag)
			}
			if !strings.Contains(stderr.String(), "--evidence-pack") {
				t.Fatalf("the refusal must name --evidence-pack: %q", stderr.String())
			}
			if _, err := os.Stat(packPath); err == nil {
				t.Fatal("a refusal at load must write no pack")
			}
		})
	}
}

func TestEvidencePackAliasedOutputCannotReplacePack(t *testing.T) {
	target := filepath.Join(moduleRoot(t), "testdata", "fixture", "discount")
	t.Run("a link to the existing pack path", func(t *testing.T) {
		t.Chdir(t.TempDir())
		privateKey, _ := packKeys(t)
		dir := t.TempDir()
		packPath := filepath.Join(dir, "review.pack")
		if err := os.WriteFile(packPath, []byte("old"), 0o600); err != nil {
			t.Fatal(err)
		}
		alias := filepath.Join(dir, "alias.txt")
		if err := os.Symlink(packPath, alias); err != nil {
			t.Skipf("no symbolic links here: %v", err)
		}
		var stdout, stderr bytes.Buffer
		exit := Run([]string{"scan", "--sample", "--evidence-pack", packPath, "-o", alias, target}, packEnv(map[string]string{evidence.EnvPackKey: privateKey}), &stdout, &stderr, nil)
		if exit == exitOK {
			t.Fatalf("an alias of the pack was accepted: %q", stderr.String())
		}
		if body, err := os.ReadFile(packPath); err != nil || string(body) != "old" {
			t.Fatalf("a refusal at load must leave the file as it was: %q, %v", body, err)
		}
	})
	t.Run("a link to a pack path not yet made", func(t *testing.T) {
		t.Chdir(t.TempDir())
		privateKey, publicKey := packKeys(t)
		dir := t.TempDir()
		packPath := filepath.Join(dir, "review.pack")
		alias := filepath.Join(dir, "alias.txt")
		if err := os.Symlink(packPath, alias); err != nil {
			t.Skipf("no symbolic links here: %v", err)
		}
		var stdout, stderr bytes.Buffer
		exit := Run([]string{"scan", "--sample", "--evidence-pack", packPath, "--text-output", alias, target}, packEnv(map[string]string{evidence.EnvPackKey: privateKey}), &stdout, &stderr, nil)
		if exit == exitOK {
			t.Fatalf("an alias of the pack was accepted: %q", stderr.String())
		}
		assertPackIntact(t, packPath, publicKey)
	})
	t.Run("a link to a directory holding the pack", func(t *testing.T) {
		t.Chdir(t.TempDir())
		privateKey, publicKey := packKeys(t)
		real := t.TempDir()
		linkDir := filepath.Join(t.TempDir(), "link")
		if err := os.Symlink(real, linkDir); err != nil {
			t.Skipf("no symbolic links here: %v", err)
		}
		packPath := filepath.Join(real, "review.pack")
		var stdout, stderr bytes.Buffer
		exit := Run([]string{"scan", "--sample", "--evidence-pack", packPath, "--json-output", filepath.Join(linkDir, "review.pack"), target}, packEnv(map[string]string{evidence.EnvPackKey: privateKey}), &stdout, &stderr, nil)
		if exit == exitOK {
			t.Fatalf("an alias of the pack was accepted: %q", stderr.String())
		}
		assertPackIntact(t, packPath, publicKey)
	})
}

// A link made after the load check is the second line: the run stops and the
// file the pack was written to is left as it was.
func TestWriteOutputsLeavesThePackWhenALinkAppearsAfterTheLoadCheck(t *testing.T) {
	dir := t.TempDir()
	packPath := filepath.Join(dir, "review.pack")
	const signed = "the signed pack"
	if err := os.WriteFile(packPath, []byte(signed), 0o600); err != nil {
		t.Fatal(err)
	}
	late := filepath.Join(dir, "late.txt")
	if err := os.Symlink(packPath, late); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	cfg := config{evidencePack: packPath, output: late, format: "text"}
	var stdout bytes.Buffer
	err := writeOutputs(cfg, review.Result{}, &stdout)
	if err == nil {
		t.Fatal("writeOutputs replaced the pack through a link made after the load check")
	}
	if !strings.Contains(err.Error(), "--evidence-pack") {
		t.Fatalf("the refusal must name --evidence-pack: %v", err)
	}
	body, rerr := os.ReadFile(packPath)
	if rerr != nil || string(body) != signed {
		t.Fatalf("the pack changed: %q, %v", body, rerr)
	}
}
