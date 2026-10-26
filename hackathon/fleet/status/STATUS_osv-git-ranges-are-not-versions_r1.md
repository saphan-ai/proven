# STATUS — osv-git-ranges-are-not-versions · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 12 min
- last commit: 33e6bfac146b7685bf57bd60a253020dd3a91aa6 [osv-git-ranges-are-not-versions] osv: skip GIT ranges (the leg's 2507105 with the tool's Co-authored-by trailers removed by the master; tree identical)
- updated: 2026-10-23T05:57:58Z

## Summary

GIT ranges are not version rows, a versions list remains when they are the only ranges, and semver refuses a number that does not fit in int.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — osv-git-ranges-are-not-versions
- state: stop-2-ready
- abstract: GIT ranges are not version rows, a versions list remains when they are the only ranges, and semver refuses a number that does not fit in int.
- last commit: 33e6bfac146b7685bf57bd60a253020dd3a91aa6 [osv-git-ranges-are-not-versions] osv: skip GIT ranges (the leg's 2507105 with the tool's Co-authored-by trailers removed by the master; tree identical)
- done: GIT ranges stay out of the table and an overflowing semver number is refused
  - TestReadDirSkipsGitRangeBesideSemver was red on the base (2 rows, the GIT hashes first) and passed after the fix
  - TestReadDirKeepsVersionsWhenRangesAreOnlyGit was red on the base (1 commit-hash row, the versions list dropped) and passed after the fix
  - TestReadDirKeepsEcosystemRangeBesideGit was red on the base (GIT row plus the ECOSYSTEM row) and passed after the fix
  - TestParseRefusesANumberThatDoesNotFit was red on the base: Parse("99999999999999999999.0.0") returned Major 7766279631452241919
  - Negative control admitted range type GIT. TestReadDirSkipsGitRangeBesideSemver exited 1. The guard was restored
  - Negative control removed the overflow check in number. TestParseRefusesANumberThatDoesNotFit exited 1. The check was restored
  - go build ./... exit 0. go vet ./... exit 0. go test -count=1 ./... exit 0 after the restore
- next / blocked-on: none
- artifacts:
  - internal/osv/convert.go
  - internal/osv/convert_test.go
  - internal/semver/semver.go
  - internal/semver/semver_test.go
  - docs/OSV_TABLE.md
- evidence:
  - baseline before any edit: go test -count=1 ./... exit 0. Every package printed ok or [no test files]. No package failed
  - toolchain: go version go1.26.8 linux/amd64. go.mod says go 1.26 and has no toolchain line. No Makefile and no workflow file were found
  - red: go test -count=1 -v ./internal/osv ./internal/semver with the four new tests, exit 1, before any production edit
  - final: go build ./... exit 0. go vet ./... exit 0 with no output. go test -count=1 ./... exit 0. Every package printed ok or [no test files]
  - commits: 5bc916e77eac20b560cfc1cb8d78d5ae3b7d3ac5 semver, then 2507105a2bbe16eebec9da50511f01501d60d744 osv and docs
  - go test -race UNRUN. This change adds no goroutine and no shared mutable state
  - testing bar UNRUN. No metric, command or baseline is declared
  - per-test pass, fail and skip counts UNRUN. The summary command does not print them. No package failed
  - session recording ADR-0024: none. No recording was provisioned for this seat and none is published in this repository for this run
- note: master-was-wrong: none. Both commits end with a Co-authored-by trailer the tool added; it was not amended, because amend puts the trailer back, and the master removes it. Rulings: a range type other than SEMVER or ECOSYSTEM is skipped, the match is case-sensitive, and the versions list is used only when no version span remains. The ECOSYSTEM test was added so dropping ECOSYSTEM cannot pass the GIT tests. A numeric pre-release that does not fit is refused at Parse. date is denied for this role, so updated is the HEAD committer time.
- updated: 2026-10-23T05:38:00Z

CLOSE-OK · osv-git-ranges-are-not-versions · commit 2507105a2bbe16eebec9da50511f01501d60d744 · branch fix/osv-git-ranges-are-not-versions · worktree ███████████████████████████████████████████████████████████████████

remark: a package line of ok is the same when a test skipped. This summary did not count skips, so a skip is not distinguished from a pass.
remark: the wrapped major 7766279631452241919 was measured on go1.26.8 linux/amd64. A 32-bit int would wrap to a different number. The refusal is that Parse returns false.
remark: the GIT negative control admitted the type GIT and changed nothing else. A defect that dropped SEMVER would not turn that test red.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
