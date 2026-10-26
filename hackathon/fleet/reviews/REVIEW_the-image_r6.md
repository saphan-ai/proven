# REVIEW — the-image · round 6

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T01:36:39Z

## Summary

c96e18f read in full: the script cannot report ok when either side's version or sample review fails or differs; normalisation is fixed-string fixture + last-line trace only; make image-check checks the tag make image builds against HEAD's label. Docker was denied on this seat, so nothing was executed (Unrun non-empty, law #53, no accept). Condition: the master's docker controls are pasted into the gate note. Owed: 4 small docs/comment nits.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-image

### Ground
- Reviewer: claude-code (Opus 5.5), REVIEW role, headless, 2026-10-25.
- `git rev-parse --short=7 HEAD` printed `c96e18f`; `git merge-base main HEAD` printed `2b8fd4a34666b6fe5406bf7e1af21932251f2582`.
- Scope read in full: `git diff 2b8fd4a34666b6fe5406bf7e1af21932251f2582...c96e18f` — `scripts/image-check.sh` (new, 80 lines), `lab/scripts/image-check.sh` (new, 8 lines), `Makefile` (+`image-check` target, .PHONY), `docs/INSTALL.md` (+"The container image", prerequisite row).
- Callers and guards read so I could judge them: `lab/██████████████` (load_versions keeps env, `oci_revision` = `git rev-parse HEAD`), `lab/████████████████` (tags `$PROVEN_IMAGE:latest` and `:$version`, sets the revision label), `lab/████████████████`, `lab/versions.env` (`PROVEN_IMAGE=saphan/proven`), `Dockerfile` (ARG REVISION default `unset`; fixture copied to `/opt/proven/fixture/discount`), `.dockerignore`, `internal/cli/run.go:20` (`versionLine = "proven development\n"`, constant), `internal/cli/format.go:60-125` (text formatter; `trace: <path>` is the last stdout line when a trace exists).
- The master's measurements in the order (docker host, 2026-10-25) are the author's claims. My seat could not run them again, so I have not used them as evidence (R1).

### Differential
How each false world reaches the script's verdict (property 1), read from `scripts/image-check.sh`:
- The binary's `--version` fails: `:49-50` `|| fail "... exited $?"` → FAIL. In `a || b`, `$?` is a's status, so the code is reported correctly.
- The image's `--version` fails (no such image, docker down, no docker): `:51-52`, `--pull=never` → FAIL (125/127). Through make, `need_docker` refuses earlier.
- The version lines differ: `:53-54` `cmp -s` → FAIL. remark: the version line is the constant `proven development` (`internal/cli/run.go:20`), so this comparison cannot tell two builds apart today; the script says so at `:67`.
- The binary's sample review exits non-zero: `:58-59` → FAIL. The run happens in an empty temp dir, never in the checkout.
- The image's sample review exits non-zero (for example a broken fixture or AppleDouble files): `:60-61` → FAIL. This is the round-1 hole (two reviews failing the same way), and it is closed.
- Both succeed but differ outside the two paths: `:62-65` `cmp -s` on normalised output → FAIL, both files printed.
- The image is from another commit (git checkout): `:69-72` label ≠ HEAD → FAIL, naming both. A plain `docker build` image carries `unset` (Dockerfile ARG default) → FAIL.
- The image is from another commit (not a git checkout, or git missing): `:74-75` prints `remark:` and continues to `ok`. The script does report this world, but its verdict reads the same as when the image is current. That is a disclosed limit, not a false green: the remark line names it.
- Both outputs are empty and both exit 0: they would compare equal. Not reachable, because the text formatter always prints `review class:` first (`format.go:61`).
Nothing above reports `ok` when either side's review fails. Property 1 holds by reading.

### Unplug
Mutations of each guard, made by reading because my seat could not execute them (R3):
- Drop `|| fail` at `:59` or `:61`: a failing side would then be compared, which is the round-1 defect. The master's negative control (a one-line change to the image's fixture → "the image's sample review exited 2") is the observable that turns red. I could not run it again: UNRUN.
- Remove the `held[NR] ~ /^trace: /` condition (`:37`): the trace line would differ (container `/work/...` against the host temp dir) and `cmp` would FAIL. This is a false red, not a false green.
- Widen the normalisation (for example back to a sed pattern): the round-4 fix replaces the fixture path with `index()` as a fixed string, so regex metacharacters in `$root` (`.`, `+`) cannot match more than the literal path.
- Remove the label check (`:69-72`): a stale image whose sample output has not changed would pass. The master measured the "commit after build → FAIL" control on docker. UNRUN here.

