# ADR-0004: The first evidence sources are OSV, CISA KEV, FIRST EPSS and the repository's own history

**Status:** accepted · **Date:** 2026-10-17
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-18
**Follows:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (evidence through Zetaris)
**Trigger:** the advisor's proposal D6 ([`planning/CONVERSATION_2026-10-17.md`](../../planning/CONVERSATION_2026-10-17.md), hops 11 and 12)

## Context

Many public sources could feed the reviewer: OSV, NVD, the GitHub Advisory Database, CISA KEV, FIRST EPSS, the CWE list, deps.dev and OpenSSF Scorecard, and the repository's own history and issues. The build window is six days. Each source costs a download script, a sample, a licence check and a table behind Zetaris.

What a reviewer of a pull request most needs to know about a dependency: does this exact version have a published advisory; is that vulnerability used in real attacks; how likely is it to be. About the code: was this changed, reverted or fixed before.

The owner agreed to the advisor's proposal:

> *"D6 - zgoda"* ("D6: agreed")

## Decision

1. **The first set:**

   | source | what it answers |
   |---|---|
   | [OSV](https://osv.dev) | which package versions an advisory covers, per ecosystem |
   | [CISA KEV](https://www.cisa.gov/known-exploited-vulnerabilities-catalog) | whether a vulnerability is known to be used in real attacks |
   | [FIRST EPSS](https://www.first.org/epss/) | how likely a vulnerability is to be used in an attack |
   | the repository's own history | what was changed, reverted or fixed before, and in which commit |

2. **Next, only if the build leaves time:** the NVD CVE records, the CWE list, OpenSSF Scorecard.
3. **All data is fetched by scripts written in the build window**, from 22 October. Nothing is downloaded before it, because data prepared earlier would have to be declared as a pre-existing component.
4. **Small samples of each source are kept in this repository**, for sample mode and for tests.

## Consequences

- OSV matches lockfile versions per ecosystem better than NVD does, which is why it comes first.
- KEV and EPSS together let a finding say not only "vulnerable" but "used in attacks" and "how likely". That is the part a reader acts on.
- The licence of every source is listed in the README once its script exists.
- The build machines must reach these sources (train order `proven-cursor-infra`, W8).
