# STATUS — the-reviewer-honours-the-request-tests · round 2

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 12 min
- last commit: 3e810674fb76136cb8e529bb4980ad08a5d50e63 [the-reviewer-honours-the-request-tests] tests: contract T1-T9
- updated: 2026-10-22T03:39:54Z

## Summary

Request-contract tests T1-T9 are committed. T8 passes at this base. T1-T7 and T9 fail because Include, Exclude, BaselineCommit and ConfigPath are ignored.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-reviewer-honours-the-request-tests
- state: stop-2-ready
- abstract: Request-contract tests T1-T9 are committed. T8 passes at this base. T1-T7 and T9 fail because Include, Exclude, BaselineCommit and ConfigPath are ignored.
- last commit: 3e810674fb76136cb8e529bb4980ad08a5d50e63 [the-reviewer-honours-the-request-tests] tests: contract T1-T9
- done: Nine request-contract tests landed; eight fail at this base for the case reason and T8 passes
  - T1-T7 and T9 fail on behaviour, proved by go test -count=1 -v ./internal/review -run TestContractW1c (exit 1)
  - T8 passes on that same run, proved by PASS TestContractW1cT8Defaults
  - every other package stayed green, proved by go test -count=1 ./... whose only FAIL lines are those eight tests
- next / blocked-on: none
- artifacts:
  - internal/review/request_contract_test.go
- evidence:
  - baseline before the file: go test -count=1 ./... exit 0. internal/review had 4 passed, 0 failed, 0 skipped: TestOpenSampleReviewKeepsF1AndF2, TestSampleReviewDoesNotReadTheEnvironment, TestCancelledContextIsPartial, TestEmptyLogDirIsRefused
  - toolchain: go version go1.26.8, module line go 1.26. No Makefile and no CI workflow in this repo. The suite command is go test -count=1 ./...
  - contract run: go test -count=1 -v ./internal/review -run TestContractW1c exit 1. 8 failed, 1 passed, 0 skipped
  - T1 FAIL request_contract_test.go:28 Exclude ignored: 2 findings remain (F1@billing/discount.go, █████████)
  - T2 FAIL request_contract_test.go:42 Include ignored: finding remains at billing/discount.go (F1@billing/discount.go, █████████)
  - T3 FAIL request_contract_test.go:65 Include checked after Exclude: finding remains at billing/discount.go (F1@billing/discount.go, █████████)
  - T4 FAIL request_contract_test.go:84 BaselineCommit ignored: review returned 2 findings and a nil error (F1@billing/discount.go, █████████)
  - T5 FAIL request_contract_test.go:96 ConfigPath ignored: Open error = <nil>, want *ConfigError for colour
  - T6 FAIL request_contract_test.go:111 timeout value not parsed: err=<nil> field=
  - T7 FAIL request_contract_test.go:120 missing file read as defaults: err=<nil>
  - T8 PASS TestContractW1cT8Defaults
  - T9 FAIL request_contract_test.go:151 config timeout never applied: Partial=false and 2 findings proved (F1@billing/discount.go, █████████)
  - after the file: go test -count=1 ./... exit 1. cmd/proven, internal/agent/judge, internal/generator, internal/netguard, internal/pipeline, internal/query, internal/testrun and internal/trace were ok. The only failures are T1 T2 T3 T4 T5 T6 T7 T9
  - pre-existing tests plus T8: go test -count=1 -v ./internal/review -run TestOpenSampleReviewKeepsF1AndF2|TestSampleReviewDoesNotReadTheEnvironment|TestCancelledContextIsPartial|TestEmptyLogDirIsRefused|TestContractW1cT8Defaults exit 0, 5 passed, 0 failed, 0 skipped
  - gofmt -l internal/review/request_contract_test.go printed nothing. go vet ./internal/review exit 0
  - race on the five passing tests: go test -race -count=1 ./internal/review -run TestOpenSampleReviewKeepsF1AndF2|TestSampleReviewDoesNotReadTheEnvironment|TestCancelledContextIsPartial|TestEmptyLogDirIsRefused|TestContractW1cT8Defaults exit 0. Race on T1-T7 and T9 is UNRUN because those tests fail on behaviour
  - testing-bar: no metric, command or baseline is declared in the order or the repo. UNRUN
  - worktree at start: d0aa8167f7031c9abe80499755e844633f753ba2. 8a356f6 is an ancestor. git diff --stat 8a356f6 HEAD -- internal/review was empty before this commit
- note: ruling: tests are package review_test — they call only Open and Review — a same-package test could pass by reading unexported fields the contract does not expose. ruling: every Open sets LogDir — an empty LogDir is already a ConfigError on Field LogDir — omitting it would fail T5-T7 for that field instead of the case. ruling: a finding is proved when it has evidence and Unresolved is false — T9 treats a zero count as no finding proved — another meaning of proved would let T9 pass while a settled finding remains. ruling: T5-T7 open with Sample true — the plan names sample on the review cases only — a refusal that exists only when Sample is false would not be seen. remark: go test ./... exit 1 is these eight tests; a different package printing FAIL is the false-world answer, and that line was absent. - updated uses the committer time from git log -1 --format=%cI because the date program is not in this role's allow list.
- updated: 2026-10-22T03:33:21+00:00

CLOSE-OK · the-reviewer-honours-the-request-tests · commit 3e810674fb76136cb8e529bb4980ad08a5d50e63 · branch test/the-reviewer-honours-the-request · worktree ██████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 5 |
