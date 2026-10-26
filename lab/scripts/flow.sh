# The lab's start-up flow (round 3). Sourced by up.sh after lib.sh, with lab/.env already read.
#
# Four things happen here, all before any container starts:
#   1. the flow's settings are validated, and a bad value is refused by name;
#   2. every port the lab publishes is checked free, and a busy one is named with the program that holds it;
#   3. the questions are asked, on a terminal only, for every answer a setting has not given in advance;
#   4. the answers become what the lab runs with: the mode, and the credentials of the path chosen.
# With no terminal and no setting the lab starts in sample mode and asks nothing, so CI and scripts never block.
#
# Settings (environment or lab/.env; a value of "" is the same as not set):
#   PROVEN_LAB_MODE          sample | live
#   PROVEN_LAB_JUDGE         yes | no            (asked only for live)
#   PROVEN_LAB_BACKEND       ours | own          (asked only for a judge; a non-judge is always own)
#   PROVEN_LAB_GATEWAY_AUTH  access | token      (path ours; default access)
#   PROVEN_LAB_OWN_SETUP     yes | no            (path own: run the automated setup)
#   PROVEN_LAB_USE_DETECTED  yes | no            (path own: use the Zetaris and Nightshift already answering here)
#   PROVEN_MARKOVIAN         on | off            Proven's own setting for bounded steps; asked only when not set
#   PROVEN_LAB_SCOUT_INTENT  on | off            the lab's spelling of proven.yaml's scout_intent; asked only when not set
# Neither has a default here: an answer of Enter, and no terminal, leave the setting unset and Proven's own default
# stands. H-MEM and the World Model have no switch in Proven, so they are not asked.
# Test seam: PROVEN_LAB_TTY names a file or FIFO that stands for the terminal. It is not for use by hand.
#
# The gateway credential never reaches a file on the host, an output line, a command line or any container's environment:
# it lives in this shell's environment, goes to curl over a pipe, and reaches the seed in the 0600 file /state/secrets, written
# over stdin (lab_write_secrets in lib.sh), never through Compose. See docs/LAB.md.

LAB_FLOW_BANNER_TITLE='Proven: a pull-request review companion that only says what it can prove (and admits what it could not check)'
LAB_FLOW_BANNER_WHO='Team Saphan.ai: built in public by one person and a fleet of AI agents.'
LAB_FLOW_Q_MODE='Do you want to see Proven in SAMPLE mode (no keys, nothing leaves this machine) or LIVE mode (real models and live data)? [sample/live]'
LAB_FLOW_Q_JUDGE='Are you a judge of the Open Agent Hackathon? [y/n]'
LAB_FLOW_JUDGE_BANNER="For Proven to prove anything live, it needs Zetaris and Meterless Nightshift. Connect your own instances, or use the ones Proven was built with, in Hua Hin, Thailand, under Marcin's desk."
LAB_FLOW_Q_TWEAK='Do you want to tweak Meterless features? [y/N]'
LAB_FLOW_NOTE_ALWAYS='H-MEM (memory of refuted findings) and World Model (claims, evidence and verdicts as a graph) are on by default. Proven has no switch for either yet, so there is nothing to ask about them.'
LAB_FLOW_Q_SCOUT='Scout Intent: plans each review before the agents run (on by default; its benefit is not measured yet). Use it? [y/n, Enter keeps the default]'
LAB_FLOW_Q_MARKOV='Markovian: reviews a large change in bounded steps (on by default). Use it? [y/n, Enter keeps the default]'
LAB_FLOW_BEST_EFFORT='We cannot vouch for results on a setup we do not control: this is best effort.'


# flow_refuse_setting NAME VALUE ALLOWED: a setting that holds something else is refused by name.
flow_refuse_setting() {
  echo "up: $1=$2 is not one of: $3. NEXT — set $1 to one of those, or unset it. Nothing was started." >&2
  exit 1
}

# flow_check_enum NAME ALLOWED...: the setting is empty or one of the words.
flow_check_enum() {
  fe_name=$1
  shift
  eval "fe_val=\${$fe_name:-}"
  [ -n "$fe_val" ] || return 0
  for fe_ok in "$@"; do
    [ "$fe_val" = "$fe_ok" ] && return 0
  done
  fe_list=
  for fe_ok in "$@"; do
    fe_list="${fe_list:+$fe_list, }$fe_ok"
  done
  flow_refuse_setting "$fe_name" "$fe_val" "$fe_list"
}

