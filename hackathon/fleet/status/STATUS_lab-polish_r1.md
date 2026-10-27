# STATUS — lab-polish · round 1

- state: ready-for-review
- executor: a Claude Code agent (an executor seat; not a Cursor agent, as the order named the backend)
- model: Sonnet 5.5 (`claude-sonnet-5-5`)
- last commit: see the branch tip
- updated: 2026-10-26 UTC

## Summary
The judges' lab now says what it will do and when it is done, ends on one line that cannot be missed, gives every demo pull request a report page set like a GitHub comment, and has six dashboards in one style instead of one. All of it is built and checked offline: 92 lab tests pass, 0 fail, 2 skip (the two that need a pseudo-terminal, which this seat does not open). **Nothing was run on Docker, so no screen was seen**: how a page or a panel looks is not checked here.

## Red items (read these first)
1. **No Docker on this seat. No `make up` was run and no screenshot was taken.** The Grafana panels, the report pages in a browser, the green and red step lines on a terminal, and START HERE in bold green are written and tested against their text and structure only. The controls the order asked for on a Docker host (one `make up`, screenshots at 1920x1080 of `/lab`, a report page and every dashboard) are **not run**.
2. **The report pages need `--md`, which this branch's CLI does not have.** The Markdown format is on `main` (commit `ec9f584`), not on the base of this branch (`daf53bf`). Until `main` is absorbed, `make reports` runs Proven's image with `--md`, the run fails, and each Report link shows the page that says the report is not ready. Nothing else breaks. The order said `--md` exists on `main`; it is not on this branch.
3. **`docs/WALKTHROUGH.md` does not exist** on this branch or on the `main` this seat last read. The ready block's Help line links to it, as the owner laid out. The walkthrough's first screenshot, which the order asked to update, could not be: there is no walkthrough here.
4. **The Nightshift dashboard has no data until the owner adds the public receipts.** The raw receipts file is inside the Nightshift VM, not on this seat. The importer, the dashboard and the privacy tests are built against the public shape the docs name; the 15 public receipts, and the step that turns raw receipts into that shape, are owed. It was not run against a real receipt.
5. **Evidence and Meterless engines show text panels, not numbers, for the metrics Proven does not emit.** Read at `origin/main` (`11112c0`) as of this run: no metric for Zetaris queries, rows, duration or refusals; for Prover runs; for Scout plans or guard overrides; for Markovian steps; for World Model claims or evidence. Each is named in a panel. When the core master adds one, that panel becomes a query.
6. **One test outside the lab changed.** `internal/telemetry/dashboard_test.go` read one dashboard file; it now reads all six and holds the union to the instruments. It is the core master's package; the change is the smallest that keeps its rule (every instrument is on a dashboard, no query names a metric nothing emits).

