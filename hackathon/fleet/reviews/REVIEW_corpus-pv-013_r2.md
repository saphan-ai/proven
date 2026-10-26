# REVIEW — corpus-pv-013 · round 2

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-23T16:18:58Z

## Summary

The check establishes the change-side refusal leak; both negative controls fail as intended. Correct stale reproduction commands and verify the original round-one conditions before gating.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-013

### Ground
Reviewed `9303b2d...4c1d2cb` at HEAD `4c1d2cb` (`git rev-parse --short=7 HEAD`). Read `test/corpus/cases/PV-013/CASE.md` §2, ran the check, then read `expected/EXPECTED.md`. `go version` returned `go version go1.26.5 linux/amd64`; both case modules declare Go 1.26. The check sets `GOTOOLCHAIN=local`, `GOPROXY=off`, `GOSUMDB=off`; the Go sources import the standard library only. `git diff --name-only 9303b2d...4c1d2cb` lists only this case outside `expected/` and `test/corpus/manifest.json`. The manifest changes only PV-013's fixture/check status strings while preserving its status object. `git diff --check 9303b2d...4c1d2cb` exited 0.

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

### Differential
`test/corpus/cases/PV-013/base/main.go` stores any field; `test/corpus/cases/PV-013/change/main.go:120-143` validates three fields and embeds the rejected name value in the refusal. Valid name and credentials/show paths agree. Each invalid field differs by acceptance versus refusal, as §2 requires. On pasted token, base stores and shows it; change refuses it and prints it on the refusal path. Both show paths keep a token placed solely in the credentials file out of output. `change/main.go:184-209` makes the shipped control use valid values only; it passes without exercising the refusal path. This establishes the expected defect without claiming that the shipped control detects it.

### Unplug
`bash test/corpus/cases/PV-013/check.sh --negative-control` printed `negative-base-only | base: tok_synthetic_0000_not_a_secret | change: refused: token_env rejected | base_token_in_output: yes | change_token_in_output: no` and exited 1. `bash test/corpus/cases/PV-013/check.sh --negative-control=no-refusal` printed `negative-no-refusal | base: none | change: accepted: tok_synthetic_0000_not_a_secret | base_token_in_output: no | change_token_in_output: yes` and exited 1. False-world answer: a base-only leak or a change message with the token but without a refusal cannot satisfy `assert_change_refusal_leaks` (`check.sh:60-63`), even if a combined token flag would read yes. These are mutations of synthetic output, not of fixture Go source. `check.sh:160-187` also asserts each added refusal and both pasted-token conditions. The second commit strengthens, rather than bends, the check.

### Closure
**OWED — `test/corpus/cases/PV-013/CASE.md:88,108,118` and `test/corpus/cases/PV-013/check.sh:3`:** Published commands refer to `tests/corpus/...`, but this case lives under `test/corpus/...`. Running the documented `bash tests/corpus/cases/PV-013/check.sh` exited 127 with `No such file or directory`. Remedy: change these four paths and run each command from the repository root. The actual shell check passed on Linux.

### Remainder
The kickoff claims the round-one review conditions were included verbatim in the round-two order, but neither original review nor round-two order text was provisioned here. `git diff d4ff8b7..4c1d2cb` confirms that the combined token flag was split into base/change flags, two discriminating negative controls were added, and the manifest status object was restored. Exact satisfaction of every original condition is unverified; the master should compare the original review records before gating. `CASE.md:74` still reads “Newly reconstructed evidence: none yet”, despite its executed evidence in §8; clarify whether this refers only to the historical review. No numeric quality metric or baseline was declared, so testing-bar has no number to compare. Check ranking: self-authored suite check, its synthetic negative controls, and independently readable Go refusal path; no external audit. Declined to judge logs, crashes, unrelated channels and the second historical finding, as excluded by CASE.md §9 and the order.

### Profile
Security: synthetic credential disclosure through user-provided name → validation → CLI refusal text. Inspected the diff for real tokens, e-mail addresses, internal host or seat names, private paths and private repository names; none found in the reviewed diff. `https://chat.example/hook` is a synthetic example hostname. Bash `local`, `mktemp -d` with a template, `grep -q -F`, `trap` and Go 1.26 require Bash and an installed Go toolchain, but use no Linux-only options; a macOS run is unverified. No lockfile, external dependency or install script changed.

### Focus
Contract §2, expected change-side refusal-and-token predicate, negative controls, diff scope, privacy and portability. A `change_token_in_output: yes` flag by itself could also fit a successful operation that printed a token; the additional `refused: token_env` text and the no-refusal negative control distinguish that world. `change/main.go:149-152` returns exit code 0 for a refusal; the contract asks for an accepted/refused answer, not a nonzero exit on refusal, so this is not a defect under this order.

### Conditions
Correct the four documented command paths before presenting the reproduction as runnable. Compare the original round-one condition text with the actual round-two diff before owner gating. Run the corrected check on macOS before claiming measured macOS support. No threshold exception was requested or granted.

### Unrun
macOS check: no macOS runner was available here. Exact round-one condition-by-condition audit: the original review text was not supplied. Full project Go suite, race detector and Go vet: outside this case-only review's scope; the case check and two negative controls were run instead.

## The review status as filed

## REVIEW — corpus-pv-013
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The check establishes the change-side refusal leak; both negative controls fail as intended. Correct stale reproduction commands and verify the original round-one conditions before gating.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-23T16:12:19Z

## REDACTED

none