flow_validate_settings() {
  flow_check_enum PROVEN_LAB_MODE sample live
  flow_check_enum PROVEN_LAB_JUDGE yes no
  flow_check_enum PROVEN_LAB_BACKEND ours own
  flow_check_enum PROVEN_LAB_GATEWAY_AUTH access token
  flow_check_enum PROVEN_LAB_OWN_SETUP yes no
  flow_check_enum PROVEN_LAB_USE_DETECTED yes no
  flow_check_enum PROVEN_MARKOVIAN on off
  flow_check_enum PROVEN_LAB_SCOUT_INTENT on off
}

# ---- ports ------------------------------------------------------------------------------------------------------

# flow_port_busy PORT: exit 0 when something listens on that port on a loopback or wildcard address. FLOW_HOLDER is
# the program's name when the host lets this user read it, and empty otherwise.
flow_port_busy() {
  fp_port=$1
  FLOW_HOLDER=
  fp_busy=1
  if command -v ss >/dev/null 2>&1; then
    fp_line=$(ss -ltnp "sport = :$fp_port" 2>/dev/null | awk '
      NR > 1 {
        a = $4
        sub(/:[0-9]+$/, "", a)
        if (a == "127.0.0.1" || a == "0.0.0.0" || a == "*" || a == "[::]" || a == "[::1]" || a == "::" || a == "::1") { print; exit }
      }')
    if [ -n "$fp_line" ]; then
      fp_busy=0
      FLOW_HOLDER=$(printf '%s\n' "$fp_line" | sed -n 's/.*users:(("\([^"]*\)".*/\1/p')
    fi
  elif command -v lsof >/dev/null 2>&1; then
    fp_name=$(lsof -nP -iTCP:"$fp_port" -sTCP:LISTEN -Fc 2>/dev/null | sed -n 's/^c//p' | head -n 1)
    if [ -n "$fp_name" ]; then
      fp_busy=0
      FLOW_HOLDER=$fp_name
    fi
  fi
  if [ "$fp_busy" -ne 0 ] && ! command -v ss >/dev/null 2>&1 && command -v nc >/dev/null 2>&1; then
    if nc -z 127.0.0.1 "$fp_port" >/dev/null 2>&1; then
      fp_busy=0
    fi
  fi
  if [ "$fp_busy" -eq 0 ] && [ -z "$FLOW_HOLDER" ] && command -v lsof >/dev/null 2>&1; then
    FLOW_HOLDER=$(lsof -nP -iTCP:"$fp_port" -sTCP:LISTEN -Fc 2>/dev/null | sed -n 's/^c//p' | head -n 1)
  fi
  return "$fp_busy"
}

# flow_lab_owns_port PORT: a container of this lab (or of its Zetaris project) already publishes the port, which is
# what a second make up on a running lab looks like. Without docker the answer is no.
flow_lab_owns_port() {
  command -v docker >/dev/null 2>&1 || return 1
  for fo_project in "$COMPOSE_PROJECT_NAME" "${ZETARIS_PROJECT:-}"; do
    [ -n "$fo_project" ] || continue
    fo_ids=$(docker ps -q --filter "label=com.docker.compose.project=$fo_project" --filter "publish=$1" 2>/dev/null || true)
    [ -z "$fo_ids" ] || return 0
  done
  return 1
}

# flow_check_ports NAME...: each setting is a port number and nothing but this lab listens on it. Every busy port is
# named in one refusal, so the judge fixes them in one pass.
flow_check_ports() {
  fc_bad=
  for fc_name in "$@"; do
    eval "fc_port=\${$fc_name:-}"
    case $fc_port in
      '' | *[!0-9]*)
        echo "up: $fc_name=$fc_port is not a port number. NEXT — set $fc_name in lab/versions.env or lab/.env to a free port. Nothing was started." >&2
        exit 1
        ;;
    esac
    if [ "$fc_port" -lt 1 ] || [ "$fc_port" -gt 65535 ]; then
      echo "up: $fc_name=$fc_port is not a port number (1 to 65535). Nothing was started." >&2
      exit 1
    fi
    if flow_port_busy "$fc_port" && ! flow_lab_owns_port "$fc_port"; then
      fc_who=${FLOW_HOLDER:-"a program"}
      fc_bad="${fc_bad}up: port $fc_port ($fc_name) is already in use by $fc_who.
"
    fi
  done
  if [ -n "$fc_bad" ]; then
    printf '%s' "$fc_bad" >&2
    echo "up: the lab publishes on 127.0.0.1 and cannot start on a port that is taken. NEXT — stop that program, or give the lab another port: put NAME=port in lab/.env (the names are in the lines above; the defaults are in lab/versions.env). Nothing was started." >&2
    exit 1
  fi
}

