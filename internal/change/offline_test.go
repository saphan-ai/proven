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
	if _, err := ProcessPolicy().gitArchive(clone, "HEAD"); err == nil || !strings.Contains(err.Error(), "PROVEN_NETWORK=local") {
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
	if !strings.Contains(joined, "protocol.allow=never") || !strings.Contains(env, "GIT_NO_LAZY_FETCH=1") || !strings.Contains(env, "GIT_TERMINAL_PROMPT=0") || !strings.Contains(env, "GIT_ATTR_NOSYSTEM=1") || !strings.Contains(joined, "log.showSignature=false") || strings.Contains(env, marker) {
		t.Fatalf("offline command: args %v env %s", c.Args, env)
	}
	for _, name := range transports {
		if !strings.Contains(joined, "protocol."+name+".allow=never") {
			t.Fatalf("offline command leaves transport %s to the configuration: %v", name, c.Args)
		}
	}
}

// TestOfflineGitRefusesATransportTheRepositoryConfigAllows proves ADR-0034 for git when a configuration
// file Proven does not own (the repository's .git/config or the user's ~/.gitconfig, which git reads
// through HOME) sets a per-protocol policy. protocol.<name>.allow wins over protocol.allow, so
// "-c protocol.allow=never" alone does not close that transport. Every offline git command also
// passes "-c protocol.<name>.allow=never" for file, git, http, https, ssh and ext.
// Mutated field: protocol.file.allow in the clone's own config.
func TestOfflineGitRefusesATransportTheRepositoryConfigAllows(t *testing.T) {
	clone := partialClone(t)
	git(t, clone, "config", "protocol.file.allow", "always")
	defer UseOffline(UseOffline(true))
	out, err := GitCommand("-C", clone, "ls-remote", "origin").CombinedOutput()
	if err == nil {
		t.Fatalf("offline git reached the origin through a per-protocol policy in the repository config:\n%s", out)
	}
}

// TestOfflineBlobNamesTheModeOnEveryGit proves the refusal is named on a git that honours
// GIT_NO_LAZY_FETCH quietly: "git show" then prints only "fatal: bad object <rev>:<path>", with no
// "lazy fetching disabled" line. The reader is still told that the mode refused the fetch.
func TestOfflineBlobNamesTheModeOnEveryGit(t *testing.T) {
	clone := partialClone(t)
	defer UseOffline(UseOffline(true))
	_, _, err := Blob(clone, "HEAD", "f.txt")
	if err == nil || !strings.Contains(err.Error(), "PROVEN_NETWORK=local") {
		t.Fatalf("offline read of a missing blob is not named: %v", err)
	}
}

// TestOfflineMessageNamesAQuietGitInAPartialClone feeds the words of a git that says only "bad object"
// and proves the sentence comes from the repository being a partial clone, not from git's words. A
// full clone with the same words is not named, and an absent path is still an answer, not a refusal.
// Mutated field: the isPartialClone check of offlineMessage.
func TestOfflineMessageNamesAQuietGitInAPartialClone(t *testing.T) {
	clone := partialClone(t)
	defer UseOffline(UseOffline(true))
	quiet := "fatal: bad object HEAD:./f.txt"
	partial := GitCommand("-C", clone, "show", "HEAD:./f.txt")
	if got := ProcessPolicy().offlineMessage(partial, quiet); !strings.Contains(got, "PROVEN_NETWORK=local") || !strings.Contains(got, quiet) {
		t.Fatalf("a partial clone's quiet refusal is not named: %q", got)
	}
	if got := ProcessPolicy().offlineMessage(partial, "fatal: path 'f.txt' does not exist in 'HEAD'"); strings.Contains(got, "PROVEN_NETWORK") {
		t.Fatalf("an absent path was named as a refused fetch: %q", got)
	}
	full := t.TempDir()
	git(t, full, "init", full)
	if got := ProcessPolicy().offlineMessage(GitCommand("-C", full, "show", "HEAD:./f.txt"), quiet); strings.Contains(got, "PROVEN_NETWORK") {
		t.Fatalf("a full clone was named as a partial one: %q", got)
	}
}
