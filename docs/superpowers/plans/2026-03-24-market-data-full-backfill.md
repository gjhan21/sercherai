# Market Data Full Backfill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有多源市场数据和任务中心基础上，落地一套覆盖 `STOCK / INDEX / ETF / LOF / CBOND` 的全市场数据回填体系，打通 `universe -> master -> quotes -> daily_basic -> moneyflow -> truth -> coverage summary` 全链路，并让后台能发起、查看、重试与治理。

**Architecture:** 继续复用现有 `scheduler_job_definitions`、`scheduler_job_runs`、`market_instruments`、`market_instrument_source_facts`、`market_symbol_aliases`、`market_daily_bars`、`market_daily_bar_truth`、`stock_daily_basic`、`stock_moneyflow_daily` 和 `market_data_quality_logs`，不新建第二套任务平台。核心做法是新增 `market-data` 域的 universe 快照、回填总单、批次明细与治理摘要接口，再把现有主数据、多源同步、truth 重建和任务中心 UI 编排成正式后台能力。

**Tech Stack:** Go, Gin, MySQL, Vue 3 Admin, Element Plus, Tushare, AKShare, TickerMD, existing scheduler job infrastructure

---

## Scope And Assumptions

- 本轮资产范围固定为 `STOCK / INDEX / ETF / LOF / CBOND`。
- 全量与增量必须共用同一套回填执行链，只是请求参数与 batch 窗口不同。
- `STOCK` 一期正式支持 `quotes + daily_basic + moneyflow`。
- `INDEX / ETF / LOF / CBOND` 一期必须支持 universe、master、quotes 与 truth；`daily_basic / moneyflow` 默认可显式 `SKIPPED`，不作为失败处理。
- 现有 `/api/v1/admin/stocks/quotes/sync` 与 `/api/v1/admin/stocks/quotes/rebuild-derived-truth` 兼容接口保留，但新的多资产回填入口统一落在 `/api/v1/admin/market-data/*`。
- 不引入新消息队列、分布式任务系统或 BFF；继续依赖现有 scheduler 与 MySQL 状态表。

## Current Baseline

- 现有股票行情同步空 `symbols` 仍回退到 15 只样本股，逻辑在 `backend/internal/growth/repo/market_data_multi_source.go` 与 `backend/internal/growth/repo/mysql_repo.go`。
- `fetchStockDailyBasicsFromTushare`、`fetchStockMoneyflowsFromTushare` 已存在于 `backend/internal/growth/repo/mysql_repo.go`，但尚未形成正式 public API 与后台任务。
- 现有 scheduler 只记录 job definitions、job runs 和 news sync 明细；尚无市场全量回填总单与批次明细表。
- `admin/src/views/DataSourcesView.vue` 具备质量日志、truth 摘要和股票/期货同步面板。
- `admin/src/views/SystemJobsView.vue` 具备任务定义、运行记录、自动重试与任务触发入口，但没有“市场数据全量回填”的业务视角。

## File Map

### Backend Schema / Models / DTO

- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260324_01_market_data_full_backfill.sql`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/dto/market_data.go`

### Backend Repo / Service / Handler / Router

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`

### Admin

- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/market-data-admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/system-jobs-admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue`

### Tests

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_status_truth_admin_test.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_backfill_repo_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/market-data-admin.test.js`

## Task 1: Add Schema, Types, And Contracts

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260324_01_market_data_full_backfill.sql`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/dto/market_data.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go`

- [ ] **Step 1: 写 failing tests，覆盖新的 backfill 请求 DTO 和详情查询接口缺失**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler -run 'Test.*MarketData.*Backfill|Test.*Coverage'`
Expected: FAIL，提示缺少 market-data backfill 请求结构、响应字段或 handler 路由不存在。

- [ ] **Step 2: 在 migration 中新增 universe、总单、明细三组表和 scheduler job seeds**

