# REVIEW — corpus-pv-003 · round 2

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T16:10:43Z

## Summary

PV-003's check fails before printing case results after the corpus path move; stale Go imports also prevent compiling it. Restore the executable paths and independently rerun before accepting the recorded outcome.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-003

### Ground

**RED — BLOCKING: the case cannot run on the reviewed tree.** HEAD measured by `git rev-parse HEAD`: `cbab00f37e57f9c8c665e4c4f2ac1c2c0f0b71f9`; scope `9303b2d...cbab00f` has two case commits. On Linux with `go version go1.26.5 linux/amd64`, `bash test/corpus/cases/PV-003/check.sh` exits 1 without printing a case. Its complete output, with the absolute checkout path redacted:

```
stat ███████████████████████████████████████████████████████████/tests/corpus/cases/PV-003/check: directory not found
```

`test/corpus/cases/PV-003/check.sh:5` invokes the former `./tests/corpus/...` path. Independently, `test/corpus/cases/PV-003/check/main.go:9-10` imports the former `tests/corpus` module packages. `go test ./test/corpus/cases/PV-003/...` exits 1 because neither imported package exists; base and change separately show `[no test files]`. Fix the script, imports and instructions in `CASE.md:94-95,116-117` to the current `test/corpus` layout; rerun the unmodified canonical check and capture its exit and all lines. This path failure is not evidence of a cooldown outcome.

### Differential

Source inspection suggests the intended comparison but does not replace execution. `check/main.go:25-57` constructs five independent inputs and runs each on base, change and reference (`:70-76`). In `:40-51`, X exhausts on slot 1, moves to slot 2, and Y takes slot 1. `base/gate.go:30-49` stores cooldown by node; `change/gate.go:29-52` uses node and slot and ignores the recorded identity for lookup; `check/reference.go:23-45` uses the tenant. Thus the source predicts the move and vacated-slot disagreement recorded in `CASE.md:103-104`, with the fixed-placement and no-cooldown inputs agreeing. These are deductions, NOT fresh output. `expected/EXPECTED.md:3-10` calls this an established defect; the identity-based rule in `CASE.md:7-10,39-44` would support it if a check ran. The reference encodes that rule, not an independent measurement.

### Unplug

False-world answer: a runner aimed at a missing path emits zero cases and exits nonzero, exactly the observed result. Both a valid outcome and its opposite remain consistent with this run because no case result was emitted. The historical five lines in `CASE.md:100-105,119-125` are not independent evidence on this checkout. No guard or output assertion was visibly bent, but a key-mutation negative control cannot run until the runner compiles. After repair, changing the moved tenant's cooldown behavior must make the move result differ from the expected line.

### Closure

**Verdict: reject** until the canonical check works and an independent five-line run confirms §2. `git diff --name-status 9303b2d...cbab00f` lists only the PV-003 case directory outside `expected/` and `test/corpus/manifest.json`; no answer-key edit. The PV-003 manifest status remains an object at `manifest.json:59-64`, but `checks: run 2026-10-22` reports a historical run, not the current tree's status. `git diff --check` is clean. No dependency or install script was added.

### Remainder

Round-two macOS condition NOT ADDRESSED: `CASE.md:127-129` explicitly says native macOS was not run. The status-object restoration IS ADDRESSED (`manifest.json:59-64`). The round-two recorded recheck output is NOT ADDRESSED for this reviewed layout: the executable check fails. Bash `exec`/`set -euo pipefail` and Go use no GNU-specific flag; the absent `tests/corpus` paths fail on either platform. No private host, seat, token, email or internal path was found in the scoped source diff; the channel/order itself contains private material and is not public copy. Licence matters are declined per order.

### Profile

Go fixture/test-corpus review. The only declared bar is a suite check (most circular); no external or project metric baseline and no exception are declared. The canonical check and `go test` fail on Linux; mutation score and numeric base/tip metric delta are UNRUN. `go.mod` pins Go 1.26; local Go is 1.26.5. No performance claim or added dependency.

### Focus

Contract §2 inputs, expected outcome, executable paths, scope, privacy, Linux/macOS portability, and round-two conditions. Source reading cannot establish the outcome of a failed check; earlier runs on the `tests/corpus` layout do not prove this `test/corpus` layout.

### Conditions

BLOCKING — `test/corpus/cases/PV-003/check.sh:5`, `check/main.go:9-10`: correct the stale paths, run the original check, verify five case lines and exit 0 against §2 and the answer key. OWED — `CASE.md:94-95,116-129`: update runnable commands and obtain native macOS verification or retain the unrun condition for the owner. Reviewer made no product changes.

### Unrun

Native macOS (no runner); key-mutation negative control (canonical check does not run); numeric base/tip Go metric comparison (base predates the fixture). No successful reproduction run is claimed.

### REDACTED

- Absolute checkout path from the `stat` error: 1 occurrence, 59 characters redacted with 59 blocks. The relative suffix is preserved.

## The review status as filed

## REVIEW — corpus-pv-003
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: PV-003's check fails before printing case results after the corpus path move; stale Go imports also prevent compiling it. Restore the executable paths and independently rerun before accepting the recorded outcome.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T16:04:40Z

## REDACTED

none
