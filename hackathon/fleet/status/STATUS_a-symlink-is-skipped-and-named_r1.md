# STATUS — a-symlink-is-skipped-and-named · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: ef7500d24e86e141790a1a50df40fbb2b6ef18ab [a-symlink-is-skipped-and-named] change: skip tracked symlinks (the leg tip 73c7e6fef3f42c89a22a3a8e2611268e61b82e55 with the tool trailer removed by the master; tree identical)
- updated: 2026-10-23T09:07:26Z

## Summary

A live review that tracks a symlink skips the link, names it once on the trace and on Result.SkippedPaths, and still reviews the changed regular file. change.Read still refuses a symlink in a prepared tree.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-symlink-is-skipped-and-named
- state: stop-2-ready
- abstract: A live review that tracks a symlink skips the link, names it once on the trace and on Result.SkippedPaths, and still reviews the changed regular file. change.Read still refuses a symlink in a prepared tree.
- last commit: ef7500d24e86e141790a1a50df40fbb2b6ef18ab [a-symlink-is-skipped-and-named] change: skip tracked symlinks (the leg tip 73c7e6fef3f42c89a22a3a8e2611268e61b82e55 with the tool trailer removed by the master; tree identical)
- done: A tracked symlink is skipped and named, and the review finishes.
  - commit mode names linked.txt and keeps the target content out of the prompt, the trace and the result; the final go test -count=1 ./... exit 0 includes TestLiveGitReviewSkipsATrackedSymlink
  - working-tree mode does the same; that suite includes TestLiveWorktreeReviewSkipsATrackedSymlink
  - negative control restored the refusal in untar: go test -count=1 -v ./internal/review -run TestLiveGitReviewSkipsATrackedSymlink exit 1, failure line change: symlink linked.txt; the skip was restored before the final suite
  - change.Read still refuses a symlink: the final go test -count=1 ./internal/change printed ok, and that package holds TestReadRefusesASymlink
- next / blocked-on: none
- artifacts:
  - docs/ARCHITECTURE.md
  - internal/agent/writer/writer.go
  - internal/candidate/candidate.go
  - internal/change/archive.go
  - internal/change/read.go
  - internal/pipeline/pipeline.go
  - internal/pipeline/trees.go
  - internal/review/git_live_test.go
  - internal/review/open.go
  - internal/review/review.go
- evidence:
  - go version go1.26.8; go.mod says go 1.26 and has no toolchain line
  - no Makefile and no .github workflow; the order's three commands are the gate, run separately
  - baseline before any edit: go test -count=1 ./... exit 0, every package ok or no test files
  - RED before the production edit: go test -count=1 -v ./internal/review -run TestLiveGitReviewSkipsATrackedSymlink exit 1 and the worktree twin exit 1, both change: symlink linked.txt
  - after the fix, the same -v command exit 0, both tests PASS
  - go build ./... exit 0 on the final tree
  - go vet ./... exit 0 on the final tree, no output
  - go test -count=1 ./... exit 0 on the final tree; summary lines are the package ok lines from that run
  - go test -race UNRUN: this change adds no goroutine and no new shared state
  - testing bar UNRUN: no metric, command or baseline is declared; no bar-weakening move found
  - per-test pass, fail and skip counts for the full suite UNRUN: go test without -v or -json does not print them; no package failed
- note: master-was-wrong: no. Co-authored-by was added by the commit tool on 73c7e6fef3f42c89a22a3a8e2611268e61b82e55; the master removes it. Session recording (ADR-0024): none was provisioned for this seat.
  - ruling: Result.SkippedPaths is a new field because no existing field is a list of unread paths. Cost if wrong: a reader of another field does not see the paths.
  - ruling: one trace event per distinct path, sorted, kind skipped, text symlink, empty finding id. Cost if wrong: a reader cannot tell which side held the link.
  - ruling: the shell printer was not changed. Cost if wrong: the CLI text does not list the paths; the result and the trace do.
  - remark: the during-Planner tree walk flags an entry named linked.txt and regular-file bytes that hold the target content. A follow that stored those bytes under another name is still caught by the prompt and the trace scan.
- updated: 2026-10-23T08:58:24Z

CLOSE-OK · a-symlink-is-skipped-and-named · commit 73c7e6fef3f42c89a22a3a8e2611268e61b82e55 · branch fix/a-symlink-is-skipped-and-named · worktree ██████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
