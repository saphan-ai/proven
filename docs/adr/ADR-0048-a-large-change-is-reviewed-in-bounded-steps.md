# ADR-0048: A large change is reviewed in bounded steps, with Meterless's Markovian engine

**Status:** accepted · **Date:** 2026-10-24
**Superseded in part by [ADR-0053](ADR-0053-all-four-meterless-engines-are-on-by-default.md):** point 4, "the benefit is measured, or the engine is not kept", no longer keeps the engine off; it is on by default.
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-24
**Follows:** [ADR-0003](ADR-0003-meterless-engines-in-go.md) and [ADR-0018](ADR-0018-meterless-engines-minimum-and-their-contract.md) (Meterless engines in Go), [ADR-0042](ADR-0042-a-proposed-fix-is-proved-by-the-same-test.md) (proved fixes)
**Trigger:** the owner, 24 October: *"Markovian from Meterless — we do it"*, and *"we do not do art for art's sake, only what is really worth using in our scenario."*

## Why we use it, and only here

A review today puts the whole change into each agent's context. A large pull request then costs more at every step and may not
fit at all. Meterless's **Markovian** engine solves exactly that: long work cut into bounded steps, with only the necessary state
carried from one step to the next, so the context and the cost per step stay flat however long the work is. We use it where
Proven has that problem, and nowhere else.

Saphan Studio has its own implementation, patent pending; for the hackathon we use Meterless's.

## Decision

1. **A change above a size set in the settings is reviewed in bounded steps**: file by file (or hunk by hunk), each step seeing
   one part plus a small carried state: what is already established, what is suspected, which claims wait for evidence.
2. **The proved-fix loop (ADR-0042) uses the same engine**: each new attempt carries only what failed and why.
3. **Built in Go to Meterless's specification and checked with its own conformance suite**, like H-MEM and the World Model;
   the scorecard is published as it is.
4. **The benefit is measured, or the engine is not kept**: on the corpus and the sample scenarios, the same findings with and
   without bounded steps, and the model tokens per review step for a large change. If it does not hold the findings and cut the
   cost, the review stays whole.

## Consequences

- Large pull requests are reviewed at a flat cost per step.
- A third Meterless engine in real use, checked by Meterless's own tests.
