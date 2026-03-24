# Market Data Full Backfill Design

## 背景

当前后端已经具备一部分市场数据底座能力，但仍处于“半成品”状态：

- 已有股票行情同步入口 `POST /api/v1/admin/stocks/quotes/sync`
- 已有股票 derived truth 重建入口 `POST /api/v1/admin/stocks/quotes/rebuild-derived-truth`
- 已有多源配置、主数据 truth、质量日志与任务中心基础设施
- 已有 `fetchStockDailyBasicsFromTushare`、`fetchStockMoneyflowsFromTushare` 等增强字段抓取函数

但系统仍缺少一条正式的“全市场数据回填链路”：

- 没有真正的全量 backfill 总入口
- 空 `symbols` 时仍回落到 15 只样本股，不会自动跑全市场
- `stock_daily_basic` 与 `stock_moneyflow_daily` 还没有进入正式的全量任务化流程
- 后台任务中心没有“市场全量回填”的中文业务视角
- 现有治理摘要还不能清楚表达 universe 覆盖率、缺口和增强因子覆盖情况

本 spec 的目标是把现有“多源同步能力 + truth 能力 + scheduler 能力”收口成一套可长期运行的全市场证券数据底座。

## 设计目标

- 建立覆盖 `沪深京 A 股 + 指数 + ETF + LOF + 可转债` 的正式 universe 快照能力
- 建立全量与增量共用的一体化任务链：`universe -> master -> quotes -> daily_basic -> moneyflow -> truth`
- 复用现有 `scheduler_job_definitions` 与 `scheduler_job_runs`，不另起一套任务系统
- 让后台可以发起任务、查看阶段进度、查看失败批次、重试失败批次、查看覆盖率缺口
- 把 `daily_basic` 与 `moneyflow` 从底层函数提升为正式同步能力
- 把“当前只有 15 只样本主数据”的状态升级为“全市场量级 universe + truth 底座”

## 非目标

- 不新建 BFF 或独立数据中台
- 不新增分布式任务系统或消息队列基础设施
- 不把期货、外汇、宏观等其他市场一起并入本轮范围
- 不在本轮接入第三方支付、前端业务页面、推荐策略逻辑
- 不追求一期内所有资产类型都具备完全相同的增强因子支持深度

## 设计范围

本轮纳入范围的资产类型固定为：

- `STOCK`
- `INDEX`
- `ETF`
- `LOF`
- `CBOND`

本轮必须落地的能力固定为：

- universe 快照
- 主数据同步
- 日线行情回填
- `daily_basic` 正式同步
- `moneyflow` 正式同步
- derived truth 重建
- 任务中心发起、查询、重试
- 治理摘要与覆盖率统计

## 总体方案

本轮采用“任务编排式全量回填”方案，而不是“大 HTTP 请求直接跑完全程”。

管理员发起的是一次“市场全量回填任务”，后端实际执行拆成多个阶段：

1. `UNIVERSE`
2. `MASTER`
3. `QUOTES`
4. `DAILY_BASIC`
5. `MONEYFLOW`
6. `TRUTH`
7. `COVERAGE_SUMMARY`

设计原则固定如下：

- 总入口只负责创建任务，不直接同步全量数据
- 每个阶段都必须可重试、可续跑、可单独执行
- 全量与增量共用同一条链路，只是参数不同
- universe 必须先快照，再驱动后续阶段
- quotes、daily_basic、moneyflow 先写事实表，再集中执行 truth 重建
- scheduler 继续作为统一任务主控，不新增平行任务中心

## 信息架构

### 1. 总入口与工具入口分层

总入口负责“一次创建完整回填任务”，工具入口负责“分阶段补数和排障”。

总入口：

- `POST /api/v1/admin/market-data/backfill`

新的多资产工具入口：

- `POST /api/v1/admin/market-data/master/sync`
- `POST /api/v1/admin/market-data/quotes/sync`
- `POST /api/v1/admin/market-data/daily-basic/sync`
- `POST /api/v1/admin/market-data/moneyflow/sync`
- `POST /api/v1/admin/market-data/truth/rebuild`

现有兼容入口继续保留：

- `POST /api/v1/admin/stocks/quotes/sync`
- `POST /api/v1/admin/stocks/quotes/rebuild-derived-truth`

查询入口：

