#!/bin/sh
# Tests the Action without GitHub: action.yml against ADR-0020 §3, and
# entrypoint.sh against the sample fixture with a temporary GITHUB_OUTPUT and
# GITHUB_WORKSPACE. Run from anywhere:  sh action/test.sh
# PROVEN_BIN names a built binary; without it the script builds one.
set -u

here=$(cd "$(dirname "$0")" && pwd)
root=$(dirname "$here")
fixture=$root/testdata/fixture/discount
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

if [ -z "${PROVEN_BIN:-}" ]; then
  (cd "$root" && go build -o "$tmp/proven" ./cmd/proven) || { echo "FAIL build: go build ./cmd/proven" >&2; exit 2; }
  PROVEN_BIN=$tmp/proven
fi
export PROVEN_BIN

pass=0
failed=0
check() {
  if [ "$2" = "$3" ]; then
    pass=$((pass + 1))
    echo "ok   $1"
  else
    failed=$((failed + 1))
    echo "FAIL $1: want '$2', got '$3'"
  fi
}

# --- action.yml: the names of ADR-0020 §3, and nothing else -------------------
names() { # section file: the keys one indent under a top-level section
  awk -v s="$1:" '
    /^[a-z]/ { in_s = ($0 == s) ; next }
    in_s && /^  [a-z-]+:/ { sub(/^  /, ""); sub(/:.*/, ""); print }
  ' "$2" | sort | tr '\n' ' '
}
check "action.yml inputs" \
  "baseline-commit comment config exclude github-token include output-format sample-mode severity strict target timeout " \
  "$(names inputs "$root/action.yml")"
check "action.yml outputs" \
  "findings-count json-file results-file sarif-file unresolved-count " \
  "$(names outputs "$root/action.yml")"
check "action.yml runs docker" "1" "$(grep -c '^  using: docker$' "$root/action.yml")"
check "action.yml strict default" "1" "$(awk '/^  strict:/ {f=1} f && /default: "true"/ {print; exit}' "$root/action.yml" | wc -l | tr -d ' ')"

# --- entrypoint.sh ------------------------------------------------------------
# run NAME VAR=value...: a fresh workspace, a fresh GITHUB_OUTPUT; sets $rc, $ws, $out.
run() {
  name=$1
  shift
  ws=$tmp/$name/ws
  out=$tmp/$name/out
  mkdir -p "$ws"
  : >"$out"
  # RUNDIR is the script's working directory: the runner starts the container in the checkout.
  (
    cd "${RUNDIR:-.}" &&
      env -i PATH="$PATH" HOME="$tmp" TMPDIR="$tmp" PROVEN_BIN="$PROVEN_BIN" \
        GITHUB_WORKSPACE="$ws" GITHUB_OUTPUT="$out" \
        INPUT_TARGET="$fixture" INPUT_SAMPLE_MODE=true "$@" \
        sh "${ENTRYPOINT:-$here/entrypoint.sh}" >"$tmp/$name/stdout" 2>"$tmp/$name/stderr"
  )
  rc=$?
}
output() { sed -n "s/^$1=//p" "$out"; }

run default
check "default: exit code (strict, findings)" 1 "$rc"
check "default: results-file" proven-results.sarif "$(output results-file)"
check "default: json-file" proven-results.json "$(output json-file)"
check "default: sarif-file" proven-results.sarif "$(output sarif-file)"
check "default: findings-count" 4 "$(output findings-count)"
check "default: unresolved-count" 0 "$(output unresolved-count)"
check "default: json file written in the workspace" 1 "$(test -s "$ws/proven-results.json" && echo 1)"
check "default: sarif file is SARIF 2.1.0" 1 "$(grep -c '"version": *"2.1.0"' "$ws/proven-results.sarif")"
check "default: workspace holds only the results" "proven-results.json proven-results.sarif " "$(ls "$ws" | tr '\n' ' ')"
check "default: stdout is the text review" 1 "$(grep -c 'finding F1' "$tmp/default/stdout")"

run lenient INPUT_STRICT=false
check "strict=false: exit code" 0 "$rc"
check "strict=false: findings-count" 4 "$(output findings-count)"

run asjson INPUT_OUTPUT_FORMAT=json INPUT_STRICT=false
check "output-format=json: results-file" proven-results.json "$(output results-file)"
check "output-format=json: stdout is JSON" 1 "$(grep -c '"findings": \[' "$tmp/asjson/stdout")"

