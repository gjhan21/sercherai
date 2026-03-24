# 专题E 第一轮（统一审计事件流与消息中心）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 review、jobs、workflow、messages、strategy-engine publish/replay、market truth 运维基线上，落一条统一 audit event stream，并让 Admin 现有页面开始回读同一套事件真相源。

**Architecture:** 不另起第二套“消息系统”，而是在现有 `/admin/audit`、`/admin/workflow`、`/admin/system`、`/admin/users`、`/admin/data-sources` 入口之上，新增统一 audit event schema、入库 repo、事件路由与 inbox 聚合接口。Admin 第一轮继续复用 `WorkflowMessagesView`、`SystemJobsView`、`ReviewCenterView`、`AuditLogsView`、`DataSourcesView` 等现有页面，只增加统一事件读取与跳转能力。

**Tech Stack:** Go, Gin, MySQL, Vue 3, Element Plus, Go test, npm build

---

最后更新: 2026-03-23  
状态: Planned

## 文档定位

这份清单是：

- [专题E-审计通知与系统消息平台.md](/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题E-审计通知与系统消息平台.md)

对应的第一轮可执行实施计划。

它不是专题E全部收官文档，而是“按当前代码入口直接开工”的文件级任务表。默认顺序固定为：

1. 包0：基线盘点与事件范围冻结
2. 包1：统一 audit event schema 与持久化 repo
3. 包2：事件写入入口与通知路由聚合
4. 包3：Admin 消息中心第一版
5. 包4：现有模块接入与回归收口

## 开工前提

- 专题A 第一轮已完成，provider routing / quality 摘要已可被消费
- `strategy_engine_job_snapshots` 与 publish/replay 本地真相源已可用
- 现有 Admin 页面继续保留，不新开第二套“消息后台”
- 第一轮不接第三方 IM、不做 webhook 平台
- 旧用户消息、旧 workflow messages、旧 operation logs 均不删除，只开始回读统一事件摘要

## 包0：基线盘点与事件范围冻结

### Task 0: 锁定第一轮事件范围与现有入口

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题E-第一轮-开发启动清单.md`
- Review: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_job_snapshot_repo.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_runtime_repo.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_admin.go`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/WorkflowMessagesView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/UserMessagesView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/AuditLogsView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`

- [ ] **Step 1: 冻结第一轮统一事件类型**

范围固定为：

- `POLICY_AUTO_MATERIALIZED`：默认策略/默认配置自动落库
- `STRATEGY_PUBLISH_FORCED`：强制发布
- `STRATEGY_PUBLISH_REPLAYED`：覆盖发布 / replay
- `REVIEW_REJECTED`：审核驳回
- `REVIEW_ROLLED_BACK`：回滚 / 人工撤销
- `PROVIDER_ROUTING_FALLBACK`：provider fallback / 切源
- `TRUTH_REBUILD_FAILED`：truth 重建失败
- `RESEARCH_RUN_FAILED`：研究运行失败
- `RESEARCH_WARNING_ESCALATED`：warning 升级

- [ ] **Step 2: 冻结第一轮严重级别与确认状态**

第一轮统一为：

- severity：`info / warning / critical`
- inbox status：`unread / acknowledged / ignored / resolved`

- [ ] **Step 3: 标记现有散落入口与未来接入点**

至少标明以下代码入口：

- `ListOperationLogs` / `ExportOperationLogsCSV`
- `ListWorkflowMessages` / `UpdateWorkflowMessageRead`
- `ListSchedulerJobRuns` / `TriggerSchedulerJob` / `RetrySchedulerJobRun`
- `ListUserMessages` / `CreateUserMessages`
- `AdminListStrategyEngineJobs` / `AdminPublishStrategyEngineJob`
- market truth 重建与 market sync summary

- [ ] **Step 4: 固化第一轮回归基线命令**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 现有 review、workflow、job、strategy-engine snapshot、market admin 相关测试保持通过

## 包1：统一 audit event schema 与持久化 repo

### Task 1: 新增 audit event migration、模型与 repo 测试骨架

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260323_00_audit_event_stream.sql`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/audit_event.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/audit_event_repo_test.go`

- [ ] **Step 1: 先写 failing repo/model 测试**

覆盖：

- 新增事件入库
- 按 `event_type / severity / object_type / object_id / status` 查询
- 事件 payload roundtrip
- 事件聚合摘要可返回 unread count 与 latest event

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestAuditEvent'`  
Expected: 因 schema / model / repo 尚不存在而失败

- [ ] **Step 2: 新增 migration**

第一轮至少落以下表：

- `audit_events`
- `audit_event_targets`
- `audit_event_inbox_states`

要求：

- 只增不改，不删除现有 `operation_logs`、`workflow_messages`、`system_job_runs`、`strategy_job_*` 表
- `audit_events` 记录统一事件本体
- `audit_event_targets` 承接一个事件关联多个对象 / 页面跳转信息
- `audit_event_inbox_states` 记录管理员对事件的 `acknowledged / ignored / resolved`

