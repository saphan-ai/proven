# The judges' lab

One command builds Proven and starts a Gitea lab on this machine. The lab binds to `127.0.0.1`. Nothing in it is published on another address: each port line of `lab/compose.yaml` names `127.0.0.1` itself, so starting that file directly, with any environment, publishes nothing beyond loopback. No setting chooses the address. The old `LAB_BIND` is retired: `make up` refuses it when it is set (an empty value too), names it, and starts no container. The lab does not serve SSH. It publishes five ports, all on that address: Gitea's web UI (`GITEA_HTTP_PORT`, 3300), Grafana's (`LGTM_GRAFANA_PORT`, 3301), the start page (`LAB_PAGE_PORT`, 3302), the documentation site (`LAB_DOCS_PORT`, 3303) and the collector's OTLP/HTTP endpoint (`LGTM_OTLP_PORT`, 4318). The start page carries the lab's password; its port line names `127.0.0.1` like the others, and a `LAB_PAGE_BIND` set to any address that is not loopback is refused by name before a container starts: `refused: LAB_PAGE_BIND=0.0.0.0 is not a loopback address`.

You need GNU make and Docker. Docker Compose v2 is the `docker compose` command. Windows is not a target. On macOS the `make` from Homebrew is GNU make; the make that ships with the system is not.

Every version, port and name is a setting in `lab/versions.env`. The environment overrides the file. The list is in [CONFIGURATION.md](CONFIGURATION.md). The image tags were chosen as pins and were not pulled when this page was written. The first `make up` is the check that each tag exists. A missing tag is changed by setting the variable, not by editing the Makefile.

## Start

