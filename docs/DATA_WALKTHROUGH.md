# Walkthrough: a Git change against a SQLite snapshot

This exercise uses a real Git comparison and the real SQLite query path, with **invented teaching data**. It needs no model key, database server or container runtime. It does not download a real feed.

The dataset and its advisory/CVE/EPSS values are synthetic. Keep it separate from operational evidence databases.

## 1. Prepare paths

Build Proven as in [Quick start](QUICKSTART.md). From the Proven checkout:

```sh
PROVEN_BIN="$PWD/bin/proven"
PROVEN_DATA="$PWD/examples/data/snapshot.json"
PROVEN_DEMO=$(mktemp -d)
mkdir "$PROVEN_DEMO/repository"
export PROVEN_EVIDENCE_STORE=sqlite
export PROVEN_EVIDENCE_FILE="$PROVEN_DEMO/evidence.sqlite"
export PROVEN_LOG_DIR="$PROVEN_DEMO/logs"
export PROVEN_GENERATOR_MODE=off
export PROVEN_GUARDRAIL_MODE=off
```

The database and logs are siblings of the target, not inside it. Do not reuse this synthetic database for real vulnerability checks.

## 2. Import three prepared tables

```sh
"$PROVEN_BIN" data import "$PROVEN_DATA"
```

Expect three imported-table messages: the OSV package table, KEV view and EPSS view. The KEV table is deliberately registered but empty. That is different from a missing source. Each imported table has a content identity; the supplied `snapshot` label contributes to that identity. Import does not certify freshness or source truth.

## 3. Create base and head

```sh
git -C "$PROVEN_DEMO/repository" init -q
printf 'example-lib==1.0.0\n' > "$PROVEN_DEMO/repository/requirements.txt"
git -C "$PROVEN_DEMO/repository" add requirements.txt
git -C "$PROVEN_DEMO/repository" -c user.name='Proven tutorial' \
  -c user.email='tutorial@example.invalid' -c commit.gpgsign=false commit -qm base
PROVEN_BASE=$(git -C "$PROVEN_DEMO/repository" rev-parse HEAD)
printf 'example-lib==1.1.0\n' > "$PROVEN_DEMO/repository/requirements.txt"
git -C "$PROVEN_DEMO/repository" add requirements.txt
git -C "$PROVEN_DEMO/repository" -c user.name='Proven tutorial' \
  -c user.email='tutorial@example.invalid' -c commit.gpgsign=false commit -qm change
```

The teaching advisory covers versions from 1.0.0 up to, but excluding, 1.2.0. Proven checks the changed requirement, 1.1.0. This exercise checks introduced/raised requirements; it is not a request to report every existing dependency issue in base.

## 4. Scan and inspect the evidence

```sh
cd "$PROVEN_DEMO"
PROVEN_REQUIRE_COMPLETE=true "$PROVEN_BIN" scan \
  --baseline-commit "$PROVEN_BASE" \
  --json-output review.json --html-output review.html \
  "$PROVEN_DEMO/repository"
```

Without `--error`, a finding remains advisory. Inspect the JSON's completion status, findings, source query, rows and snapshot identity. A finding should describe `example-lib` 1.1.0 inside the synthetic affected range and name 1.2.0 as the advisory boundary. There are no model calls in this exercise. A dependency version comparison is data evidence, not a generated-test proof.

The empty KEV table does not establish that the synthetic CVE is not exploited. The loaded table simply has no matching row. The fabricated EPSS score demonstrates enrichment and its date.

## 5. Test the boundary and the failure path

Change the head requirement to the exclusive fixed boundary and rerun against the same base:

```sh
printf 'example-lib==1.2.0\n' > "$PROVEN_DEMO/repository/requirements.txt"
git -C "$PROVEN_DEMO/repository" add requirements.txt
git -C "$PROVEN_DEMO/repository" -c user.name='Proven tutorial' \
  -c user.email='tutorial@example.invalid' -c commit.gpgsign=false commit -qm fixed
PROVEN_REQUIRE_COMPLETE=true "$PROVEN_BIN" scan \
  --baseline-commit "$PROVEN_BASE" --json-output fixed.json \
  "$PROVEN_DEMO/repository"
```

Expect a complete result with no affected-range finding. Compare the evidence rather than treating exit 0 alone as the answer.

Now select a different, empty database for one invocation:

```sh
status=0
PROVEN_EVIDENCE_FILE="$PROVEN_DEMO/empty.sqlite" \
  PROVEN_REQUIRE_COMPLETE=true "$PROVEN_BIN" scan \
  --baseline-commit "$PROVEN_BASE" --json-output empty.json \
  "$PROVEN_DEMO/repository" || status=$?
printf 'exit=%s\n' "$status"
```

Expect exit 3, `completion_status: incomplete` and a failed SQLite source naming the unregistered OSV table. The original database remains selected for later commands because the override applies only to this invocation. These three cases were executed on 16 November 2026: affected head → one finding/complete; fixed head → zero/complete; empty database → zero/incomplete/exit 3.

For real use, replace the synthetic snapshot with acquired, normalized source data. Keep provenance, acquisition/revision dates and update procedures. [Databases](DATABASES.md) describes the import contract; [Sources](SOURCES.md) separates available query adapters from planned feed connectors.