Implementation notes:
- 创建表：
  - `market_universe_snapshots`
  - `market_universe_snapshot_items`
  - `market_backfill_runs`
  - `market_backfill_run_details`
- 为 `market_backfill_run_details` 加联合索引：
  - `(run_id, stage, asset_type, status)`
  - `(scheduler_run_id, stage, status)`
- 为 `market_universe_snapshot_items` 加联合索引：
  - `(snapshot_id, asset_type, instrument_key)`
- 在同一 migration 中插入 scheduler jobs：
  - `market_data_full_backfill`
  - `market_data_incremental_sync`
  - `market_data_truth_rebuild`

Recommended schema sketch:

```sql
CREATE TABLE IF NOT EXISTS market_backfill_runs (
  id varchar(64) PRIMARY KEY,
  scheduler_run_id varchar(64) NOT NULL,
  run_type varchar(32) NOT NULL,
  asset_scope json NOT NULL,
  trade_date_from date NULL,
  trade_date_to date NULL,
  source_key varchar(64) NOT NULL,
  batch_size int NOT NULL DEFAULT 200,
  universe_snapshot_id varchar(64) NOT NULL,
  status varchar(32) NOT NULL,
  current_stage varchar(32) NOT NULL,
  stage_progress_json json NULL,
  summary_json json NULL,
  error_message varchar(1024) NULL,
  created_by varchar(64) NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  finished_at datetime NULL
);
```

- [ ] **Step 3: 在 model 与 dto 中新增类型，统一阶段、状态和支持矩阵语义**

Implementation notes:
- 在 `model/models.go` 中新增：
  - `MarketUniverseSnapshot`
  - `MarketUniverseSnapshotItem`
  - `MarketBackfillRun`
  - `MarketBackfillRunDetail`
- 在 `model/market_data_admin.go` 中新增：
  - `MarketCoverageSummary`
  - `MarketCoverageSummaryAssetItem`
  - `MarketBackfillRunSummary`
- 在 `dto/market_data.go` 中新增：
  - `MarketDataBackfillRequest`
  - `MarketDataBackfillRetryRequest`
  - `MarketDataSyncRequest`
- 固定枚举：
  - stages: `UNIVERSE`, `MASTER`, `QUOTES`, `DAILY_BASIC`, `MONEYFLOW`, `TRUTH`, `COVERAGE_SUMMARY`
  - run status: `PENDING`, `RUNNING`, `PARTIAL_SUCCESS`, `SUCCESS`, `FAILED`, `CANCELLED`
  - detail status: `PENDING`, `RUNNING`, `SUCCESS`, `FAILED`, `SKIPPED`

Recommended type sketch:

```go
type MarketDataBackfillRequest struct {
    RunType              string   `json:"run_type" binding:"required,oneof=FULL INCREMENTAL REBUILD_ONLY"`
    AssetScope           []string `json:"asset_scope" binding:"required,min=1,dive,oneof=STOCK INDEX ETF LOF CBOND"`
    SourceKey            string   `json:"source_key"`
    TradeDateFrom        string   `json:"trade_date_from"`
    TradeDateTo          string   `json:"trade_date_to"`
    BatchSize            int      `json:"batch_size" binding:"omitempty,gte=20,lte=500"`
    Stages               []string `json:"stages"`
    ForceRefreshUniverse bool     `json:"force_refresh_universe"`
    RebuildTruthAfterSync bool    `json:"rebuild_truth_after_sync"`
}
```

- [ ] **Step 4: 在 repo/service 接口层补齐新能力签名**

Implementation notes:
- 在 `repo/interfaces.go` 和 `service/service.go` 中新增：
  - `AdminCreateMarketDataBackfillRun(...)`
  - `AdminListMarketDataBackfillRuns(...)`
  - `AdminGetMarketDataBackfillRun(...)`
  - `AdminListMarketDataBackfillRunDetails(...)`
  - `AdminRetryMarketDataBackfillRun(...)`
  - `AdminListMarketUniverseSnapshots(...)`
  - `AdminGetMarketUniverseSnapshot(...)`
  - `AdminGetMarketCoverageSummary(...)`

