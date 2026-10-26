#!/bin/sh
# Posts Proven's Markdown review as ONE comment on a pull request, edited in place
# on every run (docs/ACTION.md, docs/LAB.md; ADR-0020 §3). Sourced, never run:
# entrypoint.sh sources it, and lab/workflow/proven.yml carries a byte-for-byte copy
# of the lines between the two markers below, because a workflow in a user's
# repository cannot read this file (the copy is compared by lab/scripts/post-comment-check.sh).
#
#   pc_post API_URL REPO PR RUN_NUMBER HEAD_SHA TOKEN MARKDOWN_FILE
#
# The comment is found, not remembered: the pull request's comments are listed and
# the first one written by the posting account whose body starts with
# `<!-- proven-review` is edited (PATCH .../issues/comments/{id}); a new one is
# created only when there is none. A marker comment by another account is never
# edited. The first line stays `<!-- proven-review run=N sha=H -->`, so the lab's
# `make e2e` still finds this run's review.
#
# Both Gitea and GitHub serve: GET /user, GET /repos/{r}/issues/{n}/comments,
# POST the same path, PATCH /repos/{r}/issues/comments/{id}.
#
# COMMENT_HTTP, when set, names an executable that replaces the HTTP client: it is
# called `COMMENT_HTTP METHOD URL [BODY_FILE]` and prints the response. Tests use it.
# PROVEN_COMMENT_ACCOUNT, when set, names the posting account and skips GET /user.
# PC_MAX_PAGES (default 500) bounds the walk of the comments; a walk that reaches it,
# or meets a server that ignores paging, creates nothing and fails (pc_run).

# post-comment: begin
# pc_find_comment JSON ACCOUNT — print the id of the first comment in a page of
# comments whose author is ACCOUNT and whose body starts with `<!-- proven-review`;
# print nothing when there is none. The author is read from the comment's own "user"
# object and the id from its first "id", so it holds for Gitea (user before body) and
# GitHub (body before user). A body is read only as far as its first characters.
pc_find_comment() {
  printf '%s' "$1" | awk -v account="$2" '
    { buf = buf $0 }
    END {
      # Between two comments the text is `},{"`; inside a string a quote is always
      # written \", so the same characters there never match.
      gsub(/\}, *\{ *"/, "}\n{\"", buf)
      n = split(buf, objs, "\n")
      for (i = 1; i <= n; i++) {
        obj = objs[i]
        if (!match(obj, /"id" *: *[0-9]+/)) continue
        id = substr(obj, RSTART, RLENGTH)
        sub(/^"id" *: */, "", id)
        if (!match(obj, /"body" *: *"/)) continue
        head = substr(obj, RSTART + RLENGTH, 40)
        if (index(head, "<!-- proven-review") != 1 && index(head, "\\u003c!-- proven-review") != 1) continue
        if (!match(obj, /"user" *: *\{/)) continue
        user = substr(obj, RSTART)
        if (!match(user, /"login" *: *"[^"]*"/)) continue
        login = substr(user, RSTART, RLENGTH)
        sub(/^"login" *: *"/, "", login)
        sub(/"$/, "", login)
        if (login != account) continue
        print id
        exit
      }
    }'
}

# pc_json FILE — print FILE as the JSON object {"body": "..."} the comment routes take.
pc_json() {
  awk '
    BEGIN { printf "{\"body\":\"" }
    {
      gsub(/\\/, "\\\\")
      gsub(/"/, "\\\"")
      gsub(/\t/, "\\t")
      gsub(/[\001-\010\013-\037]/, "")
      printf "%s\\n", $0
    }
    END { printf "\"}" }
  ' "$1"
}

# pc_pick_client — choose the HTTP client once. PATCH needs curl or GNU wget: BusyBox
# wget (the lab job image, Alpine docker:cli) has no --method. Without either, curl is
# installed from Alpine's packages inside this job's own container, and the choice falls
# back to BusyBox wget only when that fails; then GET and POST work and PATCH is refused.
#
# With PC_LOCAL set (PROVEN_NETWORK=local, ADR-0034) the only client is curl, run with
# -q (no .curlrc) and --noproxy '*' (no proxy, whatever the environment says), so the
# judged API address is its only route. Nothing is installed: `apk add` reaches Alpine's
# package mirrors, a host no setting names. Without curl the comment is not posted.
pc_pick_client() {
  if [ -n "${PC_LOCAL:-}" ]; then
    if command -v curl >/dev/null 2>&1; then PC_CLIENT=curl; else PC_CLIENT=none; fi
    return 0
  fi
  if command -v curl >/dev/null 2>&1; then PC_CLIENT=curl; return 0; fi
  if wget --help 2>&1 | grep -q -- '--method'; then PC_CLIENT=gnuwget; return 0; fi
  if command -v apk >/dev/null 2>&1 && apk add --no-cache curl >/dev/null 2>&1 && command -v curl >/dev/null 2>&1; then
    PC_CLIENT=curl
    return 0
  fi
  PC_CLIENT=busybox
}

# pc_curl ARGS... — curl, with no configuration file and no proxy when PC_LOCAL is set.
pc_curl() {
  if [ -n "${PC_LOCAL:-}" ]; then curl -q --noproxy '*' "$@"; else curl "$@"; fi
}

# pc_http METHOD URL [BODY_FILE] — print the response; a non-2xx answer is a failure.
pc_http() {
  if [ -n "${COMMENT_HTTP:-}" ]; then "$COMMENT_HTTP" "$@"; return; fi
  [ -n "${PC_CLIENT:-}" ] || pc_pick_client
  auth="Authorization: token $PC_TOKEN"
  case "$PC_CLIENT" in
    curl)
      if [ -n "${3:-}" ]; then
        pc_curl -fsS -X "$1" -H "$auth" -H "Content-Type: application/json" -H "Accept: application/json" --data-binary "@$3" "$2"
      else
        pc_curl -fsS -X "$1" -H "$auth" -H "Accept: application/json" "$2"
      fi
      ;;
    none)
      echo "proven: PROVEN_NETWORK=local posts the summary comment with curl only, and this image has none; install curl in the job image" >&2
      return 1
      ;;
    gnuwget)
      if [ -n "${3:-}" ]; then
        wget -q -O - --method="$1" --header="$auth" --header="Content-Type: application/json" --body-file="$3" "$2"
      else
        wget -q -O - --method="$1" --header="$auth" "$2"
      fi
      ;;
    *)
      case "$1" in
        GET) wget -q -O - --header="$auth" "$2" ;;
        POST) wget -q -O - --header="$auth" --header="Content-Type: application/json" --post-file="$3" "$2" ;;
        *)
          echo "proven: this image has no HTTP client that can $1 (BusyBox wget cannot); install curl or GNU wget in the job image" >&2
          return 1
          ;;
      esac
      ;;
  esac
}

