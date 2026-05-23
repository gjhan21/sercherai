# 深度推演二期证据增强设计规范

## Summary

深度推演一期已经把 `forecast-lab` 从运行工作台推进到了“研究入口 + 结构化报告”的基础形态，但它的核心问题仍然存在：

- 研究包仍然过薄，更多像 `thesis + highlights + history` 的通用摘要容器
- 股票与期货报告虽然已经有“证据支撑”区，但维度证据仍偏泛化，缺少真正的领域判断
- 前端报告页已经有结构，但证据内容还不足以支撑“全方位解析股票/期货状态并给出后续操作建议”的产品目标

二期的目标不是再造一套新产品，也不是提前进入三期学习闭环，而是把现有仓库里已经存在的数据真正接进深度推演主链，形成“现有数据 -> 结构化领域证据 -> 情景判断 -> 研究报告展示”的完整闭环。

本期覆盖 `股票 + 期货`，策略是：

1. 扩展 `Research Pack`，从通用摘要升级为“通用研究层 + 领域证据层”
2. 用统一接口输出股票五维、期货五维领域判断
3. 强化 `Scenario Engine` 与 `Report Schema`，让报告真正先讲状态、再讲证据、最后讲模型复核
4. 调整 `Forecast Lab` 和 `Forecast Report` 前端重心，使用户感受到“研究系统”而不是“run 工作台”

## Background

当前仓库已经具备以下基础：

