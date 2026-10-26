# REVIEW — corpus-pv-016 · round 2

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T19:33:45Z

## Summary

PV-016 at 5af73bc reproduces §2: base and change differ only by docs/; go probes match §8 (delivered=0, t- refused, retry_count and threshold absent); a send=true mutant prints delivered=1, so the reading separates the worlds.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-016

### Ground
- Reviewer: claude-code, model claude-opus-5-5. Role: review. One of two independent reviews; I have not seen the other.
- HEAD, verbatim: `git rev-parse --short=7 HEAD` → `5af73bc`; `git rev-parse HEAD` → `5af73bcfb1a81cb44a990e1a72edab75f8d93237`. Role-lock met.
- Scope: `fe146e0...5af73bc`, one commit, `5af73bc [corpus-pv-016] tests: build PV-016 base, change and check; record run`. Author date 2026-10-22T15:58:08Z, inside the 22–27 October build window.
- Toolchain on this host: `go version go1.26.5 linux/amd64` (the same version CASE.md §8 records).
- What I read: CASE.md (all of it), check.sh, claims.tsv, reference.md, base/ and change/ code, the four change pages, expected/EXPECTED.md (read only after the code and pages, as the order asks), and the manifest hunk.
- R1: the author's claim that the code was built "from §2 alone, without reading expected/" cannot be checked from the diff. I set it aside. It is not evidence either way.
- The master's claim (2026-10-22) has two parts. (a) The commit touches only its case directory and its manifest line, never expected/: CONFIRMED (`git diff --name-status`, below). (b) `check.sh` exits 0: NOT RE-MEASURED. My permission class refuses `bash` (see Unrun).

### Differential
Item 1: the fixture reproduces the condition §2 names. HOLDS, measured by probe, but not through check.sh.
- `diff -r tests/corpus/cases/PV-016/base tests/corpus/cases/PV-016/change` prints exactly one line: `Only in tests/corpus/cases/PV-016/change: docs`. Base and change share the same code byte for byte. The only difference is the pages, which matches §2 ("one version of the tool's code, used for both base and change").
- I probed the change tool directly with `go -C tests/corpus/cases/PV-016/change …`. `go vet .` is clean. Output, verbatim:
  ```
  run . dest chat:t-backup-ops   → refused: destination id "t-backup-ops" must begin with c-
  run . dest chat:c-backup-ops   → accepted: c-backup-ops
  run . emit backup_failed       → considered=1 delivered=0 reason="no transport yet"
  run . emit delivery_failed     → considered=1 delivered=0 reason="no transport yet"
  run . fields backup_failed     → error,host,job,time
  run . fields storage_low       → percent_free,time,volume
  ```
  Each line is identical to the matching "code did" text in CASE.md §8. I did not probe backup_late, restore_done or storage_low emit, or the backup_late, restore_done and delivery_failed fields. Reading the code, `catalogue` at `base/alerts.go:19-25` and `resolveSender` at `base/alerts.go:63-65` give the recorded values.
- The §2 obligations, checked against the code and the pages:
  - The sender always resolves to one that sends nothing, with reason "no transport yet": `base/alerts.go:57-65`. ✓
  - A promise of delivery: `change/docs/alerts/what-goes-out.md:3`. ✓
  - A setup destination that the code refuses: `chat-setup.md:4` (`chat:t-backup-ops`) against `base/alerts.go:36-38` (`c-` prefix). ✓
  - Field lists that claim absent fields: `retry_count` (backup_failed) and `threshold` (storage_low). Neither appears in `catalogue`. ✓
  - Four page-internal findings: the actor (`what-goes-out.md:9` against `related.md:5`), no owner of the credential and no sending machine (`chat-setup.md:5-6`), delivery_failed sent to the destination that failed (`related.md:9-10`), and `c` against `t` (`index.md:12` against `chat-setup.md:4`). ✓ All four match `reference.md:19-25`.
- No difference where §2 names none. The code is identical, so every code-facing line is the same on base and change. Only "pages say" differs.

Item 5: the diff stays in its lane. HOLDS. `git diff --name-status fe146e0...5af73bc` lists 13 added files and 1 modified file under `tests/corpus/cases/PV-016/` (none under `expected/`), plus `tests/corpus/manifest.json`. The manifest hunk is one line at line 290, inside the `"id": "PV-016"` object (id at line 279).

Item 4: nothing private is left. HOLDS. `git diff fe146e0...5af73bc | grep -niE 'macm|███████|█████████████████████████████████████████████]+\.[a-z]|tok_|seat|\.local|<IPv4>|wato|cyberasis'` matches only the synthetic `tok_synthetic_0000_not_a_secret`, three times. The contract requires that value (§2). The module path is `example.invalid/…`. No hosts, seats, private paths, real tokens or e-mail addresses.

