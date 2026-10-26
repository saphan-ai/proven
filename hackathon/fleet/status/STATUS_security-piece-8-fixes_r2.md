# STATUS — security-piece-8-fixes · round 2

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 37de26a [security-piece-8-fixes] tests: secrets meet the length floor
- updated: 2026-10-24T08:20:04Z

## Summary

All six items hold; each reproduction test failed at base fcbcb79 and passes at the tip, and reverting each fix turns it red again.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — security-piece-8-fixes
- state: stop-2-ready
- abstract: All six items hold; each reproduction test failed at base fcbcb79 and passes at the tip, and reverting each fix turns it red again.
  Build, vet and gofmt are clean; every package is green except tests that dial a local listener, which this seat refuses (named in the note).
  Judgement calls for the master: a short org id is left out of the redactor, not refused; commit 1df0382 also carries the not-yet-green tests of other items.
- last commit: 37de26a [security-piece-8-fixes] tests: secrets meet the length floor
- done: items 1-6 landed in six commits on fix/security-piece-8-fixes
  - item 1 state, trace and work area outside the reviewed target (pipeline begin; Build now writes nothing): TestReviewRefusesLogsAndStateInsideTheReviewedTarget, TestReviewRefusesATraceInsideAnEmptyTarget, TestReviewRefusesALogsDirectoryReachedThroughALinkIntoTheTarget; before exit 1, after exit 0; control: `if target != ""` in begin turned off, all three red
  - item 2 links resolved before the inside check (internal/pathguard): TestOpenRefusesAStateDirectoryReachedThroughALinkIntoTheRepository in hmem and in world; before 1, after 0; control: Resolve returning the lexical path, both red
  - item 3 existing wide state file narrowed on Open: TestOpenNarrowsAnExistingWideStateFile in hmem and world; before 1, after 0; control: the Chmod of state.json removed, both red
  - item 4 every configured secret redacted, short ones refused at load: TestTraceOmitsAConfiguredSecretOfAnyLength, TestAConfiguredValueIsRedactedWhateverItsLength, one refusal test each in generator, guardrail, query; before 1, after 0; controls: floor restored in redact.New (trace test red), TooShortSecret threshold zeroed (three refusal tests red)
  - item 5 atomic write sets the mode: TestAtomicWriteDoesNotCarryAWiderTemporaryMode in hmem and world; before 1, after 0; control: removal of an old temporary file disabled, both red
  - item 6 materialized files not group-writable: TestMaterializedTreesAreNotGroupWritableUnderAPermissiveUmask (sets umask 002 itself); before 1, after 0; control: treeFileMode returning the source mode, red
- next / blocked-on: the master runs the listener-refused tests, rules on the org id question, and the CLI train sets a default logs directory outside the target
- artifacts:
  - internal/pathguard/pathguard.go
  - internal/pipeline/memory.go
  - internal/hmem/hmem.go
  - internal/world/world.go
  - internal/redact/redact.go
  - internal/service/secret.go
  - internal/change/archive.go
- note: MASTER-WAS-WRONG: nothing refuted; every reviewer test was red at the base. One guard, TestOpenRefusesAStateDirectoryBehindALinkThatDoesNotResolve, passes at base too (MkdirAll fails through a dead link); it guards the new resolver and is not a reproduction. UNRUN, listener refused (httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted; the base had the same panics), each skipped by name so the rest of its package ran: pipeline 5 tests (TestUnsafePromptStopsTheStepAndNamesCategories, TestUnsafeAnswerIsNotUsed, TestGuardrailOffCompletesAndSaysNoGuardrail, TestClosedGuardrailRefusesTheStepByName, TestSafeGuardrailKeepsTheSampleFindings); generator 17 live tests; guardrail 8 server tests; query 4 (TestLiveSelectReadsRowsFromTheSQLEditor, TestUnauthorizedNamesZetarisAndHidesTheKey, TestClosedEndpointNamesZetarisUnreachable, TestNonSelectIsNotSent); review 10 (TestLiveReviewCarriesTheRoutedModel through TestZetarisClosedEndpointDegradesTheReview); cmd/livecheck 10. JUDGEMENT 1: the org id is an identifier and short numeric ones are accepted (fixtures use "1"); redacting "1" rewrote commit ids in the trace, so redact.FromEnv leaves an org id under 8 bytes out while redact.New has no floor and the three credentials are never excepted; if wrong a short org id can appear in a trace; alternative is to refuse it at load and change the "1" fixtures. JUDGEMENT 2: Build no longer opens the trace or stores, so a bad state directory is refused at Review, not Open; fixtures "tok" and "k" widened to 8+ bytes in generator, guardrail and query tests, and the old-floor redact test replaced by tests of its absence. JUDGEMENT 3: commit 1df0382 also holds the then-red tests of items 1, 2, 3, 5, 6; `git reset --soft` to split it was denied by the git-ref guard and not retried in another spelling. The CLI default TARGET "." with logs under the working directory is now refused at review time (run.go untouched, CLI tests pass). Not done, out of scope: operator-named output modes, the container staging mode, the conformance tool temp directory. Docs: demo/live/README.md updated; the allowed-form text of the three secret settings now says at least 8 characters.
- updated: 2026-10-24T16:00:00Z

CLOSE-OK · security-piece-8-fixes · commit 37de26a · branch fix/security-piece-8-fixes · worktree ██████████████████████████████████████████████████████████████████████████████

tip_moved_by_hand: true
moved_by: ██████████████
supersedes_commit: 7d6ea62d3179c3a0ea75091c1653d861e0919455

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| seat identity | 1 |
