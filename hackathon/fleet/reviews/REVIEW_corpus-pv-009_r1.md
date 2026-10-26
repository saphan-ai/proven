# REVIEW — corpus-pv-009 · round 1

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T17:44:56Z

## Summary

PV-009 reproduces the repaired pair scan and the retained overflow defect. The printed boundary and reference results support the expected outcome. No blocking finding in the case diff.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-009

### Ground

Reviewed `fe146e0...a0987ed`. `git rev-parse --short=7 HEAD` printed `a0987ed`; `git rev-parse HEAD` printed `a0987ed5d08bd04f1719be3146a1da42661a63e1`. The contract is `tests/corpus/cases/PV-009/CASE.md` §2. `go.mod` pins Go 1.26; local `go version` printed `go version go1.26.5 linux/amd64`.

`bash tests/corpus/cases/PV-009/check.sh` exited 0 and printed:

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

### Differential

`base/guard.go:18` selects the first pair; `change/guard.go:18` visits every pair. The `stale-after-current` and `three-pairs` lines distinguish them. Both guards use the same decimal pair expression and `strconv.ParseInt(..., 10, 64)`; both skip conversion errors. The largest signed 64-bit closing value fails the note in both guards. The next value and the twenty-digit value pass both guards but fail the text-based reference. Every §2 input appears in `check.go:44-52`. The change's six unit cases have no conversion failure, as required.

### Unplug

The false-world answer for a repaired overflow guard would be `change=fail` on the two overflow lines, matching `reference=fail`. The observed `change=pass` separates that world from this one. The false-world answer for an unrepaired pair scan would be `change=pass` on `stale-after-current`; the observed `change=fail` separates it. `check.sh` reports values but does not assert them, so exit 0 alone cannot detect a wrong answer. The printed lines are the evidence and must be read. No guard or check was altered during this review.

### Closure

Verdict: accept. `expected/EXPECTED.md` calls the retained conversion-error behavior an established defect. The required boundary and repair lines support that outcome. `change/guard.go:19-21` is the cause: conversion failure and equality share the same `continue`. This is a defect represented by the fixture. No published note is shown to contain an overflowing number.

### Remainder

The diff changes only `tests/corpus/cases/PV-009/` outside `expected/` and PV-009's own manifest status fields. The answer key is unchanged. The new test assertions are not weakened or skipped; no quality threshold or suppression changes. No private host or seat name, private path or repository name, token, or email address was found in the changed lines. No numeric quality bar was declared for this case, so there is no bar value to compare. Check strength ranking: the printed reference comparison is a suite-level check and needs human reading; the Go unit tests are also suite-level.

### Profile

Documentation and Go fixture review. `check.sh:1-4` uses Bash, `dirname`, `cd`, and `go run`; these calls are available on Linux and macOS with Bash and the pinned Go toolchain installed. The script runs from its own location, so its relative path resolves from either host. The check was executed on Linux only. `go test -count=1 ./tests/corpus/cases/PV-009/...` passed: two packages had no test files and the change package passed. `git diff --check fe146e0...a0987ed` exited 0.

### Focus

The case demonstrates the false pass caused by treating an unconvertible closing number as current while preserving evidence that the multi-pair repair works. The reference compares digit text with decimal `42` and does not call either guard's conversion logic. Its regex differs from the guards' pair regex, but the supplied nine inputs all have the intended pair shape or no arrow, so that difference does not explain the mismatch.

### Conditions

none

### Unrun

macOS execution was not run because this seat is Linux. The full repository suite and a base-revision test run were not run because this review is confined to the PV-009 case and its check; the base revision has no case code. No mutation was written to the worktree because this is a read-only review.

## The review status as filed

## REVIEW — corpus-pv-009
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: PV-009 reproduces the repaired pair scan and the retained overflow defect. The printed boundary and reference results support the expected outcome. No blocking finding in the case diff.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T17:40:06Z

## REDACTED

none