flow_check_lab_ports() {
  flow_check_ports GITEA_HTTP_PORT LGTM_GRAFANA_PORT LGTM_OTLP_PORT LAB_PAGE_PORT LAB_DOCS_PORT
}

# ---- the terminal -----------------------------------------------------------------------------------------------

# flow_open_terminal: exit 0 when there is a terminal to ask on, with it open on descriptor 4. The test is on the
# input: stdin when it is a terminal, else /dev/tty when the process has one. Output being a pipe does not matter.
FLOW_TERMINAL=no
flow_open_terminal() {
  [ "$FLOW_TERMINAL" = yes ] && return 0
  if [ -n "${PROVEN_LAB_TTY:-}" ]; then
    [ -r "$PROVEN_LAB_TTY" ] || return 1
    exec 4<"$PROVEN_LAB_TTY"
  elif [ -t 0 ]; then
    exec 4<&0
  elif ( : </dev/tty ) 2>/dev/null; then
    exec 4</dev/tty
  else
    return 1
  fi
  FLOW_TERMINAL=yes
  return 0
}

flow_close_terminal() {
  if [ "$FLOW_TERMINAL" = yes ]; then
    exec 4<&-
    FLOW_TERMINAL=no
  fi
}

# flow_ask VAR SETTING PROMPT CHOICE...: sets VAR to one of the choices. A wrong answer asks again, three times in
# all, and then stops by name. y and yes, n and no are the same answer. An input that ends is a stop, never a guess.
flow_ask() {
  fa_var=$1
  fa_setting=$2
  fa_prompt=$3
  shift 3
  fa_try=0
  while [ "$fa_try" -lt 3 ]; do
    printf '%s ' "$fa_prompt"
    fa_ans=
    if ! IFS= read -r fa_ans <&4; then
      printf '\n'
      echo "up: the input ended before an answer came. NEXT — answer in advance with $fa_setting. Nothing was started." >&2
      exit 1
    fi
    fa_ans=$(printf '%s' "$fa_ans" | tr 'A-Z' 'a-z' | tr -d ' \t\r')
    # FLOW_ASK_DEFAULT, when set, is what an empty answer means. Only the feature questions set it.
    if [ -z "$fa_ans" ] && [ -n "${FLOW_ASK_DEFAULT:-}" ]; then
      eval "$fa_var=\$FLOW_ASK_DEFAULT"
      return 0
    fi
    case $fa_ans in
      yes) fa_ans=y ;;
      no) fa_ans=n ;;
    esac
    for fa_choice in "$@"; do
      if [ "$fa_ans" = "$fa_choice" ]; then
        eval "$fa_var=\$fa_ans"
        return 0
      fi
    done
    fa_try=$((fa_try + 1))
    if [ "$fa_try" -lt 3 ]; then
      echo "That is not one of: $*. Again, please."
    fi
  done
  fa_list=
  for fa_choice in "$@"; do
    fa_list="${fa_list:+$fa_list, }$fa_choice"
  done
  echo "up: three answers in a row were not one of: $fa_list. NEXT — answer in advance with $fa_setting. Nothing was started." >&2
  exit 1
}

# flow_restore_echo: put the terminal back as it was found. Safe to call twice.
FLOW_STTY=
flow_restore_echo() {
  if [ -n "$FLOW_STTY" ]; then
    stty "$FLOW_STTY" <&4 2>/dev/null || true
    FLOW_STTY=
  fi
}

# flow_read_secret VAR PROMPT: reads one line with no echo into VAR. The terminal is restored on every way out:
# the answer, an end of input, Ctrl-C, a hang-up, a termination.
flow_read_secret() {
  fr_var=$1
  printf '%s ' "$2"
  FLOW_STTY=$(stty -g <&4 2>/dev/null || true)
  trap 'flow_restore_echo; echo; exit 130' INT
  trap 'flow_restore_echo; echo; exit 143' TERM HUP
  trap 'flow_restore_echo' EXIT
  [ -z "$FLOW_STTY" ] || stty -echo <&4 2>/dev/null || true
  fr_line=
  fr_rc=0
  IFS= read -r fr_line <&4 || fr_rc=$?
  flow_restore_echo
  trap - INT TERM HUP EXIT
  echo
  if [ "$fr_rc" -ne 0 ] && [ -z "$fr_line" ]; then
    echo "up: the input ended before the token came. NEXT — put it in lab/.env as JUDGES_GATEWAY_TOKEN. Nothing was started." >&2
    exit 1
  fi
  eval "$fr_var=\$fr_line"
  fr_line=
}