### Unplug
- R3 / law #41: what does the decisive reading say when the world is the other way? I made a mutant under `██████████████████████████████████████████████████`. It is base/ with exactly one code change, `nullSender.Send` returning `true` (`diff` shows line 58 and the header comment only). `go -C … run . emit backup_failed` → `considered=1 delivered=1 reason="no transport yet"`. The `delivered=` field therefore separates the two worlds. A tool that delivers prints `1` and this one prints `0`. The reading is a measurement, not a formality.
- The false-world trap this exposes: in the mutant, the `reason="no transport yet"` text is printed unchanged even though the alert was delivered. `Emit` copies `s.Reason()` regardless of the result (`base/alerts.go:85`). So the reason string alone is evidence of nothing. Only `delivered=0` carries the finding. EXPECTED.md keys on "delivered = 0 for each", which is correct. A reviewer who quotes only the reason would quote a line that also fits the opposite outcome. Class NOTE.
- Not unplugged: the destination prefix guard and the catalogue omissions. Both are single-line guards, and the recorded lines would flip on removal by reading.

### Closure
Item 2: the expected outcome follows from the evidence. HOLDS.
- EXPECTED.md asks for "established defect in change", proven by `delivered = 0` for each listed alert and by the setup step being refused. The recorded lines (and my probe for 2 of 5 alerts plus the setup) show exactly that. The false-world answers differ. A delivering sender gives `delivered=1` (measured, Unplug). A code that accepted `t-` gives `accepted: t-backup-ops` (read from `base/alerts.go:36`). So the output does not also fit the opposite outcome.
- "Not to be claimed: a defect in the code's other operations" is consistent with the code. `ParseDestination` and `FieldsOf` behave as their comments say.
- The pages alone do not reveal the missing transport. `chat-setup.md:7` even says "Alerts are delivered from then on". This matches the case's point that a pages-only reviewer could not see it.

Item 3: the case still proves something. HOLDS, with one weakness (C3 below).
- claims.tsv quotes the pages verbatim. The quotes are exact substrings (the check uses `grep -qF`). The `claimed` lists match the page lists. Nothing in check.sh rewrites a result. It prints what the binary returns. I found no bent check.
- Weakness: on change, a quote that no longer matches its page prints `(no such page)` and still exits 0 (`check.sh:28-33`). For `fields` lines it also silently skips the claimed-but-absent comparison (`check.sh:47`). An edit that drops the delivery promise from the pages would therefore not turn the check red. It would only change one printed column. Read from the code, not run (bash refused).

### Remainder
Findings, with location, trigger and remedy (R7). None blocks; the case's core property holds.
- **C1 · OWED · `tests/corpus/manifest.json:291`.** `"checks": "not run"` contradicts CASE.md line 5 ("checks: run 2026-10-22") and §8 (a recorded run with exit 0). The commit changed `fixture` on line 290 and left `checks` one line below. The public index states a fact the case file refutes. Remedy: set `checks` to the run date in the form the other fields use (for example `"run 2026-10-22"`).
- **C2 · OWED · `tests/corpus/cases/PV-016/check.sh:8-9,22`.** The check writes two binaries to `███████████████████████████` under the repository root. `.gitignore` does not cover `████████` (read: it lists only `.DS_Store` and `.env*`). A judge who runs the documented command in a clean clone is left with untracked binaries that a `git add .` would commit. Remedy: build into a `mktemp -d` directory removed on exit, or add the path to `.gitignore`. The second option touches a file outside the case's lane and is the owner's call.
- **C3 · OWED · `tests/corpus/cases/PV-016/check.sh:28-33`.** On the change version, a claim whose quote is not found in its page should fail the check (status 1), not print `(no such page)`. On base, `(no such page)` is the correct answer. Remedy: when `version=change` and the quote is not found, write to stderr and set `status=1`. Trigger: any edit to the change pages that drifts from claims.tsv.
- **N1 · NOTE · `base/alerts.go:85`.** `reason` is printed whether or not the alert was delivered (see Unplug). It is harmless for the case, because EXPECTED keys on `delivered`. Worth one sentence in reference.md so an evaluator does not accept a reason-only citation as proof.
- **N2 · NOTE · `base/alerts.go:86-89`.** `Delivered` is counted from `Send`'s return value, not from the in-memory receiver (`rcv.Got`). §2 says delivered "to an in-memory receiver". It is equivalent by construction. Counting `len(rcv.Got)` would tie the number to the receiver.
- **N3 · NOTE · `base/alerts.go:94-101`.** `FieldsOf` reads the catalogue, not the fields of an emitted alert, which is what §2 operation 3 says. `Emit` builds `Fields` from the same catalogue, so the answer is the same. The check also reports only claimed-but-absent fields, never emitted-but-unclaimed ones. No such field exists here.
- **N4 · NOTE · `check.sh:26`.** `IFS=$'\t'` with `read` collapses consecutive tabs, because tab is IFS whitespace. An empty column in claims.tsv would shift every later field. Today every row fills `claimed` with `-`, so it holds. It is fragile for whoever edits the table.
- **N5 · NOTE · `what-goes-out.md:13`, `chat-setup.md:7`.** "Alerts reach the chat within a minute" and "Alerts are delivered from then on" are code-facing promises with no claims.tsv row. Both are covered in substance by the P-* rows.

