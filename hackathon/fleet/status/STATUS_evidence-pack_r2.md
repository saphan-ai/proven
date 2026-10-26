# STATUS — evidence-pack · round 2

- state: ready-for-review
- executor: claude-code on macm5
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 6 min
- last commit: 24cfc28b4f175f62aece3fccfc22f7581aebafeb [evidence-pack] cli: resolve pack commits through internal/change
- updated: 2026-10-24T05:26:50Z

## Summary

Round 2: internal/cli/pack.go no longer imports os/exec. The commit is resolved by a new exported change.ResolveCommit (internal/change, git runs with childenv.Minimal and --end-of-options). TestInternalAndCmdDoNotImportBannedPackages was red at 1e83a0a (reproduced), now passes, no exemption added. Whole module: 35 ok, 6 FAIL, all `httptest: failed to listen on a port` sandbox panics (same 6 packages as base); vet and gofmt clean.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — evidence-pack
- state: stop-2-ready
- abstract: Round 2: internal/cli/pack.go no longer imports os/exec. The commit is resolved by a new exported change.ResolveCommit (internal/change, git runs with childenv.Minimal and --end-of-options). TestInternalAndCmdDoNotImportBannedPackages was red at 1e83a0a (reproduced), now passes, no exemption added. Whole module: 35 ok, 6 FAIL, all `httptest: failed to listen on a port` sandbox panics (same 6 packages as base); vet and gofmt clean.
- last commit: 24cfc28b4f175f62aece3fccfc22f7581aebafeb [evidence-pack] cli: resolve pack commits through internal/change
- done: banned-import red fixed; behaviour unchanged
  - reproduced first: `go test ./internal/pipeline/ -run TestInternalAndCmdDoNotImportBannedPackages` FAIL "internal/cli/pack.go imports os/exec"; after the change the same command is ok
  - `go test -count=1 ./internal/change ./internal/evidence ./internal/cli` all ok; `go vet ./...` clean; `gofmt -l .` empty
  - whole module `go test -count=1 ./...` capture at ████████████████████████████████████ 35 ok, 6 FAIL (cmd/livecheck, internal/generator, internal/guardrail, internal/pipeline, internal/query, internal/review), each a listen-permission panic in a different test (TestAllServicesReady, TestLiveReadsTheProvisionedStream, TestServerPostsTheMeasuredRequest, TestUnsafePromptStopsTheStepAndNamesCategories, TestLiveSelectReadsRowsFromTheSQLEditor, TestLiveReviewCarriesTheRoutedModel)
- next / blocked-on: review (named); a hand that may delete untracked █████████████████████████████
- artifacts:
  - internal/change/between.go
  - internal/cli/pack.go
- evidence:
  - the module is NOT fully green here: the 6 packages abort on a sandbox listen refusal, so their other tests (including the pipeline package's own run) are UNRUN in the full run; the import test itself was run alone and passes. The master's own unsandboxed run is the whole-module instrument.
  - ResolveCommit adds `--end-of-options` and the minimal environment, which the old call lacked; the resolved value for a real commit is unchanged. No new unit test for ResolveCommit was added (covered through the cli pack tests).
- note: master-was-wrong: none found, the red reproduced as pasted. // NOTICED, not acted on: untracked zz_probe_test.go still present (the hook refuses rm); packed-refs.lock Operation not permitted printed after the commit, which landed.
- updated: 2026-10-24T06:10:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 1 |
