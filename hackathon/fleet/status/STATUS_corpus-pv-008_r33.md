# STATUS — corpus-pv-008 · round 33

- state: ready-for-review
- executor: codex
- model: gpt-6-sol
- pool: not recorded
- cost: not recorded
- duration: 5 min
- last commit: 70b4c5e5f7170498c7913d7567862f0d4a401460 [corpus-pv-008] tests: bound PV-008 observations
- updated: 2026-10-23T09:19:34Z

## Summary

Codex (model not reported to this leg); on wedged-stop late subscription, base and change both printed no-frame; Linux check matches the four recorded lines.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — corpus-pv-008
- state: stop-2-ready
- abstract: Codex (model not reported to this leg); on wedged-stop late subscription, base and change both printed no-frame; Linux check matches the four recorded lines.
- last commit: 70b4c5e5f7170498c7913d7567862f0d4a401460 [corpus-pv-008] tests: bound PV-008 observations
- done: Bounded the observations and named failure to enter Stop.
  - The held-lock and no-stop-entry negative controls passed with go -C tests/corpus/cases/PV-008 test -count=1 ./check (exit 0).
  - bash tests/corpus/cases/PV-008/check.sh printed the four lines in CASE.md §8 (exit 0).
  - git diff --stat fe146e0..HEAD showed only PV-008 and manifest files (exit 0).
- next / blocked-on: The master or owner must run the check on macOS and compare its output and exit code with CASE.md §8 before the gate.
- artifacts:
  - tests/corpus/cases/PV-008/CASE.md
  - tests/corpus/cases/PV-008/check/main.go
  - tests/corpus/cases/PV-008/check/main_test.go
  - tests/corpus/manifest.json
- evidence:
  - go -C tests/corpus/cases/PV-008 test -count=1 ./... exited 0; one test package passed, three packages had no test files.
  - go -C tests/corpus/cases/PV-008 test -race -count=1 ./check exited 0.
  - go -C tests/corpus/cases/PV-008 vet ./... and build ./... both exited 0.
  - jq confirmed PV-008 status is an object; the master claim that round 1 changed it to a string was false.
  - macOS check not run because this seat is Linux.
- note: The Linux check exit 0 confirms execution, not an asserted outcome; CASE.md §8 records its full output and reach. The manifest was already an object, so only the check date changed. Review O4-O6 remain owner or later-round matters.
- updated: 2026-10-23T09:14:25Z

## REDACTED

none
