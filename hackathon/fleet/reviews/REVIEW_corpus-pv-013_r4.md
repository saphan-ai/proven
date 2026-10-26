# REVIEW — corpus-pv-013 · round 4

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T20:39:48Z

## Summary

PV-013 reproduces CASE.md section 8 and the round-2 path condition is addressed. One chronology discrepancy remains owed.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-013

### Ground

`git rev-parse --short=7 HEAD` returned `6932236`. The review compares `9303b2d...6932236`; the repair diff is `4c1d2cb..6932236`. Local `go version` returned `go version go1.26.5 linux/amd64`; both fixture modules declare `go 1.26`. The check builds both modules with `GOTOOLCHAIN=local`, `GOPROXY=off`, and `GOSUMDB=off`. Its result was measured on this seat, not inherited from the master's claim.

`bash test/corpus/cases/PV-013/check.sh` exited 0 and printed:

```text
control | base: none | change: pass | base_token_in_output: no | change_token_in_output: no
valid-name | base: accepted | change: accepted | base_token_in_output: no | change_token_in_output: no
token-in-credentials-show | base: (empty) | change: (empty) | base_token_in_output: no | change_token_in_output: no
bad-token-env | base: accepted | change: refused: token_env must be an environment-variable name matching [A-Z][A-Z0-9_]*; rejected value: chat-token | base_token_in_output: no | change_token_in_output: no
bad-channel | base: accepted | change: refused: channel must be a channel name matching #[a-z0-9][a-z0-9_-]*; rejected value: not a channel | base_token_in_output: no | change_token_in_output: no
bad-endpoint | base: accepted | change: refused: endpoint must be an http or https URL with a host; rejected value: not-a-url | base_token_in_output: no | change_token_in_output: no
token-in-name-then-show | base: accepted / token_env=tok_synthetic_0000_not_a_secret | change: refused: token_env must be an environment-variable name matching [A-Z][A-Z0-9_]*; rejected value: tok_synthetic_0000_not_a_secret / (empty) | base_token_in_output: yes | change_token_in_output: yes
```

The seven lines match `CASE.md` section 8 exactly. The master's recorded output is corroborated. The check's false-world answer is exit 1 when the required change-side refusal or printed synthetic value is missing; exit 0 does not establish absence of every possible leak outside the case's output paths.

### Differential

`git diff 4c1d2cb..6932236` changes four occurrences of `tests/corpus` to `test/corpus`: three commands in `CASE.md` and the command comment in `check.sh`. No other content changes. `git diff --name-status 9303b2d...6932236` names only the PV-013 case and its manifest line. The manifest change records fixture and check status. `git diff --check 9303b2d...6932236` exited 0. No dependency file outside the two case modules changed, and the repair changes neither module.

### Unplug

`bash test/corpus/cases/PV-013/check.sh --negative-control` exited 1 and printed `negative-base-only | base: tok_synthetic_0000_not_a_secret | change: refused: token_env rejected | base_token_in_output: yes | change_token_in_output: no`. This removes the value from the change-side refusal while retaining a base-side leak; the assertion refuses it.

`bash test/corpus/cases/PV-013/check.sh --negative-control=no-refusal` exited 1 and printed `negative-no-refusal | base: none | change: accepted: tok_synthetic_0000_not_a_secret | base_token_in_output: no | change_token_in_output: yes`. This removes the change-side refusal while retaining its printed value; the assertion refuses it. Neither a base-side leak nor a value alone can satisfy the case.

### Closure

`change/main.go:124` quotes the rejected name-field value in the refusal. `change/main.go:184` runs the shipped control with the token in credentials and legal field values; that path does not call the refusing branch. `check.sh:183-184` runs the pasted-token case and requires both the refusal and token in change output. `expected/EXPECTED.md`, read after the check and code, requires exactly this combination. The observed output establishes the expected defect. An opposite outcome, in which change refuses but does not print the token, would show `change_token_in_output: no` and cause the assertion to fail. The same output therefore cannot fit that opposite outcome.

### Remainder

The case's `CASE.md:5` says its code was built on 2026-10-21 and calls it declared prior art; `manifest.json` also says built 2026-10-21. The repository rule says all project code is written from 22 through 27 October 2026 and nothing written before 22 October may be copied in. This chronology discrepancy predates the path-only repair and is OWED under the re-review boundary. The owner should verify the construction date and correct the claim or replace the fixture through the authorized code process. This review does not verify the historical construction date.

No remaining `tests/`, `fleet/`, or `conformance/` reference was found by `rg -n` over the PV-013 case. A targeted search for private paths, environment variable names and host data found only `check.sh:13`, a relative scratch path inside this repository; no private material was found. The review did not judge licences, other cases, network paths, logs, crash output, or non-case production behavior.

### Profile

Docs and Go fixture. The reproduction commands in `CASE.md` run as written. The check is a project-owned suite instrument, so its pass alone is circular; the source inspection at `change/main.go:124` and the two negative controls show why the printed line has the required meaning. No numeric quality bar is declared for this case, so no metric was invented. No threshold, assertion, skip, suppression, or dependency was weakened by the repair diff.

### Focus

Round-2 path condition: ADDRESSED at `CASE.md:88,109,119` and `check.sh:3`; the corrected command executes from the repository root. The repair changes only that path. The full case stays within `test/corpus/cases/PV-013/` plus its manifest line. The case's documented seven-line result is reproduced byte for byte.

### Conditions

OWED: resolve the construction-date discrepancy in `CASE.md:5` and `test/corpus/manifest.json` before claiming the fixture satisfies the repository's 22–27 October authorship rule. This does not make the path repair or the case reproduction wrong. No BLOCKING condition remains from round 2.

### Unrun

none

## The review status as filed

## REVIEW — corpus-pv-013
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-013 reproduces CASE.md section 8 and the round-2 path condition is addressed. One chronology discrepancy remains owed.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T20:32:44Z

## REDACTED

none
