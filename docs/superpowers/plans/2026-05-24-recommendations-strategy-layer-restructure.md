# Recommendations Strategy Layer Restructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split the current mixed `/recommendations/strategies` page into a true stock recommendation strategy bridge and a separate futures strategy center, while unifying deep forecast entry semantics and removing mock fallback behavior from the real user path.

**Architecture:** Reuse existing stock recommendation detail/insight APIs to build the stock-side strategy bridge, and migrate the current futures strategy listing into a dedicated `/futures/strategies` route. Keep futures strategy detail as the futures branch detail page, fix its routing and context semantics, and remove static backtest/mock fallback from the production user flow.

**Tech Stack:** Vue 3, Vue Router, existing `newclient` shared forecast context helpers, existing growth backend APIs, Node `--test` source-contract tests

---

### File Map

**Create**
- `newclient/src/apps/pc/views/recommendations/StockStrategyBridge.vue`
- `newclient/src/apps/pc/views/futures/FuturesStrategies.vue`
- `newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs`
- `newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs`
- `newclient/src/apps/pc/views/futures/futures-strategy-detail-routing.test.mjs`

**Modify**
- `newclient/src/apps/pc/router/index.js`
- `newclient/src/apps/pc/views/recommendations/DailyRecs.vue`
- `newclient/src/apps/pc/views/recommendations/Strategies.vue`
- `newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue`
- `newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs`
- `newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs`
- `newclient/src/api/market.js` only if route consumers need convenience wrapper consistency

**Reference**
- `backend/internal/growth/repo/mysql_repo.go:4376-4502`
- `backend/internal/growth/handler/user_growth_handler.go:1029-1095`
- `backend/internal/growth/repo/mysql_repo.go:4960-5092`
- `newclient/src/shared/lib/forecast-context.js`
- `newclient/src/shared/components/deep-forecast/DeepForecastSummaryCard.vue`
- `newclient/src/shared/composables/useDeepForecastEntry.js`

### Task 1: Lock the route split with failing router and contract tests

**Files:**
- Create: `newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs`
- Create: `newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs`
- Modify: `newclient/src/apps/pc/router/index.js`
- Modify: `newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs`

- [ ] **Step 1: Write the failing source-contract tests for the route split**

Add tests that assert:
- `/recommendations/strategies` points to `StockStrategyBridge.vue`
- `/futures/strategies` exists and points to `FuturesStrategies.vue`
- the stock bridge page does not import or call `listFuturesStrategies`
- the futures strategies page does import and call `listFuturesStrategies`

Use Node source-contract style consistent with existing tests:

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const routerFile = path.join(__dirname, "..", "router", "index.js");

test("router splits stock strategy bridge and futures strategy center", () => {
  const text = fs.readFileSync(routerFile, "utf8");
  assert.match(text, /path: "\\/recommendations\\/strategies"/);
  assert.match(text, /StockStrategyBridge\\.vue/);
  assert.match(text, /path: "\\/futures\\/strategies"/);
  assert.match(text, /FuturesStrategies\\.vue/);
});
```

- [ ] **Step 2: Run the new tests to verify they fail**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
```

Expected:
- FAIL because the new files and route split do not exist yet

- [ ] **Step 3: Update the router with the new page ownership**

Modify `newclient/src/apps/pc/router/index.js`:
- keep `/recommendations/strategies`
- point it to `../views/recommendations/StockStrategyBridge.vue`
- add `/futures/strategies`
- point it to `../views/futures/FuturesStrategies.vue`
- keep `/futures/strategy/:id`
- do not change unrelated route names unless required for consistency

- [ ] **Step 4: Run the route tests again**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
```

Expected:
- Router-specific assertions now pass
- File-content assertions may still fail until the new views exist

- [ ] **Step 5: Commit the route split**

```bash
git add newclient/src/apps/pc/router/index.js newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
git commit -m "feat: split recommendations strategy and futures strategy routes"
```

### Task 2: Build the stock recommendation strategy bridge page

**Files:**
- Create: `newclient/src/apps/pc/views/recommendations/StockStrategyBridge.vue`
- Modify: `newclient/src/apps/pc/views/recommendations/DailyRecs.vue`
- Modify: `newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs`
- Test: `newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs`

- [ ] **Step 1: Expand failing tests for stock bridge behavior**

Add assertions that:
- the bridge page reads `reco_id`, `symbol`, and `name` from route query
- when `reco_id` is missing, the page shows an explicit “请先从每日推荐进入策略承接页” empty state
- the page imports `getStockRecommendationDetail` and `getStockRecommendationInsight`
- the page builds a deep forecast query with `targetType: "STOCK"`
- the page does not import `listFuturesStrategies`

- [ ] **Step 2: Run the stock bridge tests to verify failure**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
```

