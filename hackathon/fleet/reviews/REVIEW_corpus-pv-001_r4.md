# REVIEW — corpus-pv-001 · round 4

- verdict: accept-with-conditions
- reviewer: claude-code, claude-opus-5-5
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 1 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:59:17Z

## Summary

PV-001 fixture traces true to §2: base and change differ only on the six order positions, and change and the reference differ only on the duplicate attribute (static trace; check.sh UNRUN, bash denied to the review role).

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-001

### Ground

- `git rev-parse --short=7 HEAD` printed `cec11bd`. Role lock holds.
- Scope: `git diff fe146e0...cec11bd`. It is one commit. It touches 10 files: `tests/corpus/cases/PV-001/{CASE.md, base/guard.ts, change/guard.ts, check.sh, check.ts, package-lock.json, package.json, reference/read.ts, tsconfig.json}` and `tests/corpus/manifest.json` (+2/-2, PV-001's `status.fixture` and `status.checks` only).
- Contract read: `CASE.md` §2 (inputs and roles), §5 (what the reviewer under evaluation receives), §8 (recorded output). Answer key read after the static trace: `expected/EXPECTED.md`.
- Craft: `██████████████████████████` R1–R15. The author's own claims in §8 are treated as claims, not evidence (R1).
- **The check was NOT run by this leg** (see Unrun). Every statement below about the check's output is a static trace of the code against the recorded §8 output, not a measurement.

### Differential

Q1 — does the fixture reproduce §2's condition? (static trace; the run is UNRUN)

Recorded output in `CASE.md:98-112`, as committed (not re-measured):

```
comment-without-frame base=pass change=fail missing=home reference=fail missing=home
comment-with-frame base=pass change=pass reference=pass
script-without-frame base=pass change=fail missing=home reference=fail missing=home
script-with-frame base=pass change=pass reference=pass
style-without-frame base=pass change=fail missing=home reference=fail missing=home
style-with-frame base=pass change=pass reference=pass
attribute-value-without-frame base=pass change=fail missing=home reference=fail missing=home
attribute-value-with-frame base=pass change=pass reference=pass
cdata-without-frame base=pass change=fail missing=home reference=fail missing=home
cdata-with-frame base=pass change=pass reference=pass
split-line-without-frame base=pass change=fail missing=home reference=fail missing=home
split-line-with-frame base=pass change=pass reference=pass
duplicate-attribute base=pass change=pass reference=fail missing=home
```

My trace of each input through the code agrees with every line:
- base (`base/guard.ts:5`, regex `data-screen\s*=\s*"…"`) matches every input. Each holds `data-screen="home"`, and the split-line input matches through `\s*`. So base=pass on all 13 lines.
- change (`change/guard.ts`): comment skipped at :44-47; `<!CDATA…` skipped to the first `>` at :49-52 (the same bogus-comment cut the HTML tokenizer makes); script and style bodies skipped by `skipRawText` at :154-156; the attribute-value case keeps keys `alt` and `src` only; the split-line text is not a tag. So it fails on the six without-frame inputs and passes on the six with-frame inputs. On the duplicate, `attrs.set` at :151 overwrites `other` with `home`, so change=pass.
- reference (`reference/read.ts`, parse5 8.0.1): a comment, raw text, an attribute value, a CDATA-as-bogus-comment in HTML content and plain text create no `data-screen` attribute. On the duplicate, the tokenizer drops the second attribute, so the reference reads `other` and the line shows `fail missing=home`.

Where base and change differ: exactly the six "without-frame" inputs that §1 names as the old guard's weakness. They agree everywhere else, including the duplicate. Where change and reference differ: only `duplicate-attribute`, which is the one input §2 adds beyond the order's list. That is the disagreement §1 and §4 name, and nowhere else. **Q1 holds by static trace.** It is not confirmed by a run.

Q2 — does the expected outcome follow from the printed evidence? (`expected/EXPECTED.md:3-7`)
- The outcome is "established defect in `change/`". The line `duplicate-attribute … change=pass reference=fail` separates the outcomes. If a conformant parser kept the LAST duplicate (the opposite world), the reference would print `pass`, and the line would read change=pass reference=pass. The with-frame lines show all three agree on a real frame, so the guard is not simply broken (EXPECTED.md:7).
- **Gap (F2):** EXPECTED.md:6 asks for "reference = first value". The check never prints a value. It prints only `fail missing=home`. That line reads the same in a world where the reference found `other` (first value) and in a world where it found no `data-screen` at all, for example if parse5 had dropped the element or both attributes. So the printed output does not establish "first value". It establishes "not `home`".
- **Gap (F3):** `reference/read.ts:13-14` `break`s after the first `data-screen` attribute. If parse5 kept both attributes, the check's own `break` would still produce "first wins". So "first wins" is partly written by the check author, not taken from the parser that §2 (line 29) requires. If parse5 drops the duplicate, as the standard requires, the `break` is a no-op, so the recorded result is not wrong. But the output cannot tell the parser's rule apart from the check's rule.
- Name drift (NOTE, outside this diff): EXPECTED.md:5 names `checkFrames`, while the code's operation is `judge` (`change/guard.ts:162`). `expected/` is out of this diff's lane, so it is not owed here.

Q3 — was a check bent? I found no input or required id tuned to make a line come out as expected. `REQUIRED = ["home"]` is constant (check.ts:12). The inputs map one-to-one to §2:35-37. The only hand-written rule on the reference side is the `break` (F3). Separately, **F1 (BLOCKING)**: the review target tells the reviewer the answer, which defeats what the case measures (see Focus).

Q4 — privacy: I grepped the diff for e-mail shapes, `██████`, `███████`, internal host or seat names, `saphan`, `bridge`, private IP ranges and `token`. The one hit is the word "tokenize" (`change/guard.ts:1`). Lockfile `resolved` URLs are all `https://registry.npmjs.org/`. **Clean.**

Q5 — lane: only `tests/corpus/cases/PV-001/` outside `expected/`, plus PV-001's own `status` lines in `manifest.json:26-27`. `expected/` is untouched (absent from `--stat`). The master's lane claim holds. Their `check.sh exit 0` claim is UNRUN here.

Q6 — portability (static):
- `check.sh`: `#!/usr/bin/env bash`, `set -euo pipefail`, `cd "$(dirname "$0")" && pwd`, `[ -f ]`. All portable to Linux and macOS bash 3.2. The file mode is 100644, so it must be called as `bash check.sh`, which §8 does.
- The lockfile carries the `@typescript/typescript-darwin-{arm64,x64}` and `linux-*` optional binaries (package-lock.json:34-67, 102-236). No `hasInstallScript`. `npm ci` should resolve on both. This is not run on macOS.
- Non-portable, documentation only: `CASE.md:79` cites `ls -l --full-time`, which is GNU-only and not in BSD `ls` on macOS. It is the author's timestamp instrument, not part of the check.
- Node floor: entities 8.1.0 needs node `>=20.19.0` (package-lock.json:362-364). §8 records v24.5.0. No `engines` entry in `package.json` states the floor (NOTE).

### Unplug

Static mutations (R3). None was executed; the toolchain is denied to this role.
- M1: make change keep the first duplicate (`if (!attrs.has(attrName)) attrs.set(…)` at change/guard.ts:151). Expected: `duplicate-attribute` becomes change=fail, matching the reference. The case's one discriminating line flips. **Live discriminator.**
- M2: drop `skipRawText` (change/guard.ts:154-156). Expected: `script-without-frame` and `style-without-frame` flip to change=pass. **Live.**
- M3: remove the `break` in reference/read.ts:14. Expected: no change IF parse5 drops the duplicate. This is the instrument to settle F3. If the line flips, the recorded reference reading came from the check, not the parser.
- M4: unplug the reference (return `[]`). Expected: every with-frame line becomes reference=fail. **Caught.** The duplicate line does not change, which is F2's false-world reading.

### Closure

- Contract §2 roles: base (text pattern) ✓, change (tokenize → map, later wins) ✓, reference from a spec-conformant parser ✓ (with the F3 caveat), check printing name/base/change/reference per input ✓.
- Inputs: six order positions × with/without frame ✓, plus the duplicate with a non-required first value and a required second value ✓.
- Pins: typescript 7.0.2 and parse5 8.0.1, exact; lockfile v3 with integrity hashes ✓. Dependency delta (R4): parse5 8.0.1 (MIT) → entities 8.1.0 (BSD-2-Clause); typescript 7.0.2 (Apache-2.0, dev) plus 20 platform-gated optional binaries. No install scripts.
- §8 records toolchain versions and commands ✓, with a clock taken from a build artefact mtime (accepted as stated, not re-measured).

### Remainder

- Owed to the owner (R11, declined to judge): `AGENTS.md` §3 says "All project code is written between 22 and 27 October 2026. Nothing written before 22 October may be copied in." `CASE.md:5` and `:79` date this case code 2026-10-21, "declared prior art". The order says the commit was carried onto main on 22 October. Whether a declared prior-art corpus fixture falls under that rule is the owner's ruling. This review neither clears nor blocks it.
- `manifest.json:26` drops the earlier "(prior art, built before the window)" marker and now reads just `"built"`. The machine-readable record no longer carries the disclosure that `CASE.md:5` still makes (NOTE).
- The repo's `.gitignore` does not cover `node_modules/` or `dist/`. Running §8's two commands leaves `tests/corpus/cases/PV-001/node_modules/` and `dist/` as untracked trees in a judge's checkout (OWED).
- The change guard does not skip RCDATA bodies (`<textarea>`, `<title>`). A `data-screen` inside them would count where a browser would not. This is outside §2's inputs and the case's thesis. It is noted, not owed.

### Profile

`docs` plus test-fixture review (D1, D4). This is not the security profile. The diff touches no authentication, payments, data access, configuration or destructive path. It is a self-contained offline test case.

### Focus

**F1 — BLOCKING — the review target states the answer.** `change/guard.ts:1-3`: "A later duplicate attribute overwrites an earlier one. The HTML standard keeps the first duplicate; this map does not." `change/guard.ts:150`: "Later duplicate wins. This is the property the case exists to show." `CASE.md:56-59` (§5) says the reviewer under evaluation receives "the two versions of the guard and a readable diff between them" and never `expected/`. `manifest.json:12` says the answer key is withheld. These comments ARE the withheld answer, shipped inside the review target. Any reviewer that reads them is handed the established defect. So the case can no longer separate a reviewer who finds the defect from one who reads it, which is the thing it exists to measure. `base/guard.ts:1-3` likewise lists the regex's blind spots, which is milder. Trigger: hand `base/` and `change/` to a reviewer as §5 prescribes. Remedy: rewrite the comments in `base/guard.ts` and `change/guard.ts` to describe what the code does, never what is wrong with it or what the case shows. Keep the case's self-description in `CASE.md` and `check.ts`, which §5 already withholds.

**F2 — OWED — the check does not print the value EXPECTED.md requires.** `check.ts:49-53, 56-65` print only pass/fail and missing ids. Remedy: also print each side's frame-id list (for example `base-ids=… change-ids=… reference-ids=…`), so the duplicate line shows `reference-ids=other`, and re-record §8.

**F3 — OWED — the reference walker encodes first-wins itself.** `reference/read.ts:13-14` (`break`). Remedy: remove the `break` so every `data-screen` the parser keeps is reported, and let the parser alone decide. Then M3 is the control, and F2's printed id list shows whether parse5 kept one attribute or two.

### Conditions

1. **C1 · BLOCKING · F1** — strip the answer-revealing comments from `tests/corpus/cases/PV-001/change/guard.ts:1-3,150` and `base/guard.ts:1-3`. Control: `grep -nE 'standard|case exists|keeps the first|also matches' tests/corpus/cases/PV-001/base/guard.ts tests/corpus/cases/PV-001/change/guard.ts` prints nothing.
2. **C2 · BLOCKING · Unrun** — a role that may run node runs `npm ci --prefix tests/corpus/cases/PV-001 && bash tests/corpus/cases/PV-001/check.sh` at the tip and pastes the output. It must match `CASE.md:98-112` line for line, or §8 is re-recorded from that run. This review's Q1 rests on a static trace only.
3. **C3 · OWED · F2** — print the frame ids each side read, and re-record §8.
4. **C4 · OWED · F3** — remove the `break` in `reference/read.ts:14`, and record the reference ids for the duplicate input.
5. **C5 · OWED** — ignore `tests/corpus/cases/*/node_modules/` and `tests/corpus/cases/*/dist/`. This touches `.gitignore`, which is outside this diff's lane, so the master routes it.
6. **C6 · OWNER** — rule on prior-art case code against `AGENTS.md` §3's window rule (Remainder).
7. NOTE — `manifest.json:26` loses the prior-art marker. EXPECTED.md:5 `checkFrames` ≠ code's `judge`. No `engines` floor in `package.json`. `CASE.md:79` uses GNU-only `ls --full-time`.

### Unrun

- `bash tests/corpus/cases/PV-001/check.sh`: REFUSED. Quoted: `the program "bash" is NOT in the ██████████████ list of role review under permission class ██████████████████████████, which is [saphan doc check go git ls cat grep find wc head tail diff pwd cut sort] — deny by default. … allowed form: none — "bash" is denied …; do not rebuild the act from other commands, hand it to the role that holds it`. It was not rebuilt from `node`/`npm` (law #57). Q1's output is therefore NOT pasted from a run of mine. → C2.
- `npm ci --prefix tests/corpus/cases/PV-001`: not attempted; `npm` is outside the same allow list.
- Mutations M1–M4: not executed, for the same reason.
- macOS run: no macOS host in this leg.
- `git check-ignore`: needed approval and was not granted. I read `.gitignore` directly instead.

## The review status as filed

## REVIEW — corpus-pv-001
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-001 fixture traces true to §2: base and change differ only on the six order positions, and change and the reference differ only on the duplicate attribute (static trace; check.sh UNRUN, bash denied to the review role).
  BLOCKING C1: change/guard.ts:1-3,150 and base/guard.ts:1-3 comments state the withheld answer inside the review target (§5). BLOCKING C2: a run-capable role runs check.sh and pastes its output.
  OWED: print frame ids (EXPECTED needs "first value"), drop the reference/read.ts:14 break, gitignore node_modules/dist. OWNER: prior-art code vs the AGENTS.md 22-27 Oct window.
  Lane and privacy clean. The master's lane claim holds; their exit-0 claim is unmeasured here.
- artifacts:
    - _REVIEW_NOTES.md
    - _REVIEW_STATUS.md
- updated: 2026-10-22T19:20:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 2 |
| policy binding name | 2 |
| private workspace path | 1 |
