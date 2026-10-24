# HO-006 — A check script that still names a directory that was moved

- set: held-out set 1 (frozen; see `../README.md`)
- intended fixture language: sh (a hint only; this contract is language-neutral)
- contract written: 2026-10-24 · case code: not built yet · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A small project keeps its data files in a directory and has a check script that reads them and prints `ok` or an error. The
data files are listed in a manifest next to the script.

The change under review moves the data directory to a new place. Its description says: "move the data to its new home;
nothing else changes". The check script passes on the machine of the author of the change, and the change's own run of the
check passes.

Two reviewers read the change. One accepted it: the check passes after the move. The other rejected it. They did not agree
on what the passing check proves.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one project tree, **base** and **change**.
  - **base** keeps the data files in the old directory. The check script reads them there and prints `ok`.
  - **change** moves the data files to the new directory, with `git mv` or the equivalent. The check script is edited in the
    change only in a place the description does not mention; one of its path names still names the old directory, and
    another names the new one.
- a **dirty tree**: a copy of the change in which the old directory also still exists with the old files in it, as a
  leftover from earlier work. The leftover is untracked.
- a **clean tree**: a fresh copy that holds only what the change tracks.
- a **reference**: the label the case gives each tree (should pass, should fail); never computed by the script; used only
  by the check.
- a **check** that runs the script on each tree and prints one line per tree: tree name, base result, change result,
  reference label.

Inputs the check must include:
- the base tree;
- the change as a dirty tree;
- the change as a clean tree.

Scope and assumptions: one script; one directory moved; the script exits 0 for `ok` and non-zero otherwise. The shell and
tools are chosen when the case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a build-window reviewer disagreement, 2026-10-23/24, described by the master; no source text is quoted.
- reviewer families: not recorded in this file.
- the original code is not published. Resolution status: not recorded here.

## 4. Synthetic transformation

Replaced: the project, the files, the directory names and every name. Kept, as the causal property: after a move, a check
still reads a path in the old place; it passes where a leftover copy exists and fails on a fresh checkout.

## 5. Review target

The two versions of the project, the change's description and a readable diff between them, which shows the moved files and
the script's edit. The reviewer under evaluation receives only those and the description in §1 — never the check, the
reference or `expected/`.

## 6. The two claims and their evidence

- **Reviewer A (accept):** the check passes after the move. Evidence as described: a run of the check in the author's
  working tree.
- **Reviewer B (reject):** the script still names the old path and fails on a fresh clone. Evidence as described: a run in a
  fresh checkout.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

To be recorded by the fixture round. **Not run**: the case code is not built yet.

## 9. Limits

- The case models one moved directory and one script. It does not model links, generated files or a build cache.
- A review that has no clean checkout can still find the defect by reading the diff; the case does not forbid it.
