#!/bin/sh
# Offline check of the pull-request comment poster (action/post-comment.sh): the
# find-or-create choice, the edit in place, and that the lab workflow's copy of the
# poster is the same text. No network, no Docker: the HTTP client is a stub
# (COMMENT_HTTP). Run: make check-e2e-parse (it calls this), or sh lab/scripts/post-comment-check.sh
set -u
here=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT=$(CDPATH= cd -- "$here/../.." && pwd)
. "$here/e2e-parse.sh"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

fail=0
expect() { # NAME WANT GOT
  if [ "$2" != "$3" ]; then
    echo "post-comment-check: FAIL: $1: got '$3', want '$2'" >&2
    fail=1
  fi
}

# --- the workflow carries the same poster ------------------------------------
# WHY: a workflow in a user's repository cannot read action/post-comment.sh, so it
# holds a copy. The copy is indented ten spaces under `run: |`; strip them and compare.
region() { # FILE STRIP — the lines from the begin marker to the end marker
  awk -v strip="$2" '
    index($0, "# post-comment: begin") { on = 1 }
    on { print substr($0, strip + 1) }
    index($0, "# post-comment: end") { on = 0 }
  ' "$1"
}
region "$ROOT/action/post-comment.sh" 0 >"$tmp/action.region"
region "$ROOT/lab/workflow/proven.yml" 10 >"$tmp/workflow.region"
expect "the poster region exists" 1 "$([ -s "$tmp/action.region" ] && echo 1 || echo 0)"
if ! cmp -s "$tmp/action.region" "$tmp/workflow.region"; then
  echo "post-comment-check: FAIL: lab/workflow/proven.yml carries a poster that differs from action/post-comment.sh:" >&2
  diff "$tmp/action.region" "$tmp/workflow.region" | head -20 >&2
  fail=1
fi

. "$ROOT/action/post-comment.sh"

# --- pc_find_comment ----------------------------------------------------------
# Comments in the shape Gitea serves (the user object precedes the body) and in the
# shape GitHub serves (the body precedes the user object).
gitea() { # ID LOGIN BODY — body is JSON string content
  printf '{"id":%s,"html_url":"http://lab/p/d/pulls/1#issuecomment-%s","user":{"id":1,"login":"%s","login_name":"","full_name":""},"original_author":"","body":"%s","created_at":"2026-10-24T04:31:02Z"}' "$1" "$1" "$2" "$3"
}
github() { # ID LOGIN BODY
  printf '{"url":"https://api.github.com/repos/o/r/issues/comments/%s","html_url":"https://github.com/o/r/pull/1#issuecomment-%s","issue_url":"https://api.github.com/repos/o/r/issues/1","id":%s,"node_id":"IC_x","user":{"login":"%s","id":9,"type":"Bot"},"created_at":"2026-10-24T04:31:02Z","updated_at":"2026-10-24T04:31:02Z","author_association":"NONE","body":"%s","reactions":{"url":"u","total_count":0}}' "$1" "$1" "$1" "$2" "$3"
}
marker='<!-- proven-review run=3 sha=0123456789abcdef -->\n**Proven: blocked**'

c=$(gitea 11 proven-bot "$marker")
expect "gitea: own marker comment is found" 11 "$(pc_find_comment "[$c]" proven-bot)"
c=$(github 12 github-actions[bot] "$marker")
expect "github: own marker comment is found" 12 "$(pc_find_comment "[$c]" 'github-actions[bot]')"
c=$(gitea 13 proven-bot '<!-- proven-review run=3 sha=0123456789abcdef -->\n**Proven**')
expect "an escaped marker (\\u003c) is found" 13 "$(pc_find_comment "[$c]" proven-bot)"

