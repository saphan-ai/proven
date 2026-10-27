# REVIEW — corpus-pv-008 · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:23:42Z

## Summary

PV-008 reproduces the expected defect, but its check violates the contract's bounded-observation rule and can stop before reporting all inputs.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-008

### Ground

Reviewed `fe146e0...4e43d38` at measured HEAD `4e43d38571185fcf1c6f128329de6d342a324320`. The diff changes seven files in `tests/corpus/cases/PV-008/` outside `expected/` and only PV-008's status fields in `tests/corpus/manifest.json`. `git diff --check fe146e0...4e43d38` exited 0. The available toolchain reported `go version go1.26.5 linux/amd64`; the case module pins `go 1.26`. `go -C tests/corpus/cases/PV-008 test -count=1 ./...` exited 0, reporting four packages with no test files. That command is a build check, not evidence that the scenario was exercised. `saphan doc check _REVIEW_STATUS.md --family review-status` exited 0 and reported `checked=true`.

I ran `bash tests/corpus/cases/PV-008/check.sh` at the review tip. It exited 0 and printed exactly:

```text
normal-close-then-late-subscriber | no-frame | frame:ended | frame:ended
wedged-stop-then-late-subscriber | no-frame | no-frame | frame:ended+alarm:worker-not-released
wedged-stop-then-read-state | state:closed | state:closed | state:closed
wedged-stop-then-second-close | returned | blocked | returned
```

remark: These lines report one-second observations; they cannot alone distinguish a permanent wedge from a stop delayed beyond one second.
remark: Exit 0 means the check printed its table; the controlled publication-before-stop mutation also exited 0 with different results.

### Differential

The normal path separates base's missing replay from change and reference. The three wedged inputs separate change from the bounded reference on frame delivery and repeated close, while state is closed in all three. The output matches `expected/EXPECTED.md`: it establishes no frame **within the one-second wait**, state closed, no alarm in change, and a second close that has not returned within the wait. It does not by itself prove a permanent stall: a stop taking more than one second would give the same output. The source supplies the stronger premise: the wedged worker waits on `release` in `check/main.go:33-37`; change calls it before creating the frame at `change/session.go:57-64`; and the check releases the worker only after each observation at `check/main.go:110-127`. The reference bounds stop and emits the alarm at `reference/session.go:58-78`. These facts support the expected defect. The reviewers' historical severity disagreement is not a different observed fact.

### Unplug

In a disposable copy confined to `████████████████████████████████████████`, I moved change's `Stop` after final-frame publication. The check then printed `frame:ended` for change's wedged late subscriber and `returned` for change's wedged second close. Its exit status remained 0. Thus the printed comparison detects removal of the defect, but the process exit status does not assert any expected result (`check/main.go:130-138`). Reading exit 0 as a pass is invalid; the contract asks for printed observations, so this is a measurement limit rather than a separate rejection reason.

I also made a disposable copy in which change retained its mutex while entering the wedged stop, a direct inverse of the required unlock at `change/session.go:57-60`. Running that copy with an external five-second bound printed the normal line, then Go reported a deadlock in `Subscribe`; it did not print the remaining three input lines. This is the false-world answer to the check's claimed bounded-observation property.

### Closure

**BLOCKING — `check/main.go:92-98,105-118` and `CASE.md` §2.** The contract requires a fixed wait limit for each observation so the check always ends. `start` discards timeout results. Both `s.Subscribe()` and `s.State()` run synchronously outside any wait limit. If a version holds the session lock during a wedged stop, `Subscribe` blocks before `frame` can start its timer; the controlled mutation above triggered the failure. A similar held lock would block the state read. Remedy: bound entry to `Subscribe` and `State`, handle `start` timeouts explicitly, and print a distinct timeout or error result rather than confusing it with an observed `no-frame`. Add a check that introduces a held lock and verifies the harness completes with a named timeout for every input. This is a harness defect, not a claim that the current change version holds the lock.

No other blocking or owed finding was established. The fixture itself follows §2's base, change, and reference sequence. The check was not bent to make the expected lines appear; its calls use the three concrete implementations. The privacy scan over changed case code, script, module and case document found no internal host or seat name, internal path, private repository, token, or email address. The diff does not touch `expected/`.

### Remainder

The disposable mutation directory remains under `███████████████████████████████████████████████`. Removal was refused by the command approval layer: `rm -f style commands are not permitted. Use a safer approach`. Under the denied-act rule I did not attempt another spelling of the deletion. The pre-existing untracked channel and provision files remain. This prevents a claim that the worktree was restored to its initial state; the review verdict and its evidence remain complete.

### Profile

Docs and Go test-harness review. No declared numeric quality bar, baseline or metric command was found for this case, so no coverage or mutation percentage is claimed. The checks rank as suite-level evidence: the case's own printed observations, with the reviewer mutation providing an independent challenge to their reach. No dependency changed. No skip, suppression, threshold, or assertion was weakened in the diff.

### Focus

`check.sh` uses Bash arrays only through `BASH_SOURCE[0]`, `dirname`, `cd`, and `exec go run`; the script has no GNU-specific option and these constructs are available in Bash on Linux and macOS. Go source uses the standard library. The measured run was Linux only. A judge needs Bash and a compatible Go toolchain on the path. The fixed one-second observation can be sensitive to a heavily loaded host, and the check does not distinguish a slow stop from a never-returning stop by output alone. I declined to judge licence questions and the historical reviews beyond the case contract, as the order excludes them.

remark: The Linux run and shell syntax review do not establish execution on macOS.

### Conditions

Reject until the check bounds every operation named as an observation and treats inability to enter `Stop` as a named failure. The master should verify the repair with the held-lock negative control before a gate. No severity reclassification is proposed.

### Unrun

macOS execution: no macOS host was available. Full repository suite and race suite: outside this case review's bounded scope, and no such run is claimed. Public-copy publication and bridge registration: this review leg is assigned only the channel return; those acts are for the track roles that file the review and decision.

## The review status as filed

## REVIEW — corpus-pv-008
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: PV-008 reproduces the expected defect, but its check violates the contract's bounded-observation rule and can stop before reporting all inputs.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T18:18:31Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