run assarif INPUT_OUTPUT_FORMAT=sarif INPUT_STRICT=false
check "output-format=sarif: stdout is SARIF" 1 "$(grep -c '"version": *"2.1.0"' "$tmp/assarif/stdout")"

run astext INPUT_OUTPUT_FORMAT=text INPUT_STRICT=false
check "output-format=text: results-file" proven-results.txt "$(output results-file)"
check "output-format=text: text file has the review" 1 "$(grep -c 'finding F1' "$ws/proven-results.txt")"

run severe INPUT_SEVERITY=ERROR INPUT_STRICT=false
check "severity=ERROR: every fixture finding is ERROR, so all 4 stay" 4 "$(output findings-count)"

# The fixture cannot show a flag whose effect it does not exercise, so a stub
# binary records the command line each input produces.
cat >"$tmp/stub.sh" <<'EOF'
#!/bin/sh
for a in "$@"; do printf '%s\n' "$a"; done >"$STUB_ARGS"
env | grep -e '=tok-' >"$STUB_ARGS.env" || true
for a in "$@"; do
  [ -z "${STUB_NOWRITE:-}" ] || continue
  case "$a" in --sarif-output=*) printf '{}\n' >"${a#--sarif-output=}" ;; esac
  case "$a" in --json-output=*) printf '{"findings": [\n{"unresolved": true},\n{"unresolved": false},\n{"unresolved": true}\n]}\n' >"${a#--json-output=}" ;; esac
done
exit "${STUB_RC:-0}"
EOF
chmod +x "$tmp/stub.sh"
stubrun() {
  name=$1
  shift
  STUB_ARGS=$tmp/$name.args
  export STUB_ARGS
  run "$name" PROVEN_BIN="$tmp/stub.sh" STUB_ARGS="$STUB_ARGS" "$@"
  args=$(tr '\n' ' ' <"$STUB_ARGS")
}
contains() { # 1 when the recorded command line has this substring anywhere
  case "$args" in
    *"$1"*) echo 1 ;;
    *) echo 0 ;;
  esac
}
hasarg() { case " $args" in *" $1 "*) echo 1 ;; *) echo 0 ;; esac }

stubrun mapped INPUT_CONFIG=p.yaml INPUT_SEVERITY=WARNING INPUT_INCLUDE='a/*,b/*' INPUT_EXCLUDE='c/*' \
  INPUT_TIMEOUT=30 INPUT_BASELINE_COMMIT=abc123 INPUT_STRICT=true
check "maps config" 1 "$(hasarg "--config=$ws/p.yaml")"
check "maps severity" 1 "$(hasarg --severity=WARNING)"
check "maps include (first)" 1 "$(hasarg '--include=a/*')"
check "maps include (second)" 1 "$(hasarg '--include=b/*')"
check "maps exclude" 1 "$(hasarg '--exclude=c/*')"
check "maps timeout" 1 "$(hasarg --timeout=30)"
check "maps baseline-commit" 1 "$(hasarg --baseline-commit=abc123)"
check "maps sample-mode=true" 1 "$(hasarg --sample)"
check "maps strict=true" 1 "$(hasarg --error)"
check "counts findings from the JSON" 3 "$(output findings-count)"
check "counts unresolved from the JSON" 2 "$(output unresolved-count)"
check "the target is the last argument" "$fixture" "$(tail -n 1 "$STUB_ARGS")"

stubrun unmapped INPUT_SAMPLE_MODE=false INPUT_STRICT=false
check "maps strict=false" 1 "$(hasarg --no-error)"
check "sample-mode=false adds no --sample" 0 "$(hasarg --sample)"
check "empty baseline adds no --baseline-commit" 0 "$(contains --baseline-commit)"
check "empty timeout adds no --timeout" 0 "$(contains --timeout)"

stubrun failed STUB_RC=2 INPUT_STRICT=true
check "Proven's exit code 2 is the step's" 2 "$rc"

run badformat INPUT_OUTPUT_FORMAT=xml
check "output-format=xml: refused, exit 2" 2 "$rc"
check "output-format=xml: nothing written to GITHUB_OUTPUT" 0 "$(wc -c <"$out" | tr -d ' ')"

run badbool INPUT_STRICT=maybe
check "strict=maybe: refused, exit 2" 2 "$rc"

