# 历史推荐真实绩效链路 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `/recommendations/history` 从“当前推荐列表 + 前端随机绩效”改造成“真实历史推荐绩效页”。

**Architecture:** 新增独立的用户侧历史绩效接口，由后端从 `stock_recommendations + stock_reco_details + stock_market_quotes` 计算真实建仓价、最新价、区间收益和最大回撤，并返回全量汇总 summary。前端历史页切换到新接口，不再依赖 mock、随机数或当前推荐接口。

**Tech Stack:** Go, Gin, MySQL, Vue 3, Vite, Node test

---

### Task 1: 建立后端历史绩效模型与 repo 接口

**Files:**
- Modify: `backend/internal/growth/model/models.go`
- Modify: `backend/internal/growth/repo/interfaces.go`
- Test: `backend/internal/growth/repo/stock_recommendation_history_repo_test.go`

- [ ] **Step 1: 写失败的 repo 测试骨架**

在 `backend/internal/growth/repo/stock_recommendation_history_repo_test.go` 新增测试函数骨架：

- `TestListStockRecommendationHistoryBuildsRealPerformanceItems`
- `TestListStockRecommendationHistoryBuildsSummaryFromAllMatchedRows`
- `TestListStockRecommendationHistoryMapsOutcomeFromStatusAndPerformanceLabel`

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
go test ./internal/growth/repo -run 'TestListStockRecommendationHistory'
```

Expected:
- FAIL，提示方法或类型未定义

- [ ] **Step 3: 在模型层新增历史绩效对象**

在 `backend/internal/growth/model/models.go` 新增：

- `StockRecommendationHistoryItem`
- `StockRecommendationHistorySummary`

字段按 spec 定义，至少包含：

```go
type StockRecommendationHistoryItem struct {
    ID              string  `json:"id"`
    Symbol          string  `json:"symbol"`
    Name            string  `json:"name"`
    ValidFrom       string  `json:"valid_from"`
    ValidTo         string  `json:"valid_to"`
    Score           float64 `json:"score"`
    RiskLevel       string  `json:"risk_level"`
    PositionRange   string  `json:"position_range"`
    SourceType      string  `json:"source_type,omitempty"`
    StrategyVersion string  `json:"strategy_version,omitempty"`
    TakeProfit      string  `json:"take_profit,omitempty"`
    StopLoss        string  `json:"stop_loss,omitempty"`
    PerformanceLabel string `json:"performance_label,omitempty"`
    EntryPrice      float64 `json:"entry_price"`
    LatestPrice     float64 `json:"latest_price"`
    ReturnPct       float64 `json:"return_pct"`
    MaxDrawdownPct  float64 `json:"max_drawdown_pct"`
    Status          string  `json:"status"`
    Outcome         string  `json:"outcome"`
    IsClosed        bool    `json:"is_closed"`
}
```

- [ ] **Step 4: 在 repo 接口新增历史绩效方法**

在 `backend/internal/growth/repo/interfaces.go` 新增：

```go
ListStockRecommendationHistory(userID string, outcome string, tradeDateFrom string, tradeDateTo string, page int, pageSize int) ([]model.StockRecommendationHistoryItem, model.StockRecommendationHistorySummary, int, error)
```

- [ ] **Step 5: 再次运行测试确认仍然失败但进入下一层**

Run:

```bash
go test ./internal/growth/repo -run 'TestListStockRecommendationHistory'
```

Expected:
- FAIL，提示 MySQL repo 未实现接口或业务逻辑为空

- [ ] **Step 6: Commit**

```bash
git add backend/internal/growth/model/models.go backend/internal/growth/repo/interfaces.go backend/internal/growth/repo/stock_recommendation_history_repo_test.go
git commit -m "feat: add stock recommendation history models"
```

### Task 2: 实现 MySQL 历史绩效查询与真实收益计算

**Files:**
- Modify: `backend/internal/growth/repo/mysql_repo.go`
- Test: `backend/internal/growth/repo/stock_recommendation_history_repo_test.go`

- [ ] **Step 1: 扩充失败测试，锁定查询口径与收益口径**

在 `backend/internal/growth/repo/stock_recommendation_history_repo_test.go` 为测试补充 sqlmock 场景，覆盖：

- 推荐记录主查询
- 行情查询
- `entry_price`
- `latest_price`
- `return_pct`
- `max_drawdown_pct`
- summary 汇总

至少构造一个真实价格序列，如：

```text
entry=10.00
prices=[10.00, 12.00, 11.00, 9.00, 13.00]
latest=13.00
return=30.00
max_drawdown=-25.00
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
go test ./internal/growth/repo -run 'TestListStockRecommendationHistory'
```

Expected:
- FAIL，提示 repo 方法返回空结果或计算错误

- [ ] **Step 3: 在 `mysql_repo.go` 实现历史绩效查询**

新增 `ListStockRecommendationHistory` 方法，逻辑拆成三段：

1. 查推荐基础行：

```sql
SELECT r.id, r.symbol, r.name, r.score, r.risk_level,
       COALESCE(r.position_range, ''),
       r.valid_from, r.valid_to, r.status,
       COALESCE(r.source_type, ''),
       COALESCE(r.strategy_version, ''),
       COALESCE(r.performance_label, ''),
       COALESCE(d.take_profit, ''),
       COALESCE(d.stop_loss, '')