### Closure
- Property 2, the normalisation hides only the two paths: it holds. The fixture path is replaced as a fixed string on every line (`:33-35`). Only the LAST line, and only when it starts with `trace: `, is rewritten (`:37`). `format.go:119-122` writes exactly `trace: %s` as the final stdout line. If stderr adds anything after it (2>&1), the last line is no longer a trace line, it is compared as printed, and the result is a false red, never a false green. Every other line, including the same-family sentence read from the trace (`format.go:114`), is compared byte for byte.
- Property 3, `make image-check` builds what it checks: it holds. The target depends on `build image` (`Makefile:24`). The wrapper sources `lib.sh` (so `PROVEN_IMAGE` from the environment wins, as in `image.sh`) and checks `$PROVEN_IMAGE:latest` against `$ROOT/bin/proven` (`lab/scripts/image-check.sh:8`). `image.sh` tags exactly `$PROVEN_IMAGE:latest` and sets the revision label from `git -C $ROOT rev-parse HEAD`, which is the same value the script compares at `:70`. `ROOT` is the same on both paths. `--pull=never` keeps a registry image out.
- Property 4, the docs commands: they are correct by reading. `cd "$(git rev-parse --show-toplevel)"` then `sh scripts/image-check.sh saphan/proven:latest ./bin/proven` matches the default `PROVEN_IMAGE` and the script's paths. `docker run --rm saphan/proven:latest scan --sample /opt/proven/fixture/discount` matches the Dockerfile's `COPY` destination and `WORKDIR /work` (owned by `proven`). The buildx requirement matches `image.sh`'s refusal. Not executed: UNRUN.

### Remainder
Owed, none blocking (R6):
- O1 `docs/INSTALL.md:48`: the hand check hard-codes `saphan/proven:latest`. A reader who set `PROVEN_IMAGE` for `make image` checks the wrong tag, and the check fails with "the image's --version exited 125", which is red, not green. Remedy: one clause, "use `$PROVEN_IMAGE:latest` if you set PROVEN_IMAGE".
- O2 `scripts/image-check.sh:7`: the usage example `proven:local` names an image that, built without `make image`, carries revision `unset` and FAILS in a git checkout. Remedy: make the example `saphan/proven:latest`, or note the label requirement.
- O3 the revision label is HEAD even for a dirty tree (`lib.sh oci_revision`). A hand check run after an uncommitted edit, without rebuilding, passes the label check and relies only on output equality. The docs already say "the binary is compared by its output only". `make image-check` rebuilds both, so this does not affect the make path. Remedy, if wanted: also compare the `org.opencontainers.image.version` label (`git describe --dirty`).
- O4 `:71` uses `head` as a variable name and `:79` uses the `head` command. They do not collide in sh, but the reuse is confusing. Cosmetic.

### Profile
docs (D4) + correctness of a check script. No security profile: the diff adds no auth, secret or destructive path. `rm -rf "$work"` acts on a `mktemp -d` path only.
Testing-bar skill: the project declares no metric for shell scripts. The check ranks as a *suite* check (the project's own binary against its own image), and it is reported as a NOTE.

### Focus
The four properties the order names, all from the diff. On the evidence I could read, properties 1–3 hold and property 4 holds by reading. I declined to judge (R11): the Dockerfile itself (unchanged), image size or contents, the trace file's own contents (not compared by design), and licences (out of scope by the order).
Note for the master: the order asks for the H1 `# REVIEW — the-image-review-claude`. The engine-injected channel grammar requires `# REVIEW — the-image` (the reviewed stream, per law #29) and says to trust it over a brief. I followed the grammar.

### Conditions
- C1: before the gate, the master's docker measurements from the order stand as the only executed evidence: `make image-check` ok, the fixture-mutation negative control → FAIL, `PROVEN_IMAGE=x/proven` → FAIL 125, and a commit after the build → FAIL naming both revisions. The master pastes those outputs into the gate note, because this review could not run any of them (see Unrun).
- O1–O4 under Remainder are owed work. They are not conditions of this verdict.

### Unrun
- Every execution of `scripts/image-check.sh` and `make image-check`: docker is not in this seat's allow list (permission class `██████████████████████████`: `[saphan doc check go git ls cat grep find wc head tail diff pwd cut sort]`).
- The negative controls (fixture mutation, missing image, stale label): these need docker, so they were not run.
- `normalise` on a real sample output: `awk` is not admitted, and running the binary is not admitted (`mkdir` was refused with "do not rebuild the act from other commands").
- The docs commands (D4): they need docker.

## The review status as filed

## REVIEW — the-image
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: c96e18f read in full: the script cannot report ok when either side's version or sample review fails or differs; normalisation is fixed-string fixture + last-line trace only; make image-check checks the tag make image builds against HEAD's label. Docker was denied on this seat, so nothing was executed (Unrun non-empty, law #53, no accept). Condition: the master's docker controls are pasted into the gate note. Owed: 4 small docs/comment nits.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T02:00:00Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 3 |
| policy binding name | 1 |
