# 专题A：多源市场数据统一模型与供应商治理

最后更新: 2026-03-23  
状态: Planned

> 基线说明：本专题建立在阶段8、阶段9已完成的基础上。当前系统已经具备股票、期货、资讯、主数据的多源同步、truth 重建、质量日志和 Admin 数据源页，但整体仍偏“链路可跑 + 操作台可用”，尚未形成统一的供应商治理平台。

## 1. 专题目标

把当前“多个供应商已经接入、并能为研究链提供数据”的状态，升级为一个可治理、可路由、可评分、可审计、可扩展的统一市场数据底座。

专题A最终要解决的不是“再接几个源”，而是以下 5 个核心问题：

- 不同资产域、数据域对供应商能力的描述仍分散在代码分支里，缺少统一 registry
- truth 构建和上下文消费仍有不少“链路内决策”，没有完全沉到治理层统一处理
- fallback 目前可用，但策略、原因、质量影响和降级语义还没有统一表达
- Admin 数据源页偏操作台，不是完整的数据供应商治理台
- 下游虽然已经能读 freshness / warnings / source summary，但还没有稳定消费治理层摘要

## 2. 当前基线

### 2.1 已经完成的能力

- Go 后端已经接入多源股票、期货、资讯、期货仓单/库存与主数据同步
- 当前已有 source priority、requested source -> resolved source、truth 重建、质量日志记录
- `strategy-engine` 上下文默认优先使用 Go 后端统一真相源，并保留受控 fallback
- Admin 已有 `数据源管理` 页面，可执行同步、健康检查、默认源切换、truth 重建和质量日志查看
- 当前系统已接入 `TUSHARE`、`AKSHARE`、`TICKERMD`、`MYSELF`、`MOCK` 等来源，并形成股票 / 期货 / 资讯的多源切换能力

### 2.2 当前基线的代表性实现

- 多源同步与 source priority：`backend/internal/growth/repo/market_data_multi_source.go`
- 主数据同步与 truth：`backend/internal/growth/repo/market_instrument_master_data.go`
- 数据质量摘要与 truth 管理：`backend/internal/growth/repo/market_data_admin.go`
- `strategy-engine` 上下文读取：`backend/internal/growth/repo/strategy_engine_context_repo.go`
- Admin 数据源页：`admin/src/views/DataSourcesView.vue`
- Admin 数据治理辅助函数：`admin/src/lib/market-data-admin.js`

### 2.3 当前问题

当前系统的问题不是“完全没有治理”，而是“治理能力已经分散存在，但还没有被抽象成平台层”。

主要体现在：

- provider 能力是写死在同步分支和前端 provider 判断里的，没有统一 capability schema
- 股票、期货、资讯、主数据分别维护优先级和 fallback 语义，缺少统一 routing policy
- quality log 已有，但还没有沉淀为稳定的 freshness / trust / completeness 评分体系
- truth 构建部分已经集中，部分仍需在上下文消费链路里做补判断
- Admin 页面能做很多事，但用户看到的是“按钮集合”，不是“治理关系”
- 下游读取到的更多是运行结果摘要，而不是可复用的供应商治理摘要

## 3. 专题完成标准

专题A标记为 `Done` 时，至少满足以下标准：

### 3.1 统一模型

- 股票、期货、资讯、主数据、仓单/库存至少具备统一的 market data domain model
- provider registry 成为供应商声明真相源，不再由各链路硬编码支持范围
- capability schema 能明确描述“哪个 provider 在哪个资产域/数据域支持什么能力”

### 3.2 统一路由

- source routing、fallback、降级原因、禁止条件、默认优先级都走统一 policy
- truth 构建不再依赖每条链路各自拼接 source priority，而是回读治理层决策
- 上下文构建、同步作业、truth 重建、研究运行都能读到统一的 routing 摘要

### 3.3 统一评分

- freshness、quality、trust、coverage、stability 至少形成一版统一评分体系
- 评分既可用于 Admin 展示，也可用于 routing 决策和降级 warning
- 质量日志不再只是“问题明细表”，而能汇总成治理建议

### 3.4 统一后台治理

- `数据源管理` 从操作台升级为治理台，能展示 provider、能力矩阵、优先级、备源、健康度、质量分与建议动作
- 用户可以明确看到“当前默认源是谁、备源是谁、为什么切换、切换风险是什么”
- 页面不要求新建第二个大后台模块，但要形成清晰的信息架构

### 3.5 统一下游消费

