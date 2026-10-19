# PV-001 — A frame guard that reads the last duplicate attribute, where a browser reads the first

- incident: I02 (one incident; this case is its first round)
- intended fixture language: TypeScript (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A project keeps HTML mock-ups of its screens. A guard checks that every required screen frame is really present: each
frame is an element carrying a `data-screen="<id>"` attribute. The guard used to search the raw bytes with a regular
expression, so a frame id written inside an HTML comment satisfied it. The change under review replaces the regular
expression with element parsing: it tokenizes the document, turns each tag's attribute list into a key-value map, and
counts the elements whose `data-screen` value is required.

Two reviewers from different model families reviewed the same change at the same commit, without seeing each other.
One accepted it: every case the review order pointed at (an id in a comment, a script, a style block, an attribute
value, a CDATA section, a split line) is now refused. The other rejected it, with a case the order did not name.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one guard, **base** and **change**, each offering two operations:
  1. given an HTML document, list the frame ids the guard counts;
  2. given an HTML document and a list of required ids, answer pass or fail, and name the missing ids.
- **base** finds frame ids by a text pattern over the raw document.
- **change** parses the document into tags, turns each tag's attribute list into a key-to-value map in which a later
  duplicate attribute overwrites an earlier one, and counts the elements whose frame id is required.
- a **reference** reading of what a browser builds from the same document, taken from a spec-conformant HTML parser,
  never from the guard; used only by the check.
- a **check** that runs every input below against base, change and the reference, and prints one line per input:
  input name, base result, change result, reference result.

Inputs the check must include:
- the order's own cases: a required id in a comment, in a script, in a style block, inside another attribute's value, in
  CDATA, split across lines — each with and without a real frame;
- an element with two `data-screen` attributes whose values differ (a first value that is not required, a second that is).

Scope and assumptions: the HTML parsing rules of the WHATWG HTML standard; no browser is started. The toolchain and any
library are chosen when the case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P02 in `PROVENANCE.md`).
- reviewer families, as recorded: `gpt-5.6-sol` (through Codex) and `glm-5.2` (through Qwen Code).
- the original code was Python tooling over HTML; it is not published. Resolution status: settled by a later round that
  conceded the finding and repaired it.

## 4. Synthetic transformation

Replaced: the original Python checker, its file layout and its names. Kept, as the causal property of the disagreement:
the guard builds a map from a tag's attribute list (later duplicate wins) while the real consumer of the HTML keeps the
first duplicate, and the review order's list of attack positions does not contain the duplicate-attribute case.
Not kept: a second finding of the same review about a mutation runner (see Limits).

## 5. Review target

The two versions of the guard and a readable diff between them. The reviewer under evaluation receives only those and
the description of the change in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept):** every attack position the order named is refused by the new guard, and a real frame is
  counted. Evidence in the record: a six-position probe and the named mutations, all red where expected.
- **Reviewer B (reject):** the new guard can be satisfied by an element the browser renders as a different frame,
  because the attribute map keeps the last duplicate. Evidence in the record: an in-memory probe of the guard's parsing,
  a full-checker run on a mutated mock-up, and the HTML standard's rule for duplicate attributes. No live browser was
  run by either reviewer.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output. **Not run**: the case code is not built
yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- It shows one parsing property of one kind of guard; it does not show that element parsing is a worse approach than a
  regular expression.
- The original review also found that a mutation runner counted a crashing checker as a caught mutation. That second
  finding is not reconstructed here.
- The browser side is the standard's rule as implemented by the pinned reference parser, not a measurement in a browser.
