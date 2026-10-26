# REVIEW — corpus-pv-016 · round 3

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 1 OWED · 1 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T15:25:44Z

## Summary

At the reviewed tip the relocated check exits 1 before printing any claims. Its embedded tests/corpus path is stale; repair the path, rerun, and independently verify the quote negative control before accepting its recorded output.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-016

### Ground

`git rev-parse --short=7 HEAD` printed `43b3915`. `git diff --name-status ae34d44...HEAD` shows only `test/corpus/cases/PV-016/` (no `expected/`) and `test/corpus/manifest.json`. Commits `9e6896f` and `43b3915` supply the case and repair. `diff -q` found identical base and change `go.mod`, `alerts.go`, and `main.go`; only change has the four pages. The manifest changes only PV-016's status fields. No private host, seat, path, token or actual address is visible in this diff; the sample credential is explicitly synthetic.

### Differential

At HEAD, `bash test/corpus/cases/PV-016/check.sh` exited 1 and printed exactly:

```
diff: tests/corpus/cases/PV-016/base/go.mod: No such file or directory
diff: tests/corpus/cases/PV-016/change/go.mod: No such file or directory
check: base/go.mod and change/go.mod differ
```

It prints zero of the promised 24 claim lines, contradicting the master's exit-0 assertion and `CASE.md` §8's recorded 24-line output at this SHA. The first run halts before invoking Go. Direct `go -C test/corpus/cases/PV-016/base run . emit backup_failed` prints `considered=1 delivered=0 reason="no transport yet"`; direct `go -C test/corpus/cases/PV-016/change run . dest chat:t-backup-ops` prints `refused: destination id "t-backup-ops" must begin with c-`. Each succeeds. These sample only two operations; they are not a replacement for the claim-by-claim check. The script did not exist at base, so there is no base check result.

### Unplug

The round-2 missing-quote guard exists at `check.sh:31-39`, but cannot be reached: a changed quote and an unchanged quote both first produce the same path-error output. False-world answer: the same error and exit 1 even if the pages and implementation are perfectly consistent. The recorded quote mutation in `CASE.md:135-137` was not independently reproduced and does not validate the currently runnable checker. After repairing paths, change exactly one quoted sentence on a disposable copy, require `quote not found`, restore, then obtain 24 lines; this is a proposed control, not a claimed run.

### Closure

A counterexample in this actual worktree suffices: real `test/corpus` files are present and identical, but `check.sh:7,15-19` tests nonexistent `tests/corpus` files and falsely reports code differences. In `base/alerts.go:55-59` the resolver always returns the null sender, whose `Send` is false; `base/alerts.go:68-84` only increments delivered on a successful send. Change has identical source. The source inspection distinguishes a false-world tool that merely *prints* zero deliveries even when actual sends succeed. No code or test was altered in this review.

### Remainder

The script is an evidence printer, not a delivery oracle: `check.sh:44-75` will exit 0 and print `delivered=1` if the tool reports successful delivery. A human can distinguish that output from the expected outcome, but cannot interpret exit 0 alone as proof. The page quotes, including absent `retry_count` and `threshold`, are present in change. The four page-internal findings are prose, not runtime assertions. Linux only: `mktemp -d` with template, Bash substitution, `grep -qF --`, `diff -q`, and `go -C` (Go 1.26) appear usable on macOS, but were not executed there; the definite portability failure on both platforms is the wrong repository path.

### Profile

Corpus-fixture/test and documentation profile. Read contract, four pages, expected key, claim list, source pair, script, and manifest. Check ranking: source and pages are primary fixture evidence; the author-controlled script is a suite check and cannot independently certify its own answer. No quantitative baseline, changed dependency version, exemption or threshold adjustment in this scope.

### Focus

The intended source-level scenario holds: the two code variants are identical and a null sender is always selected; only change has pages promising delivery and describing a refused destination and two nonexistent fields. The expected key's no-delivery finding follows from code inspection plus sampled operations, **not from the check's actual output**. The actual output also fits the opposite delivery outcome because no tool operation was run. The pages-only review's four contradictions/omissions remain visible and its reviewer was not required to read the code.

### Conditions

- BLOCKING — `test/corpus/cases/PV-016/check.sh:7` and `test/corpus/cases/PV-016/CASE.md:102`: both still name `tests/corpus` after the move to `test/corpus`. Correct paths, rerun the full check from repository root, refresh recorded output, and verify one quote mutation fails specifically for missing quote. The current run exits 1 before any claim is tested.
- OWED — `test/corpus/cases/PV-016/check.sh:44-75`: if exit status is supposed to establish expected outcome, add a check that fails on an opposite delivery count and unplug that assertion once. Otherwise state plainly that only inspection of printed data decides the outcome.
- NOTE — `test/corpus/cases/PV-016/CASE.md:135-138` and `test/corpus/manifest.json:290-294`: round 2 added the quote guard, temporary-binary cleanup and restored the status object, but the quote-guard proof is not transferable to the moved layout. Exact verbatim round-1 conditions were not supplied; matching each condition against the original report is unverified.

### Unrun

- Full 24-claim check: attempted and failed before claims because of stale path.
- Independent missing-quote mutation: blocked behind stale path; no reviewed files modified.
- macOS run: no macOS environment available.
- Whole-repository and base-commit Go suites: outside this case-only review; no result claimed.

## The review status as filed

## REVIEW — corpus-pv-016
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: At the reviewed tip the relocated check exits 1 before printing any claims. Its embedded tests/corpus path is stale; repair the path, rerun, and independently verify the quote negative control before accepting its recorded output.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T15:19:21Z

## REDACTED

none
