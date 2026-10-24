# Held-out set 1

A held-out set is a group of fresh cases that Proven has never been run on. The cases are written and committed (frozen)
before the first run. The freeze commit is named in the [main README](../../../README.md). After the freeze there is one
run. The results are published as they come, misses included.

## The rule

- Never run before the freeze. No case here was run on Proven before the commit that holds it.
- A fix found by a miss is a general rule. It is checked on another fresh set, never on these cases.
- The set tests spread, not depth: six different defect classes, one case each.
- The cases do not change after the freeze. A mistake found later is recorded in the results, not edited away.

## What is here

- [`manifest.json`](manifest.json): the six cases, `"frozen": true`, and the base commit.
- `HO-00N/CASE.md`: the story and the contract a test author builds from. It has no answer.
- `HO-00N/expected/EXPECTED.md`: the answer key. Withhold it from the reviewer under evaluation and from the test authors.

This round is text only. The fixture code is written in a later round by test authors who do not read `expected/`.

## Where results will be published

In this directory, in a results file added after the single run, and in the project's status documents under `fleet/`.
Nothing is published here yet.
