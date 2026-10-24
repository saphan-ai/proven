#!/bin/sh
# Runs the base script and the change script against the stand-in API on three
# inputs and prints one line per input: input name, base result, change result,
# last step reached (by the change script), reference label. A result is "ok"
# or the failing step. The reference label is a literal here; no script reads it.
set -eu

case_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo=$(CDPATH= cd -- "$case_dir/../../../.." && pwd)
scratch="$repo/.saphan/work/corpus-ho-004"
mkdir -p "$scratch"
work=$(mktemp -d "$scratch/run.XXXXXX")
trap 'rm -rf "$work"' EXIT

# run_setup VERSION INPUT USER EMAIL EXISTING: a fresh API state for each run;
# EXISTING is a user the API already holds, or "-" for none.
# Prints the script's word and, on a second line, its last step.
run_setup() {
  version=$1 input=$2 user=$3 email=$4 existing=$5
  state="$work/$version-$input"
  mkdir -p "$state"
  : > "$state/users.txt"
  if [ "$existing" != "-" ]; then
    printf '%s\n' "$existing" > "$state/users.txt"
  fi
  word=$(API_STATE="$state" API_CMD="$case_dir/api.sh" \
    sh "$case_dir/$version/setup.sh" "$user" "$email" 2> "$state/trace.txt") || true
  printf '%s\n%s\n' "$word" "$(tail -n 1 "$state/trace.txt" | sed 's/^step //')"
}

for entry in \
  "new-user-good-email:alice:alice@example.org:-:should succeed" \
  "user-exists:bob:bob@example.org:bob:should succeed" \
  "new-user-refused-email:carol:carol@blocked.example:-:should fail at create"
do
  input=${entry%%:*}
  rest=${entry#*:}
  user=${rest%%:*}
  rest=${rest#*:}
  email=${rest%%:*}
  rest=${rest#*:}
  existing=${rest%%:*}
  ref=${rest#*:}
  base=$(run_setup base "$input" "$user" "$email" "$existing" | sed -n 1p)
  change=$(run_setup change "$input" "$user" "$email" "$existing" | sed -n 1p)
  last=$(run_setup change "$input" "$user" "$email" "$existing" | sed -n 2p)
  printf '%s base %s change %s last step %s reference %s\n' "$input" "$base" "$change" "$last" "$ref"
done
