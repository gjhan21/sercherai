# System Jobs Sync Model Upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the old market backfill workspace in `system-jobs` into a unified stock/futures sync task center, and retire `data-sources/sync` as the primary stock/futures sync entry.

**Architecture:** Keep the existing `marketBackfillRun` execution and detail pipeline, but wrap it in a new frontend sync-task model with four operator-facing templates: stock full sync, stock daily incremental sync, futures full sync, and futures daily incremental sync. Expand the backend DTO and execution path to admit `FUTURES`, move stock/futures operator actions into `system-jobs`, and convert `data-sources/sync` into a governance-and-redirect page.

**Tech Stack:** Vue 3, Element Plus, Vue Router, existing admin composables/helpers, Go Gin handlers, existing market backfill repo/service chain, Node `--test`, Go `test`

---

### File Map

**Create**
- `admin/src/views/system-jobs-sync-model.test.js`
- `backend/internal/growth/repo/market_data_backfill_futures_test.go`

**Modify**
- `admin/src/views/SystemJobsView.vue`
- `admin/src/lib/system-jobs-admin.js`
- `admin/src/lib/system-jobs-admin.test.js`
- `admin/src/views/DataSourcesView.vue`
- `admin/src/composables/useDataSourcesWorkspace.js`
- `admin/src/composables/useMarketSyncConsole.js`
- `admin/src/composables/useMarketSyncConsole.test.js`
- `backend/internal/growth/dto/market_data.go`
- `backend/internal/growth/handler/market_data_admin_handler_test.go`
- `backend/internal/growth/repo/market_backfill_execution.go`
- `backend/internal/growth/repo/market_data_backfill_repo_test.go`

**Reference**
- `admin/src/api/admin.js`
- `backend/internal/growth/handler/market_data_admin_handler.go`
- `backend/internal/growth/service/market_data_backfill_service.go`
- `backend/internal/growth/model/market_data_admin.go`

### Task 1: Lock the new sync-task model in helper tests before touching UI

**Files:**
- Modify: `admin/src/lib/system-jobs-admin.js`
- Modify: `admin/src/lib/system-jobs-admin.test.js`
- Create: `admin/src/views/system-jobs-sync-model.test.js`

- [ ] **Step 1: Write failing helper tests for sync task templates**

Add tests that assert:
- four templates exist:
  - `STOCK_FULL`
  - `STOCK_INCREMENTAL`
  - `FUTURES_FULL`
  - `FUTURES_INCREMENTAL`
- each template maps to the expected old payload:
  - `run_type`
  - `asset_scope`
  - `stages`
  - `force_refresh_universe`
  - `rebuild_truth_after_sync`
- existing run rows can be rendered as sync-task labels

Example structure:

```js
test("buildSyncJobTemplateOptions exposes stock and futures sync templates", () => {
  const options = buildSyncJobTemplateOptions();
  assert.deepEqual(
    options.map((item) => item.key),
    ["STOCK_FULL", "STOCK_INCREMENTAL", "FUTURES_FULL", "FUTURES_INCREMENTAL"]
  );
});
```

- [ ] **Step 2: Run the helper tests and verify they fail**

Run:

```bash
node --test admin/src/lib/system-jobs-admin.test.js admin/src/views/system-jobs-sync-model.test.js
```

Expected:
- FAIL because the new sync-task helpers do not exist yet

- [ ] **Step 3: Add the minimal helper layer**

Implement in `admin/src/lib/system-jobs-admin.js`:
- `buildSyncJobTemplateOptions()`
- `buildSyncJobPayloadFromTemplate(templateKey, overrides = {})`
- `deriveSyncJobTemplateFromBackfillRun(run)`
- `formatSyncJobTypeLabel(run)`

Keep them pure and string-driven so they can be tested without mounting Vue.

- [ ] **Step 4: Re-run helper tests**

Run:

```bash
node --test admin/src/lib/system-jobs-admin.test.js admin/src/views/system-jobs-sync-model.test.js
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add admin/src/lib/system-jobs-admin.js admin/src/lib/system-jobs-admin.test.js admin/src/views/system-jobs-sync-model.test.js
git commit -m "feat: add sync task template model for system jobs"
```

### Task 2: Expand backend DTO and execution tests to admit futures sync runs

