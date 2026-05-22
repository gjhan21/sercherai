# 技术分析主导的双层选股系统设计规范

## Summary

本设计定义一套基于当前系统演进的 `双层选股系统`：

- 第一层：从 `5000+` 全市场股票中筛出 `趋势型待选股池`
- 第二层：从待选股池中分叉出两类输出
  - `超短线主推荐`
  - `短波段辅助推荐`

系统目标不是把所有分析因子平均混合，而是建立清晰分工：

- `大盘分析` 决定市场是否适合出手以及应采取的风险姿态
- `板块 / 题材分析` 决定去哪里找机会
- `技术分析` 决定结构是否成立、时点是否合适
- `资金分析` 决定机会是否有承接与持续性
- `事件 / 风险分析` 决定是否应避雷或降低优先级

本设计强调：

- `多种分析结合`
- `技术分析主导`
- `趋势型待选池`
- `超短线为主，短波段为辅`
- `共享上游待选池，分离下游决策头`

## Product Contract

### 1. 输出结构

每日输出固定分为三块：

- `大盘结论区`
- `超短线主推荐区`
- `短波段辅助区`

三块共享一套上游数据和待选股池，但下游决策规则不同。

### 2. 主辅关系

系统默认：

- `超短线` 是主推荐位
- `短波段` 是固定辅助位

不将两个周期混为同一种推荐。

### 3. 结果表达

每个推荐结果必须明确：

- 推荐周期
- 推荐理由
- 风险提示
- 失效条件
- 观察重点

超短线和短波段必须有独立的持有逻辑与评估口径。

### 4. Admin 承接原则

本方案不新造一套后台，而是复用现有 `/admin/stock-selection` 智能选股模块。

管理端职责不是重新做一遍策略分析，而是承接 5 类动作：

- `看市场`：查看大盘结论、运行结果和风险姿态
- `配策略`：配置大盘分析、待选池、超短线头、波段头
- `跑批次`：发起运行、比较运行、追踪阶段日志
- `审结果`：审核主推荐与辅推结果，决定是否发布
- `做复盘`：按推荐头拆分命中率、收益和回撤

因此，这套业务设计必须同时满足：

- `strategy-engine` 能输出分层结果
- `backend` 能持久化并暴露分层结果
- `admin` 能以现有页面结构读懂和操作这些结果

## System Architecture

系统拆为 4 个业务层：

1. `大盘技术分析层`
2. `全市场趋势型待选池层`
3. `超短线主推荐头`
4. `短波段辅助推荐头`

### 1. 大盘技术分析层

职责：

- 判断当前市场中短期状态
- 判断明日偏向 `进攻 / 观察 / 防守`
- 为后续股票筛选提供环境标签与阈值偏置

输出两个核心结论：

- `后期趋势状态`
- `次日节奏判断`

#### 后期趋势状态

建议分类为：

- `UPTREND`
- `RANGE_STRONG`
- `RANGE_NEUTRAL`
- `RANGE_WEAK`
- `RISK_OFF`

#### 次日节奏判断

建议分类为：

- `ATTACK`
- `REPAIR`
- `NEUTRAL`
- `DEFENSE`

#### 分析维度

大盘层主要使用：

- 均线结构：`MA5 / MA10 / MA20 / MA60`
- 均线斜率
- 动量：`MACD`、`RSI`
- 波动：`ATR`、历史波动率
- 成交量与成交额变化
- 市场宽度：涨跌家数、强势股数量
- 短线情绪代理：涨停数、炸板率、连板高度

### 2. 全市场趋势型待选池层

职责：

- 从 `5000+` 股票中筛出值得继续分析的股票
- 优先保留趋势向上、结构健康、可交易性好的股票
- 不直接输出正式推荐，只输出待选股池与观察标签

#### 风格定义

第一层固定采用 `趋势型` 风格。

优先保留：

- 趋势向上的股票
- 相对大盘 / 板块更强的股票
- 放量突破或缩量回踩结构健康的股票
- 均线多头或准多头结构的股票

#### 第一层不是纯技术分析

该层采用 `多分析结合`，但由技术结构主导：

- 大盘环境过滤
- 板块 / 题材强度
- 流动性与可交易性
- 风险过滤
- 日线趋势结构
- 资金行为初筛

#### 建议流程

1. `全市场基础过滤`
2. `大盘环境偏置`
3. `板块 / 题材优先级排序`
4. `趋势型技术结构打分`
5. `资金与风险修正`
6. 输出 `待选股池`

#### 待选池规模

建议分三级：

- `全市场粗筛`：`5000+ -> 100-300`
- `重点候选`：`100-300 -> 20-50`
- `观察名单`：`20-50 -> 10-20`

