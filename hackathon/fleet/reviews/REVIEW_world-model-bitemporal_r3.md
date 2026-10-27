# REVIEW — world-model-bitemporal · round 3

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 2 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T23:48:21Z

## Summary

Round 4 closes the bounded-edge overlap on fresh stores. Go tests pass; the external Meterless runner and a live unplug were not run on this seat. The master must verify those before the gate.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — world-model-bitemporal

### Ground

Reviewed `13349e5...b112aff` at `b112aff80e920a80a7e348a861ee636fd512078e` (`git rev-parse --short=7 HEAD` printed `b112aff`). The round 4 delta is `internal/world/graph.go` and `internal/world/overlap_test.go`. The order's earlier measurements and score are claims, not this review's measurements. `go.mod` pins Go 1.26.0; this seat ran Go 1.26.5. The Makefile has lab targets but no Go test gate; the repository's established Go checks are build, vet and `go test -count=1 ./...`.

`go test -count=1 ./internal/world ./cmd/meterless-conform` exited 0 (both packages `ok`). `go test -count=1 ./...` exited 0, with no failed package. `go vet ./internal/world ./cmd/meterless-conform` exited 0. These are suite checks controlled by the repository; the Meterless runner is the external check and was not independently run here. No dependency manifest or lockfile changed. No test, threshold, or suppression was weakened in the diff. There is no declared numeric coverage or mutation bar to compare.

remark: The Go run proves the local tests pass with Go 1.26.5; it cannot prove the recorded Meterless 8/8 result against its separate pinned runner, and a passing suite can miss an untested interval.

### Differential

Review 1, same-target renewal: ADDRESSED on fresh stores. `internal/world/graph.go:196-199` returns the covering same-target edge instead of adding an overlap. `TestRepeatedTargetRenewsOneEdge` checks a later different target and the two-edge history.

Review 1, reverse-time insertion: ADDRESSED. `internal/world/graph.go:179-210` finds the next start or splits a covering interval; `TestReverseTimeInsertionKeepsIntervalsOrdered` checks both paths. Review 1, offset timestamps: ADDRESSED. `internal/world/instant.go:19-40` parses RFC 3339 instants and stores canonical UTC; `TestEquivalentOffsetsAreTheSameInstant` checks equal instants and boundary queries.

Review 2, bounded-edge overlap: ADDRESSED for new writes. `internal/world/graph.go:129-137,214-240` tests every stored edge sharing `(from,type,context)`, including explicitly bounded edges skipped by `place`; a clash restores the snapshot and returns `ErrEdgeOverlap`. `internal/world/overlap_test.go:22-83` covers open and bounded renewal, another target, and a touching boundary. Refusing a write that contradicts a caller-bounded fact is a sound rule: it preserves the caller's recorded interval and the at-most-one-valid-edge invariant rather than silently rewriting that fact.

The scorecard in `test/conformance/meterless/SCORECARD_world-model.txt` says 8/8. `README.md:124,163`, `docs/ARCHITECTURE.md:70`, and `hackathon/DESIGN.md:129` use the same World Model figure. `internal/world/world.go:247-251` only stores and returns the injected generator; the new graph and adapter paths call no model, network endpoint, or provider. Ingest's existing skip and resume path at `internal/world/world.go:307-365` is unchanged, and its tests still pass.

### Unplug

The local tests contain a specific alarm for removal of the closing assignment: `internal/world/graph_test.go:53-71` requires the old edge's `ValidTo` to equal the new start and `SupersededBy` to name the new edge. The round 4 tests require `ErrEdgeOverlap` for bounded conflicts (`internal/world/overlap_test.go:27-29,51-58,79-80`). A live source mutation and rerun were not performed because this review leg is forbidden to change the code under review. Thus the alarm is read in the test, not observed here under mutation.

### Closure

Recommendation: **accept-with-conditions**. The four earlier findings are addressed for fresh stores. The conditions below are verification gates, not another code repair round. The single-chain invariant is preserved by the tested current write paths; an explicit `at` query filters `[ValidFrom,ValidTo)` at `internal/world/graph.go:264-286` and closed edges remain in `state.json` (`internal/world/world.go:495-502`).

### Remainder

OWED, outside the round 4 fix: `internal/world/world.go:216-224` loads persisted edges after checking provenance and duplicate IDs but does not validate intervals against each other. A state file written by `e91cec0` with the bounded/open overlap is still readable by `b112aff`; `Neighbors` at `internal/world/graph.go:274-302` can then return both edges for an instant in that overlap. The pre-fix code at `e91cec0:internal/world/graph.go` inserted the edge without an overlap check. This matters if such a pre-fix state must be supported. Remedy: validate or migrate loaded edges before exposing them to queries, with a fixture made by the former writer. I did not classify it as a fresh-write blocker because the rejected revision was not landed as the product baseline.

OWED, outside the round 4 fix: an open edge with a future `ValidFrom` is returned by `Neighbors(..., at="")` because that branch checks only `ValidTo` (`internal/world/graph.go:288-289`). The API calls this view "now" (`internal/world/graph.go:257-260`). If "now" means the clock instant, use the clock to filter the start too; if it means the latest open edge regardless of date, name that semantic in the API. This review does not infer which interpretation the upstream runner intends.

Declined to judge: historical state migration policy, the upstream Meterless runner's implementation, and a performance bound; none was available as a measured contract in this worktree. Ruling: treat pre-fix state compatibility as owed rather than blocking because no rejected round was merged. Cost if wrong: an existing state can still return two neighbors after upgrade.

### Profile

Backend and docs. New graph reads an in-memory map and persists through the existing atomic state-file path. Each `Relate` scans stored edges, so the work per write grows with edge count; no bounded data-size or latency target is declared. The review found no model or network call in the added code. The changed docs stand on the committed scorecard; the separate runner was not available to corroborate it here. No new dependency, install script, skip, or coverage exclusion appears in the eleven-file diff.

### Focus

For each earlier finding, I read the public write path `Relate`, its `place` and `overlapping` helpers, the read path `Neighbors`, persistence and reopening, the adapter entry point, and the tests. The review uses the three-dot change range. All four commit subjects match the required shape and length; no AI attribution appears in those messages.

### Conditions

Before the owner gate, the master must run the pinned Meterless World Model conformance runner at `b112aff` and confirm its actual 8/8 output against the committed scorecard. The master must also perform a disposable negative control that removes the closing assignment at `internal/world/graph.go:204` and observes `TestSupersedingEdgeClosesTheOldOne` fail, then restore the source and rerun the test. These conditions discharge the two direct observations this seat did not make. No reclassification of a prior finding is proposed.

### Unrun

The external Meterless conformance runner was not run: its separate `engines/` checkout is not part of this review worktree, and `run.sh` writes scorecards into tracked source. A base-commit Go test run and a live source-mutation negative control were not run: this leg was assigned read-only review of the tip and cannot change the code under review. `go build ./...`, the full-repository `go vet ./...`, race tests, and exact per-test pass/skip counts were not run; the change has no concurrency path, and the targeted vet plus full uncached test suite cover the current review risk.

## The review status as filed

## REVIEW — world-model-bitemporal
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Round 4 closes the bounded-edge overlap on fresh stores. Go tests pass; the external Meterless runner and a live unplug were not run on this seat. The master must verify those before the gate.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T23:45:25Z

## REDACTED

none
