#!/bin/sh
# Stand-in service. No network: it answers from a fixed table.
#   service.sh list            prints the models it serves, one per line
#   service.sh answer MODEL    prints the answer text of MODEL
# Two models, one distinct answer text each.
set -eu

case "${1:-}" in
  list)
    echo model-a
    echo model-b
    ;;
  answer)
    case "${2:-}" in
      model-a) echo "answer text from the first model" ;;
      model-b) echo "answer text from the second model" ;;
      *) echo "unknown model" >&2; exit 1 ;;
    esac
    ;;
  *)
    echo "usage: service.sh list|answer MODEL" >&2
    exit 2
    ;;
esac
