# 开发环境控制脚本

## 文件

- `scripts/devctl.sh`: 统一控制 strategy-engine/backend/admin/client 的启动、停止、重启、状态查看。
- `scripts/deploy_linux_server.sh`: Linux 服务器一键部署（迁移数据库、构建 backend + strategy-engine + 前端、安装 systemd 和 nginx 配置）。
- `scripts/deploy_linux_db.sh`: 仅执行数据库初始化/迁移（可选 seed），不部署应用与 nginx。
- `scripts/deploy_linux_app.sh`: 仅部署应用（后端二进制 + admin/client 静态资源 + systemd/nginx），不执行数据库迁移。

## 支持命令

```bash
./scripts/devctl.sh start [strategy-engine|backend|admin|client|client-h5|client-pc|all]
./scripts/devctl.sh stop [strategy-engine|backend|admin|client|client-h5|client-pc|all]
./scripts/devctl.sh restart [strategy-engine|backend|admin|client|client-h5|client-pc|all]
./scripts/devctl.sh status [strategy-engine|backend|admin|client|client-h5|client-pc|all]
```

- 不传第二个参数时默认 `all`。
- `client-h5` / `client-pc` 是 `client` 的快捷别名，共享同一份 PID 和日志。
- `start` 会先清理目标服务端口占用，再启动服务。
- `start` 会在端口监听后继续做健康检查；`strategy-engine` 会校验 `/internal/v1/health`，backend 会校验 `/healthz`。
- `start` 使用真正 detached 的子进程启动服务，脚本返回后服务不会因为当前终端关闭而被一并回收。

## 默认端口

- strategy-engine: `18081`
- backend: `18080`
- admin: `5174`
- client: `5175`

## 运行产物

- PID: `./.run/<service>.pid`
- 日志: `./.run/logs/<service>.log`
- `PID` 会尽量回写为实际监听端口的进程，便于后续 `status / stop / restart` 对齐真实服务。
- strategy-engine 环境变量文件: `./.run/strategy-engine.env`（可选，`devctl` 启动时自动加载）
- backend 环境变量文件: `./.run/backend.env`（可选，`devctl` 启动时自动加载）
- admin 环境变量文件: `./.run/admin.env`（可选，`devctl` 启动时自动加载）
- client 环境变量文件: `./.run/client.env`（可选，`devctl` 启动时自动加载）

## Strategy Engine 配置

默认情况下，`devctl` 会：

- 先启动 `strategy-engine`
- 自动为 `services/strategy-engine` 创建 `.venv`
- 首次启动时自动执行 `pip install -e '.[dev]'`
- 启动 `strategy-engine` 时默认注入 `STRATEGY_ENGINE_GO_BACKEND_BASE_URL=http://127.0.0.1:<backend-port>`，让它能回调 Go backend 拉取股票/期货上下文
- 启动 backend 时默认注入 `STRATEGY_ENGINE_BASE_URL=http://127.0.0.1:18081`
- 启动 admin 时默认把 `VITE_PROXY_TARGET` 指向当前 backend 端口

如果你要覆盖 host 或 port，可在 `./.run/strategy-engine.env` 配置：

```bash
STRATEGY_ENGINE_HOST=0.0.0.0
STRATEGY_ENGINE_PORT=18081
STRATEGY_ENGINE_GO_BACKEND_BASE_URL=http://127.0.0.1:18080
```

## Tushare 行情配置

在 `./.run/backend.env` 配置：

```bash
TUSHARE_TOKEN=你的_tushare_token
```

然后重启 backend：

```bash
./scripts/devctl.sh restart backend
```

## 自定义本地端口

如果你想把本地联调改成 `19081/5176/5177` 这种端口组合，不需要再手动分开起服务，直接写到 `.run/*.env`：

`./.run/backend.env`

```bash
APP_PORT=19081
STRATEGY_ENGINE_BASE_URL=http://127.0.0.1:18081
TUSHARE_TOKEN=你的_tushare_token
```

`./.run/admin.env`

```bash
ADMIN_HOST=127.0.0.1
ADMIN_PORT=5176
VITE_PROXY_TARGET=http://127.0.0.1:19081
```

`./.run/client.env`

```bash
CLIENT_HOST=127.0.0.1
CLIENT_PORT=5177
# 可选: h5 / pc，默认 h5
CLIENT_MODE=h5
```

写完后重启对应服务：

```bash
./scripts/devctl.sh restart backend
./scripts/devctl.sh restart admin
./scripts/devctl.sh restart client
```

说明：

- `CLIENT_MODE=h5` 时，`devctl` 会执行 `npm run dev:h5`，访问地址为 `http://127.0.0.1:<port>/m/`
- `CLIENT_MODE=pc` 时，`devctl` 会执行 `npm run dev:pc`，访问地址为 `http://127.0.0.1:<port>/`
- 也可以直接使用 `./scripts/devctl.sh start client-h5` 或 `./scripts/devctl.sh start client-pc` 临时指定模式，无需改 `.run/client.env`

## 常用示例

```bash
# 启动全部服务
./scripts/devctl.sh start all

# 单独重启策略服务
./scripts/devctl.sh restart strategy-engine

# 查看全部服务状态
./scripts/devctl.sh status all

# 重启 admin
./scripts/devctl.sh restart admin

# 直接启动 H5 client
./scripts/devctl.sh start client-h5

# 直接启动 PC client
./scripts/devctl.sh start client-pc

# 停止全部服务
./scripts/devctl.sh stop all
```
