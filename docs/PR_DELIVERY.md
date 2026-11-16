# Does the PR deliver what it says?

The optional delivery check compares a PR's **title, description and commit messages** with the selected head code. It extracts concrete claims and reports code indications, contradictions or uncertainty, with checked citations. This assessment is separate from test/query/tool-proved defects.

## Turn it on

Add one input to the [CLI Action](START.md):

```yaml
- uses: saphan-ai/proven/cli@main
  with:
    backend: claude
    image: proven-claude:local
    check-delivery: true
```

The Action reads the PR event and actual Git commit messages. It writes bounded context outside the reviewed checkout; PR text cannot change configuration or supply commands. It uses the same authenticated model connection, guardrail, timeout, recovery and secret filtering as the ordinary review. No separate model account or PR API request is needed.

## Read the table

| Assessment | What it means | Evidence attached |
| --- | --- | --- |
| **Indicated by code** | The model identified code that appears to implement a stated claim. | Exact metadata excerpt and exact head-code lines, validated against the supplied context. |
| **Contradicted** | The model identified code that appears inconsistent with a stated claim. | Checked code citations; an already proved finding may also be linked as a counterexample. |
| **Unclear** | The supplied code/context does not settle the claim. | The reason and any available citations. |

For example, “reject negative amounts” may point to a negative-amount guard. Its presence is a code indication. A differential test that demonstrates an actual failure remains a separate proved finding; the delivery check does not silently promote model agreement into executable proof.

The comment and HTML report have a delivery table. JSON uses the optional `delivery` object; SARIF places it in run properties, outside vulnerability results. Evidence packs retain the assessment as report data. `proven verify` checks the signed recording and ordinary evidence decisions; it does not execute acceptance tests or independently reassess the meaning of delivery claims.

## Boundaries

The model can miss or misunderstand a requirement. Citation checks establish that quoted text exists at the stated lines, not that the semantic interpretation is correct. An assessment with no extracted claims does not certify delivery. A `complete` delivery assessment means the assessment completed, not that every requirement was implemented.

The source budget is 32 selected files, 8 KiB per file and 128 KiB total; truncation is recorded. Metadata is limited to 64 KiB and output to 24 claims with at most four citations each. Statements about omitted files, UI behavior, external services or deployment may remain unclear. Discussion-thread comments are not included by the automatic context collector.

Contradictions and unclear claims are advisory and do not become findings or alter proof counts. A failed requested assessment is reported as an incomplete service; existing findings are preserved. Standard `require-complete` policy then applies. The check is off unless requested.

## Use it outside GitHub Actions

Create an external context file:

```json
{
  "url": "https://github.com/example/project/pull/42",
  "title": "Reject negative amounts",
  "description": "Reject negative amounts before applying a discount.",
  "commits": [{"sha": "<actual commit SHA>", "message": "Add negative-amount validation"}]
}
```

Point `PROVEN_PR_CONTEXT` at that file and run a normal scan with your configured model. Keep it outside the target repository. The report records a metadata SHA-256, source commit references and the reviewed snapshot, so later changes to the PR wording do not silently become the wording used for the recording.

This is distinct from **Scout Intent**, which selects a review plan. Delivery assessment evaluates stated scope; ordinary evidence gates decide defect findings. [What Proven proves](WHAT_PROVEN_CHECKS.md) explains the difference.
