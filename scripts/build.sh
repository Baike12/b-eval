#!/usr/bin/env bash
# Build the platform: frontend bundle, then the binary with the bundle
# embedded.
#
# The frontend output is copied into internal/web/dist because the Go
# embed directive cannot reach a parent directory. The copy is the only
# reason the directory exists in the source tree at all.
set -euo pipefail
cd "$(dirname "$0")/.."

# The target platform and the output path are overridable so that the same
# script produces the local build and the cross-compiled build the
# deployment ships.
VERSION="${VERSION:-$(date -u +%Y%m%d-%H%M)}"
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
CGO_ENABLED="${CGO_ENABLED:-$(go env CGO_ENABLED)}"
OUT="${OUT:-dist/b-eval}"
# The base path is baked into the bundle, so the API paths and the asset
# URLs it produces carry it. A deployment that serves the platform from a
# prefix sets it; the default serves the platform from the root.
export BASE_PATH="${BASE_PATH:-/}"

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
mkdir -p "$(dirname "${OUT}")"
CGO_ENABLED="${CGO_ENABLED}" GOOS="${GOOS}" GOARCH="${GOARCH}" \
  go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o "${OUT}" ./cmd/b-eval
ls -l "${OUT}"
echo "Build complete: version ${VERSION} (${GOOS}/${GOARCH})"