- 一期已完成用户主链深度推演接入：
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastLabView.vue`
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastLabView.vue`
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue`
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue`
- 后端已支持严格上下文的 `forecast run` 创建、结构化报告字段、以及 `USER_REQUEST` 成功 run 的 LLM 复核降级链路：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`
- 股票侧仓库现成就有一批可用于研究结论增强的数据：
  - `Momentum5 / Momentum20 / Volatility20 / VolumeRatio / Drawdown20 / TrendStrength`
  - `FlowScore / ValueScore / NewsScore / NetMFAmount`
  - `PeTTM / PB / TurnoverRate / NewsHeat / PositiveNewsRate`
  - `RelatedNews / PerformanceStats / Explanation / ScoreFramework`
- 期货侧仓库现成也有一批可用于研究结论增强的数据：
  - `BasisPct / CarryPct / TermStructurePct / CurveSlopePct / BasisTermAlignment`
  - `InventoryLevel / InventoryChangePct / InventoryPressure`
  - `OIChangePct / TurnoverRatio / FlowBias / SpreadPressure`
  - `RelatedNews / RelatedEvents / Regime / Explanation / PerformanceStats`

因此，二期的主问题不是“缺少任何数据”，而是这些数据还没有真正进入深度推演主链，最终导致报告的领域分析层不够厚。

## Goal

在不引入新外部数据源的前提下，把深度推演二期升级为“证据层更厚、研究结论更像样、股票与期货都能成立”的研究报告系统，使用户能够：

1. 在股票和期货的深度推演报告里看到真正分维度的证据判断
2. 明确理解当前状态、主情景、风险边界、后续操作建议之间的关系
3. 区分“主报告结论”与“模型复核意见”的职责边界
4. 在 `Forecast Lab` 中优先看到“研究结论”和“研究对象”，而不是泛化的 run 视角

## Non-Goals

本设计明确不做以下内容：

- 不接入新的外部财报、行业数据库、期货宏观终端等外部数据源
- 不做三期的学习闭环：同标的多次 run 对比、命中率回写、失效识别反馈
- 不新增复杂图表终端或量化看板
- 不重写整套深度推演路由结构
- 不把股票和期货拆成两套完全不同的前端报告系统
- 不改变一期已经确定的 `LLM 复核只对 USER_REQUEST 成功 run 默认展示` 原则

## User Decisions Locked In

用户在设计阶段已确认以下边界：

- 当前进入的是“下一阶段开发”的二期，不是三期学习闭环
- 二期范围按 `股票 + 期货一起做基础增强版`
- 二期重点是把深度推演做成更完整的研究系统，而不是只改视觉
- 继续沿用一期的产品关系：`Forecast Lab` 是入口与研究中心，`Forecast Report` 是结果页
- `LLM 复核` 继续只对 `USER_REQUEST` 成功 run 默认展示

## Product Contract

### 1. 深度推演的二期产品定位

二期之后，深度推演的产品定位仍然分成两层：

- `Forecast Lab`
  - 研究入口与承接层
  - 负责承接推荐、策略、分析、详情链路
  - 负责展示“最近研究结论”和“运行动态”
- `Forecast Report`
  - 单一标的研究结果层
  - 负责按固定结构输出“状态 -> 结论 -> 情景 -> 风险 -> 动作 -> 证据 -> 模型复核”

二期不会把 `Forecast Lab` 做成新的独立内容宇宙，也不会把它退回成 run 列表页。

### 2. 报告固定骨架

`Forecast Report` 二期固定展示以下 7 个主区块：

1. `当前状态`
2. `核心判断`
3. `主情景`
4. `风险边界`
5. `后续操作建议`
6. `证据支撑`
7. `模型复核`

前端在这 7 个区块之外，可继续保留：

- `备选情景`
- `验证清单`
- `失效信号`
- `角色分歧`
- `运行证据`
- `原始报告全文`

但这些区域不再是用户理解报告的主入口。

### 3. 股票与期货的统一与分叉

二期前端不拆成股票、期货两套完全不同的报告骨架。

统一部分：

- 页面结构
- 状态页
- 当前状态 / 核心判断 / 主情景 / 风险边界 / 后续操作建议 / 模型复核

分叉部分：

- 证据支撑区的维度名称与文案
- 领域分析器与情景判断逻辑

统一骨架、分叉证据，是二期控制复杂度和保证一致体验的核心原则。

## Architecture

二期后端仍然沿用一期的主链路：

1. `Run Intake`
2. `Research Pack Builder`
3. `Domain Analyzer`
4. `Scenario Engine`
5. `Validation`
6. `Report Builder`

但二期重点增强中间三层，而不是重写整个系统。

### 1. Run Intake

二期不改变一期已经落定的严格上下文原则：

- `USER_REQUEST` 必须带 `target_id / target_key / target_type / target_label / source / source_id / source_path / reason`
- 上下文不足时，继续返回明确 `400`
- 不恢复“只有 symbol/name 也能当高质量路径发起 run”的行为

### 2. Research Pack Builder

二期把 `strategyForecastL3ResearchPack` 从“一层通用摘要包”升级为“两层结构”：

- `通用研究层`
- `领域证据层`

### 3. Domain Analyzer

二期不再继续依赖“旧角色名映射新维度”的思路来产出主证据，而是直接生成股票五维 / 期货五维判断。

### 4. Scenario Engine

二期不只是产出一个主情景名称，而要根据维度证据收敛出：

- `current_state`
- `primary_scenario`
- `secondary_scenarios`
- `trigger_conditions`
- `invalidation_conditions`
- `action_plan`

### 5. Validation

LLM 继续作为“验证器 + 反证器 + 报告整合器”，不是主结论引擎。

### 6. Report Builder

结构化字段成为主渲染源，`markdown_body / html_body` 继续保留，但降级为兼容与导出层。

## Research Pack Design

### 1. 通用研究层

所有标的共用以下研究上下文字段：

- `target_context`
- `source_context`
- `core_thesis`
- `risk_boundary`
- `invalidations`
- `historical_notes`
- `related_highlights`
- `performance_summary`
- `l2_summary`

这层的职责是：

- 记录本次推演对象是谁、从哪条主链进入
- 记录 L2 已有结论、历史版本线索、绩效摘要、风险边界
- 给后续领域证据和情景引擎提供统一背景

### 2. 股票领域证据层

股票侧二期使用现有仓库数据，拆成以下五组：

- `fundamental_evidence`
  - 用 `reason summary / explanation / performance summary / value hints` 形成代理基本面
- `technical_evidence`
  - 用 `Momentum5 / Momentum20 / TrendStrength / Volatility20 / Drawdown20 / VolumeRatio`
- `flow_evidence`
  - 用 `NetMFAmount / FlowScore / TurnoverRate`
- `valuation_evidence`
  - 用 `PeTTM / PB / ValueScore`
- `event_evidence`
  - 用 `RelatedNews / NewsHeat / PositiveNewsRate / NewsScore`

这意味着二期的股票“基本面”仍然是代理基本面，而不是外部财报深挖；这是本期主动锁定的边界，不属于缺陷。

### 3. 期货领域证据层

期货侧二期使用现有仓库数据，拆成以下五组：

- `supply_demand_evidence`
  - 用 `InventoryLevel / InventoryChangePct / InventoryPressure / InventoryBrandGradeSummary / InventoryFocusArea`
- `term_structure_evidence`
  - 用 `BasisPct / CarryPct / TermStructurePct / CurveSlopePct / BasisTermAlignment / SpreadPercentile`
- `tape_technical_evidence`
  - 用 `TrendStrength / Volatility14 / VolumeRatio / SpreadPressure`
- `position_flow_evidence`
  - 用 `OIChangePct / TurnoverRatio / FlowBias` 以及仓库中已有的公开持仓线索
- `macro_event_evidence`
  - 用 `RelatedNews / RelatedEvents / Regime`

### 4. 统一输出格式

无论股票还是期货，每个维度最终都必须产出统一结构：

- `dimension`
- `stance`
- `confidence`
- `summary`
- `supporting_points`
- `risk_points`

这样前端报告页只需要一套骨架即可消费。

## Domain Analyzer Design

### 1. 股票五维

股票的五个正式维度固定为：

- `FUNDAMENTAL`
- `TECHNICAL`
- `FLOW`
- `VALUATION`
- `EVENT`

每个维度必须独立生成：

- 当前偏向
- 置信度
- 摘要判断
- 支撑点
- 风险点

### 2. 期货五维

期货的五个正式维度固定为：

- `SUPPLY_DEMAND`
- `TERM_STRUCTURE`
- `TAPE_TECHNICAL`
- `POSITION_FLOW`
- `MACRO_EVENT`

每个维度同样生成：

- 当前偏向
- 置信度
- 摘要判断
- 支撑点
- 风险点

### 3. 角色壳的处置

一期里的 `INDUSTRY / FLOW / EVENT / MACRO / RISK` 和 `SUPPLY_DEMAND / HEDGE / SPEC_FLOW / MACRO / RISK` 角色逻辑不再承担主证据职责。

二期允许保留它们作为：

- 兼容历史字段
- 运行日志中的分析痕迹
- 原始报告全文的内部组织依据

但主报告的结构化证据必须来自领域分析器，而不是旧角色结果简单映射。

## Scenario Engine Design

二期的情景引擎从“依赖 L2 主情景 + 角色立场做轻组合”升级为“依据五维证据收敛出状态与情景”。

### 1. 固定输出

情景引擎必须产出：

- `current_state`
- `primary_scenario`
- `secondary_scenarios`
- `trigger_conditions`
- `invalidation_conditions`
- `action_plan`

### 2. 当前状态的要求

`current_state` 不再只是一个情景名，而是“当前所处状态”的归纳。

股票示例方向：

- `趋势延续`
- `等待确认`
- `高位兑现风险抬升`
- `估值承压但资金仍在支撑`

期货示例方向：

- `供需偏紧但结构分化`
- `基差走强，近月支撑占优`
- `持仓拥挤，趋势仍强但波动放大`
- `库存改善，驱动减弱`

### 3. 主情景与备选情景

`primary_scenario` 是最可能的后续演化路径。

`secondary_scenarios` 必须是真正的替代路径或反向路径，不再使用固定模板填充。

### 4. 动作计划

`action_plan` 不再只是止盈止损拼接，而是面向研究结论的动作建议，例如：

- 等待确认后再参与
- 沿趋势试探，但仓位控制
- 高位不追，优先看兑现节奏
- 驱动减弱，先观察库存与结构变化

## Validation Design

### 1. 触发原则

二期继续保持一期既定原则：

- 只对 `USER_REQUEST` 成功 run 默认展示 `模型复核`
- `ADMIN_MANUAL` 与 `AUTO_PRIORITY` 可以保留结构化报告，但不默认展示模型复核区

### 2. 输入增强

LLM 的输入不再只包含：

- `core_thesis`
- `risk_boundary`
- `invalidations`
- `action_hints`
- `role summaries`

而必须补入：

- 五维领域证据摘要
- 每个维度的支撑点
- 每个维度的风险点
- 收敛后的状态和主情景

### 3. 输出边界

LLM 继续输出结构化结果：

- `verdict`
- `scenario_consistency`
- `supporting_evidence`
- `counter_evidence`
- `blind_spots`
- `risk_review`
- `action_review`
- `llm_summary`

它的角色仍然是：

- 检查主情景是否自洽
- 补充反证与盲点
- 复核动作建议

它不负责替代主报告。

## Report Schema Design

二期结构化报告继续保留一期字段兼容，但主渲染语义固定如下：

### 1. 当前状态

由 `state_assessment` 承载：

- `current_state`
- `risk_boundary`
- `source`
- `context_quality`

### 2. 核心判断

由以下字段共同承载：

- `headline_verdict`
- `executive_summary`

### 3. 主情景

由以下字段共同承载：

- `primary_scenario`
- `scenario_assessment.primary_scenario`
- `scenario_assessment.trigger_conditions`
- `scenario_assessment.secondary_scenarios`

### 4. 风险边界

由以下字段共同承载：

- `state_assessment.risk_boundary`
- `scenario_assessment.invalidation_conditions`
- `validation_review.risk_review`

### 5. 后续操作建议

由以下字段共同承载：

- `action_guidance`
- `scenario_assessment.action_plan`
- `validation_review.action_review`

### 6. 证据支撑

由 `dimension_evidence` 承载，成为二期主渲染字段。

### 7. 模型复核

由 `validation_review` 承载。

### 8. 兼容字段策略

以下字段继续保留，但不再承担主结构职责：

- `markdown_body`
- `html_body`
- `role_disagreements`

`markdown_body / html_body` 的内容由结构化报告反向生成，而不是前端主渲染源。

## Forecast Lab Product Changes

二期不重做 `forecast-lab` 的路由与主链承接方式，但页面重心从“运行工作台”迁移到“研究入口”。

### 1. 研究聚焦区

当用户带上下文进入时，聚焦卡要额外展示：

- 标的类型
- 最近研究状态
- 是否已有结构化报告
- 是否已有模型复核

### 2. 最近研究结论区

首屏优先展示最近可读的研究卡片，而不是泛化 run 列表。每张卡片至少展示：

- 标的名称
- 当前状态
- 一句话结论
- 主情景
- 风险边界摘要
- 模型复核状态

### 3. 运行动态区

把 `queued / running / failed` 运行状态独立出来，作为“运行动态”而不是和研究结论混排。

### 4. 来源承接区

保留从推荐、策略、分析、详情链路进入的回链，但视觉上降级为辅助，不再占主舞台。

## Forecast Report Product Changes

### 1. 页面骨架不变，内容重心增强

继续沿用一期固定的 7 个主区块，但二期显著增强：

- `证据支撑`
- `当前状态`
- `后续操作建议`

### 2. 证据支撑区

股票展示：

- `基本面`
- `技术面`
- `资金面`
- `估值面`
- `事件面`

期货展示：

- `供需库存`
- `期限结构`
- `盘面技术`
- `持仓资金`
- `宏观事件`

每个维度卡片固定展示：

- 当前立场
- 置信度
- 维度摘要
- 支撑点
- 风险点

### 3. 模型复核区

`模型复核` 区只在 `USER_REQUEST + SUCCEEDED` 时主展示。

对未完成、跳过、降级的情况，页面显示：

- 未完成模型复核
- 当前主报告仍可阅读

### 4. 原始全文

`原始报告全文` 继续保留，但降级为折叠区，不再占据主阅读流。

## API and Type Contract

二期不新增新的用户主接口路径，但强化已有字段语义。

### 1. `GET /api/v1/forecast/runs/:id`

`report` 中以下结构化字段成为主消费字段：

- `state_assessment`
- `headline_verdict`
- `dimension_evidence`
- `scenario_assessment`
- `validation_review`

### 2. `StrategyForecastL3DimensionEvidence`

该结构继续作为外部接口主证据项，但二期要求其内容来自真实领域分析器，而不是泛化高亮拼装。

### 3. 兼容约束

以下兼容字段继续返回，避免影响历史详情页和历史数据读取：

- `executive_summary`
- `primary_scenario`
- `action_guidance`
- `markdown_body`
- `html_body`

## File-Level Design

### Backend

主要修改文件：

- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`

