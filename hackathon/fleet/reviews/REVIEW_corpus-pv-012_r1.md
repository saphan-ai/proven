# REVIEW — corpus-pv-012 · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 2 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:15:44Z

## Summary

PV-012 prints the intended differential, but its reference is hard-coded from input names instead of independent seeded ownership; a false ownership world prints the same evidence. The team-bearing request also exposes a contract/code mismatch. Repair the reference and settle the request semantics, then rerun the case.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-012

### Ground

`git rev-parse --short=7 HEAD` printed `657ce27`. The scope is `fe146e0...657ce27`, one commit. `git diff --name-status fe146e0...657ce27` shows only `tests/corpus/cases/PV-012/CASE.md`, `base/main.go`, `change/main.go`, `check.sh`, and `tests/corpus/manifest.json`. It does not touch `expected/`. `git diff --check fe146e0...657ce27` exited 0. The module declares Go 1.26; local `go version` printed `go version go1.26.5 linux/amd64`. No declared numeric quality bar was found for this case; the check is a suite check authored with the fixture, and no metric comparison is claimed.

`bash tests/corpus/cases/PV-012/check.sh` exited 0 and printed:

```text
control | base: 5 (Activity for all teams; control=not-present) | change: 0 (Activity scoped to ours; control=PASS) | reference: 0
vacuity | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity scoped to theirs; control=PASS) | reference: 5
application | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity for all teams; control=not-applicable) | reference: 0
scope-no-team | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity for all teams; control=not-applicable) | reference: 0
scope-with-team | base: 5 (Activity for all teams; control=not-present) | change: 5 (Activity for all teams (request team: ours); control=not-applicable) | reference: 0
extra-sync | base: 5 (Activity for all teams; control=not-present) | change: 4 (Activity scoped to ours; control=not-applicable) | reference: 0
control-no-registry | base: 5 (Activity for all teams; control=not-present) | change: 4 (Activity scoped to ours; control=FAIL) | reference: 0
```

### Differential

The named-team control reaches the registry filter: change serves 0 foreign events; removing that filter serves 4 and the control reports FAIL. The vacuity arm serves 5 for `theirs`. The application path serves 5 in both versions because `change/main.go:40` builds with an empty `team`, and `change/main.go:50` guards filtering on that page field. One extra sync by `ours` re-stamps the registry and makes 4 indirect events reappear; the direct-team audit event stays excluded. The two scope lines differ by the appended request-team clause, so a byte-identical sentence claim would be false.

### Unplug

The `control-no-registry` input at `change/main.go:83-85` disables only the registry filter, and its reported control changes from PASS to FAIL. That is a valid negative control for the named-team control. It cannot validate the production construction path, where the filter is not reached. More critically, `check.sh:20-21` assigns the reference by input name. In a false world where the four indirect events were genuinely `ours`, its printed reference would still be 0 for `extra-sync` and `application`; the check has no independent ownership reading that could answer differently. Thus the five-event defect label is not established by this instrument alone.

### Closure

**BLOCKING — reference is not independent.** `base/main.go:21-25` and `change/main.go:21-25` create the four indirect events without seeded owner fields, while `check.sh:20-21` supplies the reference from input names. Section 2 requires ownership of every event in case data, independent of the page and used by the check. The hard-coded line can agree with the expected answer even if the ownership premise is false. Remedy: put immutable seeded ownership for every event in case data, make the check derive expected counts from that data without reading page or registry results, and demonstrate a one-field ownership mutation changes the reference while page behavior stays fixed.

**BLOCKING — request semantics disagree with the contract.** `CASE.md:52-54` says change keeps the requesting team's events when the request carries a team. `change/main.go:94-96` gives the `scope-with-team` request `ours`, but `change/main.go:50` tests `page.team` and serves all 5, as the check confirms. Remedy: settle whether the filter is keyed by the request or by page construction in the contract, then make the fixture and named input follow that decision. If the request is meant only to change the sentence, say that expressly in §2.

### Remainder

The diff stays in the allowed case directory and manifest line. The reviewed files showed no internal host or seat names, internal absolute paths, private repository names, real tokens, or email addresses. The script wrote Go binaries to `██████████████████████████`. Cleanup was refused by the tool runner (`rm -f style commands are not permitted`), so those generated files remain; the owner or a permitted cleanup role must remove them. I declined to judge the historical private incident or the later repair; neither is in this diff.

### Profile

Documentation and Go fixture review. `check.sh` uses Bash arrays, `[[ ]]`, `read`, `local`, `dirname`, `pwd`, `mkdir`, and `go build`; these are available on ordinary Linux and macOS with Bash and the pinned Go toolchain. It is not POSIX `sh` compatible, but the invocation explicitly uses Bash. No network call appears. The case has no declared coverage, mutation, lint, or type-warning metric to rerun. Check ranking: the case's own check is suite-level and can be made green by fixture edits; the independent reference required by §2 would be the stronger instrument once implemented.

### Focus

The observed output matches the seven lines recorded in `CASE.md:104-110`, including the intended production 5 versus reference 0 and extra-sync 4 versus reference 0. This does not close the reference defect above. `check.sh` prints differences and exits 0 by design; exit 0 alone does not mean the page meets the reference. The `expected/EXPECTED.md` label matches those printed counts, but those counts cannot independently establish all five events' seeded owners.

### Conditions

No acceptance conditions: reject this revision. Repair both blocking findings and repeat the same check with an ownership negative control. The master must resolve the §2 request wording before a new verdict. Ruling: I treat the literal request-team requirement as operative because the fixture has a separate `request.team` field; if the intended contract used “request” to mean page construction, that wording needs correction, and the cost of this ruling is one unnecessary fixture revision.

### Unrun

macOS execution was not run because this seat is Linux. The full Go suite, race detector, and coverage were not run because this review is confined to the standalone corpus case and has no declared metric or CI gate for this change. The independent historical started-application reproduction was not run because the original private application is outside this diff.

## The review status as filed

## REVIEW — corpus-pv-012
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: PV-012 prints the intended differential, but its reference is hard-coded from input names instead of independent seeded ownership; a false ownership world prints the same evidence. The team-bearing request also exposes a contract/code mismatch. Repair the reference and settle the request semantics, then rerun the case.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T18:10:43Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