run noconfig INPUT_CONFIG=does-not-exist.yaml
check "missing config: Proven's refusal is the step's exit code (7)" 7 "$rc"
check "missing config: counts are 0" "0 0" "$(output findings-count) $(output unresolved-count)"

# A glob is Proven's to match: a path in the working directory that fits it must not
# turn it into other words. The runner starts the container in the checkout.
mkdir -p "$tmp/cwd/vendor/a" "$tmp/cwd/vendor/b"
RUNDIR=$tmp/cwd stubrun globs INPUT_EXCLUDE='vendor/*' INPUT_INCLUDE='*,vendor/*'
check "exclude glob reaches Proven verbatim" 1 "$(hasarg '--exclude=vendor/*')"
check "exclude glob is not expanded against the working directory" 0 "$(contains '--exclude=vendor/a')"
check "include glob '*' reaches Proven verbatim" 1 "$(hasarg '--include=*')"

# comment: the Action writes ONE comment per pull request, its own summary (post-comment.sh,
# tested below); reviewdog is never run, so no line comment can sit beside it. The stub
# reviewdog below is the instrument for that: it records any call, and every comment run
# asserts that no record exists (the owner's order of 2026-10-25; negative control: run
# the old entrypoint, which calls it with -reporter=github-pr-review, and these turn red).
cat >"$tmp/reviewdog.sh" <<'EOF'
#!/bin/sh
for a in "$@"; do printf '%s\n' "$a"; done >"$RD_ARGS"
cat >"$RD_STDIN"
printf '%s|%s\n' "${REVIEWDOG_GITHUB_API_TOKEN:-}" "${REVIEWDOG_GITEA_API_TOKEN:-}" >"$RD_TOKEN"
printf '%s|%s\n' "${GITHUB_API_URL:-}" "${GITEA_ADDRESS:-}" >"$RD_TOKEN.api"
exit "${RD_RC:-0}"
EOF
chmod +x "$tmp/reviewdog.sh"
rdrun() {
  name=$1
  shift
  RD_ARGS=$tmp/$name.rdargs
  RD_STDIN=$tmp/$name.rdstdin
  RD_TOKEN=$tmp/$name.rdtoken
  rm -f "$RD_ARGS" "$RD_STDIN" "$RD_TOKEN"
  run "$name" REVIEWDOG_BIN="$tmp/reviewdog.sh" RD_ARGS="$RD_ARGS" RD_STDIN="$RD_STDIN" RD_TOKEN="$RD_TOKEN" "$@"
  args=
  [ ! -f "$RD_ARGS" ] || args=$(tr '\n' ' ' <"$RD_ARGS")
}

rdrun notoken INPUT_COMMENT=true INPUT_GITHUB_TOKEN=
check "comment=true, no token: refused, exit 2" 2 "$rc"
check "comment=true, no token: names github-token" 1 "$(grep -c 'github-token is empty' "$tmp/notoken/stderr")"
check "comment=true, no token: Proven did not run" 0 "$(test -e "$tmp/notoken/ws/proven-results.json" && echo 1 || echo 0)"

rdrun nocomment INPUT_GITHUB_TOKEN=tok-1 INPUT_STRICT=false
check "comment=false: reviewdog not run" 0 "$(test -e "$RD_ARGS" && echo 1 || echo 0)"

rdrun nocommentnotoken INPUT_GITHUB_TOKEN= INPUT_STRICT=false
check "comment=false: no token needed" 0 "$rc"

rdrun nofail INPUT_COMMENT=true INPUT_GITHUB_TOKEN=tok-1 INPUT_CONFIG=does-not-exist.yaml
check "comment=true, Proven refused (7): nothing posted, exit 7" "7 0" "$rc $(test -e "$RD_ARGS" && echo 1 || echo 0)"

# No credential reaches the scan, with comment true or false (negative control:
# drop the unset and the env -u in entrypoint.sh and these two turn red).
for c in true false; do
  STUB_ARGS=$tmp/scrub-$c.args
  run "scrub-$c" PROVEN_BIN="$tmp/stub.sh" STUB_ARGS="$STUB_ARGS" REVIEWDOG_BIN="$tmp/reviewdog.sh" \
    RD_ARGS=$tmp/scrub-$c.rdargs RD_STDIN=$tmp/scrub-$c.rdstdin RD_TOKEN=$tmp/scrub-$c.rdtoken \
    INPUT_COMMENT=$c INPUT_STRICT=false INPUT_GITHUB_TOKEN=tok-1 INPUT_GITHUB-TOKEN=tok-2 \
    GITHUB_TOKEN=tok-3 ACTIONS_RUNTIME_TOKEN=tok-4 ACTIONS_ID_TOKEN_REQUEST_TOKEN=tok-5
  check "comment=$c: the scan runs with no token in its environment" "" "$(cat "$STUB_ARGS.env")"
