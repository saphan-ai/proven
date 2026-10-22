#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../../../.." && pwd)"
work="$root/.saphan/work/corpus-pv-012"
mkdir -p "$work"
go build -o "$work/base" "$root/tests/corpus/cases/PV-012/base/main.go"
go build -o "$work/change" "$root/tests/corpus/cases/PV-012/change/main.go"

foreign=(messages-1 edits-1 comments-1 uploads-1 audit-1)
result() {
    local output="$1" ids scope control count=0 item
    IFS='|' read -r ids scope control <<< "$output"
    for item in "${foreign[@]}"; do
        if [[ ",$ids," == *",$item,"* ]]; then ((count+=1)); fi
    done
    printf '%s (%s; control=%s)' "$count" "$scope" "$control"
}
for input in control vacuity application scope-no-team scope-with-team extra-sync control-no-registry; do
    reference=0
    if [[ "$input" == vacuity ]]; then reference=5; fi
    printf '%s | base: %s | change: %s | reference: %s\n' \
        "$input" "$(result "$($work/base "$input")")" "$(result "$($work/change "$input")")" "$reference"
done
