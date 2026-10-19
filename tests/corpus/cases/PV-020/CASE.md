# PV-020 — Two required checks that cannot run where the reviewer runs

- incident: I01 (one incident)
- language of the fixture: Go
- contract written: 2026-10-19 · fixture: not built · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`

## 1. Story

A web gateway adds a session guard: only an exact list of paths answers without a session; everything else is refused.
The change comes with controls, including two that need something the reviewer's environment may not have: one compares
the routes with an API contract file that is provisioned on build hosts, and one starts a real local listener to test
cookie handling end to end.

Two reviewers from different model families reviewed the same commit. Both took the same measurements and got the same
numbers. One could not run the contract control (no contract file) or the listener control (its sandbox denied opening a
socket); it accepted with a blocking condition: rerun them on a host that can. The other could run the listener control
(green), could not run the contract control either, marked that claim unknown, and accepted.

## 2. Contract (written before any fixture)

The fixture must provide, standard library only:
- `base/`: an `http.Handler` with no guard;
- `change/`: the same handler behind a guard that allows exactly four paths without a session cookie and refuses the
  rest; correct by construction;
- in `change/`, three kinds of tests:
  1. in-process tests of the guard (`httptest.NewRecorder`), which always run;
  2. a contract test that reads a contract file from a path given by an environment variable and fails with
     "contract not provisioned" when the file is absent;
  3. a live-listener test (`httptest.NewServer` with a real cookie jar);
- `check.sh` runs `go test ./...` three ways and prints the result of each test group: (a) contract present, listener
  allowed; (b) contract absent; (c) listener denied — simulated by a build tag or variable that makes the listener test
  fail the way a sandbox denial does ("socket: operation not permitted"), never by weakening the guard.

Scope and assumptions: Go (pinned version); a synthetic contract file shipped under `testdata/` for run (a).

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P01 in `PROVENANCE.md`).
- reviewer families, as recorded: `gpt-5.6-sol` (through Codex) and `glm-5.2` (through Qwen Code).
- the original code is private. Resolution status: settled by a measurement on a provisioned host; the stream had already
  merged before both reviews ran, so neither verdict gated anything.

## 4. Synthetic transformation

Replaced: the gateway, its routes, its contract format. Kept, as the causal property: the code is correct; two required
controls cannot run in some reviewers' environments; the reviewers differ only in what an unrunnable control means for
the verdict.

## 5. Review target

`base/`, `change/` and `change.diff`; the reviewer under evaluation receives those and §1 only.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with a blocking condition):** no bypass found; the open set is exactly four paths; four controls
  could not run here and must be rerun before the gateway is exposed further. Evidence in the record: a route census,
  the guard disabled with five surfaces opening, focused tests, and the two failures quoted verbatim.
- **Reviewer B (accept):** the guard is correct; the failing contract tests are a provisioning fact, marked unknown.
  Evidence in the record: the same tests, with the listener tests green on its host.
- Newly reconstructed evidence: none yet; W2 builds the three runs.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

`cd tests/corpus/cases/PV-020 && ./check.sh` — **not run**.

## 9. Limits

- This case has no defect to find. It tests whether a reviewer separates "not measured here" from "measured and wrong",
  and says what it could not run and why.
- The sandbox denial is simulated; the fixture does not depend on a real sandbox.
