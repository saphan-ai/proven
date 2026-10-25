# Proven's copy of the Scout eval runner

`index.ts` here is a copy of Meterless's `engines/scout-intent/evals/runner/index.ts` (Apache-2.0), taken at the
commit `SCORECARD_scout-intent-changes.txt` names. It scores Proven's change registry
(`internal/scout/changes.json`) on Proven's change set (`../changes/v1/`).

Meterless's own runner is not edited and is not replaced: `run_scout.sh` runs it, unchanged, on Meterless's own
regression set. This copy exists because that runner reads its registry, its corpus and its thresholds from fixed
paths beside itself, so it cannot score a second registry without being edited.

## Every difference from the published runner

Each is marked `PROVEN:` in the code.

1. **Thresholds** are read from `SCOUT_THRESHOLDS`, a JSON file, instead of `../config/thresholds.yaml`. The
   `yaml` import is gone: the copy runs with Node's own type stripping and no installed package.
   `thresholds.json` here carries the published values unchanged, with three additions: the slice value
   `change` for the `surface` and `intent_family` dimensions, and the `clarification_agreement` floor below.
2. **The registry** is `SCOUT_REGISTRY` instead of `../fixtures/intents.json`.
3. **The corpus** is every `.jsonl` file in every subdirectory of `SCOUT_CORPUS`, instead of the fixed list of
   `intents`, `injection`, `multi-intent` and `edge-cases` under `../regression-set/v1`. The command-line filter
   (`all`, `intent`, …) that chose among those directories is gone with them.
4. **The report** is written to `SCOUT_REPORT` instead of `../reports/latest.md`, so a run never writes into the
   Meterless checkout.
5. **A missing setting** (`SCOUT_REGISTRY`, `SCOUT_CORPUS`, `SCOUT_THRESHOLDS`, `SCOUT_REPORT`) stops the run with
   exit 2, like a missing `SCOUT_IMPL`. The long explanation the published runner prints for a missing
   `SCOUT_IMPL` is cut to one line.
6. **One metric is added**, `clarification_agreement`: of the examples that state `expected.clarification`, the
   share whose answer asked exactly when expected. The published runner counts how often an implementation asks
   (`clarification_rate`) but never whether it asked at the right time, and an ambiguous change is the case
   ADR-0049 most needs right. Its floor, 0.85, is Proven's own choice, set to the published `intent_top1` floor.

Nothing else differs: the tally, the metrics, the slices, the gate and the report are the published code.

## A number this set cannot meet

`clarification_rate` keeps the published ceiling of 0.25. The order for this set asks for one ambiguous fixture per
kind, so at least six of its fourteen examples must ask, and an implementation that asks when it should scores
about 0.43. The ceiling is kept as published (it is not this copy's to move), and the scorecard records the failure
as it is. `clarification_agreement` is the number that says whether the asking was right.
