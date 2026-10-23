# One change on Cursor, end to end — for the README, 22 October 2026

- from: the master of `proven-reviewer-core`, for the advisor's README sections "Running one change: the commands" and
  "What it costs"
- the example: wagon W2b `agents-ask-a-real-model`, run `run-20261022-065845-5bc1fbad`, landed on `main` as `511828c`
- every line below is a line that ran on 22 October 2026. The workspace path, the far home directory and the git e-mail
  are blacked out as `<workspace>`, `<far-home>` and `███`.

## Running one change: the commands

**1 · The order is signed.** The master writes the order (a Markdown file that names the repository, the base commit, the
branch, the backend, the model, the quote and the wall time) and commits it. The owner signs it with one touch in the
browser. The master first rehearses the signature, which changes nothing:

```bash
saphan dry-run order sign <workspace>/bridge/development/active/proven-reviewer-core/signed/agents-ask-a-real-model/EXEC_ORDER.md \
  --reply-to track:proven-reviewer-core --workspace <workspace>
saphan order sign <workspace>/bridge/development/active/proven-reviewer-core/signed/agents-ask-a-real-model/EXEC_ORDER.md \
  --reply-to track:proven-reviewer-core --workspace <workspace>
```

**2 · The engine composes the run.** `saphan dispatch command` reads the signed order and the fleet's records and prints the
one `saphan run` line that will work. The master names the Cursor seat by hand (see the note below the printout):

```bash
saphan dispatch command --machine runner-cursor --identity cursor-fleet-3-1 --stream agents-ask-a-real-model \
  --role executor --model grok-4.7-high --cost-estimate-usd 3 --wall-clock 120m \
  --order <workspace>/bridge/development/active/proven-reviewer-core/signed/agents-ask-a-real-model/EXEC_ORDER.md \
  --provision <workspace>/bridge/governance/craft/COMMON.md=.provision/craft/COMMON.md \
  --provision <workspace>/bridge/governance/craft/EXECUTOR.md=.provision/craft/EXECUTOR.md \
  --provision <workspace>/bridge/skills/lang-go=.provision/skills/lang-go \
  --provision <workspace>/bridge/skills/testing-bar=.provision/skills/testing-bar \
  --workspace <workspace>
```

What it printed (trimmed):

```text
dispatch command: agents-ask-a-real-model — composed from the record (transport=ssh):
  backend    cursor  (seat cursor-fleet-3-1 on runner-cursor)
  role       executor
  mandate    model=grok-4.7-high effort=none  (source=model=flag)
  workdir    <far-home>/.saphan/workspaces/wt-agents-ask-a-real-model  (machine record, worktrees_root + stream)
  budget     6.00 USD  (2x the declared quote 3.00 — cap is derived, not chosen)
  base       "main" — the ref the merge lands on
  [pass] far-home  <far-home>/.cursor-fleet-3-1
  [pass] binary    sha256 2ccc9a8e…a233 · 2026.10.14-941492e
  [pass] workdir   measured absent; saphan run will create it atomically via dispatch-out pack transport, CUT FROM THE BASE refs/heads/main
saphan run --workspace <workspace> --stream agents-ask-a-real-model --backend cursor --role executor --track proven-reviewer-core
  --machine runner-cursor --identity cursor-fleet-3-1 --binary <far-home>/.local/bin/agent --model grok-4.7-high --base-ref main
  --branch fix/agents-ask-a-real-model --brief <workspace>/bridge/.../signed/agents-ask-a-real-model/EXEC_ORDER.md
  --workdir <far-home>/.saphan/workspaces/wt-agents-ask-a-real-model --profile execute --wall-clock 120m --env HOME --env PATH
  --provision … (the four above) --budget-usd 6 --cost-estimate-usd 3
```

