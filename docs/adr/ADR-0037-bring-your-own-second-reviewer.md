# ADR-0037: Bring your own second reviewer: with a second model connected, every change is reviewed by two, and their disagreement is shown

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Extends:** [ADR-0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md), [ADR-0009](ADR-0009-two-reviewer-models-and-the-same-model-fallback.md) (two reviewer models), [ADR-0010](ADR-0010-every-reviewer-model-is-configurable.md)
**Follows:** [ADR-0031](ADR-0031-anyone-with-an-nvidia-key-can-run-a-live-review.md) (NVIDIA's hosted API), [ADR-0033](ADR-0033-the-text-review-is-written-for-a-person.md) (the review for a person)
**Trigger:** the owner, 23 October, after two reviewers of Proven's own security review disagreed on the same commit: *"Proven can use two models when the user connects a second one. That is our slogan."*

## Context

ADR-0009 gives Proven a default reviewer (NVIDIA) and a special-review model that doubles the security class. On
23 October the second model could only be reached through Nightshift, and it reviewed only the doubled classes. The
same day, Proven's own security review showed why two reviewers matter: on two of three pieces one model accepted the
code and the other rejected it with failing tests.

## Decision

1. **One reviewer by default, NVIDIA.** A user who connects nothing more gets one independent review.
2. **A second reviewer is any model behind an OpenAI-compatible API** (OpenAI, Anthropic through its compatible
   endpoint, a local server), named by three settings: `PROVEN_SECOND_REVIEWER_BASE_URL`, `PROVEN_SECOND_REVIEWER_KEY`
   (secret, never printed) and `PROVEN_SECOND_REVIEWER_MODEL`. The `special_reviewer` of ADR-0009 points at it.
3. **When a second reviewer is connected, every change is reviewed by both**, independently, neither seeing the other's
   answer. `proven.yaml` may narrow it to chosen review classes; security is always doubled.
4. **Disagreement is shown, never averaged.** When the two reviewers differ on a finding, the review shows a block
   *The reviewers disagree* with each model's verdict and reason, in text, Markdown, HTML and the pull-request comment;
   JSON and SARIF carry both verdicts. Evidence still decides what is reported ([ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)).
5. **Every review names its reviewers** and whether either is the author's own model (ADR-0009 point 5).
6. **The judges' lab runs with NVIDIA alone**, and `docs/LAB.md` shows how one more key adds the second reviewer.
7. **The README leads with it**: one model writes the code; Proven reviews it with another, and with a second one if
   you connect it, and shows you where they disagree.

## Consequences

- The claim Proven was built on is a feature a user can switch on with one key.
- Two reviewers cost two model calls per review step; the docs say so, and the user chooses the classes.
- A third-party API key is one more secret; it follows every rule for secrets (SECURITY.md).
