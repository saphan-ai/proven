# STATUS — cli-over-the-sample-reviewer · round 7

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 2d600b4142bc932f54b2d85321cedfb62ed60e4e [cli-over-the-sample-reviewer] cli: look up ci config in the workspace
- updated: 2026-10-22T07:54:02Z

## Summary

proven ci takes GITHUB_WORKSPACE before it looks for proven.yaml, so the opener receives that file. An explicit --config or PROVEN_CONFIG still wins.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — cli-over-the-sample-reviewer
- state: stop-2-ready
- abstract: proven ci takes GITHUB_WORKSPACE before it looks for proven.yaml, so the opener receives that file. An explicit --config or PROVEN_CONFIG still wins.
- last commit: 2d600b4142bc932f54b2d85321cedfb62ed60e4e [cli-over-the-sample-reviewer] cli: look up ci config in the workspace
- done: ci default config is the workspace file; vet clean; full suite green
  - go test -count=1 -run 'TestInvalidVariableAloneExits7|TestValidFlagBeatsInvalidVariable' ./internal/cli on ae34b74 exited 0 before any edit
  - TestCIDefaultConfigIsTheWorkspaceFile/default failed before the fix: ConfigPath was empty while proven.yaml sat in GITHUB_WORKSPACE
  - the same assertion failed again after the early workspace assignment was removed, then passed once it was restored
  - go vet ./... exited 0
  - go test -count=1 -json ./... exited 0; per-package counts are under evidence
  - go run ./cmd/proven scan --sample testdata/fixture/discount exited 0 and printed F1 and F2 with evidence, not F3
  - go run ./cmd/proven scan --sample --error testdata/fixture/discount printed the same findings and the line exit status 1
  - go run ./cmd/proven scan --config=auto testdata/fixture/discount printed the configuration refusal and the line exit status 7
  - TestValidFlagBeatsInvalidVariable and TestInvalidVariableAloneExits7 passed in the fresh suite
- next / blocked-on: none
- artifacts:
  - internal/cli/config.go
  - internal/cli/run_test.go
- evidence:
  - toolchain go1.26.8; go.mod says go 1.26; no Makefile and no CI workflow
  - baseline before edits, go test -count=1 ./... on ae34b74966dac624aac65b838c91bab0bc01e6cb, exited 0; per-test counts on that base were not collected this round
  - tip go test -count=1 -json ./... exited 0
  - cmd/proven pass 3 fail 0 skip 0
  - internal/cli pass 115 fail 0 skip 0
  - internal/agent/judge pass 3 fail 0 skip 0
  - internal/generator pass 2 fail 0 skip 0
  - internal/netguard pass 1 fail 0 skip 0
  - internal/pipeline pass 5 fail 0 skip 0
  - internal/query pass 2 fail 0 skip 0
  - internal/review pass 4 fail 0 skip 0
  - internal/testrun pass 2 fail 0 skip 0
  - internal/trace pass 1 fail 0 skip 0
  - packages with no test files, package action skip: investigator planner prover skeptic writer candidate change fixture limit service
  - unplug red, ordering reverted, go test -count=1 -v -run TestCIDefaultConfigIsTheWorkspaceFile ./internal/cli exited 1
  - unplug line: ConfigPath "", workspace file proven.yaml; default FAIL; flag PASS; variable PASS
  - remark: pass counts include subtests; a package skip means the package has no test files
  - remark: a failing suite would exit non-zero and show a fail count above zero; this fresh run did neither
  - remark: go run's process exit is 1 for every non-zero child, because cmd/go Errorf sets status 1; the child's code is the exit status N line
  - race detector UNRUN: this change adds no shared state and no goroutine
  - testing bar UNRUN: the repository declares no coverage, mutation, lint or type metric
  - internal/review diff is empty
  - saphan doc check --family exec-status exited 0; the verb printed no RESULT line, so the exit code is the verdict
- note: master-was-wrong: the invalid-variable command was not red. On ae34b74, TestInvalidVariableAloneExits7 and TestValidFlagBeatsInvalidVariable exited 0 before this round changed anything. The red this round reproduced is the empty ConfigPath. The review finding is ADDRESSED. Ruling: the workspace is chosen in resolve, before resolveConfig, and applyCI still assigns it later when the target is still "."; after the early assignment that later check does not fire. Cost if wrong: removing only the early assignment brings the empty ConfigPath back, which the unplug showed. The reviewer seam and the timeout that starts after applyCI were left with the reviewer train. packed-refs.lock Permission denied after git commit; git log -1 shows 2d600b4142bc932f54b2d85321cedfb62ed60e4e. logs/trace.jsonl from the acceptance runs is untracked and not in the commit. date was denied by the executor grant, allowed form none; updated uses the committer time from git log -1.
- updated: 2026-10-22T07:49:48Z

CLOSE-OK · cli-over-the-sample-reviewer · commit 2d600b4142bc932f54b2d85321cedfb62ed60e4e · branch fix/cli-over-the-sample-reviewer · worktree ████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