### 3. 超短线主推荐头

职责：

- 从趋势型待选池中选出当日最值得操作的 `1-2` 只股票
- 周期以 `1-2` 个交易日为主
- 更依赖分钟线确认和当日节奏

#### 主推荐逻辑

超短线主推荐的主导因素：

- 日线结构是否健康
- 当日分钟线承接是否成立
- 板块是否同步
- 资金是否确认

建议采用：

- `日线筛选`
- `分钟线确认`
- `综合阈值门槛`

#### 超短线推荐输出

每只主推荐必须输出：

- `推荐周期：1-2天`
- `核心技术结构`
- `分钟线确认结果`
- `关键买入位置`
- `风险提示`
- `失效条件`

### 4. 短波段辅助推荐头

职责：

- 从同一待选池中选出适合 `3-10` 个交易日跟踪和持有的股票
- 不与主推荐混规则
- 承担“更稳的第二层表达”

#### 辅推逻辑

短波段辅助位更看重：

- 日线趋势完整性
- 中继整理结构
- 平台突破前夜或趋势延续
- 板块中线持续性

分钟线只用于优化入场，不是决定性因素。

#### 辅推规模

建议每天固定输出 `3-5` 只。

#### 辅推输出

每只辅助推荐必须输出：

- `推荐周期：3-10天`
- `趋势逻辑`
- `关键支撑 / 压力位`
- `主要风险`
- `观察条件`

## First-Layer Model Design

### 1. 硬过滤

第一层先做硬过滤：

- `ST / *ST`
- 停牌或明显流动性异常
- 重大公告风险
- 过短次新
- 成交额不足
- 异常高波动 / 极端控盘

### 2. 待选池评分框架

建议第一层采用分层组合评分，而非平均因子加权。

建议权重：

- `大盘环境 25`
- `板块 / 题材强度 25`
- `中期技术结构 30`
- `资金行为 10`
- `事件 / 风险过滤 10`

#### 中期技术结构重点

重点看：

- `MA5 / MA10 / MA20 / MA60` 位置
- 趋势斜率
- 相对强弱
- 回调质量
- 量价匹配
- 平台突破 / 回踩结构

## Second-Layer Decision Split

### 1. 超短线头

建议权重：

- `技术分析 50-60`
- `分钟线确认 20-25`
- `板块共振 10-15`
- `资金确认 10`
- `风险过滤` 仅做扣分和否决

### 2. 波段头

建议权重：

- `日线趋势结构 45`
- `板块中期强度 20`
- `相对强弱 15`
- `资金稳定性 10`
- `事件 / 风险过滤 10`

## Current Model Reuse Strategy

基于当前代码，建议保留并重组现有模型，而不是重写全部逻辑。

### 建议作为核心保留

- `强势回踩`
- `均线支撑`
- `缩量止跌`
- 常规主链路中的 `趋势评分`

### 建议保留但降权

- `阳包阴`

### 建议只做辅助确认

- `资金背离`
- `资金流 / 事件 / 共振` 辅助因子

### 建议第一阶段移除或不进入主推荐

- `涨停板次日`
- `超跌反弹` 直接主推路径

## Market Analysis to Recommendation Flow

建议完整链路如下：

1. 盘后跑 `大盘日线分析`
2. 输出 `后期趋势状态 + 次日节奏判断`
3. 基于全市场数据构建 `趋势型待选股池`
4. 从待选股池生成：
   - `超短线观察名单`
   - `波段跟踪名单`
5. 次日开盘后：
   - 超短线头读取分钟线和盘中市场状态
   - 波段头只做轻量入场确认
6. 输出：
   - `超短线主推荐`
   - `短波段辅助推荐`

## Reporting Contract

### 大盘结论区

必须输出：

- 当前市场状态
- 明日节奏判断
- 建议仓位 / 风险姿态
- 不适合做的方向

### 超短线主推荐区

必须输出：

- `1-2` 只正式推荐
- 推荐理由
- 分钟线确认说明
- 风险与失效条件

### 短波段辅助区

必须输出：

- `3-5` 只辅助推荐
- 趋势逻辑
- 观察重点
- 风险提示

## Admin Integration Contract

### 1. 总体原则

现有 Admin 已经具备：

- 总览
- 运行中心
- 配置方案 / 模板
- 股票池规则 / 因子权重
- 候选与审核
- 事件 / 图谱
- 评估复盘

新方案不改模块边界，只改这些页面承接的业务含义。

### 2. 总览页职责

`StockSelectionOverview` 应该成为这套系统的每日驾驶舱，固定回答：