done

# A result left by an earlier run is removed before Proven runs, so a review that
# could not run (2) is never counted from a stale file.
mkdir -p "$tmp/stale/ws"; printf '{"unresolved": true}\n' >"$tmp/stale/ws/proven-results.json"
stubrun stale STUB_RC=2 STUB_NOWRITE=1
check "a stale JSON review is not counted" 0 "$(output findings-count)"

# comment: the summary beside reviewdog's line comments is ONE comment, edited in place
# (post-comment.sh). The stub server answers GET /user and GET comments from $STUB, answers
# PATCH and POST with an id, and logs every request; a body is kept as $STUB/body.N.
cat >"$tmp/http.sh" <<'EOF'
#!/bin/sh
method=$1
url=$2
n=$(wc -l <"$STUB/log" | tr -d ' ')
printf '%s %s\n' "$method" "$url" >>"$STUB/log"
[ -z "${3:-}" ] || cp "$3" "$STUB/body.$((n + 1))"
[ "${HTTP_RC:-0}" -eq 0 ] || exit "$HTTP_RC"
case "$method $url" in
  "GET "*/comments*)
    # Like a real server: comments.json is page 1; any later page is empty.
    page=${url##*page=}
    if [ "$page" = 1 ] && [ -f "$STUB/comments.json" ]; then cat "$STUB/comments.json"; else printf '[]'; fi
    ;;
  PATCH* | POST*) printf '{"id":77}' ;;
esac
EOF
chmod +x "$tmp/http.sh"
printf '{"action":"synchronize","number":7,"pull_request":{"number":7,"head":{"sha":"abc123"},"base":{"sha":"def456"}}}\n' >"$tmp/event.json"
summary() { # NAME, then VAR=value...: a comment run as the account proven-bot on pull request 7
  name=$1
  shift
  STUB=$tmp/$name.stub
  mkdir -p "$STUB"
  : >"$STUB/log"
  rdrun "$name" INPUT_COMMENT=true INPUT_GITHUB_TOKEN=tok-1 INPUT_STRICT=false \
    GITHUB_API_URL=http://api.example/api/v1 GITHUB_REPOSITORY=o/r GITHUB_EVENT_PATH="$tmp/event.json" GITHUB_RUN_NUMBER=5 \
    COMMENT_HTTP="$tmp/http.sh" STUB="$STUB" PROVEN_COMMENT_ACCOUNT=proven-bot "$@"
}
mine='{"id":31,"user":{"login":"proven-bot"},"body":"<!-- proven-review run=4 sha=old -->\\nold"}'
theirs='{"id":41,"user":{"login":"someone-else"},"body":"<!-- proven-review run=4 sha=old -->\\nold"}'

summary sumnew
check "summary: exit 0" 0 "$rc"
check "summary: no earlier comment, one comment is created" "GET http://api.example/api/v1/repos/o/r/issues/7/comments?limit=50&per_page=50&page=1
POST http://api.example/api/v1/repos/o/r/issues/7/comments" "$(cat "$STUB/log")"
check "summary: one run with findings writes exactly one comment-writing call" 1 "$(grep -c -e '^POST ' -e '^PATCH ' "$STUB/log")"
check "summary: reviewdog is not run, so no line comment sits beside the summary" 0 "$(test -e "$RD_ARGS" && echo 1 || echo 0)"
check "summary: the marker is the first line, with this run and the head commit" '{"body":"<!-- proven-review run=5 sha=abc123 -->' "$(head -c 48 "$STUB/body.2")"
check "summary: the body is Proven's Markdown" 1 "$(grep -c '\*\*Proven: blocked\*\*' "$STUB/body.2")"
check "summary: the Markdown is not left in the workspace" "proven-results.json proven-results.sarif " "$(ls "$ws" | tr '\n' ' ')"

