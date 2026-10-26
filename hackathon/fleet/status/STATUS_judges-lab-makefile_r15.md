# STATUS — judges-lab-makefile · round 15

- state: ready-for-review
- executor: claude-code
- model: claude-sonnet-5-5
- pool: not recorded
- cost: not recorded
- duration: 4 min
- last commit: 55ec036855dc4a4526e5061da6860af808adc434 [judges-lab-makefile] lab: .env.example names the built NVIDIA mode
- updated: 2026-10-24T05:43:58Z

## Summary

Round 8. The review comment's marker is now `<!-- proven-review run=N sha=<head> -->`, filled from github.run_number and the PR head sha. e2e accepts only a comment by LAB_ADMIN_USER whose marker names the highest run and the head commit; no time is compared.

## The report as filed

The executor's own status, verbatim. Acceptance, evidence and next steps are in it.

## STATUS — judges-lab-makefile
- state: stop-2-ready
- abstract: Round 8. The review comment's marker is now `<!-- proven-review run=N sha=<head> -->`, filled from github.run_number and the PR head sha. e2e accepts only a comment by LAB_ADMIN_USER whose marker names the highest run and the head commit; no time is compared.
  make check-e2e-parse covers: unrelated author, run that posted nothing, same-second older review, other head sha, escaped marker. Docker is not on this seat's allow list, so no live lab line was run; the master runs make up/e2e/e2e-failed-run.
- last commit: 55ec036855dc4a4526e5061da6860af808adc434 [judges-lab-makefile] lab: .env.example names the built NVIDIA mode
- done: Both BLOCKING findings of the codex review of dcd86fd are fixed in code and covered offline.
  - master, 2026-10-24: absorbed main d510ac5 (4241577: docs/LAB.md add/add — main's keys page kept as the runbook section \"The keys a live review needs\") and named the now-built NVIDIA mode in lab/.env.example (55ec036); go build/vet, check-links green on the absorbed tree
  - author check and run+sha marker: lab/████████████████████ review_for, lab/██████████████, lab/workflow/proven.yml; proven by `sh lab/██████████████████████████` → "e2e-parse-check: ok" (accept paths and every refusal both exercised)
  - same-second control: an older run's review (run 3) against run 4 at one timestamp is refused, no time compared
  - docs/LAB.md updated (marker, author, controls, kept-volume note)
  - `go vet ./...` clean, `go test -count=1 ./...` all ok, `sh scripts/check-links.sh` ok, `sh -n` ok on the three changed scripts, `git grep -n '\.saphan' -- Makefile lab docs README.md` prints nothing
  - no merge commit, no Co-authored-by on 9303b2d..HEAD, author Marcin Marzec <██████████████████████>
- next / blocked-on: the master runs on a Docker seat: make clean && make up && make e2e (0), make e2e-failed-run (0), runner stopped → e2e non-zero naming the runner, make down, make clean.
- artifacts:
  - lab/████████████████████
  - lab/██████████████
  - lab/██████████████████████████
  - lab/workflow/proven.yml
  - docs/LAB.md
- evidence:
  - NOT RUN (docker is not in this seat's allowed programs, so make up, e2e, e2e-failed-run and the runner-stopped control were not run here)
  - UNMEASURED: whether act on Gitea 1.22.6 fills github.run_number and github.event.pull_request.head.sha. By Gitea's design run_number equals the run index in target_url, but I did not see it live. If either is empty, the marker reads `run= sha=` and e2e fails closed (waits, then names the run and what it ignored); the master should read the posted comment once.
  - The live failed-run control (e2e-failed-run.sh) is unchanged; it still greps 'finished with status failure', which e2e still prints.
  - A kept volume whose PR branch holds an older workflow posts the old marker and e2e fails closed; docs/LAB.md says to run make clean && make up.
- note: master-was-wrong: none. Both findings are real. The old review_since took only a time and never read the author or the run.
- updated: 2026-10-24T05:10:00Z

CLOSE-OK · judges-lab-makefile · commit 712a656a1aab47fbb2177c9e811f1b039350becf · branch fix/judges-lab-makefile · worktree ███████████████████████████████████████████████████████

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| e-mail or user name | 1 |
| file and line in Saphan Studio's private code | 6 |