FROM stock_recommendations r
LEFT JOIN stock_reco_details d ON d.reco_id = r.id
WHERE r.status IN ('PUBLISHED','ACTIVE','TRACKING','HIT_TAKE_PROFIT','HIT_STOP_LOSS','INVALIDATED','REVIEWED')
```

并按 `outcome / tradeDateFrom / tradeDateTo` 过滤。

2. 批量查对应 symbol 的 `stock_market_quotes`
3. 在 Go 层逐条构建历史绩效 item

- [ ] **Step 4: 实现价格与绩效辅助函数**

在 `mysql_repo.go` 内新增私有 helper，建议最少包括：

- `findEntryQuote(...)`
- `findLatestQuote(...)`
- `calculateReturnPct(...)`
- `calculateMaxDrawdownPct(...)`
- `mapRecommendationOutcome(...)`
- `buildStockRecommendationHistorySummary(...)`

要求：
- 找不到 quote 时，不伪造价格
- 可以返回 0 值，但逻辑明确
- summary 只基于真实 items 统一构建，不依赖前端

- [ ] **Step 5: 运行 repo 测试直到通过**

Run:

```bash
go test ./internal/growth/repo -run 'TestListStockRecommendationHistory'
```

Expected:
- PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/growth/repo/mysql_repo.go backend/internal/growth/repo/stock_recommendation_history_repo_test.go
git commit -m "feat: compute stock recommendation history performance"
```

### Task 3: 暴露用户侧历史绩效接口

**Files:**
- Modify: `backend/internal/growth/handler/user_growth_handler.go`
- Test: `backend/internal/growth/handler/user_growth_handler_test.go`

- [ ] **Step 1: 写 handler 失败测试**

在 `backend/internal/growth/handler/user_growth_handler_test.go` 新增：

- `TestListStockRecommendationHistoryRequiresAuth`
- `TestListStockRecommendationHistoryReturnsItemsAndSummary`

断言接口：