# The property of the order: a marker comment by another account is NOT edited.
other=$(gitea 14 someone-else "$marker")
expect "another account's marker comment is not chosen" "" "$(pc_find_comment "[$other]" proven-bot)"
# Negative control (law #27): the one field changed is the account asked for. Asked for
# as the author, the same JSON returns the comment; the check above is about the author.
expect "control: the same JSON, asked for its own author, is chosen" 14 "$(pc_find_comment "[$other]" someone-else)"

c=$(gitea 15 proven-bot 'looks good to me')
expect "own comment without the marker is not chosen" "" "$(pc_find_comment "[$c]" proven-bot)"
own=$(gitea 16 proven-bot "$marker")
expect "another's marker first, own second: own is chosen" 16 "$(pc_find_comment "[$other,$own]" proven-bot)"
expect "own first, another's second: own is chosen" 16 "$(pc_find_comment "[$own,$other]" proven-bot)"
expect "an empty page has no comment" "" "$(pc_find_comment '[]' proven-bot)"

# A body that quotes JSON cannot pass for a comment of the posting account: inside a
# string every quote is \" so the author and the separator in it match nothing.
quote='<!-- proven-review run=1 sha=a -->\n},{\"id\":99,\"user\":{\"login\":\"proven-bot\"},\"body\":\"<!-- proven-review x\"'
c=$(gitea 17 someone-else "$quote")
expect "a body that quotes a comment of the account does not become one" "" "$(pc_find_comment "[$c]" proven-bot)"
# The same text with a real separator and a real author is two comments, and the right one wins.
expect "two comments in a row: the account's is chosen" 18 "$(pc_find_comment "[$c,$(gitea 18 proven-bot "$marker")]" proven-bot)"

# --- pc_post against a stub server -------------------------------------------
cat >"$tmp/stub.sh" <<'EOF'
#!/bin/sh
# The stub server: GET /user answers $STUB/user.json (or fails when absent), GET
# comments answers $STUB/page-N.json (or []), PATCH and POST answer {"id":77}.
# Every request is logged as "METHOD URL"; a body file is copied to $STUB/body.N.
method=$1
url=$2
n=$(wc -l <"$STUB/log" | tr -d ' ')
printf '%s %s\n' "$method" "$url" >>"$STUB/log"
[ -z "${3:-}" ] || cp "$3" "$STUB/body.$((n + 1))"
case "$method $url" in
  "GET "*/user)
    [ -f "$STUB/user.json" ] || exit 22
    cat "$STUB/user.json"
    ;;
  "GET "*/comments*)
    page=$(printf '%s' "$url" | sed 's/.*page=//')
    if [ -f "$STUB/page-$page.json" ]; then cat "$STUB/page-$page.json"; else printf '[]'; fi
    ;;
  PATCH* | POST*) printf '{"id":77}' ;;
esac
EOF
chmod +x "$tmp/stub.sh"

post() { # NAME — runs pc_post as the account in $STUB/user.json; sets $log, $STUB
  STUB=$tmp/$1
  mkdir -p "$STUB"
  : >"$STUB/log"
  printf '# Review\n\nbody "quoted" \\ back\\slash\n' >"$STUB/review.md"
}
run_post() { # run, sha
  COMMENT_HTTP=$tmp/stub.sh STUB=$STUB PROVEN_COMMENT_ACCOUNT=${ACCOUNT:-} \
    pc_post http://api.example/api/v1 owner/repo 7 "$1" "$2" tok-1 "$STUB/review.md" >"$STUB/out" 2>"$STUB/err"
  rc=$?
  log=$(cat "$STUB/log")
}
export COMMENT_HTTP STUB PROVEN_COMMENT_ACCOUNT

sha=0123456789abcdef0123456789abcdef01234567
base='http://api.example/api/v1'

# 1. no earlier comment: one POST, no PATCH.
post create
ACCOUNT=proven-bot run_post 1 "$sha"
expect "no earlier comment: exit" 0 "$rc"
expect "no earlier comment: requests" "GET $base/repos/owner/repo/issues/7/comments?limit=50&per_page=50&page=1
POST $base/repos/owner/repo/issues/7/comments" "$log"

