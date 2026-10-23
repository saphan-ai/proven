#!/usr/bin/env bash
# PV-010 check: runs every input against base, change and the reference and
# prints one line per input. Run from the repository root:
#   bash tests/corpus/cases/PV-010/check.sh
# No network, no model. Standard-library Go only.
set -u
export GOTOOLCHAIN=local

root=$(cd "$(dirname "$0")/../../../.." && pwd)
case_dir="$root/tests/corpus/cases/PV-010"
work="$root/.saphan/work/corpus-pv-010"
pinned=42

rm -rf "$work/bin" "$work/change-edited"
mkdir -p "$work/bin" "$work/change-edited"

echo "toolchain: $(go version)"

for v in base change reference; do
  (cd "$case_dir/$v" && go build -o "$work/bin/$v" .) || { echo "build failed: $v"; exit 1; }
done

# change-edited: change with only the description's width constant edited.
cp "$case_dir/change/"* "$work/change-edited/"
sed -e 's/^const statedMaxGap = 48$/const statedMaxGap = 256/' "$case_dir/change/guard.go" > "$work/change-edited/guard.go"
if cmp -s "$case_dir/change/guard.go" "$work/change-edited/guard.go"; then
  echo "change-edited: the width constant was not found"; exit 1
fi
(cd "$work/change-edited" && go build -o "$work/bin/change-edited" .) || { echo "build failed: change-edited"; exit 1; }

verdict() { # binary note -> pass | fail
  local out
  out=$("$1" check "$pinned" "$2" | head -1)
  echo "$out" | tr 'A-Z' 'a-z'
}

line() { # name note
  printf '%s | base=%s | change=%s | reference=%s\n' "$1" \
    "$(verdict "$work/bin/base" "$2")" \
    "$(verdict "$work/bin/change" "$2")" \
    "$("$work/bin/reference" note "$pinned" "$2")"
}

echo "pinned count: $pinned"
line "current"          "LANGUAGES 41 -> 42"
line "stale"            "LANGUAGES 41 -> 40"
line "twenty-digits"    "LANGUAGES 41 -> 99999999999999999999"
line "minus-999"        "LANGUAGES 41 -> -999"
line "plus-41"          "LANGUAGES 41 -> +41"
line "arabic-indic-41"  "LANGUAGES 41 -> $(printf '\xd9\xa4\xd9\xa1')"

stated() { # binary -> widths stated by the description
  local w
  w=$("$1" describe | grep -o 'up to [0-9]*' | grep -o '[0-9]*')
  if [ -z "$w" ]; then echo "none"; else echo "0..$w"; fi
}

(cd "$case_dir/change" && go test -count=1 ./... >"$work/change-tests.txt" 2>&1); change_tests=$?
(cd "$work/change-edited" && go test -count=1 ./... >"$work/change-edited-tests.txt" 2>&1); edited_tests=$?
res() { if [ "$1" -eq 0 ]; then echo pass; else echo fail; fi; }

printf 'description | stated widths: base=%s | change=%s | change-edited=%s | reference=%s | change tests on change=%s | change tests on change-edited=%s\n' \
  "$(stated "$work/bin/base")" \
  "$(stated "$work/bin/change")" \
  "$(stated "$work/bin/change-edited")" \
  "$("$work/bin/reference" widths "$work/bin/change" "$pinned")" \
  "$(res $change_tests)" "$(res $edited_tests)"
