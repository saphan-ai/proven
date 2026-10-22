#!/usr/bin/env bash
# PV-019 check: runs every input against base and change, prints one line each.
# Run from the repository root: bash tests/corpus/cases/PV-019/check.sh
set -eu
cd "$(dirname "$0")"
go run ./check
