# REVIEW — corpus-pv-020 · round 2

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T19:05:10Z

## Summary

PV-020 at 8c55cee, read in full: base and change differ only on the four non-exempt routes above 2048 bytes; runs (b) and (c) each remove exactly one precondition; no check bent; diff in lane; nothing private.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-020

Reviewer: claude-code, model `claude-opus-5-5`. Scope: `fe146e0...8c55cee`, one commit
(`[corpus-pv-020] tests: build the size-limit case and record the check`). Measured 2026-10-22.

### Ground

- `git rev-parse --short=7 HEAD` printed `8c55cee` (role-lock check passed).
- `git diff --stat fe146e0...8c55cee`: 12 files, 772 insertions, 4 deletions. Every path is under
  `tests/corpus/cases/PV-020/` except `tests/corpus/manifest.json` (one line, line 358, inside the
  `"id": "PV-020"` entry that opens at line 347).
- `git diff --name-only fe146e0...8c55cee -- tests/corpus/cases/PV-020/expected` printed nothing: `expected/` is untouched.
  This confirms the master's claim (macm5, 2026-10-22) that the diff stays out of `expected/`.
- I read every file in the diff: `base/service/service.go`, `base/██████████████████`,
  `change/service/service.go`, `change/service/controls.go`, `change/██████████████████`,
  `change/████████████████████`, `check.sh`, `description/api.json`, the CASE.md hunks and the manifest hunk.
  I read `expected/EXPECTED.md` only after I had judged the fixture against §2.
- `go -C tests/corpus/cases/PV-020/change vet ./...` → no output, exit 0.
  `go -C tests/corpus/cases/PV-020/base vet ./...` → no output, exit 0. Toolchain on this seat: `go version go1.26.5 linux/amd64`.
  So both modules compile and vet clean on Go 1.26.5. CASE.md §8 records go1.26.8.
- **`check.sh` was NOT run by me.** The permission class refused it (quoted under Unrun). The CASE.md §8 output
  pasted below is the author's recording, not mine. Under R1 it is not my evidence.

### Differential

How base and change differ, read from the code. This is not the check's output.

- `base/service/service.go:36-42`: `Decide` accepts any body on a known route and refuses an unknown route. There is no limit.
  This matches §2 ("base has no size limit").
- `change/service/service.go:70-84`: `Decide` refuses `len(body) > 2048` on every non-exempt route, and also on an
  unknown route. It accepts above the limit on the two `Exempt: true` routes, with reason `exempt upload`. A body of
  exactly 2048 bytes is accepted. This matches §2: refused on every route except the exempt list, and accepted there.
- The route sets are the same six routes in the same order in base and change. `check.sh` lines 48-75 refuse a route
  mismatch or a length mismatch with exit 1, so the per-route lines cannot drift out of step without the check failing.
- Where base and change disagree for an oversize body (2049 bytes): GET /v1/health, POST /v1/items, POST /v1/search and
  PUT /v1/items/current go from accepted to refused. The two upload routes stay accepted (the reason changes from
  `accepted` to `exempt upload`). Nowhere else. That is exactly where §1/§2 place the change.
- Runs (a)/(b)/(c) in `check.sh` lines 33-43 change one input each:
  (b) passes a configured path that does not exist (`description/not-provisioned.json`), and `loadDescription`
  (`controls.go`) maps `os.IsNotExist` to the contract text `description not provisioned`;
  (c) sets `PV020_SIMULATE_LISTEN_REFUSAL=1`, so `listenLocal` returns an EPERM `*net.OpError` before any bind, and
  `RunListener` prints `could not run: operation not permitted`. The service passed to the listener is still
  `Production()`, so the limit is not weakened. This matches §2 (c), "never simulated by weakening the limit".
- Limit-removed run: `WithoutLimit()` sets `maxBody: 0`. Read against `RunInProcess`, the failing set must be
  `inproc.limit-present`, four `inproc.over-limit-refused` and two `inproc.over-limit-exempt` (the reason drops to
  `accepted`). `inproc.at-limit` and `inproc.exempt-set` must still pass. The author's recorded line in CASE.md §8 is exactly
  that set, so the recording agrees with the code. That agreement comes from reading the code, not from running it.

