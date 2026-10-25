#!/bin/sh
# Command-line tool. Arguments: MODEL PROMPT. Prints the answer, then exits.
# The service is the program named by SERVICE_CMD.
# Exits 2 and prints an error when the model is not in the service's list.
set -eu

model=${1:?usage: tool.sh MODEL PROMPT}
prompt=${2:?usage: tool.sh MODEL PROMPT}
service=${SERVICE_CMD:?SERVICE_CMD is not set}

if ! sh "$service" list | grep -qx "$model"; then
  echo "tool: model not found: $model" >&2
  exit 2
fi

sh "$service" answer "$model"
