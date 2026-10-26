# STATUS — the-generator-through-nightshift · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 39 min
- last commit: dc276c89a5167c65f7c05dcd12ae1d3fd6968125 [the-generator-through-nightshift] generator: read the routed model
- updated: 2026-10-22T06:44:00Z

## Summary

The live generator streams an OpenAI-compatible completion, records the routed model and receipt id, refuses a non-NVIDIA answer, and degrades by name. The master verifies next.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — the-generator-through-nightshift
- state: stop-2-ready
- abstract: The live generator streams an OpenAI-compatible completion, records the routed model and receipt id, refuses a non-NVIDIA answer, and degrades by name. The master verifies next.
- last commit: dc276c89a5167c65f7c05dcd12ae1d3fd6968125 [the-generator-through-nightshift] generator: read the routed model
- done: The live generator calls an OpenAI-compatible stream, checks the routed model, and degrades by name.
  - go build ./... && go vet ./... && go test -count=1 ./... exited 0 on the tip
  - testdata/generator/nightshift-stream.sse replayed through httptest: text ready, model nvidia/nemotron-3.5-super-vl-120b-a12b, receipt rcpt_req_k67ooqsx, FromReceipt true
  - mutated meterless.routed_model to openai/gpt-4o: refused by name; with the check removed, TestLiveRefusesANonNVIDIARoutedModel failed on text "ready"
  - mutated field meterless removed: model unknown, receipt empty, text ready
  - mutated HTTP status to 401: nightshift unauthorized, one call, review degraded
  - mutated stream body to data: [DONE]: two calls, then nightshift empty answer
  - closed httptest endpoint: review completed, not partial, failed service nightshift unreachable
- next / blocked-on: none
- artifacts:
  - AGENTS.md
  - internal/candidate/candidate.go
  - internal/generator/generator_test.go
  - internal/generator/live.go
  - internal/generator/live_test.go
  - internal/generator/settings.go
  - internal/generator/settings_test.go
  - internal/generator/stream.go
  - internal/pipeline/imports_test.go
  - internal/pipeline/pipeline.go
  - internal/review/live_generator_test.go
  - internal/review/open.go
  - internal/review/open_test.go
  - internal/review/review.go
  - testdata/generator/nightshift-stream.sse
- evidence:
  - baseline before edits: go test -count=1 ./... exit 0; every package with tests printed ok
  - red: go test -count=1 -timeout 30s -run TestLiveRefusesANonNVIDIARoutedModel ./internal/generator/ exit 1; live_test.go:97 text "ready"
  - tip: go build ./... && go vet ./... && go test -count=1 ./... exit 0
  - race: go test -race -count=1 -timeout 120s ./internal/generator/ ./internal/review/ ./internal/pipeline/ exit 0
  - go version go1.26.8; go.mod says go 1.26 and has no toolchain line
  - git log --format=%B 6affa00..HEAD | grep -ci co-authored printed 0 (grep exit 1, which is zero matches)
  - remark: the test summary prints ok for a package that skipped a test, so the skip count is UNMEASURED; PROVEN_TEST_DOCKER was not set
- note: master-was-wrong: none. The provisioned stream has meterless.routed_model and meterless.receipt_id on the last chunk, as measured. An earlier chunk also has meterless_deliberation.receipt_id; the reader ignores it. winning_route.provider is absent. No session recording was handed to this leg (ADR-0024). No testing-bar metric is declared, so that bar is UNRUN. Rulings: ReceiptID was added to review.ModelUse because the result had no field for the receipt id. GET /v1/models is implemented and tested and a review does not call it. A stream with no routed model is returned as model unknown, including when NVIDIA is required. meterless-scope is the contract value user. Redirects are not followed. An empty setting is the default, except the token, which is refused. The text printer in cmd/proven still does not print models.
- updated: 2026-10-22T06:38:40Z

CLOSE-OK · the-generator-through-nightshift · commit dc276c89a5167c65f7c05dcd12ae1d3fd6968125 · branch fix/the-generator-through-nightshift · worktree ████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
