#!/bin/sh
# Merge gate. Argument: the directory that holds the recorded state of one pull request.
# Passes only when the newest run of the workflow has status success.
set -eu

state=${1:?usage: gate.sh STATE_DIR}

# runs.txt: one run per line, oldest first: <id> <commit> <status>
status=$(tail -n 1 "$state/runs.txt" | cut -d ' ' -f 3)

if [ "$status" = "success" ]; then
  echo pass
  exit 0
fi
echo fail
exit 1
