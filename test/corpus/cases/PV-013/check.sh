#!/usr/bin/env bash
# PV-013 check. From the repository root:
#   bash tests/corpus/cases/PV-013/check.sh
# Runs the shipped control, then prints one line per contract input:
#   <input> | base: <result> | change: <result> | base_token_in_output: <yes|no> | change_token_in_output: <yes|no>
# base ships no control. change ships the no-leak control.
# Credentials file for both versions: <home>/credentials

set -u

case_dir=$(cd "$(dirname "$0")" && pwd)
token='tok_synthetic_0000_not_a_secret'
work_root="$case_dir/../../../../.saphan/work/corpus-pv-013"
mkdir -p "$work_root"
work=$(mktemp -d "$work_root/check.XXXXXX")
trap 'rm -rf "$work"' EXIT

export GOTOOLCHAIN=local
export GOPROXY=off
export GOSUMDB=off

problems=0

if ! ( cd "$case_dir/base" && go build -o "$work/basebin" . ) >"$work/build-base.txt" 2>&1; then
  cat "$work/build-base.txt" >&2
  exit 1
fi
if ! ( cd "$case_dir/change" && go build -o "$work/changebin" . ) >"$work/build-change.txt" 2>&1; then
  cat "$work/build-change.txt" >&2
  exit 1
fi

new_home() {
  mktemp -d "$work/h.XXXXXX"
}

oneline() {
  local first=1 line
  first=1
  while IFS= read -r line || [ -n "$line" ]; do
    if [ "$first" -eq 1 ]; then
      printf '%s' "$line"
      first=0
    else
      printf ' / %s' "$line"
    fi
  done <"$1"
}

report() {
  local name="$1" base_file="$2" change_file="$3" base_flag=no change_flag=no
  if grep -q -F "$token" "$base_file"; then base_flag=yes; fi
  if grep -q -F "$token" "$change_file"; then change_flag=yes; fi
  printf '%s | base: %s | change: %s | base_token_in_output: %s | change_token_in_output: %s\n' \
    "$name" "$(oneline "$base_file")" "$(oneline "$change_file")" "$base_flag" "$change_flag"
}

# The pasted-value case must prove both the refusal and the value in the
# change's own output. A base-side leak cannot satisfy either assertion.
assert_change_refusal_leaks() {
  local file="$1"
  grep -q -F 'refused: token_env' "$file" && grep -q -F "$token" "$file"
}

printf 'refused: token_env rejected\n' >"$work/negative-no-value.txt"
printf 'accepted: %s\n' "$token" >"$work/negative-no-refusal.txt"
printf '%s\n' "$token" >"$work/negative-base-leak.txt"
printf 'none\n' >"$work/base-control.txt"
if assert_change_refusal_leaks "$work/negative-no-value.txt"; then
  printf 'negative control failed: missing leaked value passed\n' >&2
  problems=1
fi
if assert_change_refusal_leaks "$work/negative-no-refusal.txt"; then
  printf 'negative control failed: missing refusal passed\n' >&2
  problems=1
fi
if [ "${1-}" = '--negative-control' ]; then
  report negative-base-only "$work/negative-base-leak.txt" "$work/negative-no-value.txt"
  assert_change_refusal_leaks "$work/negative-no-value.txt"
  exit $?
fi
if [ "${1-}" = '--negative-control=no-refusal' ]; then
  report negative-no-refusal "$work/base-control.txt" "$work/negative-no-refusal.txt"
  assert_change_refusal_leaks "$work/negative-no-refusal.txt"
  exit $?
fi

# Shipped controls. base has none.
home=$(new_home)
if "$work/changebin" --home "$home" control >"$work/change-control.txt" 2>"$work/control.err"; then
  :
else
  printf ' (exit %s)\n' "$?" >>"$work/change-control.txt"
  problems=1
