# Parsers for make e2e. Sourced; they read JSON that Gitea 1.22.6 serves.
# Gitea 1.22.6 has no Actions runs or tasks route (measured: HTTP 404). It does
# serve GET /repos/{o}/{r}/commits/{sha}/statuses, one entry per status change,
# newest id first. The run number is in each entry's target_url (/actions/runs/N/).

# workflow_context FILE — print "<workflow name> / <first job id> (pull_request)",
# the context Gitea gives the statuses of that workflow's job.
workflow_context() {
  awk '
    /^name:/ && name == "" { name = $2 }
    /^jobs:/ { injobs = 1; next }
    injobs && /^  [A-Za-z0-9_-]+:/ { job = $1; sub(/:$/, "", job); exit }
    END { if (name != "" && job != "") printf "%s / %s (pull_request)\n", name, job }
  ' "$1"
}

# newest_run STATUSES_JSON CONTEXT — among the entries whose context is CONTEXT,
# take the highest run number N and print "N<TAB>status<TAB>first_created_at".
# status is the newest entry of run N (the first one listed); first_created_at is
# the oldest created_at of run N, its start. Run order is by N, never by id.
newest_run() {
  printf '%s' "$1" | awk -v ctx="$2" '
    function str(obj, key) {
      if (match(obj, "\"" key "\":\"[^\"]*\"")) {
        return substr(obj, RSTART + length(key) + 4, RLENGTH - length(key) - 5)
      }
      return ""
    }
    { buf = buf $0 }
    END {
      n = split(buf, parts, "\"status\":\"")
      for (i = 2; i <= n; i++) {
        p = parts[i]
        st = substr(p, 1, index(p, "\"") - 1)
        if (str(p, "context") != ctx) continue
        url = str(p, "target_url")
        if (!match(url, "/actions/runs/[0-9]+")) continue
        r = substr(url, RSTART + 14, RLENGTH - 14) + 0
        ts = str(p, "created_at")
        if (!(r in latest)) latest[r] = st
        if (!(r in first) || ts < first[r]) first[r] = ts
        if (r > best) best = r
      }
      if (best > 0) printf "%d\t%s\t%s\n", best, latest[best], first[best]
    }'
}

# review_since COMMENTS_JSON SINCE — exit 0 when a comment that starts with
# proven's marker was created at or after SINCE. The marker may be JSON-escaped.
review_since() {
  printf '%s' "$1" | awk -v since="$2" '
    { buf = buf $0 }
    END {
      rest = buf
      while (match(rest, /"body":"(<|\\u003c)!-- proven-review --/)) {
        rest = substr(rest, RSTART + RLENGTH)
        if (match(rest, /"created_at":"[^"]*"/)) {
          ts = substr(rest, RSTART + 14, RLENGTH - 15)
          if ((ts "") >= (since "")) exit 0
        }
      }
      exit 1
    }'
}
