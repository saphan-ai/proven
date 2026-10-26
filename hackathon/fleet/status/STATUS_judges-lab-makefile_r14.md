# STATUS — judges-lab-makefile · round 14

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 3 min
- last commit: dcd86fd [judges-lab-makefile] lab: e2e reads the head commit's statuses
- updated: 2026-10-24T02:31:34Z

## Summary

Round 7. e2e now reads commits/{head sha}/statuses (Gitea 1.22.6 has no Actions runs/tasks route), keeps the workflow job's context, takes the highest run number from target_url, and passes only on that run's newest status success plus a proven-review comment at or after the run's first status. failure/error names the run; pending or no status keeps waiting. Parsers are checked offline against the pasted JSON shape. Not run: make, docker (denied on this seat), so the live lab lines are for the master.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — judges-lab-makefile
- state: stop-2-ready
- abstract: Round 7. e2e now reads commits/{head sha}/statuses (Gitea 1.22.6 has no Actions runs/tasks route), keeps the workflow job's context, takes the highest run number from target_url, and passes only on that run's newest status success plus a proven-review comment at or after the run's first status. failure/error names the run; pending or no status keeps waiting. Parsers are checked offline against the pasted JSON shape. Not run: make, docker (denied on this seat), so the live lab lines are for the master.
- last commit: dcd86fd [judges-lab-makefile] lab: e2e reads the head commit's statuses
- done: e2e reads head-commit statuses and takes the highest run number
  - lab/████████████████████ holds newest_run, workflow_context, review_since; the context is read from lab/workflow/proven.yml, not typed twice
  - `sh lab/██████████████████████████` prints ok: runs 2 (success, id 8), 3 (pending, id 7), 1 (success) give run 3 pending; a failure after a pending gives run 3 failure with start at its pending; other context and an empty list give nothing; a comment older than the run start does not count
  - e2e-failed-run.sh unchanged: it still greps "finished with status failure", which e2e now prints for failure and error
  - docs/LAB.md updated; make check-e2e-parse added to the Makefile help
  - sh -n on the three scripts, `sh scripts/check-links.sh`, `go vet ./...`, `go test ./...` exit 0; no merge commit on the branch
- next: master runs the live lines: make clean, make up, make e2e, make e2e-failed-run, runner stopped, make down, make clean
- evidence:
  - Not run: make and docker are denied on this seat (permission class), so make up, make e2e, make e2e-failed-run and the runner-stopped control are unrun
  - The fixtures in e2e-parse-check.sh follow the shape pasted in the order; its "..." parts (creator and the rest) are my reconstruction, not captured bytes
  - master-was-wrong: no measurement here refutes it; I could not probe a live 1.22.6
  - Risk, mine: the failed-run control restores the workflow after the check, which starts one more run; e2e on the same head waits for the highest run
- updated: 2026-10-24T02:40:00Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 2 |
