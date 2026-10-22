#!/usr/bin/env bash
# Prove that the gates can block.
#
# A gate that never fails is not a gate. This script builds broken fixtures
# in a temporary directory and asserts that the matching gate refuses them.
# It is the negative half of the gate suite: the end-to-end smoke proves the
# platform works, this proves the checks notice when it does not.
set -euo pipefail
cd "$(dirname "$0")/.."

FIXTURE="$(mktemp -d)"
trap 'rm -rf "$FIXTURE"' EXIT

mkdir -p "$FIXTURE/Docs/Features/platform" "$FIXTURE/Docs/Fix"

expect_gate_to_fail() {
  local gate_id="$1"
  local label="$2"
  if go run ./cmd/b-eval gate -root "$FIXTURE" --only "$gate_id" >/dev/null 2>&1; then
    echo "FAIL: the $gate_id gate accepted $label"
    exit 1
  fi
  echo "ok: the $gate_id gate refused $label"
}

# 1. A broken link.
printf '# Broken\n\n[missing](missing.md)\n' > "$FIXTURE/Docs/Features/platform/broken-link.md"
printf '# Index\n\n[FEAT-001](platform/broken-link.md)\n' > "$FIXTURE/Docs/Features/index.md"
printf '# Index\n' > "$FIXTURE/Docs/Fix/index.md"
expect_gate_to_fail docs-links "a broken link"

# 2. Text that is not English.
#
# The fixture is written with byte escapes on purpose: the script itself must
# stay ASCII, because the English gate scans the repository's scripts too. A
# literal non-English string here would make the repository fail its own
# gate.
rm -f "$FIXTURE/Docs/Features/platform/broken-link.md"
printf '# Index\n' > "$FIXTURE/Docs/Features/index.md"
{
  printf '# Non-English\n\n'
  printf '\xe4\xb8\xad\xe6\x96\x87\n'
} > "$FIXTURE/Docs/Features/platform/non-english.md"
expect_gate_to_fail docs-english "non-English text"

# 3. A document missing from its index.
rm -f "$FIXTURE/Docs/Features/platform/non-english.md"
printf '# Index\n' > "$FIXTURE/Docs/Features/index.md"
printf '# Listed\n' > "$FIXTURE/Docs/Features/platform/listed.md"
expect_gate_to_fail docs-index "a document missing from its index"

echo "The gates block as expected."
