# ADR-0030: The product and the hackathon live apart: everything about how Proven was built moves to `hackathon/`

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Trigger:** the owner, 23 October: tidy the repository along Go's usual layout, and *"put fleet and everything else not
strictly about the project into a hackathon folder, so that after the hackathon we are left with a proper repository
without post-hackathon rubbish."*

## Context

Proven's code already follows Go's usual layout (`cmd/`, `internal/`, `docs/`, `assets/`). Its root does not: beside
the product it holds the record of the build week (`fleet/`, `planning/`, `SAPHAN.md`, `DESIGN.md`), test material in
two places (`tests/`, `conformance/`), and the hackathon's `metadata.json`. A reader cannot tell the product from the
process, and after the hackathon the process would be left behind in the product.

On 23 October 49 Markdown files linked into `fleet/` or `planning/`, 78 links pointed at `fleet/decisions/`, and
`metadata.json` named paths in both. No Go code depends on those paths.

## Decision

1. **Two halves.** Everything outside `hackathon/` is the product, and stays after the hackathon. `hackathon/` holds
   how it was built: `fleet/` (orders, status, reviews, messages, infrastructure, the cost ledger), `planning/` (the
   topic, the answers, the daily conversations), `SAPHAN.md`, and `DESIGN.md` (the plan of 20 October, kept as
   written).
2. **The target layout:**

   ```
   README.md  LICENSE  SECURITY.md  AGENTS.md  CONTRIBUTING.md  Makefile  go.mod  action.yml
   metadata.json  metadata.schema.json     (the hackathon requires them at the root until it ends)
   .github/     templates, CODEOWNERS, workflows
   cmd/  internal/
   docs/        user documentation; adr/ (the design decisions); security/ (SECURITY_REVIEW.md)
   test/        corpus/ (from tests/corpus), conformance/ (from conformance/)
   testdata/  demo/  assets/
   hackathon/   fleet/  planning/  SAPHAN.md  DESIGN.md
   ```
3. **The design decisions are the product's**, so they move to `docs/adr/` and stay.
4. **Two phases, both on 23 October** (the owner moved phase 2 forward the same day: *"today, before the push, when
   the agents stop, we migrate the whole repository"*). Phase 1: this decision, the layout map in the README,
   `CONTRIBUTING.md`. Phase 2, before the day-2 push, while the agents are paused: every move with `git mv`, every link
   and every path in `metadata.json` rewritten, and the masters' train orders updated to the new paths. The four days
   left before the last push are there to catch what the move broke.
5. **Links are checked for good, not once.** Phase 2 adds `make check-links`, run in CI: every relative link in every
   Markdown file and every path in `metadata.json` must resolve. A deliberately broken link must turn it red.
6. **After the hackathon:** tag the repository `hackathon-final`, point the links from `docs/adr/` into `hackathon/` at
   that tag, move `metadata.json` and its schema into `hackathon/`, then remove `hackathon/` from the default branch.
   The record stays readable at the tag.

## Consequences

- A judge, and later a user, sees the product first; the build's record is one folder away.
- Phase 2 touches most documents at once; the link check is what makes that safe.
- Links into the record from outside the repository (posts, the submission) should use a tag, not `main`.
