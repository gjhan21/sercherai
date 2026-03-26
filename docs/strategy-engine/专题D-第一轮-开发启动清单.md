# 专题D 第一轮（回测、实验与运营联动）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 evaluation、compare、leaderboard、experiment analytics 与运营入口基础上，把专题D细化成可直接按现有分析链路开工的文件级任务表。

**Architecture:** 不另起“全新分析平台”，而是在现有 `selection evaluation -> compare -> experiment summary -> client analytics` 主链之上，新增 backtest schema、研究效果仓库和运营联动聚合层。第一轮优先复用当前已有的 `AdminGetExperimentAnalyticsSummary`、stock/futures evaluation、compare、leaderboard 和客户端 experiment attribution 入口，不凭空发明未落地主链。

**Tech Stack:** Go, Gin, MySQL, Vue 3, npm build, Go test

---

最后更新: 2026-03-23  
状态: Planned

## 文档定位

这份清单是：

- [专题D-回测实验与自动化运营联动.md](/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题D-回测实验与自动化运营联动.md)

对应的第一轮可执行实施计划。

它不是专题D全部收官文档，而是“按当前 evaluation / experiment / attribution / analytics 主链直接开工”的文件级任务表。默认顺序固定为：

1. 包0：分析基线盘点与对象冻结
2. 包1：backtest schema 与 run 绑定
3. 包2：evaluation warehouse
4. 包3：experiment / attribution v2
5. 包4：运营联动看板与自动化动作接口

## 开工前提

- 专题A 和专题E 第一轮已完成，治理摘要与审计事件流可作为后续联动输入
- 当前 stock/futures evaluation、compare、leaderboard 和 experiment summary 已经在线可用
- 第一轮不做自动投放、不做通用 BI、不做通用工作流引擎
- 人工审核发布仍然是默认，不修改核心发布策略

## 包0：分析基线盘点与对象冻结

### Task 0: 锁定第一轮分析对象与现有入口

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题D-第一轮-开发启动清单.md`
- Review: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/stock_selection_admin_handler.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/futures_selection_admin_handler.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/experiment_analytics_test.go`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/ExperimentAnalyticsView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionEvaluationView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionEvaluationView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/lib/growth-experiments.js`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/lib/growth-analytics.js`

- [ ] **Step 1: 冻结第一轮分析对象**

范围固定为：

- `run`
- `publish_version`
- `template`
- `profile`
- `cohort`
- `order_attribution`

- [ ] **Step 2: 标记当前散落能力位置**

至少记录：

- stock/futures compare / evaluation / leaderboard 入口
- experiment summary 与客户端 experiment attribution 入口
- publish/version-history 与 evaluation meta 之间的连接点
- 系统消息 / 作业页与运营动作的可复用入口

- [ ] **Step 3: 固化回归基线命令**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: evaluation、compare、experiment summary 相关测试保持通过

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: evaluation 与 experiment analytics 页面构建保持通过

## 包1：backtest schema 与 run 绑定

### Task 1: 新增回测配置与报告对象模型

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260323_03_research_backtest_warehouse.sql`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/research_backtest.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/research_backtest_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_run.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/futures_selection_run.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/inmemory_repo.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/research_backtest_repo_test.go`

- [ ] **Step 1: 写 failing repo/model 测试**

覆盖：

- backtest config 入库
- benchmark / cost assumption / window 读取
- run 绑定 backtest report reference
- stock / futures run 同步回读 backtest summary

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestResearchBacktest'`  
Expected: 因 backtest schema / repo 尚不存在而失败

- [ ] **Step 2: 新增 migration 与模型**

第一轮至少落以下对象：

- `research_backtest_configs`
- `research_backtest_runs`
- `research_backtest_reports`

要求：

- 只增不改，不影响现有 run / publish / evaluation 表
- 先以“离线报告对象”承接，不新增独立执行服务
- run 模型只新增 reference 字段，不强制历史回填

