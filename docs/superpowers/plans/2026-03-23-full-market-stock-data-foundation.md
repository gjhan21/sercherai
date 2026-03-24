# 全市场股票数据底座 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把当前 15 只样本级股票行情链路升级为“全市场股票主数据 + 全市场日线真相 + 增强因子 + 调度治理”底座，并让智能选股 AUTO 模式真正运行在全市场 universe 上。

**Architecture:** 继续复用现有 `market_instruments`、`market_symbol_aliases`、`market_daily_bars`、`market_daily_bar_truth`、`stock_status_truth`、`stock_daily_basic`、`stock_moneyflow_daily`、`stock_news_raw` 和 `scheduler_job_definitions`，不新建第二套数据平台。核心做法是补齐全市场主数据同步、批量日线同步与增强因子同步入口，再把调度、质量摘要和 Admin 治理视图收口到现有多源治理骨架上。

**Tech Stack:** Go, MySQL, Gin, Vue 3 Admin, Tushare, AKShare, TickerMD, MYSELF, existing market truth rebuild pipeline

---

## Scope And Assumptions

- 第一轮只覆盖 A 股在市股票，不把 ETF、可转债、B 股、退市整理证券并入全市场选股 universe。
- 内部 canonical 股票键统一采用 `ts_code` 风格，例如 `600519.SH`；历史 plain key 继续兼容读取，但不再作为新增 truth 主键。
- `TUSHARE` 作为股票主数据、增强因子的默认主源；`AKSHARE`、`MYSELF`、`TICKERMD` 保持备援和补充角色。
- 本计划建立在现有专题文档 [专题A-多源市场数据统一模型与供应商治理](/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题A-多源市场数据统一模型与供应商治理.md) 上，目标是把其中“股票全市场数据底座”部分先落成一个可执行子计划。

## Current Baseline

- 股票行情同步默认空参时只会同步 15 只样本股票，入口在 [market_data_multi_source.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source.go)。
- 本地 `market_instruments` 当前仅有 15 只股票；`market_daily_bar_truth` 归一化后也只覆盖 15 只。
- `stock_daily_basic`、`stock_moneyflow_daily`、`stock_news_raw` 当前为空表。
- `fetchStockDailyBasicsFromTushare`、`fetchStockMoneyflowsFromTushare`、`fetchStockNewsFromTushare` 已存在，但只有 news 同步已经接到运行链。
- 调度当前只有 `daily_stock_quant_pipeline` 和 `daily_stock_recommendation`，没有全市场主数据、增强因子和补数任务。

## File Map

### Backend Repo

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data.go`
  - 承接股票主数据全量/增量同步、canonical key、alias 统一更新。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source.go`
  - 承接全市场股票批量日线同步、truth 重建、legacy quote 同步和增强因子同步入口。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_admin.go`
  - 输出全市场覆盖率、新鲜度、缺口和备援摘要。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go`
  - 让 AUTO universe 严格消费 canonical 股票键和全市场 truth 摘要。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
  - 复用现有 `daily_basic` / `moneyflow` / `news` 拉取与 upsert 逻辑，补齐批量入口和查询兼容。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
  - 暴露新增的主数据、增强因子和治理接口。

### Backend Handler / Service / DTO

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
  - 新增股票主数据同步、增强因子同步、批量回填和治理摘要接口。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go`
  - 若已有派生 truth 管理语义，补充全市场批量能力。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
  - 暴露 repo 层新增能力。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/dto/market_data.go`
  - 为主数据同步、增强因子同步和覆盖率摘要补请求/响应结构。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/market_data.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/market_data_admin.go`

### Router / Migration

- Modify: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
  - 补股票主数据与增强因子管理路由。
- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260323_01_stock_market_foundation_jobs.sql`
  - 种子化新的 scheduler job definitions 与必要 system config。

### Admin

- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/market-data-admin.js`
  - 封装新增主数据同步、增强因子同步、覆盖率摘要接口。
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`
  - 补股票全市场覆盖率、增强因子覆盖率、canonical key 缺口和快捷操作。
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
  - 仅保留跳转和摘要，不新增新的重叠操作中心。

