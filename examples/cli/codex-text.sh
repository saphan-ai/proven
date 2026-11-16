#!/bin/sh
# Customer-side text transport; never point it at the repository under review.
set -eu
umask 077
export CODEX_HOME="$HOME/.config/proven/codex"
mkdir -p "$CODEX_HOME"
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cd "$work"
set -- "$1"
if [ "$1" = cli-default ]; then set --; else set -- --model "$1"; fi
# Authentication still uses CODEX_HOME; user hooks, MCP, rules and project
# instructions are not loaded. Shell and code-execution tools are disabled.
codex exec --ignore-user-config --ignore-rules --ephemeral --skip-git-repo-check \
  --sandbox read-only --color never --json --output-last-message "$work/answer" \
  -c features.shell_tool=false -c features.unified_exec=false \
  -c features.code_mode=false -c features.js_repl=false \
  -c features.multi_agent=false -c features.apps=false \
  -c features.plugins=false -c features.hooks=false \
  -c approval_policy=never \
  -c features.computer_use=false -c features.browser_use=false \
  -c web_search=disabled "$@" - >/dev/null
# Codex's JSONL is an envelope, not Proven's answer protocol.
test -s "$work/answer"
cat "$work/answer"
