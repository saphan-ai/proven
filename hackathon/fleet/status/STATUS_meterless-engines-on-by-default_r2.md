# STATUS — meterless-engines-on-by-default · round 2

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- last commit: "[meterless-engines-on-by-default] settingspage: bounded-steps note"
- updated: 2026-10-26

## Summary
All four Meterless engines are on by default; each has its own off setting (`PROVEN_MARKOVIAN=off`, `scout_intent: off`, `PROVEN_HMEM=off`, `PROVEN_WORLD_MODEL=off`). ADR-0053 records it. The Scout guard: a change Scout called documentation or dependency that is not verified prose or a verified manifest runs in class `security`, doubled (a change Scout called feature, refactor or configuration runs in the class its rules give it). Round 3 added the switches for H-MEM and the World Model, which this round did not have. The benefit of Markovian and Scout Intent is still UNMEASURED, and the documents say so.

## Acceptance, item by item
1. `go build ./... && go vet ./... && go test -count=1 ./...`: build and vet pass. The test run printed 64 `ok` packages and one failure, `TestSampleSARIFMatchesTheSchema` in `internal/cli/sarifschema` (`open <user cache>/proven/logs/trace.jsonl: permission denied`). That test fails the same way on the base commit before any change, because the run's user cache is not writable.
2. The three guard cases (credentials code under `docs/`, a Polish-named file, a Thai-named file) with Scout on, before the guard: each printed `class docs (docs/**) doubled false, second reviewer calls 0`. After the guard: all three pass (class `security`, doubled, second reviewer called). With the guard's raise switched off again: the same three fail with the same line.
3. `git grep -n -iE 'off by default|OFF by default'` over README, docs and hackathon: the remaining hits are README line 433 (the day-4 row of the log, a record of what was true on 25 October), and sentences in ADR-0014, three older fleet reviews and statuses, and the planning conversation of 19 October, none about these engines.

## Evidence
- Defaults: `TestMeterlessEnginesAreOnByDefault` (both engines on unset, each off by its setting).
- Switch off restores the old review: `TestScoutSwitchedOffRunsTheFullPlanForANonProseChange`.
- A verified-prose change keeps the documentation plan and makes no model call: `TestScoutStillLightensVerifiedProse`.
- One behaviour changed besides the defaults: with Scout on, a live review of a git repository with no baseline is refused by `CheckTarget` with the old sentence, before the plan reads the change. The old test `TestLiveGitWithoutBaselineIsRefused` caught it.
- Two tests that assumed "off" were rewritten: `TestScoutIntentIsOffByDefault` (replaced) and `TestLoadDefaultsAreOnAndBadValuesAreRefusedByName` (renamed, new expectation). `TestPlanMeasurement` now sets its "without plan" side to Scout off explicitly.

## Not run
- Meterless's conformance runners (`markovian 8/8`, Scout Intent's eval): the reference `engines/` directory is not on this machine, so `run.sh` cannot start. No scorecard was changed.
- The blind set is not on this branch and was not run with the guard.

## Blocked on / next
Review. The decision document and the ledger row follow the review.
