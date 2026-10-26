# STATUS — corpus-pv-008 · round 1

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: c69475e97641a71f37b1dc9dfe2e0d0c6251e1a9 [corpus-pv-008] tests: build PV-008 base, change, reference and check
- updated: 2026-10-22T15:43:14Z

## Summary

Model: claude-sonnet-5-5 (Sonnet 5.5), as the seat reports it. Built PV-008 in Go 1.26.5, standard library only.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-008
- state: stop-2-ready
- abstract: Model: claude-sonnet-5-5 (Sonnet 5.5), as the seat reports it. Built PV-008 in Go 1.26.5, standard library only.
  On the wedged stop then late subscriber, base printed no-frame and change printed no-frame (reference: frame with alarm).
  On the wedged second close, change printed blocked; base and reference returned. check.sh exit code 0.
- last commit: c69475e97641a71f37b1dc9dfe2e0d0c6251e1a9 [corpus-pv-008] tests: build PV-008 base, change, reference and check
- done: base, change, reference and check built and run; output recorded in CASE.md section 8
  - bash tests/corpus/cases/PV-008/check.sh printed four lines, exit 0
  - gofmt -l on the case directory printed nothing
  - git diff --stat 24f27b6 HEAD lists only PV-008 files and manifest.json (main ref is not resolvable in this worktree, so the diff base is 24f27b6, the recorded HEAD at start)
- next / blocked-on: none
- artifacts:
  - tests/corpus/cases/PV-008/CASE.md
  - tests/corpus/cases/PV-008/check.sh
  - tests/corpus/manifest.json
- note: Judgement calls. (1) Where section 2 is silent, a second close on base returns at once; on change and reference it waits for the first close to publish. Cost if wrong: the blocked/returned cell for base or change would differ. (2) The manifest has fixture set to built with date and checks set to run with date, to match the CASE.md header, not only the one word built. (3) Node is not on PATH here; Go was the named language, so it did not matter. (4) expected/ was never opened. (5) git commit printed packed-refs.lock Permission denied; git log -1 shows the commit landed.
- updated: 2026-10-22T15:45:00Z

## REDACTED

none