Item 6: the check runs where a judge will run it. Read, not run on macOS.
- Non-portable or environment-bound calls: `go -C` (`check.sh:22`) needs Go ≥ 1.20. `go 1.26` in `go.mod` makes an older Go (1.21–1.25) download go1.26 over the network under the default `GOTOOLCHAIN=auto`, or refuse under `GOTOOLCHAIN=local`. That conflicts with "no network" in §2 unless the judge already has Go ≥ 1.26. §8 names go1.26.5, so the requirement is stated, but the network side effect is not. Class NOTE (N6).
- Shell constructs (`${claimed//,/ }`, `$'\t'`, `${absent:+,}`, `${absent:-none}`, `case` globbing) are all valid in macOS's bash 3.2. `diff -q`, `grep -qF --` and `mkdir -p` behave the same on BSD and GNU. `#!/usr/bin/env bash` is portable. I found no GNU-only flag.

### Profile
- docs + test-fixture. No authentication, payments, data access, configuration or destructive path in product code, so the security profile does not apply (R9). The fixture's own "credential" is synthetic by contract.
- R4 dependency delta: two new `go.mod` files, standard library only, no `require` lines, no `go.sum`. No install-time scripts.
- R13 paired paths: base/ and change/ code are identical, checked by `diff -r` and also enforced by `check.sh:12-17`.

### Focus
The order's six properties: 1 HOLDS (by probe; check.sh unrun) · 2 HOLDS · 3 HOLDS with C3 · 4 HOLDS · 5 HOLDS · 6 HOLDS by reading, with N6. Plus C1 (manifest status) and C2 (untracked build output), found outside the six items.

Declined to judge (R11): licences (out of scope by the order); whether the case's story faithfully reflects the private record in §3 and §6 (not visible from this diff); the other review of this diff (not visible to me).

Discrepancy in the order, recorded and not acted on silently: `_REVIEW_ORDER.md:62` asks for the H1 `# REVIEW — corpus-pv-016-review-claude`. The engine's channel grammar for this run asks for `# REVIEW — corpus-pv-016` and says to trust it over the brief. I used the engine's form so the verdict binds to the reviewed stream.

### Conditions
- C1 · OWED · `tests/corpus/manifest.json:291`: set `checks` to the recorded run, so the index agrees with CASE.md line 5 and §8.
- C2 · OWED · `tests/corpus/cases/PV-016/check.sh:8-9`: build into a temporary directory removed on exit (or the owner adds `████████` to `.gitignore`), so a judge's clean clone stays clean.
- C3 · OWED · `tests/corpus/cases/PV-016/check.sh:28-33`: on change, a quote not found in its page fails the check with status 1.
- Master to re-measure before the gate: `bash tests/corpus/cases/PV-016/check.sh` and its exit code, compared line by line with CASE.md §8 (see Unrun).
- NOTES N1–N6 are judgement and do not block (R6).

### Unrun
- `bash tests/corpus/cases/PV-016/check.sh`: refused by the hook. Quoted: `the program "bash" is NOT in the ██████████████ list of role review under permission class ██████████████████████████, which is [saphan doc check go git ls cat grep find wc head tail diff pwd cut sort] — deny by default … do not rebuild the act from other commands, hand it to the role that holds it`. So I did not paste the check's own output and did not measure its exit code. I did not rebuild the script from other commands (law #57). My `go run` probes test the tool's three operations directly. They are not a substitute for the check's claims.tsv walk or its quote matching. Returned to the master.
- Probes of `emit backup_late|restore_done|storage_low` and `fields backup_late|restore_done|delivery_failed`: not run, to keep the probe set small. Read from the code instead.
- macOS run of check.sh: no macOS host in this leg. Portability is judged by reading only.
- The `██████████████████████████████████████████████████` control directory, made where the order confines artifacts, makes `git status --porcelain` show `?? ████████`. My class has no `rm`, so I could not remove it. Teardown is the master's.

## The review status as filed

## REVIEW — corpus-pv-016
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-016 at 5af73bc reproduces §2: base and change differ only by docs/; go probes match §8 (delivered=0, t- refused, retry_count and threshold absent); a send=true mutant prints delivered=1, so the reading separates the worlds.
  Lane and privacy hold. Three OWED items: manifest.json:291 still says checks "not run"; check.sh leaves untracked binaries in █████████████; on change a drifted quote prints "(no such page)" with exit 0.
  Unrun: bash check.sh, refused by the review permission class; the master re-measures it before the gate.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T19:50:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 8 |
| e-mail or user name | 1 |
| policy binding name | 2 |
