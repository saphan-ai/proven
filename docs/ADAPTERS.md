# Adapters

A tool adapter is how a deterministic tool becomes evidence. A language adapter is a different interface; this page's worked example is the tool, because that is what is built. The language decision is [ADR-0038](adr/ADR-0038-a-language-is-one-small-adapter.md).

## A tool is a third kind of evidence

[ADR-0040](adr/ADR-0040-a-tool-run-is-evidence.md). A finding can be proved by a test, by a query, or by a tool. A tool proves a finding when it reports a result on the change and does not report that result on the base. The evidence records the tool, its version, the exact command, and the output that differs.

The proof level is `proved-by-tool`. It sits between a test and a query ([ADR-0044](adr/ADR-0044-every-finding-names-its-level-of-proof.md)).

A result the base already had is not evidence. A tool that does not answer is named, and it is not counted as a clean run. A run past its time limit is that kind of miss.

## What an adapter implements

One adapter is one tool. It answers four questions:

1. **Which files.** `Applies` is true when the changed paths are ones this tool can read. The race detector applies to a changed Go file. A change with no Go file does not run it.
2. **The command.** `Argv` returns the words Proven runs. There is no shell. A path that would mean something else to the tool, such as `..` or `...`, is refused and nothing is started.
3. **The output.** `Read` classifies one side. For the race detector, the line `WARNING: DATA RACE` means a race. A test that fails without that line is not a race. The spelling is the one in the race runtime of Go 1.26.
4. **The comparison.** The runner does this, not the adapter. The change runs first. Evidence is a result on the change and not on the base.

The runner lives in `internal/tool`. The sample implementation replays a recorded outcome and starts no container. The live implementation runs the command in the Prover's container.

## The container

The command is the tail of the Prover's run line ([ADR-0025](adr/ADR-0025-the-prover-runs-the-change-with-no-network-and-no-secret.md)): no network, no secret, a read-only checkout, a CPU limit and a memory limit. The same argv builder builds that line for a test and for a tool. A checkout that is a symbolic link is not mounted.

A real run needs a container runtime on the host. This wagon's tests do not start one. They check the run line and they replay recorded output.

## The time limit

`PROVEN_TOOL_TIMEOUT` is the limit for one side, the base or the change. Empty means `120s`. A value that is set and is not a Go duration greater than zero is refused before a run, and it is not replaced by the default ([ADR-0013](adr/ADR-0013-no-hardcoded-settings.md)). Sample mode does not read the variable.

The limit is the tool's own. It does not change the limit of a test run. A side that passes the limit is named `timeout` and is not evidence, even when its output already contains a result.

## The race detector

The first tool is the Go race detector. Its id is `go-race`. Its command is:

```text
go test -race ./<package>
```

One `./<package>` for each directory of a changed Go file, sorted. A Go file at the checkout root is the package `.`.

### Sample scenario

Sample mode runs no container. The recorded scenario `RaceIntroduced` has the detector's banner on the change and a passing line on the base. The version in that recording is `go1.26`, the module's go line, not a measurement from a container. A live run that does not ask the container for a version records `unknown`.

That scenario reports one finding. The Judge keeps it. Its proof level is `proved-by-tool`.

Two recordings are the controls:

- `RaceOnBoth` has the banner on both sides. It is not a finding.
- `RaceTimeout` is a change that did not answer. The reason names `go-race` and `timeout`.

## Adding the next tool

Fuzzing, gosec and coverage are later. A new tool is a new adapter with the four answers above, a recorded scenario, and tests for three outcomes: a result only on the change, the same result on the base, and a run that does not answer. It uses the same container and the same time-limit setting. It does not grow a second way to start a process.
