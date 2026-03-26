#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT_DIR/.run"
LOG_DIR="$RUN_DIR/logs"

ALL_SERVICES=("strategy-graph" "strategy-engine" "backend" "admin" "client")

resolve_go_bin() {
  if [[ -x "/opt/homebrew/bin/go" ]]; then
    echo "/opt/homebrew/bin/go"
    return
  fi
  if [[ -x "/usr/local/go/bin/go" ]]; then
    echo "/usr/local/go/bin/go"
    return
  fi
  command -v go 2>/dev/null || echo "go"
}

GO_BIN="$(resolve_go_bin)"

resolve_python_bin() {
  if command -v python3 >/dev/null 2>&1; then
    command -v python3
    return
  fi
  command -v python 2>/dev/null || echo "python3"
}

PYTHON_BIN="$(resolve_python_bin)"

usage() {
  cat <<'EOF'
用法:
  ./scripts/devctl.sh start [strategy-graph|strategy-engine|backend|admin|client|all]
  ./scripts/devctl.sh stop [strategy-graph|strategy-engine|backend|admin|client|all]
  ./scripts/devctl.sh restart [strategy-graph|strategy-engine|backend|admin|client|all]
  ./scripts/devctl.sh status [strategy-graph|strategy-engine|backend|admin|client|all]
  ./scripts/devctl.sh migrate [all|audit|market-data]

说明:
  - 默认目标为 all
  - start 前会清理目标端口占用进程
  - migrate 支持一键执行数据库迁移（默认 all）
  - 日志目录: ./.run/logs
  - PID 文件: ./.run/<service>.pid
EOF
}

ensure_dirs() {
  mkdir -p "$RUN_DIR" "$LOG_DIR"
}

service_env_file() {
  case "$1" in
    strategy-graph) echo "$ROOT_DIR/.run/strategy-graph.env" ;;
    strategy-engine) echo "$ROOT_DIR/.run/strategy-engine.env" ;;
    backend) echo "$ROOT_DIR/.run/backend.env" ;;
    admin) echo "$ROOT_DIR/.run/admin.env" ;;
    client) echo "$ROOT_DIR/.run/client.env" ;;
    *) return 1 ;;
  esac
}

read_env_override() {
  local file="$1"
  local key="$2"
  local fallback="$3"

  if [[ -f "$file" ]]; then
    local value
    value="$(
      ENV_FILE="$file" ENV_KEY="$key" python3 - <<'PY'
import os

env_file = os.environ["ENV_FILE"]
env_key = os.environ["ENV_KEY"]
value = ""

with open(env_file, "r", encoding="utf-8") as handle:
    for raw in handle:
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, val = line.split("=", 1)
        if key.strip() != env_key:
            continue
        value = val.strip().strip('"').strip("'")
        break

print(value, end="")
PY
    )"
    if [[ -n "$value" ]]; then
      echo "$value"
      return 0
    fi
  fi

  echo "$fallback"
}

service_port() {
  case "$1" in
    strategy-graph) read_env_override "$(service_env_file strategy-graph)" "STRATEGY_GRAPH_PORT" "18082" ;;
    strategy-engine) read_env_override "$(service_env_file strategy-engine)" "STRATEGY_ENGINE_PORT" "18081" ;;
    backend) read_env_override "$(service_env_file backend)" "APP_PORT" "18080" ;;
    admin) read_env_override "$(service_env_file admin)" "ADMIN_PORT" "5174" ;;
    client) read_env_override "$(service_env_file client)" "CLIENT_PORT" "5175" ;;
    *) return 1 ;;
  esac
}

service_url() {
  case "$1" in
    strategy-graph)
      echo "http://127.0.0.1:$(service_port strategy-graph)"
      ;;
    strategy-engine)
      echo "http://127.0.0.1:$(service_port strategy-engine)"
      ;;
    backend)
      echo "http://127.0.0.1:$(service_port backend)"
      ;;
    admin)
      echo "http://127.0.0.1:$(service_port admin)"
      ;;
    client)
      echo "http://127.0.0.1:$(service_port client)"
      ;;
    *) return 1 ;;
  esac
}

