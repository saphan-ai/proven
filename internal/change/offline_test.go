package change

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// partialClone returns a clone of a one-file repository that holds the commit and no blob: reading
// f.txt makes git ask the origin for it. The origin is a local path, so no byte leaves the host.
func partialClone(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	src := filepath.Join(root, "src")
	git(t, root, "init", src)
	git(t, src, "config", "uploadpack.allowFilter", "true")
	git(t, src, "config", "uploadpack.allowAnySHA1InWant", "true")
	write(t, filepath.Join(src, "f.txt"), "hello")
	git(t, src, "add", "f.txt")
	git(t, src, "commit", "-m", "[change] offline: first")
	clone := filepath.Join(root, "clone")
	out, err := exec.Command("git", "clone", "--no-checkout", "--filter=blob:none", "file://"+src, clone).CombinedOutput()
	if err != nil {
		t.Fatalf("clone: %v\n%s", err, out)
	}
	return clone
}

// TestOfflineGitNamesAMissingObjectInsteadOfFetching proves ADR-0034 for git: in a partial clone the
// blob of f.txt is on the origin only. Online, git fetches it. Offline, the read fails by name and
// the setting that caused it, and the blob is still not in the clone afterwards.
// Mutated field: the guard of GitCommand; with it removed the offline read returns "hello".
func TestOfflineGitNamesAMissingObjectInsteadOfFetching(t *testing.T) {
	clone := partialClone(t)
	defer UseOffline(UseOffline(true))
	_, _, err := Blob(clone, "HEAD", "f.txt")
	if err == nil || !strings.Contains(err.Error(), "PROVEN_NETWORK=local") || !strings.Contains(err.Error(), "partial clone") {
		t.Fatalf("offline read of a missing object: %v", err)
	}
	if _, err := gitArchive(clone, "HEAD"); err == nil || !strings.Contains(err.Error(), "PROVEN_NETWORK=local") {
		t.Fatalf("offline archive of a missing object: %v", err)
	}
	// Still missing: nothing fetched it behind the refusal.
	out, _ := exec.Command("git", "-C", clone, "rev-list", "--objects", "--missing=print", "HEAD").Output()
	if !strings.Contains(string(out), "?") {
		t.Fatalf("the blob is no longer missing: %s", out)
	}

	// The same read, online, is what the guard prevents: git fetches the blob.
	UseOffline(false)
	body, ok, err := Blob(clone, "HEAD", "f.txt")
	if err != nil || !ok || string(body) != "hello" {
		t.Fatalf("online read: %q %v %v", body, ok, err)
	}
}

// TestOfflineGitCommandCarriesBothSwitches proves what GitCommand sets, so a change of git's switch
// names is seen here and not as a fetch in the field.
func TestOfflineGitCommandCarriesBothSwitches(t *testing.T) {
	defer UseOffline(UseOffline(false))
	const marker = "hardening-marker-offline"
	t.Setenv("METERLESS_RUNTIME_TOKEN", marker)
	online := GitCommand("status")
	onlineEnv := strings.Join(online.Env, "\n")
	if strings.Contains(strings.Join(online.Args, " "), "protocol.allow") || strings.Contains(onlineEnv, marker) || !strings.Contains(onlineEnv, "PATH=") {
		t.Fatalf("online command: args %v env %v", online.Args, online.Env)
	}
	UseOffline(true)
	c := GitCommand("status")
	joined := strings.Join(c.Args, " ")
	env := strings.Join(c.Env, "\n")
	if !strings.Contains(joined, "protocol.allow=never") || !strings.Contains(env, "GIT_NO_LAZY_FETCH=1") || !strings.Contains(env, "GIT_TERMINAL_PROMPT=0") || strings.Contains(env, marker) {
		t.Fatalf("offline command: args %v env %s", c.Args, env)
	}
}
