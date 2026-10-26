# REVIEW — judges-lab-makefile · round 3

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T06:06:55Z

## Summary

A matching admin-authored marker after a successful workflow with no review passes e2e's parser; producer attribution is not verified. Offline checks pass; Docker-backed checks are unrun because Docker is absent.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — judges-lab-makefile

### Ground

**RED — the e2e proof does not identify the workflow as the comment's producer.** `lab/██████████████████████████` accepts a matching marker, run number, head SHA and author login. `lab/████████████████████` treats that as a successful workflow review. The workflow posts using the seeded admin account (`lab/seed/seed.sh:163-169`, `lab/workflow/proven.yml:30-68`), which is also the account the judge uses to sign in (`docs/LAB.md:11-15`). An admin-authored marker is indistinguishable from the workflow's comment. This breaks the promised requirement that e2e pass only when the workflow posted the review (`docs/LAB.md:122`). Remedy: establish producer-bound evidence for the successful job, for example a dedicated workflow-only identity bound to the job, and test the missing-workflow-comment world with a matching admin-authored marker. Merely checking author, marker and timestamp does not prove provenance.

Read HEAD from `git rev-parse --short=7 HEAD`: `55ec036`. Review range: `9303b2d...55ec036`. No product files changed in this review. The master's Docker measurement is a claim, not this seat's result. The lab's only published host port is HTTP on the configured bind (`lab/compose.yaml:28-29`); the inner daemon has no published host port (`lab/compose.yaml:35-47`). The seed and workflow use the lab-internal Gitea address (`lab/seed/seed.sh:128-134`). The ignored local environment file and digest pull are visible in `.gitignore:5-7` and `lab/███████████████████`; a running lab and pulled digest are not verified here.

### Differential

Previous review's unrelated-account finding: **partly addressed, not closed as a provenance guarantee**. The new author check refuses a different login (`lab/████████████████████████████████`), but the shared admin login remains a false-positive source. Previous review's same-second older-run finding: **addressed in the offline control**; the marker carries run and SHA instead of comparing timestamps (`lab/██████████████████████████`, `lab/████████████████████████████████`). The other-head control also refuses (`lab/████████████████████████████████`). No dependency manifest change in the lab diff; image and tool versions are centralized in `lab/versions.env`, but live pulls are unrun.

### Unplug

With `lab/████████████████████` sourced, passed valid JSON containing one comment by the admin with marker `run=3` and a matching 40-character head SHA, body `Injected by admin, not the workflow`; `review_for "$comment" 3 "$sha" proven` exited **0**. In the false world, a successful status exists but the workflow posted no comment, followed by a matching human admin comment: e2e's success branch accepts it. Existing offline checks instead test no comment, non-marker human text, and a marker by another login; none distinguishes this false world. No live workflow was altered. The runner-down guard in `lab/████████████████████` would return 1 before checking comments, but its live negative control could not be run without Docker.

### Closure

Verdict **reject**. The false success above is BLOCKING: the current predicate returns success for an unverified producer. Accept only after a control makes this input fail while preserving acceptance of the genuine workflow comment. The same-second and unrelated-login offline controls are green; neither cures the shared-account case.

### Remainder

No recorded numeric metric, baseline, tolerance or exception was declared for this lab change; no metric delta is asserted. Check ranking by independence: external Docker/Gitea run (unrun) > project `make check-links` (passed) > repository-owned `make check-e2e-parse` (passed). The last is circular for this missing case. The broader `go test ./...` failed only in `internal/pipeline/TestMaterializedTreesAreReadableByOthers` under this seat's `umask 0002` (two materialized files mode 664). Rerunning that test with `umask 0022` passed; outside the lab finding, this contradicts a blanket claim that the suite passes under this seat's default umask. `git diff --check 9303b2d...HEAD` found a trailing blank line at `LICENSE-docs:396`, outside the lab scope. Declined to judge: published image availability, host-specific Docker behavior and third-party image findings without a daemon.

### Profile

Security and docs: checked comment provenance across status, parser, seeded account and workflow caller; read the secret's storage path and port declarations. No live authentication or image-layer test ran. Shell scripts explain their intent; this review does not warrant line-by-line compliance with ADR-0027 over unrelated files. No private infrastructure name or address is quoted here.

### Focus

The judge-visible Makefile targets exist (`Makefile:12-64`). `make help` printed the targets and reported `docker: not found`. `make check-e2e-parse` printed `e2e-parse-check: ok`. `make check-links` printed `check-links: every relative link and metadata.json path resolves`. `make up` and `make e2e` each refused with `docker: not found` (make exit 2). Those refusals do not prove a Docker-equipped clean machine works; they only show the missing prerequisite is named on this seat. `go test ./...` exited 1 as described above.

### Conditions

Fix producer attribution and add the negative control with successful status, absent workflow review, and matching admin-authored marker. Re-run `make check-e2e-parse`, then real `make e2e` and `make e2e-failed-run` on a Docker-equipped seat. No condition is reclassified downwards by this reviewer.

### Unrun

Live `make up`, `make e2e`, `make e2e-failed-run`, runner-down control, `make down` and live image/digest verification: Docker executable is absent on this seat. The first two commands were attempted and refused as above; no containers or lab volumes started. Baseline/tip numeric metric comparison: no metric or baseline was declared. No Docker-dependent claims are reported as met.

## The review status as filed

## REVIEW — judges-lab-makefile
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: A matching admin-authored marker after a successful workflow with no review passes e2e's parser; producer attribution is not verified. Offline checks pass; Docker-backed checks are unrun because Docker is absent.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T06:01:09Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 9 |
