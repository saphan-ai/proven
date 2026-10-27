#!/bin/sh
# Create the demo repository, the workflow, the workflow's own account, the runner token, the demo pull requests
# (lab/seed/prs.tsv) and the lab's start page. The admin user already exists. The password is read from
# /state/password, a file in the lab's volume that make up wrote at mode 0600. It is never an environment variable.
# The image runs as the user proven, which owns /state, /reg and /page (lab/seed/Dockerfile).
set -eu

umask 077

# The credentials file make up wrote (lab/seed/golive.sh reads it). It is removed on every way out of this script: a
# finish, a failure under set -e, a TERM or an INT. Only a KILL leaves it, and make up removes it again after the seed.
LAB_SECRETS_FILE=${LAB_SECRETS_FILE:-/state/secrets}
trap 'rm -f "$LAB_SECRETS_FILE"' EXIT
trap 'exit 143' TERM
trap 'exit 130' INT

: "${GITEA_INTERNAL_URL:?}"
: "${LAB_ADMIN_USER:?}"
: "${LAB_BOT_USER:?}"
: "${LAB_BOT_EMAIL:?}"
: "${LAB_DEMO_REPO:?}"
: "${LAB_DEMO_BASE:?}"
: "${LAB_DEMO_HEAD:?}"
: "${PROVEN_IMAGE:?}"
: "${RUNNER_LABEL:?}"
: "${LAB_URL:?}"

# The demo pull requests and the start page's template, and the fixtures the pull requests are made from.
# The compose file mounts all three read-only.
PRS_FILE=${PRS_FILE:-/seed/prs.tsv}
PAGE_TEMPLATE=${PAGE_TEMPLATE:-/seed/lab.html.tmpl}
FIXTURES_DIR=${FIXTURES_DIR:-/fixtures}
PAGE_DIR=${PAGE_DIR:-/page}
# html_escape and render_page: the start page, kept apart so a test can render it offline.
. "$(dirname "$0")/render.sh"
# ensure_live_secrets: the keys of lab/.env as Actions secrets, kept apart so a test can run it against a stub API.
. "$(dirname "$0")/golive.sh"

api() {
  method=$1
  url=$2
  outfile=$3
  shift 3
  curl -sS --max-time 30 -o "$outfile" -w '%{http_code}' -X "$method" \
    -H "Authorization: token $API_TOKEN" \
    -H 'Content-Type: application/json' \
    "$@" \
    "$url"
}

json_string() {
  key=$1
  file=$2
  sed -n "s/.*\"${key}\":\"\\([^\"]*\\)\".*/\\1/p" "$file" | head -n 1
}

wait_ready() {
  i=0
  while [ "$i" -lt 60 ]; do
    if curl -fsS --max-time 3 "$GITEA_INTERNAL_URL/api/healthz" >/dev/null 2>&1; then
      return 0
    fi
    i=$((i + 1))
    sleep 2
  done
  echo "seed: Gitea did not become healthy" >&2
  exit 1
}

read_password() {
  if [ ! -s /state/password ]; then
    echo "seed: /state/password is missing or empty" >&2
    exit 1
  fi
  LAB_ADMIN_PASSWORD=$(cat /state/password)
}

