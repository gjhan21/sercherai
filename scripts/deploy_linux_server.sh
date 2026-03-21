#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEPLOY_DIR="${ROOT_DIR}/deploy/linux"

APP_DIR="${APP_DIR:-/opt/sercherai}"
STRATEGY_ENGINE_APP_DIR="${STRATEGY_ENGINE_APP_DIR:-${APP_DIR}/strategy-engine}"
WWW_DIR="${WWW_DIR:-/var/www/sercherai}"
SERVICE_NAME="${SERVICE_NAME:-sercherai-backend}"
STRATEGY_ENGINE_SERVICE_NAME="${STRATEGY_ENGINE_SERVICE_NAME:-sercherai-strategy-engine}"
SERVICE_USER="${SERVICE_USER:-${SUDO_USER:-$(id -un)}}"
SERVICE_GROUP="${SERVICE_GROUP:-}"
BACKEND_ENV_FILE="${BACKEND_ENV_FILE:-/etc/sercherai/backend.env}"
STRATEGY_ENGINE_ENV_FILE="${STRATEGY_ENGINE_ENV_FILE:-/etc/sercherai/strategy-engine.env}"
SYSTEMD_FILE="${SYSTEMD_FILE:-/etc/systemd/system/${SERVICE_NAME}.service}"
STRATEGY_ENGINE_SYSTEMD_FILE="${STRATEGY_ENGINE_SYSTEMD_FILE:-/etc/systemd/system/${STRATEGY_ENGINE_SERVICE_NAME}.service}"
NGINX_FILE="${NGINX_FILE:-/etc/nginx/conf.d/sercherai.conf}"

BACKEND_PORT="${BACKEND_PORT:-18080}"
STRATEGY_ENGINE_PORT="${STRATEGY_ENGINE_PORT:-18081}"
CLIENT_PORT="${CLIENT_PORT:-80}"
ADMIN_PORT="${ADMIN_PORT:-8081}"

MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-sercherai}"
MYSQL_PWD="${MYSQL_PWD:-wbdE4xkwew2TaNJL}"
MYSQL_DB="${MYSQL_DB:-sercherai}"

RUN_SEED="${RUN_SEED:-false}"
SKIP_NGINX="${SKIP_NGINX:-false}"
RUN_DB="${RUN_DB:-true}"
RUN_APP="${RUN_APP:-true}"

ROOT_CMD=()
if [ "${EUID}" -ne 0 ]; then
	ROOT_CMD=(sudo)
fi

run_root() {
	if [ "${#ROOT_CMD[@]}" -eq 0 ]; then
		"$@"
	else
		"${ROOT_CMD[@]}" "$@"
	fi
}

require_cmd() {
	local cmd="$1"
	if ! command -v "${cmd}" >/dev/null 2>&1; then
		echo "missing required command: ${cmd}" >&2
		exit 1
	fi
}

is_true() {
	case "$1" in
	1 | true | TRUE | yes | YES | on | ON)
		return 0
		;;
	*)
		return 1
		;;
	esac
}

copy_dir_without_hidden() {
	local src_dir="$1"
	local dest_dir="$2"

	run_root mkdir -p "${dest_dir}"
	if [ "${#ROOT_CMD[@]}" -eq 0 ]; then
		tar -C "${src_dir}" --exclude='./.*' --exclude='*/.*' -cf - . | tar -C "${dest_dir}" -xf -
	else
		tar -C "${src_dir}" --exclude='./.*' --exclude='*/.*' -cf - . | "${ROOT_CMD[@]}" tar -C "${dest_dir}" -xf -
	fi
}

render_systemd_file() {
	local tmp
	tmp="$(mktemp)"
	sed \
		-e "s|__SERVICE_USER__|${SERVICE_USER}|g" \
		-e "s|__SERVICE_GROUP__|${SERVICE_GROUP}|g" \
		-e "s|__APP_DIR__|${APP_DIR}|g" \
		-e "s|__BACKEND_ENV_FILE__|${BACKEND_ENV_FILE}|g" \
		"${DEPLOY_DIR}/sercherai-backend.service.template" >"${tmp}"
	run_root install -m 0644 "${tmp}" "${SYSTEMD_FILE}"
	rm -f "${tmp}"
}

