# TRAIN ORDER — proven-zetaris-local

- master: MASTER (claude-code, claude-opus-5-5, seat macm5/██████████████), the infra master of `proven-cursor-infra`, handed this train by the owner on 2026-10-21 ("we install a local VM"), mode: economy
- status: ACTIVE — amended by the master on 2026-10-21 (§1a, W1 size, Z5); waits for the owner's signature
- track: proven-zetaris-local
- repo: proven (public record and the runbook), ████████████ (the host's brief and report)
- base: main at the cut
- author: ADVISOR of the hackathon track "The Agent That Can Explain Why"

## §0 THE OWNER, VERBATIM

2026-10-20, to the advisor, after pasting the sponsor's instructions for running the Zetaris platform locally and the
address of the hosted instance's dashboard (written in Polish; given in the advisor's translation):

> *"We have Zetaris. We got access."*

> *"Let's prepare our own local cluster for it."* (written in English)

> *"We can put the containers on a new machine close to the runner, on the same cluster node."* (written in English,
> after ADR-0017; this answers Z1, and the advisor read it as covering the judges' lab rehearsal too, W9)

## §1 WHAT THIS TRAIN IS

proven's evidence is queried through Zetaris ([ADR-0002](../decisions/ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md)).
The sponsor gives every participant a hosted instance, and also the whole platform as a Docker Compose stack to run on a
machine you control. When this train is done: the platform runs on a machine of ours, reachable from the Cursor agents'
machines over the private network and from nowhere else; it passes the sponsor's own checks; proven has its own
non-admin account and API key, created by the owner; a query answers over REST from the machines the build runs on; a
refused query is shown too; the data survives a restart; and a runbook in proven says how a person sets it up and what to
do when it refuses.

**Why our own, when a hosted instance exists.** The build should not stop when a shared instance is slow or down, and the
evidence sources (OSV, CISA KEV, FIRST EPSS, git history; [ADR-0004](../decisions/ADR-0004-first-evidence-sources.md)) are
loaded and reloaded many times during the build. The hosted instance stays as the second place to run the same query
before a finding is published. Which of the two is the build's default is decision Z2.

**What the advisor read on 2026-10-20. These are CLAIMS FOR YOU TO TEST, not ground (law #44). Re-measure each one before
you build on it. Where you find the advisor wrong, write it in the wagon's `master-was-wrong:` field.**

| # | claim (advisor, READ 2026-10-20, not measured) | source | instrument |
|---|---|---|---|
| Z-A1 | The platform needs Docker with 8 GB of memory and 4 CPUs, 15 GB of free disk, and 15 to 30 minutes to set up. | the sponsor's "Run the platform locally" page, pasted by the owner | `preflight.sh` on the chosen host |
| Z-A2 | Images are about 4 GB to download; the page's last step speaks of "the 10 GB of images". The two numbers are probably compressed and unpacked size. | same page | `docker system df` after the pull |
| Z-A3 | On Apple Silicon the page asks for Rosetta, so the images are probably built for x86-64 only. | same page | `docker manifest inspect` on each image, after the registry sign-in |
| Z-A4 | The stack is six containers: postgres, init, driver, api, ui, bootstrap. Data lives in four volumes: pgdata, customer-data, driver-data, driver-logs. | same page, the sample output | `docker compose ps -a`; `docker volume ls` |
| Z-A5 | Ports: the web UI on 3000, the REST API on 8888 (base `/api/v1.0`), JDBC on 10000. | the same page (3000, 10000); the sponsor's getting-started guide, `docs/connections/client-setup.md` (8888) | `ss -ltnp` on the host |
| Z-A6 | `docker compose down -v` deletes the data **and the platform's encryption keypair**. | same page, step 10 | read only: never run it to check |
| Z-A7 | REST needs a user-created API key, the organisation's numeric ID, and a fresh request ID per call; it needs no Java. The key is shown once, carries its creator's permissions, and the guide says never to let an agent create or fetch it. | getting-started, `client-setup.md` and `scripts/HOWTO.md` | one REST call from the build machines (W5) |
| Z-A8 | The OpenAPI specification is served at `:8888/redoc/docs.yaml` behind HTTP Basic with the user's name and password; it does not accept the API key. | `client-setup.md` | `curl -u` from the host |
| Z-A9 | The Proxmox node that holds runner-cursor had about 30 GiB of memory and 616 GB of disk free after runner-cursor was sized. | the master's §1a A12 of `proven-cursor-infra`, 2026-10-19, minus runner-cursor's 8 GiB | re-measure the node; it may have changed since |

## §1a THE MASTER'S RE-MEASUREMENT (MASTER, 2026-10-21 ~01:40 UTC, read-only)

| # | verdict | what the master measured |
|---|---|---|
| Z-A9 | **WRONG (changed)** | The node now has **25 GiB available** (`free -g`: 123 total, 97 used), not 30: runner-cursor runs. The ZFS cache is capped at 6.0 GiB and may shrink to 3.9 GiB, so about **27 GiB** is the real headroom. Guests declare 110,592 MiB (108 GiB) of 123 GiB. 32 threads; `rpool` 621 G free. VMID ███ is free cluster-wide; workbench IP `███` is silent. |
| Z-A1–A8 | UNMEASURED | They need the bundle and the registry sign-in (W2–W3). |

## §2 WAGONS

### W1 · `zetaris-host-exists` — the Zetaris host exists
- **what:** a new Linux VM on the node that holds runner-cursor (Z1, decided by the owner 2026-10-20), modelled on
  runner-cursor. It carries two Compose stacks: the Zetaris platform, and the judges' lab rehearsal (W9). Proposed size,
  for the master to test against Z-A1 and Z-A9 and print the arithmetic (law #46): 8 vCPU, 24 GiB of memory (16 for the
  platform, twice the sponsor's minimum because evidence tables are loaded into it, and 8 for the lab), 160 GB of disk.
  **The master's arithmetic (2026-10-21):** 24 GiB does NOT fit: 27 GiB of real headroom − 24 = 3 GiB left for the node,
  and the guests would declare 132 GiB of 123. Candidates: **16 GiB** (platform 12 + lab 4) leaves about 11 GiB, guests
  declare 124 of 123 GiB (1 GiB over, today's guests use about 90 of their 108); **12 GiB** (platform 12, the lab
  rehearsal placed elsewhere) leaves about 15 GiB, guests declare 120 of 123. vCPU 8 of 32 and disk 160 of 621 G fit
  either way. The size is decision Z5.
  **If the node cannot spare 24 GiB beside runner-cursor and leave room for its other guests, say so with the numbers
  and propose the split; do not shrink the platform below the sponsor's 8 GB.** A tailnet member like runner-cursor. No
  port reachable from the internet.
- **owner act:** the creation script, rehearsed by the master with `DRY=1` and the output pasted (law #33).
- **dod:** the host answers; its size matches the printed arithmetic; it is on the tailnet; nothing listens on a public
  address.
- **master-was-wrong:** —

### W2 · `zetaris-platform-installed` — the platform is installed
- **what:** Docker Engine and the Compose plugin on the host; the platform bundle unpacked; the sponsor's `preflight.sh`
  passes.
- **owner acts:** the registry sign-in and the download of `zetaris-platform.zip` from the hosted dashboard. **The owner
  holds the key that registers a local setup** (2026-10-20). He types it on the host himself; it never appears in a
  script, a file in any repository, a prompt, an order or a report. The master's script stops at that step and resumes
  after it, and records only that the sign-in succeeded. The master carries the zip to the host and records its sha256. **The zip and the JDBC driver inside it are
  never committed to any repository** (Z3).
- **dod:** every line of `preflight.sh` says PASS, pasted; the zip's sha256 recorded; Z-A3 answered.
- **master-was-wrong:** —

### W3 · `zetaris-platform-runs` — the platform runs, private only
- **what:** `docker compose pull`, `docker compose up -d`, then the sponsor's `verify.sh`.
- **The ports are bound to the private network only.** Read what the compose file binds (Z-A5). If it binds all
  addresses, restrict it with a compose override file of ours (never an edit of the sponsor's file) or a host firewall,
  allowing the tailnet and nothing else.
- **dod:** `verify.sh` all PASS, pasted; `ss -ltnp` pasted; from a machine outside the tailnet, ports 3000, 8888 and
  10000 do not answer. **Negative control (law #27):** remove the restriction for one probe, show that the outside probe
  then answers, and put it back.
- **master-was-wrong:** —

### W4 · `zetaris-accounts-and-key` — proven's account and key
- **what, all owner acts, none by an agent:** sign in as the default admin and change its password; create a non-admin
  user for proven with query rights only; create that user's API key; note the organisation's numeric ID.
- **Where the key goes:** an environment variable on the machines that run proven, by the owner's hand. Never in a file
  in any repository, never in a prompt, never in an order (Z-A7; [ADR-0013](../decisions/ADR-0013-no-hardcoded-settings.md)
  point 3).
- **dod:** the owner says it is done; the master records the variable names, never their values.
- **master-was-wrong:** —

### W5 · `zetaris-answers-over-rest` — Zetaris answers over REST
- **what:** from runner-cursor and from macm5, over REST with proven's key: `SELECT 1`, then `SHOW LIGHTNING DATABASES`.
  The same two calls against the hosted instance, so the build has both measured.
- **Negative control (law #27):** the same call with one character of the key changed is refused; paste the refusal.
- **dod:** both answers and the refusal pasted, per machine; the REST base URL recorded as a setting, not a constant
  (ADR-0013).
- **master-was-wrong:** —

### W6 · `cursor-legs-reach-zetaris` — a Cursor leg reaches Zetaris
- **what:** a Cursor leg can reach the platform. This is the same question as W8 of `proven-cursor-infra`: the platform
  host becomes an admitted data host. Until the engine can do that, the gap is named, and the fallback is the owner's
  recorded `--egress-open`.
- **dod:** one leg's `SELECT 1` over REST, or the named gap with the engine change it waits for.
- **master-was-wrong:** —

### W7 · `zetaris-data-survives` — the data survives a restart
- **what:** `docker compose down`, then `up -d`, and a table created before the restart is still there.
- **Forbidden in every script of this train:** `docker compose down -v` and `--rmi all` (Z-A6: they destroy the data and
  the encryption keypair).
- **Backups:** decision Z4.
- **dod:** the table before and after, pasted; a grep of every script of this train for `down -v` prints nothing.
- **master-was-wrong:** —

### W8 · `zetaris-runbook-and-record` — the runbook and the public record
- **what:** the human ceremony of W2 and W4 written as a numbered runbook in proven, `fleet/infra/zetaris-local.md`: what
  the owner does, in which order, and what to do when a step refuses (law #39). The public copy of this order and its
  status stay current, in English, with nothing internal in them.
- **dod:** the runbook exists and a second person could follow it; the deny-list grep over every public file prints
  nothing.
- **master-was-wrong:** —

### W9 · `lab-rehearsal-host` — the host is ready for the lab rehearsal
- **what:** the judges' lab ([ADR-0017](../decisions/ADR-0017-a-gitea-lab-for-the-judges.md)) is rehearsed on this VM
  every day, from a clean checkout, with the one command a judge runs. Here, on a machine that does nothing else, the lab's
  runner may use the host's Docker socket; that says nothing about what the judges' copy does, which ADR-0017 leaves to
  the build. The lab and the Zetaris platform use separate Compose project names, networks and ports, so stopping one
  never touches the other.
- **This wagon prepares the host only.** The lab itself is product work, written from 22 October.
- **dod:** both Compose projects listed side by side; the lab's ports do not collide with 3000, 8888 and 10000; a stop
  of the lab leaves `verify.sh` all PASS.
- **master-was-wrong:** —

## §3 DECISIONS OWED BY THE OWNER

- **Z1 — the host. DECIDED by the owner 2026-10-20:** a new machine next to runner-cursor, on the same cluster node (§0).
- **Z2 — the build's default instance.** Recommended: ours for the build, the hosted one as the second place a finding's
  query is run before publication. Both are settings (ADR-0013).
- **Z3 — the bundle's terms.** Recommended: never commit the bundle or the driver, and read the sponsor's terms for
  whether they may be redistributed before anything is said about it in public.
- **Z4 — backups.** Recommended: none during the hackathon, recorded as such, because every table is reloaded from the
  public sources by scripts. Alternative: a nightly copy of the volumes.

- **Z5 — the size (added by the master 2026-10-21).** Recommended: **12 GiB**, VMID ███, IP `███`: the platform gets 12
  (one and a half times the sponsor's minimum) and the node keeps a margin; the lab rehearsal (W9) moves to a later
  decision, because the lab is product work from 22 October. Alternative: 16 GiB with the lab on this VM and the node's
  guests 1 GiB over its memory.

## §4 DEFINITION OF DONE

- The platform runs on our host, passes `verify.sh`, and is reachable from the tailnet only, with the negative control
  shown.
- proven's own non-admin account and API key exist, created by the owner; no key value appears in any file or message.
- `SELECT 1` answers over REST from runner-cursor and macm5, against ours and the hosted instance; a wrong key is refused.
- A Cursor leg reaches the platform, or the gap is named.
- A restart keeps the data; no script can delete the volumes.
- The runbook is in proven, and the public record is current.
- Unplug test (law #41): stop the platform's containers and show that W5's call fails with a named error, not an empty
  answer that reads as "no rows".

## §5 COST

- Every act of this train is local energy: a named zero, except Cursor legs in W6, which are priced and run only with the
  owner's consent at the call (law #49).
- Estimate: the master's own session, and one Cursor turn for W6.

## REDACTED

| what was blacked out | times | what it was |
|---|---|---|
| seat identity | 1 | the master's seat on macm5 |
| infrastructure identifiers | 2 | the planned VM number and the last part of its planned address |
| private repository names | 1 | our infrastructure repository (blacked out on 2026-10-21; an earlier copy showed it) |

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17 (the Cursor executors work there).
