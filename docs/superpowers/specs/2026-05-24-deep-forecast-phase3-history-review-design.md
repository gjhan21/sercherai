# 深度推演三期设计规范：同标的历史对比与完整复盘评分

## Summary

深度推演一期完成了“研究入口 + 结构化报告”的基础闭环，二期完成了“证据层增强 + 研究中心入口化”。三期不再继续扩数据源，也不先做跨市场全局驾驶舱，而是聚焦一个更直接、更高价值的能力：

- 让用户能够在同一只股票或同一期货标的上，查看多次成功深推演的历史演化
- 让用户能够默认对比“最新一次 vs 上一次”研究结论
- 让每次成功 run 都拥有完整、可解释的复盘评分，而不只是后台学习记录

三期的产品主题锁定为：

`同标的成功 run 历史 + 默认最新 vs 上一次对比 + 完整复盘评分`

这是一条典型的“学习闭环用户可见化”路径。它不追求在第一版就做全市场命中率面板，也不把失败 run、排队 run 混入核心阅读体验，而是先把“这只标的过去怎么判断、后来对没对、现在结论是不是更可靠”讲清楚。

## Background

当前仓库已经具备三期的关键基础设施：

- 深度推演用户侧入口和报告页已经存在：
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastLabView.vue`
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastLabView.vue`
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue`
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue`
- 后端已经具备结构化报告、严格上下文、股票/期货五维证据输出、LLM 复核降级链路：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`
- 学习记录与质量回写已经存在：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning_test.go`

当前 learning 侧已经能写入这些后验字段：

- `scenario_hit`
- `trigger_hit`
- `invalidation_early`
- `bias_label`
- `role_effectiveness`
- `summary`

同时，质量汇总已经支持按 `target_type` 聚合。但现状仍然有三个明显缺口：

1. 用户侧没有“同标的历史对比”入口
2. 用户侧没有“默认最新 vs 上一次”的差异视图
3. learning record 还是后台数据形态，尚未投影成用户可读的完整复盘评分

## Goals

三期目标是：

1. 为同一标的提供“全部成功 run 可查”的历史时间轴
2. 在报告页内默认展示“最新一次 vs 上一次”的结论变化与证据变化
3. 为每次成功 run 生成可解释的完整复盘评分
4. 让 `forecast-lab` 成为同标的历史研究入口，而不是只承接单次 run
5. 保持股票与期货共享同一历史对比骨架，只在证据维度文案上分叉

## Non-Goals

本期明确不做：

- 不做跨标的全局复盘驾驶舱
- 不做全市场命中率排行榜
- 不把失败 run / 排队 run 混入主历史视图
- 不做任意多 run 矩阵式复杂对比
- 不做新的外部行情、财报或期货结构化数据源接入
- 不改变一期、二期已经确定的 VIP 正文权限规则

## User Decisions Locked In

用户已明确确认以下三期边界：

- 三期优先主题：`同标的多次推演对比`
- 历史入口位置：`forecast-lab + report 页内都要`
- 历史范围：`只看已完成成功 run`
- 首屏对比顺序：`先结论变化，再证据变化`
- 历史数量：`全部成功 run 可查，但首屏先给摘要时间轴`
- 默认比较基准：`最新一次 vs 上一次`
- 第一版就带：`完整复盘评分`

## Product Contract

### 1. Forecast Lab 的三期定位

三期之后，`Forecast Lab` 不只是“研究中心入口”，还承担“同标的历史研究档案入口”的角色。

对聚焦标的，页面需要回答：

- 这只标的有多少次成功深推演
- 最近一次研究结论是什么
- 最近一次和上一次相比，结论是加强、减弱还是翻转
- 最近几次完整复盘评分是改善、恶化还是波动
- 是否值得直接进入历史对比

这意味着三期的 `forecast-lab` 至少新增一块：

- `历史研究概览卡`

它不展示全文，而是展示摘要与趋势，并提供 CTA：

- `查看同标的历史对比`

### 2. Forecast Report 的三期定位

三期之后，`Forecast Report` 从“单次研究报告页”升级为“单次报告 + 历史对比主视图”。

进入 `/forecast/:id` 后，用户不需要退回 `forecast-lab` 才能看历史。

页内需要新增：

- `同标的历史切换条`
- `成功 run 时间轴`
- `默认最新一次 vs 上一次` 对比区
- `完整复盘评分卡`

### 3. 历史对比的主阅读顺序

三期默认阅读顺序固定为：

1. `结论变化`
2. `证据变化`
3. `完整复盘评分`
4. `全部成功 run 时间轴`

不采用“两份原始报告全文直接并排”的技术型展示方式。

系统必须优先告诉用户：

- 结论变了什么
- 为什么变
- 后来验证得怎么样

## UX Design

### 1. Forecast Lab 入口结构

在现有“研究聚焦 / 最近研究结论 / 运行动态”的基础上，三期新增：

- `历史研究概览`

对聚焦标的，显示：

- `成功 run 数量`
- `最近结论等级`
- `最新 vs 上一次 结论变化`
- `最近复盘评分趋势`

