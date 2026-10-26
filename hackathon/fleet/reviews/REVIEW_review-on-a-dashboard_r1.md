# REVIEW — review-on-a-dashboard · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T13:24:55Z

## Summary

BLOCKING: failed-review and failed-agent errors bypass redaction into exported span status. Redact them and add a failed-call control. Targeted pipeline tests found an unrelated permission failure; Docker was unavailable.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — review-on-a-dashboard

### Ground
Reviewed `140c8f4` (`git rev-parse --short=7 HEAD` printed `140c8f4`) against `9b11366`, the merge's second parent. A local `main` ref is absent. Profile: security (exported diagnostics), backend, dashboard and docs. This review changed no tracked code.

### Differential
**BLOCKING — raw error text reaches exported spans.** `internal/telemetry/telemetry.go:219` sets the root span's status description to `err.Error()`; `internal/telemetry/telemetry.go:282` does the same for each failed agent. `internal/review/open.go:299-305` passes the raw error to the root, and `internal/pipeline/pipeline.go:569-571` closes the planner stage with the raw error before `state.soften` handles it. An adapter error containing a configured secret bypasses `internal/trace/trace.go:206` (`redactEvent`) and `internal/pipeline/pipeline.go:1434` (failure-text redaction). With OTLP enabled, the status description is exported. Remedy: redact those descriptions with the review's redactor, or emit fixed non-sensitive codes; test both failed agent and failed review with a configured secret and inspect the exported statuses. `TestNoSecretReachesASpanOrANumber` (`internal/review/telemetry_test.go:275`) exercises successful attributes, events and metrics, but does not inspect status descriptions or a failed call.

Off by default: `internal/telemetry/setup.go:78-95` returns nil before provider construction without an endpoint, or if the SDK is disabled; `internal/review/open.go:293-295` then bypasses telemetry. `TestAReviewIsExportedOnlyWhenAnEndpointIsSet` and `TestADisabledSDKExportsNothing` assert no collector requests in the off cases. No test compares the complete findings and exit status with telemetry on/off (OWED). A malformed telemetry setting deliberately refuses `Open` (`internal/telemetry/setup.go:96-107`), so the unqualified “never changes what a review decides” claim in `docs/CONFIGURATION.md:107` needs a configuration caveat.

`internal/pipeline/pipeline.go:569-881` starts distinct default/special planner and skeptic spans and shared investigator, prover, judge and writer spans. `internal/telemetry/telemetry.go:330-381` gives each model call its own client span and attributes. `internal/telemetry/dashboard_test.go:29` checks dashboard query names against emitted metrics. `lab/versions.env` pins `grafana/otel-lgtm:0.11.0`; `lab/compose.yaml:49-67` uses configured bind and ports and provisions the dashboard; `lab/████████████████` builds the printed URL. `lab/███████████████` and `lab/████████████████` use the compose project. `docs/LAB.md` and `docs/CONFIGURATION.md` describe the wiring. The merge's second-parent diff retains prior review decisions and adds instrumentation; no unrelated decision branch removal was observed. The new code is readable; readability alone does not block.

### Unplug
The configured-collector test has a false-world answer: disconnecting export leaves the trace/metric request counters zero and fails the test. The secret test's false-world answer covers a successful response model and trace event, not a failed status. Add the failure-status test, unplug the new redaction, and require it to turn red. Add an on/off result equality test to detect future result coupling.

### Closure
Verdict **reject** until the exported-error leak is repaired and a failed-call negative control passes. The span and dashboard structure do not cure exposed diagnostics. The master's earlier lab readings are claims, not measurements on this seat.

### Remainder
OWED: parity test for identical findings and exit code with the exporter on/off; qualify the invalid-config refusal in docs. No weakening of a threshold, assertion, skip, or suppression was observed in the telemetry-specific diff. No accepted exception was found. Declined to judge: image licensing (excluded by the order), actual visual panel rendering (Docker absent).

### Profile
Trust boundary: agent/service errors and configured secrets enter in-process telemetry and leave through OTLP. The trace writer redacts events, but the raw span status bypasses that control. Dependency delta: `go.mod` adds OpenTelemetry SDK/exporters v1.47.0 and transitive transport packages; this review did not independently audit upstream code or licenses. Check ranking: offline dashboard-query consistency is a project check; in-memory spans and HTTP collector are suite checks; the live dashboard is UNRUN. No numeric performance, coverage or mutation baseline was declared for this wagon.

### Focus
1. `internal/telemetry/telemetry.go:219,282`: raw error status, BLOCKING; redact or classify before OTLP, test failed paths.
2. `internal/review/telemetry_test.go:386`: on/off collector test does not compare findings and exit status, OWED; compare both results.
3. `docs/CONFIGURATION.md:107`: unqualified observation claim omits invalid-setting refusal, OWED; qualify it.

### Conditions
No downgrade of the BLOCKING finding is recommended. The next review should verify root and agent status descriptions with a configured secret and compare on/off result differentials. Ruling: error status disclosure is blocking because diagnostics cross the collector boundary; if wrong, the cost is an extra repair round.

### Unrun
Docker-based `make up`, `make e2e`, `make trace`, `make down`, `make clean`, and visual dashboard inspection: no Docker executable on this seat (`command -v docker` printed no path). Full `go test ./...`, links, race detector, and baseline tests: not run in this focused review. Targeted `go test ./internal/telemetry ./internal/review ./internal/pipeline -count=1`: telemetry and review passed; pipeline failed in preexisting `TestMaterializedTreesAreReadableByOthers` because temporary `go.mod` files were mode 664, unrelated to telemetry. `go vet ./internal/telemetry ./internal/review` passed. Master's earlier lab report was not rerun here.

## The review status as filed

## REVIEW — review-on-a-dashboard
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: BLOCKING: failed-review and failed-agent errors bypass redaction into exported span status. Redact them and add a failed-call control. Targeted pipeline tests found an unrelated permission failure; Docker was unavailable.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T13:16:55Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 3 |
