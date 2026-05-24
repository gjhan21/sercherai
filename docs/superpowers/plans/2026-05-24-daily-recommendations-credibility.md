# 每日推荐可信化与卡片增强 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `/recommendations` 的“今日 AI 精选”只展示真实推荐数据，严格收口到目标交易日，并把列表卡片升级为真实字段驱动的推荐卡片。

**Architecture:** 后端继续沿用 `stock_recommendations + stock_reco_details` 这套现有推荐存储结构，不改推荐生成算法，只增强用户列表查询语义和返回字段。前端在不改变推荐主链承接关系的前提下，删除随机行情与 mock 推荐回退，改为明确的登录空态和真实字段卡片。

**Tech Stack:** Go, Gin, MySQL repo layer, Vue 3, node:test, Vite

---

## File Map

### Backend

- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/backend/internal/growth/model/models.go`
  - 扩展 `StockRecommendation` 列表返回字段，补足 `take_profit / stop_loss / source_type / strategy_version / performance_label`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/backend/internal/growth/repo/mysql_repo.go`
  - 调整 `ListStockRecommendations` 查询过滤和排序
  - 复用 `stock_reco_details` 合并列表字段
  - 增加“无 trade_date 时取最新有效交易日”的逻辑
- Test: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/backend/internal/growth/repo/mysql_repo_test.go` 或同目录现有相关 slice test 文件
  - 为推荐列表查询新增回归测试

### Frontend

- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient/src/apps/pc/views/recommendations/DailyRecs.vue`
  - 默认带 `trade_date`
  - 删除随机 `price/change`
  - 删除未登录使用 `MOCK_RECS` 的真实推荐伪装
  - 调整推荐卡片字段
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient/src/api/market.js`
  - 如有必要，仅补参数文档或保持接口签名不变
- Test: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs`
  - 扩展推荐页源码契约测试
- Test: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs`
  - 新增列表数据真实性与空态行为测试

### Docs

- Already created spec:
  - `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/docs/superpowers/specs/2026-05-24-daily-recommendations-credibility-design.md`

---

### Task 1: 后端推荐列表语义收口

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/backend/internal/growth/model/models.go`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/backend/internal/growth/repo/mysql_repo.go`
- Test: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/backend/internal/growth/repo/mysql_repo_test.go`

- [ ] **Step 1: 写后端 failing test，覆盖 trade_date 收口和排序**

在推荐 repo 相关测试文件里新增测试，覆盖：
- 传 `trade_date` 时，只返回该交易日记录
- 返回状态只包含 `PUBLISHED / ACTIVE / TRACKING`
- 同日内按 `score DESC` 优先，再按 `created_at DESC`

测试示例要点：

```go
func TestListStockRecommendationsFiltersByTradeDateAndSortsByScore(t *testing.T) {
    // 准备两天数据、同一天不同 score 数据
    // 调用 ListStockRecommendations("u_demo_001", "2026-05-24", 1, 10)
    // 断言只返回 2026-05-24
    // 断言 items[0].Score > items[1].Score
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run:

```bash
go test ./internal/growth/repo -run TestListStockRecommendationsFiltersByTradeDateAndSortsByScore
```

Expected:
- FAIL
- 当前实现只按 `valid_from DESC`，不会满足新排序/收口要求

- [ ] **Step 3: 实现最小后端过滤和排序逻辑**

在 `ListStockRecommendations` 中：
- 增加“无 `trade_date` 时回退为最新有效交易日”的逻辑
- 将过滤改成基于 `DATE(valid_from)` 的目标交易日收口
- 排序改为 `score DESC, created_at DESC`

实现时保持现有状态过滤逻辑兼容。

- [ ] **Step 4: 运行测试，确认通过**

Run:

```bash
go test ./internal/growth/repo -run TestListStockRecommendationsFiltersByTradeDateAndSortsByScore
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/growth/model/models.go backend/internal/growth/repo/mysql_repo.go backend/internal/growth/repo/mysql_repo_test.go
git commit -m "feat: tighten daily stock recommendation query semantics"
```

---

### Task 2: 后端列表返回真实卡片字段

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/backend/internal/growth/model/models.go`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/backend/internal/growth/repo/mysql_repo.go`
- Test: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/backend/internal/growth/repo/mysql_repo_test.go`

- [ ] **Step 1: 写后端 failing test，覆盖列表项明细字段拼接**

新增测试验证：
- `take_profit`
- `stop_loss`
- `source_type`
- `strategy_version`
- `performance_label`

能够出现在列表项中。

测试示例要点：

