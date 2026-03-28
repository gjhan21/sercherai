# 股票与期货预测增强 L1 设计规范

## Summary

`L1` 的目标是先把现有系统从“会推荐”提升到“会做更严谨的研究编排，并能从历史结果里学习”。

本阶段不做重型模拟，不把 `MiroFish` 直接接进主链，而是在现有 `strategy-engine + insight + evaluation` 体系上补 4 件最值的事：

1. 子问题拆解式研究编排
2. 当前有效理由 / 历史失效理由分层
3. 基于真实评估结果的记忆回灌
4. 置信度与风险边界校准

本阶段交付应优先增强现有：

- 股票推荐 `insight`
- 股票推荐 `version-history`
- 期货策略 `insight`
- 期货策略 `version-history`

不要求首期就新增完整前台独立分析页。  
本阶段只要求字段结构和内部编排可以向未来“单标的分析预测页”扩展，不要求新增页面、接口或交互入口。

## Problem Statement

当前系统已经具备推荐、解释、版本历史和表现评估，但还存在以下问题：

1. 解释更像结果拼装，缺少清晰的研究步骤与子问题视角。
2. 当前理由和已经失效的旧理由没有明确拆层。
3. 评估结果虽已存在，但没有系统化进入下一轮推荐。
4. 置信度和风险边界更多是输出字段，还不是由历史表现和记忆反馈共同校准。

这些问题会直接影响：

- 推荐准确率
- 错误重复发生率
- 后续走势解释的可信度
- 用户对推荐理由的信任度

## Implementation Goals

### 目标一：让每个推荐结果都有更完整的研究拆解

系统需要能回答：

- 这只股票 / 这个合约为什么进入候选？
- 支撑它的核心事实分哪几类？
- 哪些理由是当前有效的？
- 哪些旧理由已经失效？
- 接下来应该盯哪些观察哨兵？

### 目标二：让历史评估结果变成下一轮输入

系统需要能回答：

- 过去类似推荐在哪些情况下容易失败？
- 失败是因为方向错、节奏错，还是回撤承受不了？
- 哪类逻辑在什么市场环境里更容易失效？

### 目标三：不改主产品形态，也能先提高质量

本阶段默认沿用现有接口与页面链路，让前台先通过现有 insight 页面受益。

## L1 Authoritative Contract

为了避免后续线程把 `L1` 理解成不同方向，本阶段硬约束固定如下。

### 1. 生效方式

`L1` 的 `memory_feedback` 与 `confidence calibration` 必须被下一轮 explanation 生成过程消费，不能只是写进返回字段。

但在 `L1` 中，它们的生效范围固定为：

- 可以影响 explanation 内容
- 可以影响 risk boundary
- 可以影响 confidence summary / warning flags
- 可以影响 reviewer advisory priority

在 `L1` 中明确不允许：

- 直接改动 candidate ranking
- 直接改动 portfolio weight
- 直接阻断 publish review
- 直接替代现有量化 / 规则主打分链

### 2. 最小字段 contract

`L1` 首版必须稳定输出以下结构：

- `memory_feedback`
  - 继续沿用现有 `summary / suggestions / failure_signals / items`
- `active_thesis_cards`
  - 当前仍有效的理由卡
- `historical_thesis_cards`
  - 曾成立但当前已弱化的理由卡
- `watch_signals`
  - 下一步需要观察的触发 / 验证 / 失效信号
- `confidence_calibration`
  - `base_confidence`
  - `adjusted_confidence`
  - `drivers`
  - `advisory_only`

以上字段中：

- `memory_feedback` 与现有 explanation 字段体系兼容
- 其他新增字段均为非破坏性扩展

### 3. 计算与持久化策略

`L1` 首版默认不新增新的持久化业务表作为前提。

首版规则固定为：

- 以现有 run / evaluation / version-history / report snapshot 为事实源
- 在 explanation 构建时同步计算增强结果
- 在 insight 与 version-history 读取时同步生成可读结果
- 如后续需要缓存，只能作为性能优化，不改变事实源

### 4. 刷新链路

`L1` 首版只允许两条刷新链：

1. 推荐 / 策略 explanation 生成或重建时同步刷新
2. 评估回补完成后，在下一次 insight / history 读取时消费最新结果

本阶段不额外新增独立异步研究刷新系统。

### 5. 首期覆盖范围

`L1` 首版必须覆盖且只覆盖：

- 股票推荐 `insight`
- 股票推荐 `version-history`
- 期货策略 `insight`
- 期货策略 `version-history`

本阶段不要求覆盖：

- watchlist
- archive 页面独立接口
- community
- 独立 analysis 页

## Implementation Changes

### 1. 新增研究编排层，服务于股票与期货两个域

在现有上下文构建之上，新增一个内部“研究编排层”，负责把单标的分析拆成多个固定问题槽位，而不是只给一段总理由。

股票建议的子问题模板至少覆盖：

- 趋势与价量结构
- 资金承接与流向
- 估值与基本面约束
- 行业 / 主题 / 事件催化
- 风险边界与失效条件

期货建议的子问题模板至少覆盖：

- 方向与关键价位
- 供需 / 库存 / 产业链线索
- 基差 / 期限结构 / 价差线索
- 宏观 / 政策 / 事件扰动
- 风险边界与失效条件

