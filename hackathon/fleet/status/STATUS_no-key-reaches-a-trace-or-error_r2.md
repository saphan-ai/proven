# STATUS — no-key-reaches-a-trace-or-error · round 2

- state: ready-for-review
- executor: not recorded
- model: not recorded
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: ba02f85 [no-key-reaches-a-trace-or-error] absorb main (the master: main absorbed, four conflicts hand-resolved, 35 packages ok)
- updated: 2026-10-23T17:34:02Z

## Summary

Round 2. Every finding and evidence string is redacted once at the end of Pipeline.Review (C1); the planner's duplicate-id error names a position, never model text (C2); Build refuses a live pipeline with no redactor (C3); comments cite ADR-0005 and ADR-0013 point 4 (C4). The reviewer's two tests, renamed to the property, pass. Six packages cannot bind a loopback listener on this host: UNMEASURED, not green.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — no-key-reaches-a-trace-or-error
- state: stop-2-ready
- abstract: Round 2. Every finding and evidence string is redacted once at the end of Pipeline.Review (C1); the planner's duplicate-id error names a position, never model text (C2); Build refuses a live pipeline with no redactor (C3); comments cite ADR-0005 and ADR-0013 point 4 (C4). The reviewer's two tests, renamed to the property, pass. Six packages cannot bind a loopback listener on this host: UNMEASURED, not green.
- last commit: ba02f85 [no-key-reaches-a-trace-or-error] absorb main (the master: main absorbed, four conflicts hand-resolved, 35 packages ok)
- done: C1-C4 landed in one commit
  - C1: TestTokenInAFindingDoesNotReachTheDraft passes; negative control (redaction call removed) turned it red with "the finding carries the token", then restored
  - C2: TestTokenInAFatalAgentErrorDoesNotReachTheError passes
  - C3: TestLivePipelineWithoutARedactorIsRefused passes; the one live Build in pipeline_test.go now passes a redactor
  - C4: ADR-0025 citations about the trace, result and redactor re-pointed in pipeline.go, trace.go, review/open.go, redact.go
  - TestRedactFindingsReachesEvidence covers evidence, refutation rows and the caller's copy staying unchanged
- next / blocked-on: master re-runs the six bind-blocked packages (C5) on a host that permits a loopback listener
- artifacts:
  - internal/pipeline/pipeline.go
  - internal/pipeline/result_redaction_test.go
  - internal/agent/planner/planner.go
- note: master-was-wrong: nothing refuted; the reviewer's two paths reproduced as described. One addition the order did not name: "a risk has no id" also lost the id, now "risk N has no id". Commands: gofmt -l (exit 0, no output); go build ./... && go vet ./... (exit 0); go test -count=1 ./... gave 28 ok and 6 FAIL (cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review), all "httptest: failed to listen on a port: bind: operation not permitted". In internal/pipeline the network-free tests were run with -run and pass (TestToken*, TestRedactFindingsReachesEvidence, TestLivePipeline*, TestConfiguredValueEchoedInAReceiptIsNotWritten, TestGuardrailKeyEchoedInTheCategoriesIsNotWritten, TestLiveContainerSettingIsRefused). internal/review live tests build through review.Open, which sets a redactor, so C3 does not touch them, but they are unmeasured here. The commit tool printed packed-refs.lock Permission denied; git log shows 3cfc7f8 landed; no Co-authored-by line. No session recording was made: a headless leg with no recorder (ADR-0024).
- updated: 2026-10-23

## REDACTED

none
