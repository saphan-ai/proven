# ADR-0049: Meterless's Scout Intent plans each review before a model runs

**Status:** accepted · **Date:** 2026-10-24
**Superseded in part by [ADR-0053](ADR-0053-all-four-meterless-engines-are-on-by-default.md):** point 5, "the benefit is measured, or the engine is not kept", no longer keeps the engine off; it is on by default.
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-24
**Follows:** [ADR-0012](ADR-0012-review-classes-as-in-saphan-studio.md) and [ADR-0014](ADR-0014-review-class-chosen-by-configured-rules.md) (review classes), [ADR-0018](ADR-0018-meterless-engines-minimum-and-their-contract.md) (Meterless engines)
**Trigger:** the owner, 24 October: *"which fourth one are we missing? … we must say clearly that we do not do art for art's sake, only what is really worth using in our scenario."*

## Why we use it, and only here

Every review today runs the same agents whatever the change is. A version bump in `go.mod` does not need a test-writing Prover;
a change to the login code needs more than a documentation fix does. Meterless's **Scout Intent** reads raw input, detects its
intent and risk, and routes it before any model runs. That is the review plan Proven lacks.

## Decision

1. **Before the agents run, Scout Intent classifies the change**: dependency change, refactor, feature, security-sensitive code,
   documentation only, configuration; with the signals that led there and any ambiguity.
2. **The classification chooses the plan**, inside the review classes of ADR-0014: which agents, tools (ADR-0040) and sources
   (ADR-0046) run, and whether the review is doubled. A dependency-only change runs the dependency evidence without the Prover; a
   security-sensitive change is doubled; a documentation-only change runs no model.
3. **It never lowers a configured rule**: a class set by the organisation's policy (ADR-0036) stays the minimum.
4. **Built in Go to Meterless's specification and checked with its eval harness**, published as it is.
5. **The benefit is measured, or the engine is not kept**: model calls and time per review with and without the plan, and no
   finding lost on the corpus.

## Consequences

- Cheaper, faster reviews for simple changes; more attention where the risk is.
- With ADR-0048, all four Meterless engines are used, each where it earns its place.
