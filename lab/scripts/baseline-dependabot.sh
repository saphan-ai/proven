#!/bin/sh
# Run Proven's dependency check over the repository at the commit the Dependabot
# alerts are pinned to, and write the alert-by-alert comparison beside alerts.json.
#   baseline-dependabot.sh           ask the evidence store PROVEN_EVIDENCE_STORE names
#   baseline-dependabot.sh manifests read the manifests only and ask no store
# With the local store these set the snapshot the verdict lines name:
#   BASELINE_EVIDENCE_LABEL, BASELINE_SNAPSHOT_DATE (YYYY-MM-DD), BASELINE_ADVISORIES
# BASELINE_ADVISORIES is a directory of OSV advisories the local file is built from.
set -eu
. "$(CDPATH= cd -- "$(dirname "$0")" && pwd)/lib.sh"
cd "$ROOT"
mode=${1:-evidence}
case $mode in
  manifests)
    set -- -manifest-only
    ;;
  evidence)
    set --
    if [ -n "${BASELINE_EVIDENCE_LABEL:-}" ]; then
      set -- "$@" -evidence-label "$BASELINE_EVIDENCE_LABEL"
    fi
    if [ -n "${BASELINE_SNAPSHOT_DATE:-}" ]; then
      set -- "$@" -snapshot-date "$BASELINE_SNAPSHOT_DATE"
    fi
    if [ -n "${BASELINE_ADVISORIES:-}" ]; then
      set -- "$@" -advisories "$BASELINE_ADVISORIES"
    fi
    ;;
  *)
    echo "baseline-dependabot.sh: unknown mode $mode; use evidence or manifests" >&2
    exit 2
    ;;
esac
exec go run ./cmd/dependabot-baseline "$@"
