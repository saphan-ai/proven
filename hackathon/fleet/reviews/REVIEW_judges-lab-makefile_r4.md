# REVIEW — judges-lab-makefile · round 4

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T08:00:07Z

## Summary

The previous e2e provenance defect is addressed offline. The reviewed tip fails the Go suite under this seat's 0002 umask; Docker-backed checks are unrun because Docker is absent. See notes for the blocking finding and required live checks.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — judges-lab-makefile

### Ground

**RED — BLOCKING: the reviewed tip fails the full Go suite under the seat's group-writable umask.** `internal/testrun/written.go:209-224` adds read/search bits to existing modes but preserves write bits. `internal/pipeline/trees_readable_test.go:53-58` requires materialized files not be group-writable. With `umask 0002`, `go test ./...` exits 1: `TestMaterializedTreesAreReadableByOthers` reports both `base/go.mod` and `change/go.mod` at mode 664. Independently repeating `umask 0002; go test ./internal/pipeline -run '^TestMaterializedTreesAreReadableByOthers$' -count=1` exits 1 with the same two failures; with `umask 0022` it passes. The new tree export preserves host group-write access even while the container mount is read-only. Remedy: normalize exported file modes without group/other write, retaining needed owner write, then repeat both umask controls and the full suite. This is in the reviewed `9303b2d...9678e10` diff (absorbed change), not a failure of the lab parser itself.

`git rev-parse --short=7 HEAD` printed `9678e10`. `make help` listed all documented targets and reported `docker: not found`. `make check-e2e-parse` printed `e2e-parse-check: ok`; `make check-links` printed its success line; `go vet ./...` exited 0; `git diff --check 9303b2d...9678e10 -- Makefile lab docs/LAB.md` exited 0. `go test ./...` exited 1, solely at the tree test in its output. No tracked file changed. These are tip readings, not the master's fresh-clone measurements.

remark: The successful parser check cannot establish that Gitea, Docker-in-Docker and the runner cooperate on this seat; Docker is absent.
remark: The 0022 test pass does not cancel the 0002 failure; inherited umask changes exported file modes.

### Differential

Previous rejection, ADDRESSED in the offline predicate: `lab/seed/seed.sh:169-226` makes a distinct non-admin bot, writes its token to the repository Actions secret, and `lab/workflow/proven.yml` posts with it. `lab/████████████████████` requires bot author, run and head; `lab/████████████████████████████████` supplies the exact admin-authored marker in the no-workflow-comment world and requires refusal. The offline check passed. Live producer attribution remains unverified here, not inferred from synthetic JSON.

`lab/████████████████████` distinguishes a stopped or unregistered runner before reading statuses; `lab/████████████████████` waits on the current PR head's highest run and refuses failed status or missing bot comment. `lab/compose.yaml` publishes only the configured Gitea web port on the configured loopback bind and keeps the inner daemon and runner on the lab network. `lab/███████████████████` pulls the Prover image by digest when the checked-in digest is set; the setting permits a documented unpinned override. The admin password originates in `lab/███████████████████` and stays in the seed volume; bot credentials are made in the seed container and handed to the Actions secret. No literal credential was found in tracked lab defaults. These are code observations, not a running-container audit.

The review scope includes the absorbed materialization change; the test failure is a new finding, not a resurrection of the earlier e2e finding. No new production lockfile or dependency manifest changes were found in the scoped diff. No declared numeric coverage, mutation, lint or type baseline was supplied; do not invent a target. No bar-weakening skip, suppression or test deletion was found in lab scripts or Makefile. Check ranking: Docker-backed live e2e is external integration evidence; `make check-links` and diff checking are project checks; offline parser and Go tests are suite checks, whose own fixtures can be circular.

### Unplug

The offline admin-marker control changes only author login from the accepted bot fixture and returns refusal. A same-head, same-second old-run marker and an unrelated author also return refusal (`lab/████████████████████████████████`). False world for `service_running runner`: no running runner gives an explicit error before an old comment could be accepted; read but not exercised live. False world for status selection: an old success must not beat a newer pending or failed run; recorded-shaped JSON tests cover it (`lab/████████████████████████████████`). False world for group-write safety: under 0002 the test turns red; 0022 hides this particular failure. Docker-backed negative controls remain unrun.

### Closure

Verdict: **reject**. The earlier e2e provenance defect is addressed by the distinct identity and failing admin-marker fixture, but the full-suite failure and unverified live lab prevent acceptance. Repair mode normalization, then run the full suite under 0002 and 0022 and the Docker-backed lab and negative controls on a Docker seat. No review condition was downgraded; no exception was requested or granted.

### Remainder

OWED: `lab/███████████████████████` checks any comment containing `proven-review`, without matching author, head or run. Thus `review: present` can mean an older or human-authored marker while `make e2e` correctly refuses. Make status show a weaker explicitly named observation or share the e2e predicate. Status is not itself the e2e verdict, so this is not another blocking reason.

Declined to judge: actual image availability, runtime permissions, external scanner results, registry publication, browser rendering and workflow network behavior. This seat cannot run the Docker lab. Reading the source does not prove a service is reachable.

### Profile

Security/configuration and docs profile. Read secret, API, status and runner paths together and checked the negative author control. `make clean` targets the repository's `bin`/`dist` and Compose project's volumes (`lab/█████████████████████`); it was not executed because no lab or build artifact was made. No host or repository material was intentionally changed. Read-only container mounts do not remove host group write permission.

### Focus

Judge-facing commands, comment provenance, stopped-runner and failed-run discrimination, defaults and secret placement, and the complete write-set's suite. No claim is made that a recorded-shape fixture proves a live API response. The prior admin-marker control now fails offline; the new mode finding comes from a separate absorbed change inside the requested diff.

### Conditions

BLOCKING: normalize exported tree permissions independently of umask; show the 0002 control turn green without weakening its assertion, then pass `go test ./...`. After that, obtain live `make up`, `make e2e`, `make e2e-failed-run`, stopped-runner `make e2e` refusal, `make down` and `make clean` observations on a Docker-equipped seat before requesting acceptance. This is not acceptance with conditions: the current verdict is reject.

### Unrun

Docker is absent (`make help` printed `docker: not found`; `command -v docker` found no executable): `make up`, live `make e2e`, `make e2e-failed-run`, stopped-runner control, `make down`, `make clean`, image pull/build and scanner are UNRUN. Base-commit metric comparisons are UNRUN: no numeric metric, baseline or command was declared and this review seat did not create a base checkout. No tests were claimed as live merely because the master supplied results from another seat.

## The review status as filed

## REVIEW — judges-lab-makefile
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The previous e2e provenance defect is addressed offline. The reviewed tip fails the Go suite under this seat's 0002 umask; Docker-backed checks are unrun because Docker is absent. See notes for the blocking finding and required live checks.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T07:57:48Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 10 |
