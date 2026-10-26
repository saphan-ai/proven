# STATUS — meterless-engines-on-by-default · round 3

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- last commit: "[meterless-engines-on-by-default] review: guard checks manifest names"
- updated: 2026-10-26

## Summary
Round 3 answers the two named reviews of round 2. H-MEM and the World Model now have their own off settings (`PROVEN_HMEM=off`, `PROVEN_WORLD_MODEL=off`), so all four engines can be switched off by a setting. Markovian is on by default in sample mode too, which is the lab's pull-request review. The dependency arm of the Scout guard now checks every old and new path of a change against a closed set of manifest and lockfile names. README, ADR-0053 and the earlier public status say the guard raises only "a change Scout called documentation or dependency that is not verified prose or a verified manifest". The benefit of Markovian and Scout Intent is still UNMEASURED.

## Acceptance, item by item
1. `go build ./...` and `go vet ./...` exit 0. `go test -count=1 ./...` printed 64 `ok` packages and one failure, `TestSampleSARIFMatchesTheSchema` in `internal/cli/sarifschema` (`permission denied` on the user cache). It fails the same way on the base commit; round 2 measured that before any edit.
2. Red before, green after, and a negative control for each new test:
   - Markovian in sample mode: with the sample path forced back to the zero value, `TestSampleReviewRunsMarkovianByDefault` fails with the trace line "the engine is off (PROVEN_MARKOVIAN is off)"; with the real code it passes.
   - The two store switches: with the switch ignored in `begin`, `TestHMemOffNeitherOpensReadsNorWrites`, `TestWorldOffNeitherOpensNorFolds` and `TestBothStoresOffLeaveNoState` fail ("opened with its switch off"); with the real code they pass. `requireJudged` also proves the Judge still decides every finding with a store off (a first draft skipped the Judge when H-MEM was off; the test caught it).
   - The dependency arm: `notManifestOnly` made to return "" turns `TestScoutNeverLightensVendoredCode`, `TestScoutNeverLightensAManifestRenamedToSource` and six rows of `TestNotManifestOnlyChecksEveryPath` red. Before the change, `Pipfile` renamed to `app/x.py` ran in class `backend`, not doubled, second reviewer calls 0.
3. `git grep -n -iE 'off by default|OFF by default'` over README, docs and hackathon: the same hits as round 2 (README day-4 row of the log, ADR-0014, older reviews and statuses, the planning conversation of 19 October). None is about these engines.

## What the measurement contradicted
- The order named `R go.mod -> internal/x/x.go` as a red case. With the default registry Scout answers class security for that rename (`go.mod` is also a security path), so it is green before and after. The reachable red rename is `Pipfile -> app/x.py`; both are tested.
- The order said the vendored-only test is red at the base. It is green there: the vendored check already existed, and what was missing was a test for it. The negative control turns it red.
- With Markovian on in sample mode, the `fix` fixture lost its second fix attempt (its recorded answer is keyed by an exact prompt, and the carried-state prompt differs). One answer row was added to `testdata/fixture/fix/fixture.json`; the sample output is the same as before.

## Not run
- Meterless's conformance runners (`markovian 8/8`, H-MEM, World Model, Scout Intent's eval): the reference `engines/` directory is not on this machine. No scorecard was changed.
- The blind set was not run with the guard.

## Blocked on / next
Review of this round's conditions (C1 to C4). The decision document and the ledger row follow the review.
