# REVIEW — evidence-scope-per-role · round 1

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T07:06:13Z

## Summary

W55 on claude-opus-5-5: each role's scope is checked before any store on the Investigator path and on the Skeptic (Execute) path, and removing either check turns its test red. Load-time refusals and nil-means-every-template hold, and proven.yaml is read at the baseline.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — evidence-scope-per-role

### Ground
- Reviewed `main...b0c1183` (three-dot). `git rev-parse --short=7 HEAD` printed `b0c1183`. 22 files, +621/−72: the leg's commit 2079e8f and the master's absorb of main (77a72fc).
- Model: claude-opus-5-5 (claude-code), run on 2026-10-25.
- Commands that ran here: `go test -count=1 ./internal/querytpl/ ./internal/agent/... ./internal/fixture/` passed (querytpl, answer, fixer, investigator, judge, planner, prover, skeptic, writer, fixture). `go vet` on the same packages was clean.
- `go build ./...` and the `internal/pipeline` and `internal/review` packages FAILED SETUP, not as tests: go.opentelemetry.io/otel v1.47.0 is not in the module cache, and the proxy refused it (`tls: failed to verify certificate: x509: OSStatus -26276` / `Forbidden`). See Unrun.
- I read the callers: every `q.Run` / `investigator.Execute` / `skeptic.Attempt` site in `internal/` and `cmd/` (grep), plus `review/config.go` parseConfig/applyKey, the `review/open.go` settings load and the pipeline wiring.

### Differential
- Property 1 (refused by name before any store, on every path a role can take):
  - Investigator: `internal/agent/investigator/investigator.go:113` checks the scope after the empty check and before both the nil-runner branch and `q.Run` (:120, :135). The refusal becomes `c.QueryNote = "refused: …"` and the call returns nil, so the review continues.
  - Skeptic: `investigator.go:78` (`Execute`) checks before the nil-engine branch and before `q.Run`. Its one caller is `pipeline.go:1110` (`runRefutation`), and both skeptic passes (default :745, special :780) send `req.SkepticScope`. A `*querytpl.Refusal` clears `r.Query`, sets `r.Reason` and returns nil (:1116–1119), so the review goes on.
  - Other store callers: `osv.Check` (osv/check.go:235, :332) and `epss.Attach` (epss/epss.go:155). Both are Proven's own deterministic checks, not a role's request, and no model writes their template. They are out of the scope by design (docs/CONFIGURATION.md: only the Investigator and the Skeptic have a query path). See Remainder.
  - An unknown name under a narrowed scope is refused as `ReasonUnknownTemplate` without echoing the model's text (`querytpl/scope.go:49–51`).
- Property 2 (each prompt lists exactly its scope): the Task line is built from `scope.Describe()` in `investigator.go:47` and `skeptic.go:130`. HOLDS for the Task list. PARTIAL for the Investigator prompt as a whole: its `Example` is the fixed `{"template":"kev-by-cve",…}` (`investigator.go:50`). Measured with a throwaway test: under `Scope{epss-by-cve}` the prompt lists only epss-by-cve and still shows kev-by-cve in the Example (`contains kev-by-cve: true`). That contradicts the doc comment at `investigator.go:41–43` ("the model is not offered a template the scope would refuse"). The Skeptic's Example names no template.
- Property 3 (load): unset leaves the scope nil, and `Scope.Allows` returns true for nil, so every template is allowed (`scope.go:30`; `TestNilScopeIsEveryTemplate`, `TestEvidenceScopeKeysLoad`). `ParseScope` refuses an unknown name, a duplicate, an empty entry and an empty list, each by name (`scope.go:84–97`). An unknown key goes through the existing closed-key path (`config.go:295`), and `planner_templates` is covered by `TestEvidenceScopeIsRefusedOnLoadByName`. A key written twice is refused by the existing `duplicate key` rule (`config.go:302`).
- Property 4 (no model widens a scope): the scope is loaded only from proven.yaml. It is read at the BASELINE commit when there is one (`open.go:454`, `loadSettingsAtBaseline`), so the change under review cannot widen it either. A model's text reaches only `req.Template`, which is checked against the scope and never written back into it. HOLDS.
- Absorb hunks (`pipeline.go:674/733/769`): the telemetry contexts `ictx`/`kctx` are kept and the scope argument is added. Correct as resolved.