ensure_token() {
  if [ -s /state/api-token ]; then
    API_TOKEN=$(cat /state/api-token)
    return 0
  fi
  code=$(curl -sS --max-time 30 -o /tmp/token.json -w '%{http_code}' \
    -u "$LAB_ADMIN_USER:$LAB_ADMIN_PASSWORD" \
    -H 'Content-Type: application/json' \
    -d '{"name":"lab-seed","scopes":["all"]}' \
    "$GITEA_INTERNAL_URL/api/v1/users/$LAB_ADMIN_USER/tokens")
  if [ "$code" != "201" ]; then
    code=$(curl -sS --max-time 30 -o /tmp/token.json -w '%{http_code}' \
      -u "$LAB_ADMIN_USER:$LAB_ADMIN_PASSWORD" \
      -H 'Content-Type: application/json' \
      -d '{"name":"lab-seed","scopes":["write:repository","write:user","write:admin","write:issue"]}' \
      "$GITEA_INTERNAL_URL/api/v1/users/$LAB_ADMIN_USER/tokens")
  fi
  API_TOKEN=$(json_string sha1 /tmp/token.json)
  if [ -z "$API_TOKEN" ]; then
    API_TOKEN=$(json_string token /tmp/token.json)
  fi
  if [ -z "$API_TOKEN" ]; then
    echo "seed: could not create an API token (HTTP $code)" >&2
    exit 1
  fi
  printf '%s' "$API_TOKEN" >/state/api-token
  chmod 600 /state/api-token
}

ensure_repo() {
  code=$(api POST "$GITEA_INTERNAL_URL/api/v1/user/repos" /tmp/repo.json \
    -d "{\"name\":\"$LAB_DEMO_REPO\",\"private\":false,\"auto_init\":true,\"default_branch\":\"$LAB_DEMO_BASE\"}")
  if [ "$code" != "201" ] && [ "$code" != "409" ]; then
    echo "seed: could not create the repository (HTTP $code)" >&2
    exit 1
  fi
}

# Gitea runs a workflow only for a repository whose Actions unit is on.
# A repository made through the API gets the units in [repository]
# DEFAULT_REPO_UNITS, and a repository kept from an older volume may lack it.
# A pull request opened while the unit was off has no task, so a change here
# sets RETRIGGER and ensure_pr starts the workflow again.
RETRIGGER=0

ensure_actions_unit() {
  repo="$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO"
  code=$(curl -sS --max-time 30 -o /tmp/unit.json -w '%{http_code}' \
    -H "Authorization: token $API_TOKEN" "$repo")
  if [ "$code" != "200" ]; then
    echo "seed: could not read the repository (HTTP $code)" >&2
    exit 1
  fi
  if grep -q '"has_actions":true' /tmp/unit.json; then
    return 0
  fi
  code=$(api PATCH "$repo" /tmp/unit.json -d '{"has_actions":true}')
  if [ "$code" != "200" ]; then
    echo "seed: could not turn on the Actions unit (HTTP $code)" >&2
    exit 1
  fi
  if ! grep -q '"has_actions":true' /tmp/unit.json; then
    echo "seed: Gitea did not turn on the Actions unit" >&2
    exit 1
  fi
  RETRIGGER=1
}