1. From a checkout of this repository, run `make up`. On a terminal it asks how you want to see Proven ([The start-up flow](#the-start-up-flow)); with no terminal and no setting it starts the sample lab and asks nothing.
2. Before the first Docker call, `make up` says what it is about to do, with a rough time for each step, and after each step it prints one status line with the time the step took. A status line has an empty line before and after it, so it does not drown in Docker's output. Green on success; on a failure the same line in red, naming the step that failed and the log to read. Colour is for a terminal only: nothing is coloured when the output is piped or in CI, and `NO_COLOR` is respected.

   ```text
   What will happen next:
     1. Build Proven's image (about 1 min; the first run downloads more)
     2. Start Gitea, Grafana and the lab's Docker for the CI runner
     3. Create the demo repository and 7 pull requests, and start the CI runner
     4. Proven reviews each pull request (a few minutes)
     5. Show you where to start

   ==> [1/5] Proven's image is built (58 s)

   ==> [2/5] Gitea, Grafana and the lab's Docker are up (32 s)
   ```

   Step 4 waits for Proven to review every demo pull request, with a line for each one, for at most `DEMO_WAIT_SECONDS` (20 minutes) in all, so the block below appears when the dashboards already have data. A review that has not arrived by then is named in red, with the log to read, and the lab is shown anyway: it is up, and `make demo-check` waits again. `PROVEN_LAB_WAIT=no` skips the wait. `make up` exits with the status of that wait.
3. Every run of `make up`, the first and each one after it, ends with the same block, and prints nothing after its last empty line:

   ```text
   ########################################
   ####    Your proven lab is ready    ####
   ########################################

   To see demo pull requests scanned by proven please open:

   Gitea: http://127.0.0.1:3300/user/login
   Username: proven
   Password: <password>

   Grafana dashboard: http://127.0.0.1:3301/d/proven-overview
   Wait for every review: make demo-check
   Help: https://github.com/saphan-ai/proven/blob/main/docs/WALKTHROUGH.md

   >>>>>>>>  START HERE:  http://127.0.0.1:3302/lab  <<<<<<<<

   ```

   The START HERE line is bold and green on a terminal and plain anywhere else. On a terminal the Password line carries the password, read from the lab's state file. When the output is not a terminal (CI, a redirect, a pipe) the line names the file in the lab's volume that holds it instead, and the password is printed nowhere. The step-by-step walkthrough is `docs/WALKTHROUGH.md`; this page is for troubleshooting. Outside the repository's folder, run `make -C <path to the repository> demo-check`.
4. Open the start page, `http://127.0.0.1:3302/lab`. It holds the login with copy buttons, and links to Gitea, the dashboard and the documentation. Or open Gitea's address and sign in with that user and password.
5. Open the `discount` repository. It has the demo pull requests, listed under [The demo pull requests](#the-demo-pull-requests).
6. Each pull request runs the workflow in `lab/workflow/proven.yml`. When the job finishes, the pull request has a comment whose body starts with `<!-- proven-review run=<run number> sha=<head commit> -->`. That comment is the sample review of that pull request, and the marker names the run that posted it. Reviews take a few minutes each, and the runner works on up to three at a time, so a pull request without a comment is still waiting.
7. The dashboards are at `http://127.0.0.1:3301`, the Overview first (`/d/proven-overview`). See [The dashboards](#the-dashboards).

`make status` prints the containers, Gitea's health, whether the runner is registered, which pull requests have their review comment, and the login block.

`make demo-check` is the check that the demo is whole, and the wait that `make up` runs itself. It needs the running lab. It prints one line per pull request, `[3/7] #3 demo-discount: defect, as the manifest says`, and then how many review traces reached the dashboards. It requires every pull request of the list to have Proven's comment, a clean one to say clean and a defective one to carry a finding, and the trace store to hold at least one review trace per pull request.

### The start page and the documentation

The start page is one static page. A small web server (`LAB_PAGE_IMAGE`, a build of nginx that runs as a user that is not root, with a read-only filesystem) serves it at `LAB_PAGE_PATH` on `LAB_PAGE_PORT`, published on `127.0.0.1` only (see above). The seed renders it from `lab/page/lab.html.tmpl` into the lab's page volume each time `make up` runs. The login is filled in there and nowhere else: the template holds a placeholder, the password is not in any file of this repository, and it is not in an image. The page is also staged only inside that volume: the renderer writes one file there with mode 0600 and moves it into place, and removes it if it stops, so no copy of the page with the password in it exists outside the volume. The page also carries the three ways to [go live](#go-live). The table of pull requests is written from the pull requests the seed really made. Every link on the page opens in a new tab. The page needs no network: it has no web font and no outside script. Each pull request in its table has a **Report** link: [the report pages](#the-report-pages).

The documentation site, the same `docs-site/Dockerfile` that builds Proven's documentation, is a lab service on `LAB_DOCS_PORT` (3303 by default), also published on `127.0.0.1` only. `make up` builds it. A build that fails is named and the lab goes on without it; the page's documentation link then does not open.

### The demo pull requests

The seed makes the repository's `main` from the base of the discount sample set, then one branch and one pull request for each line of `lab/seed/prs.tsv`. The branch is named `demo-<set>`, and the lab's workflow reviews it with that sample set (see [What the workflow runs](#what-the-workflow-runs)). Each change is the change tree of its sample set in `testdata/fixture`, so nothing is invented: the review a pull request gets is the review the sample set records.

| Pull request | Sample set | What Proven says |
|---|---|---|
| Round the discount down to a whole cent | `discount-clean` | Clean. The change is the correct version of the change below; nothing is reported. |
| Reword the draft note | `unknown-author` | Clean. Only a string changed; nothing is reported. |
| Apply a discount (it is taken off twice) | `discount` | A test that fails on the change and passes before it: 100 at 20 percent becomes 60, not 80. The same change adds a dependency that is on the known-exploited list. |
| Add a padding helper from evil.example/drop | `malicious-package` | The new dependency has a malicious-package advisory. |
| Bump the widget library | `osv-go-module` | The dependency version is inside an OSV affected range, with the advisory row as evidence. |
| Add a copyleft helper | `spdx-licence` | The new dependency has a licence the project's policy does not allow. |
| Add a very young dependency | `project-health` | The new dependency has no maintainers, by its project-health score. |

A clean pull request is reviewed like the others. Its comment has one more line, `clean: Proven kept no finding on this change.`, after the line that says how the review was made, and the lines of the review below it say who reviewed (the default reviewer and the second one) and `failed services: none`, which means every check ran. When a review keeps no finding but a service failed (the model did not answer, for example), that line reads `incomplete: Proven kept no finding, but a service failed (listed below), so this change was not fully reviewed.` instead: a check that did not run found nothing, and that is not a clean result. Its run is on the dashboard beside the others. The first two rows are a clean change and the botched version of the same change, so the lab shows what a score compares: a defect caught, and a clean change left alone. The check `lab/lab_test.go` keeps the list honest: every row names an existing sample set, the image holds it, and no two defective rows show the same class.

The sample generator answers for every reviewer, so the reviewers do not disagree here: that panel shows a zero, which is the true count. Sample mode reviews the sample set the branch names, not the diff itself; any other branch is reviewed with the discount sample set. A live review, with a model behind each call, is the one that reads a diff.

`make logs` follows every service. `make logs SERVICE=gitea` follows one of `gitea`, `dind`, `runner`, `lgtm`.

The seed always runs as a user that is not root, and that user owns the state volume, mode 700, with the password and tokens at mode 600, and the start page's volume, mode 755, which the page's web server reads. A state volume made by an older lab is owned by root and the seed cannot write to it: run `make clean`, then `make up`.

`make down` stops the lab and removes its containers. The volume, the admin user and the password remain. `make clean` removes the volumes, the local images, `./bin` and `./dist`.

## The dashboards

A review is observable while it runs ([ADR-0045](adr/ADR-0045-a-review-is-observable-through-opentelemetry.md)). Proven sends an OpenTelemetry trace and metrics. The lab runs `grafana/otel-lgtm` as one more container, `lgtm`: an OpenTelemetry collector, Prometheus, Tempo, Loki and Grafana in one image, pinned by `LGTM_VERSION`. It is not part of Proven's code or image. `make up` starts it and waits for it. `make down` removes the container with the others, and the dashboards with it.

The lab's pull-request review is told where to send its data, with the standard variables and nothing of Proven's own: the seeded workflow starts Proven on the lab network with `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME` and `OTEL_METRIC_EXPORT_INTERVAL` set from `lab/versions.env`, and `OTEL_RESOURCE_ATTRIBUTES` naming the review: its instance id is `<unix time>-pr<number>-<fixture>`, so a dashboard knows a review by its pull request and change, and a clean change from a defect. The review runs with a second reviewer model (`LAB_SPECIAL_REVIEWER`), so there are two reviewers to show.

Six dashboards, one style: dark, a title that says what you see, at most eight panels, a word beside every colour (green is good, red is bad, blue is a count with no verdict, and a number reads `0 · none, good` or `3 · blocked`, never a bare colour), no empty box (a bar chart with no data says why in one row, and every bar is named), and a short description in each panel (the small `i` in its corner). A table that shows only the latest rows says so in its title, and is tall enough for them. Durations are in the unit their source returns. The default time range is the last hour on all six, and the bar charts step in minutes, because the lab reviews and the receipts all arrive within a quarter of an hour. Each has a row of links to the others at the top. Sign-in is not asked: the lab turns Grafana's anonymous access on, as an administrator, because the lab binds to `127.0.0.1`.

| Dashboard | Address | What you see |
|---|---|---|
| Every review at a glance | `/d/proven-overview` | Reviews finished, how many of the broken changes Proven flagged, false alarms on the clean ones, failed agents; a table with one row per review; findings by level of proof; reviews over time; reviews that had a failed service |
| One review, step by step | `/d/proven-review` | Pick a review at the top (it is named `<time>-pr<number>-<change>`, newest first). Findings kept, claims dropped, failed model calls; each agent in order with how long it took; both reviewers' Skeptic verdicts side by side; where they disagree |
| The evidence behind each finding | `/d/proven-evidence` | Findings by level of proof, claims dropped for no evidence, how long the Investigator and the Prover ran. Zetaris queries and Prover runs are **not measured yet**: two text panels name what is missing |
| Which models answered, and what the guardrail checked | `/d/proven-models` | Model calls, calls with an error, unsafe guardrail checks; calls by agent and by the model asked and the model that answered; guardrail checks; what Proven refused to keep; model latency |
| The four Meterless engines | `/d/proven-engines` | H-MEM's suppressed claims; what each review ran as (class, two reviewers, cut short). Scout Intent, Markovian and the World Model are **not measured yet**: text panels name what is missing |
| The model gateway, from Nightshift's receipts | `/d/proven-nightshift` | Receipts loaded, calls that answered and failed, receipts that failed verification; calls over time by model and app; why calls failed; median latency; tokens ([the receipts](#the-nightshift-receipts)) |

Three things on them that a reader should know:

- **Model latency in sample mode is not latency.** The answers are recorded, so the panel's title says so. A live lab shows the real time of each call.
- **"Model calls that returned an error"** counts a model call that returned an error. In sample mode that includes a call for which the recorded answers hold no answer (`generator: no recorded response`): the review goes on without it, so the count is not a defect of the review. In a live lab it is a failed call to the model.
- **What is missing is named, not drawn.** The metrics Proven does not emit yet are listed in the text panels of the Evidence and Engines dashboards: Zetaris queries per table, rows, duration and refusals; the Prover's runs on the base and on the change, with the duration and whether the network was off; Scout's plans (light or full) and guard overrides; Markovian's steps; the World Model's claims and evidence. When Proven emits one, the panel that names it becomes a query.

The dashboards are JSON files in `lab/grafana/dashboards`, written by `python3 lab/grafana/build_dashboards.py` and provisioned on start. Nothing in them is drawn by hand, and Grafana does not let the page change them. To change one, change the generator, run it, and commit both; a test regenerates the files and fails when they differ. Another test holds every query to what Proven emits (`internal/telemetry`) and to what the Nightshift importer writes, so a renamed metric or a misspelt one turns it red. How a panel looks needs Grafana: it is checked by the screenshots of a live `make up`, not by a test.

`make trace` asks the trace store for the latest review's trace and prints its span names and the reviewers they name. It exits 1 when the trace is missing a span or a reviewer.

### Watch a review

1. Run `make up`. It ends when every review is in, and prints the Overview's address.
2. Open the Overview. It refreshes every 10 seconds. The table has one row per review, a `Verdict` of `clean` (green) or `blocked` (red), and the findings kept.
3. Open **One review** and pick a review. The agent table lists each agent in the order it ran, with a bar for how long it took.
4. Start a review again: close and reopen a pull request in Gitea. The workflow runs on the runner (`make logs SERVICE=runner`). The new review appears in the Overview's table as it finishes.

### The report pages

Each demo pull request has a **Report** page: the review Proven writes with `proven scan --md`, set in the way GitHub shows a Markdown comment. A verdict badge on top (blocked, clean), the findings table, and each finding folded under its table row with its evidence, in light or dark to match the reader's system. The page has no script and no remote address, and carries no password, no token and no name of anything private (a test holds that). It is served by the start page's nginx at `/lab/report/<pull request number>.html`.

`make up` builds the pages after the seed (`lab/scripts/reports.sh`; `make reports` builds them again): for each pull request it runs Proven's image with no network and no key, on that change's sample set, and the seed image turns the Markdown into the page (`lab/seed/report.sh`, `lab/seed/md2html.awk`). The seed writes a placeholder for every pull request first, so a Report link is never dead: the placeholder says plainly that no report has been built and names `make reports`. The directory and its pages are world-readable whatever umask the seed runs under, because the start page's nginx is another user (a 700 directory once made every Report link a 403; a test holds it). In a live lab the page says it is the sample review of the same change and that the live review is the comment in Gitea: the page is not a second live review, which would spend the lab's keys again.

The stylesheet is written for the page, in GitHub's light and dark colours. No third-party stylesheet is copied in. The Markdown the CLI writes is documented in [OUTPUT_FORMATS.md](OUTPUT_FORMATS.md).

### The Nightshift receipts

The Nightshift dashboard shows Nightshift's signed receipts, one per call Proven made through it. They are loaded into the lab's Loki from `lab/nightshift/receipts.public.jsonl` by `make up` and by `make nightshift-receipts` (`lab/seed/nightshift-import.sh`). What the lab ships is the **public** set: the receipts of Proven's own calls, with no prompt text. The full receipts file stays with whoever runs Nightshift.

One JSON object per line, flat: `ts_ms`, `app`, `status` (`ok` or `failed`), `model`, `provider`, `reason` (a failed call's reason: `provider_deadline`, `empty_response`, `http_401`, ...), `verified` (whether its signature verified), `latency_ms`, `input_tokens`, `output_tokens`, `request_id`, `receipt_id`. The importer loads only lines of app `proven`, drops `prompt_preview` from every line and refuses a line that still names it, and refuses a line that is not an object with a numeric `ts_ms` and a status. It replays the receipts: the newest is stamped with the moment of the import and the rest keep their spacing, because Loki refuses samples older than a week; each line keeps its own `ts_ms`. A set already loaded in the last hours is not loaded twice (`FORCE=1` loads it again). With no file, the dashboard says no receipts are loaded.

### Use it outside the lab

The same data goes to any OpenTelemetry backend. Set `OTEL_EXPORTER_OTLP_ENDPOINT` to its OTLP/HTTP address, and Proven exports; leave it unset and Proven exports nothing and opens no connection. For a review run on the host against the lab's collector: `OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4318 proven scan --sample <path>`. The settings are listed in [CONFIGURATION.md](CONFIGURATION.md#telemetry-opentelemetry-adr-0045). The collector's Prometheus is inside the container and is not published; to scrape Proven's numbers from an existing monitoring stack, point `OTEL_EXPORTER_OTLP_ENDPOINT` at that stack's own collector.

## The start-up flow

`make up` is an interactive flow when there is a terminal to ask on, and never blocks when there is not. It runs before any container starts, so a refusal in it has started nothing. The code is `lab/scripts/flow.sh`.

**What it does, in order**

1. It reads `lab/.env` and the environment, and refuses a bad value of any setting below by name.
2. It checks that every port the lab publishes is free, and names each busy one with the program that holds it (see [Ports](#ports)).
3. On a terminal it prints a banner, then asks what no setting has already answered:
   - `Do you want to see Proven in SAMPLE mode (no keys, nothing leaves this machine) or LIVE mode (real models and live data)? [SAMPLE/live]`
   - for LIVE: `Are you a judge of the Open Agent Hackathon? [y/N]`
   - right after the mode (sample or live), `Do you want to tweak Meterless features? [y/N]`; on `y`, one question per feature that has a switch (see [The Meterless features](#the-meterless-features)). Enter is `N`. In every question the capitalised choice is the default and Enter takes it
   - for a judge: `For Proven to prove anything live, it needs Zetaris and Meterless Nightshift. Connect your own instances, or use the ones Proven was built with, in Hua Hin, Thailand, under Marcin's desk.` and `Which do you want? [OURS/own]`
4. It starts the lab and ends with the ready block and the *Start here* page, as every `make up` does.

A question reads from the terminal only: standard input when it is a terminal, otherwise `/dev/tty`. An answer that is not one of the offered ones asks again, at most three times, and then stops by name, with the setting that answers it in advance. `y` and `yes`, `n` and `no` are the same answer. Input that ends is a stop, never a guess.

**Every answer can be given in advance.** Put the setting in the environment or in `lab/.env` (an environment value wins; an empty value is the same as not set). A question whose setting is set is not asked.

| setting | values | answers |
|---|---|---|
| `PROVEN_LAB_MODE` | `sample`, `live` | sample or live |
| `PROVEN_LAB_JUDGE` | `yes`, `no` | are you a judge (live only) |
| `PROVEN_LAB_BACKEND` | `ours`, `own` | our services through the judges' gateway, or your own setup (a judge; a non-judge is always `own`) |
| `PROVEN_LAB_GATEWAY_AUTH` | `access`, `token` | path `ours`: the Cloudflare Access e-mail sign-in (default), or the token we sent you |
| `PROVEN_LAB_OWN_SETUP` | `yes`, `no` | path `own`: run the automated setup (`make up ZETARIS=local`, then `make nightshift`) |
| `PROVEN_LAB_USE_DETECTED` | `yes`, `no` | path `own`: use the Zetaris and Nightshift that already answer on this machine |
| `PROVEN_MARKOVIAN` | `on`, `off` | Markovian, bounded steps for a large change: Proven's own setting, passed through as it is |
| `PROVEN_LAB_SCOUT_INTENT` | `on`, `off` | Scout Intent, a plan for each review: the lab's spelling of Proven's `scout_intent` key |

A value outside the list is refused by name before any container starts: `up: PROVEN_LAB_MODE=maybe is not one of: sample, live. NEXT — set PROVEN_LAB_MODE to one of those, or unset it. Nothing was started.`

**With no terminal and no setting, it is a sample lab and nothing is asked.** That is the default for CI, a pipe, a redirect and `make e2e`. The log says `up: sample mode`. Keys that happen to be in `lab/.env` are not used then, and the log names them: `up: sample mode does not use these settings: …`. A key is used only in live mode, so a scripted live start sets `PROVEN_LAB_MODE=live` (and `PROVEN_LAB_JUDGE`, and for a judge `PROVEN_LAB_BACKEND`). With `PROVEN_LAB_JUDGE=yes` and no `PROVEN_LAB_BACKEND` and no terminal, it refuses by name rather than choose for you. `PROVEN_LAB_MODE=live` with no `PROVEN_LAB_JUDGE` and no terminal means not a judge.

**Sample.** The lab starts as it always did: recorded answers, no key, no call. Any live setting is left out of the run and any secret stored earlier is removed, so the lab is a sample lab whatever `lab/.env` holds.

**Live, a judge, `ours`: our services through the judges' gateway** (path B, [ADR-0051](adr/README.md) for the gateway). The gateway address is `JUDGES_GATEWAY_URL` (in `lab/.env`, or asked for on a terminal; it must be `https://`). Two ways in, by `PROVEN_LAB_GATEWAY_AUTH`:

- `access` (default). The judge signs in with Cloudflare Access: an e-mail one-time PIN, in their browser, from an allowlist of judges' addresses; nothing secret is sent to the judge, and Access logs who entered. The flow needs `cloudflared`. If it is missing the flow says how to install it (`brew install cloudflared` on macOS; the download page for Linux and Windows) and stops before any container. It runs `cloudflared access login <address>`, discards what that prints, takes the application token with `cloudflared access token -app=<address>`, and says when the token expires (read from the token) and how to renew it: run `make up` again, which signs you in again and stores the new token.
- `token`. The per-judge token we sent privately (`JUDGES_GATEWAY_TOKEN`). On a terminal it is read with no echo; the terminal is restored on every way out, Ctrl-C included. With no terminal it must be in `lab/.env`. A token holds letters, digits and `. _ ~ + / = -` only.

Both: the credential is checked **once** against the gateway, with a request that authenticates and spends nothing, an empty JSON object sent to the evidence route (`JUDGES_GATEWAY_ZETARIS_PATH`). The gateway answers `400 not_a_template_request` to credentials it accepts, and 401 or 403 to others. The flow says `the gateway accepted your credentials: ready.` or names the refusal (`the gateway refused your credentials: HTTP 401 (token_expired)`) and stops before any container. Then the credential is stored as an Actions secret (`JUDGES_GATEWAY_TOKEN`, or `JUDGES_GATEWAY_ACCESS_TOKEN` for Access) by the seed, as every key is. It reaches the seed in the lab's state volume (see [How a credential travels](#how-a-credential-travels)), so it is never printed, never in a file on the host, never on a command line and never in Compose or any container's environment. The review job sends the Access token to the gateway as the `cf-access-token` header **only when Proven's generator and Zetaris adapters can send that header**; as of this round they send the token as a Bearer value, and the lab passes the Access token as that value and by name as `JUDGES_GATEWAY_ACCESS_TOKEN`. Whether the gateway accepts it that way is the gateway's Access-side work (W18) and has not been run against it.

**Live, `own`, and every non-judge who chooses live** (paths A and C). The flow prints where the guide is (this file, the *Go live* section, and the docs site on `127.0.0.1` once the lab is up), then:

1. It asks the two services a judge's machine may already run, one request each: Zetaris at `DETECT_ZETARIS_URL` (`http://localhost:8888/api/v1.0`) and Nightshift at `DETECT_NIGHTSHIFT_URL` (`http://127.0.0.1:8088/v1/models`). A service counts as present only when its answer has its own shape: Nightshift answers `200` with a JSON list of models (`{"data":[…]}`) or `401` with a JSON object; Zetaris answers `200`, `400`, `401` or `403` with a JSON object. A stray web server on the port (a `404`, an HTML page, a plain-text `401`) is not taken for either. These shapes are read from Proven's own adapters and are **not measured against a running Zetaris or Nightshift**: if a real one answers in another shape it is not found, and you set its address in `lab/.env` by hand. If one answers, it offers to use them as they are (`PROVEN_LAB_USE_DETECTED`). The lab's review job runs inside the lab's own Docker, so it names this machine by `LAB_JOB_HOST_ALIAS` (Docker's name for the host it runs on, set in `lab/versions.env`). A service that listens on `127.0.0.1` only cannot be reached from there; the flow says so. Your keys stay yours and come from `lab/.env`.
2. It points at what a judge's machine is set up with: Nightshift is the native app from meterless.ai (Windows, Mac M1+, Mac M5+, DGX); Zetaris is the bundle from your own account. Our Tart VM for Nightshift is an optional automation, not the first step.
3. It offers the automated setup (`PROVEN_LAB_OWN_SETUP`): `make up ZETARIS=local` for Zetaris from your bundle, then `make nightshift` for the Tart VM on a Mac with Apple silicon (on any other host it says so and goes on).
4. It prints, verbatim: `We cannot vouch for results on a setup we do not control: this is best effort.`

If no key or token is set yet, the flow says the first reviews will be sample reviews, and the start page says so.

**The mode is on the start page** (`/lab`): `Mode: sample`, `Mode: live via our gateway` or `Mode: live on your own setup`. Under it is the line `Meterless features, as this lab passed them to the review:` (see below). The page never carries a token; the only secret on it is the existing login.

### The Meterless features

Proven carries four Meterless engines: H-MEM (memory of refuted findings), the World Model (claims, evidence and verdicts as a graph), Scout Intent (plans each review; its benefit is not measured yet) and Markovian (reviews a large change in bounded steps). All four are **on by default** (the owner's ruling of 2026-10-26). The lab passes `on` for Scout Intent and Markovian unless a setting or an answer turns one off, so the default holds whether or not Proven's own default has changed yet; an answer becomes Proven's own setting. Markovian is passed in sample mode too.

| feature | asked by `make up` | Proven's setting | the lab's setting | how the review is given it |
|---|---|---|---|---|
| H-MEM | no: Proven has no switch for it | none | none | always on |
| World Model | no: Proven has no switch for it | none | none | always on |
| Scout Intent | `Scout Intent: plans each review before the agents run (on by default; its benefit is not measured yet). Use it? [Y/n]` | `scout_intent: on` or `off` in `proven.yaml`; there is no environment spelling | `PROVEN_LAB_SCOUT_INTENT` | the review job writes `scout_intent: <value>` to a file and points the review at it with `PROVEN_CONFIG` |
| Markovian | `Markovian: reviews a large change in bounded steps (on by default). Use it? [Y/n]` | `PROVEN_MARKOVIAN`, `on` or `off` (also `PROVEN_MARKOVIAN_THRESHOLD`, `…_STEP_UNIT`, `…_CARRY_TOKENS`, which the lab does not ask about) | `PROVEN_MARKOVIAN` | passed to the review by name; Proven reads it in live reviews only |

How the questions work. After sample or live, on a terminal, `make up` asks `Do you want to tweak Meterless features? [y/N]`. On `y` it prints that H-MEM and the World Model are on by default and have no switch to ask about, then asks the two questions above, each only if its setting is not already given. `y` is `on`, `n` is `off`, and Enter takes the default, the capitalised choice of the prompt (`on` for a feature). With no terminal nothing is asked and both features are on. A setting given in advance (in the environment or `lab/.env`) is never asked, so a script or CI that gives both is never asked at all. A value that is not `on` or `off` is refused by name before any container starts.

The seed stores each given value as an Actions secret, as it stores a key, but a feature is not a key: with only a feature set the lab is still a sample lab, and sample mode keeps the feature settings. The workflow passes them to the review (`lab/workflow/proven.yml`).

**The start page shows what the lab passed**, not what was typed: the line is built by the seed from the values it stored, for example `H-MEM on (always), World Model on (always), Scout Intent off, Markovian on`. H-MEM and World Model read `on (always)` for as long as Proven has no switch for them.

What the lab did not measure: whether a `scout_intent: on` file or `PROVEN_MARKOVIAN=on` changes the comment of a *sample* review (Markovian is a live-only setting by Proven's code; a sample review is recorded). The offline tests show the review process is given the file and the variable; the owner's test below shows the page.

**The names in the organisers' Participant Setup Guide** are read as they are. A judge's `.env` from the guide holds `METERLESS_BASE_URL`, `METERLESS_RUNTIME_TOKEN`, `METERLESS_MODEL`, `ZETARIS_REST_URL`, `ZETARIS_ORG_ID` and `ZETARIS_API_KEY`. The last four are the lab's names already. `METERLESS_BASE_URL` is read as `PROVEN_GENERATOR_BASE_URL` and `METERLESS_MODEL` as `PROVEN_GENERATOR_MODEL`, when the Proven name is not set (it wins when both are). This mapping is in the lab's own loader (`lab/scripts/lib.sh`); Proven's own settings loader is another track's work.

### Ports

The guide runs Zetaris on `localhost:3000` (web UI), `8888` (REST) and `10000` (JDBC), and Nightshift's runtime on `127.0.0.1:8088`. The lab therefore does not use `3000`. Its defaults are Gitea `3300` (`GITEA_HTTP_PORT`), Grafana `3301` (`LGTM_GRAFANA_PORT`), this start page `3302` (`LAB_PAGE_PORT`), the documentation site `3303` (`LAB_DOCS_PORT`) and, with `ZETARIS=local`, the Zetaris console `3304` (`ZETARIS_UI_HOST_PORT`). The collector's OTLP port `4318` (`LGTM_OTLP_PORT`) is a standard port and is checked too. Every one is a setting in `lab/versions.env`; the address stays `127.0.0.1`.

Before any container starts, `make up` checks each port. A busy one is refused by name, with the program that holds it where the host lets you read that (`ss`, then `lsof`), and `a program` where it does not:

```text
up: port 3300 (GITEA_HTTP_PORT) is already in use by docker-proxy.
up: the lab publishes on 127.0.0.1 and cannot start on a port that is taken. NEXT — stop that program, or give the lab another port: put NAME=port in lab/.env (the names are in the lines above; the defaults are in lab/versions.env). Nothing was started.
```

A port that this lab's own containers already publish is not a clash, so a second `make up` on a running lab goes on.

## Go live

Out of the box the lab reviews in sample mode: recorded answers, no key, no call. Three ways make the next review live ([ADR-0052](adr/ADR-0052-the-lab-goes-live-three-ways.md)). The owner's step by step test of all three, with the output to expect, is [LAB_GO_LIVE_TEST.md](LAB_GO_LIVE_TEST.md).

How it works, the same for all three:

- `make up` reads `lab/.env` when it exists. Keys in it are used in **live** mode only: on a terminal the flow asks, and a script sets `PROVEN_LAB_MODE=live` (see [The start-up flow](#the-start-up-flow)). Copy `lab/.env.example` to `lab/.env` and fill in what you have; the file is not committed and holds no value until you write one. A line is `NAME=value`, with no quotes.
- Each **non-empty** setting named in `LAB_LIVE_KEYS` (`lab/versions.env`) is stored as an Actions secret of the demo repository. A credential travels in the lab's state volume and a plain setting by name through Compose ([How a credential travels](#how-a-credential-travels)). Nothing is baked into an image or printed: the log names the settings it stored and never their values. An **empty** setting is not stored, and a secret you emptied is removed, so taking a key out of `lab/.env` and running `make up` again returns the lab to sample mode.
- The workflow runs a **live** review when a generator key, a Nightshift token or the gateway pair is stored, and a **sample** review otherwise. The first visible line of Proven's comment says which: `sample: no keys`, or `live: <generator>, Zetaris on|off`, for example `live: NVIDIA via Nightshift, Zetaris on`. The line above it is the `<!-- proven-review … -->` marker, an HTML comment that the browser does not show; `make e2e` reads it as the first line of the body.
- A review that already ran is not run again when a key changes. Open a new pull request, or push a commit to a demo branch.
- A live review of a demo pull request reads the sample set's change tree that is baked into Proven's image, the same files the pull request carries. It does not clone the pull request.

### How a credential travels

A credential is a value that opens something. `LAB_SECRET_KEYS` in `lab/versions.env` lists them: `PROVEN_GENERATOR_KEY`, `PROVEN_GUARDRAIL_KEY`, `METERLESS_RUNTIME_TOKEN`, `ZETARIS_API_KEY`, `JUDGES_GATEWAY_TOKEN` and `JUDGES_GATEWAY_ACCESS_TOKEN`. The other names of `LAB_LIVE_KEYS` (modes, addresses, a model name, an organisation id, the two feature switches) open nothing and are passed to the seed by name through Compose, as before.

1. `make up` reads the credential from `lab/.env` or from your shell (the Access token comes from `cloudflared` into the shell and goes nowhere else).
2. Just before the seed runs, `make up` writes every non-empty credential into the lab's state volume as one file, `/state/secrets`, over **standard input** with `umask 077` (mode 0600). `printf` is a shell builtin, so no process lists a value, and the value is not an argument of any command.
3. The seed reads the file, stores each credential as an Actions secret over a pipe, and removes the file. It removes it on every way out: a normal finish, a failure, a TERM or an INT. `make up` removes it again after the seed whatever the seed's status was, so a seed that was killed leaves none.
4. The Zetaris administrator key of path C (`ZETARIS_ADMIN_API_KEY`) goes to the register step on its standard input. No container has it in its environment.

So `docker compose config` never shows a credential (lab/compose.yaml names none), and `docker inspect` of the seed, of every other container, and of the exited seed runs shows none, because none is in any container's environment. `make livecheck` reads the keys from `lab/.env` on the host, as before, never from a container.

What this does not hide: the value is in your shell and in `lab/.env` on your machine, in the process memory of `make up`, in the state volume for the length of the seed run, and in Gitea's secret store afterwards. The Actions job receives it as a secret of the review run, as it always did.

The checks: `go test ./lab/` holds the offline controls (a credential name put back into the Compose environment goes red; a value put on a command line goes red; a loosened umask goes red; a seed that stops early removes the file). On a Docker host, `make check-lab-secrets` runs the live control after a `make up` with a sentinel value in each name: no sentinel in `docker compose config`, in `docker inspect` of any container of the project, or in a container's process arguments; each name is an Actions secret; `/state/secrets` is gone.

### Path A: one NVIDIA key

A free key is enough on its own.

1. Get a key at <https://build.nvidia.com>: sign in, open any model, choose *Get API Key*. It is shown once.
2. `cp lab/.env.example lab/.env`
3. Put the key after `PROVEN_GENERATOR_KEY=` in `lab/.env`. Leave the mode lines empty: with only a key set, the lab runs in `nvidia` mode and uses the same key for the guardrail.
4. Set `PROVEN_LAB_MODE=live` and `PROVEN_LAB_JUDGE=no` in `lab/.env` (or answer the questions on a terminal: live, then no). A key is used in live mode only. Then `make up`. The log says `seed: stored as Actions secrets: PROVEN_GENERATOR_KEY`.
5. Open a new pull request in the demo repository. Its comment starts with `live: NVIDIA direct, Zetaris off`.

### Path B: our services through the judges' gateway

The gateway holds our Nightshift token and our Zetaris key and injects them; you get an address and a token of your own, sent to you privately. The gateway is another wagon's work (W18, [ADR-0051](adr/README.md)); its order was not on this branch, so the lab names the settings `JUDGES_GATEWAY_URL` and `JUDGES_GATEWAY_TOKEN`, and where it serves Zetaris (`JUDGES_GATEWAY_ZETARIS_PATH`, default `/zetaris/api/v1.0`) and the organisation id it expects (`JUDGES_GATEWAY_ZETARIS_ORG_ID`, default `0`) are settings too. If the gateway uses other names, only those settings change.

1. `cp lab/.env.example lab/.env`
2. Set `JUDGES_GATEWAY_URL=` to the address you were sent and `JUDGES_GATEWAY_TOKEN=` to your token.
3. Set `PROVEN_LAB_MODE=live`, `PROVEN_LAB_JUDGE=yes`, `PROVEN_LAB_BACKEND=ours` and `PROVEN_LAB_GATEWAY_AUTH=token` in `lab/.env` (or answer on a terminal; the default way in is the Cloudflare Access sign-in, which needs no token, see [The start-up flow](#the-start-up-flow)). Then `make up`. The flow checks the token once and says `ready` or names the refusal. The log says `seed: stored as Actions secrets: JUDGES_GATEWAY_URL, JUDGES_GATEWAY_TOKEN`.
4. Open a new pull request. Its comment starts with `live: NVIDIA via the judges' gateway, Zetaris on`.

The guardrail (ADR-0005) is on by default and its default is your own model server. If neither the gateway nor you provide one, a review names the guardrail as the failed service. Set `PROVEN_GUARDRAIL_KEY` to an NVIDIA key and the lab uses NVIDIA's hosted guardrail.

### Path C: Zetaris and Nightshift on your own machine

Two parts. Each is optional; do either, or both.

**Zetaris** is the sponsor's platform. Its images are in the sponsor's registry behind a sign-in, and the terms under which a judge may pull and run them are not in this repository (searched on 26 October 2026: `hackathon/fleet/infra/zetaris` holds a runbook and a console tool, no licence or terms file). So `make up ZETARIS=local` does not carry them: it runs **your** bundle with **your** registry sign-in, and says so.

1. Download `zetaris-platform.zip` from your own Zetaris account and unpack it. Never commit the zip or the driver inside it.
2. Sign in to the sponsor's registry with the command that came with your account, in a shell that is not logged. Wait for `Login Succeeded`.
3. In the unpacked folder: `bash ./preflight.sh`. Every line must say PASS.
4. `cp lab/.env.example lab/.env`, then set `ZETARIS_COMPOSE_FILE=` to the full path of that folder's `docker-compose.yml`.
5. With `PROVEN_LAB_MODE=live` and `PROVEN_LAB_JUDGE=no` set (or answered on a terminal; the flow also offers this step itself, `PROVEN_LAB_OWN_SETUP`), `make up ZETARIS=local`. The log shows `zetaris: the platform answers on http://127.0.0.1:8888`. The lab runs the bundle as its own compose project with `lab/zetaris/override.yml` on top: the console and the REST service are published on `127.0.0.1` only, and the REST service also joins the lab's network as `zetaris-api`, so the lab's reviews reach it with no address to write. The evidence is **not** registered yet: the log says which settings are missing.
6. Open the console at `http://localhost:3304`. Open it as `localhost`, not as `127.0.0.1`: the console needs a secure context. This step cannot be made through an API. Sign in as the platform's default administrator (the bundle's README says how) and make an API key for the administrator. Then make a user for Proven with no role, sign in as it, and make its API key. Note the organisation's **numeric** id (the sign-in answer returns it as `orgId`; the 32-character id in the page's URL is a different value and is refused).
7. In `lab/.env` set `ZETARIS_ADMIN_API_KEY` (the administrator's key; used once, not stored as a secret of the repository), `ZETARIS_PROVEN_USER` (the user's name), `ZETARIS_API_KEY` (that user's key) and `ZETARIS_ORG_ID`.
8. `make zetaris`. It registers the definitions in `lab/zetaris/evidence.sql` as the administrator and grants them to Proven's user. Then `make livecheck ZETARIS=local` names Zetaris ready. Then `make up` again, so the keys become secrets.

What is registered: the CISA known-exploited feed as one REST table and its view (`px_kev`, `px_kev_vulns`). No row is stored: a definition holds a URL. The other sources of [ADR-0023](adr/ADR-0023-how-proven-s-evidence-is-loaded-through-zetaris.md) (EPSS, OSV by id) are registered by the administrator's steps in `hackathon/fleet/infra/zetaris/zetaris-local.md`, section C. The Zetaris images run as the sponsor built them; the lab's override does not set a user for them, because it cannot know which users they allow.

**Nightshift** runs in a macOS virtual machine, so its unsigned installer never touches your Mac. It needs a Mac with Apple silicon and about 25 GB of disk for the base image. On a Linux host `make nightshift` says that Tart is macOS-only, points to paths A and B, and exits 0.

9. `make nightshift`. If Tart is not installed it names the command to run: `/bin/bash hackathon/fleet/infra/nightshift-vm/nightshift_vm.sh install-tools` downloads a pinned Tart release, checks its checksum and signature, and asks for your password once.
10. `make nightshift` again. It checks the host, clones the base image (about 25 GB; minutes) when the VM named `NIGHTSHIFT_VM_NAME` does not exist, starts it with no shared folder and no shared clipboard, and waits for its address. The VM's window opens on your Mac.
11. Download Nightshift's macOS installer from Meterless yourself, then `/bin/bash hackathon/fleet/infra/nightshift-vm/nightshift_vm.sh copy-installer <the .dmg>`. It checks the file against Meterless's published checksums, copies it into the VM and checks it again there.
12. Inside the VM's window: open the image, drag Nightshift to Applications, start it. If macOS refuses it ("Apple could not verify…"), open System Settings → Privacy & Security, choose **Open Anyway**, confirm with the VM's password. Do this only in the VM.
13. Pass Nightshift's email access gate in the VM.
14. In Nightshift's Providers tab paste the NVIDIA key's value (`nvapi-…`, about 70 characters), never its identifier (a UUID). Nightshift's "verified" does not prove the key: the first real request does.
15. Copy Nightshift's runtime token from its own window in the VM. This script never reads or prints it. Run `make nightshift` once more: it prints `Nightshift answers at http://<the VM's address>:8088`. Put that address after `PROVEN_GENERATOR_BASE_URL=` and the token after `METERLESS_RUNTIME_TOKEN=` in `lab/.env`.
16. With `PROVEN_LAB_MODE=live` and `PROVEN_LAB_JUDGE=no` in `lab/.env` (the flow asks on a terminal), `make up`, then open a new pull request. Its comment starts with `live: NVIDIA via Nightshift, Zetaris on` (or `Zetaris off` when Zetaris is not set up).

Steps 9 to 14 are done once. What the script cannot do, and why: the installer, the email gate, the provider key and the token are in an app window with no API.

### When it does not work

| You see | Meaning | Do |
|---|---|---|
| `lab/.env: refused line 7 (a setting is NAME=value)` | line 7 of `lab/.env` is not `NAME=value` | fix that line; no quotes |
| `refused: LAB_PAGE_BIND=… is not a loopback address` | the start page carries the password and is published on `127.0.0.1` only, whatever the setting says | unset `LAB_PAGE_BIND` |
| the comment says `sample: no keys` after you set a key | the review ran before the key was stored, or the key line is empty | `make up` again, then open a **new** pull request |
| `zetaris: pulling the platform's images failed` | you are not signed in to the sponsor's registry | step 2, then `make up ZETARIS=local` |
| `zetaris: ZETARIS_COMPOSE_FILE is '', which is not a file` | the setting is empty or wrong | step 4 |
| `register: the platform answered HTTP 4xx` | the administrator's key, the organisation id or the user name is wrong | steps 6 and 7 |
| `nightshift: Nightshift is not answering` | the app is not installed or started in the VM | steps 11 to 14 |
| a live comment lists `nightshift`, `zetaris` or `guardrail` as a failed service | that service refused or did not answer; the review went on without it | check its key and address; `make livecheck` names which |

## The keys a live review needs

This page says which keys the lab needs for each way of running the model generator, and where to get each one
([ADR-0031](adr/ADR-0031-anyone-with-an-nvidia-key-can-run-a-live-review.md) point 6). It covers keys only. Every setting
and its default is in [CONFIGURATION.md](CONFIGURATION.md). Sample mode needs no key and makes no call:
[SAMPLE_MODE.md](SAMPLE_MODE.md).

### With one NVIDIA key (no Nightshift needed)

A free NVIDIA API key is enough to see Proven's agents work on live models.

1. Go to <https://build.nvidia.com> and sign in.
2. Open any model and choose *Get API Key*. Copy the key; it is shown once.
3. Put the key in the environment, never in a file that is committed:

   ```
   PROVEN_GENERATOR_MODE=nvidia
   PROVEN_GENERATOR_KEY=<the key>
   PROVEN_GUARDRAIL_MODE=nvidia
   PROVEN_GUARDRAIL_KEY=<the same key, or another>
   ```

4. Check it before a review with `livecheck`. Its generator line names the mode, and a missing or rejected key.

The address and the models have defaults, so nothing else is needed. In this mode the address is `PROVEN_NVIDIA_BASE_URL`
(`https://` only; any other value is refused by name). `PROVEN_GENERATOR_BASE_URL` is not read here, so a Nightshift
address left in the environment never receives the key. `livecheck` names the host the answer came from. The hosted API signs no receipt, so the trace says
*no receipt* for each call, and the requirement that an agent be answered by an `nvidia/*` model is checked against the
`model` field of the answer. A free key is rate limited; a review that meets the limit names it as a failed service.

### With Meterless Nightshift (the default)

Set `METERLESS_RUNTIME_TOKEN` to the runtime token of the Nightshift you run, and `PROVEN_GENERATOR_BASE_URL` to its
address (this variable is read in `nightshift` mode only). `PROVEN_GENERATOR_MODE` is left empty or set to `nightshift`. Nightshift is the team's own service: a judge
without one uses the NVIDIA key above.

### Keys of the other services

| Service | Key | Where it comes from | If it is empty |
|---|---|---|---|
| Guardrail | `PROVEN_GUARDRAIL_KEY` | the same NVIDIA key page, when `PROVEN_GUARDRAIL_MODE=nvidia` | every model call is refused and the review names the guardrail |
| Zetaris evidence | see [CONFIGURATION.md](CONFIGURATION.md) | the Zetaris account | Zetaris is down and the review names it; the rest runs |

### A second reviewer: one more key

The lab runs with NVIDIA alone ([ADR-0037](adr/ADR-0037-bring-your-own-second-reviewer.md) point 6). Every change
gets one independent review, from the default reviewer, and nothing else is needed.

Proven can review every change with two models. The second one is any model behind an OpenAI-compatible chat API:
OpenAI, Anthropic through its compatible endpoint, or a server on your own machine. You switch it on with three
settings in the environment Proven runs in, and the key is the switch:

| Variable | Value |
|---|---|
| `PROVEN_SECOND_REVIEWER_KEY` | the API key. Secret: never printed, never written to the trace or the result |
| `PROVEN_SECOND_REVIEWER_BASE_URL` | the API address, for example `https://api.openai.com/v1` or `http://localhost:11434/v1` |
| `PROVEN_SECOND_REVIEWER_MODEL` | the model name the API knows, for example `gpt-5` |

Add the three variables and run the lab as before. Take the key away and the lab is back to one reviewer. Nothing
else changes.

#### What you see once it is connected

- **Both reviewers look at every change.** They work independently: neither is shown the other's answer. A change
  under review cannot narrow this for itself: `proven.yaml` is read from the baseline.
- **Where they differ, the review says so.** Each finding carries both reviewers' verdicts and reasons, and the
  review counts the disagreements; the two are kept side by side and never averaged. The block *The reviewers
  disagree*, which shows them in the text, Markdown, HTML and pull-request comment output, and the verdicts in the
  JSON and SARIF output, arrive with the shell train (see [ARCHITECTURE.md](ARCHITECTURE.md)); the command line does
  not render them yet. Evidence still decides what is reported: a risk neither reviewer can prove is not a finding,
  however many models name it.
- **Every review names its reviewers**, and whether either of them is the model that wrote the change.

#### What it costs

Two reviewers cost two model calls per review step. To spend less, name the classes the second reviewer covers in
`proven.yaml`, for example `doubled_classes: security,backend`. The security class is always doubled when a second
reviewer is connected. See [CONFIGURATION.md](CONFIGURATION.md) for the settings and their refusals: a key with no
address or no model, or an address that is not an `http` or `https` URL, stops Proven before any review starts and
names the variable.

#### If the second reviewer is down

A review whose second reviewer fails still completes with the first reviewer's findings. The failure is listed
under `second-reviewer` in the review's failed services, and the findings show the second reviewer's verdict as
`no answer`, which is not a disagreement.

## What the workflow runs

Proven's Action is not on main yet, so the workflow does not use it. It runs:

```text
SAMPLE_MODE=true proven scan /opt/proven/fixture/<set>
```

With a key stored (see [Go live](#go-live)) the same command runs with `SAMPLE_MODE=false` and the settings that are stored, passed to the container by name. `<set>` is the name after `demo-` in the pull request's branch, and `discount` for any other branch. Each set is a recorded sample in `testdata/fixture`, baked into Proven's image (the Dockerfile copies the ones the demo uses). It does not review the diff of the demo pull request. The comment's first visible line says `sample: no keys` or `live: …`. Sample mode uses no keys, no Docker and no model: it makes no call to a service. In the lab the job also sends the review's telemetry to the lab's own collector (see [The dashboards](#the-dashboards)), which is why the seeded copy starts Proven on the inner daemon's network; the committed workflow does not. The comment is the text of that review (the report pages carry its Markdown). The lab, like the Action, posts that one comment and no line comments (ADR-0020 §7).

**One review per head commit.** A push to a pull request branch must give one review run and one comment. Two runs were reported on one head commit in a live lab (runs 18 and 19, one sha). This round could not repeat it (no Docker on the seat that wrote this) and does not claim to know which event fired twice. The workflow has: only the `pull_request` event as its trigger, with its types named (`opened`, `synchronize`, `reopened`); a `concurrency` group of pull request number and head sha, so runs of one head queue (Gitea 1.22.6 may ignore the key: **not verified**); and a check, before the review and again before posting, for a comment of Proven that already names this head sha. A run that finds one reviews nothing and posts nothing, and says `skipped: head commit … already has a Proven review`. That last guard is the one that does not depend on Gitea; its remaining gap is two runs that both look before either posts, which the second look narrows and the queue closes where it works. `make e2e` accepts the review of any run of the head that succeeded, because the second run posts nothing.

### One comment per pull request

The workflow posts ONE comment on a pull request and edits it in place on every new run, so `run=1`, `run=2` and so on no longer pile up. The comment's first line is still `<!-- proven-review run=N sha=<head commit> -->`, now with the latest run and head: `make e2e` reads that line and passes only when a comment by `proven-bot` starts with it for a run of the current head that succeeded (see **One review per head commit** above).

How the step finds its comment: it asks the server which account the token is (`GET /user`), lists the pull request's comments page by page (`GET /repos/{repo}/issues/{pr}/comments`), takes the first comment written by that account whose body starts with `<!-- proven-review`, and edits it (`PATCH /repos/{repo}/issues/comments/{id}`). It creates a comment (`POST` on the pull request's comments) only when it finds none. A comment with the marker by another account is never edited. The step carries a copy of `action/post-comment.sh` because a workflow in a user's repository cannot read files of this repository; `lab/scripts/post-comment-check.sh` fails when the two copies differ, and it checks the choices above offline (run by `make check-e2e-parse`).

Two things this asks of the environment, measured as far as this repository can measure them:

- The token needs `read:user` to answer `GET /user`. The seed now mints `lab-workflow` with `write:issue`, `write:repository` and `read:user`. A lab seeded before this change keeps the old token: run `make up` again (the seed mints a fresh token on every run), or the step stops and says to set `PROVEN_COMMENT_ACCOUNT` to the account's name. On GitHub, where `GITHUB_TOKEN` may not read `/user`, the account is `github-actions[bot]`.
- Editing needs an HTTP client that can send `PATCH`. BusyBox `wget` (the `docker:cli` job image) cannot. The step uses `curl` or GNU `wget` when the image has one, and otherwise installs `curl` with `apk add` inside the job's own container. Not measured: this step on the `docker:cli` image (no Docker on the machine that wrote it); the routes are the ones Gitea's and GitHub's API documents serve, and the offline check uses their JSON shapes.

The same file is what a GitHub user commits. On GitHub the job's Docker runs `saphan/proven:latest` from Docker Hub, and the job posts with `PROVEN_TOKEN` or with `GITHUB_TOKEN`. The image has to have been pushed first. `make push` does that, and only after Docker Scout has passed on the image. This repository does not push it for you.

### Which account posts the review

The workflow posts as its own account, `proven-bot` (`LAB_BOT_USER`). It is not an admin. The lab admin is the account you sign in with, so a review the workflow wrote and a comment you or any other person wrote must not look alike. If the admin posted the review, an admin who types the marker by hand would pass for the workflow.

- The seed creates `proven-bot`, makes it a collaborator with write access on the demo repository, and mints that account's token. The token is stored only as the repository's Actions secret `PROVEN_TOKEN`, which the workflow reads. The admin's token is not given to the workflow.
- The account's generated password stays in the lab's state volume until `make clean`. Neither it nor the token is printed, written to a tracked file or baked into an image layer. The admin's password reaches you as [Start](#start) says: on a terminal in the closing block; when stdout is not a terminal never printed, and written by the first `make up` to `lab/.state/admin-password` at mode 0600. It is never in a log, a tracked file or an image layer.
- `make e2e` accepts a review comment only from `proven-bot`, with the marker for the run and the head commit. The same marker from the admin or any other account is ignored, and e2e names the account it ignored.
- If the secret is missing, the job falls back to `GITHUB_TOKEN`; on Gitea that posts as another account, and e2e then fails closed.

The seeded copy replaces `saphan/proven` with `PROVEN_IMAGE` when that setting is not the default, and replaces `runs-on: ubuntu-latest` with `RUNNER_LABEL`.

## What `make` does besides the lab

`make` with no target prints every target, the version from `git describe`, whether `go` and `docker` were found, and the lab URL when the lab is up.

`make build` writes `./bin/proven` and `./bin/livecheck`. It uses Go when `go` is on `PATH` and the version is at least the `go` line in `go.mod`. Otherwise it builds the same binaries in `GOLANG_IMAGE` and writes them to the host. If neither Go nor Docker is available, it prints how to install Go on macOS and on Linux and exits 1.

`make build-all` cross-compiles for darwin and linux, amd64 and arm64, with `CGO_ENABLED=0`, into `./dist/`, and writes `./dist/SHA256SUMS`.

`make sample` runs the sample review from a temporary directory outside this repository. It does not pass keys and it does not call Docker.

`make livecheck` runs the live-service check **for the setup the workflow will use**. If `lab/.env` exists, it loads it, then runs the very lines of `lab/workflow/proven.yml` between `derive-begin` and `derive-end`, which decide the mode, the guardrail's mode and key and whether Zetaris is on. It prints `livecheck: the workflow will run: <the first line of the comment>`, then checks what that setup uses. So with only `PROVEN_GENERATOR_KEY` in `lab/.env` (path A) it checks NVIDIA direct and the guardrail with the same key, as the workflow does, and says `remark: the query line is not needed in this setup` for Zetaris, which is off there; the exit is `0` when nothing else is not ready. With a key, token or gateway and Zetaris set, the query line counts. With no key at all the workflow runs a sample review, so it prints `runs a SAMPLE review` and exits 1 (there is no live service to find ready). The container line is still checked in every live setup. With `ZETARIS=local` (`make livecheck ZETARIS=local`) and no `ZETARIS_REST_URL` of your own, it asks the loopback port the local platform is published on.

`make reports` builds the report page of every demo pull request again ([The report pages](#the-report-pages)). `make nightshift-receipts` loads the public Nightshift receipts into the lab's Loki ([The Nightshift receipts](#the-nightshift-receipts)).

`make nightshift` starts Nightshift in a Tart VM on a Mac with Apple silicon and prints its address (path C, steps 9 to 15). On any other host it says Tart is macOS-only and exits 0. `make zetaris` starts Zetaris from your bundle and registers the evidence (path C, steps 4 to 8); `make up ZETARIS=local` does the same as part of the lab. `make check-golive` runs the offline checks of all of this: the keys of `lab/.env` become secrets and an empty one does not (against a stub of Gitea's API), the workflow's live or sample choice and the first lines of its comment, the loopback bind of the start page, and the Zetaris registration against a stub. It needs no Docker and no key.

`make image` builds Proven's image for this machine and loads it. The load is one platform, this machine's own. A local engine cannot load a manifest list, so this target does not build both platforms. `make seed-image`, which `make up` uses, is the same kind of load: one platform, because the lab has to start the image. The image is the static `proven` and `livecheck` binaries, `git`, and the `docker` client, on `RUNTIME_IMAGE`, and it runs as a user that is not root. Local tags: `latest` and the `git describe` value. OCI labels `created`, `revision` and `version` come from the date, `git rev-parse` and `git describe`.

`make push` writes nothing to a registry until every platform of both images has been scanned, and then publishes what was scanned, and only that. Each tag is one manifest list of `PUSH_PLATFORMS` (`linux/amd64` and `linux/arm64`). Proven's tags are `latest` and the `git describe` value, on `PROVEN_REGISTRY/PROVEN_IMAGE`. The seed's only tag is `SEED_IMAGE_TAG`, on `PROVEN_REGISTRY/SEED_IMAGE`. The build asks for provenance (`--provenance=mode=max`) and an SBOM (`--sbom=true`). The sequence:

1. It checks that the `docker scout` plugin and `skopeo` are present. If either is missing, nothing is built or pushed. `skopeo` copies the scanned layouts to the registry.
2. It builds each image once, for every platform in `PUSH_PLATFORMS`, into a local OCI layout in a temporary directory (`docker buildx build --output type=oci`). No build pushes, and no tag, staging or final, is written to a registry.
3. It runs Docker Scout on the layout of each image, once per platform (`docker scout cves oci-dir://<layout> --platform <platform>`, the same severities as `make scan`, below), so the image scanned is the image that will be published, on every platform.
4. Only when every scan of every platform of both images is green, `skopeo copy --all --preserve-digests` copies each scanned layout, unchanged, to its final tags. It then reads each tag back and refuses if it carries any other digest than the layout's. Nothing is rebuilt after a scan.

It exits non-zero and names the image and the platform when the plugin or `skopeo` is missing, when a scan cannot run, when Scout reports a critical or high finding that has a fix, or when a platform is missing from a build. In any of those cases nothing has been written to any registry: no tag of any kind exists for that build, and `latest` and the version tag keep pointing where they did.

`make scan` runs Docker Scout on the images this repository builds, Proven (`saphan/proven`, tag `latest`) and the seed (`saphan/proven-lab-seed`, tag `2`, unless `SEED_IMAGE` or `SEED_IMAGE_TAG` was overridden). It fails closed: it exits non-zero and names the image when the `docker scout` plugin is missing, when a scan cannot run, or when Scout reports a critical or high finding that has a fix. The severities are `SCOUT_REFUSE_SEVERITIES`. When the images pass, it also scans the pinned third-party images (`gitea/gitea`, `gitea/act_runner`, `docker`) and writes that output between the markers at the bottom of this page.

`make check-lab-gates` checks both gates offline, with a stub `docker` and a stub `skopeo` on `PATH`: `make up` refuses the retired `LAB_BIND` before any container starts, and `make scan` refuses without a green scan. For `make push` it checks that, with Scout missing, nothing is built and nothing is written; that a failed scan of either image, or of only one platform, causes zero registry writes (the stub records every push, copy and tag, and the list must be empty); and that on success every platform of both images was scanned from a local layout before the first registry write, no build pushes, the digests copied are the digests of the layouts that were scanned, and a tag that reads back another digest is an error. Six copies of `push.sh`, each with one line changed, must fail those checks. The loopback property is checked on `lab/compose.yaml` alone, by the Go test in `lab/labcheck` and by `make check-review-gates`, each with copies that put `0.0.0.0` or a variable back in a port line. The stub cannot show what a real registry, Docker Scout or `skopeo` does; the first live `make push` is that check.

## The runner and Docker

The lab has seven services: Gitea, a Docker daemon that lives only inside the lab (`dind`), the Gitea Actions runner, the dashboard stack (`lgtm`), the start page (`page`), the documentation site (`docs`), and a one-shot seed. The runner takes up to three jobs at a time, so the demo's pull requests are reviewed side by side.

Gitea does not serve SSH (`GITEA_DISABLE_SSH`, and `GITEA_START_SSH_SERVER` is false). The image already runs OpenSSH on port 22, and Gitea's own SSH server on that same port exited the container. The lab does not need SSH. The web UI, the dashboard's two ports, the start page and the documentation site are the only published ports. `GITEA_SSH_PORT` is kept in the settings and is not published.

Every `gitea` command `make up` runs inside the container uses `GITEA_RUN_USER`. That account is `git` in this image. The same command as root is refused, and `compose exec` is root when no user is named. `make down` keeps the volume, so a second `make up` finds the password already stored. It still repairs a missing admin user, a missing runner registration and a missing pull request, then starts the runner. A branch that already exists gets the current workflow, which starts that pull request's review again; a branch that is already current is left alone.

The runner's Docker host is `tcp://dind:2375`. No service mounts the host's Docker socket. Job containers are started on that inner daemon. The inner daemon is privileged, because Docker-in-Docker needs it. That privilege stays inside the lab. The daemon's port is not published on the host.

`make up` also pulls the Prover's test image (`PROVEN_TEST_IMAGE`, `golang:1.26` by default) into that inner daemon, by digest when `TEST_IMAGE_DIGEST` is set. Without it the lab's `make livecheck` reports `No such image: golang:1.26`. The digest is pinned in `lab/versions.env` as `TEST_IMAGE_DIGEST`: it is the digest of the `golang:1.26` index (both platforms), read from Docker Hub on 2026-10-23. `make up` pulls `golang@<digest>` and tags it `golang:1.26`, so a judge gets the bytes we measured, not whatever the tag names later. To refresh it, run `docker buildx imagetools inspect golang:1.26`, copy the `Digest:` line of the index into `TEST_IMAGE_DIGEST`, and run `make up`. If you empty the setting, `make up` pulls the tag instead and prints the digest it got.

Gitea starts a workflow for a pull request only when the repository's Actions unit is on. The seed turns the unit on by name, and the Compose file adds `repo.actions` to Gitea's default units. If the unit was off when the pull request opened, the seed closes and reopens the pull request so the workflow starts. The seeded copy of the workflow posts to Gitea at `http://gitea:3000/api/v1` instead of `github.api_url`, because the job runs on the inner daemon, where `127.0.0.1` is not Gitea.

`make check-links` checks that every relative Markdown link and every path in `metadata.json` resolves.

The job image is `docker:<DOCKER_VERSION>-cli`, pinned. The runner does not pull an image it already has, so the Proven image loaded by `make up` is the one the workflow runs.

Sample mode does not start a Prover container. A later live run on this runner would start those containers on the lab's daemon, not on the host's.

## Rehearse

`make e2e` starts the lab when Gitea is not running and waits up to `E2E_WAIT_SECONDS`. Gitea 1.22.6 has no Actions runs, tasks, jobs or workflows route (HTTP 404), so e2e reads the statuses of the pull request's current head commit (`GET /repos/{owner}/{repo}/commits/{sha}/statuses`). It keeps the statuses of the workflow's job (`proven / review (pull_request)`, read from `lab/workflow/proven.yml`) and takes the highest run number in their `target_url`, because the seed may start more than one run on one head. It exits 0 only when that run's newest status is `success` and a comment by the workflow's own account (`LAB_BOT_USER`, `proven-bot` by default; see "Which account posts the review") starts with `<!-- proven-review run=N sha=<head commit> -->`, where N is that run's number and the commit is the pull request's current head. The workflow fills both from `github.run_number` and `github.event.pull_request.head.sha`. No time is compared, because two events in one second cannot be told apart by time. A run still `pending`, or no status yet, is waited for. `make check-e2e-parse` checks the parsers offline against comments and statuses shaped as Gitea serves them; it covers the lab admin posting the exact marker after a successful run that posted nothing, an unrelated account's marker, a successful run that posted nothing, an older run's review from the same second, and a review of another head commit. A review comment of an older run, of another commit, or by another account does not count and e2e says which it ignored: a lab volume kept by `make down` keeps its comments. If a kept volume still holds an older copy of the workflow on the pull request's branch, the run posts the old marker and e2e fails closed; run `make clean && make up`.

It exits 1 when that run finished as `failure` or `error`, and names the run and its status:

```text
e2e: workflow run #2 for pull request #1 (head commit <sha>) finished with status failure
```

It also exits 1 when the wait ends with no run, a run that is still going, or a successful run with no new review.

If Gitea is already running and the runner container is not, `make e2e` does not start the runner. It exits 1 and prints:

```text
e2e: runner is not running, so the workflow cannot post a review
```

That is the check that a stopped runner cannot look like a successful rehearsal. From the repository root, stop the runner and run the rehearsal again:

```bash
docker compose --env-file lab/versions.env -f lab/compose.yaml --project-directory lab stop runner
make e2e
```

A second check proves e2e cannot be fooled by an old comment. With the runner alive and an old review comment on the pull request, make the current run fail and require `make e2e` to refuse:

```bash
make e2e
make e2e-failed-run
```

`make e2e-failed-run` commits a workflow that fails on purpose to the pull request's head branch, runs `make e2e`, and puts the original workflow back. It exits 0 when e2e exited non-zero and named the failed run, and 1 when e2e accepted it. It needs the first `make e2e` to have passed, so an old comment exists. Restoring the workflow starts one more run; wait for it to finish before you run `make e2e` again.

### Test steps for the lab's look (the owner's)

Run each with `make up` on a terminal unless it says otherwise. Each step names what you see.

1. **The plan.** Run `make up`, answer `sample`, then Enter at the next question. You see `[SAMPLE/live]` in the first question and Enter takes it. Before any Docker output you see `What will happen next:` and five numbered steps, `7 pull requests` in the third.
2. **The step lines.** Between Docker's output you see `==> [1/5] Proven's image is built (NN s)`, then `[2/5]` to `[5/5]` (`The lab is ready`), each in green with an empty line above and below it.
3. **The wait.** Step 4 prints a line for each pull request, `demo-check: [1/7] ... waiting for Proven's review`, then `defect, as the manifest says` or `clean, as the manifest says`. The ready block appears after the last one.
4. **The ready block.** It ends on `>>>>>>>>  START HERE:  http://127.0.0.1:3302/lab  <<<<<<<<`, bold and green, with one empty line after it and nothing else.
5. **No colour.** `NO_COLOR=1 make up` and `make up | cat` print the same words with no colour.
6. **A failure.** Stop Docker, run `make up`: before any step line you see Docker's own refusal; stop it during step 2 instead and you see `==> [2/5] FAILED: Start Gitea, Grafana and the lab's Docker (after NN s). Read: make logs SERVICE=gitea, or SERVICE=lgtm, or SERVICE=dind` in red.
7. **The start page.** Open START HERE. The table has a **Report** column. Open the report of `Apply a discount (it is taken off twice)`: a red **Blocked** badge, the findings table, and the findings folded; open one. Switch your system to light and dark: the page follows. Open the report of a clean change: a green **Clean** badge.
8. **The dashboards.** Open the Overview. You see four numbers, each with a word: reviews finished, `5 of 5 flagged` (green), `0 false alarms` (green), `0 agents failed` (green). The table has seven rows, none hidden (it is sized for twelve). Use the links at the top to open the other five. Evidence and Meterless engines name what is missing in grey panels. Nightshift says no receipts are loaded until `lab/nightshift/receipts.public.jsonl` is there.
9. **One review.** Open **One review**, pick the newest. Each agent is a row with a bar for its duration; both reviewers' verdicts are side by side.
10. **Offline checks.** `go vet ./...`, `go test ./lab/...`, and `go test ./internal/telemetry/`.

## Third-party images

These pins are not built here. `make scan` replaces the block below with Docker Scout's output. A finding in a third-party image does not refuse the scan: we cannot patch it, and a judge should see it.

<!-- scout-third-party:start -->
Not scanned in the change that added the lab. The seat that wrote this page could not run Docker Scout. `make scan` fills this block.

Images: `gitea/gitea:1.22.6`, `gitea/act_runner:0.2.11`, `grafana/otel-lgtm:0.11.0`, `docker:29.9.1-dind`, `docker:29.9.1-cli`.
<!-- scout-third-party:end -->
