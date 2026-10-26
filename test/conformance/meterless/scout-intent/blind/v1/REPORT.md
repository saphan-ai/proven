# Scout Intent on the blind set, v1: report

Scout Intent ran once over the frozen set. Nothing was changed to improve a result. Every miss is as it came out.

- Scout Intent code measured: commit `ec7896fe80ca62bc8b95053728791f68667add49`.
- Set frozen at: commit `ec7896fe80ca62bc8b95053728791f68667add49`.
- Kind measured by: `scout.ChangeRegistry(nil)` (the shipped registry, no extra security patterns), `Decide` on the formatted change, then the two kind rules of `choosePlan` in `internal/review/plan.go` (a security candidate wins; an ambiguous or not-high decision is a clarification). This is the kind Scout Intent chose.
- Plan measured by: `review.Open` and `Review` (live mode, against a local stub that refuses every model call, so no model runs) on a real git change built from the case's file list, with `scout_intent: on` and every other setting at its shipped default. The plan is read from the plan line the review writes on its trace before any agent runs. It is measured only for the lighter-kind answers below. `scout_intent` is off in the shipped default, and then every review runs the full plan.
- Regenerate: `go test -count=1 -run TestBlindSet ./internal/scout/ -args -blind-report`.
- Not measured: the plan for any case whose kind is not a lighter-kind answer; the file contents (each file in the built change holds a one-line placeholder, no shebang, not executable, so the prose check judges the name and the git mode only); any model verdict; the doubled review a security plan adds when a special reviewer is configured (none is configured by default), so the plan comparison is on agents, tools and sources only.

## Totals

Overall: 30 of 46 match; 16 miss; 7 lighter-kind answers; 0 of those still run a lighter plan than the expected answer would.

| tag | cases | match | miss |
|---|---|---|---|
| lookalike | 4 | 2 | 2 |
| misleading | 20 | 9 | 11 |
| pl | 4 | 2 | 2 |
| plain | 18 | 17 | 1 |
| th | 4 | 2 | 2 |

## Lighter-kind answers

Cases Scout answered with a lighter kind (docs or dependency) where the expected answer is a heavier kind or a clarification: 7. This lists the kind Scout chose, then the plan the live path ran for that change. A lighter kind is not a lighter review: the plan column says which it was. 0 of the 7 ran a plan with less work than the expected answer would.

