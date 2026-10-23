#!/usr/bin/env bash
# Run from the repository root:
#   bash test/corpus/cases/PV-001/check.sh
# Dependencies come from this case's package-lock.json. Install them first:
#   npm ci --prefix test/corpus/cases/PV-001
set -euo pipefail

CASE_DIR=$(cd "$(dirname "$0")" && pwd)

if [ ! -f "$CASE_DIR/node_modules/typescript/bin/tsc" ]; then
  echo "pv-001: dependencies are not installed. Run: npm ci --prefix test/corpus/cases/PV-001" >&2
  exit 1
fi

node "$CASE_DIR/node_modules/typescript/bin/tsc" -p "$CASE_DIR/tsconfig.json"
node "$CASE_DIR/dist/check.js"
