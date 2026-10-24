# ADR-0046: A data source is one adapter: it loads into Zetaris and plugs into the decision with its own rule

**Status:** accepted · **Date:** 2026-10-24
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-24
**Follows:** [ADR-0023](ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md) (loaded through Zetaris), [ADR-0032](ADR-0032-proven-builds-every-query-from-a-template.md) (queries from templates), [ADR-0038](ADR-0038-a-language-is-one-small-adapter.md) and [ADR-0040](ADR-0040-a-tool-run-is-evidence.md) (adapters), [ADR-0039](ADR-0039-more-evidence-sources.md) (more sources)
**Trigger:** the owner, 24 October: *"what if we add an adapter for adding more data sources, feeding the database and plugging into the decision engine?"*

## Decision

1. **A third public adapter interface, for data sources**, beside languages and tools. A source adapter declares:
   - **how it loads**: where the data comes from, how often it is refreshed, and the Zetaris table it is registered as,
     with the date of each snapshot;
   - **its query templates**, which the Investigator may pick (ADR-0032); a model never writes the SQL;
   - **its rule**: what a returned row means for a change (a banned package, a critical service, a disallowed licence),
     which becomes a finding whose evidence is the query and its rows, decided by the Judge like any other;
   - **a sample table and its tests**, so sample mode covers it.
2. **The sources of ADR-0039 are written as adapters**, so the interface is proved by them first.
3. **An organisation can add its own sources** the same way: an internal list of banned packages, service
   criticality from a CMDB, code ownership, incident history. Proven then judges a change against that organisation's
   own knowledge.
4. **`docs/ADAPTERS.md`** covers all three kinds, with one worked example each.

## Consequences

- Proven's evidence grows without changing its core; each new source is one small, tested contribution.
- Zetaris is where every source meets, so an organisation that already uses it registers a table and writes a rule.
- A source's data is only as fresh as its last load; each finding states the snapshot date of its row.
