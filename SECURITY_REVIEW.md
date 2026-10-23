# Security review: how Proven's code is checked as a whole

Every change to Proven is reviewed on its own: two independent reviews of its diff before it lands. A diff review sees
one change at a time. It does not see a property that no single change breaks but that the code as a whole does not
keep. So the code is also reviewed as a whole, in pieces, each piece one boundary where something untrusted meets
something Proven protects. Decided on 23 October 2026
([ADR-0028](fleet/decisions/ADR-0028-proven-is-security-reviewed-as-a-whole-in-pieces.md)). What Proven protects and how:
[SECURITY.md](SECURITY.md).

## The pieces

Each piece names the property that must hold. A review of a piece reads every place in the code where that boundary is
crossed, not only the files where it is expected.

| # | Piece | The property that must hold | Status |
|---|---|---|---|
| 0 | **Tools first** | `govulncheck`, `gosec`, `staticcheck` and Docker Scout run over the whole repository; their output is the baseline the other pieces start from | not started |
| 1 | **The Prover's container** | code from the change runs with no network, no secret and bounded resources; nothing in the change can shape the container's arguments or reach beyond its read-only checkout | not started |
| 2 | **Secrets** | no key reaches the trace, the logs, the reports, a pull-request comment or a container; a missing key is named, never printed | not started |
| 3 | **The reviewed repository as input** | a git range, a path, a manifest, a lock file or `proven.yaml` from the change cannot inject arguments into a command or lead Proven to read or write outside the places it should | not started |
| 4 | **Evidence queries** | a value from the change enters a query only as a literal and cannot change the query's meaning | not started |
| 5 | **Models** | text in the change cannot make an agent drop its rules; the guardrail checks every prompt and answer; a model's answer without evidence never becomes a finding | not started |
| 6 | **Output and CI** | a pull-request comment carries no active content taken from the change; the workflow asks for the least token permission, never runs on `pull_request_target`, and pins every action it uses | not started |
| 7 | **The lab and the images** | the lab binds to localhost, generates its password, keeps the Prover's containers inside Docker in Docker; images are minimal, not root, and scanned | not started |
| 8 | **Data on disk** | the state and the trace are kept outside the reviewed repository, readable only by their user, and hold no secret | not started |

Pieces 0, 1 and 2 go first: they guard the machine and the keys. The others follow, two or three a day.

## How one piece is reviewed

1. **The order names the property, the code it covers and the proof expected.** It never asks a reviewer to attack:
   it asks whether the property holds, and where it does not, for a test that fails before a fix and passes after it.
2. **Two independent reviews**, each by a different model, neither seeing the other's verdict. A reviewer that cannot
   run a check says so; a review with an unrun check does not return "holds".
3. **Each weakness found becomes a fix** in the train that owns that code, with the failing test as its proof, and
   goes through the same two reviews as any other change.
4. **The piece is closed** when both reviews say the property holds, or every weakness they found is fixed and merged.

## Why one model is not enough here

Proven's code was written by one model (Grok, on Cursor). Each of its changes was reviewed by two others. For the
security pieces that is the minimum, not a luxury, for three reasons:

- **A model reviewing code tends to share the blind spots of the code it reads.** A second model from another family
  has been trained differently, reads in a different order, and stops at different lines. What one finds acceptable,
  the other may refuse.
- **In a security piece the cost of the two errors is not equal.** A false *holds* ships a weakness to every user; a
  false *reject* costs one more round of review. Two independent verdicts, where either can refuse, trade a little
  time for the error that matters.
- **Disagreement is information, not noise.** When one review accepts and the other rejects the same commit, the
  difference shows exactly where the property is in doubt, and the fix is aimed there.

The same rule holds for every change of Proven: two reviews from different model families, neither seeing the other
([ADR-0009](fleet/decisions/ADR-0009-two-reviewer-models-and-the-same-model-fallback.md)). For each closed piece, the
table below gives both verdicts side by side, on the same commit.

| Piece | Commit | claude-code | codex | Agreed? |
|---|---|---|---|---|
| filled in as each piece closes | | | | |

## What is published, and when

- This page's status column is updated as each piece moves: *not started*, *in review*, *fixes in progress*, *holds*.
- A weakness is written down publicly only **after** its fix is merged: in [SECURITY.md](SECURITY.md), with the piece,
  the property, the change that fixed it and the test that proves it.
- The review orders and the reviews themselves are published in `fleet/` with the fleet's other records, once the
  piece is closed.
