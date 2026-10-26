# STATUS — lock-files-are-checked · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: a3f0027192d7cedc42076cfbc4c7f152753d5b18 [lock-files-are-checked] manifest: check lock file versions
- updated: 2026-10-23T02:15:10Z

## Summary

go.sum, package-lock.json and poetry.lock are checked with the manifests. A package named by both is checked once, at the lock file version. A lock file that does not parse leaves a trace note and no finding.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — lock-files-are-checked
- state: stop-2-ready
- abstract: go.sum, package-lock.json and poetry.lock are checked with the manifests. A package named by both is checked once, at the lock file version. A lock file that does not parse leaves a trace note and no finding.
- last commit: a3f0027192d7cedc42076cfbc4c7f152753d5b18 [lock-files-are-checked] manifest: check lock file versions
- done: Lock files go.sum, package-lock.json and poetry.lock are checked against OSV
  - table tests for added, raised, unchanged, removed and a malformed file passed (`go test -count=1 ./internal/manifest` exit 0, and `go test -count=1 ./internal/osv` exit 0)
  - negative control mutated the go.sum `/go.mod` skip; `TestGoSumDropsTheGoModDuplicate` exited 1 and named version `v1.2.3/go.mod`; the skip was restored and `go test -count=1 ./...` exited 0
  - `go build ./...` exit 0 and `go vet ./...` exit 0 on the tip
- next / blocked-on: none
- artifacts:
  - docs/OSV_TABLE.md
  - internal/manifest/diff.go
  - internal/manifest/gomod.go
  - internal/manifest/gosum.go
  - internal/manifest/gosum_test.go
  - internal/manifest/lock_test.go
  - internal/manifest/npmlock.go
  - internal/manifest/npmlock_test.go
  - internal/manifest/poetry.go
  - internal/manifest/poetry_test.go
  - internal/osv/check.go
  - internal/osv/check_test.go
- evidence:
  - baseline before edits: `go version` printed `go version go1.26.8 linux/amd64`; `go.mod` says `go 1.26` and has no toolchain line; `go test -count=1 ./...` exit 0
  - `go build ./...` exit 0
  - `go vet ./...` exit 0
  - `go test -count=1 ./...` exit 0 after the restore; every package with tests printed ok
  - negative control exit 1: `duplicate go.mod line kept` with two changes, the second `Version:"v1.2.3/go.mod"`; file restored; the following full suite exit 0
  - sample scan CLI not run: the built binary was refused, allowed form none; `go test -count=1 ./cmd/proven` exit 0 includes `TestScanSamplePrintsF1AndF2NotF3`, which calls `run` from a temp directory with `scan --sample <repo>/testdata/fixture/discount` and requires F1, F2, F4, the OSV finding for `evil.example/leftpad` `v1.2.3`, `failed services: none`, and fails if F3 is printed; CLI stdout was not captured
  - no coverage, mutation, lint or type-warning bar is declared; those metrics are UNRUN; no threshold, skip or assertion was weakened
  - `-race` UNRUN: this change does not share state across goroutines
  - ruling: a lock file keeps every version of a package; a version already on the base is reported only when it is higher than every base version, or the two do not order — cost if wrong: a new older nested copy beside a higher pin already on the base is not checked
  - ruling: package-lock.json is read only at lockfileVersion 2 or 3, and only for paths under node_modules — cost if wrong: a workspace package and a v1 lock are not checked
  - ruling: poetry name and version are the assignments before the next table header, and the hash in go.sum must start with h1: — cost if wrong: a legal file in another shape is a trace note and no finding
  - ruling: dedup is the same directory, and PyPI names match without regard to case — cost if wrong: Flask and flask in different directories would be checked twice, or a case-only Go path would be collapsed
  - ruling: an empty lock file is zero packages — cost if wrong: a deleted lock file would become a parse note instead of a quiet removal
  - master-was-wrong: none; before the edit Check passed Go, NPM and PyPI only, as the order said
  - session recording: none in this session; no ADR-0024 recording path was given
  - commit message has no Co-authored-by; `git log -1` shows b7189c17484021b8d2d58c9b8d3aaa1af3ff1f2c
- note: master-was-wrong: none. The proven CLI was refused (allowed form none), so the sample-scan stdout is not pasted; the in-process sample test passed. No session recording was given.
- updated: 2026-10-23T01:43:44Z

CLOSE-OK · lock-files-are-checked · commit b7189c17484021b8d2d58c9b8d3aaa1af3ff1f2c · branch fix/lock-files-are-checked · worktree ██████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
