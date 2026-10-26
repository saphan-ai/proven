# ORDER — the-prover-in-a-container · round 1

- change: the-prover-in-a-container
- round: 1
- base: main @ 0ed349b
- executor: cursor on runner-cursor
- written: 2026-10-22T02:33:12Z
- task: the Prover's live test-run adapter runs the change's tests only in a container with no network and no secret, and runs no code when no container runtime is present (ADR-0025)

## The order as written

The master's order, verbatim: the owner's words, what to build, acceptance and scope are in it.

## ORDER — the-prover-in-a-container
track: proven-reviewer-core
- train: proven-reviewer-core
- wagon: W1b
- review: named
- repo: proven
- base: main @ 0ed349b
- branch: fix/the-prover-in-a-container
- lands-on: main
- worktree: █████████████████████████████████████████████████████████████
- backend: cursor
- model: grok-4.7-high
- quote-usd: 3.00
- cap-usd: 6.00
- wall: 120m
- task: the Prover's live test-run adapter runs the change's tests only in a container with no network and no secret, and runs no code when no container runtime is present (ADR-0025)

█████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████
ROLE-LOCK: YOU = EXECUTOR. One-shot headless: foreground commands only; your first act writes `_EXEC_STATUS.md`
(`- state: in-progress`). A permission denial is a fact: try once, then `- state: blocked`.
COMMIT EVERY PART BEFORE `stop-2-ready` · ZERO push/merge/PR · author = fleet identity · commit message
`[the-prover-in-a-container] <area>: <what>`.
No AI attribution anywhere: no `Co-Authored-By`, no "Generated with", no tool or model name, no session link.
After `git commit`, `packed-refs.lock: Permission denied` means the commit LANDED — check `git log -1`.

### SKILLS
- ██████████████████████████████████
- ██████████████████████████████████████

Craft: `██████████████████████████` + `████████████████████████████` — read first.

### THE OWNER'S WORDS
2026-10-21, on ADR-0025: *"OK."* 2026-10-21, adding this wagon: *"tak, dodać W1b."* (yes, add W1b).

### THE DECISION YOU BUILD — ADR-0025 (pasted, accepted by the owner)
1. The change's code, and its tests, run only inside a container that Proven starts for that one run, never in Proven's
   own process or on the host directly.
2. The container has no network. A test that needs one fails, and the review says the test could not run and why.
3. The container holds no secret. No environment variable, file or mount of Proven's own carries a token, a key or a
   credential into it.
4. The container sees only the checkout it needs, and nothing it writes leaves it except the test results Proven reads back.
5. Its image, time limit and resource limits are settings, with documented defaults, validated on load (ADR-0013). There
   is no setting that gives it a network or a secret.
6. Where no container runtime is available, the Prover does not run code. The finding then has a query as its evidence or
   is reported as unresolved, and the review says that the Prover was off and why. It never falls back to running the
   code outside a container.

### THE CONTRACT (in your base, main @ 0ed349b — implement it, do not change it)
`internal/testrun/container.go`: `ContainerConfig`, `DefaultContainerConfig`, `Validate`, `NewContainer`,
`(*Container).Argv`, `(*Container).Run`; `testrun.Request` gained `BaseDir` and `ChangeDir`. If the contract cannot carry
what ADR-0025 asks, write `SHOUT_the-contract.md` naming the field and why, and stop.

### WHAT TO BUILD
1. The bodies in `internal/testrun/container.go`: the runtime's argv for one run (no network, no environment passed in,
   one mount, of the one checkout and nothing else, the container removed after the run, CPU and memory limits); `Run` runs the command
   once against `BaseDir` and once against `ChangeDir`, each in its own container, with `Timeout` per run, and fills
   `Outcome`. A run past its time limit is never proof.
2. `NewContainer` with no runtime on the path returns `*service.Down{Service: "prover"}` with the reason; nothing runs.
3. Live mode: `review.Open` (with `Sample` false) uses `NewContainer`; when it fails, the review names the Prover as a
   failed service, as W1's `testrun.Live` did. Sample mode is unchanged. `testrun.Live` may go.
4. The defaults (image, time limit, CPU, memory) written in one short section of `AGENTS.md`.
5. Your own tests for item 3 and for anything item 1 needs beyond the test author's cases.

### ACCEPTANCE
1. `go build ./... && go vet ./... && go test -count=1 ./...`: green. Paste the summary lines.
2. A test proves that in live mode with no runtime the review completes, names the Prover off, and reports no finding as
   proved by a test.
3. `grep -rn 'os/exec' --include='*.go' .` lists only `internal/testrun/container.go` (and test files). Paste it.

### THIS ORDER CAN BE WRONG (law #44)
The master's claims, 2026-10-22, not built: (M1) Docker's `--network none`, `--rm`, `--cpus`, `--memory` and one `-v`
are enough for ADR-0025 points 1–4; (M2) a Go test can run in a stock Go image with no network when the module has no
external dependencies. If either fails, say so in `- master-was-wrong:` with the measurement, or `none`.

### WHAT IS NOT IN SCOPE
- No file ending in `_contract_test.go`, and no name starting with `TestContract` or `contract` — they belong to the test
  author, whose files you do not see.
- The shell, the generator, Nightshift, Zetaris, H-MEM.

### WHEN THIS ORDER CANNOT PROCEED
Write `SHOUT_<topic>.md` in the worktree, set `- state: blocked`, and stop.

### ARTIFACT CONFINEMENT
Every artifact this leg makes — captures, logs, evidence, scratch — lives under `███████████████████████████████████████`
in the worktree. Nothing outside the worktree; nothing committed unless this order names it.

### RETURN
`_EXEC_STATUS.md`: `- state: stop-2-ready`, `- last commit:` <sha>, an `- abstract:` of at most 5 lines, `- master-was-wrong:`,
the commands run and their exit codes, and the session recording or the reason there is none (ADR-0024). Check it with
`saphan doc check --family exec-status _EXEC_STATUS.md` before you end: its exit code is the verdict. Read every `saphan`
verb by its `RESULT:` line, its `--json` or its exit code, never by its sentences.

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| how an order is bound to the laws | 1 |
| private workspace path | 4 |
