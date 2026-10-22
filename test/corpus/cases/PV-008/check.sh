#!/usr/bin/env bash
# PV-008 check: runs every input of CASE.md §2 against base, change and the
# reference. Run from the repository root: bash tests/corpus/cases/PV-008/check.sh
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"
exec go run ./check
