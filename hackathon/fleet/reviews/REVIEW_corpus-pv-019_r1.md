# REVIEW — corpus-pv-019 · round 1

- verdict: accept-with-conditions
- reviewer: codex, gpt-6-sol
- cost: not recorded
- conditions: 0 BLOCKING · 1 OWED · 0 NOTE
- reviewed: see the notes below
- updated: 2026-10-22T18:16:03Z

## Summary

PV-019 reproduces the delivery-field holes and the failure-channel gap. The expected defect follows from the printed differential. Align the manifest's check status with the recorded run before landing.

## The notes as filed

The reviewer's own notes, verbatim: base against tip, the check broken on purpose, conditions and what was not run are in them.

## REVIEW NOTES — corpus-pv-019

### Ground

Reviewed `fe146e0...5b65589531ff66e623b98dae3078062d7e219c1b`; `git rev-parse --short=7 HEAD` printed `5b65589`. Contract: `tests/corpus/cases/PV-019/CASE.md` §2. I ran `bash tests/corpus/cases/PV-019/check.sh` on this tip with Go `go1.26.5 linux/amd64`; it exited 0 and printed:

```text
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

`go -C tests/corpus/cases/PV-019 test -count=1 ./...` passed: one test in `pv019/change`; three packages had no test files. The module pins Go 1.26 and uses only the standard library. The author's recorded run used Go 1.26.8 on macOS; mine used Go 1.26.5 on Linux.

### Differential

`base/validator.go:13-18` refuses the preview status on every input. That base result is identical for the shipped and mutated inputs, so it cannot distinguish the three delivery holes. The distinguishing instrument is change versus the separate reference. `change/validator.go:23-33` checks only the preview install route and merely nonempty failure channels; it checks no unsupported delivery field. `check/main.go:69-74` changes one named field per input. The shipped table passes both change and reference. The session switch, plugin folder, and unsupported install route pass change and fail reference. The wrong nonempty channel passes the validator and fails both reference and the content test. The empty channel fails all three relevant checks. These are the differences §2 specifies; I found no extra disagreement in these inputs.

### Unplug

I copied the case into `█████████████████████████████████████████████████████████` and changed only the preview guard to include `SessionSwitch`. The check's `preview-with-session-switch` line changed from `change: pass` to `change: fail`, while the script still exited 0. Thus the printed differential detects that fix, but the exit status does not. The script is a report for a reader, as §2 requests, not an assertion gate. A false world with no delivery holes would print `change: fail` beside the reference's `fail` for the three affected inputs, so its answer differs from the observed one. A false world in which only the known channel value is protected could still show the channel validator gap; the content-test line shows that known wrong value is caught. The check does not by itself prove the policy severity assigned to that gap.

### Closure

Verdict: accept-with-conditions. The fixture produces the required lines, and `expected/EXPECTED.md` names exactly the three delivery mismatches and the wrong-channel split visible above. It correctly says the shipped table is clean and the known wrong channel is caught by the content test. The expected class, established validator defect, follows from change versus reference. Its blocking grade also cites the historical ruling and repair; I did not verify that private history and do not derive the grade solely from the check.

### Remainder

OWED — `tests/corpus/manifest.json:342` still says `"checks": "not run"` after this commit changes `CASE.md:5,88-120` to say the check ran and records its output. Set the PV-019 manifest check field to the accurate run status. This is a public metadata contradiction, not a failed reproduction. Control: the two files disagree on the same case; my rerun also completed.

No secret, credential, email address, internal host or seat name, private repository name, or absolute private path appeared in the added case files or the changed manifest line. The case says only that its provenance is in a private record; it does not expose a path. The changed paths are all under PV-019 outside `expected/`, plus PV-019's manifest line. `check.sh` uses Bash, `dirname`, `cd`, and `go run`; none is Linux-only or macOS-only. Go 1.26 must be present. No declared numeric quality bar was found for this case; the sole shipped test is a suite check and does not cover the mutated delivery rows. The check's printed differential provides that coverage for a human reader.

### Profile

Docs and Go test fixture. The case is a standalone Go module. No dependency lockfile, downloaded package, install script, network call, or product runtime path changed. The check is a project-owned report; the shipped test is the more circular suite check. No external metric or baseline was declared, so a numeric testing-bar comparison is not applicable. Declined to judge: licences, private historical verdicts and repair, and corpus cases other than PV-019.

### Focus

The fixture reproduces §2 without bending an input or reference to make a line pass. `table.Shipped()` supplies all cases; each mutation edits one target field. The reference rules in `check/main.go:23-38` implement the status meanings and closed channel set from §2, independent of the validator. The content test in `change/content_test.go:10-12` uses the shipped table; the check explicitly calls the same `ContentTest` on both altered-channel inputs. The clean shipped row and the mutated rows establish a latent validator hole, not a bad shipped row.

### Conditions

Before landing, update PV-019's `checks` status in `tests/corpus/manifest.json:342` to agree with the documented run in CASE.md. No case-code repair is required by this review.

### Unrun

macOS execution was not run on this Linux seat; portability was checked by reading the commands and standard-library use. The private provenance and historical repair were not available in this review boundary. The full repository test suite was not run because the order confines this review to the PV-019 diff; the standalone PV-019 module suite ran.

## The review status as filed

## REVIEW — corpus-pv-019
- state: recommendation-ready
- verdict: accept-with-conditions
- ref: _REVIEW_NOTES.md
- abstract: PV-019 reproduces the delivery-field holes and the failure-channel gap. The expected defect follows from the printed differential. Align the manifest's check status with the recorded run before landing.
- artifacts:
    - _REVIEW_STATUS.md
    - _REVIEW_NOTES.md
- updated: 2026-10-22T18:12:00Z

## REDACTED

| what | times |
|---|---|
| absolute or private path | 1 |
