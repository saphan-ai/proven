#!/bin/bash
# Builds the base and the change, makes one source tree, copies it with each
# under umask 022 and under umask 002 (each run is its own process), and prints
# one line per entry and umask: entry, umask, base mode, change mode, reference
# mode. The reference is a literal here (files 0644, directories 0755); neither
# copy reads it. Then it runs the change's own test under each umask.
# A test that fails under one umask is an observation, not a script error.
set -euo pipefail

case_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo=$(CDPATH= cd -- "$case_dir/../../../.." && pwd)
scratch="$repo/.saphan/work/corpus-ho-003"
mkdir -p "$scratch"
work=$(mktemp -d "$scratch/run.XXXXXX")
trap 'rm -rf "$work"' EXIT

export GOPROXY=off
export GOSUMDB=off
export GOFLAGS=-mod=readonly

go -C "$case_dir/base" build -o "$work/base-snapcopy" ./cmd/snapcopy
go -C "$case_dir/change" build -o "$work/change-snapcopy" ./cmd/snapcopy

# The source tree: files 0600 and 0640, directories 0700. Chmod is explicit so
# the source modes do not depend on the umask of this process.
src="$work/src"
mkdir -p "$src/sub"
printf 'a\n' > "$src/a.txt"
printf 'b\n' > "$src/sub/b.txt"
chmod 0600 "$src/a.txt"
chmod 0640 "$src/sub/b.txt"
chmod 0700 "$src/sub" "$src"

# reference MODE of an entry: a literal, never computed by a copy.
reference() {
  case "$1" in
    .|sub) echo 0755 ;;
    a.txt|sub/b.txt) echo 0644 ;;
    *) echo "no reference for $1" >&2; exit 1 ;;
  esac
}

for mask in 022 002; do
  # each run is a separate process, so one umask does not leak into the other
  (umask "$mask"; "$work/base-snapcopy" "$src" "$work/base-dst-$mask") > "$work/base-$mask.txt"
  (umask "$mask"; "$work/change-snapcopy" "$src" "$work/change-dst-$mask") > "$work/change-$mask.txt"

  exec 3< "$work/change-$mask.txt"
  while IFS=' ' read -r entry bmode; do
    if ! IFS=' ' read -r centry cmode <&3 || [ "$entry" != "$centry" ]; then
      echo "entry lists differ at $entry" >&2
      exit 1
    fi
    printf '%s umask %s base %s change %s reference %s\n' \
      "$entry" "$mask" "$bmode" "$cmode" "$(reference "$entry")"
  done < "$work/base-$mask.txt"
  exec 3<&-
done

for mask in 022 002; do
  # -count=1: the test cache does not key on the umask
  if out=$(umask "$mask"; go -C "$case_dir/change" test -count=1 ./... 2>&1); then
    printf 'change test umask %s ok\n' "$mask"
  else
    n=$(printf '%s\n' "$out" | grep -c ': mode 0' || true)
    printf 'change test umask %s FAIL (%s entries with a wrong mode)\n' "$mask" "$n"
  fi
done
