# How proven's Cursor agents are set up — for the README, 22 October 2026

- from: the infrastructure master (`proven-cursor-infra`), for the advisor's README section
- every value below is the one in use on 22 October 2026, read from the fleet's registry, a seat's own files or a run's record

## 1 · How a Cursor seat is configured

A **seat** is one Cursor account on one machine, in its own home directory, so its login and session files never mix with
another account's. Four accounts are seated on runner-cursor, one on macm5. Seating is an owner act; the lines are:

```bash
saphan machine seat --identity cursor-fleet-3-1 --machine runner-cursor --slot 1 \
  --role executor --backend cursor --billing-class subscription \
  --config-dir '$HOME/.cursor-fleet-3-1'
saphan machine set-account --identity cursor-fleet-3-1 --account cursor-2df82fef6b0f1509
saphan config set probe.model.cursor grok-4.7-low,grok-4.7-medium,grok-4.7-high,grok-4.7-xhigh
```

- **role** `executor`: these seats write proven's code; reviews run on other model families.
- **billing class** `subscription`: Cursor's plan pays; the scarce thing is the account's allowance, not money, so no
  dollar cap stops a run — the clock (each order's wall time) does.
- **account**: an opaque digest that tells the scheduler which seats share one Cursor account, so its live-run ceiling is
  counted per account. Four seats, four accounts.
- **model**: Grok 4.7. Cursor takes the effort in the model's name — `grok-4.7-high` is our default; an order may step a
  wagon to `-medium` or `-xhigh`. Measured 22 October: the bracket form `grok-4.7[effort=high]` is refused by Cursor's CLI
  for Grok, so no effort is set apart.
- **the seat's own Cursor settings** (`.cursor/cli-config.json` in its home): `network.useHttp1ForAgent: true`, so the
  agent's traffic is plain HTTP/1.1 that the recording proxy can read.
- **recording**: every Cursor turn passes through Saphan Studio's recording proxy in front of `api2.cursor.sh` (codec
  `cursor-connect`); a run whose stream cannot be recorded is refused before it starts. The recording comes home with
  the run and is published here ([ADR-0024](../../../docs/adr/ADR-0024-every-cursor-coding-session-is-recorded-and-published.md)).
- **confinement**: on Linux the leg runs under Landlock (it can write only its worktree and its own seat files; the seat's
  key file and the machine's key are hidden from it) and in its own network namespace.

## 2 · The network a Cursor leg sees

A Cursor leg's network is closed by default: inside its namespace, DNS and port 443 are redirected into a per-run door
that admits exactly the hosts of Cursor's baseline — today one, `api2.cursor.sh:443` — and every other connection is
refused by the wall, not by the agent's goodwill. Everything the agent says to Cursor goes through the recording proxy
first. A run that needs more (the Go module proxy, a data source) gets it named in its order; opening the whole network
is an owner act written on the run's record.

## 3 · The four shapes and where they stand (22 October 2026)

| shape | state today |
|---|---|
| **native Linux, far** — runner-cursor, 4 seats | **in use.** Every Cursor wagon of the build runs here (W1 the command line, W1b). |
| **Docker on runner-cursor** | **not built.** The engine composes a container boundary only for local runs; a seated container run on a far host is refused. The Go + Python image is in [`fleet/infra/go-python/`](../infra/go-python/), waiting for that. |
| **macm5, local** — 1 seat | **seated, confinable** since the darwin confinement for Cursor landed; not yet used by a build wagon. |
| **macmini, far** | **not seated.** A Cursor login on a far Mac over ssh fails at the macOS keychain, and a far seat with an API key has no carrier for it yet. |

## REDACTED

Nothing was blacked out. Credential names and values, addresses and internal paths are left out rather than written;
seat names stay visible by the owner's rule of 22 October.
