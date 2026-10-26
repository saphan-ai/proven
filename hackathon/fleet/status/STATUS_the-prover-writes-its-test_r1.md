# STATUS — the-prover-writes-its-test · round 1

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 748b30de329241e7ed9e9014dcd3b4306bd95cc6 [the-prover-writes-its-test] prover: write the proving test
- updated: 2026-10-22T13:58:57Z

## Summary

The prover asks for a test file and the runner writes it into a copy of each checkout. Evidence still requires a failure on the change and a pass on the base. Sample mode still names an existing test when the model returns no file.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-prover-writes-its-test
- state: stop-2-ready
- abstract: The prover asks for a test file and the runner writes it into a copy of each checkout. Evidence still requires a failure on the change and a pass on the base. Sample mode still names an existing test when the model returns no file.
- last commit: 748b30de329241e7ed9e9014dcd3b4306bd95cc6 [the-prover-writes-its-test] prover: write the proving test
- done: The prover writes the test that proves a finding, and that test runs only in a copy of each checkout.
  - go build ./... && go vet ./... && go test -count=1 ./... exited 0
  - fake generator and fake runner: a written test that fails on the change and passes on the base is evidence (TestWrittenTestThatFailsOnChangeIsEvidence)
  - the same pair, passing on both, adds no evidence and the trace text is "written test passed on the change" (TestWrittenTestThatPassesOnBothIsNamedInTheTrace)
  - absolute path, ../, a non-test name, an existing non-test file, content over the size limit, and an empty command are refused by those names and the runner is not called
  - TestWrittenGoTestRunsInTheContainer skipped by name PROVEN_TEST_DOCKER=1
  - negative controls (a) (b) (c) went red and were restored; the three tests then passed
  - sample scan prints F1, F2 and F4 and does not print F3 (TestScanSamplePrintsF1AndF2NotF3)
- next / blocked-on: none
- artifacts:
  - internal/agent/prover/prover.go
  - internal/agent/prover/prover_test.go
  - internal/candidate/candidate.go
  - internal/change/between_test.go
  - internal/pipeline/pipeline.go
  - internal/pipeline/pipeline_test.go
  - internal/review/selection_test.go
  - internal/review/zetaris_test.go
  - internal/testrun/container.go
  - internal/testrun/container_test.go
  - internal/testrun/testrun.go
  - internal/testrun/testrun_test.go
  - internal/testrun/written.go
  - internal/testrun/written_test.go
  - testdata/fixture/discount/fixture.json
- evidence:
  - baseline go test -count=1 ./... exited 1 before the subject fix: internal/change and internal/review, commit-msg hook refused subjects first, second, base and head
  - go version go1.26.8; go.mod says go 1.26
  - go build ./... && go vet ./... && go test -count=1 ./... exited 0
  - negative control (a), passes-on-base branch removed: FAIL TestPassesOnBothWritesNoEvidence note "written test failed on the base"; restored
  - negative control (b), ../ check removed: FAIL TestRefusesPathThatLeavesTheRepository runner was called; restored
  - negative control (c), file written into the real checkout: FAIL TestWrittenTestIsCopiedAndTheCheckoutStaysUntouched wrote the checkout; restored
  - TestWrittenGoTestRunsInTheContainer skipped: PROVEN_TEST_DOCKER=1
  - go run ./cmd/proven scan --sample testdata/fixture/discount from the module root exited 2: hmem refuses a state directory inside the repository. That refusal is pre-existing. The suite runs the same scan after leaving the module.
  - scan output: finding F1 ERROR Discount is applied twice; evidence test billing/discount_test.go TestDiscounted fails_on_change=true passes_on_base=true; finding F2 ERROR Dependency leftpad v1.2.3 is on the known-exploited list; finding F4 ERROR CVE-2026-0001 for leftpad v1.2.3 is on the KEV list; failed services: none; finding F3 is absent
  - git log --format=%B a9aced6..HEAD | grep -ci co-authored printed 0
  - no declared testing-bar metric, so that bar is UNRUN
  - go test -race UNRUN: the order did not ask for it, and this change does not add a goroutine
- note: master-was-wrong: none. M1 held — the new test did not compile because Request had no Content and Prove could not take a file. M2 held — the container mounted each checkout and the request carried no file. No session recording was attached to this spawn (ADR-0024); the environment had no recording id. ruling: a test file is a base name ending in _test.go — cost if wrong, a non-Go test is refused as not a test file. ruling: the size limit is ContainerConfig.MaxWrittenBytes, default 65536, because proven.yaml's key set is closed. NOTICED: AGENTS.md is the owner's file and its container-defaults table does not list that setting. ruling: a written test that passes on the change and fails on the base is traced as "written test failed on the base". ruling: git commit appended a Co-authored-by trailer; the hook had accepted the message, and the branch objects were rewritten without the trailer. Earlier commit a5fa59cca51461bea6548198a21f7dd87d6e192f is the subject fix.
- updated: 2026-10-22T12:55:48Z

CLOSE-OK · the-prover-writes-its-test · commit f6c125939bc1d21223f67360539640235fa2872f · branch fix/the-prover-writes-its-test · worktree ██████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
