# Owner's test: the three ways to go live

One pass, in order, no hunting. Each step has the command, what you should see, what a failure looks like, and what to
send back. Run every command from the top directory of the repository. The explanation of each path is in
[LAB.md](LAB.md#go-live); the decision is [ADR-0052](adr/ADR-0052-the-lab-goes-live-three-ways.md).

**What this runbook did not check.** It was written on a host with no Docker, no network and no key. The offline checks
(Step 0) were run. Every command below that starts a container, calls a service or opens a VM was **not run** by its
author: the outputs are what the scripts print, read from the scripts. If a real output differs, that difference is the
finding; send it back.

A key never appears in any output below. If one does, stop and send back the line with the value blacked out.

## Step 0 · Offline check (about 10 seconds, no Docker)

```
make check-golive
```

You should see one line that starts `ok` and names `github.com/saphan-ai/proven/lab`.
A failure prints `--- FAIL: TestGoLive…` with a reason. Send back that output.

## Step 1 · A clean start in sample mode

```
make clean && make up
```

You should see, as the **last 12 lines** and nothing after them (on a terminal the password line carries the password):

```
Start here: http://127.0.0.1:3002/lab
########################################
####    Your proven lab is ready    ####
########################################

To see demo pull requests scanned by proven please open:

Gitea: http://127.0.0.1:3000/login
Username: proven
Password: <the password>

Grafana dashboard: http://127.0.0.1:3001/d/proven-review
```

Earlier in the output, exactly one line `seed: no key is set; the lab reviews in sample mode`.

Failure looks like: the command stops with a line starting `up:`, `seed:` or `refused:`. Send back the last 30 lines of
`make up 2>&1 | tail -30` (redirected, the password line names a file and prints no password).

## Step 2 · Path A, one NVIDIA key

1. `cp lab/.env.example lab/.env`
2. Open `lab/.env` and write your key after `PROVEN_GENERATOR_KEY=` (no quotes, no spaces). Save.
3. Check that the line is filled, without printing it: `grep -c '^PROVEN_GENERATOR_KEY=.' lab/.env`
   You should see `1`. A `0` means the line is empty.
4. `make up 2>&1 | grep 'seed:'`
   You should see `seed: stored as Actions secrets: PROVEN_GENERATOR_KEY` and, further on, the line
   `seed: demo repository, 7 pull requests (2 clean) and the start page are ready`.
5. Make a new pull request. The first seven are the demo's. Run these four commands one at a time; when `git push`
   asks for a user and password, use `proven` and the password from `make password`:
   ```
   git clone http://127.0.0.1:3000/proven/discount.git /tmp/lab-try
   git -C /tmp/lab-try switch -c try-live
   echo "// try" >> /tmp/lab-try/billing/discount.go
   git -C /tmp/lab-try commit -am "Try a live review" && git -C /tmp/lab-try push origin try-live
   ```
   Then open `http://127.0.0.1:3000/proven/discount/compare/main...try-live`, press **New Pull Request**, then **Create
   Pull Request**. It is pull request number 8.
6. Wait up to 15 minutes (the runner reviews three at a time). Then:
   ```
   curl -s http://127.0.0.1:3000/api/v1/repos/proven/discount/issues/8/comments | grep -oE 'live: [^\\]*|sample: no keys' | tail -1
   ```
   You should see exactly: `live: NVIDIA direct, Zetaris off`

   Failure looks like:
   - `sample: no keys` → the key was not stored. Send back step 4's output.
   - nothing printed after 15 minutes → no comment yet. Send back `make status`.
   - `live: NVIDIA direct, Zetaris off` but the review lists `generator` under failed services → the key was refused or
     rate limited (a free key is). Send back the comment: open pull request 8 and copy it, or
     `curl -s http://127.0.0.1:3000/api/v1/repos/proven/discount/issues/8/comments`.
7. **The negative check.** Empty the key: edit `lab/.env` so the line reads `PROVEN_GENERATOR_KEY=`. Then
   `make up 2>&1 | grep 'seed:'`. You should see `seed: no key is set; the lab reviews in sample mode`. Push another
   commit to `try-live` (`echo "// again" >> /tmp/lab-try/billing/discount.go`, commit, push); the new comment's first
   visible line is `sample: no keys`.

Send back for Path A: the outputs of steps 3, 4, 6 and 7, or the failure.

## Step 3 · Path B, our services through the judges' gateway

You need the gateway address and your own token. They are sent to you privately; they are not in this repository.

1. Edit `lab/.env`: empty `PROVEN_GENERATOR_KEY=`, and write the two values after `JUDGES_GATEWAY_URL=` and
   `JUDGES_GATEWAY_TOKEN=`. Save.
2. `make up 2>&1 | grep 'seed:'`
   You should see `seed: stored as Actions secrets: JUDGES_GATEWAY_URL, JUDGES_GATEWAY_TOKEN`.
3. Push a commit to `try-live` as in Step 2.5 (`echo "// gateway" >> …`, commit, push). The pull request's workflow runs
   again. Wait, then run the `curl` of Step 2.6 (the pull request is still number 8).
   You should see exactly: `live: NVIDIA via the judges' gateway, Zetaris on`

   Failure looks like:
   - `sample: no keys` → as in Step 2.
   - the comment lists `nightshift` or `zetaris` under failed services with a refusal → the gateway refused your token
     (expired, or not yours). That is the gateway's refusal, by name. Send it back.
   - a refusal that names `guardrail` → the guardrail's default is your own model server; see LAB.md, Path B.
4. **Expected to be unfinished:** the gateway is another wagon's work (W18). If it is not up yet, steps 1 and 2 still
   pass and step 3 fails at the gateway; say so.

Send back for Path B: the outputs of steps 2 and 3, or the failure.

## Step 4 · Path C, Zetaris and Nightshift on your machine

Only if you have your own Zetaris bundle and sign-in. Nightshift only on a Mac with Apple silicon.

### C1 · Zetaris

1. Do LAB.md Path C steps 1 to 4: unpack your bundle, sign in to the sponsor's registry, run the bundle's
   `bash ./preflight.sh` (every line PASS), set `ZETARIS_COMPOSE_FILE=` in `lab/.env` to the full path of its
   `docker-compose.yml`.
2. `make up ZETARIS=local 2>&1 | grep 'zetaris:'`
   You should see, in this order:
   ```
   zetaris: Zetaris's images come from the sponsor's registry with YOUR sign-in and YOUR bundle; this lab carries neither.
   zetaris: the platform answers on http://127.0.0.1:8888 (console: http://127.0.0.1:3004, open it as localhost)
   zetaris: the evidence definitions are NOT registered yet: ZETARIS_ADMIN_API_KEY, ZETARIS_ORG_ID and ZETARIS_PROVEN_USER must all be set in lab/.env (docs/LAB.md, path C, steps 6 to 8). Then run make zetaris.
   ```
   Failure looks like: `zetaris: pulling the platform's images failed` (not signed in: Path C step 2), or
   `zetaris: ZETARIS_COMPOSE_FILE is '', which is not a file`, or `the REST service did not answer on 127.0.0.1:8888`.
3. Check the listeners are on loopback: `docker compose -p proven-zetaris ps --format '{{.Name}} {{.Ports}}'`
   Every published port starts `127.0.0.1:`. Any `0.0.0.0:` is a failure; send it back.
4. Do LAB.md Path C steps 6 and 7 in the console (`http://localhost:3004`): this is the part with no API. Put the four
   values in `lab/.env`.
5. `make zetaris 2>&1 | grep 'register:'`
   You should see `register: ok: …` lines, the last being
   `register: the evidence definitions are in place and granted to <your Proven user>`.
   A line `register: the platform answered HTTP 4xx` names the statement and the platform's own message: a wrong key,
   organisation id or user. A `register: ok:` line is expected to be missing for a definition that already exists on a
   second run; the grants must all say ok.
6. `make livecheck ZETARIS=local`
   You should see a line `query: ready: rows <a number above 0>` and, with other keys unset, `not ready` lines for the
   other services. `summary:` ends the output. A `query: not ready: …` line is a failure; send it back.
7. `make up`, then push a commit to `try-live`. A review that has a Nightshift or NVIDIA generator as well says
   `Zetaris on` in its first line.

### C2 · Nightshift (macOS, Apple silicon)

1. On a **Linux** host first: `make nightshift; echo exit=$?`
   You should see exactly:
   ```
   nightshift: Tart runs Nightshift's macOS VM and exists for macOS on Apple silicon only; this host is Linux/x86_64.
   nightshift: NEXT — go live another way, both work here: path A (one free NVIDIA key) or path B (our services through the judges' gateway). docs/LAB.md, "Go live".
   exit=0
   ```
   (the host and architecture are yours). Anything else, or `exit=` not 0, is a failure.
2. On the Mac: `make nightshift`. The first run, with no Tart, stops with `nightshift: Tart is not installed. NEXT — run: …`
   and the command. Run that command (it asks for your password once), then `make nightshift` again.
3. The second run clones the VM and starts it. It prints the host check, `✔` lines, then
   `nightshift: the VM nightshift is running at <an address>`. The VM's window opens.
4. Do LAB.md Path C steps 11 to 14 inside the VM (installer, access gate, NVIDIA key in Providers). Then
   `make nightshift` again. You should see `nightshift: Nightshift answers at http://<the VM's address>:8088` and the line
   to put in `lab/.env`.
   Failure looks like: `nightshift: Nightshift is not answering at … within 300s` (the app is not started in the VM).
5. Put the address and the runtime token (Nightshift's window shows it) in `lab/.env`, `make up`, push a commit to
   `try-live`. The first line is `live: NVIDIA via Nightshift, Zetaris on` (or `Zetaris off`).
   If the lab cannot reach the VM's address from its containers, the review names `nightshift` as a failed service;
   send that back. It was not measured.

Send back for Path C: the outputs of C1 steps 2, 3, 5 and 6, and C2 steps 1, 3 and 4, or the failure.

## Step 5 · The negative controls (no Docker needed)

1. A start page on another address is refused before anything starts:
   `LAB_PAGE_BIND=0.0.0.0 make up`
   You should see one line, `refused: LAB_PAGE_BIND=0.0.0.0 is not a loopback address. The start page carries the lab's password and is published on 127.0.0.1 only. NEXT — unset LAB_PAGE_BIND, or set it to 127.0.0.1.`, and make's own error. No container starts.
2. A broken `lab/.env` line is refused by line number, never by content: add a line `not a setting` to `lab/.env`, then
   `make up`. You should see `lab/.env: refused line <n> (a setting is NAME=value)`. Remove the line.

## What to send back, all in one message

Which steps passed, the outputs asked for above, the first line of each live comment, and for any failure the exact
output. Never a key, a token or the password.
