# REVIEW — corpus-pv-018 · round 3

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T15:20:46Z

## Summary

The differential and surviving four-test mutant establish the expected test gap; correct the stale reproduction command in CASE.md before landing.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-018

### Ground

Reviewed `ae34d44...3de5099`; first `git rev-parse --short=7 HEAD` printed `3de5099`. `go version` printed `go version go1.26.5 linux/amd64`; the case module declares `go 1.20`. Ran `bash test/corpus/cases/PV-018/check.sh` on this HEAD: exit 0. `git diff --check ae34d44...3de5099` reported no errors. This review concerns the synthetic case, not the historical incident's original source.

### Differential

Measured output, verbatim:

```
input=shipped base=pass change=pass mutant=pass reference=pass
input=planted-citing-row-outside-the-three-pages base=pass change=fail(changelog#0:160>120) mutant=pass reference=fail(changelog#0:160>120)
input=plain-row-between-budgets-on-a-listed-page base=pass change=pass mutant=fail(install#2:160>120) reference=pass
suite change=pass(4/4)
suite mutant=pass(4/4)
```

`test/corpus/cases/PV-018/check/main.go:39` constructs the outside-page citation; `test/corpus/cases/PV-018/base/validator.go:11` uses the wide limit, and `test/corpus/cases/PV-018/change/validator.go:15` uses the content marker. They differ only on the planted citing input among these three inputs. The plain row is accepted by both; the named-page mutant refuses it. `test/corpus/cases/PV-018/data/data.go:45` places citations on exactly the three named pages, but the shipped data has no plain row on one of those pages between budgets. Thus equal *outcomes*, not equal row selection, is the accurate claim now written in CASE.md §4.

### Unplug

`test/corpus/cases/PV-018/check.sh:34-36` copies the unchanged four tests and substitutes only the named-page validator. All four pass on both versions; the planted row separates the content rule from the named-page rule. False-world answer: if the tests DID pin content derivation, the mutant suite would fail, while the planted change result could still be `fail(...160>120)`. Therefore the observed passing mutant establishes the missing test coverage; the planted line alone cannot. The reference in `check/main.go:20-33` repeats the same content rule and is not independent evidence of the historical validator's correctness. This fixture establishes a test defect, not a defect in the change validator.

### Closure

Round 2 changes `check.sh:11-23` to compare computed input lines against explicit expected lines and `check.sh:25-61` to retain test diagnostics, count four passes, and refuse failing or missing tests. It changes CASE.md §4 from equal selected rows to equal validation outcomes and adds the reference's non-independence in §9. The manifest's PV-018 status remains an object and only its fixture/check values changed. These observable repairs address the visible issues. The supplied order does not include the actual round-1 review reports or their verbatim condition text, so a one-to-one audit of each historical condition is unverified.

### Remainder

OWED — `test/corpus/cases/PV-018/CASE.md:85`: the reproduction command still names `tests/corpus/...` after the move to `test/corpus/...`. Running the documented command from the repository root prints `bash: tests/corpus/cases/PV-018/check.sh: No such file or directory` and exits 127; the corrected command above exits 0. Replace the command in §8 with the actual current path. This is a reproducibility/documentation condition, not a failure of the case mechanism. No change to `expected/` appears in the three-dot diff; its outcome class agrees with the measured surviving mutant. No internal host or seat names, internal paths, private repository names, real tokens, or email addresses appeared in the scoped files. `CASE.md:45` names a private record generically but does not expose its path.

### Profile

Test fixture and documentation, no production authentication or data access. Check ranking: the reference is an in-repository restatement of the rule; the hard-coded shell expected lines are an in-repository assertion; the unchanged four-test suite is weaker alone, but its mutant survival is a direct negative control. No external independent oracle was run. Portability inspection: `bash`, `mktemp -d`, `trap`, `cp -R`, `diff -u`, `awk`, and `go` have Linux and macOS forms used here; no GNU-only option found. `GOTOOLCHAIN=local` and offline module flags avoid toolchain or module downloads on the tested Go version; the required local Go toolchain and `bash` remain prerequisites. The check was executed on Linux only.

### Focus

The case reproduces the distinction between current validator behavior and shipped test power. The actual inputs have no unexplained base/change disagreement. The `expected/EXPECTED.md` finding follows from mutant suite success together with the planted row disagreement, not from the reported historical ruling. The scope diff contains only files under `test/corpus/cases/PV-018/` outside `expected/` and the PV-018 manifest status entry.

### Conditions

Correct the stale `CASE.md:85` path and rerun the documented command from the repository root. Class: OWED, not BLOCKING. No other repair to the fixture is requested. Historical round-1 conditions cannot be marked individually ADDRESSED without the original reports; the measurable repairs are enumerated under Closure.

### Unrun

Native macOS execution, the project's full suite, and original private historical reviews were not available in this Linux-only, scoped review. No conclusion about the original unpublished validator or other corpus cases follows from this run.

## The review status as filed

## REVIEW — corpus-pv-018
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The differential and surviving four-test mutant establish the expected test gap; correct the stale reproduction command in CASE.md before landing.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T15:14:00Z

## REDACTED

none
