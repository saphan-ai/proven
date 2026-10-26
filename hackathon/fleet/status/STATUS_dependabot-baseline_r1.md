# STATUS — dependabot-baseline · round 1

- state: ready-for-review
- executor: claude-code
- model: Sonnet 5.5 (claude-sonnet-5-5)
- last commit: see the commit that adds this file; the work commits before it are eb4e79f, 0f13b35, 5c97847 and b335314
- updated: 2026-10-26

## Summary

Proven's dependency check now runs over the manifests of the pinned commit `fbdf7d3`, read as a change from an empty base, and its result is compared with the 51 frozen Dependabot alerts, alert by alert. No evidence store and no OSV snapshot could be reached from this seat, so the committed report is the manifest-only run: Proven reads all 51 alerted packages and 51 rows are `not checked`. Whether Proven finds the advisories is not measured yet. The run against Zetaris is a numbered step in `test/README.md`, section 9.

## Acceptance, item by item

1. **A repeatable run.** `make baseline-dependabot` (asks the store `PROVEN_EVIDENCE_STORE` names) and `make baseline-dependabot-manifests` (asks none). `make` is not on this seat's allowed program list, so neither target was run through `make`. Their one body, `sh lab/scripts/baseline-dependabot.sh manifests`, was run: exit 0, and it wrote `proven-findings.json` and `REPORT.md`. `sh lab/scripts/baseline-dependabot.sh evidence` with no Zetaris settings: exit 2, "the Zetaris store is not configured … or pass -manifest-only", nothing written. `sh -n` on the script: exit 0.
2. **Evidence.** The store is the one setting `PROVEN_EVIDENCE_STORE`, with the same loaders a review uses. The local store needs `BASELINE_EVIDENCE_LABEL` and `BASELINE_SNAPSHOT_DATE`, and every verdict line carries the `evidence:` phrase. **Not run:** the local OSV snapshot the order asks for. Measured: `python3` could not resolve `api.osv.dev` (name or service not known), and no OSV advisory file exists on the seat. An OSV table built from `alerts.json` would find all 51 by construction. It was run once as a harness self-check, with the phrase `SELF-CHECK ONLY, circular`, and not committed.
3. **The comparison.** `test/corpus/reference/dependabot-fbdf7d3/REPORT.md`, generated, with counts at the top: 51 alerts, 0 found, 0 found differently, 0 missed, 51 not checked. Measured on the manifests: Proven reads all 51 alerted packages in the file the alert names, and for each alert at least one version it read lies inside the range Dependabot states (`postcss`: 8.4.31 inside, 8.5.31 outside). The section "What Proven found that Dependabot did not" says `Not checked` for this run.
4. **Docs.** `test/README.md` section 9 and the reference README: how to run, what evidence, what it shows, the honest framing, and the numbered runbook for the Zetaris run.

## Evidence

- `go build ./...` exit 0. `go vet ./...` exit 0.
- `go test -count=1 ./internal/baseline/ ./cmd/dependabot-baseline/`: ok, ok. The same two with `-race -count=3`: ok, ok.
- `go test -count=1 ./...`: every package `ok` except one. `internal/cli/sarifschema` `TestSampleSARIFMatchesTheSchema` fails with `open <user cache>/proven/logs/trace.jsonl: permission denied`: the seat's user cache directory is not writable, and the test ignores `PROVEN_LOG_DIR`. The change does not touch that package or its test. **Not run:** the same test on `main`, because creating a second worktree was denied on this seat, so "also red on main" is unverified.
- Controls, each a test in `internal/baseline` or `cmd/dependabot-baseline`:
  - Mutated field: the package's entry in the lock copy. With it, alert 1 is `found`. Without it, alert 1 is `missed` and the detail says Proven does not read the package. `TestRemovingAPackageFromTheLockTurnsItsRowMissed`.
  - Mutated field: the GHSA and CVE ids of one alert in a copy of the reference. The row is `found differently` when another advisory affects the package, `missed` when none does, and never `found`. `TestChangingAnAlertsAdvisoryIDNeverGivesFound`.
  - The report is byte-identical twice: `TestReportIsByteIdenticalTwice`, and `TestLocalRunIsRepeatableAndNamesItsEvidence` for the findings file as well. The committed report was regenerated from the repository and `git diff` showed no change.
  - Revert proof: with the "package is not read" check disabled, three tests failed. With the advisory-id match in `judgeEvidence` removed, three tests failed. Both were restored.
  - `TestCommittedReportIsRegenerated` fails when the committed report differs from the one the two JSON files give. It also fails if removing one read does not change the report.
- `scripts/check-links.sh`: three broken links, all in `docs-site/generated/settings.md` (the known defect on main). No new ones.

## Blocked on / next

The run that decides the baseline needs an evidence store this seat does not have. The owner's hand: the runbook in `test/README.md`, section 9, steps 1 to 5. Until then the report says `not checked` and claims nothing.

## NOTICED, not changed

- `internal/pipeline` keeps its own unexported copy of the store-to-query seam; `store.Runner` now exports the same thing.
- The report lists every manifest under `test/` and `testdata/` that Proven reads, because the repository at a commit includes them. A full-snapshot run will list their findings under "What Proven found that Dependabot did not".

## REDACTED

none