# ---- the gateway ------------------------------------------------------------------------------------------------

# flow_secret_ok VALUE: not empty, and nothing but LAB_FLOW_SECRET_CHARS.
flow_secret_ok() {
  case $1 in
    '' | *[!A-Za-z0-9._~+/=-]*) return 1 ;;
  esac
  return 0
}

# flow_gateway_url: JUDGES_GATEWAY_URL is set, or asked for on a terminal. https, or http on loopback (the tests).
flow_gateway_url() {
  if [ -z "${JUDGES_GATEWAY_URL:-}" ]; then
    if flow_open_terminal; then
      printf 'The gateway address you were sent (https://...): '
      # An input that ends is a stop that says so, never a guess at an address (a last line with no newline still counts).
      fg_rc=0
      IFS= read -r JUDGES_GATEWAY_URL <&4 || fg_rc=$?
      if [ "$fg_rc" -ne 0 ] && [ -z "$JUDGES_GATEWAY_URL" ]; then
        printf '\n'
        echo "up: the input ended before the gateway address came. NEXT — put it in lab/.env as JUDGES_GATEWAY_URL. Nothing was started." >&2
        exit 1
      fi
    else
      echo "up: the judges' gateway address is not set. NEXT — put the address you were sent in lab/.env as JUDGES_GATEWAY_URL. Nothing was started." >&2
      exit 1
    fi
  fi
  case $JUDGES_GATEWAY_URL in
    https://* | http://127.0.0.1[:/]* | http://localhost[:/]*) ;;
    *)
      echo "up: JUDGES_GATEWAY_URL is not an https address. NEXT — use the address you were sent, which starts with https://. Nothing was started." >&2
      exit 1
      ;;
  esac
  case $JUDGES_GATEWAY_URL in
    *[!A-Za-z0-9._~:/%@-]*)
      echo "up: JUDGES_GATEWAY_URL holds a character an address does not. Nothing was started." >&2
      exit 1
      ;;
  esac
  JUDGES_GATEWAY_URL=${JUDGES_GATEWAY_URL%/}
  export JUDGES_GATEWAY_URL
}

