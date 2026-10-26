# ADR-0053: All four Meterless engines are on by default, each with its own off setting, and Scout never lightens a change it cannot verify

**Status:** accepted · **Date:** 2026-10-26
**Decided by:** the owner · **Written by:** the executor (claude-code, `claude-sonnet-5-5`), 2026-10-26
**Supersedes in part:** [ADR-0048](ADR-0048-a-large-change-is-reviewed-in-bounded-steps.md) point 4 and [ADR-0049](ADR-0049-scout-intent-plans-each-review.md) point 5, the clause "the benefit is measured, or the engine is not kept", as far as it kept the engine off
**Follows:** [ADR-0003](ADR-0003-meterless-engines-in-go.md) and [ADR-0018](ADR-0018-meterless-engines-minimum-and-their-contract.md) (Meterless engines in Go)
**Trigger:** the owner, 26 October: all four engines are on by default, in Proven and in the lab, even where the benefit is not measured yet; each one can be switched off by its setting.

## Context

Proven carries four Meterless engines. H-MEM and the World Model ran in every review, with no way to turn either off.
Markovian (bounded steps for a large change) and Scout Intent (the plan of a review) landed switched off, by ADR-0048 and
ADR-0049, until a measurement on recorded reviews held. That measurement does not exist yet: the corpus recordings are not made.

## Decision

1. **All four engines are on by default, in Proven and in the lab.** The lab's pull-request review is a sample review, so
   sample mode reads the same switches and has the same defaults: it reads `PROVEN_MARKOVIAN` (with its three tuning
   variables), `PROVEN_HMEM` and `PROVEN_WORLD_MODEL`, and a sample review with none of them set runs Markovian, H-MEM and the
   World Model. Scout Intent is on when `scout_intent` is not set in `proven.yaml`.
2. **Each of the four has its own off setting**, named here:

   | engine | off setting | what off means |
   |---|---|---|
   | Markovian | `PROVEN_MARKOVIAN=off` | every change is reviewed whole, and a new attempt of a proposed fix sees every earlier attempt |
   | Scout Intent | `scout_intent: off` in `proven.yaml` | every review runs the full plan, and the trace says no plan was chosen |
   | H-MEM | `PROVEN_HMEM=off` | the store is not opened, not read and not written: the Judge suppresses nothing from it and the fold records no refutation |
   | World Model | `PROVEN_WORLD_MODEL=off` | the store is not opened and there is no fold of the trace, so H-MEM learns nothing from the trace either |

   The two stores are independent. With only H-MEM off, the World Model still folds and records no refutation. A wrong value
   is refused by name (ADR-0013); empty means on.
3. **The benefit of Markovian and Scout Intent stays UNMEASURED** until the corpus recordings exist, and every document that
   names these engines says so. Being on by default is a decision about what Proven carries; it is not a claim that either engine
   makes a review better.
4. **The Scout guard.** Scout reads names, and a name proves nothing. With Scout Intent on by default, Scout may lighten a review
   only for a change whose every file is verifiably prose (a documentation plan) or verifiably a manifest (a dependency plan).
   When Scout answers documentation or dependency for any other change, the review runs every agent in class `security` and
   doubled, never in class `docs` and never lighter than the files require. A class that the team's rules already set higher is kept.
   A change Scout answers as feature, refactor or configuration, or answers ambiguously, is not touched by the guard: it runs in
   the class its rules give it, as it did before this ADR.
   - **Prose** is checked against the configured prose suffixes and a closed list of source and script suffixes.
   - **A manifest** is checked against a closed set of file names (the module files, lockfiles and package manifests of the
     ecosystems Scout's dependency class names). Every path of the change is checked, the old path and the new path of a rename
     alike, because Scout scores both: a manifest renamed to a source file, or a source file renamed to a manifest, is not a
     dependency change. A path under a `vendor/` directory is vendored code and is not a manifest either.
5. **Conformance is unchanged**: `markovian 8/8`, and Scout Intent's eval as committed.

## Why the guard is part of this decision

On the blind set (46 changes frozen before the run) three security-relevant code changes were answered as documentation:
a Polish-named, a Thai-named and a credentials-handling file under a documentation directory. The plan step refused a
documentation plan for their non-prose files and ran the full agents, but in class `docs` and undoubled, where class `security` is
doubled. A review that is lighter than its files require, under a name that says it is complete, is the failure this guard closes.
Raising the class costs one more review run for a change Scout could not place; lowering it can cost a missed finding.

## Consequences

- A review of a large change is cut into bounded steps, every review is planned, and both stores run, without anyone setting anything.
- A reader who wants the review as before sets the four settings off.
- The World Model's code now accepts a missing H-MEM (the fold records no refutation then); before, it refused to open without one.
- The cost is one more run, in class `security`, for a change that Scout called documentation or dependency and that is not
  verified prose or a manifest.
- When the corpus recordings exist, the measurement of ADR-0048 and ADR-0049 is run on them. Its result goes into the README and
  here; a benefit is claimed only from it.