要求：

- 子问题模板按资产类型区分
- 输出结果保持结构化
- 研究编排结果可以回填到现有 `StrategyClientExplanation`

### 2. 建立“当前有效理由 vs 历史失效理由”分层

本阶段把现有 explanation 中的理由信息拆成四层：

1. Active Thesis
   当前仍然成立的核心理由
2. Historical Thesis
   曾经成立但当前已弱化或不再主导的理由
3. Invalidations
   已明确被证伪或失效的理由
4. Watch Signals
   尚未发生但需要持续盯住的观察点

这一步不要求做完整图谱，只要求把现有信息重新编排成稳定的研究结构。

优先复用的事实源包括：

- strategy-engine asset context
- evidence cards
- related entities / events
- reviewed stock event evidence
- version diff
- evaluation summary

### 3. 把评估回补结果升级为结构化记忆反馈

当前已有 `1 / 3 / 5 / 10 / 20` 日评估回补，本阶段要把这些结果进一步加工成策略记忆。

至少要沉淀这几类记忆：

- 高命中但回撤过深
- 方向正确但验证过慢
- 新闻驱动强但持续性差
- 主题切换快导致逻辑失效
- 高波动标的对原风险边界不友好
- 某类市场环境下某类模板持续失效

这些记忆需要：

- 可按股票 / 期货区分
- 可按市场环境聚合
- 可按模板 / profile / strategy version 聚合
- 可写回 `memory_feedback`

### 4. 增加轻量置信度与风险边界校准器

本阶段不引入复杂机器学习模型，先做基于规则和历史结果的轻量校准。

校准输入固定包括：

- 近期评估命中率
- 最大回撤表现
- 同模板在当前市场环境的稳定性
- 当前理由的证据覆盖度
- 历史失效信号是否再次出现

校准输出固定包括：

- 调整后的 confidence summary
- 更明确的 risk boundary
- 更早的 invalidation trigger
- reviewer advisory priority

### 5. 非破坏性扩展现有 explanation / history 输出

本阶段优先通过现有接口承接，不新增必须切换的前端主链路。

可在现有 `StrategyClientExplanation` 与历史项中做非破坏性扩展。`L1` 首版新增字段固定为：

- `research_outline`
- `active_thesis_cards`
- `historical_thesis_cards`
- `watch_signals`
- `confidence_calibration`

原则：

- 旧前端不依赖这些字段也能正常工作
- 新前端可以按需读取并增强展示

### 6. 建立可回补的 L1 数据刷新机制

L1 不是一次性计算完就结束，需要能持续更新。

首版固定支持：

- 新发布推荐后自动生成或刷新增强 explanation
- 评估回补完成后重算 memory feedback
- 对近期重点推荐做批量 backfill 接口复用，不另起新链路

## Candidate Module Boundaries

`L1` 首版固定落在 `backend/internal/growth` 内部，不另起全新服务。

职责边界固定为：

- `repo/context` 继续负责拉原始上下文
- `repo/evaluation` 继续负责评估回补与历史表现
- 新增 `research orchestrator` 负责子问题拆解与证据分层
- 新增 `memory calibrator` 负责从评估结果产出结构化记忆
- `strategy_client_explanation` 负责最终字段融合

## Public APIs / Interfaces / Types

### 保持不变

- 不修改现有接口路径
- 不修改现有请求参数
- 不打破现有响应主体的兼容性

### 允许的非破坏性扩展

- 扩展 `StrategyClientExplanation`
- 扩展 `StrategyVersionHistoryItem`
- 扩展现有 `evaluation_meta`
- 扩展 `memory_feedback`

### 本阶段不新增

- 不新增必须依赖的新前台页面
- 不新增外部 `MiroFish` 直连 API
- 不新增聊天式分析接口

## Test Plan

### 后端单元验证

- 股票与期货子问题模板生成正确
- evidence 分层结果正确
- active / historical / invalidated / watch 四层不混淆
- 评估结果能产出结构化记忆反馈
- 置信度校准在不同输入下输出稳定

### 集成验证

- 股票 `insight` 返回增强 explanation
- 股票 `version-history` 返回增强历史结构
- 期货 `insight` 返回增强 explanation
- 期货 `version-history` 返回增强历史结构
- 无评估数据时可以平稳降级

### 回归验证

- 现有推荐列表、详情、表现接口不回退
- 现有 explanation 旧字段不缺失
- 现有前台在不读取新字段时仍正常渲染

## Out Of Scope

- 不做完整金融知识图
- 不做多智能体深度模拟
- 不做深度采访式研究
- 不做新独立分析页完整上线
- 不做全市场任意标的一步到位搜索分析

## Exit Criteria

满足以下条件可认为 `L1` 完成：

1. 股票与期货 `insight` 均具备更强的研究结构
2. 当前有效理由与历史失效理由已明确拆层
3. 历史评估结果已被 explanation 生成过程消费，而不只是写入字段
4. 置信度与风险边界已经引入校准输入，但不直接改动排序与发布链
5. 首版仅覆盖股票 / 期货的 insight 与 version-history 四个入口
6. 不破坏现有用户侧主要链路

## Next Step

`L1` spec 确认后，应先进入实现计划拆解，再开始开发。  
`L2 / L3` 暂不实现，只冻结设计边界。
