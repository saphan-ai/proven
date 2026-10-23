# PV-002 — A guard whose test fixtures cannot express the self-closing and foreign-content cases

- incident: I02 (later rounds of the incident whose first round is PV-001; this case tells the first of them and lists
  the next as a later round)
- intended fixture language: TypeScript (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A project keeps HTML mock-ups of its screens. A guard checks that every required screen frame is really present: each
frame is an element carrying a `data-screen="<id>"` attribute. Earlier rounds moved the guard from a text pattern to
element parsing and made it read the first of two duplicate attributes, as a browser does. The change under review
teaches the guard which elements a browser does not render as part of the page: a frame inside `script`, `style` or
`template` must not count. The guard keeps an "inert" depth that an opening tag of those elements raises and their
closing tag lowers. Its tokenizer reports a tag written as `<x/>` as an opening tag immediately followed by a closing
tag. The change ships parser fixtures, and unplug controls: remove `template` from the inert set, or stop reading the
first duplicate, and named fixtures turn red.

Two reviewers from different model families reviewed the same change at the same commit, without seeing each other.
One accepted it: every fixture passes, and every unplug control turns red by name. The other rejected it: it wrote
`<template/>` in front of a real frame, and the guard still counted the frame.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- three versions of one guard, **base**, **change** and **later change**, each offering two operations:
  1. given an HTML document, list the frame ids the guard counts;
  2. given an HTML document and a list of required ids, answer pass or fail, and name the missing ids.
- **base** parses elements, reads the first duplicate attribute, and counts a frame wherever it stands, including inside
  `template`.
- **change** adds the inert depth described in §1, and takes `<x/>` as an opening tag followed by a closing tag for
  every element. It ships fixtures (none of them spells a tag as `<x/>`) and the unplug controls of §1.
- **later change** is the next round's repair: it ignores the slash on an HTML element that is not void (so `<template/>`
  opens a template), always closes void elements, and treats foreign content (`svg`, `math`) as a nesting counter inside
  which `<x/>` closes any element. It adds self-closing fixtures and foreign-content fixtures; none of them puts an HTML
  integration point (`foreignObject`, `desc`, `title`) between the `svg` and the `template`.
- a **reference** reading of what a browser builds from the same document, taken from a spec-conformant HTML parser,
  never from the guard: a frame counts only if it is in the document tree, not in template content and not in raw text.
  Used only by the check.
- a **check** that runs the shipped fixtures of change and of later change, then runs every input below against base,
  change, later change and the reference, and prints one line per input: input name, then the four results.

Inputs the check must include:
- a single real frame (all four must agree);
- a required frame inside an explicit `<template>` … `</template>`;
- `<template/>` followed by a required frame and no closing `template` tag;
- `<svg><rect/></svg>` followed by a required frame (the counter and the standard agree here);
- `<svg><foreignObject><template/>` followed by a required frame.

Scope and assumptions: the HTML parsing rules of the WHATWG HTML standard (the self-closing start tag, void elements,
foreign content and HTML integration points); no browser is started. The toolchain and any library are chosen when the
case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P07 in `PROVENANCE.md`).
- later round: pair P08, reviewing the repair, found more of the same one level deeper (the foreign-content input above).
- reviewer families, as recorded, in both pairs: `gpt-5.6-sol` (through Codex) and `glm-5.2` (through Qwen Code).
- the original code was Python tooling over HTML; it is not published. Resolution status: the first round was settled
  by a measurement taken by the orchestrating agent, then repaired; the later round has no recorded ruling (see Limits).

## 4. Synthetic transformation

Replaced: the original Python checkers, their shared module, their file layout and their names; the original change
shared the rule between two checkers, and the case keeps one guard. Kept, as the causal property of the disagreement:
the guard's tokenizer reports `<template/>` as an element that opens and closes, while the HTML tree builder ignores
that slash and keeps the template open; and the shipped fixtures contain no spelling that walks that path, so the
fixtures and their unplug controls stay green. In the later round the same property repeats: a nesting counter for
foreign content agrees with the standard on every shipped fixture, but not at an HTML integration point.
Not kept: in the first round, a missing count that the review order asked for and a change to how a mutation runner
classifies outcomes; in the later round, a claim that one unplug control cannot fail (one reviewer's measurement only).

## 5. Review target

Round 1: base, change, the change's fixtures, and a readable diff between base and change. Round 2: change, later
change, its fixtures, and a readable diff between them. Each round is evaluated on its own. The reviewer under
evaluation receives only those and the description in §1 (for round 2, also the sentence on later change in §2) —
never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

Round 1 (pair P07):
- **Reviewer A (accept):** first-duplicate reading and template inertness are verified by the author's own instruments.
  Evidence in the record: all parser fixtures green; two unplug controls that turn named fixtures red; a few edge cases
  of its own (a triple duplicate, a duplicate in other letter case, markup inside a script).
- **Reviewer B (reject):** `<template/>` in front of a real frame leaves the counted frames unchanged, while a browser
  puts the frame into inert template content. Evidence in the record: a full-checker run on a mutated mock-up (guard =
  pass), and the standard's rule for the self-closing start tag. No live browser was run.

Round 2 (pair P08), later change:
- **Reviewer A (accept):** the repaired parsing rule is verified. Evidence in the record: unplug controls that
  turn the fixtures red, eleven foreign-content boundary cases of its own, all green. None of them has an HTML integration point.
- **Reviewer B (reject):** foreign content is a counter, not namespace state, so `<svg><foreignObject><template/>`
  before a frame lets the guard count a frame a browser puts into template content. Evidence in the record: a
  full-checker run on a mutated mock-up (guard = pass), and the standard's rules for HTML integration points. No live
  browser was run.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base, change and later change, and record each line of its output. **Not run**: the case code is
not built yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- It shows how a fixture set can be unable to express a defect; it does not show that element parsing is a worse
  approach than a full HTML parser.
- The browser side is the standard's rule as implemented by the pinned reference parser, not a measurement in a browser.
  Nobody in the record ran a live browser.
- The later round has no recorded ruling: its reject stood and the work was parked. Its guard-side half was re-probed
  when this corpus was prepared; its browser-side half rests on the standard.
