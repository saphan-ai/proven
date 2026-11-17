#!/bin/sh
# Real Git + SQLite + deterministic rules. All data is synthetic teaching data.
# Run from the product image; no model, network, Docker socket or vendor CLI.
set -eu
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
export PROVEN_GENERATOR_MODE=off PROVEN_GUARDRAIL_MODE=off PROVEN_NETWORK=local
export PROVEN_EVIDENCE_STORE=sqlite PROVEN_EVIDENCE_FILE="$work/evidence.sqlite"
export PROVEN_LOG_DIR="$work/logs" PROVEN_REQUIRE_COMPLETE=true
mkdir "$work/repo"
git init -q "$work/repo"
printf 'example-lib==0.9.0\n' > "$work/repo/requirements.txt"
git -C "$work/repo" add requirements.txt
git -C "$work/repo" -c user.name=Example -c user.email=example@example.com commit -qm 'Create synthetic dependency baseline'
base=$(git -C "$work/repo" rev-parse HEAD)
printf 'example-lib==1.1.0\n' > "$work/repo/requirements.txt"
git -C "$work/repo" add requirements.txt
git -C "$work/repo" -c user.name=Example -c user.email=example@example.com commit -qm fix
proven data import /opt/proven/fixture/model-free-tables.json >/dev/null
cd "$work"
rc=0
proven scan --error --json --baseline-commit "$base" "$work/repo" > "$work/report.json" || rc=$?
[ "$rc" -eq 1 ] || { echo 'model-free smoke: expected one synthetic dependency finding' >&2; cat "$work/report.json"; exit 1; }
grep -Eq '"models"[[:space:]]*:[[:space:]]*\[\]' "$work/report.json"
grep -q 'proved-by-data' "$work/report.json"
grep -q '"placeholder"' "$work/report.json"
grep -q '"source_checks"' "$work/report.json"
grep -q '"completion_status": "complete"' "$work/report.json"
echo 'model-free smoke: Git, SQLite, dependency proof, source rules and per-commit description checks passed; no model or network'
