# STATUS — govulncheck-says-what-is-reached · round 2

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 21 min
- last commit: 6e6c09a [govulncheck-says-what-is-reached] merge: absorb main
- updated: 2026-10-24T12:10:59Z

## Summary

Go dependency findings now say reached, imported but not reached, or not measured, from govulncheck against a local file database. A reached finding ranks first. The master's verification is next.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — govulncheck-says-what-is-reached
- state: stop-2-ready
- abstract: Go dependency findings now say reached, imported but not reached, or not measured, from govulncheck against a local file database. A reached finding ranks first. The master's verification is next.
- last commit: 6e6c09a [govulncheck-says-what-is-reached] merge: absorb main
- done: govulncheck labels each Go dependency finding and ranks a reached one first
  - recorded scenario has one reached function with a call path and one imported but not reached, and a Packagist finding stays not measured: go test -count=1 ./internal/tool exit 0 inside the suite
  - negative control field Symbol.Label: labelOf returned reached for the unreached trace, TestUnreachedFunctionIsNotLabeledReached failed with Symbol.Label = "reached" path "", then the return was restored and the same test exited 0
  - a missed run and a module the scan does not mention stay not measured, never imported but not reached: TestAMissedRunIsNotMeasured and TestSilenceIsNotMeasured
  - an output with no govulncheck config object is not a measurement: TestAFailedStartIsNotAScan
  - a reached finding ranks ahead of an unreached finding that has a higher EPSS score: TestReachedRanksAheadOfAHigherScore
  - docs/CONFIGURATION.md lists PROVEN_GOVULNCHECK_DB: go test -count=1 ./internal/settingspage exit 0
- next / blocked-on: none
- artifacts:
  - internal/tool/govulncheck.go
  - internal/tool/govulncheck_test.go
  - internal/tool/govulncheck_container_test.go
  - internal/tool/race_level_test.go
  - internal/tool/sample.go
  - internal/tool/settings.go
  - internal/tool/tool.go
  - internal/pipeline/pipeline.go
  - internal/candidate/candidate.go
  - internal/agent/writer/writer.go
  - docs/ADAPTERS.md
  - docs/CONFIGURATION.md
- evidence:
  - toolchain: go.mod says go 1.26; local go version go1.26.10; no Makefile; no .github workflow in this tree
  - baseline before the edit, committed tip aaeda1a77f395d515135b6aee41c517b419ce3e1: go test -count=1 ./... exit 1. Every package ok except internal/pipeline. The only failure was TestMaterializedTreesAreReadableByOthers: go.mod mode 664 is writable by group or others
  - go build ./... exit 0
  - go vet ./... exit 0
  - go test -count=1 ./... after the change exit 1. The only failure is the same TestMaterializedTreesAreReadableByOthers, same mode 664. No new failure
  - negative control: go test -count=1 -run TestUnreachedFunctionIsNotLabeledReached ./internal/tool exit 1, then exit 0 after restore
  - docker version was denied by the executor command list. Whether golang:1.26 contains govulncheck is UNMEASURED. The container test is gated on PROVEN_TEST_DOCKER=1 and was not run
  - go test -race UNRUN: this change starts no goroutine and shares no mutable state
  - testing bar UNRUN: the order declares no coverage, mutation, lint or type metric
  - remark: an ok package line does not say how many tests skipped. This summary cannot separate a skip from a pass
- note: master-was-wrong: none of the code claims. The tool adapter, the race detector, the container runner with no network, and OSV dependency findings are on this tip, as measured. The order names base main at 656dbcb; that commit is an ancestor of aaeda1a, which adds four corpus commits, and this change sits on aaeda1a. The govulncheck JSON field names are the ones the existing comment attributes to golang.org/x/vuln v1.1.4; this leg did not open that module. Rulings: a module that is required but not imported uses the same label as imported but not reached, because the order names no fourth label; labels are read from the change scan; ranking applies inside one severity and only among findings that already have a reach label; sample mode names the tool as not run and does not replay GovulncheckMixed onto the fixture, which matches RaceIntroduced. No Co-authored-by trailer on d410e4b.
- updated: 2026-10-24T09:51:26Z

CLOSE-OK · govulncheck-says-what-is-reached · commit 6e6c09a38865cee8e3cfdae66763e5ddc3c9a41d · branch fix/govulncheck-says-what-is-reached · worktree ████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
