# REVIEW — corpus-pv-008 · round 4

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T17:52:58Z

## Summary

The wedged-stop check establishes the named change defect; update stale corpus paths in the reproduction instructions. Linux check and unit/race tests pass; macOS is unrun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-008

### Ground
`git rev-parse --short=7 HEAD` printed `1970cfc`. Reviewed `9303b2d...1970cfc` (two commits), `CASE.md` §2 before `expected/EXPECTED.md`, the three session implementations, the check, its tests, and the manifest delta. `git diff --name-status` lists only eight case files outside `expected/` and `test/corpus/manifest.json`; `git diff --check` exits 0. The manifest changes only PV-008's `fixture` and `checks` values; its status remains an object. `go version` printed `go1.26.5 linux/amd64`; the case pins Go 1.26 with standard library only.
remark: The diff census excludes untracked channel files; the toolchain reading is Linux-only.

`bash test/corpus/cases/PV-008/check.sh` exited 0 and printed exactly:
remark: Each missing result means no observation within one second, not proof from output alone of perpetual blockage.

```
normal-close-then-late-subscriber | no-frame | frame:ended | frame:ended
wedged-stop-then-late-subscriber | no-frame | no-frame | frame:ended+alarm:worker-not-released
wedged-stop-then-read-state | state:closed | state:closed | state:closed
wedged-stop-then-second-close | returned | blocked | returned
```

`go -C test/corpus/cases/PV-008 test -count=1 ./check` exited 0 (`ok pv008/check 4.005s`); the same command with `-race` exited 0 (`ok pv008/check 5.014s`).
remark: The two tests exercise the check package on this Linux seat, not a macOS run or the full repository suite.

### Differential
The normal input deliberately separates base (no late replay) from change/reference (late replay). With a wedged stop, base and change both lack a frame within one second; change alone blocks a second close, while reference returns with an alarmed final frame. State is closed in all three, as §2 predicts. The code paths corroborate this: `change/session.go:56-77` marks closed before the unbounded `Stop`, then stores and publishes the frame and closes `done` only after Stop; `reference/session.go:55-84` bounds Stop at 100 ms and publishes an alarmed frame on expiry. `base/session.go:35-57` has no late replay and returns on a second close. No check output is hard-coded in the runner: `check/main.go:163-174` executes each input on all three implementations and prints their returned observations.

The expected established-defect label follows from the wedged stop, missing frame, closed state, blocked second close and reference contrast, together with inspection of the unbounded call. False world: a slow but bounded Stop longer than the one-second observation limit would print the same `no-frame` and `blocked` lines; the output alone cannot prove never. The inspected change source and stand-in worker supply that missing causal evidence. Neither the check nor the answer key proves a severity choice between the two historical reviewers; the key correctly limits its claim to the defect.

### Unplug
`check/main_test.go:11-54` supplies a held-lock control: Subscribe, State and second Close become observably blocked, rather than silently matching the expected frame/state results. `check/main_test.go:56-69` supplies a Close that never calls Stop: the check reports `stop-not-entered`. These controls were executed by both test commands above. The test checks the timeout and stop-entry paths, not a full mutation score or an altered production implementation. The reference's alarm is derived from its own bounded Stop; the check's observations are not asserted against literal expected strings, so review of the printed values remains necessary.

### Closure
Round-two bounded observations and the two named negative controls are present and run; the manifest's status object is restored. The exact wording of both earlier reviews' conditions was not provisioned, so matching those conditions verbatim is unverified. One new owed finding follows from the later layout move: `test/corpus/cases/PV-008/CASE.md:92-107` and `test/corpus/cases/PV-008/check.sh:3` tell readers to run paths under `tests/corpus/`, but the checked-in module lives under `test/corpus/`. The documented command fails to locate the file from this repository root. Remedy: correct every path in §8 and the script header, and run the printed commands from the root. The executable script itself works at the actual path; this is a documentation failure, not a false check result.

### Remainder
The case's synthetic base intentionally lacks late replay; do not infer the original private base did. The review checked the introduced case and manifest delta, not the original historical source, reviewer severity rule, or other corpus cases. The corpus reports no declared numerical coverage/mutation/lint bar for this case; no such metric was invented. No secret, email, private repository name, machine name, or private absolute path was found in the case diff. `bash`, `dirname`, `BASH_SOURCE`, `cd` and `go run` in the script have no GNU-only options and are available in ordinary Bash/Go environments on Linux and macOS; the actual macOS run remains unverified.

### Profile
Go fixture and documentation/test-harness profile. The only manifest delta is status text, with no added dependency or install script. Check strength, negative controls, portability, and the limits of timeout observations are the material risks. Checks rank from least circular to most circular: source inspection plus contrasting reference; independent runtime outputs; unit controls over helper failures; the case author's recorded four-line output. No numeric bar was declared or changed.

### Focus
File · line · control: `change/session.go:63-77` — a wedged worker prevents final publication after state closes; the check observes it against `reference/session.go:65-84`. `check/main.go:80-109` — calls and frame reads are bounded and blocked operations have distinct labels. `check/main.go:111-127` — worker Stop entry is required before observation. `CASE.md:92-107` — the written reproduction commands point to a directory that no longer exists. The judge's `bash test/corpus/cases/PV-008/check.sh` succeeds on Linux; there is no demonstrated shell-portability defect.

### Conditions
OWED: Update stale `tests/corpus/` references in `test/corpus/cases/PV-008/CASE.md:92-107` and `test/corpus/cases/PV-008/check.sh:3` to `test/corpus/`; execute the corrected documented commands from the repository root. This does not require changing the causal fixture or its expected result. No other condition established in this review.

### Unrun
macOS execution (no macOS runner on this seat); full historical round-one condition comparison (their verbatim texts were not available in this worktree); full repository suite and a mutation score (outside this case review, no declared metric). These are limitations, not passes.

## The review status as filed

## REVIEW — corpus-pv-008
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The wedged-stop check establishes the named change defect; update stale corpus paths in the reproduction instructions. Linux check and unit/race tests pass; macOS is unrun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T17:46:04Z

## REDACTED

none
