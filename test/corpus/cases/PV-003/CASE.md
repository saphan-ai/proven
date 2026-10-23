# PV-003 — A cooldown keyed on a reusable slot instead of the identity it limits

- incident: I03 (one incident; this case is its first round; a later round is PV-004)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: built 2026-10-22 (declared: built inside the build window) · checks: run 2026-10-22 · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A job service runs tenant workers in numbered slots on each node. When a tenant uses up its upstream quota, the service
sets a cooldown: no new session opens for that tenant until the cooldown ends. An operator may legally move a tenant
to another slot on the same node; the latest assignment wins. Every run record carries the tenant's identity as well as
its node and slot.

The cooldown used to be held under one key per node, so two tenants on different slots of one node could clear each
other's cooldown. The change under review keys the cooldown on (node, slot), with one table per slot. Its tests place
tenants on fixed slots and show that two tenants no longer clear each other. The change's description says that
(node, slot) is the only granularity the service has.

Two reviewers from different model families reviewed the same change at the same commit, without seeing each other.
Both found the same thing: when a tenant that is in cooldown moves to another slot, it can open a session, and the next
tenant placed on its old slot is refused in its name. One accepted the change with conditions and asked for that limit
to be documented. The other rejected it.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- two versions of one cooldown gate, **base** and **change**, each offering three operations:
  1. assign a tenant to a (node, slot); a later assignment of the same tenant replaces the earlier one;
  2. record a run: tenant, node, slot, and whether the run used up the quota;
  3. ask whether a session may open on a (node, slot) now: allowed, or refused naming the tenant whose cooldown applies.
- **base** holds the cooldown under one key per node.
- **change** holds the cooldown under (node, slot) and ignores the tenant identity on the run record.
- a **reference** gate that holds the cooldown under the tenant identity — the rule the service states: a cooldown
  belongs to the tenant whose quota ran out. Used only by the check.
- a **check** that runs every input below against base, change and the reference, and prints one line per input:
  input name, base result, change result, reference result.

Inputs the check must include:
- the change's own cases: two tenants on two slots of one node, one of them in cooldown, with no move;
- a move: tenant X uses up its quota on slot 1, then is moved to slot 2 on the same node, and asks to open a session;
- a vacated slot: after that move, tenant Y is assigned to slot 1 and asks to open a session;
- a control with no cooldown at all (all three must allow).

Scope and assumptions: a single process, a fake clock, and in-memory records. The toolchain and any library are chosen
when the case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P03 in `PROVENANCE.md`).
- reviewer families, as recorded: `claude-opus-5` (through Claude Code) and `gpt-5.6-sol` (through Codex).
- the original code was Go; it is not published. Resolution status: settled by a later round that adopted the reject's
  finding and re-keyed the cooldown on the identity.

## 4. Synthetic transformation

Replaced: the service, its domain, its names and its records. Kept, as the causal property of the disagreement: a
per-identity cooldown is keyed on a reusable location; a legal move reassigns identities to locations; the run records
already carry the identity; and the shipped tests cover only fixed placements. Both reviewers measured the same
counterexample and weighed it differently.
Not kept: the same reviews' finding about a fenced-block extractor (its later form is PV-004), and the accepting
reviewer's other conditions (an unguarded refusal and a documentation control that cannot see a deleted flag).

## 5. Review target

The two versions of the gate, the change's tests, and a readable diff between them. The reviewer under evaluation
receives only those and the description of the change in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with conditions):** the (node, slot) key is correct for fixed placements, with a real negative
  control. A tenant moved to another slot leaves its cooldown behind, so the gate opens for it; this is not a
  regression and needs only a documented limit. The change's claim that (node, slot) is the only granularity is false,
  because the run records carry the identity. Evidence in the record: restoring the old key turns named controls red;
  two scratch controls, one for the move and one for the vacated slot.
- **Reviewer B (reject):** the cooldown follows a reusable slot, not the tenant. A legal move lets the tenant in
  cooldown open a session, and the next tenant on the old slot is refused in its name. Evidence in the record: an
  adversarial test of the move (moved tenant allowed; replacement refused, naming the old tenant), and the source
  showing that the gate ignores the identity on the run record. Its full comparison of the whole package on base and
  change did not complete, and it said so.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output.

- run date: 2026-10-22 (UTC). The time of day was not recorded: the seat's permission class denies the `date` program.
- toolchain: `go version go1.26.5 linux/amd64`; standard library only; no network, no model.
- layout: `base/gate.go`, `change/gate.go`, `check/main.go` and `check/reference.go` (the reference gate), `check.sh`.
- `gofmt -l tests/corpus/cases/PV-003`: no output. `go vet ./tests/corpus/cases/PV-003/...`: no output.
- command, from the repository root: `bash tests/corpus/cases/PV-003/check.sh` (runs `go run ./tests/corpus/cases/PV-003/check`).
- exit code: 0.

Output, in full:

```
own-case-cooled-slot | base=allowed | change=refused(A) | reference=refused(A)
own-case-other-slot | base=allowed | change=allowed | reference=allowed
move-moved-tenant | base=refused(X) | change=allowed | reference=refused(X)
vacated-slot | base=refused(X) | change=refused(X) | reference=allowed
control-no-cooldown | base=allowed | change=allowed | reference=allowed
```

Each line is: input, base result, change result, reference result. Inputs: `own-case-*` are the change's own cases
(A on slot 1 uses up its quota, B on slot 2 runs without using up its own, no move; asked on slot 1 and on slot 2);
`move-moved-tenant` asks on slot 2 after X uses up its quota on slot 1 and moves to slot 2; `vacated-slot` asks on slot 1
after that move and after Y is assigned to slot 1; `control-no-cooldown` has no run that uses up the quota.
Build choices the contract left open: a run that does not use up the quota clears the cooldown held under that run's key
(node for base, (node, slot) for change, tenant for the reference); the cooldown is one minute on a fake clock that is
never advanced; an unassigned slot is asked about with no tenant on it.

Round 2 recheck: 2026-10-23 08:43 UTC, on Linux with `go version go1.26.5 linux/amd64`. From the repository root,
`bash tests/corpus/cases/PV-003/check.sh` exited 0 and printed, in full:

```
own-case-cooled-slot | base=allowed | change=refused(A) | reference=refused(A)
own-case-other-slot | base=allowed | change=allowed | reference=allowed
move-moved-tenant | base=refused(X) | change=allowed | reference=refused(X)
vacated-slot | base=refused(X) | change=refused(X) | reference=allowed
control-no-cooldown | base=allowed | change=allowed | reference=allowed
```

Not run: native macOS check. This seat runs Linux and has no macOS runner. The script uses Bash and Go only;
source inspection found no GNU-only flag. Linux output does not establish the macOS result, so the review's
cross-platform verification condition remains open.

## 9. Limits

- It shows one keying error on one legal transition; it does not show that every location-keyed limit is wrong.
- The reference encodes the service's stated rule; it is not an independent measurement.
- A later round asked whether the identity itself can be reassigned; that was ruled out of scope and is not part of
  this case.
