# ADR-0010: Every reviewer model is configurable, NVIDIA's included; NVIDIA is only the default first reviewer, and no setup is blocked

**Status:** accepted · **Date:** 2026-10-19
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-19
**Extends:** [ADR-0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md), [ADR-0009](ADR-0009-two-reviewer-models-and-the-same-model-fallback.md)
**Trigger:** the owner, after ADR-0009 ([`planning/CONVERSATION_2026-10-19.md`](../../planning/CONVERSATION_2026-10-19.md), hop 5)

## Context

[ADR-0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md) and [ADR-0009](ADR-0009-two-reviewer-models-and-the-same-model-fallback.md) name NVIDIA as the default reviewer and recommend three models: A writes, NVIDIA reviews by default, B reviews special cases. They could be read as requiring NVIDIA, or as refusing a setup without an independent model family. The owner:

> *"And we make NVIDIA also configurable. Maybe some customers will want to use coding A, reviewing A, without NVIDIA, or A and B. We don't block this. NVIDIA is just a model like the others, just the default first reviewing model."*

## Decision

1. **Every reviewer model is configuration**: the default reviewer and the special-review model alike. NVIDIA is the value proven ships with for the default reviewer, nothing more.
2. **No setup is refused.** A team may run:
   - **A writes, A reviews**, with no second model;
   - **A writes, B reviews**, without NVIDIA;
   - **A writes, NVIDIA reviews, B reviews special cases**, the recommended setup ([ADR-0009](ADR-0009-two-reviewer-models-and-the-same-model-fallback.md)).
3. **Independence is recommended and reported, not enforced.** Every review states which model it used and whether that model is the author's ([ADR-0009](ADR-0009-two-reviewer-models-and-the-same-model-fallback.md), point 5), so a reader always sees which kind of independence the review had, including none beyond a fresh agent.
4. **Evidence is never configurable.** Every finding needs a test or a query and its rows ([ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)), whatever models are configured.

## Consequences

- Proven fits whatever models a team already has, including a single one.
- A team that reviews with its writing model gets fresh-agent reviews, not cross-family ones, and every review says so.
- Proven's own runs at the hackathon use the default, NVIDIA, as the first reviewer.
- The open point of ADR-0009, a change written by the default reviewer's own family, becomes a configuration choice like any other: the review runs and says the reviewer's model is the author's.