```go
func TestListStockRecommendationsIncludesDetailFields(t *testing.T) {
    // 准备 stock_recommendations + stock_reco_details 测试数据
    // 调用 ListStockRecommendations
    // 断言 items[0].TakeProfit / StopLoss / SourceType 不为空
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run:

```bash
go test ./internal/growth/repo -run TestListStockRecommendationsIncludesDetailFields
```

Expected:
- FAIL
- 当前列表查询没有 join detail 表

- [ ] **Step 3: 实现最小字段扩展**

在模型层扩展 `StockRecommendation` 列表返回字段。  
在 repo 层通过 `LEFT JOIN stock_reco_details` 或二次查询，把：
- `take_profit`
- `stop_loss`
- `source_type`
- `strategy_version`
- `performance_label`

拼回列表项。

注意：
- 列表字段允许为空，不要强制所有历史记录都有 detail
- 不在本期引入复杂四维评分数组

- [ ] **Step 4: 运行测试，确认通过**

Run:

```bash
go test ./internal/growth/repo -run TestListStockRecommendationsIncludesDetailFields
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/growth/model/models.go backend/internal/growth/repo/mysql_repo.go backend/internal/growth/repo/mysql_repo_test.go
git commit -m "feat: expose real recommendation card fields"
```

---

### Task 3: 前端去掉假行情并改为真实字段卡片

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient/src/apps/pc/views/recommendations/DailyRecs.vue`
- Test: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs`

- [ ] **Step 1: 写前端 failing test，锁住“无随机 price/change、请求带 trade_date”**

新增或扩展测试，至少断言：
- `loadDailyRecs()` 请求会携带 `trade_date`
- 页面源码不再使用 `Math.random()` 生成卡片 `price/change`
- 推荐卡片使用 `risk_level / position_range / take_profit / stop_loss`

测试示例：

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

test("daily recs view requests trade_date and avoids fake quote fields", () => {
  const source = fs.readFileSync(FILE, "utf8");
  assert.match(source, /trade_date/);
  assert.doesNotMatch(source, /Math\\.random\\(\\) \\* 6 - 1/);
});
```

- [ ] **Step 2: 运行测试，确认失败**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs
```

Expected:
- FAIL
- 当前视图里仍存在随机 `price/change`

- [ ] **Step 3: 实现最小前端映射改造**

在 `DailyRecs.vue` 中：
- 请求参数补 `trade_date`
- 删除随机 `price/change`
- 删除“把 score 当 price”的映射
- 卡片展示改成真实字段：
  - `score`
  - `risk_level`
  - `position_range`
  - `take_profit`
  - `stop_loss`
  - `reason_summary`

同时保持：
- `查看对应策略`
- `进入深度推演`

上下文链不变。

- [ ] **Step 4: 运行测试，确认通过**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add newclient/src/apps/pc/views/recommendations/DailyRecs.vue newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs
git commit -m "feat: remove fake quote data from daily recommendations"
```

---

### Task 4: 前端未登录空态替换 mock 推荐池

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient/src/apps/pc/views/recommendations/DailyRecs.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient/src/mock/recommendations.js`（如清理引用所需）
- Test: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs`

- [ ] **Step 1: 写前端 failing test，锁住未登录空态**

测试覆盖：
- 未登录时不再依赖 `MOCK_RECS` 作为真实推荐列表
- 页面出现“登录后查看今日 AI 精选”或等价空态文案

测试示例：

```js
test("daily recs view shows login gate instead of mock rec list", () => {
  const source = fs.readFileSync(FILE, "utf8");
  assert.match(source, /登录后查看今日 AI 精选/);
  assert.doesNotMatch(source, /const dailyRecs = ref\\(MOCK_RECS\\)/);
});
```

- [ ] **Step 2: 运行测试，确认失败**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
```

Expected:
- FAIL
- 当前仍以 `MOCK_RECS` 初始化

- [ ] **Step 3: 实现最小未登录空态改造**

在 `DailyRecs.vue` 中：
- 取消 `dailyRecs = ref(MOCK_RECS)` 作为默认真实池
- 改为：
  - 空数组初始化
  - 登录态成功后填充真实数据
  - 未登录态显示模块空态和登录提示

如果 `recommendations.js` 仅剩无用引用，清理不再使用的 `DAILY_RECS` 引入。

- [ ] **Step 4: 运行测试，确认通过**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add newclient/src/apps/pc/views/recommendations/DailyRecs.vue newclient/src/mock/recommendations.js newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
git commit -m "feat: replace mock daily recommendations with login empty state"
```

---

### Task 5: 全链路回归验证

**Files:**
- Verify only

- [ ] **Step 1: 运行后端推荐切片测试**

Run:

```bash
go test ./internal/growth/repo -run 'TestListStockRecommendationsFiltersByTradeDateAndSortsByScore|TestListStockRecommendationsIncludesDetailFields'
```

Expected:
- PASS

- [ ] **Step 2: 运行前端推荐页测试**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/daily-recs-data-contract.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
```

Expected:
- PASS

- [ ] **Step 3: 运行已有推荐/承接链相关测试**

Run:

```bash
node --test newclient/src/apps/pc/views/forecast-lab-view.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
```

Expected:
- PASS

- [ ] **Step 4: 构建前端**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/daily-recommendations-credibility/newclient && npm run build
```

Expected:
- PASS
- PC/H5 构建成功

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "test: verify daily recommendations credibility upgrade"
```

