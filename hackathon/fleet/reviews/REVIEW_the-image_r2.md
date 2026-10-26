# REVIEW — the-image · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T00:37:49Z

## Summary

Read-only review of b4769f8 (docker and sh refused on this seat, so Unrun is not empty). The script fails closed whenever either side's review fails or the outputs differ.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-image

Reviewer: claude-code (Opus 5.5), REVIEW leg `the-image-review-r3`, 2026-10-25.
`git rev-parse --short=7 HEAD` printed `b4769f8`. Merge base with `main` (from `git merge-base main HEAD`): `2b8fd4a34666b6fe5406bf7e1af21932251f2582`.
Scope: `git diff 2b8fd4a…b4769f8`: `scripts/image-check.sh` (new, 54 lines), `Makefile` (+`image-check`, +.PHONY), `docs/INSTALL.md` (+"The container image").
This seat ran no script: the permission class `██████████████████████████` refuses both `docker` and `sh`/`env` (refusals quoted under Unrun). Everything below comes from reading the code. Per R1, I set the master's measurements aside as claims, not evidence.

### Ground

Callers and guards read for every finding:
- `scripts/image-check.sh:1-54`, the whole file.
- `Makefile:15-25`: `build`, `image`, `image-check`.
- `lab/█████████████████████`: tags `$PROVEN_IMAGE:latest` and `$PROVEN_IMAGE:$version`, build context `$ROOT`.
- `lab/██████████████████████████████`: `load_versions` runs when the file is sourced, and **a variable already set in the environment wins** (`lib.sh:35-38`).
- `lab/████████████████`: `go build -o $ROOT/bin/proven ./cmd/proven`.
- `Dockerfile`: builds from `COPY . .` with `.dockerignore`, copies `testdata/fixture/discount` to `/opt/proven/fixture/discount`, `WORKDIR /work`, `ENTRYPOINT ["proven"]`.
- `.dockerignore`.
- `internal/cli/run.go:19-20,45-46`: the version line.
- `docs/SAMPLE_MODE.md:55-97`: the sample output and the trace line.
- `docs/INSTALL.md:1-56`.

### Differential

Property 1 asks what the script answers in each false world (read, not run):

| False world | Script's answer | Where |
|---|---|---|
| Binary's sample review exits ≠0 | FAIL "the binary's sample review exited N" | `image-check.sh:44-45`. `$?` after `||` is the subshell's status, correct |
| Image's sample review exits ≠0 | FAIL "the image's sample review exited N" | `:46-47` |
| Both fail identically (the AppleDouble case) | FAIL at the binary leg first | `:44`, which requires success on each side separately |
| Reviews differ in any byte outside the normalised spans | FAIL "the sample reviews differ" | `:48-51`, `cmp -s` on whole files |
| Version lines differ | FAIL | `:38-40` |
| Docker absent, or the image is missing and not pullable | FAIL "the version lines differ", a misleading cause but fail-closed | `:37`, `|| true` swallows the exit status |
| Image is missing locally but pullable | FAIL at the version step, because the pull chatter goes to stderr, which is captured (`2>&1`) | `:37` |
| **Image built from a DIFFERENT commit, sample behaviour unchanged** | **ok**. The version line is the constant `proven development\n` (`internal/cli/run.go:20`) for every build, so the version step cannot tell commits apart | `:36-40` → finding C2 |
| **`make image-check` with `PROVEN_IMAGE` set in the environment** | **ok against a stale local `saphan/proven:latest`**. The image just built is never examined | `Makefile:25` → finding C1 |

The script's internal comparison is fail-closed in every world where one of the two runs is broken. Both "ok" rows are cases where the script compares against the wrong subject.

### Unplug

Mutations read rather than run, because `sh` is refused:
- Drop `|| fail` at `:45` or at `:47`: a broken side is reported only if its output differs. The AppleDouble world (both broken identically) then turns green. That is exactly the regression the master saw, so the guard is load-bearing.
- Drop the `cmp` at `:49`: any difference turns green.
- The master's recorded negative control (image fixture changed by one line → "the image's sample review exited 2") exercises `:46-47` only. It does not exercise the `cmp` at `:49-50`, because that mutant never reaches the comparison. **No recorded control shows `:49-50` firing.**
- Remedy (owed): a negative control where both runs exit 0 but one byte of output differs. Example: a fixture change that alters a finding's message but not the exit code.

