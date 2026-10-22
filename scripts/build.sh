#!/usr/bin/env bash
# Build the platform: frontend bundle, then the binary with the bundle
# embedded.
#
# The frontend output is copied into internal/web/dist because the Go
# embed directive cannot reach a parent directory. The copy is the only
# reason the directory exists in the source tree at all.
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION="$(date -u +%Y%m%d-%H%M)"

echo ">>> [1/3] Build the frontend"
(
  cd web
  if [ -f package-lock.json ]; then
    npm ci
  else
    npm install
  fi
  npm run build
)

echo ">>> [2/3] Copy the frontend bundle into the embed directory"
rm -rf internal/web/dist
mkdir -p internal/web/dist
cp -R web/dist/. internal/web/dist/

echo ">>> [3/3] Compile the backend"
mkdir -p dist
go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o dist/b-eval ./cmd/b-eval
ls -l dist/b-eval
echo "Build complete: version ${VERSION}"
