# Traces, metrics and diagnostics

The result answers “what did Proven keep, and what coverage is missing?” The trace answers “what happened during this run?” Optional OpenTelemetry export adds remote monitoring. Observability does not change adjudication, and a failed export does not change a review's verdict.

## Local JSONL trace

The CLI prints the trace path. By default it is under `<user cache>/proven/logs/trace.jsonl`; `PROVEN_LOG_DIR` selects an external directory. The same directory can hold H-MEM and World Model state. Keep it outside the reviewed repository and restrict access. See [Operations](OPERATIONS.md).

Each line is a JSON event. Inspect the event's `kind` and stage/reviewer context rather than depending on its display text. The file is diagnostic history, not a substitute for a signed evidence pack. Preserve a run's trace with its report before sharing or rotating it.

| Event or information | Use |
| --- | --- |
| Agent/stage events | Determine what ran, what was skipped and where execution stopped. |
| `model` | Successful logical model use and answering identity when known. |
| `model-attempt` | Individual successful/failed HTTP requests, timings, status, finish reason and output/reasoning counts. |
| Guardrail event | Classification result, backend and latency; explicit disabled state when off. |
| Query/test/tool evidence | Connect a candidate to observed rows or controlled execution. |
| `unresolved-hypothesis` | A final pending planning description with reviewer role, independent of finding IDs. |
| Dropped/refuted candidate | Understand why a proposed concern did not become a finding. |

For model performance, compare requested/reported model, attempt duration, time to headers, first body byte, HTTP status, response ID, `finish_reason`, answer/reasoning byte counts and available token usage. Missing provider counters remain unknown. Successful logical calls and HTTP attempts are different counts because failures, retry and recovery can add requests.

### Inspect without extra tools

This prints the structured model-attempt objects from a selected trace. Substitute the actual path printed by your run:

```sh
python3 - /path/to/trace.jsonl <<'PY'
import json, sys
with open(sys.argv[1]) as trace:
    for line in trace:
        event = json.loads(line)
        if event.get("kind") == "model-attempt":
            print(json.dumps(event, sort_keys=True))
PY
```

Do not publish the full trace without inspecting its source excerpts and model text. Known configured secrets are redacted, but source code can contain unrelated confidential information.

## Reasoning capture

Reasoning byte counts are retained without reasoning text by default. `PROVEN_GENERATOR_TRACE_REASONING_BYTES` permits a bounded amount of text per attempt; the second reviewer has an independent/inherited override. Text passes the same configured-secret redactor as other trace fields.

Capture only what is needed to diagnose a specific model. It increases retained sensitive content and is not an extra evidence class. Reasoning is never used as final agent JSON merely because the answer was empty.

## OpenTelemetry

Export is off when no endpoint is configured. To send traces and metrics to a local OTLP/HTTP collector:

```sh
export OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
export OTEL_SERVICE_NAME=proven
```

The base endpoint receives `/v1/traces` and `/v1/metrics`. Signal-specific `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` and `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` are full URLs and override the base for that signal. `OTEL_SDK_DISABLED=true` disables export. Only HTTP/protobuf is built in; a configured gRPC protocol is refused by name. The SDK reads its standard header, certificate and timeout settings.

One review is one trace, with agent/model spans and bounded local-trace events. Each process has a service instance ID, so counters restarting at zero are not presented as one continuous series. Sample mode also reads the `OTEL_*` variables; unset/disable them for a completely disconnected sample.

| Instrument | Measures |
| --- | --- |
| `proven.review.findings` | Kept findings by proof level. |
| `proven.review.claims.dropped` | Candidates dropped by reason. |
| `proven.reviewer.verdicts` | Skeptic verdicts by reviewer. |
| `proven.reviewer.disagreements` | Reviewer disagreements. |
| `proven.model.calls` | Logical model calls by agent, reviewer and outcome. |
| `gen_ai.client.operation.duration` | Model-call duration in seconds. |
| `proven.guardrail.checks` / `proven.guardrail.unsafe` | Guardrail activity and unsafe verdicts. |
| `proven.agent.state` | Agent state: 0 waiting, 1 working, 2 done, 3 failed. |

Build dashboards around completion, service failures, dropped claims, latency and evidence mix as well as findings. A lower finding count alone is not a quality improvement. Logical-call metrics do not replace `model-attempt` events when diagnosing HTTP retries.

## Local network policy

`PROVEN_NETWORK=local` checks collector endpoints at configuration and connection time, including sample mode. Public destinations are refused even if source/model processing would otherwise stay local. Configure an allowed internal collector or disable export. See [Air-gapped operation](AIR_GAPPED.md).

## Guardrail accounting

The included helper groups structured guardrail events by UTC day, side, verdict and backend, with median latency:

```sh
go run ./cmd/guardcount /path/to/trace.jsonl
```

It needs a Go checkout, not just the Proven binary. With the guardrail off, there are no classification checks to count. The [configuration reference](CONFIGURATION.md#counting-guardrail-checks) documents structured fields and older-trace compatibility.

## A useful incident bundle

Retain the result and trace together with base/head, Proven revision, sanitized limits/profile, source snapshot identities and runner image. A signed pack adds offline integrity and decision checks; it does not establish that the recorded execution actually occurred. See [Evidence packs](EVIDENCE_PACK.md) and [Troubleshooting](TROUBLESHOOTING.md).
