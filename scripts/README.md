# 开发环境控制脚本

## 文件

- `scripts/devctl.sh`: 统一控制 backend/admin/client 的启动、停止、重启、状态查看。

## 支持命令

```bash
./scripts/devctl.sh start [backend|admin|client|all]
./scripts/devctl.sh stop [backend|admin|client|all]
./scripts/devctl.sh restart [backend|admin|client|all]
./scripts/devctl.sh status [backend|admin|client|all]
```

- 不传第二个参数时默认 `all`。
- `start` 会先清理目标服务端口占用，再启动服务。

## 默认端口

- backend: `18080`
- admin: `5174`
- client: `5175`

## 运行产物

- PID: `./.run/<service>.pid`
- 日志: `./.run/logs/<service>.log`
- backend 环境变量文件: `./.run/backend.env`（可选，`devctl` 启动时自动加载）

## Tushare 行情配置

在 `./.run/backend.env` 配置：

```bash
TUSHARE_TOKEN=你的_tushare_token
```

然后重启 backend：

```bash
./scripts/devctl.sh restart backend
```

## 常用示例

```bash
# 启动全部服务
./scripts/devctl.sh start all

# 查看全部服务状态
./scripts/devctl.sh status all

# 重启 admin
./scripts/devctl.sh restart admin

# 停止全部服务
./scripts/devctl.sh stop all
```
