# REVIEW — corpus-pv-020 · round 4

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T17:28:29Z

## Summary

The case reproduces the three control environments and catches a removed limit; correct the stale reproduction command in CASE.md before landing. macOS remains unrun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-020

### Ground

Reviewed `9303b2d...e98e4e5`; `git rev-parse --short=7 HEAD` printed `e98e4e5`. The two commits change only `test/corpus/cases/PV-020/` outside `expected/`, plus PV-020's status object in `test/corpus/manifest.json:356`. `git diff --check 9303b2d...e98e4e5` passed. No new third-party dependency or install script; both Go modules declare Go 1.26. The review uses the contract in `CASE.md` §2, not the author's pass claim.

`bash test/corpus/cases/PV-020/check.sh` returned exit 0 on Linux with Go 1.26.5. Exact output:

```text
run a description-present-listener-allowed control in-process result passed
run a description-present-listener-allowed control description result passed
run a description-present-listener-allowed control listener result passed
run b description-absent control in-process result passed
run b description-absent control description result could not run: description not provisioned
run b description-absent control listener result passed
run c listener-refused control in-process result passed
run c listener-refused control description result passed
run c listener-refused control listener result could not run: operation not permitted
route GET /v1/health oversize base accepted (accepted) change refused (body exceeds limit)
route POST /v1/items oversize base accepted (accepted) change refused (body exceeds limit)
route POST /v1/search oversize base accepted (accepted) change refused (body exceeds limit)
route PUT /v1/items/current oversize base accepted (accepted) change refused (body exceeds limit)
route POST /v1/uploads/media oversize base accepted (accepted) change accepted (exempt upload)
route POST /v1/uploads/bundle oversize base accepted (accepted) change accepted (exempt upload)
run limit-removed control in-process result failed: inproc.limit-present, inproc.over-limit-refused GET /v1/health, inproc.over-limit-refused POST /v1/items, inproc.over-limit-refused POST /v1/search, inproc.over-limit-refused PUT /v1/items/current, inproc.over-limit-exempt POST /v1/uploads/media, inproc.over-limit-exempt POST /v1/uploads/bundle
```

`GOPROXY=off GOSUMDB=off go -C test/corpus/cases/PV-020/change test -count=1 ./...` passed (`service`); command packages have no tests.

### Differential

The six route lines show base accepts all six oversize bodies; change refuses exactly four non-exempt routes and accepts both documented uploads. `change/service/service.go:78` applies the decision to the HTTP handler via `decideSize`, and `change/service/controls.go:212` exercises actual streamed requests. Present description and allowed listener pass. Removing the description makes only the description group unrunnable; simulated `EPERM` makes only the listener group unrunnable. Neither changes the production limit. This is the distinction required by `CASE.md` §2.

### Unplug

`check.sh:79` removes the limit only in the in-process control invocation: its printed failure names the missing limit and all six oversize-route checks. This negative control would not produce the same answer if the limit were present. It does not prove that the listener or description controls would fail on a defective implementation: their own mutation controls were not run. The provided description is synthetic; a different external description could disagree despite these same 16 lines. The same lines also permit either merge threshold for an unrunnable control; they establish a human scope decision, not which human threshold is right. The historical provisioned-host measurement asserted in `expected/EXPECTED.md:14` does not follow from this synthetic check alone.

### Closure

The round-two handler change removes the conflicting 8 MiB ceiling for exempt uploads in `change/service/service.go:129`; `change/service/service_test.go:11` checks both uploads above that former ceiling. The portable scratch path and cleanup are in `check.sh:8-13`, and the manifest restores a status object. Those measurable round-two repairs are present. The exact wording of both earlier review conditions was not supplied in this worktree; their verbatim completeness cannot be certified against the prior verdicts.

### Remainder

**Condition — `test/corpus/cases/PV-020/CASE.md:98`:** The newly recorded reproduction command still names `tests/corpus/cases/PV-020/check.sh`, but this tree uses `test/corpus/`. Running the documented command returned `bash: tests/corpus/cases/PV-020/check.sh: No such file or directory`, exit 127. Correct the command to `bash test/corpus/cases/PV-020/check.sh`, then rerun it from the repository root; otherwise a judge following §8 cannot reproduce the 16 lines. The actual script runs from any working directory when invoked with its valid path.

### Profile

Go/backend/docs. `check.sh` uses Bash, `mktemp -d` with a template, `go -C` (Go 1.26), `cd --`, and standard POSIX tools; no Linux-only call was found in its source. Its real listener uses loopback TCP and may be refused by host policy; the case simulates that refusal for run (c). Linux ran; macOS behavior was not measured. The reviewed files contain no observed secret, e-mail address, internal host, machine name, private repository name, or absolute private path. The synthetic description is case-local. Check ranking: the synthetic description is project-owned; in-process and listener checks are suite-owned, so a green line is not independent external validation.

### Focus

Six requested properties: fixture reproduction yes on Linux; expected **class** follows while historical resolution and owner threshold are not proved by the synthetic output; removed-limit negative control fires; no private data found; changed paths stay in lane; script source is portable on its stated prerequisites, but §8's pasted invocation fails. The primary finding is the stale invocation in `CASE.md:98`, not a limit defect.

### Conditions

Fix the command in `test/corpus/cases/PV-020/CASE.md:98` to the current `test/corpus` path and rerun the published invocation. Do not downgrade or hide the two deliberately unrunnable control groups. This review does not reclassify earlier review conditions.

### Unrun

macOS execution: no macOS host in this leg. Independent mutation of the description and listener assertions: not run; only the supplied limit-removed control was run. Historical provisioned-host controls and the two original review documents: not available in this review's supplied evidence.

## The review status as filed

## REVIEW — corpus-pv-020
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The case reproduces the three control environments and catches a removed limit; correct the stale reproduction command in CASE.md before landing. macOS remains unrun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T17:20:36Z

## REDACTED

none
