## 背景

当前 `newclient` 的 [\`/recommendations/strategies\`](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/index.js:10) 在产品叙事上属于“每日推荐”的第二步，文案表达是“第 2 步：形成策略”。但实际页面 [Strategies.vue](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/Strategies.vue) 加载的是 [listFuturesStrategies](/Users/gjhan21/cursor/sercherai/newclient/src/api/market.js:44)，后端来源是 [ListFuturesStrategies](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go:4960) 查询 `futures_strategies` 表。

这导致当前模块存在根本性错位：

- 产品主线表达的是“股票推荐 -> 形成策略 -> 深度推演”
- 实际数据对象却是“期货策略发布记录”
- 从股票推荐页带过来的 `symbol/name` 只用于顶部说明文案，不参与策略筛选
- 同一页内存在两套深度推演目标语义：
  - 顶部流程按钮按 `STOCK` 承接
  - 策略卡片按钮按 `FUTURES` 承接
- 策略卡片“查看回测”当前仍然跳向 [BacktestView.vue](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/BacktestView.vue)，而该页面仍消费 [mock 回测数据](/Users/gjhan21/cursor/sercherai/newclient/src/mock/recommendations.js:39)
- 页面接口失败时会静默回退到 [mock 策略](/Users/gjhan21/cursor/sercherai/newclient/src/mock/recommendations.js:33)，掩盖真实业务状态

需要一次正式重构，把“推荐主线策略承接层”和“期货策略中心”拆开，避免继续用文案强行串联两条不同业务线。

## 问题定义

当前实现的核心问题不是样式或单点 bug，而是业务身份混淆：

1. `\`/recommendations/strategies\`` 被放在股票推荐主线中，但实际承载期货策略列表
2. 推荐上下文没有进入策略筛选，页面无法真正承接某条股票推荐
3. 期货策略列表、期货策略详情、期货深度推演链路本身是真实存在的，但被错误地嵌入股票推荐主线
4. mock 回退和静态回测使页面在真实数据缺失时看起来“像有功能”，削弱可信度
5. 深度推演入口上下文不统一，用户会在同一页中遭遇 `STOCK` 与 `FUTURES` 两套目标含义

## 目标

本次重构的目标是：

1. 让股票推荐主线变成真正闭环：
   - `/recommendations`
   - `/recommendations/history`
   - `/recommendations/strategies`
   - `/forecast-lab`
   - `/forecast/:id`

2. 让期货策略主线独立成立：
   - `/futures/strategies`
   - `/futures/strategy/:id`
   - `/forecast-lab`
   - `/forecast/:id`

3. 保持深度推演作为共用研究层，但在进入前统一上下文语义

4. 第一阶段优先复用现有后端能力，不先新增大规模聚合接口

## 非目标

本次重构不包含：

- 新增股票策略回测引擎
- 新增期货策略回测引擎
- 新增个性化策略推荐排序
- 重写深度推演报告结构
- 新增策略引擎产物 schema
- H5 端同步重构整套股票/期货策略导航结构

H5 端若受 PC 路由或上下文规则影响，仅做必要兼容，不在本 spec 内扩展为完整对等改造。

## 最终产品结构

### 1. 股票推荐主线

#### 页面职责

- [\`/recommendations\`](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/DailyRecs.vue)
  负责“发现机会”
- [\`/recommendations/history\`](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/HistoryPerf.vue)
  负责“回看历史真实绩效”
- `\`/recommendations/strategies\``
  负责“把当前推荐机会收敛成执行策略”
- [\`/forecast-lab\`](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastLabView.vue)
  负责“进入研究中心”
- [\`/forecast/:id\`](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue)
  负责“阅读深度推演报告”

#### 核心语义

`/recommendations/strategies` 不再是策略池列表，而是**当前推荐标的的策略承接页**。

它不负责“给用户看还有哪些别的策略”，而负责：

- 解释这条推荐为什么值得执行
- 把推荐 detail / insight 转成用户可执行策略
- 给出下一步动作：回推荐、去股票分析、进深度推演

### 2. 期货策略主线

#### 页面职责

- `\`/futures/strategies\``
  负责“期货策略列表”
- [\`/futures/strategy/:id\`](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue)
  负责“期货策略详情、绩效、解释、深度推演入口”

#### 核心语义

这条线保留现有 `futures_strategies` 能力，不再混入股票推荐主线。  
它的定位是“期货策略中心”，不是“股票推荐的第二步”。

## 页面设计

### A. 新的 `/recommendations/strategies`

该页不再加载 [listFuturesStrategies](/Users/gjhan21/cursor/sercherai/newclient/src/api/market.js:44)，而改为围绕当前股票推荐上下文组织内容。

#### 入口要求

必须从推荐主线带入完整上下文，至少包括：

- `reco_id`
- `symbol`
- `name`

如果缺失 `reco_id`：

- 页面不请求股票策略数据
- 页面显示空态：
  - `请先从每日推荐进入策略承接页`
- 提供返回 [\`/recommendations\`](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/DailyRecs.vue) 的动作

#### 页面区块

1. `承接头部`
   - 当前推荐标的
   - 来源：每日推荐
   - 步骤条：`每日推荐 -> 执行策略 -> 深度推演`

2. `策略结论卡`
   - 推荐评分
   - 风险等级
   - 建议仓位
   - 止盈
   - 止损
   - 风险备注

3. `执行计划卡`
   - 当前更适合：等待确认 / 分批试探 / 趋势跟随 / 仅观察
   - 策略理由摘要
   - 关键观察信号

4. `证据维度卡`
   - 技术面
   - 基本面代理
   - 资金面
   - 情绪面

5. `动作区`
   - 回每日推荐
   - 去股票分析
   - 进入深度推演

#### 数据来源

第一阶段不新增后端聚合接口，直接复用：

- [GetStockRecommendationDetail](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go:4376)
- [GetStockRecommendationInsight](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go:4438)

以 recommendation detail / insight 作为“策略承接层”的真实数据源。

### B. 新的 `/futures/strategies`

当前 [Strategies.vue](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/Strategies.vue) 的期货策略列表逻辑迁移到该页面。

#### 页面调整

- 去掉股票推荐主线文案：
  - `第 2 步：形成策略`
  - `回看推荐来源`
- 改成期货策略中心文案：
  - `期货策略`
  - `查看最新发布与进行中的期货策略`
- 保留：
  - 策略卡片
  - 风险 / 仓位 / 理由
  - 进入深度推演
- “查看回测”从真实卡片中移除，直到真实回测链路上线

### C. `FuturesStrategyDetail`

保留其现有角色，但至少修正：

1. 返回路径从错误的 `/strategies` 改到 `/futures/strategies`
2. 明确 performance 的来源：
   - 来自 `futures_reviews`
   - 或 fallback 估算
3. 深度推演入口继续保留，但目标上下文始终为 `FUTURES`

## 上下文与深度推演规则

### 股票策略承接页

“进入深度推演”统一构造：

- `targetType = STOCK`
- `targetId = reco_id`
- `targetKey = symbol`
- `targetLabel = name`
- `source = STRATEGY`
- `sourceId = reco_id`
- `sourcePath = /recommendations/strategies`

### 期货策略中心

“进入深度推演”统一构造：

- `targetType = FUTURES`
- `targetId = strategy.id`
- `targetKey = contract`
- `targetLabel = strategy.name`
- `source = STRATEGY`
- `sourceId = strategy.id`
- `sourcePath = /futures/strategies` 或 `/futures/strategy/:id`

### 明确禁止

禁止在同一页面中同时出现：

- 顶部入口按 `STOCK` 承接
- 卡片入口按 `FUTURES` 承接

也就是说，页面语义必须单一。

## Mock 与回退策略

### 前端 mock

以下 mock 不应继续出现在真实用户链路中：

- [newclient/src/mock/recommendations.js](/Users/gjhan21/cursor/sercherai/newclient/src/mock/recommendations.js) 的 `STRATEGIES`
- [BacktestView.vue](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/BacktestView.vue) 的静态回测数据

#### 规则

- 接口失败：显示错误或空态
- 未登录：显示登录门槛
- 无上下文：显示承接空态
- 不再默默显示 mock 数据

### 后端 fallback

后端 fallback 允许继续存在，例如 [AdminGenerateDailyFuturesStrategies](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go:6835) 的样本策略插入链。  
但它只属于后台生成兜底，不应通过前端 mock 回退进一步放大误导。

## 回测模块处理

当前 [BacktestView.vue](/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/BacktestView.vue) 仍然消费静态 [BACKTEST_RESULTS](/Users/gjhan21/cursor/sercherai/newclient/src/mock/recommendations.js:39)。

### 第一阶段策略

- 从真实策略卡片中移除“查看回测”
- 或将入口替换为中性提示：
  - `回测能力建设中`

### 第二阶段策略

未来若上线真实股票/期货策略回测链路，再分别接入，不在本次重构中处理。

## 实施分期

### 第一阶段：业务身份纠偏

目标：先把错位感和错误入口修掉。

内容：

- 新增 `/futures/strategies`
- 迁移当前期货策略列表逻辑
- 修正 `FuturesStrategyDetail` 返回路径
- `/recommendations/strategies` 改成上下文门槛页
- 移除 mock 回退
- 移除真实策略卡片上的静态回测入口

### 第二阶段：股票策略承接页上线

目标：让股票推荐主线闭环。

内容：

- 基于 recommendation detail / insight 实现策略结论卡
- 实现执行计划卡与证据维度卡
- 对接股票分析与深度推演动作

### 第三阶段：真实感增强

目标：提升策略中心可信度。

内容：

- 在期货详情页标明真实与估算绩效来源
- 优化股票策略承接页中的执行建议表达
- 后续条件成熟后再接入真实回测

## 测试策略

### 前端测试

必须新增或调整测试以保证：

1. `/recommendations/strategies` 缺少 `reco_id` 时显示承接空态
2. `/recommendations/strategies` 不再调用 `listFuturesStrategies`
3. `/futures/strategies` 才调用 `listFuturesStrategies`
4. 股票策略承接页中的深度推演入口只构造 `STOCK` 上下文
5. 期货策略页中的深度推演入口只构造 `FUTURES` 上下文
6. `FuturesStrategyDetail` 返回路径为 `/futures/strategies`
7. 真实策略页不再渲染 mock 回测入口

### 后端测试

第一阶段以回归为主，不新增大规模新接口。  
如果后续新增股票策略桥接聚合接口，再单独补 handler/repo tests。

## 方案取舍

曾考虑过两种替代方案：

### 方案 A：继续沿用当前页面，只修 bug

优点：

- 改动最小

缺点：

- 页面仍会是“股票主线里装期货策略”
- 业务身份问题不解决

### 方案 B：彻底把当前页定义成期货策略中心

优点：

- 页面本身更自洽

缺点：

- 股票推荐主线失去“策略承接”这一层

### 最终选择：拆成双主线

即本 spec 方案：

- 股票推荐主线拥有自己的策略承接页
- 期货策略拥有独立策略中心
- 深度推演作为共用研究层

这是当前代码库下改动成本最低、业务定义最清晰、后续扩展最稳的方案。
