# ADR-0012: Proven uses Saphan Studio's review classes (security, backend, user interface, mobile, docs); security reviews are doubled by default, and which classes are doubled is configurable

**Status:** accepted · **Date:** 2026-10-19
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-19
**Extends:** [ADR-0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md), [ADR-0009](ADR-0009-two-reviewer-models-and-the-same-model-fallback.md)
**Trigger:** an open point of ADR-0008 and ADR-0009, which kinds of review are doubled ([`planning/CONVERSATION_2026-10-19.md`](../../hackathon/planning/CONVERSATION_2026-10-19.md), hop 6)

## Context

[ADR-0008](ADR-0008-reviewer-model-chosen-against-the-authors-model.md) says that special kinds of review, security among them, are doubled. It did not say what the kinds are. The owner:

> *"Take the same classes that we have in Saphan Studio today."*

In Saphan Studio, every review of a change carries one class, chosen by what the change touches, not by its size. The classes are security, backend, user interface, mobile and docs, and each adds its own checks to the review. Security reviews are done twice unless the person who ordered the review writes down why not.

## Decision

1. **Proven's review classes are Saphan Studio's:** security, backend, user interface, mobile, docs.
2. **A change gets one class, chosen by what it touches.** A change to authentication, payments, data access, configuration or anything destructive is a security review, however small.
3. **Security reviews are doubled by default**: one review by the default reviewer and one by the special-review model ([ADR-0009](ADR-0009-two-reviewer-models-and-the-same-model-fallback.md)).
4. **Which classes are doubled is configurable.** A team may double more classes, or none.
5. **Each review states its class**, and why the change was given it.

## Consequences

- Proven and Saphan Studio speak the same language about reviews, which makes taking part of proven into Saphan Studio simpler.
- A small change to a sensitive path gets the full security review; size never lowers the class.
- Doubling costs a second review. The default doubles only security, where a second opinion is worth most.
- What each class checks in detail is part of proven's build, from 22 October.