**Files:**
- Modify: `backend/internal/growth/dto/market_data.go`
- Modify: `backend/internal/growth/handler/market_data_admin_handler_test.go`
- Modify: `backend/internal/growth/repo/market_data_backfill_repo_test.go`
- Create: `backend/internal/growth/repo/market_data_backfill_futures_test.go`
- Modify: `backend/internal/growth/repo/market_backfill_execution.go`

- [ ] **Step 1: Write failing backend tests for futures asset scope**

Add tests that cover:
- `CreateMarketDataBackfillRun` accepts `asset_scope=["FUTURES"]`
- `FULL + FUTURES` can create and execute a run
- `INCREMENTAL + FUTURES` uses a futures-safe stage set and does not require stock-only enhancement stages

Example handler test outline:

```go
func TestCreateMarketDataBackfillRunAcceptsFuturesAssetScope(t *testing.T) {
    // request body uses asset_scope: ["FUTURES"]
    // expect 200 instead of validation error
}
```

- [ ] **Step 2: Run backend tests to verify they fail**

Run:

```bash
go test ./backend/internal/growth/handler -run 'TestCreateMarketDataBackfillRunAcceptsFuturesAssetScope'
go test ./backend/internal/growth/repo -run 'TestExecuteMarketDataBackfillRunSupportsFutures'
```

Expected:
- FAIL because `FUTURES` is not accepted by DTO or creation rules yet

- [ ] **Step 3: Extend DTO validation and futures run normalization**

Modify `backend/internal/growth/dto/market_data.go`:
- extend `MarketDataBackfillRequest.AssetScope` to allow `FUTURES`

Modify `backend/internal/growth/repo/market_backfill_execution.go`:
- normalize futures incremental stages
- ensure futures runs do not attempt stock-only enhancement stages like `DAILY_BASIC` and `MONEYFLOW`
- keep old internal model names intact

- [ ] **Step 4: Re-run backend tests**

Run:

```bash
go test ./backend/internal/growth/handler -run 'TestCreateMarketDataBackfillRunAcceptsFuturesAssetScope'
go test ./backend/internal/growth/repo -run 'TestExecuteMarketDataBackfillRunSupportsFutures|TestExecuteMarketDataBackfillRunFuturesIncrementalSkipsStockEnhancements'
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/growth/dto/market_data.go backend/internal/growth/handler/market_data_admin_handler_test.go backend/internal/growth/repo/market_backfill_execution.go backend/internal/growth/repo/market_data_backfill_repo_test.go backend/internal/growth/repo/market_data_backfill_futures_test.go
git commit -m "feat: support futures in sync task backfill model"
```

### Task 3: Convert System Jobs UI from backfill language to sync-task language

**Files:**
- Modify: `admin/src/views/SystemJobsView.vue`
- Modify: `admin/src/views/system-jobs-view.test.js`
- Test: `admin/src/views/system-jobs-sync-model.test.js`

- [ ] **Step 1: Write failing UI contract tests for sync-task language**

Extend tests to assert:
- page copy uses `市场同步任务中心`
- the create panel uses `新建同步任务`
- the list uses `同步任务列表`
- the form offers the four operator-facing templates
- legacy `回填` wording is removed from the primary market-data tab

- [ ] **Step 2: Run UI tests to verify they fail**

Run:

```bash
node --test admin/src/views/system-jobs-view.test.js admin/src/views/system-jobs-sync-model.test.js
```

Expected:
- FAIL because the old wording is still present

- [ ] **Step 3: Refactor `SystemJobsView.vue` to use template-driven sync tasks**

Implement the following:
- rename the market-data workspace copy to sync-task language
- add a template selector for:
  - 股票全量同步
  - 股票每日增量同步
  - 期货全量同步
  - 期货每日增量同步
- generate the old backfill payload through `buildSyncJobPayloadFromTemplate`
- display run rows via `formatSyncJobTypeLabel`
- keep detail drawer and retry mechanics unchanged except for user-facing copy

Do not rewrite the entire market-data tab structure; keep overview, create form, list, snapshot, and detail flow.

- [ ] **Step 4: Re-run UI tests**

Run:

```bash
node --test admin/src/views/system-jobs-view.test.js admin/src/views/system-jobs-sync-model.test.js
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add admin/src/views/SystemJobsView.vue admin/src/views/system-jobs-view.test.js admin/src/views/system-jobs-sync-model.test.js
git commit -m "feat: upgrade system jobs market tab to sync task center"
```

