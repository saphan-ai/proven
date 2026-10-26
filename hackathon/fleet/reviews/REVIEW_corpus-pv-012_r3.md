# REVIEW — corpus-pv-012 · round 3

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T22:04:28Z

## Summary

PV-012 runs from its documented path and reproduces section 8; the round-2 path condition is addressed. No blocking or owed conditions.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-012

### Ground

Review target `479cfb8`; `git rev-parse --short=7 HEAD` printed `479cfb8`. Scope: `9303b2d...479cfb8`, with the re-review focused on `99d1b6c..479cfb8`. The locally installed toolchain printed `go version go1.26.5 linux/amd64`; `go.mod` declares `go 1.26`. I read the contract before the implementations and `expected/EXPECTED.md` after running the check. The check from the repository root, `bash test/corpus/cases/PV-012/check.sh`, exited 0 and printed:

```text
control | base: 5 (Activity for all teams; control=not-present) | change: 0 (Activity scoped to ours; control=PASS) | reference: 0
vacuity | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity scoped to theirs; control=PASS) | reference: 5
application | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity for all teams; control=not-applicable) | reference: 0
scope-no-team | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity for all teams; control=not-applicable) | reference: 0
scope-with-team | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity for all teams (request team: ours); control=not-applicable) | reference: 0
extra-sync | base: 5 (Activity for all teams; control=not-present) | change: 4 (Activity scoped to ours; control=not-applicable) | reference: 0
control-no-registry | base: 5 (Activity for all teams; control=not-present) | change: 4 (Activity scoped to ours; control=FAIL) | reference: 0
control-wrong-ownership | base: 5 (Activity for all teams; control=not-present) | change: 4 (Activity scoped to ours; control=FAIL) | reference: 0
```

These eight lines match `test/corpus/cases/PV-012/CASE.md:113-120` exactly. The script exited 0; its documented `exit=0` is a transcript annotation, not output printed by the script. This confirms the master's reported output rather than relying on it.

### Differential

`git diff 99d1b6c..479cfb8` changes only `test/corpus/cases/PV-012/CASE.md:104` and `test/corpus/cases/PV-012/check.sh:9-10`, replacing `tests/corpus` with `test/corpus`. `git diff --name-only 9303b2d...479cfb8` lists only the four case files and `test/corpus/manifest.json`. `git diff --check 9303b2d...479cfb8` passed. No dependency, threshold, test assertion, or unrelated source changes in the path fix.

### Unplug

The executed `control-no-registry` arm disables the registry filter and returns four foreign events and `control=FAIL` instead of zero and `PASS`. `control-wrong-ownership` changes the sync order and returns four and `FAIL`. The false-world answer for a genuinely safe production construction would be `application | change: 0 | reference: 0`, not the observed 5 versus 0. A bare exit code 0 cannot distinguish those worlds: the script prints counts but does not assert the reference. The explicit count comparison establishes this synthetic case's defect, not an actual deployed service's state.

### Closure

The retired-path condition is ADDRESSED: the documented invocation and both Go build inputs point to existing files and execute. The reproduced `application` (5 versus 0), `control` (0 versus 0), and `extra-sync` (4 versus 0) lines establish the expected outcome in `expected/EXPECTED.md`; the reference is seeded independently of page output. The negative control confirms the control detects an absent filter; it does not establish that production enters the filter. The scope sentences differ when the request names a team, so whole-sentence equality is not claimed.

### Remainder

No hits for `tests/`, `fleet/`, or `conformance/` in the case. The change is limited to the public test corpus and carries no newly introduced private path or secret. Declined to judge: the historical production incident, licensing, and performance of any deployed application; this in-memory fixture cannot measure them. No declared coverage, mutation, lint, or type metric and baseline in this order; no invented threshold. Check ranking: project Go build, then suite-local printed fixture checks; no external check, so the suite-only evidentiary limit is a NOTE.

### Profile

Docs and Go fixture. The documented command ran and its output matches. The paired base and change programs were both built and exercised for all eight inputs by the same script. No new external I/O, authentication, or dependency delta. No separate full repository suite was requested for this path-only re-review.

### Focus

`test/corpus/cases/PV-012/check.sh:9-10` — previously missing build inputs now exist. `test/corpus/cases/PV-012/CASE.md:104` — the documented command now runs. `test/corpus/cases/PV-012/CASE.md:115,118` — the production-construction and extra-sync counts match the expected outcome.

### Conditions

No BLOCKING or OWED condition. Round-2 path condition ADDRESSED.

### Unrun

No full repository suite or live application: outside the scoped path-only re-review. No independent external reference audit. No fresh run at the pre-fix commit; its build path is absent in that commit, and the reviewed fix was evaluated at `479cfb8`.

## The review status as filed

## REVIEW — corpus-pv-012
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: PV-012 runs from its documented path and reproduces section 8; the round-2 path condition is addressed. No blocking or owed conditions.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T22:01:10Z

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

     # REVIEW NOTES — corpus-pv-012

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
| file and line in Saphan Studio's private code | 2 |
