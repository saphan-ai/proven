# STATUS — kotlin-gradle-adapter · round 1

- state: ready-for-review
- executor: cursor
- model: grok-4.7-high
- duration: 37 minutes, from the kickoff file mtime 2026-10-25T06:59:13Z to the docs commit 2026-10-25T07:35:54Z. That mtime is the file's, not a clock the engine printed for the spawn.
- last commit: a582b9767433056e302a27486c8907175485b0e6 "[kotlin-gradle-adapter] docs: describe the Kotlin Gradle adapter"
- updated: 2026-10-25T07:35:54Z

Cost and the monthly pool are omitted. A public document from 25 October 2026 carries neither.

## Summary

Kotlin and Gradle are one Maven adapter. `gradle/libs.versions.toml` and `build.gradle.kts` are two manifest readers, because `Diff` is given bytes and not a path. `gradle.lockfile` is the lock reader. All three compare with `internal/mavenver`. The sample `osv-kotlin-gradle` raises `com.example:widget` from 1.2.0 to 1.4.0 and reports one finding on `gradle.lockfile` line 5. The full suite on the tip exited 0.

## Acceptance, item by item

1. Reproduce first. Before the readers existed, a sample scan of `osv-kotlin-gradle` was a security review under the rule `build.gradle.kts`, with no dependency finding, the section "Agreed by two models, not proven", and `failed services: none`. The full stdout of that run was not kept after compaction, so this line is the recorded result, not a second run.
2. On the tip, after the docs commit: `go build ./...` exit 0, no output. `go vet ./...` exit 0, no output. `go test -count=1 ./...` exit 0. The three commands were run separately. Go is `go1.26.10`. `go.mod` says `go 1.26.0` and has no toolchain line. No package was listener-refused. Summary lines are under Evidence.
3. Negative control. Mutated field: `Compare` in `internal/mavenver/mavenver.go`. The first statement was `return 0, true`, so every pair compared equal. While that line was in place it was line 80. `go test -count=1 -run TestOSVKotlinGradleReportsTheDependency ./cmd/proven` exited 1. The failure is `missing "finding osv:com.example:widget ERROR Dependency com.example:widget 1.4.0 is in an OSV affected range"`, and the body had no dependency finding. The statement was removed. The same command then exited 0. The full suite after the restore exited 0.

## Evidence

`master-was-wrong: none`. At the base, before the edit, `Readers()` was Go, GoSum, NPM, NPMLock, PyPI, Poetry, Composer, ComposerLock, Cargo, CargoLock. Maven had no comparison. The three tests that used `Maven` as an unknown ecosystem were `adapter_test.go`, `osv/fix_test.go` and `osv/range_test.go`. They now use `NuGet` and still prove an unknown ecosystem is not compared. Sample fixtures live under `testdata/fixture/`.

Rulings where the order was silent:

- Two manifest readers and one lock reader. One reader would have to guess which file the bytes came from.
- In this subset a hyphen between numbers is the same separator as a dot, so `1.0-1` and `1.0.1` are one version. A test names that. Maven's list-item quirk is not implemented.
- A qualifier that is not in the Maven list sorts after `sp`, then as lowercase text. That placement was not checked against one named Maven release.
- A plugin, a missing `version.ref`, a rich version, and a `$` version are notes. `[bundles]` is not read. Groovy `build.gradle` and `settings.gradle.kts` are not matched. The lock file has no classifier field.
- The sample catalog does not raise `com.example:widget`, so the one finding is the lock file at the project root. Taking only `GradleLock` out of `Readers` leaves the literal in `build.gradle.kts`, and the finding remains. Taking `GradleKts` and `GradleLock` out together removes it. The negative control above is the one the order names.

`go test -count=1 ./...` exit 0, on the tip, after `Compare` was restored:

