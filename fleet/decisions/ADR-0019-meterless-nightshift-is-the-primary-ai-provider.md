# ADR-0019: Meterless Nightshift is proven's primary AI provider; NVIDIA's models are reached through it by default

**Status:** accepted · **Date:** 2026-10-21
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-21
**Amends:** [ADR-0010](ADR-0010-every-reviewer-model-is-configurable.md) (every reviewer model configurable): the default route, not the freedom to change it
**Follows:** [ADR-0013](ADR-0013-no-hardcoded-settings.md) (no hardcoded settings), [ADR-0018](ADR-0018-meterless-engines-minimum-and-their-contract.md) (one injected generator)
**Trigger:** the organisers' setup guide ([`planning/CONVERSATION_2026-10-21_advisor.md`](../../planning/CONVERSATION_2026-10-21_advisor.md), hop 6)

## Context

The organisers' setup guide (last updated 19 October) sets a Power rule: *"your app must support Meterless Nightshift as a primary AI provider."* Nightshift is Meterless's application that serves AI requests from apps on a machine: one OpenAI-compatible endpoint (by default `http://127.0.0.1:8088/v1`), a runtime token, and lanes that route each request to local models or to cloud providers configured inside it. The guide says the NVIDIA key goes in Nightshift's Providers tab, *"not in the Console and not in your app's .env file"*. The NVIDIA Nemotron award asks for *"at least one NVIDIA Nemotron model through Meterless Nightshift in a working feature"*.

[ADR-0010](ADR-0010-every-reviewer-model-is-configurable.md) made every reviewer model configurable, with an NVIDIA model as the default first reviewer. It did not say which route reaches that model.

The owner approved the advisor's proposal: *"ADRs: yes."* (translated from Polish)

## Decision

1. **Nightshift is the default route for every model call proven makes**, through the generator of [ADR-0018](ADR-0018-meterless-engines-minimum-and-their-contract.md). The defaults ship pointing at Nightshift.
2. **NVIDIA's models are reached through Nightshift by default.** The default reviewer and the Skeptic are Nemotron models in a Nightshift lane. The NVIDIA key lives in Nightshift, never in proven.
3. **proven implements the Nightshift runtime contract** as the guide gives it: `Authorization: Bearer <token>`, the `meterless-app` header, `meterless-scope: user`, model discovery with `GET /v1/models`, and `POST /v1/chat/completions`. A connection check sends a real request; a healthy `/healthz` alone is not proof that a model answers.
4. **Each review names the route and the model that answered**, read from Nightshift's response, not assumed from the configuration. `meterless/auto` is a routing mode that may pick a model from another family, so the default lanes pin Nemotron models by name.
5. **Other routes stay settings** ([ADR-0010](ADR-0010-every-reviewer-model-is-configurable.md)): a team may point the generator at another OpenAI-compatible endpoint. Nightshift is the default, not a lock.
6. **The settings are the base URL, the runtime token and the routing mode or model**, from the environment, validated when they load ([ADR-0013](ADR-0013-no-hardcoded-settings.md)). The token is a secret and lives in the environment only.
7. **Sample mode needs no Nightshift**, so a judge without it still sees the whole review ([ADR-0001](ADR-0001-go-behind-adapters.md)).

## Consequences

- The local guardrail model on our own machine stays behind its own adapter ([ADR-0005](ADR-0005-guardrail-on-by-default.md)); whether it is also served through a Nightshift lane is not part of this decision.
- Nightshift ships for Windows, macOS and NVIDIA DGX, not for other Linux machines. proven running on Linux reaches a Nightshift on another machine over the private network, with its own token.
- On 21 October the macOS installer was neither notarized by Apple nor signed with a developer identity; Meterless's own note says so. Where Nightshift runs, and how it is kept apart from the rest of the machine, is decided before it is installed.
- The judges' lab ([ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md)) runs in sample mode by default; with Nightshift on the judge's machine, the lab reaches it at the host's address.
