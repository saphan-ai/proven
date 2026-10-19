# ADR-0001: proven is written in Go, and every outside service sits behind an adapter

**Status:** accepted · **Date:** 2026-10-17
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-18
**Trigger:** the hackathon requires Zetaris and Meterless in every entry, and the design was reshaped around them ([`planning/CONVERSATION_2026-10-17.md`](../../planning/CONVERSATION_2026-10-17.md), hops 11 and 12)

## Context

proven has two lives. It is open source, entered in the Open Agent Hackathon 2026. And one part of it will be taken into Saphan Studio, the platform that runs our fleet.

The advisor recommended TypeScript, because Meterless publishes its engine references and conformance suites in TypeScript. The owner chose Go:

> *"D5 - nie, napiszemy w go z adapterem. Potem wykorzystamy to w Saphan - taki jest cel tej zabawy."*
> ("D5: no, we will write it in Go with an adapter. Later we will use it in Saphan; that is the point of this game.")

proven calls several services it does not own: Zetaris, NVIDIA's hosted models, the model providers, and Meterless's conformance suites at test time. The hackathon's technical guide asks that each sponsor integration sit behind an adapter, so that an outage does not take the project down during judging, and a judge without our keys can still run it in sample mode.

## Decision

1. **The product is written in Go.**
2. **Every outside service sits behind an adapter**: one Go interface per service, with a live implementation and a sample implementation. No agent and no part of the review logic calls a service directly.
3. **Every adapter call has a timeout and a fallback.** A failing service degrades the review, and the review says which service failed. It never hangs.
4. **Keys come from environment variables only.** An adapter with no key uses its sample implementation and says so in the trace.

## Consequences

- The part taken into Saphan Studio needs no rewrite of the language.
- Meterless's engines are not reused as code; they are built in Go (see [ADR-0003](ADR-0003-meterless-engines-in-go.md)).
- `SAMPLE_MODE=true` is a set of sample adapters, not a separate code path, so sample mode runs the same agent logic as a live run, as the hackathon's rules require.
- Each adapter is one more interface to write and test. That is the price of a review that survives an outage.
