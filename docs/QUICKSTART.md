# Quick start: your first checkable review

Start with a recorded example, inspect its evidence, then decide which live components you want. You need the Go version specified in `go.mod`, Git and a shell. The recorded example needs no API key, database server or container runtime.

## 1. Build the CLI

From the repository root:

```sh
go build -trimpath -o bin/proven ./cmd/proven
./bin/proven --help
```

Keep this checkout separate from the repository being reviewed. Set a shell variable for the executable before changing directories:

```sh
PROVEN_BIN="$PWD/bin/proven"
PROVEN_FIXTURE="$PWD/testdata/fixture/discount"
PROVEN_RUN_DIR=$(mktemp -d)
cd "$PROVEN_RUN_DIR"
```

The run directory is deliberately outside the target. Logs, evidence stores and persistent state must not be written into the source under review.

## 2. Run the recorded defect

```sh
"$PROVEN_BIN" scan --sample \
  --json-output review.json \
  --html-output review.html \
  "$PROVEN_FIXTURE"
```

The text report lists four findings in this fixture: a discount applied twice and dependency findings backed by synthetic data. It names the proving test, base/head outcome, query templates and returned rows. `review.html` is a standalone report you can open locally; `review.json` is the machine-readable result. The last text line gives the trace path.

Sample mode executes the decision and formatting code, but replays the model replies, test outcomes and source rows. It does not prove that your machine can call a live model or run a container. See the [walkthrough](WALKTHROUGH.md).

## 3. Understand the exit code

The first invocation is advisory: a finding does not by itself make `scan` fail. To make findings fail:

```sh
status=0
"$PROVEN_BIN" scan --sample --error "$PROVEN_FIXTURE" || status=$?
printf 'exit=%s\n' "$status"
```

Expect exit 1 for this recorded defective change. Required-completion failure is separate and has exit 3; a completed advisory command can still describe incomplete analysis. [Readiness](READINESS.md#reading-completion-correctly) explains the codes.

## 4. Choose your next path

| Your goal | Next guide | What it adds |
| --- | --- | --- |
| Understand why a finding was reported | [Read one review](WALKTHROUGH.md) | Follow a claim through evidence, refutation and Judge. |
| Use local data without a model | [Data walkthrough](DATA_WALKTHROUGH.md) | Create a real Git change and query an imported SQLite snapshot. |
| Review code with your model | [Models](MODELS.md) | Configure the endpoint or CLI, check compatibility and tune limits. |
| Execute generated tests | [Restricted runner](RUNNER.md) | Prepare the target project's toolchain and dependency cache. |
| Check pull requests | [GitHub Action](ACTION.md) | Pin tooling, choose enforcement and retain reports. |
| Operate for a team | [Operations](OPERATIONS.md) | Separate state, credentials, policy and updates. |

For a live Git change, the basic command is:

```sh
proven scan --baseline-commit <base-sha> /path/to/repository
```

Use an actual base commit available in that clone. It is not a scan of every historical defect. With no model configuration, model stages are off. A new database has no public feeds: import prepared source tables before relying on dependency evidence. A real generated-test run additionally needs a prepared [runner image](RUNNER.md).

## 5. Before enforcement

Check `completion_status`, failed services and `unresolved_hypotheses`, not only the findings count. Review the supported languages, dataset revisions, configured model identity and tool availability. A recorded success or an API probe is not an accuracy benchmark. Start with cases whose expected outcome you know, including a deliberate defect and a change without a defect.

See [readiness and limits](READINESS.md), [troubleshooting](TROUBLESHOOTING.md) and the [complete documentation map](README.md).