- [ ] **Step 5: 运行 targeted tests，确认 contracts 通过**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler -run 'Test.*MarketData.*Backfill|Test.*Coverage'`
Expected: PASS 或进入下一轮 repo 断言失败，但不再报缺少 DTO / model / interface。

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/migrations/20260324_01_market_data_full_backfill.sql \
  backend/internal/growth/model/models.go \
  backend/internal/growth/model/market_data_admin.go \
  backend/internal/growth/dto/market_data.go \
  backend/internal/growth/repo/interfaces.go \
  backend/internal/growth/service/service.go \
  backend/internal/growth/handler/market_data_admin_handler_test.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add market data backfill schema and contracts"
```

## Task 2: Implement Universe Snapshot And Master Sync

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_backfill_repo_test.go`

- [ ] **Step 1: 写 failing tests，覆盖 universe snapshot 创建与读取**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestMarketUniverse|TestMarketInstrument.*Snapshot'`
Expected: FAIL，提示 snapshot 表读写逻辑、全市场资产类型拆分或主数据 enrich 缺失。

- [ ] **Step 2: 在 repo 中实现全市场 universe 生成**

Implementation notes:
- 在 `market_instrument_master_data.go` 中新增：
  - `buildMarketUniverseSnapshot`
  - `fetchMarketUniverseItemsForSource`
  - `normalizeUniverseAssetType`
- universe 生成必须覆盖：
  - `STOCK`
  - `INDEX`
  - `ETF`
  - `LOF`
  - `CBOND`
- `STOCK` 继续优先使用 Tushare `stock_basic`
- 其他资产类型优先按现有 provider 能力抽象，不足时先落 fallback/source fact 最小记录，但必须进入 snapshot

Recommended function sketch:

```go
func (r *MySQLGrowthRepo) AdminBuildMarketUniverseSnapshot(
    sourceKey string,
    assetScope []string,
    operator string,
) (model.MarketUniverseSnapshot, error)
```

- [ ] **Step 3: 基于 snapshot 跑 master sync，并按资产类型分段写 source facts / aliases / truth**

Implementation notes:
- 扩展 `syncMarketInstrumentMasterData`，支持直接消费 `snapshot_id`
- `MASTER` 阶段按资产类型执行，不按全量混跑
- `market_symbol_aliases` 必须始终保留 canonical 映射
- 对主数据缺口写 `market_data_quality_logs`

- [ ] **Step 4: 实现 snapshot 与 universe 查询接口所需 repo 查询**

Implementation notes:
- 在 `mysql_repo.go` 中新增：
  - `AdminListMarketUniverseSnapshots`
  - `AdminGetMarketUniverseSnapshot`
  - `listMarketUniverseSnapshotItems`
- 返回结果中包含每个资产类型的数量和状态分布摘要

- [ ] **Step 5: 运行 repo tests，确认 snapshot 与 master sync 通过**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestMarketUniverse|TestMarketInstrument.*Snapshot|TestMarketInstrument.*Master'`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/market_instrument_master_data.go \
  backend/internal/growth/repo/mysql_repo.go \
  backend/internal/growth/repo/market_instrument_master_data_test.go \
  backend/internal/growth/repo/market_data_backfill_repo_test.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add market universe snapshots and master sync"
```

## Task 3: Implement Quotes, Daily Basic, Moneyflow, And Truth Orchestration

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_status_truth_admin_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_backfill_repo_test.go`

- [ ] **Step 1: 写 failing tests，覆盖 batch quotes、daily_basic、moneyflow、truth rebuild 与 SKIPPED 逻辑**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestMarketDataBackfill|TestMarketDailyBar|TestStockMoneyflow|TestDerivedTruth'`
Expected: FAIL，提示不支持 snapshot-driven batches、detail status `SKIPPED`、或 daily_basic / moneyflow public API 缺失。

