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
