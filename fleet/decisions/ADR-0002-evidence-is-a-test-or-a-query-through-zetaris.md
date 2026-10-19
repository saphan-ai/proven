# ADR-0002: A finding counts only with evidence, and evidence is a test or a query through Zetaris with the rows it returned

**Status:** accepted · **Date:** 2026-10-17
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-18
**Extends:** the advisor's proposal of 16 October, *proof-carrying code review* ([`planning/PROPOSAL_advisor.md`](../../planning/PROPOSAL_advisor.md))
**Trigger:** Zetaris is required in every entry ([`planning/CONVERSATION_2026-10-17.md`](../../planning/CONVERSATION_2026-10-17.md), hop 11)

## Context

proven's idea from the start: a review finding counts only with evidence. On 16 October the evidence was *a test that fails on the change and passes on the base, or a citation from the code history, an issue or the docs*.

A citation is weak evidence. The reader has to trust that the reviewer read the source correctly, and cannot run it again.

The hackathon requires every entry to integrate Zetaris and warns that a sponsor *"bolted on"* at the end scores badly. Zetaris runs SQL over many data sources where the data already lives, without copying it first. The owner asked for more data for the evaluating agent:

> *"Jak dostaniemy platformę to możemy poszaleć. Np. ściąganie baz CVE jak data, inne zródła danych dla agenta oceniającego."*
> ("Once we get the platform we can go wild. For example, pulling CVE databases as data, other data sources for the evaluating agent.")

## Decision

1. **A finding counts only with one of two kinds of evidence:**
   - **a test** that fails on the change and passes on the base; or
   - **a query and the rows it returned.** The SQL is written into the review next to the finding, so any reader can run it again and get the same rows.
2. **A finding with neither is dropped**, not softened.
3. **Every evidence source is a table behind Zetaris.** The Investigator agent asks its questions as queries through Zetaris, across all sources at once.
4. **The same queries run on a local engine over small samples** when Zetaris is not reachable or `SAMPLE_MODE=true` (see [ADR-0001](ADR-0001-go-behind-adapters.md)).

Who tries to refute each finding is [ADR-0006](ADR-0006-a-skeptic-from-another-model-family.md). What kind of evidence proven never produces is [ADR-0007](ADR-0007-no-exploit-code.md).

## Consequences

- Zetaris carries the core of the product. Without it there is no query evidence, only tests.
- The answer to "explain why" is a query a reader can run, which is stronger than a citation.
- Every query must be deterministic over the same data. The review records which snapshot of each source it read.
- The local engine for sample mode must accept the same SQL as Zetaris, or the queries need one translation layer. Which one is not known until we have the Zetaris instance.
