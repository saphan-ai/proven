#!/bin/sh
# Create the demo repository, the workflow, the runner token and the pull request.
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

render_workflow() {
  sed "s|saphan/proven:latest|${PROVEN_IMAGE}:latest|g; s|runs-on: ubuntu-latest|runs-on: ${RUNNER_LABEL}|" \
    /workflow/proven.yml >/tmp/proven.yml
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

ensure_secret() {
  code=$(api PUT "$GITEA_INTERNAL_URL/api/v1/repos/$LAB_ADMIN_USER/$LAB_DEMO_REPO/actions/secrets/PROVEN_TOKEN" /tmp/secret.json \
    -d "{\"data\":\"$API_TOKEN\"}")
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
ensure_workflow
ensure_secret
ensure_runner_token
ensure_branch
ensure_change
ensure_pr
printf '%s\n' ready >/state/ready
echo "seed: demo repository and pull request are ready"