# flow_token_expiry TOKEN: the expiry a JWT names, as a date, or nothing when it cannot be read.
flow_token_expiry() {
  ft_payload=$(printf '%s' "$1" | cut -d. -f2 | tr '_-' '/+')
  [ -n "$ft_payload" ] || return 0
  while [ $((${#ft_payload} % 4)) -ne 0 ]; do
    ft_payload="$ft_payload="
  done
  ft_json=$(printf '%s' "$ft_payload" | base64 -d 2>/dev/null || printf '%s' "$ft_payload" | base64 -D 2>/dev/null || true)
  ft_exp=$(printf '%s' "$ft_json" | sed -n 's/.*"exp"[[:space:]]*:[[:space:]]*\([0-9][0-9]*\).*/\1/p')
  [ -n "$ft_exp" ] || return 0
  date -u -d "@$ft_exp" '+%Y-%m-%d %H:%M UTC' 2>/dev/null || date -u -r "$ft_exp" '+%Y-%m-%d %H:%M UTC' 2>/dev/null || true
}

# flow_gateway_access: Cloudflare Access with the e-mail one-time PIN. cloudflared signs the judge in, in their
# browser, and hands back an application token. The token is read from cloudflared's output into a variable and goes
# nowhere else but the check and the secret store.
flow_gateway_access() {
  if ! command -v cloudflared >/dev/null 2>&1; then
    echo "up: the Access sign-in needs cloudflared, and it is not installed. NEXT — install it: macOS: brew install cloudflared; Linux and Windows: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/ . Then run make up again. Or use a token instead: PROVEN_LAB_GATEWAY_AUTH=token. Nothing was started." >&2
    exit 1
  fi
  echo "up: opening the sign-in for $JUDGES_GATEWAY_URL. Cloudflare sends a one-time PIN to your e-mail; enter it in the browser."
  # The login prints the token on its standard output in some versions: that stream is discarded here. Its standard
  # error carries the sign-in address for a machine whose browser did not open.
  if ! cloudflared access login "$JUDGES_GATEWAY_URL" >/dev/null; then
    echo "up: the Access sign-in did not complete. NEXT — run make up again, and enter the PIN Cloudflare e-mailed to you. Nothing was started." >&2
    exit 1
  fi
  fg_token=$(cloudflared access token -app="$JUDGES_GATEWAY_URL" 2>/dev/null || true)
  if ! flow_secret_ok "$fg_token"; then
    fg_token=
    echo "up: cloudflared signed you in but returned no usable token for $JUDGES_GATEWAY_URL. NEXT — run make up again; if it repeats, the address is not the one Access protects. Nothing was started." >&2
    exit 1
  fi
  JUDGES_GATEWAY_ACCESS_TOKEN=$fg_token
  export JUDGES_GATEWAY_ACCESS_TOKEN
  FLOW_CRED_HEADER=cf-access-token
  FLOW_CRED_VALUE=$fg_token
  fg_exp=$(flow_token_expiry "$fg_token")
  fg_token=
  if [ -n "$fg_exp" ]; then
    echo "up: signed in. The token expires $fg_exp. To renew it, run make up again: it signs you in again and stores a new one."
  else
    echo "up: signed in. This token does not say when it expires; Cloudflare Access ends it when your session ends. To renew it, run make up again."
  fi
}

# flow_gateway_token: the per-judge token, from the settings, or asked for with no echo.
flow_gateway_token() {
  if [ -z "${JUDGES_GATEWAY_TOKEN:-}" ]; then
    if ! flow_open_terminal; then
      echo "up: the judge token is not set and there is no terminal to ask on. NEXT — put the token you were sent in lab/.env as JUDGES_GATEWAY_TOKEN. Nothing was started." >&2
      exit 1
    fi
    ftk_try=0
    while [ "$ftk_try" -lt 3 ]; do
      flow_read_secret JUDGES_GATEWAY_TOKEN 'The judge token we sent you (nothing is shown as you type):'
      if flow_secret_ok "$JUDGES_GATEWAY_TOKEN"; then
        break
      fi
      JUDGES_GATEWAY_TOKEN=
      ftk_try=$((ftk_try + 1))
      if [ "$ftk_try" -lt 3 ]; then
        echo "That is not a token this lab can send (letters, digits and . _ ~ + / = - only). Again, please."
      fi
    done
    if [ -z "$JUDGES_GATEWAY_TOKEN" ]; then
      echo "up: three tries and no usable token. NEXT — copy the token again from the message we sent you. Nothing was started." >&2
      exit 1
    fi
  elif ! flow_secret_ok "$JUDGES_GATEWAY_TOKEN"; then
    echo "up: JUDGES_GATEWAY_TOKEN holds a character this lab does not send (letters, digits and . _ ~ + / = - only). Nothing was started." >&2
    exit 1
  fi
  export JUDGES_GATEWAY_TOKEN
  FLOW_CRED_HEADER=Authorization
  FLOW_CRED_VALUE="Bearer $JUDGES_GATEWAY_TOKEN"
}

# flow_check_gateway: one request that authenticates and spends nothing: an evidence-route request with an empty
# JSON object. Credentials the gateway accepts get 400 not_a_template_request; refused ones get 401 or 403. The
# credential goes to curl as a config on its standard input, and printf is a builtin, so no process lists it.
flow_check_gateway() {
  fk_url=$JUDGES_GATEWAY_URL$JUDGES_GATEWAY_ZETARIS_PATH
  if ! command -v curl >/dev/null 2>&1; then
    echo "up: checking the gateway needs curl, and it is not installed. NEXT — install curl, then run make up again. Nothing was started." >&2
    exit 1
  fi
  fk_out=$(printf 'header = "%s: %s"\n' "$FLOW_CRED_HEADER" "$FLOW_CRED_VALUE" |
    curl -sS --max-time "${JUDGES_GATEWAY_CHECK_SECONDS:-20}" -K - -X POST -H 'Content-Type: application/json' \
      --data '{}' -w '\n%{http_code}' "$fk_url" 2>/dev/null || true)
  FLOW_CRED_VALUE=
  fk_code=$(printf '%s\n' "$fk_out" | tail -n 1)
  fk_body=$(printf '%s\n' "$fk_out" | sed '$d' | head -c 400)
  fk_name=$(printf '%s' "$fk_body" | sed -n 's/.*"error"[[:space:]]*:[[:space:]]*"\([A-Za-z0-9_ -]*\)".*/\1/p' | head -n 1)
  case $fk_code in
    400)
      case $fk_body in
        *not_a_template_request*)
          echo "up: the gateway accepted your credentials: ready."
          return 0
          ;;
      esac
      echo "up: the gateway answered 400 but not not_a_template_request, so these credentials are not confirmed. NEXT — check JUDGES_GATEWAY_URL is the address you were sent. Nothing was started." >&2
      ;;
    401 | 403)
      echo "up: the gateway refused your credentials: HTTP $fk_code${fk_name:+ ($fk_name)}. NEXT — a token that is unknown, expired or revoked, or an Access sign-in that was refused, is replaced by the organisers: ask them for a new one, or run make up again to sign in again. Nothing was started." >&2
      ;;
    '' | 000)
      echo "up: the gateway did not answer at $JUDGES_GATEWAY_URL. NEXT — check the address and your network, then run make up again. Nothing was started." >&2
      ;;
    *)
      echo "up: the gateway answered HTTP $fk_code${fk_name:+ ($fk_name)}, which is not an answer to a check. Your credentials are not confirmed. Nothing was started." >&2
      ;;
  esac
  exit 1
}

