# REVIEW — corpus-pv-008 · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:34:49Z

## Summary

PV-008 (fe146e0...4e43d38) read against CASE.md §2 and expected/: base/change/reference predict exactly §8's four lines; change differs from the reference only on the wedged rows, so the output discriminates a correct change. Diff stays in its lane; expected/ untouched; nothing private; vet clean.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-008

Reviewer: claude-code (Anthropic family), stream corpus-pv-008-review-r2, 2026-10-23.
Scope: `fe146e0...4e43d38`, one commit (`[corpus-pv-008] tests: build PV-008 base, change, reference and check`).
Profile: none named by the order; the diff is a test fixture (no auth, data access, config or destructive path), so no
security profile applies (R9).

### Ground

- `git rev-parse --short=7 HEAD` printed `4e43d38` (ROLE-LOCK satisfied).
- `git diff --name-only fe146e0...4e43d38` printed exactly eight paths:
  `tests/corpus/cases/PV-008/{CASE.md, base/session.go, change/session.go, check.sh, check/main.go, go.mod, reference/session.go}`
  and `tests/corpus/manifest.json`.
- `git diff fe146e0...4e43d38 -- tests/corpus/cases/PV-008/expected` printed nothing: `expected/` is untouched.
- The manifest hunk is lines 147–152; `grep -n '"id"' tests/corpus/manifest.json` puts PV-008 at line 139 and PV-009 at
  159, so the hunk lies inside PV-008's own entry (`status.fixture`, `status.checks`).
- Local toolchain: `go version go1.26.8 darwin/arm64`. The module pins `go 1.26` (no `toolchain` line); §8 records
  `go1.26.5 linux/amd64`. Different patch and platform from the author's run.
- `go -C tests/corpus/cases/PV-008 vet ./...` → no output, exit 0. `go -C tests/corpus/cases/PV-008 build ./base ./change ./reference` → no output, exit 0.
- `go list -deps ./...` at the repository root lists no `pv008` package: the nested `go.mod` keeps the case out of the
  root module's build and test, so the root suite cannot be broken by it.
- **The check itself was NOT run** (see Unrun). `bash tests/corpus/cases/PV-008/check.sh` was refused by this seat's
  permission class; I did not rebuild it as `go run ./check`, which is the same act in another spelling (law #57, R15).
  Every statement below about what the check prints is therefore derived from READING the code, never from executing it.
- Master's claim (macm5, 2026-10-22) re-measured: "touches only its own case directory and manifest line, never
  `expected/`" — CONFIRMED by the three git reads above. "`check.sh` exit 0" — UNMEASURED here.

### Differential

What §2 names as the disagreement, against what the code does (read, not run):

| input | base | change | reference | follows from |
|---|---|---|---|---|
| normal close → late subscriber | no-frame | frame:ended | frame:ended | base `Close` feeds only `s.subs` present at publish (base/session.go:49-54); change stores `final` and `Subscribe` replays it (change/session.go:34-37, 63-66) |
| wedged stop → late subscriber | no-frame | no-frame | frame:ended+alarm:worker-not-released | change publishes only after `s.worker.Stop()` returns (change/session.go:60-71); the check releases the wedge only in `defer` AFTER `frame()` returns (check/main.go:111-113), so nothing can reach the channel within the 1 s limit; reference bounds stop at 100 ms (reference/session.go:58-68) |
| wedged stop → read state | state:closed | state:closed | state:closed | all three set `closed` under the lock before `Stop` |
| wedged stop → second close | returned | blocked | returned | change's second call waits on `<-s.done` (change/session.go:52-55), closed only after publish; reference's `done` closes ~100 ms in; base returns at once (base/session.go:40-43) |

- Change differs from the reference exactly on the three wedged inputs' observable outcomes (rows 2 and 4; row 3 is equal
  in all three, as §2 implies — the state is terminal in every version) and agrees on the normal path (row 1). That is the
  condition §2 and §4 name: a terminal state, an unbounded stop, and a final frame that late subscribers wait on.
- Base differs from change in rows 1 and 4. Row 1 is the §2 contract (base has no replay). **Row 4's base value is an
  author choice §2 does not make** — declared openly in §8 ("author's choice where §2 is silent"). It is not used by
  `expected/`, which compares change against the reference only.
- The behaviour of all three is deterministic under the check's ordering, with one timing margin: the reference's
  wedged-row frame relies on 100 ms < 1 s; a 900 ms scheduler stall would flip it to `no-frame`. That is a ~10× margin;
  not a finding, named as the one timing dependence.

