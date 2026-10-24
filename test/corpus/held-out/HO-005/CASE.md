# HO-005 — A CLI that serves another model and reports success

- set: held-out set 1 (frozen; see `../README.md`)
- intended fixture language: sh (a hint only; this contract is language-neutral)
- contract written: 2026-10-24 · case code: not built yet · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A small command-line tool asks a language-model service for an answer. The caller names the model it wants. The tool lists
the models it can serve. A wrapper script runs the tool for a batch of jobs and writes a line into a log for each job: the
job, and the model it asked for.

The change under review adds a fallback to the tool. Its description says: "if the requested model is busy, the tool must
not fail the job". The change makes the tool serve the answer from another model when the requested one is not in its list,
and exit 0.

Two reviewers read the change. One accepted it: jobs no longer fail. The other rejected it. They did not agree on whether
the fallback is safe.

## 2. Contract (written before any fixture)

The case must provide, with no network and no model:
- two versions of one tool, **base** and **change**, and one wrapper script that is the same in both. The tool takes a
  model name and a prompt and prints the answer, then exits.
  - **base** exits 2 and prints an error when the model is not in its list.
  - **change** serves the answer from its default model when the model is not in its list, prints the answer and exits 0.
    The answer itself carries no field that names the model that served it.
- a **stand-in service**: a local program that answers from a fixed table, one distinct answer text for each of two models,
  and lists those two models. No call leaves the machine.
- the **wrapper**: runs the tool for a list of jobs and appends one line to a log for each job that exited 0, naming the job
  and the model it asked for.
- a **reference**: the model that really served each job, which the case knows because it built the stand-in; never
  computed by the tool or the wrapper; used only by the check.
- a **check** that prints one line per job: job, model asked for, tool exit code, model logged, model that really served,
  reference label.

Inputs the check must include:
- a job that asks for a model in the list;
- a job that asks for a model not in the list.

Scope and assumptions: two listed models; one default model; no retries. The shell and tools are chosen when the case code
is written, pinned, and recorded in §8.

## 3. Provenance

- source: a build-window reviewer disagreement, 2026-10-23/24, described by the master; no source text is quoted.
- reviewer families: not recorded in this file.
- the original code is not published. Resolution status: not recorded here.

## 4. Synthetic transformation

Replaced: the tool, the service, the models, the wrapper, every name and every file. Kept, as the causal property: a CLI
asked for a model it does not list serves a different model and reports success, and the caller records the model it asked
for.

## 5. Review target

The two versions of the tool, the wrapper, the change's description and a readable diff between them. The reviewer under
evaluation receives only those and the description in §1 — never the check, the reference or `expected/`.

## 6. The two claims and their evidence

- **Reviewer A (accept):** jobs no longer fail when the requested model is missing. Evidence as described: a run that asks
  for a model not in the list and gets an answer.
- **Reviewer B (reject):** the answer comes from another model while the log names the requested one. Evidence as
  described: the log line and the answer text of the same job.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

To be recorded by the fixture round. **Not run**: the case code is not built yet.

## 9. Limits

- The case models a tool with no field that names the serving model. A tool that names it would be a different design.
- The case does not rule whether a fallback may exist; it asks whether a fallback may be silent.
