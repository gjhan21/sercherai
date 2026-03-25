# 专题E：默认策略自动落库审计通知与系统消息平台

## 目标

把当前散落在操作日志、workflow messages、审核动作、发布动作和数据源告警里的消息能力，收口为统一 `audit event` 真相源，再让消息中心、审核中心、任务中心逐步回读这条主链。

## 第一轮范围

第一轮先做 `E0 / E1 / E2`：

- 建立统一 `admin_audit_events` schema
- 建立 `AdminCreateAuditEvent / AdminListAuditEvents` repo + service + handler 入口
- 先接 4 类关键事件：
  - 默认策略自动落库
  - 策略发布 / 强制发布
  - 审核通过 / 驳回
  - 数据源健康告警 / 恢复
- 保留原有 `operation_logs` 与 `workflow_messages`，不做替换删除

## 当前代码锚点

- 后端模型：`/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/model/admin_audit_event.go`
- 后端 repo：`/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/internal/growth/repo/admin_audit_event_repo.go`
- migration：`/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/migrations/20260324_00_admin_audit_events.sql`
- Admin 读取入口：`/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/backend/router/router.go`
- Admin API：`/Users/gjhan21/cursor/sercherai/.worktrees/topic-a-governance-a0/admin/src/api/admin.js`

## 已完成

- `E0`：统一 audit event schema 已落地，字段包含 `event_domain / event_type / level / module / object_type / object_id / actor_user_id / metadata / dedupe_key`
- `E1`：统一事件入库入口已落地，且默认策略自动落库、策略发布、审核结果、数据源健康事件开始写入 `admin_audit_events`
- 最小 Admin 读取 API 已落地：`GET /api/v1/admin/audit/events`
- `E2`：统一事件流已开始驱动现有 `workflow_messages`
  - 数据源健康告警 / 恢复改为通过 `audit event` 集中路由消息
  - 审核结果通知改为通过 `audit event` 集中路由消息
  - 发布事件可按 `actor_user_id` 形成最小自通知
  - 新增 `GET /api/v1/admin/audit/events/summary` 供后台读取聚合摘要
  - open dedupe event 不再直接吞掉重复写入，而是会把 `occurrence_count / last_seen_at` 合并进已打开事件，并按最新严重级别升级
  - `AuditLogsView` 与 `WorkflowMessagesView` 已开始直接回读统一 `audit summary / open events`
- `E3`：统一消息中心第一刀已落地
  - `WorkflowMessagesView` 已升级为“消息中心”壳，包含 `流程待办 / 开放事件` 双视图
  - `AuditLogsView` 明确退回“审计与操作日志”定位，并提供返回消息中心入口
  - 旧 `workflow_messages` / `operation_logs` 入口继续保留，尚未进入硬切阶段
- `E4`：Review / Jobs / Data Sources 已开始回读统一事件流
  - `ReviewCenterView` 现在展示审核事件摘要，并可跳回统一消息中心处理开放事件
  - `SystemJobsView` 现在展示任务事件摘要，直接暴露 `SCHEDULER_JOB` open events 与严重级别统计
  - `DataSourcesView` 现在展示数据事件摘要，统一暴露 `DATA_SOURCE` open events 与严重级别统计
  - 三个模块都保留原有主体页面，只在原位叠加统一事件流摘要，不另起第二套告警台
- `E4` 深化第一刀已落地
  - `WorkflowMessagesView` 的开放事件视图已补 `事件域 / 等级 / 模块 / 对象类型 / 状态` 过滤组合
  - `WorkflowMessagesView` 与 `AuditLogsView` 都已补“跳转对象页”，可直接回到 `Review / Jobs / Data Sources` 等对象入口
  - 消息中心与审计日志都开始显示 `已关闭事件` 聚合，不再只强调 open count
- `E4` 深化第二刀已落地
  - 审计事件对象跳转已从“模块级入口”深化到“对象级深链”：`review_id / run_id / source_key`
  - `ReviewCenterView`、`SystemJobsView`、`DataSourcesView` 已开始消费这些 query deep link 并自动打开对应对象上下文
  - 股票/期货运行中心继续复用既有 `run_id` deep link，不新增第二套跳转协议
- `E4` 深化第三刀已落地
  - `WorkflowMessagesView` 与 `AuditLogsView` 已补 `STRATEGY_JOB / STRATEGY_PUBLISH_POLICY` 对象类型深链
  - `MarketCenterView` 已可消费 `publish_id / view / job_type / policy_id` query，并自动打开股票/期货发布详情或定位发布策略
  - `StrategyEngineConfigPanel` 已支持发布策略高亮与对象定位，消息中心到策略中心的跨对象跳转不再停留在模块级

## 下一步

- `E4` 后续深化：继续补更多对象类型的深链映射与跨对象跳转策略
- 等 Topic E 第一轮 commit 整理后，再决定是否把更多旧入口改为默认跳回统一消息中心
