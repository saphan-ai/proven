package lab

// Round 5 (ADR-0052, docs/LAB.md "How a credential travels"): no credential is in Compose or in a container's
// environment. make up writes them to the seed's state volume over stdin at mode 0600, the seed stores each one as an
// Actions secret and removes the file. Every control below mutates exactly one thing and names it.

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// sentinel is the value a test gives a credential's name, so a leak is a plain substring search.
func sentinel(name string) string { return "sentinel-" + name + "-0123456789" }

// secretNames is LAB_SECRET_KEYS of lab/versions.env.
func secretNames(t *testing.T) []string {
	t.Helper()
	names := strings.Fields(versionsList(t, "LAB_SECRET_KEYS"))
	if len(names) == 0 {
		t.Fatal("LAB_SECRET_KEYS is empty")
	}
	return names
}

// composeNamesAny returns the names of want that the compose text mentions as an environment entry (NAME: or NAME=),
// as an interpolation (${NAME), or in a list item (- NAME or - NAME=...), in any service.
func composeNamesAny(compose string, want []string) []string {
	var hit []string
	for _, n := range want {
		re := regexp.MustCompile(`(?m)^\s*(- )?` + n + `\s*[:=]|\$\{` + n + `\b|^\s*- ` + n + `\s*$`)
		if re.MatchString(compose) {
			hit = append(hit, n)
		}
	}
	return hit
}

func TestSecretsNoCredentialNameIsInCompose(t *testing.T) {
	names := append(secretNames(t), "ZETARIS_ADMIN_API_KEY")
	compose := read(t, "lab", "compose.yaml")
	if hit := composeNamesAny(compose, names); len(hit) > 0 {
		t.Errorf("lab/compose.yaml names a credential, so docker compose config and docker inspect would show its value: %v", hit)
	}
	// Negative control, mutating one field: put one credential name back into the seed's environment block.
	back := strings.Replace(compose, "      PROVEN_GENERATOR_MODE:\n", "      PROVEN_GENERATOR_MODE:\n      JUDGES_GATEWAY_ACCESS_TOKEN:\n", 1)
	if back == compose {
		t.Fatal("the mutation target is not in lab/compose.yaml")
	}
	if hit := composeNamesAny(back, names); len(hit) != 1 || hit[0] != "JUDGES_GATEWAY_ACCESS_TOKEN" {
		t.Errorf("the check did not go red for a credential put back into Compose: %v", hit)
	}
	// And the Zetaris administrator's key, which the register service now reads from its standard input.
	back = strings.Replace(compose, "      ZETARIS_PROVEN_USER:\n", "      ZETARIS_PROVEN_USER:\n      ZETARIS_ADMIN_API_KEY:\n", 1)
	if back == compose || len(composeNamesAny(back, names)) != 1 {
		t.Error("the check did not go red for the administrator key put back into Compose")
	}
}

func TestSecretsTheListsAreHeldTogether(t *testing.T) {
	live := strings.Fields(versionsList(t, "LAB_LIVE_KEYS"))
	secret := secretNames(t)
	isSecret := map[string]bool{}
	for _, s := range secret {
		isSecret[s] = true
		found := false
		for _, l := range live {
			found = found || l == s
		}
		if !found {
			t.Errorf("%s is a credential but LAB_LIVE_KEYS does not list it, so it would never be stored", s)
		}
	}
	// The seed's environment names exactly the settings: LAB_LIVE_KEYS minus the credentials.
	compose := read(t, "lab", "compose.yaml")
	seed := compose[strings.Index(compose, "  seed:"):]
	seed = seed[:strings.Index(seed, "    volumes:")]
	var got, want []string
	for _, m := range regexp.MustCompile(`(?m)^      (PROVEN_[A-Z0-9_]*|ZETARIS_[A-Z0-9_]*|METERLESS_[A-Z0-9_]*|JUDGES_GATEWAY_URL):$`).FindAllStringSubmatch(seed, -1) {
		got = append(got, m[1])
	}
	for _, l := range live {
		if !isSecret[l] {
			want = append(want, l)
		}
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("the seed's environment names %v, want the settings %v", got, want)
	}
}

// fakeDockerRecording stands in for docker: it records every argument, one per line, to $FAKE_DIR/argv, runs the
// script of a `-c` (the seed's `sh -c`) against $FAKE_DIR/state instead of /state, and leaves its standard input to
// that script. Anything else gets the input recorded to $FAKE_DIR/stdin.
const fakeDockerRecording = `#!/bin/sh
printf '%s\n' "$@" >> "$FAKE_DIR/argv"
mkdir -p "$FAKE_DIR/state"
prev=
for a in "$@"; do
  if [ "$prev" = "-c" ]; then
    script=$(printf '%s' "$a" | sed "s|/state/|$FAKE_DIR/state/|g")
    exec sh -c "$script"
  fi
  prev=$a
done
cat > "$FAKE_DIR/stdin"
`

