# Proposed fixes and whole-change autofix

Proven can propose edits after finding evidence. It validates edits on copies and reports what passed. It does not commit, merge or silently change the author's branch. A model's proposed patch is not a proved fix until the required execution checks pass.

## Two workflows

| Workflow | Default | Scope |
| --- | --- | --- |
| Per-finding suggestions, `fixes` | `on` | A source edit for a test-proved finding, or a fixed dependency version from advisory data. |
| Whole-change attempt, `autofix` | `off` | One patch addressing all proved findings, followed by evidence and collateral checks. |

Configure these repository policy keys in `proven.yaml`. The complete contract is in [Configuration](CONFIGURATION.md#provenyaml). They do not enable a model by themselves; a live edit proposal needs an appropriate configured generator.

## A per-finding source fix

The Fixer receives a proved finding and proposes an edit. Proven applies it to a scratch copy and uses the finding's own test. The acceptance conditions are:

| Gate | Required observation |
| --- | --- |
| Base | The proving test passes. |
| Head | The same test fails. |
| Patched head | The same test passes after the proposed edit. |
| Other tests in the package | Existing package tests pass on the patched copy. |

A run that could not start, timed out or failed to compile does not stand in for a passing gate. A patch without the required proof is withheld as an accepted fix; the report explains that it was not proved.

`fix_attempts` defaults to 2 and accepts 1–5. A retry receives the previous failure. The transport's own recovery budget is separate, and all work remains bounded by the review deadline. Increasing attempts also increases runtime and cost.

## Dependency suggestions

An advisory's first fixed version can be offered with a dependency finding. This is a data-derived boundary, not an observed build of the project after an upgrade. The data may be stale, the upgrade may be incompatible and no version may be supplied. Keep that distinction when presenting or applying the suggestion.

The affected range uses ecosystem-specific version comparison. Fixed is exclusive; `last_affected` is inclusive when no fixed boundary exists. See [OSV schema and version rules](OSV_TABLE.md).

## Whole-change autofix

An explicit policy might contain:

```yaml
autofix: on
autofix_generator: default
autofix_timeout: 10m
```

This makes one proposal for the whole change, with every proved finding and its evidence. `special` selects the configured second reviewer's generator; when unavailable, Proven names the failed service instead of borrowing an undisclosed replacement.

Each finding is checked with its own kind of evidence:

- A test-backed finding closes when its proving test and the package's other tests pass on the patched copy.
- A data-backed finding closes when the same evidence query no longer returns the relevant row under the recheck rules.
- A tool-backed finding closes when the same tool run no longer reports it.

Proven also runs package tests for every package touched by the patch, or the whole module for a module-file edit. A failing/no-run collateral check rejects the attempt. A touched package without tests cannot demonstrate that nothing broke and is rejected. Edits outside a package of the module, including scripts, documents or a nested module, can produce a no-run rejection too.

The result is accepted for human review (“fixed automatically, waiting for a person”) or rejected with open findings and failed checks. There is no second whole-change attempt. Per-finding suggestions are not generated while autofix is on. `fixes: off` governs those suggestions and does not disable an explicitly enabled whole-change attempt.

## Offline execution still applies

Fix validation uses the [restricted runner](RUNNER.md). A dependency upgrade that needs an unavailable module cannot fetch it during proof. Prepare the cache/image or vendored dependencies first; a missing dependency is a rejected/no-run outcome, not an acceptable fix.

The patch remains a proposal even when the recorded checks pass. The checks establish the implemented acceptance conditions, not exhaustive semantic correctness or deployment suitability. Review the diff, evidence and coverage gaps before applying it.

## Learn with recorded controls

The [sample catalogue](SAMPLE_MODE.md#the-scenario-testdatafixturefix) includes a per-finding fix with its proof. `testdata/fixture/autofix-fixed` and `testdata/fixture/autofix-rejected` exercise whole-change accepted/rejected results without live models or containers. Their outcomes are recorded; use a prepared real runner to validate your operational environment.
