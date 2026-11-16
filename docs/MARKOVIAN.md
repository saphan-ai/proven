# Bounded planning and unresolved hypotheses

Markovian splits a large change into Planner steps. Each step reads one changed file (or one hunk), plus a bounded summary carried from earlier steps. It reduces the context sent on each call. It can lose context across files; its effect on defect-detection quality has not been established.

## Configuration

| Setting | Default | Meaning |
| --- | --- | --- |
| `PROVEN_MARKOVIAN` | `on` | Enable bounded planning above the threshold; `off` reads the changed files together. |
| `PROVEN_MARKOVIAN_THRESHOLD` | `8000` | Estimated tokens above which planning is split. This estimate is characters divided by four, not provider-reported tokens. |
| `PROVEN_MARKOVIAN_STEP_UNIT` | `file` | `file` or `hunk`. A single file may itself be large. |
| `PROVEN_MARKOVIAN_CARRY_TOKENS` | `800` | Bound on carried state. Claims awaiting evidence are kept whole. |

Whole-change planning needs a model context large enough for all selected changed files. Turning the engine off does not change file selection, evidence requirements or runner isolation. In the repository's own workflow, `PROVEN_REVIEW_MARKOVIAN` selects this setting. The application default and a repository workflow's selected profile are separate configurations.

## What happens at the last step

The Planner returns identified risks and a state containing established points, suspicions and `awaiting_evidence` claims. A final claim awaiting evidence is still a hypothesis. It has no proof and is not a finding.

When the last step leaves such claims:

1. Identified risks continue through Investigator, Prover, Skeptic and Judge. Findings supported by evidence are retained.
2. The remaining descriptions are kept in `unresolved_hypotheses`, with reviewer role and the step after which they remained open. The trace records each as `unresolved-hypothesis`.
3. The report has `completion_status: incomplete`. `partial` remains false if the deadline did not interrupt execution. Model and database services can all have answered while coverage is still incomplete.
4. Required completion exits with code 3 after writing the report and posting the configured summary. Advisory scans can return 0 while still reporting incomplete coverage.

For example, a run that finished planning and found no proved defect can contain:

```json
{
  "completion_status": "incomplete",
  "partial": false,
  "findings": [],
  "failed_services": [],
  "unresolved_hypotheses": [{
    "reviewer": "default",
    "after_step": 33,
    "description": "A caller outside the selected change is needed to check this claim."
  }]
}
```

This is a coverage gap, not a clean bill of health or an independently verified defect. Read the descriptions and supply the missing context or evidence before relying on completeness. Reports redact configured secrets and escape model text for their output format. The full description is in JSON and SARIF; compact comments may shorten it.

## Other stopping conditions

A full carried state is different from a final open claim. If all awaiting claims cannot fit within the carry bound, planning still stops rather than silently discarding them. That refusal currently returns exit 2 and may leave only a trace. Increasing the carry bound or selecting whole-change planning are operator choices, with different context costs.

Timeouts and unavailable model calls retain their own failure handling. See [readiness and limitations](READINESS.md) and [configuration](CONFIGURATION.md).

## Observed case and migration

On 16 November 2026, a self-review completed 33 bounded calls successfully but retained one final awaiting claim. The earlier implementation returned a fatal `OpenClaimsError` before evidence gathering and wrote no result report; its trace retained only the count. Whole-change planning was temporarily used to test the model transport. The current implementation preserves the final descriptions and continues verification as described above. It does not reconstruct missing descriptions from older traces.

This fixes result handling. It does not establish that bounded planning finds as many defects as whole-change planning. Compare both against controlled defects and changes without defects before selecting a profile for enforcement.
