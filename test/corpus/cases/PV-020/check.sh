#!/bin/bash
# Runs the PV-020 controls three ways, then one line per route for an
# oversize body, then the in-process controls with the limit removed.
# Observations are printed. A control result of failed or "could not run"
# is not a script error. A crash is.
set -euo pipefail

case_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo=$(git -C "$case_dir" rev-parse --show-toplevel)
work="$repo/.saphan/work/corpus-pv-020"
mkdir -p "$work"

export GOPROXY=off
export GOSUMDB=off
export GOFLAGS=-mod=readonly

go -C "$case_dir/base" build -o "$work/base-decide" ./cmd/decide
go -C "$case_dir/change" build -o "$work/change-decide" ./cmd/decide
go -C "$case_dir/change" build -o "$work/change-controls" ./cmd/controls

desc="$case_dir/description/api.json"
absent="$case_dir/description/not-provisioned.json"

emit() {
  local run=$1
  local group=$2
  shift 2
  local line
  line=$("$work/change-controls" "$@")
  printf 'run %s control %s result %s\n' "$run" "$group" "$line"
}

emit "a description-present-listener-allowed" "in-process" -group in-process
emit "a description-present-listener-allowed" "description" -group description -description "$desc"
emit "a description-present-listener-allowed" "listener" -group listener

emit "b description-absent" "in-process" -group in-process
emit "b description-absent" "description" -group description -description "$absent"
emit "b description-absent" "listener" -group listener

emit "c listener-refused" "in-process" -group in-process
emit "c listener-refused" "description" -group description -description "$desc"
PV020_SIMULATE_LISTEN_REFUSAL=1 emit "c listener-refused" "listener" -group listener

limit=$("$work/change-decide" -print-limit)
over=$((limit + 1))
"$work/base-decide" -bytes "$over" > "$work/base-routes.txt"
"$work/change-decide" -bytes "$over" > "$work/change-routes.txt"

exec 3< "$work/change-routes.txt"
while IFS= read -r bline; do
  if ! IFS= read -r cline <&3; then
    echo "change route list ended before base" >&2
    exit 1
  fi
  broute=${bline%%|*}
  brest=${bline#*|}
  bacc=${brest%%|*}
  breason=${brest#*|}
  croute=${cline%%|*}
  crest=${cline#*|}
  cacc=${crest%%|*}
  creason=${crest#*|}
  if [ "$broute" != "$croute" ]; then
    echo "route mismatch: $broute vs $croute" >&2
    exit 1
  fi
  printf 'route %s oversize base %s (%s) change %s (%s)\n' \
    "$broute" "$bacc" "$breason" "$cacc" "$creason"
done < "$work/base-routes.txt"
if IFS= read -r _ <&3; then
  echo "change route list is longer than base" >&2
  exit 1
fi
exec 3<&-

emit "limit-removed" "in-process" -group in-process -limit-removed