**Note — the chooser (law #46).** On 22 October the engine's automatic placement (`saphan dispatch place`) could not choose
a Cursor seat: Cursor reports no allowance reading, and an absent reading is not taken as a free seat. So the master
names the seat with `--machine` and `--identity`, and `dispatch command` checks it. There is no chooser printout for a
Cursor leg yet; the printout above is the check of the seat the master named.

**3 · The leg runs.** The master runs the printed `saphan run` line. The engine copies the base to runner-cursor, starts
Cursor's agent there with Grok `grok-4.7-high`, records every model turn through its recording proxy, and brings the
branch, the transcript and the leg's `_EXEC_STATUS.md` home. W2b ran 28 min 40 s (06:58:45Z to 07:27:24Z), exit 0.

**What happens inside step 3.** Cursor's CLI never runs on its own. Before it starts, the engine builds an environment
around it on runner-cursor, and the CLI runs only inside that environment. Each item below is taken from W2b's run
record.

1. **A fresh checkout.** The engine first moves runner-cursor's copy of proven forward to the commit the master named,
   and reads the result back. Then it cuts the leg's own worktree from that commit. The leg sees only that worktree.
2. **A file-system sandbox.** `saphan confine-exec` applies a Landlock policy (Linux, ABI 4) before it starts the CLI. The
   CLI can write only to its worktree, the git paths of its own branch and its seat's Cursor directory. It can write to no
   other stream's checkout and to no other directory in the home. A seccomp filter also stops it from opening another
   process's sockets or driving another process (`ptrace`).
3. **No network except one door.** The CLI and every process it starts run in their own network namespace
   (`unshare --user --net`), so a direct connection out goes nowhere. The one way out is this run's **egress door**, a
   per-run socket served by `saphan egress-door`. The door judges every connection against the fleet's egress policy for
   Cursor (`baseline-cursor-v1`). It opens before the CLI starts and closes when the CLI ends.
4. **A recording proxy in front of the model.** `saphan guardrail-proxy` stands on loopback in front of
   `api2.cursor.sh` and speaks Cursor's own protocol. The CLI is started with `-e <proxy> --agent-endpoint <proxy>`, so
   every model turn goes through the proxy. Every exchange goes to an audit file, and the whole model transcript is
   sealed and brought home with the run (W2b: 1.8 MB). If the session cannot be recorded, the run does not start.
5. **Only what it needs.** The CLI is given `HOME` and `PATH` and nothing else. Cursor's own endpoint variables are never
   passed on, and no key value travels on the ssh hop. The order, the craft files and the skills are placed under
   `.provision/` in the worktree.
6. **Git guarded.** A git hook on runner-cursor refuses `fetch`, `merge`, `pull`, `push` and `reset --hard` for this leg.
   The leg can commit to its own branch and to nothing else.
7. **Then the CLI.** Only now does Cursor's `agent` start in that worktree, with Grok `grok-4.7-high`. W2b: 60 model
   turns, 137 tool calls, outcome `ok`.

**4 · The master verifies.** The master checks the returned branch in a fresh worktree: `go build ./...`, `go vet ./...`,
`go test -count=1 ./...`, the order's negative controls, and a live review through Nightshift. The result is a
`MASTER_VERIFY` file, landed in the record:

```bash
saphan channel put --stream agents-ask-a-real-model --kind master-verify --round 1 \
  <workspace>/bridge/development/active/proven-reviewer-core/MASTER_VERIFY_agents-ask-a-real-model_round-1_2026-10-22.md \
  --workspace <workspace>
```

**5 · The owner lands it.** W2b was a master-verified wagon (no named review), so its landing is one line, run by the owner:

```bash
saphan accept agents-ask-a-real-model --reason "<the reason, pre-filled by the master>" --workspace <workspace>
```

A wagon with a named review lands with `saphan land <stream> --review-decision accept --review-delivery <id>` instead.

## What it costs

The engine's run record for `run-20261022-065845-5bc1fbad`:

| field | value |
|---|---|
| quote (in the order) | 3.00 USD |
| cap | 6.00 USD (always 2 × the quote) |
| billing class of the seat | `subscription` |
| tokens (Cursor's own count) | 31,446 in · 21,891 out · 4,023,040 read from cache |
| cost recorded | **0 USD** — `subscription-covered — marginal USD 0, energy not metered` |
| price row used | `bridge/registry/pricing.json#cursor-grok-4.7-high-subscription` |

- **Why 0.** The seat's billing class is `subscription`: the Cursor plan pays. Under a subscription the scarce thing is the
  account's allowance, so a run is bounded by its wall time (the order's `wall`), not by dollars.
- **The quote still matters.** Every order declares one, the cap is derived from it, and the engine compares quote and
  actual at the close of every run.
- **Where the price table comes from.** `bridge/registry/pricing.json` in the fleet's own records holds one row per
  backend, model and billing class. The four Cursor Grok rows (`grok-4.7-low`, `-medium`, `-high`, `-xhigh`) were added on
  22 October as `subscription-covered`. A run cites the row it used, with the table's commit (`pricing_ref` in the run
  record).
