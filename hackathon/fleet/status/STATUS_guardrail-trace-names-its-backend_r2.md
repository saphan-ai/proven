# STATUS — guardrail-trace-names-its-backend · round 2

- state: ready-for-review
- executor: claude-code
- model: claude-opus-5-5
- pool: not recorded
- cost: not recorded
- duration: 16 min
- last commit: 9cef733b1da90787cbd23ecce1c4d7e1539d6d8a [guardrail-trace-names-its-backend] guardcount: count checks per day
- updated: 2026-10-25T00:36:20Z

## Summary

Each guardrail check on the trace now names backend (server, nvidia, sample, sample-recorded, absent, off) and model; a check that could not be made carries its side and a closed down value (unknown errors land as "other"); text unchanged.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — guardrail-trace-names-its-backend
- state: stop-2-ready
- abstract: Each guardrail check on the trace now names backend (server, nvidia, sample, sample-recorded, absent, off) and model; a check that could not be made carries its side and a closed down value (unknown errors land as "other"); text unchanged.
  New command cmd/guardcount counts checks per UTC day, side, verdict, backend with median latency; golden test plus a test holding docs/CONFIGURATION.md to the golden output.
  Master was wrong once: the down set was not closed (the empty-key reason and raw errors escaped). Build, vet, full suite green at 9cef733.
- last commit: 9cef733b1da90787cbd23ecce1c4d7e1539d6d8a [guardrail-trace-names-its-backend] guardcount: count checks per day
- done: trace.Check gains backend, model and down; guardcount counts them; docs section and tests landed in two commits (21f741b, 9cef733)
  - trace fields + Wrap/Identity + closed down set: internal/pipeline/guardrail_backend_test.go (8 tests), internal/guardrail/down_test.go (census + computed reasons)
  - instrument: cmd/guardcount/main_test.go TestCountsMatchTheGoldenFile against █████████████████████████████████lden
  - docs: docs/CONFIGURATION.md "Counting guardrail checks", held to the golden file by TestDocsShowTheGoldenOutput
  - no secret: TestServerCheckNamesServerAndItsModel and TestNVIDIACheckNamesNVIDIAAndItsModel grep the written trace for the configured key
- next / blocked-on: master verification, then the owner's gate and merge
- artifacts:
  - internal/trace/trace.go
  - internal/trace/redact.go
  - internal/guardrail/guardrail.go
  - internal/guardrail/down.go
  - internal/guardrail/down_test.go
  - internal/guardrail/wrap.go
  - internal/guardrail/live.go
  - internal/guardrail/absent.go
  - internal/guardrail/sample.go
  - internal/pipeline/pipeline.go
  - internal/pipeline/guardrail_backend_test.go
  - cmd/guardcount/main.go
  - cmd/guardcount/main_test.go
  - cmd/guardcount/testdata/live.trace.jsonl
  - cmd/guardcount/testdata/sample.trace.jsonl
  - █████████████████████████████████lden
  - docs/CONFIGURATION.md
- evidence:
  - toolchain: go.mod pins go 1.26; local go1.26.10 linux/amd64 (measured, go version)
  - baseline on 64b5d0b before any edit: go build ./... && go vet ./... && go test -count=1 ./... exit 0, no FAIL line (measured)
  - step 0 red, compile: go test -count=1 ./internal/pipeline/ exit 1, "c.Backend undefined (type trace.Check has no field or method Backend)", same for Model and Down; nothing else failed to build
  - step 0 red, run time (fields added, no behaviour): go test -count=1 -run 'Backend|Down|Other|NamesServer|NamesNVIDIA|SampleCheck' ./internal/pipeline/ exit 1; all 8 new tests FAIL, e.g. "check 0 names backend \"\" model \"\", want \"server\" \"nemotron-3.5-content-safety\"" and "check (*trace.Check)(nil), want down \"timeout\" side \"prompt\"" (confirms MC-4: a down check had no Guardrail block)
  - negative control (law 27), mutated field trace.Check.Backend in guardEvent set to "": same command exit 1; FAIL TestServerCheckNamesServerAndItsModel, TestNVIDIACheckNamesNVIDIAAndItsModel, TestSampleCheckNamesTheSampleBackend, TestRecordedSampleCheckNamesItsBackend, TestRefusingAdaptersNameBackendAndDown, TestLiveUnauthorizedNamesBackendAndDown; restored, exit 0
  - unplug (law 41): TestGuardrailOffTraceHasNoRows runs a sample review with GuardOff; exactly one guardrail line, text "no guardrail", agent pipeline, no guardrail object; guardcount prints the header only, exit 0, empty stderr
  - tip 9cef733: go build ./... && go vet ./... && go test -count=1 ./... exit 0; 50 packages ok, 0 FAIL, 10 without test files (measured, grep of the saved output)
  - attribution: git log --format=%B db9f7c2..HEAD | grep -ci co-authored printed 0
  - UNRUN: proven scan --sample by hand (running the built binary is denied with allowed form none); the docs output comes from the golden test, not a hand run
- note: master-was-wrong: (1) MC-4 confirmed by the run-time red. (2) The down set was NOT closed: settings.go builds "████████████████████ is empty" with service.Down directly and review.Open passes it through guardrail.Absent, so it never went through the listed reasons; and Safe/Sample return the raw ctx error when limit.Bound's own deadline passes, which asDown passed through as text. The first is now down "no-key", the second lands as "other". (3) Observe: confirmed it carried neither backend nor model; smallest change taken is three fields on Verdict filled by Wrap plus an optional Identity interface, so neither the Adapter interface nor the Observe signature changed and no caller moved.
  ruling: down values are hyphenated tokens (not-configured, no-recorded-verdict, http-503) not the reason text with spaces — a space-split table keeps its columns — cost if wrong: a rename in down.go and the golden file.
  ruling: backend names are the mode for live adapters (server, nvidia, off) and sample, sample-recorded, absent for the others; an adapter without Identity prints unknown — cost if wrong: renames in the Identity methods and the docs table.
  ruling: http-<status> counts as closed only for 100..599; anything else is other — cost if wrong: one condition in downCode.
  ruling: guardcount counts an older down line with no guardrail object as down:unstructured with side and backend "-", rather than dropping it — cost if wrong: one branch in checkOf.
  ruling: the docs section sits after the generated settings block in docs/CONFIGURATION.md, because the block is regenerated from internal/settingspage and refuses hand prose (TestConfigurationPageMatchesTheRegistry went red when it sat inside) — cost if wrong: a move.
  ruling: the order's `- master-was-wrong:` key is not in the closed status key set, so it lives in this note.
  NOTICED: the worktree was forked at 64b5d0b, not at the order's base db9f7c2 (db9f7c2 is its ancestor). NOTICED: docs/SAMPLE_MODE.md shows a guardrail trace line for the unsafe-prompt scenario that now also carries "backend":"sample-recorded"; I could not run the scan to re-measure it, so it is left as written.
- updated: 2026-10-25T01:20:00Z

CLOSE-OK · guardrail-trace-names-its-backend · commit 9cef733b1da90787cbd23ecce1c4d7e1539d6d8a · branch fix/guardrail-trace-names-its-backend · worktree █████████████████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 2 |
| private credential or variable name | 1 |
