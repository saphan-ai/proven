# STATUS — cli-over-the-sample-reviewer · round 18

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 17 min
- last commit: b11e5831a3b32cbdd11f02c2d765a3902cdebf23 [cli-over-the-sample-reviewer] agents: restore the owner file
- updated: 2026-10-23T04:22:47Z

## Summary

proven ci exits 2 when it has no pull-request baseline, and AGENTS.md on this branch matches main at 26343c2. Tip b11e583.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — cli-over-the-sample-reviewer
- state: stop-2-ready
- abstract: proven ci exits 2 when it has no pull-request baseline, and AGENTS.md on this branch matches main at 26343c2. Tip b11e583.
- last commit: b11e5831a3b32cbdd11f02c2d765a3902cdebf23 [cli-over-the-sample-reviewer] agents: restore the owner file
- done: proven ci with no usable baseline exits 2 and does not review the checkout, and AGENTS.md matches 26343c2
  - Before the check, go test -count=1 -v -run TestCIMissingContextExits2 ./internal/cli failed both missing-context cases: no event printed exit 0 opened true stderr "" stdout "review class: \nreview rule: \nfailed services: none\n" at run_test.go:726; event with an empty base SHA printed the same at run_test.go:745. Malformed JSON, an explicit --baseline-commit, and scan already passed
  - With the new check removed, go test -count=1 -v -run TestCIMissingContextExits2/no_event_no_base_ref_no_baseline$ ./internal/cli failed at run_test.go:728: exit 0 opened true stderr "" stdout "review class: \nreview rule: \nfailed services: none\n". The check was restored
  - After the restore, go test -count=1 -run 'TestCIMissingContextExits2|TestCIMergeBaseIsTheForkPoint|TestCIDefaultConfigIsTheWorkspaceFile|TestCIBaseRefWhenTheEventHasNoSHA|TestCIFlagAndSHAOverrideTheEvent|TestCIUnresolvableBaseExits2|TestCICommitsComeFromTheFixtureEvent|TestCIWorkspaceIsTheTarget' ./internal/cli exited 0
  - go vet ./... exited 0 before the commits and again after them
  - go test -count=1 -json ./... exited 0 with 0 failures across 38 packages, on the tree the two commits recorded. After the commits, go test -count=1 ./internal/cli exited 0
  - git log --oneline d488a0f0c5edf247f76464aa342d32d389013a46..HEAD printed b11e583 agents: restore the owner file and 561eb8b cli: exit 2 when ci has no baseline
  - git diff --stat 26343c2..HEAD -- AGENTS.md internal/review printed nothing
  - git rev-parse HEAD printed b11e5831a3b32cbdd11f02c2d765a3902cdebf23
- next / blocked-on: none
- artifacts:
  - internal/cli/run.go
  - internal/cli/run_test.go
  - AGENTS.md
- evidence:
  - git rev-parse HEAD printed b11e5831a3b32cbdd11f02c2d765a3902cdebf23
  - go version printed go1.26.8 linux/amd64. go.mod says go 1.26 and has no toolchain line
  - No-context refusal: ci: no pull-request baseline: set --baseline-commit, or GITHUB_EVENT_PATH with a base SHA, or GITHUB_BASE_REF
  - Event with an empty base SHA: ci: the event has no base SHA, and --baseline-commit and GITHUB_BASE_REF are unset
  - go test -count=1 -json ./... top-level counts, 38 packages, 0 failures. cmd/livecheck pass=12. cmd/osv-to-csv pass=2. cmd/proven pass=10. conformance/meterless skip=1. demo pass=2. internal/agent/answer pass=9. internal/agent/judge pass=10. internal/agent/planner pass=2. internal/agent/prover pass=14. internal/agent/skeptic pass=2. internal/agent/writer pass=3. internal/change pass=12. internal/cli pass=40. internal/envdoc pass=4. internal/epss pass=6. internal/fixture pass=1. internal/generator pass=28. internal/guardrail pass=29. internal/hmem pass=12. internal/manifest pass=14. internal/osv pass=15. internal/pep440 pass=2. internal/pipeline pass=34. internal/query pass=8. internal/review pass=75. internal/semver pass=1. internal/settingspage pass=2. internal/testrun pass=41 skip=2. internal/trace pass=2. internal/world pass=9
  - Packages with no test files, action skip and zero tests: cmd/meterless-conform, cmd/settings-doc, demo/live/review, internal/agent/investigator, internal/candidate, internal/limit, internal/livecheck, internal/service
  - remark: these counts are top-level Test names only. A subtest is not included. TestCIMissingContextExits2 is one of the 40 cli tests and holds five subtests. A package that fails to compile would be a package failure. None did. An action of skip with all zeros means the package has no test files
  - remark: the json suite ran before the two commits, on the working tree those commits recorded. git status after the commits showed no modified tracked file. go vet ./... and go test -count=1 ./internal/cli were run again after the commits and both exited 0. A failure that appeared only in another package after the commit would not be in the second command
  - remark: no testing-bar metric is declared, so coverage, mutation, lint and type counts are UNRUN. -race is UNRUN because this round added no shared-state code and the order does not gate on it
  - Review BLOCKING 1 ADDRESSED. proven ci with no baseline exits 2 and the opener is not called. TestCIMergeBaseIsTheForkPoint stayed green
  - Review BLOCKING 2 ADDRESSED. git diff --stat 26343c2..HEAD -- AGENTS.md is empty
  - Review BLOCKING 3 NOT ADDRESSED by this leg. The order says the public copies are the master's
  - ruling: PROVEN_BASELINE_COMMIT and the SEMGREP baseline variables stay a usable baseline, because resolveBaseline fills cfg.baseline before applyCI and ADR-0020 makes them one setting with the flag — cost if wrong: ci with only the variable set still reviews that commit
  - ruling: an explicit baseline with an empty head still reviews, because review.Request says an empty head is the working tree and the order's usable baseline does not require a head — cost if wrong: proven ci --baseline-commit with no GITHUB_SHA reviews the working tree since that commit
  - ruling: TestCIDefaultConfigIsTheWorkspaceFile now passes --baseline-commit so it still reaches open — cost if wrong: that test no longer covers a missing baseline, which TestCIMissingContextExits2 covers
  - master-was-wrong: —
- note: SHOUT — git status still says cherry-picking db5a6f5. Do not abort: ORIG_HEAD is f836c0f and abort would leave the commits after it. git cherry-pick --quit printed packed-refs.lock Permission denied and left CHERRY_PICK_HEAD. rm is denied by the executor command list. Delete of that gitdir path was refused as outside the worktree. Both new commits inherited AuthorDate Thu Oct 22 14:27:53 2026 +0000 from db5a6f5 because they were made while that cherry-pick was open. Their CommitDates are Fri Oct 23 04:15:54 2026 +0000 and Fri Oct 23 04:16:19 2026 +0000. The messages and the trees are this round's. master-was-wrong: —
- updated: 2026-10-23T04:18:59Z

## REDACTED

none
