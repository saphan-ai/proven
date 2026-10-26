# REVIEW — the-image · round 3

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T00:47:26Z

## Summary

Trace normalization can hide non-path differences; sequential install commands lose the checkout directory. Docker checks unrun on this seat.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-image

### Ground
Reviewed HEAD `078bc85` against `2b8fd4a` (the measured `git merge-base origin/main HEAD`). The author's Docker success and negative-control claims are not this seat's measurements: `command -v docker` returned 1. No numeric quality metric, command, baseline, or tolerance is declared in the order. No manifest, lockfile, test assertion, skip, lint exclusion, or quality threshold changes in the diff. Check ranking: shell and make are project checks; the sample and its own comparisons are suite checks, not external audits. No numeric bar or exception can be scored.

### Differential
BLOCKING — `scripts/image-check.sh:28,51-54`: the round-two narrowing does not meet path-only normalization. `sed -E 's#^trace: .*$#trace: <trace path>#'` masks every line beginning `trace: ` even when its content is not a path. Independently measured false world: `printf 'trace: SUCCESS\ntrace: FAILURE\n' | sed -E 's#^trace: .*$#trace: <trace path>#'` produced the identical `trace: <trace path>` line twice. Two otherwise identical successful reviews whose output differs in such a line can pass the final comparison. `internal/cli/format.go:119` emits the real trace line from `result.TracePath`, but this script accepts arbitrary binary/image outputs and verifies neither the line's path shape nor its uniqueness or final position. Remedy: replace only the actual expected trace-path line and reject unexpected or duplicate trace lines; demonstrate that non-path trace-prefixed differences remain different. Earlier broad trace erasure is narrowed but NOT ADDRESSED as a path-only guard.

OWED — `docs/INSTALL.md:31,47-48`: the previous sample instructions change CWD to `/tmp/proven-run`. The new manual check uses `scripts/image-check.sh` and `./bin/proven` relative to the clone, without a return to it. Running the guide sequentially therefore cannot find the script or binary. Remedy: instruct the reader to return to the checkout before make and the manual check, or give anchored paths. This predates round three; report it without opening another round solely for it.

### Unplug
By inspection, `scripts/image-check.sh:38-50` fails on each nonzero binary/image version and sample-review exit; `cmp` at lines 42 and 53 detects unnormalized version/review differences. False-world answers: exit at the corresponding guard, except the measured non-path `trace: ` discrepancy, which normalizes to equality. `lab/scripts/image-check.sh:6-8` resolves the tag through the loader also used by `lab/████████████████`; `Makefile:25-26` builds both prerequisites. Both container runs disable pulling, and lines 58-61 compare image revision label to checkout HEAD. No Docker-backed unplug control or 97 MB size measurement was possible here.

### Closure
`3f4550f` ADDRESSED the identical failed `--version` exits by requiring success, but the path-only trace-normalization finding remains NOT ADDRESSED. `078bc85` ADDRESSED the tag override, no-pull, and revision-label concerns by inspection. A manually supplied binary has no independent commit attestation; `make image-check` instead builds both sides from the checkout. The no-git fallback expressly skips revision comparison, so it cannot establish same-commit provenance outside a checkout.

### Remainder
No dependency delta or quality-bar exception. Declined to judge licenses, live review, publishing, and provenance beyond the label or changes outside the review diff.

### Profile
Docs and shell/CLI. The prose describes the check but the sequential manual command is not runnable from the preceding CWD. Image build and checker use the same tag loader. No external or declared numeric testing bar; measure a baseline first rather than inventing a target.

### Focus
1. Reject either broken run: exit guards exist (`scripts/image-check.sh:38-50`), but output equality can be false-positive for `trace: ` content (`:28,51-54`).
2. Hide only two paths: NOT MET for non-path trace-prefixed text; fixture replacement is a literal host/container fixture pathname (`scripts/image-check.sh:51-52`).
3. Build the matching image and binary: MET by source inspection of `Makefile:25-26`, `lab/scripts/image-check.sh:6-8`, `lab/████████████████`, and `lab/████████████████`; runtime UNRUN.
4. Run docs as written: NOT MET from the documented CWD (`docs/INSTALL.md:31,47-48`); runtime UNRUN.

### Conditions
Reject until the trace path is the only trace-prefixed text normalized and the negative control distinguishes non-path lines. Correct the guide's directory transition. Docker-backed runs and negative controls need repeating on an enabled seat before claiming their results as this leg's evidence.

### Unrun
`make image-check`, image execution and inspection, image size, and the author's Linux Docker negative controls: `command -v docker` exited 1 on this seat. No numeric base/tip testing-bar comparison exists because no metric, command or baseline was declared. The shell normalization differential was measured independently. Did not run `proven scan` inside the checkout.

## The review status as filed

## REVIEW — the-image
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: Trace normalization can hide non-path differences; sequential install commands lose the checkout directory. Docker checks unrun on this seat.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-25T00:44:57Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 3 |