- [ ] **Step 3: 新增统一模型**

至少新增：

- `AuditEvent`
- `AuditEventTarget`
- `AuditEventInboxState`
- `AuditEventFilter`
- `AuditEventSummary`
- `AuditEventRoutingDecision`

- [ ] **Step 4: 跑 repo/model 测试**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestAuditEvent'`  
Expected: schema/model 层测试通过

- [ ] **Step 5: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/backend/migrations/20260323_00_audit_event_stream.sql /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/audit_event.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/audit_event_repo_test.go
git commit -m "feat: add audit event stream schema"
```

### Task 2: 扩展 repo interface，并实现 MySQL / in-memory 事件读写能力

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/audit_event_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/inmemory_repo.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/audit_event_repo_test.go`

- [ ] **Step 1: 扩展 repo interface**

至少暴露：

- `AdminCreateAuditEvent`
- `AdminBatchCreateAuditEvents`
- `AdminListAuditEvents`
- `AdminGetAuditEvent`
- `AdminUpdateAuditEventStatus`
- `AdminGetAuditEventSummary`

- [ ] **Step 2: 实现 MySQL repo**

要求：

- 事件本体、目标对象、inbox 状态一次性写入
- 查询支持分页与多条件过滤
- summary 支持：`unread_count / warning_count / critical_count / latest_items`
- detail 支持回传 `targets[]`、`payload`、`routing`、`status_history`

- [ ] **Step 3: 实现 in-memory repo**

要求：

- 测试环境不依赖 migration 也能跑
- 预置 demo 事件样本，覆盖 job / publish / provider / truth / workflow 五类对象

- [ ] **Step 4: 跑 repo 全量回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo/...`  
Expected: 新 repo 测试通过，现有 snapshot/runtime/market admin 测试不回归

- [ ] **Step 5: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/audit_event_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/inmemory_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/audit_event_repo_test.go
git commit -m "feat: add audit event repo"
```

## 包2：事件写入入口与通知路由聚合

### Task 3: 新增统一 audit event service 与 Admin 路由

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/audit_event_service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/audit_event_repo_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler_test.go`

- [ ] **Step 1: 为 admin audit events API 写 failing handler/service 测试**

第一轮固定新增接口：

- `GET /api/v1/admin/audit/events`
- `GET /api/v1/admin/audit/events/:id`
- `PUT /api/v1/admin/audit/events/:id/status`
- `GET /api/v1/admin/audit/events/summary`

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler -run 'TestAdminAuditEvent'`  
Expected: 因 handler / route / service 尚未接入而失败

- [ ] **Step 2: 新增统一 routing / aggregation service**

要求：

- 输入事件后返回去向：`INBOX / REVIEW_CENTER / SYSTEM_JOBS / DATA_SOURCES / DIGEST_ONLY`
- 第一轮先做规则表驱动，不做复杂策略引擎
- 聚合键至少支持：`event_type + object_type + object_id + trade_date`

- [ ] **Step 3: 在 handler 中暴露统一 inbox API**

要求：

- `list` 返回事件列表 + summary
- `detail` 返回 payload、targets、建议跳转对象
- `status update` 支持 `acknowledged / ignored / resolved`
- 权限建议挂在 `audit.view` / `audit.edit`，不新增一套独立 permission 前缀

- [ ] **Step 4: 跑 handler/service 回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler/... ./internal/growth/service/...`  
Expected: 新 audit events handler/service 测试通过

- [ ] **Step 5: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/backend/internal/growth/service/audit_event_service.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go /Users/gjhan21/cursor/sercherai/backend/router/router.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler_test.go
git commit -m "feat: expose admin audit event inbox api"
```

### Task 4: 把现有关键行为统一写入 audit event stream

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_job_snapshot_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_runtime_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/market_data_admin_service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_archive_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_job_snapshot_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_runtime_repo_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go`

- [ ] **Step 1: 接 publish / replay / default policy materialize 事件**

要求：

- `AdminPublishStrategyEngineJob` 成功后写 `STRATEGY_PUBLISH_REPLAYED` 或 `STRATEGY_PUBLISH_FORCED`
- `strategy_engine.publish_policy.*` 默认策略自动落库时写 `POLICY_AUTO_MATERIALIZED`
- 事件 payload 至少带：`job_id / publish_id / job_type / operator / force_publish / warning_count / vetoed_assets`

- [ ] **Step 2: 接 market routing / truth 事件**

要求：

- provider fallback、truth 重建失败、truth 重建 warning 升级进入统一事件流
- market 事件 payload 至少带：`asset_class / data_kind / selected_source / fallback_chain / warnings / trade_date`

- [ ] **Step 3: 接 workflow / system job / research failure 事件**

要求：

- review 驳回 / 回滚进入 `REVIEW_REJECTED` / `REVIEW_ROLLED_BACK`
- scheduler trigger / retry 失败、研究运行失败进入 `RESEARCH_RUN_FAILED`
- warning 升级写 `RESEARCH_WARNING_ESCALATED`

- [ ] **Step 4: 跑关键链路回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: publish/replay、policy materialize、market truth、workflow、scheduler 相关测试通过

- [ ] **Step 5: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_job_snapshot_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_runtime_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_admin.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/service/market_data_admin_service.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_archive_test.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_job_snapshot_test.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_runtime_repo_test.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go
git commit -m "feat: route strategy and market events into audit stream"
```

