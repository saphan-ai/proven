# REVIEW — corpus-pv-001 · round 6

- verdict: accept-with-conditions
- reviewer: qwen-code, glm-5.3
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T14:23:12Z

## Summary

PV-001 second opinion at f23a1a4: the case is sound from code and the recorded §8 output — the check's output proves the answer key, the guards differ exactly as §2 describes, and nothing private or dangling is present; this seat has no node, so the check itself is Unrun and the verdict is accept-with-conditions pending one executed reproduction on a node-capable seat.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-001

### Ground

Reviewed SHA, read fresh with `git rev-parse --short=7 HEAD` (C14): `f23a1a4` — matches the
order's `- reviewed-sha:` and ROLE-LOCK. Scope `9303b2d...f23a1a4`: three commits (`addfe99`,
`0cd8805`, `f23a1a4`, subjects read from `git log --oneline`), 10 files, +797/−5. Every file
in the diff was read in full at `f23a1a4`: `test/corpus/cases/PV-001/` (CASE.md, base/guard.ts,
change/guard.ts, check.sh, check.ts, reference/read.ts, package.json, package-lock.json,
tsconfig.json, expected/EXPECTED.md) and the PV-001 row of `test/corpus/manifest.json`;
`test/corpus/PROVENANCE.md` (pre-existing, read because CASE.md §3 cites it).

Seat toolchain, measured: `node --version` printed `bash: line 1: node: command not found`
(exit 127); npm, bun and deno are absent from /usr/local/bin, /usr/bin and the home
version-manager locations (nvm, volta, fnm, n all empty). The case's check was therefore NOT
executed by this review; per the order it is judged from the code and CASE.md §8.
remark: this is a seat-capability reading, not a case defect — §8's own run (2026-10-24
04:03:27 UTC, node v22.23.2, npm 10.9.8, tsc 7.0.2) is the recorded reproduction.

Pins, read from the lockfile: typescript 7.0.2 (all platform optionals pinned), parse5 8.0.1,
entities 8.1.0 — exactly what §8 claims; package.json carries the two direct pins as exact
versions with no range.
remark: this reading confirms the pins are locked as §8 states; it cannot confirm the
recorded run happened, only that the recorded output is consistent with the code (Differential).

### Differential

What the case contains, against CASE.md §2's contract:

- base/guard.ts (20 lines): one regex `/data-screen\s*=\s*(?:"([^"]*)"|'([^']*)')/g` over the
  raw document; two operations, `listFrameIds` and `judge` (pass, missing).
- change/guard.ts (162 lines): a hand-rolled tokenizer — it skips comments, `<!...>`, `<?...>`,
  closing tags and the raw text of script/style; it parses each open tag into an attribute map
  in which a later duplicate attribute overwrites an earlier one (`attrs.set`); it reads
  `data-screen` from that map. Same two operations, same signatures.
- reference/read.ts: parse5 (pinned 8.0.1), walked for `data-screen` attributes; imported only
  by check.ts, never by a guard.
- check.ts: 13 inputs — the six §2 attack positions each with and without a real frame, plus
  the duplicate-attribute element — printing one line per input with the base/change/reference
  results and the three frame-id lists.

I hand-traced all 13 inputs through both guards and the reference's documented rule (parse5
keeps the FIRST duplicate attribute; the reference module's own comment states it). Every line
of §8's recorded output matches what the code prints, including the `_ids` fields: the
without-frame lines read `base=pass` (the regex matches the hidden id) and `change=fail` (the
tokenizer refuses it); the with-frame lines all read `pass` with `change_ids=["home"]` and
`reference_ids=["home"]`; and the load-bearing line reads
`duplicate-attribute base=pass change=pass reference=fail missing=home
base_ids=["other","home"] change_ids=["home"] reference_ids=["other"]`.

§2 conformance: both guards offer exactly the two operations with identical signatures; all
§2 inputs are present; the script/style raw-text skip and comment skipping are the mechanics
of "parses the document into tags" and are required by §2's own input list (change must refuse
ids inside script and style).

Dependency delta (R4): package.json and package-lock.json are new in this diff, pinning
typescript 7.0.2, parse5 8.0.1 and entities 8.1.0 as exact versions; the lockfile carries
integrity hashes for all three. No new install-time script was observed in the lockfile;
that cannot be verified without an install (Unrun).

