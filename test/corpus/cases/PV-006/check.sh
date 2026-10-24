#!/usr/bin/env bash
# PV-006 check: runs every input of CASE.md §2 against base, change and the
# reference, and prints one line per input. Run from the repository root.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec go -C "$here" run ./check
