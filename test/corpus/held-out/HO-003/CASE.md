# HO-003 — A tree copy whose file modes depend on the umask

- set: held-out set 1 (frozen; see `../README.md`)
- intended fixture language: Go (a hint only; this contract is language-neutral)
- contract written: 2026-10-24 · case code: not built yet · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A small tool copies a directory tree, for example to make a read-only snapshot of a build. The snapshot must be readable by
the owner, the group and others, and must not be writable by the group or by others: every file in the copy has mode `0644`
and every directory `0755`, whatever mode its source had.

The change under review rewrites the copy. Its description says: "the copy adds the read bits for everyone and leaves the
rest as it was". The change's own test copies a small tree and checks that every file and directory in the copy has the mode
the description promises.

Two reviewers read the change. One accepted it: the test passes. The other rejected it. They did not agree on what the test
proves.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one copy function, **base** and **change**. Each takes a source tree and a destination, copies the tree,
  and returns an error or nil.
  - **base** copies each entry and gives every file and directory the mode `0644` and `0755` respectively.
  - **change** copies each entry, then sets the mode to the mode the new entry has OR the read bits for owner, group and
    others. New entries are created with the language's ordinary creation call.
- a **source tree**: a small tree made by the case, with files of mode `0600` and `0640`, and directories of mode `0700`.
- a **reference**: for each entry, the mode the snapshot must have: `0644` for every file and `0755` for every directory,
  under every umask the check uses and whatever the source mode; written as literals, never computed by the copy; used
  only by the check.
- a **check** that runs the copy under two umasks, `022` and `002`, each in a separate process so one run does not change
  the other, and prints one line per entry and umask: entry name, umask, base mode, change mode, reference mode.

Inputs the check must include:
- the same source tree under umask `022`;
- the same source tree under umask `002`.

Scope and assumptions: a Unix-like system; no symlinks, no special files, no ownership change. The toolchain is chosen when
the case code is written, pinned, and recorded in §8. The test the change brings is part of the change and is shown to the
reviewer.

## 3. Provenance

- source: a build-window reviewer disagreement, 2026-10-23/24, described by the master; no source text is quoted.
- reviewer families: not recorded in this file.
- the original code is not published. Resolution status: not recorded here.

## 4. Synthetic transformation

Replaced: the tool, the tree, every name and every file. Kept, as the causal property: a copy that adds read bits and keeps
the group-write bit its creation call left in place, so the result depends on the umask of the process; and a test that
passes under one umask and fails under the other.

## 5. Review target

The two versions of the copy, the change's test, the change's description and a readable diff between them. The reviewer
under evaluation receives only those and the description in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence

- **Reviewer A (accept):** the change's test passes. Evidence as described: a run of the test in one environment.
- **Reviewer B (reject):** the same test fails in another environment. Evidence as described: a run with the group-write
  bit left on.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

To be recorded by the fixture round. **Not run**: the case code is not built yet.

## 9. Limits

- The case needs a way to run a process under a chosen umask; the way is written in §8 by the fixture round.
- The case does not model access control lists, other file systems, or a platform with no umask.
- Which mode is the right one is the description's choice; the case does not rule on it.
