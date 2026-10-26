# STATUS — json-sarif-outputs · round 3

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 12 min
- last commit: 3d54504 [json-sarif-outputs] cli: follow dangling output links, guard the write
- updated: 2026-10-23T21:16:07Z

## Summary

Output identity now follows dangling final symlinks (Readlink, relative targets, bounded at 40 hops; loops refused at load, exit 7, flag named). The write refuses a file that is the same as one already written (os.SameFile, exit 2, both flags named). One commit on 04c5815; internal/cli green, vet and gofmt clean. Full suite has sandbox-only failures (port bind denied; schema validator module not downloadable), none in the changed code.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — json-sarif-outputs
- state: stop-2-ready
- abstract: Output identity now follows dangling final symlinks (Readlink, relative targets, bounded at 40 hops; loops refused at load, exit 7, flag named). The write refuses a file that is the same as one already written (os.SameFile, exit 2, both flags named). One commit on 04c5815; internal/cli green, vet and gofmt clean. Full suite has sandbox-only failures (port bind denied; schema validator module not downloadable), none in the changed code.
- last commit: 3d54504 [json-sarif-outputs] cli: follow dangling output links, guard the write
- done: dangling-alias collision refused at load, write-time second line of defence
  - reproduction before the fix: all 12 ordered flag pairs with alias.json -> actual.json (target absent) returned exit 0 (`go test ./internal/cli -run 'Dangling|LinkLoop|AliasCreated'`); the chain case and the write-time case also returned 0; loop and long chain exited 2 from the kernel
  - TestDanglingAliasOutputsAreRefused: 12 ordered pairs, exit 7, both flags named, no review, no file; red with the Readlink step disabled (exit 2 there, because the write guard catches it; exit 0 before either fix)
  - TestDanglingAliasChainsAreFollowed: two hops, a relative target in another directory
  - TestLinkLoopAndLongChainAreRefusedAtLoad: loop and 41-hop chain, exit 7, flag named
  - TestAliasCreatedAfterLoadIsRefusedAtWrite: link made by the review itself; exit 2, both flags named, first output intact; red with the SameFile check disabled: `exit 0, want 2`
- next: none
- artifacts:
  - internal/cli/config.go
  - internal/cli/run.go
  - internal/cli/outputs_test.go
- evidence:
  - git log --oneline 04c5815..HEAD: 3d54504 [json-sarif-outputs] cli: follow dangling output links, guard the write
  - git diff --stat 04c5815..HEAD -- AGENTS.md internal/review: empty
  - gofmt -l .: empty; go vet ./internal/cli: clean (exit 0)
  - go test -count=1 ./internal/cli: ok
  - go test -count=1 ./...: 30 packages ok; FAIL in cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review (httptest cannot bind a port here, operation not permitted) and internal/cli/sarifschema (setup failed: jsonschema/v5 not downloadable, TLS verify). `go vet ./...` fails only on that same missing module. These were not run green here; the master runs them.
  - leftover: ███████████████████████ (untracked, only `package cli`) is a scratch file I could not delete (rm denied for this role); not committed, the master should remove it at teardown
  - master-was-wrong: — (the order's measurement held; one note: with only the Readlink step removed the test returns exit 2 not 0, because the write-time guard now catches it; exit 0 was measured before both)
- updated: 2026-10-24T00:00:00Z

CLOSE-OK · json-sarif-outputs · commit 3d54504 · branch fix/json-sarif-outputs · worktree ██████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
