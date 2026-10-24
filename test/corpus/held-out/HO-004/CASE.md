# HO-004 — A client that reads every 422 as "already exists"

- set: held-out set 1 (frozen; see `../README.md`)
- intended fixture language: sh (a hint only; this contract is language-neutral)
- contract written: 2026-10-24 · case code: not built yet · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A small setup script prepares a service through its web API. It creates a user, then it adds that user to a team. The
script is meant to be run again and again, so creating a user that already exists must not stop it.

The change under review makes the script safe to re-run. Its description says: "when the API answers 422 to the create
call, the user already exists, so go on". The script then adds the user to the team.

Two reviewers read the change. One accepted it: a second run now works. The other rejected it. They did not agree on what a
422 means.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one setup script, **base** and **change**. Each runs the two calls in order, and prints `ok` and exits 0,
  or prints the failing step and exits 1.
  - **base** stops at any answer to the create call that is not 201.
  - **change** treats 201 and 422 the same way and goes on.
- a **stand-in API**: a local program or a set of recorded answers that behaves like the service, with no network. For the
  create call it answers 201 for a new user, 422 for a user that exists, and 422 for an input it refuses (an email address
  it will not take). It answers the team call 204 when the user exists and 404 when it does not. The body of each 422 says
  which of the two reasons applies.
- a **reference**: the label the case gives each input (should succeed, should fail at create); never computed by the
  script; used only by the check.
- a **check** that prints one line per input: input name, base result, change result, last step reached, reference label.

Inputs the check must include:
- a new user with a good email;
- a user that already exists;
- a new user with an email the API refuses.

Scope and assumptions: one create call and one team call; no authentication. The shell and tools are chosen when the case
code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a build-window reviewer disagreement, 2026-10-23/24, described by the master; no source text is quoted.
- reviewer families: not recorded in this file.
- the original code is not published. Resolution status: not recorded here.

## 4. Synthetic transformation

Replaced: the service, the API, the user and team, every name and every file. Kept, as the causal property: one status
code stands for two different causes, the client reads it as only one, and the object is never created, so a later step
fails with 404.

## 5. Review target

The two versions of the script, the stand-in API's documented answers as the change's description gives them, and a
readable diff between them. The reviewer under evaluation receives only those and the description in §1 — never the check,
the reference or `expected/`.

## 6. The two claims and their evidence

- **Reviewer A (accept):** a second run now works. Evidence as described: a run against a user that already exists.
- **Reviewer B (reject):** a 422 can also mean the input was refused; then the user is never created and the team step
  fails. Evidence as described: a run with a refused email that ended at the team step with 404.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

To be recorded by the fixture round. **Not run**: the case code is not built yet.

## 9. Limits

- The case models one API. It does not rule on how any real API should use 422.
- The error text a user sees is not part of the case; the step reached and the exit code are.
