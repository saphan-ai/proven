package review

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The tests of the third round of ADR-0036: a repository cannot weaken the policy in any combination,
// doubled review included; a policy file is refused anywhere in the repository under review; the two
// default reads of proven.yaml and the case of no file are guarded.

// TestDoubledPathsStayDoubled: a policy that doubles a class must keep every path it doubles doubled.
// Each case is a pair of settings that each pass alone and together review a doubled path once.
func TestDoubledPathsStayDoubled(t *testing.T) {
	cases := []struct{ name, policy, local, path string }{
		{"no class key in the policy", "doubled_classes: security\n", "path_rules: **=docs\nfallback_class: docs\n", "go.mod"},
		{"a raised fallback leaves the doubled class", "fallback_class: backend\ndoubled_classes: backend\n", "fallback_class: security\n", "main.go"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := loadPolicy(writeFile(t, "policy.yaml", c.policy))
			if err != nil {
				t.Fatal(err)
			}
			base := p.base()
			polClass, _ := chooseClass([]string{c.path}, base.PathRules, base.ClassOrder, base.Fallback)
			cfg, err := p.parseLocal([]byte(c.local))
			if err != nil {
				return // refused: the property holds
			}
			class, _ := chooseClass([]string{c.path}, cfg.PathRules, cfg.ClassOrder, cfg.Fallback)
			if base.Doubled[polClass] && !cfg.Doubled[class] {
				t.Fatalf("%s: the policy doubles it as %s; the repository's settings review it once as %s, and were not refused", c.path, polClass, class)
			}
		})
	}
}

// TestADoubledPathIsRefusedByName: the refusal names the key that moves the path, the local value and
// the class the policy doubles, so the reader knows which of the repository's lines to change.
func TestADoubledPathIsRefusedByName(t *testing.T) {
	cases := []struct {
		name, policy, local, key string
		why                      []string
	}{
		{
			"a rule that puts every path below a doubled class", "doubled_classes: security\n", "path_rules: **=docs\n",
			"path_rules", []string{"weaker than the organisation policy"},
		},
		{
			"a fallback above a doubled class the repository does not double", "fallback_class: backend\ndoubled_classes: backend\n", "fallback_class: security\n",
			"fallback_class", []string{"security", "backend", "doubles"},
		},
		{
			"a rule above a doubled class the repository does not double", "path_rules: **/*.go=backend\nfallback_class: docs\ndoubled_classes: backend\n", "path_rules: **/*.go=backend, go.mod=security\n",
			"path_rules", []string{"security", "backend", "doubles"},
		},
		{
			"a new order of the classes under a doubling policy", "doubled_classes: security\n", "class_order: backend, security, mobile, user interface, docs\n",
			"class_order", []string{"fixes the order of the classes"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, err := loadPolicy(writeFile(t, "policy.yaml", c.policy))
			if err != nil {
				t.Fatal(err)
			}
			_, err = p.parseLocal([]byte(c.local))
			ce, ok := err.(*ConfigError)
			if !ok || ce.Field != c.key {
				t.Fatalf("want a *ConfigError on %s, got %v", c.key, err)
			}
			for _, want := range c.why {
				if !strings.Contains(ce.Reason, want) {
					t.Errorf("reason lacks %q: %s", want, ce.Reason)
				}
			}
		})
	}
}

