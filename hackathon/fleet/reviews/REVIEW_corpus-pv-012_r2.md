# REVIEW — corpus-pv-012 · round 2

- verdict: accept-with-conditions
- reviewer: claude-code on macm5, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:27:16Z

## Summary

PV-012 fixture realises §2 on all seven inputs by hand trace (application 5/0, control 0/0, extra-sync 4/0, no-registry 4 FAIL) and EXPECTED follows; diff stays in lane, nothing private. check.sh itself UNRUN here (bash denied) — C1: master runs it at 657ce27. Owed: driver leaks probes into review target (F4), crash reads as 0 with exit 0 (F2), build output into unignored ████████████ (F3).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-012

### Ground

- `git rev-parse --short=7 HEAD` printed `657ce27` (ROLE-LOCK satisfied).
- Scope `git diff --name-only fe146e0...657ce27` printed exactly five paths: `tests/corpus/cases/PV-012/CASE.md`,
  `tests/corpus/cases/PV-012/base/main.go`, `tests/corpus/cases/PV-012/change/main.go`,
  `tests/corpus/cases/PV-012/check.sh`, `tests/corpus/manifest.json`. Nothing under `expected/`. The manifest hunk is
  `@@ -222,2 +222,2 @@`, PV-012's own `fixture`/`checks` status lines only. Master's claim "touches only its own case
  directory and manifest line, never `expected/`" — RE-MEASURED, holds (item 5).
- Master's claim "`check.sh` exit 0" — NOT re-measured: `bash tests/corpus/cases/PV-012/check.sh` was refused on this
  seat (quoted under Unrun). Per law #57 / craft R15 I did not rebuild the check from `go run` calls. What follows for
  items 1–3 is a hand trace of the two programs and the script, input by input, plus `go vet` (allowed instrument).
- Toolchain on this seat: `go version go1.26.8 darwin/arm64`; `go.mod`: `module github.com/saphan-ai/proven`, `go 1.26`,
  no `toolchain` line. `go vet tests/corpus/cases/PV-012/base/main.go` and `go vet .../change/main.go`: both clean
  (no output, exit 0).
- The author's self-assessment (§8's recorded output) was set aside and re-derived (R1); my trace agrees with it on
  every line.
- Channel H1: the order asks for `# REVIEW — corpus-pv-012-review-claude`; the engine-injected grammar names
  `# REVIEW — corpus-pv-012` (the reviewed stream, the slug `merge` reads). I followed the engine and record the conflict
  here for the master.

### Differential

Hand trace (store after `run()` prologue: events `ours-1`(direct ours), `messages-1 edits-1 comments-1 uploads-1`
(no direct team), `audit-1`(direct theirs); registry after `sync("ours")`, `addForeign()`, `sync("theirs")` = every item
→ `theirs`). Foreign count = items of `check.sh:10` `foreign` served.

| input | base | change (filter path, `change/main.go:48-58`) | reference | §2 requirement |
|---|---|---|---|---|
| control | 5 | 0 — four dropped by registry≠ours, audit by direct theirs; control PASS | 0 | narrowing works when reached ✓ |
| vacuity | 5 | 5 — registry=theirs, audit direct theirs; `ours-1` dropped; PASS | 5 | all five served ✓ |
| application | 5 | 5 — `fromSettings` sets no team (`settings struct{}`, `change/main.go:34,40`), filter guarded by `page.team != ""` never entered | 0 | production path never narrows ✓ |
| scope-no-team | 5 | 5, "Activity for all teams" | 0 | ✓ |
| scope-with-team | 5 | 5, "Activity for all teams (request team: ours)" | 0 | sentence partly identical ✓ (see F1) |
| extra-sync | 5 | 4 — `sync("ours")` re-stamps registry; audit still dropped by its own column | 0 | four of five come back ✓ |
| control-no-registry | 5 | 4 — `registryFilter=false` keeps the four registry items; control FAIL | 0 | control notices absence ✓ |

Item 1: base and change differ exactly on control, extra-sync and control-no-registry counts and on the scope sentence;
they agree on application/scope counts, which is precisely where §2 locates the disagreement (narrowing exists but is
not reached from settings). Vacuity agrees by count (5/5) — base's difference there (`ours-1` served) is not visible in
a foreign-only count; harmless, the arm's job is non-vacuity of the change.

