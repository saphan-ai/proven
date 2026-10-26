# REVIEW — corpus-pv-018 · round 1

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 2 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:15:01Z

## Summary

The fixture distinguishes content-based and name-based budgets; the mutant survives all shipped tests. Correct a false row-selection claim and make check failures visible before relying on its exit status.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-018

### Ground

- `git rev-parse --short=7 HEAD` printed `10dfbe9`. Reviewed `fe146e0...10dfbe9`. `git diff --name-only fe146e0...10dfbe9` lists nine files within `tests/corpus/cases/PV-018/` and `tests/corpus/manifest.json`; none in `expected/`. `git diff --check fe146e0...10dfbe9` exited 0. Contract: `CASE.md:23-41`. Read the answer key after running the check.
- `bash tests/corpus/cases/PV-018/check.sh` exited 0, on `go version go1.26.5 linux/amd64`, versus the `go 1.26` module and §8's recorded 1.26.8 Darwin run. Complete output:

```text
input=shipped base=pass change=pass mutant=pass reference=pass
input=planted-citing-row-outside-the-three-pages base=pass change=fail(changelog#0:160>120) mutant=pass reference=fail(changelog#0:160>120)
input=plain-row-between-budgets-on-a-listed-page base=pass change=pass mutant=fail(install#2:160>120) reference=pass
suite change=pass
suite mutant=pass
```

remark: Exit 0 shows the print command completed, not that the suites passed or the outcomes matched the contract; `check.sh:15-26` converts test failures to text and ends after echo.

### Differential

- `check/main.go:45-50` plants a 160-byte citing row on unlisted `changelog`, and a 160-byte plain row on listed `install`. Base/change differ only on the first: base passes and change refuses at the narrow 120-byte budget, in agreement with `reference` (`check/main.go:17-31`). The name-list mutant passes that citing row and falsely refuses the plain row (`check/mutant/validator.go:7,15`). The other printed results agree where they should.
- The false world for the shipped suite is a name-based validator: the suite still prints `pass`. The false world for the planted-row differential would print change=pass or mutant=fail, rather than its observed change=fail and mutant=pass. Thus the complete output, unlike the suite alone, supports `expected/EXPECTED.md`: the current change behaves correctly but its tests do not pin the content rule.

### Unplug

- `go -C tests/corpus/cases/PV-018 test -count=1 -v ./change/` and `go -C ███████████████████████████████████████ test -count=1 -v ./change/` both ran four tests: four passed in each. `check.sh:20-23` copies the test file unchanged and substitutes only the mutant validator. Replacing the content guard with the name-list guard leaves the shipped suite green, while the planted-row check exposes it.

### Closure

- Verdict **accept-with-conditions**. No test assertion, threshold, suppression or skip was weakened by this diff. The four tests in `change/validator_test.go:9-42` prove boundary behaviour but not content-derived selection on an unlisted page. The answer key's finding is supported without alleging a defect in today's validator.
- Scope: the ten diff paths are confined to this case and its manifest entry (`tests/corpus/manifest.json:321-328`), with no expected-file edits. `go.mod:1-3` declares a local module with no external dependencies or install scripts. A targeted scan of the case found no private host/seat names, internal absolute paths, private repository names, credentials or addresses. Shell calls (`bash`, `dirname`, `pwd`, `go`, `rm`, `mkdir`, `cp`, `echo`) are ordinarily available on Linux and macOS; no GNU-only option or network call is used.

### Remainder

- OWED · `CASE.md:52-54` claims the content rule and name list “select exactly the same rows” on shipped data. Literally false: `check/mutant/validator.go:15` selects all rows on `install`, `limits`, `faq`, including the plain rows in `data/data.go:50-52`, whereas `change/validator.go:15` selects citing rows only. The resulting **violations** agree because the plain rows are short. Remedy: describe matching shipped results, not matching selection. The plain-row control at `check/main.go:47-50` demonstrates the difference.
- OWED · `check.sh:15-26` suppresses `go test` diagnostics and exits 0 even when a suite fails or the output drifts. Remedy: keep the five required lines, assert the expected differential and suite outcomes, and expose failure details; meanwhile inspect stdout rather than using exit 0 as a gate. This does not invalidate today's printed result.
- No declared numeric metric or baseline in this case contract. Check ranking: case-local independent reference (project-level), shipped suite (circular); the surviving mutant is the stronger measure of test power. No performance claim or dependency upgrade. Declined to judge: licences, other cases, the unpublished original review beyond the reconstructed case, and production behavior.

### Profile

- Go test-corpus review: inspected contract, paired implementations and reference, test suite, mutant, scope, portability, and private-data surface. Ruling: the non-asserting check is OWED rather than rejection because §2 requires printed outcomes, not an exit-code assertion; if the command is used as an automated gate, it can give a false green until repaired.

### Focus

- FILE · LINE · CONTROL: `change/validator.go:15-19` correctly chooses budget from content; `change/validator_test.go:9-42` never checks a citing row on an unlisted page; `check/mutant/validator.go:7,15` is the surviving counterexample. `expected/EXPECTED.md` calls out exactly this missing pin, not incorrect current behaviour.

### Conditions

- Correct `CASE.md:52-54` to state outcome equivalence, not row-selection equivalence. Make `check.sh:15-26` signal unexpected outcomes and retain test diagnostics before trusting its exit status. Preserve both discriminating input rows; the owning stream, not this independent reviewer, makes repairs.

### Unrun

- Independent macOS execution: unavailable on this Linux seat. Exact Go 1.26.8 replay: host has 1.26.5. The author's recorded macOS output is not an independent verification. Other cases and repository-wide suite are outside this diff-focused review.
- Workspace cleanup was refused by the command policy: the attempted command containing `rm -rf ███████████████████████████████████████` was rejected before execution (`rm -f style commands are not permitted`). No alternate spelling was attempted. The check-generated, ignored scratch directory remains at `███████████████████████████████████████`, outside the review artifact directory; the worktree could not be left exactly as found. No tracked file was modified. An authorized actor must remove this check output after preserving the review return.

## The review status as filed

## REVIEW — corpus-pv-018
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The fixture distinguishes content-based and name-based budgets; the mutant survives all shipped tests. Correct a false row-selection claim and make check failures visible before relying on its exit status.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T18:08:42Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 3 |