### Unplug

Item 2 — does the printed output establish `expected/EXPECTED.md`'s outcome, or would it also fit the opposite (law #52)?

- EXPECTED's claim: with a wedged stop, change gives a late subscriber no frame within the limit, state reads closed, no
  alarm, and a second close does not return; normal-close change and reference agree.
- **False world A — change is correct (bounded stop, frame published anyway).** Then the change column would equal the
  reference column on rows 2 and 4 (`frame:…+alarm…`, `returned`). The predicted output (`no-frame`, `blocked`) is a
  different answer, so the check discriminates this world. ✔
- **False world B — change fails on a normal close too.** Row 1 would read `no-frame` for change. Predicted `frame:ended`
  rules it out, which supports EXPECTED's "must NOT be claimed" clause. ✔
- **False world C — change bounds the stop, but with a limit longer than the 1 s wait limit** (say 5 s). The check would
  print the SAME change column (`no-frame` / `blocked`). The check cannot tell "unbounded" from "bounded above 1 s"; no
  finite observation can. EXPECTED says "within the wait limit" and "does not return [within it]", so it does not
  overclaim — but the code reading (no timer anywhere in change/session.go) is what carries "unbounded", not the output.
  Owed note O3.
- **False world D — change never reaches `Stop` at all** (e.g. a Close that blocks before stopping). `start()`
  (check/main.go:92-94) returns silently when `w.entered` does not fire within 1 s, and the rows then print the same
  `no-frame` / `blocked`. The output does not distinguish "wedged in stop" from "never got to stop". Owed O2.
- **Row 3 alone is not evidence:** `state:closed` is the same in all three columns, including the reference. It proves
  only that the state is terminal while no frame exists, which is what EXPECTED asserts in combination with row 2.
- "No alarm": change has no alarm path (`composeFinal` returns the constant `"ended"`, change/session.go:43-45), and in
  the wedged rows it emits no frame at all, so the absence of an alarm is shown by row 2 plus the code, not by a frame
  that lacks one.

Verdict on item 2: under the predicted output, the evidence establishes EXPECTED's outcome against worlds A and B; it
does not by itself exclude C and D, and EXPECTED's wording stays inside what the output can carry.

### Closure

