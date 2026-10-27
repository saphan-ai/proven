# STATUS — lab-demo · round 4

- state: ready-for-review
- executor: a Claude Code agent (an executor seat; not a Cursor agent, as the order named the backend)
- model: Sonnet 5.5 (`claude-sonnet-5-5`)
- last commit: see the branch tip
- updated: 2026-10-26 UTC

## Summary
`make up` now asks `Do you want to tweak Meterless features? [y/N]` after sample or live, and on `y` asks one y/n for each feature that has a switch in Proven: Scout Intent and Markovian. Each answer is Proven's own setting, can be given in advance, and the lab sets no default of its own. H-MEM and the World Model have no switch in Proven, so they are not asked. The start page shows what the seed stored. `make livecheck` now runs the workflow's own derivation. A head commit gets one review and one comment. Nothing was run against Docker.

## Red items (read these first)
1. **H-MEM and the World Model have no switch in Proven at this commit.** Searched the settings of `internal/hmem`, `internal/world`, `internal/pipeline`, `internal/review` and the docs: none. The order said to offer no prompt for such a feature, so there is none. The page says `on (always)` for both. The owner's reading, "each one can be switched off by its setting", does not hold for these two until Proven adds one.
2. **The duplicate review run was not measured.** The seat has no Docker (the run was refused), so which events Gitea 1.22.6 fires for one push was not observed. The workflow's guards are built so that the result does not depend on the cause (see item 6 below), but the cause is unknown, and whether Gitea 1.22.6 honours the `concurrency` key is not verified.
3. **Scout Intent has no environment setting.** It is a key of `proven.yaml`. The lab writes `scout_intent: <value>` to a file in the review container and points the review at it with `PROVEN_CONFIG`. `PROVEN_LAB_SCOUT_INTENT` is the lab's name for it.
4. **Proven's own defaults today are off for Markovian and Scout Intent.** The owner's order says all four are on by default, landing as another track's change. The lab passes nothing when nothing is set, so the lab follows Proven's default whichever it is. The prompt sentences say "on by default", as the order told.
5. **The Zetaris shape is a claim.** `flow_probe` takes Zetaris as present when its API root answers 200, 400, 401 or 403 with a JSON object. The shape was read from Proven's adapters, not measured against a running Zetaris or Nightshift.

## Acceptance, item by item
1. **The prompts, as the owner worded them.** `TestFeaturesTheQuestionsAreAskedOnATerminalAndEachAnswerIsTheSetting`: `Do you want to tweak Meterless features? [y/N]`, the sentence about H-MEM and the World Model, the Scout Intent question and the Markovian question, each saying "on by default". `y` is `on`, `n` is `off`. Negative control (mutated field: the mapping of `n` in `flow_feature_answer`): the lab got `on`, red.
2. **Settable in advance, no prompt in CI.** `TestFeaturesNothingIsAskedOrPassedWithNoTerminalAndNoSetting` (nothing asked, nothing passed); `TestFeaturesEachSettingReachesTheLabInSampleModeToo` (both `off` reach the lab; negative control, mutated field: the exemption of the features in `flow_drop_keys`: red); `TestFeaturesASettingGivenInAdvanceIsNotAskedAgain`; `TestFeaturesEnterAndNKeepProvensOwnDefault`; `TestFeaturesAWrongValueIsRefusedByName`; `TestFeaturesAnInputThatEndsAtAFeatureQuestionIsAStopByName`.
3. **The page shows what was passed.** `TestFeaturesThePageSaysWhichAreOnFromWhatWasStored`: the page line is built from the values the seed stored (`H-MEM on (always), World Model on (always), Scout Intent off, Markovian on`), and `Proven's default` for an unset one. Negative control (mutated field: the Scout Intent setting read by `golive.sh`): the page no longer says `off`, red. `TestFeaturesTheSeedNamesWhatItStoredAndNeverAValue`, `TestFeaturesAFeatureAloneDoesNotMakeTheLabLive`.
4. **The review is given the setting.** `TestFeaturesTheReviewIsGivenEachSettingAsProvensOwn` runs the workflow's review step with a stand-in container that runs the step's script with only the variables named by `-e`: the review process saw `scout_intent: off` in the file `PROVEN_CONFIG` names, and `PROVEN_MARKOVIAN=off`. Two negative controls (mutated fields: the value of `SCOUT_INTENT` handed to the container; `PROVEN_MARKOVIAN` in the passed names): both red. A value that is not `on` or `off` stops the step before the container.
5. **`make livecheck` and the workflow agree for Path A.** `TestLivecheckJudgesPathAWithTheWorkflowsOwnDerivation`: with only `PROVEN_GENERATOR_KEY` in a test `lab/.env`, livecheck says `the workflow will run: live: NVIDIA direct, Zetaris off`, generator ready, guardrail ready, the query line not needed, exit 0. The stand-in checker answers by the settings it is given, so the old behaviour (`METERLESS_RUNTIME_TOKEN is empty`) would show. Negative control (mutated field: the line that runs the workflow's derivation, replaced by livecheck's own defaults): red. `TestLivecheckOtherSetupsAgreeWithTheWorkflow` (no key; Nightshift token alone; Zetaris on) and `TestLivecheckRefusesWhenTheWorkflowHasNoDerivationBlock`.
6. **One push, one review and one comment.** `TestWorkflowOnlyThePullRequestEventStartsItAndRunsOfOneHeadQueue` holds the trigger configuration: only `pull_request`, with its types named, and a `concurrency` group of pull request number and head sha. `TestWorkflowAHeadThatIsAlreadyReviewedIsNotReviewedOrCommentedAgain`: with a comment that names the head sha, neither the review nor a post happens; with none, or with a comment of another head, both do. Negative control (mutated field: the sha the guard looks for): the duplicate runs, red. `make e2e` now accepts the review of any succeeded run of the head (`success_runs`, checked in `e2e-parse-check.sh`), because the second run posts nothing.
7. **Finding 3, the end of input at the gateway address.** `TestFlowAnInputThatEndsAtTheGatewayAddressSaysSo`: `up: the input ended before the gateway address came.` Negative control (mutated field: the end-of-input test): the old message appears, red.
8. **Finding 4, the probe.** `TestProbeCountsAServiceOnlyWhenItAnswersInItsOwnShape`: a stray server answering 404, one answering 200 HTML, and one answering a plain 401 are taken for neither service; a JSON list of models and a JSON 401 are Nightshift; a JSON root is Zetaris. Negative control (mutated field: the shape test; any status counts): a stray 404 taken for Nightshift, red.
9. **macOS controls.** The two negative controls of the terminal tests are now subtests that skip on macOS by name (`skipControlOnDarwin`); the positive half of each runs there. The macOS result is the master's measurement; it was not repeated here.
10. **Docs.** `docs/LAB.md` (the questions, the mapping table, the settings, livecheck, one review per head, the probe shapes) and `docs/LAB_GO_LIVE_TEST.md` (Steps 1c and 1d with the expected output).