mkdir -p "$tmp/sumedit.stub" && printf '[%s]' "$mine" >"$tmp/sumedit.stub/comments.json"
summary sumedit
check "summary: an earlier comment of the account is edited, none is added" "GET http://api.example/api/v1/repos/o/r/issues/7/comments?limit=50&per_page=50&page=1
PATCH http://api.example/api/v1/repos/o/r/issues/comments/31" "$(cat "$STUB/log")"

mkdir -p "$tmp/sumother.stub" && printf '[%s]' "$theirs" >"$tmp/sumother.stub/comments.json"
summary sumother
check "summary: another account's marker comment is not edited" 0 "$(grep -c PATCH "$STUB/log")"
check "summary: another account's marker comment, a new one is created" 1 "$(grep -c '^POST ' "$STUB/log")"

summary sumstrict INPUT_STRICT=true
check "summary, strict, findings: still posted once, exit is Proven's 1" "1 1" "$(grep -c -e '^POST ' -e '^PATCH ' "$STUB/log") $rc"

summary sumgitea GITEA_ACTIONS=true GITHUB_SERVER_URL=https://gitea.example
check "summary on Gitea: one comment-writing call, reviewdog not run" "1 0" "$(grep -c -e '^POST ' -e '^PATCH ' "$STUB/log") $(test -e "$RD_ARGS" && echo 1 || echo 0)"

summary sumfail HTTP_RC=22
check "summary: a failed post fails the step (2)" 2 "$rc"
check "summary: the failure names the summary comment" 1 "$(grep -c 'posting the summary comment on the pull request failed' "$tmp/sumfail/stderr")"

rdrun sumnoevent INPUT_COMMENT=true INPUT_GITHUB_TOKEN=tok-1 INPUT_STRICT=false COMMENT_HTTP="$tmp/http.sh" STUB="$tmp/sumnoevent.stub"
check "summary: no pull request in the event, nothing is posted, exit 0" "0 0" "$rc $(test -e "$tmp/sumnoevent.stub/log" && echo 1 || echo 0)"
check "summary: no pull request, the notice says so" 1 "$(grep -c 'the summary comment was not posted' "$tmp/sumnoevent/stderr")"

rdrun sumoff INPUT_COMMENT=false INPUT_GITHUB_TOKEN=tok-1 INPUT_STRICT=false GITHUB_API_URL=http://api.example/api/v1 GITHUB_REPOSITORY=o/r GITHUB_EVENT_PATH="$tmp/event.json" COMMENT_HTTP="$tmp/http.sh" STUB="$tmp/sumoff.stub"
check "comment=false: no summary is posted" 0 "$(test -e "$tmp/sumoff.stub/log" && echo 1 || echo 0)"

# PROVEN_NETWORK=local (ADR-0034): the API address the summary is posted to is judged by the
# same rule as every other address, before the scan runs, by Proven's own reading of the mode
# (`proven check-address` prints it). An outside address, or a name the HTTP client would
# resolve with its own resolver, is refused by name: exit 7, nothing scanned, nothing posted.
# An allowed one is posted to with curl only, with no configuration file and no proxy.
# A stub curl on PATH records its arguments and answers like the forge.
mkdir -p "$tmp/fakebin"
cat >"$tmp/fakebin/curl" <<'EOF2'
#!/bin/sh
printf '%s\n' "$*" >>"$STUB/curl"
case "$*" in
  *" GET "*comments*) printf '[]' ;;
  *" POST "* | *" PATCH "*) printf '{"id":77}' ;;
  *) printf '{}' ;;
