# STATUS — a-fix-is-proved-by-the-same-test · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 446d6ac [a-fix-is-proved-by-the-same-test] merge: absorb main
- updated: 2026-10-24T05:16:30Z

## Summary

A Fixer step after the Judge asks the model for a patch (exact-text edits) for each kept finding a test proved, applies it to a scratch copy only, and shows it as a fix only when the proving test passes on base, fails on change, passes on change+patch and the package's other tests pass; each run is recorded with command and exit code.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-fix-is-proved-by-the-same-test
- state: stop-2-ready
- abstract: A Fixer step after the Judge asks the model for a patch (exact-text edits) for each kept finding a test proved, applies it to a scratch copy only, and shows it as a fix only when the proving test passes on base, fails on change, passes on change+patch and the package's other tests pass; each run is recorded with command and exit code.
  A failed patch is retried by the same model with the failed run in hand (`fix_attempts`, default 2); then the review says "fix not proved" and shows no patch. `fixes: off` turns it all off. A dependency finding names the smallest fixed version with its OSV row.
  Sample scenario `testdata/fixture/fix` has one proved fix (F1) and one failed attempt (F2, two patches); recorded runs are real `go test` output. UNRUN: a real container run (no Docker on this seat) — the master runs it. JSON, SARIF and markdown writers do not exist at the base, so the fix is rendered in text only.
- last commit: 446d6ac [a-fix-is-proved-by-the-same-test] merge: absorb main
- done: ADR-0042 points 1 to 8 are built in the reviewer core and proven in sample mode; the container path is proven only against a fake runtime.
  - the three proving runs plus the package run are recorded with command and exit code: TestRunFixRunsTheProofThreeTimesAndThePackage (internal/testrun), TestAProvedPatchIsShownAsAFix (internal/agent/fixer), TestFixScenarioShowsOneProvedFixAndOneFailedAttempt (cmd/proven)
  - a patch that breaks another test of the package is not shown as a fix: TestAPatchThatBreaksAnotherTestIsNotShownAsAFix, TestRunFixReportsABrokenNeighbourTest, and F2 attempt 2 in the sample scenario
  - retry with the failure in hand, bounded by the setting: TestAFailedPatchIsRetriedWithTheFailureInHand, TestTheAttemptsSettingBoundsTheModelCalls, TestFixAttemptsBoundsTheModelCalls, TestOneAttemptSettingEndsTheFailedFixAtOnce
  - fixes off asks neither the model nor the runner: TestFixesOffAsksNeitherTheModelNorTheRunner, TestFixesOffProposesNoFix, TestFixSettingsDefaultAndOtherValues
  - the fix is never written to the change: TestAFixIsNeverWrittenToTheChange (every file of the fixture before and after), TestRunFixRunsTheProofThreeTimesAndThePackage (the change tree and the scratch copy)
  - a patch cannot edit a test file, leave the checkout, follow a link, or apply half: TestRunFixStartsNothingForARefusedPatch, TestRefuseEditsSeesEarlierEditsOfTheSameFile, TestRefuseEditsDoesNotFollowALink
  - dependency fix is the smallest fixed version no returned range holds, with the advisory row, compared in Go from the existing osv-by-package rows: TestSmallestFix* (internal/osv), TestADependencyFindingNamesTheSmallestFixedVersion
  - a dropped finding gets no fix and a fix is redacted like the rest of the finding: TestADroppedFindingGetsNoFix, TestAFixIsRedactedLikeTheRestOfTheFinding
- next / blocked-on: master: run the real container proof on a seat with Docker (UNRUN here), check the two stray untracked paths under NOTICED, and dispatch the output-format train for JSON, SARIF and markdown fix rendering.
- artifacts:
  - internal/agent/fixer/fixer.go
  - internal/testrun/fix.go
  - internal/osv/fix.go
  - testdata/fixture/fix/fixture.json
  - docs/SAMPLE_MODE.md