flow_path_ours() {
  flow_gateway_url
  FLOW_CRED_HEADER=
  FLOW_CRED_VALUE=
  # One method per run: a credential of the other left in lab/.env must not outrank this one in the workflow, and an
  # unset one makes the seed remove the secret stored earlier.
  case ${PROVEN_LAB_GATEWAY_AUTH:-access} in
    token)
      unset JUDGES_GATEWAY_ACCESS_TOKEN
      flow_gateway_token
      ;;
    *)
      unset JUDGES_GATEWAY_TOKEN
      flow_gateway_access
      ;;
  esac
  flow_check_gateway
  LAB_MODE="live via our gateway"
}

# ---- your own setup ---------------------------------------------------------------------------------------------

# flow_origin URL HOST: URL with its host replaced by HOST, which is how the lab's review job reaches a service that
# runs on this machine. The job runs inside the lab's own Docker, where 127.0.0.1 is the job itself.
flow_rehost() {
  printf '%s' "$1" | sed "s|^\(https\{0,1\}://\)[^/:]*|\1$2|"
}

# flow_probe KIND URL: exit 0 when the service named KIND answers at URL, and answers in its own shape. Any other
# listener (a stray web server on the port, a proxy's error page) is not that service. The shapes are claims about
# the services, read from Proven's own adapters and not measured against a running Zetaris (docs/LAB.md says so):
#   nightshift  GET /v1/models: 200 and a JSON list of models ({"data":[...]}), or 401 and a JSON object, which is
#               what its runtime answers without its token;
#   zetaris     its API root: 200, 400, 401 or 403 and a JSON object (never 404, never HTML).
flow_probe() {
  fq_kind=$1
  fq_url=$2
  command -v curl >/dev/null 2>&1 || return 1
  fq_out=$(curl -sS --max-time 3 -w '\n%{http_code}' "$fq_url" 2>/dev/null || true)
  fq_code=$(printf '%s\n' "$fq_out" | tail -n 1)
  fq_flat=$(printf '%s\n' "$fq_out" | sed '$d' | head -c 8192 | tr -d ' \t\r\n')
  case $fq_kind in
    nightshift)
      case $fq_code in
        200)
          case $fq_flat in
            '{'*'"data":['*) return 0 ;;
          esac
          ;;
        401)
          case $fq_flat in
            '{'*'}') return 0 ;;
          esac
          ;;
      esac
      ;;
    zetaris)
      case $fq_code in
        200 | 400 | 401 | 403)
          case $fq_flat in
            '{'*'}') return 0 ;;
          esac
          ;;
      esac
      ;;
  esac
  return 1
}