- `strategy-engine`、`strategy-graph`、Admin、Client 只消费 Go 后端统一真相源与治理摘要
- 新增治理能力全部走“只增不改”，不破坏现有阶段8/9已收官链路

## 4. 非目标

以下内容明确不并入专题A验收：

- 实时 tick 级行情撮合与毫秒级路由
- 商业采购、账单、供应商成本中心与合同系统
- 高频资讯语义理解与事件审核台
- 更深层股票事件图谱与行业/主题语义建模
- 更重的期货库存、交割地、品牌、等级、跨市场研究底座

这些方向分别在后续专题B、专题C、专题D、专题E、专题F中处理。

## 5. 设计原则

- 统一真相源：下游不能各自再决定“到底信谁”，治理层给出稳定答案
- 只增不改：旧同步接口、旧 truth 表、旧上下文字段、旧 Admin 操作不立即删除
- 路由先于消费：provider 选择与 fallback 先统一，之后再扩下游展示
- 评分先服务治理，再谨慎介入业务：第一版先让评分解释决策，不直接做黑盒自动切源
- 资产域与数据域分离：股票 / 期货是资产域，日行情 / 主数据 / 资讯 / 仓单是数据域，避免混在一起
- 可降级：单个 provider 失败不应让 run 或发布链路整体崩溃

## 6. 统一数据域模型

专题A推荐将多源治理抽象为 4 层模型。

### 6.1 资产域

- `STOCK`
- `FUTURES`

### 6.2 数据域

- `DAILY_BARS`
- `INSTRUMENT_MASTER`
- `NEWS_ITEMS`
- `FUTURES_INVENTORY`
- 预留后续扩展：
  - `RESEARCH_EVENTS`
  - `FUND_FLOW`
  - `BASIS_TERM_STRUCTURE`
  - `WAREHOUSE_RECEIPTS`

### 6.3 治理域对象

- `ProviderRegistry`
  - 描述一个供应商实体是谁、状态如何、认证方式是什么、默认超时与重试策略是什么
- `ProviderCapability`
  - 描述供应商在某个资产域 + 数据域下支持什么操作、覆盖什么粒度、是否允许 truth、是否允许 fallback
- `RoutingPolicy`
  - 描述默认优先级、备源顺序、禁用条件、质量阈值、降级策略
- `QualityProfile`
  - 描述 freshness、coverage、trust、stability、completeness 的分值与原因
- `TruthDecision`
  - 描述某个 trade date / instrument / data kind 最终选择了谁、为什么、是否降级、有哪些 warning

### 6.4 下游消费对象

治理层对下游提供的不是 provider 明细堆砌，而是一层稳定摘要：

- `selected_source`
- `selected_trade_date`
- `fallback_chain`
- `quality_score`
- `freshness_status`
- `warnings`
- `decision_reason`

## 7. Provider Registry 与 Capability Schema

### 7.1 Provider Registry

当前代码里已经存在 `DataSource` 和 provider config，但专题A需要把它升级为“可治理声明”。

建议 registry 至少稳定承载：

- `provider_key`
- `provider_name`
- `provider_type`
- `status`
- `auth_mode`
- `endpoint`
- `timeout_ms`
- `retry_policy`
- `health_policy`
- `rate_limit_policy`
- `cost_tier`
- `supports_truth_write`
- `supports_manual_sync`
- `supports_auto_sync`

### 7.2 Capability Schema

capability 不再写死在 `supportsSyncKind` 或多处 switch/case，而是抽为结构化声明。

建议最少包含：

- `asset_class`
- `data_kind`
- `supports_sync`
- `supports_truth_rebuild`
- `supports_context_seed`
- `supports_research_run`
- `supports_backfill`
- `supports_batch`
- `supports_intraday`
- `supports_history`
- `supports_metadata_enrichment`
- `requires_auth`
- `fallback_allowed`
- `priority_weight`

### 7.3 第一版 provider 口径

结合当前已接入来源，第一版建议先稳定治理以下 provider：

- `TUSHARE`
  - 股票、期货日线与主数据强
  - 资讯能力有限，名称和深度语义仍需其他源补充
- `AKSHARE`
  - 资讯、股票补充信息、部分市场详情能力较强
  - 稳定性与字段统一性需要治理层兜底
- `TICKERMD`
  - 作为备源和异常兜底使用
- `MYSELF`
  - 作为内部桥接 provider，承接新浪、腾讯等公开接口的封装接入
