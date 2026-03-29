# Stock Futures Forecast L3 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 `L1 / L2` explanation 基线之上，为股票推荐与期货策略引入“异步深推演 + 报告资产 + 长期学习回灌”的 `L3` 能力，并保持它始终是增强层，而不是新的平行推荐主系统。

**Architecture:** 继续复用当前 `strategy-engine publish history + report_snapshot + StrategyClientExplanation + evaluation backfill + admin config` 体系，不重造主推荐链。`L3` 新增一条独立的异步深推演运行链：先把高价值标的排进 `forecast_l3` 队列，再由本地编排器构建研究包、调用可替换的深推演 adapter、落库报告与日志、回写长期学习记录，最后把摘要只读接回现有 `insight / version-history / admin` 页面。

**Tech Stack:** Go, Gin, MySQL repo layer, existing strategy-engine artifacts, existing scheduler/system-config patterns, Vue 3 client, Vue 3 admin, current `strategy_client_explanation` and `forecast-admin` helpers.

---

## Scope And Guardrails

- 本计划只覆盖 `L3`，不回头重写 `L1 / L2` 主线。
- `L3` 只允许作为增强层：
  - 不直接替换股票推荐或期货策略的发布主链
  - 不直接修改 candidate ranking / portfolio weight / publish approval 主流程
  - 没有 `L3` 结果时必须优雅降级到现有 `L2`
- `L3` 首期只对少量高价值对象触发，不允许全市场全量自动跑：
  - `ADMIN_MANUAL`
  - `AUTO_PRIORITY`
  - `USER_REQUEST`
- `L3` 首期默认使用本地可执行 adapter，复用现有上下文、研报、事件、L1/L2 explanation、评估与模板能力，不把外部 MiroFish 引擎接入当成本阶段阻塞项。
- `MiroFish` 只借鉴“研究闭环、角色化推演、长期学习”的方法，不照搬其社交世界模拟产品形态。
- `L3` 可以新增独立 admin 工作台，但 client 侧仍优先复用现有阅读链与页面，不重做全站交互。
- 首期允许新增持久化表，因为 `L3` 的异步任务、报告资产、日志与学习记录需要真实落库。

## Current Baseline

- `L1 / L2` 已完成并在 `main`：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l1_config.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l2_config.go`
- 当前已经存在可复用的 strategy-engine 运行产物与模板体系：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_client.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_runtime_repo.go`
- 当前 explanation / history 已经能读到：
  - `memory_feedback`
  - `relationship_snapshot`
  - `scenario_snapshots`
  - `agent_opinions`
  - `scenario_meta`
- 当前已有后验评估与回补入口：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/futures_selection_run_repo.go`
- 当前 admin 已能承接 `L1 / L2` 配置与摘要：
  - `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.js`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`
- 当前 client 已能承接 explanation / history：
  - `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
  - `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
  - `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`

## L3 Delivery Shape

首期 `L3` 交付固定拆成 5 个层次：

1. **运行层**
   - 创建、排队、执行、完成、失败、重跑
   - 手动触发、自动高优先级触发、用户主动请求

2. **研究层**
   - 从现有 publish history / report snapshot / L1/L2 explanation / related events / evaluation 中组装研究包
   - 使用金融专用角色函数集生成深推演结果

3. **资产层**
   - 深推演报告
   - 研究步骤日志
   - 中间证据快照
   - 质量回写记录

4. **读链层**
   - `insight / version-history` 增加 `L3` 摘要与报告引用
   - client 与 admin 可读详细报告

5. **学习层**
   - 将成熟结果回写为长期学习记录
   - 输出命中率、失效提前识别率、角色有效性、偏差标签

## File Map

### Backend Contracts / Config / Persistence

- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260329_01_strategy_forecast_l3.sql`
  - 创建 `L3` 任务、报告、日志、学习记录表，并种子化调度任务、系统配置、权限。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go`
  - 定义 `L3` 任务、报告、日志、学习记录、摘要 contract。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/dto/strategy_forecast_l3.go`
  - 定义 user/admin 请求与响应 DTO。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_config.go`
  - 定义 `L3` runtime config 读取、默认值和准入判断。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`
  - MySQL 层的 runs / reports / logs / learning records 增删查改。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo_test.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_config_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/inmemory_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/strategy_forecast_l3.go`

