package testrun

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// runtimeScript writes an executable named docker in its own 0700 directory and returns its path.
func runtimeScript(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "docker")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

// loadRuntime loads the settings with the given runtime, in local or public mode.
func loadRuntime(t *testing.T, runtime string, local bool) (ContainerConfig, error) {
	t.Helper()
	extra := map[string]string{"PROVEN_TEST_RUNTIME": runtime}
	if !local {
		extra["PROVEN_NETWORK"] = "public"
	}
	return Load(localEnv(extra))
}

// wantRuntimeRefusal checks that err is a ConfigError that names the runtime setting and a reason fragment.
func wantRuntimeRefusal(t *testing.T, err error, fragment string) {
	t.Helper()
	var ce *ConfigError
	if !errors.As(err, &ce) {
		t.Fatalf("not refused as a ConfigError: %v", err)
	}
	if ce.Field != EnvRuntime || !strings.Contains(ce.Reason, fragment) || len(ce.Allowed) == 0 {
		t.Fatalf("the refusal does not say what it found and what to do: %+v (want %q)", ce, fragment)
	}
}

// TestLocalRefusesARelativeRuntimePath: a runtime given with a separator that is not absolute names a
// different program in each directory. Local mode refuses it by name and public mode does not.
// Mutated field: PROVEN_TEST_RUNTIME, from an absolute path to a relative one with the same base name.
func TestLocalRefusesARelativeRuntimePath(t *testing.T) {
	abs := runtimeScript(t, "exit 0")
	if _, err := loadRuntime(t, abs, true); err != nil {
		t.Fatalf("control: the absolute path was refused: %v", err)
	}
	for _, rel := range []string{"./docker", "bin/docker", "../docker", "sub/dir/podman"} {
		_, err := loadRuntime(t, rel, true)
		wantRuntimeRefusal(t, err, "relative path")
		if _, err := loadRuntime(t, rel, false); err != nil {
			t.Fatalf("public mode refused %q: %v", rel, err)
		}
	}
}

// TestLocalRefusesAWorldWritableRuntime: any user could replace a runtime file every user can write.
// Mutated field: the mode of the runtime file, 0700 to 0777 (set with chmod, so no umask hides it).
func TestLocalRefusesAWorldWritableRuntime(t *testing.T) {
	path := runtimeScript(t, "exit 0")
	cfg, err := loadRuntime(t, path, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewContainer(cfg, nil); err != nil {
		t.Fatalf("control: a 0700 runtime was refused: %v", err)
	}
	for _, mode := range []os.FileMode{0o777, 0o702, 0o757} {
		if err := os.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
		box, err := NewContainer(cfg, nil)
		if err == nil || box != nil {
			t.Fatalf("mode %04o: a world-writable runtime was accepted", mode)
		}
		wantRuntimeRefusal(t, err, "writable by every user")
	}
	// Public mode runs it, as before.
	pub, err := loadRuntime(t, path, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewContainer(pub, nil); err != nil {
		t.Fatalf("public mode refused a world-writable runtime: %v", err)
	}
}

// TestLocalRefusesAWorldWritableRuntimeThroughALink: the file that runs is the one judged.
// Mutated field: the mode of the link's target.
func TestLocalRefusesAWorldWritableRuntimeThroughALink(t *testing.T) {
	target := runtimeScript(t, "exit 0")
	if err := os.Chmod(target, 0o777); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "docker")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("no symbolic link here: " + err.Error())
	}
	cfg, err := loadRuntime(t, link, true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewContainer(cfg, nil)
	wantRuntimeRefusal(t, err, "writable by every user")
}

// TestLocalRefusesAWorldWritableRuntimeDirectoryWithoutSticky: a directory every user can write lets any
// user replace the file in it, unless the sticky bit stops them. Mutated field: the mode of the directory.
func TestLocalRefusesAWorldWritableRuntimeDirectoryWithoutSticky(t *testing.T) {
	path := runtimeScript(t, "exit 0")
	dir := filepath.Dir(path)
	cfg, err := loadRuntime(t, path, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	_, err = NewContainer(cfg, nil)
	wantRuntimeRefusal(t, err, "no sticky bit")
	// The sticky bit makes the same directory acceptable, so the refusal above is about that bit.
	if err := os.Chmod(dir, 0o777|os.ModeSticky); err != nil {
		t.Fatal(err)
	}
	if _, err := NewContainer(cfg, nil); err != nil {
		t.Fatalf("a world-writable directory with the sticky bit was refused: %v", err)
	}
}

// TestLocalResolvesABareRuntimeNameThroughPath: a bare name is found through PATH once, and the
// container keeps the absolute path it found.
func TestLocalResolvesABareRuntimeNameThroughPath(t *testing.T) {
	path := runtimeScript(t, "exit 0")
	t.Setenv("PATH", filepath.Dir(path))
	cfg, err := loadRuntime(t, "docker", true)
	if err != nil {
		t.Fatal(err)
	}
	box, err := NewContainer(cfg, nil)
	if err != nil {
		t.Fatalf("a bare name found in PATH was refused: %v", err)
	}
	if box.runtime != path || !filepath.IsAbs(box.runtime) {
		t.Fatalf("the container keeps %q, want the absolute path %q", box.runtime, path)
	}
	if !box.Local() {
		t.Fatal("a container loaded in local mode does not say so")
	}
}

// TestDescribeNamesThePathAndTheVersion: the start line carries the absolute path and the first line of
// --version; a runtime that cannot say is "version unknown", not a refusal.
func TestDescribeNamesThePathAndTheVersion(t *testing.T) {
	path := runtimeScript(t, "echo 'Docker version 29.9.1, build abc'; echo 'second line'")
	cfg, err := loadRuntime(t, path, true)
	if err != nil {
		t.Fatal(err)
	}
	box, err := NewContainer(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := box.Describe(context.Background())
	if want := "container runtime " + path + ": Docker version 29.9.1, build abc"; got != want {
		t.Fatalf("start line %q, want %q", got, want)
	}

	failing := runtimeScript(t, "echo 'no such flag' >&2; exit 2")
	cfg, err = loadRuntime(t, failing, true)
	if err != nil {
		t.Fatal(err)
	}
	box, err = NewContainer(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	got = box.Describe(context.Background())
	if want := "container runtime " + failing + ": version unknown"; got != want {
		t.Fatalf("start line %q, want %q", got, want)
	}
}

// TestDescribeIsBounded: a runtime that does not answer is "version unknown" when the context ends.
func TestDescribeIsBounded(t *testing.T) {
	path := runtimeScript(t, "exec sleep 30")
	cfg, err := loadRuntime(t, path, true)
	if err != nil {
		t.Fatal(err)
	}
	box, err := NewContainer(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if got := box.Describe(ctx); !strings.HasSuffix(got, ": version unknown") {
		t.Fatalf("start line %q", got)
	}
}

// TestPublicModeDoesNotDescribe: public mode loads as before and prints no start line.
func TestPublicModeDoesNotDescribe(t *testing.T) {
	path := runtimeScript(t, "exit 0")
	cfg, err := loadRuntime(t, path, false)
	if err != nil {
		t.Fatal(err)
	}
	box, err := NewContainer(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if box.Local() {
		t.Fatal("a container loaded in public mode reports local mode")
	}
}
