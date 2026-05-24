# 系统任务中心同步模型升级设计

## Summary

当前 admin 端存在两套并行的市场同步入口：

- [data-sources/sync](/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue) 中的股票/期货同步操作台，直接调用同步接口
- [system-jobs](/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue) 中的“市场数据回填工作台”，走 `marketBackfillRun` 总单/明细/重试链路

这会造成三个问题：

1. 操作员需要理解两套任务模型：`同步操作` 和 `回填总单`
2. 股票/期货同步缺少统一任务中心，追踪、重试、审计分散在两个页面
3. `system-jobs` 里的旧“市场数据回填”语义，与一线运营更常用的“全量同步 / 每日增量同步”心智不一致

本次升级不新建第二套任务表，也不重做底层执行引擎，而是：

- 保留现有 `marketBackfillRun` 执行链、详情链、审计链
- 将其前台模型整体升级为“市场同步任务”模型
- 让 `system-jobs` 成为股票/期货同步的唯一主入口
- 让 `data-sources/sync` 退化为治理与引导页面

第一阶段仅覆盖：

- 股票全量同步
- 股票每日增量同步
- 期货全量同步
- 期货每日增量同步

资讯同步本期不纳入。

## 背景

### 当前 `data-sources/sync`

股票/期货同步操作目前集中在：

- [admin/src/composables/useMarketSyncConsole.js](/Users/gjhan21/cursor/sercherai/admin/src/composables/useMarketSyncConsole.js)

其中股票侧有三类动作：

- `全量同步`
- `每日增量同步`
- `仅行情`

期货侧有两类动作：

- `全量同步`
- `仅行情`

这些动作直接调用：

