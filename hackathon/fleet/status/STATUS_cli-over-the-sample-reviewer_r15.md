# STATUS — cli-over-the-sample-reviewer · round 15

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 30 min
- last commit: 9412b829960ded9756f18905b0a70583b74741f9 [cli-over-the-sample-reviewer] cli: isolate fixture git config
- updated: 2026-10-22T17:35:26Z

## Summary

W1b is on fe146e0. The text format prints the review class first. Fixture git ignores host config, and go test ./... is green.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — cli-over-the-sample-reviewer
- state: stop-2-ready
- abstract: W1b is on fe146e0. The text format prints the review class first. Fixture git ignores host config, and go test ./... is green.
- last commit: 9412b829960ded9756f18905b0a70583b74741f9 [cli-over-the-sample-reviewer] cli: isolate fixture git config
- done: W1b is on fe146e0, the text format prints the review class, and fixture git ignores host config
  - git rev-parse HEAD printed 9412b829960ded9756f18905b0a70583b74741f9
  - git log --oneline fe146e0..HEAD lists nine commits and no merge commit
  - git diff --stat fe146e0..HEAD -- internal/review printed no output
  - go test -count=1 -run 'TestScanOpenFamilyNameIsAccepted|TestScanMalformedFamilyExits7|TestScanUnknownClassExits7|TestScanSamplePrintsF1AndF2NotF3' ./cmd/proven exited 0
  - go test -count=1 -v -run 'TestCIBaseRefWhenTheEventHasNoSHA|TestCIWorkspaceIsTheTarget|TestCIMergeBaseIsTheForkPoint|TestCIUnresolvableBaseExits2' ./internal/cli printed four PASS lines and exited 0
  - With cfg.baseline = base discarded, TestCIMergeBaseIsTheForkPoint failed because the baseline SHA equalled the base tip; after restore the same command exited 0
  - go vet ./... exited 0 and go test -count=1 ./... exited 0
- next / blocked-on: none
- artifacts:
  - internal/cli/format.go
  - internal/cli/run_test.go
  - cmd/proven/main_test.go
- evidence:
  - git cat-file -t fe146e0 printed commit
  - git merge-tree --write-tree fe146e0 HEAD exited 1 and printed CONFLICT for AGENTS.md, cmd/proven/main.go, cmd/proven/main_test.go, and other files including internal/review
  - git rebase --onto fe146e0 2a1ea87 stopped at c6fefdc; git rebase --continue exited 1 with could not remove CHERRY_PICK_HEAD after packed-refs.lock Permission denied. The remaining commits were cherry-picked. Each commit command printed packed-refs.lock Permission denied and exited 0. git log -1 showed the commit.
  - At 6f30e14, withRefusingGlobalHook plus go test -count=1 -run 'TestCI' ./internal/cli failed four tests. The failure text was git-rule-subject-pattern for subjects root and only, not the helper hook's empty exit 1.
  - printenv showed GIT_CONFIG_GLOBAL unset, GIT_CONFIG_COUNT=1, GIT_CONFIG_KEY_0=core.hooksPath, GIT_CONFIG_VALUE_0 ending in ████████████████████ git --version printed git version 2.43.0. man was denied by the grant, so the precedence claim is this measurement.
  - go test -count=1 ./... : 23 packages ok, 0 FAIL, 9 packages printed [no test files]. cmd/proven 3.432s, internal/cli 5.646s, internal/review 30.936s, internal/pipeline 23.920s, internal/testrun 12.164s, internal/change 8.434s. The other ok packages were under 1s except those named.
  - remark: go test ./... prints ok or FAIL per package and does not print pass, fail, or skip counts inside a package. A failed package would print FAIL. None did.
  - remark: local go version is go1.26.8 linux/amd64. go.mod says go 1.26 and has no toolchain line. No Makefile and no workflow file was found.
  - remark: no testing-bar metric is declared, so coverage, mutation, lint, and type counts are UNRUN. -race is UNRUN because this round does not change shared state and the order does not gate on it.
- note: master-was-wrong — merge-tree conflicts were not limited to cmd/proven/main.go and cmd/proven/main_test.go, because fe146e0 and 6f30e14 diverged at bc41431634dcf4bfb9e9a8d92e4479aec099fa00. The refusing hook is injected by GIT_CONFIG_COUNT, not by a GIT_CONFIG_GLOBAL file, so the three variables named in the order leave core.hooksPath in force and test: root is still refused. runGit also drops GIT_CONFIG_COUNT, GIT_CONFIG_KEY_*, GIT_CONFIG_VALUE_*, and GIT_CONFIG_PARAMETERS. Review condition ADDRESSED by the order's hermetic fixtures, not by subjects the active hook accepts. The review said do not suppress the hook. Fixture git does not see that hook. The hook still ran on this round's commits. The status key set has no master-was-wrong field, so this note holds it. ruling: formatter write errors are checked, and the five format strings stay in main's order — cost is a stricter Write than main's printResult. ruling: the three scan tests call cli.Run with os.Getenv and keep their assertions — cost is host environment visible to sample mode. ruling: assertDiscountText also requires main's review-class lines — cost is a stricter sample test than the three named tests.
- updated: 2026-10-22T17:27:55Z

## REDACTED

| what | times |
|---|---|
| private workspace path | 1 |