# The job runs on the lab's inner daemon, where 127.0.0.1 is not Gitea.
# github.api_url is built from ROOT_URL, so the seeded copy posts to the
# address the lab network serves instead.
render_workflow() {
  sed "s|saphan/proven:judging|${PROVEN_IMAGE}:latest|g; s|runs-on: ubuntu-latest|runs-on: ${RUNNER_LABEL}|; s|\${{ github.api_url }}|${GITEA_INTERNAL_URL}/api/v1|" \
    /workflow/proven.yml >/tmp/proven.yml
  # ADR-0045: the review in the lab is exported to the lab's collector. The job's `docker run` starts
  # Proven on the inner daemon, whose default network cannot resolve the lab's service names, so the
  # container joins the daemon's own network, which is on the lab network. Proven reads the standard
  # OTEL_* variables and nothing else. Without an endpoint the workflow is left as it is.
  # The review's instance id carries the pull request and the change: "pr<number>-<fixture>", with no time in it. A pull
  # request reviewed again (a new push) keeps the same id, so its series continue and the dashboards count the latest
  # review of each pull request, one thing, never a review per run. Proven's metrics are labelled with it (instance), so
  # the dashboards name a review by its pull request, and the clean changes from the defect ones. Proven's own setup lets OTEL_RESOURCE_ATTRIBUTES say more than the id it made (internal/telemetry/setup.go).
  if [ -n "${OTEL_EXPORTER_OTLP_ENDPOINT:-}" ]; then
    sed "s|docker run --rm |docker run --rm --network host -e OTEL_EXPORTER_OTLP_ENDPOINT=${OTEL_EXPORTER_OTLP_ENDPOINT} -e OTEL_SERVICE_NAME=${OTEL_SERVICE_NAME:-proven} -e OTEL_METRIC_EXPORT_INTERVAL=${OTEL_METRIC_EXPORT_INTERVAL:-1000} -e OTEL_RESOURCE_ATTRIBUTES=service.instance.id=pr\${PR}-\${fixture} |" \
      /tmp/proven.yml >/tmp/proven.otel
    mv /tmp/proven.otel /tmp/proven.yml
  fi
  # ADR-0009, ADR-0045: the lab's review runs a second reviewer model, so the dashboard shows two.
  # The sample generator answers for both, so the findings do not change; the review names the second reviewer.
  # ADR-0050 point 6: the one automated fix attempt is on in the judges' lab and off everywhere else, so the
  # lab's review runs with a proven.yaml that turns it on, with or without the second reviewer. The command
  # sits inside single quotes in the workflow, so the file is written with echo and no quote of its own.
  # The second reviewer is a sample-mode device (the sample generator answers for it); a live review has one real
  # model behind each call, so the line is written only while SAMPLE_MODE is true.
  second=""
  if [ -n "${LAB_SPECIAL_REVIEWER:-}" ]; then
    second=" \&\& { [ \$SAMPLE_MODE != true ] || echo special_reviewer: ${LAB_SPECIAL_REVIEWER} >> lab.yaml; }"
  fi
  # The delimiter is #: the spliced text carries "||", which would end a |-delimited expression.
  sed "s#proven scan --md #echo autofix: on > lab.yaml${second} \&\& proven scan --config lab.yaml --md #" \
    /tmp/proven.yml >/tmp/proven.two
  mv /tmp/proven.two /tmp/proven.yml
  # Path B: where the gateway serves Zetaris, and the organisation id it expects, are settings (ADR-0013).
  sed "s|GATEWAY_ZETARIS_PATH: /zetaris/api/v1.0|GATEWAY_ZETARIS_PATH: ${JUDGES_GATEWAY_ZETARIS_PATH:-/zetaris/api/v1.0}|; s|GATEWAY_ZETARIS_ORG_ID: \"0\"|GATEWAY_ZETARIS_ORG_ID: \"${JUDGES_GATEWAY_ZETARIS_ORG_ID:-0}\"|" \
    /tmp/proven.yml >/tmp/proven.gw
  mv /tmp/proven.gw /tmp/proven.yml
}

ensure_workflow() {
  render_workflow
  b64=$(base64 </tmp/proven.yml | tr -d '\n')
  url="$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/contents/.github/workflows/proven.yml"
  body=$(printf '{"message":"Add the proven workflow","branch":"%s","content":"%s"}' "$LAB_DEMO_BASE" "$b64")
  code=$(api POST "$url" /tmp/wf.json -d "$body")
  if [ "$code" = "201" ]; then
    return 0
  fi
  get=$(curl -sS --max-time 30 -o /tmp/wfget.json -w '%{http_code}' \
    -H "Authorization: token $API_TOKEN" \
    "$url?ref=$LAB_DEMO_BASE")
  sha=$(json_string sha /tmp/wfget.json)
  if [ "$get" != "200" ] || [ -z "$sha" ]; then
    echo "seed: could not add the workflow (HTTP $code)" >&2
    exit 1
  fi
  body=$(printf '{"message":"Update the proven workflow","branch":"%s","sha":"%s","content":"%s"}' \
    "$LAB_DEMO_BASE" "$sha" "$b64")
  code=$(api PUT "$url" /tmp/wf.json -d "$body")
  # 422: the file is already the copy we would write. A second make up must not stop.
  if [ "$code" != "200" ] && [ "$code" != "201" ] && [ "$code" != "422" ]; then
    echo "seed: could not update the workflow (HTTP $code)" >&2
    exit 1
  fi
}