### Tests

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_candidates_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_status_truth_admin_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/market-data-admin.test.js`

## Package Breakdown

### Package 1: 股票主数据全市场化与代码键统一

**Outcome:** `market_instruments` 从样本池扩成全市场股票主数据表，`display_name`、`list_date`、行业和公司画像可用；新写入全部走 canonical `ts_code` 键。

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source_test.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260323_01_stock_market_foundation_jobs.sql`

- [ ] **Step 1: 写 failing tests，覆盖股票主数据 canonical key 与全量同步入口**

Run: `go test ./backend/internal/growth/repo -run 'TestMarketInstrument|TestMarketDailyBar'`
Expected: 现有样本逻辑不支持全市场主数据批量同步或 plain key / dotted key 去重断言失败。

- [ ] **Step 2: 扩展主数据同步逻辑，支持“先拉股票清单，再补公司画像”**

Implementation notes:
- 在 `market_instrument_master_data.go` 中新增“全量股票清单”拉取入口，优先用 `stock_basic` 批量拉清单。
- 对清单中的批量结果先写 `market_instrument_source_facts`，再调用现有 truth rebuild。
- 对少数字段缺口再使用 `stock_company` 做二次 enrichment。

- [ ] **Step 3: 统一股票 canonical key 为 `ts_code`**

Implementation notes:
- 所有新增股票主数据与日线 truth 写入都统一落 `600519.SH` 风格。
- plain key 只保留 alias，不再作为 `market_instruments.instrument_key` 的新值。
- 复用现有 `market_symbol_aliases` 表，把 plain key、AKShare key、MYSELF key 映射到 canonical key。

- [ ] **Step 4: 为全量主数据同步补最小管理入口和种子任务**

Implementation notes:
- 在 migration 中新增 scheduler job definitions：
  - `stock_master_full_sync`
  - `stock_master_incremental_sync`
- 在 system config 中补股票主数据默认批次大小、回看窗口和缺口补跑开关。

- [ ] **Step 5: 回归测试**

Run: `go test ./internal/growth/repo/... ./router/...`
Expected: 主数据与 alias 测试通过，无新增 schema 兼容错误。

- [ ] **Step 6: Commit**

```bash
git add backend/internal/growth/repo/market_instrument_master_data.go \
  backend/internal/growth/repo/market_data_multi_source.go \
  backend/internal/growth/repo/market_instrument_master_data_test.go \
  backend/internal/growth/repo/market_data_multi_source_test.go \
  backend/migrations/20260323_01_stock_market_foundation_jobs.sql
git commit -m "feat: expand stock instrument master to full market"
```

### Package 2: 全市场日线真相与 AUTO universe 扩池

**Outcome:** 股票日线同步不再默认停留在 15 只样本；`market_daily_bar_truth` 以 canonical key 覆盖全市场；AUTO universe 真正按全市场 truth 候选工作。

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_status_truth_admin_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_candidates_test.go`

- [ ] **Step 1: 写 failing tests，覆盖全市场批量日线与 AUTO universe 去重**

Run: `go test ./backend/internal/growth/repo -run 'TestStrategyEngineStockContext|TestMarketStatusTruth'`
Expected: 样本池限制、双键重复或覆盖率断言失败。

- [ ] **Step 2: 给股票行情同步增加“按股票池批量拉取”的显式模式**

Implementation notes:
- 不再把“symbols 为空”默认等同于样本池。
- 新增显式全市场模式：从 `market_instruments` 取 ACTIVE 股票分批同步。
- 样本池逻辑保留为 `debug/mock` 模式，不再充当生产默认。

- [ ] **Step 3: 调整 truth 重建与 legacy quote 补写逻辑**

Implementation notes:
- `rebuildMarketDailyBarTruth` 输出 canonical key。
- `syncLegacyStockQuotesFromTruthBars` 继续兼容旧 `stock_market_quotes` 读路径，但统一 symbol 规范。
- `stock_status_truth` 继续沿用当前 truth 重建链，只需要确保键一致。

- [ ] **Step 4: 更新 AUTO universe 读取逻辑**

Implementation notes:
- `loadStrategyStockContextCandidates` 只从 canonical key truth 中去重后取候选。
- 对 plain key 历史记录只作兼容读取，不再让它们放大 candidate 数量。

- [ ] **Step 5: 回归测试**

Run: `go test ./internal/growth/repo/... ./router/...`
Expected: AUTO universe 不再重复，同一股票不再以 plain key 和 dotted key 同时出现。

- [ ] **Step 6: Commit**

```bash
git add backend/internal/growth/repo/market_data_multi_source.go \
  backend/internal/growth/repo/strategy_engine_context_repo.go \
  backend/internal/growth/repo/market_status_truth_admin_test.go \
  backend/internal/growth/repo/strategy_engine_context_test.go \
  backend/internal/growth/repo/strategy_engine_context_candidates_test.go