### Task 4: Retire stock/futures sync as a primary action from Data Sources

**Files:**
- Modify: `admin/src/views/DataSourcesView.vue`
- Modify: `admin/src/composables/useDataSourcesWorkspace.js`
- Modify: `admin/src/composables/useMarketSyncConsole.js`
- Modify: `admin/src/composables/useMarketSyncConsole.test.js`

- [ ] **Step 1: Write failing tests for the redirect-only sync entry**

Add tests that assert:
- stock/futures sync controls are no longer the primary execution actions in the data sources workspace
- the sync section shows a redirect/guidance action to `/system-jobs?tab=market-data`
- governance and health functionality remain present

- [ ] **Step 2: Run the data sources tests to verify they fail**

Run:

```bash
node --test admin/src/composables/useMarketSyncConsole.test.js
```

Expected:
- FAIL because the old sync cards still expose stock/futures execution actions

- [ ] **Step 3: Remove stock/futures execution buttons from the data sources primary path**

Implement the minimal safe migration:
- keep the data source workspace and its governance helpers
- stop exposing stock/futures full/incremental sync actions as the primary operator flow
- replace the sync card actions with a guidance card / redirect action to the system jobs market-data tab

Do not remove truth rebuild, registry, health, or quality tooling.

- [ ] **Step 4: Re-run data sources tests**

Run:

```bash
node --test admin/src/composables/useMarketSyncConsole.test.js
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add admin/src/views/DataSourcesView.vue admin/src/composables/useDataSourcesWorkspace.js admin/src/composables/useMarketSyncConsole.js admin/src/composables/useMarketSyncConsole.test.js
git commit -m "feat: retire data sources sync actions in favor of system jobs"
```

### Task 5: Full verification and regression pass

**Files:**
- No new files; verify all touched surfaces

- [ ] **Step 1: Run focused frontend tests**

Run:

```bash
node --test admin/src/lib/system-jobs-admin.test.js admin/src/views/system-jobs-view.test.js admin/src/views/system-jobs-sync-model.test.js admin/src/composables/useMarketSyncConsole.test.js
```

Expected:
- PASS

- [ ] **Step 2: Run focused backend tests**

Run:

```bash
go test ./backend/internal/growth/handler -run 'TestCreateMarketDataBackfillRun'
go test ./backend/internal/growth/repo -run 'TestExecuteMarketDataBackfillRun|TestAdminCreateMarketDataBackfillRun'
```

Expected:
- PASS

- [ ] **Step 3: Run the admin build**

Run:

```bash
cd admin && npm run build
```

Expected:
- PASS with updated `system-jobs` and `data-sources` bundles

- [ ] **Step 4: Manually sanity-check the target pages in-browser**

Open and verify:
- `http://127.0.0.1:5174/system-jobs?tab=market-data`
- `http://127.0.0.1:5174/data-sources/sync`

Checklist:
- `system-jobs` shows sync-task language, not backfill language
- stock/futures templates exist
- data sources sync no longer exposes stock/futures direct actions
- redirect to system jobs works

- [ ] **Step 5: Commit verification-only fixes if needed**

If any build/test/manual regression is discovered, fix minimally and commit:

```bash
git add <files>
git commit -m "fix: polish sync task center migration"
```

### Task 6: Final integration checkpoint

**Files:**
- Review only

- [ ] **Step 1: Diff review**

Run:

```bash
git diff --stat main...HEAD
git diff -- admin/src/views/SystemJobsView.vue admin/src/composables/useMarketSyncConsole.js backend/internal/growth/dto/market_data.go
```

Expected:
- Only the intended sync-task migration files are changed

- [ ] **Step 2: Summarize migration impact**

Prepare a short release note covering:
- system-jobs market tab now hosts stock/futures sync tasks
- data-sources sync buttons retired as primary flow
- futures now supported in unified sync task model

- [ ] **Step 3: Commit any final docs or test touch-ups**

```bash
git add docs/superpowers/specs/2026-05-24-system-jobs-sync-model-upgrade-design.md docs/superpowers/plans/2026-05-24-system-jobs-sync-model-upgrade.md
git commit -m "docs: add sync task center migration plan"
```
