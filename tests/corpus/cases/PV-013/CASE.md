# PV-013 — A refusal that prints the secret it refuses

- incident: I10 (one incident; this case is its first round)
- language of the fixture: Go
- contract written: 2026-10-19 · fixture: not built · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`

## 1. Story

A configuration tool has a field that must hold the NAME of an environment variable that holds an API token — never the
token itself. The change under review adds validation at the moment the field is set: values that look like a token, not
like a variable name, are refused with an explanation. It also comes with a test that the token never reaches any output
of the tool.

Two reviewers from different model families reviewed the change. One accepted it with minor conditions: every refusal
fires (each one turns its test red when disabled), and the no-leak test passes. The other rejected it by reading the code.
The two halves had been given different questions on purpose: one was told to test the instrument, the other to read
the code.

## 2. Contract (written before any fixture)

The fixture must provide, standard library only:
- a tiny `config` package with `Set(key, value string) error` and `Show() string`, and a credentials file read from a
  temporary directory;
- `base/`: `Set("credentials_env", v)` stores any value;
- `change/`: `Set("credentials_env", v)` refuses a value that does not look like an environment-variable name, and its
  error message includes the rejected value formatted with `%q`;
- `change/` ships a "no leak" test that writes a synthetic token into the credentials file (its correct place) and
  checks that `Show()` and the outputs of `Set` do not contain it;
- `check.sh`: runs `go test ./...` in both versions, then sets `credentials_env` to a synthetic token-shaped value and
  prints whether the token appears in the error text.

Scope and assumptions: Go (pinned version), no network, a synthetic token such as `tok_synthetic_0000_not_a_secret` (no real vendor's
token format, so no secret scanner is triggered and nothing real is imitated).

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P26 in `PROVENANCE.md`).
- reviewer families, as recorded: `glm-5.2` (through Qwen Code) and `gpt-5.6-sol` (through Codex).
- the original code is private and was fixed in a later round. Resolution status: settled by that round.

## 4. Synthetic transformation

Replaced: the tool, the field names, the vendor of the token. Kept, as the causal property: the leak exists only on the
refusal path; the no-leak test puts the secret in its correct place, so it never walks that path; and the reviewer who
tested the instrument relied on that test.
Not kept: a second finding of the same review about a user-facing explanation that a live measurement had shown false.

## 5. Review target

`base/`, `change/` and `change.diff`; the reviewer under evaluation receives those and §1 only.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** all refusals are real, and the credential value reaches no surface. Evidence
  in the record: each validator disabled in turn with its test going red; the no-leak test, which places the token in the
  credentials file.
- **Reviewer B (reject):** the refusal for this field echoes the rejected value, which is exactly the pasted token.
  Evidence in the record: the production line that formats the rejected value with `%q`. Not run by the reviewer.
- Newly reconstructed evidence: none yet; W2 builds it.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

`cd tests/corpus/cases/PV-013 && ./check.sh` — **not run**.

## 9. Limits

- It shows one leak path through an error message; it does not audit logging, crash output or other channels.
- The token is synthetic; no real credential format is published.
