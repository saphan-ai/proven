# STATUS — cli-escapes-output-strips-git-env · round 4

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: e8c9b28 [cli-escapes-output-strips-git-env] cli: fix the JSON encoder comment
- updated: 2026-10-24T06:40:03Z

## Summary

Round 4. The over-bound stderr case now reaches its own error on 45c805e. Sample mode resets the file bound to the default, so the case uses a sparse file one byte over DefaultLimits().File. Main's fixture check is untouched. The test is red with `visible` removed at run.go:117 (both cases). One commit on 45c805e; internal/review, internal/change, internal/fixture and AGENTS.md unchanged. Six packages fail only on the sandbox listener refusal.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — cli-escapes-output-strips-git-env
- state: stop-2-ready
- abstract: Round 4. The over-bound stderr case now reaches its own error on 45c805e. Sample mode resets the file bound to the default, so the case uses a sparse file one byte over DefaultLimits().File. Main's fixture check is untouched. The test is red with `visible` removed at run.go:117 (both cases). One commit on 45c805e; internal/review, internal/change, internal/fixture and AGENTS.md unchanged. Six packages fail only on the sandbox listener refusal.
- last commit: e8c9b28 [cli-escapes-output-strips-git-env] cli: fix the JSON encoder comment
- done: the over-bound case reaches its own error and the test is green; it is red without `visible` at the stderr print site
  - MASTER (2026-10-24): e8c9b28 corrects the comment at format.go:247 (claude review r6 C1, comment only); merged with main d510ac5: vet clean, 40 packages ok, 0 FAIL.
  - MASTER (2026-10-24): commit subjects brought to the git rules (the master's absorb subject rewritten; every tree unchanged, git diff empty); tip a78d7d8.
  - `go test -count=1 -run TestStderrCarries -v ./internal/cli` — both subtests PASS (exit 0)
  - mutation `fmt.Fprintln(stderr, err.Error())` at internal/cli/run.go:117 — both subtests FAIL (red pasted below), run.go restored with git checkout
- next / blocked-on: none
- artifacts:
  - internal/cli/stderr_escape_test.go
- evidence:
  - `git log --oneline 45c805e..HEAD` printed `1bd720a [cli-escapes-output-strips-git-env] cli: stderr test reaches its error`
  - `git diff --stat 45c805e..HEAD -- AGENTS.md internal/review internal/change internal/fixture` printed nothing
  - `go vet ./...` exit 0; `gofmt -l .` printed nothing
  - `go test -count=1 ./...`: ok in 34 packages incl. internal/cli, internal/change, internal/fixture, cmd/proven; FAIL in cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review. 6 `--- FAIL` lines, each `httptest: failed to listen on a port: listen tcp6 [::1]:0: bind: operation not permitted`, a sandbox refusal; none of those packages is in the diff. Not green here; the master runs the full suite.
  - the diagnosis of the failure on 45c805e: main's change is not what stopped the test. Review's Open puts DefaultLimits back in sample mode (internal/review/open.go:53), so the test's UseLimits(File: 2048) was discarded and the 4096-byte file was read under the 8 MiB bound. With the hash rebound by hand the case exited 0, stderr empty (measured).
  - red with `visible` removed at run.go:117:
    - symlink_named_in_the_error: stderr_escape_test.go:60 stderr does not show the name with its control byte shown: "change: symlink …/fixture/change/x\x1b[2Jy\n"; :63 raw byte 0x1b at offset 222
    - over-bound_file_named_in_the_error: stderr_escape_test.go:60 stderr does not show the name with its control byte shown: "change: the file x\x1b[2Jy is larger than the 8388608-byte bound (setting PROVEN_MAX_FILE_BYTES)\n"; :63 raw byte 0x1b at offset 18
  - commit note: the first `git commit` printed `Unable to create '.../sources/proven/.git/packed-refs.lock': Operation not permitted` after the commit object landed; HEAD is 1bd720a and `git log` shows it
  - not mine, not committed: untracked █████████████████████████████ was present at the start
- note: master-was-wrong — the order says the test "stops earlier than the error it means to provoke" because main checks the planner prompt hash first. Measured: the hash check was reached only because sample mode discards the test's lowered bound (internal/review/open.go:53), so the file was never over-bound; rebinding the hash alone gives exit 0 and empty stderr. The fix is the file size, not the fixture hash.
- updated: 2026-10-24T05:10:00Z

CLOSE-OK · cli-escapes-output-strips-git-env · commit 1bd720a19dd0f926fa5c3d33ea882243d531ae74 · branch fix/cli-escapes-output-strips-git-env · worktree █████████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
