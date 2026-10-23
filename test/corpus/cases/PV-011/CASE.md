# PV-011 — A per-value budget that does not bound a whole row

- incident: I07 (an incident of several review rounds; this case is a later round than PV-009 and PV-010, and merges two pairs)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A help site serves its sections from a table of rows. Each row has named fields of several shapes: a single text, a
list of texts, and others. Over earlier rounds, long history notes had piled up in a few rows, and one row had grown to
about 25 KB. The change under review moves that history off the rows and leaves a short pointer; the largest row drops
below 1 KB. To keep it that way, the change adds a budget: when rows are loaded, every text value above 800 bytes is
refused. A comment in the code says that no field of a row is outside the budget.

Two reviewers from different model families reviewed the same change. They found largely the same facts: the budget
looks only at fields that are a single text or a list of texts, and it bounds each value, not the row. They disagreed on
the grade. One accepted with conditions: nothing the site serves today is false, so each finding is a sentence to fix.
The other rejected: the mechanism does not hold what the code says it holds.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- two versions of one row loader, **base** and **change**, each offering one operation: given a row, answer accept or
  refuse, and name the field over budget.
- rows used by the check have fields of these shapes: a single text; a list of texts; a map from text to text; a list
  of lists of texts; a nested record with a text field; a byte sequence that holds text.
- **base** has no budget; it accepts every row.
- **change** refuses a row when a single-text field, or one item of a list-of-texts field, is longer than 800 bytes.
  It does not look at fields of any other shape. It does not bound the sum of a row's texts.
- **change** ships a boundary test whose expected results are computed from the same 800 constant the loader uses.
- a third version, **change-raised**: change with only the constant raised to 900. Nothing else differs.
- a **reference** measurement, independent of the loader and used only by the check: it walks every text value
  reachable in the row, whatever its shape, and reports the largest single value and the row's total.
- a **check** that runs every input below against base, change and the reference, and prints one line per input:
  input name, base result, change result, reference largest value, reference total. It also runs change's own tests on
  change-raised and prints whether they pass.

Inputs the check must include:
- a compliant row like the ones served today (all three agree: accept, small total);
- a single-text field of 801 bytes; a list-of-texts field with one 801-byte item (change refuses both);
- a map field with one 900-byte value; a list-of-lists field with one 900-byte value; a nested record whose text field
  holds 801 bytes; a byte-sequence field holding 3,750 bytes of text;
- a list-of-texts field with forty values of 700 bytes each (28,000 bytes in one row).

Scope and assumptions: sizes are counted in bytes of UTF-8 text. The toolchain is chosen when the case code is written,
pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P21 in `PROVENANCE.md`); a
  later round of the same work is pair P25.
- reviewer families, as recorded: P21 — `claude-opus-5` (through Claude Code) and `gpt-5.6-sol` (through Codex); P25 —
  `claude-opus-5` (through Claude Code) and `glm-5.2` (through Qwen Code).
- the original code was Go; it is not published. Resolution status: P21 was settled by the next round, which carried
  both reviews and closed every escape they measured. That round's own review pair (P25) found more of the same; its
  repair round was written and reviewed but did not merge.

## 4. Synthetic transformation

Replaced: the site, the rows, the field names, the sizes of the original rows and the names of the original code. Kept,
as the causal property of the disagreement: a budget that applies to some field shapes only and to each value only,
with a boundary test that derives its expectation from the constant it guards; a code comment that claims more than the
mechanism holds; and rows served today that all fit, so the grade "nothing served is false today" stays available.
Not kept: a date check that read source files instead of served text, and a claim that four earlier false sentences
could no longer be written (both shown false in the record; see Limits).

## 5. Review target

The two versions of the loader, the change's tests and a readable diff between them. The reviewer under evaluation
receives only those and the description of the change in §1 — never the check, the reference, change-raised or
`expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** the budget is real and the size reduction reproduces. But a map field or a
  list-of-lists field of 900 bytes passes every test, so the comment "no field is outside the budget" is false; the
  bound is per value, and nothing bounds how many values a row holds; and earlier false sentences can be written back
  with every test green. None of this makes a served sentence false today, so none blocks; each is a sentence to fix.
  Evidence in the record: deleting the budget call turned tests red; a new text field over budget turned them red;
  planted map and list-of-lists fields of 900 bytes stayed green; a 790-byte false sentence written back stayed green
  and was rendered by the site. These measurements were accurate. It did not test whether the 800 ceiling was held by
  anything independent of the constant.
- **Reviewer B (reject):** the per-value budget does not bound a served row, misses nested and map values, allows the
  ceiling to move from 800 to 900 with every test green because each boundary test derives from the same constant, and
  the earlier false sentences can be restored with tests green. Evidence in the record: a nested-record field of 801
  bytes accepted; a map field accepted; forty 700-byte values accepted (28,000 bytes); the constant raised to 900 with
  tests green. Its findings file is not in the part of the record we read; its content is known from the next round's
  order, which restates it. Some of its test suites could not run in its environment, which it recorded.
- **Later round (P25):** the repair replaced the per-value budget with a walk over every field of a row that bounds the
  row's total at 2,800 bytes, and added a comment saying three rows could take no further sentence. `claude-opus-5`
  accepted with conditions and called one finding blocking: the comment is false for rows other than the widest, which
  is the only row the test measures; a map with non-text keys is counted as one value; a byte-sequence field holding
  3,750 bytes is reported as carrying no text; a timestamp field makes loading fail. `glm-5.2` accepted: the walk has
  two blind shapes, no row carries them today, and a second check would flag such a field as unmeasured. That last claim
  was asserted, not planted; the other reviewer planted the field and the second check reported it as not applicable.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base, change and change-raised and record each line of its output. **Not run**: the case code is
not built yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- It shows that a per-value, per-shape budget does not bound a row; it does not say what the right bound is.
- The disagreement in P21 is partly about grading (does a false code comment block when every served sentence is
  true?). The case can establish the facts; the grade is a policy, recorded in `expected/` as the record settled it.
- The later round's repair (P25) did not merge, and whether its gaps remain in the original code was not measured. The
  check covers the first round's change only; the byte-sequence input is the one overlap.
- The date check and the four restorable false sentences are not reconstructed.