- [syncStockInstrumentMaster](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [syncStockQuotes](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [incrementalSyncStockQuotes](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [syncFuturesQuotes](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [syncFuturesInventory](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)

也就是说，这里是“直接触发同步接口”的操作台。

### 当前 `system-jobs`

`system-jobs` 已经存在一套成熟任务模型：

- 创建总单
- 列表查询
- 总单详情
- 阶段明细
- 重试失败批次

核心前端页面在：

- [admin/src/views/SystemJobsView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue)

核心接口在：

- [createMarketDataBackfillRun](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [listMarketDataBackfillRuns](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [getMarketDataBackfillRun](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [listMarketDataBackfillRunDetails](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [retryMarketDataBackfillRun](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)

后端 handler 在：

- [backend/internal/growth/handler/market_data_admin_handler.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go)

底层执行链在：

- [backend/internal/growth/repo/market_backfill_execution.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_backfill_execution.go)

说明现有 `system-jobs` 其实已经是“同步任务中心”的半成品，只是对外仍暴露旧“回填”语义。

## 问题定义

本次需要解决的不是“再加两个按钮”，而是统一模型。

当前问题主要有四类：

1. **入口双轨**
   股票/期货同步既能在 `data-sources/sync` 发起，也能在 `system-jobs` 发起回填，总线不统一。

2. **任务语义错位**
   运营心智是：
   - 股票全量同步
   - 股票每日增量同步
   - 期货全量同步
   - 期货每日增量同步

   但 `system-jobs` 暴露的是：
   - `FULL`
   - `INCREMENTAL`
   - `REBUILD_ONLY`
   - `asset_scope`
   - `stages`

3. **期货未正式纳入统一任务模板**
   当前 DTO [MarketDataBackfillRequest](/Users/gjhan21/cursor/sercherai/backend/internal/growth/dto/market_data.go) 的 `asset_scope` 只允许：
   - `STOCK`
   - `INDEX`
   - `ETF`
   - `LOF`
   - `CBOND`

   前台虽然已有期货直接同步按钮，但未进入 `marketBackfillRun` 统一任务语义。

4. **旧模型词汇暴露过深**
   “回填总单 / 回填批次 / 发起回填”这些词会让一线使用者误以为这里只是补数后台，而不是股票/期货同步任务中心。

## 目标

### 产品目标

1. 让 `system-jobs` 成为股票/期货同步的唯一主入口
2. 让一线用户只看到“同步任务”语义，不需要理解“回填模型”
3. 用统一任务模板表达四类任务：
   - 股票全量同步
   - 股票每日增量同步
   - 期货全量同步
   - 期货每日增量同步
4. 保留现有总单、明细、重试、审计能力

### 技术目标

1. 复用现有 `marketBackfillRun` 执行链，不新建 `syncRun` 体系
2. 补齐 `FUTURES` 进入统一任务模型的后端校验与执行链
3. 将旧模型命名留在内部，前台全面切换为新语义

## 非目标

本次不做以下内容：

- 不重命名数据库表或历史 Go 结构体
- 不新增第二套同步任务表
- 不重做资讯同步模型
- 不重写市场数据底层执行引擎
- 不新增更复杂的任务编排器
- 不在本期清理所有内部 `backfill` 命名

## 最终模型

### 1. 前台任务模板

前台统一暴露四个任务模板：

- `STOCK_FULL`
- `STOCK_INCREMENTAL`
- `FUTURES_FULL`
- `FUTURES_INCREMENTAL`

注意：

这些模板是 **前台语义模板**，第一阶段不要求后端数据库层也立刻改成相同枚举。

### 2. 后台执行映射

#### 股票全量同步

映射为：

- `run_type = FULL`
- `asset_scope = ["STOCK", "INDEX", "ETF", "LOF", "CBOND"]`
- `stages = ["UNIVERSE", "MASTER", "QUOTES", "DAILY_BASIC", "MONEYFLOW", "TRUTH", "COVERAGE_SUMMARY"]`
- `force_refresh_universe = true`
- `rebuild_truth_after_sync = true`

#### 股票每日增量同步

映射为：

- `run_type = INCREMENTAL`
- `asset_scope = ["STOCK", "INDEX", "ETF", "LOF", "CBOND"]`
- `stages = ["QUOTES", "DAILY_BASIC", "MONEYFLOW", "TRUTH", "COVERAGE_SUMMARY"]`
- `force_refresh_universe = false`
- `rebuild_truth_after_sync = true`

#### 期货全量同步

映射为：

- `run_type = FULL`
- `asset_scope = ["FUTURES"]`
- `stages = ["UNIVERSE", "MASTER", "QUOTES", "TRUTH", "COVERAGE_SUMMARY"]`
- `force_refresh_universe = true`
- `rebuild_truth_after_sync = true`

如果底层第一阶段可稳定支持期货库存阶段，可追加：
- `INVENTORY`

#### 期货每日增量同步

映射为：

- `run_type = INCREMENTAL`
- `asset_scope = ["FUTURES"]`
- `stages = ["QUOTES", "TRUTH", "COVERAGE_SUMMARY"]`
- `force_refresh_universe = false`
- `rebuild_truth_after_sync = true`

第一阶段若期货 `INVENTORY` 在增量链上不稳定，不强行纳入默认模板。

## 页面设计

### A. `system-jobs`

#### 模块级重命名

统一替换前台文案：

- `市场数据回填工作台` -> `市场同步任务中心`
- `新建回填任务` -> `新建同步任务`
- `回填总单` -> `同步任务列表`
- `回填批次明细` -> `同步批次明细`
- `发起回填` -> `发起同步`

#### 顶部概览卡

从“回填视角”改成“同步任务视角”，例如：

- 今日同步任务数
- 最近成功同步
- 最近失败同步
- 最近股票全量
- 最近股票增量
- 最近期货同步

不再强调“回填总单”“Universe 快照”这样的底层术语。

#### 新建任务表单

新建表单改为 **任务模板驱动**：

1. 先选任务模板：
   - 股票全量同步
   - 股票每日增量同步
   - 期货全量同步
   - 期货每日增量同步

2. 再显示模板可调参数：
   - 来源
   - 起止日期
   - 批量大小
   - 是否重建 Truth
   - 是否强制刷新 Universe

用户不再直接操作“原始回填参数模型”。

#### 同步任务列表

列表保留现有技术字段，但展示名统一做前台映射：

- `FULL + 股票资产范围` -> `股票全量同步`
- `INCREMENTAL + 股票资产范围` -> `股票每日增量同步`
- `FULL + FUTURES` -> `期货全量同步`
- `INCREMENTAL + FUTURES` -> `期货每日增量同步`

列表列建议保留：

- 任务ID
- 任务类型
- 资产范围
- 来源
- 当前阶段
- 状态
- 调度运行ID
- 创建时间
- 操作

#### 同步任务详情

详情弹窗保留现有：

- 总单状态
- 阶段进度
- 分阶段明细
- 重试失败批次

只做前台语义重命名。

### B. `data-sources/sync`

该页不再作为股票/期货同步主入口。

第一阶段保留：

- 数据源治理
- 健康状态
- 质量与 truth rebuild

移除或替换：

- 股票全量同步
- 股票每日增量同步
- 期货全量同步

新的表现方式是：

- 提示卡：`股票/期货同步任务已迁移到任务中心`
- 跳转按钮：`前往任务中心`

这样可以避免双入口继续并行。

## 后端设计

### 1. DTO 扩展

文件：
- [backend/internal/growth/dto/market_data.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/dto/market_data.go)

`MarketDataBackfillRequest.AssetScope` 需要支持：

- `FUTURES`

否则期货无法纳入统一任务模型。

### 2. handler 保持接口稳定

文件：
- [backend/internal/growth/handler/market_data_admin_handler.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go)

第一阶段不新增新接口，继续复用：

- `CreateMarketDataBackfillRun`
- `ListMarketDataBackfillRuns`
- `GetMarketDataBackfillRun`
- `ListMarketDataBackfillRunDetails`
- `RetryMarketDataBackfillRun`

外部协议保持稳定，只更新前端使用方式与必要错误提示。

### 3. 执行链补齐 `FUTURES`

文件：
- [backend/internal/growth/repo/market_backfill_execution.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_backfill_execution.go)

当前执行链已存在若干 `FUTURES` 分支：

- `normalizeMarketStageInstrumentKeys`
- `marketStageSyncPriority`
- `resolveMarketBackfillQuoteRoute`

第一阶段需要做的是：

1. 放开 `FUTURES` 进入 run 创建
2. 确认 `FULL + FUTURES` 执行链可达
3. 为 `INCREMENTAL + FUTURES` 增加阶段裁剪或容错
4. 保证前端模板发起的默认阶段组合不会直接触发非法执行路径

### 4. 旧模型保留为内部实现

第一阶段允许继续保留：

- `MarketBackfillRun`
- `MarketBackfillRunDetail`
- `market_backfill_runs`
- `market_backfill_run_details`

前台完成概念切换即可，底层结构不做破坏性重命名。

## 测试要求

### 前端

应补充以下测试：

1. `SystemJobsView` 任务模板测试
   - 存在四个同步任务模板
   - 模板能正确映射为后端 payload

2. `system-jobs-admin.js` helper 测试
   - `formatSyncJobTypeLabel`
   - `buildSyncJobPayloadFromTemplate`
   - `deriveSyncJobTemplateFromBackfillRun`

3. `DataSources` 收口测试
   - 股票/期货同步操作不再作为该页主入口
   - 页面引导跳转 `/system-jobs?tab=market-data`

### 后端

应补充以下测试：

1. handler 测试
   - `CreateMarketDataBackfillRun` 支持 `asset_scope=["FUTURES"]`

2. repo / 执行链测试
   - `FULL + FUTURES`
   - `INCREMENTAL + FUTURES`
   - 期货增量默认阶段不会错误要求股票增强阶段

## 分阶段实施建议

### 第一阶段：模型升级与入口统一

目标：

- `system-jobs` 完成同步任务语义切换
- `data-sources/sync` 退出股票/期货同步主入口
- `FUTURES` 纳入统一任务模板

### 第二阶段：任务细化与更多期货增强

后续可考虑：

- 期货库存与结构同步进一步模板化
- 同步任务摘要卡增加真实成功率和耗时统计
- 再考虑资讯同步是否迁入同模型

## 结论

最佳路线不是再造一套 `syncRun`，而是：

**保留旧 `marketBackfillRun` 引擎与明细链，整体升级为新的股票/期货同步任务模型。**

这样既能统一入口和用户心智，也能最大化复用已有总单、分阶段明细、重试和审计能力。