# 2. an earlier comment of the account: PATCH it, never POST, and the marker names this run.
post edit
printf '[%s]' "$(gitea 31 proven-bot '<!-- proven-review run=1 sha=aaa -->\nold')" >"$STUB/page-1.json"
ACCOUNT=proven-bot run_post 2 "$sha"
expect "earlier comment: exit" 0 "$rc"
expect "earlier comment: it is edited in place, nothing is created" "GET $base/repos/owner/repo/issues/7/comments?limit=50&per_page=50&page=1
PATCH $base/repos/owner/repo/issues/comments/31" "$log"
sent=$(cat "$STUB/body.2")
case $sent in
  '{"body":"<!-- proven-review run=2 sha='"$sha"' -->\n# Review\n\nbody \"quoted\" \\ back\\slash\n"}') ;;
  *) echo "post-comment-check: FAIL: the edited body is not the new marker plus the Markdown: $sent" >&2; fail=1 ;;
esac
# The body the poster sends is what `make e2e` looks for: the same parser accepts it.
review_for "[$(gitea 31 proven-bot "$(printf '%s' "$sent" | sed 's/^{"body":"//; s/"}$//')")]" 2 "$sha" proven-bot >/dev/null ||
  { echo "post-comment-check: FAIL: e2e's parser does not accept the posted body as run 2's review" >&2; fail=1; }

# 3. the only marker comment is another account's: it is NOT edited; one is created.
post other
printf '[%s]' "$(gitea 41 someone-else '<!-- proven-review run=1 sha=aaa -->\nold')" >"$STUB/page-1.json"
ACCOUNT=proven-bot run_post 3 "$sha"
expect "another account's marker: a comment is created, not edited" "GET $base/repos/owner/repo/issues/7/comments?limit=50&per_page=50&page=1
GET $base/repos/owner/repo/issues/7/comments?limit=50&per_page=50&page=2
POST $base/repos/owner/repo/issues/7/comments" "$log"
case $log in *PATCH*) echo "post-comment-check: FAIL: another account's comment was edited" >&2; fail=1 ;; esac

# 4. the account's comment is on the second page: found there.
post paged
printf '[%s]' "$(gitea 51 someone-else 'hello')" >"$STUB/page-1.json"
printf '[%s]' "$(gitea 52 proven-bot '<!-- proven-review run=1 sha=aaa -->\nold')" >"$STUB/page-2.json"
ACCOUNT=proven-bot run_post 4 "$sha"
expect "second page: edited there" "GET $base/repos/owner/repo/issues/7/comments?limit=50&per_page=50&page=1
GET $base/repos/owner/repo/issues/7/comments?limit=50&per_page=50&page=2
PATCH $base/repos/owner/repo/issues/comments/52" "$log"

# 4b. the account's comment is on page 21, past the old 20-page bound: it is edited and
# nothing is created. (Red at 031efe4: the walk stopped at page 20 and POSTed a second comment.)
post page21
p=1
while [ "$p" -le 20 ]; do
  printf '[%s]' "$(gitea "$((100 + p))" someone-else "hello $p")" >"$STUB/page-$p.json"
  p=$((p + 1))
done
printf '[%s]' "$(gitea 121 proven-bot '<!-- proven-review run=1 sha=aaa -->\nold')" >"$STUB/page-21.json"
ACCOUNT=proven-bot run_post 4 "$sha"
expect "page 21: exit" 0 "$rc"
expect "page 21: the comment is edited" 1 "$(printf '%s\n' "$log" | grep -c "^PATCH $base/repos/owner/repo/issues/comments/121\$")"
expect "page 21: nothing is created" 0 "$(printf '%s\n' "$log" | grep -c '^POST ')"

