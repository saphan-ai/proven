# STATUS — a-fix-is-proved-by-the-same-test · round 2

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 6e2d6bc [a-fix-is-proved-by-the-same-test] merge: absorb main
- updated: 2026-10-24T06:19:58Z

## Summary

Round 2. A proof run now counts as passed only when its go test -json output holds a pass event for the proving test (a package run: the package pass and at least one passed test); exit code 0 alone is recorded as not run. The Prover's runs, the sample replay and the fixer share the judgement (C1).

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-fix-is-proved-by-the-same-test
- state: stop-2-ready
- abstract: Round 2. A proof run now counts as passed only when its go test -json output holds a pass event for the proving test (a package run: the package pass and at least one passed test); exit code 0 alone is recorded as not run. The Prover's runs, the sample replay and the fixer share the judgement (C1).
  A test in internal/pipeline shows the guardrail covers the Fixer (C2). readRegular now refuses a link on any parent segment (C3).
  The fix scenario's recorded runs were regenerated with -json. UNRUN: a real container run; the six listener-refused packages.
- last commit: 6e2d6bc [a-fix-is-proved-by-the-same-test] merge: absorb main
- done: C1, C2 and C3 of the review are closed, each with a test that goes red when only its guarded piece is removed.
  - C1 reproduction and negative control: with the pass check forced to accept (`if true || ranAndPassed(...)` in judgeProof) TestARunPassesOnlyWhenItsOutputShowsTheTestPassed (8 cases red), TestAnEarlyExitWithCodeZeroIsNotAPassInTheProofRuns ("recorded as FixRun{Label:change+patch ... Result:passed, ExitCode:0, Output:''}") and TestTheProversBaseRunNeedsTheTestsPassEvent ("passes_on_base=true: a base run with no pass event passed") fail; restored, green
  - C1 end to end: TestARunThatShowedNoPassIsNotAProvedFix (internal/agent/fixer) feeds the sample runner recorded runs marked passed with no pass line and the fix is not proved, keeps no edits; only the output field changes between its cases
  - C1 sample: TestARecordedPassNeedsAPassLineToBeReplayed
  - C2: TestTheGuardrailCoversTheFixer; negative control, `p.fix(ctx, st, raw, ...)` in place of `gen` turned it red for both the prompt-refused and answer-refused cases (proved fix shown, 3 proofs ran, 3 model calls); restored, green
  - C3 was done, not re-worded: readRegular does Lstat on every segment and refuses a link or a non-directory parent; TestAnEditCannotReachThroughALinkedParentDirectory; negative control, skipping the check on non-final segments turned it red (RefuseEdits returned empty); restored, green
- next / blocked-on: master: run the real container proof on a seat with Docker (UNRUN here), run the six listener-refused packages, and arrange the independent second review (C4).
- artifacts:
  - internal/testrun/evidence.go
  - internal/testrun/evidence_test.go
  - internal/testrun/container.go
  - internal/testrun/fix.go
  - internal/testrun/sample.go
  - internal/agent/fixer/fixer_test.go
  - internal/pipeline/fix_test.go
  - testdata/fixture/fix/fixture.json
  - docs/SAMPLE_MODE.md
- evidence:
  - `go build ./...` exit 0; `go vet ./...` exit 0; `gofmt -l internal cmd` prints nothing. measured 2026-10-24
  - `go test -count=1 ./...` exit 1: ok in every package except the six that panic in `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted` and are UNRUN here: cmd/livecheck (TestAllServicesReady), internal/generator (TestLiveReadsTheProvisionedStream), internal/guardrail (TestServerPostsTheMeasuredRequest), internal/pipeline (TestUnsafePromptStopsTheStepAndNamesCategories), internal/query (TestLiveSelectReadsRowsFromTheSQLEditor), internal/review (TestLiveReviewCarriesTheRoutedModel). internal/testrun, internal/agent/fixer, cmd/proven and internal/cli are ok
  - internal/pipeline run on its own for the fix tests: `go test -count=1 -v -run 'TestEachProvedFindingGetsItsFixAttemptsInOrder|TestFixesOffAsksNeitherTheModelNorTheRunner|TestFixAttemptsBoundsTheModelCalls|TestADroppedFindingGetsNoFix|TestAFixIsNeverWrittenToTheChange|TestAFixIsRedactedLikeTheRestOfTheFinding|TestTheGuardrailCoversTheFixer|TestTraceCarriesEvidenceAndRefutation' ./internal/pipeline` exit 0, all PASS. The other tests of that package after the listener panic did not run
  - recorded sample runs: the fix scenario's `recorded_fixes` and the `command` of its `recorded_tests` were regenerated (real go test -json output, condensed to the text its events carry); the discount and refuted fixtures' `recorded_tests` carry a Prover summary and booleans and were not regenerated, so their displayed command still lacks -json
  - after each commit git printed `packed-refs.lock: Operation not permitted`; `git log -2` shows both commits (3950a35, 1ffe358)
- note: master-was-wrong — the key `master-was-wrong:` is not in the closed key set, so it is written here. (1) The order says to record `passed` on a pass event "or, for a package run, the package's pass event and at least one test": that held, with one addition that was needed: a package-level pass event alone does not show a test ran, so the package judgement also requires a test-level pass event, and `go test -json` emits the package pass event even when no test ran. (2) The review says the Prover's runs share the exit-code trust; they did, through the same `sideExit`, so one change there serves the Prover, RunFix and the sample, and no second judge was needed in the fixer. (3) Not refuted: the early-exit probe — I could not run a real container, so the reproduction is on a fake runtime that returns exit 0 with no events; the reviewer's measurement of `go test` printing `ok` for an init-time os.Exit(0) is unrepeated here.
  rulings: (a) -json rather than -v, because the go tool encodes program output into output events, so a program cannot forge a pass event by printing a `--- PASS:` line — cost if wrong: none seen; -v would be forgeable. (b) a run that exits 0 without evidence is `could not run`, exit code kept as 0 — cost if wrong: a reader sees an exit 0 beside a not-run result, which the output line explains. (c) a failed run is still decided by the exit code alone — a patch that exits 1 early reads as failed on the change, which could prove a bogus finding; the order asked only for the passed side; cost if wrong: a bogus failure counts as the change failing. (d) the displayed command now carries -json — cost if wrong: a reader who pastes it gets JSON events. (e) the Prover's recorded outcomes in sample fixtures keep their recorded booleans; only fix runs are re-judged in the sample.
  NOTICED, not acted on: untracked `logs/`, `███████████████████████████` (an emptied `package main` file I could not remove, staged by a broad `git add` and unstaged again, not committed), `████████████████` with the recorder scripts, which now also hold refresh_fixture.py.
- updated: 2026-10-24T06:40:00Z

CLOSE-OK · a-fix-is-proved-by-the-same-test · commit 6e2d6bc · branch fix/a-fix-is-proved-by-the-same-test · worktree ████████████████████████████████████████████████████████████████████████████████████████

tip_moved_by_hand: true
moved_by: ██████████████
supersedes_commit: 1ffe3581e81e2d8490b15f27c16228b5a1c43e96

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| file and line in Saphan Studio's private code | 1 |
| seat identity | 1 |
