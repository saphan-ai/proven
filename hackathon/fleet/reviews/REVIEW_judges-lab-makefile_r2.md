# REVIEW — judges-lab-makefile · round 2

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T04:39:44Z

## Summary

BLOCKING: e2e accepts an unrelated account's marker after a successful run without a review, and an older comment from the same second. Offline controls reproduced; live Docker tests unrun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — judges-lab-makefile

### Ground
Reviewed `9303b2d...dcd86fd8243401d3c6063fb253f53e1c1c0e2ce4` at the worktree tip. `git rev-parse --short=7 HEAD` returned `dcd86fd`; `git rev-parse HEAD` returned `dcd86fd8243401d3c6063fb253f53e1c1c0e2ce4`. This is a file and offline-parser review, not an independent live lab run. The master's reported live Docker measurements are claims, not my results. `make check-e2e-parse` returned `e2e-parse-check: ok`; `make check-links` returned `check-links: every relative link and metadata.json path resolves`; `sh -n lab/scripts/*.sh lab/seed/seed.sh` and `git diff --check 9303b2d...dcd86fd` exited 0. `make help` listed every target and identified Docker as missing. `make up` and `make e2e` both exited 2 from make, with `docker: not found` as the script's named cause. `git check-ignore -v lab/.env` identified `.gitignore:5`.

### Differential
BLOCKING — `lab/██████████████████████████`, called from `lab/████████████████████`, accepts any comment whose body starts with the marker and whose creation time is at least the latest successful run's first status. It never checks the comment author or associates the comment with that run. `docs/LAB.md:79` explicitly says a comment by anyone else does not count. If a workflow succeeds but posts no review, a different account's marker comment posted after it makes e2e print success and exit 0. An offline Gitea-shaped JSON control returned `UNRELATED-AUTHOR-ACCEPTED` with `user.login=unrelated`, comment at `2026-10-24T04:31:02Z`, and run start `2026-10-24T04:31:01Z`; the matching successful status was parsed as `5 success 2026-10-24T04:31:01Z`. Remedy: bind proof to a verified workflow-produced review and run identity, rather than trusting time and marker alone; test a successful workflow with no review and an unrelated account's marker.

BLOCKING — `lab/██████████████████████████`, `docs/LAB.md:79`: comparing second-resolution timestamps with `>=` accepts an older review posted earlier in the *same second* as a later run's first status. A direct offline control with both timestamps `2026-10-24T04:31:01Z` returned `SAME-SECOND-OLDER-COMMENT-ACCEPTED`. This remains a false-positive even if author is checked. Remedy: bind review proof to the run by identity; a timestamp cannot distinguish events within one second. Add the same-second old-review control.

The prior review's failed-current-run condition has a source-level guard at `lab/████████████████████` and a lab control in `lab/███████████████████████████████`; I cannot independently confirm its live execution. The guide's digest paragraph at `docs/LAB.md:67` now describes the pull and tag steps at `lab/███████████████████`. The index digest was not independently fetched here.

### Unplug
`make check-e2e-parse` covers highest-run selection, failed status, and an old comment separated by one second; its false-world answer is FAIL for those synthetic inputs. It has no author mismatch, successful-run/no-post, or same-second old-comment control: `ok` is identical for correct and defective versions for those cases. Direct `review_since` probes produced the false-world acceptances above. `lab/████████████████████` names a stopped runner before accepting a comment; actual stopped-runner execution is unrun without Docker. `lab/█████████████████████████` cannot detect an unrelated fresh comment after a *successful* run.

### Closure
Reject this tip until e2e verifies the current run produced the review, including an author control and same-second event order. The master's reported positive and failed-run tests do not cover these false positives. I made no product-code changes, commits, or live lab changes.

### Remainder
OWED — `Dockerfile:5` retains `golang:1.26.8` as a standalone default while `lab/versions.env:11` uses `golang:1.26.10`. Make's image/push scripts pass `GOLANG_IMAGE`, so this is not a proven default-path defect using Make; clarify duplicated defaults if direct Docker builds are in scope. Clean-machine build, version availability, volume behavior, image-digest bytes, port bindings at runtime, and actual review posting remain unverified. Licences and unrelated source behaviors were deliberately not judged.

### Profile
Security/configuration: `lab/compose.yaml:28-29` publishes only the Gitea HTTP port on `${LAB_BIND}`; `lab/compose.yaml:35-47` declares the inner daemon without published ports; `lab/compose.yaml:49-68` does not mount the host Docker socket. `lab/seed/seed.sh:60-86,172-196` writes API and runner tokens into named volumes; `lab/███████████████████` stores and intentionally prints the generated password once as documented. `.gitignore:5` excludes `lab/.env`. These are source checks only; runtime, log, and image-layer checks require Docker. Dependency inputs: `lab/versions.env:11-12,36-45` sets the Go/Alpine, runner/Docker/Gitea, and test-image versions and digest, while `lab/seed/Dockerfile:7` installs unpinned Alpine packages; image availability/digest were not checked. No new Go module dependency was added. A literal scan of changed tracked files found no newly added private hostname, absolute workspace path, or private repository name.

### Focus
The proof boundary is e2e's success criterion, not whether one earlier lab run happened to leave a legitimate comment. `lab/█████████████████` prints complete comment JSON; that output alone does not prove run association. The scripts generally state their purpose; `lab/██████████████████████████` accurately describes its incomplete timestamp rule.

### Conditions
None. Both findings are blocking correctness defects; verdict `reject`, not a conditional accept. Reclassification requires a master gate-note reason, not a silent downgrade.

### Unrun
Live `make up`, `make e2e`, `make e2e-failed-run`, stopped-runner control, `make down`, `make clean`, `make scan`, image builds/pushes, and port/container/volume checks: Docker is not installed (`make up` and `make e2e` printed `docker: not found`; make exited 2). Go build, `go vet ./...`, `go test -count=1 ./...`, race and mutation checks: not run; this review targeted the lab and no Go product code changed. The master's reported runs do not substitute for mine.

## The review status as filed

## REVIEW — judges-lab-makefile
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: BLOCKING: e2e accepts an unrelated account's marker after a successful run without a review, and an older comment from the same second. Offline controls reproduced; live Docker tests unrun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T04:33:27Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 11 |