# The workflow posts as its own account, so a review the workflow wrote and a
# comment a person wrote (the admin signs in to the lab) are told apart by
# author. The account is not an admin. Its password stays in /state and is
# never printed; its token goes only into the repository's Actions secret.
ensure_bot() {
  if [ ! -s /state/bot-password ]; then
    head -c 36 /dev/urandom | base64 | tr -d '/+=\n' >/state/bot-password
    chmod 600 /state/bot-password
  fi
  bot_pass=$(cat /state/bot-password)
  code=$(api POST "$GITEA_INTERNAL_URL/api/v1/admin/users" /tmp/bot.json \
    -d "{\"username\":\"$LAB_BOT_USER\",\"email\":\"$LAB_BOT_EMAIL\",\"password\":\"$bot_pass\",\"must_change_password\":false,\"send_notify\":false}")
  if [ "$code" != "201" ] && [ "$code" != "422" ]; then
    echo "seed: could not create the workflow's account (HTTP $code)" >&2
    exit 1
  fi
  # 422 means "already exists" only if the account is there: Gitea also answers 422 for an email it will not take
  # (proven-bot@localhost was refused, measured on Gitea 1.22.6), and then the account was never made.
  code=$(api GET "$GITEA_INTERNAL_URL/api/v1/users/$LAB_BOT_USER" /tmp/bot-get.json)
  if [ "$code" != "200" ]; then
    echo "seed: the workflow's account $LAB_BOT_USER does not exist after creating it (HTTP $code); check LAB_BOT_EMAIL in lab/versions.env" >&2
    exit 1
  fi
  # A kept Gitea volume may hold the account with another password.
  code=$(api PATCH "$GITEA_INTERNAL_URL/api/v1/admin/users/$LAB_BOT_USER" /tmp/bot.json \
    -d "{\"login_name\":\"$LAB_BOT_USER\",\"source_id\":0,\"password\":\"$bot_pass\",\"must_change_password\":false,\"admin\":false}")
  if [ "$code" != "200" ]; then
    echo "seed: could not set the workflow account's password (HTTP $code)" >&2
    exit 1
  fi
  code=$(api PUT "$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/collaborators/$LAB_BOT_USER" /tmp/bot.json \
    -d '{"permission":"write"}')
  if [ "$code" != "204" ] && [ "$code" != "200" ]; then
    echo "seed: could not make the workflow's account a collaborator (HTTP $code)" >&2
    exit 1
  fi
  # A token is shown once, so mint a fresh one on every run. The same name
  # cannot be made twice: delete the old one first (404 when there is none).
  # read:user lets the workflow ask GET /user which account it posts as, so it can
  # find and edit its own earlier comment instead of adding one per run.
  tokens="$GITEA_INTERNAL_URL/api/v1/users/$LAB_BOT_USER/tokens"
  curl -sS --max-time 30 -o /dev/null -X DELETE -u "$LAB_BOT_USER:$bot_pass" "$tokens/lab-workflow" || true
  code=$(curl -sS --max-time 30 -o /tmp/bot-token.json -w '%{http_code}' \
    -u "$LAB_BOT_USER:$bot_pass" \
    -H 'Content-Type: application/json' \
    -d '{"name":"lab-workflow","scopes":["write:issue","write:repository","read:user"]}' \
    "$tokens")
  BOT_TOKEN=$(json_string sha1 /tmp/bot-token.json)
  if [ -z "$BOT_TOKEN" ]; then
    BOT_TOKEN=$(json_string token /tmp/bot-token.json)
  fi
  rm -f /tmp/bot-token.json
  if [ -z "$BOT_TOKEN" ]; then
    echo "seed: could not create the workflow account's token (HTTP $code)" >&2
    exit 1
  fi
}