- `GET /api/v1/admin/market-data/backfill-runs`
- `GET /api/v1/admin/market-data/backfill-runs/:id`
- `GET /api/v1/admin/market-data/backfill-runs/:id/details`
- `POST /api/v1/admin/market-data/backfill-runs/:id/retry`
- `GET /api/v1/admin/market-data/universe-snapshots`
- `GET /api/v1/admin/market-data/universe-snapshots/:id`
- `GET /api/v1/admin/data-sources/market-coverage-summary`

### 2. 后台任务中心分区

任务中心不再只展示原始 job name，而要补出“市场数据全量回填”中文业务视角。

后台固定分为以下区块：

- 市场数据全量回填
- 回填任务记录
- 回填批次明细
- Universe 快照
- 覆盖率与缺口

管理员的主操作路径固定为：

1. 发起全量回填
2. 查看阶段进度
3. 查看失败批次
4. 重试失败批次或从阶段继续
5. 查看覆盖率与缺口摘要

## 数据模型设计

### 1. 继续复用的现有表

以下表继续作为现有底座使用，不另做替换：

- `scheduler_job_definitions`
- `scheduler_job_runs`
- `market_instruments`
- `market_instrument_source_facts`
- `market_symbol_aliases`
- `market_daily_bars`
- `market_daily_bar_truth`
- `stock_market_quotes`
- `stock_daily_basic`
- `stock_moneyflow_daily`
- `market_data_quality_logs`

### 2. 新增表：Universe 快照

#### `market_universe_snapshots`

字段建议：

- `id`
- `scope`
- `source_key`
- `snapshot_date`
- `summary_json`
- `created_by`
- `created_at`

职责：

- 标识一次固定的 universe 基准
- 作为整次全量任务的唯一分母
- 为复盘、重试、覆盖率统计提供稳定输入

#### `market_universe_snapshot_items`

字段建议：

- `id`
- `snapshot_id`
- `asset_type`
- `instrument_key`
- `external_symbol`
- `display_name`
- `exchange_code`
- `status`
- `list_date`
- `delist_date`
- `raw_metadata_json`
- `created_at`

职责：

- 记录一次快照中每个证券的标准化结果
- 支持后续按资产类型和 symbol 批次切分
- 支持回填复盘和差异分析

### 3. 新增表：全量回填总单

#### `market_backfill_runs`

字段建议：

- `id`
- `scheduler_run_id`
- `run_type`
- `asset_scope`
- `trade_date_from`
- `trade_date_to`
- `source_key`
- `batch_size`
- `universe_snapshot_id`
- `status`
- `current_stage`
- `stage_progress_json`
- `summary_json`
- `error_message`
- `created_by`
- `created_at`
- `updated_at`
- `finished_at`

状态值固定为：

- `PENDING`
- `RUNNING`
- `PARTIAL_SUCCESS`
- `SUCCESS`
- `FAILED`
- `CANCELLED`

职责：

- 表示一笔完整的市场数据回填任务
- 挂接 scheduler run 主记录
- 汇总每个阶段的总体进度、摘要和最终结果

### 4. 新增表：批次明细

#### `market_backfill_run_details`

字段建议：

- `id`
- `run_id`
- `scheduler_run_id`
- `stage`
- `asset_type`
- `batch_key`
- `source_key`
- `symbol_count`
- `symbol_sample`
- `trade_date_from`
- `trade_date_to`
- `status`
- `fetched_count`
- `upserted_count`
- `truth_count`
- `warning_text`
- `error_text`
- `started_at`
- `finished_at`
- `created_at`
- `updated_at`

阶段值固定为：

- `UNIVERSE`
- `MASTER`
- `QUOTES`
- `DAILY_BASIC`
- `MONEYFLOW`
- `TRUTH`
- `COVERAGE_SUMMARY`

批次状态固定为：

- `PENDING`
- `RUNNING`
- `SUCCESS`
- `FAILED`
- `SKIPPED`

职责：

- 记录每个阶段每个资产类型每个批次的执行结果
- 为任务中心展示明细、失败原因和重试入口提供事实来源

## 任务编排设计

### 1. 阶段顺序

全量任务执行顺序固定为：

1. `UNIVERSE`
2. `MASTER`
3. `QUOTES`
4. `DAILY_BASIC`
5. `MONEYFLOW`
6. `TRUTH`
7. `COVERAGE_SUMMARY`

任何阶段失败都必须写入总单和批次明细，不能只靠日志表达。

### 2. Universe 阶段

职责：

- 按资产范围拉全集
- 标准化为内部 `instrument_key`
- 写入快照表
- 统计每类资产数量与状态分布

要求：

- 同一任务一旦生成快照，后续所有阶段必须绑定该 `snapshot_id`
- 重试批次时不重新改写本次任务的 universe

