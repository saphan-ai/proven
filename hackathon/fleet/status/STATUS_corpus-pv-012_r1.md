# STATUS — corpus-pv-012 · round 1

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 8b50d74464903d77c6eb2cc436aa6e7f8329ab90 [corpus-pv-012] tests: build scoped page corpus case
- updated: 2026-10-22T15:57:53Z

## Summary

Seat model unreported by this run (unknown); application's page: base 5 foreign events, change 5; named-team control: base 5, change 0.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-012
- state: stop-2-ready
- abstract: Seat model unreported by this run (unknown); application's page: base 5 foreign events, change 5; named-team control: base 5, change 0.
- last commit: 8b50d74464903d77c6eb2cc436aa6e7f8329ab90 [corpus-pv-012] tests: build scoped page corpus case
- done: Built and checked the PV-012 base, change, control, vacuity arm, and independent reference.
  - `bash tests/corpus/cases/PV-012/check.sh` exit 0: seven input lines, recorded verbatim in CASE.md §8.
  - `GOTOOLCHAIN=local GOPROXY=off go test -count=1 ./...` exit 0: all packages passed or had no test files.
  - `go vet ./tests/corpus/cases/PV-012/base ./tests/corpus/cases/PV-012/change` exit 0.
  - `git diff --stat origin/main...HEAD` lists only PV-012 outside expected/ and tests/corpus/manifest.json; origin/main is 6bf302d, and local main is unavailable.
- next / blocked-on: master verification; no push, merge, or PR by this leg.
- artifacts:
  - tests/corpus/cases/PV-012/CASE.md
  - tests/corpus/cases/PV-012/base/main.go
  - tests/corpus/cases/PV-012/change/main.go
  - tests/corpus/cases/PV-012/check.sh
  - tests/corpus/manifest.json
  - ███████████████████████████████████████████
  - ██████████████████████████████████████████
- note: Local toolchain go version go1.26.5 linux/amd64. The control-no-registry input reports control=FAIL while the script exits 0 as a measurement. The model was not exposed in the run metadata available to this leg; no model identity was inferred. No independent review was run.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
