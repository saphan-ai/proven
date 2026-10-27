# STATUS — lab-polish · round 3

- state: ready-for-review
- executor: a Claude Code agent (an executor seat; not a Cursor agent, as the order named the backend)
- model: Sonnet 5.5 (`claude-sonnet-5-5`)
- last commit: see the branch tip
- updated: 2026-10-27 UTC

## Summary
Round 3 makes the lab read as finished. The docs site's generated page is regenerated and a Go test now holds it. Every dashboard number counts one thing, per pull request, and opens with one sentence (`5 of 7 pull requests blocked`). Bars are named and draw no zero bar, tables fit their rows, the default range is half an hour, the lab opens in the light theme, each report page has a folded "How this was answered" block, and the start page has six scenarios and a security section. All of it is built and checked offline. **Nothing was run on Docker and no screenshot was taken.**

## Red items (read these first)
1. **No Docker, no browser, no `node` on this seat. The 1920x1080 screenshots of `/lab`, a report page and the six dashboards are not taken, and the start page has no screenshots at all.** The security section links to the live pages (the build pipeline on GitHub Actions, the comments in the lab, the image on Docker Hub) and shows no picture of them. The master takes the screenshots.
2. **The Grafana changes are written against Grafana's documented behaviour and checked as JSON, not seen.** The likeliest to differ live: bar names from the series legend (the old display-name template is gone, which is the fix for `Value #A`); the `noValue` sentence of an empty bar gauge; `filterFieldsByName` on Tempo's span tables, which keeps only the columns the panel names (the raw `service.instance.id` column is kept only on the Engines table, renamed "Review"); and `label_replace` on the asked-and-answered model label.
3. **`make try SCENARIO=forged` and `make try SCENARIO=prover` were not run.** They need Docker. Tests hold their flags (the Prover's flags equal `DefaultContainerConfig`), that an unknown scenario is refused by name, and that the page and the script agree on the two names.
4. **Three tests fail on this seat for a reason outside the change.** `TestWorkflowAHeadThatIsAlreadyReviewedIsNotReviewedOrCommentedAgain` and `TestGoLiveTheCommentStartsWithTheMarkerThenHowItWasMade` call `mktemp` in the system temp directory, which this seat may not write. `TestSampleSARIFMatchesTheSchema` cannot open the user cache directory. All three: `permission denied`.
5. **The Docker Scout figures are not stated on the page.** The order's example reads 0 critical, 0 high, 0 medium, 0 low with the bill of materials and provenance attested. This seat could not read the live Scout page, so the page says the image is scanned and links to it, and states no number.
6. **The docs site's Node guard was not run** (`node` was refused). A Go test runs the same generator and compares.
7. Evidence and Meterless engines still show text panels for the metrics Proven does not emit: the telemetry package at this branch's tip has nine instruments, none for Zetaris, the Prover or the engines.
8. The findings table in a report page has the columns `--md` writes (severity, finding, location, rule). The owner's example adds evidence kind, level of proof and why. That is a change to the CLI's Markdown, not a page change, and was not made.

## What now holds (by the order's numbering)
1. **Docs site.** `docs-site/generated/settings.md` regenerated with `sh docs-site/scripts/generate.sh docs-site/generated`. `TestTheDocsSiteGeneratedPagesAreWhatTheGeneratorWrites` runs the generator and compares; its control changes one page by a byte and goes red.
2. **Numbers that agree.** A pull request now keeps one review instance id, `pr<number>-<fixture>` (no time in it), so a new push continues the same series and every number counts the latest review of each pull request. A stat that reads `N of M` has M equal to the number of demo pull requests and counts over those ids. The Findings column and the proof bars sum the same series. The "Reviews with a failed service" table (a bare "No data") is gone; failed agents and failed calls are stats.
3. **No `Value #A`, no zero bar.** Bars are named by their legend. An empty bar chart says why in one sentence and draws no bar. The reviewer-disagreement panel is a sentence when empty.
4. **Labels.** The model panel reads the model asked, and `asked → answered` only when the provider names one. Refusal reasons read in words (no evidence, refuted by the Skeptic, already refuted earlier). Bar labels are 16 px and values 20 px. Tempo tables keep only named columns.
5. **Units.** Prover and Investigator durations stay in nanoseconds, the unit Tempo returns, and the panel says that in sample mode they are recorded answers and not real runs (a few microseconds is the recorded answer). The model-latency bars say the same.
6. **Time range.** Half an hour on all six.
7. **Tables.** Each is tall enough for every row it can hold. The "(the latest N)" titles are gone. The Engines table names the pull request.
8. **Report pages.** The footer tells a reader to open a block only when the page has one. The page's review runs with the same settings as the workflow (the special reviewer is named). A folded "How this was answered" block lists the agents, the number of queries and their snapshot, the reviewer model, and what a receipt is.
9. **Grammar and logs.** The comment headline reads `4 errors, 0 warnings, 0 info` (`1 error` for one); the golden file moved with it. One empty line, not two, between step lines 4 and 5. The receipts script names its file from the repository root.
10. **Light theme.** Grafana's default theme is light; `/lab` and the report pages are light by default and follow a dark system.
Showcase: six scenarios on the start page and a "Security: code and image" section; each outside link is a public page of ours.

## Controls
Each names the field it mutates.
- `TestTheVisualBarTurnsRedWhenEachRuleIsBroken`: eighteen mutations, one field each: the Tempo Duration unit; a table's `noValue`; a bar gauge's `noValue`; a bar's expression (a zero row added); a bar's `legendFormat` set to `Value #A`; a bar's `options.text.titleSize`; a stat's base colour; a stat's unit and mappings; a stat's mappings (a plural with no sentence for 1); a stat's unit set to `of 5` ("8 of 5"); a stat's expression (no longer one review per pull request); a table's `gridPos.h`; a table's transformations (raw columns); a chart's interval; a title's length; a bar gauge's unit; the first panel's height; and the default range. Each goes red.
- `TestEveryDashboardHoldsTheVisualBar`, `TestEveryProvisionedDashboardIsBuiltToTheOneStyleAndQueriesWhatIsEmitted` (a misspelt metric and a misspelt label: red), `TestTheDashboardsAreTheGeneratorsOutput`.
- `TestUpPrintsOneGreenLineAfterEachPhaseWithAnEmptyLineOnEachSide`: the line that lets two step lines share an empty line removed: red.
- `TestReportFooterOnlyTellsTheReaderToOpenABlockWhenThereIsOne`, `TestReportPageSetsTheReviewLikeAGitHubComment` (the how block).
- `TestEveryScenarioButtonOpensAPullRequestTheLabSeeds`: a token for a change nobody seeds: red. `TestThePagesCarryNoPrivatePathAddressLoginOrMachineName` (no image is in the page folder; the check also reads an image's name). `TestEveryOutsideLinkOnThePagesIsAPublicPageOfOurs`.
- `TestTheDocsSiteGeneratedPagesAreWhatTheGeneratorWrites`: one page changed by a byte: red.

## Evidence
- `go vet ./...`: exit 0.
- `go test ./...`: every package passes except the three named in red item 4.
- `sh -n` on `try.sh`, `ui.sh`, `up.sh`, `reports.sh`, `nightshift-receipts.sh`, `report.sh`: exit 0.
- `sh scripts/check-links.sh`: exit 0. It had reported three broken links in the generated settings page, which resolve from `docs/` on the site; the script now resolves a generated page from there.

## Judgement calls
- A pull request keeps one instance id. Cost if wrong: a pull request reviewed twice with a different set of proof levels shows both sets on the proof bars until the series ages out; the table is unaffected.
- Half an hour, not fifteen minutes: a lab opened later than that reads zero until the range is widened.
- The plural fix is in the CLI's Markdown (`internal/cli/format_md.go`), not in the lab, because the lab only renders it.
- `info` stays singular in `0 info` (it is a level, not a count noun).

## Not run
Docker, `make up`, `make try`, screenshots, the docs site's Node guards, the three tests of red item 4.
