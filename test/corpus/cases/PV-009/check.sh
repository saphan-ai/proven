#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../../../.."
go run ./tests/corpus/cases/PV-009/check.go
