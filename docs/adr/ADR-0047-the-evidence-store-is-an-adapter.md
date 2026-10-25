# ADR-0047: The evidence store is an adapter: Zetaris, or a local store, chosen by one setting

**Status:** accepted · **Date:** 2026-10-24
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-24
**Amends:** [ADR-0046](ADR-0046-a-data-source-is-one-adapter.md) (a source loads into the configured evidence store, not into Zetaris by name)
**Follows:** [ADR-0001](ADR-0001-go-behind-adapters.md) (Go behind adapters), [ADR-0015](ADR-0015-zetaris-through-its-rest-api.md) and [ADR-0023](ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md) (Zetaris), [ADR-0032](ADR-0032-proven-builds-every-query-from-a-template.md) (queries from templates), [ADR-0034](ADR-0034-proven-runs-air-gapped.md) (air-gapped)
**Trigger:** the owner, 24 October: *"Meterless is open source, so it can stay. But for Zetaris we must think about running it through a data-store adapter, because in Saphan we may not want Zetaris."*

## Context

Proven's evidence queries go to Zetaris, a commercial platform. Sample mode already runs the same SQL over local tables, so a
local query engine exists. An organisation, and Saphan Studio itself, may want Proven without Zetaris.

## Decision

1. **One evidence-store interface**: register a source's table, run a query built from a template, return its columns and rows
   with the snapshot date.
2. **Three stores**: **Zetaris** (over its REST API, as today); **a local store** in one file with no server (DuckDB or
   SQLite, chosen when built), grown from the engine sample mode already uses; and **CouchDB** (added the same day: *"we have
   CouchDB"*), the document store Saphan Studio already runs, so Proven inside Saphan keeps its evidence where Saphan keeps the
   rest. Postgres may follow.
3. **One setting chooses the store** (`PROVEN_EVIDENCE_STORE=zetaris|local|couchdb`). For the hackathon the default stays
   Zetaris.
4. **A template has one form per query language**: SQL for Zetaris and the local store, a Mango selector for CouchDB. The
   evidence in a review is the query in the form that actually ran, with the rows or documents it returned. A conformance test
   runs one scenario on every store and compares the evidence they produce.
5. **Source adapters (ADR-0046) load into the configured store**, whichever it is.
6. **The local store works fully air-gapped** (ADR-0034).
7. **A review's evidence records go to CouchDB when configured, in sample mode too** (the owner, 2026-10-25): the
   judges' lab runs a sample review and its records belong in the lab's CouchDB. Sample mode therefore reads
   `PROVEN_COUCHDB_RECORDS` and `PROVEN_EVIDENCE_LINK_BASE_URL` besides ADR-0020 §1's keys and ADR-0045's `OTEL_*`;
   with both unset it reads nothing more and writes nothing.

## Consequences

- Proven no longer depends on one vendor's platform; Zetaris stays a first-class choice.
- Each store must keep up with the templates; the conformance test is what catches a store that drifts.
