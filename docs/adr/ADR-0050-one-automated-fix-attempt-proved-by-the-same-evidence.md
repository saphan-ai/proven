# ADR-0050: One automated fix attempt, proved by the same evidence

**Status:** accepted · **Date:** 2026-10-25
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-25
**Follows:** [ADR-0042](ADR-0042-a-proposed-fix-is-proved-by-the-same-test.md) (a proposed fix is proved by the same
test), [ADR-0044](ADR-0044-every-finding-names-its-level-of-proof.md) (levels of proof),
[ADR-0025](ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md) (the change's code runs with no network
and no secret)
**Trigger:** the owner, 25 October: *"A feedback loop from the pull request's result to a coding agent: once it knows
exactly what is wrong, it will manage. A pull request plus an automated fix. If it does not manage, too bad. After it
fixes, only the deterministic evidence is run again to check. If it is still wrong after that one try, the pull request
is rejected. We have everything already."*

## Context

Proven already proposes a fix for a problem proved by a test, and proves the fix with the same test (ADR-0042). That fix
is a suggestion for one finding at a time, and only for findings proved by a test. A review also proves findings by data
(a query and its rows) and by tools. And the person who opened the pull request still has to read every finding and
decide what to do.

The evidence Proven attaches to each finding is exactly what a coding agent needs: what is wrong, where, and the check
that shows it. The checks are deterministic, so whether a fix worked can be decided without asking any model.

## Decision

1. **After a review with findings, Proven makes ONE fix attempt for the whole change.** A coding agent gets every
   finding with its evidence: the failing test and its output, the query and its rows, the tool's report. The agent is
   the configured generator; its default is NVIDIA's Nemotron (ADR-0019, ADR-0031). It writes one patch.
2. **The patch is judged by the same evidence, and by no model.** Proven runs again only the deterministic checks:
   - a finding proved by a test: the test passes on the change with the patch, and every other test of the package still
     passes;
   - a finding proved by data: the same query on the patched change no longer returns the row (for example, the
     dependency is now outside the affected range);
   - a finding proved by a tool: the same tool run no longer reports it.

   A finding that is only an opinion is not part of the attempt.
3. **One attempt, then a verdict.** If every proved finding is closed and nothing else broke, the verdict is
   **fixed automatically, waiting for a person**. If anything is still open or something broke, the verdict is
   **rejected**, with each open finding and its evidence. There is no second attempt.
4. **Proven never commits to the author's branch.** The patch is published as suggested changes in the review comment
   (GitHub and Gitea), or as a separate branch the author can merge in one step. A person always decides.
5. **The patch is untrusted input.** It runs only where the change's code runs: in the container with no network and no
   secret (ADR-0025). The command run there is built by Proven, never taken from the agent.
6. **Configured, never hardcoded** (ADR-0013): on or off, which generator writes the patch, the time limit of the
   attempt. Default: on in the judges' lab, off elsewhere until measured.
7. **Measured.** The corpus score reports, per set, how many reviews ended "fixed automatically" and how many
   "rejected", and whether any "fixed" patch was wrong against the answer key.

## Consequences

- A pull request with proved problems can come back already fixed, with the proof that the fix works, and the person
  only decides.
- Every fix is checked by evidence, not by a model saying it looks right.
- "Rejected" means something exact: the problems named here are still there after one honest try.
- ADR-0042's single-finding suggestion becomes one case of this loop.
