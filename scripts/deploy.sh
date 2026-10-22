#!/usr/bin/env bash
# Deploy the platform to a server as a container.
#
# The binary is cross-compiled for the target architecture and shipped over
# ssh; the runtime image is then built on the server. The server therefore
# needs neither a Go toolchain nor a Node toolchain, and the transfer is the
# size of the binary rather than the size of an image.
set -euo pipefail
cd "$(dirname "$0")/.."

# The deployment target has no default on purpose. A host address stored in
# version control publishes the server to every reader of the repository, and
# the history keeps the address long after the file is edited. The operator
# names the target through B_EVAL_HOST or --host.
HOST="${B_EVAL_HOST:-}"
TAG="${B_EVAL_TAG:-$(date -u +%Y%m%d-%H%M)}"
APP_DIR="${B_EVAL_APP_DIR:-/opt/b-eval}"
TARGET_GOOS="${B_EVAL_GOOS:-linux}"
TARGET_GOARCH="${B_EVAL_GOARCH:-amd64}"
BASE_PATH="${B_EVAL_BASE_PATH:-/eval}"
BINARY="dist/b-eval-${TARGET_GOOS}-${TARGET_GOARCH}"
SKIP_BUILD=0

usage() {
  cat <<EOF
Usage: bash scripts/deploy.sh [options]

Options:
  --host HOST      ssh target, including the user       (required)
  --tag TAG        image tag and build version          (default: ${TAG})
  --app-dir DIR    application directory on the server  (default: ${APP_DIR})
  --base-path PATH path the platform is served under    (default: ${BASE_PATH})
  --skip-build     ship the binary that is already in dist/
  -h, --help       show this help

Environment: B_EVAL_HOST, B_EVAL_TAG, B_EVAL_APP_DIR, B_EVAL_BASE_PATH,
             B_EVAL_GOOS, B_EVAL_GOARCH
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --host) HOST="$2"; shift 2 ;;
    --tag) TAG="$2"; shift 2 ;;
    --app-dir) APP_DIR="$2"; shift 2 ;;
    --base-path) BASE_PATH="$2"; shift 2 ;;
    --skip-build) SKIP_BUILD=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if [ -z "${HOST}" ]; then
  echo "no deployment target: set B_EVAL_HOST or pass --host" >&2
  usage >&2
  exit 2
fi

if [ "${SKIP_BUILD}" -eq 0 ]; then
  echo ">>> Build the frontend and cross-compile the binary"
  # CGO is disabled because a cross-compiled binary has no C toolchain
  # available, and the platform needs none: it reaches the network through
  # the pure-Go resolver.
  #
  # The base path is baked into the bundle, so the asset URLs and the API
  # paths the browser uses carry it. Vite wants a trailing slash and the
  # nginx location does not, so the two forms are kept apart here.
  VERSION="${TAG}" GOOS="${TARGET_GOOS}" GOARCH="${TARGET_GOARCH}" CGO_ENABLED=0 \
    OUT="${BINARY}" BASE_PATH="${BASE_PATH%/}/" bash scripts/build.sh
else
  echo ">>> Reuse the existing ${BINARY}"
fi

if [ ! -f "${BINARY}" ]; then
  echo "${BINARY} does not exist" >&2
  exit 1
fi

echo ">>> Ship the deployment to ${HOST}"
# The host drops a fraction of the ssh connections it accepts, so every step
# is retried. A dropped connection is a property of the network, not of the
# deployment, and a retry costs five seconds where a failure costs the whole
# shipping step.
retry() {
  local attempt=0
  until "$@"; do
    attempt=$((attempt + 1))
    if [ "${attempt}" -ge 6 ]; then
      echo "giving up after ${attempt} attempts: $*" >&2
      exit 1
    fi
    echo "retry ${attempt}: $*" >&2
    sleep 5
  done
}

retry ssh -o BatchMode=yes -o ConnectTimeout=20 "${HOST}" "mkdir -p '${APP_DIR}/dist' '${APP_DIR}/nginx'"
retry scp -o BatchMode=yes -o ConnectTimeout=20 deploy/Dockerfile deploy/compose.yaml deploy/server_setup.sh deploy/remote_run.sh "${HOST}:${APP_DIR}/"
retry scp -o BatchMode=yes -o ConnectTimeout=20 deploy/nginx/b-eval.conf.template "${HOST}:${APP_DIR}/nginx/"
retry scp -o BatchMode=yes -o ConnectTimeout=20 "${BINARY}" "${HOST}:${APP_DIR}/dist/b-eval-linux-amd64"

echo ">>> Prepare the host and start the container"
retry ssh -o BatchMode=yes -o ConnectTimeout=20 "${HOST}" "cd '${APP_DIR}' && B_EVAL_BASE_PATH='${BASE_PATH}' bash server_setup.sh && TAG='${TAG}' bash remote_run.sh"

echo ">>> Verify the platform through the nginx route"
retry ssh -o BatchMode=yes -o ConnectTimeout=20 "${HOST}" "curl -fsS 'http://127.0.0.1${BASE_PATH}/api/healthz'"
echo
echo "Deployed ${TAG}. The platform is at http://${HOST##*@}${BASE_PATH}/"
