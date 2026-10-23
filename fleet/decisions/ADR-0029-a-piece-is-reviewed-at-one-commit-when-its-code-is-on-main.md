# ADR-0029: A security piece is reviewed at one commit of `main`, and only once its code is on `main`

**Status:** proposed · **Date:** 2026-10-23
**Decided by:** the owner, when he signs the train order `proven-security-review` · **Written by:** the master of
proven-security-review, 2026-10-23
**Follows:** [ADR-0028](ADR-0028-proven-is-security-reviewed-as-a-whole-in-pieces.md) (the review in pieces),
[ADR-0026](ADR-0026-one-makefile-builds-proven-and-runs-the-judges-lab.md) (the Makefile, the images, Docker Scout)
**Trigger:** writing the train order for ADR-0028. On 23 October `main` held no Makefile, no lab, no image of Proven,
no action and no workflow.

## Context

ADR-0028 reviews the code as a whole, in nine pieces. The code changes every day of the build week. Two of the pieces
cover code that is decided but not built: the output and CI piece needs the action and the workflow, and the lab piece
needs the Makefile, the lab and the images. Piece 0 asks for Docker Scout, which scans images, and no image is built
from `main` yet.

## Decision

1. **Each piece reviews `main` at one named commit.** Its order names the commit. Both reviews of a piece read the
   same commit, so their verdicts can be compared.
2. **Code that lands after the cut is covered by the diff reviews of its own change**, as before. A piece is not
   reopened for a later change unless that change crosses the piece's boundary in a new place; the master then orders
   one more review of that change against the piece's property.
3. **A piece whose code is not on `main` waits for it.** The output and CI piece reviews the output formats first and
   the action and workflow when they land. The lab piece starts when the Makefile, the lab and the image are on `main`.
   Until then its status stays *not started*, and SECURITY_REVIEW.md says why.
4. **Docker Scout runs in the lab piece (7), not in piece 0.** Piece 0 records `govulncheck`, `gosec` and
   `staticcheck`. Scout scans images, and the lab piece is where the images are reviewed.

## Consequences

- A piece's verdict is about one commit, and the reader can check out that commit and see what was reviewed.
- Pieces 6 and 7 may close last, after the build week's code lands; the public status says so.
- A change after the cut is trusted to its diff reviews unless it opens a new crossing of a boundary.
