# STATUS — the-prover-runs-only-its-test · round 1

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 23 min
- last commit: 29e566ffcf241fbf58c3b8f843ae45a3b67cad20 [the-prover-runs-only-its-test] testrun: build the test command itself
- updated: 2026-10-23T11:47:20Z

## Summary

The container command is now built by Proven from the written test's file and name (Go template, separate argv words, no shell). An answer carrying a command, or a name or path failing the strict checks, is refused with a named reason and no container starts.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-prover-runs-only-its-test
- state: stop-2-ready
- abstract: The container command is now built by Proven from the written test's file and name (Go template, separate argv words, no shell). An answer carrying a command, or a name or path failing the strict checks, is refused with a named reason and no container starts.
- last commit: 29e566ffcf241fbf58c3b8f843ae45a3b67cad20 [the-prover-runs-only-its-test] testrun: build the test command itself
- done: Property holds: nothing a model answers can shape the container's arguments or the command it runs
  - Step 1 red at base 21af174: TestModelTextDoesNotBecomeTheContainerCommand and TestRunnerRefusesACommandOutsideTheTemplate both FAILED (model text "make proof" reached the runner; "printf ready" started 2 containers); output in ███████████████████████████████████████████ and quoted under evidence
  - Acceptance 2: TestAnswerThatCarriesACommandIsRefused (written and named) and TestRunnerStartsNoContainerForAHostileName pass; the runner is not called and no container is started
  - Acceptance 3: TestArgvIsTheTemplateWithNoShell, TestRunnerRunsExactlyTheTemplate and TestArgvIsOneReadOnlyMountWithNoNetworkAndNoEnv pass; the run line ends with the image and go test -run ^Name$ ./pkg, no sh and no -c
  - Acceptance 4 negative control: the command check in prover.Prove disabled (if false &&) turned TestAnswerThatCarriesACommandIsRefused red in both subtests; restored, green
  - Acceptance 5: proven scan --sample on testdata/fixture/discount from a scratch directory prints F2, F4, osv:evil.example/leftpad and F1 as docs/SAMPLE_MODE.md records for main; only the command line of F1 differs
- next / blocked-on: master review (named) and the owner's gate; nothing blocks this leg; W28 container flags and W30 traces are untouched
- artifacts:
  - ███████████████████████████████████████████████████████████████
  - ████████████████████████████████████████████████████████████
  - ███████████████████████████████████████████████████████
- evidence:
  - Red at base (go test -count=1 -run TestModelTextDoesNotBecomeTheContainerCommand|TestRunnerRefusesACommandOutsideTheTemplate ./internal/agent/prover ./internal/testrun, exit 1): command_test.go:21 the command is the model's text, or not Proven's template: "make proof"; command_test.go:19 found true, containers started 2
  - go build ./... exit 0; go vet ./... exit 0; gofmt -l on internal and cmd clean after gofmt -w of three files
  - go test -count=1 on agent/..., testrun, fixture, cmd/proven, demo, pipeline, review: all ok. Pipeline and review were run with -skip of 12 tests that open a local listener (guardrail_test, zetaris_test, live_generator_test, TestLiveModeWithoutRuntimeNamesProverOff)
  - UNRUN: cmd/livecheck, internal/generator, internal/guardrail, internal/query tests and the 12 skipped tests of pipeline and review. This seat refuses a listener (httptest: failed to listen on a port: operation not permitted), the panic aborts the package, and none of them touch the changed code. The full go test ./... is therefore not green on this seat; the master runs it
  - UNRUN: TestContractW1bT10 and TestWrittenGoTestRunsInTheContainer need PROVEN_TEST_DOCKER=1 and a runtime. T10 was rewritten from go run dial.go to a dial_test.go with a go.mod, so it is untested here
  - UNRUN: a fresh scan on main 21af174 for a side-by-side diff. tar and rm are denied to this seat, so the base was compared with the output recorded in docs/SAMPLE_MODE.md
  - Session recording: none; this one-shot headless leg keeps no recording (ADR-0024)
  - Commands and exit codes: go test red-at-base 1; go build 0; go vet 0; sample scan 0; negative-control run 1 (red as intended); commit landed, packed-refs.lock Permission denied as the order says
  - Co-authored-by trailer: none in the commit message
- note: master-was-wrong: none of the order's claims was contradicted at base 21af174; prover.go carried the model's command into Request.Command and the container ran it with sh -c. What the order did not say and I did: (1) I removed the command field from the Prover and Skeptic prompts and from Request and Refutation, so a fixture's recorded prompts changed in discount and refuted, and the recorded F1 command line now reads go test -run '^TestDiscounted$' ./billing; (2) the Skeptic's refuting test goes through the same checks via prover.Execute; (3) the named-test path gives a refused note, where it gave none before. Rulings: Go only, other languages are refused as not a test file, cost if wrong is a later template per language; name pattern Test followed by an upper-case letter, digit, underscore or nothing, cost if wrong is a legal lower-case test name refused; directory segments limited to letters, digits, underscore, dot and dash, cost if wrong is an unusual directory refused; no -count flag in the template, as the order's template has none
- updated: 2026-10-23T12:00:00Z

CLOSE-OK · the-prover-runs-only-its-test · commit 29e566ffcf241fbf58c3b8f843ae45a3b67cad20 · branch fix/the-prover-runs-only-its-test · worktree █████████████████████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 5 |