Expected:
- FAIL because `StockStrategyBridge.vue` does not exist yet

- [ ] **Step 3: Implement `StockStrategyBridge.vue` with strict context gate**

Build a focused page that:
- reads `reco_id`, `symbol`, `name` from `route.query`
- refuses to load strategy data without `reco_id`
- renders:
  - bridge header with “每日推荐 -> 执行策略 -> 深度推演”
  - strategy verdict card from recommendation detail
  - execution plan card from recommendation insight
  - evidence cards from detail/score framework
  - actions:
    - back to `/recommendations`
    - go to `/identify/:symbol` or `/identify` with context
    - go to `/forecast-lab` using `buildForecastContextQuery`
- uses existing API methods from `market.js`
- uses explicit loading/error/empty states
- does not silently fallback to mock data

- [ ] **Step 4: Update daily recommendations entry to pass strict stock bridge context**

Modify `newclient/src/apps/pc/views/recommendations/DailyRecs.vue` so the “查看对应策略” action sends:
- `reco_id`
- `symbol`
- `name`
- `from`

Do not rely on only `symbol` and `name` anymore.

- [ ] **Step 5: Run tests for the stock bridge**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
```

Expected:
- PASS

- [ ] **Step 6: Commit the stock bridge**

```bash
git add newclient/src/apps/pc/views/recommendations/StockStrategyBridge.vue newclient/src/apps/pc/views/recommendations/DailyRecs.vue newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
git commit -m "feat: add stock recommendation strategy bridge"
```

### Task 3: Move the current futures strategies page into the futures module

**Files:**
- Create: `newclient/src/apps/pc/views/futures/FuturesStrategies.vue`
- Modify: `newclient/src/apps/pc/views/recommendations/Strategies.vue`
- Test: `newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs`

- [ ] **Step 1: Strengthen the futures page tests**

Add assertions that the futures strategies page:
- imports `listFuturesStrategies`
- does not use recommendation-step copy
- does not render “回看推荐来源”
- uses futures-oriented page copy
- keeps deep forecast entry with futures semantics

- [ ] **Step 2: Run the futures page tests to verify failure**

Run:

```bash
node --test newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs
```

Expected:
- FAIL because `FuturesStrategies.vue` does not exist yet

- [ ] **Step 3: Implement `FuturesStrategies.vue` by migrating the current list logic**

Create `newclient/src/apps/pc/views/futures/FuturesStrategies.vue` by adapting the current `Strategies.vue` logic:
- keep `listFuturesStrategies` data loading
- remove stock-mainline copy
- remove “回看推荐来源”
- remove the static backtest button from real strategy cards
- keep the card click to `/futures/strategy/:id`
- keep the futures deep forecast CTA with `targetType: "FUTURES"`
- replace mock fallback with explicit loading/error/empty behavior

- [ ] **Step 4: Turn the old `Strategies.vue` into a thin compatibility redirect or remove it from active routing**

Preferred approach:
- keep the file only if helpful during migration
- otherwise leave it unused once router points to `StockStrategyBridge.vue`
- do not keep real business logic duplicated in both pages

- [ ] **Step 5: Run futures page tests**

Run:

```bash
node --test newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs
```

Expected:
- PASS

- [ ] **Step 6: Commit the futures strategy center page**

```bash
git add newclient/src/apps/pc/views/futures/FuturesStrategies.vue newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs newclient/src/apps/pc/views/recommendations/Strategies.vue
git commit -m "feat: move futures strategies into futures center"
```

### Task 4: Fix futures strategy detail routing and semantics

**Files:**
- Modify: `newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue`
- Modify: `newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs`
- Create: `newclient/src/apps/pc/views/futures/futures-strategy-detail-routing.test.mjs`

- [ ] **Step 1: Add failing tests for futures strategy detail routing and context**

Assertions should cover:
- back button routes to `/futures/strategies`
- `forecastLabEntryTo` uses `targetType: "FUTURES"`
- `sourcePath` points to `/futures/strategy/:id` or `/futures/strategies`, not `/recommendations/strategies`
- the page still uses `DeepForecastSummaryCard` and `useDeepForecastEntry`

- [ ] **Step 2: Run the futures detail tests to verify failure**

Run:

```bash
node --test newclient/src/apps/pc/views/futures/futures-strategy-detail-routing.test.mjs newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs
```

Expected:
- FAIL because the old route path is still `/strategies` and legacy source path assumptions still exist

- [ ] **Step 3: Update `FuturesStrategyDetail.vue`**

Make the following targeted changes:
- back button goes to `/futures/strategies`
- `sourcePath` fallback no longer references `/recommendations/strategies`
- optional copy adjustments make clear this page belongs to the futures branch
- keep existing detail and deep forecast loading behavior intact

- [ ] **Step 4: Run futures detail tests again**

Run:

```bash
node --test newclient/src/apps/pc/views/futures/futures-strategy-detail-routing.test.mjs newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs
```

Expected:
- PASS

- [ ] **Step 5: Commit the futures detail fixes**

```bash
git add newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue newclient/src/apps/pc/views/futures/futures-strategy-detail-routing.test.mjs newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs
git commit -m "fix: align futures strategy detail with futures route semantics"
```

### Task 5: Remove mock backtest from the real strategy flow and verify the whole PC strategy split

**Files:**
- Modify: `newclient/src/apps/pc/views/futures/FuturesStrategies.vue`
- Modify: `newclient/src/apps/pc/views/recommendations/StockStrategyBridge.vue`
- Modify: `newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs`
- Modify: `newclient/src/apps/pc/views/recommendations/history-perf-data-contract.test.mjs` only if route text assumptions changed indirectly

- [ ] **Step 1: Add failing assertions that real strategy cards no longer route to static mock backtest**

Update source-contract tests to assert:
- real stock bridge actions do not link to `/recommendations/backtest`
- futures strategy cards do not expose the static backtest button

- [ ] **Step 2: Run the strategy-flow test suite**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs newclient/src/apps/pc/views/futures/futures-strategy-detail-routing.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs
```