flow_detect_own() {
  fd_z=no
  fd_n=no
  flow_probe zetaris "$DETECT_ZETARIS_URL" && fd_z=yes
  flow_probe nightshift "$DETECT_NIGHTSHIFT_URL" && fd_n=yes
  if [ "$fd_z" = no ] && [ "$fd_n" = no ]; then
    echo "up: neither Zetaris ($DETECT_ZETARIS_URL) nor Nightshift ($DETECT_NIGHTSHIFT_URL) answers on this machine."
    return 0
  fi
  [ "$fd_z" = yes ] && echo "up: Zetaris already answers on this machine: $DETECT_ZETARIS_URL"
  [ "$fd_n" = yes ] && echo "up: Nightshift already answers on this machine: $DETECT_NIGHTSHIFT_URL"
  fd_use=${PROVEN_LAB_USE_DETECTED:-}
  if [ -z "$fd_use" ]; then
    if flow_open_terminal; then
      flow_ask fd_use PROVEN_LAB_USE_DETECTED 'Use what is already running, as it is? [y/n]' y n
      [ "$fd_use" = y ] && fd_use=yes || fd_use=no
    else
      fd_use=no
    fi
  fi
  [ "$fd_use" = yes ] || return 0
  if [ "$fd_z" = yes ] && [ -z "${ZETARIS_REST_URL:-}" ]; then
    ZETARIS_REST_URL=$(flow_rehost "$DETECT_ZETARIS_URL" "$LAB_JOB_HOST_ALIAS")
    export ZETARIS_REST_URL
  fi
  if [ "$fd_n" = yes ] && [ -z "${PROVEN_GENERATOR_BASE_URL:-}" ]; then
    PROVEN_GENERATOR_BASE_URL=$(flow_rehost "$DETECT_NIGHTSHIFT_URL" "$LAB_JOB_HOST_ALIAS" | sed 's|^\(https\{0,1\}://[^/]*\).*|\1|')
    export PROVEN_GENERATOR_BASE_URL
  fi
  echo "up: the lab's review job reaches them as $LAB_JOB_HOST_ALIAS. The keys stay yours: METERLESS_RUNTIME_TOKEN, ZETARIS_API_KEY and ZETARIS_ORG_ID come from lab/.env, with the names your guide uses."
  echo "up: if a service listens on 127.0.0.1 only, the lab's job cannot reach it: it must listen on an address the lab's Docker can reach."
}

flow_path_own() {
  echo "up: your own setup. The guide is docs/LAB.md, the section \"Go live\" (http://127.0.0.1:$LAB_DOCS_PORT/ once the lab is up)."
  echo "up: Nightshift: install the app from meterless.ai (Windows, Mac M1+, Mac M5+, DGX). Zetaris: the bundle from your own account. Our automation (Zetaris from your bundle, Nightshift in a Tart VM) is optional."
  flow_detect_own
  fo_setup=${PROVEN_LAB_OWN_SETUP:-}
  if [ -z "$fo_setup" ]; then
    if flow_open_terminal; then
      flow_ask fo_setup PROVEN_LAB_OWN_SETUP 'Run the automated setup now (make up ZETARIS=local, then make nightshift)? [y/n]' y n
      [ "$fo_setup" = y ] && fo_setup=yes || fo_setup=no
    else
      fo_setup=no
    fi
  fi
  if [ "$fo_setup" = yes ]; then
    ZETARIS=local
    export ZETARIS
    flow_check_ports ZETARIS_UI_HOST_PORT ZETARIS_API_HOST_PORT
    echo "up: running the automated setup: Zetaris first (it starts with the lab), then Nightshift."
    if ! sh "$ROOT/lab/scripts/nightshift.sh"; then
      echo "up: make nightshift did not complete (see above); the lab goes on." >&2
    fi
  fi
  if [ -z "${PROVEN_GENERATOR_KEY:-}" ] && [ -z "${METERLESS_RUNTIME_TOKEN:-}" ]; then
    echo "up: no key or token is set yet, so the first reviews will be sample reviews (the start page says so). NEXT — add yours to lab/.env, then make up again."
  fi
  echo "$LAB_FLOW_BEST_EFFORT"
  LAB_MODE="live on your own setup"
}

# ---- the flow ---------------------------------------------------------------------------------------------------

# flow_drop_keys: sample mode uses no key. The names a live review reads are removed from this run's environment, so
# the seed removes any secret stored earlier and the lab is a sample lab, whatever lab/.env holds.
flow_drop_keys() {
  fd_dropped=
  for fd_name in $LAB_LIVE_KEYS; do
    # The Meterless features are settings of a review, not keys: sample mode keeps them.
    case " ${LAB_FEATURE_KEYS:-} " in
      *" $fd_name "*) continue ;;
    esac
    eval "fd_val=\${$fd_name:-}"
    if [ -n "$fd_val" ]; then
      fd_dropped="${fd_dropped:+$fd_dropped, }$fd_name"
    fi
    unset "$fd_name"
  done
  unset ZETARIS
  if [ -n "$fd_dropped" ]; then
    echo "up: sample mode does not use these settings: $fd_dropped. For a live review set PROVEN_LAB_MODE=live."
  fi
}

# ---- the Meterless features -------------------------------------------------------------------------------------

