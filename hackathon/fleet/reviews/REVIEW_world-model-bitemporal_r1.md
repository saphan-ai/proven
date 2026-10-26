# REVIEW — world-model-bitemporal · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 3 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T15:39:41Z

## Summary

Three temporal graph cases return wrong results: same-target renewal, reverse-time insertion, and timestamp offsets. Go tests pass; external Meterless runner was unavailable on this seat. Reject pending repair and rerun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — world-model-bitemporal

### Ground

RED — Reject the bitemporal claim. Three valid write/query sequences return the wrong history or current edge. The committed 8/8 scorecard covers a narrower sequence. Review tip `d7ffd71`, diff `13349e5...d7ffd71` (two commits, nine files); `go version` reported 1.26.5 against `go.mod` 1.26.0. The executor's 8/8 and 54/0 claims are claims, not independent measurements here.

### Differential

1. **BLOCKING — repeat of the same target creates overlapping history.** `internal/world/graph.go:52-69`: the content address includes `ValidFrom`, but the closure loop excludes an open edge with the same `To`. Control: `internal/world/graph_test.go:52-91` changes A to B only and never repeats A at a later time. Reproduction: A→B at 2087, A→B again at 2088, A→C at 2089. `Neighbors(A, member, 2088-06-01)` returned `[B,B]`, though one relationship is valid. The 2087 edge remains valid through 2089 and overlaps the 2088 edge. Remedy: define whether repeating the target renews or preserves one edge, then close or reject the overlapping interval and test the temporal result.
2. **BLOCKING — out-of-order writes close an edge before it starts.** `internal/world/graph.go:57-69`: any open edge is closed at the new edge's start without checking order. Control: `internal/world/graph_test.go:73-91` inserts only increasing times. Reproduction: A→B at 2089, then A→C at 2088. The store accepts both; `Neighbors(A, late, 2089-06-01)` returned B, but the open/current neighbor is C. B has `ValidTo=2088` before its `ValidFrom=2089`; a historical query treats it as absent and current query treats C as current. Remedy: reject retroactive supersession or re-split intervals in temporal order; test reverse insertion and valid interval bounds.
3. **BLOCKING — timestamp strings are compared as bytes, not instants.** `internal/world/graph.go:102-104` compares caller strings without parsing or normalizing; `Relate` accepts timestamp offsets at lines 41-52. Control: `internal/world/graph_test.go:80-84` uses only one fixed UTC format. Reproduction: edge starts `2087-01-01T01:00:00+01:00` (00:00Z); query at `2087-01-01T00:30:00Z` returned zero neighbors, though the edge had started. Remedy: parse and compare instants, or enforce one canonical timestamp format at writes and queries; cover equivalent offsets.

All three reproductions ran with `go run ██████████████████████████████████████████████████████` against the unmodified reviewed code. The three temporary reproduction and overlay files remain under that confined directory: the cleanup command was denied (`rm -f style commands are not permitted. Use a safer approach`), so I did not restate the denied deletion through another command. No tracked source file changed.

### Unplug

The closing rule was replaced only in a temporary Go overlay (`if e.ValidTo == ""` → `if false`). `go test -count=1 -overlay ██████████████████████████████████████████████████████████ -run 'TestSupersedingEdgeClosesTheOldOne|TestPointInTimeQueryRespectsValidity' -v ./internal/world` failed both named tests: current returned `[team:g:a team:g:b]`, and the closing instant also returned both. The existing tests do detect removal of the simple A→B closure, but not the three sequences above.

### Closure

`go test -count=1 ./internal/world ./cmd/meterless-conform ./test/conformance/meterless` passed all three packages. `go test -count=1 ./...` also exited 0, with every tested package passing and the listed no-test packages having no tests. The ingest implementation and its idempotency tests were unchanged by the graph diff; the new edge map is included in fold snapshot and persistence in `internal/world/world.go:213-222,492-501,513-539`. Static reading found no generator call, provider, or network endpoint in `internal/world`. The two commit subjects satisfy the supplied message form. README, ARCHITECTURE, DESIGN, and the checked-in World Model scorecard all say 8/8 consistently; the card prints eight PASS rows and `8/8 checks passed`. This is document consistency, not a rerun of Meterless's runner.

### Remainder

