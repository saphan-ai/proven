#!/bin/sh
# Command-line tool. Arguments: MODEL PROMPT. Prints the answer, then exits.
# The service is the program named by SERVICE_CMD.
# When the model is not in the service's list, serves the default model.
set -eu

default_model=model-a

model=${1:?usage: tool.sh MODEL PROMPT}
prompt=${2:?usage: tool.sh MODEL PROMPT}
service=${SERVICE_CMD:?SERVICE_CMD is not set}

if ! sh "$service" list | grep -qx "$model"; then
  model=$default_model
fi

sh "$service" answer "$model"
