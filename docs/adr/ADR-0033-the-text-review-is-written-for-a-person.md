# ADR-0033: The text review is written for a person; the machine detail moves to `--verbose`, JSON and SARIF

**Status:** accepted · **Date:** 2026-10-23
**Decided by:** the owner · **Written by:** the advisor (claude-code, `claude-opus-5-5`), 2026-10-23
**Follows:** [ADR-0002](ADR-0002-evidence-is-a-test-or-a-query-through-zetaris.md) (evidence), [ADR-0020](ADR-0020-proven-command-line-and-action-interface.md) (the command line and its formats)
**Trigger:** the owner, 23 October, reading `proven scan --sample` on `main` (`ae34d44`): *"maybe give this stream of machine babble some notes for a human? Nobody but a machine will read it."* And: *"and some separators between the findings."*

## Context

On 23 October the text output printed, for each finding, its rule id, every evidence query's full SQL, its column list
and every row, and repeated the same EPSS line under each finding. The sample review of `discount` printed three
separate findings about one dependency, each with two or three queries. A person had to read SQL to learn what was
wrong and how Proven knew.

## Decision

1. **The default text output is for a person.** JSON and SARIF stay complete and unchanged; `--verbose` adds the full
   machine detail to the text output.
2. **A summary first**: what was reviewed, how many problems were found, how many possible problems were dropped for
   lack of evidence.
3. **Each finding is a block, separated from the next by a header line** that carries its number, its severity, its
   title and its place in the code: `━━ 1 of 3 ━━ ERROR ━━ Discount is applied twice ━━ billing/discount.go:5`.
   Colour only when the output is a terminal.
4. **Inside a block, three plain parts**: what is wrong, in one or two sentences; *How we know*, the evidence in plain
   words (a test that passes on the old code and fails on the new; the known-exploited list; the OSV range and the
   fixed version; the EPSS percentile), with the one command or query a reader can rerun; *Fix*, when it is known.
5. **One problem, one block.** Findings about the same dependency and the same advisory are shown as one block with
   all their evidence; the separate findings stay separate in JSON and SARIF.
6. **The dropped findings are shown**, one line each, with the reason, because a dropped claim is the point of Proven.
7. **The last line** names the reviewing models, whether the guardrail checked every call, and the trace's path.
8. **Two formats for people, two for machines** (the owner, the same day). For people: the text output above;
   **Markdown** (`--md`, `--md-output`) in the same shape as the pull-request comment, to paste into an issue, a wiki or
   a chat; and **HTML** (`--html`, `--html-output`), one self-contained report file with no external resources: the
   summary, one card per problem with *How we know*, the evidence (SQL, rows, test output) folded open on demand, and
   the dropped claims. For machines: **JSON** and **SARIF**, complete and unchanged.

## Consequences

- A developer reads a review the way they read a colleague's comment; the SQL is one flag away.
- Tests of the text output change with it; the JSON and SARIF tests do not.
- The pull-request comment, which is rendered from the same findings, should follow the same shape.
