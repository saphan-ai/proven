# STATUS — the-owed-security-items · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 8ea6817 state: narrow an existing state directory and file (the master: the review's C1 and C3; C2 ruled skip-and-name; 39 packages ok)
- updated: 2026-10-23T21:03:50Z

## Summary

All 17 items are handled in 6 commits. Fixed with a red test first: 1, 2, 3, 4, 5, 6, 8, 13, 14, 16. Fixed at base, tests added and green: 9, 10, 11, 12 (a to e), 15, 17. Item 7 had no test and the redactor was sound, so only the test was added. Negative controls for 1 (trace), 3, 8 and 17 went red when the fix was undone, and the fixes were restored. `go build` and `go vet` are green. `go test ./...` shows the same 6 red packages as the baseline, all because the sandbox refuses a listener (`httptest`).

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-owed-security-items
- state: stop-2-ready
- abstract: All 17 items are handled in 6 commits. Fixed with a red test first: 1, 2, 3, 4, 5, 6, 8, 13, 14, 16. Fixed at base, tests added and green: 9, 10, 11, 12 (a to e), 15, 17. Item 7 had no test and the redactor was sound, so only the test was added. Negative controls for 1 (trace), 3, 8 and 17 went red when the fix was undone, and the fixes were restored. `go build` and `go vet` are green. `go test ./...` shows the same 6 red packages as the baseline, all because the sandbox refuses a listener (`httptest`).
- last commit: 8ea6817 state: narrow an existing state directory and file (the master: the review's C1 and C3; C2 ruled skip-and-name; 39 packages ok)
- done: the owed security items are in 6 commits on top of 16cf7b7, each behaviour with a test that fails before the fix, or the proof it already holds at base
  - 1 and 2, directories 0700 and files 0600 in world, hmem, trace and cmd/meterless-conform: TestStateIsReadableOnlyByItsUser (world, hmem), TestTraceIsReadableOnlyByItsUser, TestSessionDirectoriesAreReadableOnlyByTheUser. Red before (modes 755 and 644), green after. Commit b81467a.
  - 3, a linked proven.yaml is refused by name without quoting its target, in loadSettingsInTarget: TestLinkedConfigIsRefusedByName red before, TestRegularConfigInTheTargetIsStillRead green. An explicit --config path is left as is, because the operator named it. Commit 9fd1970.
  - 4, the worktree export skips a tracked path whose parent directory is a link and names it as skipped: TestExportWorktreeDoesNotCopyThroughALinkedParent red before. The file is internal/change/archive.go. Commit b5a4c41.
  - 5, a linked base or change is not a prepared tree (internal/pipeline/trees.go, Lstat) and a linked checkout is not a mount source (container.go Run): TestLinkedTreesAreNotAPreparedPair (red when the Lstat is put back to Stat) and TestLinkedCheckoutIsNotMounted, red before. Commit b5a4c41.
  - 6, the run line passes --pull never, and an absent image is the Prover down naming the pull: TestRunLineNeverPullsTheImage and TestAbsentImageIsTheProverDownAndNamesThePull, red before. Commit b5a4c41.
  - 7 and 8, the demo redactor test (TestRedactTextHidesAConfiguredSecret in demo/live/review) and redaction of the %q, %+q and JSON spellings of a secret (internal/redact, spellings): TestAnEscapedFormOfAValueIsRedactedToo red before. The raw-value half of the demo test passed at base, as expected. Commit 7e13e0d.
  - 9, 10 and 17 hold at base: the literal builder already refuses every control byte, DEL, a byte of 0x80 or more, a quote and a backslash. Tests added, green at base: TestPackageNamesRefuseEveryControlByte, TestEveryIdParameterRefusesLineBreaksNulAndAForeignId, TestLiteralByteBackstopRefusesWhatAFormatShouldNeverAdmit. Removing the backstop loop turned the last one red, and I restored it. Commit ec76eb7.
  - 11 holds at base: osv.ValidTable, osv.ValidView and epss.ValidView are one-line wrappers over querytpl.ValidRelation, so one rule exists. TestOneRuleNamesAValidRelation holds the three equal to it. I kept the wrappers, because they are exported names. Commit ec76eb7.
  - 12, tests (a) size-limit refusals: TestLoadLimitsRefusesAWrongValueByName and TestUseLimitsKeepsTheDefaultForAZeroField, (b) the baseline proven.yaml bound: TestBaselineConfigOverTheConfigBoundIsRefused, (c) the record sum: TestAuthorRecordIsBoundedAsASum, (d) Open's wiring: TestOpenAppliesTheSizeSettings, (e) the commit-log bound on its own: TestMessagesRefuseALogOverTheFileBound. All were green at base except the sample-mode half of (d), which is item 14. Commits b6d7450 and 9fd1970.
  - 13, git's failure text now comes from clipped (512 byte) stderr and never from stdout: TestFailureMessageNeverQuotesStandardOutput and TestFailureMessageIsClipped, red before. Commit b6d7450.
  - 14, sample-mode Open puts the default size bounds back: the red assertion is in TestOpenAppliesTheSizeSettings. I took the "or" branch of the review condition. The bounds are still process state, so two live reviews with different settings in one process still share them. Commit 9fd1970.
  - 15, the baseline proven.yaml is read with the config bound: fixed at base (loadSettingsAtBaseline), proved by TestBaselineConfigOverTheConfigBoundIsRefused, green at base.
  - 16, the live gate refuses a double quote and a backtick: TestSelectGateSeesAStatementHiddenBehindAnotherQuoteCharacter red before. The comment in internal/query/live.go and the "The second check" section of docs/QUERIES.md say what the gate reads. Commit ec76eb7.
- next / blocked-on: none
- artifacts:
- note: master-was-wrong (the closed key set has no key of that name): the order named ███████████████████████████ and trees.go, which do not exist. The export is in internal/change/archive.go and the prepared-tree check is in internal/pipeline/trees.go. Items 9, 10, 11, 15 and 17 were already fixed at the base, because ADR-0032's querytpl replaced the builders the reviewers read. The review notes for piece 4 describe quoteLiteral, which no longer exists. P4-F5 is therefore a test addition, and the tests pass at base.
- evidence:
  - baseline at 16cf7b7, go build ./... and go vet ./...: exit 0. go test -count=1 ./...: 6 packages red (cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review), every one a panic "httptest: failed to listen on a port: operation not permitted". The sandbox refuses a listener. A panic ends its package run, so tests after the first listener test were not seen in those packages.
  - tip ec76eb7, go build ./... and go vet ./...: exit 0. gofmt -l: empty. go test -count=1 ./...: 32 packages ok, the same 6 red for the same reason, nothing new red.
  - go test -count=1 -skip on the four listener tests of internal/query: ok. internal/review and internal/pipeline were run with -run on the new tests and the config, record and sample tests, which were green, and the full package could not be read past the first listener panic. This is not a full green for those two packages.
  - Negative controls, fix undone then test red then fix restored: item 1 (trace MkdirAll 0755, TestTraceIsReadableOnlyByItsUser), item 3 (link check disabled, TestLinkedConfigIsRefusedByName), item 8 (spellings returns only the raw value, TestAnEscapedFormOfAValueIsRedactedToo), item 5 (Lstat back to Stat, TestLinkedTreesAreNotAPreparedPair), item 17 (byte loop disabled, TestLiteralByteBackstopRefusesWhatAFormatShouldNeverAdmit).
  - git log 16cf7b7..HEAD carries no Co-authored-by line. The packed-refs.lock "Operation not permitted" message appeared after each commit and each commit landed, checked with git log.
  - session recording: none, this headless leg has no recording tool (ADR-0024).
- updated: 2026-10-24T00:00:00Z

CLOSE-OK · the-owed-security-items · commit ec76eb7 · branch fix/the-owed-security-items · worktree ███████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