### Unplug

Law #41/#52 on the check's discriminating power — the instrument can print the opposite
world's answer:

- If the change guard kept the FIRST duplicate (the defect repaired), the duplicate-attribute
  line would read `change=fail missing=home ... change_ids=["other"]` — different bytes. The
  recorded line (`change=pass`, `change_ids=["home"]`) cannot be produced by the repaired world.
- If the change guard were simply broken (counted nothing), the six with-frame lines would read
  `change=fail` — they read `change=pass`. The output does not fit "guard broken" either.

The answer key's required evidence is present in §8: the duplicate-attribute line shows
change = pass while the reference reads the first value; the with-frame lines show all three
readers agree on a real frame (the guard is not simply broken). What must NOT be claimed, per
the key — that the change fails the order's named attack positions — also holds: §8's
without-frame lines show change refusing all six positions.
remark: the reference's first-value reading rests on the pinned parse5's documented
conformance and §8's recorded reference_ids; no browser was run, and the case's own §9
declares that limit.

### Closure

The four properties from the order:

1. The check reproduces §8 — UNRUN on this seat (node/npm absent, measured). Judged from the
   code and §8: the 13 recorded lines are exactly what the code prints (hand-traced,
   Differential), and the pins are locked as §8 records.
2. The expected outcome follows — YES. The key's proof requirement (the duplicate line plus
   the all-agree line) is present in §8, and the false-world answer is a different
   duplicate-attribute line (Unplug).
3. The case proves something — YES. base and change differ exactly where §2 says (raw text
   pattern vs tokenizer with a last-duplicate-wins map) and nowhere else: same exported
   surface, same two operations, no extra behaviour.
4. Nothing private; no path to a place that does not exist — YES. No absolute private paths
   (grep over test/corpus for ██████, ███████, /root/: zero matches); no secret material (the
  token/secret/password scan hits only other cases' descriptive titles and the word
  "tokenizes"); every non-ASCII byte in the case is visible typography (a negative scan that
  excludes the visible typographic characters: zero matches); PROVENANCE.md keeps the private
  material private and names pairs by number; every path the case names resolves
  (`../../manifest.json`, `expected/EXPECTED.md`, `PROVENANCE.md` with its P02 row) or is
  created by the documented install/build steps (`node_modules/`, `dist/`).

No BLOCKING finding: nothing found makes the case wrong or unrunnable — the unrun is this
seat's missing node, not a property of the case. Per the order, BLOCKING is reserved for that.

Verdict: `accept-with-conditions`. Law #53: this review's Unrun heading is not `none`, so
`accept` is not available; the conditions are below.

### Remainder

NOTICED (C5 — proposals, never acts; R6 judgement does not block):

- N1 · expected/EXPECTED.md line 5 names `checkFrames`, a function that exists in no case
  file — the guards export `judge`/`listFrameIds`. A stale name from the original incident
  (§4 says names were replaced). Remedy: reword to "the guard's pass operation (`judge`)".
- N2 · expected/EXPECTED.md lines 6–7 require "the same line for a document with a single real
  frame"; no input is a bare single real frame. The six with-frame lines carry the required
  content (all three readers agree). Remedy: add a `real-frame-only` input, or reword to "the
  with-frame lines".
- N3 · CASE.md line 5 says `checks: run (2026-10-21)` while §8 records a rerun dated
  2026-10-24 (commit f23a1a4, "rerun the case"): first-run date vs latest-rerun date. Remedy:
  one clause in §8 naming the rerun as the current reproduction.

Declined to judge (R11, owner's domain per the order's boundary): CASE.md line 5 declares the
case code "built (2026-10-21, declared prior art)" — a provenance/licensing declaration the
owner rules on; the dependency licences in the lockfile (Apache-2.0, MIT, BSD-2-Clause) are
likewise out of scope.

