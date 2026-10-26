# REVIEW — corpus-pv-020 · round 5

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 2 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T22:30:05Z

## Summary

Stale-path condition addressed; 16 lines reproduced on Linux with exit 0; only a non-blocking documentation clarification remains.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-020

### Ground
Reviewed `9303b2d...ec1c97f4b696f56369db26e18d5c6e77ab1eb56a` on `review/corpus-pv-020-review-r7`. `git rev-parse --short=7 HEAD` returned `ec1c97f`. `go version` returned `go1.26.5 linux/amd64`; `uname -sm` returned `Linux x86_64`. The master's macOS check is a claim, not this seat's measurement. I read the answer key after the contract, source and observed output.

### Differential
`git diff e98e4e5..ec1c97f` changes only `test/corpus/cases/PV-020/CASE.md:98`: `tests/corpus` becomes `test/corpus`. `git diff --check e98e4e5..ec1c97f` returned zero. `git diff --name-only 9303b2d...ec1c97f` lists only PV-020 files and `test/corpus/manifest.json:347`; no file outside this lane. `rg -n 'tests/|fleet/|conformance/' test/corpus/cases/PV-020 test/corpus/manifest.json` found no matches. The documented command exists, the retired command does not. Last-commit diff contains no private paths. Earlier stale-path findings: ADDRESSED.

Measured from the repository root: `bash test/corpus/cases/PV-020/check.sh` returned exit 0 and exactly the 16 lines recorded at `test/corpus/cases/PV-020/CASE.md:102`–`117`:

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

### Unplug
`test/corpus/cases/PV-020/check.sh:79` invokes the in-process controls with the limit removed: the last output line reports `failed` and names missing checks and newly accepted oversize bodies. False-world answer if the limit were missing: `failed`, not the normal `passed`. This suite-level negative control is author-controlled and by itself cannot prove the source was not made to hard-code output. The paired route output and `test/corpus/cases/PV-020/change/service/service.go:71`–`95` independently show the intended production decision. The check's exit 0 means execution succeeded, not that the negative control passed.

### Closure
Accept the path fix and the reproduced fixture. The answer key's human-scope-decision label follows from no demonstrated limit defect and the deliberately unmeasured claims. Run (a) closes both controls on this Linux seat; run (b) leaves description unmeasured (`description not provisioned`), and run (c) leaves listener unmeasured (`operation not permitted`). The same output fits the opposite *policy choice* about whether an unmeasured control blocks merging; evidence cannot settle that owner's policy. It does not establish historical reviewer environments. No BLOCKING finding.

### Remainder
OWED, non-blocking: `test/corpus/cases/PV-020/CASE.md:92` says macOS was not run for its original Linux capture, whereas the master's later macOS claim is unverified on this seat. If this sentence is intended to summarize later runs, clarify its date scope in a future documentation edit. No change to reviewed code requested. Declined to judge: licenses, historical reviewer runtime permissions, macOS behavior. No numeric test metric, baseline, or tolerance was declared, so no metric was inferred. Check ranking: the shell fixture is suite-level, source diff and command existence are project-level, no external check ordered. No bar-weakening move or exception in the path-only repair.

### Profile
Documentation-only Go fixture repair. Checked the documented path, script, paired base/change decision and controls; no code behavior or dependency delta in the repair.

### Focus
FILE·LINE·CONTROL: `test/corpus/cases/PV-020/CASE.md:98` names the existing script. `test/corpus/cases/PV-020/check.sh:35`–`79` runs three groups three ways, prints the paired routes and executes the unplug control. `test/corpus/cases/PV-020/expected/EXPECTED.md:3`–`15` distinguishes an environmental non-run from a product defect. Previous path finding ADDRESSED; no previous finding NOT ADDRESSED.

### Conditions
OWED only: clarify the time scope of the macOS sentence if documentation is revised. No blocking condition.

### Unrun
none

## The review status as filed

## REVIEW — corpus-pv-020
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Stale-path condition addressed; 16 lines reproduced on Linux with exit 0; only a non-blocking documentation clarification remains.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T22:27:01Z

## REDACTED

none
