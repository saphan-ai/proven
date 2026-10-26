# STATUS — corpus-pv-019 · round 1

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 2 min
- last commit: be9cc026b63654f94af38fe31d44a8921ebf4170 [corpus-pv-019] tests: build PV-019 case code and record the check run
- updated: 2026-10-22T16:09:59Z

## Summary

Seat model as the run reports it: claude-sonnet-5-5 (Sonnet 5.5). PV-019 case code built in Go go1.26.8 (stdlib only) and check.sh run.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-019
- state: stop-2-ready
- abstract: Seat model as the run reports it: claude-sonnet-5-5 (Sonnet 5.5). PV-019 case code built in Go go1.26.8 (stdlib only) and check.sh run.
  For the input the case is about (a preview row with session switch / plugin folder, an unsupported row with an install route): base fails all (unknown status "preview"), change passes all three; the reference refuses them.
  Wrong non-empty channel: change passes, shipped content test fails; empty channel: change and content test both fail.
- last commit: be9cc026b63654f94af38fe31d44a8921ebf4170 [corpus-pv-019] tests: build PV-019 case code and record the check run
- done: base, change, content test, reference and check.sh built under tests/corpus/cases/PV-019/; check run; CASE.md §8 and manifest status updated
  - bash tests/corpus/cases/PV-019/check.sh printed 9 lines, no failure reported (output pasted in CASE.md §8)
  - go -C tests/corpus/cases/PV-019 test ./change: ok
  - git diff --stat main...HEAD lists only PV-019 files outside expected/ and tests/corpus/manifest.json
- next: none
- note: git commit printed packed-refs.lock Operation not permitted; git log -1 shows the commit landed. Worktree base is main at 5d8a8b8, not d81a93d named in the order. The exact exit code of check.sh could not be printed (hook refuses $?); date was denied, so the run time is the run-directory stamp 16:05 UTC. Manifest: only the fixture field was set to built; checks left as it was, per the order's words.
- artifacts:
  - tests/corpus/cases/PV-019/CASE.md
  - tests/corpus/cases/PV-019/check.sh
  - tests/corpus/manifest.json
- updated: 2026-10-22

CLOSE-OK · corpus-pv-019 · commit be9cc026b63654f94af38fe31d44a8921ebf4170 · branch fix/corpus-pv-019 · worktree █████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
