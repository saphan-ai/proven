#!/bin/sh
# Runs the base gate and the change gate on four recorded states and prints one
# line per state: state name, base result, change result, reference label.
# The reference label is written here as a literal. No gate reads it.
# Each state records the bot account's name in bot.txt; the gates do not read it.
set -eu

case_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
work=$(mktemp -d "${TMPDIR:-/tmp}/corpus-ho-002.XXXXXX")
trap 'rm -rf "$work"' EXIT

tab=$(printf '\t')

# mkstate NAME RUN : write run.txt, bot.txt and an empty comments.txt
mkstate() {
  mkdir -p "$work/$1"
  printf '%s\n' "$2" > "$work/$1/run.txt"
  printf 'ci-bot\n' > "$work/$1/bot.txt"
  : > "$work/$1/comments.txt"
}

# comment NAME AUTHOR BODY
comment() {
  printf '%s%s%s\n' "$2" "$tab" "$3" >> "$work/$1/comments.txt"
}

mkstate bot-marker-current "7 c0ffee7 success"
comment bot-marker-current ci-bot "verified: run 7 commit c0ffee7"

mkstate no-marker "7 c0ffee7 failure"

mkstate person-marker-current "7 c0ffee7 failure"
comment person-marker-current reviewer "verified: run 7 commit c0ffee7"

mkstate person-marker-older "7 c0ffee7 success"
comment person-marker-older ci-bot "verified: run 7 commit c0ffee7"
comment person-marker-older reviewer "verified: run 6 commit badf00d"

run_gate() {
  # prints the gate's word; a failing exit is a result, not a script error
  sh "$1/gate.sh" "$2" || true
}

for entry in \
  "bot-marker-current:should pass" \
  "no-marker:should fail" \
  "person-marker-current:should fail" \
  "person-marker-older:should pass"
do
  name=${entry%%:*}
  ref=${entry#*:}
  b=$(run_gate "$case_dir/base" "$work/$name")
  c=$(run_gate "$case_dir/change" "$work/$name")
  printf '%s base %s change %s reference %s\n' "$name" "$b" "$c" "$ref"
done