service_health_url() {
  case "$1" in
    strategy-graph)
      echo "$(service_url strategy-graph)/health"
      ;;
    strategy-engine)
      echo "$(service_url strategy-engine)/internal/v1/health"
      ;;
    backend)
      echo "$(service_url backend)/healthz"
      ;;
    *)
      echo ""
      ;;
  esac
}

service_cmd() {
  case "$1" in
    strategy-graph)
      cat <<EOF
STRATEGY_GRAPH_ENV_FILE="$ROOT_DIR/.run/strategy-graph.env"
if [ -f "\$STRATEGY_GRAPH_ENV_FILE" ]; then
  set -a
  . "\$STRATEGY_GRAPH_ENV_FILE"
  set +a
fi
cd "$ROOT_DIR/services/strategy-graph"
if [ ! -x ".venv/bin/python" ]; then
  "$PYTHON_BIN" -m venv .venv
fi
if [ ! -f ".venv/.deps_ready" ] || [ pyproject.toml -nt ".venv/.deps_ready" ]; then
  .venv/bin/python -m pip install -q -e '.[dev]'
  touch .venv/.deps_ready
fi
STRATEGY_GRAPH_HOST="\${STRATEGY_GRAPH_HOST:-0.0.0.0}"
STRATEGY_GRAPH_PORT="\${STRATEGY_GRAPH_PORT:-18082}"
exec env STRATEGY_GRAPH_HOST="\${STRATEGY_GRAPH_HOST}" STRATEGY_GRAPH_PORT="\${STRATEGY_GRAPH_PORT}" .venv/bin/python -m uvicorn app.main:app --host "\${STRATEGY_GRAPH_HOST}" --port "\${STRATEGY_GRAPH_PORT}"
EOF
      ;;
    strategy-engine)
      cat <<EOF
STRATEGY_ENGINE_ENV_FILE="$ROOT_DIR/.run/strategy-engine.env"
if [ -f "\$STRATEGY_ENGINE_ENV_FILE" ]; then
  set -a
  . "\$STRATEGY_ENGINE_ENV_FILE"
  set +a
fi
cd "$ROOT_DIR/services/strategy-engine"
if [ ! -x ".venv/bin/python" ]; then
  "$PYTHON_BIN" -m venv .venv
fi
if [ ! -f ".venv/.deps_ready" ] || [ pyproject.toml -nt ".venv/.deps_ready" ]; then
  .venv/bin/python -m pip install -q -e '.[dev]'
  touch .venv/.deps_ready
fi
STRATEGY_ENGINE_HOST="\${STRATEGY_ENGINE_HOST:-0.0.0.0}"
STRATEGY_ENGINE_PORT="\${STRATEGY_ENGINE_PORT:-18081}"
STRATEGY_ENGINE_GO_BACKEND_BASE_URL="\${STRATEGY_ENGINE_GO_BACKEND_BASE_URL:-$(service_url backend)}"
STRATEGY_ENGINE_GRAPH_SERVICE_BASE_URL="\${STRATEGY_ENGINE_GRAPH_SERVICE_BASE_URL:-$(service_url strategy-graph)}"
exec env STRATEGY_ENGINE_HOST="\${STRATEGY_ENGINE_HOST}" STRATEGY_ENGINE_PORT="\${STRATEGY_ENGINE_PORT}" STRATEGY_ENGINE_GO_BACKEND_BASE_URL="\${STRATEGY_ENGINE_GO_BACKEND_BASE_URL}" STRATEGY_ENGINE_GRAPH_SERVICE_BASE_URL="\${STRATEGY_ENGINE_GRAPH_SERVICE_BASE_URL}" .venv/bin/python -m uvicorn app.main:app --host "\${STRATEGY_ENGINE_HOST}" --port "\${STRATEGY_ENGINE_PORT}"
EOF
      ;;
    backend)
      cat <<EOF
