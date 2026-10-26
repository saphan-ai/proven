# REVIEW — corpus-pv-008 · round 3

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T15:18:58Z

## Summary

Four observations support the expected defect. Correct the stale documented root command before delivery. Linux checks pass; macOS was not run.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-008

### Ground

Review of `ae34d44...8adaca2`; `git rev-parse --short=7 HEAD` printed `8adaca2`. The change adds `base/`, `change/`, `reference/`, `check/`, `check.sh`, and `go.mod`, edits `CASE.md`, and changes only PV-008's manifest status fields. `git diff --name-only ae34d44...HEAD -- test/corpus/cases/PV-008/expected` returned nothing. No dependencies beyond the standard library (`go.mod` declares Go 1.26). `go version` printed `go1.26.5 linux/amd64`.

I ran `bash test/corpus/cases/PV-008/check.sh`; it exited 0 and printed:

```text
normal-close-then-late-subscriber | no-frame | frame:ended | frame:ended
wedged-stop-then-late-subscriber | no-frame | no-frame | frame:ended+alarm:worker-not-released
wedged-stop-then-read-state | state:closed | state:closed | state:closed
wedged-stop-then-second-close | returned | blocked | returned
```

`go -C test/corpus/cases/PV-008 test -count=1 ./check` and `go -C test/corpus/cases/PV-008 test -race -count=1 ./check` both passed; `go -C test/corpus/cases/PV-008 vet ./...` passed. `git diff --check ae34d44...HEAD` passed. This is an independent Linux measurement, not an adoption of the author's exit-code claim.

### Differential

`test/corpus/cases/PV-008/check/main.go:125`: on normal close, base has no late replay, while change and reference replay `ended`. With the stop held, change's `Close` sets closed before `Stop` (`change/session.go:57-60`), never reaches final-frame publication (`change/session.go:62-71`), and its second `Close` waits on `done` (`change/session.go:52-55`). Base's second close returns (`base/session.go:40-42`); reference times out its stop, publishes an alarm frame, then completes (`reference/session.go:58-78`). Those four observations match §2 and the three wedged lines plus normal line required by `expected/EXPECTED.md:3-9`. The change has no path to an alarm before the stop returns. The evidence establishes the liveness defect, not that a particular severity ruling is uniquely correct.

The printed `no-frame`/`blocked` alone would ALSO appear if the stop returned only after the one-second observation window: that is the false-world answer. `change/session.go:60` supplies the distinct source evidence: it has no bound, and the worker remains held during observation (`check/main.go:32-36,133-162`). This is a bounded observation of an unbounded code path, not proof from elapsed time alone. Normal close is not a defect in change.

### Unplug

The round-two tests exercise a held lock: `go -C test/corpus/cases/PV-008 test -count=1 ./check` passes `TestHeldLockObservationsAreBounded`, which expects `subscribe:blocked` and bounded state and second-close observations (`check/main_test.go:38-54`). `TestStopNotEnteredIsNamed` checks a close that never reaches stop and expects `stop-not-entered` (`check/main_test.go:56-69`). Both are negative controls for the newly bounded observations. These tests rank as suite checks, weaker than the source-and-output comparison with the independent contract; the command's zero exit alone does not assert the four expected output strings (`check/main.go:166-174`). I did not edit the review target to run additional mutations.

### Closure

The current tree addresses the two visible round-two failure classes: observation calls are bounded (`check/main.go:76-104`), and missing stop entry and unfinished normal close have named outcomes (`check/main.go:106-118`). The manifest preserves a status object (`test/corpus/manifest.json:148-154`). The order says round-one conditions were carried into an earlier order, but does not supply their verbatim text here; I cannot certify that no other original condition remains. The defect's expected class is established, but this review has one delivery condition below.

### Remainder

**OWED — documented command points at a nonexistent directory.** `test/corpus/cases/PV-008/CASE.md:92-95` still directs a judge to `tests/corpus/cases/PV-008/`, as does `test/corpus/cases/PV-008/check.sh:3`. From this root, `bash tests/corpus/cases/PV-008/check.sh` failed with `No such file or directory` (exit 127), whereas `bash test/corpus/cases/PV-008/check.sh` succeeds. Update the run commands and module path in §8 and the script comment to the moved `test/corpus/` layout, then execute the documented command. This is a user-facing reproducibility defect, not a failure of the actual fixture.

No private machine names, internal paths, credentials, tokens, or real addresses appeared in the reviewed case diff. Its provenance uses generic references. The modified manifest remains limited to PV-008. No extra dependencies or bar-weakening skips were found. A late base subscriber registered *during* a wedged stop would receive a frame if the stop were released later (`base/session.go:29-32,49-54`); §2's base wording concerns subscribers after the end, and the check never releases until after observation, so this does not change the measured four lines.

### Profile

Go concurrency and documentation. `check.sh` uses Bash built-ins and `dirname`; `go run ./check` and `go -C ...` require Go 1.26 and work on both Linux and macOS in principle. The script has no GNU-only flag or platform-specific syscall. Only Linux was tested here. The shell path mistake is platform-independent.

### Focus

The property under review is whether the check reproduces §2, distinguishes normal from wedged stop, and supports the expected factual finding without inferring severity. It does, subject to the documented invocation correction. A check with the same output could fit a slow bounded stop, so source inspection is essential to the word “unbounded.”

### Conditions

Correct all `tests/corpus/cases/PV-008/` run and module path references in `CASE.md:92-100` and `check.sh:3`, then run the command literally from the repository root. The master should compare the original round-one conditions against this diff if their text is available in the private record; this order did not reproduce them.

### Unrun

macOS execution: no macOS host on this seat. Additional source mutations: review target was not edited. Original round-one review documents: not furnished with the order; only the current observable fixes were assessed. No claim of macOS pass, mutation score, or complete historical condition closure is made.

## The review status as filed

## REVIEW — corpus-pv-008
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Four observations support the expected defect. Correct the stale documented root command before delivery. Linux checks pass; macOS was not run.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T15:11:03Z

## REDACTED

none
