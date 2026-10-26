# REVIEW — corpus-pv-019 · round 2

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:28:52Z

## Summary

PV-019 code matches §2 line for line, and the recorded check output establishes EXPECTED.md's outcome (output differs in the false world). The diff stays in its lane; expected/ untouched; go test ./change ok, go vet clean.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-019

### Ground
- `git rev-parse --short=7 HEAD` printed `5b65589`, as the order requires.
- Scope: `git diff fe146e0...5b65589`, one commit, 10 files, +259 −4.
- `git diff fe146e0...5b65589 --name-only -- tests/corpus/cases/PV-019/expected` printed nothing: `expected/` is untouched.
- Files touched: `tests/corpus/cases/PV-019/{CASE.md, base/validator.go, change/content.go, change/content_test.go, change/validator.go, check.sh, check/main.go, go.mod, table/table.go}` and `tests/corpus/manifest.json` (1 line, inside the PV-019 entry at line 341). The master's claim "touches only its own case directory and manifest line, never `expected/`" holds.
- `go -C tests/corpus/cases/PV-019 test -count=1 ./change` → `ok  	pv019/change	0.002s` (go1.26.5 linux/amd64).
- `go -C tests/corpus/cases/PV-019 vet ./...` → no output, exit 0.
- `bash tests/corpus/cases/PV-019/check.sh` → REFUSED by this seat's hook (see Unrun). I did NOT see the check print. The output quoted in CASE.md §8 is the author's; I checked it by reading the code, not by running it.
- The author's self-reading in CASE.md §8 ("reading, by the author…") was set aside (R1).

### Differential
Static trace of `check/main.go` over `table.Shipped()` and each input, against the output recorded in CASE.md:98–106. Every line matches what the code must print:
- shipped-table: base refuses `ed-cedar` (`unknown status "preview"`, base/validator.go:17); change passes (preview row has no install route, channel non-empty); reference passes (all channels in the closed set main.go:16). Matches CASE.md:98.
- preview-with-install-route: change refuses (change/validator.go:23–25); reference refuses. CASE.md:99.
- preview-with-session-switch, preview-with-plugin-folder: change passes (preview branch checks only `InstallRoute`); reference refuses (main.go:30–32). CASE.md:100–101.
- unsupported-with-install-route: change passes (change/validator.go:21, no check); reference refuses `ed-birch` (main.go:26–28). CASE.md:102.
- channel-wrong-nonempty: change passes (only emptiness checked, change/validator.go:30); reference refuses; content test fails (content.go:20–22). CASE.md:103–104.
- channel-empty: change refuses `failure channel is empty`; reference refuses; content test fails. CASE.md:105–106.

Item 1 (fixture reproduces §2): yes. base, change, content test, reference and the seven inputs each match §2 line for line. Change and reference disagree on exactly the four inputs §2/§6 name as holes (session switch, plugin folder, unsupported install route, wrong non-empty channel) and agree on the other three.
Base vs change: base fails all 7 inputs with the same refusal, because every input is built from `table.Shipped()`, which carries the `preview` row (main.go:69–77). This follows from §2 (base refuses an unknown status), but it means the base column carries no information past "base does not know preview". See F3.