- `GET /api/v1/stocks/recommendations/history`
- 返回 `items + summary + page + page_size + total`

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
go test ./internal/growth/handler -run 'TestListStockRecommendationHistory'
```

Expected:
- FAIL，提示 handler 或 route 未实现

- [ ] **Step 3: 实现 handler**

在 `user_growth_handler.go` 新增：

```go
func (h *UserGrowthHandler) ListStockRecommendationHistory(c *gin.Context)
```

逻辑：
- `requireUserID`
- `loadAccessProfile`
- 读取 `page/page_size/outcome/trade_date_from/trade_date_to`
- 调 `service.ListStockRecommendationHistory(...)`
- 返回统一响应

- [ ] **Step 4: 挂用户路由**

如果路由定义不在 handler 文件，补到对应 router 注册位置，确保：

```text
GET /api/v1/stocks/recommendations/history
```

- [ ] **Step 5: 运行 handler 测试确认通过**

Run:

```bash
go test ./internal/growth/handler -run 'TestListStockRecommendationHistory'
```

Expected:
- PASS

- [ ] **Step 6: Commit**

```bash
git add backend/internal/growth/handler/user_growth_handler.go backend/internal/growth/handler/user_growth_handler_test.go
git commit -m "feat: expose stock recommendation history endpoint"
```

### Task 4: 前端切换到真实历史绩效接口

**Files:**
- Modify: `newclient/src/api/market.js`
- Modify: `newclient/src/apps/pc/views/recommendations/HistoryPerf.vue`
- Test: `newclient/src/apps/pc/views/recommendations/history-perf-data-contract.test.mjs`

- [ ] **Step 1: 写前端失败契约测试**

新增 `newclient/src/apps/pc/views/recommendations/history-perf-data-contract.test.mjs`，锁定以下约束：

- 使用 `listStockRecommendationHistory`
- 不再 import `MOCK_HISTORY`
- 不再出现 `Math.random()`
- 不再调用 `listStockRecommendations`
- 顶部统计来自 `summary`
- 表头包含：
  - `建仓价`
  - `最新/结算价`
  - `真实收益`
  - `最大回撤`
- 未登录态显示登录门槛

- [ ] **Step 2: 运行测试确认失败**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/history-perf-data-contract.test.mjs
```

Expected:
- FAIL

- [ ] **Step 3: 在 API client 新增历史绩效请求**

在 `newclient/src/api/market.js` 新增：

```js
export function listStockRecommendationHistory(params) {
  return http.get("/stocks/recommendations/history", { params: buildParams(params) });
}
```

- [ ] **Step 4: 重写 `HistoryPerf.vue`**

改造方向：

- 删除 `MOCK_HISTORY`
- 删除随机 `currentPrice/actualReturn/outcome`
- 使用 `listStockRecommendationHistory`
- 用后端 `summary` 渲染顶部统计
- 用真实字段渲染表格
- 未登录态显示登录提示，不再展示 mock

推荐的核心响应式状态：

```js
const historyData = ref([]);
const historySummary = ref(null);
const showLoginGate = computed(() => !isLoggedIn.value);
```

- [ ] **Step 5: 运行前端测试直到通过**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/history-perf-data-contract.test.mjs
```

Expected:
- PASS

- [ ] **Step 6: Commit**

```bash
git add newclient/src/api/market.js newclient/src/apps/pc/views/recommendations/HistoryPerf.vue newclient/src/apps/pc/views/recommendations/history-perf-data-contract.test.mjs
git commit -m "feat: switch recommendation history to real performance data"
```

### Task 5: 全链路验证与文档收尾

**Files:**
- Verify only unless docs need touch

- [ ] **Step 1: 运行后端 repo 与 handler 相关测试**

Run:

```bash
go test ./internal/growth/repo -run 'TestListStockRecommendationHistory'
go test ./internal/growth/handler -run 'TestListStockRecommendationHistory'
```

Expected:
- PASS

- [ ] **Step 2: 运行前端相关测试**

Run:

```bash
node --test \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/history-perf-data-contract.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
```

Expected:
- PASS

- [ ] **Step 3: 构建 `newclient`**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/newclient && npm run build
```

Expected:
- PASS，PC/H5 都成功构建

- [ ] **Step 4: 人工检查关键行为**

确认：
- `/recommendations/history` 未登录态不再显示 mock 历史绩效
- 登录后页面不再显示随机收益或随机结果
- 顶部统计和列表字段与“历史推荐表现”语义一致

- [ ] **Step 5: Commit（若本任务产生补丁）**

```bash
git add -A
git commit -m "test: verify real recommendation history flow"
```

仅当这个任务有真实文件改动时执行；如果只是验证，无需空提交。