## What was built
- **`make up` tells you what happens.** `lab/scripts/ui.sh`: the plan after the questions and before the first Docker call; one status line per phase with its time and an empty line on each side, green on success; red naming the phase and the log to read on failure; colour only on a terminal, `NO_COLOR` and `TERM=dumb` respected. `up.sh` runs five phases; phase 4 waits for every demo review (`demo-check.sh`: a line per pull request, `DEMO_WAIT_SECONDS` for the whole wait, a second bound on the trace count) and the lab is shown even if a review is late. `PROVEN_LAB_WAIT=no` skips the wait; `make e2e` sets it.
- **Enter takes the default.** Every question writes its default in capitals (`[SAMPLE/live]`, `[y/N]`, `[Y/n]`, `[OURS/own]`) and Enter takes it. `flow_ask` now takes the default as an argument.
- **The ready block.** Exactly the owner's layout, ending on `>>>>>>>>  START HERE:  http://127.0.0.1:3302/lab  <<<<<<<<` (bold green on a terminal) and one empty line, nothing after. The Grafana link is the Overview (`LGTM_DASHBOARD_UID=proven-overview`).
- **Report pages.** `lab/seed/md2html.awk`, `lab/seed/report.sh`, `lab/page/report.html.tmpl`: the review of `proven scan --md` as a page with a verdict badge, the findings table, each finding folded, light and dark by the reader's system; no script, no remote address, no stylesheet copied from elsewhere (this seat has no network to fetch one byte for byte). `lab/scripts/reports.sh` builds one per pull request with no network and no key; the seed writes a not-ready page first so no link is dead; nginx serves `/lab/report/<n>.html`; `/lab` has a Report column, the help line, the `make demo-check` line and the six dashboard links. In a live lab the page says it is the sample review of the same change and that the live review is the comment.
- **Six dashboards.** `lab/grafana/build_dashboards.py` writes them: Overview, One review, Evidence, Models and trust, Meterless engines, Nightshift. Dark, at most eight panels each, a word beside every colour, a sentence in every panel that can be empty, a description in every panel, the last hour (six for Nightshift). The unreadable agents-and-reviewers timeline is gone; One review lists each agent with a duration bar from the trace. Sample-mode latency is titled as recorded answers. The shared model-call errors are explained in the panel: in sample mode a call for which the recorded answers hold no answer counts as an error (`generator: no recorded response`). A review is named by pull request and change through its instance id (`<unix time>-pr<number>-<fixture>`), set by the seeded workflow.
- **Nightshift importer.** `lab/seed/nightshift-import.sh`: only app `proven`, `prompt_preview` dropped and a line that still names it refused, replayed to end at the moment of import, one push to the lab's own Loki; `make nightshift-receipts`.
- **Docs.** `docs/LAB.md` (the plan and step lines, the ready block, the six dashboards, the report pages, the receipts, the owner's test steps), `docs/CONFIGURATION.md`, `docs/LAB_GO_LIVE_TEST.md`.

## Controls
Each names the field it mutates.
- `TestFlowEnterTakesTheCapitalisedDefaultOfEveryQuestion`. Mutated field: the empty-answer branch of `flow_ask` (Enter refused again): red.
- `TestUpPrintsThePlanAfterTheQuestionsAndBeforeTheFirstDockerCall`. Mutated field: where `up.sh` calls `ui_plan` (after the first Docker call): red. The order is read from one output file the stand-in Docker also writes to.
- `TestUpPrintsOneGreenLineAfterEachPhaseWithAnEmptyLineOnEachSide`, `TestUpNamesTheFailedPhaseInRedAndTheLogToRead` (three phases). Mutated field: `ui_exit` removed from the EXIT trap: red.
- `TestUpEndsOnStartHereAndOneEmptyLine`. Mutated field: the newlines after the START HERE line: red. `make check-banner` holds the same, and on a terminal the bold green line and `NO_COLOR`.
- `TestUpStillShowsTheLabWhenAReviewDoesNotArrive`, `TestUpDoesNotWaitWhenToldNotTo`, `TestE2EDoesNotWaitForTheDemoReviewsItsOwnPullRequestStandsFor`.
- `TestUiPaintsGreenAndRedOnlyWhereColourIsAllowed` (painting, with the terminal test answered). `TestUiColourOnATerminalHonoursNoColorAndDumbTerminals`: mutated field the `NO_COLOR` line of `ui_color`; **skipped here (no pseudo-terminal), not run**.
- `TestReportPageSetsTheReviewLikeAGitHubComment`, `TestReportPageWritesNoTagTheReviewDidNotMean` (script, image, iframe and link in the review are written as text), `TestReportStubSaysTheReportIsNotReadyAndNamesTheCommand`.
- `TestReportPageCarriesNoPasswordTokenOrPrivateName`. Mutated field: the template given the password: red.
- `TestReportsBuildsOnePagePerPullRequestWithNoNetworkAndNoKey`. Mutated field: `--network none` in `reports.sh`: red. No credential in any Docker call.
- `TestEveryProvisionedDashboardIsBuiltToTheOneStyleAndQueriesWhatIsEmitted`: six files, uid equals file name, a title, a time range, 1 to 8 panels, none overlapping, a description on each, every metric from `internal/telemetry`, every label a series carries. Mutated field: a misspelt metric name, then a misspelt Loki label: red each time.
- `TestTheDashboardsAreTheGeneratorsOutput`, `TestTheOverviewHoldsTheDemoManifestToItsCounts`, `TestTheDashboardsThatLackMetricsSayWhichOnesAreMissing`, `TestTheLabNamesEachReviewByItsPullRequestAndChange`.
- `TestNightshiftImportLoadsOnlyProvenOwnCallsAndDropsThePrompt`. Mutated field: the drop and the refusal of `prompt_preview`: the prompt reaches the push, red. `TestNightshiftImportRefusesWhatItCannotTrust`.

## Evidence
- `go vet ./...`: exit 0. `gofmt -l lab internal cmd`: nothing. `sh -n` on every `.sh` under `lab`: exit 0.
- `go test -count=1 ./lab/...`: ok. 92 tests pass, 0 fail, 2 skip (`TestFlowLeavesARealTerminalWithEchoOnAfterCtrlC`, `TestUiColourOnATerminalHonoursNoColorAndDumbTerminals`: no pseudo-terminal on this seat).
- `go test -count=1 ./...`: one failure, `TestSampleSARIFMatchesTheSchema` in `internal/cli/sarifschema`: `open <user cache>/proven/logs/trace.jsonl: permission denied`. It writes to the user's cache directory, which this seat may not write. It does not touch the lab; not caused by this change, and not re-run with another log directory.
- `awk` and `cp` are not available to this seat as commands of its own; the awk renderer and the importer were exercised only through the project's own tests (`go test`).

## Judgement calls
- The report pages are built after the seed from Proven's sample review of each change, not at seed time, because the review does not exist at seed time. Cost if wrong: a live lab's report is a sample report, said on the page.
- The Grafana link of the ready block is the Overview, not `proven-review` as the example showed. Cost: one setting.
- The Overview names a review for a clean or a defect pull request by the fixture in its instance id, so "false alarms on the clean changes" is a number. The ids are set by the seeded workflow. Cost if wrong: Overview counts read zero, in green; the table still lists reviews.
- Instance ids replace the id Proven made for itself, through `OTEL_RESOURCE_ATTRIBUTES`, which Proven's setup reads after its own id. Not run.
- The Nightshift receipts are replayed to end at the moment of import (Loki refuses samples older than a week). Each line keeps its own time.

## Blocked on / next
Absorb `main` (for `--md`), then one `make up` on a Docker host with the screenshots the order names. The owner's test steps are in `docs/LAB.md`, "Test steps for the lab's look".
