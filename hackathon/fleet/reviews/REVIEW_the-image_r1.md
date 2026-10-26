# REVIEW — the-image · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T00:18:34Z

## Summary

Matching failed version commands can pass; trace normalization hides other differences. Docker is absent here, so image execution remains unrun.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-image

### Ground
`git rev-parse --short=7 HEAD` returned `b4769f8`. `git diff HEAD^...HEAD -- Makefile docs/INSTALL.md scripts/image-check.sh` shows one commit changing only these files. The requested `main` reference is not present: `git diff $(git merge-base main HEAD)...HEAD` failed with `fatal: Not a valid object name main`. Read the script, Makefile target, installation section, build and image scripts, Dockerfile, and `internal/cli/format.go:120`. `command -v docker` returned absent. The author's Linux-host numbers and asserted negative control were not rerun. No dependency manifest or lockfile changed, and no new install-time script was introduced.

### Differential
BLOCKING — `scripts/image-check.sh:36-39`: both `--version` invocations end in `|| true`, then only captured output is compared. False world: both executables return nonzero with the same error (including empty output), while sample scans return zero and match. The check reports `ok`, not a version failure. Remedy: require both version commands to exit zero before comparing their output; test each failure independently and both failing identically.

BLOCKING — `scripts/image-check.sh:27,47-50`: `sed -E 's#(trace[^:]*: ).*#\1<trace path>#'` matches a `trace...: ` substring anywhere on a line and erases the rest, not just an exact `trace: PATH` line. Local read-only probe `printf '%s\n' 'finding trace: evidence-A' 'finding trace: evidence-B' | sed -E 's#(trace[^:]*: ).*#\1<trace path>#' | sort -u` returned only `finding trace: <trace path>`. False world: a review line containing `trace: ` with differing evidence collapses to identical normalized text, so the script reports `ok` when other lines match. Remedy: anchor normalization to the actual path line (`internal/cli/format.go:120`) and test differences on a non-path line containing `trace: `.

OWED — `docs/INSTALL.md:39-52`: direct-check command requires `./bin/proven` and a local image, but the earlier build creates `/tmp/proven` and the image build is described, not invoked. Docker is listed as necessary for live mode only at `docs/INSTALL.md:14`, although the new commands require it. Remedy: add `make build` and `make image` prerequisites or give `make image-check` as the first-run command; update the prerequisite table. Actual commands UNRUN here.

### Unplug
Version false-world answer: equal failed output passes; `sh -c 'false > /dev/null 2>&1 || true; printf "exit=%s\n" "$?"'` printed `exit=0`. Binary scan failure: lines 43-44 refuse; image scan failure: lines 45-46 refuse; both fail identically: binary-side refusal first. Different surviving output: lines 49-50 refuse. Different version output: lines 38-39 refuse. A trace-path-only difference is intended to normalize; a difference on an ordinary line containing `trace: ` also normalizes, wrongly. These are control-flow readings and a sed probe, not executed Docker tests. Independently mutate both scan exit statuses and a non-path output on a Docker-capable seat before gating.

### Closure
Reject. Two blocking false-green paths violate the comparison contract. `Makefile:25-26` orders `build` then `image`; `lab/████████████████` builds the host binary and `lab/████████████████` builds the image from the same `ROOT` checkout, tagging `$PROVEN_IMAGE:latest`, which the target passes to the script. `Dockerfile` copies the fixture to `/opt/proven/fixture/discount`. These are static readings, not a successful build or content-identity proof.

### Remainder
No code under review changed. Declined to judge: licensing, container-size claim, provenance beyond same-checkout flow, and other reviewers. `internal/cli/run.go:20` holds the constant version line `proven development`; it does not encode a commit identifier. The make target, rather than the version string, supplies the checkout relationship. The manual command accepts arbitrary image and binary arguments, so it does not guarantee same-commit input alone.

### Profile
Docs and shell/build flow. No recorded coverage, mutation, lint, or type metric with a baseline is declared in the order; numerical testing-bar comparison is not applicable. Check ranking: Docker-backed sample comparison would be a project-owned integration check; static shell and sed probes are weaker. No bar-weakening move or exception in the three-file diff.

### Focus
1. `scripts/image-check.sh:36-39`: version success is not checked; identical failures can pass.
2. `scripts/image-check.sh:27`: trace normalization is not confined to the path line.
3. `Makefile:25-26`: prerequisites and tag resolve to one checkout by inspection.
4. `docs/INSTALL.md:39-52`: direct commands omit preparation and Docker prerequisites.

### Conditions
Before accepting: require successful version invocations; restrict normalization to intended trace and fixture paths; demonstrate negative controls above on a Docker-capable seat. Repair direct-command instructions and verify from documented starting state. No review condition was reclassified.

### Unrun
`make image-check`, `make image`, container scan, image size, and both documentation Docker commands: Docker is absent (`command -v docker` returned absent). Claimed Linux Docker-host outcomes and 97 MB size are not independent evidence on this seat. No Docker simulation substituted. Base/tip numeric tests: no metric and baseline declared. Complete `main...HEAD` comparison: local `main` ref absent, so inspected `HEAD^...HEAD`. No `proven scan` inside checkout.

## The review status as filed

## REVIEW — the-image
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Matching failed version commands can pass; trace normalization hides other differences. Docker is absent here, so image execution remains unrun.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T23:47:08Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 2 |
