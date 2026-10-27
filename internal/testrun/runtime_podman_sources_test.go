package testrun

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// systemConf points the system containers directory at a new directory and returns it.
func systemConf(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	saved := podmanSystemConfs
	podmanSystemConfs = []string{filepath.Join(dir, "containers.conf")}
	t.Cleanup(func() { podmanSystemConfs = saved })
	return dir
}

const (
	farConf = "[engine.service_destinations.far]\nuri = \"tcp://%s:2375\"\n"
	farJSON = `{"Connection":{"Connections":{"far":{"URI":"tcp://%s:2375"}}}}`
)

// podmanSourceFiles are the places podman reads a destination from in one configuration
// directory, relative to it.
var podmanSourceFiles = []string{
	"containers.conf",
	"containers.rootless.conf",
	"containers.rootful.conf",
	"containers.conf.d/10.conf",
	"containers.rootless.conf.d/10.conf",
	"containers.rootless.conf.d/1000/10.conf",
	"containers.rootful.conf.d/10.conf",
	"podman-connections.json",
}

func writeSource(t *testing.T, dir, rel, host string) {
	t.Helper()
	body := farConf
	if strings.HasSuffix(rel, ".json") {
		body = farJSON
	}
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(body, "%s", host)), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestLocalReadsEveryPodmanSource: a destination outside the network is refused in every file podman
// reads, in the vendor directory, the system directory and the user's directory: the main file and its
// rootless and rootful variants, the drop-ins of each, a per-user-id drop-in, and the connections file.
// Mutated field: the file that holds the destination; the same file with 10.0.0.5 must load, so the
// refusal is the address and not the place.
func TestLocalReadsEveryPodmanSource(t *testing.T) {
	for _, where := range []string{"vendor", "system", "user"} {
		for _, rel := range podmanSourceFiles {
			t.Run(where+"/"+rel, func(t *testing.T) {
				vendor, system, home := t.TempDir(), t.TempDir(), t.TempDir()
				saved := podmanSystemConfs
				podmanSystemConfs = []string{filepath.Join(vendor, "containers.conf"), filepath.Join(system, "containers.conf")}
				defer func() { podmanSystemConfs = saved }()
				dir := map[string]string{"vendor": vendor, "system": system, "user": filepath.Join(home, ".config", "containers")}[where]
				writeSource(t, dir, rel, "203.0.113.8")
				err := loadPodman(home)
				if err == nil || !strings.Contains(err.Error(), "203.0.113.8") {
					t.Fatalf("local mode loaded a podman destination 203.0.113.8 in %s/%s: %v", where, rel, err)
				}
				writeSource(t, dir, rel, "10.0.0.5")
				if err := loadPodman(home); err != nil {
					t.Fatalf("a podman destination 10.0.0.5 in %s/%s was refused: %v", where, rel, err)
				}
			})
		}
	}
}

// TestLocalRefusesAPodmanDirectoryItCannotList: a configuration or drop-in directory that exists and
// cannot be listed is refused by name, not read as empty. A regular file where the directory should be
// cannot be listed, whatever user runs the test. Mutated field: the kind of the path.
func TestLocalRefusesAPodmanDirectoryItCannotList(t *testing.T) {
	for _, rel := range []string{"containers.conf.d", "containers.rootless.conf.d", "containers.rootful.conf.d"} {
		t.Run(rel, func(t *testing.T) {
			dir := systemConf(t)
			path := filepath.Join(dir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("not a directory"), 0o600); err != nil {
				t.Fatal(err)
			}
			err := loadPodman(t.TempDir())
			var ce *ConfigError
			if !errors.As(err, &ce) || !strings.Contains(ce.Error(), "cannot be listed") {
				t.Fatalf("local mode did not refuse the unlistable %s: %v", rel, err)
			}
		})
	}
}