- 今天市场处于什么状态
- 今天更适合做超短线还是偏防守
- 最新一轮运行产出了多少 `趋势型待选股`
- 最新一轮运行产出了多少 `超短线主推荐`
- 最新一轮运行产出了多少 `短波段辅助推荐`
- 当前有哪些预警、待审核项和数据新鲜度问题

总览页要新增或强化的卡片：

- `大盘结论卡`
- `次日节奏卡`
- `主推荐 / 辅推产出卡`
- `待选池规模与压缩率`
- `分头复盘摘要`

### 3. 配置方案页职责

`Profiles / Templates / Rules / Factors` 不再只服务旧的“泛候选池 + 最终组合”模型，而是服务新分层结构。

配置必须显式拆成：

- `market_analysis_config`
- `candidate_pool_config`
- `short_term_head_config`
- `swing_head_config`
- `publish_config`

为了兼容现有模型，旧字段仍保留，但业务含义需要重映射：

- `universe_config`：股票池硬过滤与基础约束
- `seed_mining_config`：过渡期承接待选池粗筛参数
- `factor_config`：过渡期承接第一层 / 第二层权重
- `portfolio_config`：承接主推数量、辅推数量、观察池上限等输出约束

### 4. 运行中心职责

`Runs` 页面需要把一次运行明确展示为分层流水线，而不是单一候选生成作业。

建议阶段键演进为：

- `MARKET_ANALYSIS`
- `TREND_CANDIDATE_POOL`
- `SHORT_TERM_PRIMARY`
- `SWING_AUXILIARY`
- `REVIEW_PAYLOAD`
- `FORWARD_EVALUATION`

运行详情必须能看见：

- 大盘分析摘要
- 趋势型待选池压缩结果
- 主推荐与辅推各自产出
- 每一层的样本数、耗时、淘汰原因和预警

### 5. 候选与审核页职责

`Candidates / Reviews` 页面要承接 3 份不同性质的数据，而不是只看一个候选池：

- `共享待选池`
- `超短线主推荐`
- `短波段辅助推荐`

审核动作仍然走现有发布链路，但审核对象要表达清楚：

- 主推荐是核心发布决策
- 波段辅助是固定副区
- 两者共享同一次运行上下文

候选页建议至少支持以下筛选：

- 按 `stage`
- 按 `portfolio_role`
- 按 `recommendation_head`
- 按 `selected / vetoed`

### 6. 评估复盘页职责

`Evaluation` 页面必须停止把所有结果混成一个榜单口径。

评估维度至少拆成：

- `SHORT_TERM_PRIMARY`
- `SWING_AUXILIARY`
- `CANDIDATE_POOL`

评估看板需要支持：

- 同模板下主推和辅推分开看
- 同市场状态下主推和辅推分开看
- 同 profile 下不同推荐头的命中率 / 收益 / 回撤差异

### 7. 事件与图谱页职责

`Events / Graph` 在新方案里不负责最终推荐，而负责解释与辅助确认。

它们的角色是：

- 给大盘分析和板块方向提供上下文
- 给待选池和推荐头提供辅助证据
- 在审核页里解释“为什么这只票入选 / 为什么这只票被否决”

### 8. 权限与发布流

沿用现有权限模型：

- `stock_selection.view`：看总览、运行、候选、复盘
- `stock_selection.manage`：改配置、发起运行、审核发布

但审核流要更加清晰地区分：

- `策略配置动作`
- `运行触发动作`
- `推荐审核发布动作`

这样 Admin 才不会把“配模型”和“发推荐”混成同一种管理行为。

## Admin Data Contract Sketch

### 1. Overview 返回结构

`GET /admin/stock-selection/overview` 在兼容现有字段的前提下，建议新增：

```json
{
  "market_analysis": {
    "trend_state": "UPTREND",
    "next_day_rhythm": "ATTACK",
    "risk_posture": "NORMAL",
    "summary": "指数趋势健康，适合进攻型短线与趋势跟踪",
    "breadth_snapshot": {
      "advancers": 3120,
      "decliners": 1680,
      "limit_up_count": 76,
      "limit_down_count": 2,
      "broken_limit_rate": 0.18
    }
  },
  "candidate_pool_summary": {
    "universe_count": 5123,
    "coarse_pool_count": 240,
    "focus_pool_count": 36,
    "watch_pool_count": 12
  },
  "head_summary": {
    "short_term_primary_count": 2,
    "swing_auxiliary_count": 4
  },
  "evaluation_split_summary": {
    "short_term_primary": {},
    "swing_auxiliary": {}
  }
}
```

### 2. Run Detail 返回结构

`GET /admin/stock-selection/runs/:run_id` 建议在现有 `stage_logs` 和 `context_meta` 基础上，固定暴露：

