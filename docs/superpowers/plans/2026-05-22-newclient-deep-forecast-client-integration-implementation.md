# Newclient Deep Forecast Client Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add deep-forecast read-side support to `newclient` so PC and H5 users can see a summary card inside existing stock/futures pages and open a dedicated forecast detail page for full L3 reports.

**Architecture:** Keep `newclient`’s existing page structure intact and add deep forecast as a pluggable enhancement layer. Build one shared summary-normalization layer, one shared summary-card surface, and one dedicated forecast-detail container per platform shell, then wire those into the existing stock and futures host pages through small page-specific adapters.

**Tech Stack:** Vue 3, Vue Router 4, Vite, Axios, Node.js `node:test`, source-contract tests, shared client-side formatting helpers.

---

## Scope

This plan covers one coherent sub-project:

- connect existing backend `forecast run` read APIs into `newclient` as a two-level client experience

It does not include:

- adding a top-level “深度推演” navigation module
- adding a forecast list/history page
- letting users create forecast runs from `newclient`
- redesigning the broader recommendations IA
- creating a brand-new H5 futures strategy host page

## Current Code Reality

### What already exists

- `/Users/gjhan21/cursor/sercherai/newclient/src/api/forecast.js` already exposes:
  - `createForecastRun`
  - `listForecastRuns`
  - `getForecastRunDetail`
- `/Users/gjhan21/cursor/sercherai/newclient/src/api/membership.js` already exposes `getMembershipQuota()`
- `PC` already has deep host pages that can accept inserted summary cards:
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/analysis/StockAnalysis.vue`
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue`
- `H5` already has a stock host page:
  - `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/StockDetail.vue`
- old `/Users/gjhan21/cursor/sercherai/client` already contains proven behavior references:
  - `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
  - `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.test.js`
  - `/Users/gjhan21/cursor/sercherai/client/src/apps/pc/views/PcForecastRunView.vue`
  - `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5ForecastRunView.vue`

### What is missing

- `newclient` has no deep forecast summary normalization helper
- `newclient` has no dedicated forecast detail routes on `PC` or `H5`
- `newclient` has no shared forecast summary component
- stock/futures host pages do not expose deep forecast entry points yet
- `newclient` currently has no local frontend test harness of its own, so this feature should lean on repository-established `node:test` source-contract and pure-helper tests instead of introducing a full DOM test stack in this project

### Important constraint

`H5` does not currently have a dedicated futures strategy detail page matching the `PC` host. The implementation must therefore:

- fully support `FUTURES` in the shared summary model and forecast detail pages
- provide at least a minimal `H5` futures entry surface in an existing host page
- avoid inventing a large new H5 futures page as part of this feature

## File Structure

### New shared utility files

- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.js`

### New shared component files

- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/components/deep-forecast/DeepForecastSummaryCard.vue`

### New shared composable files

- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/composables/useDeepForecastEntry.js`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/composables/useForecastRunDetail.js`

### New platform view files

- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue`

### Existing platform files to modify

- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/analysis/StockAnalysis.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/StockDetail.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/futures/FuturesArbitrageDetail.vue`

### New tests

- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/pc-forecast-route.test.mjs`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/h5-forecast-route.test.mjs`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-futures-detail-deep-forecast.test.mjs`

## Testing Strategy

Use the repository’s established lightweight frontend test style for `newclient`:

- `node:test`
- `node:assert/strict`
- source-contract tests that inspect files for required wiring
- pure helper tests for business logic normalization

Do **not** introduce `Vitest`, `jsdom`, or `@vue/test-utils` as part of this feature unless implementation becomes blocked and there is no simpler path. The design goal is to keep the feature aligned with current repo testing habits and minimize setup churn.

## Shared Reference Behavior

The following old-client behaviors are treated as requirements to preserve semantically:

