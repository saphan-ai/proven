# ADR-0032: Proven builds every query from a fixed template; a model picks the template and its values

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the master of the reviewer core (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (evidence is a test or a query),
[ADR-0015](ADR-0015-zetaris-through-its-rest-api.md) and [ADR-0023](ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md)
(Zetaris over REST), [ADR-0025](ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md) point 7 (Proven
builds the Prover's command)
**Trigger:** the owner, 23 October, choosing how a query reaches Zetaris

## Context

The Investigator and the Skeptic gather query evidence through Zetaris (ADR-0002). A query is evidence only if its
meaning is the one Proven intends: the rows it returns must answer the question the review asked. ADR-0025 point 7 has
already settled the same question for the Prover's command: Proven builds what is run, from a fixed template, and a
model only supplies values.

## Decision

1. **No query text written by a model reaches Zetaris.** The Investigator and the Skeptic choose one query template from
   a closed list that Proven defines (for example: an OSV advisory by package and version, a KEV entry by CVE id, an
   EPSS score by CVE id) and give the values for its parameters.
2. **Proven validates every value against a strict format before it builds the query**: a package name per ecosystem, a
   version, a CVE id, an advisory id. A value that fails its format is refused with a named reason, and that finding has
   no query evidence (the Judge decides as today).
3. **Proven builds the SQL itself** from the template and the validated values, with each value as a literal that cannot
   change the query's meaning. The query and its rows are recorded in the trace as today.
4. **A read-only Zetaris role is a second layer only if Zetaris offers one cheaply.** This is measured, not assumed:
   until a measurement shows that a read-only key exists and refuses a write, Proven relies on points 1-3 alone, and the
   documentation does not claim a read-only key.

## Consequences

- A model can no longer ask Zetaris a question outside the templates. A new kind of evidence needs a new template,
  which is a change to Proven reviewed like any other.
- The templates are listed in one place, with their parameters and formats, so a reader can see every question Proven
  can ask.
- The query check that read a model's SQL text is replaced by the template list; the sample fixtures' recorded answers
  change shape once.