### Backend Orchestration / Learning

- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`
  - 组装研究包、执行 adapter、落结果、写日志。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator_test.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder.go`
  - 生成结构化深推演报告、摘要、风险补充与报告快照。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder_test.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning.go`
  - 将成熟 run 和后验结果回写成长期学习记录与质量摘要。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/futures_selection_run_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
  - 注册 user/admin 路由，并启动 `L3` dispatcher / quality worker。

### Backend Read-Side Integration

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
  - 为 explanation / history 增加 `L3` 摘要和报告引用字段。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
  - 从最新成功 `L3` run 中拼接摘要；无结果时继续退回 `L2`。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`

### Backend API Layer

- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/user_growth_handler_forecast_l3.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler_forecast_l3.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/forecast_l3_handler_test.go`

### Client

- Create: `/Users/gjhan21/cursor/sercherai/client/src/api/forecast.js`
  - 用户创建 `L3` 请求、查看列表、查看详情。
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/router/index.js`
  - 新增 `L3` 报告详情路由。
- Create: `/Users/gjhan21/cursor/sercherai/client/src/views/ForecastRunView.vue`
  - 渲染深推演报告详情、步骤日志与结果状态。
- Create: `/Users/gjhan21/cursor/sercherai/client/src/views/forecast-run-view.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.test.js`

### Admin

- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/admin-navigation.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/admin-navigation.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
- Create: `/Users/gjhan21/cursor/sercherai/admin/src/views/ForecastLabView.vue`
- Create: `/Users/gjhan21/cursor/sercherai/admin/src/views/forecast-lab-view.test.js`

### Docs / Handoff

- Modify: `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md`
  - 把 `L3` 实施计划文档路径加入后续线程阅读顺序。

