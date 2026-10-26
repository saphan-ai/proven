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
