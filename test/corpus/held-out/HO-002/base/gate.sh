#!/bin/sh
# Merge gate. Argument: the directory that holds the recorded state of one pull request.
# Passes when a comment holds a marker line for any run.
set -eu

state=${1:?usage: gate.sh STATE_DIR}

# comments.txt: one comment per line, oldest first: <author><TAB><body>
if grep -Eq '	verified: run [0-9]+ commit [0-9a-f]+$' "$state/comments.txt"; then
  echo pass
  exit 0
fi
echo fail
exit 1