- [ ] **Step 2: 把现有 quotes 同步改成 snapshot/batch 驱动，并保留兼容接口**

Implementation notes:
- 在 `market_data_multi_source.go` 中新增 batch 入口：
  - `runMarketQuotesBackfillStage`
  - `buildMarketBackfillBatches`
- 保留 `AdminSyncStockQuotesDetailed`，但内部允许调用新的 batch logic
- `symbols` 为空时不再默认等于生产全市场，仅在 debug/mock 显式路径使用样本池

- [ ] **Step 3: 把 `fetchStockDailyBasicsFromTushare` 和 `fetchStockMoneyflowsFromTushare` 提升为正式 public API**

Implementation notes:
- 在 `mysql_repo.go` 中新增：
  - `AdminSyncMarketDailyBasicDetailed`
  - `AdminSyncMarketMoneyflowDetailed`
- 对 `INDEX / ETF / LOF / CBOND` 一期按支持矩阵写 detail `SKIPPED`
- 对 `STOCK` 正式 upsert 到：
  - `stock_daily_basic`
  - `stock_moneyflow_daily`

Recommended support matrix helper:

```go
func marketAssetEnhancementSupport(assetType string) (dailyBasic bool, moneyflow bool) {
    switch strings.ToUpper(strings.TrimSpace(assetType)) {
    case "STOCK":
        return true, true
    default:
        return false, false
    }
}
```

- [ ] **Step 4: 实现整单阶段编排与 truth rebuild**

Implementation notes:
- 在 repo 中新增：
  - `executeMarketBackfillRun`
  - `runMarketMasterStage`
  - `runMarketQuotesStage`
  - `runMarketDailyBasicStage`
  - `runMarketMoneyflowStage`
  - `runMarketTruthStage`
  - `finalizeMarketCoverageSummaryStage`
- `TRUTH` 阶段允许：
  - 按 `snapshot_id`
  - 按 `asset_type`
  - 按 `trade_date range`
- `PARTIAL_SUCCESS` 判定按 spec 中核心阶段规则执行

- [ ] **Step 5: 运行 repo tests，确认编排链和 truth 通过**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestMarketDataBackfill|TestMarketDailyBar|TestDerivedTruth|TestStockMoneyflow'`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/market_data_multi_source.go \
  backend/internal/growth/repo/mysql_repo.go \
  backend/internal/growth/repo/interfaces.go \
  backend/internal/growth/service/service.go \
  backend/internal/growth/repo/market_data_multi_source_test.go \
  backend/internal/growth/repo/market_status_truth_admin_test.go \
  backend/internal/growth/repo/market_data_backfill_repo_test.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add market backfill execution pipeline"
```

## Task 4: Expose Admin API, Scheduler Integration, And Retry Flow

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go`

