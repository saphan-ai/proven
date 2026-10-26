# REVIEW — corpus-pv-006 · round 2

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 2 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T18:09:24Z

## Summary

PV-006 reproduces two change-side misses with raw-token evidence; Linux check passes. Confirm macOS run and reconcile documentation chronology.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-006

### Ground
HEAD `0b5a194` (`git rev-parse --short=7 HEAD`); range `9303b2d...0b5a194`. Ran `bash test/corpus/cases/PV-006/check.sh` on Linux with `go1.26.5 linux/amd64`; exit 0:

```text
well-formed base=pass change=pass reference=pass
listed-token base=fail({limit}) change=fail({limit}) reference=fail
unlisted-token base=pass change=fail({window}) reference=fail
dollar-token base=pass change=pass reference=fail raw=$wait
duration-0-vs-1 base=pass change=pass reference=fail shown=0 retry-after=1
duration-equal base=pass change=pass reference=pass shown=30 retry-after=30
```

### Differential
`test/corpus/cases/PV-006/base/base.go:9` lists `{limit}` but not `{window}`; `change/change.go:10` checks both using the same brace pattern as `product/product.go:15`. Both controls reject listed tokens, only change rejects unlisted braces, and both pass the filled entry. `product/product.go:20` copies `$wait` unchanged; both text controls pass it, whereas the hand-written exact-text oracle (`reference/reference.go:8`) rejects it. For 0/1, the base always passes (`base/base.go:23`), change allows a two-second shortfall (`change/change.go:22`) and exact equality rejects (`reference/reference.go:23`). At 30/30 all pass. This supports the expected established defect, not a claim that the repaired brace controls miss earlier attacks.

### Unplug
`test/corpus/cases/PV-006/check/main.go:40` requires the *rendered* notice to contain `$wait` before printing `raw=$wait`. `go -C test/corpus/cases/PV-006 test -count=1 -v ./check` exits 0 and logs `negative control alarm: rendered notice does not contain "$wait"` on the intentionally replaced token (`check/main_test.go:15`). False-world answer: if the renderer removes or fills this token, the evidence check returns an error and the main check panics rather than printing the claimed dollar-token observation. A reference mismatch without this assertion could also fit another text difference. The test does not establish the behavior of arbitrary future token shapes.

### Closure
`git diff --name-only 9303b2d...0b5a194` lists only PV-006 non-`expected/` files and the manifest; its diff changes only PV-006's status fields (`test/corpus/manifest.json:112`). `git diff --check` exits 0. Round 2's rendered-token assertion and missing-token negative control address the raw-token evidence gap, and the printed shown/retry-after values address the duration-verdict ambiguity. No control was relaxed. A scan of the case and manifest for absolute host paths, e-mail addresses, URLs and secret-variable names found no matches; no private source, credential or internal host appears in the inspected diff.

### Remainder
- OWED · `test/corpus/cases/PV-006/CASE.md:93` records the check run as 2026-10-23, while `test/corpus/manifest.json:113` says checks ran 2026-10-22. Reconcile with the run record; the discrepancy does not change the observed result.
- OWED · The order says both first-round reviews' conditions were carried verbatim into round 2, but their exact text was not provisioned here. The two observable evidence gaps above are paid; one-to-one closure of every earlier condition cannot be verified without the condition list. CASE §6 explicitly excludes reconstruction of other predicates and source-scan blind spots.

### Profile
Test-corpus and documentation review. Ranking: the hand-written exact-text reference is a separate project oracle; the check and unplug test are suite-level instruments. No external auditor applies. No new dependency or install script; the case module uses the standard library only. No baseline metric and command for coverage, mutation, lint or types was provided, so none is claimed. The checker uses Bash, `dirname`, `pwd`, and Go `-C`, without an apparent GNU-only flag. Go 1.26 is pinned; macOS execution remains unmeasured.

### Focus
FILE · LINE · CONTROL: `test/corpus/cases/PV-006/check/main.go:60` passes each text input through the shared product and three evaluators; `check/main.go:76` applies all three duration evaluators to each pair. The check has no golden-output assertion for all six lines, so this review compared printed lines against the called implementations. Declined to judge: historical private reviews, original production implementation, other languages/browser behavior and unreconstructed controls; licenses are out of scope.

### Conditions
Run the check and negative-control test on macOS with Go 1.26. Reconcile the CASE/manifest check date. Provide verbatim earlier review conditions for one-to-one closure or explicitly record that the remaining original controls are outside the synthetic contract. These are validation and provenance gaps, not an observed failure of the constructed case.

### Unrun
macOS check and test: no macOS runtime on this Linux seat. Original private first-round review-condition document: not provisioned. No project-wide metric or test result claimed.

## The review status as filed

## REVIEW — corpus-pv-006
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-006 reproduces two change-side misses with raw-token evidence; Linux check passes. Confirm macOS run and reconcile documentation chronology.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T18:03:33Z

## REDACTED

none
