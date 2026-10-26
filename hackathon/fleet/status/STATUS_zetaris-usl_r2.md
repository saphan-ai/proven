# STATUS — zetaris-usl · round 2

- state: ready-for-review
- executor: claude-code leg of the fleet (not a Cursor agent)
- model: Sonnet 5.5 (`claude-sonnet-5-5`), as the run named it
- last commit: the public copy of this status, on `fix/zetaris-usl` (`git log -1`)
- updated: 2026-10-25, UTC

## Summary
Round 1 was rejected by the named review on three blocking findings. All three are repaired, each with a negative control that was red before the repair or red when the repair is removed. The DROP syntax measured live after round 1 is kept. Nothing was sent to Zetaris: the leg had no network and no credential.

## Acceptance, item by item
1. **The gate reads only a well-formed answer** — ADDRESSED. `apply.sh` no longer uses `sed`. `lab/zetaris/usl/count.py` (run as `python3 -I`) requires one JSON document with `headers` of exactly `["violations"]` (case aside), `data` of exactly one row of one cell, the cell a non-negative whole number (an integer or a string of digits), and no `error`, `errors`, `exception`, `message` or `msg` field. Anything else is exit 5 naming what was wrong, and no settings are printed. `TestApplyRefusesAnUnreadableGateAnswer` now holds 17 bodies, including the review's two (`{"data":[[0],[7]]}`, `{"message":"partial","data":[[0]]}`), each mutating one field of the valid answer; before the repair 8 of them exited 0. `TestApplyReadsAWellFormedCount` holds the passing shapes. A missing `python3` is exit 2 before anything is sent.
2. **No diagnostic prints the organisation id or the key** — ADDRESSED. `redact` in `apply.sh` removes both (`[key]`, `[org]`) and is the only path of every refusal, transport error and printed row. `TestApplyRedactsTheOrganisationIdEverywhere` echoes the id and the key in a statement refusal, a transport error, a refused gate query and the rows of a refused gate, and asserts neither reaches stdout or stderr. With the id removed from `redact` the test went red in 4 places; restored, green.
3. **The relation grain is what the docs say** — ADDRESSED by documenting the real grain: one row per package range, per alias of its advisory, per EPSS score date. The view gained one column, `epss_score_date`, so rows of one CVE can be told apart. `docs/ZETARIS_USL.md` has a section "The grain of the relation view" and says how to count ranges or take the latest score. `TestRelationViewGrainIsPerAliasPerScoreDate` runs the view's own SELECT in `python3`'s sqlite3 over a fixture with two aliases and two score dates (3 rows for that range, 1 row of empty columns for a range with no advisory); `TestRelationViewGrainFollowsTheScoreDates` removes one score date and the rows fall from 4 to 3. The four entity views are unchanged, so parity is unchanged.

## Evidence
- `go vet ./...`: exit 0. `gofmt -l internal cmd`: no output.
- `go test ./... -count=1`: no package failed; `internal/usl` 24 tests (23 pass, 1 skipped by name: `TestParityLive`).
- `sh -n lab/zetaris/usl/apply.sh`: exit 0. `sh scripts/check-links.sh`: exit 0.

## Not run
- Anything against Zetaris. The relation view's new column and its grain on live data were not run there. The grain test runs in sqlite, not in Zetaris's dialect.
- `make zetaris-usl-parity`: `make` is not available to this leg.
- The review's owed remainder, a tolerated refusal that checks the specific condition and not the word "exist", is not changed.
- If a live Zetaris success answer carries a `message` field, the gate now refuses it as exit 5 and names the field. That answer shape has not been seen.

## Master was wrong
—

## Judgement calls
- A JSON parser in `python3` and not a Go tool: the apply script is shell, and `python3` is already on the runbook's path. Cost if wrong: a host without `python3` cannot apply (exit 2, nothing sent).
- Fields that refuse the answer are a closed list of five names. Cost if wrong: a live success answer carrying one of them is refused; the fix is one edit of `count.py`.
- Grain documented, not aggregated: aggregation across a container is unmeasured on Zetaris, and the review allowed either form. Cost if wrong: a reader who wanted one row per range filters as the docs say.

## REDACTED
None.