如果该标的历史成功 run 只有 1 条，则展示：

- `当前仅有 1 次成功深推演，暂不能形成历史对比`

### 2. Forecast Report 顶部历史切换条

在报告页 hero 区附近新增：

- 当前 run 在成功历史中的位置
  - 例如：`第 5 次成功推演 / 最近一次`
- 切换动作：
  - `上一条`
  - `下一条`
  - `查看全部`
- 默认提示：
  - `当前默认对比：最新一次 vs 上一次`

### 3. 历史主视图

默认主视图分两层：

#### 3.1 结论变化卡

对比以下差异：

- `headline verdict`
- `primary scenario`
- `risk boundary`
- `action guidance`
- `review grade / review score`

输出必须是“差异摘要”，不是简单原文拼接。

#### 3.2 证据变化区

按标的类型展示：

- 股票：
  - `基本面`
  - `技术面`
  - `资金面`
  - `估值面`
  - `事件面`
- 期货：
  - `供需库存`
  - `期限结构`
  - `盘面技术`
  - `持仓资金`
  - `宏观事件`

每个维度都要先给变化类型：

- `增强`
- `减弱`
- `翻转`
- `基本不变`

再允许用户下钻看两次 run 的具体差异。

### 4. 全部成功 run 时间轴

时间轴是三期的“全量浏览模式”，而不是表格。

每条成功 run 至少显示：

- 时间
- 核心判断
- 主情景
- 复盘等级
- 复盘分数
- 是否命中主情景
- 是否提前失效
- `查看该次报告`
- `设为对比对象`

### 5. H5 与 PC 差异

PC：

- 完整展示结论变化、证据变化、评分卡和时间轴

H5：

- 首屏优先展示：
  - 结论变化
  - 简版复盘评分
  - 历史时间轴入口
- 证据变化默认折叠展开

## Data Model Design

三期不建议把历史与复盘直接散塞进现有 `StrategyForecastL3Report`，而是新增独立的历史/复盘对象。

### 1. 单次 run 复盘对象

新增：

```go
type StrategyForecastL3RunReview struct {
    RunID              string             `json:"run_id"`
    TargetType         string             `json:"target_type"`
    TargetKey          string             `json:"target_key"`
    ReviewScore        int                `json:"review_score"`
    ReviewGrade        string             `json:"review_grade"`
    ReviewVerdict      string             `json:"review_verdict"`
    ScenarioHit        bool               `json:"scenario_hit"`
    TriggerHit         bool               `json:"trigger_hit"`
    InvalidationEarly  bool               `json:"invalidation_early"`
    BiasLabel          string             `json:"bias_label,omitempty"`
    RoleEffectiveness  map[string]float64 `json:"role_effectiveness,omitempty"`
    ReviewNotes        []string           `json:"review_notes,omitempty"`
    ReviewedAt         string             `json:"reviewed_at,omitempty"`
}
```

### 2. 历史时间轴项

新增：

```go
type StrategyForecastL3HistoryItem struct {
    RunID            string `json:"run_id"`
    TargetType       string `json:"target_type"`
    TargetKey        string `json:"target_key"`
    TargetLabel      string `json:"target_label"`
    GeneratedAt      string `json:"generated_at,omitempty"`
    HeadlineVerdict  string `json:"headline_verdict,omitempty"`
    PrimaryScenario  string `json:"primary_scenario,omitempty"`
    ActionGuidance   string `json:"action_guidance,omitempty"`
    ReviewScore      int    `json:"review_score,omitempty"`
    ReviewGrade      string `json:"review_grade,omitempty"`
    ReviewVerdict    string `json:"review_verdict,omitempty"`
}
```

### 3. 历史对比结果

新增：

```go
type StrategyForecastL3HistoryCompare struct {
    LeftRun       *StrategyForecastL3HistoryItem `json:"left_run,omitempty"`
    RightRun      *StrategyForecastL3HistoryItem `json:"right_run,omitempty"`
    VerdictShift  []string                       `json:"verdict_shift,omitempty"`
    EvidenceShift []StrategyForecastL3EvidenceDiff `json:"evidence_shift,omitempty"`
    ReviewShift   []string                       `json:"review_shift,omitempty"`
}
```

`StrategyForecastL3EvidenceDiff` 至少包含：

- `dimension`
- `change_type`
- `left_summary`
- `right_summary`

## Review Score Design

三期第一版采用“可解释规则评分”，不使用复杂黑箱模型。

### 1. 对外展示形式

用户看到的是两层：

1. `等级`
   - `A`
   - `B`
   - `C`
   - `D`
2. `分数`
   - 用于趋势和辅助排序，例如：
     - `85`
     - `70`
     - `55`
     - `35`

### 2. 第一版评分映射

建议规则：

- `scenario_hit && trigger_hit && !invalidation_early`
  - `A / 85`
  - `判断有效，执行窗口清晰`
- `scenario_hit && !trigger_hit`
  - `B / 70`
  - `方向正确，但确认不足`
- `!scenario_hit && !invalidation_early`
  - `C / 55`
  - `主情景未完全走出，需复核证据链`