## Task 1: Define L3 Contracts And Runtime Config

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/dto/strategy_forecast_l3.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_config.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_config_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`

- [ ] **Step 1: 写 failing tests，先把 `L3` config 与 contract 钉死**

Add tests similar to:

```go
func TestParseForecastL3RuntimeConfig(t *testing.T) {
	config := parseForecastL3RuntimeConfig(map[string]string{
		"growth.forecast_l3.enabled":                    "true",
		"growth.forecast_l3.admin_manual_enabled":       "true",
		"growth.forecast_l3.user_request_enabled":       "false",
		"growth.forecast_l3.max_active_runs":            "6",
		"growth.forecast_l3.max_user_runs_per_day":      "2",
		"growth.forecast_l3.min_priority_threshold":     "0.70",
		"growth.forecast_l3.dispatch.interval_minutes":  "8",
		"growth.forecast_l3.quality.interval_minutes":   "30",
		"growth.forecast_l3.default_engine_key":         "LOCAL_SYNTHESIS",
		"growth.forecast_l3.require_vip_for_full_report":"true",
	})
	if !config.Enabled || !config.AdminManualEnabled {
		t.Fatalf("expected l3 runtime switches to be enabled: %+v", config)
	}
	if config.MaxActiveRuns != 6 || config.MaxUserRunsPerDay != 2 {
		t.Fatalf("expected l3 runtime thresholds to be parsed: %+v", config)
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestParseForecastL3RuntimeConfig'`

Expected: FAIL，因为 `L3` config 文件与结构尚不存在。

- [ ] **Step 2: 在 model 中定义 `L3` 核心枚举和摘要结构**

Implementation notes:
- 至少新增：
  - `StrategyForecastL3Run`
  - `StrategyForecastL3Report`
  - `StrategyForecastL3Log`
  - `StrategyForecastL3LearningRecord`
  - `StrategyForecastL3Summary`
  - `StrategyForecastL3ReportRef`
- 固定枚举：
  - `TargetType`: `STOCK` / `FUTURES`
  - `TriggerType`: `ADMIN_MANUAL` / `AUTO_PRIORITY` / `USER_REQUEST`
  - `Status`: `QUEUED` / `RUNNING` / `SUCCEEDED` / `FAILED` / `CANCELLED`
- `StrategyForecastL3Summary` 必须能直接挂进现有 explanation / history，不要求前端理解整个深推演报告结构。

- [ ] **Step 3: 定义 runtime config 与默认值**

Implementation notes:
- `L3` 首期默认 config 至少包括：
  - `growth.forecast_l3.enabled`
  - `growth.forecast_l3.admin_manual_enabled`
  - `growth.forecast_l3.user_request_enabled`
  - `growth.forecast_l3.auto_priority_enabled`
  - `growth.forecast_l3.client_read_enabled`
  - `growth.forecast_l3.require_vip_for_full_report`
  - `growth.forecast_l3.max_active_runs`
  - `growth.forecast_l3.max_runs_per_day`
  - `growth.forecast_l3.max_user_runs_per_day`
  - `growth.forecast_l3.min_priority_threshold`
  - `growth.forecast_l3.dispatch.enabled`
  - `growth.forecast_l3.dispatch.interval_minutes`
  - `growth.forecast_l3.quality.enabled`
  - `growth.forecast_l3.quality.interval_minutes`
  - `growth.forecast_l3.default_engine_key`
- 默认 engine 先固定为 `LOCAL_SYNTHESIS`。

- [ ] **Step 4: 扩展 repo / service interface**

Implementation notes:
- 在 `interfaces.go` 和 `service.go` 中暴露：
  - create / list / get / retry / cancel run
  - list logs
  - list quality summary
  - execute queued runs
  - run quality backfill

- [ ] **Step 5: 运行 targeted tests**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestParseForecastL3RuntimeConfig'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/model/strategy_forecast_l3.go \
  backend/internal/growth/dto/strategy_forecast_l3.go \
  backend/internal/growth/repo/strategy_forecast_l3_config.go \
  backend/internal/growth/repo/strategy_forecast_l3_config_test.go \
  backend/internal/growth/repo/interfaces.go \
  backend/internal/growth/service/service.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: define forecast l3 contracts and config"
```

## Task 2: Add L3 Persistence Schema And Repository

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260329_01_strategy_forecast_l3.sql`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/inmemory_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/scripts/seed_admin_extra.sql`

- [ ] **Step 1: 写 failing repository tests，覆盖创建 / 列表 / 明细 / 日志**

Add tests similar to:

```go
func TestCreateForecastL3RunPersistsQueuedRecord(t *testing.T) {
	repo, mock := newGrowthRepoWithMock(t)
	mock.ExpectExec(`INSERT INTO strategy_forecast_l3_runs`).
		WithArgs(sqlmock.AnyArg(), "STOCK", "600519.SH", "ADMIN_MANUAL", "QUEUED", "admin_001", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	run, err := repo.AdminCreateForecastL3Run(model.StrategyForecastL3CreateInput{
		TargetType:  "STOCK",
		TargetKey:   "600519.SH",
		TriggerType: "ADMIN_MANUAL",
		RequestedBy: "admin_001",
	})
	if err != nil {
		t.Fatalf("AdminCreateForecastL3Run() error = %v", err)
	}
	if run.Status != "QUEUED" {
		t.Fatalf("expected queued run, got %+v", run)
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'ForecastL3Run|ForecastL3Report'`

Expected: FAIL，因为表和 repo 尚未实现。

- [ ] **Step 2: 新增 migration，创建 4 张表 + 2 个调度任务 + 2 个权限**

Implementation notes:
- 新增表：
  - `strategy_forecast_l3_runs`
  - `strategy_forecast_l3_reports`
  - `strategy_forecast_l3_logs`
  - `strategy_forecast_l3_learning_records`
- 新增 scheduler job definitions：
  - `forecast_l3_dispatch_pending`
  - `forecast_l3_quality_backfill`
- 新增权限：
  - `forecast_l3.view`
  - `forecast_l3.edit`
- 新增默认 system configs，对应 Task 1 的 config keys。

- [ ] **Step 3: 实现 MySQL repo 与 in-memory fallback**

Implementation notes:
- `strategy_forecast_l3_repo.go` 至少实现：
  - `AdminCreateForecastL3Run`
  - `UserCreateForecastL3Run`
  - `AdminListForecastL3Runs`
  - `UserListForecastL3Runs`
  - `GetForecastL3RunDetail`
  - `AppendForecastL3RunLog`
  - `PersistForecastL3Report`
  - `AdminRetryForecastL3Run`
  - `AdminCancelForecastL3Run`
- `inmemory_repo.go` 不要求完整业务模拟，但必须返回结构合法的空结果，确保 handler/unit tests 不炸。

- [ ] **Step 4: 为 run detail 补齐最小聚合**

Implementation notes:
- detail 至少返回：
  - run 基本信息
  - report 摘要
  - report markdown/html
  - steps logs
  - learning summary 占位

- [ ] **Step 5: 运行 repository tests**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'ForecastL3Run|ForecastL3Report'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/migrations/20260329_01_strategy_forecast_l3.sql \
  backend/internal/growth/repo/strategy_forecast_l3_repo.go \
  backend/internal/growth/repo/strategy_forecast_l3_repo_test.go \
  backend/internal/growth/repo/inmemory_repo.go \
  backend/scripts/seed_admin_extra.sql
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add forecast l3 persistence"
```

## Task 3: Add User/Admin Queue APIs For L3 Runs

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/strategy_forecast_l3.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/user_growth_handler_forecast_l3.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler_forecast_l3.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/forecast_l3_handler_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`

- [ ] **Step 1: 写 failing handler tests，先固定权限边界**

Add tests similar to:

```go
func TestCreateForecastL3RunRequiresAuth(t *testing.T) {
	handler := newUserGrowthHandlerForTest(t)
	router := gin.New()
	router.POST("/api/v1/forecast/runs", handler.CreateForecastL3Run)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/runs", strings.NewReader(`{"target_type":"STOCK","target_key":"600519.SH"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler -run 'ForecastL3'`

Expected: FAIL，因为 handler 和 route 尚未存在。

- [ ] **Step 2: 定义首期 user/admin API**

Recommended routes:
- user:
  - `POST /api/v1/forecast/runs`
  - `GET /api/v1/forecast/runs`
  - `GET /api/v1/forecast/runs/:id`
- admin:
  - `GET /api/v1/admin/forecast/runs`
  - `POST /api/v1/admin/forecast/runs`
  - `GET /api/v1/admin/forecast/runs/:id`
  - `POST /api/v1/admin/forecast/runs/:id/retry`
  - `POST /api/v1/admin/forecast/runs/:id/cancel`
  - `GET /api/v1/admin/forecast/quality`

- [ ] **Step 3: 在 service 层实现准入校验与限流判断**

Implementation notes:
- user 请求至少校验：
  - `growth.forecast_l3.enabled = true`
  - `growth.forecast_l3.user_request_enabled = true`
  - 当前用户是否登录
  - 超过 `max_user_runs_per_day` 时拒绝
- auto/admin 请求至少校验：
  - `max_active_runs`
  - `min_priority_threshold`
  - `target_type / target_key` 是否可映射到现有推荐/策略上下文

- [ ] **Step 4: 让 create API 只负责排队，不在请求内执行重型逻辑**

Implementation notes:
- create 成功只返回：
  - `run_id`
  - `status = QUEUED`
  - `estimated_next_action`
- 不允许在 handler 内直接跑深推演正文。

- [ ] **Step 5: 运行 handler tests**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler -run 'ForecastL3'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/service/strategy_forecast_l3.go \
  backend/internal/growth/handler/user_growth_handler_forecast_l3.go \
  backend/internal/growth/handler/admin_growth_handler_forecast_l3.go \
  backend/internal/growth/handler/forecast_l3_handler_test.go \
  backend/router/router.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add forecast l3 queue apis"
```

## Task 4: Implement L3 Orchestrator, Report Builder, And Scheduler Workers

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator_test.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_runtime_repo.go`

- [ ] **Step 1: 写 failing tests，固定研究包、角色集和 fallback 行为**

Add tests similar to:

```go
func TestExecuteForecastL3RunBuildsReportAndLogs(t *testing.T) {
	repo := newForecastL3RepoForTest(t)
	run := seedQueuedForecastL3Run(t, repo, "STOCK", "600519.SH")

	result, err := repo.ExecuteForecastL3Run(run.ID)
	if err != nil {
		t.Fatalf("ExecuteForecastL3Run() error = %v", err)
	}
	if result.Status != "SUCCEEDED" {
		t.Fatalf("expected succeeded run, got %+v", result)
	}
	if len(result.Logs) == 0 || result.Report.Ref.ReportID == "" {
		t.Fatalf("expected logs and report ref, got %+v", result)
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'ExecuteForecastL3Run|BuildForecastL3Report'`

Expected: FAIL，因为编排器与报告构造器尚未实现。

- [ ] **Step 2: 实现研究包组装器**

Implementation notes:
- 研究包输入固定来自现有事实源：
  - `findStrategyEngineAssetContext`
  - `StrategyClientExplanation`
  - `report_snapshot / publish_payloads`
  - related events / event evidence
  - stock/futures evaluation summary
  - `scenario_templates / agents / publish_policy`
- 输出 `research pack` 至少包含：
  - 目标对象基础信息
  - 当前有效理由与历史失效理由
  - `L2` 三情景与 veto 信息
  - 关联事件与研报摘要
  - 评估回看摘要

- [ ] **Step 3: 实现可替换 adapter，首期先提供 `LOCAL_SYNTHESIS`**

Implementation notes:
- adapter interface 至少定义：
  - `RunDeepForecast(ctx, pack) (report, logs, error)`
- `LOCAL_SYNTHESIS` 先做可执行版本：
  - 股票角色：`INDUSTRY` / `FLOW` / `EVENT` / `RISK` / `MACRO`
  - 期货角色：`SUPPLY_DEMAND` / `HEDGE` / `SPEC_FLOW` / `MACRO` / `RISK`
- 每个角色输出：
  - 立场
  - 置信度
  - 关键依据
  - 风险或否决理由
- 本地 adapter 不要求调用外部服务，但必须保留后续扩展外部引擎的接口位置。

- [ ] **Step 4: 实现结构化报告与日志**

Implementation notes:
- `strategy_forecast_l3_report_builder.go` 必须生成：
  - executive summary
  - scenario map
  - trigger checklist
  - invalidation signals
  - role disagreement summary
  - action guidance
  - markdown/html
- 日志步骤最少固定为：
  - `LOAD_CONTEXT`
  - `BUILD_RESEARCH_PACK`
  - `RUN_DEEP_FORECAST`
  - `BUILD_REPORT`
  - `PERSIST_REPORT`

- [ ] **Step 5: 注册两个 worker**

Implementation notes:
- 在 `router.go` 里按现有模式启动：
  - `forecast_l3_dispatch_pending`
  - `forecast_l3_quality_backfill`
- 每次执行都写 `scheduler_job_runs`
- worker config 从 `growth.forecast_l3.dispatch.*` 与 `growth.forecast_l3.quality.*` 读取

- [ ] **Step 6: 运行 orchestrator tests**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'ExecuteForecastL3Run|BuildForecastL3Report'`

Expected: PASS

- [ ] **Step 7: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go \
  backend/internal/growth/repo/strategy_forecast_l3_orchestrator_test.go \
  backend/internal/growth/repo/strategy_forecast_l3_report_builder.go \
  backend/internal/growth/repo/strategy_forecast_l3_report_builder_test.go \
  backend/internal/growth/repo/strategy_engine_runtime_repo.go \
  backend/router/router.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add forecast l3 orchestration"
```

## Task 5: Wire L3 Summary Into Explanation / History And Client Surfaces

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`
- Create: `/Users/gjhan21/cursor/sercherai/client/src/api/forecast.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/router/index.js`
- Create: `/Users/gjhan21/cursor/sercherai/client/src/views/ForecastRunView.vue`
- Create: `/Users/gjhan21/cursor/sercherai/client/src/views/forecast-run-view.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.test.js`

- [ ] **Step 1: 写 failing tests，钉死 `L3` 摘要字段和前端 helper**

Add tests similar to:

```go
func TestBuildStockStrategyExplanationCarriesL3Summary(t *testing.T) {
	explanation := repo.buildStockStrategyExplanation(item, detail)
	if explanation.DeepForecastSummary.RunID == "" {
		t.Fatalf("expected deep forecast summary on explanation, got %+v", explanation)
	}
}
```

```js
test("buildStrategyDeepForecastSummary returns primary scenario and report status", () => {
  const summary = buildStrategyDeepForecastSummary({
    deep_forecast_summary: {
      run_id: "l3_001",
      status: "SUCCEEDED",
      primary_scenario: "base",
      executive_summary: "主逻辑延续，但需警惕需求侧验证。"
    }
  });
  assert.equal(summary.status, "SUCCEEDED");
  assert.match(summary.note, /base/i);
});
```

Run:
- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'DeepForecastSummary'`
- `cd /Users/gjhan21/cursor/sercherai/client && node --test src/lib/strategy-version.test.js`

Expected: FAIL，因为 `L3` 摘要字段和 helper 还不存在。

- [ ] **Step 2: 在 explanation / history 中新增非破坏性 `L3` 摘要**

Implementation notes:
- 在 `StrategyClientExplanation` 与 `StrategyVersionHistoryItem` 中新增：
  - `DeepForecastSummary`
  - `DeepForecastReportRef`
- 只挂摘要，不把整份报告直接塞进现有 insight/history payload。
- summary 至少包含：
  - `run_id`
  - `status`
  - `primary_scenario`
  - `executive_summary`
  - `updated_at`

- [ ] **Step 3: explanation 构建时优先加载最近成功或运行中的 `L3` run**

Implementation notes:
- 只读取与当前 `asset_key + trade_date / publish_id` 最相关的 run。
- 没有 `L3` 时保持当前 `L2` 行为不变。
- 有运行中的 `L3` 时显示“深推演进行中”状态，但不阻塞页面。

- [ ] **Step 4: 增加 client 详情页与入口**

Implementation notes:
- 新增 `ForecastRunView.vue`：
  - 顶部显示状态、对象、更新时间
  - 中段显示 executive summary、情景路径、触发/失效条件、角色分歧
  - 下段显示步骤日志与学习回写摘要
- `StrategyView.vue`：
  - 有 `L3` 摘要时显示“查看深推演”
  - 满足条件且已登录时显示“发起深推演”
- `RecommendationArchiveView.vue`：
  - 历史版本条目中显示 `L3` 是否已有报告

- [ ] **Step 5: 运行 backend/client verification**

Run:
- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'DeepForecastSummary'`
- `cd /Users/gjhan21/cursor/sercherai/client && node --test src/lib/strategy-version.test.js src/views/forecast-run-view.test.js`
- `cd /Users/gjhan21/cursor/sercherai/client && npm run build`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/model/strategy_client_explanation.go \
  backend/internal/growth/repo/strategy_client_explanation.go \
  backend/internal/growth/repo/strategy_client_explanation_test.go \
  client/src/api/forecast.js \
  client/src/router/index.js \
  client/src/views/ForecastRunView.vue \
  client/src/views/forecast-run-view.test.js \
  client/src/views/StrategyView.vue \
  client/src/views/RecommendationArchiveView.vue \
  client/src/lib/strategy-version.js \
  client/src/lib/strategy-version.test.js
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: surface forecast l3 in client read chain"
```

## Task 6: Add Learning Writeback, Quality Calibration, And Admin Forecast Lab

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/futures_selection_run_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/admin-navigation.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/admin-navigation.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
- Create: `/Users/gjhan21/cursor/sercherai/admin/src/views/ForecastLabView.vue`
- Create: `/Users/gjhan21/cursor/sercherai/admin/src/views/forecast-lab-view.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md`

- [ ] **Step 1: 写 failing tests，固定学习回写与 admin 摘要**

Add tests similar to:

```go
func TestRunForecastL3QualityBackfillWritesLearningRecord(t *testing.T) {
	count, err := repo.AdminRunForecastL3QualityBackfill(20)
	if err != nil {
		t.Fatalf("AdminRunForecastL3QualityBackfill() error = %v", err)
	}
	if count == 0 {
		t.Fatalf("expected learning records to be written")
	}
}
```

```js
test("parseForecastAdminConfigMap reads l3 switches", () => {
  const config = parseForecastAdminConfigMap({
    "growth.forecast_l3.enabled": "true",
    "growth.forecast_l3.max_active_runs": "8"
  });
  assert.equal(config.l3Enabled, true);
  assert.equal(config.l3MaxActiveRuns, 8);
});
```

Run:
- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'ForecastL3Quality|LearningRecord'`
- `cd /Users/gjhan21/cursor/sercherai/admin && node --test src/lib/forecast-admin.test.js`

Expected: FAIL

- [ ] **Step 2: 实现质量回写**

Implementation notes:
- `strategy_forecast_l3_learning.go` 要把成熟 `L3` run 与后验结果对齐：
  - 股票：复用 `stock_selection_run_evaluation_backfill`
  - 期货：复用 `futures_selection_run_repo.go` 当前可得的表现 / guidance / invalidation 信息
- 至少产出：
  - `scenario_hit`
  - `trigger_hit`
  - `invalidation_early`
  - `bias_label`
  - `role_effectiveness`

- [ ] **Step 3: 把 `L3` config 和质量摘要接到 admin**

Implementation notes:
- `SystemConfigsView.vue` 新增 `L3` 配置区：
  - 全局开关
  - user/admin/manual 开关
  - max active runs
  - per-user quota
  - dispatcher / quality worker interval
- `ForecastLabView.vue` 新增独立工作台：
  - run 列表
  - 状态筛选
  - 报告详情
  - 研究日志
  - 质量摘要卡
  - 手动重跑 / 取消
- `MarketCenterView.vue` 可增加到 `ForecastLabView` 的跳转入口，但不承载重操作。

- [ ] **Step 4: 更新 handoff 文档**

Implementation notes:
- 在线程交接文档中加入：
  - `L3` 实施计划路径
  - 推荐阅读顺序从 spec -> plan -> code baseline

- [ ] **Step 5: 运行 backend/admin verification**

Run:
- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'ForecastL3Quality|LearningRecord'`
- `cd /Users/gjhan21/cursor/sercherai/admin && node --test src/lib/forecast-admin.test.js src/lib/admin-navigation.test.js src/views/forecast-lab-view.test.js`
- `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/strategy_forecast_l3_learning.go \
  backend/internal/growth/repo/strategy_forecast_l3_learning_test.go \
  backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go \
  backend/internal/growth/repo/futures_selection_run_repo.go \
  admin/src/api/admin.js \
  admin/src/router/index.js \
  admin/src/lib/admin-navigation.js \
  admin/src/lib/admin-navigation.test.js \
  admin/src/lib/forecast-admin.js \
  admin/src/lib/forecast-admin.test.js \
  admin/src/views/SystemConfigsView.vue \
  admin/src/views/MarketCenterView.vue \
  admin/src/views/ForecastLabView.vue \
  admin/src/views/forecast-lab-view.test.js \
  docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add forecast l3 learning and admin lab"
```

## Task 7: Full Verification And Exit Review

**Files:**
- Verify only

- [ ] **Step 1: 跑完整 backend 相关测试**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo ./internal/growth/handler`

Expected: PASS

- [ ] **Step 2: 跑 client / admin 构建**

Run:
- `cd /Users/gjhan21/cursor/sercherai/client && npm run build`
- `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`

Expected: PASS

- [ ] **Step 3: 手工验收 `L3` 核心链**

Checklist:
- admin 能创建 `L3` run，状态从 `QUEUED -> RUNNING -> SUCCEEDED / FAILED`
- worker 能写 scheduler job run
- 成功 run 能生成 report、log、summary
- strategy / archive 页能读到 `L3` 摘要
- 无 `L3` 结果时页面仍显示 `L2`
- 质量回写能生成 learning summary

- [ ] **Step 4: Exit review**

Before considering `L3` MVP complete, confirm:
- 没有直接改动推荐排序或发布主链
- 没有把 `L3` 做成必须存在才能打开页面的阻塞项
- 没有把外部引擎适配做成首期硬依赖
- `L3` 只对少量对象触发，而不是误跑全市场
- 研究日志和报告可审计，不只是最终一段大文本

- [ ] **Step 5: Final commit**

```bash
git -C /Users/gjhan21/cursor/sercherai commit -am "docs: close out forecast l3 implementation rollout"
```
