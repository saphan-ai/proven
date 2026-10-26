# STATUS — corpus-pv-006 · round 32

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 4 min
- last commit: 658808c165d17f068f7826b078e7e0ce19d467d1 [corpus-pv-006] tests: prove the raw token reaches the visitor
- updated: 2026-10-23T09:12:04Z

## Summary

GPT-6; dollar-token base=pass change=pass reference=fail, with raw=$wait asserted in the rendered notice. The negative control raises an alarm when the token is removed; duration values are now printed.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-006
- state: stop-2-ready
- abstract: GPT-6; dollar-token base=pass change=pass reference=fail, with raw=$wait asserted in the rendered notice. The negative control raises an alarm when the token is removed; duration values are now printed.
- last commit: 658808c165d17f068f7826b078e7e0ce19d467d1 [corpus-pv-006] tests: prove the raw token reaches the visitor
- done: The review's checker condition is addressed and the manifest condition was already met.
  - `bash tests/corpus/cases/PV-006/check.sh` exited 0 and printed six input lines; CASE.md §8 has the full output.
  - `go -C tests/corpus/cases/PV-006 test -count=1 -v ./...` exited 0; its negative control logged `rendered notice does not contain "$wait"` after removing the token.
  - `go -C tests/corpus/cases/PV-006 vet ./...` exited 0; `gofmt -l` printed no paths; `git diff --check` exited 0.
  - `git diff --stat fe146e0..HEAD` lists only PV-006 files outside expected/ and tests/corpus/manifest.json.
- next / blocked-on: Master verification.
- artifacts:
  - tests/corpus/cases/PV-006/CASE.md
  - tests/corpus/cases/PV-006/check/main.go
  - tests/corpus/cases/PV-006/check/main_test.go
- evidence:
  - Baseline `bash tests/corpus/cases/PV-006/check.sh` exited 0 with six lines before the checker change.
  - `jq` parsed tests/corpus/manifest.json; PV-006 status is already an object with contract, fixture, checks and independent_check.
  - Review condition ADDRESSED: rendered `$wait` is asserted and reported; a missing raw token raises an error in the isolated negative control.
  - Master-was-wrong: the order's round 1 instruction to use a string manifest status contradicted the existing object schema; no manifest edit was needed in round 2.
  - macOS execution UNRUN: no macOS host in this leg; check.sh uses no GNU-only flags.
- note: CASE.md records the 2026-10-23 UTC check run. No expected/ file was read or changed. No coverage or mutation bar was declared.
- updated: 2026-10-23T09:07:42Z

## REDACTED

none
