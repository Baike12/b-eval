#!/usr/bin/env bash
# One-time server preparation for the b-eval platform. Idempotent: running
# it again on a prepared host changes nothing.
#
#   1. Create the application directory and the data directory
#   2. Give the data directory to the container's unprivileged user
#   3. Make sure the base image the runtime image is built from is present
#   4. Install the nginx snippet and route the base path to the container
#
# The script does not change the firewall. The platform is reached through
# the host's nginx, on the port nginx already serves.
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/b-eval}"
DATA_DIR="${B_EVAL_DATA_DIR:-${APP_DIR}/data}"
CONTAINER_UID="${B_EVAL_UID:-10001}"
BASE_IMAGE="${B_EVAL_BASE_IMAGE:-alpine:3.20}"
MIRROR_IMAGE="${B_EVAL_MIRROR_IMAGE:-docker.m.daocloud.io/library/alpine:3.20}"
BASE_PATH="${B_EVAL_BASE_PATH:-/eval}"
NGINX_SNIPPET="/etc/nginx/snippets/b-eval.conf"
NGINX_TEMPLATE="${APP_DIR}/nginx/b-eval.conf.template"
ITEACH_CONF="/etc/nginx/sites-available/iteach.conf"

echo ">>> [1/4] Create the application directories"
mkdir -p "${APP_DIR}/dist" "${DATA_DIR}"

echo ">>> [2/4] Hand the data directory to the container user"
# The store creates the record sub-directories itself, but it can only do
# that if it can write to the root of the data directory.
chown -R "${CONTAINER_UID}:${CONTAINER_UID}" "${DATA_DIR}"
chmod 755 "${DATA_DIR}"

echo ">>> [3/4] Provide the base image ${BASE_IMAGE}"
if docker image inspect "${BASE_IMAGE}" >/dev/null 2>&1; then
  echo "    ${BASE_IMAGE} is already present"
else
  # registry-1.docker.io is not reachable from this network, so the base
  # image is fetched through a registry mirror and retagged.
  if ! docker image inspect "${MIRROR_IMAGE}" >/dev/null 2>&1; then
    docker pull "${MIRROR_IMAGE}"
  fi
  docker tag "${MIRROR_IMAGE}" "${BASE_IMAGE}"
  echo "    tagged ${MIRROR_IMAGE} as ${BASE_IMAGE}"
fi

echo ">>> [4/4] Install the nginx route for ${BASE_PATH}"
if [ ! -f "${NGINX_TEMPLATE}" ]; then
  echo "    ${NGINX_TEMPLATE} is missing" >&2
  exit 1
fi
mkdir -p /etc/nginx/snippets
sed "s|{{BASE_PATH}}|${BASE_PATH}|g" "${NGINX_TEMPLATE}" > "${NGINX_SNIPPET}"

if ! grep -q 'snippets/b-eval.conf' "${ITEACH_CONF}"; then
  cp "${ITEACH_CONF}" "${ITEACH_CONF}.bak.$(date -u +%Y%m%d%H%M%S)"
  # The include goes before the last closing brace, which is the end of
  # the server block. Every other line of the site is left untouched.
  awk -v snippet="    include ${NGINX_SNIPPET};" '
    { lines[NR] = $0 }
    END {
      last = NR
      while (last > 0 && lines[last] ~ /^[[:space:]]*$/) last--
      for (i = 1; i <= NR; i++) {
        if (i == last && lines[i] ~ /^[[:space:]]*}/) {
          print snippet
        }
        print lines[i]
      }
    }
  ' "${ITEACH_CONF}" > "${ITEACH_CONF}.tmp"
  mv "${ITEACH_CONF}.tmp" "${ITEACH_CONF}"
  echo "    added the include to ${ITEACH_CONF}"
else
  echo "    the include is already present"
fi

# A broken nginx configuration would take the site's other applications
# down with it, so the reload happens only when the configuration parses.
if ! nginx -t; then
  echo "    nginx rejected the configuration; restoring the backup" >&2
  cp "${ITEACH_CONF}.bak."* "${ITEACH_CONF}"
  nginx -t
  exit 1
fi
systemctl reload nginx
echo "    the route is live at ${BASE_PATH}/"

echo
echo "Prepared ${APP_DIR}. Next step: bash ${APP_DIR}/remote_run.sh with TAG set."