// writeSecrets runs lab_write_secrets of the tree with a sentinel in every credential's name, and returns the
// recorded arguments, the file the seed's side wrote, and the file's mode.
func writeSecrets(t *testing.T, tree string, empty ...string) (argv, file string, mode os.FileMode) {
	t.Helper()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	writeExec(t, bin, "docker", fakeDockerRecording)
	env := []string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "TMPDIR=" + tmp, "FAKE_DIR=" + tmp}
	for _, n := range secretNames(t) {
		v := sentinel(n)
		for _, e := range empty {
			if e == n {
				v = ""
			}
		}
		env = append(env, n+"="+v)
	}
	cmd := exec.Command("sh", "-c", `ROOT="$1"; export ROOT; . "$1/lab/scripts/lib.sh"; lab_write_secrets`, "sh", tree)
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("lab_write_secrets: %v\n%s", err, out)
	}
	a, _ := os.ReadFile(filepath.Join(tmp, "argv"))
	f := filepath.Join(tmp, "state", "secrets")
	b, err := os.ReadFile(f)
	if err != nil {
		t.Fatalf("the seed's side has no credentials file: %v", err)
	}
	info, _ := os.Stat(f)
	return string(a), string(b), info.Mode().Perm()
}

func TestSecretsTravelOnStandardInputIntoAFileOnlyTheSeedCanRead(t *testing.T) {
	tree := copyLab(t)
	argv, file, mode := writeSecrets(t, tree, "ZETARIS_API_KEY")
	for _, n := range secretNames(t) {
		if strings.Contains(argv, sentinel(n)) {
			t.Errorf("the value of %s is an argument of a process", n)
		}
		line := n + "=" + sentinel(n) + "\n"
		if n == "ZETARIS_API_KEY" {
			if strings.Contains(file, n+"=") {
				t.Errorf("an empty credential was written: %q", file)
			}
			continue
		}
		if !strings.Contains(file, line) {
			t.Errorf("the file does not hold %s: %q", n, file)
		}
	}
	if mode != 0o600 {
		t.Errorf("the credentials file has mode %o, want 600", mode)
	}
	// The mutated field is the argument list of the one docker call: a value appended to it is an argument.
	mut := copyLab(t)
	mutate(t, mut, "lab/scripts/lib.sh", `-c 'umask 077 && cat > /state/secrets'`, `-c "umask 077 && cat > /state/secrets # $lws_val"`)
	argv, _, _ = writeSecrets(t, mut)
	leaked := false
	for _, n := range secretNames(t) {
		leaked = leaked || strings.Contains(argv, sentinel(n))
	}
	if !leaked {
		t.Error("the argument check did not go red when a value was put on the command line")
	}
	// The mutated field is the mode: without the umask the file is not 0600.
	mut = copyLab(t)
	mutate(t, mut, "lab/scripts/lib.sh", `-c 'umask 077 && cat > /state/secrets'`, `-c 'umask 022 && cat > /state/secrets'`)
	if _, _, mode = writeSecrets(t, mut); mode == 0o600 {
		t.Error("the mode check did not go red when the umask was loosened")
	}
}

func TestSecretsAValueWithALineBreakIsRefusedBeforeAnythingIsWritten(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin")
	writeExec(t, bin, "docker", fakeDockerRecording)
	cmd := exec.Command("sh", "-c", `ROOT="$1"; export ROOT; . "$1/lab/scripts/lib.sh"; lab_write_secrets`, "sh", mustAbs(t, root))
	cmd.Env = []string{"PATH=" + bin + ":" + os.Getenv("PATH"), "HOME=" + tmp, "TMPDIR=" + tmp, "FAKE_DIR=" + tmp,
		"PROVEN_GENERATOR_KEY=line one\nMETERLESS_RUNTIME_TOKEN=smuggled"}
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "PROVEN_GENERATOR_KEY holds a line break") {
		t.Errorf("a value with a line break was not refused by name: %v\n%s", err, out)
	}
	if _, serr := os.Stat(filepath.Join(tmp, "argv")); serr == nil {
		t.Error("docker was called although the value was refused")
	}
}

