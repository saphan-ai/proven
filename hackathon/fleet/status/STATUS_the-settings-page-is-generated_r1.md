# STATUS — the-settings-page-is-generated · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: f29f2ac01bd075db1ccef0e3b15c02e62a7d47db [the-settings-page-is-generated] envdoc: generate the settings tables
- updated: 2026-10-22T15:37:51Z

## Summary

Environment tables in docs/CONFIGURATION.md are generated from the code's declarations, and the drift test names a setting that disagrees. No cell changed. The master's verification is next.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-settings-page-is-generated
- state: stop-2-ready
- abstract: Environment tables in docs/CONFIGURATION.md are generated from the code's declarations, and the drift test names a setting that disagrees. No cell changed. The master's verification is next.
- last commit: f29f2ac01bd075db1ccef0e3b15c02e62a7d47db [the-settings-page-is-generated] envdoc: generate the settings tables
- done: The environment tables are generated from the packages' declarations and the drift test fails when the page disagrees.
  - go test -count=1 ./internal/settingspage -run TestConfigurationPageMatchesTheRegistry exit 0 after go run ./cmd/settings-doc
  - negative control added PROVEN_FAKE_SETTING to generator.Environment without regenerating; the same test exited 1 and named PROVEN_FAKE_SETTING; the row was then removed
  - git diff --stat -- docs/CONFIGURATION.md against the parent is 2 insertions, the marker lines only
- next / blocked-on: none
- artifacts:
  - cmd/settings-doc/main.go
  - cmd/livecheck/main.go
  - cmd/livecheck/main_test.go
  - docs/CONFIGURATION.md
  - internal/envdoc/envdoc.go
  - internal/envdoc/envdoc_test.go
  - internal/generator/settings.go
  - internal/guardrail/settings.go
  - internal/livecheck/settings.go
  - internal/query/settings.go
  - internal/settingspage/catalog.go
  - internal/settingspage/page_test.go
- evidence:
  - go version exit 0: go1.26.8 linux/amd64; go.mod says go 1.26 and has no toolchain line
  - baseline before edits: go test -count=1 ./... exit 0; every package ok or no test files; no FAIL line
  - RED before generation: go test -count=1 -v ./internal/settingspage -run TestConfigurationPageMatchesTheRegistry exit 1, missing <!-- generated: settings -->
  - go run ./cmd/settings-doc exit 0, printed the configuration path
  - git diff --stat -- docs/CONFIGURATION.md: 1 file changed, 2 insertions
  - negative control exit 1: docs/CONFIGURATION.md drifted on PROVEN_FAKE_SETTING; run: go run ./cmd/settings-doc
  - revert of the page then the same test exit 1 again (markers gone); go run ./cmd/settings-doc exit 0 restored them
  - second go run ./cmd/settings-doc exit 0 and the diff stayed 2 insertions
  - go build ./... exit 0
  - go vet ./... exit 0
  - tip go test -count=1 ./... exit 0; new packages internal/envdoc, internal/settingspage and cmd/settings-doc included; no FAIL line
  - remark: a package line of ok is the same when a test inside it skipped, so skip counts are UNMEASURED
  - remark: go test -race UNRUN; this change adds no shared state and no goroutine
  - remark: testing-bar UNRUN; the order and the repository declare no coverage, mutation, lint or type metric
  - remark: no .github workflow was found, so the acceptance commands are the commands that were run
  - session recording ADR-0024: no recording file is in this worktree and this order does not publish one; whether a proxy recorded the session is UNMEASURED from here
- note: master-was-wrong: M2 holds, nothing tied the page to the code. M1 is wrong that every source declares an Env constant: livecheck used unexported envEvidenceView in package main. The page intro also cites internal/review/config.go, which is proven.yaml and was not generated. The commit carries a Co-authored-by trailer the tool added; the master removes it. Amend was not used.
  - ruling: the evidence-view strings moved to internal/livecheck beside EnvEvidenceView — why the generator cannot import package main — cost if wrong is a different view name; livecheck tests passed
  - ruling: one marker pair also holds the three headings and the sentence after the guardrail table — why the order names one pair — cost if wrong is that a hand edit of that sentence is overwritten
  - ruling: PROVEN_EVIDENCE_VIEW stays on the Zetaris table — why the hand-written page already put it there — cost if wrong is a new heading
  - ruling: duration cells use a whole-seconds spelling (60s) rather than time.Duration.String (1m0s) — why the hand-written rows and the refusal examples use 60s — cost if wrong is a page that says 1m0s, which Load still accepts
  - ruling: Secret is a bool, and the table still says Secret in the description; TestSecretFlagMatchesTheDescription requires they agree
- updated: 2026-10-22T15:06:53+00:00

CLOSE-OK · the-settings-page-is-generated · commit 642afcb84c108afa1a60d00dd3b33dcce2339537 · branch fix/the-settings-page-is-generated · worktree ██████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
