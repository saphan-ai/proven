# STATUS — the-reviewer-model-and-the-review-class · round 3

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: cbae71c532ddb31f07931f9f1f41d351bbeedfb8 [the-reviewer-model-and-the-review-class] docs: list setting defaults
- updated: 2026-10-22T16:02:29Z

## Summary

Round 2 closes B1 and O1 to O6. A symlinked author record is refused, and a change's proven.yaml does not choose how strictly it is reviewed.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-reviewer-model-and-the-review-class
- state: stop-2-ready
- abstract: Round 2 closes B1 and O1 to O6. A symlinked author record is refused, and a change's proven.yaml does not choose how strictly it is reviewed.
- last commit: cbae71c532ddb31f07931f9f1f41d351bbeedfb8 [the-reviewer-model-and-the-review-class] docs: list setting defaults
- done: B1 and O1 to O6 are closed.
  - B1: TestSymlinkedRecordOutsideTheTargetIsRefused and TestSymlinkedRecordGlobIsRefused failed at 2d373c2 with "symlink was read: author Outside:outside-model family family unknown from record", and pass after the fix.
  - O1: TestChangeCannotChooseItsOwnStrictness failed before the fix with "class docs rule **; the change's proven.yaml was applied", and passes after with class security, doubled, reviewer from-base, and "the change edits proven.yaml".
  - O2: docs/SETTINGS.md lists every setting, its default, what it does, and its allowed values. AGENTS.md was not edited.
  - O3: TestEachSettingAtDefaultAndOneOtherValue passes, one subtest per key, including fallback_class, author_commit_trailers off, default_reviewer, doubled_classes as a list, and author_record as a glob, an absolute path, and a path with "..".
  - O4: TestUnreadableTrailerIsNamed and TestUnreadableDescriptionIsNotCalledUnstated pass. A recognised key that is not agent:model prints "author statement unreadable: <line>" and does not say "not stated".
  - O5: TestOpenFamilyNameIsAccepted, TestOpenFamilyNameIsAcceptedAndUnknownClassIsRefused, and TestScanOpenFamilyNameIsAccepted pass. gemini-*=Google is accepted. A malformed family and an unknown class still exit 7.
  - O6: TestSecurityIsDoubledAndBackendIsNot asserts the statements applyDecision produced for a security file and for a backend file.
  - suite: go build ./... && go vet ./... && go test -count=1 ./... exit 0 on go1.26.8. go test -count=1 -race ./internal/review/ exit 0.
- next / blocked-on: none
- artifacts:
  - internal/review/decide.go
  - internal/review/decision_test.go
  - internal/review/config.go
  - internal/review/open.go
  - internal/review/review.go
  - internal/review/decide_test.go
  - internal/review/settings_table_test.go
  - internal/change/between.go
  - cmd/proven/main.go
  - cmd/proven/main_test.go
  - docs/SETTINGS.md
- evidence:
  - B1 red at 2d373c2, exit 1: symlink was read: author Outside:outside-model family family unknown from record (literal and glob).
  - O1 red before the load change, exit 1: class docs rule **; the change's proven.yaml was applied.
  - Negative control (a), symlink check removed, exit 1: symlink was read: author Outside:outside-model family family unknown from record. Restored. Both symlink tests then exit 0.
  - Negative control (b), settings read from the change, exit 1: class docs rule **; the change's proven.yaml was applied. Restored. TestChangeCannotChooseItsOwnStrictness then exit 0.
  - go build ./... && go vet ./... && go test -count=1 ./... exit 0. Every package with tests reported ok. gofmt -l on the changed trees printed nothing.
  - go test -count=1 -race ./internal/review/ exit 0.
  - go run of the package from a scratch directory with no go.mod: exit 1, go.mod file not found. go run -C of the module from that directory: exit 2, hmem state directory inside the repository, because -C makes the process cwd the repository. TestScanSamplePrintsF1AndF2NotF3, which runs the same scan after changing to a temporary directory, is green in the suite.
- note: master-was-wrong: none. B1 reproduced at 2d373c2 before the fix. When the review names a baseline, proven.yaml is read from that commit with git show; the whole file is used, including timeout, so a change cannot supply a key the strictness list forgot. Cost if wrong: a change's timeout is ignored whenever a baseline is set. An explicit config path still wins. No environment variable sets these keys. A review with no baseline has no base tree, so the target file stays the source; that is the sample fixture. A symlink is refused by name even when it points inside the target. Session recording (ADR-0024): none; no recording was provisioned into this worktree. logs/trace.jsonl is an empty untracked file left by the go run -C attempt; rm is not in this role's command list, so it remains. No declared testing-bar metric. Go is go1.26.8, matching the go 1.26 line. No Co-authored-by trailer on the commits.
- updated: 2026-10-22T14:30:16Z

CLOSE-OK · the-reviewer-model-and-the-review-class · commit dd3e5f31cf404d8d1a3d217e9825b6cda0102677 · branch fix/the-reviewer-model-and-the-review-class · worktree ███████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
