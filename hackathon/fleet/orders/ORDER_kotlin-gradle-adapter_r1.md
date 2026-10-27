# ORDER — kotlin-gradle-adapter · round 1

- change: kotlin-gradle-adapter
- round: 1
- base: main @ 74d46e7
- executor: cursor on runner-cursor
- written: 2026-10-25T06:59:27Z
- task: Kotlin and Gradle as ONE language adapter (ADR-0038): Maven dependencies from gradle/libs.versions.toml, gradle.lockfile and build.gradle.kts, with a sample scenario and its ADAPTERS.md example

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — kotlin-gradle-adapter
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W57
- review: master-verify-only
- repo: proven
- base: main @ 74d46e7
- branch: fix/kotlin-gradle-adapter
- lands-on: main
- backend: cursor
- model: grok-4.7-high
- quote-usd: 5.00
- cap-usd: 10.00
- wall: 180m
- task: Kotlin and Gradle as ONE language adapter (ADR-0038): Maven dependencies from gradle/libs.versions.toml, gradle.lockfile and build.gradle.kts, with a sample scenario and its ADAPTERS.md example

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit subject at most 72
characters, `[kotlin-gradle-adapter] <area>: <what>` included, the area ONE token. No AI attribution anywhere. The tool re-adds a
`Co-authored-by:` line on amend; if it is there at the end, say so (the master removes it).
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S ORDER (2026-10-25, via the advisor)
Kotlin and Gradle through ONE language adapter, like PHP and Rust: read dependencies from gradle/libs.versions.toml and
the Gradle lockfile (gradle.lockfile) or build.gradle.kts, ecosystem Maven for OSV, with a sample scenario and its
docs/ADAPTERS.md example. Why: the QuakeWatch Android app's pull requests will be reviewed by Proven, and today Proven
sees nothing in a Kotlin project.

### MASTER'S READINGS (claims, measured by the master 2026-10-25 on main — re-measure, you may refute them)
- A language is `manifest.Adapter` (= `Reader`, a struct of function fields: Ecosystem, Match, Diff, Compare, Explain,
  Notes, Lock, NoteErr; internal/manifest/gomod.go). `Readers()` in internal/manifest/adapter.go lists them in a fixed
  order: Go, GoSum, NPM, NPMLock, PyPI, Poetry, Composer, ComposerLock, Cargo, CargoLock. The Rust pair
  (internal/manifest/cargo.go, cargolock.go) and the PHP pair (composer.go, composerlock.go) are the models to follow.
- Maven has NO version comparison today, and three tests use "Maven" as the example of an unknown ecosystem:
  internal/manifest/adapter_test.go:35, internal/osv/fix_test.go:106, internal/osv/range_test.go:80. Move them to
  another unknown name; do not delete what they prove.
- Sample scenarios live in testdata/fixture/<name>/ (models: osv-php-composer, osv-rust-cargo).
If any of this is false at your base, say so in `- master-was-wrong:` and build on what is there.

### WHAT TO BUILD
1. A Maven version comparison (Maven's own ordering rules: numeric segments, qualifiers alpha < beta < milestone < rc <
   snapshot < release < sp) good enough for OSV ranges, with table tests that name each rule.
2. A manifest adapter for gradle/libs.versions.toml (and a build.gradle.kts with literal coordinates; a version from a
   variable or a plugin it cannot resolve is a named note, never a guess), and a lock adapter for gradle.lockfile; both
   ecosystem "Maven", package names `group:artifact`.
3. A sample scenario `osv-kotlin-gradle`: a change that raises a dependency into a vulnerable range, reviewed end to end
   in sample mode.
4. docs/ADAPTERS.md: a worked example "Kotlin and Gradle" beside PHP.

### ACCEPTANCE
1. Reproduce first: at your base, the sample change is reviewed with no dependency finding; paste it.
2. `go build ./... && go vet ./... && go test -count=1 ./...`: green, or each listener-refused package named UNRUN with
   the quoted refusal (the master runs those).
3. **Negative control (law #27):** the Maven comparison made to treat every version as equal → the scenario's finding
   disappears and its test goes red; restore. Name the mutated line.

### WHAT IS NOT IN SCOPE
Groovy build.gradle files; Gradle plugins resolution; running Gradle.

### COMMENTS FOR THE JUDGES (ADR-0027, a standing rule of this train)
(1) every package you touch keeps its package comment true; (2) every exported name has a doc comment beginning with the
name; (3) every step a reader would stop at has a comment saying WHY; (4) code that carries out a decision names it;
(5) every test says which property it proves; (6) English, plain; no AI attribution, no host names, internal paths or
credential names, no commented-out code.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes lives under `███████████████████████████████████` in the worktree, except the files this order names.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes. Check it with `saphan doc check --family exec-status _EXEC_STATUS.md` before you
end: its exit code is the verdict.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
