#!/usr/bin/env bash
set -euo pipefail
CASE_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$CASE_DIR"
CACHE_DIR="$CASE_DIR/../../../../.saphan/work/corpus-pv-004/npm-cache"
mkdir -p "$CACHE_DIR"
npm ci --ignore-scripts --cache "$CACHE_DIR" >&2
"$CASE_DIR/node_modules/.bin/tsc" -p "$CASE_DIR/tsconfig.json" >&2
node "$CASE_DIR/dist/check/run.js" > "$CACHE_DIR/check.stdout"
diff -u "$CASE_DIR/check/expected.stdout" "$CACHE_DIR/check.stdout" >&2
cat "$CACHE_DIR/check.stdout"
