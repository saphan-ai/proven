#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
(cd "$work" && PROVEN_LOG_DIR="$work/logs" "$root/bin/proven" scan --sample "$root/testdata/fixture/discount")
