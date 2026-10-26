# REVIEW — world-model-bitemporal · round 4

- verdict: accept-with-conditions
- reviewer: qwen-code, glm-5.3
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-26T04:06:01Z

## Summary

Round 4 holds: all four earlier findings are repaired and unreachable on the public write path, refusing an overlap is sound for the specification, and the three named docs state 8 of 8 as the committed scorecard prints. Zero blocking findings; four NOTE findings owed (now-query skips a bounded edge valid at the present; a repeat write's differing valid_to is ignored; Open does not parse-check edge times; the infra README row still prints 7 of 8). accept-with-conditions: go is denied on this seat, so every go run is Unrun and no accept is legal; the gate names the master's measured 8 of 8 run at b112aff as the executed proof, or the suite is re-run before merge.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — world-model-bitemporal

### Ground

- Reviewed tip, read from the worktree with `git rev-parse --short=7 HEAD` before any act: `b112aff`. It matches the order's `- reviewed-sha: b112aff` and its base line `b112aff80e920a80a7e348a861ee636fd512078e`.
- Scope read in full: `13349e5...b112aff` — 4 commits (c9e12f2, d7ffd71, e91cec0, b112aff), 11 files, +751/−13. Every file was read at the tip; `internal/world/world.go` was read both at the tip and as a diff against the base.
- Toolchain discovered (C1): `go.mod` pins `go 1.26.0`; there is no `go.work`; the Makefile holds build and lab targets and no test target; the project's test entry points are `go test ./...` and `test/conformance/meterless/run.sh`, which builds `cmd/meterless-conform`, needs `node` and a `METERLESS_ENGINES` checkout of Meterless's engines, and writes the scorecards.
- The `go` program is denied on this seat. The single-program attempt and its refusal are quoted verbatim under Unrun. That is a permissions fact, not a missing binary: it was tried once as one program and then stopped; no grant was self-made and no other spelling was used. Every command that needs `go` is UNRUN; this review rests on the diff, and the executed suite proof on record is the master's.
- The order's ground, quoted as the master's claim and not re-measured here: Meterless's runner at b112aff (pinned 1742dec5) → world-model 8/8, whole suite 54 ok, 0 FAIL; round 4's three tests fail on e91cec0's graph code and pass at b112aff.
- Two independent reviews run on this tip in parallel; the other review's notes were not read, as the order directs.
- Callers of the new API, grepped across the repository: `Relate` and `Neighbors` are called by the `internal/world` tests and by `cmd/meterless-conform` only. No pipeline code calls them, and the ingest path writes no edges.

remark: the SHA above is one read of this worktree on 2026-10-26; a commit made later on this branch would not appear in it.
remark: the go refusal is a permissions fact on this seat; a green run on another seat proves nothing about this seat, and this review claims no run.
remark: the 8/8 and 54-ok figures are the master's measurement quoted from the order; the wrong reading they invite is treating them as this review's own runs.

### Differential

The order's five properties, judged against the diff:

1. Supersession closes and keeps — met. `place` (internal/world/graph.go:165) closes the covering edge at the new edge's start and points it at the new edge (graph.go:204); nothing is deleted, and `Canonical`, `persist` and `Open` all carry the edges (world.go: the edges map, the disk field, the Open validation, `foldSnap`, `cloneState`, `restoreState`). `Neighbors` (graph.go:264) answers an `at` query as ValidFrom ≤ at < ValidTo, compared as instants. `TestSupersedingEdgeClosesTheOldOne`, `TestPointInTimeQueryRespectsValidity` and `TestEdgesSurviveReopen` pin the three facts.
2. Ingest idempotent and resumable — met. The world.go delta is additive only: the edges map, its disk field, its Open validation, and its place in `Canonical`, `cloneState` and `restoreState`. `Ingest` and `apply` are untouched and fold no edges, so `idem-001`/`idem-002` behaviour is unchanged; a failed pass now restores the edges with the rest of the fold. The committed scorecard prints both checks PASS.
3. Tests cover the bitemporal rules — met by reading; the executed mutation proof is UNRUN (Unplug). The assertions name `ErrEdgeOverlap`, the closed edge's `ValidTo` and `SupersededBy`, exact neighbour sets and edge counts, so each guard has a test whose failure names the removed rule.
4. No engine code calls a model, a network endpoint or a provider — met. Every changed Go file was read in full: the imports are standard library plus `internal/trace`, `internal/hmem`, `internal/generator` and `internal/pathguard`. The generator is stored by `SetGenerator` and never called; `cmd/meterless-conform` states "No network" and its imports confirm it.
5. The docs state 8 of 8 — met for the three docs the order names: README.md:124 and README.md:163, docs/ARCHITECTURE.md:70 and hackathon/DESIGN.md:129 all state 8 of 8 (ARCHITECTURE and DESIGN add the runner pin 1742dec5), matching the committed scorecard, which prints "8/8 checks passed" with bitemp-001 PASS. No other World Model figure appears in them.

The four earlier findings at this tip, each ADDRESSED:

- Review 1, same-target renewal left two open edges: `place` returns the covering edge when its target matches (graph.go:197-198) and stores nothing; `TestRepeatedTargetRenewsOneEdge` holds the count at two edges after A, A, then B.
- Review 1, out-of-order insertion closed an edge before it began: a write that starts before a later edge takes that edge's start as its end (graph.go:208-209); `TestReverseTimeInsertionKeepsIntervalsOrdered` also asserts no edge ends before it begins.
- Review 1, offset timestamps compared as text: `ParseInstant` canonicalises every bound to UTC (internal/world/instant.go), and the edge id, the stored bounds and the comparisons all use the canonical form; `TestEquivalentOffsetsAreTheSameInstant` pins the equivalence and the one-edge outcome.
- Review 2, the overlap stayed reachable when the first edge was bounded: `Relate` now runs the `overlapping` check after `place` on every write, open or bounded (graph.go:133), against any earlier edge of the same from, type and context — caller-bounded facts included, because `overlapping` (graph.go:219) does not filter on `sameChain`. A refusal restores the snapshot taken before `place`, so nothing changes. The three round-4 tests pin bounded-then-open, bounded-then-bounded in three shapes, and open-over-bounded-of-another-target.

Commit hygiene (the order's BOUNDARY): the four subjects match `^\[[a-z0-9-]+\] [a-z0-9/._-]+: .+$`; the longest, b112aff's, is 72 characters by hand count, at the limit the engine's commit hook enforces; no body carries an AI-attribution line.

### Unplug

The order asks whether the tests would fail if the closing rule were removed. The executed mutations are UNRUN (`go` denied, refusal under Unrun); each argument below is read from the assertions and names the control it mutates:

- Remove the overlap refusal (drop the branch at graph.go:133, or make `overlapping` always return false): the three round-4 tests fail at their `errors.Is(err, world.ErrEdgeOverlap)` assertions — the writes would succeed and `err` would be nil. The mutated field is the refusal itself.
- Remove the closing rule (drop graph.go:204): `TestSupersedingEdgeClosesTheOldOne` fails — the old edge would keep no `ValidTo` and no `SupersededBy`, and the "now" query would return two neighbours. The mutated fields are the old edge's `ValidTo` and `SupersededBy`.
- Remove the renewal rule (drop graph.go:197-198): `TestRepeatedTargetRenewsOneEdge` fails — the second write of the same target would store a second edge and the total would pass 2.
- Remove instant canonicalisation (compare bounds as text): `TestEquivalentOffsetsAreTheSameInstant` fails — an edge written 01:00+01:00 would no longer count at 00:30Z, and two spellings of one instant would name two edges.

The alarm each guard raises is a named test failure. Which verb refuses when the mechanism fails: `go test ./internal/world` — UNRUN on this seat, so the alarms are argued, not shown.

### Closure

- Verdict: accept-with-conditions. Zero blocking findings; four NOTE findings under Remainder, none of the class that blocks — judgement findings are owed work, never a reason for another round.
- Why not `accept`: this review's Unrun heading is not none. The Go toolchain is denied on this seat, so the suite, the base run and the mutation runs are unproven by this leg. The review-return-refusal law forbids an accept over a non-empty Unrun, so the executed proof must come from the master's recorded run or a fresh run on a seat that holds the toolchain; the conditions below carry exactly that, and only the master may close them.
- Ruling declared (C6): the order's artifact-confinement sentence names `██████████████████████████████████████████████` while its review-output section names the two root channel files as the return manifest. The specific section wins: the two channel files are the return, and this leg made no other artifacts, so no directory was created and the worktree stays at the protocol family. Cost if wrong: none — the manifest names exactly the files the delivery reads.
- Grammar check: `saphan doc check _REVIEW_STATUS.md --family review-status` was attempted once, as one program, and refused by the permission layer; the refusal is quoted under Unrun. The status file follows the engine-emitted skeleton's forms exactly: the H1, the six keys, the verdict token alone on its line, `ref` naming a declared artifact verbatim, nested dashes for the manifest.
- Worktree state: nothing committed, nothing staged. This leg's writes are the two channel files only. The final `git rev-parse --short=7 HEAD` and `git status --porcelain`, read after these notes were written, printed verbatim:
  b112aff
  ?? ███████████
  ?? .saphan-kickoff-brief
  ?? _REVIEW_NOTES.md
  ?? _REVIEW_ORDER.md
  ?? _REVIEW_STATUS.md
  The four first entries were untracked when the leg started; `_REVIEW_NOTES.md` is this return's own artifact, named in the status manifest. The worktree is otherwise as it was found.

### Remainder

Owed, non-blocking. FILE·LINE·CONTROL for each:

1. `internal/world/graph.go:288` — control `else if r.ValidTo != ""`. `Neighbors` with an empty `at` ("now") skips every edge that has a `ValidTo`, including one whose interval contains the physical present: a bounded edge written with a past start and a future end is invisible to "now". The semantics is documented in the method's comment ("only open edges count") and the `at` query is exact, so no history is lost; the trap is a caller reading "now" as the present instant. Remedy: resolve an empty `at` to the present UTC instant, or state the open-only semantics in the public docs.
2. `internal/world/graph.go:116` — control `if prev, ok := s.edges[e.ID]`. A repeat write with the same from, type, to, context and start but a different `valid_to` returns the stored edge and ignores the offered end, with no error. The caller sees the stored truth, so the fact is not hidden, but nothing distinguishes a repeat from a conflicting end. Remedy: refuse when the stored `ValidTo` differs from the offered one, or document the identity rule beside the API.
3. `internal/world/world.go:216` — control `for _, r := range file.Edges`. `Open` validates an edge's provenance and id but not that its times parse; an edge with a bad timestamp would load and then be silently skipped by `overlapping`, `place` and `Neighbors`. Reachable only through an externally damaged state.json — `persist` writes canonical times — so this is hardening, not a live defect. Remedy: parse-check `ValidFrom` and `ValidTo` in `Open`, as the store already refuses its other damaged shapes.
4. `hackathon/fleet/infra/README.md:32` — the "Build: Meterless engines" row still prints "World Model 7 of 8" with its 22 October run date. It is a dated record and the three docs the order names are correct, but the sweep missed this row, and a judge reading the infra page sees the stale figure. Remedy: refresh the row with the new run date and 8 of 8, or keep the dated record and add the current score beside it.

Declined to judge (R11): Meterless's reference suite and runner code (the order says judge Proven's code, not the reference); the H-MEM scorecard (outside this diff's scope); the store's unsynchronised methods (a pre-existing shape — the fold is sequential by design, and this diff did not introduce it).

### Profile

No profile is named in the order. The diff touches a local data store: no network, no auth, no secrets, no destructive path, so no security profile applies. Backend observations: writes are idempotent by content address and rollback is a snapshot swap; every mutation is observable in `world.jsonl` and `state.json` without exposing secrets; `Neighbors` and `overlapping` scan the whole map per call, which is linear and proportionate at one review's scale.

Testing bar: the project declares no numeric metric — no coverage, mutation, lint or type-warning bar. Per the testing-bar skill, that is itself the finding: no declared bar; the first baseline should be measured and recorded before a bar is set. Nothing in the diff weakens a bar: no skips, no loosened assertions, no suppression comments, no thresholds moved; the scorecard is committed as the runner prints it.

Check ranking by circularity: external — Meterless's published conformance suite at a pinned upstream commit, the strongest class; project — the engine's commit-message rules, enforced on every commit of the branch; suite — the Go tests in `internal/world`. The primary acceptance instrument is the external one, not a suite the author could shape.

### Focus

The round's focus, as the order directs: the overlap rule.

- Is any earlier finding still reachable, on any public path? No. The only write path is `Relate`; every write runs the `overlapping` check after `place`, open or bounded, and `overlapping` compares against every earlier edge of the same from, type and context, not only chain members. The public surface — `relate` in `cmd/meterless-conform` — forwards straight into it.
- Is refusing an overlap a sound rule for the specification? Yes. The specification's check is "a superseding edge closes the old one; point-in-time queries respect validity", which is the invariant that a chain holds at most one valid edge at an instant. An edge the caller bounded is a recorded fact the store refuses to rewrite (`sameChain`, graph.go:59), so an overlapping write has no faithful automatic resolution: closing or splitting the fact would rewrite it, and storing the write would answer a point-in-time query with two truths. Refusal with a named error hands the decision to the caller. Touching intervals stay writable, and the tests pin that.
- Are the three round-4 tests adequate? They cover bounded-then-open (same target), bounded-then-bounded (same target inside, same target straddling the start, other target inside) and open-over-bounded-of-another-target, each asserting the refusal by name, the unchanged state and the neighbours at the contested instant. A fourth shape — a bounded write straddling an open edge's start — is refused by the same check, by reading, since the interval test is symmetric in the two edges; it has no dedicated test, which is owed work of the smallest class, not a hole in the rule.

### Conditions

1. The gate names the executed suite proof before the merge: either the master's recorded run — Meterless's runner at b112aff, pinned 1742dec5, world-model 8/8, whole suite 54 ok, 0 FAIL, measured 2026-10-26 — is accepted in the gate note as the run this review could not repeat, or the suite is re-run on a seat whose grant admits the toolchain. Only the master may close this condition, and a condition's class changes only upward.
2. The four NOTE findings under Remainder land on the owed list with their remedies; none blocks the merge.

### Unrun

- `go test ./... -count=1` at the tip b112aff — UNRUN. `go` is denied on this seat. Single-program attempt: `go version`; refusal, verbatim: "Qwen Code requires permission to use \"run_shell_command\", but that permission was declined (non-interactive mode cannot prompt for confirmation)."
- The same suite at the base 13349e5 — UNRUN, same refusal. That the base holds no world tests is a reading of the tree, not a run.
- `go vet ./...` and `gofmt -l` on the changed files — UNRUN, same refusal.
- `test/conformance/meterless/run.sh` — UNRUN: it builds with `go` (denied) and needs `node` plus a `METERLESS_ENGINES` checkout. The committed scorecard and the master's 2026-10-26 measurement are the executed proof on record.
- The mutation runs under Unplug — UNRUN, same refusal; the alarms are argued from the assertions.
- `saphan doc check _REVIEW_STATUS.md --family review-status` — UNRUN. Attempted once, as one program; refusal, verbatim: "Qwen Code requires permission to use \"run_shell_command\", but that permission was declined (non-interactive mode cannot prompt for confirmation)." The status file follows the engine-emitted skeleton's forms exactly; the check could not be re-run here.

## The review status as filed

## REVIEW — world-model-bitemporal
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Round 4 holds: all four earlier findings are repaired and unreachable on the public write path, refusing an overlap is sound for the specification, and the three named docs state 8 of 8 as the committed scorecard prints. Zero blocking findings; four NOTE findings owed (now-query skips a bounded edge valid at the present; a repeat write's differing valid_to is ignored; Open does not parse-check edge times; the infra README row still prints 7 of 8). accept-with-conditions: go is denied on this seat, so every go run is Unrun and no accept is legal; the gate names the master's measured 8 of 8 run at b112aff as the executed proof, or the suite is re-run before merge.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-26T00:52:13Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| private workspace path | 1 |
