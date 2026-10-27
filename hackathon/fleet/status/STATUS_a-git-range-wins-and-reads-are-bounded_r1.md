# STATUS — a-git-range-wins-and-reads-are-bounded · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 8726c4e [a-git-range-wins-and-reads-are-bounded] absorb main (the master: C1 of the review fixed in c74d2a4, a size bound above 1 TiB refused; then main absorbed, clean, 34 packages ok)
- updated: 2026-10-23T17:47:45Z

## Summary

With a baseline commit the git range always decides the trees; every read of repository content (archive stream, one file, proven.yaml, the author record, git output) has a named size bound and is refused over it; ExportCommit checks its revision, and a dash-leading revision is refused by every reader. Written by the MASTER: the leg (run-20261023-125102-a28d8295) was killed by a session restart after its three commits and before its status.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-git-range-wins-and-reads-are-bounded
- state: stop-2-ready
- abstract: With a baseline commit the git range always decides the trees; every read of repository content (archive stream, one file, proven.yaml, the author record, git output) has a named size bound and is refused over it; ExportCommit checks its revision, and a dash-leading revision is refused by every reader. Written by the MASTER: the leg (run-20261023-125102-a28d8295) was killed by a session restart after its three commits and before its status.
- last commit: 8726c4e [a-git-range-wins-and-reads-are-bounded] absorb main (the master: C1 of the review fixed in c74d2a4, a size bound above 1 TiB refused; then main absorbed, clean, 34 packages ok)
- done: the three items of the order, each with tests
  - TestBaselineCommitDecidesTheTreesNotTheRepositoryDirectories and TestPreparedTreesAreReadWhenNoBaselineIsGiven
  - TestDashLeadingRevisionIsRefusedByEveryReader; the file, archive, config and record bound tests
  - negative control 1 (master): prepared trees allowed to win again with a baseline -> TestBaselineCommitDecidesTheTreesNotTheRepositoryDirectories red; restored
  - negative control 2 (master): the file bound lifted in change.Read -> TestReadRefusesAFileOverTheFileBound red; restored
  - go build ./... exit 0; go vet ./... exit 0; gofmt clean; go test -count=1 ./... 33 packages ok (master, macm5, go1.26.8)
- next / blocked-on: the named review
- artifacts:
  - internal/pipeline/trees.go
  - internal/change/bounds_test.go
  - internal/pipeline/trees_range_test.go
  - internal/review/bounds_test.go
  - docs/CONFIGURATION.md
- note: master-was-wrong: not measured by the leg (its status was never written); the master re-measured every item at the tip. No step-1 red output was recorded by the leg. No session recording (ADR-0024).
- updated: 2026-10-23T13:50:00Z

CLOSE-OK · a-git-range-wins-and-reads-are-bounded · commit 0a4341c · branch fix/a-git-range-wins-and-reads-are-bounded · worktree ██████████████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
