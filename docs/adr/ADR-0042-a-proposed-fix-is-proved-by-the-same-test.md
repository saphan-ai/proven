# ADR-0042: A proposed fix is proved by the same test that proved the problem

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (evidence), [ADR-0007](ADR-0007-no-exploit-code.md) (no exploit code), [ADR-0025](ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md) (the Prover's container)
**Trigger:** the owner, 23 October, after reading that large vendors now validate findings and deliver fixes: *"we fit the big-tech trend; we have to show something more."*

## Context

Proven proves a problem with a test that passes on the old code and fails on the change. A fix suggested by a model is
only a suggestion until something shows it works. Proven already has the test that would show it.

## Decision

1. **For a finding proved by a test, Proven may propose a fix**: a patch to the change, written by a model.
2. **The fix is proved by the same test**, run three times in the Prover's container: it passes on the base, fails on
   the change, and passes on the change with the patch applied. Only then is the fix shown as *proved*.
3. **A patch that does not make the test pass, or that breaks another test of the package, is not shown as a fix**;
   the review says the attempt failed.
4. **The fix is a suggestion on the pull request**, never a commit: a person applies it.
5. **For a dependency finding, the fix is the smallest version that leaves the advisory's affected range**, named with
   the advisory row that shows it.

## Consequences

- A reviewer gets not only the problem and its proof, but a fix and its proof.
- Three test runs instead of two for each fixed finding; a setting can turn proposed fixes off.