- evidence:
  - baseline at base 8770336 plus the branch's d82cac2: `go test -count=1 ./...` exit 1; ok everywhere except 6 packages that panic in `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted` (this seat's sandbox forbids a listener): cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review. measured 2026-10-24
  - after: `go build ./... && go vet ./...` exit 0. `go test -count=1 ./...` exit 1 with the SAME six packages red for the same reason and every other package ok (fixer, testrun, osv, cli, cmd/proven, demo, fixture, agents, change, hmem, world, trace, conformance all ok). measured 2026-10-24
  - with the 44 listener-opening tests skipped by name (`go test -count=1 -skip '<44 names>' ./internal/generator ./internal/guardrail ./internal/pipeline ./internal/query ./internal/review`) exit 0, all five ok; cmd/livecheck still panics on a helper-based listener test and was not touched. measured 2026-10-24
  - negative control (law #27), mutated field `recorded_fixes[F1, attempt 1].runs[change+patch].result` (and its exit_code) in testdata/fixture/fix/fixture.json set to failed: TestFixScenarioShowsOneProvedFixAndOneFailedAttempt went red ("F1 is missing fix proved after 1 attempt"), the output showed "fix not proved after 1 attempt"; restored, green again
  - negative control at the unit level, mutated field the `Result` of the `package on change+patch` run: TestAPatchThatBreaksAnotherTestIsNotShownAsAFix flips between proved and not proved on that one field
  - two more mutations, each restored: the `Dropped` guard in Pipeline.fix disabled turned TestADroppedFindingGetsNoFix red; the `redactFix` call removed turned TestAFixIsRedactedLikeTheRestOfTheFinding red
  - two defects my own tests caught during the work: defaultSettings() lacked `Fixes: true` (so fixes would have been off by default) and my first osv test expectations ignored that a lower fixed bound can still lie inside another row's range; both fixed before commit
  - session recording (ADR-0024): none; this headless leg has no recorded coding session to publish
- note: master-was-wrong — (1) your reading that the review output has JSON, SARIF and markdown is false at this base: `cli.FormatterFor` returns text only and refuses json and sarif, and no markdown writer exists, so the fix is carried in `review.Finding.Fix` and rendered in text only; the other formats are unbuilt, not skipped. (2) the Prover, testrun, proof levels and querytpl readings held; the affected-range lookup needed no new template: it reads the rows the existing osv-by-package template already returns. The status key `master-was-wrong:` is not in the closed key set of this file, so it is written here under `note:`.
  rulings: (a) fix_attempts counts attempts including the first, default 2 (one try and one retry), bounded 1 to 5 — cost if wrong: the owner meant two retries, one more model call and proof per failed fix. (b) fixes are on by default in live mode too, not only in sample mode — cost if wrong: a live review spends extra model calls and container runs the owner did not choose. (c) a patch is a list of exact-text edits shown as - and + lines, not a unified diff, so no patch tool is needed and every path is checked first — cost if wrong: it cannot be fed to `git apply` as printed. (d) a patch may not edit any _test.go file, so it cannot rewrite the proof — cost if wrong: a fix that legitimately needs a test change is never shown. (e) a premise that does not show again (base not passing, change not failing) stops the retries — cost if wrong: one lost retry in a flaky case. (f) the package run is skipped when the patched test already failed — cost if wrong: a failed attempt's record has no package line. (g) in sample mode a finding with no recorded fixer answer shows no fix and no error — cost if wrong: a fixture author can forget the recording and see nothing. (h) the dependency fix is never marked proved, and `fixes: off` also removes it — cost if wrong: a reader may expect a version hint with fixes off. (i) `fixer` added to the generator's agent list so PROVEN_REQUIRE_NVIDIA may name it, and the review seam got additive types although review.go asks both sides to agree — cost if wrong: the shell train reviews an extra optional field. (j) one existing test was edited and justified: config_test.go's expected key list, because the closed key set grew by `fixes` and `fix_attempts`.
  NOTICED, not acted on: an untracked `logs/` at the worktree root, made by my first `go run ./cmd/proven scan` from inside the repository (hmem refused, logs were already written); an untracked `███████████████████████████` I emptied to `package main` because this seat has no rm; and `██████████████████████████████████████████████` with my recorder and fixture scripts. None is committed; the master should remove the first two. The text output prints a multi-line test `output:` with no indent for its later lines (existing behaviour at the base).
  UNRUN: any real container run of RunFix (Docker is not on this seat; `docker` is outside the leg's grant); the 44 listener tests; JSON, SARIF and markdown rendering (no writers exist).
- updated: 2026-10-24T04:30:00Z

CLOSE-OK · a-fix-is-proved-by-the-same-test · commit 446d6ac · branch fix/a-fix-is-proved-by-the-same-test · worktree ████████████████████████████████████████████████████████████████████████████████████████

tip_moved_by_hand: true
moved_by: ██████████████
supersedes_commit: 29544679fdee78cd1e5c468115722d6acd893d02

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| file and line in Saphan Studio's private code | 1 |
| seat identity | 1 |
