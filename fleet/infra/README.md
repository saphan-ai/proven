# fleet/infra — how Proven's setup is built, and the evidence that it runs

The organisers ask every entry to show its setup, not only its result: Zetaris running on our own machine, Meterless Nightshift as the AI provider with NVIDIA's Nemotron inside it, the Meterless engines in a working feature, and Cursor as the place the code is written. This page shows where each part runs and links the evidence that it does. Every screenshot hides keys, tokens and passwords.

## The setup at a glance

```
            ┌───────────────────── our private network ─────────────────────┐
            │                                                                │
  macm5 (the owner's Mac)                    one cluster node
  ├─ the owner, the advisor, the masters     ├─ runner-cursor (Linux VM)
  ├─ a Cursor seat                           │    Cursor agents write Proven's code
  └─ Nightshift VM (macOS, Tart)             │    and run its tests
       └─ Meterless Nightshift ──────┐       │
            NVIDIA key inside it     │       └─ Zetaris VM (Linux)
                                     │            Zetaris platform, Docker Compose
                                     ▼            web UI · REST API · JDBC
                         NVIDIA hosted Nemotron
```

**Two kinds of AI, kept apart, as the organisers' guide asks.** Writing Proven uses Cursor's own models (Composer, Grok). Running Proven uses Nemotron, reached through Nightshift. No provider key is ever placed in Cursor.

**Why separate machines.** Zetaris needs its own memory and disk, so it has its own VM. Nightshift's macOS installer is not signed or notarized by Apple, so it runs inside a macOS VM with no shared folder or clipboard, away from the fleet's keys ([order](../orders/TRAIN_ORDER_proven-nightshift-vm.md)). All of them are machines we own and run, which is what *"running on your machine"* asks.

## The layers, and the evidence for each

| layer | requirement (organisers' words) | where it runs | how it is checked | evidence |
|---|---|---|---|---|
| **Data: Zetaris** | *"Zetaris and Meterless must be running on your machine"*; *"everyone runs Zetaris locally in Docker"* | Zetaris VM ([order](../orders/TRAIN_ORDER_proven-zetaris-local.md)) | `verify.sh` all checks pass; `SELECT 1` in the SQL editor; a query over REST from runner-cursor | setup runbook: [`zetaris/`](zetaris/); checks: not yet |
| **Power: Meterless Nightshift** | *"your app must support Meterless Nightshift as a primary AI provider"* | Nightshift VM on macm5 | NVIDIA in the Providers tab; every request's signed receipt names the model that answered | [`nightshift-vm/`](nightshift-vm/); 206 signed receipts, all verified, 22 October ([`nvidia/`](nvidia/)) |
| **Power: NVIDIA Nemotron** | *"at least one NVIDIA Nemotron model through Meterless Nightshift in a working feature"* | NVIDIA's hosted API, through Nightshift | the model named in each request's signed receipt, matched to the chat response `id` | [`nvidia/`](nvidia/): 145 answers from `nemotron-3.5-super-vl-120b-a12b`, 22 October |
| **Build: Meterless engines** | *"implement at least one Meterless Engine in a working feature"* | Proven's code ([ADR-0018](../decisions/ADR-0018-meterless-engines-minimum-and-their-contract.md)) | Meterless's conformance suites, or a reproducible feature test; JSONL in `logs/` | not yet: built from 22 October |
| **Build: Cursor** | *"you must build in Cursor"* | runner-cursor and macm5 | each change's order, the agent's report with its model and run id, the review | [`cursor/`](cursor/), [`../status/`](../status/), [`../LEDGER.md`](../LEDGER.md) |
| **The build image** | the environment the agents build and test in | runner-cursor | built and checked on 21 October | [`go-python/`](go-python/) |

The evidence column is filled in as each part runs: a screenshot or the pasted output of its check, dated, with keys hidden, saved under `assets/setup/`.

## What the demo shows

The organisers' guide lists what judges should see. The demo video and the deck show each of these, from this setup, not from a mock-up:

1. Proven reviewing a real pull request in the judges' lab, each finding with its evidence.
2. The Zetaris platform: its check passing, and the query a finding cites, run again in its SQL editor with the same rows.
3. Nightshift: NVIDIA in its Providers tab, a lane pinned to a Nemotron model.
4. One real request going from Proven through Nightshift to Nemotron and back, with the model named in the answer.
5. The Meterless engines: the conformance or feature test passing, and the JSONL trace in `logs/` that shows a refuted finding not returning in the next review.
6. Cursor: one change, from its order to the agent's report to the review.

## Not run

Measured so far: the build image (21 October) and the Nightshift receipts (22 October). Zetaris checks and the Meterless engines are not run yet; their rows say so until their evidence is linked.
