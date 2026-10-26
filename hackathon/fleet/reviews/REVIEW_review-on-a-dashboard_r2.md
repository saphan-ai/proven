# REVIEW — review-on-a-dashboard · round 2

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 1 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T01:08:50Z

## Summary

Prior unredacted status finding addressed; focused Go tests pass. Live Docker lab is unrun on this seat and requires independent verification before a live-demo claim.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — review-on-a-dashboard

### Ground

HEAD `c540b3b` (measured by `git rev-parse --short=7 HEAD`). Comparison base `db9f7c2` (measured by `git merge-base origin/main HEAD`). Second review of the span-status redaction finding at `140c8f4`; review only, no product edits. The master's lab measurements in the order are claims, not results of this review.

### Differential

- ADDRESSED: `internal/pipeline/pipeline.go:182` installs the configured-secret redactor on telemetry before a review starts. `internal/telemetry/telemetry.go:78`, `:242`, `:305` redact root and agent error statuses; `:397` uses constant text for model-call failures. The prior code sent raw `err.Error()` to root and agent statuses. `internal/review/telemetry_status_test.go:125` and `:144` exercise failing review and degraded service, with secret-bearing errors and a redaction-marker control. Both passed.
- Off/on: `internal/telemetry/setup.go:78` returns nil without an endpoint; `internal/review/open.go:350` runs without telemetry in that state. `internal/review/telemetry_status_test.go:93` compares full sample results with the exporter on and off and passed. It does not compare executable exit codes.
- Both reviewer planners and skeptics have distinct spans at `internal/pipeline/pipeline.go:561`, `:597`, `:724`, `:760`; shared agents also have spans. `internal/telemetry/telemetry.go:284` sets reviewer and agent attributes; `:322` emits their state. `internal/telemetry/dashboard_test.go:29` checks queried metric names, not rendered panels.
- Lab: `lab/versions.env:40` pins `grafana/otel-lgtm:0.11.0`; `lab/compose.yaml:61` binds ports through `LAB_BIND`. `lab/█████████████` prints the dashboard address; `docs/LAB.md` and `docs/CONFIGURATION.md` name the settings. The added spans wrap the original default/special calls, and `internal/review/open.go:350` returns the review's original result/error after bounded flushing. This is source inspection, not a Docker observation.

### Unplug

At `140c8f4`, both status setters used raw `err.Error()`. The new failed-error controls at `internal/review/telemetry_status_test.go:125` and `:144` put a configured secret in the error and assert the status contains only its marker. Reverting those setters would fail the assertions. I did not execute a mutant or the old revision, so historical RED is unmeasured. The lab negative control `make e2e-failed-run` is also unrun.

### Closure

The earlier BLOCKING status-description leak is ADDRESSED. No further blocking defect established in this re-review scope. Verdict is `accept-with-conditions`, not `accept`, because the live lab checks are unrun. No code, threshold or exception change is proposed.

### Remainder

- NOTE: `internal/review/telemetry_status_test.go:93` compares result objects but not CLI exit codes; the master may require process-level parity proof. No difference was measured in review results.
- Declined to judge: live Grafana panel appearance, production collector availability, image licensing (owner decision), and Docker cleanup; this seat has no Docker executable.
- No numeric coverage, mutation, lint or type-warning metric and baseline was declared in the supplied order. Testing-bar ranking: suite-owned metric-name test is most circular; Go vet is a toolchain check; an actual running collector/dashboard is least circular and remains unrun. No numeric baseline or delta is claimed.

### Profile

Security (secret disclosure and collector boundary), backend (telemetry I/O and failure isolation), UI (Grafana), docs. `go.mod` pins Go 1.26.0 and OpenTelemetry SDK v1.47.0; measured seat version is Go 1.26.5. `go test -count=1 ./internal/telemetry ./internal/review ./internal/pipeline` passed all three packages. Focused status, off/on and dashboard tests, `go vet ./internal/telemetry ./internal/review ./internal/pipeline`, `make check-links`, and `git diff --check origin/main...HEAD` passed. These are not a full-suite or live-lab claim.

### Focus

The prior finding was exported status descriptions bypassing the configured-secret redactor. The revision routes failed review and failed agent errors through it, with failed-call controls. Static dashboard query matching does not prove the panels render.

### Conditions

No new BLOCKING or OWED defect in the re-review scope. Before a live-demo claim or gate, the master must verify `make up`, `make e2e`, `make trace`, `make e2e-failed-run`, `make down`, and `make clean` on a Docker-capable seat and inspect the actual Grafana dashboard. A failed step needs a new classification; these static checks cannot stand in for a live pass.

### Unrun

- The six Docker-backed make targets in Conditions: Docker command absent on this seat. The master's reported run was not re-measured. The false-world answer for the static dashboard test is an unknown instrument name, not a blank live panel; rendering is UNMEASURED.
- Full `go test ./...`, a mutant/old-revision RED, and base-vs-tip numeric quality comparison: not executed; the supplied order declares no numeric quality baseline.

## The review status as filed

## REVIEW — review-on-a-dashboard
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Prior unredacted status finding addressed; focused Go tests pass. Live Docker lab is unrun on this seat and requires independent verification before a live-demo claim.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T01:04:05Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 1 |
