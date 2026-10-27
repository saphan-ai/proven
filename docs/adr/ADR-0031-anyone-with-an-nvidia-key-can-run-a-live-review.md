# ADR-0031: Anyone with an NVIDIA API key can run a live review: the generator can also call NVIDIA's hosted API, and `proven scan` runs live

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0005](ADR-0005-guardrail-on-by-default.md) (the guardrail), [ADR-0013](ADR-0013-no-hardcoded-settings.md) (settings), [ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md) point 2 (real keys turn on the live services), [ADR-0019](ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md) (Nightshift is the primary provider), [ADR-0020](ADR-0020-proven-command-line-and-action-interface.md) (the command line)
**Trigger:** the owner, 23 October, on the judges' lab: *"and the models beyond sample mode? live mode?"*

## Context

Measured on 23 October: `proven scan` accepts only `--sample`; a live review runs only through the demo in
`demo/live/`. The generator calls Meterless Nightshift and nothing else, and Nightshift runs on the team's own machine
with the team's runtime token. A judge, or any user, has neither, so nobody outside the team can see Proven's models
work live. The guardrail can already call NVIDIA's hosted API (`PROVEN_GUARDRAIL_MODE=nvidia`), and NVIDIA hands out
API keys for that API at no cost.

## Decision

1. **The generator has a second provider: NVIDIA's hosted API** (`https://integrate.api.nvidia.com/v1`, the
   OpenAI-compatible chat completions it serves), chosen by a setting, `PROVEN_GENERATOR_MODE=nightshift|nvidia`.
2. **Nightshift stays the default** ([ADR-0019](ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md)). The
   hosted API is the path for whoever has no Nightshift.
3. **In `nvidia` mode the generator reads its key from `PROVEN_GENERATOR_KEY`**, its address and models from settings
   with defaults, and never prints the key. The rules that hold in `nightshift` mode hold here: every agent answered by
   an `nvidia/*` model where `PROVEN_GENERATOR_REQUIRE_NVIDIA` asks for it, the Skeptic from another family than the
   finder ([ADR-0006](ADR-0006-a-skeptic-from-another-model-family.md)), every call checked by the guardrail. Where
   Nightshift gives a signed receipt and the hosted API does not, the trace says *no receipt* instead of inventing one.
4. **`proven scan` runs live** when `--sample` is not given, with the same settings `livecheck` reads; a service whose
   key is empty is down and named in *failed services*, the rest of the review runs (the owner's decision on W27).
5. **The judges' lab turns live on from `lab/.env`.** With one NVIDIA key in it (`PROVEN_GENERATOR_MODE=nvidia`,
   `PROVEN_GENERATOR_KEY`, `PROVEN_GUARDRAIL_MODE=nvidia`, `PROVEN_GUARDRAIL_KEY`) the pull-request review runs on
   live models; Zetaris stays down unless its keys are given, and the review says so.
6. **The docs say, per mode, which keys are needed and where to get them**: `docs/CONFIGURATION.md`, `docs/LAB.md`.
7. **A reasoning model's thinking is bounded by a setting, never by a default** (the owner, 2026-10-27).
   `PROVEN_GENERATOR_THINKING_BUDGET` sends `thinking_budget` on every call; empty sends no such field, so a provider that
   does not know it sees the body it saw before.

## Consequences

- A judge with a free NVIDIA key sees Proven's agents work on live models, in a terminal and on a pull request.
- Two generator paths to keep working; each needs its own live check in `livecheck`.
- The hosted API has rate limits on a free key; a review that hits one names it as a failed service.
