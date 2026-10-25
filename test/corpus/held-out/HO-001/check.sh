#!/bin/sh
# Runs the base gate and the change gate on four recorded states and prints one
# line per state: state name, base result, change result, reference label.
# The reference label is written here as a literal. No gate reads it.
set -eu

case_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
tab=$(printf '\t')

# mkstate NAME RUNS... : write runs.txt from the arguments, one run per argument.
# The comments are written by the caller into comments.txt.
work=$(mktemp -d "${TMPDIR:-/tmp}/corpus-ho-001.XXXXXX")
trap 'rm -rf "$work"' EXIT

mkstate() {
  name=$1
  shift
  mkdir -p "$work/$name"
  : > "$work/$name/runs.txt"
  : > "$work/$name/comments.txt"
  for run in "$@"; do
    printf '%s\n' "$run" >> "$work/$name/runs.txt"
  done
}

comment() {
  printf '%s%s%s\n' "$2" "$tab" "$3" >> "$work/$1/comments.txt"
}

mkstate one-run-success "1 aaa111 success"
comment one-run-success ci-bot "run passed"

mkstate one-run-failure "1 aaa111 failure"

mkstate success-then-failure "1 aaa111 success" "2 bbb222 failure"
comment success-then-failure ci-bot "run passed"

mkstate failure-then-success "1 aaa111 failure" "2 bbb222 success"
comment failure-then-success ci-bot "run passed"

run_gate() {
  # prints the gate's word; a failing exit is a result, not a script error
  sh "$1/gate.sh" "$2" || true
}

for entry in \
  "one-run-success:should pass" \
  "one-run-failure:should fail" \
  "success-then-failure:should fail" \
  "failure-then-success:should pass"
do
  name=${entry%%:*}
  ref=${entry#*:}
  b=$(run_gate "$case_dir/base" "$work/$name")
  c=$(run_gate "$case_dir/change" "$work/$name")
  printf '%s base %s change %s reference %s\n' "$name" "$b" "$c" "$ref"
done