必要时扩展读取现有期货库存、结构、持仓线索的 reader 接口，但不引入外部服务。

### Frontend

主要修改文件：

- `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastLabView.vue`
- `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastLabView.vue`
- `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue`
- `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue`
- `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.js`
- `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-localization.js`

允许新增一个专用的前端 report view model 层，避免详情页组件继续膨胀。

## Testing Strategy

二期继续沿用当前仓库的 `node:test + go test + 源码契约测试` 方式，不引入新测试框架。

### Backend Tests

必须覆盖：

- `Research Pack Builder`
  - 股票能抽出估值、资金、技术、事件线索
  - 期货能抽出库存、基差、结构、持仓线索
- `Dimension Evidence Builder`
  - 股票五维输出稳定
  - 期货五维输出稳定
- `Scenario Assessment`
  - `current_state / trigger / invalidation / action_plan` 有稳定输出
- `Validation Prompt`
  - 五维证据进入 prompt
  - 无 LLM key 或 LLM 失败仍可安全降级

### Frontend Tests

必须覆盖：

- `Forecast Detail`
  - 股票成功态渲染五个股票维度
  - 期货成功态渲染五个期货维度
- `Forecast Lab`
  - 研究结论区和运行动态区同时存在
  - 聚焦标的卡在有无结构化报告时文案正确
