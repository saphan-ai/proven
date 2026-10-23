# ADR-0035: Every review can leave a signed evidence pack an auditor can verify

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (evidence), [ADR-0019](ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) (signed receipts)
**Trigger:** the owner, 23 October: features closer to enterprise

## Context

A review already writes a trace with every claim, piece of evidence, refutation and verdict, and Nightshift signs a
receipt for every model answer. A compliance audit (SOC 2, ISO 27001) asks to see that a change was reviewed, by what,
and on what evidence, long after the run. Today that record is a loose folder.

## Decision

1. **`--evidence-pack <file>`** writes one archive per review: the findings, the trace, every test command and its
   output, every query with its rows, the Nightshift receipts, the reviewed commits and the settings used (never a
   secret).
2. **The pack is signed** with a key the organisation configures, and holds a manifest with a hash of every file.
3. **`proven verify <pack>`** checks the signature, every hash and every receipt, and prints what was reviewed, when,
   by which models, and on what evidence.
4. **The proof is a test**: a pack verifies; a pack with one byte changed is refused by name.

## Consequences

- An auditor gets one file per review and one command to check it.
- The pack carries source excerpts and query rows; where it is kept is the organisation's decision, and the docs say so.
