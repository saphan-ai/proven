# REVIEW — corpus-pv-013 · round 1

- verdict: reject
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 1 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:56:15Z

## Summary

The fixture demonstrates the expected refusal leak, but the check's leak flag cannot distinguish a change-side leak from a base-only leak and its exit status does not assert the expected outcome. Repair the instrument before relying on its green result.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-013

### Ground
`git rev-parse --short=7 HEAD` printed `bf6722d`. Compared `fe146e0...bf6722d`: only `tests/corpus/cases/PV-013/CASE.md`, `base/`, `change/`, `check.sh`, and the PV-013 row in `tests/corpus/manifest.json` changed. `expected/` did not change. The new `go.mod` files pin Go 1.26 and have no external dependencies or install scripts. I read §2 before running the check and read `expected/EXPECTED.md` only afterward.

`bash tests/corpus/cases/PV-013/check.sh` exited 0 on Linux and printed (redactions preserve the two occurrences of the synthetic credential value):

```
control | base: none | change: pass | token_in_output: no
valid-name | base: accepted | change: accepted | token_in_output: no
token-in-credentials-show | base: (empty) | change: (empty) | token_in_output: no
bad-token-env | base: accepted | change: refused: token_env must be an environment-variable name matching [A-Z][A-Z0-9_]*; rejected value: chat-token | token_in_output: no
bad-channel | base: accepted | change: refused: channel must be a channel name matching #[a-z0-9][a-z0-9_-]*; rejected value: not a channel | token_in_output: no
bad-endpoint | base: accepted | change: refused: endpoint must be an http or https URL with a host; rejected value: not-a-url | token_in_output: no
token-in-name-then-show | base: accepted / token_env=███████████████████████████████ | change: refused: token_env must be an environment-variable name matching [A-Z][A-Z0-9_]*; rejected value: ███████████████████████████████ / (empty) | token_in_output: yes
```

### Differential
The valid input is accepted by both; the three bad fields are accepted on base and refused on change. With the synthetic value in the credentials file, neither version shows it. Pasting it into the name field is accepted and shown on base; change refuses it, prints it in the refusal, and subsequently shows an empty configuration. This matches §2's intended difference and directly supports `expected/EXPECTED.md`'s established-defect outcome. The shipped control passes because it only sets legal values from a credentials-file scenario (`tests/corpus/cases/PV-013/change/main.go:184-209`), not the pasted-value refusal (`tests/corpus/cases/PV-013/change/main.go:122-125`). Do not infer that credentials-file show leaks, or that any refusal failed to fire.

### Unplug
**BLOCKING — `tests/corpus/cases/PV-013/check.sh:52-58,142`.** `report` combines base and change output before testing for the synthetic value. In the opposite world where base prints it but change's refusal is repaired, `token_in_output` is still `yes`. A read-only probe combining `base: [synthetic value]` with `change: refused` returned `yes`. The printed change-result text currently distinguishes the worlds, but the dedicated flag does not. Also, `problems` is set only for command failures (`check.sh:71-115`), not for a missing refusal or a leak; thus exit 0 reports execution success, not proof of the expectation. The remedy is to report per-side leak presence, assert the change-side refusal and leaked value separately for this reproduction, and show a negative control that goes red when either assertion is removed. Do not change the case's target code to make a check green.

The proposed local mutation that removed the name guard was not run: the command to create and then remove a temporary fixture was refused by the execution tool at the cleanup verb. I did not restate that denied cleanup via a different command.

### Closure
**Verdict: reject the current check as a reliable green signal, not the reproduced finding.** The observed change-side refusal visibly contains the synthetic value, so the expected defect is established by the full text. The Boolean flag alone and exit code 0 cannot establish it. `tests/corpus/cases/PV-013/check.sh:50-58,142` is the blocking location; fix the check, then demonstrate both the normal output and its failure on the negative control. No code under review was changed.

### Remainder
No private machine, seat, internal path, private repository, real credential, or e-mail address was found in the reviewed diff. The value printed above is explicitly synthetic, but is redacted in this public review as a credential-shaped string. `https://chat.example/hook` is a placeholder. The Go build ran with `GOTOOLCHAIN=local`, `GOPROXY=off`, and `GOSUMDB=off`; there are no downloaded modules. No benchmark or project metric is declared for this case.

### Profile
Configuration/credential exposure: the boundary is between credentials and user-facing output. The change refuses invalid values while echoing the rejected argument (`tests/corpus/cases/PV-013/change/main.go:120-135`); the shipped control inspects only accepted-value outputs. Check ranking: the printed check is an author-owned suite check, not an external or independent auditor. No exception, baseline change, skip, or suppressed check was found. Declined to judge: unrelated product logging, crash paths, real services, licence policy, and the original historical review.

### Focus
The case uses Bash `local`, `mktemp -d` with a template, `grep -q -F`, `cat`, `pwd`, `rm -rf` in the temporary-directory exit trap, and `go build`; these have Linux and macOS equivalents with Bash and Go installed. The script is Bash-specific, not POSIX sh; no GNU-only option was identified. macOS was not available here. Review scope was limited to the one commit and its own manifest line.

### Conditions
Repair the combined leak flag and add a failing negative control as above; rerun the seven contract inputs and verify the explicit change-side refusal still prints the synthetic value. No accepted exception or downgrade of this blocking condition is recommended.

### Unrun
macOS execution: this host is Linux. Mutated-case execution: the tool refused the temporary fixture's cleanup act, so no mutation outcome is claimed. No network, external model, or historical private repository was accessed.

### REDACTED
- Synthetic credential-shaped value, 2 occurrences in the pasted check output; each 31-character value is replaced by 31 blocks. No real secret was observed.

## The review status as filed

## REVIEW — corpus-pv-013
- state: recommendation-ready
- verdict: reject
- ref: _REVIEW_NOTES.md
- abstract: The fixture demonstrates the expected refusal leak, but the check's leak flag cannot distinguish a change-side leak from a base-only leak and its exit status does not assert the expected outcome. Repair the instrument before relying on its green result.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T18:52:43Z

## REDACTED

none
