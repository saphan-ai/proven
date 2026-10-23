# ADR-0019: Meterless Nightshift is proven's primary AI provider; NVIDIA's models are reached through it by default

**Status:** accepted · **Date:** 2026-10-21
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-21
**Amends:** [ADR-0010](ADR-0010-every-reviewer-model-is-configurable.md) (every reviewer model configurable): the default route, not the freedom to change it
**Follows:** [ADR-0013](ADR-0013-no-hardcoded-settings.md) (no hardcoded settings), [ADR-0018](ADR-0018-meterless-engines-minimum-and-their-contract.md) (one injected generator)
**Trigger:** the organisers' setup guide ([`planning/CONVERSATION_2026-10-21_advisor.md`](../../hackathon/planning/CONVERSATION_2026-10-21_advisor.md), hop 6)

## Context

The organisers' setup guide (last updated 19 October) sets a Power rule: *"your app must support Meterless Nightshift as a primary AI provider."* Nightshift is Meterless's application that serves AI requests from apps on a machine: one OpenAI-compatible endpoint (by default `http://127.0.0.1:8088/v1`), a runtime token, and lanes that route each request to local models or to cloud providers configured inside it. The guide says the NVIDIA key goes in Nightshift's Providers tab, *"not in the Console and not in your app's .env file"*. The NVIDIA Nemotron award asks for *"at least one NVIDIA Nemotron model through Meterless Nightshift in a working feature"*.

[ADR-0010](ADR-0010-every-reviewer-model-is-configurable.md) made every reviewer model configurable, with an NVIDIA model as the default first reviewer. It did not say which route reaches that model.

The owner approved the advisor's proposal: *"ADRs: yes."* (translated from Polish)

## Decision

1. **Nightshift is the default route for every model call proven makes**, through the generator of [ADR-0018](ADR-0018-meterless-engines-minimum-and-their-contract.md). The defaults ship pointing at Nightshift.
2. **NVIDIA's models are reached through Nightshift by default.** The default reviewer and the Skeptic are Nemotron models in a Nightshift lane. The NVIDIA key lives in Nightshift, never in proven.
3. **proven implements the Nightshift runtime contract** as the guide gives it: `Authorization: Bearer <token>`, the `meterless-app` header, `meterless-scope: user`, model discovery with `GET /v1/models`, and `POST /v1/chat/completions`. A connection check sends a real request; a healthy `/healthz` alone is not proof that a model answers.
4. **Each review names the route and the model that answered**, read from Nightshift's signed receipt in the response (`meterless.routed_model`, `meterless.winning_route.provider`), not assumed from the configuration. **A review whose answering model is not an NVIDIA model (`nvidia/*`, provider `nvidia`) is refused**: proven does not show it as a verdict, it says which model answered and why the answer was set aside, and it counts toward nothing. A response without a receipt is refused the same way. Nightshift cannot be made to pin a model from outside (measured 2026-10-21, see Amendment 1), so this check is where the Nemotron guarantee lives.
5. **Other routes stay settings** ([ADR-0010](ADR-0010-every-reviewer-model-is-configurable.md)): a team may point the generator at another OpenAI-compatible endpoint. Nightshift is the default, not a lock.
6. **The settings are the base URL, the runtime token and the routing mode or model**, from the environment, validated when they load ([ADR-0013](ADR-0013-no-hardcoded-settings.md)). The token is a secret and lives in the environment only.
7. **Sample mode needs no Nightshift**, so a judge without it still sees the whole review ([ADR-0001](ADR-0001-go-behind-adapters.md)).

## Consequences

- The local guardrail model on our own machine stays behind its own adapter ([ADR-0005](ADR-0005-guardrail-on-by-default.md)); whether it is also served through a Nightshift lane is not part of this decision.
- Nightshift ships for Windows, macOS and NVIDIA DGX, not for other Linux machines. proven running on Linux reaches a Nightshift on another machine over the private network, with its own token.
- On 21 October the macOS installer was neither notarized by Apple nor signed with a developer identity; Meterless's own note says so. Where Nightshift runs, and how it is kept apart from the rest of the machine, is decided before it is installed.
- The judges' lab ([ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md)) runs in sample mode by default; with Nightshift on the judge's machine, the lab reaches it at the host's address.

## Amendment 1 — 2026-10-21: the route is checked, because it cannot be pinned

**Decided by:** the owner (*"tak, wszystkie 3"*, translated: "yes, all three") · **Written by:** the master of
`proven-nightshift-vm`, from measurements recorded in [`fleet/orders/TRAIN_ORDER_proven-nightshift-vm.md`](../../hackathon/fleet/orders/TRAIN_ORDER_proven-nightshift-vm.md) (W3, W5, N7).

Point 4 said the default lanes pin Nemotron models by name. Measured against Nightshift 1.2.1 that is not possible:

- For an identified app (one that sends `meterless-app`), Nightshift's proxy routes by the mode the operator set for
  that app in Nightshift and **records, but does not obey,** the model named in the request (Nightshift's own code
  comment, `engine/proxy.js`, NS-APP-STREAM AS-31/34/35). A request naming `nvidia/nemotron-3.5-lightning-30b-a3b` was
  answered by `nvidia/nemotron-3-ultra-550b-a55b` under mode `auto`.
- Under mode `deep` the same request went to a **local 3B model** inside Nightshift (`local-runtime/llama-3.2-3b-instruct`),
  with no NVIDIA call at all.
- The *Default model* setting offers modes only (`auto`, `fast`, `deep`, `cheap`, `local`, `fusion`).

So point 4 now checks the route instead of trusting a pin. Alongside it, the machine that runs Nightshift for proven
keeps no local chat models, and app `proven` runs mode `auto` in Nightshift; both are operator settings that a receipt
re-measures, and neither replaces the check.
