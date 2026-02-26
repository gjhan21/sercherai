#!/usr/bin/env bash
set -euo pipefail

MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PWD="${MYSQL_PWD:-abc123}"
MYSQL_DB="${MYSQL_DB:-sercherai}"
SEED_FILE="${SEED_FILE:-$(cd "$(dirname "$0")" && pwd)/seed_demo.sql}"

echo "[1/3] Checking mysql client..."
command -v mysql >/dev/null 2>&1 || { echo "mysql client not found"; exit 1; }

echo "[2/3] Seeding demo data from ${SEED_FILE} ..."
mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PWD}" "${MYSQL_DB}" < "${SEED_FILE}"

echo "[3/3] Done. Quick check:"
mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PWD}" "${MYSQL_DB}" \
  -e "SELECT id, member_level FROM users WHERE id='u_demo_001'; SELECT id, contract_a, contract_b, status FROM arbitrage_recos LIMIT 1; SELECT id, contract, valid_to FROM futures_guidances LIMIT 1;"

echo "Demo seed completed."
