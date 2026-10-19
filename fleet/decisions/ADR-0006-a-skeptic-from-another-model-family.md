# ADR-0006: Every finding faces a Skeptic from a different model family, NVIDIA Nemotron, and the evidence settles the disagreement, never a vote

**Status:** accepted · **Date:** 2026-10-18
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-18
**Follows:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (what counts as evidence)
**Trigger:** first part of ADR-0002 as written; the owner asked for it as a decision of its own ([`planning/CONVERSATION_2026-10-18.md`](../../planning/CONVERSATION_2026-10-18.md), hop 12)

## Context

On 16 October the advisor's proposal asked for two reviewers from different model families. Our own fleet showed why: two model families reviewing the same changes disagreed in 31 of 41 cases ([`planning/answers/README.md`](../../planning/answers/README.md)). A reviewer that checks its own findings with the same model shares its blind spots.

The hackathon asks for agents that critique each other, not one agent with several labels, and for NVIDIA technology wherever it helps. The owner, when the design was reshaped:

> *"NVIDIA mamy, musimy do tego doczepić jeszcze Zetaris i Meterless"*
> ("We have NVIDIA; we still have to attach Zetaris and Meterless to it.")

The README pushed on 17 October already named an NVIDIA Nemotron model as the Skeptic.

## Decision

1. **A Skeptic agent tries to refute every finding** before it reaches the review.
2. **The Skeptic runs on a different model family from the agent that made the finding.** Its model is an NVIDIA Nemotron model, through NVIDIA's hosted API.
3. **A disagreement is settled by the evidence** ([ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)): a test that runs, or a query whose rows say it. **Never by a vote**, and never by which agent is more confident.
4. **A finding the Skeptic refutes, with evidence, is dropped**, and the refutation is written to the trace and to the reviewer's memory ([ADR-0003](ADR-0003-meterless-engines-in-go.md)), so it does not come back.
5. **The Skeptic's model is a setting**, like every model in proven. Nemotron is the default, not a fixed part of the code.

## Consequences

- The trace shows critique between agents, which the judges look for.
- Every finding costs one more model call. Its cost and latency are in the trace.
- A disagreement neither side can prove stays in the review as unproven. It is not settled by either side's say-so.
- Which Nemotron model plays the Skeptic is chosen in the build, by measurement on our own examples.
