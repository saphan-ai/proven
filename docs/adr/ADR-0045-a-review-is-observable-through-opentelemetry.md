# ADR-0045: A review is observable through OpenTelemetry, and the lab shows it live on a dashboard

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0013](ADR-0013-no-hardcoded-settings.md) (settings), [ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md) and [ADR-0026](ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md) (the lab), [ADR-0037](ADR-0037-bring-your-own-second-reviewer.md) (two reviewers)
**Trigger:** the owner, 23 October: *"a nice dashboard in a separate container… show how the two models go, one already has its verdict and the other is still working… it has to be really nice… take a framework and only put the data in… Proven sends its data, and the dashboard exports Prometheus metrics."* Then: *"it can be plugged in wherever you want."*

## Context

A run in a forge's CI page shows a log. The interesting part of a Proven review, the agents and the two reviewers
working side by side, the verdicts, what was dropped, is not visible there. The fleet should not draw its own UI:
standard tools already do this well.

## Decision

1. **Proven emits OpenTelemetry (OTLP)**: a trace per review with one span per agent and per reviewer model, and
   metrics (findings by level of proof, claims dropped, reviewer verdicts and disagreements, model calls and their
   duration, guardrail checks and unsafe verdicts). Events of the trace file go along as span events.
2. **Configured only by the standard variables** (`OTEL_EXPORTER_OTLP_ENDPOINT` and the other `OTEL_*`); **off unless
   set**. Because it is the standard, it plugs into whatever an organisation already runs: Grafana, Datadog, New Relic,
   Honeycomb, Elastic, a plain Prometheus. Never a secret in an attribute (the redactor applies).
3. **The lab gets one more container, `grafana/otel-lgtm`** (OpenTelemetry collector, Prometheus, Tempo, Loki and
   Grafana in one image, pinned by version), with a **provisioned dashboard**: a state timeline of the agents and the
   two reviewers, verdicts side by side, disagreements, findings by level of proof, dropped claims, model latency and
   guardrail checks. The dashboard is a JSON file; nothing is drawn by hand.
4. **Prometheus metrics are exported** from the collector, so the numbers can be scraped by an existing monitoring
   stack.
5. **`make up` starts it** and prints its address; `docs/LAB.md` shows the dashboard during a pull-request review.
6. **Grafana runs as its own container** and is not part of Proven's code or image.

## Consequences

- A judge watches a review happen: which agent works, which model has answered, where the two disagree.
- An enterprise gets Proven's numbers in its own monitoring with no adapter to write.
- One more container in the lab; it binds to localhost like the rest.
