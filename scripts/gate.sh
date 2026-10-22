#!/usr/bin/env bash
# Quality gate pipeline.
#
# The stages run in order and a failing stage stops the pipeline. The
# structure follows bblog's scripts/build.sh and scripts/publish.sh: numbered
# stages, each one a single command, and a failure that blocks the next
# stage. The verdict inside a stage comes from `b-eval gate`, which appends
# its evidence to .gates/evidence.jsonl.
set -euo pipefail
cd "$(dirname "$0")/.."

GATE=(go run ./cmd/b-eval gate -root .)

echo ">>> [1/4] Go format, vet, tests and build"
"${GATE[@]}" --only go-fmt,go-vet,go-test,go-build

echo ">>> [2/4] Frontend type check, tests and build"
"${GATE[@]}" --only web-typecheck,web-test,web-build

echo ">>> [3/4] Documentation gates (English, links, indexes)"
"${GATE[@]}" --only docs-english,docs-links,docs-index

echo ">>> [4/4] End-to-end smoke"
"${GATE[@]}" --only e2e-smoke

echo "All gates passed. Evidence: .gates/evidence.jsonl"