render_strategy_engine_systemd_file() {
	local tmp
	tmp="$(mktemp)"
	sed \
		-e "s|__SERVICE_USER__|${SERVICE_USER}|g" \
		-e "s|__SERVICE_GROUP__|${SERVICE_GROUP}|g" \
		-e "s|__STRATEGY_ENGINE_APP_DIR__|${STRATEGY_ENGINE_APP_DIR}|g" \
		-e "s|__STRATEGY_ENGINE_ENV_FILE__|${STRATEGY_ENGINE_ENV_FILE}|g" \
		"${DEPLOY_DIR}/sercherai-strategy-engine.service.template" >"${tmp}"
	run_root install -m 0644 "${tmp}" "${STRATEGY_ENGINE_SYSTEMD_FILE}"
	rm -f "${tmp}"
}

render_nginx_file() {
	local tmp
	tmp="$(mktemp)"
	sed \
		-e "s|__CLIENT_PORT__|${CLIENT_PORT}|g" \
		-e "s|__ADMIN_PORT__|${ADMIN_PORT}|g" \
		-e "s|__CLIENT_ROOT__|${WWW_DIR}/client|g" \
		-e "s|__ADMIN_ROOT__|${WWW_DIR}/admin|g" \
		-e "s|__BACKEND_PORT__|${BACKEND_PORT}|g" \
		"${DEPLOY_DIR}/sercherai.nginx.conf.template" >"${tmp}"
	run_root install -m 0644 "${tmp}" "${NGINX_FILE}"
	rm -f "${tmp}"
}

echo "[0/10] checking prerequisites..."
require_cmd go
require_cmd npm
require_cmd python3
require_cmd mysql
require_cmd make
require_cmd systemctl
require_cmd curl
require_cmd tar
if ! is_true "${SKIP_NGINX}"; then
	require_cmd nginx
fi

if ! id -u "${SERVICE_USER}" >/dev/null 2>&1; then
	echo "SERVICE_USER does not exist: ${SERVICE_USER}" >&2
	exit 1
fi
if [ -z "${SERVICE_GROUP}" ]; then
	SERVICE_GROUP="$(id -gn "${SERVICE_USER}")"
fi

echo "[1/10] preparing directories..."
run_root mkdir -p "${APP_DIR}/bin" "${APP_DIR}/uploads" "${STRATEGY_ENGINE_APP_DIR}" "${WWW_DIR}/admin" "${WWW_DIR}/client" "/etc/sercherai"
run_root chown -R "${SERVICE_USER}:${SERVICE_GROUP}" "${APP_DIR}" "${WWW_DIR}"

if is_true "${RUN_DB}"; then
	echo "[2/10] running mysql migrations..."
	(
		cd "${ROOT_DIR}/backend"
		MYSQL_HOST="${MYSQL_HOST}" \
		MYSQL_PORT="${MYSQL_PORT}" \
		MYSQL_USER="${MYSQL_USER}" \
		MYSQL_PWD="${MYSQL_PWD}" \
		MYSQL_DB="${MYSQL_DB}" \
		make init-db
	)

	if is_true "${RUN_SEED}"; then
		echo "[3/10] seeding demo data..."
		(
			cd "${ROOT_DIR}/backend"
			MYSQL_HOST="${MYSQL_HOST}" \
			MYSQL_PORT="${MYSQL_PORT}" \
			MYSQL_USER="${MYSQL_USER}" \
			MYSQL_PWD="${MYSQL_PWD}" \
			MYSQL_DB="${MYSQL_DB}" \
			make seed-db
		)
	else
		echo "[3/10] skipping demo seed (RUN_SEED=${RUN_SEED})..."
	fi
else
	echo "[2/10] skipping mysql migrations (RUN_DB=${RUN_DB})..."
	echo "[3/10] skipping demo seed because database step is disabled..."
fi

