package testrun

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLocalRuntimeRefusesAWritableAncestor: a directory above the runtime's own
// that any user can write, without the sticky bit, lets that user replace the
// directory that holds the runtime, so local mode refuses it. Mutated field:
// the ancestor's mode (0777 instead of 0700); its child and the file stay private.
func TestLocalRuntimeRefusesAWritableAncestor(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	runtime := filepath.Join(bin, "docker")
	if err := os.WriteFile(runtime, []byte("#!/bin/sh\necho docker-test\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultContainerConfig()
	cfg.local = true
	cfg.Runtime = runtime
	if _, err := NewContainer(cfg, nil); err != nil {
		t.Fatalf("control: a private ancestor must pass: %v", err)
	}
	if err := os.Chmod(root, 0o777); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o700) })
	if _, err := NewContainer(cfg, nil); err == nil {
		t.Fatal("local mode accepted a runtime whose ancestor directory is world-writable without the sticky bit")
	}
}

// TestLocalRuntimeRefusesAWritableLinkHop: the runtime is reached through two
// links, bin/docker -> hop/docker -> real/docker. The directory of the middle
// link is writable by every user without the sticky bit, so any user could
// replace hop/docker, although bin and real are private. Mutated field: the
// mode of the middle link's directory (0777 instead of 0700).
func TestLocalRuntimeRefusesAWritableLinkHop(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"bin", "hop", "real"} {
		if err := os.Mkdir(filepath.Join(root, d), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(root, "real", "docker")
	if err := os.WriteFile(file, []byte("#!/bin/sh\necho docker-test\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(file, filepath.Join(root, "hop", "docker")); err != nil {
		t.Fatal(err)
	}
	runtime := filepath.Join(root, "bin", "docker")
	if err := os.Symlink(filepath.Join(root, "hop", "docker"), runtime); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultContainerConfig()
	cfg.local = true
	cfg.Runtime = runtime
	if _, err := NewContainer(cfg, nil); err != nil {
		t.Fatalf("control: private directories along the links must pass: %v", err)
	}
	hop := filepath.Join(root, "hop")
	if err := os.Chmod(hop, 0o777); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(hop, 0o700) })
	if _, err := NewContainer(cfg, nil); err == nil {
		t.Fatal("local mode accepted a runtime reached through a link whose directory is world-writable without the sticky bit")
	}
}

// TestLocalRuntimeRefusesALinkBeforeDotDot: the runtime path passes through a
// link and then "..", writable/link/../../runtime/docker. Cleaning the text
// first would drop "writable", the directory that holds the link, although any
// user who can write it could point the link elsewhere. Mutated field: the mode
// of "writable" (0777 instead of 0700).
func TestLocalRuntimeRefusesALinkBeforeDotDot(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"writable", "safe/deep/sub", "safe/runtime", "runtime"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range []string{"safe/runtime/docker", "runtime/docker"} {
		if err := os.WriteFile(filepath.Join(root, d), []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(root, "safe/deep/sub"), filepath.Join(root, "writable/link")); err != nil {
		t.Fatal(err)
	}
	raw := filepath.Join(root, "writable") + "/link/../../runtime/docker"
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": raw, "HOME": root}
	cfg, err := Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewContainer(cfg, nil); err != nil {
		t.Fatalf("control: a private link path must pass: %v", err)
	}
	if err := os.Chmod(filepath.Join(root, "writable"), 0o777); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "writable"), 0o700) })
	if _, err := NewContainer(cfg, nil); err == nil {
		t.Fatal("local mode accepted a runtime path through a world-writable link parent that dot-dot hides")
	}
}

// TestLocalRuntimeAcceptsAStickyTmp: a runtime under a sticky, world-writable
// directory such as /tmp is accepted; the sticky bit stops other users from
// replacing what is in it.
func TestLocalRuntimeAcceptsAStickyTmp(t *testing.T) {
	root := t.TempDir()
	tmp := filepath.Join(root, "tmp")
	if err := os.Mkdir(tmp, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(tmp, 0o777|os.ModeSticky); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(tmp, 0o700) })
	own := filepath.Join(tmp, "own")
	if err := os.Mkdir(own, 0o700); err != nil {
		t.Fatal(err)
	}
	runtime := filepath.Join(own, "docker")
	if err := os.WriteFile(runtime, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := DefaultContainerConfig()
	cfg.local = true
	cfg.Runtime = runtime
	if _, err := NewContainer(cfg, nil); err != nil {
		t.Fatalf("a runtime under a sticky world-writable directory was refused: %v", err)
	}
}