esac
EOF2
chmod +x "$tmp/fakebin/curl"
localpost() { # NAME, then VAR=value...: a comment run on pull request 7 with the stub curl
  name=$1
  shift
  STUB=$tmp/$name.stub
  mkdir -p "$STUB"
  : >"$STUB/curl"
  rdrun "$name" INPUT_COMMENT=true INPUT_GITHUB_TOKEN=tok-1 INPUT_STRICT=false GITHUB_REPOSITORY=o/r \
    GITHUB_EVENT_PATH="$tmp/event.json" GITHUB_RUN_NUMBER=5 PROVEN_COMMENT_ACCOUNT=proven-bot \
    PATH="$tmp/fakebin:$PATH" STUB="$STUB" "$@"
}
localrefused() { # NAME: exit 7, the refusal names GITHUB_API_URL, nothing posted or scanned
  check "$1: exit 7" 7 "$rc"
  check "$1: names GITHUB_API_URL" 1 "$(grep -c GITHUB_API_URL "$tmp/$1/stderr")"
  check "$1: nothing posted, nothing scanned" "0 0" "$(wc -l <"$STUB/curl" | tr -d ' ') $(test -e "$ws/proven-results.json" && echo 1 || echo 0)"
}
# Mutated field: GITHUB_API_URL (outside, then a name).
localpost localoutside PROVEN_NETWORK=local GITHUB_API_URL=https://203.0.113.7/api/v3
localrefused localoutside
localpost localname PROVEN_NETWORK=local GITHUB_API_URL=https://ghe.example/api/v3
localrefused localname
# Mutated field: PROVEN_NETWORK (local; local with a trailing carriage return; unset).
localpost localallowed PROVEN_NETWORK=local GITHUB_API_URL=http://10.1.2.3/api/v1 HTTPS_PROXY=http://203.0.113.9:3128
check "local, private address: exit 0" 0 "$rc"
check "local, private address: one comment-writing call" 1 "$(grep -c -e ' POST ' -e ' PATCH ' "$STUB/curl")"
check "local: every curl call has no configuration file and no proxy" "$(wc -l <"$STUB/curl" | tr -d ' ')" "$(grep -c "^-q --noproxy \* " "$STUB/curl")"
cr=$(printf '\r')
localpost localcr "PROVEN_NETWORK=local$cr" GITHUB_API_URL=http://10.1.2.3/api/v1
check "local with a carriage return: the comment is posted" 1 "$(grep -c -e ' POST ' -e ' PATCH ' "$STUB/curl")"
check "local with a carriage return: curl has no proxy either" "$(wc -l <"$STUB/curl" | tr -d ' ')" "$(grep -c "^-q --noproxy \* " "$STUB/curl")"
# Local mode with no curl in the image: the comment is not posted, the step says so, and
# neither wget nor an install is tried (apk add would reach a package mirror). PATH is a
# directory of links to every program on this host's PATH except curl, wget and apk, with
# stubs for wget and apk that record any call. Mutated field: PATH (curl absent).
mkdir -p "$tmp/nocurl"
oldifs=$IFS
IFS=:
for d in $PATH; do
  IFS=$oldifs
  [ -d "$d" ] || continue
  for f in "$d"/*; do
    b=${f##*/}
    case "$b" in curl | wget | apk) continue ;; esac
    [ -e "$tmp/nocurl/$b" ] || ln -s "$f" "$tmp/nocurl/$b" 2>/dev/null || true
  done
done
IFS=$oldifs
for b in wget apk; do
  printf '#!/bin/sh\nprintf "%%s %%s\\n" %s "$*" >>"$STUB/other"\nexit 1\n' "$b" >"$tmp/nocurl/$b"
  chmod +x "$tmp/nocurl/$b"
done
localpost localnocurl PROVEN_NETWORK=local GITHUB_API_URL=http://10.1.2.3/api/v1 PATH="$tmp/nocurl"
check "local, no curl: the post fails the step (2)" 2 "$rc"
check "local, no curl: the step says curl is needed" 1 "$(grep -c 'with curl only' "$tmp/localnocurl/stderr")"
check "local, no curl: neither wget nor apk is run" 0 "$(test -s "$STUB/other" && echo 1 || echo 0)"
localpost publiccurl GITHUB_API_URL=http://10.1.2.3/api/v1
check "public mode: curl is run as before" 0 "$(grep -c -- "--noproxy" "$STUB/curl")"
check "public mode: the comment is posted" 1 "$(grep -c -e ' POST ' -e ' PATCH ' "$STUB/curl")"

# pc_json: a backslash in the review (Markdown escapes a version as 1\.2.3) is doubled, so the
# body is valid JSON on every awk. BusyBox awk turned gsub's "\\\\" into ONE backslash and the
# lab's server refused the comment with 422.
printf '1\\.2.3 "q"\tx\\\n' >"$tmp/pc-json.md"
check "pc_json: backslash, quote and tab escaped" '{"body":"1\\.2.3 \"q\"\tx\\\n"}' \
  "$(sh -c '. "$1"; pc_json "$2"' sh "$here/post-comment.sh" "$tmp/pc-json.md")"

echo "passed $pass, failed $failed"
[ "$failed" -eq 0 ]
