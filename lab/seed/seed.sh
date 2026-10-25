#!/bin/sh
# Create the demo repository, the workflow, the workflow's own account, the runner token and the pull request.
# The admin user already exists. The password arrives in LAB_ADMIN_PASSWORD.
set -eu

if [ "$(id -u)" -eq 0 ]; then
  mkdir -p /state /reg
  chown proven:proven /state /reg
  chmod 700 /state /reg
  exec su-exec proven "$0" "$@"
fi

: "${GITEA_INTERNAL_URL:?}"
: "${LAB_ADMIN_USER:?}"
: "${LAB_ADMIN_PASSWORD:?}"
: "${LAB_BOT_USER:?}"
: "${LAB_BOT_EMAIL:?}"
: "${LAB_DEMO_REPO:?}"
: "${LAB_DEMO_BASE:?}"
: "${LAB_DEMO_HEAD:?}"
: "${PROVEN_IMAGE:?}"
: "${RUNNER_LABEL:?}"

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

save_password() {
  if [ ! -s /state/password ]; then
    printf '%s' "$LAB_ADMIN_PASSWORD" >/state/password
    chmod 600 /state/password
  fi
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
  sed "s|saphan/proven:latest|${PROVEN_IMAGE}:latest|g; s|runs-on: ubuntu-latest|runs-on: ${RUNNER_LABEL}|; s|\${{ github.api_url }}|${GITEA_INTERNAL_URL}/api/v1|" \
    /workflow/proven.yml >/tmp/proven.yml
  # ADR-0045: the review in the lab is exported to the lab's collector. The job's `docker run` starts
  # Proven on the inner daemon, whose default network cannot resolve the lab's service names, so the
  # container joins the daemon's own network, which is on the lab network. Proven reads the standard
  # OTEL_* variables and nothing else. Without an endpoint the workflow is left as it is.
  if [ -n "${OTEL_EXPORTER_OTLP_ENDPOINT:-}" ]; then
    sed "s|docker run --rm --entrypoint sh|docker run --rm --network host -e OTEL_EXPORTER_OTLP_ENDPOINT=${OTEL_EXPORTER_OTLP_ENDPOINT} -e OTEL_SERVICE_NAME=${OTEL_SERVICE_NAME:-proven} -e OTEL_METRIC_EXPORT_INTERVAL=${OTEL_METRIC_EXPORT_INTERVAL:-1000} --entrypoint sh|" \
      /tmp/proven.yml >/tmp/proven.otel
    mv /tmp/proven.otel /tmp/proven.yml
  fi
  # ADR-0009, ADR-0045: the lab's review runs a second reviewer model, so the dashboard shows two.
  # The sample generator answers for both, so the findings do not change; the review names the second reviewer.
  # ADR-0050 point 6: the one automated fix attempt is on in the judges' lab and off everywhere else, so the
  # lab's review runs with a proven.yaml that turns it on, with or without the second reviewer. The command
  # sits inside single quotes in the workflow, so the file is written with echo and no quote of its own.
  second=""
  if [ -n "${LAB_SPECIAL_REVIEWER:-}" ]; then
    second=" \&\& echo special_reviewer: ${LAB_SPECIAL_REVIEWER} >> lab.yaml"
  fi
  sed "s|proven scan --sample /opt|echo autofix: on > lab.yaml${second} \&\& proven scan --sample --config lab.yaml /opt|" \
    /tmp/proven.yml >/tmp/proven.two
  mv /tmp/proven.two /tmp/proven.yml
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
  tokens="$GITEA_INTERNAL_URL/api/v1/users/$LAB_BOT_USER/tokens"
  curl -sS --max-time 30 -o /dev/null -X DELETE -u "$LAB_BOT_USER:$bot_pass" "$tokens/lab-workflow" || true
  code=$(curl -sS --max-time 30 -o /tmp/bot-token.json -w '%{http_code}' \
    -u "$LAB_BOT_USER:$bot_pass" \
    -H 'Content-Type: application/json' \
    -d '{"name":"lab-workflow","scopes":["write:issue","write:repository"]}' \
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

ensure_branch() {
  code=$(api POST "$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/branches" /tmp/branch.json \
    -d "{\"new_branch_name\":\"$LAB_DEMO_HEAD\",\"old_ref_name\":\"$LAB_DEMO_BASE\"}")
  if [ "$code" != "201" ] && [ "$code" != "409" ]; then
    code=$(api POST "$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/branches" /tmp/branch.json \
      -d "{\"new_branch_name\":\"$LAB_DEMO_HEAD\",\"old_branch_name\":\"$LAB_DEMO_BASE\"}")
  fi
  if [ "$code" != "201" ] && [ "$code" != "409" ]; then
    echo "seed: could not create the branch (HTTP $code)" >&2
    exit 1
  fi
}

ensure_change() {
  code=$(curl -sS --max-time 30 -o /tmp/readme.json -w '%{http_code}' \
    -H "Authorization: token $API_TOKEN" \
    "$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/contents/README.md?ref=$LAB_DEMO_BASE")
  if [ "$code" != "200" ]; then
    echo "seed: could not read README.md (HTTP $code)" >&2
    exit 1
  fi
  sha=$(json_string sha /tmp/readme.json)
  note='The lab changes this line so the pull request has a diff. The workflow runs a sample review.'
  b64=$(printf '%s\n' "# discount" "" "$note" | base64 | tr -d '\n')
  body=$(printf '{"message":"Open a change for the sample review","branch":"%s","sha":"%s","content":"%s"}' \
    "$LAB_DEMO_HEAD" "$sha" "$b64")
  code=$(api PUT "$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/contents/README.md" /tmp/put.json \
    -d "$body")
  if [ "$code" != "200" ] && [ "$code" != "201" ] && [ "$code" != "422" ]; then
    echo "seed: could not write the change (HTTP $code)" >&2
    exit 1
  fi
}

ensure_pr() {
  code=$(curl -sS --max-time 30 -o /tmp/pulls.json -w '%{http_code}' \
    -H "Authorization: token $API_TOKEN" \
    "$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/pulls?state=open")
  if [ "$code" = "200" ] && grep -q '"number":' /tmp/pulls.json; then
    # The pull request exists. If the Actions unit was off when it opened,
    # it has no task; closing and opening it again starts the workflow.
    if [ "$RETRIGGER" = "1" ]; then
      num=$(sed -n 's/.*"number":\([0-9][0-9]*\).*/\1/p' /tmp/pulls.json | head -n 1)
      pr="$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/pulls/$num"
      api PATCH "$pr" /tmp/pr.json -d '{"state":"closed"}' >/dev/null
      code=$(api PATCH "$pr" /tmp/pr.json -d '{"state":"open"}')
      if [ "$code" != "201" ] && [ "$code" != "200" ]; then
        echo "seed: could not reopen the pull request (HTTP $code)" >&2
        exit 1
      fi
    fi
    return 0
  fi
  code=$(api POST "$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/pulls" /tmp/pr.json \
    -d "{\"title\":\"Review the sample change\",\"head\":\"$LAB_DEMO_HEAD\",\"base\":\"$LAB_DEMO_BASE\",\"body\":\"The lab opens this pull request so proven can comment.\"}")
  if [ "$code" != "201" ] && [ "$code" != "409" ]; then
    echo "seed: could not open the pull request (HTTP $code)" >&2
    exit 1
  fi
}

wait_ready
save_password
ensure_token
ensure_repo
ensure_actions_unit
ensure_workflow
ensure_bot
ensure_secret
ensure_runner_token
ensure_branch
ensure_change
ensure_pr
printf '%s\n' ready >/state/ready
echo "seed: demo repository and pull request are ready"
