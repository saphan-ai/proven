#!/usr/bin/env bash
# PV-018 check. Run from the repository root: bash tests/corpus/cases/PV-018/check.sh
# Prints one line per input (base, change, mutant, reference) and one line per
# version for the shipped tests (change suite, mutant suite).
set -u

case_dir="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "$case_dir/../../../.." && pwd)"
scratch="$root/.saphan/work/corpus-pv-018/mutant-suite"

cd "$case_dir" || exit 2

go run ./check || exit 2

# The shipped tests, unchanged, against change.
if go test -count=1 ./change/ >/dev/null 2>&1; then change_suite=pass; else change_suite=fail; fi

# The same tests, byte for byte, against the mutant: a copy of the module whose
# change/validator.go is the mutant's.
rm -rf "$scratch" && mkdir -p "$scratch" || exit 2
cp -R go.mod data change "$scratch"/ || exit 2
cp check/mutant/validator.go "$scratch/change/validator.go" || exit 2
if (cd "$scratch" && go test -count=1 ./change/ >/dev/null 2>&1); then mutant_suite=pass; else mutant_suite=fail; fi

echo "suite change=$change_suite"
echo "suite mutant=$mutant_suite"
