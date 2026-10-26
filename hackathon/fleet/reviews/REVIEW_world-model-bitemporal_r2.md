# REVIEW — world-model-bitemporal · round 2

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T21:55:05Z

## Summary

Two repairs hold for open edges, but bounded edges bypass renewal and permit overlapping history; external conformance was not rerun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — world-model-bitemporal

### Ground
Reviewed `13349e5...e91cec0` and repair `d7ffd71...e91cec0`; `git rev-parse --short=7 HEAD` returned `e91cec0`. Go module pins 1.26.0. No dependency files changed. Recorded scorecard says World Model 8/8 (`test/conformance/meterless/SCORECARD_world-model.txt:13`); README (`README.md:124`, `README.md:163`), architecture (`docs/ARCHITECTURE.md:70`), and design (`hackathon/DESIGN.md:129`) agree. This is the committed score, not my independent run. The adapter forwards `validFrom` and `validTo` (`cmd/meterless-conform/main.go:382`); it checks malformed query timestamps (`cmd/meterless-conform/main.go:203`). Source search found no engine-owned model/provider/network call in `internal/world` or the conformance adapter.

### Differential
BLOCKING — original repeat-target overlap remains reachable. `internal/world/graph.go:89`: `Relate` calls `place` only when incoming `ValidTo` is empty. `internal/world/graph.go:46`: `sameChain` excludes an originally bounded edge (`ValidTo != ""`, `SupersededBy == ""`) even if the new edge is open. Write A from 2087-01-01 until 2089-01-01; write A again, open, from 2088-01-01 with the same from/type/context. The second has a different ID, skips the first in `place`, and is stored. At 2088-06-01, `Neighbors` (`internal/world/graph.go:214`) returns the same target twice. That violates renewal and gives duplicate point-in-time neighbors. Named remedy: test a bounded first edge followed by same-target renewal, expecting one interval and one neighbor; validate or refuse overlaps for bounded edges. `TestRepeatedTargetRenewsOneEdge` (`internal/world/graph_test.go:141`) covers only an open first edge. Finding 1: NOT ADDRESSED on this public bounded-edge path. Reproduced with a temporary Go caller under `████████████`: `go run` returned `distinct_ids=true neighbors_at_2088=2`.

Finding 2 ADDRESSED for open-edge reverse insertion: `place` chooses the earliest later start and bounds the inserted edge (`internal/world/graph.go:156`, `internal/world/graph.go:185`); `TestReverseTimeInsertionKeepsIntervalsOrdered` checks before, within, and after (`internal/world/graph_test.go:160`). Finding 3 ADDRESSED: writes parse and canonicalize instants and refuse invalid bounds (`internal/world/graph.go:84`, `internal/world/instant.go:14`); reads compare parsed instants (`internal/world/graph.go:204`); offset and malformed-input tests are `internal/world/graph_test.go:184` and `internal/world/graph_test.go:209`.

### Unplug
The repair tests assert repeated open writes, reverse insertion, equivalent offsets, and invalid timestamps. Source inspection indicates removing the close/bound branches changes the asserted neighbors or edge counts; removing timestamp parsing changes the offset assertions. Mutation harness and base-commit red run were not run. No existing test reddens on the bounded-edge overlap; in that false world the suite still passes.

### Closure
`go test -count=1 ./internal/world ./cmd/meterless-conform` passed (both packages). `go test -count=1 ./...` failed in `internal/store`: `TestViewsReturnTheSampleScenarioRecords` reported `node is not installed`. That failure is outside this write set and does not establish a product regression. `git diff --check 13349e5...e91cec0` passed. Ingest idempotency and resume tests (`internal/world/world_test.go:82`, `internal/world/world_test.go:107`) passed in the targeted run; the committed scorecard lists `idem-001` and `idem-002` PASS. No test suppression, removed assertion, or altered threshold was found in the diff.

### Remainder
The minimum graph has valid-time intervals but no separate transaction-time query; a fuller bitemporal API beyond the owner's `bitemp-001` minimum is declined to judge. `Open` checks edge provenance and duplicate IDs (`internal/world/world.go:213`), not interval integrity in persisted data; that is owed work, not a new round under R14. Ruling: bounded edges are within repeat-target scope because `relateW` accepts `validTo`. If they instead express independent concurrent facts, the owner must define that alternate semantics; cost if wrong is an unnecessarily restrictive validator. No exception was added, used, or expired.

### Profile
Backend/data persistence and docs. Query scans all stored edges (`internal/world/graph.go:214`); no database index applies. Reopen/history coverage is `internal/world/graph_test.go:112`. No new authorization boundary or network I/O. Check ranking: the Meterless runner is external-project conformance, but only its committed scorecard was available; Go tests are author-owned suite checks. No coverage or mutation metric baseline was declared, so percentages are UNRUN, not zero. Per-instance concurrency was not run.

### Focus
Re-review of three prior findings and `d7ffd71...e91cec0`; the bounded-edge path leaves original finding 1 reachable. The other two prior paths are addressed for open edges. Full-diff inspection checked the adapter and stated scores; unrelated deferred H-MEM features were declined to judge.

### Conditions
Reject until original finding 1 handles or refuses explicitly bounded overlap with a regression that goes red before repair and passes after. Re-run external conformance and the full Go suite on a host with Node available. This is not an owner gate decision.

### Unrun
- External Meterless conformance: no configured runner checkout; 8/8 is the committed scorecard only.
- Mutation and base-commit red checks, not performed; bounded-edge behavior was reproduced by the Go caller.
- Complete full Go suite: attempted, but `internal/store` failed because Node is unavailable; completed package results, including `internal/world`, passed.
- Cleanup of the temporary reproduction under `██████████████████████████████████████████████` was refused by the command guard; it remains untracked, and I did not bypass that refusal.

## The review status as filed

## REVIEW — world-model-bitemporal
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Two repairs hold for open edges, but bounded edges bypass renewal and permit overlapping history; external conformance was not rerun.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T21:50:49Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
