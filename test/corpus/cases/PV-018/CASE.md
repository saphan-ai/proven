# PV-018 — A budget derivation whose tests pass with a hard-coded list

- incident: I16 (one incident; this case is its first round)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: built 2026-10-22 (declared: built inside the build window) · checks: run 2026-10-22 · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A site generator builds help pages from rows of text. Each row has a byte budget, and a validator refuses a row over its
budget. The change under review adds a narrower budget for rows that cite a measurement, because a citation expands when
it is rendered. The change says the narrower budget is **derived from the row's content** (does it cite a measurement?),
never from a list of page names. On today's data, exactly three pages have citing rows. The change ships four tests of
the budgets.

Two reviewers from different model families reviewed the same change at the same commit, without seeing each other.
One accepted it with conditions: it planted a citing row on a page outside those three, and the validator refused it at
the narrow budget. The other rejected it: it replaced the content rule with a list of the three page names, and all four
budget tests stayed green.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one validator, **base** and **change**, each offering one operation: given a set of pages and their
  rows, answer pass or fail, and name each row over budget with the budget it was held to.
  - **base** holds every row to one wide budget.
  - **change** holds a row to a narrow budget when the row cites a measurement, and to the wide budget otherwise.
- the four budget tests shipped with **change**, run unchanged.
- a **mutant** of change, used only by the check: the content rule replaced by "the page is one of these three names".
- a **reference**: the stated rule (a citing row gets the narrow budget, wherever it is), applied by the case to each row,
  never by the validator; used only by the check.
- a **check** that prints one line per input: input name, base result, change result, mutant result, reference result;
  and one line per version for the shipped tests: change suite result, mutant suite result.

Inputs the check must include:
- the shipped data, where the citing rows sit on exactly the three listed pages;
- a planted citing row, between the narrow and the wide budget in size, on a page that is not one of the three;
- a row on one of the three pages that does not cite a measurement, between the two budgets in size.

Scope and assumptions: byte budgets; a fixed citation marker. The toolchain is chosen when the case code is written,
pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P37 in `PROVENANCE.md`).
- reviewer families, as recorded: `claude-opus-5` (through Claude Code) and `glm-5.2` (through Qwen Code).
- the original code was a Go validator, reviewed inside a larger catch-up merge; it is not published. Resolution status:
  settled by a ruling of the orchestrating agent; later repair rounds took the same class further.

## 4. Synthetic transformation

Replaced: the pages, their rows, the marker and the numbers. Kept, as the causal property of the disagreement: on the
shipped data the content rule and the list of names give the same validation outcomes, so the shipped tests cannot tell
them apart, while a planted row elsewhere can. The rules need not select the same rows when both budgets pass.
Not kept: the other parts of the catch-up merge and two further findings about them (see Limits).

## 5. Review target

The two versions of the validator, the shipped tests, the data and a readable diff between them. The reviewer under
evaluation receives only those and the description of the change in §1 — never the mutant, the check, the reference
or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** the narrow budget is derived, not listed. Evidence in the record: a data probe
  that planted a citing row on a page outside the list, refused at the narrow budget; citations respelled away; no new
  failing package against the base. This was accurate about behaviour. It exercised the code, not whether the shipped
  tests would notice the rule being replaced.
- **Reviewer B (reject):** the claim "derived, never listed" is not pinned. Evidence in the record: the content rule
  replaced by a three-name list in the code, with all four budget tests passing. This was accurate. It did not test the
  behaviour on a planted row, and some of its package runs were left incomplete.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base, on change and on the mutant, and record each line of its output.

- run: 2026-10-23T02:03Z (UTC), from the repository root; the original five lines were also measured on commit `10dfbe9` before this repair
- toolchain: `go version go1.26.5 linux/amd64`, standard library only, no network (module `pv018` in this directory, `go 1.20`;
  `GOTOOLCHAIN=local` refuses an older local toolchain instead of downloading one)
- command: `bash tests/corpus/cases/PV-018/check.sh` (runs `go run ./check`, then `go test -count=1 -v ./change/` in the case module,
  then the same test file unchanged against a temporary copy of the module whose `change/validator.go` is the mutant's)
- exit code: 0
- output, in full:

```
input=shipped base=pass change=pass mutant=pass reference=pass
input=planted-citing-row-outside-the-three-pages base=pass change=fail(changelog#0:160>120) mutant=pass reference=fail(changelog#0:160>120)
input=plain-row-between-budgets-on-a-listed-page base=pass change=pass mutant=fail(install#2:160>120) reference=pass
suite change=pass(4/4)
suite mutant=pass(4/4)
```

- budgets: wide 200 bytes, narrow 120 bytes, marker `{{measure:`; a violation prints as `page#row:size>budget`.

## 9. Limits

- Both readings are true; they answer different questions (does it behave, and would the tests catch a regression). The
  case is about the second.
- The original review also split on two other parts of the merge: a command that printed a false source for a setting,
  and whether a credential field refused a value. On the credential field, reviewer A was right and reviewer B wrong. Not
  reconstructed here.
- It shows one gap in test power; it does not show that the validator is wrong today.
- The reference repeats the stated content rule; its agreement with change is not independent evidence of correctness.