- `invalidation_early`
  - `D / 35`
  - `风险边界过早触发，当前判断失效`

### 3. Bias Label 映射

`bias_label` 不直接主导总分，但决定复盘语义说明，例如：

- `RISK_FIRST`
  - 风险边界识别偏早或市场先进入风险态
- `OVERCONFIDENT`
  - 证据不足但结论过强
- `UNDERCONFIRMED`
  - 方向可能正确，但触发条件不充分
- `LATE_CONFIRMATION`
  - 确认滞后于市场演化

## API Design

三期新增用户侧接口，不直接复用管理端 quality 接口。

### 1. 同标的成功历史

`GET /api/v1/forecast/targets/history`

参数：

- `target_type`
- `target_key`
- `page`
- `page_size`

返回：

- 只包含 `SUCCEEDED`
- 默认按时间倒序
- 包含 `StrategyForecastL3HistoryItem[]`

### 2. 默认历史对比

`GET /api/v1/forecast/targets/history/compare`

参数：

- `target_type`
- `target_key`
- `left_run_id` 可选
- `right_run_id` 可选

默认行为：

- 未显式传入时，自动使用 `latest vs previous`

返回：

- `StrategyForecastL3HistoryCompare`

### 3. 单次 run 复盘评分

`GET /api/v1/forecast/runs/:id/review`

返回：

- `StrategyForecastL3RunReview`

## Backend Architecture

### 1. Learning Record 升级

当前 `buildStrategyForecastL3LearningRecord(...)` 仍偏占位。三期不废弃它，而是在其上增加投影：

- `raw learning signals`
- `review projection`

即：

1. 先生成底层后验信号：
   - `scenario_hit`
   - `trigger_hit`
   - `invalidation_early`
   - `role_effectiveness`
2. 再从这些信号生成：
   - `review_score`
   - `review_grade`
   - `review_verdict`
   - `review_notes`

### 2. Repo 层新增职责

在 `strategy_forecast_l3_repo.go` 中新增：

- 按 `target_type + target_key` 获取成功历史
- 构建默认 `latest vs previous`
- 生成 `history compare`
- 生成 `run review`

### 3. 权限与可见性

三期不改变正文 VIP 规则：

- 非 VIP 仍可查看：
  - 历史时间轴摘要
  - 结论变化摘要
  - 复盘评分
- 非 VIP 不能借历史接口绕过正文权限

## Frontend Architecture

### 1. Forecast Lab

在现有研究中心基础上新增：

- `历史研究概览卡`
- `最近几次成功 run 变化摘要`
- `查看完整历史对比`

### 2. Forecast Report

在现有结构化报告页基础上新增：

- `同标的历史切换条`
- `结论变化卡`
- `证据变化区`
- `完整复盘评分卡`
- `成功 run 时间轴`

### 3. Shared View Model

建议新增：

- `forecast-history-view-model.js`

负责：

- 默认 pair 选择
- 时间轴排序
- 结论变化摘要
- 证据变化标签
- 评分趋势映射

## Testing Strategy

### 后端

补四类测试：

1. `Learning -> Review`
   - 不同 `scenario_hit / trigger_hit / invalidation_early` 组合映射到正确 `review_grade / review_score / review_verdict`
2. `History List`
   - 只返回成功 run
   - 只按同标的聚合
3. `History Compare`
   - 默认 `latest vs previous`
   - 只有 1 条成功 run 时稳定降级
4. `User Visibility`
   - 非 VIP 仍可看历史与评分，但正文权限不被绕过

### 前端

补四类测试：

1. `Forecast Detail`
   - 包含 `同标的历史`
   - 包含 `最新一次 vs 上一次`
   - 包含 `完整复盘评分`
2. `Forecast Lab`
   - 包含 `历史研究概览`
   - 包含 `查看完整历史对比`
3. `History View Model`
   - 默认 pair
   - 时间轴排序
   - 差异标签生成
4. `Stock + Futures`
   - 股票和期货都能在同一骨架下渲染历史对比，只分叉维度文案

## Risks

### 1. 后验评分过于武断

应对：

- 第一版只做规则评分
- 分数必须带解释，不允许黑箱

### 2. 历史对比信息过载

应对：

- 默认只比较 `最新 vs 上一次`
- 全量时间轴放在下层
- H5 默认折叠证据变化

### 3. 历史接口绕过正文权限

应对：

- 历史接口只返回摘要与评分
- 不返回可绕过权限的正文全文

## Success Criteria

三期第一版完成后，用户应该能做到：

1. 在 `forecast-lab` 里发现一只标的的历史研究入口
2. 在 `forecast/:id` 里直接看到“最新一次 vs 上一次”结论变化
3. 下钻查看全部成功 run 时间轴
4. 理解每次成功 run 的完整复盘评分，而不是只看到后台布尔字段
5. 在股票和期货上都获得一致的历史对比体验

## Out of Scope For This Phase

这些能力留给后续迭代，不属于三期第一版：

- 跨标的全局复盘驾驶舱
- 全市场命中率与排行
- 失败 run 历史对比
- 复杂图表与可视化终端
- 自动学习调参与模型自适应闭环
