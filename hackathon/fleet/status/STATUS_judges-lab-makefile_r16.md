# STATUS — judges-lab-makefile · round 16

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 5 min
- last commit: 02981c86f3c476c41ad9d20058a77e7f1a7b3381 [judges-lab-makefile] merge: absorb main c1e4ba8
- updated: 2026-10-24T10:41:48Z

## Summary

Round 9. The seed creates a non-admin bot account (proven-bot), makes it a write collaborator on the demo repo, and stores only its token as the ████████████ secret; the admin's token no longer reaches the workflow. e2e accepts a review only from that account, and names an admin or other author it ignored. New offline control: admin posts the exact marker after a successful run that posted nothing, refused. Live make up/e2e/e2e-failed-run NOT run: this seat has no docker or make.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — judges-lab-makefile
- state: stop-2-ready
- abstract: Round 9. The seed creates a non-admin bot account (proven-bot), makes it a write collaborator on the demo repo, and stores only its token as the ████████████ secret; the admin's token no longer reaches the workflow. e2e accepts a review only from that account, and names an admin or other author it ignored. New offline control: admin posts the exact marker after a successful run that posted nothing, refused. Live make up/e2e/e2e-failed-run NOT run: this seat has no docker or make.
- last commit: 02981c86f3c476c41ad9d20058a77e7f1a7b3381 [judges-lab-makefile] merge: absorb main c1e4ba8
- done: producer attribution fixed by a dedicated bot account; five round-8 controls kept and one new one added
  - master, 2026-10-24: absorbed main c1e4ba8 (02981c8): main's new LAB.md section (a second reviewer, ADR-0037) joins the runbook's keys section; go build, go vet, check-links green on the merged tree
  - master, 2026-10-24: make up failed at f162084 (seed: HTTP 404 setting the bot password). Measured on a live lab: Gitea 1.22.6 refuses proven-bot@localhost with 422, read as already-exists. Fixed in 9678e10 (email ████████████████████; the seed checks the account exists). At 9678e10: make up 0, e2e 0 (bot run #1), e2e-failed-run 0, runner stopped 2, down/clean/check-links 0
  - sh lab/██████████████████████████ prints "e2e-parse-check: ok" (the same script make check-e2e-parse runs), with the admin-exact-marker control in it
  - sh -n on seed.sh, e2e.sh, e2e-parse-check.sh: ok
  - sh scripts/check-links.sh: ok
  - go vet ./...: ok
  - docs/LAB.md section "Which account posts the review" and docs/CONFIGURATION.md rows LAB_BOT_USER, LAB_BOT_EMAIL
- next / blocked-on: the master runs make clean && make up && make e2e, make e2e-failed-run, runner stopped, make down on a Docker seat
- note: master-was-wrong: not measured. Gitea 1.22.6 giving a non-admin account a token usable as an Actions secret is UNVERIFIED here (no docker). The seed fails closed with a named HTTP code at each bot step. NOT RUN on this seat (no docker, no make on the executor's allowed programs): make up, make e2e, make e2e-failed-run, runner-stopped control, make down, live image pulls. Seed bot steps (admin/users POST+PATCH, collaborators PUT, users/bot/tokens DELETE+POST with scopes write:issue,write:repository) are written against the Gitea API from memory and unmeasured. go test ./... fails only in internal/pipeline TestMaterializedTreesAreReadableByOthers (umask 0002 on this seat, as the round-9 review also found); no Go file changed. Attribution: commit author is the fleet identity, no Co-authored-by trailer.
- updated: 2026-10-24
- artifacts:
  - lab/seed/seed.sh
  - lab/██████████████
  - lab/██████████████████████████
  - lab/compose.yaml
  - lab/versions.env
  - docs/LAB.md
  - docs/CONFIGURATION.md

CLOSE-OK · judges-lab-makefile · commit f162084e2302ea4b31b4fa4d8995ef979913c705 · branch fix/judges-lab-makefile · worktree ███████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 1 |
| file and line in Saphan Studio's private code | 3 |
| private credential or variable name | 2 |