BACKEND_ENV_FILE="$ROOT_DIR/.run/backend.env"
if [ -f "\$BACKEND_ENV_FILE" ]; then
  set -a
  . "\$BACKEND_ENV_FILE"
  set +a
fi
APP_PORT="\${APP_PORT:-$(service_port backend)}"
STRATEGY_ENGINE_BASE_URL="\${STRATEGY_ENGINE_BASE_URL:-$(service_url strategy-engine)}"
STRATEGY_GRAPH_BASE_URL="\${STRATEGY_GRAPH_BASE_URL:-$(service_url strategy-graph)}"
cd "$ROOT_DIR/backend" && exec env APP_PORT="\${APP_PORT}" TUSHARE_TOKEN="\${TUSHARE_TOKEN:-}" STRATEGY_ENGINE_BASE_URL="\${STRATEGY_ENGINE_BASE_URL}" STRATEGY_GRAPH_BASE_URL="\${STRATEGY_GRAPH_BASE_URL}" GOCACHE=\$(pwd)/.gocache GOMODCACHE=\$(pwd)/.gomodcache GOPATH=\$(pwd)/.gopath "$GO_BIN" run .
EOF
      ;;
    admin)
      cat <<EOF
ADMIN_ENV_FILE="$ROOT_DIR/.run/admin.env"
if [ -f "\$ADMIN_ENV_FILE" ]; then
  set -a
  . "\$ADMIN_ENV_FILE"
  set +a
fi
ADMIN_HOST="\${ADMIN_HOST:-0.0.0.0}"
ADMIN_PORT="\${ADMIN_PORT:-$(service_port admin)}"
VITE_PROXY_TARGET="\${VITE_PROXY_TARGET:-$(service_url backend)}"
cd "$ROOT_DIR/admin" && exec env VITE_PROXY_TARGET="\${VITE_PROXY_TARGET}" npm run dev -- --host "\${ADMIN_HOST}" --port "\${ADMIN_PORT}"
EOF
      ;;
    client)
      cat <<EOF
CLIENT_ENV_FILE="$ROOT_DIR/.run/client.env"
if [ -f "\$CLIENT_ENV_FILE" ]; then
  set -a
  . "\$CLIENT_ENV_FILE"
  set +a
fi
CLIENT_HOST="\${CLIENT_HOST:-0.0.0.0}"
CLIENT_PORT="\${CLIENT_PORT:-$(service_port client)}"
cd "$ROOT_DIR/client" && exec npm run dev -- --host "\${CLIENT_HOST}" --port "\${CLIENT_PORT}"
EOF
      ;;
    *)
      return 1
      ;;
  esac
}

service_wait_checks() {
  case "$1" in
    strategy-graph) echo "160" ;;
    strategy-engine) echo "320" ;;
    *) echo "80" ;;
  esac
}

migration_files_for_target() {
  case "$1" in
    audit)
      echo "$ROOT_DIR/backend/migrations/20260324_00_admin_audit_events.sql"
      ;;
    market-data)
      cat <<EOF
$ROOT_DIR/backend/migrations/20260323_00_market_provider_governance.sql
$ROOT_DIR/backend/migrations/20260324_01_market_data_full_backfill.sql
EOF
      ;;
    *) return 1 ;;
  esac
}

resolve_mysql_config() {
  local key="$1"
  local fallback="$2"
  read_env_override "$(service_env_file backend)" "$key" "$fallback"
}

