# STATUS — the-engines-keep-memory-and-disk-equal · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 25 min
- last commit: f4b656d86cadd7f0798481476138ff45b8108f1c [the-engines-keep-memory-and-disk-equal] testrun: say the prover is off (the leg tip 860c8f913668e328fa8903d8a8182e3ba19825ec with the tool trailer removed by the master; tree identical)
- updated: 2026-10-23T08:19:49Z

## Summary

A failed H-MEM or world write leaves memory as it was. Clear audits and resets the id sequence. Supersession sets both sides. A nil query or test adapter is a named refusal. A missing written test is noted as not found. The test runner has one refusal type, and Live's off reason is true.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-engines-keep-memory-and-disk-equal
- state: stop-2-ready
- abstract: A failed H-MEM or world write leaves memory as it was. Clear audits and resets the id sequence. Supersession sets both sides. A nil query or test adapter is a named refusal. A missing written test is noted as not found. The test runner has one refusal type, and Live's off reason is true.
- last commit: f4b656d86cadd7f0798481476138ff45b8108f1c [the-engines-keep-memory-and-disk-equal] testrun: say the prover is off (the leg tip 860c8f913668e328fa8903d8a8182e3ba19825ec with the tool trailer removed by the master; tree identical)
- done: H-MEM, the world model, the two agents and the test runner match the order
  - failed audit or persist leaves H-MEM memory unchanged, proved by TestFailedWriteLeavesMemoryAsItWas
  - Clear writes a clear audit line and the next id after reopen is mem-1, proved by TestClearAuditsAndResetsSeq
  - both supersession sides are set and the actor is the replacing id, proved by TestMarkSupersededSetsBothSides
  - a failed world upsert or ingest leaves the fold unchanged, proved by TestFailedWriteLeavesMemoryAsItWas in world
  - a nil query adapter is service query / no query engine, proved by TestNilQueryIsANamedRefusal
  - a nil test runner is service prover / no test runner, proved by TestNilRunnerIsANamedRefusal
  - a written test that was not found is noted as not found, proved by TestWrittenTestNotFoundSaysNotFound
  - one container refusal type ConfigError, proved by go vet ./... and go test ./internal/testrun
  - Live's reason no longer says the container runner is missing, proved by TestProverOffReasonIsTrue
- next / blocked-on: none
- artifacts:
  - internal/hmem/hmem.go
  - internal/hmem/hmem_test.go
  - internal/hmem/retrieve.go
  - internal/world/world.go
  - internal/world/world_test.go
  - internal/agent/investigator/investigator.go
  - internal/agent/investigator/investigator_test.go
  - internal/agent/prover/prover.go
  - internal/agent/prover/prover_test.go
  - internal/testrun/container.go
  - internal/testrun/container_test.go
  - internal/testrun/settings.go
  - internal/testrun/live.go
  - internal/testrun/testrun_test.go
- evidence:
  - baseline before any edit: go test -count=1 ./... exit 0, every package ok or no test files. Go go1.26.8. go.mod says go 1.26 and has no toolchain line. No Makefile and no workflow file. The acceptance command is the gate.
  - red before the fix, exit 1: hmem kept the record after a failed audit and a failed persist, kept a refutation, kept feedback, dropped records on a failed clear, and had no clear audit line. supersedes was empty. world kept the entity and the finding. investigator and prover panicked on a nil adapter. the not-found note was "written test passed on the change". ProverOffReason was "Prover off: no container runner yet".
  - negative control: the H-MEM publish rollback was removed. TestFailedWriteLeavesMemoryAsItWas exit 1 on add audit, add persist, refutation, feedback, clear and supersede. The query subtest stayed green because its own restore was not the field removed. The rollback was restored.
  - go build ./... exit 0. go vet ./... exit 0. gofmt -l on the edited files printed nothing.
  - go test -count=1 ./... exit 0 after the restore. Every package ok or no test files. Per-test skip counts were not printed by that command.
  - go test -race -count=1 on hmem, world, investigator, prover and testrun exit 0.
  - testing-bar: no metric is declared, so the bar is UNRUN.
- note: master-was-wrong: none. No session recording was produced in this worktree, and this leg was not given a recording path (ADR-0024). Every commit message carries a Co-authored-by trailer the tool added; the master removes it. Clear's audit actor is store and its memory id is all, because Clear takes no caller. MarkSuperseded's actor is the replacing record id and the signature stays (id, by). A nil adapter is refused at the call, so an empty proposal with a nil adapter still returns success. The not-found note is "written test was not found". ConfigError is the one refusal type because callers outside the package already use it. World ingest rolls the fold back and leaves an H-MEM refutation that already committed; the retry finds that same record. Feedback and Query roll back too, because the task covers every failed write.
- updated: 2026-10-23T08:08:00Z

CLOSE-OK · the-engines-keep-memory-and-disk-equal · commit 860c8f913668e328fa8903d8a8182e3ba19825ec · branch fix/the-engines-keep-memory-and-disk-equal · worktree ██████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
