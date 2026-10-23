# ADR-0008: The reviewer's model is chosen against the model that wrote the change, as stated in the commit or pull request; by default it is NVIDIA's, and some kinds of review are doubled

**Status:** accepted · **Date:** 2026-10-19
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-19
**Extends:** [ADR-0006](ADR-0006-a-skeptic-from-another-model-family.md) (a Skeptic from another model family)
**Trigger:** the open question in [`DESIGN.md`](../../hackathon/DESIGN.md), "Independent reviewers" ([`planning/CONVERSATION_2026-10-19.md`](../../hackathon/planning/CONVERSATION_2026-10-19.md), hop 3)

## Context

The owner's design asks that the reviewer come, where possible, from a different model family than the agent that wrote the code, and that some reviews be done twice, by two different models. [ADR-0006](ADR-0006-a-skeptic-from-another-model-family.md) covered a narrower relationship: the Skeptic differs from the agent that *made the finding*. It did not say how proven learns which model *wrote the change*, or how it picks the reviewer. Proven cannot detect the author's model from the code itself, and this design does not try.

The owner's ruling:

> *"The model is stated in the commit and PR notes. If it is not stated, the default model from the configuration is taken. By default it is NVIDIA. A second model is also configured: for special types of review, like security, the review is doubled."*

## Decision

1. **The author's model is what the change says it is.** Proven reads the model named in the change's commit messages or pull-request description. It does not guess.
2. **The reviewer's model comes from configuration.** If the change names no author model, the configured default reviewer model is used.
3. **The default reviewer model is an NVIDIA model.**
4. **A second reviewer model is configured as well.** For special kinds of review, security among them, the review is doubled: two independent reviews, one per configured model.
5. **Every finding in every review still needs evidence** ([ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)). When the two reviews of a doubled review conflict, the evidence behind both is examined, never a vote; what stays unsettled is reported as unsettled, and a person decides.

## Consequences

- Proven's reviewer independence depends on the change being honest about its author. A change that names no model, or names the wrong one, gets the default reviewer. The review says which model reviewed it and whether the author's model was stated.
- Teams that already record the model in their commits or pull requests get independence with no extra step. Our own fleet records the model of every run.
- A doubled review costs two reviews. Which kinds of review are doubled is configuration, starting with security.
- **Not decided yet:**
  - the exact place and form in which a commit or pull request names its model;
  - what happens when a configured reviewer model is the same as the author's: **settled by [ADR-0009](ADR-0009-two-reviewer-models-and-the-same-model-fallback.md)** for the second review of a doubled review; the case where the author's model is the default reviewer's family stays open there;
  - which kinds of review, besides security, are doubled.
