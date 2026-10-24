# PV-020 — Two required checks that cannot run where the reviewer runs

- incident: I01 (one incident; this case is a second review of a change that had already merged)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: built (2026-10-21, declared prior art) · checks: run (2026-10-23) · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A small web service accepts requests on many routes. The change under review adds a size limit: a request whose body is
larger than the limit is refused on every route, except on an exact, documented list of upload routes. The change is
correct. It comes with three kinds of controls. Two of them need something a reviewer's environment may not have. One
compares the service's routes and limits with an API description that exists only on provisioned build hosts. The other
starts a real local listener and sends real requests over a connection, because an in-process call cannot exercise the
way a real client streams a body.

Two reviewers from different model families reviewed the same commit, without seeing each other. They took the same
measurements and got the same numbers. The first could run neither of the two controls: the description was absent, and
its environment refused to open a listener. It accepted with a blocking condition: rerun both on a host that can. The
second could run the listener control, and it passed. It could not run the description control either. It marked that
claim unknown and accepted.

## 2. Contract (written before any code)

The case must provide, with no network beyond the local machine and no model:
- two versions of one service, **base** and **change**, each offering one operation: given a route and a request body,
  answer accepted or refused, with the reason.
- **base** has no size limit.
- **change** refuses a body above the limit on every route except the exempt list, and accepts it there. It is correct by
  construction: the case must contain no defect in the limit.
- three controls in **change**:
  1. in-process controls of the limit, which always run;
  2. a description control, which reads an API description from a location given by configuration and stops with the
     message "description not provisioned" when it is absent;
  3. a listener control, which starts a real local listener and sends real requests to it.
- a synthetic API description, shipped with the case, for the run in which it is present.
- a **check** that runs the controls three ways and prints one line per control group and run: run name, control group,
  result (passed, failed, or could not run with its reason):
  (a) description present, listener allowed;
  (b) description absent;
  (c) listener refused. The refusal is simulated so that the listener fails to start the way a restricted environment
  refuses it ("operation not permitted"). It is never simulated by weakening the limit.

The check must also include one line per route showing base and change results for an oversize body, and one run of the
in-process controls with the limit removed, to show the controls notice its absence.

Scope and assumptions: one local machine; no restricted environment is needed, because the refusal is simulated. The
toolchain is chosen when the case code is written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P01 in `PROVENANCE.md`).
- reviewer families, as recorded: `gpt-5.6-sol` (through Codex) and `glm-5.2` (through Qwen Code).
- the original code is private and is not published. Resolution status: settled by a measurement on a provisioned host.
  The change had already merged before both reviews ran, so neither verdict gated anything.

## 4. Synthetic transformation

Replaced: the original service, what its change protected, its routes and the format of its description. Kept, as the
causal property of the disagreement: the code is correct; two required controls cannot run in some reviewers'
environments, for two different reasons (a missing provisioned file, a refused listener); and the reviewers differ only
in what an unrunnable control means for the verdict. Not kept: a non-blocking hardening remark (see Limits).

## 5. Review target

The two versions of the service, their controls, and a readable diff between them. The reviewer under evaluation
receives only those and the description of the change in §1 — never the check, the synthetic description or `expected/`.
It runs the controls in its own environment, whatever that environment allows.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (accept with a blocking condition):** no way around the limit was found; the exempt set is exactly the
  documented list; with the limit removed, more routes accept an oversize body. Four controls could not run in its environment, two for
  each reason, and must be rerun on a provisioned host that allows listeners before the service is exposed further.
  Evidence in the record: an independent route census, a run with the limit removed, focused controls, the two failure
  messages quoted word for word, and a count of control names before and after the change (none removed). Its measurements
  were accurate. Its condition rested on an absence of evidence caused by its environment, not on a demonstrated defect.
- **Reviewer B (accept):** the limit is correct; the failing description controls are a provisioning fact, not a code
  regression, so their claims are marked unknown and do not block. Evidence in the record: the same controls, with the
  listener controls green on its host, plus a wider surface checked by hand. It closed one of A's two unknowns by running
  it. The description agreement stayed unmeasured, exactly as for A.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check and record each line of its output for runs (a), (b) and (c).

Recorded 2026-10-23T02:24:28Z on Linux. Toolchain: `go version go1.26.5 linux/amd64`. `base/go.mod` and `change/go.mod` declare `go 1.26`. Standard library only. The same command first ran on the unchanged tracked tree at `8c55cee` (exit 0, with the 16 lines below byte for byte), then ran after this round's edits (exit 0, same output). macOS was not run: no macOS host was available.

The check needs Bash and Go 1.26 or later. It can run from any directory in a checkout or source archive. It writes temporary build outputs under `.saphan/work/corpus-pv-020/` inside that tree.

Command, from the repository root:

    bash test/corpus/cases/PV-020/check.sh

Exit code: 0

    run a description-present-listener-allowed control in-process result passed
    run a description-present-listener-allowed control description result passed
    run a description-present-listener-allowed control listener result passed
    run b description-absent control in-process result passed
    run b description-absent control description result could not run: description not provisioned
    run b description-absent control listener result passed
    run c listener-refused control in-process result passed
    run c listener-refused control description result passed
    run c listener-refused control listener result could not run: operation not permitted
    route GET /v1/health oversize base accepted (accepted) change refused (body exceeds limit)
    route POST /v1/items oversize base accepted (accepted) change refused (body exceeds limit)
    route POST /v1/search oversize base accepted (accepted) change refused (body exceeds limit)
    route PUT /v1/items/current oversize base accepted (accepted) change refused (body exceeds limit)
    route POST /v1/uploads/media oversize base accepted (accepted) change accepted (exempt upload)
    route POST /v1/uploads/bundle oversize base accepted (accepted) change accepted (exempt upload)
    run limit-removed control in-process result failed: inproc.limit-present, inproc.over-limit-refused GET /v1/health, inproc.over-limit-refused POST /v1/items, inproc.over-limit-refused POST /v1/search, inproc.over-limit-refused PUT /v1/items/current, inproc.over-limit-exempt POST /v1/uploads/media, inproc.over-limit-exempt POST /v1/uploads/bundle

## 9. Limits

- This case has no defect to find. It tests whether a reviewer separates "not measured here" from "measured and wrong",
  and says what it could not run and why.
- The refused listener is simulated; the case does not depend on a real restricted environment.
- Reviewer A also raised a non-blocking hardening remark about how the open exceptions are keyed. The record never
  settled it, and it is not reconstructed here.
