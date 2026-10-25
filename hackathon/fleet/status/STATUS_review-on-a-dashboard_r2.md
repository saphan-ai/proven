# STATUS — review-on-a-dashboard · round 2

- state: ready-for-review
- executor: Claude Code executor
- model: Sonnet 5.5 (claude-sonnet-5-5)
- pool: not shown to this leg
- cost: about USD 4.03 of the USD 22 cap, as the session budget line showed before this file was written
- last commit: see `git log -1` on the branch; the four code and doc commits are `03ebea7`, `c09bd7e`, `cfed420`, `fab29b5`
- updated: 2026-10-24

## Summary

Proven now emits OpenTelemetry (OTLP over HTTP): one trace per review, one span per agent and per reviewer model, the trace file's events as span events, and nine instruments. It is off unless a standard `OTEL_*` endpoint is set. The lab gets `grafana/otel-lgtm` as one more service with a dashboard provisioned from a JSON file; the seeded workflow sends the lab review to it with two reviewer models. `go vet` and `go test` pass except one test that fails on this seat's umask and passes under 022. `make`, `docker` and the dashboard itself were not run here.

## What was built

1. **`internal/telemetry`.** A nil `*T` is off, and every call on it does nothing. `Setup` reads only `OTEL_*`: `OTEL_EXPORTER_OTLP_ENDPOINT` (and the traces and metrics variants), `OTEL_SDK_DISABLED`, the protocol variables (only `http/protobuf` is built in; anything else is refused by name), `OTEL_SERVICE_NAME`, `OTEL_METRIC_EXPORT_INTERVAL`. Each process gets its own `service.instance.id`.
2. **Spans.** `proven.review` is the root. Each agent is `agent <name>` with `proven.agent` and `proven.reviewer` (`default`, `special`, `shared`). A model call is a `chat <model>` client span under its agent, with the gen_ai attributes. The trace writer calls an observer after redaction, so each trace line is a span event and no secret reaches an attribute.
3. **Instruments.** `proven.review.findings` (by level of proof), `proven.review.claims.dropped` (by reason), `proven.reviewer.verdicts`, `proven.reviewer.disagreements`, `proven.model.calls`, `gen_ai.client.operation.duration`, `proven.guardrail.checks`, `proven.guardrail.unsafe`, and the gauge `proven.agent.state` that feeds the state timeline.
4. **Lab.** Service `lgtm` in `lab/compose.yaml`, pinned in `lab/versions.env`; Grafana on `127.0.0.1:3001`, OTLP/HTTP on `127.0.0.1:4318`. `make up` starts it, waits for Grafana and prints the dashboard address. `make down` and `make clean` remove it. `make trace` asks Tempo for the latest review's trace and exits 1 when a span or a reviewer is missing. The dashboard is `lab/grafana/dashboards/proven-review.json` (11 panels), provisioned by `lab/grafana/proven-dashboards.yaml`.
5. **Docs.** `docs/LAB.md` has the section "The dashboard": how to open it during a pull-request review and what each panel shows. `docs/CONFIGURATION.md` lists the new lab settings, and its generated table lists the `OTEL_*` variables.

## Acceptance, item by item

