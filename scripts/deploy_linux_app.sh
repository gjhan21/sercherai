#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

RUN_DB=false \
RUN_APP=true \
"${ROOT_DIR}/scripts/deploy_linux_server.sh"
