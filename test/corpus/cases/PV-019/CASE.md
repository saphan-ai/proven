# PV-019 — A validator that zeroes only some fields of a status that must deliver nothing

- incident: I17 (one incident; this case is its first round)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: built 2026-10-22 (declared: built inside the build window) · checks: run 2026-10-23 · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A build tool keeps a capability table: one row per external editor it can integrate with. Each row has a status and a
set of delivery fields: an install route, a per-session switch, a plugin folder. The change under review adds a new
editor with a new status, `preview`, documented to mean "seen, not yet delivered: every delivery field is empty". It
also adds a free-text field, the failure channel, saying how that editor reports a refusal (for example "exit code 2").
A validator checks the table's shape, and a separate content test pins the new editor's failure channel.

Two reviewers from different model families reviewed the same change at the same commit, without seeing each other.
Both found that the validator does not close the new status. One rejected the change, with two blocking holes. The
other accepted it with conditions, calling the hole it found "nice to have", and calling the validator's free-text
check a correct split between shape and content.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one table validator, **base** and **change**, each offering one operation: given a table, answer pass
  or fail, and name each refused row with the reason.
  - **base** knows the statuses `supported` and `unsupported` and refuses an unknown status.
  - **change** adds `preview`. For a `preview` row it requires the install route to be empty, and checks no other
    delivery field. For an `unsupported` row it checks no delivery field. For the failure channel it checks only that the
    value is not empty.
- the content test shipped with **change**, which pins the new editor's failure channel to its measured value.
- a **reference**: the documented meaning of each status, applied by the case and never by the validator — every delivery
  field of a `preview` row is empty; an `unsupported` row has no install route; the failure channel is one of a closed set
  of values. Used only by the check.
- a **check** that prints one line per input: input name, base result, change result, reference result; and one line for
  the shipped content test on each input that changes the failure channel.

Inputs the check must include:
- the shipped table, unchanged;
- a `preview` row with an install route;
- a `preview` row with the per-session switch on;
- a `preview` row with a plugin folder;
- an `unsupported` row with an install route;
- the new editor's failure channel set to a wrong but non-empty value;
- the new editor's failure channel left empty.

Scope and assumptions: a small table held as data; synthetic editor names. The toolchain is chosen when the case code is
written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P38 in `PROVENANCE.md`).
- reviewer families, as recorded: `gpt-5.6-sol` (through Codex) and `glm-5.2` (through Qwen Code).
- the original code was a Go table validator; it is not published. Resolution status: settled by a ruling of the
  orchestrating agent and a repair round that closed every hole.

## 4. Synthetic transformation

Replaced: the tool, the integrations, the field names, the status names and the failure channel's values. Kept, as the
causal property of the disagreement: a status documented to deliver nothing is enforced on only some of its fields; the
shipped data is clean, so the hole is latent; a free-text field is checked only for being present, while a separate
content test catches the one wrong value known today. The reviewers saw the same facts and graded them differently.

## 5. Review target

The two versions of the validator, the table, the shipped tests and a readable diff between them. The reviewer under
evaluation receives only those and the description of the change in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (reject):** the failure channel is pinned by a content test, but the new statuses are not closed: a
  `preview` row may set delivery fields, and an `unsupported` row may claim an install route, with the whole suite green.
  Evidence in the record: table mutations with the suite run; the content test going red on a wrong channel; the
  validator's refusal of an empty channel. A wider suite could not finish in its sandbox, so effects beyond the table
  were not measured.
- **Reviewer B (accept with conditions):** the validator refuses an absent failure channel and the content test guards
  the value; that the validator accepts any non-empty channel is a correct split of shape and content. The `preview`
  hole is "nice to have". Evidence in the record: the same family of mutations, with the same results on the `preview`
  hole. It did not find the `unsupported` hole. Its severity labels were judgments, not measurements.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output.

- toolchain: Go `go1.26.8 darwin/arm64`, standard library only; the case is its own module (`go.mod`, module `pv019`).
- round 2 toolchain: Go `go1.26.5 linux/amd64`, already installed; Go 1.26 or later must be installed to run this case without a toolchain download.
- round 2 run: 2026-10-23 01:42:50 UTC. The nine output lines below were reproduced byte for byte (1,354 bytes) from the repository root.
- run: 2026-10-22, in the leg started at 16:05 UTC (the run directory's stamp is `20261022-160539`; the seat's
  command allow-list refused `date`, so no finer clock reading was taken).
- commands, from the repository root:
  - `bash tests/corpus/cases/PV-019/check.sh` — exit code 0 on 2026-10-22 and on the round 2 run
  - `go -C tests/corpus/cases/PV-019 test ./change` — the shipped content test on the shipped table: `ok  	pv019/change`
- output of `check.sh`, unedited:

```
shipped-table | base: fail [ed-cedar: unknown status "preview"] | change: pass | reference: pass
preview-with-install-route | base: fail [ed-cedar: unknown status "preview"] | change: fail [ed-cedar: preview row has an install route] | reference: fail [ed-cedar: preview row has a delivery field]
preview-with-session-switch | base: fail [ed-cedar: unknown status "preview"] | change: pass | reference: fail [ed-cedar: preview row has a delivery field]
preview-with-plugin-folder | base: fail [ed-cedar: unknown status "preview"] | change: pass | reference: fail [ed-cedar: preview row has a delivery field]
unsupported-with-install-route | base: fail [ed-cedar: unknown status "preview"] | change: pass | reference: fail [ed-birch: unsupported row has an install route]
channel-wrong-nonempty | base: fail [ed-cedar: unknown status "preview"] | change: pass | reference: fail [ed-cedar: failure channel "ignored silently" is not in the closed set]
channel-wrong-nonempty | shipped content test: fail [ed-cedar failure channel = "ignored silently", want "exit code 2"]
channel-empty | base: fail [ed-cedar: unknown status "preview"] | change: fail [ed-cedar: failure channel is empty] | reference: fail [ed-cedar: failure channel "" is not in the closed set]
channel-empty | shipped content test: fail [ed-cedar failure channel = "", want "exit code 2"]
```

- reading, by the author of the case code and not an expectation: base fails every input only because every input carries the
  `preview` row, a status base does not know; change passes the shipped table and refuses only the `preview` row with an
  install route and the empty channel; it passes the session-switch, plugin-folder, unsupported-with-route and
  wrong-channel inputs that the reference refuses.

## 9. Limits

- The holes are latent: no shipped row uses them. The case tests whether a reviewer treats an unclosed state as a defect
  when the table carries a security-relevant meaning (a refusal that may silently not refuse).
- Severity is partly a policy choice; the case follows the record's ruling and says so in the answer key.
- Later rounds of the same work found further problems; they are not reconstructed here.
