# The judges' lab: the keys a live review needs

This page says which keys the lab needs for each way of running the model generator, and where to get each one
([ADR-0031](adr/ADR-0031-anyone-with-an-nvidia-key-can-run-a-live-review.md) point 6). It covers keys only. Every setting
and its default is in [CONFIGURATION.md](CONFIGURATION.md). Sample mode needs no key and makes no call:
[SAMPLE_MODE.md](SAMPLE_MODE.md).

## With one NVIDIA key (no Nightshift needed)

A free NVIDIA API key is enough to see Proven's agents work on live models.

1. Go to <https://build.nvidia.com> and sign in.
2. Open any model and choose *Get API Key*. Copy the key; it is shown once.
3. Put the key in the environment, never in a file that is committed:

   ```
   PROVEN_GENERATOR_MODE=nvidia
   PROVEN_GENERATOR_KEY=<the key>
   PROVEN_GUARDRAIL_MODE=nvidia
   PROVEN_GUARDRAIL_KEY=<the same key, or another>
   ```

4. Check it before a review with `livecheck`. Its generator line names the mode, and a missing or rejected key.

The address and the models have defaults, so nothing else is needed. The hosted API signs no receipt, so the trace says
*no receipt* for each call, and the requirement that an agent be answered by an `nvidia/*` model is checked against the
`model` field of the answer. A free key is rate limited; a review that meets the limit names it as a failed service.

## With Meterless Nightshift (the default)

Set `METERLESS_RUNTIME_TOKEN` to the runtime token of the Nightshift you run, and `PROVEN_GENERATOR_BASE_URL` to its
address. `PROVEN_GENERATOR_MODE` is left empty or set to `nightshift`. Nightshift is the team's own service: a judge
without one uses the NVIDIA key above.

## Keys of the other services

| Service | Key | Where it comes from | If it is empty |
|---|---|---|---|
| Guardrail | `PROVEN_GUARDRAIL_KEY` | the same NVIDIA key page, when `PROVEN_GUARDRAIL_MODE=nvidia` | every model call is refused and the review names the guardrail |
| Zetaris evidence | see [CONFIGURATION.md](CONFIGURATION.md) | the Zetaris account | Zetaris is down and the review names it; the rest runs |
