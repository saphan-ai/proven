package labcheck

// Checks on the judges' lab and the images we build: the lab publishes only on
// loopback, no container gets the host Docker socket, only dind is privileged, the
// seed and the images run as a user that is not root, make scan refuses to pass when
// nothing was scanned, the build context is an allow-list, the Dockerfile defaults
// equal lab/versions.env, and no password is committed.
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

var (
	portEntry    = regexp.MustCompile(`^\s*-\s*(.*?)\s*$`)
	loopbackPort = regexp.MustCompile(`^"?(127(\.[0-9]{1,3}){3}|\[::1\]):[0-9$]`)
	hostNetwork  = regexp.MustCompile(`(?m)^\s*network_mode:\s*"?host"?\s*$`)
)

// checkLoopbackPorts reads lab/compose.yaml alone. The file must be loopback by
// itself, because `docker compose -f lab/compose.yaml up` does not run any lab
// script. Every published port names a literal loopback address (127.x.x.x or
// [::1]) in its own line: no variable, no empty host part, no long form, no inline
// list. A service on the host network publishes everything and is refused too.
func checkLoopbackPorts(compose string) []string {
	var bad []string
	if hostNetwork.MatchString(compose) {
		bad = append(bad, "network_mode: host")
	}
	lines := strings.Split(compose, "\n")
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if !strings.HasPrefix(t, "ports:") {
			continue
		}
		if rest := strings.TrimSpace(strings.TrimPrefix(t, "ports:")); rest != "" && !strings.HasPrefix(rest, "#") {
			bad = append(bad, t)
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if s := strings.TrimSpace(lines[j]); s == "" || strings.HasPrefix(s, "#") {
				continue
			}
			m := portEntry.FindStringSubmatch(lines[j])
			if m == nil {
				break
			}
			if !loopbackPort.MatchString(m[1]) {
				bad = append(bad, m[1])
			}
		}
	}
	return bad
}

func TestLabPublishesOnlyOnLoopback(t *testing.T) {
	compose := read(t, "lab/compose.yaml")
	if bad := checkLoopbackPorts(compose); len(bad) > 0 {
		t.Errorf("lab/compose.yaml publishes beyond a literal loopback address: %v", bad)
	}
	if !strings.Contains(compose, `"127.0.0.1:${GITEA_HTTP_PORT}:`) {
		t.Error("the check found no published Gitea port: it would pass an empty file")
	}
	if n := strings.Count(compose, "ports:"); n != 2 {
		t.Errorf("lab/compose.yaml has %d ports: blocks, want 2 (gitea, lgtm)", n)
	}
	// Negative controls: each mutates the one port line it names, and must turn red.
	for name, mut := range map[string]string{
		"wildcard literal":      strings.Replace(compose, `"127.0.0.1:${GITEA_HTTP_PORT}`, `"0.0.0.0:${GITEA_HTTP_PORT}`, 1),
		"interpolated address":  strings.Replace(compose, `"127.0.0.1:${GITEA_HTTP_PORT}`, `"${LAB_BIND}:${GITEA_HTTP_PORT}`, 1),
		"interpolated default":  strings.Replace(compose, `"127.0.0.1:${LGTM_OTLP_PORT}`, `"${LAB_BIND:-127.0.0.1}:${LGTM_OTLP_PORT}`, 1),
		"no host address":       strings.Replace(compose, `"127.0.0.1:${LGTM_GRAFANA_PORT}`, `"${LGTM_GRAFANA_PORT}`, 1),
		"private address":       strings.Replace(compose, `"127.0.0.1:${LGTM_OTLP_PORT}`, `"192.168.1.5:${LGTM_OTLP_PORT}`, 1),
		"lookalike address":     strings.Replace(compose, `"127.0.0.1:${LGTM_OTLP_PORT}`, `"128.0.0.1:${LGTM_OTLP_PORT}`, 1),
		"long form":             strings.Replace(compose, "    ports:\n", "    ports:\n      - target: 3000\n        published: 3000\n", 1),
		"inline list":           strings.Replace(compose, "    ports:\n", "    ports: [\"3000:3000\"]\n", 1),
		"host network":          strings.Replace(compose, "  dind:\n", "  dind:\n    network_mode: host\n", 1),
		"extra unprefixed port": strings.Replace(compose, "    ports:\n", "    ports:\n      - \"2375:2375\"\n", 1),
	} {
		if mut == compose {
			t.Errorf("negative control %q did not change the file", name)
			continue
		}
		if len(checkLoopbackPorts(mut)) == 0 {
			t.Errorf("negative control %q: check accepted it", name)
		}
	}
	// The other loopback spellings are accepted.
	ok := strings.Replace(compose, `"127.0.0.1:${GITEA_HTTP_PORT}`, `"[::1]:${GITEA_HTTP_PORT}`, 1)
	if bad := checkLoopbackPorts(ok); len(bad) > 0 {
		t.Errorf("check refused [::1]: %v", bad)
	}
}

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
		"lab/compose.yaml":        read(t, "lab/compose.yaml"),
		"lab/runner/config.yaml":  read(t, "lab/runner/config.yaml"),
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
// The file is either a deny-list that names each path, or an allow-list: the first
// rule is `*` and no later `!` rule re-admits a wanted path.
func checkDockerignore(ignore string) []string {
	var rules []string
	for _, l := range strings.Split(ignore, "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			rules = append(rules, l)
		}
	}
	allowList := len(rules) > 0 && rules[0] == "*"
	have := map[string]bool{}
	for _, l := range rules {
		have[strings.TrimSuffix(strings.TrimPrefix(l, "/"), "/")] = true
	}
	var missing []string
	for _, want := range []string{".saphan", ".provision", "logs", "hackathon", ".env"} {
		if allowList {
			if dockerignoreAdmits(rules, want) {
				missing = append(missing, want)
			}
			continue
		}
		if !have[want] && !have["**/"+want] {
			missing = append(missing, want)
		}
	}
	return missing
}

