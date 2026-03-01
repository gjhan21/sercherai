#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT_DIR/.run"
LOG_DIR="$RUN_DIR/logs"

ALL_SERVICES=("backend" "admin" "client")

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

usage() {
  cat <<'EOF'
用法:
  ./scripts/devctl.sh start [backend|admin|client|all]
  ./scripts/devctl.sh stop [backend|admin|client|all]
  ./scripts/devctl.sh restart [backend|admin|client|all]
  ./scripts/devctl.sh status [backend|admin|client|all]

说明:
  - 默认目标为 all
  - start 前会清理目标端口占用进程
  - 日志目录: ./.run/logs
  - PID 文件: ./.run/<service>.pid
EOF
}

ensure_dirs() {
  mkdir -p "$RUN_DIR" "$LOG_DIR"
}

service_port() {
  case "$1" in
    backend) echo "18080" ;;
    admin) echo "5174" ;;
    client) echo "5175" ;;
    *) return 1 ;;
  esac
}

service_url() {
  case "$1" in
    backend) echo "http://127.0.0.1:18080" ;;
    admin) echo "http://127.0.0.1:5174" ;;
    client) echo "http://127.0.0.1:5175" ;;
    *) return 1 ;;
  esac
}

service_cmd() {
  case "$1" in
    backend)
      cat <<EOF
BACKEND_ENV_FILE="$ROOT_DIR/.run/backend.env"
if [ -f "\$BACKEND_ENV_FILE" ]; then
  set -a
  . "\$BACKEND_ENV_FILE"
  set +a
fi
cd "$ROOT_DIR/backend" && exec env APP_PORT=18080 TUSHARE_TOKEN="\${TUSHARE_TOKEN:-}" GOCACHE=\$(pwd)/.gocache GOMODCACHE=\$(pwd)/.gomodcache GOPATH=\$(pwd)/.gopath "$GO_BIN" run .
EOF
      ;;
    admin)
      cat <<EOF
cd "$ROOT_DIR/admin" && exec npm run dev -- --host 0.0.0.0 --port 5174
EOF
      ;;
    client)
      cat <<EOF
cd "$ROOT_DIR/client" && exec npm run dev -- --host 0.0.0.0 --port 5175
EOF
      ;;
    *)
      return 1
      ;;
  esac
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
  nohup bash -lc "$cmd" >>"$logfile" 2>&1 &
  pid=$!
  echo "$pid" > "$pidfile"

  sleep 0.2
  if ! is_pid_running "$pid"; then
    echo "[ERROR] $service 启动失败，进程未存活。日志: $logfile"
    rm -f "$pidfile"
    tail -n 40 "$logfile" || true
    return 1
  fi

  if wait_for_port "$port"; then
    echo "[OK] ${service} 已启动: PID=${pid} PORT=${port} URL=$(service_url "$service")"
  else
    echo "[ERROR] ${service} 未能监听端口 ${port}。日志: ${logfile}"
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
    start|stop|restart|status) ;;
    *)
      echo "[ERROR] 不支持的命令: $action"
      usage
      exit 1
      ;;
  esac

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
