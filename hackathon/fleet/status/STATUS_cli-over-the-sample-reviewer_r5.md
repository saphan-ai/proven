# STATUS — cli-over-the-sample-reviewer · round 5

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 19 min
- last commit: ae34b74966dac624aac65b838c91bab0bc01e6cb [cli-over-the-sample-reviewer] cli: repair three shell findings
- updated: 2026-10-22T04:00:16Z

## Summary

CI baseline is the merge base, a valid flag beats an invalid variable, and a panic while opening exits 2. The reviewer seam is unchanged.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — cli-over-the-sample-reviewer
- state: stop-2-ready
- abstract: CI baseline is the merge base, a valid flag beats an invalid variable, and a panic while opening exits 2. The reviewer seam is unchanged.
- last commit: ae34b74966dac624aac65b838c91bab0bc01e6cb [cli-over-the-sample-reviewer] cli: repair three shell findings
- done: three shell repairs landed, vet clean, full suite green
  - go vet ./... exited 0
  - go test -count=1 ./... exited 0; per-package counts are under evidence
  - TestCIMergeBaseIsTheForkPoint red with the merge-base call reverted: baseline was the base tip 49907be3a88a146a8380e5a710411a6082fc7572, fork was 6ca0647817e306288d809d962e5060abad626bbf
  - TestValidFlagBeatsInvalidVariable and TestSampleFlagIgnoresInvalidSampleMode red with the parsers reverted: each valid flag exited 7 and named the invalid variable, including PROVEN_SEVERITY=invalid under --severity=INFO
  - TestOpenerPanicExits2 red with openCall reverted: panic open-boom, recovered and repanicked by the test harness
  - go run ./cmd/proven scan --sample testdata/fixture/discount exited 0 and printed F1 and F2 with evidence, not F3
  - go run ./cmd/proven scan --sample --error testdata/fixture/discount printed the same findings and the go tool line exit status 1
  - go run ./cmd/proven scan --config=auto testdata/fixture/discount printed the configuration refusal and the go tool line exit status 7
- next / blocked-on: none
- artifacts:
  - internal/cli/config.go
  - internal/cli/run.go
  - internal/cli/run_test.go
- evidence:
  - toolchain go1.26.8; go.mod says go 1.26; no Makefile and no CI workflow
  - baseline before edits, go test -count=1 ./... on cca7e5c, exited 0; per-test counts on that base were not collected this round
  - tip go test -count=1 -json ./... exited 0
  - cmd/proven pass 3 fail 0 skip 0
  - internal/cli pass 111 fail 0 skip 0
  - internal/agent/judge pass 3 fail 0 skip 0
  - internal/generator pass 2 fail 0 skip 0
  - internal/netguard pass 1 fail 0 skip 0
  - internal/pipeline pass 5 fail 0 skip 0
  - internal/query pass 2 fail 0 skip 0
  - internal/review pass 4 fail 0 skip 0
  - internal/testrun pass 2 fail 0 skip 0
  - internal/trace pass 1 fail 0 skip 0
  - packages with no test files, package action skip: investigator planner prover skeptic writer candidate change fixture limit service
  - remark: pass counts include subtests; a package skip means the package has no test files
  - remark: a failing suite would exit non-zero and show a fail count above zero; this run did neither
  - remark: go run's process exit is 1 for every non-zero child, because cmd/go Errorf sets status 1; the child's code is the exit status N line
  - race detector UNRUN: this change adds no shared state and no goroutine
  - testing bar UNRUN: the repository declares no coverage, mutation, lint or type metric
  - internal/review diff is empty
  - unresolvable base exits 2, not 7: ADR-0020 section 2, Proven could not review; exit 7 is a configuration refused at load
- note: master-was-wrong: the shell could refuse a non-empty --include or --exclude today, because review.Open drops them and a green scan would claim a filter it did not apply. Not built. Do not refuse BaselineCommit or HeadCommit; this round passes the merge base, and the reviewer train owns honouring it. Cost if that proposal is wrong: path filters stay silent until that train lands. Finding 1 (selection ignored by the reviewer) and finding 4 (config file never loaded) are NOT ADDRESSED. Finding 2 ADDRESSED. Finding 3 ADDRESSED. The open-panic OWED is ADDRESSED; TestReviewerPanicExits2 still covers the review stage. Ruling: an explicit baseline flag or variable is not rewritten by merge-base, because ADR-0020 says that flag means review since that commit; cost if wrong: a pipeline that sets the baseline variable to an advanced branch tip diffs against the tip. Ruling: --output, --include, --exclude and the baseline variables have no refusal grammar, so they have no exit-7 invalid-value test; --config already chose the winner before checking the path, and TestConfigFlagIgnoresInvalidVariable stayed green when the five parsers were reverted. The review's shell prefix PROVEN_SEVERITY=invalid was refused by the grant (allowed form none). The same call was red in TestValidFlagBeatsInvalidVariable/severity before the fix. Old tests edited because they asserted the copied SHA: TestCICommitsComeFromTheFixtureEvent now expects exit 2; TestCIBaseRefWhenTheEventHasNoSHA expects the fork; TestCIWorkspaceIsTheTarget uses a real repository so the target is still visible. packed-refs.lock Permission denied after git commit; git log -1 shows ae34b74966dac624aac65b838c91bab0bc01e6cb. logs/trace.jsonl from the acceptance run is untracked and not in the commit.
- updated: 2026-10-22T03:54:57Z

CLOSE-OK · cli-over-the-sample-reviewer · commit ae34b74966dac624aac65b838c91bab0bc01e6cb · branch fix/cli-over-the-sample-reviewer · worktree ████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
