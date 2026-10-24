#!/bin/bash
# Runs the data check on three trees and prints one line per tree: tree name,
# base result, change result, reference label. The reference label is a literal
# here; the script does not read it.
#   base           the base tree (data in the old directory)
#   change-dirty   a working copy of the change, with the old directory left
#                  over, untracked, from earlier work
#   change-clean   a fresh clone of the change: only what the change tracks
# The base column of the two change rows is "-": base is one tree, not a pair.
# A check that exits non-zero is a result; a crash of this script is not.
set -euo pipefail

case_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo=$(CDPATH= cd -- "$case_dir/../../../.." && pwd)
scratch="$repo/.saphan/work/corpus-ho-006"
mkdir -p "$scratch"
work=$(mktemp -d "$scratch/run.XXXXXX")
trap 'rm -rf "$work"' EXIT

# a repository of our own, whatever the surroundings say
export GIT_CONFIG_GLOBAL=/dev/null
export GIT_CONFIG_SYSTEM=/dev/null
export GIT_AUTHOR_NAME=case GIT_AUTHOR_EMAIL=case@example.invalid
export GIT_COMMITTER_NAME=case GIT_COMMITTER_EMAIL=case@example.invalid

# the base tree
cp -R "$case_dir/base" "$work/base-tree"

# the change, committed in a repository of its own
cp -R "$case_dir/change" "$work/change-dirty"
rm -f "$work/change-dirty/DESCRIPTION.md"
git -C "$work/change-dirty" init -q
git -C "$work/change-dirty" add -A
git -C "$work/change-dirty" -c core.hooksPath=/dev/null commit -q -m "change"

# the fresh clone holds only what the change tracks
git clone -q "$work/change-dirty" "$work/change-clean"

# the leftover: the old directory, untracked, in the dirty tree only
cp -R "$case_dir/base/data" "$work/change-dirty/data"

if [ -e "$work/change-clean/data" ]; then
  echo "the clean tree holds the old directory" >&2
  exit 1
fi
if [ "$(git -C "$work/change-dirty" status --porcelain)" != "?? data/" ]; then
  echo "the leftover is not exactly one untracked directory" >&2
  exit 1
fi

# the word is "ok" when the script exits 0, otherwise "fail"; its own message
# is not printed (it can name a temporary path)
run_check() {
  if sh "$1/verify.sh" > /dev/null 2>&1; then
    echo ok
  else
    echo fail
  fi
}

printf '%s base %s change %s reference %s\n' base "$(run_check "$work/base-tree")" - "should pass"
printf '%s base %s change %s reference %s\n' change-dirty - "$(run_check "$work/change-dirty")" "should fail"
printf '%s base %s change %s reference %s\n' change-clean - "$(run_check "$work/change-clean")" "should fail"
