#!/bin/sh
# Merge gate. Argument: the directory that holds the recorded state of one pull request.
# Passes only when a comment holds a marker line that names the current run and the current commit.
set -eu

state=${1:?usage: gate.sh STATE_DIR}

# run.txt: <id> <commit> <status>
run_id=$(cut -d ' ' -f 1 "$state/run.txt")
commit=$(cut -d ' ' -f 2 "$state/run.txt")

# comments.txt: one comment per line, oldest first: <author><TAB><body>
tab=$(printf '\t')
if grep -qx ".*${tab}verified: run ${run_id} commit ${commit}" "$state/comments.txt"; then
  echo pass
  exit 0
fi
echo fail
exit 1
