# ADR-0018: proven builds the minimum of two Meterless engines, H-MEM and World Model, to the sponsor's common contract; Markovian only if time is left

**Status:** accepted · **Date:** 2026-10-21
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-21
**Extends:** [ADR-0003](ADR-0003-meterless-engines-in-go.md) (Meterless engines built in Go)
**Trigger:** the sponsor's engine guide and the organisers' setup guide ([`planning/CONVERSATION_2026-10-21_advisor.md`](../../planning/CONVERSATION_2026-10-21_advisor.md), hop 6)

## Context

[ADR-0003](ADR-0003-meterless-engines-in-go.md) decided to build two Meterless engines in Go and check them with Meterless's conformance suites. It did not say how much of each engine, or what contract they share. On 21 October the owner shared the sponsor's guides. They say:

- the engines are *"reference architectures, not a library you pip install. You implement the services described, in whatever language/framework your agent uses"*;
- the Build rule: *"implement at least one Meterless Engine in a working feature"*, with its *"conformance suite, evaluation or a reproducible feature test"* passing; *"simply mentioning an engine … does not satisfy the Build rule"*;
- the minimum to implement: for H-MEM, *"MemoryStoreService + MemoryRetrievalService (skip dreaming/sleep at first)"*; for World Model, *"WorldStoreService + IngestPipelineService + ContextBuilderService"*;
- the common contract: headless; every model call through *"an injected generator"*, never inside the engine; append-only JSONL audit, traces and provenance written to the run's `logs/`; a sample mode with no credentials; World Model ingest *"idempotent and resumable"*; the World Model's memory backed by H-MEM;
- the warning: *"Implementing everything before anything works … Six days is not long."*

The owner approved the advisor's proposal: *"ADRs: yes."* (translated from Polish)

## Decision

1. **Two engines, the minimum first.** H-MEM: the store and retrieval services. World Model: the store, the ingest pipeline and the context builder. H-MEM's dreaming and sleep, and the World Model's enrichment and control plane, are built only after the minimum works.
2. **One generator interface in Go**, passed to every engine and agent. No engine calls a model, an endpoint or a provider itself. The generator is configured, never hardcoded ([ADR-0013](ADR-0013-no-hardcoded-settings.md)); its default route is [ADR-0019](ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md).
3. **Every engine writes JSONL to the run's `logs/`**: audit, traces and provenance, append-only. The review's trace is that JSONL, so a judge can follow each finding from claim to verdict.
4. **The World Model's memory is backed by H-MEM.**
5. **World Model ingest is idempotent and resumable.** A run killed halfway and started again gives the same state, and a test does exactly that.
6. **Each engine works in a feature a user sees**: H-MEM keeps a finding refuted in one review from returning in the next; the World Model holds each review's claims, evidence, refutations and verdicts.
7. **Each engine proves itself** with Meterless's conformance suite through the TypeScript adapter ([ADR-0003](ADR-0003-meterless-engines-in-go.md)), or, where that does not run, with a reproducible feature test. The README names each engine, where it is in the code and the test that shows it working.
8. **Markovian is optional.** It is built only if the minimum of both engines works and a review of a large pull request outgrows a model's context.

## Consequences

- The Build rule is met by two engines in a working feature, with evidence, not by a mention.
- The engines run the same in sample mode, because the generator is the only thing sample mode replaces ([ADR-0001](ADR-0001-go-behind-adapters.md)).
- Markovian's "projected against actual savings" chart, which the sponsor suggests for the demo, is given up unless point 8 is reached.
- Scout Intent and Swarm, the sponsor's other engines, are not part of this decision.
