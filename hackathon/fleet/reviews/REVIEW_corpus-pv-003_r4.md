# REVIEW — corpus-pv-003 · round 4

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T20:34:41Z

## Summary

PV-003 reproduces CASE.md §8 at ceba195. The path repair is complete. Independent macOS execution remains unrun on this Linux seat; the master's macOS result is a claim, not this review's measurement.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-003

### Ground

`git rev-parse --short=7 HEAD` printed `ceba195`. Reviewed `9303b2d...ceba195`; repair diff `cbab00f..ceba195`. The repair changes only three PV-003 files, replacing `tests/corpus` with `test/corpus` (CASE.md:94-95,118; check.sh:3,5; check/main.go:9-10). The full diff names only PV-003 files and the manifest. CASE.md §2 is the contract. I read `expected/` last.

### Differential

`bash test/corpus/cases/PV-003/check.sh` exited 0 on `go1.26.5 linux/amd64` and printed exactly:

```
own-case-cooled-slot | base=allowed | change=refused(A) | reference=refused(A)
own-case-other-slot | base=allowed | change=allowed | reference=allowed
move-moved-tenant | base=refused(X) | change=allowed | reference=refused(X)
vacated-slot | base=refused(X) | change=refused(X) | reference=allowed
control-no-cooldown | base=allowed | change=allowed | reference=allowed
```

This matches CASE.md §8 byte for byte within its output block. `go vet ./test/corpus/cases/PV-003/...` and `gofmt -l test/corpus/cases/PV-003` exited 0 with no output. `rg -n 'tests/|fleet/|conformance/' test/corpus/cases/PV-003` found no matches. The repair diff changes no dependency, threshold, skip, suppression, or baseline. No numeric quality bar is declared for this case.

### Unplug

False-world answer: these five lines alone would also fit a service that deliberately limits slots rather than tenants. CASE.md §1-2 states that cooldown belongs to the tenant and that moving a tenant is legal. Against that rule, `move-moved-tenant` and `vacated-slot` distinguish the slot key from the identity key. Removing the move would leave only fixed placements, where change and reference agree. The reference implements the stated rule; it is not independent evidence that the rule is correct.

### Closure

The round-2 path condition is ADDRESSED at CASE.md:94-95,118; check.sh:3,5; check/main.go:9-10. The documented command runs from the root. The code uses standard Go and Bash with no platform-specific command. The master's reported macOS output matches this Linux run, but is second-hand evidence here.

### Remainder

OWED: attach the master's macOS command, exit code, and output to the gate record, or state that independent macOS confirmation remains open. This Linux seat cannot certify that host. This does not make the case wrong or unrunnable on the measured host. No other finding.

### Profile

Docs and Go case. The check is a suite check under the case author's control, so it is more circular than the stated tenant rule and preserved reviewer record. No external quality metric was declared. No security, dependency, or performance claim changed in this repair.

### Focus

Expected outcome: established defect. The change allows X after a legal move while the reference refuses X. It refuses the vacated slot in X's name while the reference allows Y. Fixed-placement and no-cooldown controls agree with the reference. This proves the defect under the case's stated tenant rule, not every possible assignment. The historical accepting reviewer measured the move too; the reviewers differed on its weight.

### Conditions

Accept with conditions. OWED: keep the independent macOS verification limit in the gate record, alongside the master's measured macOS evidence if available. No BLOCKING condition.

### Unrun

Native macOS execution: this seat is Linux and has no macOS runner. Full repository test suite: outside this repair-only review. Separate mutation run: not performed; the move and vacated-slot inputs discriminate the keying behavior.

## The review status as filed

## REVIEW — corpus-pv-003
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-003 reproduces CASE.md §8 at ceba195. The path repair is complete. Independent macOS execution remains unrun on this Linux seat; the master's macOS result is a claim, not this review's measurement.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md

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

     # REVIEW NOTES — corpus-pv-003

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
