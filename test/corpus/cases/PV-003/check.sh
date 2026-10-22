#!/usr/bin/env bash
# PV-003 check: run every contract input against base, change and the reference.
# Run from the repository root: bash tests/corpus/cases/PV-003/check.sh
set -euo pipefail
exec go run ./tests/corpus/cases/PV-003/check