Author's recorded output (CASE.md §8; NOT re-measured by me, see Unrun):

    run a description-present-listener-allowed control in-process result passed
    run a description-present-listener-allowed control description result passed
    run a description-present-listener-allowed control listener result passed
    run b description-absent control in-process result passed
    run b description-absent control description result could not run: description not provisioned
    run b description-absent control listener result passed
    run c listener-refused control in-process result passed
    run c listener-refused control description result passed
    run c listener-refused control listener result could not run: operation not permitted
    route GET /v1/health oversize base accepted (accepted) change refused (body exceeds limit)
    route POST /v1/items oversize base accepted (accepted) change refused (body exceeds limit)
    route POST /v1/search oversize base accepted (accepted) change refused (body exceeds limit)
    route PUT /v1/items/current oversize base accepted (accepted) change refused (body exceeds limit)
    route POST /v1/uploads/media oversize base accepted (accepted) change accepted (exempt upload)
    route POST /v1/uploads/bundle oversize base accepted (accepted) change accepted (exempt upload)
    run limit-removed control in-process result failed: inproc.limit-present, inproc.over-limit-refused GET /v1/health, inproc.over-limit-refused POST /v1/items, inproc.over-limit-refused POST /v1/search, inproc.over-limit-refused PUT /v1/items/current, inproc.over-limit-exempt POST /v1/uploads/media, inproc.over-limit-exempt POST /v1/uploads/bundle

### Unplug

What each control would print if the thing it guards were gone or wrong. I worked this out by reading the code; I did not mutate and run it.

- Limit removed → the limit-removed line turns `failed` and names seven controls. This is in the check itself (§2's
  "controls notice its absence"). It is a real negative control.
- Off-by-one (refusing at 2048) → `inproc.at-limit` fails for the four non-exempt routes (`controls.go`, `inprocAtLimitAccepted`).
- A non-exempt route wrongly flagged `Exempt: true` → `inproc.exempt-set` fails (the flags are compared with the separate
  `DocumentedExempt` list), the description control fails `description.limits`, and that route's line reads
  `change accepted (exempt upload)`.
- Description file absent → run (b) `could not run: description not provisioned`. Description present but different →
  `failed: description.routes` / `description.limits`. These are separate outputs, so absence is not confused with disagreement.
- Listener refused → run (c) `could not run: operation not permitted`. A refusal other than EPERM is printed with its own
  error text and is not relabelled. A listener that starts while the handler is broken → `failed: listener.…`.
- **What cannot fire:** a defect in `Handler` on any route other than POST /v1/items and POST /v1/uploads/media. The listener
  control only streams to those two routes, and the route census uses `Decide`, not HTTP. See Remainder R-1 for the one
  HTTP/`Decide` disagreement this leaves unseen.

### Closure

Each acceptance item, as a property that must hold:

1. **The fixture reproduces the condition §2 names: HOLDS by reading, UNRUN by measurement.** By code reading, base differs from
   change exactly on the four non-exempt routes above the limit, and nowhere else. Runs (b) and (c) each remove exactly one
   precondition. I could not run `check.sh` and paste its output, as the order requires (see Unrun). Condition C1.
