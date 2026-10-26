# STATUS — cli-escapes-output-strips-git-env · round 2

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 6 min
- last commit: 40a957022529a5bbbd2a41708ab8571ce5a807d8 [cli-escapes-output-strips-git-env] cli: field lines stay marked
- updated: 2026-10-24T02:38:39Z

## Summary

Round 2 is one commit, 40a9570, on top of 1e77598. A continuation line of a multi-line field is printed behind the constant prefix "    | " (four spaces and a bar), which no line Proven writes starts with. A tab is escaped like the rest of C0, and a malformed UTF-8 byte is shown as \xNN of the original byte. The escape test fills every printed field with raw bytes. go vet and gofmt are clean and internal/cli and internal/childenv are green; six other packages fail because this sandbox refuses a loopback bind, which is unrelated to this change.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — cli-escapes-output-strips-git-env
- state: stop-2-ready
- abstract: Round 2 is one commit, 40a9570, on top of 1e77598. A continuation line of a multi-line field is printed behind the constant prefix "    | " (four spaces and a bar), which no line Proven writes starts with. A tab is escaped like the rest of C0, and a malformed UTF-8 byte is shown as \xNN of the original byte. The escape test fills every printed field with raw bytes. go vet and gofmt are clean and internal/cli and internal/childenv are green; six other packages fail because this sandbox refuses a loopback bind, which is unrelated to this change.
- last commit: 40a957022529a5bbbd2a41708ab8571ce5a807d8 [cli-escapes-output-strips-git-env] cli: field lines stay marked
- done: format.go and terminal_escape_test.go changed, the git-environment half untouched, internal/review and AGENTS.md unchanged
  - reproduced before any change: visibleLines("first\nfailed services: none\nlast") printed "first\nfailed services: none\nlast" (the forged line at column 0), and visible("a\tb\xffc") printed "a\tb" + U+FFFD + "c" (tab raw, byte lost). Instrument: a scratch test with t.Logf("%q").
  - no forged lines: continuationPrefix = "    | ". Every Proven line starts at column 0 or with exactly two spaces, and none with four spaces and a bar, so a prefixed line cannot be one of Proven's own. Test TestMultiLineFieldCannotForgeProvenLine, with a real failed service in the result.
  - whole byte range: tab is shown as \x09, and a malformed byte as its original \xNN, by iterating with utf8.DecodeRuneInString and treating RuneError with size 1 as one malformed byte. Test TestControlContractBytes, plus tab and 0xff in every field of the coverage test.
  - coverage: TestTextOutputShowsEveryPrintedFieldWithoutControlBytes gives each of 25 printed fields its own name plus raw bytes. The call sites are Class, ClassRule, SettingsStatement, AuthorModelStatement, ReviewerStatements, FailedServices (service and reason), finding ID, Severity, Title, EPSS phrase, Location.Path, Message, RuleID, test Path, Name, Command and Output, query Engine, Snapshot, SQL, Columns and Rows, the family sentence from the trace, and TracePath.
  - the byte contract is written once, in the doc comment of visible, with the ADR-0027 reference
- evidence:
  - git log --oneline 1e77598..HEAD: 40a9570 [cli-escapes-output-strips-git-env] cli: field lines stay marked
  - git diff --stat 1e77598..HEAD -- AGENTS.md internal/review: empty (0 lines)
  - RED prefix removed (continuationPrefix set to ""): TestMultiLineFieldCannotForgeProvenLine fails: "1 lines start at column 0 with "failed services:", want 0", "2 lines start at column 0 with "finding ", want 1" (same for "review class:" and "trace:")
  - RED tab kept (isControl excludes 0x09): TestTextOutputShows... fails for every field (field Class is not printed with its control bytes shown), and TestControlContractBytes fails
  - RED malformed byte replaced (RuneError case removed): the same two tests fail
  - RED one call site (visible removed from finding.Title): TestTextOutputShows... fails with "field Title is not printed with its control bytes shown" and "raw byte 0x1b at offset 314"
  - the reds were run by ██████████████████████████████████████████████████████, which restores format.go each time (cmp reports identical)
  - gofmt -l: empty. go vet ./...: exit 0.
  - go test -count=1 ./... under umask 022: 31 packages ok, 5 with no test files, 6 FAIL. The 6 are cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query and internal/review (live tests). Each panics with "httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted", which is this sandbox refusing a loopback listener. None touches the changed files, and I did not run them on the base for comparison, so that is UNRUN. internal/cli ok (42s) and internal/childenv ok.
  - git commit printed "Unable to create packed-refs.lock: Operation not permitted" but HEAD and the branch both read 40a957022529a5bbbd2a41708ab8571ce5a807d8
- note: master-was-wrong: — (the order's claims held; one added finding: JSON turns a malformed byte into U+FFFD, so the family-sentence field cannot carry 0xff and the test says so). An untracked scratch file █████████████████████████████ (my step-1 reproduction, logs only, not committed) remains because rm is outside this role's grant; the master removes it at teardown.
- updated: 2026-10-24T02:55:00Z
- artifacts:
  - internal/cli/format.go
  - internal/cli/terminal_escape_test.go

CLOSE-OK · cli-escapes-output-strips-git-env · commit 40a957022529a5bbbd2a41708ab8571ce5a807d8 · branch fix/cli-escapes-output-strips-git-env · worktree █████████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| file and line in Saphan Studio's private code | 1 |