# flow_feature_answer VAR PROMPT: one feature. A setting given in advance is left as it is. Otherwise y is on,
# n is off and Enter leaves the setting unset, so Proven's own default stands.
flow_feature_answer() {
  ff_var=$1
  eval "ff_cur=\${$ff_var:-}"
  [ -z "$ff_cur" ] || return 0
  FLOW_ASK_DEFAULT=keep
  flow_ask ff_ans "$ff_var" "$2" y n
  unset FLOW_ASK_DEFAULT
  case $ff_ans in
    y) eval "$ff_var=on" ;;
    n) eval "$ff_var=off" ;;
    *) return 0 ;;
  esac
  export "$ff_var"
}

# flow_features: after sample or live. Asked on a terminal only, and only for a feature whose setting is not given.
# The owner's 2026-10-26 ruling: all four Meterless engines are on by default, in Proven and in the lab. The lab
# therefore passes on for any switch that no setting and no answer turned off, instead of leaving Proven's own default.
flow_features() {
  flow_features_ask
  : "${PROVEN_MARKOVIAN:=on}" "${PROVEN_LAB_SCOUT_INTENT:=on}"
  export PROVEN_MARKOVIAN PROVEN_LAB_SCOUT_INTENT
}

# No terminal and no setting asks nothing; flow_features then passes the defaults.
flow_features_ask() {
  if [ "$FLOW_ASK" != yes ]; then
    return 0
  fi
  if [ -n "${PROVEN_MARKOVIAN:-}" ] && [ -n "${PROVEN_LAB_SCOUT_INTENT:-}" ]; then
    return 0
  fi
  FLOW_ASK_DEFAULT=n
  flow_ask ff_tweak "PROVEN_MARKOVIAN and PROVEN_LAB_SCOUT_INTENT (on or off)" "$LAB_FLOW_Q_TWEAK" y n
  unset FLOW_ASK_DEFAULT
  [ "$ff_tweak" = y ] || return 0
  echo "$LAB_FLOW_NOTE_ALWAYS"
  flow_feature_answer PROVEN_LAB_SCOUT_INTENT "$LAB_FLOW_Q_SCOUT"
  flow_feature_answer PROVEN_MARKOVIAN "$LAB_FLOW_Q_MARKOV"
}

flow_banner() {
  echo
  echo "$LAB_FLOW_BANNER_TITLE"
  echo "$LAB_FLOW_BANNER_WHO"
  echo "No humans were harmed. Several agents were asked to try again."
  echo
}

# lab_flow: the whole of it. Leaves LAB_MODE exported: "sample", "live via our gateway" or "live on your own setup".
lab_flow() {
  flow_validate_settings
  flow_check_lab_ports

  FLOW_ASK=no
  if flow_open_terminal; then
    FLOW_ASK=yes
    flow_banner
  fi

  fl_mode=${PROVEN_LAB_MODE:-}
  if [ -z "$fl_mode" ]; then
    if [ "$FLOW_ASK" = yes ]; then
      flow_ask fl_mode PROVEN_LAB_MODE "$LAB_FLOW_Q_MODE" sample live
    else
      fl_mode=sample
    fi
  fi

  flow_features

  if [ "$fl_mode" = sample ]; then
    LAB_MODE=sample
    flow_drop_keys
    echo "up: sample mode: recorded answers, no keys, nothing leaves this machine."
  else
    fl_judge=${PROVEN_LAB_JUDGE:-}
    if [ -z "$fl_judge" ]; then
      if [ "$FLOW_ASK" = yes ]; then
        flow_ask fl_judge PROVEN_LAB_JUDGE "$LAB_FLOW_Q_JUDGE" y n
        [ "$fl_judge" = y ] && fl_judge=yes || fl_judge=no
      else
        fl_judge=no
      fi
    fi
    fl_backend=own
    if [ "$fl_judge" = yes ]; then
      fl_backend=${PROVEN_LAB_BACKEND:-}
      if [ -z "$fl_backend" ]; then
        if [ "$FLOW_ASK" = yes ]; then
          echo "$LAB_FLOW_JUDGE_BANNER"
          flow_ask fl_backend PROVEN_LAB_BACKEND 'Which do you want? [ours/own]' ours own
        else
          echo "up: PROVEN_LAB_JUDGE=yes needs PROVEN_LAB_BACKEND=ours or own, and there is no terminal to ask on. Nothing was started. NEXT — set PROVEN_LAB_BACKEND=ours or own, or run make up on a terminal." >&2
          exit 1
        fi
      fi
    fi
    if [ "$fl_backend" = ours ]; then
      flow_path_ours
    else
      flow_path_own
    fi
    echo "up: live mode: $LAB_MODE."
  fi
  export LAB_MODE
  flow_close_terminal
}