### Closure

Property 2 (normalisation):
- `normalise` at `:27` is `s#(trace[^:]*: ).*#\1<trace path>#`. **The pattern has no anchor**, so it rewrites the rest of ANY line that contains `trace`, followed by non-colon characters, then `: `. Examples: a finding body with "stack trace: …", or "retrace…: …". In today's discount output only the final `trace: …/logs/trace.jsonl` line matches (`SAMPLE_MODE.md:93`; no other `trace` occurs in the output block). So in practice it hides only the path. But it is not structurally confined to that line, and a future finding text containing `trace…: ` would have its difference hidden. Finding C3.
- The fixture substitution at `:48` puts `$fixture` into a sed regex without escaping it. A `.` in the path over-matches only same-length lookalikes. A `#` in the path makes sed error, `set -e` exits non-zero, and the script fails closed. Acceptable.
- The image side substitutes only `/opt/proven/fixture/discount` and the binary side only `$fixture`, so neither side's substitution can make the other side's text match. Correct.
- The comment at `:25-26` says the trace path "names a per-user cache directory". The code and the docs disagree: the trace is `./logs/trace.jsonl` in the start directory (`SAMPLE_MODE.md:22,96`), i.e. `$work/bin/logs/…` against `/work/logs/…`. The comment is wrong and the code is right. Owed.

### Remainder

Owed, non-blocking (R6):
- O1 `:36-37`: the version step ignores both exit statuses (`|| true`), so a missing docker reads as "the version lines differ". Require exit 0 on each side as `:44-47` already do, and name the failing side.
- O2 `:54`: a failing `docker image inspect` inside `$(…)` does not trip `set -e`. The script prints `size  bytes` and exits 0 after already printing ok. Capture it into a variable first, so a failure exits non-zero.
- O3 `:25-26`: the wrong comment described under Closure.
- O4 `:1-5` and the ok line at `:52`: they claim the image "runs the binary it claims to carry … built from the same commit". C2 shows that the check by hand cannot establish this.

Declined to judge (R11), for the master to rule on: the `Dockerfile` build flags against the `build.sh` flags (`-trimpath -s -w` against none) and the Go version skew between the host and `GOLANG_IMAGE`. These do not change the sample output's bytes as read, and the Dockerfile is outside this diff.

### Profile

Script and build tooling, plus docs (D4: the commands are judged under Focus). No security profile: the diff touches no authentication, grant or egress code. R13 paired paths: the image half is `lab/████████████████` and the binary half is `lab/████████████████`. I read both, and their settings source is the cause of C1.

### Focus

1. **Property 1 holds for broken sides, and fails for a wrong subject.** See the Differential table. Every broken-side world is fail-closed. The two green false worlds are C1 and C2.
2. **Property 2 holds on today's fixture output, and is not structurally confined.** See C3.
3. **Property 3 holds only when `PROVEN_IMAGE` is unset.** `Makefile:25` runs `. $(ROOT)/lab/versions.env`, a plain `KEY=VALUE` source that **overwrites** the environment. `make image` → `image.sh` → `lib.sh:127 load_versions` **keeps** an environment value (`lib.sh:35-38`, documented at `versions.env:2`). So `PROVEN_IMAGE=me/proven make image-check` builds and tags `me/proven:latest`, then checks `saphan/proven:latest`. If a stale `saphan/proven:latest` exists locally, the check reports ok about an image it did not build. This is the only recipe in the Makefile that sources `versions.env` itself; every other consumer goes through `lib.sh`. The binary half is correct: `build.sh` writes `$(ROOT)/bin/proven` from the same checkout, and the recipe checks that path.
4. **Docs commands, read only.** `docs/INSTALL.md` runs `sh scripts/image-check.sh saphan/proven:latest ./bin/proven` with relative paths. A reader following the page top-down is in `/tmp/proven-run` (`INSTALL.md:32`), and the page's own Build step produced `/tmp/proven`, not `./bin/proven` (`INSTALL.md:23`). Taken in order, the command fails: it fails closed, since `cd "$(dirname ./bin/proven)"` errors, but it does not work as written. The `docker run --rm saphan/proven:latest scan --sample /opt/proven/fixture/discount` line agrees with the Dockerfile (fixture path, `ENTRYPOINT proven`, writable `WORKDIR /work` owned by uid 65532). The Prerequisites table (`INSTALL.md:14`) still says Docker is for "live mode only", and `make image` additionally needs buildx (`image.sh:7-9`). Finding C4.