# 4c. a walk that is cut off creates nothing and says why: the bound is reached with
# every page still full, so the comment may be on a page not read.
post cutoff
p=1
while [ "$p" -le 5 ]; do
  printf '[%s]' "$(gitea "$((100 + p))" someone-else "hello $p")" >"$STUB/page-$p.json"
  p=$((p + 1))
done
PC_MAX_PAGES=3 ACCOUNT=proven-bot run_post 4 "$sha"
expect "cut-off walk: the step fails" 1 "$rc"
expect "cut-off walk: nothing is created or edited" 0 "$(printf '%s\n' "$log" | grep -c -e '^POST ' -e '^PATCH ')"
case $(cat "$STUB/err") in *"were not read to the end after 3 pages"*) ;; *) echo "post-comment-check: FAIL: the cut-off refusal does not say why: $(cat "$STUB/err")" >&2; fail=1 ;; esac

# 4d. a server that ignores paging answers the same page twice: the list cannot be read
# to the end, so nothing is created.
post nopaging
printf '[%s]' "$(gitea 51 someone-else 'hello')" >"$STUB/page-1.json"
printf '[%s]' "$(gitea 51 someone-else 'hello')" >"$STUB/page-2.json"
ACCOUNT=proven-bot run_post 4 "$sha"
expect "a server that ignores paging: the step fails" 1 "$rc"
expect "a server that ignores paging: nothing is created" 0 "$(printf '%s\n' "$log" | grep -c '^POST ')"
case $(cat "$STUB/err") in *"ignores paging"*) ;; *) echo "post-comment-check: FAIL: the paging refusal does not say why: $(cat "$STUB/err")" >&2; fail=1 ;; esac

# 5. the account is asked of the server when it is not named: GET /user.
post whoami
printf '{"id":3,"login":"proven-bot","email":"x"}' >"$STUB/user.json"
printf '[%s]' "$(gitea 61 proven-bot '<!-- proven-review run=1 sha=aaa -->\nold')" >"$STUB/page-1.json"
ACCOUNT= run_post 5 "$sha"
expect "account from GET /user: edited" "GET $base/user
GET $base/repos/owner/repo/issues/7/comments?limit=50&per_page=50&page=1
PATCH $base/repos/owner/repo/issues/comments/61" "$log"

# 6. /user refused: on GitHub the account is github-actions[bot]; on Gitea there is no guess.
post nouser
printf '[%s]' "$(github 71 'github-actions[bot]' '<!-- proven-review run=1 sha=aaa -->\nold')" >"$STUB/page-1.json"
ACCOUNT= run_post 6 "$sha"
expect "no /user on GitHub: github-actions[bot] is the account" 0 "$rc"
case $log in *"PATCH $base/repos/owner/repo/issues/comments/71"*) ;; *) echo "post-comment-check: FAIL: the bot's comment was not edited: $log" >&2; fail=1 ;; esac
post nouser-gitea
GITEA_ACTIONS=true ACCOUNT= run_post 6 "$sha"
expect "no /user on Gitea: refused" 1 "$rc"
case $(cat "$STUB/err") in *PROVEN_COMMENT_ACCOUNT*) ;; *) echo "post-comment-check: FAIL: the refusal does not name PROVEN_COMMENT_ACCOUNT" >&2; fail=1 ;; esac
expect "no /user on Gitea: nothing was posted" "GET $base/user" "$log"

# 7. a server answer without an id is a failure, not a success.
post noid
printf '#!/bin/sh\ncase "$1" in PATCH|POST) printf "{}" ;; *) printf "[]" ;; esac\n' >"$STUB/noid.sh"
chmod +x "$STUB/noid.sh"
COMMENT_HTTP=$STUB/noid.sh PROVEN_COMMENT_ACCOUNT=proven-bot pc_post "$base" owner/repo 7 1 "$sha" tok-1 "$STUB/review.md" >/dev/null 2>&1
expect "an answer without an id fails" 1 "$?"

[ "$fail" -eq 0 ] && echo "post-comment-check: ok"
exit "$fail"
