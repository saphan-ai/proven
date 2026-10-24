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

The address and the models have defaults, so nothing else is needed. In this mode the address is `PROVEN_NVIDIA_BASE_URL`
(`https://` only; any other value is refused by name). `PROVEN_GENERATOR_BASE_URL` is not read here, so a Nightshift
address left in the environment never receives the key. `livecheck` names the host the answer came from. The hosted API signs no receipt, so the trace says
*no receipt* for each call, and the requirement that an agent be answered by an `nvidia/*` model is checked against the
`model` field of the answer. A free key is rate limited; a review that meets the limit names it as a failed service.

## With Meterless Nightshift (the default)

Set `METERLESS_RUNTIME_TOKEN` to the runtime token of the Nightshift you run, and `PROVEN_GENERATOR_BASE_URL` to its
address (this variable is read in `nightshift` mode only). `PROVEN_GENERATOR_MODE` is left empty or set to `nightshift`. Nightshift is the team's own service: a judge
without one uses the NVIDIA key above.

## Keys of the other services

| Service | Key | Where it comes from | If it is empty |
|---|---|---|---|
| Guardrail | `PROVEN_GUARDRAIL_KEY` | the same NVIDIA key page, when `PROVEN_GUARDRAIL_MODE=nvidia` | every model call is refused and the review names the guardrail |
| Zetaris evidence | see [CONFIGURATION.md](CONFIGURATION.md) | the Zetaris account | Zetaris is down and the review names it; the rest runs |

## A second reviewer: one more key

The lab runs with NVIDIA alone ([ADR-0037](adr/ADR-0037-bring-your-own-second-reviewer.md) point 6). Every change
gets one independent review, from the default reviewer, and nothing else is needed.

Proven can review every change with two models. The second one is any model behind an OpenAI-compatible chat API:
OpenAI, Anthropic through its compatible endpoint, or a server on your own machine. You switch it on with three
settings in the environment Proven runs in, and the key is the switch:

| Variable | Value |
|---|---|
| `PROVEN_SECOND_REVIEWER_KEY` | the API key. Secret: never printed, never written to the trace or the result |
| `PROVEN_SECOND_REVIEWER_BASE_URL` | the API address, for example `https://api.openai.com/v1` or `http://localhost:11434/v1` |
| `PROVEN_SECOND_REVIEWER_MODEL` | the model name the API knows, for example `gpt-5` |

Add the three variables and run the lab as before. Take the key away and the lab is back to one reviewer. Nothing
else changes.

### What you see once it is connected

- **Both reviewers look at every change.** They work independently: neither is shown the other's answer. A change
  under review cannot narrow this for itself: `proven.yaml` is read from the baseline.
- **Where they differ, the review says so.** Each finding carries both reviewers' verdicts and reasons, and the
  review counts the disagreements; the two are kept side by side and never averaged. The block *The reviewers
  disagree*, which shows them in the text, Markdown, HTML and pull-request comment output, and the verdicts in the
  JSON and SARIF output, arrive with the shell train (see [ARCHITECTURE.md](ARCHITECTURE.md)); the command line does
  not render them yet. Evidence still decides what is reported: a risk neither reviewer can prove is not a finding,
  however many models name it.
- **Every review names its reviewers**, and whether either of them is the model that wrote the change.

### What it costs

Two reviewers cost two model calls per review step. To spend less, name the classes the second reviewer covers in
`proven.yaml`, for example `doubled_classes: security,backend`. The security class is always doubled when a second
reviewer is connected. See [CONFIGURATION.md](CONFIGURATION.md) for the settings and their refusals: a key with no
address or no model, or an address that is not an `http` or `https` URL, stops Proven before any review starts and
names the variable.

### If the second reviewer is down

A review whose second reviewer fails still completes with the first reviewer's findings. The failure is listed
under `second-reviewer` in the review's failed services, and the findings show the second reviewer's verdict as
`no answer`, which is not a disagreement.