- [ ] **Step 3: 跑后端 repo 回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo/...`  
Expected: 新 repo 测试通过

## 包2：evaluation warehouse

### Task 2: 收口研究效果指标仓库与聚合接口

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/research_evaluation_warehouse.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/research_evaluation_warehouse_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/stock_selection_service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/futures_selection_service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/stock_selection_admin_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/futures_selection_admin_handler.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/research_evaluation_warehouse_repo_test.go`

- [ ] **Step 1: 为 run group / version group 聚合写 failing tests**

覆盖：

- 收益 / 回撤 / 波动 / 命中率 / 暴露 / 行业集中度 聚合
- stock / futures 双资产聚合摘要
- compare / leaderboard 继续兼容旧接口

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestResearchEvaluationWarehouse|TestStockSelectionEvaluation'`  
Expected: 因 warehouse 聚合尚未接入而失败

- [ ] **Step 2: 扩展 repo / service / handler**

要求：

- evaluation warehouse 成为统一聚合底层
- 旧 `evaluation` / `leaderboard` 接口继续可用
- 不要求第一轮就重做所有 Admin 页面，只先增强数据来源

- [ ] **Step 3: 跑后端全量回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: evaluation 与 compare 相关测试通过

## 包3：experiment / attribution v2

### Task 3: 统一 experiment analytics、cohort 与订单归因

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/research_experiment_v2.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/research_experiment_v2_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/user_growth_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/ExperimentAnalyticsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/growth-experiments.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/growth-analytics.js`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/experiment_analytics_test.go`

- [ ] **Step 1: 为 experiment v2 聚合写 failing tests**

覆盖：

- cohort / journey 聚合
- 模板表现对比
- 订单级唯一归因
- experiment summary 接口与客户端埋点兼容

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestExperimentAnalytics'`  
Expected: 因 v2 聚合模型尚未接入而失败

- [ ] **Step 2: 扩展 summary 接口与前端消费**

要求：

- 复用现有 `/admin/market/experiments/summary` 路由组语义
- 客户端 experiment 埋点协议继续保持兼容
- Admin 先显示统一摘要，不要求第一轮重做全部实验 UI

- [ ] **Step 3: 跑后端 + Admin 回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: experiment 与 attribution 相关测试通过

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: Experiment Analytics 页面构建通过

## 包4：运营联动看板与自动化动作接口

### Task 4: 让 Admin 能串起研究效果、用户承接与追踪动作

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/ExperimentAnalyticsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/WorkflowMessagesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/research_ops_action.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/research_ops_action_repo.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/research_ops_action_repo_test.go`

- [ ] **Step 1: 为研究型自动化动作接口写 failing tests**

覆盖：

- 发布后追踪任务
- 研究复盘提醒
- 实验摘要任务
- 系统消息 / 作业页可读取联动摘要

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestResearchOpsAction'`  
Expected: 因自动化动作 repo / API 尚不存在而失败

- [ ] **Step 2: 新增最小联动接口**

要求：

- 只做面向研究平台的动作：提醒、复盘、实验摘要、发布后追踪
- 不实现通用工作流引擎
- 可消费专题E统一审计事件摘要，但不依赖第三方消息系统

- [ ] **Step 3: 跑 Backend / Admin 回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 联动接口与现有研究页面测试通过

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: Admin 构建通过

- [ ] **Step 4: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/admin/src/views/ExperimentAnalyticsView.vue /Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue /Users/gjhan21/cursor/sercherai/admin/src/views/WorkflowMessagesView.vue /Users/gjhan21/cursor/sercherai/admin/src/api/admin.js /Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/research_ops_action.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/research_ops_action_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/research_ops_action_repo_test.go
git commit -m "feat: add research ops linkage actions"
```

## 第一轮默认验收清单

### Backend

- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo/...`
- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`

### Admin

- `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`

## 第一轮明确不并入

- 自动广告投放
- 通用 BI 替代
- 任意工作流编排引擎
- 全量实时回测执行服务

## 完成后文档回写要求

第一轮完成后，至少回写：

- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题D-回测实验与自动化运营联动.md`
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/README.md`
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/未完成专题细化任务表.md`
