# REVIEW — corpus-pv-020 · round 1

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 0 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:03:13Z

## Summary

The case reproduces its contract and supports a human scope decision; correct the PV-020 manifest check status before landing.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-020

### Ground

Reviewed `fe146e0...8c55cee829936c5308b6393319b640d737d2856e` (`git rev-parse HEAD`). The diff changes only `tests/corpus/cases/PV-020/` outside `expected/` and PV-020's manifest entry. `git diff --check fe146e0...8c55cee` passed. The local toolchain was `go version go1.26.5 linux/amd64`; both fixture modules declare Go 1.26. CASE.md records an earlier run with 1.26.8, which I did not use.

I ran `bash tests/corpus/cases/PV-020/check.sh` on this checkout. It exited 0 and printed exactly:

```text
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
```

### Differential

At 2,049 bytes, base accepts all six offered routes. Change refuses the four non-exempt routes and accepts the two documented upload routes. This is the difference in CASE.md §2. The modules use the same route names and order; `check.sh` compares each pair and rejects a missing or extra route. The check's exit code means the commands completed; it deliberately does not turn a reported control failure into a shell failure.

### Unplug

The `limit-removed` run makes the in-process group fail on the limit-presence check and every route assertion, so the group observes the limit's removal. The refused-listener run sets a simulated listen error before binding; `Production()` is still passed to `RunListener`. The absent-description run points at a nonexistent file; it does not alter the service. A false world with a wrong real API description would make the description group fail when its route set, limit, or exemptions differ. A false world in which the service refuses all oversize routes would make the exempt-route assertions fail. The output does not prove that a deployment's own description exists or agrees; it proves agreement with the synthetic file shipped here. It also cannot settle the policy choice of whether an unavailable control blocks a merge.

### Closure

The expected outcome's factual core follows: no demonstrated limit defect, all controls pass when available, two environment conditions yield `could not run`, and the limit removal is detected. The remaining verdict threshold is a human scope decision. The historical claim that an orchestrator later ran controls on another host is outside this synthetic check and was not independently verified here.

### Remainder

**Condition — manifest state is false.** `tests/corpus/manifest.json:359` still says PV-020's checks are `not run`, while `CASE.md:93` records an exit-0 run and this review repeated it. A consumer of the manifest gets the opposite status. Set that field to a truthful completed-check value before landing. The manifest line is the control; the run above is the trigger and evidence. This is an order-scope metadata defect, not a product limit defect.

No private host or seat names, private paths, real credentials, or email addresses appeared in the reviewed diff. The loopback literal in `controls.go:246` is a public local-listener address, not an internal host. No threshold, skip, test exclusion, or assertion weakening appeared. No numeric coverage or lint bar is declared for this case; no such metric was inferred. The case's own suite check is more circular than an external check, so its result is supported by source inspection and the limit-removal control.

### Profile

Docs and Go test fixture. Linux check ran. The shell uses Bash, Git, and Go `-C`; the Go code uses standard-library `syscall.EPERM` and a loopback TCP listener. Those calls have Linux and macOS forms, but macOS execution was not measured. No dependencies or lockfiles changed; each fixture module uses the standard library.

### Focus

All six requested properties were inspected. The fixture matches §2, the expected factual outcome follows within the stated reach, the negative limit control fails as intended, the diff stays in lane, and the privacy scan found no private material. The portability claim is limited to source inspection on macOS. The review does not decide licences, as ordered.

### Conditions

Correct `tests/corpus/manifest.json:359` to record that PV-020's check has run. Do not treat my Linux run as a macOS result. No condition is proposed on the service limit itself.

### Unrun

macOS execution: no macOS host in this leg. Repository-wide Go suite and CI gates: the order confines this review to the PV-020 diff and its check; no repository-wide result is claimed. The historical provisioned-host resolution: its private record was not available to this leg.

## The review status as filed

## REVIEW — corpus-pv-020
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: The case reproduces its contract and supports a human scope decision; correct the PV-020 manifest check status before landing.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T17:58:54Z

## REDACTED

none