- [ ] **Step 1: 写 failing tests，覆盖新的 `/admin/market-data/*` 路由和 retry 行为**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler ./router -run 'Test.*MarketData.*Backfill|Test.*Scheduler.*MarketData'`
Expected: FAIL，提示 market-data 路由、retry 请求、detail 查询或 coverage summary 不存在。

- [ ] **Step 2: 在 handler 中新增总入口与查询入口**

Implementation notes:
- 在 `market_data_admin_handler.go` 中新增：
  - `CreateMarketDataBackfillRun`
  - `ListMarketDataBackfillRuns`
  - `GetMarketDataBackfillRun`
  - `ListMarketDataBackfillRunDetails`
  - `RetryMarketDataBackfillRun`
  - `ListMarketUniverseSnapshots`
  - `GetMarketUniverseSnapshot`
  - `GetMarketCoverageSummary`
- 参数校验统一使用 `dto/market_data.go`

- [ ] **Step 3: 将总入口与现有 scheduler run 关联**

Implementation notes:
- `POST /admin/market-data/backfill` 创建 scheduler run，状态先置 `RUNNING`
- 执行过程中持续更新：
  - `scheduler_job_runs.result_summary`
  - `market_backfill_runs.stage_progress_json`
- 失败批次重试继续使用现有 `scheduler_job_runs.parent_run_id` 与 `retry_count`

- [ ] **Step 4: 在 router 中挂载新路由并保留兼容股票接口**

Implementation notes:
- 新增 group：`/api/v1/admin/market-data`
- 保持已有：
  - `/api/v1/admin/stocks/quotes/sync`
  - `/api/v1/admin/stocks/quotes/rebuild-derived-truth`
- market-data group 中新增：
  - `/backfill`
  - `/backfill-runs`
  - `/backfill-runs/:id`
  - `/backfill-runs/:id/details`
  - `/backfill-runs/:id/retry`
  - `/universe-snapshots`
  - `/universe-snapshots/:id`
  - `/master/sync`
  - `/quotes/sync`
  - `/daily-basic/sync`
  - `/moneyflow/sync`
  - `/truth/rebuild`

- [ ] **Step 5: 运行 handler/router tests，确认 API 链通过**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler ./router -run 'Test.*MarketData.*Backfill|Test.*Scheduler.*MarketData'`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/handler/admin_growth_handler.go \
  backend/internal/growth/handler/market_data_admin_handler.go \
  backend/internal/growth/service/service.go \
  backend/router/router.go \
  backend/internal/growth/handler/market_data_admin_handler_test.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add market data backfill admin APIs"
```

## Task 5: Extend Governance Summary And Data Source Management View

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/market-data-admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`
- Test: `/Users/gjhan21/cursor/sercherai/admin/src/lib/market-data-admin.test.js`

- [ ] **Step 1: 写 failing tests，覆盖 coverage summary 字段与前端格式化缺失**

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm test -- market-data-admin.test.js`
Expected: FAIL，提示缺少 coverage summary 映射、资产分布格式化或新的 API 封装。

- [ ] **Step 2: 在 repo 中扩展治理摘要**

Implementation notes:
- 在 `market_data_admin.go` 中新增或扩展：
  - `AdminGetMarketCoverageSummary`
- 摘要至少返回：
  - universe 覆盖数
  - master truth 覆盖数
  - quotes 覆盖数
  - daily_basic 覆盖数
  - moneyflow 覆盖数
  - latest trade date
  - fallback source 分布
  - canonical key gap 数
  - display name gap 数

- [ ] **Step 3: 在 admin API 和格式化工具层补新请求与中文文案**

Implementation notes:
- `admin/src/api/admin.js` 增加：
  - `createMarketDataBackfillRun`
  - `listMarketDataBackfillRuns`
  - `getMarketDataBackfillRun`
  - `listMarketDataBackfillRunDetails`
  - `retryMarketDataBackfillRun`
  - `listMarketUniverseSnapshots`
  - `getMarketUniverseSnapshot`
  - `getMarketCoverageSummary`
- `admin/src/lib/market-data-admin.js` 增加：
  - coverage cards 构造
  - asset scope 标签格式化
  - batch detail 状态格式化

- [ ] **Step 4: 改造 DataSourcesView，加入 coverage 与缺口卡组**

Implementation notes:
- 保留现有质量日志、truth 摘要、同步面板
- 增加独立区块：
  - 覆盖率总览卡
  - 资产类型覆盖卡
  - 最新 trade date / fallback source 摘要
  - universe 快照入口

- [ ] **Step 5: 运行前后端测试**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler ./internal/growth/repo -run 'Test.*Coverage|Test.*Quality'`
Expected: PASS

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm test -- market-data-admin.test.js`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/market_data_admin.go \
  backend/internal/growth/model/market_data_admin.go \
  backend/internal/growth/handler/market_data_admin_handler.go \
  admin/src/api/admin.js \
  admin/src/lib/market-data-admin.js \
  admin/src/views/DataSourcesView.vue \
  admin/src/lib/market-data-admin.test.js
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add market data coverage governance dashboard"
```

