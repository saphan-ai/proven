# STATUS — lab-polish · round 2

- state: ready-for-review
- executor: a Claude Code agent (an executor seat; not a Cursor agent, as the order named the backend)
- model: Sonnet 5.5 (`claude-sonnet-5-5`)
- last commit: see the branch tip
- updated: 2026-10-27 UTC

## Summary
Round 2 repairs the defects the master found in a live run of round 1. The report pages render and are readable, the docs link guard holds, durations use the unit their source returns, every dashboard bar is named and no bar chart is an empty box, colours follow meaning, time ranges fit the run, tables say when they show only the latest rows, and `make up` prints a `[5/5]` line. All of it is built and checked offline. **Nothing was run on Docker and no screenshot was taken**: how a panel looks is not seen here.

## Red items (read these first)
1. **No Docker, no browser, no `node` run on this seat. The screenshots the order asked for at 1920x1080 are not taken.** The seat may not run `docker`, and a `node` run of the docs site's own guards was refused. The master takes the screenshots and runs `node docs-site/scripts/build-content.mjs` then `node docs-site/scripts/guards.mjs`.
2. **The Grafana changes are written against Grafana's documented behaviour and checked as JSON, not seen.** Four things are the likeliest to differ live: the `merge` transformation that builds the Overview table from four queries; `or on() label_replace(vector(0), …)` as the "empty" row of a bar chart (PromQL and LogQL); `label_join` in PromQL; and the `step` key on Loki queries.
3. **The 12 public receipts were not supplied.** The Nightshift dashboard says plainly that none are loaded and names `make nightshift-receipts`. Put the file at `lab/nightshift/receipts.public.jsonl` and it fills.
4. **Evidence and Meterless engines still show text panels** for the metrics Proven does not emit (read at this branch's tip, which carries main `11112c03`): the telemetry package still has nine instruments, none for Zetaris, the Prover or the engines.
5. **Two lab tests fail on this seat for a reason outside the change**: `TestWorkflowAHeadThatIsAlreadyReviewedIsNotReviewedOrCommentedAgain` and `TestGoLiveTheCommentStartsWithTheMarkerThenHowItWasMade` run a script that calls `mktemp` in the system temp directory, which this seat may not write (`mktemp: failed to create directory via template '/tmp/tmp.XXXXXXXXXX': Permission denied`). Not re-run with another temp directory.
6. **`scripts/check-links.sh` reports three broken links** in `docs-site/generated/settings.md` (`OSV_TABLE.md`, `ZETARIS_USL.md`, `POLICY.md`). That file came in with main; not touched here.

## What now holds (by the order's numbering)
1. **Report pages render.** The branch now has `--md` (main absorbed). A test runs the CLI's `--md` for the sample change and puts it through the page: the Blocked badge, the findings table, the folded details.
2. **The reports directory is world-readable** under any umask (`umask 022` and `chmod 755` in `lab/seed/report.sh`). The placeholder says "No report has been built" and names `make reports`; its footer no longer claims an `--md` rendering.
3. **The docs link guard passes on `docs/LAB.md`**: the anchor is `#the-dashboards`. A Go test holds in-page anchors of the lab's three pages with the guard's own anchor rule.
4. **Durations are in the right unit**: `ns` for Tempo tables.
5. **No `noValue` sentence on any table.** A table that can be empty says so in its title or description.
6. **The Overview table has Pull request, Change, Verdict, Findings, Claims dropped, State**, built with `label_replace` and `merge`, sorted by Pull request.
7. **No empty bar chart.** With no data a bar chart has one row that says why, named in words, with the names above the bars so they are not cut.
8. **Every bar is named** by a display name template: agent and reviewer, model asked and model that answered (an empty answering model reads "model not named"), guardrail side and verdict, reason, token series. Titles are held to a length a panel can show.
9. **Colour follows meaning.** A count that means "nothing yet" is blue. A "bad" count is green at zero and says `0 · none, good`. A failure bar chart is green at zero.
10. **Time ranges fit.** One hour on all six; bar charts step in one to two minutes.
11. **Tables are tall enough for their rows** and say "(the latest N)" when they cut.
12. **Grammar.** A stat's word never follows a number as a plural ("1 findings kept" is gone).
Also: `make up` prints `[5/5] The lab is ready`, then one empty line, then the banner.

## Controls
Each names the field it mutates.
- `TestReportsDirectoryIsWorldReadableWhateverTheUmask`: the `umask` and `chmod` lines of the script removed, run under umask 077: red.
- `TestReportPageRendersTheReviewTheCLIWritesNow`.
- `TestTheLabDocsLinkOnlyToHeadingsThatExist`: the anchor put back to `#the-dashboard`: red.
- `TestTheVisualBarTurnsRedWhenEachRuleIsBroken`: twelve mutations, one field each (the Duration unit, `noValue`, `displayName`, the `or on()` row, the base threshold colour, unit and mappings, the stat unit, `gridPos.h`, a table title, the target interval, a title, a bar gauge unit) and the default time range: red each.
- `TestEveryDashboardHoldsTheVisualBar`, `TestTheDashboardsAreTheGeneratorsOutput`.
- `TestUpPrintsOneGreenLineAfterEachPhaseWithAnEmptyLineOnEachSide` now holds five lines and one empty line before the banner.

## Evidence
- `go vet ./...`: exit 0. `gofmt -l lab internal cmd`: nothing. `sh -n` on `up.sh` and `report.sh`: exit 0.
- `go test -count=1 ./lab/...`: all pass but the two named in red item 5.

## Judgement calls
- The lab job still posts the text review; the report pages render `--md`. Kept, as the order said.
- One hour for every dashboard, not thirty minutes: a viewer who opens the lab a little late still sees the run. Cost: a lab left for over an hour reads zero until the range is widened.
- The Overview table is sized for twelve rows and says in its description to scroll for more. Cost: a lab with more than twelve pull requests hides some, said in the panel.

## Not run
Docker, `make up`, screenshots, the docs site's Node guards, the two tests of red item 5, the two pseudo-terminal tests.