- `buildStrategyDeepForecastSummary` in `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
- `node:test` examples in `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.test.js`
- route/view surface tests like:
  - `/Users/gjhan21/cursor/sercherai/client/src/apps/pc/views/forecast-run-view.test.js`
  - `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/router/h5-search-surface.test.js`
  - `/Users/gjhan21/cursor/sercherai/client/src/apps/pc/router/pc-search-layout.test.js`

That means the implementation should preserve:

- normalized deep forecast status labels
- readable summary extraction rules
- `VIP` summary-vs-body boundary
- dedicated route wiring assertions
- source-level proof that hosts render summary entry points

## Task 1: Add Shared Forecast Summary Normalization

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.js`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs`
- Reference: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
- Reference: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.test.js`

- [ ] **Step 1: Write the failing normalization test**

```js
import test from "node:test";
import assert from "node:assert/strict";
import { buildDeepForecastSummary } from "./forecast-summary.js";

test("buildDeepForecastSummary normalizes readable L3 summary fields", () => {
  const result = buildDeepForecastSummary({
    deep_forecast_summary: {
      run_id: "l3run_demo_001",
      status: "SUCCEEDED",
      executive_summary: "主情景继续有效。",
      primary_scenario: "bull",
      action_guidance: "沿确认信号执行",
      generated_at: "2026-05-22T12:30:00Z",
      report_available: true
    },
    deep_forecast_report_ref: {
      run_id: "l3run_demo_001",
      report_id: "l3report_demo_001",
      requires_vip: true,
      full_readable: false
    }
  });

  assert.equal(result.runId, "l3run_demo_001");
  assert.equal(result.status, "SUCCEEDED");
  assert.equal(result.statusLabel, "已完成");
  assert.equal(result.requiresVip, true);
  assert.equal(result.reportAvailable, true);
});

test("buildDeepForecastSummary keeps running and failed states readable", () => {
  const running = buildDeepForecastSummary({
    deep_forecast_summary: { run_id: "l3run_running", status: "RUNNING", executive_summary: "推演中" }
  });
  const failed = buildDeepForecastSummary({
    deep_forecast_summary: { run_id: "l3run_failed", status: "FAILED", executive_summary: "失败" }
  });

  assert.equal(running.statusLabel, "推演中");
  assert.equal(failed.statusLabel, "已失败");
});

test("buildDeepForecastSummary returns null when no usable summary exists", () => {
  assert.equal(buildDeepForecastSummary({}), null);
});
```

- [ ] **Step 2: Run the helper test to verify it fails**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs
```

Expected:

- FAIL because `forecast-summary.js` does not exist yet

- [ ] **Step 3: Implement the shared normalization helper**

Required implementation:

- create `buildDeepForecastSummary(source)`
- accept both `deep_forecast_summary` and `report_ref`-style inputs
- normalize at minimum:
  - `runId`
  - `reportId`
  - `status`
  - `statusLabel`
  - `summary`
  - `scenario`
  - `actionGuidance`
  - `generatedAt`
  - `requiresVip`
  - `reportAvailable`
  - `tone`
- keep behavior semantically aligned with old client summary-building logic

- [ ] **Step 4: Run the helper test to verify it passes**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs
```

Expected:

- PASS

- [ ] **Step 5: Commit**

```bash
git add \
  /Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.js \
  /Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs
git commit -m "feat: add shared deep forecast summary normalization"
```

## Task 2: Add Shared Forecast Detail Composables

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/composables/useDeepForecastEntry.js`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/composables/useForecastRunDetail.js`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/api/forecast.js` only if a small helper export is truly needed
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs`

- [ ] **Step 1: Extend the failing helper test to cover entry derivation**

Add a test like:

```js
test("buildDeepForecastSummary keeps CTA-safe run id and VIP metadata", () => {
  const result = buildDeepForecastSummary({
    deep_forecast_summary: {
      run_id: "l3run_entry_001",
      status: "QUEUED",
      executive_summary: "排队中"
    },
    deep_forecast_report_ref: {
      run_id: "l3run_entry_001",
      requires_vip: false
    }
  });

  assert.equal(result.runId, "l3run_entry_001");
  assert.equal(result.statusLabel, "排队中");
});
```

