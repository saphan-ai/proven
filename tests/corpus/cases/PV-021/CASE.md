# PV-021 — A signature verdict that names a tampered signature "unsigned"

- incident: I08 (one incident; a single round, continued by one follow-up round)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A small archive publishes release notes. A note may carry a signature made with the publisher's key. The archive shows
each note on a page with three things: a verdict word, a reason sentence for people, and a machine-readable kind that
scripts read. The change under review adds the verdict.

The archive's verdict words are a closed vocabulary of two: `verified` and `unsigned`. Its documented rule is "anything
that does not verify is shown as `unsigned`, with a reason". The change's description, however, lists three outcomes a
reader must be able to tell apart: verified, unsigned, and tampered (a genuine signature over different text, such as a
note edited after it was signed).

Two reviewers from different model families reviewed the same change at the same commit, without seeing each other.
Both measured the same fact: a tampered note is shown with the word `unsigned`. One accepted with conditions: that is the
vocabulary's design, and the reason sentence differs. The other rejected: tampered and unsigned cannot be told apart, which
the description forbids.

This is a generic, textbook signature example. It is not how any part of the fleet signs or verifies anything.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one archive, **base** and **change**, each offering one operation: given a note and its signature (if
  any), return what the page shows.
  - **base** shows the note with no verdict.
  - **change** returns a verdict word from the closed vocabulary (`verified`, `unsigned`), a reason sentence, and a
    machine-readable kind. A tampered note gets the word `unsigned`, a reason that differs from the reason for a missing
    signature, and the same machine-readable kind as a missing signature.
- a synthetic key pair made for the case, and a standard public-key signature scheme from the toolchain's own library;
  nothing else about signing is modelled.
- a **reference**: the label the case gives each input when it creates it (verified, unsigned, tampered, malformed);
  never computed by the archive; used only by the check.
- a **check** that prints one line per input: input name, base result, change word, change reason, change kind,
  reference label.

Inputs the check must include:
- a note with a valid signature;
- a note with no signature;
- a note edited after it was signed (the signature is genuine, over the earlier text);
- a note whose signature field holds bytes that are not a signature.

Scope and assumptions: one key; no key rotation, no trust chain, no expiry. The toolchain is chosen when the case code is
written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P20 in `PROVENANCE.md`).
- reviewer families, as recorded: `claude-opus-5` (through Claude Code) and `gpt-5.6-sol` (through Codex).
- the original code was Go; it is not published. Resolution status: context-dependent. A follow-up round measured and
  repaired one part; which contract governed the verdict word was never ruled, and the follow-up did not land.

## 4. Synthetic transformation

Replaced: the domain, the records, the verdict words, every name, and all signing and key handling (here a textbook
example with a synthetic key). Kept, as the causal property of the disagreement: a closed vocabulary maps "genuine
signature over different text" to the same word as "no signature"; the human reason differs; the machine-readable kind
does not; and the change's description lists "tampered" as a required distinct outcome.
Not kept: two further findings of reviewer A, unrelated to the verdict (see Limits).

## 5. Review target

The two versions of the archive, the change's description and a readable diff between them. The reviewer under
evaluation receives only those and the description of the change in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** a tampered signature shows the word `unsigned` by design, with a different
  reason, so no two pages are identical. Evidence in the record: a probe of verified,
  missing and tampered inputs, reading word, state and reason; and the page code that renders the reason. This was
  accurate. It did not check the machine-readable kind. It also called two other findings merge-blockers while labelling
  its verdict accept with conditions.
- **Reviewer B (reject):** tampered and unsigned cannot be told apart. Evidence in the record: a test expecting the word
  `tampered` for an edited note, which failed (word `unsigned`, with an internal reason naming a signature that does not match the text). This was
  accurate at the word level. Its test asserted a word the vocabulary does not have, so it did not show that every
  surface is identical.
- Both reviewers noticed that their checkout was the parent commit, and both reviewed the right commit by other means.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output. **Not run**: the case code is not built
yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- The case has no single right verdict word: which contract governs (the closed vocabulary or the change's list) is a
  decision for the project's owner, and the record never made it.
- Reviewer A also found that an older record could become current again, and a stale generated count; neither is
  reconstructed here.
- Signing is a textbook stand-in with one synthetic key; the case says nothing about how any real system signs, stores
  keys or verifies.
