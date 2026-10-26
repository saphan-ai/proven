# REVIEW — the-image · round 5

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T01:12:48Z

## Summary

Static review finds the round-four fixes addressed; Docker-backed negative controls remain required before acceptance.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-image

### Ground
Reviewed c96e18fd4bd647cd1be8909692d380da6023880f (fresh `git rev-parse HEAD`) against the `origin/main` merge base; local `main` does not exist. The four-file diff adds `scripts/image-check.sh`, `lab/scripts/image-check.sh`, the `Makefile` target, and the install instructions. Docker is not on this seat's PATH. The master's Docker measurements are claims, not reproduced evidence.

### Differential
`scripts/image-check.sh:49-54` refuses a nonzero exit from either version command and compares the full outputs. `scripts/image-check.sh:58-65` refuses a nonzero exit from either sample review before comparing normalised outputs. False-world answers: one failing binary version, image version, binary review, or image review returns FAIL; different successful version text or review text outside the normalised paths returns FAIL. These are deductions from shell control flow, not observed runs. `scripts/image-check.sh:32-38` substitutes only literal fixture-path occurrences and rewrites only the final line if it begins `trace: `; intervening `trace: ` lines remain compared. The final line's entire suffix is treated as the path, so this check does not independently validate that suffix's structure; the current renderer at `internal/cli/format.go:120` emits only the trace path there. This is a scope limit, not an observed failure.

`Makefile:25-26` runs `build` and `image` before the wrapper; `lab/scripts/image-check.sh:5-8` reads the same image setting as `lab/████████████████` and supplies that `:latest` tag. `scripts/image-check.sh:51,60` refuses pulling another tag; `scripts/image-check.sh:69-76` compares the image revision label with HEAD in a checkout. A missing image, stale revision, or failed inspect does not yield `ok` by this path. The revision label attests the build input's checkout HEAD, not a clean tree or the binary's embedded revision. `docs/INSTALL.md:39-58` supplies hand commands from the repository with the default image tag; custom `PROVEN_IMAGE` users must substitute their configured tag in the hand commands. The Make target itself handles that override.

### Unplug
Static negative controls: invert either success test at `scripts/image-check.sh:49-52` or `:58-61` and two identically failing commands could pass `cmp`; delete `cmp` at `:53-54` or `:64-65` and divergent outputs could pass; remove the revision check at `:69-73` and a stale image could pass. The current code has these guards. No live unplug experiment ran because Docker is absent; therefore these controls remain unverified by execution.

### Closure
Round-two version-exit and overbroad trace-line issues: ADDRESSED at `scripts/image-check.sh:49-54,37-38`. Round-three wrong image tag, implicit pulls, and stale image: ADDRESSED at `lab/scripts/image-check.sh:5-8` and `scripts/image-check.sh:51,60,69-73`. Round-four last-line-only trace substitution, literal fixture replacement, and repository-anchored documentation: ADDRESSED at `scripts/image-check.sh:32-38` and `docs/INSTALL.md:44-49`. `sh -n scripts/image-check.sh`, `sh -n lab/scripts/image-check.sh`, and `git diff --check` passed; `make -n image-check` lists build, image, check in that order. These checks do not run either binary or image.

### Remainder
No new blocking source finding. No dependency manifest, lockfile, quality threshold, test assertion, skip, or suppression changed. No declared numeric testing-bar metric, command, baseline, or tolerance exists in this order; base/tip values and ranking are not applicable. Declined to judge licences, unrelated changes, live-service behaviour, and any claim that a revision label proves a clean source tree.

### Profile
Docs and shell/build integration. Docs commands were read against the Makefile, image builder, and Dockerfile; they were not executed on this seat. Check ranking: shell syntax and Make dry run are project-tool checks; the Docker image comparison is a suite-like self-check that needs independent negative controls. No external metric was declared.

### Focus
FILE·LINE·CONTROL: `scripts/image-check.sh:49-65` separates success from output equality; `scripts/image-check.sh:32-38` narrows normalisation to fixture strings and the last trace line; `Makefile:25-26` and `lab/scripts/image-check.sh:5-8` bind build and tag; `docs/INSTALL.md:39-58` gives the default-tag commands. Verdict is conditional because the runtime and negative controls were not run here.

### Conditions
Before accept, run `make image-check` on a Docker-capable checkout and independently force one nonzero version exit on either side, one nonzero sample exit on each side, one review-output difference outside the paths, and a stale image revision. Require a nonzero check for each false world and a successful, equal baseline. Report outputs and image size; this seat cannot certify the master's claimed 97 MB.

### Unrun
Docker-backed `make image-check`, both review processes, image size, stale-image and forced-failure negative controls, and the Docker hand commands: `command -v docker` found no executable. No Go build or `proven scan` was run, to avoid generating checkout artifacts and because Docker is required for the comparison. The absence of Docker cannot prove the image succeeds or fails.

## The review status as filed

## REVIEW — the-image
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Static review finds the round-four fixes addressed; Docker-backed negative controls remain required before acceptance.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T01:09:05Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 1 |
