# STATUS — the-prover-in-a-container · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 3ec484a0120c6687ed15526216df96ad565e5065 [the-prover-in-a-container] testrun: run tests in a container
- updated: 2026-10-22T03:00:45Z

## Summary

The live Prover runs each checkout in a container with no network and no secret, and names itself off when no runtime is on PATH.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-prover-in-a-container
- state: stop-2-ready
- abstract: The live Prover runs each checkout in a container with no network and no secret, and names itself off when no runtime is on PATH.
- last commit: 3ec484a0120c6687ed15526216df96ad565e5065 [the-prover-in-a-container] testrun: run tests in a container
- done: The container runner and the live review wiring are committed.
  - go build ./... exited 0
  - go vet ./... exited 0
  - go test -count=1 ./... exited 0, and go test -race -count=1 ./... exited 0
  - TestLiveModeWithoutRuntimeNamesProverOff passed: the review completed, named the Prover off, and proved no finding by a test
  - grep -rn os/exec --include=*.go . lists internal/testrun/container.go and two test files
- next / blocked-on: none
- artifacts:
  - AGENTS.md
  - internal/testrun/container.go
  - internal/testrun/container_test.go
  - internal/testrun/live.go
  - internal/pipeline/pipeline.go
  - internal/pipeline/pipeline_test.go
  - internal/pipeline/imports_test.go
  - internal/review/open.go
  - internal/review/open_test.go
- evidence:
  - go build ./... exit 0
  - go vet ./... exit 0
  - go test -count=1 ./... exit 0. Packages ok: cmd/proven, internal/agent/judge, internal/generator, internal/netguard, internal/pipeline, internal/query, internal/review, internal/testrun, internal/trace. No package failed. No test skipped.
  - go test -race -count=1 ./... exit 0, same packages
  - internal/testrun -v: 14 tests passed, plus 9 subtests, 0 failed. internal/review -v: 5 passed, 0 failed. internal/pipeline -v: 6 passed, 0 failed.
  - RED: TestDefaultContainerConfigValidates failed while DefaultContainerConfig returned an empty config, then passed after the defaults were restored
  - baseline before the edit: go test -count=1 ./... exit 0
  - grep -rn 'os/exec' --include='*.go' . exit 0: internal/testrun/container.go, internal/testrun/container_test.go, internal/pipeline/imports_test.go
- note: master-was-wrong: UNMEASURED. M1 and M2 were not run in a container. On 2026-10-22 ls showed /usr/bin/docker and /bin/docker. The executor grant refuses the docker program, allowed form none, so the claims were not executed. docker-run(1) from Docker 29.9.1 documents --network none, --rm, --cpus, --memory and one -v with the ro option. A host go test of the fixture, not a container, exited 0 on base and 1 on the change (got 60 want 80). That host run is not M2.
  - commands: go build ./... exit 0; go vet ./... exit 0; go test -count=1 ./... exit 0; go test -race -count=1 ./... exit 0
  - session recording: none. This leg was not given a recorder (ADR-0024).
  - testing bar: no metric is declared, so the bar is UNRUN
  - toolchain: go.mod says go 1.26; local go version is go1.26.8; no Makefile; no .github CI config
  - ruling: a nil lookPath uses exec.LookPath, so os/exec stays in container.go. Cost if wrong: a nil caller gets a PATH lookup instead of an error.
  - ruling: defaults are docker, golang:1.26, 2 minutes, 1 CPU and 512m. Cost if wrong: a contract test that pins other numbers fails.
  - ruling: the checkout is mounted read-only at /checkout and the command is sh -c. Cost if wrong: a test that must write into the tree cannot, and is not proof.
  - ruling: docker exit 125, 126 and 127 mean the test did not run. Cost if wrong: a program that exits 127 is not counted as a failing test.
  - ruling: the Prover-off reason is recorded when the review starts, because the live generator is also down and the prover agent never runs. Cost if wrong: the name appears only when NewContainer failed, which is the case that runs no code.
  - ruling: testrun.Live stays for the existing unplug test. Cost if wrong: two off adapters remain.
  - ruling: no --pull flag, because that docker-run(1) does not document one. Cost if wrong: a missing image is pulled on the host before the container starts.
  - attribution: the message file has no Co-authored-by. git commit and git commit --amend both still record Co-authored-by Cursor. The git hooks do not add that line. --no-verify was not used.
  - noticed: a live review still stops at the generator, which has no provider yet, so Container.Run is not reached until that wagon. The container path is covered by unit tests.
- updated: 2026-10-22T02:56:22Z

CLOSE-OK · the-prover-in-a-container · commit 3ec484a0120c6687ed15526216df96ad565e5065 · branch fix/the-prover-in-a-container · worktree █████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
