# STATUS — the-prover-in-a-container · round 5

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: c91f8282d456402e9defc54526618e914f31be82 [the-prover-in-a-container] testrun: bound removal after a timeout
- updated: 2026-10-22T04:22:12Z

## Summary

A timed-out run returns within its limit plus 3 seconds, and a failed removal names the container that may remain. The commits since the base carry no co-author trailer.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-prover-in-a-container
- state: stop-2-ready
- abstract: A timed-out run returns within its limit plus 3 seconds, and a failed removal names the container that may remain. The commits since the base carry no co-author trailer.
- last commit: c91f8282d456402e9defc54526618e914f31be82 [the-prover-in-a-container] testrun: bound removal after a timeout
- done: A timed-out run stays inside its limit, and a failed removal is named.
  - go build ./... exited 0
  - go vet ./... exited 0
  - go test -count=1 ./... exited 0
  - TestRunReturnsWithinTenSecondsWhenTheRuntimeHangs passed in 6.00s, elapsed 6.003s
  - git log --format=%B e25d5c51d..HEAD | grep -ci co-authored printed 0
- next / blocked-on: none
- artifacts:
  - AGENTS.md
  - internal/testrun/container.go
  - internal/testrun/container_test.go
- evidence:
  - baseline before the edit: go test -count=1 ./... exit 0. testrun 0.046s. No package failed.
  - RED before the fix: TestRunReturnsWithinTenSecondsWhenTheRuntimeHangs failed, elapsed 22.019s, calls run 1s, rm 10.009s, run 1s, rm 10.009s. TestRunNamesTheContainerWhenRemovalFails failed: the output did not name proven-798967-3.
  - RED after reverting only the removal fix: the same two tests failed again, elapsed 22.02s, calls run 1s, rm 10.009s, run 1s, rm 10.01s, and the output did not name proven-800331-3.
  - GREEN after the fix: go test -count=1 -v -run TestRunReturnsWithinTenSecondsWhenTheRuntimeHangs ./internal/testrun/ passed in 6.00s, log elapsed 6.003s, calls run 1s, rm 2.001s, run 1s, rm 2.001s. TestRunNamesTheContainerWhenRemovalFails passed in 0.04s.
  - go build ./... exit 0. go vet ./... exit 0. go test -count=1 ./... exit 0. Packages ok: cmd/proven, internal/agent/judge, internal/generator, internal/netguard, internal/pipeline, internal/query, internal/review, internal/testrun 6.092s, internal/trace. No package failed. No test skipped in that run.
  - go test -race -count=1 ./internal/testrun/ exit 0, 7.098s.
  - Contract tests from commit 33c0dcf, run and not committed: TestContractW1bT1 through T9 passed. T7 passed in 6.00s. T10 skipped on its own guard, PROVEN_TEST_DOCKER was unset.
  - trailer count: git log --format=%B e25d5c51d..HEAD | grep -ci co-authored printed 0 and exited 1, which is grep's no-match status.
- note: master-was-wrong: none. Measured 2026-10-22 before the fix, Timeout 1s, fake sleeps 30s until its context ends: Run held 22.019s because each removal used its own 10s limit (run 1s, rm 10.009s, twice). The review's demand to revert AGENTS.md is not followed: this order says the round-1 section stays. The margin sentence was added there.
  - ADDRESSED: cleanup bound. Removal's own limit is 2s, inside a margin of 3s. One side returns within Timeout plus 3s. T7-equal test 6.003s for both sides. Contract T7 6.00s.
  - ADDRESSED: a failed rm is in the outcome, with the container name and the error. TestRunNamesTheContainerWhenRemovalFails.
  - ADDRESSED: the co-author trailer. git commit and git commit --amend both re-added Co-authored-by. A commit object built from the message file did not. Both commits on e25d5c51d..HEAD are clean. Count 0.
  - NOT ADDRESSED as a revert: AGENTS.md. Review said revert it. This order says keep the section the round-1 order required. It stays, with the margin sentence added.
  - commands: go test -count=1 ./... baseline exit 0; reproduction exit 1; gofmt -w exit 0; tests after the fix exit 0; revert-the-fix exit 1; go build ./... exit 0; go vet ./... exit 0; go test -count=1 ./... exit 0; go test -race -count=1 ./internal/testrun/ exit 0; contract suite exit 0 with T10 skipped.
  - session recording: none. This leg was not given a recorder (ADR-0024).
  - testing bar: no metric is declared, so the bar is UNRUN.
  - toolchain: go.mod says go 1.26; local go version is go1.26.8; no Makefile; no CI workflow file was found.
  - ruling: the T7 fake returns when its context ends, which is what CommandContext does to a hang. A fake that ignores the context cannot return inside any bound. Cost if wrong: a runtime that keeps running after its client is killed still holds the caller.
  - ruling: the removal bound is a fixed 2s inside a fixed 3s margin, not a new setting. Cost if wrong: a runtime that needs longer than 2s to delete a container is reported as still there.
  - ruling: removal uses its own short context, so a timed-out run is still removed. Cost if wrong: a caller that cancels during removal waits up to 2s.
  - ruling: contract T10 was not given PROVEN_TEST_DOCKER=1, so the real-container dial stayed skipped. Cost if wrong: a network path the unit tests do not exercise is still open.
- updated: 2026-10-22T04:19:02Z

CLOSE-OK · the-prover-in-a-container · commit c91f8282d456402e9defc54526618e914f31be82 · branch fix/the-prover-in-a-container · worktree █████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