1. `go vet ./...`: exit 0. `go test -count=1 ./...`: 40 packages `ok`, 3 with no test files, 1 FAIL: `internal/pipeline` `TestMaterializedTreesAreReadableByOthers`. The failure is the seat's umask (`0002`, so files are mode 664). `umask 022; go test ./internal/pipeline -run TestMaterializedTreesAreReadableByOthers` passes. The test does not touch telemetry. I did not run the suite on the base to show it fails there too.
2. New tests, all passing, in `internal/review/telemetry_test.go` and `internal/telemetry`: `TestReviewIsOneTraceWithAnAgentSpanAndTheTwoReviewers`, `TestReviewCountsWhatADashboardShows`, `TestDroppedClaimsAreCounted`, `TestNoSecretReachesASpanOrANumber`, `TestAReviewIsExportedOnlyWhenAnEndpointIsSet`, `TestADisabledSDKExportsNothing`, `TestABadTelemetrySettingIsRefusedByName`, `TestANilTelemetryIsOff`, `TestTheTimelineShowsEveryAgentFromTheStart`, `TestAStageLeftOpenIsFailedWhenTheReviewEnds`, `TestADropReasonIsAClosedSet`, `TestSetupIsOffWithoutAnEndpoint`, `TestTheDashboardQueriesOnlyWhatProvenEmits`.
3. **Unset exports nothing, with its control.** `TestAReviewIsExportedOnlyWhenAnEndpointIsSet` runs the same sample review twice against a local stand-in collector. With the endpoint set it is sent both a trace and metrics (the control). With it unset it is sent 0 requests. Mutation: with the observer wiring removed, `TestReviewIsOneTraceWithAnAgentSpanAndTheTwoReviewers` fails with `no span event of kind "claim"`. Mutation: a metric renamed in the dashboard makes `TestTheDashboardQueriesOnlyWhatProvenEmits` fail naming the panel.
4. `sh scripts/check-links.sh`: exit 0. `make check-links` itself was not run: `make` is refused on this seat.
5. `git rev-list --merges 54a41c4..HEAD` prints nothing, and no commit of mine has a `Co-authored-by` line. The author of the four commits is the fleet identity. `git rev-list --merges 9303b2d..HEAD` prints 22 commits: those are the judges-lab stream's merges that were already under this branch, not mine.
6. `sh lab/scripts/e2e-parse-check.sh`: `e2e-parse-check: ok`. `sh -n` on `up.sh`, `trace.sh`, `status.sh`, `help.sh`, `seed.sh`: exit 0, syntax only. The seeded workflow's `sed` was run on `lab/workflow/proven.yml` by hand: the `docker run` line came out with `--network host`, the three `OTEL_*` variables and the `--config lab.yaml` step. `proven scan --sample --config <file with special_reviewer: qwen>` prints the same four findings as without the file, plus the line `reviewer qwen; role special`.

## Not run

- `make clean && make up`, `make e2e`, `make trace`, `make down`: `make` and `docker` are refused on this seat. The master runs them. So the dashboard's trace query and its answer (span names, two reviewer spans) are not shown here. `make trace` is the command that shows them.
- Grafana itself. The dashboard JSON parses (`python3 -I -c "import json; json.load(...)"`) and its metric names are held to the instruments by a test, but no Grafana rendered it. Not checked against a real Grafana: the panel JSON, the state-timeline value mappings, the TraceQL table query, the PromQL names an OTLP-to-Prometheus path gives (`_total`, `_seconds_sum`).
- The image tag `grafana/otel-lgtm:0.11.0` was not pulled. The seat had no network, and WebFetch was refused. I chose the tag from memory. The first `make up` is the check; if the tag does not exist, set `LGTM_VERSION`.
- Facts about the `otel-lgtm` image that I wrote from memory and did not read: it serves Grafana on 3000 and OTLP/HTTP on 4318; Grafana reads extra provider files from `/otel-lgtm/grafana/conf/provisioning/dashboards/`; the data source is named `Tempo`; `/data` exists. If a panel is empty or `make trace` says there is no data source named Tempo, look here first.
- That the job's `docker run --network host` resolves the name `lgtm` through the inner daemon's DNS: reasoned, not run.

## Rulings

- **Sample mode reads the `OTEL_*` variables.** The lab's pull-request review is `proven scan --sample`, so the dashboard shows nothing unless sample mode may export. Four existing tests listed the keys sample mode may read, and each now also allows the `OTEL_` prefix (`cmd/proven/main_test.go`, `cmd/proven/scenario_test.go`, `internal/review/open_test.go`, `internal/cli/run_test.go`). No test was loosened for any other key. Cost if wrong: the master removes the prefix and the lab review exports nothing.
- **The lab review runs with a second reviewer** (`LAB_SPECIAL_REVIEWER=qwen`, written into a `proven.yaml` the job creates). Without it there is one reviewer and nothing to show side by side. The sample generator answers for both, and the findings did not change. Cost if wrong: set `LAB_SPECIAL_REVIEWER=` empty.
- **Anonymous Grafana as admin** (`GF_AUTH_ANONYMOUS_*`). `make trace` reads through it and a judge is not asked to sign in. The lab binds to 127.0.0.1; `docs/LAB.md` says what changes if `LAB_BIND` does not.
- **`go.mod`:** I removed `// indirect` from the seven modules the code now imports by hand. `go mod tidy` could not run: it verifies against the checksum database and the seat cannot write that cache.
- **`--network host` on the lab job's `docker run`**, only in the seeded copy of the workflow. It joins the inner daemon's network, which is on the lab network. The committed workflow is unchanged.

## master-was-wrong

— Nothing in the order or in ADR-0045 was contradicted. The master's grep was right: before this change `git grep -niE 'otlp|opentelemetry' -- '*.go'` found nothing.
