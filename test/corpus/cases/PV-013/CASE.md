# PV-013 — A refusal that prints the secret it refuses

- incident: I10 (one incident; this case is its first round)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: built (2026-10-21, declared prior art) · checks: run (2026-10-23) · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A configuration tool sends notices to a chat service. One of its fields must hold the NAME of an environment variable
that holds the service's API token — never the token itself. The token belongs in a separate credentials file. The
change under review adds validation when a field is set: a wrong value is refused at once, with an explanation, instead
of failing later at send time. For the name field, a value that looks like a token, not like a variable name, is
refused. The change also ships a control that the token never reaches any output of the tool.

Two reviewers from different model families reviewed the same change. They were given different questions on purpose.
One was told to test the instrument. It accepted with minor conditions: every refusal fires, and the no-leak control
passes. The other was told to read the code, and was not allowed to run the tests. It rejected the change.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- two versions of one configuration tool, **base** and **change**, each offering two operations:
  1. given a field name and a value, set the field, and answer accepted, or refused with a message;
  2. show the configuration, as text.
- a credentials file, in a fresh temporary place for each run, that holds the token.
- **base** stores any value in any field.
- **change** refuses a bad value for several fields, the name field among them. For the name field it refuses a value
  that does not look like an environment-variable name. Its refusal message quotes the rejected value in full.
- **change** ships its own no-leak control. The control writes a synthetic token into the credentials file (its correct
  place), then checks that the show output and the messages of the set operation do not contain it.
- no reference implementation. The property is a plain text test: the synthetic token must not appear in any text the
  tool outputs.
- a **check** that runs the shipped controls on both versions, then runs every input below against base and change, and
  prints one line per input: input name, base result, change result, and whether the token appears in any output.

Inputs the check must include:
- a valid variable name in the name field;
- the token in the credentials file, followed by show;
- for each refusal the change adds, one bad value;
- the synthetic token `tok_synthetic_0000_not_a_secret` pasted into the name field, followed by show.

Scope and assumptions: one process, one temporary directory, no network. The token is synthetic and follows no real
vendor's format, so no secret scanner is triggered and nothing real is imitated. The toolchain is chosen when the case
code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P26 in `PROVENANCE.md`).
- reviewer families, as recorded: `glm-5.2` (through Qwen Code) and `gpt-5.6-sol` (through Codex).
- the original code was Go; it is not published. Resolution status: settled by a later round that conceded the finding
  and repaired it.

## 4. Synthetic transformation

Replaced: the tool, its fields and names, and the chat vendor. Kept, as the causal property of the disagreement: the
leak exists only on the refusal path; the shipped no-leak control puts the secret in its correct place, so it never
walks that path; and the reviewer who tested the instrument relied on that control.
Not kept: a second finding of the same review, about a user-facing explanation that a live measurement had shown false
(see Limits).

## 5. Review target

The two versions of the tool, with the change's own controls, and a readable diff between them. The reviewer under
evaluation receives only those and the description of the change in §1 — never the check or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** every refusal is real, and the credential value reaches no output. Evidence
  in the record: each refusal disabled in turn, with its control going red each time; the no-leak control, which places
  the token in the credentials file and searches the outputs. Its conditions were minor gaps in what the refusals catch.
- **Reviewer B (reject):** the refusal for the name field repeats the rejected value, and that value is exactly the
  pasted token, so the token is printed to the terminal. Evidence in the record: reading the production code, where the
  refusal message quotes the rejected value. The reviewer did not run it.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output.

Command, from the repository root:

```
bash test/corpus/cases/PV-013/check.sh
```

Toolchain: `go version` reports `go version go1.26.5 linux/amd64`. The case modules pin `go 1.26`. The check builds with `GOTOOLCHAIN=local`, `GOPROXY=off` and `GOSUMDB=off`.
Run at: 2026-10-23 02:20 UTC.
Exit code: 0.

```
control | base: none | change: pass | base_token_in_output: no | change_token_in_output: no
valid-name | base: accepted | change: accepted | base_token_in_output: no | change_token_in_output: no
token-in-credentials-show | base: (empty) | change: (empty) | base_token_in_output: no | change_token_in_output: no
bad-token-env | base: accepted | change: refused: token_env must be an environment-variable name matching [A-Z][A-Z0-9_]*; rejected value: chat-token | base_token_in_output: no | change_token_in_output: no
bad-channel | base: accepted | change: refused: channel must be a channel name matching #[a-z0-9][a-z0-9_-]*; rejected value: not a channel | base_token_in_output: no | change_token_in_output: no
bad-endpoint | base: accepted | change: refused: endpoint must be an http or https URL with a host; rejected value: not-a-url | base_token_in_output: no | change_token_in_output: no
token-in-name-then-show | base: accepted / token_env=tok_synthetic_0000_not_a_secret | change: refused: token_env must be an environment-variable name matching [A-Z][A-Z0-9_]*; rejected value: tok_synthetic_0000_not_a_secret / (empty) | base_token_in_output: yes | change_token_in_output: yes
```

Negative controls, from the repository root, at 2026-10-23 02:20 UTC:

```
bash test/corpus/cases/PV-013/check.sh --negative-control
```

Exit code: 1 (expected: base contains the token, but change does not).

```
negative-base-only | base: tok_synthetic_0000_not_a_secret | change: refused: token_env rejected | base_token_in_output: yes | change_token_in_output: no
```

```
bash test/corpus/cases/PV-013/check.sh --negative-control=no-refusal
```

Exit code: 1 (expected: change prints the token without a refusal).

```
negative-no-refusal | base: none | change: accepted: tok_synthetic_0000_not_a_secret | base_token_in_output: no | change_token_in_output: yes
```

## 9. Limits

- It shows one leak path, through a refusal message. It does not audit logs, crash output or other channels.
- The original review also found that a user-facing explanation stated as fact something a live measurement had shown
  false. That second finding is not reconstructed here.
- On base, a pasted token is stored and then shown as the name. That is the problem the change set out to fix, so base
  is not a leak-free baseline; the case is about the new path the change opens, not about base being safe.
- The token is synthetic; no real credential format is published.
