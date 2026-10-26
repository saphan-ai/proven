# REVIEW — held-out-set-1 · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-24T14:45:33Z

## Summary

HO-003's promised directory mode is ambiguous, making its answer key unfair. Record a post-freeze correction or exclude it before scoring. The other five cases are distinct; the freeze and README commits have the required boundaries.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — held-out-set-1

### Ground

`git rev-parse --short=7 HEAD` printed `ece0efb`. I read all six contracts, answer keys, the manifest and both commits. `git diff-tree --no-commit-id --name-only -r 0fabc62` listed 14 files, all under `test/corpus/held-out/`; for `ece0efb` it listed only `README.md`, which names `0fabc62`. The comparison from `9b11366` includes intervening unrelated commits, so the review focuses on these two commits. `git diff --check 0fabc62^..ece0efb` exited 0. The manifest declares `run_before_freeze: false`, which is not independent proof of pre-freeze history.

### Differential

**BLOCKING — ambiguous reference makes HO-003 unfair.** `test/corpus/held-out/HO-003/CASE.md:12` says the copy adds read bits and leaves the rest as it was. Section 2 at lines 25–33 supplies source directories of `0700`, a base that sets destination directories to `0755`, and a change that ORs read bits onto newly created entries. If “rest as it was” means source mode, the promised directory mode is `0744` (not traversable by group or others, contrary to the readability promise at line 10); if it means the default newly created destination under umask `022`, it is `0755`. The file references are similarly unspecified as between source and new destination modes. `test/corpus/held-out/HO-003/expected/EXPECTED.md:4-14` assumes the `022` result is promised and group-write under `002` is the sole defect. A test author using only §2 can choose either reading, so an otherwise correct review of directory traversal or mode preservation can be scored as a miss. Remedy: fix the file/directory reference modes and state what “rest” refers to in a recorded post-freeze correction, or exclude HO-003 from scoring; do not rewrite the freeze.

The other five mechanisms are distinct and their keys follow their contracts: HO-001 accepts a previous run's comment; HO-002 verifies the current marker's text but not its author; HO-004 conflates two meanings of 422; HO-005 substitutes a serving model without reporting it; HO-006 reads an obsolete path only covered by an untracked leftover. HO-001 and HO-002 share the gate setting, but test independent evidence dimensions. None of these five §2 contracts requires an answer key to build the fixture, and §5 withholds checks and expected outcomes from the reviewer under evaluation.

### Unplug

The contrasting inputs specify what should fail when each mechanism is absent: old success with new failure (HO-001); human-authored current marker (HO-002); `002` versus `022` modes (HO-003, conditional on fixed reference); refused input returning 422 (HO-004); unlisted model (HO-005); clean versus dirty checkout (HO-006). These are contract deductions, not executed fixture runs. The ambiguous reference prevents HO-003's negative control from giving one fair answer.

### Closure

Verdict: `reject` on HO-003's scoring contract. The frozen corpus was not edited and no product code was changed. The freeze contains 14 corpus files; the separate README commit contains one README change.

### Remainder

A name search and manual reading found no private person, host, vendor or product names in the six stories or keys. Generic public technology names do appear: `Go` and `Unix-like` at `test/corpus/held-out/HO-003/CASE.md:4` and :38, and `git mv` at `test/corpus/held-out/HO-006/CASE.md:25`. If “no product name” literally includes public tools and languages, that criterion is not met. Remedy: clarify the rule and record a correction; these terms do not expose the incidents' identities. Declined to judge: actual reviewer detection rate and the independently unobserved pre-freeze execution history.

### Profile

Text-only contract and provenance review, with an expected-key separation boundary. No production code, dependency manifest or runtime guard changed in these two commits. Check ranking: Git's path/diff check is independent of the answer keys; comparing the project's own keys to its own contracts is circular; no suite or mutation check exists until fixtures are built. No metric baseline was supplied for this text-only change.

### Focus

File · line · control: `test/corpus/held-out/HO-003/CASE.md:12` and :25–33 versus `test/corpus/held-out/HO-003/expected/EXPECTED.md:4`; the missing control is a fixed reference independent of umask and source-versus-new-entry interpretation. Ruling: the likely intended reference is `0644` files and `0755` directories, but it is not a contractual fact; if wrong, a valid source-mode-preserving review would be misgraded.

### Conditions

Define and record the reference modes before scoring HO-003, or exclude it; preserve the frozen commit. Clarify whether public generic technology names are excluded from the no-product-name requirement.

### Unrun

Fixture base/change/check programs and negative-control execution: not run, because they do not exist yet. Full Go suite and review-model evaluation: not run, because the order confines this text-only review to the six contracts. The manifest's claim that there was no earlier run is unverified independently.

## The review status as filed

## REVIEW — held-out-set-1
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: HO-003's promised directory mode is ambiguous, making its answer key unfair. Record a post-freeze correction or exclude it before scoring. The other five cases are distinct; the freeze and README commits have the required boundaries.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-24T14:39:53Z

## REDACTED

none
