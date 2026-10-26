# REVIEW — corpus-pv-008 · round 5

- verdict: accept
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T21:53:41Z

## Summary

PV-008 reproduces the recorded four lines at d27b0d7; the path-only repair pays the round-2 condition. No blocking or owed findings.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-008

### Ground
HEAD: `d27b0d722163eab2c0ffc8f0b64e5441aa2ebd4a` (`git rev-parse HEAD`); `git rev-parse --short=7 HEAD`: `d27b0d7`. Scope: `9303b2d...d27b0d7`. Reviewed the contract in `test/corpus/cases/PV-008/CASE.md:24`, implementations and check before opening `expected/EXPECTED.md`. The fixture, check and recorded output agree with the contract. The manifest entry at `test/corpus/manifest.json:139` identifies PV-008; no other case changed in the scoped diff. This review does not claim to re-create the private original incident.

### Differential
`git diff 1970cfc..d27b0d7` changes four `tests/corpus` path mentions to `test/corpus` in `CASE.md:92,95,106,107` and one in `check.sh:3`, with no other change. `git diff --name-only 9303b2d...d27b0d7` lists eight files under PV-008 plus `test/corpus/manifest.json`; `git diff --check` succeeds. `rg -n 'tests/|fleet/|conformance/' test/corpus/cases/PV-008` found no matching retired paths. No credentials or private paths appear in the inspected scoped diff. No dependencies, thresholds or test assertions changed in the repair.

### Unplug
False-world answer for the measured claim: a bounded stop completing after the 1 s observation limit could also produce `no-frame` and `blocked`; the check alone cannot prove infinite waiting. `change/session.go:59-81` calls `Stop` synchronously without a bound, before storing the frame and closing `done`; the stand-in stop waits on release at `check/main.go:35-39`. A stop with a 100 ms bound produces the different reference frame and returning second close, so the check discriminates this implemented wedge from the reference. `check/main_test.go` tests held-lock and stop-not-entered failure controls. The path-only fix would fail at launch if the old `tests/corpus` command were retained; the corrected command ran from the repository root.

### Closure
`bash test/corpus/cases/PV-008/check.sh` exited 0 and printed exactly the four lines recorded in `CASE.md:97-103`:

```
normal-close-then-late-subscriber | no-frame | frame:ended | frame:ended
wedged-stop-then-late-subscriber | no-frame | no-frame | frame:ended+alarm:worker-not-released
wedged-stop-then-read-state | state:closed | state:closed | state:closed
wedged-stop-then-second-close | returned | blocked | returned
```

`go -C test/corpus/cases/PV-008 test -count=1 ./check`: `ok pv008/check 4.003s`; `go -C test/corpus/cases/PV-008 test -race -count=1 ./check`: `ok pv008/check 5.009s`. The wedged rows show the established defect in change versus the reference; the normal row shows that change and reference agree normally. The answer key in `expected/EXPECTED.md:3-13`, read last, states the same outcome and does not prescribe a severity verdict for the original pair of reviews.

### Remainder
No blocking or owed findings in this repair. Declined to judge the historical review severity, the undisclosed original code, licences, and source-scan/build-pipeline blind spots outside the case contract. The output does not measure behaviour after the 1 s limit. This is a same-host Linux reproduction, not an independent macOS run. No cost or pool usage was supplied by the seat's measurement record; do not infer one.

### Profile
Docs and Go test fixture. Commands and snippets in the repair are runnable at their new paths. No dependency or configuration delta. Testing-bar metrics have no declared baseline or threshold for this case: no base/tip metric delta is asserted. Check ranking: the Go check and its Go unit/race tests are suite-level, not an external oracle; the contract and separate expected key supply the independent written comparison but not an external execution.

### Focus
FILE·LINE·CONTROL: `test/corpus/cases/PV-008/check.sh:3-6` launches the module from the corrected directory; `CASE.md:92-107` documents the same path and output; `check/main.go:117-158` exercises all four contract inputs; `change/session.go:59-81` shows why the observed wedge prevents final publication, while `reference/session.go:55-88` bounds it. Re-review disposition: round-2 retired-path condition ADDRESSED; no new out-of-scope finding opened.

### Conditions
none

### Unrun
Independent macOS reproduction and tests outside PV-008: not run; not requested in this repair scope. Historical base/tip testing-bar metrics: no declared metric/baseline command to compare. No independent check beyond the named same-host executions.

## The review status as filed

## REVIEW — corpus-pv-008
- state: recommendation-ready
- verdict: accept
- ref: _REVIEW_NOTES.md
- abstract: PV-008 reproduces the recorded four lines at d27b0d7; the path-only repair pays the round-2 condition. No blocking or owed findings.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T21:50:35Z

## REDACTED

none
