# fleet/infra/zetaris

Fleet tooling for proven's self-hosted Zetaris platform. Environment, not product code (2026-10-21).

The whole setup, step by step, is the runbook [`zetaris-local.md`](zetaris-local.md).

## zetaris_console.sh — the console on your machine

The Zetaris web console calls `crypto.randomUUID()`. Browsers offer it only in a "secure context": https, or
`http://localhost`. Opened as `http://<address>:3000`, the console fails at sign-in with
`crypto.randomUUID is not a function`. This tool opens an ssh tunnel, so the console runs as `http://localhost:3000`.

1. Write your settings once, to `~/.config/proven/zetaris-console.env` (mode 600), or export them:
   ```
   ZETARIS_SSH=<user>@<the machine that runs the platform>
   ZETARIS_UI=<the console's address as that machine sees it, e.g. 127.0.0.1:3000>
   ```
2. `fleet/infra/zetaris/zetaris_console.sh up` — the tunnel runs in the background (it survives closing the terminal).
3. Open `http://localhost:3000`.
4. `zetaris_console.sh status` — up or down, and whether the console answers. `zetaris_console.sh down` closes it.

When it refuses:
- `port 3000 is taken` — something else listens there; stop it, or set `ZETARIS_LOCAL=<another port>`.
- `the tunnel did not open` — check `ssh <ZETARIS_SSH>` works on its own.
- `console answers 000` — the platform is down on the host; run its `verify.sh` there.