- `MOCK`
  - 仅用于测试、演示和极端 fallback，不作为生产默认真相源

## 8. Routing / Fallback Policy

专题A的重点不是“能 fallback”，而是把 fallback 升级为可解释、可治理、可约束的统一策略。

### 8.1 Routing 决策输入

- 资产域
- 数据域
- 请求类型
- instrument / symbol 范围
- trade date / date range
- provider capability
- provider status / health
- freshness score
- quality score
- trust score
- 用户或系统显式指定 source

### 8.2 Routing 决策输出

- `requested_source`
- `resolved_source_chain`
- `selected_source`
- `selected_reason`
- `fallback_applied`
- `fallback_reason`
- `blocked_sources`
- `warnings`

### 8.3 第一版统一路由规则

- 显式指定 source 时，先校验 capability 和状态，再决定是否放行
- 未显式指定时，先按数据域读取 routing policy，而不是直接读散落 config key
- provider 不支持当前数据域或能力时，视为“不可选”，而不是运行时报错后才回退
- freshness / trust 低于阈值时，可继续回退，但必须记录 decision reason
- `MOCK` 默认仅在显式允许或本地开发/测试场景可进入最终回退链

### 8.4 与当前实现的关系

当前已有：

- source priority config key
- requested source -> resolved source keys
- provider 分支处理
- fallback warning

专题A不是推倒这些逻辑，而是把它们抽出成为统一 routing layer，然后让现有同步、truth、context 逐步回读这层结果。

## 9. Quality / Freshness / Trust Scoring

当前已有 quality log，但还缺少统一评分模型。专题A建议第一版先形成可解释的治理评分，而不是复杂机器学习打分。

### 9.1 Freshness

衡量是否“够新”：

- 最新 trade date 是否达到预期
- source updated_at 是否超时
- 是否连续多日未更新
- 日终数据是否在 SLA 时间窗内完成

### 9.2 Coverage

衡量是否“够全”：

- 请求 symbols / contracts 覆盖率
- 历史区间完整度
- 关键字段缺失率
- 主数据别名、名称、行业、交易所等补齐程度

### 9.3 Trust

衡量是否“够可信”：

- 历史稳定性
- 与其他 provider 的偏差率
- 最近异常率
- 是否频繁回退
- 是否依赖大量代理逻辑

### 9.4 Stability

衡量是否“够稳”：

- 健康检查成功率
- 超时率
- 限流率
- 单次同步失败率

### 9.5 评分输出

每个 provider 在每个资产域 / 数据域维度都至少输出：

- `freshness_score`
- `coverage_score`
- `trust_score`
- `stability_score`
- `overall_score`
- `score_reasons`

这些评分既供 Admin 展示，也可供 routing policy 作为辅助条件，但第一版不做“评分自动完全接管优先级”。

## 10. Truth 构建如何升级

当前系统已经有 truth 表和重建逻辑，但 truth 决策还没有完全上升为治理层对象。

专题A建议升级方向如下：

### 10.1 从“重建结果”升级为“决策记录”

truth 不只保存最终结果，还应能稳定追踪：

- 候选来源链
- 最终选中源
- 选中原因
- 降级原因
- 评分摘要
- warning 摘要

### 10.2 从“链路内判断”升级为“治理层统一输出”

`strategy-engine` 上下文构建不再需要自己再推断太多 source 语义，而是直接读取：

- truth 结果
- truth 决策摘要
- freshness / quality 摘要

### 10.3 从“单表结果”升级为“可审计快照”

第一版不要求新增复杂大表群，但至少要支持：

- 针对每次重建或同步保留 routing / truth 决策摘要
- 能从 Admin 页面追溯“这一批 truth 为何选了该 provider”
- 能给 run / explanation 附上稳定来源说明

## 11. Admin 数据源管理如何升级为治理台

当前 `数据源管理` 页面已经可以做很多动作，但信息组织仍偏“运维面板”。

专题A建议保留原页面入口，但把内容升级为 4 个治理区块：

### 11.1 供应商总览

- provider 列表
- provider 状态
- 默认源 / 备源
- 最新健康状态
- 主要能力标签
- 最近质量分

### 11.2 能力矩阵

- 按资产域 + 数据域展示 provider capability
- 清楚看到谁是主源、谁是备源、谁不支持
- 支持查看 capability 缺口和建议补位

### 11.3 路由与真相源治理

- 查看默认 routing policy
- 查看 fallback 链
- 查看当前 truth 选源摘要
- 查看最近 source 切换、切换原因和风险提示

