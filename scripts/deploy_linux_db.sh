#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

RUN_DB=true \
RUN_APP=false \
SKIP_NGINX=true \
"${ROOT_DIR}/scripts/deploy_linux_server.sh"
