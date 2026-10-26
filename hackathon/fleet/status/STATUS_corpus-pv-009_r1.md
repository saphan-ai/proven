# STATUS — corpus-pv-009 · round 1

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 018e8f15ca5bfe82a3c73af60b078c0584aee803 [corpus-pv-009] tests: build overflow validation corpus
- updated: 2026-10-22T15:48:38Z

## Summary

Codex seat (model not exposed to this leg); overflow-int64 printed base=pass, change=pass, reference=fail; committed the case and check.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-009
- state: stop-2-ready
- abstract: Codex seat (model not exposed to this leg); overflow-int64 printed base=pass, change=pass, reference=fail; committed the case and check.
- last commit: 018e8f15ca5bfe82a3c73af60b078c0584aee803 [corpus-pv-009] tests: build overflow validation corpus
- done: Built and ran PV-009 from its contract.
  - `bash tests/corpus/cases/PV-009/check.sh` exited 0 and printed nine input lines; §8 reproduces all lines.
  - `go test -count=1 ./tests/corpus/cases/PV-009/...` exited 0; one test function with six subtests passed.
  - `go vet ./tests/corpus/cases/PV-009/...` exited 0.
  - `git diff --stat origin/main...HEAD` lists only PV-009 and the manifest; local `main` is unavailable, so `git diff --stat main...HEAD` was not run successfully.
- next / blocked-on: Master verifies the recorded result; no owner act requested by this leg.
- artifacts:
  - tests/corpus/cases/PV-009/CASE.md
  - tests/corpus/cases/PV-009/base/guard.go
  - tests/corpus/cases/PV-009/change/guard.go
  - tests/corpus/cases/PV-009/change/guard_test.go
  - tests/corpus/cases/PV-009/check.go
  - tests/corpus/cases/PV-009/check.sh
  - tests/corpus/manifest.json
- note: Toolchain go1.26.5 linux/amd64. No declared numerical testing-bar metric was present in the case order. Model identity beyond the codex backend was not made available to the leg. The referenced base d81a93d is older than the branch's parent 24f27b6; origin/main at 24f27b6 is the appropriate comparison. Commit succeeded despite the post-commit packed-refs.lock permission message, confirmed by git log -1.
- updated: 2026-10-22T15:46:17Z

## REDACTED

none
