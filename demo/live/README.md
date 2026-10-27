# Live demo

A small Go change a judge can review live. One finding is a logic bug a test can prove. One finding is a dependency a KEV query can confirm.

`base/` is the tree before the change. `change/` is the tree after it. That is the same shape as `testdata/fixture/discount`: `base/`, `change/`, `fixture.json` and `tables.sql`.

## The bug

`split.Share` divides cents across a number of people and returns the leftover cents. In `change/` the leftover is always 0. `TestShare` expects 100 cents across 3 people to leave 1 cent. It passes on `base/` and fails on `change/`.

## The dependency

`change/go.mod` requires `golang.org/x/net v0.16.0` and does not import it.

CVE-2023-44487 is the HTTP/2 rapid reset flaw. The Go release that fixed `golang.org/x/net` is v0.17.0, so v0.16.0 is the release before that fix. This run did not read that advisory or the CISA catalogue file, so both claims are UNVERIFIED here. A review that asks Zetaris

```sql
SELECT cve_id, vendor_project, product, date_added, known_ransomware
FROM px_evidence.px_kev_vulns
WHERE cve_id = 'CVE-2023-44487'
```

gets a row when that CVE is on the feed.

The one KEV row this repository records offline is CVE-2021-44228, Apache Log4j2, in `testdata/fixture/discount/tables.sql`. That product is not a Go module, so it is not this dependency. The live KEV view is the full feed (`fleet/status/STATUS_evidence-through-zetaris_r1.md`), so a CVE that is on the feed returns a row. If CVE-2023-44487 is not on the feed, that query returns no row and the dependency finding has no evidence.

No `go.sum` line. `go build ./...` and `go test -count=1 ./...` on a copy of each tree, with `GOPROXY=off` and `GOSUMDB=off`, do not ask for one. The module is not imported, so the build does not load it.

## How a live review takes this directory

A review with an empty baseline reads `base/` and `change/`. It does not need a git repository of its own. `run.sh` does not create one.

`fixture.Load` still runs for a live review. It refuses a directory that has no `fixture.json` or no `tables.sql`, and it checks `planner_prompt_sha256` against the trees. Those two files are here so the directory loads. Live mode does not use the recorded planner response, and it does not read `tables.sql`.

`proven scan` on this branch accepts only `--sample`. `run.sh` builds that binary, then runs the live review through `demo/live/review`, which calls `review.Open` with sample mode off and an empty baseline.

## Run

From any directory:

```sh
bash demo/live/run.sh
```

The script builds `proven`, `livecheck` and the live review command. It then runs `livecheck`. If any service is not ready, it prints that output and stops. It does not start the review. If every service is ready, it reviews this directory and prints each finding, its evidence, and the trace path. The trace is `logs/trace.jsonl` in the proven repository.

Settings come from the environment only. The script does not print a secret. A value of `METERLESS_RUNTIME_TOKEN`, `PROVEN_GUARDRAIL_KEY`, `ZETARIS_API_KEY` or `ZETARIS_ORG_ID` is replaced with `[redacted: NAME]`, naming its variable, if a service echoes it. The same marker replaces it in the trace file, in the review's models and in a failed service's reason. A runtime token, guardrail key or API key shorter than eight characters is refused when the settings load, naming its variable and never printing the value, so every secret that is configured can be replaced. An empty `METERLESS_RUNTIME_TOKEN`, `PROVEN_GUARDRAIL_KEY`, `ZETARIS_API_KEY`, `ZETARIS_REST_URL` or `ZETARIS_ORG_ID` makes that service down and names it, and an empty guardrail key refuses every model call unless `PROVEN_GUARDRAIL_MODE` is `off`.

### Generator

| name | default |
|---|---|
| `METERLESS_RUNTIME_TOKEN` | none; required. Secret. |
| `PROVEN_GENERATOR_BASE_URL` | `http://127.0.0.1:8088` |
| `PROVEN_GENERATOR_APP` | `proven` |
| `PROVEN_GENERATOR_MODEL` | `meterless/auto` |
| `PROVEN_GENERATOR_TIMEOUT` | `60s` |
| `PROVEN_GENERATOR_REQUIRE_NVIDIA` | `all` |
| `PROVEN_SKEPTIC_MODEL` | `nvidia/nemotron-3.5-super-vl-120b-a12b` |

`PROVEN_GENERATOR_REQUIRE_NVIDIA` is `all`, `none`, or a comma-separated list of `planner`, `investigator`, `prover`, `skeptic`, `judge`, `writer`.

### Guardrail

| name | default |
|---|---|
| `PROVEN_GUARDRAIL_MODE` | `server` (`server`, `nvidia`, or `off`) |
| `PROVEN_GUARDRAIL_BASE_URL` | `http://127.0.0.1:8081` in server mode; `https://integrate.api.nvidia.com/v1` in nvidia mode |
| `PROVEN_GUARDRAIL_MODEL` | `nemotron-3.5-content-safety` in server mode; `nvidia/nemotron-3.5-content-safety` in nvidia mode |
| `PROVEN_GUARDRAIL_KEY` | none; required unless mode is `off`. Secret. |
| `PROVEN_GUARDRAIL_TIMEOUT` | `30s` |

### Zetaris

| name | default |
|---|---|
| `ZETARIS_REST_URL` | none; required |
| `ZETARIS_API_KEY` | none; required. Secret. |
| `ZETARIS_ORG_ID` | none; required. Digits, at most 18. |
| `ZETARIS_ROW_LIMIT` | `100` |
| `ZETARIS_TIMEOUT` | `60s` |

### Livecheck

| name | default |
|---|---|
| `PROVEN_EVIDENCE_VIEW` | `px_evidence.px_kev_vulns` |

The container check has no setting. It uses runtime `docker` and image `golang:1.26`. It does not pull the image.
