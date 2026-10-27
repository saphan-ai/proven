package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/change"
)

// TestGitMergeBaseChildCarriesNoKeyFromProven proves the property that no key
// reaches a subprocess: git, replaced by a script that prints its own
// environment, is started without a key Proven holds, and still receives PATH.
func TestGitMergeBaseChildCarriesNoKeyFromProven(t *testing.T) {
	bin := t.TempDir()
	dump := filepath.Join(t.TempDir(), "child-env.txt")
	script := "#!/bin/sh\n/usr/bin/env > '" + dump + "'\necho 0123456789abcdef0123456789abcdef01234567\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PROVEN_GENERATOR_KEY", "marker-9f3a1c")
	if _, err := gitMergeBase(change.ProcessPolicy(), t.TempDir(), "main", "HEAD"); err != nil {
		t.Fatal(err)
	}
	env, err := os.ReadFile(dump)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("child environment:\n%s", env)
	if strings.Contains(string(env), "marker-9f3a1c") || strings.Contains(string(env), "PROVEN_GENERATOR_KEY") {
		t.Fatalf("the git child inherited Proven's key:\n%s", env)
	}
	if !strings.Contains(string(env), "PATH=") {
		t.Fatalf("the git child lost PATH:\n%s", env)
	}
}