// TestAStricterClassStaysFreeWhenItIsDoubled: a repository may raise a path above a doubled class when
// it doubles the higher class too, and may copy the policy's own settings.
func TestAStricterClassStaysFreeWhenItIsDoubled(t *testing.T) {
	p, err := loadPolicy(writeFile(t, "policy.yaml", "fallback_class: backend\ndoubled_classes: backend\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, local := range []string{
		"fallback_class: security\ndoubled_classes: security, backend\n",
		"fallback_class: backend\ndoubled_classes: backend\n",
		"fallback_class: backend\n",
	} {
		if _, err := p.parseLocal([]byte(local)); err != nil {
			t.Errorf("%q was refused: %v", local, err)
		}
	}
}

// TestNoCombinationOfSettingsUndoublesAPath sweeps every pairing of a policy and a repository from a small
// grammar of values and checks the property of the order for each path of a list: when the policy
// doubles a path, the repository's settings double it too, with and without a second reviewer. It also
// counts the pairs accepted and refused, so a sweep that refuses everything, or nothing, fails.
func TestNoCombinationOfSettingsUndoublesAPath(t *testing.T) {
	policyDoubled := []string{"", "doubled_classes: off\n", "doubled_classes: security\n", "doubled_classes: backend\n", "doubled_classes: security, backend\n", "doubled_classes: docs\n"}
	policyClass := []string{"", "fallback_class: security\n", "fallback_class: backend\n", "fallback_class: docs\n", "path_rules: **/*.go=backend, go.mod=security\n"}
	localFallback := []string{"", "fallback_class: security\n", "fallback_class: backend\n", "fallback_class: mobile\n", "fallback_class: user interface\n", "fallback_class: docs\n"}
	localRules := []string{"", "path_rules: **=security\n", "path_rules: **=backend\n", "path_rules: **=docs\n", "path_rules: **/*.go=docs\n", "path_rules: **/*.go=security, **=docs\n", "path_rules: go.mod=security, **/*.go=backend, **/*.md=docs\n"}
	localDoubled := []string{"", "doubled_classes: off\n", "doubled_classes: security\n", "doubled_classes: backend\n", "doubled_classes: security, backend\n", "doubled_classes: security, backend, mobile, user interface, docs\n"}
	paths := []string{"go.mod", "main.go", "README.md", "docs/a.md", "vendor/x.go", "app/auth/login.go", "ui/view.tsx"}
	accepted, refused := 0, 0
	for _, pd := range policyDoubled {
		for _, pc := range policyClass {
			polBody := pd + pc
			if polBody == "" {
				continue
			}
			p, err := loadPolicy(writeFile(t, "policy.yaml", polBody))
			if err != nil {
				t.Fatal(err)
			}
			_, doubledSet := p.given["doubled_classes"]
			for _, lf := range localFallback {
				for _, lr := range localRules {
					for _, ld := range localDoubled {
						local := lf + lr + ld
						cfg, err := p.parseLocal([]byte(local))
						if err != nil {
							refused++
							continue
						}
						accepted++
						if !doubledSet {
							continue
						}
						for _, second := range []string{"", "second-model"} {
							pol, repo := withSecond(p.settings, second), withSecond(cfg, second)
							for _, path := range paths {
								polClass, _ := chooseClass([]string{path}, pol.PathRules, pol.ClassOrder, pol.Fallback)
								class, _ := chooseClass([]string{path}, repo.PathRules, repo.ClassOrder, repo.Fallback)
								if pol.Doubled[polClass] && !repo.Doubled[class] {
									t.Fatalf("policy %q, repository %q, second reviewer %q: %s is doubled as %s by the policy and reviewed once as %s by the repository",
										polBody, local, second, path, polClass, class)
								}
							}
						}
					}
				}
			}
		}
	}
	if accepted == 0 || refused == 0 {
		t.Fatalf("the sweep is not a control: %d accepted, %d refused", accepted, refused)
	}
}

// TestAPolicyInTheRepositoryOfASubdirectoryTargetIsRefused: a policy file in the git repository of a
// subdirectory target is inside the repository under review, so the change could edit it.
func TestAPolicyInTheRepositoryOfASubdirectoryTargetIsRefused(t *testing.T) {
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	target := filepath.Join(repo, "service")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	policyPath := filepath.Join(repo, "policy.yaml")
	if err := os.WriteFile(policyPath, []byte("fallback_class: security\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := loadPolicy(policyPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.refuseInside(target); err == nil {
		t.Fatal("a policy in the same git repository as the target was not refused")
	}
	// A link into the repository from outside is the same file.
	link := filepath.Join(t.TempDir(), "policy-link.yaml")
	if err := os.Symlink(policyPath, link); err != nil {
		t.Fatal(err)
	}
	linked, err := loadPolicy(link)
	if err != nil {
		t.Fatal(err)
	}
	if err := linked.refuseInside(target); err == nil {
		t.Fatal("a link to a policy in the git repository of the target was not refused")
	}
	// A policy beside the repository, not in it, is accepted.
	outside, err := loadPolicy(writeFile(t, "policy.yaml", "fallback_class: security\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err := outside.refuseInside(target); err != nil {
		t.Fatalf("a policy outside the repository was refused: %v", err)
	}
}

// TestAPolicyInTheRepositoryIsRefusedByAReview: the refusal holds through Review, not only the method.
func TestAPolicyInTheRepositoryIsRefusedByAReview(t *testing.T) {
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	target := filepath.Join(repo, "service")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	policyPath := filepath.Join(repo, "policy.yaml")
	if err := os.WriteFile(policyPath, []byte("fallback_class: security\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rev, err := Open(Options{Sample: true, LogDir: t.TempDir(), PolicyPath: policyPath})
	if err != nil {
		t.Fatal(err)
	}
	_, err = rev.Review(t.Context(), Request{Target: target})
	ce, ok := err.(*ConfigError)
	if !ok || ce.Field != EnvPolicy || !strings.Contains(ce.Reason, "inside the repository under review") {
		t.Fatalf("want the policy in the repository refused, got %v", err)
	}
}

// TestNoLocalFileRunsOnThePolicy: with no proven.yaml in the target, the review runs on the policy's settings.
func TestNoLocalFileRunsOnThePolicy(t *testing.T) {
	res, err := reviewUnderPolicy(t, "fallback_class: mobile\npath_rules: nothing-matches/**=docs\n", "", "2026-10-24")
	if err != nil {
		t.Fatal(err)
	}
	if res.Class != "mobile" {
		t.Fatalf("class %q rule %q, want the policy's fallback mobile", res.Class, res.ClassRule)
	}
}

// TestTheDefaultReadsRefuseAWeakerProvenYaml: the two default reads of proven.yaml — the target's working
// tree and the baseline commit — refuse a value weaker than the policy, as an explicit --config does.
func TestTheDefaultReadsRefuseAWeakerProvenYaml(t *testing.T) {
	git := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=r", "-c", "user.email=r@example.invalid"}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "proven.yaml"), []byte("doubled_classes: off\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(target, "init", "-q")
	git(target, "add", "proven.yaml")
	git(target, "commit", "-q", "-m", "[probe] test: base")
	head := git(target, "rev-parse", "HEAD")
	for name, req := range map[string]Request{
		"working tree": {Target: target},
		"baseline":     {Target: target, BaselineCommit: head, HeadCommit: head},
	} {
		t.Run(name, func(t *testing.T) {
			rev, err := Open(Options{Sample: true, LogDir: t.TempDir(), PolicyPath: writeFile(t, "policy.yaml", "doubled_classes: security\n")})
			if err != nil {
				t.Fatal(err)
			}
			_, err = rev.Review(t.Context(), req)
			ce, ok := err.(*ConfigError)
			if !ok || ce.Field != "doubled_classes" || !strings.Contains(ce.Reason, "weaker than the organisation policy") {
				t.Fatalf("want the weaker doubled_classes refused, got %v", err)
			}
		})
	}
}
