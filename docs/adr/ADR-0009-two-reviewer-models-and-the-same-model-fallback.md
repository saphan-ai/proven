# ADR-0009: Two reviewer models are configured; a doubled review may use the author's own model as its second reviewer, because a fresh agent still sees the code for the first time

**Status:** accepted · **Date:** 2026-10-19
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-19
**Extends:** [ADR-0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md) (the reviewer's model chosen against the author's)
**Trigger:** an open point of ADR-0008, the case where a configured reviewer model is the same as the author's ([`planning/CONVERSATION_2026-10-19.md`](../../hackathon/planning/CONVERSATION_2026-10-19.md), hop 4)

## Context

[ADR-0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md) says that the reviewer's model comes from configuration, NVIDIA by default, and that special reviews such as security are doubled by a second configured model. It left open what happens when a configured reviewer model is the same as the model that wrote the change.

Many teams have only two models: the one they write code with, and one more. The owner's ruling:

> *"Two second models can be configured. Then, when a customer has model A and NVIDIA, the review is done by NVIDIA; this meets the requirement of an independent reviewing model. For special classes like security, the same model A is used. That is still valuable, because the review is done by a fresh agent who sees the code for the first time. The best setup is: model A writes the code, NVIDIA reviews by default, and model B reviews special cases like security."*

## Decision

1. **Two reviewer models are configured:** a **default reviewer** (NVIDIA by default) and a **special-review model** for the review classes that are doubled, security among them.
2. **Every review uses the default reviewer.** When the author's model is A and the default reviewer is NVIDIA, that review is independent of the author's model family.
3. **A doubled review adds the special-review model.** If that model is the author's own model A, the review still runs. Its value is not a different model family but a **fresh agent**: a new context that has not seen the code or the author's reasoning before.
4. **The recommended setup has three models:** model A writes the code, NVIDIA reviews by default, and model B reviews the special cases. Then both reviews of a doubled review are independent of the author's family.
5. **Each review states which model it used, and whether that model is the author's.** A reader can see which independence a review had: a different model family, or a fresh agent of the same model.
6. **Evidence is still required for every finding** ([ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)), whichever model found it.

## Consequences

- A team with only two models still gets one independent review on every change, and a second, fresh review on the changes that matter most.
- A same-model second review is weaker independence than a different family. Proven says which kind each review had, rather than presenting both as equal.
- With three models, a security review is checked by two families other than the author's.
- **Not decided yet:**
  - the exact place and form in which a commit or pull request names its model (open since ADR-0008);
  - which review classes besides security are doubled (open since ADR-0008);
  - what happens when the author's model is itself the default reviewer's family, for example a change written by an NVIDIA model.
