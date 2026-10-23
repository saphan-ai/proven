# ADR-0005: Every prompt and answer passes a guardrail by default: NVIDIA's Nemotron 3.5 Content Safety, local or hosted, with a switch to turn it off

**Status:** accepted · **Date:** 2026-10-18
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-18
**Follows:** [ADR-0001](ADR-0001-go-behind-adapters.md) (adapters)
**Trigger:** the owner's request for a local NVIDIA guardrail model, and his ruling on it ([`planning/CONVERSATION_2026-10-18.md`](../../planning/CONVERSATION_2026-10-18.md), hops 1, 7 and 10)

## Context

Saphan Studio checks the content of its agents' work with guardrail models. The owner wants proven to do the same:

> *"For the judges we will add an option to run without guardrails, but guardrails are on by default, as in Saphan."*

NVIDIA's Nemotron 3.5 Content Safety (4B) judges whether a prompt, or a response, is safe. It answers `User Safety: safe|unsafe`, `Response Safety: safe|unsafe` and `Safety Categories: …`. We are installing it on our own model server (train order `nemotron-content-safety`). A judge's machine cannot reach our server.

NVIDIA also serves the model on its hosted API, on the same endpoint as the Skeptic. The owner (translated from Polish):

> *"We have our own developer key, so we are good to go."*

## Decision

1. **The guardrail is on by default.** Every prompt an agent sends to a model, and every answer it gets back, passes the guardrail first.
2. **One guardrail adapter, two backends:** our own model server, and NVIDIA's hosted API. Our runs use either; a judge with an NVIDIA key uses the hosted one.
3. **A setting turns the guardrail off**, for anyone with neither. The trace says that the run had no guardrail.
4. **An `unsafe` verdict stops that step** and is written to the trace with its categories. It is never silently dropped.
5. **A guardrail that is down never lets a model call through unchecked** (the owner, 2026-10-23). In live mode, when
   the guardrail's key is empty or its server cannot be reached, every model call is refused and the trace says why;
   the review names the guardrail in its failed services. Only the setting `PROVEN_GUARDRAIL_MODE=off` turns the
   guardrail off, on purpose, and the trace then says the run had no guardrail (point 3). A missing key is never read as
   "off".

## Consequences

- A judge with an NVIDIA key runs proven exactly as we do, guardrail included.
- Every model call costs one guardrail call more. Its latency is measured and shown in the trace.
- **Not confirmed yet:** the hosted model's id (NVIDIA's code sample leaves it blank) and its limits (NVIDIA's page states no free tier and no rate limit). The first live call settles both.
- **Not measured yet:** whether our server renders the model's prompt template correctly. The train order `nemotron-content-safety` measures it before the model is used.
