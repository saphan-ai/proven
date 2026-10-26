# STATUS — json-sarif-outputs · round 5

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: ae1a88e [json-sarif-outputs] cli: refuse an output that is stdout's own file
- updated: 2026-10-24T05:03:04Z

## Summary

Standard output on a regular file is now compared (os.SameFile, links followed) with -o and every --*-output at load, before the review: exit 7 naming "standard output" and the flag. The write-time guard counts stdout's file too (exit 2, stdout kept). One commit on f4819cf; AGENTS.md and internal/review untouched.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — json-sarif-outputs
- state: stop-2-ready
- abstract: Standard output on a regular file is now compared (os.SameFile, links followed) with -o and every --*-output at load, before the review: exit 7 naming "standard output" and the flag. The write-time guard counts stdout's file too (exit 2, stdout kept). One commit on f4819cf; AGENTS.md and internal/review untouched.
- last commit: ae1a88e [json-sarif-outputs] cli: refuse an output that is stdout's own file
- done: C1 closed; each claim below names its instrument
  - reproduction before the fix: `go test -run TestStdout ./internal/cli/` was red, 8 rows "exit 0, want 7" (4 flags x direct/symlink) and the write-time test "exit 0, want 2" (captured in ███████████████████████████████████████)
  - after the fix: TestStdoutRedirectedOntoAnOutputIsRefused, TestStdoutOnAnotherFileOrAWriterIsNotACollision, TestStdoutAliasCreatedAfterLoadIsRefusedAtWrite PASS
  - unplug of the write-time stdout entry (`&& false`): TestStdoutAliasCreatedAfterLoadIsRefusedAtWrite red "exit 0, want 2"; restored by git checkout
  - `go test -count=1 ./internal/cli/` ok 34.9s; gofmt -l empty; `git diff --stat f4819cf..HEAD -- AGENTS.md internal/review` empty; `git log --oneline f4819cf..HEAD` is only ae1a88e
- evidence:
  - go vet ./... fails only at internal/cli/sarifschema: module santhosh-tekuri/jsonschema/v5 not fetchable here (tls x509 OSStatus -26276); go vet ./internal/cli/ clean
  - go test ./...: all ok except sarifschema [setup failed, same cause] and six packages panicking on httptest listen (cmd/livecheck, internal/generator, guardrail, pipeline, query, review): sandbox denies bind, outside this diff
  - master-was-wrong: — (the order's claims held; one note: with -o set and stdout on the same file as -o the run is refused too, as the order says "like -o")
  - untracked residue not mine: ███████████████████████ (reviewer's one-line probe) left untouched
- artifacts:
  - internal/cli/config.go
  - internal/cli/run.go
  - internal/cli/outputs_stdout_test.go
- next: none
- updated: 2026-10-24

CLOSE-OK · json-sarif-outputs · commit ae1a88e · branch fix/json-sarif-outputs · worktree ██████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| file and line in Saphan Studio's private code | 1 |
