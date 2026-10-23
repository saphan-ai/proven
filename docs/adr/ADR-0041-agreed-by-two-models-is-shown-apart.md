# ADR-0041: A problem two models agree on, without proof, is shown apart and labelled as an opinion

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (evidence), [ADR-0037](ADR-0037-bring-your-own-second-reviewer.md) (the second reviewer)
**Trigger:** the owner, 23 October: *"or some other verification, not from a database, but from judgement, the way we do it with our reviewers?"*

## Context

Proven reports a problem only with evidence: a test, a query, a tool's run. Some real problems cannot be proved that
way in one review, for example a confusing design or a missing check that no test exercises. With two reviewers from
different model families (ADR-0037), their independent agreement is a signal worth showing, but it is not proof.

## Decision

1. **A finding that both reviewers name independently, at the same place, and that no evidence proves**, is shown in a
   separate section: *Agreed by two models, not proven*.
2. **It is never mixed with proven findings**: it does not fail a pipeline, does not count in the totals, and carries
   both models' reasons word for word.
3. **One model alone is not enough**: a claim only one reviewer makes stays dropped, as before.
4. **JSON and SARIF mark these with their own level** (`note`), so a machine never reads them as proven.

## Consequences

- The reader gets the judgement of two independent reviewers without Proven claiming more than it can prove.
- Needs a second reviewer; with one, the section is empty and says why.
