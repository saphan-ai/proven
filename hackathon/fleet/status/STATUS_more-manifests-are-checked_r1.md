# STATUS — more-manifests-are-checked · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 14303b7c6d2147d445978d087b7f63c1e111c4e3 [more-manifests-are-checked] osv: check npm and requirements.txt
- updated: 2026-10-23T01:17:52Z

## Summary

package.json and requirements.txt are checked against OSV the same way go.mod is. npm uses semantic versions. PyPI compares PEP 440 release, pre-release and post-release forms. An unparsed version is not a finding, and the trace says why.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — more-manifests-are-checked
- state: stop-2-ready
- abstract: package.json and requirements.txt are checked against OSV the same way go.mod is. npm uses semantic versions. PyPI compares PEP 440 release, pre-release and post-release forms. An unparsed version is not a finding, and the trace says why.
- last commit: 14303b7c6d2147d445978d087b7f63c1e111c4e3 [more-manifests-are-checked] osv: check npm and requirements.txt
- done: npm and PyPI packages added or raised are checked against the OSV table
  - go test -count=1 ./... on the tip, exit 0
  - negative control TestRequirementsDiffAddsAndRaises went red when pinned lines were dropped, then green after restore
- next / blocked-on: none
- artifacts:
  - internal/pep440/pep440.go
  - internal/pep440/pep440_test.go
  - internal/manifest/diff.go
  - internal/manifest/gomod.go
  - internal/manifest/gomod_test.go
  - internal/manifest/npm.go
  - internal/manifest/npm_test.go
  - internal/manifest/requirements.go
  - internal/manifest/requirements_test.go
  - internal/osv/check.go
  - internal/osv/check_test.go
  - internal/osv/range.go
  - internal/osv/range_test.go
  - docs/OSV_TABLE.md
  - docs/ARCHITECTURE.md
- evidence:
  - go version go1.26.8 linux/amd64. go.mod says go 1.26 and has no toolchain line. Local toolchain matches the pin.
  - baseline before any edit, at ee6e55f730e4138ef735b0c4eaf8d9121edba3d7: go test -count=1 ./... exit 0. Every package printed ok or [no test files]. No package failed.
  - RED before the symbols existed: go test -count=1 ./internal/pep440 ./internal/manifest exit 1. The compiler named undefined Compare, Parse, Reject, diffNPM, npmNotes, NPM, diffRequirements and requirementsNotes.
  - negative control. Mutated field: pin acceptance in parseRequirements. Every line containing == was dropped. go test -count=1 -v ./internal/manifest -run TestRequirementsDiffAddsAndRaises exit 1. requirements_test.go:22 printed []manifest.Change(nil). FAIL TestRequirementsDiffAddsAndRaises.
  - restored. go test -count=1 ./internal/manifest -run TestRequirements exit 0.
  - tip go build ./... exit 0. tip go vet ./... exit 0, no output. tip go test -count=1 ./... exit 0. Every package printed ok or [no test files]. No package failed. Per-test skip counts were not taken.
  - earlier commit cbd977131ffbe2b56f759eccdfccdf0ac149e966 [more-manifests-are-checked] pep440: compare release, pre and post. No Co-authored-by on either commit.
  - sample scan. The proven binary is not in the executor allow list. allowed form: none. The binary was built and not executed. TestScanSamplePrintsF1AndF2NotF3 passed inside the tip suite. It calls run from a temp directory with scan --sample testdata/fixture/discount. It requires F1, F2, F4, the OSV finding for evil.example/leftpad, and failed services: none, and it fails if F3 is printed. The CLI stdout itself was not produced.
  - go test -race UNRUN. This change adds no goroutine and no shared mutable state.
  - testing-bar UNRUN. No metric, command or baseline is declared.
- note: master-was-wrong: no. W13's comparer is semantic versions. internal/osv/range.go InRange uses semver.Parse and semver.Compare only. npm reuses that comparer and does not coerce a range or a two-part number. PyPI accepts a PEP 440 release, pre-release (a, b, rc and alpha, beta, rc, pre, preview, c) and post-release (post, rev, r, and the -N form). An epoch, a dev release, a local version or a wildcard is refused with a sentence. 1.0 and 1.0.0 compare equal. requirements.txt blank lines and full-line comments are not traced; every other non-pin line is. Package names are not case-folded. When dependencies and devDependencies both name one package, the later section wins. No new setting. PROVEN_OSV_TABLE is unchanged, and the settings drift tests stayed green. docs/ARCHITECTURE.md now names the three manifests; the order named docs/OSV_TABLE.md, and the architecture sentence would otherwise still say only go.mod. Session recording (ADR-0024): none. The run directory evidence, provision and state directories are empty. No recording was provisioned. updated is the tip commit's committer time, 2026-10-22T17:23:01Z. The tip suite ran after that commit.
- updated: 2026-10-22T17:23:01Z

CLOSE-OK · more-manifests-are-checked · commit 566f5c34c8a35994fd55ef5c6fbb88dd49c9058a · branch fix/more-manifests-are-checked · worktree ██████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