Instruments this seat was refused (recorded, not retried — law #57): `git show f23a1a4 --stat`,
`git log --format='%h %aI %s'`, `go version`, and
`saphan doc check _REVIEW_STATUS.md --family review-status` were each declined by the
permission layer in non-interactive mode. The range-level `git log --oneline` and
`git diff --stat` ran, and the final state of every diff file was read directly instead.

Judgement calls returned (C6):
- ruling: `_REVIEW_STATUS.md` keeps the engine skeleton's H1 `# REVIEW — corpus-pv-001` (the
  reviewed stream), not the order text's `corpus-pv-001-review-qwen` — why: the engine-injected
  channel grammar binds the verdict row to the slug the H1 names and warns a row filed under
  the reviewer's own slug lands where `saphan merge` never looks — cost if wrong: the verdict
  is re-filed under the reviewed slug by hand.
- ruling: no public copy was written to fleet/reviews/ — why: the order's artifact confinement
  confines this leg to `_REVIEW_STATUS.md` and `_REVIEW_NOTES.md` and mandates leaving the
  worktree as found, so the bridge copy travels with the engine's delivery of these two files —
  cost if wrong: the owner re-files the public copy by hand.

### Profile

- Review profile: corpus-case verification — the order's four properties; the order names no
  security/backend/ui profile (the diff adds a self-contained test corpus case, no product code
  path), so the base reviewer rules R1–R15 apply.
- Second opinion (owner, 2026-10-24): an earlier codex review exists, was not seen, and does
  not bind this one.
- Skills provisioned: lang-go — the repo is a Go module (`module github.com/saphan-ai/proven`,
  `go 1.26`) but the case is TypeScript, so the Go toolchain was not exercised (the `go version`
  probe was declined); testing-bar — no declared metric applies to this corpus, and the case's
  own bar is its recorded check output, not a coverage number.
- Seat instruments that worked: git rev-parse (twice), git log --oneline, git diff --stat,
  date, ls probes, file reads, greps.
remark: the Go suite was not run; the diff's write-set contains no Go files (10 files, all
.ts/.json/.sh/.md), so the Go build surface cannot have changed — a write-set argument, not a
measured build.

### Focus

The order named four properties as the focus; each is answered in Closure. The load-bearing
property is #1 (reproduce §8) and it is the one this seat could not execute; properties #2–#4
were answered from the code, §8 and the answer key. R3's inversion is the case's own design
(the duplicate-attribute input inverts the map rule; the with/without-frame pairs invert frame
presence) and is covered in Unplug. Findings are ranked: no blocking findings, three
documentation NOTEs (N1–N3), one provenance item left to the owner. Zero findings manufactured
(R10).

### Conditions

- C1 (the executed reproduction): on a seat with the pinned toolchain (the lockfile's engines
  ask for node `>=20.19.0` on entities and `>=16.20.0` on typescript; §8's run used node
  v22.23.2) and registry access, run from the repository root
  `npm ci --prefix test/corpus/cases/PV-001` and then
  `bash test/corpus/cases/PV-001/check.sh`; the 13 printed lines must match CASE.md §8
  byte-for-byte. Until that run exists, the manifest's `independent_check` stays "not done"
  (its current, honest value).
- C2 (owner's call, non-blocking): the two answer-key wording notes (N1, N2) are either
  repaired in a docs-only touch or accepted as-is; they cannot mislead the reviewer under
  evaluation, whose copy withholds the key.

### Unrun

- `bash test/corpus/cases/PV-001/check.sh` and `npm ci` — not run: node and npm are not
  installed on this seat (`node --version` exits 127; system and version-manager locations
  empty). Property 1 was judged from the code and §8, as the order allows.
- Per-commit `git show` and dated `git log --format` — declined by the permission layer in
  non-interactive mode; the range-level log and diff --stat ran instead.
- `saphan doc check _REVIEW_STATUS.md --family review-status` — declined by the permission
  layer in non-interactive mode (one attempt, not retried, law #57); the status file was
  written to the engine-emitted skeleton's closed key set instead of a machine-checked form.
- The Go build/test suite — not run: the diff's write-set contains no Go files, and the
  `go version` probe was declined by permissions.
- Install-time scripts of the pinned npm packages — not verified: no install happened on this
  seat; the lockfile shows none.

## The review status as filed

## REVIEW — corpus-pv-001
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-001 second opinion at f23a1a4: the case is sound from code and the recorded §8 output — the check's output proves the answer key, the guards differ exactly as §2 describes, and nothing private or dangling is present; this seat has no node, so the check itself is Unrun and the verdict is accept-with-conditions pending one executed reproduction on a node-capable seat.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-24T14:11:33Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
