# Security review: how Proven's code is checked as a whole

Every change to Proven is reviewed on its own: two independent reviews of its diff before it lands. A diff review sees
one change at a time. It does not see a property that no single change breaks but that the code as a whole does not
keep. So the code is also reviewed as a whole, in pieces, each piece one boundary where something untrusted meets
something Proven protects. Decided on 23 October 2026
([ADR-0028](../adr/ADR-0028-proven-is-security-reviewed-as-a-whole-in-pieces.md)). What Proven protects and how:
[SECURITY.md](../../SECURITY.md).

## The pieces

Each piece names the property that must hold. A review of a piece reads every place in the code where that boundary is
crossed, not only the files where it is expected.

| # | Piece | The property that must hold | Status |
|---|---|---|---|
| 0 | **Tools first** | `govulncheck`, `gosec` and `staticcheck` run over the whole repository; their output is the baseline the other pieces start from | fixes merged (`f9628fd`); `govulncheck` runs inside Proven (`891a7e9`) |
| 1 | **The Prover's container** | code from the change runs with no network, no secret and bounded resources; nothing in the change can shape the container's arguments or reach beyond its read-only checkout | fixes merged (`1d23854`, `175fc39`) |
| 2 | **Secrets** | no key reaches the trace, the logs, the reports, a pull-request comment or a container; a missing key is named, never printed | fixes merged (`37eabfb`, `2f0f44a`, `310bce5`) |
| 3 | **The reviewed repository as input** | a git range, a path, a manifest, a lock file or `proven.yaml` from the change cannot inject arguments into a command or lead Proven to read or write outside the places it should | fixes merged (`b3e2648`) |
| 4 | **Evidence queries** | a value from the change enters a query only as a literal and cannot change the query's meaning | fixes merged (`460da35`, `222dcca`) |
| 5 | **Models** | text in the change cannot make an agent drop its rules; the guardrail checks every prompt and answer; a model's answer without evidence never becomes a finding | fixes merged (`53d88ba`) |
| 6 | **Output and CI** | a pull-request comment carries no active content taken from the change; the workflow asks for the least token permission, never runs on `pull_request_target`, and pins every action it uses | fixes merged (`f723510`) |
| 7 | **The lab and the images** | the lab binds to localhost, generates its password, keeps the Prover's containers inside Docker in Docker; images are minimal, not root, and scanned by Docker Scout | fixes merged (`f2f3b03`) |
| 8 | **Data on disk** | the state and the trace are kept outside the reviewed repository, readable only by their user, and hold no secret | fixes merged (`492f03b`) |

Pieces 0, 1 and 2 go first: they guard the machine and the keys. The others follow, two or three a day. Each piece reviews `main` at one named commit, and a piece whose code is not on `main` yet waits for it
([ADR-0029](../adr/ADR-0029-a-piece-is-reviewed-at-one-commit-when-its-code-is-on-main.md)). Pieces 0, 1 and 2 are reviewed at
`90d6951`, pieces 3 and 4 at `436ef04`, pieces 5 and 8 at `d82cac2`.

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
([ADR-0009](../adr/ADR-0009-two-reviewer-models-and-the-same-model-fallback.md)). The table gives both verdicts side
by side, on the same commit, as they came back on 23 October. What each reviewer found is published only once its
fix is merged ([SECURITY.md](../../SECURITY.md)).

| Piece | Commit | Anthropic Claude (claude-code) | OpenAI (codex) | Agreed? |
|---|---|---|---|---|
| 0, tools first | `90d6951` | reject | reject | yes |
| 1, the Prover's container | `90d6951` | accept with conditions | **reject**, with failing tests | **no** |
| 2, secrets | `90d6951` | accept with conditions | **reject**, with five failing tests | **no** |
| 3, the reviewed repository as input | `436ef04` | reject | reject | yes |
| 4, evidence queries | `436ef04` | reject | reject | yes |

On pieces 1 and 2, one reviewer would have let the code through. On piece 1 both reviewers named the same weakness,
and Claude still accepted with conditions; OpenAI rejected. That weakness is the first one fixed. On piece 2 Claude
found nothing that broke the property; OpenAI rejected the same commit and proved it with five tests that fail.
"Reject" means the property did not hold at that commit; each piece's fixes are made in the train that owns the code,
and each fix gets its own two reviews.

**A third family, on 24 October.** On 23 October the pieces were reviewed by Anthropic's Claude (on claude-code) and
OpenAI's model (on codex). On 24 October the review runs again with NVIDIA's Nemotron as a reviewer, through a governed
agent loop, on the commit with the first fixes in it: the pieces already reviewed, to see what a third family finds that
the first two did not, and the pieces still ahead. Its verdicts are added beside the others.

## Dependency alerts on this repository

On 26 October 2026 GitHub's Dependabot reported **51 open alerts** on `main` (4 critical, 20 high, 24 moderate, 3 low).
We read every one:

- **45 were npm packages of the documentation site** (`docs-site/`): Next.js above all, plus sharp, postcss and zod. The
  site's dependencies were upgraded to patched versions (Next.js 15.5.28; commit `f1567e9`) and the site rebuilt. The
  open alerts fell from 51 to 6.
- **The other 6 are on purpose.** They all point at `golang.org/x/net v0.16.0` in `demo/live/change/go.mod`. That module
  is the "after" side of the [live demo](../../demo/live/README.md): a small change written to carry a known-exploited
  dependency, so a judge can watch Proven find it and prove it with a query. The module does not import the package, and
  it is never part of Proven itself. These alerts stay open because the demo needs them.

We also turned the 51 alerts into a **reference set**: Proven reviews the repository at the commit where Dependabot
counted them (`fbdf7d3`), and for every alert we record whether Proven found it with its own evidence, missed it, or
found it differently, and what Proven finds that Dependabot does not. The comparison is published whatever it shows.

## What is published, and when

- This page's status column is updated as each piece moves: *not started*, *in review*, *fixes in progress*, *fixes merged*
  (every weakness the two reviews found is fixed on `main`; the piece's second review is still to say so), *holds*.
- A weakness is written down publicly only **after** its fix is merged: in [SECURITY.md](../../SECURITY.md), with the piece,
  the property, the change that fixed it and the test that proves it.
- The review orders and the reviews themselves are published in `hackathon/fleet/` with the fleet's other records, once the
  piece is closed.