### 11.4 质量与异常

- 质量分趋势
- freshness 告警
- 异常 issue 分桶
- 需要人工处理的 provider 建议

第一版不要求新开复杂二级系统，但至少要从“按钮集合”升级为“治理视图 + 操作入口”。

## 12. 与其他系统的边界

### 12.1 与 `strategy-engine`

- `strategy-engine` 不直接对接第三方 provider
- 只消费 Go 后端统一上下文和 truth / routing 摘要
- explanation 可读取 `selected_source / warnings / freshness / decision_reason`

### 12.2 与 `strategy-graph`

- 图服务不是 provider 治理层的一部分
- 但图谱 enrich 所依赖的市场数据和资讯数据应回读治理层真相源
- run graph snapshot 可附带治理摘要，但不接管 provider 路由

### 12.3 与 Admin

- Admin 不直接拼底层 provider 逻辑
- 所有治理展示统一走 Go 后端治理接口
- 保留同步、重建、健康检查动作，但这些动作应回流治理摘要

### 12.4 与 Client

- Client 不直接关心 provider 明细
- 只消费解释级摘要，例如来源、鲜度、warning、可信度说明
- 新增字段走“只增不改”，旧 explanation fallback 保留

## 13. 建议实施分包

专题A建议按 4 个包落地，每个包都可以独立验收。

### 包A：Provider Registry 与 Capability Schema

目标：

- 把当前 provider 分支、supports 判断、source type 判断抽成统一声明层

交付物：

- provider registry schema
- capability schema
- 初始 provider 能力填充
- 现有同步链路改为优先读取 capability

### 包B：Routing Policy 与 Fallback 治理

目标：

- 把 source priority、fallback、禁用逻辑、MOCK 约束统一到 routing layer

交付物：

- routing policy schema
- 统一 resolved source 决策器
- truth 决策摘要
- 同步、truth、context 三条链路回读 routing 决策

### 包C：Quality / Freshness / Trust Scoring

目标：

- 让质量日志升级为评分与治理建议系统

交付物：

- freshness / coverage / trust / stability / overall score
- 评分原因摘要
- provider 质量画像
- Admin 可读的治理建议摘要

### 包D：Admin 数据治理台第一版

目标：

- 在现有 `数据源管理` 页基础上升级治理体验

交付物：

- provider 总览区
- capability matrix
- routing / truth 治理区
- 质量与异常区
- 下游依赖摘要展示

## 14. 测试与验收

### 14.1 Backend

- provider registry / capability / routing / scoring repo 与 service 测试
- 多个数据域的 fallback 与禁用规则测试
- truth 决策摘要写入与查询测试
- 上下文构建读取治理摘要的兼容测试

推荐回归命令：

- `cd backend && go test ./internal/growth/... ./router/...`

### 14.2 Admin

- 治理页构建通过
- provider、能力矩阵、质量摘要、truth 路由摘要正常展示
- 同步、重建、健康检查仍可用且动作结果能回流治理区

推荐回归命令：

- `cd admin && npm run build`

### 14.3 端到端验收

- 股票：同步 -> truth 重建 -> 上下文 -> run -> explanation，来源与 warning 一致
- 期货：同步 -> truth 重建 -> 上下文 -> run -> explanation，来源与 warning 一致
- 指定 source、自动 source、fallback source 三种路径都能追溯决策摘要
- provider 故障时 run 不丢失，Admin 可看见降级原因和建议动作

## 15. 默认约定

- 专题A不推翻当前多源实现，而是在现有基线上抽象治理层
- 旧 config key、旧同步接口、旧 truth 重建入口先保留，直到新治理层完全接管
- `MYSELF` 继续作为内部桥接 provider，可承接新浪、腾讯等公开接口的封装能力
- `MOCK` 继续保留，但默认不作为生产真相源主来源
- 专题A是后续专题B、专题C、专题D、专题E、专题F的共同底座，优先级最高

## 16. 与总路线图的关系

本文件是：

- `docs/strategy-engine/专题路线图-多源市场与研究治理平台.md`

中“专题A：多源市场数据统一模型与供应商治理”的正式专题文档。

后续如果进入开发，应以本文件作为专题A真相源；总路线图继续只负责专题排序、依赖关系和整体节奏，不再承载专题A全部细节。

专题A第一轮的可执行启动清单见：

- `docs/strategy-engine/专题A-第一轮-开发启动清单.md`
