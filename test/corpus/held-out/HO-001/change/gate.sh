#!/bin/sh
# Merge gate. Argument: the directory that holds the recorded state of one pull request.
# Passes when the bot has commented that the run passed.
set -eu

state=${1:?usage: gate.sh STATE_DIR}
bot=ci-bot

# comments.txt: one comment per line, oldest first: <author><TAB><body>
tab=$(printf '\t')
if grep -qx "$bot${tab}run passed" "$state/comments.txt"; then
  echo pass
  exit 0
fi
echo fail
exit 1