ensure_secret() {
  code=$(api PUT "$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/actions/secrets/PROVEN_TOKEN" /tmp/secret.json \
    -d "{\"data\":\"$BOT_TOKEN\"}")
  if [ "$code" != "201" ] && [ "$code" != "204" ] && [ "$code" != "200" ]; then
    echo "seed: could not store the actions secret (HTTP $code)" >&2
    exit 1
  fi
}

ensure_runner_token() {
  # Always write a token. A kept volume may hold one from a run that died
  # before the runner registered, and that token may already be spent.
  # The runner reads this file only when it has no registration of its own.
  repo_token="$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/actions/runners/registration-token"
  code=$(api POST "$repo_token" /tmp/reg.json -d '{}')
  token=$(json_string token /tmp/reg.json)
  if [ -z "$token" ]; then
    code=$(api GET "$repo_token" /tmp/reg.json)
    token=$(json_string token /tmp/reg.json)
  fi
  if [ -z "$token" ]; then
    code=$(api GET "$GITEA_INTERNAL_URL/api/v1/admin/actions/runners/registration-token" /tmp/reg.json)
    token=$(json_string token /tmp/reg.json)
  fi
  if [ -z "$token" ]; then
    code=$(api GET "$GITEA_INTERNAL_URL/api/v1/admin/runners/registration-token" /tmp/reg.json)
    token=$(json_string token /tmp/reg.json)
  fi
  if [ -z "$token" ]; then
    echo "seed: could not get a runner registration token (HTTP $code)" >&2
    exit 1
  fi
  printf '%s' "$token" >/reg/token
  chmod 644 /reg/token
}

repo_api="$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO"
TAB=$(printf '\t')

# put_file BRANCH PATH LOCALFILE MESSAGE: create the file on the branch, or update it. A file that already holds
# these bytes is left alone, so a second make up makes no commit and starts no review.
put_file() {
  pf_branch=$1
  pf_path=$2
  pf_src=$3
  pf_msg=$4
  pf_b64=$(base64 <"$pf_src" | tr -d '\n')
  pf_url="$repo_api/contents/$pf_path"
  code=$(curl -sS --max-time 30 -o /tmp/pf-get.json -w '%{http_code}' \
    -H "Authorization: token $API_TOKEN" "$pf_url?ref=$pf_branch")
  if [ "$code" = "200" ]; then
    pf_have=$(sed -n 's/.*"content":"\([^"]*\)".*/\1/p' /tmp/pf-get.json | head -n 1 | sed 's/\\n//g')
    if [ "$pf_have" = "$pf_b64" ]; then
      return 0
    fi
    pf_sha=$(json_string sha /tmp/pf-get.json)
    code=$(api PUT "$pf_url" /tmp/pf.json \
      -d "$(printf '{"message":"%s","branch":"%s","sha":"%s","content":"%s"}' "$pf_msg" "$pf_branch" "$pf_sha" "$pf_b64")")
  elif [ "$code" = "404" ]; then
    code=$(api POST "$pf_url" /tmp/pf.json \
      -d "$(printf '{"message":"%s","branch":"%s","content":"%s"}' "$pf_msg" "$pf_branch" "$pf_b64")")
  else
    echo "seed: could not read $pf_path on $pf_branch (HTTP $code)" >&2
    exit 1
  fi
  # 422: the file is already the copy we would write.
  if [ "$code" != "200" ] && [ "$code" != "201" ] && [ "$code" != "422" ]; then
    echo "seed: could not write $pf_path on $pf_branch (HTTP $code)" >&2
    exit 1
  fi
}