## Evidence
- `go vet ./...`: exit 0. `gofmt -l lab cmd internal`: nothing. `go test -count=1 ./lab/... ./cmd/livecheck/`: ok.
- `sh -n` on every changed `.sh`: exit 0. `sh lab/scripts/e2e-parse-check.sh`, `golive-check.sh`, `check-lab-gates.sh`, `check-review-gates.sh`, `check-lab-password.sh`: exit 0 (the terminal case of the last one is not run: no pseudo-terminal).
- `scripts/check-links.sh`: three broken links, all in `docs-site/generated/settings.md`, as in round 3; not touched.

## Not run
- Docker, Compose, Gitea, the runner: none. So not run: one push to a pull request branch (the duplicate-run control), `make up` with the prompts on a real terminal, the start page in a browser, a live review with each feature set.
- The macOS behaviour of the two skipped controls.
- The docs site's own text guard (`node docs-site/scripts/guards.mjs`): the command needed approval and was not run. The added text was searched by hand for the shapes the guard names: none found.
- `make` targets: `make` is not on the seat's command list; the scripts and tests behind them were run directly.

## Judgement calls
1. Enter at a feature question, and `N` at the first, leave the setting unset rather than choosing `on`. The order says the lab never defaults for Proven; the cost if wrong is one line.
2. Features are stored as Actions secrets with the keys, so the page shows what was stored. A feature is kept out of the "keys stored" list and out of the sample-mode drop list, so it never makes a lab live and survives sample mode.
3. A second run of a head already reviewed ends successfully and posts nothing, rather than failing. A failing run would turn the head's status red for a harmless duplicate.
4. `make livecheck` with no key exits 1 and says the workflow runs a sample review. It could exit 0; 1 keeps the old exit and says "nothing live is ready".
5. The livecheck derivation is shared by running marked lines of the workflow file, because the job image has no copy of any lab script. The markers are `derive-begin` and `derive-end`; livecheck refuses by name when they are gone.

## master-was-wrong
(a) The order's reading that all four features can be switched off by a Proven setting is wrong for H-MEM and the World Model: no switch exists (red item 1). (b) "Each answer maps to the existing Proven setting" is true for Markovian (`PROVEN_MARKOVIAN`) and only through a file for Scout Intent (`scout_intent` has no environment spelling; red item 3). (c) The order read livecheck's disagreement as two copies of the defaults; the measured cause is the same: `livecheck.sh` had no derivation at all and left the generator mode to Proven's own default. (d) Not measured: the cause of the two runs.

## Blocked on / next
Nothing blocks. Next: the master runs one push to a pull request branch on a Docker host and the owner's Steps 1c and 1d, and decides whether Proven should add switches for H-MEM and the World Model.

## REDACTED
Nothing was blacked out: no credentials, host names, user names, absolute paths or private-record names appear above (0 items).
