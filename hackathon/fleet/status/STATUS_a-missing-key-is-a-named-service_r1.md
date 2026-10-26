# STATUS — a-missing-key-is-a-named-service · round 1

- state: ready-for-review
- executor: cursor on runner-cursor
- model: grok-4.7-high
- pool: not recorded
- cost: not recorded
- duration: 27 min
- last commit: 53a0d964ac1edcd99e3403480d91e661ce373777 [a-missing-key-is-a-named-service] review: empty secret is down (the leg tip 04f111cbc8ef24ef8da2c748094af7133abb4ea4 with the tool trailer removed by the master; tree identical)
- updated: 2026-10-23T12:10:13Z

## Summary

Empty secrets name nightshift, zetaris and the guardrail and the review continues; an empty guardrail key refuses every model call, and sample mode is shown to dial nothing.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — a-missing-key-is-a-named-service
- state: stop-2-ready
- abstract: Empty secrets name nightshift, zetaris and the guardrail and the review continues; an empty guardrail key refuses every model call, and sample mode is shown to dial nothing.
- last commit: 53a0d964ac1edcd99e3403480d91e661ce373777 [a-missing-key-is-a-named-service] review: empty secret is down (the leg tip 04f111cbc8ef24ef8da2c748094af7133abb4ea4 with the tool trailer removed by the master; tree identical)
- done: Empty secrets name their services down, a down guardrail refuses every model call, and sample mode opens no connection.
  - no secrets, live review of a scratch git repo with a tracked symlink completes and names nightshift, zetaris and the guardrail, and names linked.txt — TestLiveReviewWithEmptySecretsNamesDownServices
  - empty guardrail key, other secrets set, stub generator call count is 0, the refusal is in the trace, guardrail is in failed services — TestEmptyGuardrailKeyRefusesEveryModelCall
  - PROVEN_GUARDRAIL_MODE=off with an empty key lets model calls through and the trace says no guardrail — TestGuardrailOffWithEmptyKeyLetsModelCallsThrough
  - a line break in each secret is still a configuration refusal before the review — TestWrongSecretsAreRefusedBeforeTheReview
  - sample review builds no live client and trips on a dial through the default transport or resolver — TestSampleReviewDialsNothing
- next / blocked-on: none
- artifacts:
  - internal/generator/absent.go
  - internal/generator/generator.go
  - internal/generator/live.go
  - internal/generator/settings.go
  - internal/generator/settings_test.go
  - internal/guardrail/absent.go
  - internal/guardrail/guardrail.go
  - internal/guardrail/live.go
  - internal/guardrail/settings.go
  - internal/guardrail/settings_test.go
  - internal/pipeline/pipeline.go
  - internal/query/absent.go
  - internal/query/live.go
  - internal/query/live_test.go
  - internal/query/query.go
  - internal/query/settings.go
  - internal/review/missing_secret_test.go
  - internal/review/open.go
  - internal/review/sample_nodial_test.go
  - internal/settingspage/catalog.go
  - docs/CONFIGURATION.md
  - docs/INSTALL.md
  - demo/live/README.md
- evidence:
  - before the fix, go test -count=1 -run TestLiveReviewWithEmptySecretsNamesDownServices ./internal/review/ exit 1: configuration refused: ███████████████████████=""
  - go build ./... exit 0
  - go vet ./... exit 0
  - go test -count=1 ./... exit 0; go 1.26.8, go.mod says go 1.26; no Makefile and no CI workflow in the tree
  - go test -race -count=1 ./internal/generator/ ./internal/query/ ./internal/guardrail/ ./internal/review/ ./internal/pipeline/ exit 0
  - negative control, empty guardrail key treated as off: TestEmptyGuardrailKeyRefusesEveryModelCall exit 1, generator calls 4; restored
  - negative control, one http.Get on the sample path: TestSampleReviewDialsNothing exit 1, sample review dialed; restored
- note: master-was-wrong: no. The tool appended Co-authored-by to 04f111c; the master removes it. One trace line per refused call. An empty ZETARIS_REST_URL or ZETARIS_ORG_ID is still a refusal, because neither is a secret. No testing bar is declared, so that metric is UNRUN. No session recording was given to this leg.
- updated: 2026-10-23T11:38:00Z

CLOSE-OK · a-missing-key-is-a-named-service · commit 04f111cbc8ef24ef8da2c748094af7133abb4ea4 · branch fix/a-missing-key-is-a-named-service · worktree ████████████████████████████████████████████████████████████████████

```owner-block-v1
version: 1
stream: a-missing-key-is-a-named-service
elements:
  - kind: none
    abstract: Nothing on your hand
```

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| private credential or variable name | 1 |
