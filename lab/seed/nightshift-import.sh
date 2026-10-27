#!/bin/sh
# Load the public receipts of Proven's own Nightshift calls into the lab's Loki, for the Nightshift dashboard.
#
#   nightshift-import.sh < receipts.public.jsonl
#
# Input: one JSON object per line, flat, in the lab's public receipt shape (docs/LAB.md, "The Nightshift receipts"):
#   ts_ms, app, status (ok or failed), model, provider, reason (a failed call's reason), verified (true or false),
#   latency_ms, input_tokens, output_tokens, request_id, receipt_id.
# What it keeps and what it refuses, in this order:
#   - only app "proven": Proven's own calls. A line of any other app (the fleet's review legs) is counted and not loaded;
#   - prompt_preview is dropped from every line before it is sent, and a line that still names it afterwards is refused;
#   - a line that is not a flat object with a numeric ts_ms and a status of ok or failed is refused and counted.
# The receipts are replayed: the newest one is stamped with the moment of the import and the rest keep their spacing, so the
# dashboard's last hours hold them whatever day the lab is started (Loki refuses samples older than a week). The line keeps
# its own ts_ms.
#
# It runs in the seed image over the lab network. It reads no secret and sends nothing outside the lab: LAB_LOKI_URL is the
# lab's own Loki. A set of receipts already loaded in the last hours is not loaded twice; FORCE=1 loads them again.
set -eu

LOKI_URL=${LAB_LOKI_URL:-}
# The most lines one import takes in: the public set is a handful.
IMPORT_MAX_LINES=2000
# How far back a loaded set counts as "already there", in seconds.
ALREADY_WINDOW=18000

fail() {
  echo "nightshift-import: $1" >&2
  exit 1
}

[ -n "$LOKI_URL" ] || fail "LAB_LOKI_URL is not set"
case $LOKI_URL in
  http://lgtm:* | http://127.0.0.1:* | http://localhost:*) ;;
  *) fail "LAB_LOKI_URL is not the lab's own Loki" ;;
esac

now_s=$(date +%s)
if [ "${FORCE:-}" != 1 ]; then
  start_ns="$((now_s - ALREADY_WINDOW))000000000"
  loaded=$(curl -sS --max-time 10 "$LOKI_URL/loki/api/v1/label/job/values?start=$start_ns" 2>/dev/null || true)
  case $loaded in
    *'"nightshift"'*)
      echo "nightshift-import: the receipts are already loaded (in the last hours); FORCE=1 loads them again"
      exit 0
      ;;
  esac
fi

payload=$(mktemp)
trap 'rm -f "$payload"' EXIT INT TERM
head -n "$IMPORT_MAX_LINES" | tr -d '\r' | awk -v now_ms="$((now_s * 1000))" '
  # jstr(s, key): the string value of "key" in the flat object s, "" when it is absent or not a string.
  function jstr(s, key,    i, rest, out, c, n) {
    i = index(s, "\"" key "\"")
    if (i == 0) return ""
    rest = substr(s, i + length(key) + 2)
    if (!match(rest, /^[ \t]*:[ \t]*"/)) return ""
    rest = substr(rest, RLENGTH + 1)
    out = ""
    n = length(rest)
    for (i = 1; i <= n; i++) {
      c = substr(rest, i, 1)
      if (c == "\\") { out = out c substr(rest, i + 1, 1); i++; continue }
      if (c == "\"") return out
      out = out c
    }
    return ""
  }
  # jnum(s, key): the number value of "key", "" when it is absent or not a number.
  function jnum(s, key,    i, rest) {
    i = index(s, "\"" key "\"")
    if (i == 0) return ""
    rest = substr(s, i + length(key) + 2)
    if (!match(rest, /^[ \t]*:[ \t]*-?[0-9]+(\.[0-9]+)?/)) return ""
    rest = substr(rest, 1, RLENGTH)
    sub(/^[ \t]*:[ \t]*/, "", rest)
    return rest
  }
  function jbool(s, key) {
    if (s ~ ("\"" key "\"[ \t]*:[ \t]*true")) return "true"
    if (s ~ ("\"" key "\"[ \t]*:[ \t]*false")) return "false"
    return ""
  }
  # label(v): a label value Loki takes without surprise.
  function label(v) {
    if (v == "") return "none"
    gsub(/[^A-Za-z0-9._\/:-]/, "_", v)
    return v
  }
  function esc(s) {
    gsub(/\\/, "\\\\", s)
    gsub(/"/, "\\\"", s)
    return s
  }
  NF == 0 { next }
  {
    line = $0
    if (line !~ /^[ \t]*\{.*\}[ \t]*$/) { refused++; next }
    # prompt_preview is dropped: its string value, its comma, from the line sent.
    gsub(/,[ \t]*"prompt_preview"[ \t]*:[ \t]*("([^"\\]|\\.)*"|null)/, "", line)
    gsub(/"prompt_preview"[ \t]*:[ \t]*("([^"\\]|\\.)*"|null)[ \t]*,?/, "", line)
    if (index(line, "prompt_preview") > 0) { refused++; next }
    if (jstr(line, "app") != "proven") { other++; next }
    st = jstr(line, "status")
    ts = jnum(line, "ts_ms")
    if ((st != "ok" && st != "failed") || ts == "") { refused++; next }
    n++
    TS[n] = ts + 0
    LINE[n] = line
    ST[n] = st
    MODEL[n] = label(jstr(line, "model"))
    PROV[n] = label(jstr(line, "provider"))
    REASON[n] = (st == "ok") ? "none" : label(jstr(line, "reason"))
    v = jbool(line, "verified")
    VER[n] = (v == "") ? "unknown" : v
    if (TS[n] > max) max = TS[n]
  }
  END {
    printf "nightshift-import: %d loaded, %d of other apps left out, %d refused\n", n, other, refused > "/dev/stderr"
    if (n == 0) exit 3
    shift = now_ms - max
    printf "{\"streams\":["
    for (i = 1; i <= n; i++) {
      printf "%s{\"stream\":{\"job\":\"nightshift\",\"app\":\"proven\",\"model\":\"%s\",\"provider\":\"%s\",\"status\":\"%s\",\"reason\":\"%s\",\"verified\":\"%s\"},", \
        (i > 1 ? "," : ""), MODEL[i], PROV[i], ST[i], REASON[i], VER[i]
      printf "\"values\":[[\"%.0f000000\",\"%s\"]]}", TS[i] + shift, esc(LINE[i])
    }
    printf "]}\n"
  }
' >"$payload" || {
  rc=$?
  if [ "$rc" -eq 3 ]; then
    fail "no receipt of Proven's own was found in the input; nothing was loaded"
  fi
  fail "the receipts could not be read"
}

code=$(curl -sS --max-time 20 -o /dev/null -w '%{http_code}' -X POST -H 'Content-Type: application/json' \
  --data-binary @"$payload" "$LOKI_URL/loki/api/v1/push" 2>/dev/null || true)
if [ "$code" != 204 ]; then
  fail "Loki answered HTTP ${code:-none}; the receipts were not loaded"
fi
echo "nightshift-import: the receipts are in Loki, replayed to end now"
