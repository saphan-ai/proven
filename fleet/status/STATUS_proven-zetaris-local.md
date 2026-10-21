# STATUS — proven-zetaris-local · master progress

- state: in-progress
- executor: none yet — this train runs owner-act scripts and the master's measurements; no Cursor agent has run in it
- model: claude-opus-5-5 (the master's session)
- pool: not used — no Cursor turn has run
- cost: not measured — an interactive master session; no priced act has run
- duration: not measured
- last commit: 13d337a "[proven-zetaris-local] zetaris_console.sh: the file restored after a broken edit; nohup for the background tunnel"
- updated: 2026-10-21 (W1–W5, W7 and W8 done; the evidence exploration recorded as ADR-0023)

## Summary
Our own Zetaris platform runs on the Zetaris host, private to the tailnet, and passes the sponsor's `verify.sh` (9 of 9).
Proven has its own non-admin user and key, created by the owner; REST answers from macm5 and runner-cursor, and a wrong
key is refused. A restart keeps everything. The runbook is in proven. The evidence exploration is decision
[ADR-0023](../decisions/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md). W6 waits for a working far Cursor
leg; W9 is deferred by Z5.

## Done, 2026-10-21
1. The owner handed the train to the infrastructure master. His words: *"We install a local VM."* (translated from
   Polish). The master re-measured the node: **25 GiB available**, not 30, because runner-cursor now runs; about
   **27 GiB** with the file-system cache. The size became decision Z5, and the owner chose **12 GiB**.
2. The owner signed the train order (the second signing worked, after the master moved the title line to the top).
3. **W1 done** (`zetaris-host-exists`). A first run of the creation script failed: the master's script computed a memory
   size on the wrong side of an ssh call; nothing was created. The owner: *"Fix yourself."* (translated). The corrected
   run was clean. The Zetaris host runs with 8 vCPU, 12 GiB of memory, a 160 GB disk and the guest agent, and it is a
   tailnet member. On its LAN address only ssh listens.
4. **W2 done** (`zetaris-platform-installed`). Docker 29.9.2 and Compose 5.6.0 installed. The zip's sha256 recorded; the
   zip is in no repository. The owner signed in to the registry (on the second try; the session is not logged). Every
   line of `preflight.sh` said PASS. The images take 10.15 GB (Z-A2: the larger figure is the right one).
5. **W3 done** (`zetaris-platform-runs`). `docker compose pull` and `up -d`; `verify.sh` 9 of 9 PASS. Measured: the
   sponsor binds REST 8888 and JDBC 10000 to loopback and the UI 3000 to all addresses (Z-A5 was wrong for remote use).
   Our `docker-compose.override.yml` puts the UI and REST on the tailnet address only; JDBC 10000 and 8889 stay on
   loopback. Docker waits for the tailnet address after a reboot. Negative control, before the change: the LAN reached
   3000. After it: 3000, 8888 and 10000 closed on the LAN; 3000 and 8888 open on the tailnet.
6. The console fails at sign-in when opened by its address: it needs `crypto.randomUUID`, which a browser offers only in
   a secure context. At the owner's request, the fleet tool [`zetaris_console.sh`](../infra/zetaris/zetaris_console.sh)
   `up | down | status` opens it as `http://localhost:3000` through an ssh tunnel, in the background (13d337a).
7. **W4 done** (`zetaris-accounts-and-key`). The owner kept the default administrator password: *"It did not ask for a
   new password. We go on with the same."* (translated). He created Proven's non-admin user and its key. At his word,
   the master wrote `ZETARIS_REST_URL`, `ZETARIS_ORG_ID` and `███████████████` to an env file with mode 600 on macm5 and
   runner-cursor, without printing a value. Z-A7 was partly wrong: `X-Org-ID` takes the numeric org ID, not the hex ID
   in the page's URL.
8. **W5 done** (`zetaris-answers-over-rest`). From macm5 and from runner-cursor, with Proven's key: `select 1` → 200,
   `{"one":"1"}`. One character of the key changed → 400 `{"msg":"Not Allowed"}`, from both. `/query/sql/start`
   accepts SELECT only; `SHOW LIGHTNING DATABASES` goes to `/sql-editor/sqls/run`.
9. **The evidence exploration.** Proven's key may not register a source (administrator only). The owner: *"Now only
   exploration, no data in the database. Preparation only."* and *"And you may do b."* (translated). The administrator
   account registered `px_` definitions; Proven's key measured them: KEV 1754 rows (the same as the feed), EPSS one CVE
   1 row, one EPSS page 100 rows, OSV by id 1 row, KEV × EPSS 1 row; OSV by package does not work. No cache, no stored
   rows. The owner: *"OK. Save it in decisions."* (translated): [ADR-0023](../decisions/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md).
10. **W7 done** (`zetaris-data-survives`), as the owner's rule allows today: no data, so a restart of the definitions.
    `docker compose down`, then `up -d`, never `-v`: all four volumes kept, healthy after 44 s, `verify.sh` 9 of 9, the
    ports unchanged, and the `px_` KEV view answers 1754 after the restart. No script of this train contains `down -v`.
11. **W8 done** (`zetaris-runbook-and-record`). The runbook: [`fleet/infra/zetaris/zetaris-local.md`](../infra/zetaris/zetaris-local.md)
    (the order named `fleet/infra/zetaris-local.md`; it lives beside the console tool). This status page, the
    conversation page and ADR-0023 are the public record.

## Wagon state
| wagon | state | next |
|---|---|---|
| W1 `zetaris-host-exists` | done: 8 vCPU, 12 GiB, 160 GB, guest agent, tailnet member; only ssh listens on the LAN address | — |
| W2 `zetaris-platform-installed` | done: Docker, the bundle, the owner's registry sign-in, preflight all PASS, images 10.15 GB | — |
| W3 `zetaris-platform-runs` | done: `verify.sh` 9/9; UI and REST on the tailnet only, JDBC on loopback; negative control shown | — |
| W4 `zetaris-accounts-and-key` | done: Proven's non-admin user and key by the owner; the default admin password kept by his decision | — |
| W5 `zetaris-answers-over-rest` | done: `select 1` from macm5 and runner-cursor; a wrong key refused from both | the hosted instance: not measured |
| W6 `cursor-legs-reach-zetaris` | waits | a working far Cursor leg (the Cursor backend train) |
| W7 `zetaris-data-survives` | done: restart without `-v`, four volumes kept, 9/9, the `px_` definitions answer 1754 | a stored table survives, after the first data |
| W8 `zetaris-runbook-and-record` | done: the runbook and this record | kept current with each act |
| W9 `lab-rehearsal-host` | deferred by Z5 | the lab rehearsal is product work, and the 12 GiB host has no room for it |

## Blocked on / next
- **W6** waits for a far Cursor leg that works (train `proven-cursor-infra` and the Cursor backend train).
- **Owed by the owner:** Z2 (the build's default instance), Z3 (the bundle's terms), Z4 (backups); and the path for OSV
  by package (ADR-0023, Consequences).

## Not run
- W6: no far Cursor leg has run.
- W5 against the hosted instance: no address or key for it was handed over.
- W7's "a table survives" half: no data is stored, by the owner's rule of 2026-10-21.
- W9: deferred by Z5.
- The unplug test (§4, law #41): not run yet.

## REDACTED
| kind | count |
|---|---|
| the new machine's host name (written as "the Zetaris host") | 3 |
| the VM number, the cluster node and the machine's addresses (left out) | 4 |
| the zip's sha256 and the env file's path on each machine (left out) | 2 |
| the name of the variable that holds Proven's key (AGENTS.md §2) | 1 |

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17.
