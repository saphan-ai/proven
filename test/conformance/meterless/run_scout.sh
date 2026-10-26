#!/usr/bin/env bash
# Build meterless-conform and run Scout Intent's eval runner against the Go engine (ADR-0049 point 4), twice:
# Meterless's own runner on Meterless's own regression set, unchanged, and the copy under scout-intent/runner on
# Proven's change registry and change set. Writes SCORECARD_scout-intent.txt and SCORECARD_scout-intent-changes.txt
# beside this script, as the runners print them.
# Works from any directory. Exit 0 when both pass, 1 when a gate fails, 2 when the shim does not load, 3 when the
# helper does not build, 4 when no runner is found.
# METERLESS_ENGINES names a Meterless engines/ directory (required; it is not part of this repository).
# METERLESS_COMMIT, when set, is the reference commit the scorecard names; otherwise git is asked.
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

REF="${METERLESS_ENGINES:-}"
RUNNER="$REF/scout-intent/evals/runner/index.ts"
if [[ ! -f "$RUNNER" ]]; then
  echo "run_scout.sh: no Scout runner at $RUNNER; set METERLESS_ENGINES to a Meterless engines/ directory" >&2
  exit 4
fi
COMMIT="${METERLESS_COMMIT:-}"
if [[ -z "$COMMIT" ]]; then
  COMMIT="$(git -C "$REF" rev-parse HEAD 2>/dev/null || echo unknown)"
fi

# The scorecards are public, so a printed absolute path is cut to the repository-relative or reference-relative
# part, and never names the machine or the user that ran it.
scrub() {
  sed -i.bak -e "s#${REF}/#engines/#g" -e "s#${ROOT}/##g" -e "s#${WORK}/#<temp>/#g" "$1" && rm -f "$1.bak"
}

# 1. Meterless's runner and corpus, unchanged. It writes its report into the reference's own evals/reports, so that
# file is put back afterwards: the reference stays as it was provisioned.
CARD="$HERE/SCORECARD_scout-intent.txt"
KEEP="$WORK/latest.md"
[[ -f "$REF/scout-intent/evals/reports/latest.md" ]] && cp "$REF/scout-intent/evals/reports/latest.md" "$KEEP"
sstatus=0
{
  echo "Scout Intent eval: Meterless runner and regression set v1, unchanged"
  echo "reference: github.com/Meterless/Meterless at commit ${COMMIT}, engines/scout-intent"
  echo "implementation: test/conformance/meterless/scout-intent/index.ts -> internal/scout (Stage 1 only; Stage 2 not built)"
  echo
  SCOUT_IMPL="$HERE/scout-intent/index.ts" node "$RUNNER" all 2>&1 || sstatus=$?
  echo
  echo "exit ${sstatus}"
} >"$CARD"
[[ -f "$KEEP" ]] && cp "$KEEP" "$REF/scout-intent/evals/reports/latest.md"
scrub "$CARD"
printf '%s\n' "----- scout-intent exit ${sstatus} -----"
cat "$CARD"

# 2. Proven's change registry and change set, through the copy of the runner (scout-intent/runner/README.md names
# every line that differs). The registry is the file the review embeds.
CCARD="$HERE/SCORECARD_scout-intent-changes.txt"
cstatus=0
{
  echo "Scout Intent eval: Proven's change registry (internal/scout/changes.json) and change set v1"
  echo "runner: test/conformance/meterless/scout-intent/runner/index.ts, a copy of engines/scout-intent/evals/runner/index.ts at commit ${COMMIT}"
  echo
  SCOUT_IMPL="$HERE/scout-intent/index.ts" \
    SCOUT_REGISTRY="$ROOT/internal/scout/changes.json" \
    SCOUT_CORPUS="$HERE/scout-intent/changes/v1" \
    SCOUT_THRESHOLDS="$HERE/scout-intent/runner/thresholds.json" \
    SCOUT_REPORT="$WORK/changes-report.md" \
    node "$HERE/scout-intent/runner/index.ts" 2>&1 || cstatus=$?
  echo
  echo "exit ${cstatus}"
} >"$CCARD"
scrub "$CCARD"
printf '%s\n' "----- scout-intent-changes exit ${cstatus} -----"
cat "$CCARD"

if [[ "$sstatus" -eq 2 || "$cstatus" -eq 2 ]]; then
  exit 2
fi
if [[ "$sstatus" -ne 0 || "$cstatus" -ne 0 ]]; then
  exit 1
fi
exit 0
