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

# success_runs STATUSES_JSON CONTEXT — the numbers of the runs, highest first, one per line, whose newest status among
# the entries of CONTEXT is success. The statuses are those of one head commit, so these are the runs that finished
# a review of it. Two runs for one head are possible (the workflow lets the second post nothing), and the review is
# then the first one's.
success_runs() {
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
        if (!(r in latest)) { latest[r] = st; order[++m] = r }
      }
      for (j = 1; j <= m; j++) if (latest[order[j]] == "success") print order[j]
    }' | sort -rn
}

# review_for COMMENTS_JSON RUN SHA AUTHOR — exit 0 when a comment by AUTHOR
# starts its body with `<!-- proven-review run=RUN sha=SHA -->`. The marker
# names the run and the head commit, so no time is compared: two events in one
# second cannot be told apart by time. Gitea may write < and > as < and
# >. On stdout it prints one line per proven marker it ignored (another
# author, another run or commit), so the caller can say why.
review_for() {
  printf '%s' "$1" | awk -v run="$2" -v sha="$3" -v author="$4" '
    { buf = buf $0 }
    END {
      want = "<!-- proven-review run=" run " sha=" sha " -->"
      n = split(buf, parts, "\"body\":\"")
      ok = 0
      for (i = 2; i <= n; i++) {
        # The poster is the last "login" before this body: the user object precedes it.
        before = parts[i - 1]
        login = ""
        while (match(before, /"login":"[^"]*"/)) {
          login = substr(before, RSTART + 9, RLENGTH - 10)
          before = substr(before, RSTART + RLENGTH)
        }
        body = parts[i]
        gsub(/\\u003c/, "<", body)
        gsub(/\\u003e/, ">", body)
        if (index(body, "<!-- proven-review") != 1) continue
        if (index(body, want) != 1) {
          printf "ignored: a proven review by %s names another run or commit\n", login
        } else if (login != author) {
          printf "ignored: a proven review for run #%s by %s, not by %s\n", run, login, author
        } else {
          ok = 1
        }
      }
      exit ok ? 0 : 1
    }'
}