```json
{
  "run_id": "ssr_xxx",
  "market_regime": "UPTREND",
  "context_meta": {
    "market_analysis": {},
    "candidate_pool_summary": {},
    "head_output_summary": {}
  },
  "stage_logs": [
    { "stage_key": "MARKET_ANALYSIS" },
    { "stage_key": "TREND_CANDIDATE_POOL" },
    { "stage_key": "SHORT_TERM_PRIMARY" },
    { "stage_key": "SWING_AUXILIARY" }
  ]
}
```

关键是让 Runs 页不需要自己猜“当前阶段代表什么”。

### 3. Candidate Snapshot 返回结构

`GET /admin/stock-selection/runs/:run_id/candidates` 建议补充这些维度：

- `recommendation_head`
  可选值：
  - `SHARED_CANDIDATE_POOL`
  - `SHORT_TERM_PRIMARY`
  - `SWING_AUXILIARY`
- `selection_layer`
  可选值：
  - `L1_CANDIDATE_POOL`
  - `L2_PRIMARY`
  - `L2_AUXILIARY`
- `reason_tags`
- `veto_tags`
- `technical_pattern`
- `sector_strength_label`
- `market_rhythm_match`

这样候选页才能一眼分清：

- 这只股票只是进入了共享待选池
- 还是已经晋级成了超短线主推
- 还是进入了波段辅助区

### 4. Evaluation 返回结构

`GET /admin/stock-selection/runs/:run_id/evaluation` 与 leaderboard 都建议显式带上：

- `evaluation_scope`
  可选值：
  - `SHORT_TERM_PRIMARY`
  - `SWING_AUXILIARY`
  - `CANDIDATE_POOL`
- `head_label`
- `holding_contract`

这样复盘页就能按推荐头拆开，而不是继续把所有样本揉成一个统计口径。

### 5. Profile / Template 配置结构

在保持旧字段兼容的前提下，建议 profile/template 逐步过渡到下面这组 bucket：

```json
{
  "market_analysis_config": {
    "breadth_weight": 25,
    "momentum_weight": 25,
    "volume_weight": 20,
    "sentiment_weight": 30
  },
  "candidate_pool_config": {
    "trend_weight": 30,
    "sector_weight": 25,
    "market_weight": 25,
    "flow_weight": 10,
    "risk_weight": 10,
    "coarse_pool_limit": 300,
    "focus_pool_limit": 50,
    "watch_pool_limit": 20
  },
  "short_term_head_config": {
    "max_picks": 2,
    "daily_line_weight": 55,
    "minute_line_weight": 25,
    "sector_resonance_weight": 10,
    "capital_confirmation_weight": 10
  },
  "swing_head_config": {
    "max_picks": 5,
    "trend_weight": 45,
    "sector_weight": 20,
    "relative_strength_weight": 15,
    "stability_weight": 10,
    "risk_weight": 10
  }
}
```

第一阶段可以先把这些 bucket 挂在现有 profile/template JSON 下，不要求一开始就彻底废弃旧字段。

## Admin Page Card Checklist

### 1. 总览页

建议固定 6 张卡：

- `市场状态卡`
- `次日节奏卡`
- `待选池压缩卡`
- `主推产出卡`
- `辅推产出卡`
- `复盘摘要卡`

### 2. 配置方案页

建议固定 5 个配置分组：

- `股票池边界`
- `大盘分析参数`
- `待选池参数`
- `超短线主推参数`
- `短波段辅推参数`

### 3. 运行中心页

建议详情页固定 4 个页签或区块：

- `市场分析`
- `待选池流转`
- `主推 / 辅推结果`
- `阶段日志与预警`

### 4. 候选与审核页

建议固定 3 个列表切换：

- `共享待选池`
- `超短线主推`
- `短波段辅推`

### 5. 评估复盘页

建议固定 3 个维度切换：

- 按 `模板 / 配置`
- 按 `市场状态`
- 按 `推荐头`

## Non-Goals

第一阶段不做：

- 创业板 / 全市场不同涨跌幅制度统一建模
- 高频盘口超短择时
- 复杂机器学习黑箱买点模型
- 把超短线和短波段混成同一个推荐信号
- 纯事件驱动或纯接力模型作为主链

## First-Phase Success Criteria

第一阶段成功标准不是“覆盖所有机会”，而是：

- 大盘判断能稳定约束推荐节奏
- `5000+ -> 待选池` 压缩链路稳定
- 超短线主推荐与波段辅推分工清晰
- 推荐逻辑能解释为什么入选、为什么没入选
- 技术分析真正成为最终推荐层的主引擎
