# ADR-0028: Proven's code is security-reviewed as a whole, in pieces, each piece one trust boundary

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0005](ADR-0005-guardrail-on-by-default.md), [ADR-0007](ADR-0007-no-exploit-code.md), [ADR-0025](ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md), [ADR-0026](ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md)
**Trigger:** the owner, 23 October: *"Maybe start a systematic security review? We do it in pieces, but a whole one."*

## Context

Each change to Proven gets two independent reviews of its diff. By the second build day the code held over twenty
packages written by several executors, and no review had looked at it as a whole. Proven runs code from the change it
reviews, holds keys, asks models about untrusted text and posts to pull requests, so a property can be broken by the
way the pieces fit together without any one diff breaking it.

## Decision

1. **The code is reviewed as a whole, in nine pieces**, each one boundary where something untrusted meets something
   Proven protects: tools first; the Prover's container; secrets; the reviewed repository as input; evidence queries;
   models; output and CI; the lab and the images; data on disk. Their properties are listed in
   [SECURITY_REVIEW.md](../../SECURITY_REVIEW.md).
2. **Pieces 0, 1 and 2 go first**, then two or three a day.
3. **Each piece gets two independent reviews** by different models, phrased as the property that must hold and the
   proof expected, never as an attack.
4. **A weakness becomes a fix in the train that owns the code**, proved by a test that fails before and passes after.
5. **Two public documents at the repository's root**: `SECURITY.md` (what the product protects and how, and how to
   report a vulnerability) and `SECURITY_REVIEW.md` (this process and each piece's status).
6. **A weakness is published only after its fix is merged.**

## Consequences

- The judges can read what Proven protects, how it was checked, and what was fixed.
- Review seats spend part of the build week on the pieces; fixes compete with features for executor time.
- `SECURITY.md` must follow the code: a protection it describes is one that is on `main`, or it says it is not yet.