- [ ] **Step 2: Run the helper test to verify the new case fails if needed**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs
```

Expected:

- either FAIL if a field is missing, or PASS if the helper already covers it

- [ ] **Step 3: Implement `useDeepForecastEntry` and `useForecastRunDetail`**

Required implementation:

- `useDeepForecastEntry`
  - accept page source data
  - call `buildDeepForecastSummary`
  - build route target for `/forecast/:id` or `/m/forecast/:id`
  - expose a minimal surface the host page can consume
- `useForecastRunDetail`
  - load `getForecastRunDetail(id)`
  - expose `run`, `report`, `logs`, `loading`, `errorMessage`
  - start polling only for `QUEUED` / `RUNNING`
  - stop polling on unmount / cleanup

- [ ] **Step 4: Re-run helper tests to confirm shared behavior still passes**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs
```

Expected:

- PASS

- [ ] **Step 5: Commit**

```bash
git add \
  /Users/gjhan21/cursor/sercherai/newclient/src/shared/composables/useDeepForecastEntry.js \
  /Users/gjhan21/cursor/sercherai/newclient/src/shared/composables/useForecastRunDetail.js \
  /Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs
git commit -m "feat: add shared deep forecast entry and detail composables"
```

## Task 3: Add PC and H5 Forecast Detail Routes

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/index.js`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/pc-forecast-route.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/h5-forecast-route.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs`

- [ ] **Step 1: Write the failing route contract tests**

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

