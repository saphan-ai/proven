# The judges' lab: one more key adds a second reviewer

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

## What you see once it is connected

- **Both reviewers look at every change.** They work independently: neither is shown the other's answer. A change
  under review cannot narrow this for itself: `proven.yaml` is read from the baseline.
- **Where they differ, the review says so.** Each finding carries both reviewers' verdicts and reasons, and the
  review counts the disagreements; the two are kept side by side and never averaged. The block *The reviewers
  disagree*, which shows them in the text, Markdown, HTML and pull-request comment output, and the verdicts in the
  JSON and SARIF output, arrive with the shell train (see [ARCHITECTURE.md](ARCHITECTURE.md)); the command line does
  not render them yet. Evidence still decides what is reported: a risk neither reviewer can prove is not a finding,
  however many models name it.
- **Every review names its reviewers**, and whether either of them is the model that wrote the change.

## What it costs

Two reviewers cost two model calls per review step. To spend less, name the classes the second reviewer covers in
`proven.yaml`, for example `doubled_classes: security,backend`. The security class is always doubled when a second
reviewer is connected. See [CONFIGURATION.md](CONFIGURATION.md) for the settings and their refusals: a key with no
address or no model, or an address that is not an `http` or `https` URL, stops Proven before any review starts and
names the variable.

## If the second reviewer is down

A review whose second reviewer fails still completes with the first reviewer's findings. The failure is listed
under `second-reviewer` in the review's failed services, and the findings show the second reviewer's verdict as
`no answer`, which is not a disagreement.
