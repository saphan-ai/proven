# ADR-0013: Proven hardcodes no settings; every value it runs with comes from configuration, with a documented default

**Status:** accepted · **Date:** 2026-10-19
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-19
**Generalizes:** [ADR-0010](ADR-0010-every-reviewer-model-is-configurable.md), [ADR-0011](ADR-0011-how-a-change-names-its-model.md), [ADR-0012](ADR-0012-review-classes-as-in-saphan-studio.md), each of which made one group of settings configurable
**Trigger:** the owner, while ADR-0011 and ADR-0012 were being written ([`planning/CONVERSATION_2026-10-19.md`](../../planning/CONVERSATION_2026-10-19.md), hop 6)

## Context

Three decisions in a row made one more value configurable: the reviewer models, where a change names its model, which review classes are doubled. The owner made it the rule for everything:

> *"As a rule, we NEVER hardcode any settings. Everything is configurable. Write this as an ADR in Proven."*

## Decision

1. **No setting is hardcoded.** Every value proven runs with comes from configuration: models and their providers, endpoints, keys' variable names, trailer keys and sources, the model-family table, review classes and which are doubled, evidence sources, timeouts, retry and round limits, the guardrail and its backend, sample mode, logging, ports and paths.
2. **Every setting has a documented default**, so proven runs out of the box. The defaults are listed in one place, with what each one does.
3. **Configuration comes from a file and from environment variables**, environment winning, and secrets from environment variables only ([ADR-0001](ADR-0001-go-behind-adapters.md)).
4. **A run records the configuration it used**, secrets masked, so a review can be traced to its settings.
4a. **An empty secret is a missing service; a wrong value is a refusal** (the owner, 2026-10-23, after the train's
   tester found a live review with no keys stopping on the first missing token). In live mode, a secret that is not set
   (empty) makes its service down: the review goes on with the other services and names that one in its failed
   services. A secret or setting that is set but wrong (malformed, too long, a line break) is still refused before the
   review starts. The guardrail is the exception that keeps its own rule: see [ADR-0005](ADR-0005-guardrail-on-by-default.md) point 5.
4a. **Every setting is validated when the configuration is loaded**, against its allowed values or type. A wrong key or value stops proven before any work, and the refusal says what was wrong, what is allowed, and how to fix it. Proven never falls back silently to a default for a value that was set wrongly. These are the same principles the owner set for Saphan Studio's settings (see [ADR-0014](ADR-0014-review-class-chosen-by-configured-rules.md), Context).
5. **What is a rule, not a setting, stays a rule.** A finding needs evidence ([ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)), proven writes no exploit code ([ADR-0007](ADR-0007-no-exploit-code.md)), and a review says which model reviewed it ([ADR-0009](ADR-0009-two-reviewer-models-and-the-same-model-fallback.md)). These define what proven is; they are not values to tune.

## Consequences

- Every executor order of the build says which settings its change introduces, and with which defaults. A reviewer treats a literal value in the code as a defect unless it is a rule under point 5.
- The list of defaults is part of the README for judges: it is how they run proven without our keys.
- More settings mean more to test. Each one is tested at its default and at least one other value.
