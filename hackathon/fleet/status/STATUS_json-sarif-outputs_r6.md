# STATUS — json-sarif-outputs · round 6

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 18 min
- last commit: 1bc7db1a40a7585fef0b3d6996166749f9d3ae95 [json-sarif-outputs] docs: letter-case outputs fail at the write
- updated: 2026-10-24T08:49:36Z

## Summary

A path through an absent directory is not the same file as the collapsed name. missing/../report.json fails at the write with exit 2. A real alias still exits 7.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — json-sarif-outputs
- state: stop-2-ready
- abstract: A path through an absent directory is not the same file as the collapsed name. missing/../report.json fails at the write with exit 2. A real alias still exits 7.
- last commit: 1bc7db1a40a7585fef0b3d6996166749f9d3ae95 [json-sarif-outputs] docs: letter-case outputs fail at the write
- done: A later ".." after a missing component is unreachable, so the load check does not call it the same file.
  - TestMissingDirectoryThenDotDotIsUnwritable absent exited 2 and present exited 7 (`go test -count=1 -run TestMissingDirectoryThenDotDotIsUnwritable ./internal/cli`, exit 0, including once on the committed tree)
  - With the missing-component return removed, the absent subtest exited 7 and the present subtest passed (`go test -count=1 -v ./internal/cli/ -run TestMissingDirectoryThenDotDotIsUnwritable`, exit 1). The return was restored.
  - Every alias-table row still exits 7 (`go test -count=1 -run TestAliasTableIsRefusedAtLoad ./internal/cli` and `TestAliasTableIsCaughtByTheLoadCheckNotTheWriteGuard`, exit 0)
  - `go vet ./...` exit 0. `go test -count=1 ./...` exit 0. `gofmt -l` on the three edited files printed nothing and exited 0
- next / blocked-on: none
- artifacts:
  - internal/cli/config.go
  - internal/cli/outputs_alias_test.go
  - internal/cli/outputs_test.go
- evidence:
  - git rev-parse HEAD printed e289682156c29dfd049c30b9820138297e179984
  - git log --oneline e446790..HEAD printed one line, e289682 [json-sarif-outputs] cli: a missing directory blocks parent dots
  - git diff --stat e446790..HEAD -- AGENTS.md internal/review printed nothing and exited 0
  - Before any edit, on e446790, go test -count=1 ./... exited 0. The output was 37 lines starting with ok and 5 lines of [no test files]. No FAIL line.
  - Before the fix, go test -count=1 -v ./internal/cli/ -run TestMissingDirectoryThenDotDotIsUnwritable exited 1. The absent subtest printed exit 7, want 2, and stderr said the sarif path was the same file as --json-output. The present subtest passed.
  - After the fix, the same test exited 0. TestCollidingOutputPathsAreRefused, TestOutputWriteFailureExits2, TestDanglingAliasOutputsAreRefused and TestDanglingAliasChainsAreFollowed also exited 0.
  - Unplug of the rule: deleting the return that runs when a ".." follows a missing component made the absent subtest exit 7 again and left the present subtest passing. Restored. The committed test then exited 0.
  - go version printed go1.26.8 linux/amd64. go.mod says go 1.26 and has no toolchain line. No Makefile. No .github workflow in this worktree.
  - The built binary was not run. go build -o ██████████████████████████████████████ ./cmd/proven exited 0. Executing that program was denied, allowed form none. cli.Run is what cmd/proven/main.go calls.
  - BLOCKING outputIdentity false collision: ADDRESSED
  - go test -race UNRUN: this change adds no goroutine and no shared mutable state
  - testing-bar UNRUN: no metric, command or baseline is declared
- note: master-was-wrong: on e446790 the suite printed 37 ok and 5 no-test-files, not 41 ok; 0 FAIL agrees, and the collision claim held (absent pair exited 7 before the fix). ruling: the unreachable identity is the prefix unreachable: plus the raw path, because a resolved identity is absolute — cost if wrong, two different unreachable spellings are not refused at load and the first write exits 2. ruling: TestCollidingOutputPathsAreRefused now creates sub, because the old fixture left sub absent and was this false collision — cost if wrong, that table no longer covers an absent sub. NOTICED: docs/OUTPUT_FORMATS.md still says paths are compared after they are cleaned. NOTICED: when Getwd fails, outputIdentity still returns filepath.Clean, which collapses ".." lexically.
- note: master-commits-after-review: bf4897b (docs line 5, codex r8 condition) and 1bc7db1 (docs letter-case sentence, claude r9 C3, measured on macOS 2026-10-24: second write exit 2, first output kept) are the master's, docs only; e289682..1bc7db1 changes docs/OUTPUT_FORMATS.md and nothing else. master-verify on 1bc7db1 merged with main 656dbcb: build ok, vet ok, internal/cli and internal/cli/sarifschema ok; U11 unplug reddens TestSARIFShapesMatchTheSchema/line_0 (claude r9 C1).
- updated: 2026-10-24T16:05:00Z

remark: the 37-and-5 package split is a hand count of go test lines; a red suite prints FAIL and a non-zero exit, which this run did not.
remark: exit 7 before the fix was measured by cli.Run in the new test, not by the proven binary; the binary was denied and was not executed.
remark: the hand count of packages cannot see a package go test did not list.

CLOSE-OK · json-sarif-outputs · commit 1bc7db1a40a7585fef0b3d6996166749f9d3ae95 · branch fix/json-sarif-outputs · worktree ██████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
