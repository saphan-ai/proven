#!/usr/bin/env bash
# Build meterless-conform and run the three Meterless runners (H-MEM, World Model, Markovian) against the shims.
# Works from any directory. Writes SCORECARD_<engine>.txt beside this script.
# Exit 0 when every runner passes, 1 when a check fails, 2 when a shim does not load, 4 when no runner is found.
# METERLESS_ENGINES names a Meterless engines/ directory (required; it is not part of this repository).
set -u
# HERE is this directory (test/conformance/meterless); ROOT is the repository, three levels up.
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$HERE/../../.." && pwd)"
# The helper binary is built into a temporary directory, removed on exit, so the repository is never written to.
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
if ! go -C "$ROOT" build -o "$WORK/meterless-conform" ./cmd/meterless-conform; then
  exit 3
fi
export METERLESS_CONFORM_BIN="$WORK/meterless-conform"

# The Meterless engines directory (a checkout of Meterless's engines/, Apache-2.0). Not in this repository.
REF="${METERLESS_ENGINES:-}"
for e in hmem world-model markovian; do
  if [[ ! -f "$REF/$e/conformance/runner.ts" ]]; then
    echo "run.sh: no Meterless runner at $REF/$e/conformance/runner.ts; set METERLESS_ENGINES to a Meterless engines/ directory" >&2
    exit 4
  fi
done

run_engine() {
  local engine="$1"
  local var="$2"
  local impl="$3"
  local runner="$4"
  local card="$HERE/SCORECARD_${engine}.txt"
  local status=0
  env "${var}=${impl}" node "$runner" >"$card" 2>&1 || status=$?
  # The runner prints the implementation's absolute path; the scorecard is public, so it keeps the path relative to
  # the repository and never names the machine or the user that ran it.
  sed -i.bak "s#${ROOT}/##g" "$card" && rm -f "$card.bak"
  printf '%s\n' "----- ${engine} exit ${status} -----"
  cat "$card"
  return "$status"
}

hstatus=0
run_engine hmem HMEM_IMPL \
  "$HERE/hmem/index.ts" \
  "$REF/hmem/conformance/runner.ts" || hstatus=$?

wstatus=0
run_engine world-model WORLD_MODEL_IMPL \
  "$HERE/world-model/index.ts" \
  "$REF/world-model/conformance/runner.ts" || wstatus=$?

# ADR-0048 point 3: the Markovian engine is checked by Meterless's own runner, like the other two.
mstatus=0
run_engine markovian MARKOVIAN_IMPL \
  "$HERE/markovian/index.ts" \
  "$REF/markovian/conformance/runner.ts" || mstatus=$?

# The worst status wins, in the order the header names: a shim that does not load (2) before an infrastructure
# failure (above 2, the highest first) before a failed check (1).
worst=0
for s in "$hstatus" "$wstatus" "$mstatus"; do
  if [[ "$s" -eq 2 ]]; then
    exit 2
  fi
  if [[ "$s" -gt 2 && "$s" -gt "$worst" ]]; then
    worst="$s"
  fi
done
if [[ "$worst" -gt 2 ]]; then
  exit "$worst"
fi
if [[ "$hstatus" -ne 0 || "$wstatus" -ne 0 || "$mstatus" -ne 0 ]]; then
  exit 1
fi
exit 0
