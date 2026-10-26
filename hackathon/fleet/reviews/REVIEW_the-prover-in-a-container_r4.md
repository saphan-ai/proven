# REVIEW — the-prover-in-a-container · round 4

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T05:05:47Z

## Summary

Codex review of bc41431: failed-removal warning survives long output on both timed-out sides; cancellation paths and the red mutant checked. Accept; real Docker T10 skipped.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-prover-in-a-container

### Ground
Reviewed `e25d5c5...bc41431`, with re-review focus `c91f828..bc41431`. The tip was `bc41431` when measured. `go.mod` pins Go 1.26; local `go version` reported go1.26.5. No Makefile, task runner, linter configuration or CI workflow was found. The test author's object `33c0dcf` was available locally; no fetch was needed. Review against the container contract and the actual `Container.Run`, `side`, `formatSides`, `trimOutput`, and cancellation paths, not the master's assessment.

### Differential
ADDRESSED — `internal/testrun/container.go:329-338` now puts the failed-removal note before the runtime text. `formatSides` trims each side separately to 8192 bytes at `internal/testrun/container.go:386-405`; the note therefore survives long runtime text on both the change and base sides. The new `internal/testrun/container_test.go:506-528` returns 20000 bytes on each timed-out side and checks for two surviving notes. It passed. Cancellation returns a wrapped `context.Canceled` error through `withRemoval`, without applying `trimOutput`; a review-only two-case test returned 20000 bytes with failed removal on either the change or base side and passed both cases. No production code changed beyond `joinRemoval`. Its separator is now an unconditional newline when text exists, so a runtime output already ending in a newline produces an extra blank line after the removal note; no proof or safety behaviour changes. No changed test was removed or weakened. The master's dated account of the scope and red/green outcome matched the measured diff and mutant run.

### Unplug
Mapped `internal/testrun/container.go` to its `c91f828` blob in a read-only Go overlay, preserving the new test. `go test -overlay ... -count=1 -run '^TestRemovalFailureSurvivesLongOutput$' -v ./internal/testrun` failed as expected: the assertion reported 16540 bytes and 0 surviving notes. On `bc41431`, that same test passed. The review-only cancellation probe passed for both change-side and base-side cancellation with 20000 bytes of runtime output and a failed removal. This proves the ordering guard is observed by its new regression test; the two cancellation cases inspect the separate error path.

### Closure
`internal/testrun/container.go:329-338` — previous BLOCKING finding ADDRESSED. For ordinary long runtime output, the remedy places the removal warning ahead of truncated text on both timed-out sides. No new blocking or owed finding in the focus diff. Recommendation: accept.

### Remainder
The removal warning itself is not separately bounded; an unusually long removal error could be truncated before its full detail, although its `container <name> may remain` prefix remains visible. No additional follow-up is required by this re-review. Declined to judge: licensing, host image-pull policy, and unmodified container-runtime isolation behaviour; outside this fix diff.

### Profile
Security profile for container cleanup and confinement; backend profile for external runtime failure reporting. The failed-removal boundary is observable in the returned outcome or cancellation error and cannot become proof. `go test -overlay ... -count=1 -v ./internal/testrun`: 26 passed, 1 skipped (T10 requires an explicitly enabled real runtime), 0 failed. `go test -overlay ... -count=1 ./...`: all tested packages passed. `go vet ./...` and `git diff c91f828..bc41431 --check` exited 0. No declared metric/baseline, exception, dependency change or bar-weakening move. Check ranking: project vet, then suite and mutant; the suite alone is circular and not an external audit. Model: Codex; exact model identifier was not exposed to this leg.

### Focus
The re-review was limited to the previous failed-removal visibility finding and `c91f828..bc41431`. The Go overlay used a local test-author blob, not a tracked source edit. Review probes and logs remained in `█████████████` and were not returned as source artifacts. Ruling: an added blank separator is presentation-only; if this is wrong, a consumer matching exact output bytes could observe a changed string.

### Conditions
none

### Unrun
Real Docker T10 was skipped by the test author's explicit opt-in guard; no real container was started. No live-runtime integration, race detector, or base-revision full-suite rerun was performed; none was required to establish the ordering regression in this re-review.

## The review status as filed

## REVIEW — the-prover-in-a-container
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: Codex review of bc41431: failed-removal warning survives long output on both timed-out sides; cancellation paths and the red mutant checked. Accept; real Docker T10 skipped.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T05:01:37Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
