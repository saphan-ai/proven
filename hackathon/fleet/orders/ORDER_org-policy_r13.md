# ORDER — org-policy · round 13

- change: org-policy
- round: 13
- base: main @ 16cf7b7
- executor: claude-code
- written: 2026-10-24T15:18:05Z
- task: round 3 — a repository cannot weaken the organisation policy in any combination, doubled review included; the policy file is refused anywhere in the reviewed repository; the default reads are guarded by tests

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — org-policy
track: proven-enterprise
- train: proven-enterprise
- wagon: W3
- review: named
- repo: proven
- base: main @ 16cf7b7
- branch: fix/org-policy
- lands-on: main
- backend: claude-code
- quote-usd: 12.00
- cap-usd: 24.00
- wall: 150m
- task: round 3 — a repository cannot weaken the organisation policy in any combination, doubled review included; the policy file is refused anywhere in the reviewed repository; the default reads are guarded by tests

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message `[org-policy] <area>: <what>`.
No AI attribution anywhere: no `Co-authored-by:` line, no tool or model name in a commit.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-23: ADR-0036 — *"A repository may make the policy stricter, never weaker"*; ADR-0036 Consequences: *"teams cannot
quietly lower it"*. 2026-10-24: the JSON condition is the CLI train's debt, not this wagon's.

### THE REVIEW OF ROUND 2 (tip f56ff13) — PASTED (law #44: carry, do not point)
Verdict: reject. Its BLOCKING findings and what a fix must make:
- **B1 · the comparison checks the class rank, not whether a path is doubled.** `internal/review/policy.go:282-288`,
  `:334-341`: `fixesClasses` does not count `doubled_classes`; `policy.go:290`, `:361` treat any higher class as
  stricter even when the policy doubles the lower class and not the higher one; `policy_test.go:157-164` and
  `docs/POLICY.md:75` assert the first case is allowed. Trigger 1: policy `doubled_classes: security`; repository
  `path_rules: **=docs`, `fallback_class: docs` → `go.mod` is doubled under the policy, once under the repository.
  Trigger 2: policy `fallback_class: backend`, `doubled_classes: backend`; repository `fallback_class: security` →
  `main.go` doubled under the policy, once under the repository. **A fix must make:** for every path, "doubled under
  the policy" imply "doubled under the repository's settings". Change the committed test and the docs line that
  assert the opposite.
- **B2 · "inside the repository" is checked against the target, not the repository.** `policy.go:408-428` (called at
  `open.go:214`) compares the policy file with `req.Target`; a target can be a subdirectory of a git repository.
  **A fix must make:** a policy file inside the git working tree that holds the target (its top level, links
  resolved) is refused by `PROVEN_POLICY`, like a file inside the target.
- **B3 · the two default reads of proven.yaml and the no-file case have no test.** `config.go:185-198` (baseline,
  used by `ci`), `:207-228` (working tree), `:191`/`:212` (no file → policy base). The code holds; the control is
  missing. **A fix must make:** the reviewer's probes P3 and P4 (below), or equal tests, committed.
- **O1 · OWED:** `TestARepositoryInheritsThePolicyValueItDoesNotSet` asserts only the statement, not the inherited value.

The reviewer's probe tests (red at f56ff13 except P3/P4, which guard B3), verbatim:
```go
// PROBE (review r3, not committed): a policy that doubles a class must keep every path it doubles doubled.
func TestProbeR3DoubledPathsStayDoubled(t *testing.T) {
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
```
```go
// PROBE (review r3, not committed): a policy file in the git repository of a subdirectory target is inside
// the repository under review, so the change could edit it; it must be refused.
func TestProbeR3APolicyInTheRepositoryOfASubdirectoryTargetIsRefused(t *testing.T) {
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
}
```
```go
// PROBE (review r3, not committed): with no proven.yaml in the target, the review runs on the policy's settings.
func TestProbeR3NoLocalFileRunsOnThePolicy(t *testing.T) {
	res, err := reviewUnderPolicy(t, "fallback_class: mobile\npath_rules: nothing-matches/**=docs\n", "", "2026-10-24")
	if err != nil {
		t.Fatal(err)
	}
	if res.Class != "mobile" {
		t.Fatalf("class %q rule %q, want the policy's fallback mobile", res.Class, res.ClassRule)
	}
}

// PROBE (review r3, not committed): the two default reads of proven.yaml — the target's working tree and the
// baseline commit — refuse a value weaker than the policy, as an explicit --config does.
func TestProbeR3TheDefaultReadsRefuseAWeakerProvenYaml(t *testing.T) {
	git := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=r", "-c", "user.email=█████████████████"}, args...)...)
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
```

### WHAT TO BUILD
1. **Rebase first:** `git rebase --onto main 16cf7b7` (main has moved). Do not merge.
2. B1, B2, B3 and O1 as above; commit the reviewer's probes (renamed to the package's style) as the regression tests.
3. `docs/POLICY.md`: the doubled-path rule and the repository-wide inside check.

### ACCEPTANCE
1. `go test -count=1 ./...` (whole module) green; `go vet ./...` clean.
2. The probes for B1 and B2 fail on f56ff13's behaviour and pass after; P3 and P4 are committed and pass.
3. Negative controls, named in your status: remove the doubled-path check — B1's test turns red; remove the
   repository-wide inside check — B2's test turns red. Restore both.
4. Subjects match `^\[org-policy\] [a-z0-9/._-]+: .+$`, at most 72 characters; no `Co-authored-by` line.

### WHAT IS NOT IN SCOPE
- JSON, SARIF or PR-comment writers (the CLI train's debt). W1, W2. Licences.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `████████████████████████` in the worktree.
Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, each finding with
its test, the negative controls, the whole-module counts. Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end.
Read every `saphan` verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences (`craft/COMMON.md`).

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
