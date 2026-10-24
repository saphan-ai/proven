# ADR-0044: Every finding names its level of proof, in every format

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md), [ADR-0033](ADR-0033-the-text-review-is-written-for-a-person.md), [ADR-0040](ADR-0040-a-tool-run-is-evidence.md), [ADR-0041](ADR-0041-agreed-by-two-models-is-shown-apart.md)
**Trigger:** the owner, 23 October: *"we have to show something more."*

## Decision

1. **Each finding carries one named level of proof**, from strongest to weakest:
   - *proved by a test*: passes on the old code, fails on the new (ADR-0002);
   - *proved by a tool*: a deterministic tool's result differs between old and new (ADR-0040);
   - *proved by data*: a query and the rows it returned (ADR-0002);
   - *agreed by two models, not proven* (ADR-0041).
2. **The level is shown first** in text, Markdown, HTML and the pull-request comment, and is a field in JSON and
   SARIF. A proposed fix carries its own level: *fix proved by the same test* or *fix not proved* (ADR-0042).
3. **The totals count by level**, so a reader sees at once how much is proven and how much is opinion.

## Consequences

- One word tells a reader how far to trust a finding, without reading its evidence.