### 3. Master 阶段

职责：

- 同步主数据
- 补齐 alias、source facts 与 instrument truth

要求：

- 先按资产类型分段执行，再按需要细分批次
- 必须先于 quotes 阶段完成
- 必须支持重复 upsert，保持幂等

### 4. Quotes 阶段

职责：

- 回填日线行情
- 写入事实表
- 记录批次级抓取与入库结果

批次策略建议：

- `STOCK`：每批 `200-300`
- `INDEX`：每批 `100-200`
- `ETF`：每批 `100-200`
- `LOF`：每批 `100-200`
- `CBOND`：每批 `100-200`

要求：

- 批次必须写明资产类型、日期范围、symbol 数量和 sample
- 支持失败批次重试
- 不在每一批结束后立即全表 truth rebuild

### 5. Daily Basic 阶段

职责：

- 把当前已存在的 `daily_basic` 抓取能力接入正式回填任务

支持矩阵要求：

- `STOCK`：正式支持
- `INDEX`：默认 `SKIPPED`
- `ETF / LOF / CBOND`：一期默认 `SKIPPED`

要求：

- 不支持的资产类型必须明确 `SKIPPED`
- 不能把“不支持”误记为“失败”
- 如后续实施中确认某一资产类型已有稳定真实源支持，可作为增量增强单独开启，但不作为本期验收前提

### 6. Moneyflow 阶段

职责：

- 把当前已存在的 `moneyflow` 抓取能力接入正式回填任务

一期支持策略固定为：

- `STOCK`：正式支持
- `INDEX / ETF / LOF / CBOND`：默认 `SKIPPED`

要求：

- 一期优先保证股票 moneyflow 正式可跑
- 其他资产类型如暂不稳定，不强行扩张

### 7. Truth 阶段

职责：

- 对本次任务受影响范围执行集中 truth 重建

重建粒度要求：

- 支持按 `snapshot_id`
- 支持按 `asset_type`
- 支持按 `trade_date range`

执行建议：

- quotes、daily_basic、moneyflow 先完成事实写入
- 再按资产类型分段 rebuild
- 最后输出总覆盖率摘要

## 失败重试与幂等设计

### 1. 重试层级

必须支持以下三层重试：

- 批次级重试
- 阶段级重试
- 从指定阶段继续

默认不建议整单从头重复执行，除非管理员明确选择。

### 2. 幂等要求

- `master sync` 必须支持重复 upsert
- `quotes / daily_basic / moneyflow` 必须以稳定主键或唯一约束 upsert
- `truth rebuild` 必须支持重复执行并稳定覆盖
- universe 快照一旦绑定任务，不因重试而变化

### 3. 总任务成功判定

成功判定规则固定为：

- 所有阶段全部成功：`SUCCESS`
- 核心阶段成功但部分批次失败：`PARTIAL_SUCCESS`
- 核心阶段失败导致结果不可用：`FAILED`

核心阶段固定为：

- `UNIVERSE`
- `MASTER`
- `QUOTES`
- `TRUTH`

`DAILY_BASIC` 与 `MONEYFLOW` 在一期属于增强层：

- 应尽量完成
- 但少量失败不必强制打整单 `FAILED`
- 必须在摘要与治理页中清楚暴露缺口

## 接口设计

### 1. 总入口

#### `POST /api/v1/admin/market-data/backfill`

请求体建议：

- `run_type`
- `asset_scope`
- `source_key`
- `trade_date_from`
- `trade_date_to`
- `batch_size`
- `stages`
- `force_refresh_universe`
- `rebuild_truth_after_sync`

返回值建议：

- `run_id`
- `scheduler_run_id`
- `status`

### 2. 阶段工具接口

新增市场数据工具接口：

- `POST /api/v1/admin/market-data/master/sync`
- `POST /api/v1/admin/market-data/quotes/sync`
- `POST /api/v1/admin/market-data/daily-basic/sync`
- `POST /api/v1/admin/market-data/moneyflow/sync`
- `POST /api/v1/admin/market-data/truth/rebuild`

现有股票兼容接口保留：

- `POST /api/v1/admin/stocks/quotes/sync`
- `POST /api/v1/admin/stocks/quotes/rebuild-derived-truth`

要求：

- 支持按 `snapshot_id` 或 `asset_scope` 执行
- 支持小范围补数与排障，不局限于总任务内部调用

### 3. 查询接口

新增：

