# STATUS — corpus-pv-010 · round 1

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 3 min
- last commit: b4092e4bd71990e9e04fc81b265bb49d09b35a73 [corpus-pv-010] tests: build PV-010 base, change and check, run it
- updated: 2026-10-22T15:49:05Z

## Summary

Seat model as the run reports it: claude-sonnet-5-5 (Sonnet 5.5). PV-010 built in Go 1.26.8 (stdlib only): base, change, reference, check.sh.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-010
- state: stop-2-ready
- abstract: Seat model as the run reports it: claude-sonnet-5-5 (Sonnet 5.5). PV-010 built in Go 1.26.8 (stdlib only): base, change, reference, check.sh.
  Input `-999`: base=pass, change=pass, reference=stale — change still passes it; so do `+41` and Arabic-Indic 41. The twenty-digit note: base=pass, change=fail.
  Description: change states 0..48, change-edited 0..256, reference probes 0..48; change's own tests pass on both.
- last commit: b4092e4bd71990e9e04fc81b265bb49d09b35a73 [corpus-pv-010] tests: build PV-010 base, change and check, run it
- done: PV-010 case code built and check.sh run; output and exit code recorded in CASE.md §8
  - `bash tests/corpus/cases/PV-010/check.sh` exit 0, printed one line per §2 input (pasted in CASE.md §8)
  - `git diff --stat main...HEAD` lists only files under tests/corpus/cases/PV-010/ (none in expected/) and tests/corpus/manifest.json
- next / blocked-on: none
- artifacts:
  - tests/corpus/cases/PV-010/check.sh
  - tests/corpus/cases/PV-010/CASE.md
  - tests/corpus/manifest.json
- evidence:
  - toolchain go version go1.26.8 darwin/arm64; check run 2026-10-22T15:46Z
  - base and change printed identically for -999, +41, Arabic-Indic 41; only the twenty-digit note differs between them
  - git commit printed packed-refs.lock Operation not permitted; git log -1 shows the commit landed
  - manifest: I set status.fixture to "built" and left checks as "not run"; the order said "status only, to built", so this reading is a judgement call, and checks could also be set to run
  - ████████ scratch is untracked and not committed
- note: I read neither expected/ nor any other case's files. `saphan doc check --family exec-status _EXEC_STATUS.md` exited 0 (a deprecated-alias warning on `next` is now fixed).
- updated: 2026-10-22T15:50:00Z

CLOSE-OK · corpus-pv-010 · commit b4092e4bd71990e9e04fc81b265bb49d09b35a73 · branch fix/corpus-pv-010 · worktree █████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