git commit -m "feat: backfill full-market stock truth and canonical universe"
```

### Package 3: 增强因子同步入口、调度与批量补数

**Outcome:** `stock_daily_basic`、`stock_moneyflow_daily`、`stock_news_raw` 不再是空表；增强因子同步有后台入口、调度任务和缺口补数能力。

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/dto/market_data.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260323_01_stock_market_foundation_jobs.sql`

- [ ] **Step 1: 写 failing tests，覆盖 daily_basic / moneyflow / news 同步接口**

Run: `go test ./backend/internal/growth/handler ./backend/internal/growth/repo -run 'Test.*MarketData|Test.*DataSource'`
Expected: 新接口不存在或增强数据同步结果为空。

- [ ] **Step 2: 把现有 `fetchStockDailyBasicsFromTushare` / `fetchStockMoneyflowsFromTushare` 接到 repo public API**

Implementation notes:
- 复用 `upsertStockDailyBasics`、`upsertStockMoneyflows`。
- 设计与 `AdminSyncStockQuotesDetailed` 一致的结果结构：支持 source、symbols、days、batch summary。

- [ ] **Step 3: 新增管理端接口**

Recommended routes:
- `POST /api/v1/admin/stocks/master/sync`
- `POST /api/v1/admin/stocks/daily-basic/sync`
- `POST /api/v1/admin/stocks/moneyflow/sync`
- `POST /api/v1/admin/stocks/news/sync`
- `POST /api/v1/admin/stocks/backfill`

- [ ] **Step 4: 新增调度任务**

Recommended jobs:
- `stock_quotes_incremental_sync`
- `stock_daily_basic_incremental_sync`
- `stock_moneyflow_incremental_sync`
- `stock_news_incremental_sync`
- `stock_truth_rebuild`
- `stock_data_backfill`

- [ ] **Step 5: 回归测试**

Run: `go test ./internal/growth/... ./router/...`
Expected: 新接口、调度入口和 repo 同步测试全部通过。

- [ ] **Step 6: Commit**

```bash
git add backend/internal/growth/repo/mysql_repo.go \
  backend/internal/growth/repo/interfaces.go \
  backend/internal/growth/service/service.go \
  backend/internal/growth/dto/market_data.go \
  backend/internal/growth/handler/admin_growth_handler.go \
  backend/router/router.go \
  backend/internal/growth/handler/market_data_admin_handler_test.go \
  backend/migrations/20260323_01_stock_market_foundation_jobs.sql
git commit -m "feat: add stock factor sync and backfill jobs"
```

### Package 4: 数据治理摘要与 Admin 收口

