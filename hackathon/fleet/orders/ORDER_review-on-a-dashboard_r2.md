# ORDER — review-on-a-dashboard · round 2

- change: review-on-a-dashboard
- round: 2
- base: main @ 9303b2d
- executor: claude-code
- written: 2026-10-24T09:54:15Z
- task: round 1 — Proven emits OpenTelemetry per ADR-0045, and the judges' lab shows a review live on a provisioned Grafana dashboard

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — review-on-a-dashboard
track: proven-cursor-infra
- train: proven-cursor-infra
- wagon: W13
- round: 1
- review: named
- repo: proven
- base: main @ 9303b2d
- branch: fix/review-on-a-dashboard
- lands-on: main
- backend: claude-code
- quote-usd: 11
- cap-usd: 22
- wall: 150m
- task: round 1 — Proven emits OpenTelemetry per ADR-0045, and the judges' lab shows a review live on a provisioned Grafana dashboard

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · commit subject `[review-on-a-dashboard] <area>: <what>`
matching `^\[[a-z0-9-]+\] [a-z0-9/._-]+: .+$`, one-token area, at most 72 characters, English.
No AI attribution anywhere. After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-23 (ADR-0045's trigger, translated): "a nice dashboard in a separate container… show how the two models go, one
already has its verdict and the other is still working… it has to be really nice… take a framework". 2026-10-24, the
owner's goal for today via the advisor: the judges' lab end to end with Gitea, and Grafana in the same lab per ADR-0045.

### WHAT THE MASTER MEASURED (2026-10-24 ~08:25Z; claims to test, law #44)
- Proven emits NO OpenTelemetry today: `git grep -niE 'otlp|opentelemetry' -- '*.go'` finds nothing on main or any branch.
- Your branch starts at the judges' lab tip 9678e10 (stream judges-lab-makefile, in review): Gitea 1.22.6 + DinD +
  act_runner + seed; `make up` / `make e2e` green on runner-cursor; the workflow posts as the bot account `proven-bot`.
  If that stream changes before it lands, the master carries your branch; you build on 9678e10.
- The decision is `docs/adr/ADR-0045-a-review-is-observable-through-opentelemetry.md` at your base: read it first, it is
  the specification. ADR-0013: no hardcoded settings — every version, port and address comes from `lab/versions.env`
  or the standard `OTEL_*` variables.

### WHAT TO BUILD
1. **Proven emits OTLP (ADR-0045 points 1–2).** One trace per review; one span per agent and per reviewer model (so the
   two reviewers show side by side, one finished while the other still works); span events for the trace file's events;
   metrics: findings by level of proof, claims dropped, reviewer verdicts and disagreements, model calls and their
   duration, guardrail checks and unsafe verdicts. Configured ONLY by the standard `OTEL_*` variables; OFF unless
   `OTEL_EXPORTER_OTLP_ENDPOINT` (or the signal-specific one) is set; never a secret in an attribute (the redactor
   applies). Use the official OpenTelemetry Go SDK (Apache-2.0). Follow the gen_ai semantic conventions for model calls.
   Tests: a review with an in-memory exporter produces the trace and spans named above; unset → nothing is exported
   and no network is opened (prove it with a control that fails when it is not so).
2. **The lab runs `grafana/otel-lgtm`** as one more compose service, pinned by version in `lab/versions.env`, on the lab
   network, its Grafana port published as `docs/LAB.md` states; `make up` starts it and prints its address; `make down`
   and `make clean` remove it. The workflow job gets `OTEL_EXPORTER_OTLP_ENDPOINT` (and service name) pointing at it, so a
   pull-request review in the lab is exported.
3. **A provisioned dashboard** (a JSON file, provisioned on start; nothing drawn by hand): a state timeline of the agents
   and the two reviewers, verdicts side by side, disagreements, findings by level of proof, dropped claims, model
   latency, guardrail checks.
4. **`docs/LAB.md`** gains a section: how to open the dashboard during a pull-request review, and what each panel shows.
   `docs/CONFIGURATION.md` lists every new setting.

### ACCEPTANCE (the owner reads these)
- `go vet ./...` and `go test ./...` exit 0; the new tests named in your status.
- On a seat with Docker: `make clean && make up` prints the dashboard address; `make e2e` still passes; after it, the
  dashboard's trace query returns the review's trace (show the query and its answer: span names, two reviewer spans).
  If your seat cannot run make or docker, say which lines you could not run — the master runs them.
- With `OTEL_*` unset, `proven scan --sample` exports nothing (the control above).
- `make check-links` exit 0; `git log main..HEAD` has no merge commit and no `Co-authored-by`.

### THIS ORDER CAN BE WRONG (law #44)
Write a `- master-was-wrong:` line: `—`, or what you measured that contradicts this order or ADR-0045 (for example, an
emission already present that the master's grep missed).

### WHAT IS NOT IN SCOPE
- Pushing images; the GitHub Action; changing what a review decides. Telemetry only observes.

### HEADLESS TURNS
A headless leg never ends its turn with prose: a message without a tool call ends the session. Write `_EXEC_STATUS.md`
before your last message; if you cannot finish, write it with `- state: blocked`. Quote `git rev-parse HEAD` output
verbatim; never restate HEAD from memory.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `███████████████████████████████████` in
the worktree. Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, the commands run
and their exit codes, the acceptance lines run and not run. Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict.

### ATTRIBUTION (owner, 2026-10-23)
No commit of yours may carry a `Co-authored-by:` trailer or a seat identity as its author. Check
`git log --format="%an <%ae>%n%B" main..HEAD` before you end and say so in your status.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