run_migration() {
  local target="$1"
  local mysql_host mysql_port mysql_user mysql_pwd mysql_db migration_files
  local -a table_checks

  mysql_host="$(resolve_mysql_config MYSQL_HOST 127.0.0.1)"
  mysql_port="$(resolve_mysql_config MYSQL_PORT 3306)"
  mysql_user="$(resolve_mysql_config MYSQL_USER root)"
  mysql_pwd="$(resolve_mysql_config MYSQL_PWD abc123)"
  mysql_db="$(resolve_mysql_config MYSQL_DB sercherai)"

  if ! command -v mysql >/dev/null 2>&1; then
    echo "[ERROR] 未找到 mysql 客户端，请先安装 mysql 并确保可执行"
    return 1
  fi

  if [[ "$target" == "all" ]]; then
    echo "[INFO] 执行全量数据库迁移 (target=all, db=${mysql_db}@${mysql_host}:${mysql_port})"
    (
      cd "$ROOT_DIR/backend"
      MYSQL_HOST="$mysql_host" \
      MYSQL_PORT="$mysql_port" \
      MYSQL_USER="$mysql_user" \
      MYSQL_PWD="$mysql_pwd" \
      MYSQL_DB="$mysql_db" \
      ./scripts/init_mysql.sh
    )
    echo "[OK] 数据库全量迁移完成"
    return 0
  fi

  migration_files="$(migration_files_for_target "$target")"
  if [[ -z "$migration_files" ]]; then
    echo "[ERROR] 未找到迁移文件配置: ${target}"
    return 1
  fi
  while IFS= read -r migration_file; do
    [[ -z "$migration_file" ]] && continue
    if [[ ! -f "$migration_file" ]]; then
      echo "[ERROR] 迁移文件不存在: $migration_file"
      return 1
    fi
  done <<< "$migration_files"

  mysql -h"${mysql_host}" -P"${mysql_port}" -u"${mysql_user}" -p"${mysql_pwd}" \
    -e "CREATE DATABASE IF NOT EXISTS \`${mysql_db}\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
  while IFS= read -r migration_file; do
    [[ -z "$migration_file" ]] && continue
    echo "[INFO] 执行数据库迁移 (${target}): $(basename "$migration_file")"
    mysql -h"${mysql_host}" -P"${mysql_port}" -u"${mysql_user}" -p"${mysql_pwd}" "${mysql_db}" < "$migration_file"
  done <<< "$migration_files"

  case "$target" in
    audit)
      table_checks=("admin_audit_events")
      ;;
    market-data)
      table_checks=("market_provider_registry" "market_provider_capabilities" "market_provider_routing_policies" "market_universe_snapshots")
      ;;
    *)
      table_checks=()
      ;;
  esac

  if [[ "${#table_checks[@]}" -eq 0 ]]; then
    echo "[OK] 迁移完成: ${target}"
    return 0
  fi

  local table_check exists
  for table_check in "${table_checks[@]}"; do
    exists="$(
      mysql -h"${mysql_host}" -P"${mysql_port}" -u"${mysql_user}" -p"${mysql_pwd}" -N -e \
        "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='${mysql_db}' AND table_name='${table_check}';"
    )"
    if [[ "${exists}" == "0" ]]; then
      echo "[ERROR] 迁移执行后仍未检测到表: ${table_check}"
      return 1
    fi
  done
  echo "[OK] 迁移完成，已检测到关键表: ${table_checks[*]}"
}

pid_file() {
  echo "$RUN_DIR/$1.pid"
}

log_file() {
  echo "$LOG_DIR/$1.log"
}

