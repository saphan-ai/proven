# BUG REPORT to NVIDIA — `nvidia/nemotron-3.5-super-vl-120b-a12b` returns empty responses and 500s under steady load

- from: the proven team (Open Agent Hackathon 2026), written by the infrastructure master; to be sent by the owner
- date: 22 October 2026
- product: NVIDIA's hosted API, model `nvidia/nemotron-3.5-super-vl-120b-a12b`, chat completions, streaming
- reached through: Meterless Nightshift (an OpenAI-compatible gateway) running on our machine, which calls NVIDIA's hosted
  API with our NVIDIA key and signs one receipt per call

## Summary

During two code reviews running at the same time, about one call in three to `nemotron-3.5-super-vl-120b-a12b` failed:
the model returned an **empty response**, or NVIDIA answered **HTTP 500 "Failed to generate completions"**, or the first
byte did not arrive in time. The successful calls in the same minutes were normal (2–33 s, up to 101,685 input tokens),
so the failures are not explained by the size of the request.

## What we measured

Nightshift's own signed receipts, 22 October 2026, 04:40–05:05 UTC, one application (`proven-review`), one model:

| outcome | calls |
|---|---|
| ok | 63 |
| failed: empty response from the model (`…_empty_response`) | 19 |
| failed: aborted by the client after the provider did not answer (`This operation was aborted`) | 8 |
| failed: NVIDIA HTTP 500 `{"message":"Failed to generate completions","type":"Internal Server Error","code":500}` | 3 |
| failed: time to first byte exceeded (`ttfb_timeout`) | 1 |
| **total** | **94** (31 failed, 33 %) |

- Successful calls in the same window: input 21,530–101,685 tokens; output 105–2,004 tokens; latency 1,996–33,359 ms;
  48–128 output tokens per second.
- Failures were interleaved with successes of similar size within seconds (for example 04:48:28 empty response, 04:48:29
  ok with 33,956 input tokens; 04:49:19 and 04:49:23 empty responses, 04:49:27 ok with 72,414 input tokens).
- The first 500 was at 04:51:45 UTC.
- Load: two agent sessions sending requests in turn, at most two requests in flight from our side.

## What we expected

A non-empty completion or a clear, retryable error with a status code (for example 429 or 503 with `Retry-After`), not an
empty body with a success status.

## What it costs us

Our review agent stops after a few failed retries, so a review of several minutes is lost and has to be run again;
on 22 October two reviews were lost this way.

## Reproduction

A coding agent (qwen-code) in a loop of chat completions with tool use, streaming, 20k–100k input tokens per call,
calls a few seconds apart, two such loops at once, against `nvidia/nemotron-3.5-super-vl-120b-a12b`. We can share the
receipt ids and timestamps of every failed call on request.

## Not measured

- The raw upstream response bodies of the empty responses: Nightshift records the classification
  (`…_empty_response`), not the bytes; we can capture them on a next run if that helps.
- Whether one loop alone fails at the same rate.

## REDACTED

Nothing was blacked out. Our NVIDIA key, the gateway's address and our machines' names are left out rather than written.
