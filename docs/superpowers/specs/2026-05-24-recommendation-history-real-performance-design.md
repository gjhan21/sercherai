# 历史推荐真实绩效链路设计

## 背景

当前 `/recommendations/history` 页面标题是“历史推荐表现”，但实际数据源并不是历史绩效。

现状链路存在三个核心问题：

1. 页面调用的是当前推荐接口 `GET /stocks/recommendations`，拿到的是“当前/最新有效推荐集合”，不是历史推荐结果集合。
2. 页面把 `recPrice`、`currentPrice`、`actualReturn`、`outcome` 通过 `score` 和 `Math.random()` 在前端伪造出来。
3. 顶部胜率、平均收益、最高收益、最大回撤都是前端基于伪造字段计算，因此没有业务可信度。

因此，这一页当前并不是“历史推荐表现”，而只是“当前推荐列表伪装成历史绩效页”。

## 目标

把 `/recommendations/history` 正式修复为“真实历史推荐绩效页”。

修复后的页面必须满足：

- 数据来源是独立的历史绩效接口，而不是复用当前推荐接口。
- `entry_price`、`latest_price / exit_price`、`return_pct`、`max_drawdown_pct` 必须来自真实行情数据。
- 顶部统计由后端基于全量历史数据计算，不能由前端基于当前页 items 估算。
- 未登录态不再显示 mock 历史绩效。
- 页面标题、表头、统计口径与数据语义一致。

## 范围

本次只修复“股票历史推荐表现页”的真实绩效链路，不扩展到：

- 期货历史绩效页
- 推荐算法升级
- 推荐个性化
- 复杂回测指标
- 独立复盘平台

## 数据来源策略

### 推荐记录来源

仍然使用现有推荐主表与详情表：

- `stock_recommendations`
- `stock_reco_details`

但历史页查询状态范围会扩大到：

- `PUBLISHED`
- `ACTIVE`
- `TRACKING`
- `HIT_TAKE_PROFIT`
- `HIT_STOP_LOSS`
- `INVALIDATED`
- `REVIEWED`

这样历史页既能看到已结束推荐，也能看到仍在跟踪中的推荐。

### 行情来源

第一版使用股票专用行情表：

- `stock_market_quotes`

而不是直接抽象到 `market_daily_bars` 多源真值体系，原因是：

- 这次只修复股票历史绩效页。
- `stock_market_quotes` 字段已经足够支持第一版真实绩效计算。
- 可以降低实现复杂度和联调风险。

## 新接口设计

新增用户侧接口：

- `GET /api/v1/stocks/recommendations/history`

### Query 参数

- `page`
- `page_size`
- `outcome`
- `trade_date_from`
- `trade_date_to`

第一版不增加更多筛选项。

### 返回结构

```json
{
  "items": [],
  "summary": {
    "total_count": 0,
    "success_count": 0,
    "fail_count": 0,
    "neutral_count": 0,
    "ongoing_count": 0,
    "win_rate": 0,
    "avg_return_pct": 0,
    "max_return_pct": 0,
    "max_drawdown_pct": 0
  },
  "page": 1,
  "page_size": 20,
  "total": 0
}
```

## 后端模型设计

新增模型对象：

### `StockRecommendationHistoryItem`

字段建议：

- `id`
- `symbol`
- `name`
- `valid_from`
- `valid_to`
- `score`
- `risk_level`
- `position_range`
- `source_type`
- `strategy_version`
- `take_profit`
- `stop_loss`
- `performance_label`
- `entry_price`
- `latest_price`
- `return_pct`
- `max_drawdown_pct`
- `status`
- `outcome`
- `is_closed`

### `StockRecommendationHistorySummary`

字段建议：

- `total_count`
- `success_count`
- `fail_count`
- `neutral_count`
- `ongoing_count`
- `win_rate`
- `avg_return_pct`
- `max_return_pct`
- `max_drawdown_pct`

## 绩效计算口径

### 建仓价 `entry_price`

取推荐生效日 `valid_from` 之后最近一个交易日的 `close_price`。

如果 `valid_from` 当天非交易日，则向后寻找最近可用交易日。

### 最新/结算价 `latest_price`

- 若推荐已经结束，则取 `valid_to` 当天或之前最近一个交易日的 `close_price`
- 若推荐仍在跟踪，则取该标的最新交易日的 `close_price`

### 区间收益 `return_pct`

```text
(latest_price - entry_price) / entry_price * 100
```

### 最大回撤 `max_drawdown_pct`

第一版使用推荐存续期间的真实收盘价序列计算最大回撤。

口径：

- 从建仓后收盘价序列中寻找阶段最高点
- 计算其后回落到最低点的最大跌幅百分比

这是“真实行情驱动但简化版”的回撤口径，优先保证真实性，再考虑高级回测细化。

## 结果映射规则

### `outcome`

第一版采用规则映射：

- `HIT_TAKE_PROFIT` -> `success`
- `HIT_STOP_LOSS` -> `fail`
- `INVALIDATED` -> `fail`
- `ACTIVE / TRACKING / PUBLISHED` -> `ongoing`
- `REVIEWED` -> 基于 `performance_label` 映射

### `REVIEWED` 的 performance_label 映射

第一版采用保守规则：

- 明显正向标签 -> `success`
- 明显负向标签 -> `fail`
- 其他 -> `neutral`

该规则在 repo 层集中实现，避免前端二次猜测。

## 前端页面改造

页面：

- `newclient/src/apps/pc/views/recommendations/HistoryPerf.vue`

### 改造原则

- 删除 mock 数据和随机收益逻辑
- 改为调用新的历史绩效接口
- 顶部统计使用后端 summary
- 表格展示真实绩效字段

### 表格字段调整

建议表头改为：

- 推荐日期
- 股票
- 名称
- 建仓价
- 最新/结算价
- 真实收益
- 最大回撤
- 推荐评分
- 策略版本
- 结果

### 未登录态

不再显示 mock 历史表现，改为登录门槛提示。

## 非目标

本次不做：

- 期货历史推荐绩效
- 实时行情接入
- 分时回测
- 年化收益、Sharpe、信息比率
- 推荐复盘平台
- 用户自定义绩效筛选器

## 测试策略

### 后端

新增 repo 测试，覆盖：

- 历史记录筛选状态范围
- `entry_price` 取推荐起始后最近交易日
- `latest_price` 取结束日或最新交易日
- `return_pct` 真实计算
- `max_drawdown_pct` 正确计算
- `summary` 基于全量记录，而不是当前分页

新增 handler 测试，覆盖：

- 登录校验
- 参数透传
- 返回 `{ items, summary, page, page_size, total }`

### 前端

新增页面契约测试，覆盖：

- 使用 `listStockRecommendationHistory`
- 不再 import mock history
- 不再使用 `Math.random()`
- 使用后端 summary
- 表头为真实绩效字段
- 未登录态显示登录门槛

## 风险与兼容性

### 风险

- 部分历史推荐可能缺少对应行情数据
- 某些状态记录可能没有清晰的终态日期
- `performance_label` 的语义可能存在脏数据

### 第一版兼容策略

- 找不到行情时，单条记录允许价格字段为空，但不能伪造
- 价格缺失的记录仍可展示推荐基础信息和状态
- `summary` 只统计成功算出真实收益的记录，并在实现中保持规则清晰

## 成功标准

修复完成后：

- `/recommendations/history` 不再依赖当前推荐列表
- 页面不再使用任何随机数生成绩效
- 历史推荐收益率和回撤来自真实行情
- 页面统计项来自后端 summary
- 用户看到的“历史推荐表现”与页面标题语义一致