if is_true "${RUN_APP}"; then
	echo "[4/10] building backend binary..."
	tmp_backend="$(mktemp)"
	(
		cd "${ROOT_DIR}/backend"
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${tmp_backend}" .
	)
	run_root install -m 0755 "${tmp_backend}" "${APP_DIR}/bin/sercherai-backend"
	rm -f "${tmp_backend}"
	run_root chown "${SERVICE_USER}:${SERVICE_GROUP}" "${APP_DIR}/bin/sercherai-backend"

	echo "[5/10] publishing strategy-engine runtime..."
	run_root rm -rf "${STRATEGY_ENGINE_APP_DIR}/app"
	run_root mkdir -p "${STRATEGY_ENGINE_APP_DIR}"
	copy_dir_without_hidden "${ROOT_DIR}/services/strategy-engine/app" "${STRATEGY_ENGINE_APP_DIR}/app"
	run_root install -m 0644 "${ROOT_DIR}/services/strategy-engine/pyproject.toml" "${STRATEGY_ENGINE_APP_DIR}/pyproject.toml"
	run_root install -m 0644 "${ROOT_DIR}/services/strategy-engine/README.md" "${STRATEGY_ENGINE_APP_DIR}/README.md"
	run_root bash -lc "cd '${STRATEGY_ENGINE_APP_DIR}' && python3 -m venv .venv && .venv/bin/python -m pip install -q --upgrade pip && .venv/bin/python -m pip install -q ."
	run_root chown -R "${SERVICE_USER}:${SERVICE_GROUP}" "${STRATEGY_ENGINE_APP_DIR}"

	echo "[6/10] building admin frontend..."
	(
		cd "${ROOT_DIR}/admin"
		npm ci
		npm run build
	)

	echo "[7/10] building client frontend..."
	(
		cd "${ROOT_DIR}/client"
		npm ci
		npm run build
	)

	echo "[8/10] publishing static assets..."
	run_root rm -rf "${WWW_DIR}/admin" "${WWW_DIR}/client"
	run_root mkdir -p "${WWW_DIR}/admin" "${WWW_DIR}/client"
	copy_dir_without_hidden "${ROOT_DIR}/admin/dist" "${WWW_DIR}/admin"
	copy_dir_without_hidden "${ROOT_DIR}/client/dist" "${WWW_DIR}/client"
	run_root chown -R "${SERVICE_USER}:${SERVICE_GROUP}" "${WWW_DIR}"

	echo "[9/10] installing service config..."
	if ! run_root test -f "${BACKEND_ENV_FILE}"; then
		run_root install -m 0640 "${DEPLOY_DIR}/backend.env.example" "${BACKEND_ENV_FILE}"
		echo "created ${BACKEND_ENV_FILE}; please update secrets before production launch."
	fi
	if ! run_root test -f "${STRATEGY_ENGINE_ENV_FILE}"; then
		run_root install -m 0640 "${DEPLOY_DIR}/strategy-engine.env.example" "${STRATEGY_ENGINE_ENV_FILE}"
		echo "created ${STRATEGY_ENGINE_ENV_FILE}; please update service config before production launch."
	fi
	render_systemd_file
	render_strategy_engine_systemd_file
	run_root systemctl daemon-reload
	run_root systemctl enable --now "${STRATEGY_ENGINE_SERVICE_NAME}"
	run_root systemctl enable --now "${SERVICE_NAME}"
	run_root systemctl restart "${STRATEGY_ENGINE_SERVICE_NAME}"
	run_root systemctl restart "${SERVICE_NAME}"

	if ! is_true "${SKIP_NGINX}"; then
		echo "[10/10] installing nginx config..."
		render_nginx_file
		run_root nginx -t
		run_root systemctl enable --now nginx
		run_root systemctl reload nginx
	else
		echo "[10/10] skipping nginx setup (SKIP_NGINX=${SKIP_NGINX})..."
	fi
else
	echo "[4/10] skipping backend build (RUN_APP=${RUN_APP})..."
	echo "[5/10] skipping strategy-engine runtime publish..."
	echo "[6/10] skipping admin frontend build..."
	echo "[7/10] skipping client frontend build..."
	echo "[8/10] skipping static publish..."
	echo "[9/10] skipping service install/restart..."
	echo "[10/10] skipping nginx setup because app step is disabled..."
fi

echo
echo "deploy completed."
echo "strategy-engine health: http://127.0.0.1:${STRATEGY_ENGINE_PORT}/internal/v1/health"
echo "backend health: http://127.0.0.1:${BACKEND_PORT}/healthz"
echo "client url:     http://<server-ip>:${CLIENT_PORT}/"
echo "admin url:      http://<server-ip>:${ADMIN_PORT}/"
