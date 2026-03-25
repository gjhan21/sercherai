# 专题E 第一轮开发启动清单

## Goal

先把统一审计事件真相源立起来，再逐步让通知中心与旧消息入口回读它。

## 第一轮包顺序

### 包1：E0 / 审计事件 schema

- 已完成
- Files:
  - Create: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/migrations/20260324_00_admin_audit_events.sql`
  - Create: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/model/admin_audit_event.go`
  - Create: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/repo/admin_audit_event_repo.go`
- Expected:
  - 后端具备统一 `audit event` 数据模型
  - 支持 `metadata / dedupe_key / level / domain`

### 包2：E1 / 统一事件入库入口

- 已完成深化第三刀
- Files:
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/repo/interfaces.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/service/service.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/repo/inmemory_repo.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/repo/strategy_engine_config_repo.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/repo/mysql_repo.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/handler/admin_growth_handler.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/handler/strategy_engine_config_handler.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/router/router.go`
- Expected:
  - 关键系统动作通过统一入口写入 `admin_audit_events`
  - 默认策略自动落库、发布、审核结果、数据源告警至少有一条主链事件

### 包3：E1.5 / 最小 Admin 读取入口

- 已完成 API，UI 待下一包
- Files:
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/api/admin.js`
- Expected:
  - Admin 可通过 `GET /api/v1/admin/audit/events` 拉取统一审计事件
  - 旧操作日志页暂不重写

### 包4：E2 / 通知路由与聚合

- 已完成第二刀
- Files:
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/repo/admin_audit_event_repo.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/repo/inmemory_repo.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/repo/mysql_repo.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/handler/admin_growth_handler.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/router/router.go`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/AuditLogsView.vue`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/WorkflowMessagesView.vue`
  - Test: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/repo/admin_audit_event_repo_test.go`
  - Test: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/handler/market_data_admin_handler_test.go`
  - Test: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/audit-logs-view.test.js`
  - Test: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/workflow-messages-view.test.js`
- Expected:
  - `audit event` 可自动路由到现有 `workflow_messages`
  - 后台可读取统一 audit summary
  - open dedupe event 会聚合 `occurrence_count / last_seen_at`，并按最新严重级别升级
  - `AuditLogsView` 与 `WorkflowMessagesView` 都开始回读统一 audit summary / open events

### 包5：E3 / 统一消息中心第一版

- 已完成第一刀
- Files:
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/WorkflowMessagesView.vue`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/AuditLogsView.vue`
  - Test: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/workflow-messages-view.test.js`
  - Test: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/audit-logs-view.test.js`
- Expected:
  - `WorkflowMessagesView` 升级为“消息中心”壳，至少包含 `流程待办 / 开放事件` 两个视图
  - `AuditLogsView` 退回“审计与操作日志”定位，并可跳回消息中心
  - 不删除旧 `workflow_messages` 与旧 `operation_logs` 路径

### 包6：E4 / Review、Jobs、Data Sources 回读统一事件流

- 已完成深化第三刀
- Files:
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/ReviewCenterView.vue`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/SystemJobsView.vue`
  - Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/DataSourcesView.vue`
  - Test: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/review-center-view.test.js`
  - Test: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/system-jobs-view.test.js`
  - Test: `/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/data-sources-view.test.js`
- Expected:
  - 审核中心、系统任务、数据源治理台都直接读取统一 `audit summary / open events`
  - 各模块事件摘要可一键跳转回 `/workflow-messages` 消息中心，不再各自维持弱语义告警入口
  - 旧页面主体结构保持不变，只在原入口上叠加统一事件流摘要
  - `WorkflowMessagesView` / `AuditLogsView` 开始支持对象页跳转、`对象类型` 过滤与 `已关闭事件` 聚合
  - `ReviewCenterView` / `SystemJobsView` / `DataSourcesView` 开始消费对象级 deep link：`review_id / run_id / source_key`
  - `WorkflowMessagesView` / `AuditLogsView` 已补 `STRATEGY_JOB / STRATEGY_PUBLISH_POLICY` 深链，`MarketCenterView` 与 `StrategyEngineConfigPanel` 已可消费 `publish_id / policy_id` 自动定位对象

## 验证命令

- `cd /Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend && go test ./internal/growth/... ./router/...`
- `node --test /Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/review-center-view.test.js /Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/system-jobs-view.test.js /Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/views/data-sources-view.test.js`
- `cd /Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin && npm run build`

## 不并入第一轮

- 第三方 IM 集成
- 对外 webhook
- 全量消息中心 UI 重做