No dependency manifest or lockfile changed. No declared coverage, mutation, lint-warning, or type-warning floor was found; no numeric bar can be judged. The checked-in scorecard is a suite check owned by this project and is less independent than the external runner. Declined to judge Meterless's upstream reference implementation, H-MEM's unbuilt sections, and product behavior outside this diff. No exception or owner ruling is requested for the three wrong results.

### Profile

Backend data-access and docs profiles apply: this diff adds durable graph state and updates public score claims. The graph is held per store and persisted; the review found no new network I/O. Source files and direct callers (`meterless-conform`, TypeScript shim) were read. There is no new dependency or client/server pair beyond that adapter.

### Focus

Properties 1 and 3 fail for the three differential cases; the simple closing-rule negative control passes. Property 2's existing ingest tests passed locally; an edge/ingest interleaving was not independently exercised. Property 4 passed static inspection of engine code. Property 5 passed a comparison of the changed prose with the committed scorecard, subject to the unrun external runner.

### Conditions

Reject. Repair the three temporal failures and add tests whose expected neighbor sets cover same-target renewal, reverse time insertion, and equivalent timestamp offsets. Rerun the external conformance runner before claiming its 8/8 score as independently verified.

### Unrun

Meterless's upstream runner was not run: `METERLESS_ENGINES` did not identify a runner on this seat. The base commit's Go test counts were not measured because this review was confined to the checked-out review tip and may not create another checkout. Race and mutation-score tools were not run; this diff introduces no goroutines or declared mutation metric.

## The review status as filed

## REVIEW — world-model-bitemporal
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Three temporal graph cases return wrong results: same-target renewal, reverse-time insertion, and timestamp offsets. Go tests pass; external Meterless runner was unavailable on this seat. Reject pending repair and rerun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T15:34:53Z

<!-- ENGINE-EMITTED REVIEW SKELETON, rendered from ██████████████████████████████.
     Everything below this line contains WORKED FORMS to copy as applicable.
     inside an HTML comment, and this engine's parsers skip comment interiors,
     so nothing here is read as a declaration — copy the applicable form OUT.

     Write-legal `- state:` tokens: spawned · recommendation-ready.
     Last act before done: set `- state:` above to recommendation-ready and add every
     return-time key shown below. Their names, value classes, and vocabularies
     are projected from the review family's rows, never copied here.

     `verdict`: exactly one write-legal verdict token, alone on the line — mandatory for `recommendation-ready`; no parenthetical and no conditions after it; the conditions belong in the review body the `- artifacts:` manifest names
     - verdict: <accept | accept-with-conditions | reject>

     `ref`: the review's returned artifact — mandatory for `recommendation-ready`; one of this document's own `- artifacts:` entries, VERBATIM (a basename that is not a declared entry is refused at delivery); never a git ref
     - ref: _REVIEW_NOTES.md

     `artifacts`: NESTED DASH LIST of paths the review returns; mandatory for `recommendation-ready`, including an explicit empty list when nothing is returned
     - artifacts:
       - _REVIEW_NOTES.md
       - _REVIEW_STATUS.md

     An entry names a FILE, or a DIRECTORY: a directory is delivered as ONE
     deterministic `.tar` archive (sorted entries, fixed mtimes — the same
     bytes, and so the same digest, every time) recorded under the entry's
     own name plus that suffix, which is the name the projected manifest and
     the teardown gate then read. A path that EXISTS and is neither is
     refused by what it IS, never as `missing`.

     Explicit empty manifest form when the review returns no files:
     - artifacts:

     🔴 THE ANALYSIS DOES NOT GO IN THIS FILE. It goes in _REVIEW_NOTES.md, whose
     mandatory members are H2 HEADINGS — never top-level dash bullets in
     _REVIEW_STATUS.md. A section heading written as `- ground:` here is an unknown
     key: it is now reported as a WARNING rather than refused, so your
     verdict still lands — but the analysis lands nowhere at all.

     The complete _REVIEW_NOTES.md form, projected from the same one section source
     this engine renders it from (██████████████████████████████):

     # REVIEW NOTES — world-model-bitemporal

     ## Ground
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Differential
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Unplug
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Closure
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Remainder
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Profile
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Focus
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Conditions
     <replace with measured review content; write `none` only when none is the measured answer>

     ## Unrun
     <replace with measured review content; write `none` only when none is the measured answer>
-->

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| file and line in Saphan Studio's private code | 2 |