2. **The expected outcome follows from the evidence: HOLDS, with a stated reach.** EXPECTED.md says the class is "human scope
   decision — no defect in change". The evidence it requires is: the in-process controls green, red with the limit removed,
   (a) all green, (b)/(c) one control each `could not run` with its reason. The recorded lines carry exactly that.
   False-world answer (law #52): if change had a defect in `Decide`'s limit or in the exempt set, the output would differ
   (Unplug). If change had a defect only in the HTTP path on an untested route, the output would be **identical**. So the
   check establishes "the controls ran or could not run, as listed". It does **not** establish "no defect". "No defect" rests
   on §2's correct-by-construction clause and on reading the code. One divergence weakens that clause (R-1). Condition C3.
   Small wording difference: EXPECTED.md:7 names the (c) reason "listener refused". The check prints
   "operation not permitted", which is §2's own text (CASE.md:42). It does not change the outcome. NOTE only; `expected/`
   is outside this diff.
3. **The case still proves something; no check bent: HOLDS.** The refusal is simulated in `listenLocal` and never in the limit.
   The limit-removed run uses a separate constructor, and the shipped `Production()` is unchanged. The description controls
   tell "absent" apart from "different". I found no assertion loosened to make a line come out "as expected". Testing-bar
   ranking: every control here is a **suite** check written by the same author as the service. That is circular by nature
   (NOTE). The base/change route census is the only line that contrasts two independent versions.
4. **Nothing private is left: HOLDS.** I scanned the diff (`git diff -U0 | grep -niE '@|/home|/users|macm|saphan|token|secret|
   password|wato|192\.168|10\.[0-9]+\.|\.local|ssh|seat|bridge'`). The only hits are hunk headers and `check.sh:10`
   `$repo/.saphan/work/corpus-pv-020`, a relative work directory named after the public product. There is no host, seat,
   private path, e-mail address or token. CASE.md §3 names reviewer models (`gpt-5.6-sol`, `glm-5.2`). That text sits in
   unchanged context and is not part of this diff.
5. **The diff stays in its lane: HOLDS.** See Ground: only `tests/corpus/cases/PV-020/` (outside `expected/`) and manifest line 358
   inside the PV-020 entry.
6. **The check runs where a judge runs it: PARTLY.** It runs on Linux by reading; macOS is UNRUN. Every non-portable or
   environment-bound call:
   - `check.sh:9` `git -C "$case_dir" rev-parse --show-toplevel`: fails outside a git checkout (for example a GitHub
     "Download ZIP" or a release tarball), and `set -e` then aborts the script.
   - `check.sh:10-11`: writes three binaries under `<repo>/.saphan/work/corpus-pv-020/`. `.gitignore` does not ignore
     `████████`, so a judge's `git status` shows an untracked directory after one run.
   - `check.sh:13` with `go.mod` `go 1.26`: needs a local Go of 1.26 or later. With `GOPROXY=off`, an older Go cannot fetch the
     toolchain the module asks for, and the build fails. `go -C` (lines 17-19) needs Go 1.20 or later.
   - Bash constructs (`local`, `${v%%|*}`, `exec 3<`, `read -r <&3`, `set -o pipefail`) all exist in macOS's bash 3.2.
     `syscall.EPERM` exists on darwin. I did not verify whether the macOS application firewall prompts for a loopback bind
     by an unsigned binary.

### Remainder

- **R-1 · OWED · `change/service/service.go:97`, `:111`, `:132`.** Over HTTP, an exempt upload is read through
  `MaxBytesReader(w, rc, readCeiling)` with `readCeiling = 8 << 20`. Above 8 MiB the handler answers
  `400 body unreadable`. `Decide` (`:70-84`) accepts an exempt body of any size. §2 says change "accepts it there", and
  EXPECTED.md says "must NOT be claimed: a defect in the limit (none exists)". A reviewer under evaluation who reads both
  paths (R13, paired paths) can truthfully report that the exempt routes are not unlimited over HTTP. The answer key would
  then grade a true observation as a false claim. Trigger: POST of 8 MiB + 1 byte to /v1/uploads/media through `Handler`.
  I checked the callers: `RunListener` is the only caller of `Handler`, and it posts 2049 bytes. Remedy: one of two.
  (a) State the ceiling as part of the documented behaviour, in §1 or in the description handed to the reviewer, so it is not
  a hidden divergence. (b) Make `Handler` and `Decide` agree. Either is a fixture change for the author, not for me.
- **R-2 · OWED · `tests/corpus/manifest.json:358-359`.** The manifest now reads `"fixture": "built"` and
  `"checks": "not run"`. CASE.md:5 reads `case code: built (2026-10-21, declared prior art) · checks: run (2026-10-21)`.
  The two records disagree on `checks`. The manifest also drops the "prior art, built before the window" disclosure that
  CASE.md keeps. Remedy: make the manifest line say what CASE.md says, including the prior-art disclosure.
- **R-3 · NOTE · `check.sh:9-11`, `:13`.** The portability items listed under Closure item 6. Remedy: document the
  preconditions (a git checkout, Go 1.26 or later) in CASE.md §8, and either write to a temporary directory or ignore
  `████████`.
- **R-4 · NOTE · `controls.go`, `loadDescription`.** A read error other than not-exist (for example permission denied) is
  printed as `failed: open <absolute path>: …`. It is classified as a failed control, not as "could not run". That is
  defensible. Noted only because it prints an absolute path from the judge's machine.

### Profile

Default reviewer profile plus the testing-bar skill. This is a test-fixture diff with no authentication, payments,
persistence or configuration surface, so the security profile does not apply (R9). The lang-go skill applied: both modules
are standard-library only, vet clean, and have no dependency delta (R4: two new `go.mod` files, neither has a `require`).

### Focus

- The order's six items, answered in Closure.
- R13 paired paths: `Decide` against `Handler` (R-1).
- Law #52: the false-world answer for item 2 (Closure 2, Unplug "What cannot fire").
- Testing-bar: the project declares no numeric bar for corpus cases, so there is no metric to re-run. All controls are suite checks.
- Declined to judge (R11), for the master or owner to rule:
  - AGENTS.md §3 says "Nothing written before 22 October may be copied in". The commit's author date is
    2026-10-21 17:36:28 +0000, and CASE.md records "declared prior art". Whether a declared prior-art fixture is allowed
    is the owner's decision. Licences are out of scope by order.
  - Whether EXPECTED.md:7's reason wording should match the printed reason. `expected/` is outside this diff.

### Conditions

- **C1 · BLOCKING until discharged.** Someone whose role may execute it runs `bash tests/corpus/cases/PV-020/check.sh` from the
  repository root on a fresh checkout at `8c55cee`, and pastes the full output and the exit code. Discharged when the 16 lines
  match CASE.md §8 byte for byte and the exit code is 0. Ideally run on Linux and on macOS (item 6).
- **C2 · OWED.** Make `tests/corpus/manifest.json:358-359` agree with CASE.md:5 (`checks`, and the prior-art disclosure). See R-2.
- **C3 · OWED.** Resolve the `Handler` 8 MiB ceiling against §2's "accepts it there". Either document it or align it. See R-1.
- **C4 · NOTE.** The portability preconditions and the work directory inside the checkout. See R-3.

### Unrun

- `bash tests/corpus/cases/PV-020/check.sh`: refused by this leg's permission class, verbatim:
  `the program "bash" is NOT in the ██████████████ list of role review under permission class ██████████████████████████, which is [saphan doc check go git ls cat grep find wc head tail diff pwd cut sort] — deny by default. NEXT — use one of those programs, or one of this run's typed tools [tool-command-not-in-class]` /
  `allowed form: none — "bash" is denied under permission class ██████████████████████████; do not rebuild the act from other commands, hand it to the role that holds it`.
  Per R15 and law #57, I did not rebuild the check from `go build` plus running the binaries. Item 1's measurement is therefore owed (C1).
- The check on macOS: I have no macOS seat in this leg.
- Mutation runs (R3): I worked out the controls' response to each mutation by reading the code (Unplug). I mutated and executed nothing,
  because executing a built binary is not in the permission class either.
- (Run, not unrun: `saphan doc check _REVIEW_STATUS.md --family review-status` → `family=review-status ok (checked=true)`.)

## The review status as filed

## REVIEW — corpus-pv-020
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-020 at 8c55cee, read in full: base and change differ only on the four non-exempt routes above 2048 bytes; runs (b) and (c) each remove exactly one precondition; no check bent; diff in lane; nothing private.
  check.sh UNRUN here (bash denied by the permission class), so C1 (BLOCKING) asks for a run whose output matches CASE.md §8.
  OWED: C2, the manifest says checks "not run" and drops the prior-art note that CASE.md keeps; C3, Handler refuses exempt uploads above 8 MiB while Decide accepts them, which goes against the "no defect" key.
  NOTE: needs a git checkout and Go 1.26 or later, and writes into <repo>/.saphan/. The false-world reach of the output is in the notes.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T17:30:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| file and line in Saphan Studio's private code | 3 |
| policy binding name | 3 |