// dockerignoreAdmits reports whether the path is sent after the rules are applied in
// order, the last matching rule winning. A `!` rule for something under the path
// admits that part of it, so it counts as admitting the path.
func dockerignoreAdmits(rules []string, want string) bool {
	admitted := false
	for _, r := range rules {
		neg := strings.HasPrefix(r, "!")
		p := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(r, "!"), "/"), "/")
		hit := p == want || p == "**" || p == "**/"+want
		if ok, err := filepath.Match(p, want); err == nil && ok {
			hit = true
		}
		if neg && strings.HasPrefix(p, want+"/") {
			hit = true
		}
		if hit {
			admitted = neg
		}
	}
	return admitted
}

func TestDockerignoreKeepsWorkAreaOut(t *testing.T) {
	if m := checkDockerignore(read(t, ".dockerignore")); len(m) > 0 {
		t.Errorf(".dockerignore does not exclude %v", m)
	}
	if len(checkDockerignore(".saphan\nlogs\nhackathon\n")) == 0 {
		t.Error("negative control: check accepted a file without .env")
	}
	// Negative control: an allow-list that re-admits .saphan must be refused.
	allow := read(t, ".dockerignore")
	if m := checkDockerignore(allow + "\n!.saphan\n"); len(m) != 1 || m[0] != ".saphan" {
		t.Errorf("negative control: an allow-list with !.saphan gave %v, want [.saphan]", m)
	}
	// Negative control: an allow-list that re-admits only .provision must be refused.
	if m := checkDockerignore(allow + "\n!.provision/skills\n"); len(m) != 1 || m[0] != ".provision" {
		t.Errorf("negative control: an allow-list with !.provision/skills gave %v, want [.provision]", m)
	}
	// An allow-list must start with `*`: without it the file is read as a deny-list.
	if len(checkDockerignore(strings.Replace(allow, "\n*\n", "\n", 1))) == 0 {
		t.Error("negative control: check accepted an allow-list without its first * rule")
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
	// The password never enters the Compose environment: Compose renders every
	// interpolated value, so the file must not name the variable at all.
	if c, ok := texts["lab/compose.yaml"]; ok && strings.Contains(c, "LAB_ADMIN_PASSWORD") {
		bad = append(bad, "compose environment")
	}
	return bad
}

func TestLabPasswordHasNoCommittedDefault(t *testing.T) {
	texts := map[string]string{
		"lab/versions.env": read(t, "lab/versions.env"),
		"lab/.env.example": read(t, "lab/.env.example"),
		"lab/compose.yaml": read(t, "lab/compose.yaml"),
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
	texts["lab/versions.env"] = read(t, "lab/versions.env")
	texts["lab/compose.yaml"] += "\n      LAB_ADMIN_PASSWORD: ${LAB_ADMIN_PASSWORD:-}\n"
	if len(checkNoCommittedPassword(texts)) == 0 {
		t.Error("negative control: check accepted the password in the Compose environment")
	}
}
