# PV-004 — A "strict" fenced-block extractor whose refusal suite only tests closing fences

- incident: I03 (this case is the second round of the incident; the first round is a different finding, PV-003)
- language of the fixture: TypeScript
- contract written: 2026-10-19 · fixture: not built · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`

## 1. Story

A runbook written in Markdown contains one command in a fenced code block. A tool extracts that command and shows it to
an operator. The first version was lenient. The change under review makes the extractor "strict, by the CommonMark
rules": a malformed fence is refused with a named refusal class, and a test suite named for that promise ("refuses a
malformed fence by name") has eight cases.

Two reviewers from different model families reviewed the same change at the same commit. One accepted it: it ran the
suite, every case was refused by name, and it reported that the extractor refuses malformed fences by class. The other
rejected it: it wrote two inputs of its own, and the extractor accepted both.

## 2. Contract (written before any fixture)

The fixture must provide, with no network and no model:
- `extractCommand(markdown: string): { command: string } | { refusal: string }` in `base/` and `change/`.
- `base/`: the lenient extractor (first fenced block wins, no refusals).
- `change/`: the "strict" extractor with named refusal classes for unclosed fences, mismatched closing fence length and
  character, and trailing text after a closing fence; it trims each line before classifying it, and it never inspects
  the opening fence's info string.
- `change/` ships its own suite of eight refusal cases, all of which vary only the closing fence or its absence.
- `check.sh` runs the shipped suite on `change/` and then the inputs below on both versions, one line per input.

Inputs the check must include:
- the eight shipped cases;
- a backtick fence whose info string contains a backtick (CommonMark: not a fence opener);
- a fence opener indented four spaces (CommonMark: an indented code block, not a fence).

Scope and assumptions: CommonMark 0.31 §4.5 (fenced code blocks) as the reference; Node LTS, pinned TypeScript toolchain.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P06 in `PROVENANCE.md`).
- reviewer families, as recorded: `gpt-5.6-sol` (through Codex) and `glm-5.2` (through Qwen Code).
- the original code was Go; it is not published. Resolution status: settled by a measurement taken by the
  orchestrating agent, then repaired in a later round.

## 4. Synthetic transformation

Replaced: the language (Go → TypeScript), the names and the runbook. Kept, as the causal property of the disagreement:
a suite named for a whole class of inputs that exercises only one part of it, so a reviewer who reads the suite's name and
sees it green generalises to the class; and an implementation that loses the indentation rule by trimming lines.
Not kept: the same review's second, separate question about how a rate-limit cooldown is keyed (a scope question in the
original, see `PROVENANCE.md`, incident I03).

## 5. Review target

`base/`, `change/` (with its suite) and `change.diff`. The reviewer under evaluation receives those and §1 only.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (reject):** two malformed openers are accepted without a refusal. Evidence in the record: an executed
  probe returning a command and no error for both inputs, and the source lines showing the trim and the missing info-string
  check.
- **Reviewer B (accept):** the extractor refuses all malformed fence shapes by named class. Evidence in the record: the
  author's suite, run green. The suite's cases vary only the closing fence.
- Newly reconstructed evidence: none yet; W2 builds it.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

`cd tests/corpus/cases/PV-004 && ./check.sh` — **not run**.

## 9. Limits

- It shows that a green, well-named suite can be narrower than its name; it says nothing about extractors in general.
- The reference is the CommonMark specification text, applied by hand in the check's expected lines, unless W2 pins a
  conformant Markdown parser for the reference.
