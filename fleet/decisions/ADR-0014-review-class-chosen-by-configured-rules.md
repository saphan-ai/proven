# ADR-0014: A change's review class is chosen by configured rules over the files it touches, not by a model; the highest-risk class wins

**Status:** accepted · **Date:** 2026-10-19
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-19
**Extends:** [ADR-0012](ADR-0012-review-classes-as-in-saphan-studio.md) (Saphan Studio's review classes)
**Follows:** [ADR-0013](ADR-0013-no-hardcoded-settings.md) (no hardcoded settings)
**Trigger:** the owner, on who decides a change's class ([`planning/CONVERSATION_2026-10-19.md`](../../planning/CONVERSATION_2026-10-19.md), hop 7)

## Context

[ADR-0012](ADR-0012-review-classes-as-in-saphan-studio.md) adopts Saphan Studio's five review classes and says a change gets one, chosen by what it touches. It did not say who chooses. The owner:

> *"Then you need to look at Saphan again and check what it means, security or backend class or the others. Someone must decide about it. It could be a model or a deterministic configuration. For this hackathon we take the configuration."*

What the classes mean in Saphan Studio, in short:

| class | what the review looks at |
|---|---|
| **security** | trust boundaries, authorization, secrets, dependency changes, destructive operations |
| **backend** | data access and queries, caching, behaviour with more than one instance, observability |
| **user interface** | every screen state, keyboard and accessibility, the requests each control sends |
| **mobile** | app lifecycle, offline and slow networks, small screens, data stored on the device |
| **docs** | the change explained on its own, decisions recorded, every command re-run |

In Saphan Studio the class is declared by the agent that orders the work, from what the change touches, and size never lowers it. Proven reviews other people's pull requests, where no such agent exists, so something else has to choose.

## Decision

1. **For the hackathon, the class is chosen by deterministic configuration, not by a model.** The same change always gets the same class, and anyone can see why.
2. **The configuration is a list of rules**, each mapping file-path patterns to a class, for example `**/auth/**`, `**/payment*/**`, `**/*secret*`, `**/migrations/**` and dependency manifests to **security**; `**/*.md` and `docs/**` to **docs**.
3. **A change touching several classes gets the highest-risk one**, in a configurable order. The default order is security, backend, mobile, user interface, docs.
4. **A change that matches no rule gets a configurable fallback class**, backend by default.
5. **Proven ships a default rule set**, documented with the other defaults ([ADR-0013](ADR-0013-no-hardcoded-settings.md)). A team replaces or extends it for its own layout.
6. **Each review states its class and the rule that chose it** ([ADR-0012](ADR-0012-review-classes-as-in-saphan-studio.md), point 5).

## Consequences

- Classification costs nothing and cannot vary between runs.
- Path rules can miss: a security-relevant change in a file with an ordinary name gets a lower class. The review names the rule that matched, so a team sees the miss and adds a rule.
- Choosing the class with a model, or a model checking the rules' choice, is a later option and needs its own decision.
- What each class checks in proven is built from 22 October.
