# STATUS — cli-escapes-output-strips-git-env · round 3

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 4 min
- last commit: 3f96bf094172ce61a7a138ab837f2c973b341e68 [cli-escapes-output-strips-git-env] cli: stderr shows control bytes
- updated: 2026-10-24T04:02:46Z

## Summary

Round 3 is one commit, 3f96bf0, on top of 40a9570. Every stderr print site in internal/cli/run.go now goes through printErr, which passes the sentence through the same visible encoder the text output uses, so a file name from the change shows its control bytes as \xNN on stderr. The test drives Run into the symlink error and the over-bound file error with a change file named x\x1b[2Jy; it was red before the fix and is red again with visible removed. go vet and gofmt are clean, internal/cli and internal/change are green, and six packages fail only on this sandbox refusing a loopback bind.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — cli-escapes-output-strips-git-env
- state: stop-2-ready
- abstract: Round 3 is one commit, 3f96bf0, on top of 40a9570. Every stderr print site in internal/cli/run.go now goes through printErr, which passes the sentence through the same visible encoder the text output uses, so a file name from the change shows its control bytes as \xNN on stderr. The test drives Run into the symlink error and the over-bound file error with a change file named x\x1b[2Jy; it was red before the fix and is red again with visible removed. go vet and gofmt are clean, internal/cli and internal/change are green, and six packages fail only on this sandbox refusing a loopback bind.
- last commit: 3f96bf094172ce61a7a138ab837f2c973b341e68 [cli-escapes-output-strips-git-env] cli: stderr shows control bytes
- done: internal/cli/run.go and internal/cli/stderr_escape_test.go changed; text writer, git environment, internal/review, internal/change and AGENTS.md untouched
  - reproduced before any change: TestStderrCarriesNoRawControlByteFromTheChange failed on both cases, symlink case "raw byte 0x1b at offset 222 of stderr" ("change: symlink …/change/x\x1b[2Jy"), over-bound case "raw byte 0x1b at offset 18 of stderr" ("change: the file x\x1b[2Jy is larger than the 2048-byte bound …"); instrument: the test's own bytes.IndexByte on the stderr buffer.
  - one encoder: printErr (run.go) is the only error print and calls visible; the verbose target/baseline/head line also goes through visible. Instrument: grep of stderr in internal/cli non-test files shows no other print site.
  - red with visible removed from printErr: both cases fail again, "raw byte 0x1b at offset 223" and "at offset 18"; run.go restored by git checkout.
- evidence:
  - git log --oneline 40a9570..HEAD: 3f96bf0 [cli-escapes-output-strips-git-env] cli: stderr shows control bytes
  - git diff --stat 40a9570..HEAD -- AGENTS.md internal/review internal/change: empty
  - gofmt -l .: empty. go vet ./...: exit 0.
  - go test -count=1 ./...: internal/cli ok 35.972s, internal/change ok 36.573s, internal/childenv ok 3.905s, 30 packages ok in total, 6 FAIL (cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review), each "httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted", a sandbox refusal; those six are UNRUN for this change.
  - master-was-wrong: — (one measured difference: the order's case (b) cannot be driven through Run with PROVEN_MAX_FILE_BYTES, because sample mode keeps default bounds; the test sets the bound with change.UseLimits instead)
- note: an untracked █████████████████████████████ left over from round 2 is not committed and was not removable by this leg (rm is outside its class); it is not part of any commit.
- next: none
- updated: 2026-10-24T04:10:00Z

CLOSE-OK · cli-escapes-output-strips-git-env · commit 3f96bf0 · branch fix/cli-escapes-output-strips-git-env · worktree █████████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
