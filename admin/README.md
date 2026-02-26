# SercherAI Admin

管理端前端（Vue 3 + Vite + Element Plus）首版，已接入以下后端模块：

- 登录鉴权（`/api/v1/auth/login`、`/api/v1/auth/mock-login`）
- 仪表盘（`/api/v1/admin/dashboard/overview`）
- 用户管理（`/api/v1/admin/users*`）
- 新闻管理（`/api/v1/admin/news*`）
- 审核中心（`/api/v1/admin/workflow/reviews*`、`/api/v1/admin/workflow/metrics`）
- 任务中心（`/api/v1/admin/system/job-definitions*`、`/api/v1/admin/system/job-runs*`）
- 数据源管理与健康检查（`/api/v1/admin/data-sources*`）
- 流程消息（`/api/v1/admin/workflow/messages*`）

## 启动方式

```bash
cd /Users/gjhan21/cursor/sercherai/admin
npm install
npm run dev
```

默认端口 `5174`，默认代理到 `http://127.0.0.1:18080`。

## 环境变量

复制 `.env.example` 到 `.env` 并按需修改：

- `VITE_API_BASE_URL`：前端请求基地址，默认 `/api/v1`
- `VITE_PROXY_TARGET`：Vite 开发代理后端地址

## 联调建议

1. 启动后端（建议开启 `ALLOW_MOCK_LOGIN=true` 便于开发）
2. 打开管理端，使用 Mock 登录（`admin_001` / `ADMIN`）
3. 先联调「用户管理 / 新闻管理 / 审核中心 / 任务中心 / 数据源管理 / 流程消息」