## Task 6: Add Market Backfill Experience To System Jobs

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/system-jobs-admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/market-data-admin.js`

- [ ] **Step 1: 写 failing tests，覆盖市场数据任务卡与分区文案**

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm test -- system-jobs-admin.test.js`
Expected: FAIL，提示市场数据全量回填卡片、标签或说明文案不存在。

- [ ] **Step 2: 在 system jobs 工具层增加市场数据专区配置**

Implementation notes:
- 在 `system-jobs-admin.js` 中新增：
  - 市场数据任务概览卡
  - 中文操作说明卡
  - 快捷任务入口：全量回填、增量同步、truth 重建
- 任务说明必须采用中文产品语气，例如：
  - “先生成证券全集，再按阶段补数”
  - “失败批次可重试，不需要从头跑”

- [ ] **Step 3: 在 SystemJobsView.vue 中增加市场数据专区与详情联动**

Implementation notes:
- 不把市场数据任务和所有其他 tab 混在一起
- 增加独立 tab 或独立大卡区：
  - 发起市场回填
  - 回填任务列表
  - 批次明细抽屉
  - 最近 universe 快照
- 操作入口至少支持：
  - 新建全量任务
  - 查看任务详情
  - 重试失败批次

- [ ] **Step 4: 运行 admin build 与 tests**

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm test -- system-jobs-admin.test.js`
Expected: PASS

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  admin/src/api/admin.js \
  admin/src/lib/system-jobs-admin.js \
  admin/src/views/SystemJobsView.vue \
  admin/src/lib/market-data-admin.js
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add market backfill workspace to system jobs"
```

## Task 7: Run Full Regression And Final Verification

**Files:**
- No new product files
- Verify all files touched above

- [ ] **Step 1: 运行 backend 全量相关测试**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo ./internal/growth/handler ./router`
Expected: PASS

- [ ] **Step 2: 运行 admin 测试与构建**

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm test`
Expected: PASS

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`
Expected: PASS

- [ ] **Step 3: 做数据库与运行态人工验收**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && MYSQL_PWD=abc123 mysql -h 127.0.0.1 -P 3306 -u root sercherai -e "SHOW TABLES LIKE 'market\\_universe\\_%'; SHOW TABLES LIKE 'market\\_backfill\\_%';"`
Expected: 新表存在。

Run: `curl -s http://127.0.0.1:19081/healthz`
Expected: `{"status":"ok"}`

Manual checklist:
- 任务中心可创建市场回填任务
- DataSourcesView 可看到 coverage summary
- 失败批次可重试
- summary 能区分 `SUCCESS / PARTIAL_SUCCESS / FAILED`

- [ ] **Step 4: 最终 commit**

```bash
git -C /Users/gjhan21/cursor/sercherai status --short
git -C /Users/gjhan21/cursor/sercherai add backend admin docs/superpowers/plans/2026-03-24-market-data-full-backfill.md
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: implement market data full backfill workflow"
```

## Acceptance Checklist

- [ ] 能创建一次覆盖 `STOCK / INDEX / ETF / LOF / CBOND` 的 universe 快照
- [ ] 能创建并查询 market data full backfill 总任务
- [ ] 总任务按 `UNIVERSE -> MASTER -> QUOTES -> DAILY_BASIC -> MONEYFLOW -> TRUTH -> COVERAGE_SUMMARY` 推进
- [ ] `STOCK` 的 `daily_basic` 与 `moneyflow` 正式入链，不再是空表或孤立函数
- [ ] 其他资产类型对增强因子明确 `SKIPPED`，不被错误记为失败
- [ ] 后台支持查看任务详情、批次明细、失败原因与重试入口
- [ ] 数据源治理页能看到全市场覆盖率、latest trade date、fallback source 和缺口
- [ ] 兼容股票旧接口仍可用，但新的多资产回填入口统一走 `/admin/market-data/*`
