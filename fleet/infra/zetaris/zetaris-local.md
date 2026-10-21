# Runbook — our own Zetaris platform, step by step

Train [`proven-zetaris-local`](../../orders/TRAIN_ORDER_proven-zetaris-local.md), wagon W8 (law #39: the human ceremony,
numbered). This is what a person does to run the sponsor's Zetaris platform on a machine of ours, give Proven its own
key, and register an evidence source. Every step was run on 21 October 2026, unless it says otherwise. Environment, not
product code.

**Who does what.** Steps marked **owner** need a secret or an administrator's right. An agent never does them, never sees
the values, and never asks for them.

## What you need first

- A Linux host with Docker Engine and the Compose plugin. Ours: 8 vCPU, 12 GiB of memory, a 160 GB disk. The sponsor's
  minimum is 8 GB of memory, 4 CPUs and 15 GB of free disk. The images take about 10 GB.
- The host is a Tailscale member. No port of the platform is reachable from the internet or from the LAN.
- Your Zetaris access: a sign-in to the hosted dashboard and the registry sign-in command that came with it.

## A · Install and start

1. **owner** — Download `zetaris-platform.zip` from the hosted dashboard. Write down its sha256 (`shasum -a 256
   zetaris-platform.zip`). **Never commit the zip or the JDBC driver inside it** to any repository.
2. Copy the zip to the host and unpack it: `unzip zetaris-platform.zip`. You get the folder `zetaris-platform/`.
3. **owner** — Sign in to the sponsor's registry. Open a shell on the host that is **not logged** (not a script, not a
   recorded session). Paste the sign-in command from your own file. Wait for `Login Succeeded`. Type `exit`.
4. In `zetaris-platform/`: `bash ./preflight.sh`. Every line must say PASS.
5. Put our override file beside the sponsor's compose file, as `zetaris-platform/docker-compose.override.yml`. Never edit
   the sponsor's `docker-compose.yml`. Replace `<tailnet address>` with the host's own Tailscale IPv4 address:
   ```yaml
   # ours, never an edit of the sponsor's docker-compose.yml.
   # The UI and the REST API answer on the host's tailnet address only; JDBC (plaintext) and the API's internal login port
   # stay on loopback, as the sponsor ships them.
   services:
     ui:
       ports: !override
         - "<tailnet address>:3000:3000"
     api:
       ports: !override
         - "<tailnet address>:8888:8888"
         - "127.0.0.1:8889:8889"
   ```
   JDBC (10000) is not in the file: the sponsor's file already binds it to loopback, and it stays there.
6. Make Docker wait for the tailnet address after a reboot. Otherwise Docker starts first and cannot bind to it. Write
   this systemd drop-in as `docker.service.d/10-after-tailscale.conf` under the system's unit directory, then
   `sudo systemctl daemon-reload`:
   ```ini
   [Unit]
   After=tailscaled.service
   Wants=tailscaled.service

   [Service]
   ExecStartPre=/bin/sh -c 'for i in $(seq 1 90); do ip -4 addr show tailscale0 2>/dev/null | grep -q "inet <tailnet address>/" && exit 0; sleep 1; done; exit 0'
   ```
7. `docker compose pull`, then `docker compose up -d`.
8. `bash ./verify.sh`. Every check must say PASS (ours: 9 of 9).
9. Check the listeners: `sudo ss -ltnp | grep -E ':(3000|8888|8889|10000)\b'`. 3000 and 8888 are on the tailnet address;
   8889 and 10000 are on `127.0.0.1`. From a machine on the LAN, 3000, 8888 and 10000 must not answer.

## B · Proven's account and key

10. Open the console on your own machine through a tunnel: [`zetaris_console.sh`](README.md) `up`, then
    `http://localhost:3000`. Why: the console calls `crypto.randomUUID()`, which a browser offers only in a secure
    context (https, or `localhost`). Opened as `http://<tailnet address>:3000`, sign-in fails.
11. **owner** — Sign in as the default administrator that the sponsor's README documents. On our install it did not ask
    for a new password, and the owner kept it (W4): the console and REST answer on the tailnet only. If you expose
    anything wider, change it first.
12. **owner** — Create a non-admin user for Proven. Give it no role.
13. **owner** — Sign in as Proven's user and create its API key. The key carries its creator's permissions, and it is
    shown once.
14. **owner** — Note the organisation's **numeric** ID. The answer to the console's sign-in returns it as `orgId`. The
    32-character hex ID in the page's URL is a different value and is refused.
15. **owner** — On each machine that runs Proven (ours: macm5 and runner-cursor), write an env file with mode 600, for
    example `~/.config/proven/zetaris.env`, and source it from the shell's profile. **Never in a repository.**
    ```sh
    export ZETARIS_REST_URL=http://<tailnet address>:8888/api/v1.0
    export ZETARIS_ORG_ID=<the numeric org id>
    export ZETARIS_API_KEY=<proven's key>
    ```
