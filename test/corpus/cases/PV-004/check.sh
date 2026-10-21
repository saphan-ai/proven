#!/usr/bin/env bash
set -euo pipefail
CASE_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$CASE_DIR"
npm ci --ignore-scripts >&2
"$CASE_DIR/node_modules/.bin/tsc" -p "$CASE_DIR/tsconfig.json" >&2
node "$CASE_DIR/dist/check/run.js"
