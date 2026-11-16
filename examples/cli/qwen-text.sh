#!/bin/sh
set -eu
umask 077
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cd "$work"
if [ "$1" = cli-default ]; then set --; else set -- --model "$1"; fi
# Zero is an execution budget, not an empty allowlist (which enables defaults).
# A model trying any tool fails this invocation before executing that tool.
qwen --safe-mode --output-format text --chat-recording=false \
  --max-tool-calls 0 --max-session-turns 1 --advisor off "$@"