test("pc router registers dedicated forecast detail route", () => {
  const text = fs.readFileSync("/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/index.js", "utf8");
  assert.match(text, /path:\s*\"\/forecast\/:id\"/);
  assert.match(text, /name:\s*\"forecast-detail\"/);
});

test("h5 router registers dedicated forecast detail route", () => {
  const text = fs.readFileSync("/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/index.js", "utf8");
  assert.match(text, /path:\s*\"\/forecast\/:id\"/);
  assert.match(text, /name:\s*\"h5-forecast-detail\"/);
});
```

- [ ] **Step 2: Write the failing view surface tests**

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

test("pc forecast detail view renders report and log surfaces", () => {
  const text = fs.readFileSync("/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue", "utf8");
  assert.match(text, /深推演报告/);
  assert.match(text, /运行日志/);
  assert.match(text, /getForecastRunDetail/);
  assert.match(text, /getMembershipQuota/);
});
```

- [ ] **Step 3: Run the route and view tests to verify they fail**

Run:

```bash
node --test \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/pc-forecast-route.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/h5-forecast-route.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs
```

Expected:

- FAIL because the routes and views do not exist yet

- [ ] **Step 4: Implement both platform detail pages and route wiring**

Required implementation:

- add `PC` route `/forecast/:id`
- add `H5` route `/forecast/:id` under H5 base `/m/`
- build both detail views around `useForecastRunDetail`
- include:
  - hero status
  - structured summary
  - action guidance
  - alternative scenarios
  - logs
  - VIP body lock state
  - back navigation behavior
- keep `PC` denser and `H5` single-column, but share semantics

- [ ] **Step 5: Re-run the route and view tests**

Run:

```bash
node --test \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/pc-forecast-route.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/h5-forecast-route.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs
```

Expected:

- PASS

- [ ] **Step 6: Commit**

```bash
git add \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/index.js \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/index.js \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/pc-forecast-route.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/h5-forecast-route.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs
git commit -m "feat: add deep forecast detail routes for pc and h5"
```

## Task 4: Add Shared Summary Card Surface

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/components/deep-forecast/DeepForecastSummaryCard.vue`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs`

- [ ] **Step 1: Write the failing summary-surface tests**

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

test("stock analysis view renders deep forecast summary entry surface", () => {
  const text = fs.readFileSync("/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/analysis/StockAnalysis.vue", "utf8");
  assert.match(text, /DeepForecastSummaryCard/);
  assert.match(text, /深度推演/);
});
```

- [ ] **Step 2: Run the source-contract tests to verify they fail**

Run:

```bash
node --test \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs
```

Expected:

- FAIL because the summary card is not wired yet

- [ ] **Step 3: Implement the shared summary card**

Required implementation:

- show title and status chip
- show summary
- show scenario and action guidance when present
- show “查看完整深度推演” CTA when a `runId` exists
- support at least:
  - `pc` layout mode
  - `h5` layout mode
  - `running`
  - `failed`
  - `vip`

- [ ] **Step 4: Re-run the source-contract tests**

Run:

```bash
node --test \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs
```

Expected:

- still FAIL until host pages are wired; summary card itself is now ready

- [ ] **Step 5: Commit**

```bash
git add \
  /Users/gjhan21/cursor/sercherai/newclient/src/shared/components/deep-forecast/DeepForecastSummaryCard.vue
git commit -m "feat: add shared deep forecast summary card"
```

## Task 5: Wire PC Stock Analysis Entry Surface

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/analysis/StockAnalysis.vue`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs`

- [ ] **Step 1: Expand the failing stock analysis source-contract test**

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

test("stock analysis wires deep forecast summary before long-form report", () => {
  const text = fs.readFileSync("/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/analysis/StockAnalysis.vue", "utf8");
  assert.match(text, /useDeepForecastEntry/);
  assert.match(text, /DeepForecastSummaryCard/);
  assert.match(text, /forecast/);
});
```

- [ ] **Step 2: Run the stock analysis test to verify it fails**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs
```

Expected:

- FAIL

- [ ] **Step 3: Wire `StockAnalysis.vue` to the shared deep forecast entry**

Required implementation:

- derive deep forecast summary from the loaded stock insight
- insert the summary card in the explanation zone
- route CTA to `/forecast/:id`
- keep page behavior safe when summary is missing

- [ ] **Step 4: Re-run the stock analysis test**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs
```

Expected:

- PASS

- [ ] **Step 5: Commit**

```bash
git add \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/analysis/StockAnalysis.vue \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs
git commit -m "feat: add deep forecast entry to pc stock analysis"
```

## Task 6: Wire PC Futures Strategy Entry Surface

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs`

- [ ] **Step 1: Write the failing futures strategy contract test**

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

test("pc futures strategy detail exposes deep forecast summary entry", () => {
  const text = fs.readFileSync("/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue", "utf8");
  assert.match(text, /DeepForecastSummaryCard/);
  assert.match(text, /useDeepForecastEntry/);
  assert.match(text, /深度推演/);
});
```

- [ ] **Step 2: Run the futures strategy test to verify it fails**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs
```

Expected:

- FAIL

- [ ] **Step 3: Wire `FuturesStrategyDetail.vue` to the shared deep forecast entry**

Required implementation:

- derive summary from futures insight or explanation payload
- insert summary under the existing AI insight section
- CTA goes to `/forecast/:id`
- do not break current strategy and guidance layout

- [ ] **Step 4: Re-run the futures strategy test**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs
```

Expected:

- PASS

- [ ] **Step 5: Commit**

```bash
git add \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs
git commit -m "feat: add deep forecast entry to pc futures strategy detail"
```

## Task 7: Wire H5 Stock Detail Entry Surface

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/StockDetail.vue`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs`

- [ ] **Step 1: Expand the failing H5 stock detail test**

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

test("h5 stock detail renders deep forecast summary card below AI brief", () => {
  const text = fs.readFileSync("/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/StockDetail.vue", "utf8");
  assert.match(text, /DeepForecastSummaryCard/);
  assert.match(text, /useDeepForecastEntry/);
  assert.match(text, /深度推演/);
});
```

- [ ] **Step 2: Run the H5 stock detail test to verify it fails**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs
```

Expected:

- FAIL

- [ ] **Step 3: Wire `StockDetail.vue` to the shared deep forecast entry**

Required implementation:

- insert a compact summary card below “AI 快评”
- route CTA to `/m/forecast/:id`
- keep the current “查看完整 AI 分析报告” action intact
- safely hide the summary when data is absent

- [ ] **Step 4: Re-run the H5 stock detail test**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs
```

Expected:

- PASS

- [ ] **Step 5: Commit**

```bash
git add \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/StockDetail.vue \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs
git commit -m "feat: add deep forecast entry to h5 stock detail"
```

## Task 8: Add Minimal H5 Futures Entry Surface

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/futures/FuturesArbitrageDetail.vue`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-futures-detail-deep-forecast.test.mjs`

- [ ] **Step 1: Write the failing H5 futures contract test**

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

test("h5 futures detail exposes a minimal deep forecast entry surface", () => {
  const text = fs.readFileSync("/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/futures/FuturesArbitrageDetail.vue", "utf8");
  assert.match(text, /深度推演|查看推演|forecast/i);
});
```

- [ ] **Step 2: Run the H5 futures test to verify it fails**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-futures-detail-deep-forecast.test.mjs
```

Expected:

- FAIL

- [ ] **Step 3: Add a minimal H5 futures entry surface**

Required implementation:

- add a compact summary or deep-link block to the existing H5 futures host page
- do not over-design the page into a full futures strategy experience
- keep copy honest if only a summary or jump action is available

- [ ] **Step 4: Re-run the H5 futures test**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-futures-detail-deep-forecast.test.mjs
```

Expected:

- PASS

- [ ] **Step 5: Commit**

```bash
git add \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/futures/FuturesArbitrageDetail.vue \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-futures-detail-deep-forecast.test.mjs
git commit -m "feat: add minimal deep forecast entry to h5 futures detail"
```

## Task 9: Run Full Regression For The Feature Slice

**Files:**
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/pc-forecast-route.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/h5-forecast-route.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-futures-detail-deep-forecast.test.mjs`

- [ ] **Step 1: Run the full node:test forecast suite**

Run:

```bash
node --test \
  /Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/pc-forecast-route.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/h5-forecast-route.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-futures-detail-deep-forecast.test.mjs
```

Expected:

- PASS across the new forecast feature slice

- [ ] **Step 2: Run the production build**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/newclient && npm run build
```

Expected:

- PASS
- `dist/index.html` exists
- `dist-h5/m/index.html` exists

- [ ] **Step 3: Sanity-check route strings and summary entry strings**

Run:

```bash
rg -n "/forecast/:id|深度推演|查看完整深度推演" /Users/gjhan21/cursor/sercherai/newclient/src
```

Expected:

- route declarations found in both routers
- summary entry copy found in host pages and/or shared component

- [ ] **Step 4: Commit**

```bash
git add \
  /Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/pc-forecast-route.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/h5-forecast-route.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/h5-futures-detail-deep-forecast.test.mjs \
  /Users/gjhan21/cursor/sercherai/newclient/src
git commit -m "test: verify deep forecast client integration"
```

## Final Verification Checklist

- [ ] `PC` route `/forecast/:id` exists
- [ ] `H5` route `/m/forecast/:id` exists
- [ ] shared deep forecast summary helper exists and is tested
- [ ] stock analysis page shows deep forecast entry when data exists
- [ ] PC futures strategy page shows deep forecast entry when data exists
- [ ] H5 stock detail page shows deep forecast entry when data exists
- [ ] H5 futures host exposes at least a minimal deep forecast entry surface
- [ ] `VIP` gate only blocks body content, not summary usefulness
- [ ] forecast detail pages handle running, failed, and readable states
- [ ] `npm run build` passes in `newclient`
