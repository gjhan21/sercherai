# 每日推荐优化设计：今日 AI 精选可信化与卡片增强

## Summary

`/recommendations` 的“今日 AI 精选”已经具备真实的后端推荐产出链路，但当前用户体验仍停留在“可演示”状态，离“可信的每日精选推荐产品”还有明显差距。

当前系统的问题集中在四点：

1. “今日”语义不严格  
   用户页默认不传 `trade_date`，后端也只按 `valid_from DESC` 取最近记录，导致用户看到的更像“最近可展示推荐”，而不是“今日推荐”。

2. 列表卡片存在假数据  
   前端当前用 `score` 和 `Math.random()` 拼出 `price/change`，并把 `entry/stopLoss/takeProfit` 统一写成“待确认”，这会直接伤害模块可信度。

3. 列表结果过于扁平  
   后端实际上已经有 `risk_level / position_range / take_profit / stop_loss / tech_score / fund_score / sentiment_score / money_flow_score` 等信息，但列表页只消费了很薄的一层。

4. 未登录态混用了 mock 推荐  
   当前未登录直接回退到本地 `MOCK_RECS`，用户很难分辨“真实推荐”和“演示数据”。

本次优化不改底层推荐算法，不进入个性化排序，不接新外部数据源，而是聚焦第一阶段高价值修正：

- 把“今日 AI 精选”从“演示列表”升级为“可信的每日推荐入口”
- 统一前后端对“今日”的语义
- 清理前端假行情和假涨跌幅
- 让列表卡片尽可能展示真实存在的推荐字段

## Background

当前推荐链路已经是完整的：

- 前端页面在 `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/DailyRecs.vue`
- 前端接口在 `/Users/gjhan21/cursor/sercherai/newclient/src/api/market.js`
- 用户接口在 `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/user_growth_handler.go`
- 查询落在 `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go` 的 `ListStockRecommendations`
- 真实推荐记录存放在 `stock_recommendations` 和 `stock_reco_details`
- 推荐生成来源有两条：
  - 策略引擎 `stock-selection` 正式链
  - 本地量化 fallback 链

现有核心行为：

- 已登录用户看到的股票列表来自 `GET /api/v1/stocks/recommendations`
- 后端仅返回 `PUBLISHED / ACTIVE / TRACKING` 的推荐记录
- 列表页只消费 `symbol / name / score / reason_summary / position_range`
- 价格、涨跌幅、具体交易位不是后端返回，而是前端临时拼装

因此，这块模块已经有“真实推荐池”，但仍缺“真实推荐展示层”。

## Goals

本次优化的目标是：

1. 让“今日 AI 精选”严格对应“某个交易日的有效推荐集合”
2. 列表卡片不再展示假价格、假涨跌幅、假交易位
3. 列表卡片优先展示真实存在的推荐字段
4. 未登录态不再伪装成真实推荐池
5. 保持当前后端推荐生成链和深度推演承接链不被破坏

## Non-Goals

本次明确不做：

- 不改策略引擎 `stock-selection` 算法
- 不重写量化评分模型
- 不做个性化推荐排序
- 不做跨用户的风险偏好适配
- 不引入新的实时行情服务
- 不新增复杂图表、回测面板或多级筛选器

## Product Contract

### 1. “今日”语义

用户侧“今日 AI 精选”必须绑定一个明确的交易日。

第一阶段采用以下收口规则：

- 前端列表请求默认带当天日期 `trade_date`
- 后端按 `trade_date` 优先过滤
- 若无 `trade_date`，后端回退为“当前最新有效交易日”

“有效推荐”的定义为：

- 状态属于 `PUBLISHED / ACTIVE / TRACKING`
- 且推荐属于目标交易日

本期不额外引入新的复杂有效期语义，只在现有 `valid_from / valid_to` 之上维持兼容。

### 2. 列表卡片展示原则

列表卡片只允许展示真实字段或明确标记为缺省状态的字段，不允许再出现：

- 随机 `price`
- 随机 `change`
- 从 `score` 伪装出来的价格

卡片的基础字段改为：

- 股票名称
- 股票代码
- 推荐评分
- 风险等级
- 建议仓位
- AI 推荐理由摘要

如果后端能提供明细信息，则卡片再展示：

- 止盈
- 止损
- 可选：简化四维评分摘要

如果某字段不存在，允许显示空态文案，但不得显示伪造数据。

### 3. 未登录态

未登录态不再展示和已登录态同构的 mock 推荐池。

改为：

- 保留页面骨架
- 模块文案改成“登录后查看今日 AI 精选”
- 可选展示 1-2 条静态示意卡，但必须视觉上明确标注“示意”

本期推荐直接采用更简单、更安全的方案：

- 未登录态不展示 mock 股票池
- 只展示引导登录的模块空态

### 4. 推荐详情承接

“进入深度推演”和“查看对应策略”的上下文链不变。

本次优化只修列表数据可信度，不改变：

- `/recommendations -> /forecast-lab`
- `/recommendations -> /recommendations/strategies`

的承接路径。

## API / Data Design

### 1. 用户列表接口扩展

当前接口：

- `GET /api/v1/stocks/recommendations`

保留接口路径，但增强语义和字段。

请求参数：

- `trade_date`：前端默认传入，格式 `YYYY-MM-DD`
- `page`
- `page_size`

