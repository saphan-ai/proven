# STATUS — judges-lab-makefile · round 13

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 6 min
- last commit: 278307bfb1ab1a77a33e585caf622b17adabe4db [judges-lab-makefile] lab: e2e passes only on the current run's review
- updated: 2026-10-23T20:40:55Z

## Summary

Round 6. `make e2e` now passes only when the workflow run of the PR's current head commit succeeded and a proven-review comment was created at or after that run started; a failed run exits 1 naming run and status, even with an old comment. New negative control `make e2e-failed-run`. docs/LAB.md digest paragraph fixed. Docker, make and network were not available on this seat: the Docker lines are unrun and the master runs them.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — judges-lab-makefile
- state: stop-2-ready
- abstract: Round 6. `make e2e` now passes only when the workflow run of the PR's current head commit succeeded and a proven-review comment was created at or after that run started; a failed run exits 1 naming run and status, even with an old comment. New negative control `make e2e-failed-run`. docs/LAB.md digest paragraph fixed. Docker, make and network were not available on this seat: the Docker lines are unrun and the master runs them.
- last commit: 278307bfb1ab1a77a33e585caf622b17adabe4db [judges-lab-makefile] lab: e2e passes only on the current run's review
- done: e2e checks the current run's own review, adds a failed-current-run negative control and fixes the digest paragraph; shell parsers and static checks run, Docker lines unrun
  - the master's defect was right: the old e2e.sh grepped every comment for `proven-review` (measured by reading lab/██████████████ at d8514fe; the review's predicate reproduction agrees); master-was-wrong: —
  - Actions API named: Gitea 1.22.6 `GET /repos/{owner}/{repo}/actions/tasks?limit=50` (one entry per run: head_sha, run_number, status, run_started_at, created_at). /actions/runs is a later release. UNVERIFIED on this seat: no network and no Gitea here, so the route and the field names come from my knowledge of Gitea 1.22, not from a measurement. If the route or a field differs, e2e reports "no workflow run has started" after E2E_WAIT_SECONDS and exits 1: it fails closed, never green.
  - the tasks route may need a token; the calls use the admin's basic auth (password read from the lab state volume by lab/████████████████████)
  - parsers run offline on synthetic Gitea-shaped JSON (███████████████████████████████████████████████): newest run picked by run_number for the head sha (failure picked over an older success), unknown sha gives nothing, a comment older than the run start is refused, a fresh one passes, the JSON-escaped marker (<) matches, a person's comment without the marker at its start is ignored, a non-UTC time is refused
  - `sh -n` exit 0 on lab/██████████████, gitea-api.sh, e2e-failed-run.sh; `sh scripts/check-links.sh` exit 0; `go vet ./...` and `go test -count=1 ./...` exit 0 (all packages ok)
  - `git rev-list --merges --count d8514fe^..HEAD` prints 0; `grep -ci co-authored` over the commit messages prints 0; author is Marcin Marzec <██████████████████████>
- next: the master runs the Docker lines below on a seat with Docker
- evidence:
  - Negative control, exact commands, from the repository root on a seat with Docker: `make clean && make up && make e2e` (exit 0, sets an old comment); `make e2e-failed-run` (exit 0 = e2e refused; it commits a workflow that fails on purpose to branch `change`, runs make e2e, which must print "workflow run #N for pull request #1 (head commit <sha>) finished with status failure" and exit non-zero, then restores the original workflow); wait for the restore run, then `make e2e` again exits 0
  - Existing control kept: `docker compose --env-file lab/versions.env -f lab/compose.yaml --project-directory lab stop runner && make e2e` exits non-zero naming the runner
  - UNRUN here: make (denied), docker (denied), `make clean`, `make up`, `make e2e`, `make e2e-failed-run`, runner-stopped control, `make down`; the whole Gitea behaviour (tasks route, that a pull_request run reads the head commit's workflow, that a push to `change` starts a run) rests on knowledge of Gitea 1.22, not measurement
  - Judgement call: provenance is by time (comment created_at >= run_started_at, UTC strings, second resolution), not by a run id in the comment, because a kept volume's `change` branch holds an older copy of the workflow that would not carry a new marker. Cost if wrong: a comment from a different run in the same second as the start would pass
  - Judgement call: the control replaces the whole workflow file on the head branch with a minimal failing one and restores the original bytes it read first
- note: laws echoed: operator-paste-law, command-delivery-law, commands-anchored-in-canonical-root, dogfood-gate-sequence, sacred-injection-hierarchy, red-escalation-law, fix-or-shout-law, pre-flight-budget-check, backend-routing-by-work-class, plan-default-opus-max, law-at-act, literal-executability, owner-next-block, cap-is-2x-estimate, fresh-checkout-law, full-surface-verify, commandment-01-archive-before-teardown, worktree-fork-from-current-HEAD, amendment-edits-the-text-it-amends, law-refresh-has-a-dod, verify-before-assert, executor-toolchain-denial, a-head-refuses-work-outside-its-track, a-known-next-act-is-performed-not-announced, a-denied-act-stays-denied-in-every-spelling, and the rest of 1 to 59.
- updated: 2026-10-23T20:45:00Z

CLOSE-OK · judges-lab-makefile · commit 278307bfb1ab1a77a33e585caf622b17adabe4db · branch fix/judges-lab-makefile · worktree ███████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| e-mail or user name | 1 |
| file and line in Saphan Studio's private code | 3 |
