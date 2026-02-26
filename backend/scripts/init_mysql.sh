#!/usr/bin/env bash
set -euo pipefail

MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PWD="${MYSQL_PWD:-abc123}"
MYSQL_DB="${MYSQL_DB:-sercherai}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$(cd "$(dirname "$0")/../migrations" && pwd)}"

echo "[1/4] Checking mysql client..."
command -v mysql >/dev/null 2>&1 || { echo "mysql client not found"; exit 1; }

echo "[2/4] Creating database ${MYSQL_DB} if not exists..."
mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PWD}" \
  -e "CREATE DATABASE IF NOT EXISTS \`${MYSQL_DB}\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

echo "[3/4] Running migrations from ${MIGRATIONS_DIR} ..."
for file in $(ls "${MIGRATIONS_DIR}"/*.sql | sort); do
  echo "  -> $(basename "${file}")"
  mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PWD}" "${MYSQL_DB}" < "${file}"
done

echo "[4/4] Done. Table count:"
mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PWD}" "${MYSQL_DB}" \
  -e "SELECT COUNT(*) AS table_count FROM information_schema.tables WHERE table_schema='${MYSQL_DB}';"

echo "Database initialization completed."