```
ok  	github.com/saphan-ai/proven/cmd/guardcount	1.180s
ok  	github.com/saphan-ai/proven/cmd/livecheck	0.045s
ok  	github.com/saphan-ai/proven/cmd/meterless-conform	0.004s
ok  	github.com/saphan-ai/proven/cmd/osv-to-csv	0.008s
ok  	github.com/saphan-ai/proven/cmd/proven	14.754s
ok  	github.com/saphan-ai/proven/cmd/settings-doc	0.008s
ok  	github.com/saphan-ai/proven/demo	0.344s
ok  	github.com/saphan-ai/proven/demo/live/review	0.006s
ok  	github.com/saphan-ai/proven/examples/formatter	0.004s
ok  	github.com/saphan-ai/proven/internal/agent/answer	0.003s
ok  	github.com/saphan-ai/proven/internal/agent/fixer	0.005s
ok  	github.com/saphan-ai/proven/internal/agent/investigator	0.005s
ok  	github.com/saphan-ai/proven/internal/agent/judge	0.003s
ok  	github.com/saphan-ai/proven/internal/agent/planner	0.006s
ok  	github.com/saphan-ai/proven/internal/agent/prover	0.005s
ok  	github.com/saphan-ai/proven/internal/agent/skeptic	0.005s
ok  	github.com/saphan-ai/proven/internal/agent/writer	0.004s
ok  	github.com/saphan-ai/proven/internal/cargover	0.001s
ok  	github.com/saphan-ai/proven/internal/change	21.281s
ok  	github.com/saphan-ai/proven/internal/childenv	0.001s
ok  	github.com/saphan-ai/proven/internal/cli	26.591s
ok  	github.com/saphan-ai/proven/internal/cli/sarifschema	2.019s
ok  	github.com/saphan-ai/proven/internal/composerver	0.002s
ok  	github.com/saphan-ai/proven/internal/envdoc	0.008s
ok  	github.com/saphan-ai/proven/internal/epss	0.004s
ok  	github.com/saphan-ai/proven/internal/fixture	0.008s
ok  	github.com/saphan-ai/proven/internal/generator	0.316s
ok  	github.com/saphan-ai/proven/internal/guardrail	0.213s
ok  	github.com/saphan-ai/proven/internal/hmem	0.443s
ok  	github.com/saphan-ai/proven/internal/limit	0.003s
ok  	github.com/saphan-ai/proven/internal/manifest	0.003s
ok  	github.com/saphan-ai/proven/internal/mavenver	0.003s
ok  	github.com/saphan-ai/proven/internal/osv	0.011s
ok  	github.com/saphan-ai/proven/internal/pathguard	0.004s
ok  	github.com/saphan-ai/proven/internal/pep440	0.002s
ok  	github.com/saphan-ai/proven/internal/pep503	0.001s
ok  	github.com/saphan-ai/proven/internal/pipeline	43.378s
ok  	github.com/saphan-ai/proven/internal/query	0.015s
ok  	github.com/saphan-ai/proven/internal/querytpl	0.010s
ok  	github.com/saphan-ai/proven/internal/redact	0.006s
ok  	github.com/saphan-ai/proven/internal/review	85.564s
ok  	github.com/saphan-ai/proven/internal/semver	0.001s
ok  	github.com/saphan-ai/proven/internal/settingspage	0.004s
ok  	github.com/saphan-ai/proven/internal/source	0.007s
ok  	github.com/saphan-ai/proven/internal/store	0.019s
ok  	github.com/saphan-ai/proven/internal/telemetry	0.010s
ok  	github.com/saphan-ai/proven/internal/testrun	12.227s
ok  	github.com/saphan-ai/proven/internal/tool	0.008s
ok  	github.com/saphan-ai/proven/internal/trace	0.128s
ok  	github.com/saphan-ai/proven/internal/world	0.412s
ok  	github.com/saphan-ai/proven/test/conformance/meterless	0.002s
ok  	github.com/saphan-ai/proven/test/corpus/cases/PV-009/change	0.004s
```

Packages with no test files, not failures: `internal/candidate`, `internal/livecheck`, `internal/service`, `test/corpus/cases/PV-003/base`, `test/corpus/cases/PV-003/change`, `test/corpus/cases/PV-003/check`, `test/corpus/cases/PV-009`, `test/corpus/cases/PV-009/base`, `test/corpus/cases/PV-012/base`, `test/corpus/cases/PV-012/change`.

Per-test pass, fail and skip counts for the whole suite were not printed by `go test` without `-v` or `-json`. UNRUN at that grain. No package failed.

## Blocked on / next

Nothing blocked. Next is the master's verification. This round is master-verify-only. Merge stays the owner's hand.