16. Check from each machine. Every call carries the key as a Bearer token, the numeric org ID, and a fresh request ID:
    ```sh
    . ~/.config/proven/zetaris.env
    curl -s -X POST "$ZETARIS_REST_URL/sql-editor/sqls/run" \
      -H "Authorization: Bearer $ZETARIS_API_KEY" -H "X-Org-ID: $ZETARIS_ORG_ID" \
      -H "X-Request-ID: $(uuidgen)" -H "Content-Type: application/json" \
      -d "{\"queryId\":\"$(uuidgen)\",\"sql\":\"select 1 as one\",\"source\":\"SqlEditor\",\"limit\":100}"
    ```
    The answer has one row with `1`. With one character of the key changed, the answer is 400 `{"msg":"Not Allowed"}`.
    Proven itself reads with `POST $ZETARIS_REST_URL/query/sql/start`, which accepts SELECT only.

## C · Register an evidence source (administrator) and grant it to Proven

17. **owner** — As the administrator, send each statement on its own to `POST <REST base URL>/sql-editor/sqls/run`.
    Proven's key is refused here. Use names with the prefix `px_` for exploration objects:
    ```sql
    CREATE LIGHTNING DATABASE <DATABASE> DESCRIBE BY "<letters digits spaces only>"
    CREATE SCHEMASTORE CONTAINER <container>

    CREATE LIGHTNING REST TABLE <table> FROM <DATABASE> REQUEST(
        endpoint "<the full URL, with every parameter>",
        method "get", response_type "json", http_encoding "URLENCODED"
    ) HEADER () BODY ()

    CREATE SCHEMASTORE VIEW <view> WITH CONTAINER <container> AS
    SELECT v.<field> AS <column>, …
    FROM <container>.<table>
    LATERAL VIEW explode(<the JSON array>) AS v
    ```
    The URL and its parameters are fixed when the table is created. One table per fixed query or page
    ([ADR-0023](../../decisions/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md)). Do not run `CACHE TABLE`,
    `INSERT` or `CREATE TABLE AS` while the owner's rule "no data in the database" stands.
18. **owner** — Grant it to Proven. All three are needed; without the grants, Proven sees the database but a SELECT is
    refused:
    ```sql
    ASSIGN LIGHTNING DATABASE <DATABASE> TO USER '<proven user>'
    GRANT SELECT ON LIGHTNING DATABASE NAMESPACE <DATABASE> TO USER '<proven user>'
    GRANT SELECT ON VIEW NAMESPACE <container> TO USER '<proven user>'
    ```
19. Check with Proven's key, step 16 with `select count(*) from <container>.<view>`. Every SELECT fetches the source
    live.

## D · Restart

20. In `zetaris-platform/`: `docker compose down`, then `docker compose up -d`, then `bash ./verify.sh`. Ours: all four
    volumes kept, healthy after 44 s, 9 of 9 PASS, the ports unchanged, the `px_` definitions still answer.
21. **NEVER `docker compose down -v`, and never `--rmi all`.** They delete the data **and the platform's encryption
    keypair**.

## When a step refuses

| step | what you see | what to do |
|---|---|---|
| 4 | `preflight.sh` FAIL: not signed in to the registry | Redo step 3, then step 4. |
| 10 | the console says `crypto.randomUUID is not a function` | You opened it by its address. Use the tunnel, `zetaris_console.sh up`, and `http://localhost:3000`. |
| 11, 16 | `{"msg":"Not Allowed"}` | A wrong password or a wrong key. For REST, also check the org ID is the numeric one (step 14). |
| 17 | `{"msg":"User is not permitted to run this command"}` | Registration is the administrator's only. Proven's key cannot do it. |
| 16, 17 | 400 `Bad Select SQL` | A statement other than SELECT went to `/query/sql/start`. DDL goes to `/sql-editor/sqls/run`. |
| 19 | `does not have privileges to execute 'SELECT'` | The assignment is there, a grant is not. Redo step 18. |

## Not run

- The hosted instance with the same calls: not measured; no address or key for it was handed over.
- A Cursor leg's call (W6): it waits for a working Cursor agent.
- A backup of the volumes: decision Z4 of the train order.

## REDACTED

| what | times | how |
|---|---|---|
| the host's tailnet address | 4 | written as `<tailnet address>` |
| the host's name, user and home directory | — | left out |
| Proven's user name, the org ID and the key | — | written as placeholders |

macm5, macmini and runner-cursor stay visible by the owner's rule of 2026-10-17.
