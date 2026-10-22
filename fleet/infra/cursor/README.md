# Cursor: how Proven is coded with it

Proven's product code is written by **Cursor agents**. If you use Cursor in your editor, this is not quite what you know. Nobody sits in front of a Cursor window. Each agent runs **unattended, on a remote Linux machine**. It gets a written order, works on its own copy of the repository, and returns a report. Then other agents check its work, and a human approves it before it is merged.

This page explains how that works, step by step, with the actual commands and configuration. Who does what in the fleet is explained in [SAPHAN.md](../../../SAPHAN.md); the other technologies' setups are in [`fleet/infra/`](../README.md).

## In one paragraph

[Saphan Studio](https://saphan.ai/) runs the Cursor agent as one of its **backends**, next to Claude Code, Codex and others. A **master** (a coordinating agent) writes an order for one small change. The owner signs it. Saphan Studio places the run on a free **Cursor seat**: one Cursor account, logged in on one machine. It starts Cursor's command-line agent there, headless, in a fresh copy of the repository, inside a sandbox whose only way out is Saphan Studio's **recording proxy**. Every request the agent makes to its model passes through the proxy and is recorded. When the agent is done, the master verifies the result on another machine. Two independent reviewers check it, one of them always an NVIDIA Nemotron model. The owner decides whether it is merged.

## What a Cursor agent may and may not do

| | |
|---|---|
| **may** | read and write files in its own copy of the repository; run the build, the tests and the tools its permission class allows (Go, Python, Node, gcc); commit to its own branch |
| **may not** | reach the internet: its only route out is the recording proxy, which lets through only the model's traffic (`api2.cursor.sh:443`) and what its order names, such as the Go module proxy; read any data source directly (evidence comes through Zetaris, tests use recorded fixtures); hold any credential of Proven's data sources or model providers; merge anything; review its own work |
| **model** | Grok 4.7, effort `high` (`grok-4.7-high`), named explicitly on every run; Cursor's `auto` is refused, because it may switch to another vendor's model |

## The life of one change

1. **Order.** The master writes an order for one change: what to build, how it will be checked (the *definition of done*), and what is out of scope. Every order is public in [`fleet/orders/`](../../orders/).
2. **Signature.** The owner signs the order. A signature binds the exact bytes, so an order cannot change after it is signed. If it must change, its master amends it and the owner signs it again.
3. **Placement.** Saphan Studio picks the seat. It reads every admitted Cursor seat's state and prints why the winner won.
4. **Run.** The Cursor agent works on the remote machine, behind the proxy. Every model request is recorded ([ADR-0024](../../decisions/ADR-0024-every-cursor-coding-session-is-recorded-and-published.md)).
5. **Report.** The agent writes its status report: what it built, what it ran, what failed, and what it could not check.
6. **Verify.** The master checks the result out at the exact commit, on a different machine, and runs the build and the full test suite again. It does not trust the agent's report.
7. **Review.** Two independent reviews of the same diff run in parallel, and neither sees the other's verdict. One is always an NVIDIA Nemotron model through Meterless Nightshift, with a signed receipt per request. The other is a different model family.
8. **Decision.** The owner reads the reviews and decides. Only the owner merges.

## A Cursor seat: the setup

A **seat** is one Cursor account on one machine, in its own home directory, so its login and session files never mix with
another account's. Four accounts are seated on runner-cursor, one on macm5. Seating is the owner's act:

```bash
saphan machine seat --identity cursor-fleet-3-1 --machine runner-cursor --slot 1 \
  --role executor --backend cursor --billing-class subscription \
  --config-dir '$HOME/.cursor-fleet-3-1'
saphan machine set-account --identity cursor-fleet-3-1 --account cursor-2df82fef6b0f1509
saphan config set probe.model.cursor grok-4.7-low,grok-4.7-medium,grok-4.7-high,grok-4.7-xhigh
```

- **role** `executor`: these seats write proven's code; reviews run on other model families.
- **billing class** `subscription`: Cursor's plan pays, so the scarce thing is the account's allowance; no dollar cap
  stops a run — the order's wall time does.
- **account**: an opaque digest that tells the scheduler which seats share one Cursor account. Four seats, four accounts.
- **model**: Grok takes its effort in its name: `grok-4.7-high` by default, `-medium` or `-xhigh` when an order says so.
  The bracket form `grok-4.7[effort=high]` is refused by Cursor's CLI for Grok (measured 22 October).
- **the seat's Cursor settings** (`.cursor/cli-config.json` in its home): `network.useHttp1ForAgent: true`, so the
  agent's traffic is HTTP/1.1 that the recording proxy can read.
- **network**: closed. Inside the run's own network namespace, DNS and port 443 go to a per-run door that admits only
  Cursor's own host; everything the agent sends to Cursor passes the recording proxy first (codec `cursor-connect`).
- **confinement**: Landlock on Linux; the agent can write only its copy of the repository and its own seat files, and
  cannot read the seat's key or the machine's key.

## Running one change: the commands

<!-- DRAFT: filled from the reviewer train master's answer 818c8bdf84310f1b -->

## What it costs

<!-- DRAFT: filled from the reviewer train master's answer 818c8bdf84310f1b -->

## Where it runs today

| shape | state, 22 October 2026 |
|---|---|
| native Linux, far — runner-cursor, 4 seats | **in use**: every Cursor wagon of the build runs here |
| Docker on runner-cursor | **not built**: a container boundary is composed only for local runs; the Go + Python image waits in [`../go-python/`](../go-python/) |
| macm5, local — 1 seat | seated and confinable; not yet used by a build wagon |
| macmini, far | not seated: a Cursor login over ssh fails at the macOS keychain, and a far seat's API key has no carrier yet |