is_pid_running() {
  local pid="$1"
  [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

kill_with_timeout() {
  local pid="$1"
  local label="$2"
  local wait_count=0

  if ! is_pid_running "$pid"; then
    return 0
  fi

  kill "$pid" 2>/dev/null || true

  while is_pid_running "$pid" && [[ $wait_count -lt 20 ]]; do
    sleep 0.2
    wait_count=$((wait_count + 1))
  done

  if is_pid_running "$pid"; then
    echo "[WARN] $label(PID=$pid) 未在 4 秒内退出，执行强制终止"
    kill -9 "$pid" 2>/dev/null || true
  fi
}

kill_port_processes() {
  local port="$1"
  local pids

  pids="$(lsof -ti tcp:"$port" 2>/dev/null || true)"
  if [[ -z "$pids" ]]; then
    return 0
  fi

  echo "[INFO] 端口 ${port} 已被占用，清理进程: $(echo "$pids" | tr '\n' ' ')"
  while IFS= read -r pid; do
    [[ -z "$pid" ]] && continue
    kill_with_timeout "$pid" "port-${port}"
  done <<< "$pids"
}

wait_for_port() {
  local port="$1"
  local max_checks="${2:-80}"
  local checks=0

  while [[ $checks -lt $max_checks ]]; do
    if lsof -ti tcp:"$port" >/dev/null 2>&1; then
      return 0
    fi
    checks=$((checks + 1))
    sleep 0.25
  done
  return 1
}

is_port_listening() {
  local port="$1"
  lsof -ti tcp:"$port" >/dev/null 2>&1
}

wait_for_http_ready() {
  local url="$1"
  local max_checks="${2:-40}"
  local checks=0

  if [[ -z "$url" ]]; then
    return 0
  fi
  if ! command -v curl >/dev/null 2>&1; then
    return 0
  fi

  while [[ $checks -lt $max_checks ]]; do
    if curl -fsS -m 2 "$url" >/dev/null 2>&1; then
      return 0
    fi
    checks=$((checks + 1))
    sleep 0.25
  done
  return 1
}

first_listener_pid() {
  local port="$1"
  lsof -ti tcp:"$port" 2>/dev/null | head -n 1
}

spawn_detached() {
  local logfile="$1"
  local cmd="$2"

  "$PYTHON_BIN" - "$logfile" "$cmd" <<'PY'
import subprocess
import sys

logfile = sys.argv[1]
cmd = sys.argv[2]

with open(logfile, "ab", buffering=0) as log_handle:
    proc = subprocess.Popen(
        ["/bin/bash", "-lc", cmd],
        stdin=subprocess.DEVNULL,
        stdout=log_handle,
        stderr=subprocess.STDOUT,
        start_new_session=True,
        close_fds=True,
    )

print(proc.pid)
PY
}

start_service() {
  local service="$1"
  local port cmd pidfile logfile old_pid pid

  ensure_dirs
  port="$(service_port "$service")"
  cmd="$(service_cmd "$service")"
  pidfile="$(pid_file "$service")"
  logfile="$(log_file "$service")"

  if [[ -f "$pidfile" ]]; then
    old_pid="$(tr -d '[:space:]' < "$pidfile")"
    if is_pid_running "$old_pid"; then
      echo "[INFO] $service 已在运行(PID=$old_pid)，先停止旧进程"
      kill_with_timeout "$old_pid" "$service"
    fi
    rm -f "$pidfile"
  fi

  kill_port_processes "$port"

  echo "[INFO] 启动 $service ..."
  pid="$(spawn_detached "$logfile" "$cmd")"
  echo "$pid" > "$pidfile"

  sleep 0.2
  if ! is_pid_running "$pid"; then
    echo "[ERROR] $service 启动失败，进程未存活。日志: $logfile"
    rm -f "$pidfile"
    tail -n 40 "$logfile" || true
    return 1
  fi

  if wait_for_port "$port" "$(service_wait_checks "$service")"; then
    local health_url listener_pid
    health_url="$(service_health_url "$service")"
    if ! wait_for_http_ready "$health_url" 24; then
      echo "[ERROR] ${service} 端口已监听，但健康检查失败。日志: ${logfile}"
      rm -f "$pidfile"
      tail -n 80 "$logfile" || true
      return 1
    fi
    sleep 0.5
    if ! is_pid_running "$pid" || ! is_port_listening "$port"; then
      echo "[ERROR] ${service} 启动后未稳定运行。日志: ${logfile}"
      rm -f "$pidfile"
      tail -n 80 "$logfile" || true
      return 1
    fi
    listener_pid="$(first_listener_pid "$port")"
    if [[ -n "$listener_pid" ]]; then
      echo "$listener_pid" > "$pidfile"
    fi
    echo "[OK] ${service} 已启动: PID=${pid} PORT=${port} URL=$(service_url "$service")"
  else
    echo "[ERROR] ${service} 未能监听端口 ${port}。日志: ${logfile}"
    rm -f "$pidfile"
    tail -n 80 "$logfile" || true
    return 1
  fi
}

stop_service() {
  local service="$1"
  local port pidfile pid pids stopped=0

  port="$(service_port "$service")"
  pidfile="$(pid_file "$service")"

  if [[ -f "$pidfile" ]]; then
    pid="$(tr -d '[:space:]' < "$pidfile")"
    if is_pid_running "$pid"; then
      echo "[INFO] 停止 $service(PID=$pid)"
      kill_with_timeout "$pid" "$service"
      stopped=1
    fi
    rm -f "$pidfile"
  fi

  pids="$(lsof -ti tcp:"$port" 2>/dev/null || true)"
  if [[ -n "$pids" ]]; then
    echo "[INFO] 清理 ${service} 端口 ${port} 进程: $(echo "$pids" | tr '\n' ' ')"
    while IFS= read -r pid; do
      [[ -z "$pid" ]] && continue
      kill_with_timeout "$pid" "${service}-port-${port}"
    done <<< "$pids"
    stopped=1
  fi

  if [[ $stopped -eq 1 ]]; then
    echo "[OK] $service 已停止"
  else
    echo "[INFO] $service 未运行"
  fi
}

status_service() {
  local service="$1"
  local port pidfile pid process_status port_pids

  ensure_dirs
  port="$(service_port "$service")"
  pidfile="$(pid_file "$service")"
  process_status="stopped"

  if [[ -f "$pidfile" ]]; then
    pid="$(tr -d '[:space:]' < "$pidfile")"
    if is_pid_running "$pid"; then
      process_status="running (PID=$pid)"
    else
      process_status="stale pid file (PID=$pid)"
    fi
  fi

  port_pids="$(lsof -ti tcp:"$port" 2>/dev/null | tr '\n' ' ' | sed 's/[[:space:]]*$//' || true)"
  if [[ -n "$port_pids" ]]; then
    echo "[STATUS] ${service}: ${process_status}, port=${port}(LISTEN:${port_pids}), log=$(log_file "$service")"
  else
    echo "[STATUS] ${service}: ${process_status}, port=${port}(free), log=$(log_file "$service")"
  fi
}

is_valid_service() {
  local target="$1"
  local item
  for item in "${ALL_SERVICES[@]}"; do
    if [[ "$item" == "$target" ]]; then
      return 0
    fi
  done
  return 1
}

main() {
  local action="${1:-}"
  local target="${2:-all}"
  local services=()
  local service

  if [[ -z "$action" ]]; then
    usage
    exit 1
  fi

  case "$action" in
    start|stop|restart|status|migrate) ;;
    *)
      echo "[ERROR] 不支持的命令: $action"
      usage
      exit 1
      ;;
  esac

  if [[ "$action" == "migrate" ]]; then
    case "$target" in
      all|audit|market-data)
        run_migration "$target"
        ;;
      *)
        echo "[ERROR] 不支持的迁移目标: $target"
        usage
        exit 1
        ;;
    esac
    return 0
  fi

  if [[ "$target" == "all" ]]; then
    services=("${ALL_SERVICES[@]}")
  elif is_valid_service "$target"; then
    services=("$target")
  else
    echo "[ERROR] 不支持的服务: $target"
    usage
    exit 1
  fi

  case "$action" in
    start)
      for service in "${services[@]}"; do
        start_service "$service"
      done
      ;;
    stop)
      for service in "${services[@]}"; do
        stop_service "$service"
      done
      ;;
    restart)
      for service in "${services[@]}"; do
        stop_service "$service"
      done
      for service in "${services[@]}"; do
        start_service "$service"
      done
      ;;
    status)
      for service in "${services[@]}"; do
        status_service "$service"
      done
      ;;
  esac
}

main "$@"