Item 2 — false-world answer (law #52). EXPECTED: established defect in change; evidence application (5 vs 0), control
(0 vs 0), extra-sync (4 vs 0). In the world where change were correct in production, the application line would read
change 0, and the extra-sync line 0 — the output discriminates. The control 0/PASS and control-no-registry 4/FAIL lines
also discriminate the "must NOT claim" items (narrowing wrong where it runs; control weak against mutation): both would
read differently if those claims were true. Reach of that discrimination: it proves the FIXTURE realises the contract;
it cannot prove the record's label, because the defect is built in by contract (`settings` has no team field by §2).
The label rests on EXPECTED's cited reproduction-and-repair, as EXPECTED itself says. One normative assumption is
carried by the reference, not by seeded data: the bare page's reference 0 presumes the requester is "ours".

Item 3: no check was bent to the expected line. `extra-sync` reports `control=not-applicable` rather than FAIL
(`change/main.go:84-86`); that hides nothing EXPECTED needs (it keys on the count 4), and §2 gives that input no control.

### Unplug

- Unplug the narrowing via the registry: realised as the `control-no-registry` input → change 4, control FAIL (traced).
- Unplug the production defect (give `fromSettings` a team): application line would read 0 — not executed (no edits;
  reviewer never fixes code), reasoned from `change/main.go:51`.
- Unplug a binary (crash/non-zero exit): `check.sh:23` puts `$($work/base "$input")` inside an argument of `result`,
  inside an argument of `printf`; under `set -e` a failing command substitution in an argument is not fatal. The line
  prints `0 (; control=)` and the script exits 0. ⇒ §8's `exit=0` proves the two `go build` calls succeeded and nothing
  about the runs (F2).
- The script never compares against the reference; it is a printer, so exit 0 never means "as expected". Allowed by §2
  (the check "prints one line per input"), named so nobody reads the exit code as a verdict.

### Closure

Diff closure: page, store, sync, two constructors, request, control, mutation switch `registryFilter`, input driver
`run()`, `main()`, and the shell check. Base and change share the store/sync code byte-for-byte (paired paths, R13);
only `page`, its constructors and `list`/`run` differ. The reference lives in `check.sh` (`foreign` list +
per-input literal) — never derived from the page, as §2 requires; it is in the check, not "in the case data" (§2 wording),
which I judge equivalent.

### Remainder

- F1 (owed, non-blocking) — `change/main.go:48-66`, `check.sh` scope-with-team line. §2's change bullet says change
  "keeps only the requesting team's events when the request carries a team". The fixture keys the filter on `page.team`
  (constructor) and lets `request.team` change the sentence only, so the scope-with-team line shows a request carrying
  "ours" and 5 foreign events served. That is a defensible reading of §2 ops 2–3 (team supplied at construction; request
  "with no filter"), but read against the change bullet it looks like the narrowing failing where it runs — the very
  claim EXPECTED says must not be made. Remedy: one sentence in CASE.md §8 (or a code comment) stating that the team
  narrowing is keyed on the page's team and the request's team only feeds the sentence.
- F2 (owed) — `check.sh:23`: a crashed binary is printed as count 0 and exit stays 0. Remedy: capture each output in a
  variable on its own line (`out=$("$work/base" "$input")`) so `set -e` aborts, or check the status explicitly.
- F3 (owed) — `check.sh:5`: build output goes to `$root/.saphan/work/corpus-pv-012`, the fleet's leg-confinement
  directory, inside a public repository whose `.gitignore` (`.DS_Store`, `.env*` only) does not ignore it. A judge who
  runs the check is left with two untracked binaries, and the public corpus names an internal tooling convention.
  Remedy: build into `$(mktemp -d)` with a `trap` cleanup, or a path under the case that is ignored.
- F4 (owed, for the case's later use, not this reproduction) — `base/main.go:52-73`, `change/main.go:69-104`: the input
  driver `run()` lives in the same file as the page under review, with inputs named `application`, `scope-no-team`,
  `scope-with-team`, `extra-sync` and `control-no-registry`, and `control-no-registry` flips `registryFilter`. §5 says the
  reviewer under evaluation receives the page, store, sync, control and diff, "never the check". Handing over
  `change/main.go` as-is hands over the check's probes, two of which ARE the expected findings. Remedy: move the driver
  and the mutation switch into a separate file (e.g. `driver.go` in the same package) that is withheld with the check,
  or state in §5 which parts form the target.
- F5 (note) — `check.sh:23` `$work/base` / `$work/change` unquoted: breaks if the checkout path contains a space.
- F6 (note, portability) — with `go.mod` declaring `go 1.26` and no `GOTOOLCHAIN` set, a judge whose local Go is older
  gets an automatic toolchain download (network) under the default `GOTOOLCHAIN=auto`; §2 says "no network". §8 should
  name the minimum local Go.

### Profile

Corpus-fixture review (testing-bar + lang-go). No security profile: no auth, data access, configuration or destructive
path in product code; the only destructive-adjacent act is `mkdir -p` into the repo root (F3). Check ranking
(testing-bar step 10): the whole case is a SUITE check — the fixture, its control and its reference are authored by the
same leg; nothing external constrains them. Reported as a NOTE; the independent W3 check is the mitigation the
manifest already tracks (`independent_check: not done`).

### Focus

1. Fixture reproduces §2 — YES by hand trace on all seven inputs; script run UNRUN on this seat (Unrun).
2. Expected follows from evidence — YES; the false-world output differs on application, control, extra-sync and
   control-no-registry lines. Limit: the fixture realises a contract that encodes the defect; the label's warrant is the
   record's repair, as EXPECTED states.
3. Case still proves something — YES; no bent check found. F1 and F4 weaken how it reads, not what it proves.
4. Nothing private — no host/seat names, no `/Users` paths, no tokens or e-mail addresses in the diff (grep of the
   unified diff). §8's `go1.26.5 linux/amd64` is a toolchain string, not a host. Only hit: `█████████████` (F3). The
   pre-existing §3 reviewer-family names are outside this diff.
5. Lane — YES (Ground). The header/manifest wording "built inside the build window" matches `tests/README.md:111,114-115`
   (16 cases coded in the window from 22 October).
6. Portability — bash-only constructs (`[[ ]]`, arrays, `<<<`, `(( ))`, `local`) all exist in macOS's bash 3.2; no GNU-
   only utility is called (`cd`, `dirname`, `pwd`, `mkdir -p`, `printf`, `go`). Non-portable/fragile calls named: F5
   (unquoted path), F6 (toolchain auto-download), F3 (write into repo root). macOS run itself UNRUN.

### Conditions

- C1 — the master (or a seat that admits `bash`) runs `bash tests/corpus/cases/PV-012/check.sh` at `657ce27` and confirms
  its seven lines equal the Differential table above and CASE.md §8; this reviewer could not (Unrun). Class: blocking
  until done.
- C2 — F4: separate the input driver and mutation switch from the review target, or state in §5 what the target is,
  before PV-012 is used in an evaluation. Class: owed.
- C3 — F2 and F3 in `check.sh`. Class: owed.
- C4 — F1 and F6, one sentence each in CASE.md §8. Class: owed.

Declined to judge (R11): licences (out of scope by order); whether `.saphan` as a name is "private" (the repo is
`github.com/saphan-ai/proven` with a root `SAPHAN.md`, so I judged it public and raised only the convention, F3).

### Unrun

- `bash tests/corpus/cases/PV-012/check.sh` — refused: `the program "bash" is NOT in the ██████████████ list of role
  review under permission class ██████████████████████████ ... allowed form: none`. Not rebuilt from `go run` (law #57).
  Covered by C1.
- macOS and Linux runs of the check — same refusal.
- `git check-ignore -v ███████████████████████████████` — refused ("This command requires approval"); replaced by
  reading `.gitignore` directly, which is a read, not the same act.
- Mutation of `fromSettings` (Unplug, second bullet) — not executed; a reviewer does not edit the code under review.

## The review status as filed

## REVIEW — corpus-pv-012
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-012 fixture realises §2 on all seven inputs by hand trace (application 5/0, control 0/0, extra-sync 4/0, no-registry 4 FAIL) and EXPECTED follows; diff stays in lane, nothing private. check.sh itself UNRUN here (bash denied) — C1: master runs it at 657ce27. Owed: driver leaks probes into review target (F4), crash reads as 0 with exit 0 (F2), build output into unignored ████████████ (F3).
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T00:00:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 4 |
| policy binding name | 2 |
