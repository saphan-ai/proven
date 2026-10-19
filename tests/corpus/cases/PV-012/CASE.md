# PV-012 — A filter that production never exercises, tested through a configuration production cannot build

- incident: I09 (one incident; this case is a later round of it, and the round after it settled the finding)
- intended fixture language: Go (a hint only; this contract is language-neutral; see `../../manifest.json`)
- contract written: 2026-10-19 · case code: not built yet — built before the window, declared prior art · checks: not run · independent check: not done
- the expected outcome is NOT in this file: it is in `expected/EXPECTED.md`, kept apart so an evaluation can withhold it

## 1. Story

A shared web application serves several teams from one store. It has an activity page that lists recent events and
offers filter chips. The page must show nothing that belongs to another team. Events live in several tables. Some tables
carry a team column of their own. Others do not: their team is looked up through a registry of work items, and a sync
step writes each discovered item into that registry.

The change under review narrows the default activity page to the requesting team: when the request carries a team, the
query keeps only that team's events, in every table. It also prints a sentence on the page saying what the page is
scoped to. Its control builds the page for a named team, seeds five events of another team, and checks that none of the
five is served. A second, vacuity arm builds the page for the other team and checks that all five are served.

Two reviewers from different model families reviewed the same commit, without seeing each other. One accepted it:
removing the narrowing turns the control red, the vacuity arm holds, and every mutation it tried is caught. The other
rejected it: it read how the running application builds the page and found that the narrowing is never entered there.

## 2. Contract (written before any code)

The case must provide, with no network and no model:
- a small in-memory store with two teams, "ours" and "theirs", and five events of "theirs": four in tables without a
  team column of their own (their team comes from the registry), one in a table with its own team column.
- a sync operation: given a requesting team, it discovers every work item and writes each one into the registry, stamped
  with the requesting team.
- two versions of the page, **base** and **change**, each offering three operations:
  1. build the page the way the running application builds it, from its settings;
  2. build the page for an explicitly named team, the way the change's control builds it;
  3. given a built page and a request with no filter, list the events it serves and the scope sentence it prints.
- **base** serves every event of the store.
- **change** keeps only the requesting team's events when the request carries a team, and serves every event when it
  does not. It narrows a table with its own team column by that column, and the other tables through the registry. Its settings have no place for a team, so operation 1 always builds a page with no team.
- the change's own control and vacuity arm, as described in §1.
- a **reference**: the seeded ownership of every event, written in the case data and never derived from the page; used
  only by the check.
- a **check** that prints one line per input: input name, base result, change result, reference result, where a result
  is the number of events of "theirs" that are served (and, where it applies, the scope sentence).

Inputs the check must include:
- the change's control: page built for "ours", sync run once by each team in creation order;
- the vacuity arm: page built for "theirs";
- the running application's page: built from settings, a bare request;
- the same bare request, comparing the scope sentence with and without a team on the request;
- order of syncs: page built for "ours", and one extra sync by "ours" after the items of "theirs" exist;
- the change's control with the narrowing through the registry removed, to show the control notices its absence.

Scope and assumptions: a single process; no real listener is required. The toolchain is chosen when the case code is
written, pinned, and recorded in §8.

## 3. Provenance

- source: a pair of concurrent reviews in the fleet's private record, September 2026 (pair P24 in `PROVENANCE.md`).
- reviewer families, as recorded: `claude-opus-5` (through Claude Code) and `glm-5.2` (through Qwen Code).
- the original code is private and is not published. Resolution status: settled by the next round, which reproduced the
  finding through a started application and repaired it.

## 4. Synthetic transformation

Replaced: the original application, its page, its tables, its notion of a team and every name. Kept, as the causal
properties of the disagreement: a filter that runs only when the request carries a value; a production construction
path that cannot supply that value; a control that supplies it by hand, so a mutation of the filter turns it red while
production never filters; and a sync step that re-stamps ownership with whoever ran it last.

## 5. Review target

The two versions of the page, the store, the sync operation, the change's control, and a readable diff between them. The
reviewer under evaluation receives only those and the description of the change in §1 — never the check, the reference
or `expected/`.

## 6. The two claims and their evidence (as preserved in the record)

- **Reviewer A (reject):** the running application's default page still serves all five foreign events, because it is
  built with no team and its settings cannot carry one, so the narrowing is never entered; the change's control builds
  the page in a way the application cannot. The scope sentence is the same with and without a team. Separately, for a
  named team, the result depends on the order of syncs: one extra ordinary sync re-stamps the foreign items and four of
  the five come back. Evidence in the record: a reading of how the application builds the page, a probe of a started
  application at its bare address, the sentence compared with and without a team, and one added sync in the control
  showing four of five. Where it went beyond its evidence: it said the whole sentence was byte-identical; only part of it
  was.
- **Reviewer B (accept):** the default page is scoped. Removing the narrowing through the registry leaks four of five, the vacuity arm serves
  all five, deleting registry entries turns the independent arms red, and the rest is a nice-to-have. Evidence in the
  record: named mutations, each with the control it turned red, and its own count of unscoped queries. Its measurements
  were accurate, inside the control's own construction, which supplies a team. It did not check whether the running
  application supplies one, or whether sync order changes ownership.
- Newly reconstructed evidence: none yet; it is built with the case code.

## 7. Expected outcome

In `expected/EXPECTED.md`.

## 8. Reproduction plan

Run the check on base and on change and record each line of its output. **Not run**: the case code is not built
yet. The exact commands, the toolchain and its version are written here when they exist.

## 9. Limits

- It shows one property of one control: a mutation that turns a test red proves the code matters when it is reached, not
  that production reaches it.
- The two reviewers counted unscoped queries in different units; the record never reconciled the counts, and they are
  not reconstructed here.
- A later fix extended ownership to two more tables. It is not part of this case.