# pc_account API_URL — print the login of the account the token acts as. A token that
# may not read /user (GitHub's installation token) acts as github-actions[bot]; on Gitea
# there is no such default, so the account must then be named in PROVEN_COMMENT_ACCOUNT.
pc_account() {
  if [ -n "${PROVEN_COMMENT_ACCOUNT:-}" ]; then printf '%s\n' "$PROVEN_COMMENT_ACCOUNT"; return 0; fi
  me=$(pc_http GET "$1/user" 2>/dev/null | awk 'match($0, /"login" *: *"[^"]*"/) { s = substr($0, RSTART, RLENGTH); sub(/^"login" *: *"/, "", s); sub(/"$/, "", s); print s; exit }')
  if [ -n "$me" ]; then printf '%s\n' "$me"; return 0; fi
  if [ "${GITEA_ACTIONS:-}" = true ]; then
    echo "proven: the token cannot read $1/user, so the posting account is unknown; set PROVEN_COMMENT_ACCOUNT" >&2
    return 1
  fi
  printf '%s\n' 'github-actions[bot]'
}

# pc_post API_URL REPO PR RUN_NUMBER HEAD_SHA TOKEN MARKDOWN_FILE — find, then edit or create.
pc_post() {
  PC_TOKEN=$6
  pc_dir=$(mktemp -d) || return 1
  pc_run "$@"
  pc_rc=$?
  rm -rf "$pc_dir"
  return "$pc_rc"
}

# pc_run — the steps of pc_post, with its arguments and $pc_dir for the files it writes.
pc_run() {
  {
    printf '<!-- proven-review run=%s sha=%s -->\n' "$4" "$5"
    cat "$7"
  } >"$pc_dir/comment.md"
  pc_json "$pc_dir/comment.md" >"$pc_dir/comment.json"
  pc_me=$(pc_account "$1") || return 1
  pc_id=
  pc_page=1
  pc_prev=
  while [ -z "$pc_id" ]; do
    # WHY a refusal and not a POST when the walk is cut off: the comment may be on a page
    # not read yet, and a second Proven comment on the pull request is the defect this
    # poster exists to prevent. The client prints no headers, so the end of the list is
    # the first page with no comment in it (the same end a missing rel="next" announces).
    if [ "$pc_page" -gt "${PC_MAX_PAGES:-500}" ]; then
      echo "proven: the comments of $2 #$3 were not read to the end after ${PC_MAX_PAGES:-500} pages, so no Proven comment is created; set PC_MAX_PAGES higher or remove old comments" >&2
      return 1
    fi
    pc_list=$(pc_http GET "$1/repos/$2/issues/$3/comments?limit=50&per_page=50&page=$pc_page") || return 1
    case "$pc_list" in *'{'*) ;; *) break ;; esac
    if [ "$pc_list" = "$pc_prev" ]; then
      echo "proven: page $pc_page of the comments of $2 #$3 is the same as page $((pc_page - 1)), so the server ignores paging and the list cannot be read to the end; no Proven comment is created" >&2
      return 1
    fi
    pc_prev=$pc_list
    pc_id=$(pc_find_comment "$pc_list" "$pc_me")
    pc_page=$((pc_page + 1))
  done
  if [ -n "$pc_id" ]; then
    pc_out=$(pc_http PATCH "$1/repos/$2/issues/comments/$pc_id" "$pc_dir/comment.json") || return 1
    pc_verb="updated comment $pc_id"
  else
    pc_out=$(pc_http POST "$1/repos/$2/issues/$3/comments" "$pc_dir/comment.json") || return 1
    pc_verb="created a comment"
  fi
  case "$pc_out" in *'"id"'*) ;; *) echo "proven: the server's answer to the comment carries no id: $pc_out" >&2; return 1 ;; esac
  echo "proven: $pc_verb on $2 #$3 as $pc_me (run $4)"
}
# post-comment: end
