#!/usr/bin/env bash
# PV-018 check. Run from any directory with: bash path/to/PV-018/check.sh
set -u

case_dir="$(cd "$(dirname "$0")" && pwd)" || exit 2
scratch="$(mktemp -d)" || exit 2
trap 'rm -rf "$scratch"' EXIT
cd "$case_dir" || exit 2
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off

if ! go run ./check >"$scratch/inputs" 2>"$scratch/input-errors"; then
  cat "$scratch/input-errors" >&2
  exit 2
fi
printf '%s\n' \
  'input=shipped base=pass change=pass mutant=pass reference=pass' \
  'input=planted-citing-row-outside-the-three-pages base=pass change=fail(changelog#0:160>120) mutant=pass reference=fail(changelog#0:160>120)' \
  'input=plain-row-between-budgets-on-a-listed-page base=pass change=pass mutant=fail(install#2:160>120) reference=pass' \
  >"$scratch/expected"
cat "$scratch/inputs"
if ! diff -u "$scratch/expected" "$scratch/inputs" >&2; then
  exit 1
fi

# The shipped tests, unchanged, against change. Keep diagnostics until checked.
if go test -count=1 -v ./change/ >"$scratch/change.log" 2>&1; then
  change_exit=0
else
  change_exit=1
fi

# The same test file against a copy with only the validator replaced.
mkdir -p "$scratch/module" || exit 2
cp -R go.mod data change "$scratch/module/" || exit 2
cp check/mutant/validator.go "$scratch/module/change/validator.go" || exit 2
if (cd "$scratch/module" && go test -count=1 -v ./change/) >"$scratch/mutant.log" 2>&1; then
  mutant_exit=0
else
  mutant_exit=1
fi

change_pass=$(awk '/^--- PASS: / { n++ } END { print n+0 }' "$scratch/change.log")
mutant_pass=$(awk '/^--- PASS: / { n++ } END { print n+0 }' "$scratch/mutant.log")
change_fail=$(awk '/^--- FAIL: / { n++ } END { print n+0 }' "$scratch/change.log")
mutant_fail=$(awk '/^--- FAIL: / { n++ } END { print n+0 }' "$scratch/mutant.log")
result=0
if [ "$change_exit" -eq 0 ] && [ "$change_pass" -eq 4 ] && [ "$change_fail" -eq 0 ]; then
  echo 'suite change=pass(4/4)'
else
  echo "suite change=fail($change_pass/4)"
  cat "$scratch/change.log" >&2
  result=1
fi
if [ "$mutant_exit" -eq 0 ] && [ "$mutant_pass" -eq 4 ] && [ "$mutant_fail" -eq 0 ]; then
  echo 'suite mutant=pass(4/4)'
else
  echo "suite mutant=fail($mutant_pass/4)"
  cat "$scratch/mutant.log" >&2
  result=1
fi
exit "$result"
