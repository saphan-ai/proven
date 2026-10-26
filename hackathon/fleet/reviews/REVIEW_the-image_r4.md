# REVIEW — the-image · round 4

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 3 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-25T00:53:45Z

## Summary

By reading 078bc85, the script never prints ok when either side fails or differs: both sides must exit 0, cmp is byte-exact, --pull=never, label must equal HEAD.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — the-image

### Ground
- Reviewed sha: `git rev-parse --short=7 HEAD` printed `078bc85` (ROLE-LOCK satisfied). Merge base with main: `2b8fd4a34666b6fe5406bf7e1af21932251f2582`.
- Scope: `git diff 2b8fd4a...078bc85`. It changes four files: `scripts/image-check.sh` (new, 69 lines), `lab/scripts/image-check.sh` (new, 8 lines), `Makefile` (an `image-check` target plus its `.PHONY` entry), and `docs/INSTALL.md` (the Docker table row and a new "The container image" section). The order names three files. The fourth, `lab/scripts/image-check.sh`, is the make wrapper, and I judged it as part of item 3.
- Read for context, not changed by the diff: `lab/██████████████`, `lab/████████████████`, `lab/████████████████`, `lab/versions.env`, `Dockerfile`, `.dockerignore`, `internal/cli/format.go` (the text writer that prints the `trace:` line), `internal/cli/run.go:19-20` (the version line), `docs/SAMPLE_MODE.md` (the documented sample output).
- My seat ran no docker, make, sh or proven. The permission class `██████████████████████████` allows only `saphan doc check go git ls cat grep find wc head tail diff pwd cut sort`. Every runtime claim below therefore comes from reading the code (see Unrun). The master's Linux docker measurements from 2026-10-25 are claims I could not re-measure. I set them aside per R1.
- Channel shout (law #7): the order's "Review output" section asks for H1 `# REVIEW — the-image-review-claude`. The engine's channel grammar says the H1 must name the reviewed stream, `# REVIEW — the-image`, and that the engine's line wins. I wrote the engine's H1. The order template has a defect: an H1 written that way would file the verdict where `saphan merge` never looks.

### Differential
Property 1. The script must never print ok when either side fails or when the outputs differ in anything but the two normalised paths. Each possible false world, with what the script answers:

| False world | What the script does | Line |
|---|---|---|
| `./bin/proven` is missing or broken, so `--version` exits non-zero | FAIL "the binary's --version exited N". `$?` in `a \|\| fail "...$?"` expands after `a` exits, so it carries `a`'s status | scripts/image-check.sh:38-39 |
| The image is missing, or the tag names an image not present locally | FAIL "the image's --version exited 125". `--pull=never` stops a registry image from standing in for the local one | :40-41 |
| The image's entrypoint is wrong or prints nothing, with exit 0 | FAIL "the version lines differ", because the binary side anchors the comparison with `proven development` | :42-43 |
| Either sample review exits non-zero, including two runs that break the same way | FAIL "...sample review exited N". Both sides must exit 0 before any comparison | :47-50 |
| Both reviews exit 0 but any line other than the trace line differs (a finding, a statement, an extra stderr line) | FAIL "the sample reviews differ". `cmp -s` compares bytes, with stderr merged in | :51-54 |
| The image was built from another commit, or an older `:latest` from another worktree is in place (in a git checkout) | FAIL "the image was built from X, this checkout is Y" | :58-61 |
| An image with no revision label (in a git checkout) | FAIL, because the empty label does not equal HEAD | :59-60 |
| A normaliser `sed` errors, for example a checkout path that contains `#` | Exits non-zero through `set -e`: the pipeline's status is `sed`'s. It prints no "FAIL" sentence, only sed's own error. Fails closed | :51 |

I found no path to ok while either side fails. One more read of R3 (mutate the guards): deleting any single `|| fail` still does not reach ok, because `set -eu` makes the bare failing command end the script with a non-zero status. The `|| fail` clauses add the message, and `set -e` is the backstop. The first round was falsely green because two failing runs printed the same output. That world is now closed by the exit-0 requirement at :47-50.

The version comparison cannot tell builds apart (law #52). `versionLine` is the constant `"proven development\n"` (internal/cli/run.go:20), so every build of every commit prints the same line. The version comparison therefore proves only that the entrypoint is `proven` and that it runs. The script says so itself (:56) and closes the gap with the revision label. That is correct, and I raise no finding. The output line `image-check: version proven development` (:68) carries no build identity, and a reader should not take it as one.

### Unplug
- Remove the exit-0 requirement, so two identically broken runs get compared: the master's macOS-tar run of 2026-10-25 shows this world turned green. In the current code a non-zero exit fails at :47-50 before any comparison. Measured by reading.
- Remove the revision check (:58-64): a stale `saphan/proven:latest` whose sample output matches would pass. Under `make image-check` the `image` prerequisite rebuilds the image first, so the check matters by hand and when two checkouts share the tag. The master measured "a commit after the build → FAIL naming both revisions" (claim, not re-measured).
- Remove `--pull=never`: a missing local tag would be pulled from Docker Hub (`saphan/proven` is the published name, `PROVEN_REGISTRY=docker.io`), and an image from the registry would be checked in place of the one just built. With the flag, the master measured `PROVEN_IMAGE=x/proven` → FAIL 125 (claim, not re-measured).
- Nothing turns red if `make image-check` is never run. The check is opt-in and wired into no CI stage. This is outside the order's question. I report it under Remainder and do not hold it against the diff.

### Closure
Property 2. The normalisation hides only the trace path and the fixture path.
- Trace: `sed -E 's#^trace: .*$#trace: <trace path>#'` (:28) rewrites only a line that begins with `trace: `. The text writer prints that line exactly once, last, as `trace: <TracePath>` (internal/cli/format.go:119-121). The two runs genuinely differ there: the binary writes `$work/bin/logs/trace.jsonl` and the image writes `/work/logs/trace.jsonl`. The family sentence that is read back from the trace (format.go:114-118) is not rewritten, so the trace's content is still compared. Residue, non-blocking: a column-0 statement line from the review that began with `trace: ` would also be blanked. In sample mode every such line is a fixed recorded string, so no real difference can ride on it today.
- Fixture: the host side replaces `$fixture` (:51) and the image side replaces `/opt/proven/fixture/discount` (:52), each only in its own output. `$fixture` is spliced into a BRE unescaped. A `.` in the checkout path matches any one character, which could only blank a string that is the fixture path itself with one character changed. `[` or `*` would make the replacement miss, giving a false FAIL. `#` errors the pipeline, failing closed (see Differential). None of these can hide a real difference in the review's content.
- Not normalised, and correctly left alone: the state directory, timestamps and stderr. If the image's run printed `$work`, a platform warning or anything else, the check would go red rather than green.
- Verdict on property 2: holds.

Property 3. `make image-check` checks what it built.
- `image-check: build image` (Makefile:24) depends on both producers. `build.sh` writes `$ROOT/bin/proven` (build.sh:20 with Go, :38 through docker using host GOOS/GOARCH). `image.sh` tags `$PROVEN_IMAGE:latest` (image.sh:33). `lab/scripts/image-check.sh` sources the same `lib.sh`, whose `load_versions` keeps a `PROVEN_IMAGE` already in the environment. It passes `$PROVEN_IMAGE:latest` and `$ROOT/bin/proven` (lab/scripts/image-check.sh:6-8), the same names the producers wrote. The inner script derives its own root from `$0` (scripts/image-check.sh:18), which is the same checkout. `image.sh` labels the image with `git -C "$ROOT" rev-parse HEAD` (lib.sh `oci_revision`), and the check compares against the same command (:58). Verdict: holds.
- Residue, non-blocking: `oci_revision` has no `--dirty`. On a dirty tree the label still says HEAD while the image holds the working tree. The binary is built from that same working tree in the same make run, so behaviour is still compared. The printed sentence "the image's label and this checkout agree" (:61) is true of the label only.

Property 4. The docs commands work as written. Two defects, both owed (see Conditions):
- docs/INSTALL.md:46-48: `sh scripts/image-check.sh saphan/proven:latest ./bin/proven` uses two relative paths that only resolve from the repository root. The section never says so, and the section just above it (INSTALL.md:31-34) leaves the reader in `/tmp/proven-run`. Read top to bottom, the command fails with "No such file". The next section already uses the form this one needs: "From the repository:" (INSTALL.md:60).
- docs/INSTALL.md:44: "The same check by hand, for any image and binary". In a git checkout the script refuses every image whose revision label is not this checkout's HEAD (:58-60), so the check is not for *any* image. The sentence promises what the script deliberately refuses. It should say that the image must be built from this checkout's HEAD, and that the binary is compared by output only, since its version line names no build.
- Fine as written: `docker run --rm saphan/proven:latest scan --sample /opt/proven/fixture/discount` (INSTALL.md:52-54) matches the Dockerfile's `COPY --chown=proven:proven testdata/fixture/discount /opt/proven/fixture/discount`, `WORKDIR /work` (owned by `proven`) and `ENTRYPOINT ["proven"]`. It has no `--pull=never`, so on a host without a local build it would pull the published image. That is acceptable in a section that begins with `make image`. The prerequisites row (INSTALL.md:14) is accurate.

### Remainder
- The check compares the image with the binary. It does not check that the sample review is *correct*. A defect shared by the source of both artefacts (a lost finding) passes as ok. That is the right scope for an image-versus-binary check. SAMPLE_MODE's scenario tests hold the correctness bar, not this script.
- By hand, a stale `./bin/proven` from another commit with identical sample output passes, because the binary has no identity to compare (the constant version line). Under make the `build` prerequisite makes it fresh. This lifts when the version line is stamped at build time (run.go:19 already plans it).
- Running `docker run` without `-t` sends stdout and stderr down separate streams. If a sample scan ever writes to both, the interleaving in the merged capture may differ from the host's and give a false FAIL. It can never give a false ok.
- An image for a foreign platform, checked by hand, makes the docker CLI print a platform warning into the capture, giving a false FAIL. `make image` builds for the server's own platform, so the make path does not meet this.
- A source tree unpacked inside some other git repository makes `git -C "$root" rev-parse HEAD` read the enclosing repository's HEAD. `make image` labels the image with that same HEAD, so the check passes and says "agree" about an unrelated commit. This happens only on an unusual layout and is never a false ok on the output comparison.
- The check is not run by any CI stage, so nothing reddens if it stops being run (law #41). I report this and do not judge it. CI is outside this order.

### Profile
docs + build tooling (a shell script, a make target, and the docs). Not the security profile: the diff adds no grant, secret, network listener or authorization path. D4 (re-run every quoted command) could not be done on this seat. See Unrun.

### Focus
The four properties the order names, in order. In particular, which false world the script answers with ok (none found), and whether the normaliser can hide a real difference (it cannot, beyond the residues named in Closure). I judged nothing else. Declined to judge (R11): CI wiring of the check, the Dockerfile's contents beyond what the script and docs rely on, the image's size budget, and licences (out of scope).

### Conditions
1. OWED (docs), docs/INSTALL.md:44-48. Make the hand command runnable as written. Either add the starting directory ("From the repository:", the form INSTALL.md:60 already uses, or `cd` back to the clone), or use paths that do not depend on the current directory. Today a reader who follows the page top to bottom is in `/tmp/proven-run` and the command fails.
2. OWED (docs), docs/INSTALL.md:44. Replace "for any image and binary" with what the script enforces: in a git checkout the image must carry this checkout's HEAD as its `org.opencontainers.image.revision` label, or the check fails naming both revisions, and the binary is compared by its output only.
3. OWED (verification, master), before the gate. The runtime claims are unmeasured on this seat (see Unrun). They are the master's Linux docker measurements of 2026-10-25: the ok run, the one-line fixture negative control, `PROVEN_IMAGE=x/proven` → 125, and a commit after the build → revision FAIL. Each run's output should be attached to the record if it is not there already. Reading supports each claim, but reading is not a measurement.
Non-blocking notes (R6): anchor the trace normaliser to the last line only (:28). Escape `$fixture` before splicing it into the `sed` pattern (:51), or compare with a fixed-string substitution.

### Unrun
- `make image-check` (ok path): not run. `make` and `docker` are not in the review class's command list (`██████████████████████████`: saphan doc check go git ls cat grep find wc head tail diff pwd cut sort). The same class refused `echo` with "do not rebuild the act from other commands", which binds `make` and `docker` too.
- The negative controls (fixture changed in the image only, `PROVEN_IMAGE` naming a missing image, a commit after the build): not run, same reason.
- `sh scripts/image-check.sh saphan/proven:latest ./bin/proven` and `docker run --rm saphan/proven:latest scan --sample /opt/proven/fixture/discount` (D4, the docs commands re-run): not run, same reason (`sh`, `docker` not admitted).
- `sh -n scripts/image-check.sh` (syntax check): not run. `sh` is not admitted.
- Every judgement above about runtime behaviour comes from reading the code, with the line cited.

## The review status as filed

## REVIEW — the-image
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: By reading 078bc85, the script never prints ok when either side fails or differs: both sides must exit 0, cmp is byte-exact, --pull=never, label must equal HEAD.
  The normaliser hides only the trace line and the fixture path. make image-check checks the tag and binary it builds.
  OWED (docs): INSTALL.md:46-48 uses relative paths after the page cds to /tmp/proven-run. INSTALL.md:44 says "any image", but the revision check refuses other commits.
  Unrun: docker, make and sh are denied on this seat, so every runtime claim is the master's 2026-10-25 measurement, not re-measured (condition 3).
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-25T01:10:00Z

## REDACTED

| what | times |
|---|---|
| file and line in Saphan Studio's private code | 3 |
| policy binding name | 2 |