- Item 1 (fixture reproduces §2's condition): HOLDS by reading; the four-line output in §8 is exactly what the code
  predicts. NOT re-measured by execution on this seat.
- Item 2 (expected follows from evidence): HOLDS, with the reach named in Unplug.
- Item 3 (no bent check): no line is forced. The check prints observations; it has no expected values in it and no
  branch keyed on the implementation's name. `os.Exit(0)` is unconditional (check/main.go:138), so exit 0 says only that
  the check ended — it never asserts the outcome (consistent with §2, which asks it only to print). The one author
  choice (base row 4) is declared in §8.
- Item 4 (nothing private): `grep -rniE 'macm|marzec|wato|cyberasis|saphan|█████████████████████████████████]+\.(com|org)|ghp_|sk-' tests/corpus/cases/PV-008`
  → no matches. Model names in §3 (`claude-opus-5`, `gpt-5.6-sol`) pre-date this commit (contract text, not in the diff).
  The commit's author e-mail is git metadata, not case content.
- Item 5 (lane): HOLDS (Ground).
- Item 6 (portability), read:
  - `check.sh`: `#!/usr/bin/env bash`, `set -euo pipefail`, `cd "$(dirname "${BASH_SOURCE[0]}")"`, `exec go run ./check`.
    All valid in bash 3.2 (macOS) and bash 5 (Linux). No GNU-only or BSD-only flags.
  - The file is mode `100644` (`git ls-files -s`): `./check.sh` fails; `bash check.sh`, as documented, works.
  - Go code uses `fmt`, `os`, `strings`, `sync`, `time` only; no OS-specific call.
  - Non-portable dependency: `go.mod` says `go 1.26`. On a host with Go < 1.26 and default `GOTOOLCHAIN=auto`, `go run`
    tries to DOWNLOAD a toolchain — a network act §2 ("no network") excludes; with `GOTOOLCHAIN=local` it refuses. The
    code needs nothing newer than Go 1.x basics. Owed O6.
  - `go run` writes to `GOCACHE`; a read-only home makes it fail. Environmental, not a defect.

### Remainder

Owed, non-blocking (R6; each with location and remedy, R7):

- **O1 · check/main.go:108, 113, 118 — unbounded calls against §2's "every wait in the check is bounded".**
  `s.Subscribe()` and `s.State()` are called synchronously; a version that holds its mutex across `Stop` would hang the
  check forever. None of the three shipped versions does, so the output is unaffected today. Remedy: run each
  observation through the same bounded helper (goroutine + `time.After(waitLimit)`), printing e.g. `state:blocked`.
- **O2 · check/main.go:92-94 — silent fallback when the stop is never entered.** `start()` discards the `false` from
  `waitFor(w.entered)` and the row prints as if the stop were wedged. Remedy: return a sentinel and print
  `stop-not-entered` for that column.
- **O3 · CASE.md §8 / EXPECTED — the reach of "unbounded".** The output shows "not within 1 s"; the code carries
  "unbounded". Remedy: one sentence in §9 Limits saying the check cannot distinguish an unbounded stop from one bounded
  above the wait limit.
- **O4 · CASE.md:106-107 — base row 4 is an author choice.** Declared; owner may wish §2 to say whether base's second
  close is in the contract at all, since it is the only base/change difference not derived from §2.
- **O5 · CASE.md:5, manifest.json:150 — provenance declaration changed.** Every other case (20 of 20) still reads
  "built before the window, declared prior art"; PV-008 now declares "built inside the build window". That is an honest
  disclosure, and a change of the case's provenance class. "Window" is defined nowhere under `tests/corpus/`. Owner's
  ruling, not a code defect.
- **O6 · go.mod:3 — `go 1.26` forces a toolchain fetch on older Go.** Remedy: lower the `go` line to the lowest version
  the code needs, or record in §8 that Go ≥ 1.26 is required and `GOTOOLCHAIN=local` should be set so no network is used.

Quality (R8): `os.Exit(0)` at check/main.go:138 is redundant (main returning exits 0). Nothing else.

Declined to judge (R11): licences (order boundary); whether the corpus SHOULD contain a check that self-asserts against
`expected/` (§2 asks only for printing); the content of §1–§7 that pre-dates this commit.

### Profile

No named profile; test-fixture diff. testing-bar: no metric is declared for this module, so there is no bar to ratchet
(skill step: report, do not invent). Check ranking: the check is a *suite* check written by the same author as the
reference it compares against — the only ranking available; it is not external. lang-go: no `_test.go`, no race-sensitive
shared state outside mutex-guarded fields; `go vet` clean. `-race` run UNRUN (it would execute the check, see Unrun).

### Focus

Order items 1–6, in order; weight on item 2 (false-world answer) and item 3 (bent check). R3 mutation by reading: inverting
change's replay guard (`s.final != ""`, change/session.go:34) would flip row 1 change to `no-frame`; removing
`<-s.done` (change/session.go:54) would flip row 4 change to `returned`; giving change the reference's `select` would flip
rows 2 and 4. Each mutant is visible in the output — none survives. NOT executed.

### Conditions

- **C1 (blocking for the gate, owner/master act): run `bash tests/corpus/cases/PV-008/check.sh` from the repository root on
  Linux AND on macOS, paste the output, and confirm it equals the four lines in CASE.md §8 with exit 0.** This seat's
  permission class refused `bash`; the code reading predicts exactly §8's output, but a prediction is not a measurement.
- O1–O6 above are owed, not conditions.

### Unrun

- `bash tests/corpus/cases/PV-008/check.sh` — refused by the seat: `the program "bash" is NOT in the ██████████████ list
  of role review under permission class ██████████████████████████ ... do not rebuild the act from other commands, hand
  it to the role that holds it`. Not rebuilt as `go run ./check` (law #57). Handed to the master as C1.
- `go run -race ./check` — same act (executing the check); not run.
- Execution of the R3 mutants listed in Focus — would require editing files under review and executing the check; not run.

## The review status as filed

## REVIEW — corpus-pv-008
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-008 (fe146e0...4e43d38) read against CASE.md §2 and expected/: base/change/reference predict exactly §8's four lines; change differs from the reference only on the wedged rows, so the output discriminates a correct change. Diff stays in its lane; expected/ untouched; nothing private; vet clean.
  Condition C1: check.sh was NOT executed — this seat's class refused bash — so the master runs it on Linux and macOS and confirms §8's output. Owed O1–O6: unbounded Subscribe/State calls, silent stop-not-entered, the reach of "unbounded", the base row-4 choice, the prior-art→build-window declaration, and the go 1.26 toolchain fetch.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T00:00:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
| policy binding name | 2 |