// TestLocalRefusesAPodmanDropInDirectoryWithoutPermission: a drop-in directory the user cannot open is
// refused, not read as empty. Mutated field: the mode of the directory. A superuser opens any
// directory, so the test is skipped for one; TestLocalRefusesAPodmanDirectoryItCannotList runs for every user.
func TestLocalRefusesAPodmanDropInDirectoryWithoutPermission(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("a superuser opens a directory of mode 0; the unlistable-directory test runs for every user")
	}
	dir := systemConf(t)
	drop := filepath.Join(dir, "containers.conf.d")
	if err := os.Mkdir(drop, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(drop, "10.conf"), []byte(strings.ReplaceAll(farConf, "%s", "203.0.113.8")), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(drop, 0); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(drop, 0o700)
	err := loadPodman(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "cannot be listed") {
		t.Fatalf("local mode read an unreadable drop-in directory as empty: %v", err)
	}
}

// TestLocalRefusesAPodmanSystemFileItCannotJudge: the system files are judged as the user's is. A
// system containers.conf that is not TOML, or cannot be read, is refused by name.
// Mutated field: the content or kind of the system file.
func TestLocalRefusesAPodmanSystemFileItCannotJudge(t *testing.T) {
	dir := systemConf(t)
	conf := filepath.Join(dir, "containers.conf")
	if err := os.WriteFile(conf, []byte("[engine\nremote = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := loadPodman(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "TOML") {
		t.Fatalf("local mode loaded a system file that is not TOML: %v", err)
	}
	if err := os.Remove(conf); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(conf, 0o700); err != nil {
		t.Fatal(err)
	}
	err = loadPodman(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "cannot be read") {
		t.Fatalf("local mode loaded a system file it cannot read: %v", err)
	}
}

// TestLocalRefusesAPodmanTableSetInTwoCases: engine and Engine are one table to podman's decoder, so a
// file that sets both holds two answers and is refused. Mutated field: the case of the second table name.
func TestLocalRefusesAPodmanTableSetInTwoCases(t *testing.T) {
	systemConf(t)
	body := "[engine]\nremote = false\n[Engine]\nremote = true\n"
	err := loadPodman(confHome(t, body))
	if err == nil || !strings.Contains(err.Error(), "more than once") {
		t.Fatalf("local mode loaded a table set in two cases: %v", err)
	}
}

// TestLocalRefusesWhenHomeIsUnset: with HOME unset the runtime finds the user's files under the home of
// the passwd entry, which Proven does not read, so local mode refuses and names HOME.
// Mutated field: HOME, set or not.
func TestLocalRefusesWhenHomeIsUnset(t *testing.T) {
	systemConf(t)
	env := map[string]string{"PROVEN_NETWORK": "local", "PROVEN_TEST_RUNTIME": "podman"}
	_, err := Load(func(k string) string { return env[k] })
	if err == nil || !strings.Contains(err.Error(), "HOME") {
		t.Fatalf("local mode loaded with HOME unset: %v", err)
	}
	env["HOME"] = t.TempDir()
	if _, err := Load(func(k string) string { return env[k] }); err != nil {
		t.Fatalf("local mode refused with HOME set: %v", err)
	}
}

// TestPodmanConfigVariablesDoNotReachTheRuntime: the variables that move podman's configuration are set
// in Proven's own environment and are absent from the runtime process, so the files Proven judges are the
// files the runtime reads. Mutated field: runtimeEnvNames; adding one of them turns this red.
func TestPodmanConfigVariablesDoNotReachTheRuntime(t *testing.T) {
	names := []string{"CONTAINERS_CONF", "CONTAINERS_CONF_OVERRIDE", "CONTAINER_HOST", "XDG_CONFIG_HOME", "PODMAN_CONNECTIONS_CONF", "CONTAINERS_REGISTRIES_CONF"}
	for _, name := range names {
		t.Setenv(name, "/elsewhere")
	}
	t.Setenv("HOME", "/home/prover")
	text, err := commandOutput(context.Background(), []string{"/usr/bin/env"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "HOME=/home/prover") {
		t.Fatalf("positive control: the runtime process did not get HOME: %q", text)
	}
	for _, name := range names {
		if strings.Contains(text, name+"=") {
			t.Errorf("%s reached the runtime process", name)
		}
	}
}