Item 2 (expected outcome follows): EXPECTED.md:4–8 asks for change=pass/reference=fail on three inputs, agreement on the shipped table, and validator=pass with content test red on a wrong channel. The recorded lines 100–102, 98, 103–104 carry exactly that.
False-world answer (law #52): if the change validator closed the holes, lines 100–102 would read `change: fail`; if the content test missed the wrong channel, line 104 would read `pass`. The output differs between the two worlds, so it is a measurement. It is not the same answer under "no defect".
One mismatch in wording, outside this diff: EXPECTED.md:7 says the shipped-table line shows "all agree"; the line shows `base: fail`. Only change and reference agree. See F4.

Item 3 (nothing bent): change/validator.go implements §2's holes exactly as written. It does not narrow them to fit the answer key. The reference (main.go:19–41) applies the §2 meaning in full. No input is weakened. No line is filtered.

Item 4 (nothing private): grep over the case (excluding `expected/`) for host names, user names, absolute paths, e-mail, tokens and private record names: one hit, CASE.md:90–91 ("the seat's command allow-list"). That is fleet vocabulary, not a name. No secret, host, path or address is present. CASE.md:113 names commit `5d8a8b8`, which does not exist in this repository, and "the order", which is an internal document (F2).

Item 5 (lane): holds; see Ground.

Item 6 (portability): `check.sh` uses `#!/usr/bin/env bash`, `set -eu`, `cd "$(dirname "$0")"`, `go run ./check`. Each works the same on Linux and macOS (no GNU-only flags, quoted path). The file mode is `100644` (`git ls-files -s`), so it runs only as `bash check.sh`, which is how CASE.md and the header say to run it. Non-portable dependency: the toolchain, see F7.

### Unplug
- `check.sh` always exits 0 while main.go prints its lines: main.go has no assertion and no non-zero exit. A run whose lines contradict EXPECTED.md also exits 0. The master's claim "`check.sh` exit 0" has the same value in the true and the false world, so it is not evidence about the outcome; it proves only that the program compiled and ran (F5).
- Mutation by reasoning, not run (R3): removing the install-route guard at change/validator.go:23 would flip CASE.md:99 to `change: pass`. Removing the empty check at change/validator.go:30 would flip line 105 to `change: pass`. Changing `PinnedChannel` would turn line 98's companion `go test ./change` red. Each guard is observed by at least one line. Not executed: see Unrun.
- `edit()` (main.go:53–60) returns the table unchanged when the editor name is absent. A misspelt editor would silently produce a copy of the shipped table. It is not live: each recorded line differs from shipped-table where it should (F6).

### Closure
Findings, most severe first. None is blocking.
- F1 · OWED · tests/corpus/manifest.json:342 — `"checks": "not run"` contradicts CASE.md:5 (`checks: run 2026-10-22`) and §8. Line 341 `"fixture": "built"` drops the date that CASE.md:5 and the neighbouring `"contract": "written 2026-10-19"` carry. Trigger: any reader or tool of the manifest is told the checks were never run. Remedy: set `"checks": "run 2026-10-22"` and `"fixture": "built 2026-10-22"`, or reconcile CASE.md:5 with whichever value the corpus owner rules on.
- F2 · OWED · tests/corpus/cases/PV-019/CASE.md:113 — `the worktree's base is main at 5d8a8b8, not the d81a93d the order names`. `git cat-file -t 5d8a8b8` → `fatal: Not a valid object name` (the main rewrite removed it). "the order" points at a document the public cannot read. Remedy: drop the line, or restate it against a commit that exists here (the commit was carried onto `fe146e0`).
- F3 · NOTE · check/main.go:69–77 — every input carries the `preview` row, so the base column is `fail [ed-cedar: unknown status "preview"]` on all 7 lines. Base's own behaviour on, for example, `unsupported-with-install-route` is masked. §2 does not ask for more, and EXPECTED.md needs no base evidence. Remedy (optional): add preview-free variants for base, or say in §8 that the base column is uninformative by construction.
- F4 · NOTE · tests/corpus/cases/PV-019/expected/EXPECTED.md:7 — out of this diff's lane. "the line for the shipped table showing all agree" does not hold literally: base fails it. For the corpus owner to word as "change and reference agree". Not a change this diff may make.
- F5 · NOTE · check.sh:6 / check/main.go:63–85 — exit 0 is not a verdict (see Unplug). Remedy (optional): have `check` compare its lines with the recorded §8 block, or state in §8 that exit 0 means "ran", not "as expected".
- F6 · NOTE · check/main.go:53–60 — `edit` is silent on a missing editor. Remedy (optional): panic when no row matched.
- F7 · NOTE · tests/corpus/cases/PV-019/go.mod:3 — `go 1.26` with no toolchain line. A judge with Go 1.21–1.25 and the default `GOTOOLCHAIN=auto` makes `go run` download a newer toolchain over the network, against §2's "no network". With Go older than 1.21 the build fails. Remedy: name "Go 1.26 or later already installed" in §8, or note `GOTOOLCHAIN=local` and its refusal.

### Remainder
- I did not see the check run. The match between code and recorded output is a static trace (Differential), not a measurement.
- The reference's closed set of channels (main.go:16) was chosen by the case author; §2 names no members. It contains every shipped value, so it does not bend the shipped-table line.
- CASE.md:89 records `go1.26.8 darwin/arm64`. The run here used go1.26.5 linux/amd64 for `go test` and `go vet` only. macOS was not exercised by this review.
- The order's H1 for `_REVIEW_STATUS.md` (`corpus-pv-019-review-claude`) disagrees with the engine's channel grammar (`corpus-pv-019`). The engine's form was used, as the grammar instructs.
- Declined to judge (R11): licences (out of scope by the order); whether the corpus manifest has a documented status vocabulary (none was found under `tests/corpus` outside `cases/`).

### Profile
Test-fixture review (docs/test corpus). No security, backend or UI surface. Skills read from the order: lang-go, testing-bar (provisioned). Reviewer: claude-code, model claude-opus-5-5. The `updated` stamp in `_REVIEW_STATUS.md` is the run's start stamp (`20261022-182343`): `date` is not on this seat's allow-list.

### Focus
The order's six properties, in that order. Most weight was on item 2 (does the evidence establish the outcome, and what does it read in the false world) and on item 1's base-vs-change shape.

### Conditions
1. The master runs `bash tests/corpus/cases/PV-019/check.sh` from the repository root at `5b65589` and compares the output byte for byte with CASE.md:98–106 before the gate. This reviewer could not run it (Unrun). If it differs, this verdict does not stand.
2. F1: reconcile `tests/corpus/manifest.json:341–342` with CASE.md:5.
3. F2: remove or restate CASE.md:113 so it names no missing commit and no internal order.
F3–F7 are notes; they do not condition the verdict.

### Unrun
- `bash tests/corpus/cases/PV-019/check.sh` — refused by the seat's hook: `the program "bash" is NOT in the ██████████████ list of role review under permission class ██████████████████████████ … allowed form: none — "bash" is denied …; do not rebuild the act from other commands, hand it to the role that holds it`. Not rebuilt as `go run ./check` (law #57). Handed to the master as Condition 1.
- Mutation runs of change/validator.go and check/main.go (R3) — they need the check to run; reasoned only.
- A run on macOS — no macOS host in this leg.

## The review status as filed

## REVIEW — corpus-pv-019
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-019 code matches §2 line for line, and the recorded check output establishes EXPECTED.md's outcome (output differs in the false world). The diff stays in its lane; expected/ untouched; go test ./change ok, go vet clean.
  Conditions: (1) the master runs check.sh and diffs against CASE.md §8, which this seat refused (bash denied, UNRUN); (2) manifest.json:342 still says checks "not run"; (3) CASE.md:113 names commit 5d8a8b8, absent from this repo, and an internal order.
  Notes: base column uninformative by construction; check.sh exit 0 is not a verdict; go.mod may fetch a toolchain on Go <1.26.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T18:23:43Z

## REDACTED

| what | times |
|---|---|
| policy binding name | 2 |
