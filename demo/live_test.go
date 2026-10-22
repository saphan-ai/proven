package demo

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saphan-ai/proven/internal/fixture"
)

func TestShareFailsOnChangeAndPassesOnBase(t *testing.T) {
	root := repoRoot(t)
	demo := filepath.Join(root, "demo", "live")
	for _, side := range []string{"base", "change"} {
		out, code := goCmd(t, filepath.Join(demo, side), "build", "./...")
		if code != 0 {
			t.Fatalf("%s: go build exit %d\n%s", side, code, out)
		}
	}
	baseOut, baseCode := goCmd(t, filepath.Join(demo, "base"), "test", "-count=1", "./...")
	if baseCode != 0 {
		t.Fatalf("base: go test exit %d\n%s", baseCode, baseOut)
	}
	changeOut, changeCode := goCmd(t, filepath.Join(demo, "change"), "test", "-count=1", "./...")
	if changeCode == 0 {
		t.Fatalf("change: go test passed\n%s", changeOut)
	}
	if !strings.Contains(changeOut, "TestShare") || !strings.Contains(changeOut, "want 33 left 1") {
		t.Fatalf("change: failed for another reason\n%s", changeOut)
	}
}

func TestDemoFixtureLoads(t *testing.T) {
	root := repoRoot(t)
	if _, err := fixture.Load(filepath.Join(root, "demo", "live")); err != nil {
		t.Fatal(err)
	}
}

func goCmd(t *testing.T, src string, args ...string) (string, int) {
	t.Helper()
	dst := t.TempDir()
	if err := copyTree(src, dst); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = dst
	cmd.Env = offlineEnv()
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return string(out), exit.ExitCode()
	}
	t.Fatalf("go: %v\n%s", err, out)
	return "", 1
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink %s", path)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	})
}

// offlineEnv is the process environment with the module proxy turned off,
// so a test that needs a download fails instead of fetching.
func offlineEnv() []string {
	var out []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOPROXY=") || strings.HasPrefix(e, "GOSUMDB=") || strings.HasPrefix(e, "GOFLAGS=") {
			continue
		}
		out = append(out, e)
	}
	return append(out, "GOPROXY=off", "GOSUMDB=off")
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
