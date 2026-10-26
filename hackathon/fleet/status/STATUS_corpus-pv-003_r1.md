# STATUS — corpus-pv-003 · round 1

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 2 min
- last commit: 6167baba76819aa9e6a548dae02c14c213f6393e [corpus-pv-003] tests: build PV-003 case code, record the run
- updated: 2026-10-22T15:31:02Z

## Summary

Model: Sonnet 5.5 (claude-sonnet-5-5). Built PV-003 in Go 1.26.5 (stdlib only): base, change, reference, check.sh.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-003
- state: stop-2-ready
- abstract: Model: Sonnet 5.5 (claude-sonnet-5-5). Built PV-003 in Go 1.26.5 (stdlib only): base, change, reference, check.sh.
  Input move-moved-tenant: base=refused(X), change=allowed, reference=refused(X). Input vacated-slot: base=refused(X), change=refused(X), reference=allowed.
  check.sh exit 0; output pasted in CASE.md §8; manifest status set to built.
- last commit: 6167baba76819aa9e6a548dae02c14c213f6393e [corpus-pv-003] tests: build PV-003 case code, record the run
- done: PV-003 case code built and run, record written
  - `bash tests/corpus/cases/PV-003/check.sh` printed five lines, exit 0 (CASE.md §8)
  - `gofmt -l` and `go vet` on the case directory: no output
  - `git diff --stat HEAD~1 HEAD`: only files under tests/corpus/cases/PV-003 (not expected/) and tests/corpus/manifest.json
- next: none
- note: `git diff main...HEAD` could not run, because no local ref `main` exists in this worktree; the diff was taken against HEAD~1, the only commit of this leg. The `date` program is denied for this seat, so the UTC time of day is not recorded in §8, only the date. `git commit` printed `packed-refs.lock: Permission denied`; `git log -1` shows the commit landed. Order claim refuted: none; §2 was buildable as written. Build choices the contract left open are listed in §8.
- artifacts:
  - tests/corpus/cases/PV-003/check.sh
  - tests/corpus/cases/PV-003/base/gate.go
  - tests/corpus/cases/PV-003/change/gate.go
  - tests/corpus/cases/PV-003/check/main.go
  - tests/corpus/cases/PV-003/check/reference.go
  - tests/corpus/cases/PV-003/CASE.md
  - tests/corpus/manifest.json
- updated: 2026-10-22

## REDACTED

none