# The base tree of the demo repository is the base tree of the discount fixture. Every demo pull request is a
# change on top of it.
ensure_base() {
  base_dir="$FIXTURES_DIR/discount/base"
  if [ ! -d "$base_dir" ]; then
    echo "seed: the discount fixture is not mounted at $FIXTURES_DIR" >&2
    exit 1
  fi
  (cd "$base_dir" && find . -type f | sed 's|^\./||' | sort) >/tmp/base-files.txt
  while IFS= read -r rel; do
    put_file "$LAB_DEMO_BASE" "$rel" "$base_dir/$rel" "Add the discount package: $rel"
  done </tmp/base-files.txt
}

ensure_branch() {
  eb_branch=$1
  code=$(api POST "$repo_api/branches" /tmp/branch.json \
    -d "{\"new_branch_name\":\"$eb_branch\",\"old_ref_name\":\"$LAB_DEMO_BASE\"}")
  if [ "$code" != "201" ] && [ "$code" != "409" ]; then
    code=$(api POST "$repo_api/branches" /tmp/branch.json \
      -d "{\"new_branch_name\":\"$eb_branch\",\"old_branch_name\":\"$LAB_DEMO_BASE\"}")
  fi
  if [ "$code" != "201" ] && [ "$code" != "409" ]; then
    echo "seed: could not create the branch $eb_branch (HTTP $code)" >&2
    exit 1
  fi
}

# apply_change FIXTURE BRANCH TITLE: commit the fixture's change tree on the branch. go.mod is the one file whose
# module name differs between fixtures, so it is the base go.mod plus the `require` lines the fixture adds.
apply_change() {
  ac_fixture=$1
  ac_branch=$2
  ac_title=$3
  ac_dir="$FIXTURES_DIR/$ac_fixture/change"
  if [ ! -d "$ac_dir" ]; then
    echo "seed: no fixture $ac_fixture under $FIXTURES_DIR" >&2
    exit 1
  fi
  (cd "$ac_dir" && find . -type f | sed 's|^\./||' | sort) >/tmp/change-files.txt
  while IFS= read -r rel; do
    if [ "$rel" = "go.mod" ]; then
      {
        cat "$FIXTURES_DIR/discount/base/go.mod"
        if grep -q '^require ' "$ac_dir/go.mod"; then
          echo
          grep '^require ' "$ac_dir/go.mod"
        fi
      } >/tmp/go.mod.new
      put_file "$ac_branch" go.mod /tmp/go.mod.new "$ac_title"
    else
      put_file "$ac_branch" "$rel" "$ac_dir/$rel" "$ac_title"
    fi
  done </tmp/change-files.txt
}

# open_pr BRANCH TITLE SAYS FIXTURE: open the pull request, or find the one that is open; sets PR_NUMBER.
open_pr() {
  op_branch=$1
  op_title=$2
  op_says=$3
  op_fixture=$4
  body=$(printf '{"title":"%s","head":"%s","base":"%s","body":"%s\\n\\nSample set: testdata/fixture/%s in the Proven repository."}' \
    "$op_title" "$op_branch" "$LAB_DEMO_BASE" "$op_says" "$op_fixture")
  code=$(api POST "$repo_api/pulls" /tmp/pr.json -d "$body")
  if [ "$code" != "201" ] && [ "$code" != "409" ]; then
    echo "seed: could not open the pull request for $op_branch (HTTP $code)" >&2
    exit 1
  fi
  code=$(api GET "$repo_api/pulls/$LAB_DEMO_BASE/$op_branch" /tmp/pr-get.json)
  PR_NUMBER=$(sed -n 's/.*"number":\([0-9][0-9]*\).*/\1/p' /tmp/pr-get.json | head -n 1)
  if [ "$code" != "200" ] || [ -z "$PR_NUMBER" ]; then
    echo "seed: could not read the pull request for $op_branch (HTTP $code)" >&2
    exit 1
  fi
  # A pull request opened while the Actions unit was off has no task; closing and opening it again starts the workflow.
  if [ "$RETRIGGER" = "1" ]; then
    api PATCH "$repo_api/pulls/$PR_NUMBER" /tmp/pr.json -d '{"state":"closed"}' >/dev/null
    code=$(api PATCH "$repo_api/pulls/$PR_NUMBER" /tmp/pr.json -d '{"state":"open"}')
    if [ "$code" != "201" ] && [ "$code" != "200" ]; then
      echo "seed: could not reopen the pull request $PR_NUMBER (HTTP $code)" >&2
      exit 1
    fi
  fi
}

