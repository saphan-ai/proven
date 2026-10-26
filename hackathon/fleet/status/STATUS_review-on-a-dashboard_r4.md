# STATUS — review-on-a-dashboard · round 4

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: not recorded
- last commit: 140c8f448e1b32962b2e764c9f07a05b53e05eaf [review-on-a-dashboard] merge: absorb main 9b11366
- updated: 2026-10-24T12:57:35Z

## Summary

Proven emits OTLP (one trace per review, a span per agent and per reviewer model, span events, nine instruments), off unless an OTEL_* endpoint is set. The lab runs grafana/otel-lgtm with a provisioned 11-panel dashboard; make up prints its address, make trace queries the trace. go vet exit 0; go test: 40 packages ok, 1 FAIL (TestMaterializedTreesAreReadableByOthers, this seat's umask 0002; passes under 022). NOT run: make, docker, Grafana, the otel-lgtm tag (chosen from memory).

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — review-on-a-dashboard
- state: stop-2-ready
- abstract: Proven emits OTLP (one trace per review, a span per agent and per reviewer model, span events, nine instruments), off unless an OTEL_* endpoint is set. The lab runs grafana/otel-lgtm with a provisioned 11-panel dashboard; make up prints its address, make trace queries the trace. go vet exit 0; go test: 40 packages ok, 1 FAIL (TestMaterializedTreesAreReadableByOthers, this seat's umask 0002; passes under 022). NOT run: make, docker, Grafana, the otel-lgtm tag (chosen from memory).
- last commit: 140c8f448e1b32962b2e764c9f07a05b53e05eaf [review-on-a-dashboard] merge: absorb main 9b11366
- done: ADR-0045 built in four parts, committed, no push and no merge commit of mine
  - master, 2026-10-24: absorbed main 9b11366 (judges lab landed, second reviewer) at 140c8f4: go.mod/go.sum union; pipeline.go and review/open.go clean unions (spans wrap both planners and both skeptics); main's fix_scenario_test env guard allows OTEL_*; build, vet, all tests and check-links green
  - emission: internal/telemetry, the trace observer, the pipeline and review wiring (03ebea7, c09bd7e); proven by go test ./internal/review and ./internal/telemetry, 13 new tests
  - lab: lgtm service, versions.env pins, workflow rendering, make trace, dashboard JSON and its test (cfed420); proven by go test ./internal/telemetry and sh lab/██████████████████████████ (ok)
  - docs: docs/LAB.md section "The dashboard", docs/CONFIGURATION.md rows (fab29b5); proven by sh scripts/check-links.sh exit 0
  - public status copy: hackathon/fleet/status/STATUS_review-on-a-dashboard_r2.md (a45aed9)
- next: the master runs make clean, make up, make e2e, make trace on a seat with Docker and opens the dashboard; if the image tag, the Tempo data source name or a panel is wrong, see the Not run list in the public status
- artifacts:
  - hackathon/fleet/status/STATUS_review-on-a-dashboard_r2.md
  - internal/telemetry/telemetry.go
  - internal/telemetry/setup.go
  - lab/grafana/dashboards/proven-review.json
  - lab/████████████████
- evidence:
  - go vet ./... exit 0
  - go test -count=1 ./... 40 ok, 3 no test files, 1 FAIL internal/pipeline TestMaterializedTreesAreReadableByOthers (mode 664 files from umask 0002; umask 022 go test ./internal/pipeline -run that test: ok)
  - control, endpoint set: the stand-in collector is sent a trace and metrics; endpoint unset: 0 requests (TestAReviewIsExportedOnlyWhenAnEndpointIsSet)
  - mutation, observer wiring removed: TestReviewIsOneTraceWithAnAgentSpanAndTheTwoReviewers fails with no span event of kind claim; mutation, dashboard metric renamed: TestTheDashboardQueriesOnlyWhatProvenEmits fails naming the panel
  - sh scripts/check-links.sh exit 0; make check-links not run (make refused on this seat)
  - git rev-list --merges 54a41c4..HEAD prints nothing; no Co-authored-by in 54a41c4..HEAD; author Marcin Marzec for all five commits; 9303b2d..HEAD holds 22 merges that were under the branch before this round
  - not run: make clean, make up, make e2e, make trace, make down, Grafana rendering, the pull of grafana/otel-lgtm:0.11.0, the job's docker run --network host reaching the name lgtm
- note: master-was-wrong: — (nothing contradicted; the master's grep of no OTLP in Go code was right). Judgement calls, with their cost if wrong, are in the public status under Rulings: sample mode reads OTEL_ names (four tests allow the prefix), the lab review runs a second reviewer (LAB_SPECIAL_REVIEWER=qwen), anonymous Grafana as admin on 127.0.0.1, go.mod indirect markers hand-edited because go mod tidy cannot verify on this seat, --network host only in the seeded workflow.
- updated: 2026-10-24T11:45:00Z

CLOSE-OK · review-on-a-dashboard · commit a45aed90b8e8d0269af92246c68fd946ba60ee96 · branch fix/review-on-a-dashboard · worktree █████████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| file and line in Saphan Studio's private code | 2 |