### Conditions

- **C1 (blocking class: a wrong result silently trusted), `Makefile:25`.**
  - Trigger: `PROVEN_IMAGE` set in the environment while a stale local `saphan/proven:latest` exists.
  - Remedy: resolve the tag the way `image.sh` does. Either add a `lab/scripts/image-check.sh` wrapper that sources `lib.sh` and calls `scripts/image-check.sh "$PROVEN_IMAGE:latest" "$ROOT/bin/proven"`, or keep an environment value in the recipe.
  - Also pass `--pull=never` to both `docker run` calls in `scripts/image-check.sh`, so a missing local image is refused by name instead of being fetched.
  - Control: with `PROVEN_IMAGE=x/proven` and an older `saphan/proven:latest` present, the target must examine `x/proven:latest`.
- **C2 (owed, order-adjacent), `scripts/image-check.sh:36-40`.**
  - Defect: the version line is constant (`run.go:20`), so the "same commit" claim rests on the sample review alone.
  - Remedy: compare `docker image inspect --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'` against `git -C "$root" rev-parse HEAD`, or drop "same commit" from the header (`:1-5`) and the docs, and add a `remark:` line saying the version step cannot tell builds apart until releases are stamped.
  - Inside `make image-check` the image is rebuilt from the checkout, so C2 bites only the check by hand "for any image and binary".
- **C3 (owed), `scripts/image-check.sh:27`.** Anchor the normaliser to the one line it means: `s#^trace: .*#trace: <trace path>#`.
- **C4 (owed), `docs/INSTALL.md` "The container image".**
  - State "from the repository root, after `make build`" (or give absolute forms).
  - Note that Docker with buildx is needed for this section.
- **C5 (owed), control.** Record one negative control where both runs exit 0 and the outputs differ, to show `:49-50` firing (see Unplug).
- O1–O4 under Remainder are owed and never reopen a round (R6).

### Unrun

Every runtime check is unrun on this seat. Refusals, verbatim:
- `docker version --format '{{.Server.Version}}'` → `the program "docker" is NOT listed in permission class ██████████████████████████ ([saphan doc check go git ls cat grep find wc head tail diff pwd cut sort]) … allowed form: none — "docker" is denied`.
- `env PROVEN_IMAGE=me/proven sh -c '…'` (the C1 precedence probe, docker-free) → `the program "env" is NOT in the ██████████████ list of role review … deny by default … allowed form: none`.

Not run, therefore:
- the image-against-binary comparison;
- `make image-check`;
- the master's one-line-fixture negative control;
- the C1 precedence probe (established by reading `lib.sh:35-38` against `Makefile:25`);
- the R3 mutants of `:45`, `:47` and `:49`;
- the two docs commands;
- the 97 MB size claim.

Per law #53 this review therefore does not return `accept`. The master should run the C1 probe and the C5 control on the Linux docker host before the gate.

## The review status as filed

## REVIEW — the-image
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: Read-only review of b4769f8 (docker and sh refused on this seat, so Unrun is not empty). The script fails closed whenever either side's review fails or the outputs differ.
  Blocking condition C1: Makefile:25 sources versions.env, which overrides an environment PROVEN_IMAGE, while make image (lib.sh) keeps it. The target can then check a stale saphan/proven:latest instead of the image it just built.
  Owed: C2, the version line is constant ("proven development"), so it cannot tell commits apart. C3, anchor the trace normaliser. C4, the docs command assumes the repo root and ./bin/proven. C5, add a control that reaches the cmp.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T00:00:00Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 5 |
| policy binding name | 3 |