响应项在现有基础上建议补充：

- `source_type`
- `strategy_version`
- `performance_label`
- `take_profit`
- `stop_loss`

第一阶段不强制返回实时行情字段，因为当前后端链路并没有稳定的“推荐时行情快照”契约。

### 2. Repo 查询规则

`ListStockRecommendations` 从“只按状态 + valid_from 倒序”调整为：

- 按 `trade_date` 精确收口
- 默认只返回当天集合
- 同一天内优先按 `score DESC`
- 再按 `created_at DESC`

也就是说，“今天的 AI 精选”更像“今日推荐榜”，而不是“最近插入的几条推荐记录”。

### 3. Detail 数据复用

当前 `stock_reco_details` 已经存在：

- `tech_score`
- `fund_score`
- `sentiment_score`
- `money_flow_score`
- `take_profit`
- `stop_loss`
- `risk_note`

第一阶段不单独新增新表，而是直接复用这张表，为列表卡片补充：

- `take_profit`
- `stop_loss`

是否同时补四维评分，视实现复杂度而定：

- 首选：列表不展示四维分数，只展示止盈止损和风险等级
- 可选：后续小步增量再加四维摘要

## UX Design

### 1. PC 列表卡片

PC 端“今日 AI 精选”卡片改成：

- 顶部：名称、代码、评分
- 中部：风险等级、建议仓位
- 主体：AI 推荐理由摘要
- 底部：
  - 止盈
  - 止损
  - CTA：`查看对应策略`
  - CTA：`进入深度推演`

不再显示：

- 假价格
- 假涨跌幅

### 2. H5 一致性

本次任务主要作用于 `/recommendations` 的 PC 页面，但推荐数据契约属于通用能力。  
如果 H5 未来接同一数据接口，应自动获得同样的“真实字段优先”收益。

本期不额外新增 H5 的推荐专页重构。

### 3. 空态

未登录或接口无数据时，显示清晰空态：

- 未登录：`登录后查看今日 AI 精选`
- 无数据：`今日暂无可发布推荐`

不再使用 mock 列表掩盖真实状态。

## Implementation Outline

### 后端

涉及文件：

- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/user_growth_handler.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go`

主要改动：

1. 扩展 `StockRecommendation` 或新增更适合列表的返回字段
2. 调整 `ListStockRecommendations` 的筛选与排序
3. 将 `stock_reco_details` 中的止盈止损拼回列表项
4. 兼容 `trade_date` 默认策略

### 前端

涉及文件：

- `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/DailyRecs.vue`
- `/Users/gjhan21/cursor/sercherai/newclient/src/api/market.js`
- 相关推荐页源码契约测试文件

主要改动：

1. `loadDailyRecs()` 默认带 `trade_date`
2. 删除随机 `price/change`
3. 删除 mock 推荐池作为未登录兜底的行为
4. 卡片改展示真实字段
5. 保持既有 CTA 链路不变

## Testing Strategy

### 前端

新增或增强测试，至少覆盖：

1. 推荐页请求携带 `trade_date`
2. 推荐卡片不再包含随机 `price/change`
3. 未登录态显示登录引导，而不是 mock 推荐池
4. 推荐卡片包含：
   - `风险等级`
   - `建议仓位`
   - `止盈`
   - `止损`

### 后端

新增或增强测试，至少覆盖：

1. `ListStockRecommendations` 在传 `trade_date` 时只返回目标交易日记录
2. 同日排序优先按 `score DESC`
3. 返回项包含从 `stock_reco_details` 合并的 `take_profit / stop_loss`
4. 无 `trade_date` 时，回退到最新有效交易日

### Verification

交付前至少验证：

- 推荐页源码契约测试
- 后端 repo/handler slice tests
- `newclient` build

## Risks and Mitigations

### 风险 1：历史数据没有统一 `trade_date`

现有表结构更多依赖 `valid_from`，如果历史记录没有严格每日批次语义，可能出现部分记录无法精确映射到“今日”。

缓解：

- 第一阶段用 `DATE(valid_from)` 兼容映射
- 不先做 schema 迁移

### 风险 2：部分推荐没有 detail 明细

已有代码里 `GetStockRecommendationInsight` 已经对 detail 缺失做过估算兜底。列表若强依赖 detail，可能出现缺字段。

缓解：

- 列表的止盈止损字段允许为空态
- 不要求所有历史推荐都强补 detail

### 风险 3：未登录态体验变“空”

删掉 mock 列表后，未登录首屏会从“热闹”变成“更克制”。

缓解：

- 空态文案必须清楚
- 视觉上保留模块骨架与 CTA

## Rollout Recommendation

本次建议作为单独的小阶段上线，不与个性化推荐、推荐算法升级、实时行情接入混在一起。

推荐的上线顺序：

1. 后端列表语义收口 + 列表字段增强
2. 前端去假数据、改卡片字段
3. 验证深度推演和策略页承接链不受影响
4. 后续再单开二期做个性化推荐排序与算法增强

## Success Criteria

本次优化完成后，应满足：

1. 已登录用户在 `/recommendations` 看到的是明确的“今日推荐集合”
2. 页面不再出现随机行情或伪价格
3. 推荐卡片展示的都是后端真实字段
4. 未登录用户不会被 mock 推荐误导
5. `查看对应策略` 与 `进入深度推演` 链路继续可用