### Unplug
- `investigator.go:113` scope.Check → `error(nil)`: **red** — `TestInvestigateRefusesOutOfScopeBeforeAnyStore` ("the store was asked: [SELECT … px_kev_vulns …]").
- `investigator.go:78` scope.Check → `error(nil)`: **red** — `TestExecuteRefusesOutOfScopeBeforeAnyStore` (store up: `<nil>`, store asked; no store: reported as service down). This reproduces the master's measurement.
- `skeptic.go:130` `scope.Describe()` → `querytpl.Describe()`: **red** — skeptic `TestPromptListsOnlyTheScopesTemplates`.
- `querytpl/scope.go:93` duplicate check disabled: **red** — `TestParseScope`.
- All mutants were restored with `git checkout -- <file>`. `git diff` on tracked files is empty.
- Mutants judged BY READING ONLY, because pipeline/review do not build here. Under each, I found no test in the tree that would turn red (grep: no test sets `InvestigatorScope`, and none loads a scope from proven.yaml through `open.go`):
  - `open.go:513` or `:514` deleted (config never reaches the pipeline);
  - `pipeline.go:674` `req.InvestigatorScope` → nil;
  - `pipeline.go:733` / `:769` Attempt scope → nil (prompt only; enforcement still holds at Execute);
  - `pipeline.go:780` (special-skeptic `runRefutation`) scope → nil, unless the discount sample runs a special reviewer, which I did not verify.

### Closure
- The guards in the role functions are each closed by a test that turns red. The WIRING between them is not: proven.yaml → `fileSettings` → `pipeline.Request` → `Investigate`/`runRefutation` (`open.go:513–514`, `pipeline.go:674`, `:780`). Today it is correct by reading. If a refactor drops one of those lines, every role silently goes back to every template and the suite stays green.

### Remainder
- `osv.Check` and `epss.Attach` query stores with fixed templates, not role requests. No scope applies to them. This is the documented design, not a defect; listed so the master can rule on it (R11, declined to judge).
- `pipeline.go:1118`: an out-of-scope query refusal overwrites `r.Reason` even when a refuting test already set "test X passes on the change". The overwrite is older than this diff (any Refusal did it), but scopes make it easier to reach. Declined to judge its effect on the Judge.
- Scratch left in the worktree: `███████████████████████████████████████████████████`. It is the throwaway test that measured the Example finding, and it passes with logs only. I could not remove it because `rm` is denied under `██████████████████████████` ("do not rebuild the act from other commands"). Owed: the master deletes it. It is untracked and not part of b0c1183.

### Profile
- security (R9: configuration and data access, an authorization-style scope). S3 was run on both authorization checks (Unplug). S1: the trust boundary is model text → `req.Template`. It can be spoofed to any string, and that string is closed by Check plus `source.Lookup`. No secrets, dependency changes or destructive paths are in the diff.

### Focus
- The four properties, in order. Every guard was mutated where the package builds.

### Conditions
1. **owed (property 2)** — `internal/agent/investigator/investigator.go:50`: the Investigator Example names `kev-by-cve` whatever the scope. Remedy: derive the Example from the scope's first template, or use an empty-template example like the Skeptic's. Add a test with a scope that excludes kev-by-cve and assert that kev-by-cve does not appear anywhere in the prompt.
2. **owed (property 1 wiring)** — `internal/review/open.go:513–514`, `internal/pipeline/pipeline.go:674`, `:780`: add one test that loads a proven.yaml naming both keys, runs the sample review through `review`, and asserts an out-of-scope refusal in the trace for the Investigator AND the Skeptic. Removing any of those lines must turn it red.
3. **owed (hygiene)** — remove the untracked scratch `███████████████████████████████████████████████████` before teardown.

### Unrun
- `go build ./...`, `go vet ./...` and `go test ./...` for every package that imports `internal/telemetry`, which includes `internal/pipeline` and `internal/review` and so `TestSkepticRequestOutsideItsScopeAsksNoStore`, `TestSkepticIsNarrowerThanTheInvestigator`, `TestEvidenceScopeKeysLoad` and `TestEvidenceScopeIsRefusedOnLoadByName`. Quoted refusal: `go.opentelemetry.io/████████████: Get "https://proxy.golang.org/go.opentelemetry.io/otel/@v/v1.47.0.zip": tls: failed to verify certificate: x509: OSStatus -26276` (later attempts: `Forbidden`). The master's macm5 measurement (full suite green at b0c1183) is the only evidence for these, and I did not reproduce it.
- The pipeline/review wiring mutants listed under Unplug were judged by reading, not run.

## The review status as filed

## REVIEW — evidence-scope-per-role
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: W55 on claude-opus-5-5: each role's scope is checked before any store on the Investigator path and on the Skeptic (Execute) path, and removing either check turns its test red. Load-time refusals and nil-means-every-template hold, and proven.yaml is read at the baseline.
  Conditions (owed): the Investigator prompt Example names kev-by-cve whatever the scope (property 2, measured), and no test covers proven.yaml→pipeline scope wiring (open.go:513–514, pipeline.go:674/780).
  Unrun: pipeline/review packages (otel v1.47.0 not in module cache, proxy TLS refused). An untracked probe test was left in the worktree because rm is denied.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T07:30:00Z

## REDACTED

| what | times |
|---|---|
| e-mail or user name | 1 |
| file and line in Saphan Studio's private code | 2 |
| policy binding name | 1 |