# One branch, one commit per changed file, one pull request per line of prs.tsv. The workflow file goes on every
# branch, so a kept volume's branches run the current workflow. The rows of the start page's table are written
# from the pull requests that really exist, in the order of the file.
ensure_prs() {
  if [ ! -f "$PRS_FILE" ]; then
    echo "seed: $PRS_FILE is not mounted" >&2
    exit 1
  fi
  : >/tmp/pr-rows.html
  CLEAN_COUNT=0
  PR_COUNT=0
  while IFS="$TAB" read -r branch fixture class verdict title says; do
    case $branch in
      '' | \#*) continue ;;
    esac
    if [ -z "$fixture" ] || [ -z "$class" ] || [ -z "$verdict" ] || [ -z "$title" ] || [ -z "$says" ]; then
      echo "seed: a line of $PRS_FILE for $branch has fewer than six fields" >&2
      exit 1
    fi
    ensure_branch "$branch"
    apply_change "$fixture" "$branch" "$title"
    put_file "$branch" .github/workflows/proven.yml /tmp/proven.yml "Update the proven workflow"
    open_pr "$branch" "$title" "$says" "$fixture"
    PR_COUNT=$((PR_COUNT + 1))
    if [ "$verdict" = "clean" ]; then
      CLEAN_COUNT=$((CLEAN_COUNT + 1))
      chip='<span class="v ok">clean</span>'
    else
      chip='<span class="v bad">defect</span>'
    fi
    etitle=$(printf '%s' "$title" | html_escape)
    esays=$(printf '%s' "$says" | html_escape)
    printf '    <tr><td class="n">%s</td><td><a href="%s/%s/%s/pulls/%s" target="_blank" rel="noopener">%s</a></td><td class="muted">%s</td><td>%s</td><td><a class="btn-sm" href="%s/report/%s.html" target="_blank" rel="noopener">Report</a></td></tr>\n' \
      "$PR_NUMBER" "${LAB_URL%/}" "$LAB_ADMIN_USER" "$LAB_DEMO_REPO" "$PR_NUMBER" "$etitle" "$esays" "$chip" "${LAB_PAGE_URL:-/lab}" "$PR_NUMBER" >>/tmp/pr-rows.html
    printf '%s\t%s\t%s\t%s\n' "$PR_NUMBER" "$branch" "$fixture" "$verdict" >>/tmp/pr-numbers.tsv
  done <"$PRS_FILE"
}

wait_ready
read_password
ensure_token
ensure_repo
ensure_actions_unit
ensure_workflow
ensure_bot
ensure_secret
ensure_live_secrets
rm -f "$LAB_SECRETS_FILE"
ensure_runner_token
ensure_base
: >/tmp/pr-numbers.tsv
ensure_prs
cp /tmp/pr-numbers.tsv /state/prs.tsv
# Every pull request gets a report page now, so its Report link is never dead. A page the lab already built for it (a kept
# volume) stays; make up builds the real one from Proven's --md review (lab/scripts/reports.sh).
while IFS="$TAB" read -r report_num _; do
  [ -n "$report_num" ] || continue
  if [ ! -f "$PAGE_DIR/reports/$report_num.html" ]; then
    sh "${REPORT_SH:-/usr/local/bin/report.sh}" stub "$report_num"
  fi
done </state/prs.tsv
render_page
printf '%s\n' ready >/state/ready
echo "seed: demo repository, $PR_COUNT pull requests ($CLEAN_COUNT clean) and the start page are ready"
