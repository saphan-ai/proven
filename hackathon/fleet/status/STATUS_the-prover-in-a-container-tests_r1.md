# STATUS — the-prover-in-a-container-tests · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 18 min
- last commit: 33c0dcf199f6167d40b57549223f17c3326b23d3 [the-prover-in-a-container-tests] tests: add container contract tests
- updated: 2026-10-22T02:54:27Z

## Summary

Added the container contract tests. At this base T1–T9 fail on the stubs and T10 skips; master verification is next.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-prover-in-a-container-tests
- state: stop-2-ready
- abstract: Added the container contract tests. At this base T1–T9 fail on the stubs and T10 skips; master verification is next.
- last commit: 33c0dcf199f6167d40b57549223f17c3326b23d3 [the-prover-in-a-container-tests] tests: add container contract tests
- done: T1–T9 fail at this base for the plan's reasons, and T10 skips.
  - T1–T9 FAIL and T10 SKIP: go test -count=1 -v ./internal/testrun/ -run TestContractW1b exit 1
  - the two tests already in the package still pass: go test -count=1 -v ./internal/testrun/ -run 'TestSampleReturnsTheRecordedTest|TestLiveNamesProverOff' exit 0, 2 passed, 0 failed, 0 skipped
  - no other package failed: go test -count=1 ./... exit 1, and the only FAIL line is internal/testrun
  - the new file is gofmt-clean and vet-clean: gofmt -l on the file printed nothing, exit 0; go vet ./internal/testrun/ exit 0
- next / blocked-on: none
- artifacts:
  - internal/testrun/container_contract_test.go
- evidence:
  - command: go test -count=1 -v ./internal/testrun/ -run TestContractW1b; exit 1; 9 FAIL, 1 SKIP, 0 PASS
  - T1 FAIL container_contract_test.go:50: T1: argv must carry --network none or --network=none; NewContainer: container=false err=testrun: container runner not implemented
  - T2 FAIL container_contract_test.go:61: T2: argv must pass no Proven environment (no -e, --env, or --env-file, and no element contains s3cret-T2); NewContainer: container=false err=testrun: container runner not implemented
  - T3 FAIL container_contract_test.go:74: T3: argv must mount exactly the given dir and no other host path; NewContainer: container=false err=testrun: container runner not implemented
  - T4 FAIL container_contract_test.go:91: T4: argv must carry --rm, the CPU limit 1 and the memory limit 512m; NewContainer: container=false err=testrun: container runner not implemented
  - T5 FAIL container_contract_test.go:119: T5: NewContainer must return *service.Down with Service prover when lookPath fails; got testrun: container runner not implemented (service "")
  - T6 FAIL container_contract_test.go:135: T6: a command that fails on the change and passes on the base is proved, and the runtime runs twice; NewContainer: container=false err=testrun: container runner not implemented
  - T7 FAIL container_contract_test.go:150: T7: a run past its time limit returns within 10s and is not proof, or the error names the time limit; NewContainer: container=false err=testrun: container runner not implemented
  - T8 FAIL container_contract_test.go:176: T8: Validate must refuse an empty Image by name; got testrun: container runner not implemented
  - T9 FAIL container_contract_test.go:198: T9: the two runs mount BaseDir and ChangeDir, one each; NewContainer: container=false err=testrun: container runner not implemented
  - T10 SKIP container_contract_test.go:217: PROVEN_TEST_DOCKER=1
- note: Go 1.26.8, go.mod says go 1.26 and has no toolchain line. No Makefile and no CI workflow in the repo. Baseline before the edit: go test -count=1 ./... exit 0. Race on the two pre-existing tests: go test -race -count=1 -run 'TestSampleReturnsTheRecordedTest|TestLiveNamesProverOff' ./internal/testrun/ exit 0. Contract tests under -race are UNRUN; they are red before any goroutine starts. golangci is UNRUN; the repo has no linter config. No testing-bar metric is declared. The stub file returns ErrContainerNotImplemented and zero values, so the behaviour was not already present. No case was untestable through the contract. ruling: package testrun_test — the package already tests the exported surface that way — cost is none, the contract is exported. ruling: an empty argv fails T2 — a stub must not pass by returning nothing — cost is a correct runner must return a non-empty argv. ruling: mounts are -v, --volume, or --mount with source= or src= — the plan names -v or --mount — cost is another spelling stays red. ruling: another host path is an absolute token that exists, excluding the checkout and the resolved runtime — a container target usually does not exist on the host — cost is a container path that exists on the host is rejected. ruling: CPU is --cpus value 1, memory is --memory or -m value 512m, and --rm is exact — the plan names --rm and the two limit strings — cost is 1.0 or 512MB stays red. ruling: proved means FailsOnChange and PassesOnBase; a time-limit error contains timeout, timed out, time limit, deadline, or context.DeadlineExceeded; Run is given context.Background and the test bound is 10s — so an ignored Timeout cannot hide behind the passed context — cost is an error without those words stays red. ruling: T8 matches the field names Image and Timeout — the contract spells them that way — cost is a lowercase-only error stays red. ruling: T9 wants each checkout once, in either order — the plan does not name an order — cost is none. ruling: T10 runs go run dial.go with the checkout as the working directory, allows Run 2 minutes, and skips with the text PROVEN_TEST_DOCKER=1 — the image that runs Go tests is assumed to contain go — cost is an image without go, or a different workdir, fails when the master runs T10. remark: go test ./... without -v does not print a skip, so T10 is taken from the -v run. remark: ok on a package line does not give a pass count; the count of 2 is from -v on internal/testrun before the edit. remark: updated is the committer time of 33c0dcf199f6167d40b57549223f17c3326b23d3, because the clock command is denied to this role.
- updated: 2026-10-22T02:47:55Z

CLOSE-OK · the-prover-in-a-container-tests · commit 33c0dcf199f6167d40b57549223f17c3326b23d3 · branch test/the-prover-in-a-container · worktree ███████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
