# REVIEW — judges-lab-makefile · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T19:28:21Z

## Summary

The e2e success check accepts stale or unrelated comments; Docker-backed validation was unavailable on this seat.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — judges-lab-makefile

### Ground

RED — BLOCKING: The e2e check can report a successful rehearsal when the current workflow has not posted a review. It reads any occurrence of `proven-review` in the first open PR's comments, not a comment from the current run. This breaks the central e2e proof boundary.

Reviewed `9303b2d...d8514fe`; `git rev-parse --short=7 HEAD` printed `d8514fe`. The diff adds a Makefile, lab scripts, Dockerfiles, Compose, workflow, documentation and status records. The master's Docker results remain claims, not measurements from this seat. Here `make help` enumerated the documented targets and printed `go1.26.5`, `docker: not found`, `lab: not running`; `make check-links` printed `check-links: every relative link and metadata.json path resolves`. `sh -n lab/scripts/*.sh lab/runner/start.sh lab/seed/seed.sh` and `git diff --check 9303b2d...d8514fe` exited zero. `make e2e` printed `docker: not found` and Make exited 2: a missing-prerequisite refusal, not a lab test.

### Differential

BLOCKING · `lab/████████████████████` · The seed installs the workflow and opens a PR (`lab/seed/seed.sh:131-166`, `lab/seed/seed.sh:233-270`); the workflow writes a marker to a comment (`lab/workflow/proven.yml:46-65`). The e2e predicate instead searches the entire JSON reply for a substring, without checking body prefix, comment provenance, workflow run, creation time or revision. An unrelated comment containing the word or an earlier successful comment causes success despite a failed current job. The runner-presence check (`lab/████████████████████`) catches a stopped container, but a running, registered runner with a failed job remains a false green. Remedy: identify a fresh run and verify its new expected PR comment and successful job. Add a negative control with runner alive, prior matching comment present, and current job failing; e2e must refuse. A limited predicate reproduction, `printf '%s' '[{"body":"unrelated text says proven-review but no workflow ran"}]' | grep -q 'proven-review'`, exited 0; a stale marked comment also exited 0. This proves only the predicate, not the full Docker behavior.

OWED · `docs/LAB.md:67` · The guide says `TEST_IMAGE_DIGEST` is empty and first up pulls by tag, but `lab/versions.env:45` has a digest and `lab/███████████████████` pulls by digest. This gives judges the wrong expected output and repair instruction. Remedy: update the guide to describe the committed digest and actual pull path; documentation alone is not blocking.

### Unplug

An unrelated comment and a stale comment both give the substring predicate the same zero exit code as a valid review. The stopped-container runner check does not discriminate a running runner with a failed current job. Static syntax and link checks cannot measure lab liveness. The Docker-integrated runner-stop and failed-current-job negative controls were unrun because Docker is absent.

### Closure

Reject until e2e verifies the current run, demonstrated by a negative control and a successful clean run on a Docker-equipped seat. No product code was changed. Verify the documented output after the digest paragraph is corrected.

### Remainder

No further verified blocker. `lab/compose.yaml:28-29` binds the only published port using `LAB_BIND`, whose default is `127.0.0.1`; dind has no published port. `lab/███████████████████` prints the generated password as explicitly promised and stores it in a lab volume. `lab/.env` matches `.gitignore:5-7`. No tracked credential or private machine path was established from the reviewed diff; live leak inspection is unrun. A test-image digest is configured, though third-party image references are mutable tags and package installs are not pinned; no claim of byte-for-byte reproduction follows from this inspection. Declined to judge licenses (excluded by the order), registry availability and runtime containment without Docker. Ruling: the first-print password is intentional; if wrong, terminal logs may retain it.

### Profile

Security and docs profiles apply: inspect localhost bindings, inner Docker, tokens, ignored secret files and judge-facing instructions. No changed Go dependency manifest or lockfile. No numeric metric/baseline is declared for this wagon: coverage, mutation, lint and type counts are UNRUN/not declared, not met. Check ranking: external = shell parser (run) and Docker runtime (unavailable); project = link check and docs consistency; suite = Docker e2e (unavailable; most circular). No baseline/tolerance exception or test suppression observed in the diff; there is no measured base/tip numeric delta.

### Focus

Contract items 1–7: help and links checked locally; e2e positive predicate has a reproducible false-world answer; stopped-runner behavior, credential exposure, runtime network containment and actual digest pull need Docker; versions and scripts inspected; no private host or path identified in reviewed files. The digest paragraph in the guide is stale.

### Conditions

BLOCKING: reject stale or unrelated matching comments when the current workflow fails; exercise a failed-current-run negative control and a clean positive run. OWED: fix the guide's digest paragraph. This is reject, not accept over unrun checks.

### Unrun

Docker is absent (`make help`; `make e2e`: `docker: not found`). `make up`, clean e2e, stopped-runner test, `make down`, live secret/port inspection and image scanning were not run. `make build`, `make sample`, `make livecheck` and Go suites were not run because they write artifacts outside this leg's confined allowance; local Go is `go1.26.5`, distinct from the configured build image `golang:1.26.10`. `make scan` needs Docker and can rewrite tracked `docs/LAB.md`. No base-commit metric measured; no recorded numeric baseline declared.

## The review status as filed

## REVIEW — judges-lab-makefile
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The e2e success check accepts stale or unrelated comments; Docker-backed validation was unavailable on this seat.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T19:25:52Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 4 |
