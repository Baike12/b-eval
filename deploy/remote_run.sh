#!/usr/bin/env bash
# Build the runtime image on the server from the shipped binary, then
# (re)start the container. Invoked on the server by scripts/deploy.sh.
#
# The image is built here rather than loaded from a tarball because the
# build context is a single static binary and the base image is already on
# the host. The server therefore performs no registry pull of its own.
set -euo pipefail

: "${TAG:?TAG is required}"

APP_DIR="${APP_DIR:-/opt/b-eval}"
BASE_IMAGE="${B_EVAL_BASE_IMAGE:-alpine:3.20}"
IMAGE="b-eval:${TAG}"

cd "${APP_DIR}"

if [ ! -f dist/b-eval-linux-amd64 ]; then
  echo "dist/b-eval-linux-amd64 is missing in ${APP_DIR}" >&2
  exit 1
fi

echo ">>> Build the image ${IMAGE}"
docker build --build-arg "BASE_IMAGE=${BASE_IMAGE}" -t "${IMAGE}" "${APP_DIR}"

echo ">>> Point the compose project at this tag"
printf 'B_EVAL_TAG=%s\n' "${TAG}" > "${APP_DIR}/.env"

echo ">>> Start the container"
docker compose --project-directory "${APP_DIR}" up -d --force-recreate

echo ">>> Keep the three most recent tags for rollback"
docker images --format '{{.Repository}}:{{.Tag}} {{.CreatedAt}}' b-eval \
  | grep -v ':latest ' \
  | sort -k2 -r \
  | awk 'NR>3 {print $1}' \
  | xargs -r docker rmi -f >/dev/null 2>&1 || true

sleep 2
docker ps --filter name=b-eval --format 'container: {{.Status}} | image: {{.Image}} | ports: {{.Ports}}'
