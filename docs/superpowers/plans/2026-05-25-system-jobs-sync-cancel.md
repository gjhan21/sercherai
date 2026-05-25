# System Jobs Sync Cancel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan step by step.

**Goal:** Add a cancellable sync-task action to admin `system-jobs` so operators can stop `PENDING` or `RUNNING` market sync runs, with graceful execution-chain cancellation and preserved retry capability.

**Architecture:** Extend the existing `marketBackfillRun` model with a cancel endpoint and repo/service/handler chain, add cooperative cancellation checks to the execution pipeline at stage/batch boundaries, and expose cancel actions in the sync-task list and detail drawer.

**Tech Stack:** Vue 3, Element Plus, admin API client, Go Gin handlers, existing market backfill repo/service chain, Node `--test`, Go `test`

---

## File Map

**Modify**
- `admin/src/api/admin.js`
- `admin/src/views/SystemJobsView.vue`
- `admin/src/views/system-jobs-view.test.js`
- `backend/router/admin.go`
- `backend/internal/growth/handler/market_data_admin_handler.go`
- `backend/internal/growth/handler/market_data_admin_handler_test.go`
- `backend/internal/growth/service/service.go`
- `backend/internal/growth/service/market_data_backfill_service.go`
- `backend/internal/growth/repo/interfaces.go`
- `backend/internal/growth/repo/market_data_backfill.go`
- `backend/internal/growth/repo/market_backfill_execution.go`
- `backend/internal/growth/repo/market_data_backfill_repo_test.go`

**Reference**
- `admin/src/views/ForecastLabView.vue`
- `admin/src/api/admin.js` (forecast cancel pattern)
- `backend/internal/growth/handler/admin_forecast_handler.go`
- `backend/internal/growth/repo/strategy_forecast_l3_repo.go`

---

## Task 1: Lock the frontend cancel affordance with failing tests

- [ ] Extend `admin/src/views/system-jobs-view.test.js` with assertions that:
  - `SystemJobsView.vue` references `cancelMarketDataBackfillRun`
  - sync-task list shows cancel action logic for `PENDING` / `RUNNING`
  - detail drawer exposes cancel task action

- [ ] Run:

```bash
node --test admin/src/views/system-jobs-view.test.js
```

Expected: fail because cancel API/action does not exist yet.

- [ ] Add frontend API stub in `admin/src/api/admin.js`:
  - `cancelMarketDataBackfillRun(id, payload)`

- [ ] Update `SystemJobsView.vue`:
  - add list action button `取消`
  - add detail action `取消任务`
  - confirm dialog + refresh list/detail
  - only show for `PENDING` / `RUNNING`

- [ ] Re-run frontend test:

```bash
node --test admin/src/views/system-jobs-view.test.js
```

- [ ] Commit:

```bash
git add admin/src/api/admin.js admin/src/views/SystemJobsView.vue admin/src/views/system-jobs-view.test.js
git commit -m "feat: add sync task cancel action in admin ui"
```

## Task 2: Add backend cancel endpoint and state transition rules

- [ ] Extend `backend/internal/growth/handler/market_data_admin_handler_test.go` with failing tests for:
  - successful cancel of a `PENDING`/`RUNNING` run
  - rejecting cancel for `SUCCESS`/`FAILED`/`CANCELLED`

- [ ] Run:

```bash
go test ./backend/internal/growth/handler -run 'TestCancelMarketDataBackfillRun'
```

Expected: fail because endpoint does not exist yet.

- [ ] Implement backend API chain:
  - route in `backend/router/admin.go`
  - handler method in `backend/internal/growth/handler/market_data_admin_handler.go`
  - service interface + impl in `backend/internal/growth/service/service.go` and `market_data_backfill_service.go`
  - repo interface + impl in `backend/internal/growth/repo/interfaces.go` and `market_data_backfill.go`

- [ ] Enforce state rules:
  - allow only `PENDING`, `RUNNING`
  - reject `SUCCESS`, `FAILED`, `PARTIAL_SUCCESS`, `CANCELLED`
  - update status to `CANCELLED`
  - set `finished_at`, `updated_at`, `error_message`
  - write operation log `CANCEL_BACKFILL_RUN`

- [ ] Re-run handler tests:

```bash
go test ./backend/internal/growth/handler -run 'TestCancelMarketDataBackfillRun'
```

- [ ] Commit:

```bash
git add backend/router/admin.go backend/internal/growth/handler/market_data_admin_handler.go backend/internal/growth/handler/market_data_admin_handler_test.go backend/internal/growth/service/service.go backend/internal/growth/service/market_data_backfill_service.go backend/internal/growth/repo/interfaces.go backend/internal/growth/repo/market_data_backfill.go
git commit -m "feat: add market sync run cancel endpoint"
```

## Task 3: Make execution cooperative-cancellable

- [ ] Extend `backend/internal/growth/repo/market_data_backfill_repo_test.go` with failing tests that verify:
  - canceled runs stop before later stages
  - canceled runs stop long-history quote chunk loops
  - canceled runs preserve existing successful details

- [ ] Run:

```bash
go test ./backend/internal/growth/repo -run 'TestAdminCancelMarketDataBackfillRun|TestExecuteMarketDataBackfillRunStopsWhenCancelled'
```

Expected: fail because execution does not check cancellation yet.

- [ ] Implement cooperative cancellation in `backend/internal/growth/repo/market_backfill_execution.go`:
  - add helper to read current run status
  - check before each stage
  - check before each asset loop where practical
  - check before each long-history chunk/batch
  - stop with `CANCELLED` final state instead of continuing

- [ ] Re-run repo tests:

```bash
go test ./backend/internal/growth/repo -run 'TestAdminCancelMarketDataBackfillRun|TestExecuteMarketDataBackfillRunStopsWhenCancelled'
```

- [ ] Commit:

```bash
git add backend/internal/growth/repo/market_backfill_execution.go backend/internal/growth/repo/market_data_backfill_repo_test.go
git commit -m "feat: stop market sync execution when run is cancelled"
```

## Task 4: Full verification

- [ ] Run frontend/admin verification:

```bash
node --test admin/src/views/system-jobs-view.test.js
cd admin && npm run build
```

- [ ] Run backend verification:

```bash
go test ./backend/internal/growth/handler -run 'TestCancelMarketDataBackfillRun'
go test ./backend/internal/growth/repo -run 'TestAdminCancelMarketDataBackfillRun|TestExecuteMarketDataBackfillRunStopsWhenCancelled'
```

- [ ] If all pass, keep worktree clean and prepare for integration review.
