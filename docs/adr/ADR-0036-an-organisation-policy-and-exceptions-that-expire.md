# ADR-0036: An organisation sets one policy for every repository, and accepts a risk only with an owner and an expiry

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0013](ADR-0013-no-hardcoded-settings.md) (settings), [ADR-0014](ADR-0014-review-class-chosen-by-configured-rules.md) (review classes by configured rules)
**Trigger:** the owner, 23 October: features closer to enterprise

## Context

Each repository can carry its own `proven.yaml`. An organisation with many repositories wants one policy (for example:
a dependency on the known-exploited list always blocks), and a way to accept a known risk without silencing it for
ever.

## Decision

1. **An organisation policy file**, named by a setting (a path or a repository), is read before each repository's
   `proven.yaml`. A repository may make the policy stricter, never weaker; a weaker local rule is refused by name.
2. **An exception** names the finding it covers (rule and location), a reason, an owner and an expiry date. Until it
   expires the finding is reported as *accepted* with those four fields, never hidden; after it expires the finding
   counts again, and the review says the exception expired.
3. **An exception without an owner or an expiry is refused.**
4. **The proof is a test** for each: a weaker local rule refused, an accepted finding shown with its owner, an expired
   exception no longer accepted.

## Consequences

- Security teams set the bar once; teams cannot quietly lower it.
- Accepted risks stay visible and come back on their own.