## 包3：Admin 消息中心第一版

### Task 5: 复用现有页面升级统一 inbox 视图与摘要卡片

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/WorkflowMessagesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/AuditLogsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/UserMessagesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js`

- [ ] **Step 1: 在 `admin.js` 增加统一事件 API**

至少新增：

- `listAuditEvents(params)`
- `getAuditEvent(id)`
- `getAuditEventSummary(params)`
- `updateAuditEventStatus(id, status)`

要求：

- 旧 `listWorkflowMessages`、`listUserMessages`、`listSchedulerJobRuns`、`listOperationLogs` API 保留
- 新 API 作为“统一审计流”补充，不破坏旧页面

- [ ] **Step 2: 把 `WorkflowMessagesView` 升级为第一版 inbox 主视图**

要求：

- 默认增加统一事件列表区块，不删除旧 workflow message 表格
- 支持按 `event_type / severity / object_type / status` 过滤
- 点击事件可打开 detail drawer，展示 payload、targets、跳转按钮

- [ ] **Step 3: 在其他页面增加 audit summary 卡片或跳转锚点**

要求：

- `SystemJobsView`：展示与当前 job_name 相关的最新 audit events
- `ReviewCenterView`：展示 review 事件摘要与驳回/回滚跳转
- `DataSourcesView`：展示 provider fallback / truth failure 摘要
- `AuditLogsView`：增加“操作日志 vs 统一事件流”关联入口
- `UserMessagesView`：保留用户直发消息，但可跳转到统一 inbox 查看系统事件

- [ ] **Step 4: 跑 Admin 构建验证**

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: 统一 inbox API 与现有页面升级后编译通过

- [ ] **Step 5: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/admin/src/api/admin.js /Users/gjhan21/cursor/sercherai/admin/src/views/WorkflowMessagesView.vue /Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue /Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue /Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue /Users/gjhan21/cursor/sercherai/admin/src/views/AuditLogsView.vue /Users/gjhan21/cursor/sercherai/admin/src/views/UserMessagesView.vue /Users/gjhan21/cursor/sercherai/admin/src/router/index.js
git commit -m "feat: add admin audit inbox views"
```

## 包4：现有模块接入与回归收口

### Task 6: 让旧入口开始回读统一事件摘要，并固化验收顺序

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题E-审计通知与系统消息平台.md`
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/未完成专题细化任务表.md`
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/README.md`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/audit_event_repo_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go`

- [ ] **Step 1: 验证五类对象都能看到统一事件摘要**

覆盖：

- strategy job publish / replay
- default publish policy materialize
- workflow review reject / rollback
- scheduler research failure
- provider fallback / truth rebuild failure

- [ ] **Step 2: 验证旧入口仍可继续使用**

至少验证：

- `WorkflowMessagesView` 旧读写不报错
- `UserMessagesView` 仍能发站内消息
- `SystemJobsView` 原有 trigger / retry / metrics 正常
- `AuditLogsView` 现有 CSV 导出不回归

- [ ] **Step 3: 作为第一轮收官命令**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: backend 全量回归通过

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: Admin 构建通过

- [ ] **Step 4: 回写文档状态**

至少回写：

- 专题E主文档的“当前基线 / 完成标准 / 第一轮完成情况”
- 总任务表里专题E状态
- README 中专题E入口与依赖关系

## 第一轮默认验收清单

### Backend

- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo/...`
- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler/...`
- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`

### Admin

- `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`

### 手动验证

- 打开 `/Users/gjhan21/cursor/sercherai/admin/src/views/WorkflowMessagesView.vue` 对应页面，确认 unified inbox 可筛选事件
- 打开 `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue`，确认 job 详情可看到 audit summary
- 打开 `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`，确认 provider fallback / truth failure 可见
- 打开 `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`，确认驳回/回滚可见

## 第一轮明确不并入

- 第三方 IM 编排
- 对外 webhook 市场
- 统一 digest 调度中心
- 完整多租户消息偏好
- 替代现有 `operation_logs` 的全量审计职责

## 完成后文档回写要求

第一轮完成后，至少回写：

- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题E-审计通知与系统消息平台.md`
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/README.md`
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/未完成专题细化任务表.md`