// runSeedUntilItStops runs the real seed.sh with the credentials file in place, and no Gitea behind it.
func runSeedUntilItStops(t *testing.T, tree string, env ...string) (string, error) {
	t.Helper()
	tmp := t.TempDir()
	secrets := secretsFile(t, "PROVEN_GENERATOR_KEY="+sentinel("PROVEN_GENERATOR_KEY")+"\n")
	cmd := exec.Command("sh", filepath.Join(tree, "lab", "seed", "seed.sh"))
	cmd.Env = cleanEnv(t, tmp, append([]string{"LAB_SECRETS_FILE=" + secrets}, env...)...)
	out, err := cmd.CombinedOutput()
	if _, serr := os.Stat(secrets); serr == nil {
		return string(out), errFileRemains
	}
	return string(out), err
}

var errFileRemains = &os.PathError{Op: "remove", Path: "the credentials file", Err: os.ErrExist}

func TestSecretsTheFileIsGoneAfterASeedThatFailedMidway(t *testing.T) {
	needTool(t, "curl")
	// The seed stops at its first check (a required setting is missing). The file made for it is gone.
	out, err := runSeedUntilItStops(t, mustAbs(t, root))
	if err == errFileRemains || err == nil {
		t.Fatalf("a seed that failed left the credentials file or did not fail: %v\n%s", err, out)
	}
	// The control mutates the trap: without it the file stays.
	tree := t.TempDir()
	for _, f := range []string{"seed.sh", "golive.sh", "render.sh"} {
		b := mustRead(t, filepath.Join(root, "lab", "seed", f))
		if err := os.MkdirAll(filepath.Join(tree, "lab", "seed"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tree, "lab", "seed", f), b, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mutate(t, tree, "lab/seed/seed.sh", `trap 'rm -f "$LAB_SECRETS_FILE"' EXIT`, `trap ':' EXIT`)
	if _, err := runSeedUntilItStops(t, tree); err != errFileRemains {
		t.Errorf("the control did not go red when the trap was removed: %v", err)
	}
}

func TestSecretsTheSeedStoresACredentialFromTheFileAndRemovesNothingItDidNotWrite(t *testing.T) {
	needTool(t, "curl")
	srv, calls, mu := stubGitea(t, http.StatusCreated)
	cmd := exec.Command("sh", "-c", `. "$ROOT/lab/seed/golive.sh"; ensure_live_secrets`)
	// Nothing is in the environment: the one credential is in the file only, so a seed reading the environment stores
	// nothing and this test goes red.
	cmd.Env = cleanEnv(t, t.TempDir(),
		"ROOT="+mustAbs(t, root), "GITEA_INTERNAL_URL="+srv.URL,
		"LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount", "API_TOKEN=x",
		"LAB_LIVE_KEYS=JUDGES_GATEWAY_ACCESS_TOKEN PROVEN_GENERATOR_MODEL", "LAB_SECRET_KEYS=JUDGES_GATEWAY_ACCESS_TOKEN",
		"PROVEN_GENERATOR_MODEL=some/model",
		"LAB_SECRETS_FILE="+secretsFile(t, "JUDGES_GATEWAY_ACCESS_TOKEN="+sentinel("JUDGES_GATEWAY_ACCESS_TOKEN")+"\n"))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	mu.Lock()
	got := map[string]string{}
	for _, c := range *calls {
		got[c.method+" "+c.name] = c.data
	}
	mu.Unlock()
	if got["PUT JUDGES_GATEWAY_ACCESS_TOKEN"] != sentinel("JUDGES_GATEWAY_ACCESS_TOKEN") {
		t.Errorf("the Access token was not stored from the file: %v", got)
	}
	if got["PUT PROVEN_GENERATOR_MODEL"] != "some/model" {
		t.Errorf("a plain setting was not stored from the environment: %v", got)
	}
	// A credential found only in the environment is not a credential of this lab: the environment holds none.
	cmd = exec.Command("sh", "-c", `. "$ROOT/lab/seed/golive.sh"; ensure_live_secrets`)
	cmd.Env = cleanEnv(t, t.TempDir(),
		"ROOT="+mustAbs(t, root), "GITEA_INTERNAL_URL="+srv.URL,
		"LAB_ADMIN_USER=proven", "LAB_DEMO_REPO=discount", "API_TOKEN=x",
		"LAB_LIVE_KEYS=JUDGES_GATEWAY_ACCESS_TOKEN", "LAB_SECRET_KEYS=JUDGES_GATEWAY_ACCESS_TOKEN",
		"JUDGES_GATEWAY_ACCESS_TOKEN=from-the-environment", "LAB_SECRETS_FILE=/nonexistent/secrets")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	mu.Lock()
	defer mu.Unlock()
	for _, c := range *calls {
		if c.data == "from-the-environment" {
			t.Error("a credential was read from the environment")
		}
	}
}
