# REVIEW — corpus-pv-020 · round 3

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T15:38:58Z

## Summary

PV-020 at 3cd2e64 reproduces §2 by code trace (base/change differ on exactly the 4 non-exempt routes; could-not-run kept apart from failed); lane confirmed; nothing private; go test/vet green; round-2 HTTP guard kills its mutant.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-020

Reviewer: claude-code, model claude-opus-5-5 (the Anthropic family). Scope: `ae34d44...3cd2e64`, 2 commits
(`75fd68f` round 1, `3cd2e64` round 2). Measured 2026-10-23 on Linux, `go version go1.26.5 linux/amd64`.

### Ground

- `git rev-parse --short=7 HEAD` printed `3cd2e64`. ROLE-LOCK holds.
- `git diff --name-status ae34d44...3cd2e64` lists 13 paths: 11 added and `CASE.md` modified, all under
  `test/corpus/cases/PV-020/` and none under `expected/`, plus `test/corpus/manifest.json`. The manifest hunk
  (`-U12`) changes two values inside the `"id": "PV-020"` status object (`fixture`, `checks`) and nothing else.
  The master's lane claim (macm5, 2026-10-22) is **confirmed**.
- `bash test/corpus/cases/PV-020/check.sh` was **not run**: my permission class refuses `bash` (quoted under Unrun).
  So I cannot paste its output. I did not rebuild the check from its `go` parts (law #57). The master's claim
  "`check.sh` exit 0" is **unverified by me**. It is also weak on its own: `check.sh` exits 0 whatever a control
  reports (its header, lines 4–5, says so), so exit 0 means only "did not crash". The real evidence is the 16 lines
  compared with CASE.md §8.
- What I did run: `go -C test/corpus/cases/PV-020/change test -count=1 ./...` → `ok pv020change/service 0.018s`;
  `go vet ./...` in `change/` and in `base/` → no output (clean).
- Instead of running the check, I traced by hand each line `check.sh` prints through `change/service/service.go`
  `decideSize`, `controls.go` `RunInProcess` / `RunDescription` / `RunListener` / `outcomeFromErrs`, and
  `cmd/controls`, `cmd/decide`. Every one of the 16 lines in CASE.md:102–117 follows from the code, including the
  limit-removed failure list in the order printed (limit-present; over-limit-refused ×4; over-limit-exempt ×2).
  Run (a) and run (b) listener `passed` depend on the host allowing a loopback listener. I did not measure that here.

### Differential

Property 1 — the fixture reproduces the condition §2 names. **Holds, by reading.**
- base (`base/service/service.go:37-43`) accepts any body on a known route. change (`change/service/service.go:74-87`)
  refuses `size > 2048` on every non-exempt route and accepts with `exempt upload` on the two documented uploads.
- The recorded route lines (CASE.md:111–116) differ between base and change on exactly the 4 non-exempt routes. They
  agree (accepted) on the 2 exempt routes. That is where §2 places the disagreement, and nowhere else.
- One unprinted difference, consistent with §2 ("every route"): an oversize body on an unknown route is
  `body exceeds limit` in change and `unknown route` in base (`service.go:77-82`).
- The three runs: (b) passes an absent path, and `loadDescription` maps ENOENT to `description not provisioned`
  (`controls.go:159-168`). (c) uses `PV020_SIMULATE_LISTEN_REFUSAL=1`, which returns EPERM before any bind
  (`controls.go:242-247`), and leaves the limit alone. §2's rule "never simulated by weakening the limit" holds.

Property 2 — the expected outcome follows from the evidence. **Holds for the label. Two sentences in EXPECTED.md say
more than the check prints.**
- False-world answer for the label ("no defect in change"): a defect in the limit (an off-by-one, a drifted exempt
  list, a bypass) turns the in-process line `failed` and moves a route line. The description control compares against
  a separate file, so a drift between the code and the description is also visible. The printed lines therefore tell
  the two worlds apart, with one exception. A defect shared by `Routes`, `DocumentedExempt` and `description/api.json`
  (one author, one commit) passes all three controls. That is inside the case's own definition of "documented", so it
  is not a gap in the case.
- `expected/EXPECTED.md:4-5`: "with the limit removed … oversize bodies are accepted on more routes". The check prints
  no route table with the limit removed. It prints only control names. `inproc.over-limit-refused <route>` fires on
  `d.Accepted || d.Reason != "body exceeds limit"` (`controls.go:73`). A service with no limit that *refused* with some
  other reason would print the same line. "Accepted" is true by code reading, but the check does not print it (N2).
- `expected/EXPECTED.md:7` quotes the listener reason as "listener refused". The check prints
  `could not run: operation not permitted`, which is §2's wording (CASE.md:41-42, 110). An evaluator who matches the
  answer key's quoted string will not find it (N1).

Property 3 — the case still proves something. **Holds.** No check is bent to make a line come out as expected.
`could not run` (absence) and `failed` (a measured defect) are separate paths (`controls.go:326-346`). A malformed
description gives `failed: description.unreadable`, not `could not run`. The limit-removed run turns red on a real
property (`MaxBody`, and the refusals), not on a planted flag. One weak point, not a bend: the two
`over-limit-exempt` failures in the limit-removed run come from the reason label (`exempt upload` is emitted only
when `maxBody > 0`, `service.go:84`), not from an exempt route changing its answer.

Property 4 — nothing private is left. **Holds.** A grep of the added lines for home or user paths, mail addresses,
host or seat names, tokens and `bridge` matched only the relative scratch path `██████████████████████████`
(`check.sh:10`, CASE.md:94). That path is inside the reader's own checkout, not ours. The commit author
`██████████████████████` is git metadata set by the engine, not a file in the diff. CASE.md:52-53 names reviewer model
families, which the corpus publishes on purpose (PROVENANCE.md:22).

Property 5 — the diff stays in its lane. **Holds** (Ground).

Property 6 — the check runs where a judge runs it. **Portable by reading. Not run on macOS** (no host, the same as the
author).
- `#!/bin/bash`, `set -euo pipefail`, `CDPATH= cd -- "$(dirname "$0")" && pwd`, `mktemp -d "$scratch/run.XXXXXX"`
  (BSD and GNU both accept a path template), `exec 3<`, `read -r <&3`, `${var%%|*}` and `$(( ))`. All of these work
  on macOS's Bash 3.2. There is no GNU-only flag (no `sed -i`, `readlink -f` or `date -d`).
- `go -C` needs Go 1.20 or later. The documented floor is Go 1.26, so that is covered. With an older Go,
  `GOPROXY=off` (`check.sh:15`) blocks the automatic toolchain download, so the check stops with a Go error. That
  matches the documented "Go 1.26 or later".
- The repository root is `case_dir/../../../..` (`check.sh:9`). That is correct for `test/corpus/cases/PV-020`.
- Not portable as a document: the command in CASE.md:98 is `bash tests/corpus/cases/PV-020/check.sh`. After the
  ADR-0030 move that path does not exist, so a judge who copies §8 gets "No such file" (C1).

### Unplug

- Round 2's HTTP fix (`service.go:127-133`, "count streamed bodies without a second ceiling"). I unplugged it: a mutant
  copy under `█████████████████████████████████████████████████` adds `http.MaxBytesReader(w, rc, 8<<20)` to the
  exempt branch of `readBodySize`. `go -C …/mut1 test -count=1 ./service/` →
  `FAIL … status=400 decision={Accepted:false Reason:body unreadable}` on both `/v1/uploads/media` and
  `/v1/uploads/bundle`. The guard fires. **Mutant killed.**
- That guard is a `go test` test. `check.sh` never runs `go test`, and the listener control sends only 2049 bytes
  (`controls.go:261`). So a judge who runs only the check would not see a returning ceiling (N6).
- Limit unplugged: the contract's own limit-removed run does this. By code trace it prints 7 named failures
  (CASE.md:117). I did not execute it (Unrun).

### Closure

- The round-1 conditions were not carried into this order. It says they were "carried verbatim into its order",
  meaning round 2's order, which was not provisioned here. I cannot mark each one ADDRESSED or NOT ADDRESSED.
  **Unverified** (C3).
- What round 2's commit message says it paid, checked against the diff:
  - "align HTTP uploads": `service.go:127-133` plus `service_test.go`. Present, and the guard is live (Unplug).
  - "record the measured check": CASE.md §8 holds the 16 lines, the toolchain and the exit code. But the command line
    is stale (C1).
  - "prior-art status": CASE.md:5 and manifest `fixture`/`checks`. Present.
  - "portable scratch handling": `check.sh:8-13`, relative to the repository root. Present.
- The status object in `test/corpus/manifest.json` is an OBJECT with 4 keys. "Restored" is confirmed.

### Remainder

NOTEs (non-blocking, R6). Each names a location and a remedy.
- N1 `expected/EXPECTED.md:7`: "listener refused" ≠ printed `operation not permitted`. The remedy is the scribe's
  (expected/ is outside this lane): quote the check's string.
- N2 `expected/EXPECTED.md:4-5`: "accepted on more routes" is not printed with the limit removed. Remedy: either the
  check prints a route table for the limit-removed service, or EXPECTED.md says "the controls turn red".
- N3 `check.sh`: exits 0 whatever the controls report, and compares nothing. Agreement with §8 is left to the reader.
  Remedy (optional): print a final line comparing the output with §8, or document that exit 0 means "ran".
- N4 `check.sh:10-13`: `████████` is not in `.gitignore` (`grep -n saphan .gitignore` matched nothing). The scratch
  directory stays as untracked `████████` in a judge's checkout after the trap removes only `run.*`. My own mutant
  directory showed up the same way in `git status --porcelain --ignored`. Remedy: remove the empty scratch directory
  in the trap, or say in §8 that it stays.
- N5 CASE.md:92 cites `8c55cee`. `git cat-file -t` finds the object locally, but `git branch --contains 8c55cee`
  prints nothing: it is a commit from before the rewrite, which no branch reaches. A stranger cannot resolve it.
  Remedy: cite `75fd68f`, or say that the hash is from before the rewrite.
- N6 `service_test.go` is not run by `check.sh` (Unplug). Remedy (optional): add a `go test` line to the check.
- N7 CASE.md:5 and the manifest say "built 2026-10-21, declared prior art". Round 2 changed product code of the
  fixture on 2026-10-23 (`3cd2e64` author date), which is inside the window. Whether a change inside the window keeps
  the prior-art label is the owner's call. Remedy: name the 2026-10-23 change in the status line.
- N8 `controls.go:163-168`: a description file that exists but cannot be read (EACCES) is reported `failed`, not
  `could not run`. Outside the contract's two cases, so it is not a defect here.

### Profile

None of the named profiles (security, backend, ui, mobile, docs) fits a test fixture. I judged it as a test case
against its contract (`testing-bar`, `lang-go`). R13 (paired paths): `Decide` and `Handler` share `decideSize`, and
`limitApplies` agrees with `decideSize` on unknown routes (both refuse oversize). A test of the same shape covers only
the exempt side over HTTP. The non-exempt side over HTTP is covered by the listener control.

### Focus

The six properties the order names, round 2's claims, and the master's two claims (the lane: confirmed; exit 0:
unverified). Declined to judge (R11): licences (out of scope by order), the case's design choices beyond §2
(for example, refusing oversize bodies on unknown routes), and EXPECTED.md's wording beyond N1 and N2.

The order's H1 for `_REVIEW_STATUS.md` (`# REVIEW — corpus-pv-020-review-r2-claude`) contradicts the engine's channel
grammar (`# REVIEW — corpus-pv-020`). I followed the engine, which states that the verdict binds to the reviewed
stream's slug.

### Conditions

- C1 (OWED) `test/corpus/cases/PV-020/CASE.md:98`: `bash tests/corpus/cases/PV-020/check.sh` → change it to
  `bash test/corpus/cases/PV-020/check.sh`. The ADR-0030 move left the reproduction command pointing at a path that
  is gone. "Renames only, byte-identical" is the reason it is stale.
- C2 (OWED, before the gate, master) Run `bash test/corpus/cases/PV-020/check.sh` from the repository root. Record
  its exit code, and `diff` its 16 lines against CASE.md:102–117. I could not run it (Unrun). My by-hand trace
  predicts a byte-for-byte match on a host that allows a loopback listener.
- C3 (OWED, before the gate, master) Check each round-1 review condition against `3cd2e64`, or carry their text into
  a re-review order (law #44 (3)). They did not reach this leg.

### Unrun

- `bash test/corpus/cases/PV-020/check.sh`. Refused by the hook: "the program "bash" is NOT in the ██████████████ list
  of role review under permission class ██████████████████████████ … allowed form: none". Not rebuilt from its
  `go build` / binary calls (law #57). Its output is not pasted, and the master's "exit 0" is not re-measured. → C2.
- The limit-removed run, runs (a)–(c), and the route table as live output. Same cause. Traced by hand only.
- macOS run of the check. No macOS host.
- Round-1 conditions closure (R14). Their text was not provisioned. → C3.
- Removal of my mutant directory `████████████████████████████████████████████` (3 files). `rm` is not in my class.
  It is the one leftover of mine in `git status --porcelain`, and the master removes it.

## The review status as filed

## REVIEW — corpus-pv-020
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-020 at 3cd2e64 reproduces §2 by code trace (base/change differ on exactly the 4 non-exempt routes; could-not-run kept apart from failed); lane confirmed; nothing private; go test/vet green; round-2 HTTP guard kills its mutant.
  check.sh itself UNRUN (bash refused to this class) — master runs it and diffs against CASE.md:102-117 (C2); round-1 conditions not provisioned (C3).
  C1: CASE.md:98 still says tests/corpus after ADR-0030. NOTES N1-N8, incl. EXPECTED.md quoting "listener refused" vs printed "operation not permitted".
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-23T15:55:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 5 |
| e-mail or user name | 1 |
| policy binding name | 2 |
