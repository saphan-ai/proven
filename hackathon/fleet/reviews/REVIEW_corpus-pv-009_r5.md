# REVIEW — corpus-pv-009 · round 5

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T21:41:50Z

## Summary

PV-009 reproduces all nine recorded lines at 8640a72. The path repair is complete; one stale macOS verification sentence is owed.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-009

### Ground

`git rev-parse --short=7 HEAD` printed `8640a72`; `git rev-parse HEAD` printed `8640a72db2312a75d62fa18a6eba8cdf91289d1a`. The local toolchain printed `go version go1.26.5 linux/amd64`, matching the version recorded in CASE.md §8. `go.mod` requires Go 1.26. I read the case contract, both guards, the check and tests before opening `expected/EXPECTED.md` last. The master's macOS result was treated as a claim, not as my measurement.

`bash test/corpus/cases/PV-009/check.sh` exited 0 and printed:

```text
one-current base=pass change=pass reference=pass
one-stale base=fail:LANGUAGES 41 -> 43 change=fail:LANGUAGES 41 -> 43 reference=fail:LANGUAGES 41 -> 43
stale-after-current base=pass change=fail:LANGUAGES 41 -> 42; 42 -> 43 reference=fail:LANGUAGES 41 -> 42; 42 -> 43
current-after-stale base=fail:LANGUAGES 41 -> 43; 43 -> 42 change=fail:LANGUAGES 41 -> 43; 43 -> 42 reference=fail:LANGUAGES 41 -> 43; 43 -> 42
three-pairs base=pass change=fail:LANGUAGES 40 -> 42; 41 -> 42; 42 -> 43 reference=fail:LANGUAGES 40 -> 42; 41 -> 42; 42 -> 43
no-arrow base=pass change=pass reference=pass
max-int64 base=fail:LANGUAGES 41 -> 9223372036854775807 change=fail:LANGUAGES 41 -> 9223372036854775807 reference=fail:LANGUAGES 41 -> 9223372036854775807
overflow-int64 base=pass change=pass reference=fail:LANGUAGES 41 -> 9223372036854775808
twenty-digits base=pass change=pass reference=fail:LANGUAGES 41 -> 10000000000000000000
```

Those nine lines match CASE.md §8 lines 101–109. `go test -count=1 ./test/corpus/cases/PV-009/...` exited 0: the `change` package passed; the case root and `base` have no test files.

remark: The check prints observations and exits 0 even when change and reference disagree; its exit status alone cannot prove the expected defect.

### Differential

`git diff 627907d..8640a72 -- test/corpus/cases/PV-009` shows only four substitutions of `tests/corpus` with `test/corpus` in CASE.md:96, check.go:9–10 and check.sh:4. `git diff --name-only 9303b2d...8640a72` lists only the six PV-009 case files and `test/corpus/manifest.json`. No dependency or quality-bar configuration changed. `rg -n 'tests/|fleet/|conformance/' test/corpus/cases/PV-009` returned no matches. I found no private path, credential, machine name, or private store identifier in this diff.

### Unplug

The false-world answer for the hypothesis “change rejects every stale closing number” is `change=pass` beside `reference=fail` on a stale note. Both `overflow-int64` and `twenty-digits` give that answer, while `max-int64` gives `change=fail`. In change/guard.go:19–21, a failed `ParseInt` follows the same `continue` as equality with 42. The reference in check.go:22–25 compares digit text with the pinned count and marks those two numbers stale. If conversion failure were refused, these two `change=pass` cells would change to `change=fail`; the recorded output would then fail its exact comparison even though check.sh itself would still exit 0.

remark: This instrument covers the nine listed inputs, not every possible note or a published note carrying an overflowing number.

### Closure

The expected outcome is established for the case: change retains the overflow defect already present in base, while its multi-pair repair is real. CASE.md:103 and :105 show base passing stale later pairs and change refusing them; :107–109 locate the conversion boundary. The same printed lines do not fit the opposite outcome for these inputs. No blocking finding against the case or the path repair remains.

### Remainder

The review is confined to PV-009 and the path repair at 8640a72. I declined to judge licences, other corpus cases, signed and other-script numbers, whether a published note ever overflowed, and the provenance of the historical reviews. No code was changed.

### Profile

Docs and Go. The documented command ran as written. The case's test suite passed. No project quality metric with a command and recorded baseline is declared for this case, so no numeric bar was inferred. The independent reference and recorded output are stronger evidence here than the suite alone; the suite has no overflow test.

### Focus

The round-2 path condition is ADDRESSED: imports, script and documented command now point to existing `test/corpus` paths. The case's output and expected outcome agree. The path change introduced no new behaviour in the guard.

### Conditions

OWED — CASE.md:98: The sentence that the literal command has not run on macOS may now be stale in light of the master's dated macOS run claim. The master should reconcile that sentence with its run record before publication. This is wording and does not make the case wrong or unrunnable. No BLOCKING condition.

### Unrun

The literal check on macOS was not run on this Linux seat; the master's macOS output was supplied in the order but was not independently verified here. No full-repository suite was run because the review scope is PV-009 and the path-only repair.

## The review status as filed

## REVIEW — corpus-pv-009
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-009 reproduces all nine recorded lines at 8640a72. The path repair is complete; one stale macOS verification sentence is owed.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T21:37:49Z

## REDACTED

none
