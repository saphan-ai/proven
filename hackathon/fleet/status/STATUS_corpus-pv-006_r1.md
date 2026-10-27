# STATUS — corpus-pv-006 · round 1

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 2 min
- last commit: 13c27e650a793f154aa808f4d338cc3cc1c71ce2 [corpus-pv-006] tests: build PV-006 case code and record its run
- updated: 2026-10-22T15:37:26Z

## Summary

Model: Sonnet 5.5 (claude-sonnet-5-5). PV-006 built in Go 1.26.5 (stdlib only); check.sh exit 0.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-006
- state: stop-2-ready
- abstract: Model: Sonnet 5.5 (claude-sonnet-5-5). PV-006 built in Go 1.26.5 (stdlib only); check.sh exit 0.
  For the input the case is about, the `$wait` token: base=pass, change=pass, reference=fail (raw token reaches the visitor, both controls green).
  Also: duration 0 vs 1 base=pass change=pass reference=fail; unlisted `{window}` base=pass change=fail reference=fail.
- last commit: 13c27e650a793f154aa808f4d338cc3cc1c71ce2 [corpus-pv-006] tests: build PV-006 case code and record its run
- done: PV-006 case code built, check run, CASE.md §8 and manifest status updated
  - check.sh: `bash tests/corpus/cases/PV-006/check.sh` exit 0, six lines, pasted in CASE.md §8
  - scope: `git diff --stat 0afaeb2 HEAD` lists only tests/corpus/cases/PV-006 (outside expected/) and tests/corpus/manifest.json
  - gofmt -l on the case directory: clean after one gofmt -w
- next: none
- note: base ref differs from the order's d81a93d; the worktree started at 0afaeb2 and the diff was measured from it. `main` is not a ref in this worktree, so `main...HEAD` could not run. `git commit` printed packed-refs.lock Permission denied; git log -1 shows the commit landed. The `date` program is denied for this role, so the UTC time came from the bash printf time builtin. expected/ was not opened. No tests were written beyond check.sh; no go vet run.
- artifacts:
  - tests/corpus/cases/PV-006/check.sh
  - tests/corpus/cases/PV-006/CASE.md
  - tests/corpus/manifest.json
- updated: 2026-10-22T15:40:00Z

## REDACTED

none
