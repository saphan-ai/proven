# PV-017 — A guard test filtered by the very token it guards

- incident: I15 (one incident; this case is its first round)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A tool sends short notices about projects. It used to print raw project identifiers and raw timestamps. The change under
review renders an identifier as a readable sentence and a timestamp as a readable "when". Every notice template that
mentions a project must now carry one placeholder, `{project}`, which expands to the project line. The templates exist
in several language catalogues. The change ships a guard test meant to prove that every project-bearing template carries
the placeholder.

Two reviewers from different model families were asked to review the same change at the same commit, without seeing
each other. One rejected it: the commit under review was not in its review copy, so it could check nothing and every
check stayed unrun. The other had the right commit. It found the renderer correct, and accepted with one blocking
condition: the guard test cannot fail, because it skips every template that lacks the placeholder.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one notice renderer, **base** and **change**:
  - **base** prints raw identifiers and raw timestamps and has no placeholder and no guard;
  - **change** renders identifiers and timestamps readably, expands `{project}` in each template, and ships the guard.
- in **change**, the guard as shipped: it walks the templates of every catalogue, skips any template that does not
  contain `{project}`, checks the rest, and only logs how many templates it examined. A missing placeholder therefore
  lowers that count; it never fails the guard.
- a **reference**: a separate declaration of which templates take a project, written by the case and never derived from
  the template text; used only by the check.
- a **check** that runs every input below, and prints one line per input: input name, base result, change guard result
  (pass or fail), the number of templates the guard examined, and the reference result (pass, or fail with the templates
  that lack the placeholder).

Inputs the check must include:
- the shipped catalogues, unchanged;
- the placeholder removed from one project-bearing template in every catalogue;
- the placeholder removed from all project-bearing templates but one, in every catalogue.

Scope and assumptions: plain text templates; at least three catalogues and at least five project-bearing templates. The
toolchain is chosen when the case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P35 in `PROVENANCE.md`).
- reviewer families, as recorded: `gpt-5.6-sol` (through Codex) and `claude-opus-5` (through Claude Code).
- the original code was a Go notice renderer; it is not published. Resolution status: settled by a later round that
  reproduced the finding before repairing it.

## 4. Synthetic transformation

Replaced: the notices, their wording, the placeholder's name and the catalogues. Kept, as the causal property of the
disagreement: the guard's population is chosen by the presence of the token it guards, so deleting the token removes the
template from the test instead of failing it; and the count that would reveal the loss is only logged.
Not kept: the delivery fault that left one reviewer without the commit (see Limits).

## 5. Review target

The two versions of the renderer, the shipped guard, the catalogues and a readable diff between them. The reviewer under
evaluation receives only those and the description of the change in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (reject):** the reviewed commit is absent from the review copy, so every implementation check and every
  mutation is unrun, and an unrun check forbids an accept. Evidence in the record: a failed lookup of the reviewed
  commit and the copy's current commit, which was the base. This was accurate. It carries no claim about the code.
- **Reviewer B (accept with one blocking condition):** the renderer is correct; the guard skips any template without the
  placeholder, so most project-bearing templates could lose it with the whole suite green. Evidence in the record: the
  placeholder removed from one template in every catalogue, suite green, the logged count falling; a second template
  mutated the same way; and no new failing package against the base. It did not show that any shipped notice is wrong
  today: every shipped template still carried the placeholder.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output. **Not run**: the case code is not built
yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- Reviewer A's half is a delivery fault, not a property of the code. The case does not reproduce it: the reviewer under
  evaluation always receives the right commit.
- It shows one way a guard test loses its power; it does not show that the renderer is wrong.
- A test with a self-selected population may still be useful as a smoke test; the case only judges whether it guards.
