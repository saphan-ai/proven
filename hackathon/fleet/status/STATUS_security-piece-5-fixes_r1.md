# STATUS — security-piece-5-fixes · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: d3b4c24 [security-piece-5-fixes] fixture: fix scenario prompts as data
- updated: 2026-10-24T07:09:30Z

## Summary

Items 1-5 are fixed in five commits, each with a test that failed at the base and passes now. Evidence rows must be bound to the change, a bare refuted flag is not stored, change text enters prompts as JSON data, a disabled guardrail is named in the result, the stores get the guarded generator. Build and vet are green; listener-refused tests are UNRUN (named below).

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — security-piece-5-fixes
- state: stop-2-ready
- abstract: Items 1-5 are fixed in five commits, each with a test that failed at the base and passes now. Evidence rows must be bound to the change, a bare refuted flag is not stored, change text enters prompts as JSON data, a disabled guardrail is named in the result, the stores get the guarded generator. Build and vet are green; listener-refused tests are UNRUN (named below).
- last commit: d3b4c24 [security-piece-5-fixes] fixture: fix scenario prompts as data
- note: master-was-wrong: nothing; every reviewer test failed at the base (exit 1) for the stated reason.
- artifacts:
- done: five items fixed, test red before and green after for each
  - item 1 evidence bound to the change: TestQueryRowsUnrelatedToTheChangeAreNotEvidence, TestDropsQueryRowsThatOnlyRepeatTheFindingsWords; before exit 1, after exit 0; reverted line: the Bound check in judge counts back to queryRelates
  - item 2 refutation carries evidence: TestSkepticWordAloneDoesNotSuppressTheNextReview; before exit 1, after exit 0; reverted line: the "&& evidenced" condition in world apply. Per your ruling TestRefutedFindingIsAbsentOnTheNextReview and TestNegativeControl_ClearHMEMStore were rewritten to refute with a test that passes on the change
  - item 3 change text is data: TestTwoDifferentChangesNeverShareOnePrompt, TestChangeContentCannotCreateASecondFileFrame, TestTitleWithLineBreakAddsNoLineToThePrompt in investigator, prover, skeptic; before exit 1, after exit 0; reverted line: the JSON escaping in answer Encode. Recorded sample prompts and planner hashes regenerated for five fixtures and demo/live. Task text was not moved to a system role (live.go untouched)
  - item 4 guardrail off is named: TestGuardOffIsNamedInTheDraft, TestResultCarriesTheGuardrailOffStatement, TestTextSaysWhenTheGuardrailWasOff; before exit 1, after exit 0; reverted line: the guardOff block in Pipeline Review. JSON and SARIF have no writer in this build, only text
  - item 5 stores get the guarded generator: TestStoresHoldTheGuardedGenerator; before exit 1, after exit 0; reverted line: SetGenerator(gen) changed to SetGenerator(raw)
- evidence:
  - go build ./... and go vet ./... green
  - go test per package: all green except listener-refused tests, which were skipped and are UNRUN
  - three tests not named in the ruling encoded the old word-match behaviour and were changed: TestNegativeControl_F3RecordedQueryRow, TestNegativeControl_PackageTheAdvisoryDoesNotName, TestCVEWithNoEPSSRowAttachesNothing
  - UNRUN pipeline (listener refused, httptest failed to listen on a port: bind: operation not permitted): TestUnsafePromptStopsTheStepAndNamesCategories TestUnsafeAnswerIsNotUsed TestGuardrailOffCompletesAndSaysNoGuardrail TestClosedGuardrailRefusesTheStepByName TestSafeGuardrailKeepsTheSampleFindings
  - UNRUN review (same refusal): TestLiveReviewCarriesTheRoutedModel TestLiveReviewRefusesANonNVIDIAModelByName TestUnauthorizedReviewDegrades TestClosedEndpointDegradesTheReview TestLiveModeWithoutRuntimeNamesProverOff TestRuntimeTokenEchoedInTheReceiptIsNotWritten TestGuardrailKeyEchoedInTheCategoriesIsNotWritten TestOrgIDEchoedInTheReceiptIsNotWritten TestZetarisUnauthorizedReviewCompletes TestZetarisClosedEndpointDegradesTheReview
  - UNRUN generator, guardrail, query and cmd/livecheck: every httptest test in each package, same refusal; the master runs them
- next / blocked-on: the master runs the UNRUN tests; delete internal/review/git_live_test.go.bak (rm is denied to this role); NOT DONE by order: sample guardrail trace wording, a Prover test binding to the claim. QueryContradicts still matches by word, outside this order
- updated: 2026-10-24T07:10:00Z

tip_moved_by_hand: true
moved_by: ██████████████
supersedes_commit: 5cd6e693b31f0d9964a931bbf78839da6f9df8f8

## REDACTED

| what | times |
|---|---|
| seat identity | 1 |