fi
report control "$work/base-control.txt" "$work/change-control.txt"
if ! grep -q -F -x 'pass' "$work/change-control.txt"; then
  printf 'shipped control failed: change did not report pass\n' >&2
  problems=1
fi

run_set() {
  local bin="$1" field="$2" value="$3" out="$4" home rc
  home=$(new_home)
  if "$work/$bin" --home "$home" set "$field" "$value" >"$out" 2>"$work/step.err"; then
    :
  else
    rc=$?
    problems=1
    printf 'error exit %s\n' "$rc" >"$out"
    cat "$work/step.err" >>"$out"
  fi
}

run_set_show() {
  local bin="$1" field="$2" value="$3" out="$4" home rc
  home=$(new_home)
  if ! "$work/$bin" --home "$home" set "$field" "$value" >"$work/step1.txt" 2>"$work/step.err"; then
    rc=$?
    problems=1
    printf 'error exit %s\n' "$rc" >"$out"
    cat "$work/step.err" >>"$out"
    return
  fi
  if ! "$work/$bin" --home "$home" show >"$work/step2.txt" 2>"$work/step.err"; then
    rc=$?
    problems=1
    printf 'error exit %s\n' "$rc" >"$out"
    cat "$work/step.err" >>"$out"
    return
  fi
  cat "$work/step1.txt" "$work/step2.txt" >"$out"
}

run_cred_show() {
  local bin="$1" out="$2" home rc
  home=$(new_home)
  printf '%s\n' "$token" >"$home/credentials"
  if "$work/$bin" --home "$home" show >"$out" 2>"$work/step.err"; then
    :
  else
    rc=$?
    problems=1
    printf 'error exit %s\n' "$rc" >"$out"
    cat "$work/step.err" >>"$out"
  fi
}

run_set basebin token_env CHAT_API_TOKEN "$work/b-valid.txt"
run_set changebin token_env CHAT_API_TOKEN "$work/c-valid.txt"
report valid-name "$work/b-valid.txt" "$work/c-valid.txt"

run_cred_show basebin "$work/b-cred.txt"
run_cred_show changebin "$work/c-cred.txt"
report token-in-credentials-show "$work/b-cred.txt" "$work/c-cred.txt"

run_set basebin token_env chat-token "$work/b-bad-env.txt"
run_set changebin token_env chat-token "$work/c-bad-env.txt"
report bad-token-env "$work/b-bad-env.txt" "$work/c-bad-env.txt"
if ! grep -q -F 'refused: token_env' "$work/c-bad-env.txt"; then
  printf 'bad-token-env case failed: change did not refuse\n' >&2
  problems=1
fi

run_set basebin channel 'not a channel' "$work/b-bad-ch.txt"
run_set changebin channel 'not a channel' "$work/c-bad-ch.txt"
report bad-channel "$work/b-bad-ch.txt" "$work/c-bad-ch.txt"
if ! grep -q -F 'refused: channel' "$work/c-bad-ch.txt"; then
  printf 'bad-channel case failed: change did not refuse\n' >&2
  problems=1
fi

run_set basebin endpoint not-a-url "$work/b-bad-ep.txt"
run_set changebin endpoint not-a-url "$work/c-bad-ep.txt"
report bad-endpoint "$work/b-bad-ep.txt" "$work/c-bad-ep.txt"
if ! grep -q -F 'refused: endpoint' "$work/c-bad-ep.txt"; then
  printf 'bad-endpoint case failed: change did not refuse\n' >&2
  problems=1
fi

run_set_show basebin token_env "$token" "$work/b-paste.txt"
run_set_show changebin token_env "$token" "$work/c-paste.txt"
report token-in-name-then-show "$work/b-paste.txt" "$work/c-paste.txt"
if ! assert_change_refusal_leaks "$work/c-paste.txt"; then
  printf 'pasted-value case failed: change output lacks refusal or synthetic value\n' >&2
  problems=1
fi

exit "$problems"