- `blind-mis-docs-code-01` (misleading): expected change.feature, actual kind change.docs (band high, confidence 1.000). Files: `A docs/generate.go`. Live plan: the full plan runs in class docs (docs/**). This plan is not lighter than expected. Reviewer's reason: a Go source file under docs/ is code, not prose
- `blind-mis-docs-code-02` (misleading): expected change.feature, actual kind change.docs (band high, confidence 1.000). Files: `M docs/conf.py`. Live plan: the full plan runs in class docs (docs/**). This plan is not lighter than expected. Reviewer's reason: a Python file under docs/ runs at build time; it is code
- `blind-mis-docs-sec-01` (misleading): expected change.security, actual kind change.docs (band high, confidence 1.000). Files: `A docs/credentials.go`. Live plan: the full plan runs in class docs (docs/**). This plan is not lighter than expected. Reviewer's reason: a credentials-handling source file under docs/ is security-relevant code
- `blind-mis-rename-hides-01` (misleading): expected clarification, actual kind change.docs (band high, confidence 1.000). Files: `R docs/notes.md -> internal/authz/check.go`. Live plan: the full plan runs in class backend (fallback). This plan is not lighter than expected. Reviewer's reason: a rename that moves a prose file onto a code path hides a new source file
- `blind-mis-dep-vendor-01` (misleading): expected change.feature, actual kind change.dependency (band high, confidence 1.000). Files: `A vendor/github.com/x/y/y.go; M go.mod; M go.sum`. Live plan: class security (go.mod), agents planner+investigator+prover+skeptic+fixer, tools govulncheck, sources osv+epss, doubled yes. This plan is not lighter than expected. Reviewer's reason: vendored source is added with the manifest; new code enters the build
- `blind-pl-04` (pl, misleading): expected change.security, actual kind change.docs (band high, confidence 1.000). Files: `A docs/hasła.go`. Live plan: the full plan runs in class docs (docs/**). This plan is not lighter than expected. Reviewer's reason: a Go file named for passwords hides under docs/ with a Polish name
- `blind-th-04` (th, misleading): expected change.security, actual kind change.docs (band high, confidence 1.000). Files: `A docs/เข้าสู่ระบบ.js`. Live plan: the full plan runs in class docs (docs/**). This plan is not lighter than expected. Reviewer's reason: a script named for login under docs/ with a Thai name is security-relevant code

## All misses

16 cases.

- `blind-plain-sec-03` (plain): expected change.security, actual change.feature (band high, confidence 1.000). Files: `M internal/acl/policy.go; A internal/acl/policy_test.go`. Reviewer's reason: access-policy code changes
- `blind-mis-docs-code-01` (misleading): expected change.feature, actual change.docs (band high, confidence 1.000). Files: `A docs/generate.go`. Reviewer's reason: a Go source file under docs/ is code, not prose
- `blind-mis-docs-code-02` (misleading): expected change.feature, actual change.docs (band high, confidence 1.000). Files: `M docs/conf.py`. Reviewer's reason: a Python file under docs/ runs at build time; it is code
- `blind-mis-docs-sec-01` (misleading): expected change.security, actual change.docs (band high, confidence 1.000). Files: `A docs/credentials.go`. Reviewer's reason: a credentials-handling source file under docs/ is security-relevant code
- `blind-mis-cfg-logic-01` (misleading): expected change.security, actual clarification (band low, confidence 0.667). Files: `M .github/workflows/release.yml; A .github/scripts/publish.sh`. Reviewer's reason: a workflow plus a new script that publishes artifacts touches the release path
- `blind-mis-cfg-logic-02` (misleading): expected change.feature, actual clarification (band reject, confidence 0.000). Files: `A config/rules.lua`. Reviewer's reason: a config-folder file in a scripting language carries logic
- `blind-mis-cfg-logic-03` (misleading): expected change.security, actual clarification (band reject, confidence 0.000). Files: `M .env; M .env.production`. Reviewer's reason: environment files hold secrets; a config-looking change is security-relevant
- `blind-mis-rename-hides-01` (misleading): expected clarification, actual change.docs (band high, confidence 1.000). Files: `R docs/notes.md -> internal/authz/check.go`. Reviewer's reason: a rename that moves a prose file onto a code path hides a new source file
- `blind-mis-dep-vendor-01` (misleading): expected change.feature, actual change.dependency (band high, confidence 1.000). Files: `A vendor/github.com/x/y/y.go; M go.mod; M go.sum`. Reviewer's reason: vendored source is added with the manifest; new code enters the build
- `blind-mis-lock-only-01` (misleading): expected change.security, actual clarification (band medium, confidence 0.800). Files: `M package-lock.json; M scripts/preinstall.js`. Reviewer's reason: a lock file beside an install-time script is a supply-chain risk
- `blind-pl-02` (pl): expected change.security, actual change.feature (band high, confidence 1.000). Files: `A internal/uwierzytelnianie/sesja.go; M internal/uwierzytelnianie/hasło.go`. Reviewer's reason: Polish-named authentication and password code
- `blind-pl-04` (pl, misleading): expected change.security, actual change.docs (band high, confidence 1.000). Files: `A docs/hasła.go`. Reviewer's reason: a Go file named for passwords hides under docs/ with a Polish name
- `blind-th-02` (th): expected change.security, actual change.feature (band high, confidence 1.000). Files: `A internal/รหัสผ่าน/ตรวจสอบ.go`. Reviewer's reason: Thai-named password-check code
- `blind-th-04` (th, misleading): expected change.security, actual change.docs (band high, confidence 1.000). Files: `A docs/เข้าสู่ระบบ.js`. Reviewer's reason: a script named for login under docs/ with a Thai name is security-relevant code
- `blind-lookalike-02` (lookalike): expected change.docs, actual clarification (band reject, confidence 0.000). Files: `M README.mԁ`. Reviewer's reason: a Cyrillic-letter extension is odd but the file is still a readme-like prose file; no code is touched
- `blind-lookalike-04` (lookalike): expected change.dependency, actual clarification (band reject, confidence 0.000). Files: `M gо.mod`. Reviewer's reason: manifest name has a Cyrillic o; a careful reviewer treats it as a manifest edit, a dependency change

## Every case

| id | tags | expected | actual | match |
|---|---|---|---|---|
| blind-plain-dep-01 | plain | change.dependency | change.dependency | yes |
| blind-plain-dep-02 | plain | change.dependency | change.dependency | yes |
| blind-plain-dep-03 | plain | change.dependency | change.dependency | yes |
| blind-plain-ref-01 | plain | change.refactor | change.refactor | yes |
| blind-plain-ref-02 | plain | change.refactor | change.refactor | yes |
| blind-plain-ref-03 | plain | change.refactor | change.refactor | yes |
| blind-plain-feat-01 | plain | change.feature | change.feature | yes |
| blind-plain-feat-02 | plain | change.feature | change.feature | yes |
| blind-plain-feat-03 | plain | change.feature | change.feature | yes |
| blind-plain-sec-01 | plain | change.security | change.security | yes |
| blind-plain-sec-02 | plain | change.security | change.security | yes |
| blind-plain-sec-03 | plain | change.security | change.feature | MISS |
| blind-plain-docs-01 | plain | change.docs | change.docs | yes |
| blind-plain-docs-02 | plain | change.docs | change.docs | yes |
| blind-plain-docs-03 | plain | change.docs | change.docs | yes |
| blind-plain-cfg-01 | plain | change.config | change.config | yes |
| blind-plain-cfg-02 | plain | change.config | change.config | yes |
| blind-plain-cfg-03 | plain | change.config | change.config | yes |
| blind-mis-docs-code-01 | misleading | change.feature | change.docs | MISS (lighter kind) |
| blind-mis-docs-code-02 | misleading | change.feature | change.docs | MISS (lighter kind) |
| blind-mis-docs-sec-01 | misleading | change.security | change.docs | MISS (lighter kind) |
| blind-mis-docs-sec-02 | misleading | change.security | change.security | yes |
| blind-mis-md-name-01 | misleading | change.feature | change.feature | yes |
| blind-mis-md-name-02 | misleading | change.feature | change.feature | yes |
| blind-mis-md-name-03 | misleading | change.security | change.security | yes |
| blind-mis-manifest-code-01 | misleading | clarification | clarification | yes |
| blind-mis-manifest-code-02 | misleading | clarification | clarification | yes |
| blind-mis-cfg-logic-01 | misleading | change.security | clarification | MISS |
| blind-mis-cfg-logic-02 | misleading | change.feature | clarification | MISS |
| blind-mis-cfg-logic-03 | misleading | change.security | clarification | MISS |
| blind-mis-rename-hides-01 | misleading | clarification | change.docs | MISS (lighter kind) |
| blind-mis-rename-hides-02 | misleading | change.security | change.security | yes |
| blind-mis-dep-vendor-01 | misleading | change.feature | change.dependency | MISS (lighter kind) |
| blind-mis-lock-only-01 | misleading | change.security | clarification | MISS |
| blind-pl-01 | pl | change.docs | change.docs | yes |
| blind-pl-02 | pl | change.security | change.feature | MISS |
| blind-pl-03 | pl | change.feature | change.feature | yes |
| blind-pl-04 | pl, misleading | change.security | change.docs | MISS (lighter kind) |
| blind-th-01 | th | change.docs | change.docs | yes |
| blind-th-02 | th | change.security | change.feature | MISS |
| blind-th-03 | th | change.refactor | change.refactor | yes |
| blind-th-04 | th, misleading | change.security | change.docs | MISS (lighter kind) |
| blind-lookalike-01 | lookalike, misleading | change.security | change.security | yes |
| blind-lookalike-02 | lookalike | change.docs | clarification | MISS |
| blind-lookalike-03 | lookalike, misleading | change.security | change.security | yes |
| blind-lookalike-04 | lookalike | change.dependency | clarification | MISS |