Expected:
- FAIL until the static backtest entry is gone from real strategy paths

- [ ] **Step 3: Remove the static backtest entry from real strategy cards**

Ensure:
- no real-stock or real-futures strategy surface links to `/recommendations/backtest`
- if needed, leave the backtest route in the router for legacy/manual access, but it must not be part of the real strategy user flow

- [ ] **Step 4: Run the full PC strategy split verification**

Run:

```bash
node --test newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs newclient/src/apps/pc/views/futures/futures-strategy-detail-routing.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs
```

Expected:
- PASS

- [ ] **Step 5: Run the app build**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/newclient && npm run build
```

Expected:
- PASS
- PC and H5 bundles compile successfully even though this change is PC-focused

- [ ] **Step 6: Commit the final flow cleanup**

```bash
git add newclient/src/apps/pc/views/recommendations/StockStrategyBridge.vue newclient/src/apps/pc/views/futures/FuturesStrategies.vue newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs newclient/src/apps/pc/views/futures/futures-strategy-detail-routing.test.mjs newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs
git commit -m "feat: separate stock strategy bridge from futures strategy center"
```

### Task 6: Final verification and handoff

**Files:**
- No new files expected

- [ ] **Step 1: Run the focused backend sanity check only if any backend file changed during implementation**

Run only if implementation ended up touching backend APIs unexpectedly:

```bash
go test ./internal/growth/handler ./internal/growth/repo
```

Expected:
- PASS for the touched slice

- [ ] **Step 2: Run the final frontend verification set**

```bash
node --test newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs newclient/src/apps/pc/views/futures/futures-strategy-detail-routing.test.mjs newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs
cd /Users/gjhan21/cursor/sercherai/newclient && npm run build
```

Expected:
- PASS

- [ ] **Step 3: Inspect git diff for accidental unrelated changes**

Run:

```bash
git diff --stat
```

Expected:
- only route split, strategy pages, and related tests changed
- existing user local homepage edits remain untouched

- [ ] **Step 4: Commit any final cleanup**

```bash
git add newclient/src/apps/pc/router/index.js newclient/src/apps/pc/views/recommendations/DailyRecs.vue newclient/src/apps/pc/views/recommendations/StockStrategyBridge.vue newclient/src/apps/pc/views/futures/FuturesStrategies.vue newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue newclient/src/apps/pc/views/recommendations/recommendation-flow-entry.test.mjs newclient/src/apps/pc/views/recommendations/stock-strategy-bridge.test.mjs newclient/src/apps/pc/views/futures/futures-strategies-page.test.mjs newclient/src/apps/pc/views/futures/futures-strategy-detail-routing.test.mjs newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs
git commit -m "chore: finalize recommendations strategy layer split"
```
