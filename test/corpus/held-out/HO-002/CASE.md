# HO-002 — A gate that trusts a success marker from any account

- set: held-out set 1 (frozen; see `../README.md`)
- intended fixture language: sh (a hint only; this contract is language-neutral)
- contract written: 2026-10-24 · case code: not built yet · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A small project merges pull requests through a forge. A workflow runs on each pull request. When a run succeeds, a bot
posts a comment that holds a marker line. The marker names the run and the commit, for example
`verified: run <id> commit <sha>`. A merge gate script reads the comments of the pull request and decides whether the pull
request may merge.

The change under review makes the gate strict about the run. Its description says: "the gate passes only when a marker
names the current run and the current commit". The change does that: it compares both values.

Two reviewers read the change. One accepted it: the marker is matched to the current run and commit. The other rejected
it. They did not agree on whether matching the marker's text is enough.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one gate script, **base** and **change**. Each takes the recorded state of one pull request and prints
  `pass` or `fail`, and exits 0 for pass and 1 for fail.
  - **base** passes when a comment holds a marker line for any run of the pull request.
  - **change** passes only when a comment holds a marker line whose run id is the current run and whose commit is the
    current commit.
- a **recorded state**: a small set of files that stands in for the forge. It holds the current run (id, commit, status)
  and the comments in order, each with an author and a body. No call leaves the machine.
- a **reference**: the label the case gives each recorded state (should pass, should fail); never computed by the gate;
  used only by the check.
- a **check** that prints one line per state: state name, base result, change result, reference label.

States the check must include:
- the current run succeeded; the bot posted the exact marker for that run and commit;
- the current run failed or posted nothing; no marker exists;
- the current run posted nothing; a person's account posted the exact marker for the current run and commit;
- the current run succeeded and posted its marker; a person's account posted a marker for an older run.

Scope and assumptions: one pull request; one workflow; one bot account whose name the state records. The marker's text is
public: any person who can read the run can write the same text. The shell and tools are chosen when the case code is
written, pinned, and recorded in §8.

## 3. Provenance

- source: a build-window reviewer disagreement, 2026-10-23/24, described by the master; no source text is quoted.
- reviewer families: not recorded in this file.
- the original code is not published. Resolution status: not recorded here.

## 4. Synthetic transformation

Replaced: the forge, the workflow, the bot, the marker format, every name and every file. Kept, as the causal property: a
gate that checks what a marker says and not who posted it, so the source of the evidence is never checked.

## 5. Review target

The two versions of the gate, the change's description and a readable diff between them. The reviewer under evaluation
receives only those and the description in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence

- **Reviewer A (accept):** the marker is matched to the current run and commit, which is what the description asks.
  Evidence as described: the gate fails on a marker for an older run.
- **Reviewer B (reject):** a person can post the exact marker after a run that posted nothing, and the gate passes.
  Evidence as described: a state in which a person's account holds the matching marker.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

To be recorded by the fixture round. **Not run**: the case code is not built yet.

## 9. Limits

- The case does not model a forge's own permissions, signed comments or a bot with a token. It models only an author
  field.
- The marker's text is assumed knowable; a secret marker would be a different design.
- The case says nothing about any real forge.
