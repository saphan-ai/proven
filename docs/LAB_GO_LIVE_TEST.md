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

```
make check-flow
```

You should see `--- PASS:` for every `TestFlow…` test and a last line `ok  github.com/saphan-ai/proven/lab`. One test,
`TestFlowLeavesARealTerminalWithEchoOnAfterCtrlC`, needs a pseudo-terminal and `python3`; on a host without one it prints
`--- SKIP` with the reason, and Step 1b.3 is the check that replaces it. On macOS the two negative controls
(`…/negative_control:_the_INT_trap_removed`) skip by name: the master measured that they cannot turn red there. The
positive half of each test still runs.

## Step 1 · A clean start in sample mode

```
make clean && PROVEN_LAB_MODE=sample make up
```

You should see, as the **last 12 lines** and nothing after them (on a terminal the password line carries the password):

```
########################################
####    Your proven lab is ready    ####
########################################

To see demo pull requests scanned by proven please open:

Gitea: http://127.0.0.1:3300/user/login
Username: proven
Password: <the password>

Grafana dashboard: http://127.0.0.1:3301/d/proven-overview
Wait for every review: make demo-check
Help: https://github.com/saphan-ai/proven/blob/main/docs/WALKTHROUGH.md

>>>>>>>>  START HERE:  http://127.0.0.1:3302/lab  <<<<<<<<

```

Before the first Docker call it prints `What will happen next:` and five steps; after each step, one status line
`==> [n/5] ... (NN s)`; and it waits for every demo review before it prints this block (`make demo-check` waits again).

Earlier in the output, exactly one line `seed: no key is set; the lab reviews in sample mode`.

Failure looks like: the command stops with a line starting `up:`, `seed:` or `refused:`. Send back the last 30 lines of
`make up 2>&1 | tail -30` (redirected, the password line names a file and prints no password).

## Step 1b · The start-up flow (round 3), on a terminal and with no terminal

