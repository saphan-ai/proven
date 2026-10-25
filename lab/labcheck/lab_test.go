package labcheck

// Review checks for security-piece-7 (lab and images), read against main at 177bba6.
// Each check is a function over file text. Every test runs the check on the real file
// and on one mutated copy (the negative control), so a check that cannot fail is caught.

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "lab", "compose.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repository root with lab/compose.yaml not found")
		}
		dir = parent
	}
}

func read(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// serviceBlock returns the lines of one top-level Compose service.
func serviceBlock(compose, name string) string {
	var out []string
	in := false
	for _, line := range strings.Split(compose, "\n") {
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "   ") && strings.HasSuffix(strings.TrimSpace(line), ":") {
			in = strings.TrimSpace(line) == name+":"
			continue
		}
		if !strings.HasPrefix(line, " ") && line != "" {
			in = false
		}
		if in {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

var portEntry = regexp.MustCompile(`^\s*-\s*"?([^"]+)"?\s*$`)

// checkLoopbackPorts: every published port names 127.0.0.1 literally, or names
// ${LAB_BIND} and a lab script refuses a LAB_BIND that is not 127.0.0.1.
func checkLoopbackPorts(compose, scripts string) []string {
	guard := regexp.MustCompile(`case\s+"?\$\{?LAB_BIND\}?"?\s+in`).MatchString(scripts)
	var bad []string
	lines := strings.Split(compose, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "ports:" {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			m := portEntry.FindStringSubmatch(lines[j])
			if m == nil {
				break
			}
			v := m[1]
			switch {
			case strings.HasPrefix(v, "127.0.0.1:"):
			case strings.HasPrefix(v, "${LAB_BIND}:") && guard:
			default:
				bad = append(bad, v)
			}
		}
	}
	return bad
}

func TestLabPublishesOnlyOnLoopback(t *testing.T) {
	compose := read(t, "lab/compose.yaml")
	scripts := read(t, "lab/scripts/lib.sh") + read(t, "lab/scripts/up.sh")
	if bad := checkLoopbackPorts(compose, scripts); len(bad) > 0 {
		t.Errorf("published ports not fixed to 127.0.0.1 and no script refuses another LAB_BIND: %v", bad)
	}
	// Negative control: a literal wildcard binding must be refused even with a guard present.
	mut := strings.Replace(compose, `"${LAB_BIND}:${GITEA_HTTP_PORT}`, `"0.0.0.0:${GITEA_HTTP_PORT}`, 1)
	if len(checkLoopbackPorts(mut, `case "$LAB_BIND" in`)) == 0 {
		t.Error("negative control: check accepted a 0.0.0.0 binding")
	}
}

// checkNoHostSocket: no volume, mount or -v names docker.sock.
func checkNoHostSocket(texts map[string]string) []string {
	mount := regexp.MustCompile(`(-v\s|--volume|--mount|^\s*-\s+[^#]*:)[^\n]*docker\.sock`)
	var bad []string
	for name, text := range texts {
		for _, line := range strings.Split(text, "\n") {
			if strings.Contains(line, "--host=unix:///var/run/docker.sock") {
				continue // dind's own listener inside its container, not a mount
			}
			if mount.MatchString(line) {
				bad = append(bad, name+": "+strings.TrimSpace(line))
			}
		}
	}
	return bad
}

func labTexts(t *testing.T) map[string]string {
	texts := map[string]string{
		"lab/compose.yaml":       read(t, "lab/compose.yaml"),
		"lab/runner/config.yaml": read(t, "lab/runner/config.yaml"),
		"lab/workflow/proven.yml": read(t, "lab/workflow/proven.yml"),
	}
	entries, err := os.ReadDir(filepath.Join(repoRoot(t), "lab", "scripts"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		texts["lab/scripts/"+e.Name()] = read(t, "lab/scripts/"+e.Name())
	}
	return texts
}

func TestNoContainerReceivesTheHostDockerSocket(t *testing.T) {
	texts := labTexts(t)
	if bad := checkNoHostSocket(texts); len(bad) > 0 {
		t.Errorf("host Docker socket reaches a container: %v", bad)
	}
	mut := map[string]string{"compose": strings.Replace(texts["lab/compose.yaml"],
		"      - runner-data:/data", "      - runner-data:/data\n      - /var/run/docker.sock:/var/run/docker.sock", 1)}
	if len(checkNoHostSocket(mut)) == 0 {
		t.Error("negative control: check accepted a docker.sock mount")
	}
}

// checkOnlyDindPrivileged: privileged: true appears in the dind service and nowhere else.
func checkOnlyDindPrivileged(compose, runnerCfg string) []string {
	var bad []string
	for _, svc := range []string{"gitea", "runner", "seed"} {
		if strings.Contains(serviceBlock(compose, svc), "privileged: true") {
			bad = append(bad, svc)
		}
	}
	if !strings.Contains(runnerCfg, "privileged: false") || strings.Contains(runnerCfg, "--privileged") {
		bad = append(bad, "runner job containers")
	}
	return bad
}

func TestOnlyDindIsPrivileged(t *testing.T) {
	compose, cfg := read(t, "lab/compose.yaml"), read(t, "lab/runner/config.yaml")
	if bad := checkOnlyDindPrivileged(compose, cfg); len(bad) > 0 {
		t.Errorf("privileged beyond dind: %v", bad)
	}
	if len(checkOnlyDindPrivileged(compose, strings.Replace(cfg, "privileged: false", "privileged: true", 1))) == 0 {
		t.Error("negative control: check accepted privileged job containers")
	}
}

// checkSeedNotRoot: the seed image we build is never started as root by the lab.
func checkSeedNotRoot(compose, scripts string) []string {
	var bad []string
	if regexp.MustCompile(`(?m)^\s*user:\s*"?(0|0:0|root)(:\d+)?"?\s*$`).MatchString(serviceBlock(compose, "seed")) {
		bad = append(bad, "compose.yaml seed: user 0:0")
	}
	if n := strings.Count(scripts, "--user 0:0"); n > 0 {
		bad = append(bad, "scripts: --user 0:0 x"+strconv.Itoa(n))
	}
	return bad
}

func TestSeedRunsAsNonRoot(t *testing.T) {
	compose := read(t, "lab/compose.yaml")
	scripts := read(t, "lab/scripts/up.sh") + read(t, "lab/scripts/gitea-api.sh")
	if bad := checkSeedNotRoot(compose, scripts); len(bad) > 0 {
		t.Errorf("the seed image is started as root: %v", bad)
	}
	clean := strings.Replace(compose, `    user: "0:0"`+"\n", "", 1)
	if len(checkSeedNotRoot(clean, "")) != 0 {
		t.Error("check refuses a compose file with no root user")
	}
	if len(checkSeedNotRoot(clean+"\n", "--user 0:0")) == 0 {
		t.Error("negative control: check accepted --user 0:0")
	}
}

// finalUser returns the last USER of the last stage.
func finalUser(dockerfile string) string {
	user := ""
	for _, line := range strings.Split(dockerfile, "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 && strings.EqualFold(f[0], "FROM") {
			user = ""
		}
		if len(f) >= 2 && strings.EqualFold(f[0], "USER") {
			user = f[1]
		}
	}
	return user
}

func rootUser(u string) bool {
	return u == "" || u == "root" || u == "0" || strings.HasPrefix(u, "0:") || strings.HasPrefix(u, "root:")
}

func TestImagesWeBuildEndAsNonRoot(t *testing.T) {
	for _, f := range []string{"Dockerfile", "lab/seed/Dockerfile"} {
		if u := finalUser(read(t, f)); rootUser(u) {
			t.Errorf("%s: final user %q is root", f, u)
		}
	}
	if !rootUser(finalUser(strings.Replace(read(t, "Dockerfile"), "USER proven", "USER root", 1))) {
		t.Error("negative control: check accepted USER root")
	}
}

// checkScanFailsClosed: make scan never exits 0 having scanned nothing.
func checkScanFailsClosed(scan string) bool {
	return !regexp.MustCompile(`(?m)^\s*exit 0\s*$`).MatchString(scan)
}

func TestScanRefusesWhenNothingIsScanned(t *testing.T) {
	scan := read(t, "lab/scripts/scan.sh")
	if !checkScanFailsClosed(scan) {
		t.Error("lab/scripts/scan.sh exits 0 when docker scout is missing: make scan is green with no scan")
	}
	if checkScanFailsClosed("if x; then\n  exit 0\nfi\n") {
		t.Error("negative control: check accepted exit 0")
	}
}

// checkDockerignore: the fleet work area, the trace directory and the fleet record stay out of the context.
func checkDockerignore(ignore string) []string {
	have := map[string]bool{}
	for _, l := range strings.Split(ignore, "\n") {
		have[strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(l), "/"), "/")] = true
	}
	var missing []string
	for _, want := range []string{".saphan", "logs", "hackathon", ".env"} {
		if !have[want] && !have["**/"+want] {
			missing = append(missing, want)
		}
	}
	return missing
}

func TestDockerignoreKeepsWorkAreaOut(t *testing.T) {
	if m := checkDockerignore(read(t, ".dockerignore")); len(m) > 0 {
		t.Errorf(".dockerignore does not exclude %v", m)
	}
	if len(checkDockerignore(".saphan\nlogs\nhackathon\n")) == 0 {
		t.Error("negative control: check accepted a file without .env")
	}
}

// checkDefaultsMatch: the Dockerfile ARG defaults equal lab/versions.env, as the Dockerfile says.
func checkDefaultsMatch(dockerfile, versions string, keys ...string) []string {
	var bad []string
	for _, k := range keys {
		arg := regexp.MustCompile(`(?m)^ARG ` + k + `=(\S+)`).FindStringSubmatch(dockerfile)
		env := regexp.MustCompile(`(?m)^` + k + `=(\S+)`).FindStringSubmatch(versions)
		if arg == nil || env == nil || arg[1] != env[1] {
			bad = append(bad, k)
		}
	}
	return bad
}

func TestDockerfileDefaultsMatchVersions(t *testing.T) {
	versions := read(t, "lab/versions.env")
	if bad := checkDefaultsMatch(read(t, "Dockerfile"), versions, "GOLANG_IMAGE", "RUNTIME_IMAGE"); len(bad) > 0 {
		t.Errorf("Dockerfile defaults differ from lab/versions.env: %v", bad)
	}
	if bad := checkDefaultsMatch(read(t, "lab/seed/Dockerfile"), versions, "RUNTIME_IMAGE"); len(bad) > 0 {
		t.Errorf("lab/seed/Dockerfile defaults differ from lab/versions.env: %v", bad)
	}
	if len(checkDefaultsMatch("ARG X=a\n", "X=b\n", "X")) == 0 {
		t.Error("negative control: check accepted a mismatch")
	}
}

// checkNoCommittedPassword: no committed lab file carries a password value, and the
// Compose default for the admin password is empty.
func checkNoCommittedPassword(texts map[string]string) []string {
	val := regexp.MustCompile(`(?m)^[ \t]*[A-Z_]*(PASSWORD|TOKEN|_KEY)[ \t]*=[ \t]*\S+`)
	var bad []string
	for name, text := range texts {
		if val.MatchString(text) {
			bad = append(bad, name)
		}
	}
	if c, ok := texts["lab/compose.yaml"]; ok && !strings.Contains(c, "LAB_ADMIN_PASSWORD: ${LAB_ADMIN_PASSWORD:-}") {
		bad = append(bad, "compose default")
	}
	return bad
}

func TestLabPasswordHasNoCommittedDefault(t *testing.T) {
	texts := map[string]string{
		"lab/versions.env":  read(t, "lab/versions.env"),
		"lab/.env.example":  read(t, "lab/.env.example"),
		"lab/compose.yaml":  read(t, "lab/compose.yaml"),
	}
	if bad := checkNoCommittedPassword(texts); len(bad) > 0 {
		t.Errorf("a committed default: %v", bad)
	}
	if !strings.Contains(read(t, "lab/scripts/up.sh"), "/dev/urandom") {
		t.Error("up.sh does not generate the admin password from /dev/urandom")
	}
	texts["lab/versions.env"] += "\nLAB_ADMIN_PASSWORD=changeme\n"
	if len(checkNoCommittedPassword(texts)) == 0 {
		t.Error("negative control: check accepted a committed password")
	}
}