- `入口链路`
  - 从 `recommendations / strategies / analysis / detail` 进入后承接态不丢

## Rollout Notes

二期是“一期结构上的增强”，不是产品关系翻转。

因此上线后预期变化应当是：

- 用户明显感觉深度推演更像一份研究报告
- `Forecast Lab` 更像研究中心，而不是运行清单
- 股票与期货都能看到更像样的分维度证据

而不是：

- 多出一个全新的一级模块
- 进入三期学习闭环
- 变成复杂的量化终端

## Open Risks

### 1. 股票基本面仍是代理基本面

这是二期主动接受的边界，不是遗漏。若要做真正财报与行业景气深挖，属于后续阶段。

### 2. 期货持仓证据的可读性依赖现有仓库线索质量

二期优先把现有线索接入；如果仓库已有字段粒度不足，允许先以摘要结论呈现。

### 3. Frontend 详情页复杂度继续增长

为控制风险，二期允许新增 shared view-model 层或局部子组件拆分，但不做无关大重构。

## Assumptions

- 一期已经在 `main` 上可用，二期建立在当前深度推演主链之上
- 本期不引入新的外部数据源
- 本期不改变 `USER_REQUEST` 的模型复核展示策略
- 本期不触碰三期学习记录与命中率回写能力
- 当前仓库里的股票 quant 数据与期货 context 数据足以支撑二期基础增强版
