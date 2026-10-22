#!/usr/bin/env bash
# Build meterless-conform and run both Meterless runners against the shims.
# Works from any directory. Writes SCORECARD_<engine>.txt beside this script.
# Exit 0 when both runners pass, 1 when a check fails, 2 when a shim does not load.
set -u
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORK="$ROOT/.saphan/work/meterless-conformance-shim"
mkdir -p "$WORK"
if ! go -C "$ROOT" build -o "$WORK/meterless-conform" ./cmd/meterless-conform; then
  exit 3
fi
export METERLESS_CONFORM_BIN="$WORK/meterless-conform"

run_engine() {
  local engine="$1"
  local var="$2"
  local impl="$3"
  local runner="$4"
  local card="$ROOT/conformance/meterless/SCORECARD_${engine}.txt"
  local status=0
  env "${var}=${impl}" node "$runner" >"$card" 2>&1 || status=$?
  printf '%s\n' "----- ${engine} exit ${status} -----"
  cat "$card"
  return "$status"
}

hstatus=0
run_engine hmem HMEM_IMPL \
  "$ROOT/conformance/meterless/hmem/index.ts" \
  "$ROOT/.provision/reference/meterless/hmem/conformance/runner.ts" || hstatus=$?

wstatus=0
run_engine world-model WORLD_MODEL_IMPL \
  "$ROOT/conformance/meterless/world-model/index.ts" \
  "$ROOT/.provision/reference/meterless/world-model/conformance/runner.ts" || wstatus=$?

if [[ "$hstatus" -eq 2 || "$wstatus" -eq 2 ]]; then
  exit 2
fi
if [[ "$hstatus" -gt 2 || "$wstatus" -gt 2 ]]; then
  if [[ "$hstatus" -gt "$wstatus" ]]; then
    exit "$hstatus"
  fi
  exit "$wstatus"
fi
if [[ "$hstatus" -ne 0 || "$wstatus" -ne 0 ]]; then
  exit 1
fi
exit 0
