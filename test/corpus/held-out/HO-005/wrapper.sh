#!/bin/sh
# Batch wrapper, the same for both versions of the tool.
# Arguments: TOOL LOG JOBS. The tool is the program TOOL (run with sh).
# JOBS has one job per line: <job> <model> <prompt>.
# For each job whose tool run exits 0, appends "<job> <model>" to LOG:
# the job and the model it asked for.
set -eu

tool=${1:?usage: wrapper.sh TOOL LOG JOBS}
log=${2:?usage: wrapper.sh TOOL LOG JOBS}
jobs=${3:?usage: wrapper.sh TOOL LOG JOBS}

while read -r job model prompt; do
  if sh "$tool" "$model" "$prompt" > /dev/null 2>&1; then
    printf '%s %s\n' "$job" "$model" >> "$log"
  fi
done < "$jobs"