The flow is `make up` asking what you want to see. The explanation is [LAB.md](LAB.md#the-start-up-flow). Every command in
this step is run by you in a terminal; none of the output below was seen by the author on a Docker host (see "What this
runbook did not check"). The offline controls of Step 0 (`make check-flow`) did run it with stand-ins.

**1b.1 Ask, and answer sample.** With nothing set, and `lab/.env` holding none of the `PROVEN_LAB_` settings:

```
make clean && make up
```

The first lines are the banner, a blank line, then the question, with the cursor after it:

```
Proven: a pull-request review companion that only says what it can prove (and admits what it could not check)
Team Saphan.ai: built in public by one person and a fleet of AI agents.
No humans were harmed. Several agents were asked to try again.

Do you want to see Proven in SAMPLE mode (no keys, nothing leaves this machine) or LIVE mode (real models and live data)? [SAMPLE/live]
```

Type `sample`. You should see `up: sample mode: recorded answers, no keys, nothing leaves this machine.`, then the lab
starts and the output ends with the same 12-line block as Step 1, *Start here* first. Open the start page: under the
introduction it says `Mode: sample.`

**1b.2 A wrong answer asks again, three times, then stops.** `make up`, then type `maybe`, `perhaps`,
`nope`. After the first two you see `That is not one of: sample live. Again, please.` After the third:
`up: three answers in a row were not one of: sample, live. NEXT — answer in advance with PROVEN_LAB_MODE. Nothing was started.`
Check nothing started: `docker compose -p proven-lab ps` lists no new container.

**1b.3 Live, a judge, ours.** `make up`; answer `live`, then `y`. You see `Are you a judge of the Open Agent Hackathon? [y/N]`,
then `For Proven to prove anything live, it needs Zetaris and Meterless Nightshift. Connect your own instances, or use the ones Proven was built with, in Hua Hin, Thailand, under Marcin's desk.`
and `Which do you want? [OURS/own]`. Answer `ours`. It asks for the gateway address if `JUDGES_GATEWAY_URL` is not in
`lab/.env`.
- **By Cloudflare Access (default).** You need `cloudflared`. Without it: `up: the Access sign-in needs cloudflared, and it is not installed. NEXT — install it: …` and nothing starts.
  With it: `up: opening the sign-in for <address>. Cloudflare sends a one-time PIN to your e-mail; enter it in the browser.`
  Your browser opens; enter the PIN. Then `up: signed in. The token expires <date> UTC. To renew it, run make up again: …`
  and `up: the gateway accepted your credentials: ready.` No token is printed anywhere.
- **By token.** `PROVEN_LAB_GATEWAY_AUTH=token make up`; at `The judge token we sent you (nothing is shown as you type):`
  paste the token. Nothing appears as you type or paste. Press Enter. You see `up: the gateway accepted your credentials: ready.`
  **Press Ctrl-C at that prompt once instead:** the shell prompt comes back and what you type next is shown (echo is on).
  If it is not, run `stty sane` and send back what happened.
- A refused credential: `up: the gateway refused your credentials: HTTP 401 (<name>)`, nothing starts.
The start page then says `Mode: live via our gateway.`

**1b.4 Live, own.** `make up`; answer `live`, then `n`. You see the guide's place (`docs/LAB.md`, the *Go live* section),
the Nightshift and Zetaris lines, and, if either already answers on this machine, `up: Zetaris already answers on this machine: …`
with `Use what is already running, as it is? [Y/n]`. Then `Run the automated setup now (make up ZETARIS=local, then make nightshift)? [y/N]`.
The last line before the lab starts is, word for word: `We cannot vouch for results on a setup we do not control: this is best effort.`
The start page says `Mode: live on your own setup.`

**1b.5 No terminal, no setting: sample, nothing asked.** Run it with no controlling terminal and nothing on stdin:
`setsid make up < /dev/null 2>&1 | head -8`. You see no question, and
the log has the line `up: sample mode: recorded answers, no keys, nothing leaves this machine.`

**1b.6 Every answer in advance.** `PROVEN_LAB_MODE=live PROVEN_LAB_JUDGE=no PROVEN_LAB_OWN_SETUP=no make up` asks nothing
on a terminal either (the banner still shows). A bad value is refused by name before any container:
`PROVEN_LAB_MODE=maybe make up` prints `up: PROVEN_LAB_MODE=maybe is not one of: sample, live. NEXT — set PROVEN_LAB_MODE to one of those, or unset it. Nothing was started.`

**1b.7 The port check.** In another terminal: `python3 -m http.server 3300 --bind 127.0.0.1`. Then `make up`. You see
`up: port 3300 (GITEA_HTTP_PORT) is already in use by <the program>.` (`python3` on most hosts) and the NEXT line, and
no container starts. Stop the server and `make up` goes on. If Zetaris from the organisers' guide is running on
`localhost:3000`, the lab does not clash with it: the lab's Gitea is on 3300.

Send back for Step 1b: the first screen of 1b.1, the three lines of 1b.2, the Access or token lines of 1b.3 (never a token),
the last line of 1b.4, and the `stty` result of the Ctrl-C test.

## Step 1c · The Meterless features (round 4)

The explanation is [LAB.md](LAB.md#the-meterless-features). The offline controls ran with stand-ins (`make check-flow`,
`make check-golive`); the commands below were **not run** against a Docker host by their author.

**1c.1 Ask.** With none of `PROVEN_MARKOVIAN` and `PROVEN_LAB_SCOUT_INTENT` set in the environment or `lab/.env`:
`make up`. Answer `sample`. You should see, word for word:
```
Do you want to tweak Meterless features? [y/N]
```
Answer `y`. You should see the sentence `H-MEM (memory of refuted findings) and World Model (claims, evidence and verdicts as a graph) are on by default. Proven has no switch for either yet, so there is nothing to ask about them.`
then `Scout Intent: plans each review before the agents run (on by default; its benefit is not measured yet). Use it? [Y/n]`
Answer `n`. Then `Markovian: reviews a large change in bounded steps (on by default). Use it? [Y/n]`
Answer `y`.

**1c.2 The page says what was passed.** When the lab is up, open the start page (the *Start here* address that
`make up` prints). Under `Mode: sample.` you should read:
`Meterless features, as this lab passed them to the review: H-MEM on (always), World Model on (always), Scout Intent off, Markovian on.`
The seed's log, in the output of that same `make up`, names what it stored (names only, never a value):
`seed: stored the Meterless feature settings: PROVEN_MARKOVIAN, PROVEN_LAB_SCOUT_INTENT`.

**1c.3 Enter keeps Proven's default.** `make up` again, answer `sample`, `y`, then press Enter at both feature questions.
The page now says `Scout Intent Proven's default, Markovian Proven's default`, and the seed log has no
`stored the Meterless feature settings` line: the two secrets of 1c.2 were removed.

**1c.4 In advance, no question.** `PROVEN_MARKOVIAN=off PROVEN_LAB_SCOUT_INTENT=off make up` asks no tweak question
(the sample/live question is still asked) and the page says `Scout Intent off, Markovian off`. With no terminal
(`setsid make up < /dev/null 2>&1 | head -12`) nothing is asked and nothing is set.

**1c.5 A bad value.** `PROVEN_MARKOVIAN=maybe make up` prints
`up: PROVEN_MARKOVIAN=maybe is not one of: on, off. NEXT — set PROVEN_MARKOVIAN to one of those, or unset it. Nothing was started.`

**1c.6 The review was given it.** After 1c.4, open a new pull request as in Step 2.5 (any branch). In the runner's log of
that run (`make logs SERVICE=runner`) the review step ran; the proof the review process got the file is the offline test
`TestFeaturesTheReviewIsGivenEachSettingAsProvensOwn`, not this log. If you want to see it live, set
`PROVEN_LAB_MODE=live` with a key and compare a comment made with `PROVEN_LAB_SCOUT_INTENT=on` and with `off`: the trace
line `scout-intent:` names whether a plan was chosen.

Send back for Step 1c: the screen of 1c.1, the page line of 1c.2 and 1c.4, and the secret names of 1c.2.

## Step 1d · One review per push, and `make livecheck` for path A (round 4)

**1d.1 One push, one review.** After Step 2.5 made pull request 8, push one more commit to `try-live`:
`echo "// once" >> /tmp/lab-try/billing/discount.go`, commit, push. Wait for the run, then count the comments that carry
the new head commit:
```
head=$(git -C /tmp/lab-try rev-parse HEAD)
curl -s http://127.0.0.1:3300/api/v1/repos/proven/discount/issues/8/comments | grep -o "proven-review run=[0-9]* sha=$head" | sort | uniq -c
```
You should see exactly one line, with a count of `1`. If you see `2`, send back both lines and
`make logs SERVICE=runner`: that is the duplicate this round could not repeat. If the count is `1` but the runner's log or
the Actions tab shows two runs for that commit, the second one should say `skipped: head commit … already has a Proven review`.
The author did not measure which events Gitea 1.22.6 fires: the offline test (`TestWorkflowAHeadThatIsAlreadyReviewedIsNotReviewedOrCommentedAgain`) holds the guard, not the trigger.

**1d.2 `make livecheck` agrees with the workflow, path A.** With only `PROVEN_GENERATOR_KEY` filled in `lab/.env`
(Step 2.1 to 2.3): `make livecheck`. You should see, in this order:
```
livecheck: the workflow will run: live: NVIDIA direct, Zetaris off
generator: ready: mode nvidia, model …
guardrail: ready: verdict …
query: not ready: …
container: ready
summary: 3 ready, 1 not ready
remark: the query line is not needed in this setup: Zetaris is off in the workflow …
```
and the exit status `0` (`echo $?`). The `query` and `container` lines depend on your host (the container line needs the
inner Docker of the lab). If `generator` or `guardrail` says `not ready` while Step 2.6 shows a live review, send back
the whole output with the key blacked out.

## Step 2 · Path A, one NVIDIA key

1. `cp lab/.env.example lab/.env`
2. Open `lab/.env` and write your key after `PROVEN_GENERATOR_KEY=` (no quotes, no spaces). Save.
3. Check that the line is filled, without printing it: `grep -c '^PROVEN_GENERATOR_KEY=.' lab/.env`
   You should see `1`. A `0` means the line is empty.
4. `PROVEN_LAB_MODE=live PROVEN_LAB_JUDGE=no PROVEN_LAB_OWN_SETUP=no make up 2>&1 | grep 'seed:'`
   You should see `seed: stored as Actions secrets: PROVEN_GENERATOR_KEY` and, further on, the line
   `seed: demo repository, 7 pull requests (2 clean) and the start page are ready`.
5. Make a new pull request. The first seven are the demo's. Run these four commands one at a time; when `git push`
   asks for a user and password, use `proven` and the password from `make password`:
   ```
   git clone http://127.0.0.1:3300/proven/discount.git /tmp/lab-try
   git -C /tmp/lab-try switch -c try-live
   echo "// try" >> /tmp/lab-try/billing/discount.go
   git -C /tmp/lab-try commit -am "Try a live review" && git -C /tmp/lab-try push origin try-live
   ```
   Then open `http://127.0.0.1:3300/proven/discount/compare/main...try-live`, press **New Pull Request**, then **Create
   Pull Request**. It is pull request number 8.
6. Wait up to 15 minutes (the runner reviews three at a time). Then:
   ```
   curl -s http://127.0.0.1:3300/api/v1/repos/proven/discount/issues/8/comments | grep -oE 'live: [^\\]*|sample: no keys' | tail -1
   ```
   You should see exactly: `live: NVIDIA direct, Zetaris off`

   Failure looks like:
   - `sample: no keys` → the key was not stored. Send back step 4's output.
   - nothing printed after 15 minutes → no comment yet. Send back `make status`.
   - `live: NVIDIA direct, Zetaris off` but the review lists `generator` under failed services → the key was refused or
     rate limited (a free key is). Send back the comment: open pull request 8 and copy it, or
     `curl -s http://127.0.0.1:3300/api/v1/repos/proven/discount/issues/8/comments`.
7. **The negative check.** Empty the key: edit `lab/.env` so the line reads `PROVEN_GENERATOR_KEY=`. Then
   `PROVEN_LAB_MODE=live PROVEN_LAB_JUDGE=no PROVEN_LAB_OWN_SETUP=no make up 2>&1 | grep 'seed:'`. You should see `seed: no key is set; the lab reviews in sample mode`. Push another
   commit to `try-live` (`echo "// again" >> /tmp/lab-try/billing/discount.go`, commit, push); the new comment's first
   visible line is `sample: no keys`.

Send back for Path A: the outputs of steps 3, 4, 6 and 7, or the failure.

## Step 3 · Path B, our services through the judges' gateway

You need the gateway address and your own token. They are sent to you privately; they are not in this repository.

1. Edit `lab/.env`: empty `PROVEN_GENERATOR_KEY=`, and write the two values after `JUDGES_GATEWAY_URL=` and
   `JUDGES_GATEWAY_TOKEN=`. Save.
2. `PROVEN_LAB_MODE=live PROVEN_LAB_JUDGE=yes PROVEN_LAB_BACKEND=ours PROVEN_LAB_GATEWAY_AUTH=token make up 2>&1 | grep -E 'seed:|up:'`
   You should see `up: the gateway accepted your credentials: ready.`, then `seed: stored as Actions secrets: JUDGES_GATEWAY_URL, JUDGES_GATEWAY_TOKEN`. A refused token stops before any container with `up: the gateway refused your credentials: HTTP 401 …`.
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
2. `PROVEN_LAB_MODE=live PROVEN_LAB_JUDGE=no PROVEN_LAB_OWN_SETUP=no make up ZETARIS=local 2>&1 | grep 'zetaris:'`
   You should see, in this order:
   ```
   zetaris: Zetaris's images come from the sponsor's registry with YOUR sign-in and YOUR bundle; this lab carries neither.
   zetaris: the platform answers on http://127.0.0.1:8888 (console: http://127.0.0.1:3304, open it as localhost)
   zetaris: the evidence definitions are NOT registered yet: ZETARIS_ADMIN_API_KEY, ZETARIS_ORG_ID and ZETARIS_PROVEN_USER must all be set in lab/.env (docs/LAB.md, path C, steps 6 to 8). Then run make zetaris.
   ```
   Failure looks like: `zetaris: pulling the platform's images failed` (not signed in: Path C step 2), or
   `zetaris: ZETARIS_COMPOSE_FILE is '', which is not a file`, or `the REST service did not answer on 127.0.0.1:8888`.
3. Check the listeners are on loopback: `docker compose -p proven-zetaris ps --format '{{.Name}} {{.Ports}}'`
   Every published port starts `127.0.0.1:`. Any `0.0.0.0:` is a failure; send it back.
4. Do LAB.md Path C steps 6 and 7 in the console (`http://localhost:3304`): this is the part with no API. Put the four
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
7. `PROVEN_LAB_MODE=live PROVEN_LAB_JUDGE=no PROVEN_LAB_OWN_SETUP=no make up`, then push a commit to `try-live`. A review that has a Nightshift or NVIDIA generator as well says
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