**Outcome:** `数据源管理` 页面能看见股票主数据、日线 truth、增强因子覆盖率、新鲜度、缺口和备援摘要；`MarketCenter` 不再承接重复数据操作。

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/market-data-admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/market-data-admin.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`

- [ ] **Step 1: 写 failing tests，覆盖治理摘要的股票全市场覆盖率**

Run: `go test ./backend/internal/growth/handler ./backend/internal/growth/repo -run 'Test.*DerivedTruth|Test.*Quality'`
Expected: 现有摘要不含主数据、增强因子覆盖率或 canonical key 缺口。

- [ ] **Step 2: 扩展治理摘要返回字段**

Recommended summary fields:
- `stock_master_coverage`
- `stock_truth_coverage`
- `stock_daily_basic_coverage`
- `stock_moneyflow_coverage`
- `stock_news_coverage`
- `latest_trade_date`
- `fallback_source_summary`
- `canonical_key_gap_count`
- `display_name_missing_count`
- `list_date_missing_count`

- [ ] **Step 3: 收口 Admin 页面**

Implementation notes:
- `DataSourcesView.vue` 成为股票数据底座治理主入口。
- `MarketCenterView.vue` 只保留跳转、最新同步摘要和研究模块入口，不再承载重复操作表单。

- [ ] **Step 4: 前后端回归**

Run: `go test ./internal/growth/... ./router/...`
Expected: 治理摘要接口通过。

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`
Expected: Admin 构建通过。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/growth/repo/market_data_admin.go \
  backend/internal/growth/model/market_data_admin.go \
  backend/internal/growth/handler/market_data_admin_handler.go \
  backend/internal/growth/handler/market_data_admin_handler_test.go \
  admin/src/lib/market-data-admin.js \
  admin/src/lib/market-data-admin.test.js \
  admin/src/views/DataSourcesView.vue \
  admin/src/views/MarketCenterView.vue
git commit -m "feat: add stock data governance coverage dashboard"
```

## Acceptance Checklist

- [ ] `market_instruments` 股票数从 15 扩到全市场量级，`display_name` 不再主要等于代码。
- [ ] `market_daily_bar_truth` 最新交易日股票覆盖达到全市场量级。
- [ ] 新写入股票 truth 不再新增 plain key 主记录。
- [ ] `stock_daily_basic`、`stock_moneyflow_daily`、`stock_news_raw` 均不再为空表。
- [ ] `BuildStrategyEngineStockSelectionContext` 的 AUTO universe 能从全市场 truth 中去重选候选。
- [ ] `DataSourcesView` 能显示股票主数据、truth、增强因子和缺口治理摘要。
- [ ] 旧 `stock_market_quotes`、旧量化入口和旧客户端消费契约继续兼容。

## Verification Commands

### Backend

```bash
cd /Users/gjhan21/cursor/sercherai/backend
go test ./internal/growth/... ./router/...
```

### Admin

```bash
cd /Users/gjhan21/cursor/sercherai/admin
npm run build
```

### Database Spot Checks

```bash
mysql -uroot -pabc123 -h127.0.0.1 -P3306 sercherai -e "
SELECT COUNT(*) AS stock_instruments FROM market_instruments WHERE asset_class='STOCK';
SELECT COUNT(DISTINCT instrument_key) AS stock_truth_symbols, MAX(trade_date) AS max_trade_date FROM market_daily_bar_truth WHERE asset_class='STOCK';
SELECT COUNT(*) AS daily_basic_rows FROM stock_daily_basic;
SELECT COUNT(*) AS moneyflow_rows FROM stock_moneyflow_daily;
SELECT COUNT(*) AS news_rows FROM stock_news_raw;
"
```

## Risks And Guardrails

- `Tushare` 积分和频率限制可能让“全市场一次性回填”失败，必须采用批量分页与断点重跑。
- `MYSELF` 目前未加入 `market.stock.daily.source_priority`，若要作为正式日线备援，需要同步更新 routing config 与测试。
- plain key 历史数据不能粗暴删表清空，必须走 alias 兼容和 truth 去重迁移。
- `stock_news_raw` 可能受资讯接口和 symbol 映射影响，第一轮先保证“有数据可消费”，不把语义抽取并入本子计划。
- 现有 `MarketCenter` 和旧调度保留兼容，不在本计划里删除。

## Recommended Execution Order

1. Package 1
2. Package 2
3. Package 3
4. Package 4

不要并行反转顺序。`Package 2` 依赖 `Package 1` 的 canonical key；`Package 3` 依赖 `Package 1` 的全市场股票清单；`Package 4` 依赖前三包的真实数据覆盖结果。