- `GET /api/v1/admin/market-data/backfill-runs`
- `GET /api/v1/admin/market-data/backfill-runs/:id`
- `GET /api/v1/admin/market-data/backfill-runs/:id/details`
- `POST /api/v1/admin/market-data/backfill-runs/:id/retry`
- `GET /api/v1/admin/market-data/universe-snapshots`
- `GET /api/v1/admin/market-data/universe-snapshots/:id`
- `GET /api/v1/admin/data-sources/market-coverage-summary`

## 页面与运维承接

### 1. 任务中心承接方式

任务中心需要提供更贴近中文后台习惯的业务视角，而不是只展示抽象 job 列表。

页面结构建议：

- 页面定位与说明
- 发起回填卡
- 运行中的任务
- 历史任务列表
- 任务详情抽屉或详情页
- 失败批次与重试区
- Universe 快照与覆盖率摘要

使用说明必须采用中文产品语气，避免技术术语堆叠。

### 2. 数据源治理页承接方式

治理摘要页需要扩展以下指标：

- 各资产类型 universe 覆盖数
- 各资产类型主数据 truth 覆盖数
- quotes 覆盖数
- `daily_basic` 覆盖数
- `moneyflow` 覆盖数
- latest trade date
- fallback source 分布
- canonical key 缺口数
- display name 缺失数
- list date 缺失数

## 实施顺序建议

### 阶段 1：Universe 与主数据

- 新增快照表
- 建立全市场 universe 拉取与标准化
- 打通主数据 source facts、aliases、truth

### 阶段 2：Quotes 主链

- 新增 backfill 总入口
- 跑通 quotes 批次同步
- 接入批次明细与失败重试

### 阶段 3：Daily Basic 与 Moneyflow

- 把现有抓取函数提升为正式同步任务
- 建立支持矩阵与 `SKIPPED` 规则

### 阶段 4：Truth 与治理摘要

- 支持按范围 rebuild truth
- 扩展治理摘要和缺口统计

### 阶段 5：后台任务中心收口

- 发起、查看、重试、覆盖率查看统一收口

## 验收标准

### 数据范围验收

- 能生成一次包含 `STOCK / INDEX / ETF / LOF / CBOND` 的 universe 快照
- universe 快照中每类资产都有数量统计与状态分布
- 主数据 truth 不再停留在样本级别

### 同步链路验收

- 管理端可创建一次全量回填任务
- 总任务可依次推进 `UNIVERSE -> MASTER -> QUOTES -> DAILY_BASIC -> MONEYFLOW -> TRUTH`
- 每个阶段均有批次明细
- 失败批次可重试
- 可从阶段继续执行

### 数据结果验收

- `market_daily_bars` 与 `market_daily_bar_truth` 的覆盖从样本级显著扩展到全市场量级
- `stock_daily_basic` 不再为 0
- `stock_moneyflow_daily` 不再为 0
- `market_symbol_aliases` 与 `market_instrument_source_facts` 不再停留在 15 条样本级别

### 治理与可观测性验收

- 后台可查看总任务状态、当前阶段、失败原因、覆盖率摘要与 latest trade date
- 后台可查看 fallback source 使用情况与缺口数量
- 任务中心中文说明清晰，非开发人员可理解基本操作路径

### 稳定性验收

- 同一批次重复执行不会产生脏重复数据
- 失败后可只重试失败批次
- 中断后可从指定阶段继续
- truth rebuild 可重复执行且结果稳定

## 风险与控制点

### 1. 数据源支持深度不一致

不同资产类型对 `daily_basic`、`moneyflow` 的支持深度不一致，一期必须用支持矩阵表达，而不是强行统一到同一深度。

### 2. 全量任务耗时长

全量任务必须以后台任务形式执行，不能依赖同步 HTTP 请求完成。

### 3. universe 变化带来的结果漂移

必须以 `snapshot_id` 固定一次任务的 universe 分母，防止长任务运行中 universe 漂移。

### 4. 任务中心可用性

如果只有通用 job name 而没有业务明细，管理员仍然很难用。因此任务明细表和中文业务视图是必要设计，不属于额外装饰。

## 最终结论

本轮不是单纯补一个“全量下载接口”，而是要在现有多源市场数据与任务中心基础上，新增一套面向 `A 股 + 指数 + ETF + LOF + 可转债` 的正式市场数据回填体系。

这套体系以 universe 快照为分母，以 scheduler 任务为主控，以批次明细和 truth 重建为核心，同时把 `daily_basic` 与 `moneyflow` 从底层函数提升为正式可运维能力，最终让后台具备“可发起、可查看、可重试、可治理”的全市场证券数据底座。
