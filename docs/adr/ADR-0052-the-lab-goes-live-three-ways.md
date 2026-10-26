# ADR-0052: The judges' lab goes live three ways, each wired with no hand edit

**Status:** accepted · **Date:** 2026-10-26
**Decided by:** the owner · **Written by:** the executor of wagon W17, round 2
**Follows:** [ADR-0013](ADR-0013-no-hardcoded-settings.md) (no hardcoded settings),
[ADR-0017](ADR-0017-a-gitea-lab-for-the-judges.md) (the lab), [ADR-0019](ADR-0019-meterless-nightshift-is-the-primary-ai-provider.md),
[ADR-0023](ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md), [ADR-0026](ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md),
[ADR-0031](ADR-0031-anyone-with-an-nvidia-key-can-run-a-live-review.md)
**Trigger:** the owner's four relays of 26 October 2026, 00:45 to 01:00 UTC. *Number note:* ADR-0051 is held for the
judges' gateway (wagon W18). This record takes the next number after it.

## Context

The lab's workflow ran `proven scan --sample` and passed no key, so the pull-request reviews in the lab could never be
live, even for a judge who owned an NVIDIA key. `make livecheck` read `lab/.env`; nothing else did. Judges have three
different starting points: a free NVIDIA key and nothing else; a token for our services through the judges' gateway; or
their own Zetaris account and a Mac. Each must reach a live review with no hand edit of a file in the repository, and
each must be written down step by step (every act a person must do has numbered steps in the docs).

## Decision

1. **One mechanism for the keys.** `make up` reads `lab/.env`. Each non-empty setting named in `LAB_LIVE_KEYS`
   (`lab/versions.env`) becomes an Actions secret of the demo repository. A credential (`LAB_SECRET_KEYS`) reaches the
   seed as one 0600 file of the lab's state volume, `/state/secrets`, written by `make up` over standard input and
   removed by the seed on every exit and again by `make up`; it is named nowhere in Compose and is in no container's
   environment, so `docker compose config` and `docker inspect` never show it. A setting that opens nothing (a mode, an
   address, a model name) still passes by name through Compose. Each value goes to Gitea over a pipe and is never
   written to an image or a log; the log names the settings it stored. An empty setting is not stored, and a secret
   that was emptied is removed, so the lab can return to sample mode. (Amended in round 5: before it, the credentials
   too were passed by name through Compose, and a render of the Compose file or an inspect of the seed showed them.)
2. **The workflow chooses.** It runs a live review when the gateway pair, a Nightshift token or a generator key is
   stored, in that order of precedence, and a sample review otherwise. The first visible line of Proven's comment says
   which: `sample: no keys` or `live: <generator>, Zetaris on|off`. The marker comment stays the first line of the body,
   because `make e2e` and the comment-update logic read it there.
3. **Path A, one NVIDIA key.** The key alone makes the lab run in `nvidia` mode and serves the guardrail too, unless a
   guardrail key is stored.
4. **Path B, the judges' gateway.** The gateway's address and the judge's own token are two settings
   (`JUDGES_GATEWAY_URL`, `JUDGES_GATEWAY_TOKEN`). The workflow points Proven's Nightshift address and Zetaris address at
   the gateway and sends the judge's token; our own Nightshift token and Zetaris key never reach the lab. The names are
   this lab's, chosen because W18's order was not available to this wagon; if the gateway uses others, only the settings
   change. The gateway's boundary (who may call, rate limits, logging, read-only Zetaris) is W18's and is decided in its
   own record.
5. **Path C, on the judge's machine.** *Zetaris:* the sponsor's platform is a bundle from the judge's own account, with
   images in the sponsor's registry behind a sign-in. No terms in this repository say a judge may pull and run them
   (searched 26 October 2026), so `make up ZETARIS=local` runs the judge's bundle with the judge's sign-in and says so;
   the lab carries neither. Our override publishes the console and REST service on `127.0.0.1` only and attaches the REST
   service to the lab's network. A compose profile service registers the evidence definitions as the judge's Zetaris
   administrator; the key is used once and never stored. *Nightshift:* `make nightshift` orchestrates the measured Tart
   steps on macOS with Apple silicon, and on any other host says Tart is macOS-only, points to paths A and B, and exits 0.
   The installer, the email gate, the provider key and the runtime token are acts in an app window with no API; they are
   numbered steps in `docs/LAB.md`.
6. **Every address, port, image and VM name is a setting with a default** (`lab/versions.env`). No secret is baked in or
   printed. The runtime token is never read or printed by the lab: Nightshift shows it in its own window.
7. **The start page is loopback only.** It shows the lab's password, so it is published on `127.0.0.1`, named in its
   own port line of `lab/compose.yaml` like every port of the lab (ADR-0026 point 7; no setting moves it), and `make up`
   refuses a `LAB_PAGE_BIND` that names any other address, by name, before it starts a container. The renderer stages the
   page only inside the page volume, with mode 0600, and removes the staging file if it stops. (Both are repairs from the
   named review of round 1.)

## Consequences

- A judge makes one file, `lab/.env`, and runs `make up`. A review that already ran is not run again when a key changes:
  the judge opens a new pull request.
- A live review of a demo pull request reviews the sample set's change tree baked into Proven's image, not a clone of the
  pull request. That is the same set of files the pull request carries; cloning the pull request belongs to the CLI
  track.
- The guardrail's default is a model server of the judge's own. With a Nightshift or gateway generator and no guardrail
  key, a review names the guardrail as a failed service rather than skipping it (ADR-0005 stays on by default).
- Not measured by this change: a live review against any real service, `docker compose` on the files, the sponsor's
  images, Tart. They are stated as not run in the status report of the round.
