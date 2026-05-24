# Deep Forecast Phase 3 History Review Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build phase 3 deep forecast history comparison so users can browse all successful runs for the same target, compare latest vs previous by default, and read an explainable review score for each successful run.

**Architecture:** Keep the existing phase 2 report and research-center architecture, then add a new history/review layer around it. Backend work should project existing learning signals into explicit user-facing review objects and history comparison summaries. Frontend work should consume those new APIs with a single shared history view-model so PC and H5 keep one interaction model with different density.

**Tech Stack:** Go backend (`go test`), Vue 3 / Vite frontend, `node:test` source-contract tests, existing forecast L3 repo/handler/report pipeline

---

## File Structure

### Backend core files

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go`
  - Add `StrategyForecastL3RunReview`
  - Add `StrategyForecastL3HistoryItem`
  - Add `StrategyForecastL3HistoryCompare`
  - Add any supporting diff structs such as `StrategyForecastL3EvidenceDiff`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
  - Extend repository interface with user-facing history/review methods
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning.go`
  - Project learning signals into review score / grade / verdict / notes
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`
  - Add same-target successful history lookup
  - Add default `latest vs previous` comparison builder
  - Add run review lookup
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/user_growth_handler_forecast_l3.go`
  - Add user endpoints for history, compare, and review
- Modify if route wiring is needed: `/Users/gjhan21/cursor/sercherai/backend/router/user.go`
  - Register new forecast phase 3 endpoints if not already grouped in user route file

### Backend test files

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/forecast_l3_handler_test.go`

### Frontend core files

- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/api/forecast.js`
  - Add history / compare / review API helpers
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue`
  - Add history switcher, comparison summary, and review score panel
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastLabView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastLabView.vue`
  - Add target history overview and CTA into comparison flow
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-history-view-model.js`
  - Shared logic for sorting successful runs, choosing default pair, building verdict/evidence/review shifts
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-localization.js`
  - Localize review grades, review verdicts, and evidence diff labels

### Frontend test files

- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-lab-view.test.mjs`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-lab-view.test.mjs`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-history-view-model.test.mjs`

## Task 1: Add Review Objects And Scoring Projection

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning_test.go`

- [ ] **Step 1: Write the failing review projection tests**

Add tests in `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning_test.go` covering the required grading logic:

```go
func TestBuildStrategyForecastL3RunReviewGradesScenarioAndTriggerHit(t *testing.T) {
	record := model.StrategyForecastL3LearningRecord{
		RunID:            "l3run_review_a",
		TargetType:       model.StrategyForecastL3TargetTypeStock,
		TargetKey:        "600519.SH",
		ScenarioHit:      true,
		TriggerHit:       true,
		InvalidationEarly:false,
		BiasLabel:        "UNDERCONFIRMED",
		RoleEffectiveness: map[string]float64{"TECHNICAL": 0.78},
		Summary:          "主情景得到验证",
		CreatedAt:        "2026-05-24T12:00:00Z",
	}

	review := buildStrategyForecastL3RunReview(record)
	if review.ReviewGrade != "A" || review.ReviewScore != 85 {
		t.Fatalf("expected A/85, got %+v", review)
	}
	if review.ReviewVerdict == "" {
		t.Fatalf("expected human-readable review verdict, got %+v", review)
	}
}

func TestBuildStrategyForecastL3RunReviewPenalizesEarlyInvalidation(t *testing.T) {
	record := model.StrategyForecastL3LearningRecord{
		RunID:             "l3run_review_d",
		TargetType:        model.StrategyForecastL3TargetTypeFutures,
		TargetKey:         "AU2408",
		ScenarioHit:       false,
		TriggerHit:        false,
		InvalidationEarly: true,
		BiasLabel:         "RISK_FIRST",
		Summary:           "风险边界过早触发",
		CreatedAt:         "2026-05-24T13:00:00Z",
	}

	review := buildStrategyForecastL3RunReview(record)
	if review.ReviewGrade != "D" || review.ReviewScore != 35 {
		t.Fatalf("expected D/35, got %+v", review)
	}
	if len(review.ReviewNotes) == 0 {
		t.Fatalf("expected review notes, got %+v", review)
	}
}
```

- [ ] **Step 2: Run the new learning tests to verify they fail**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuildStrategyForecastL3RunReviewGradesScenarioAndTriggerHit|TestBuildStrategyForecastL3RunReviewPenalizesEarlyInvalidation'
```

Expected: FAIL because `buildStrategyForecastL3RunReview` and new review structures do not exist yet.

- [ ] **Step 3: Add phase 3 public review models**

Extend `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go` with:

```go
type StrategyForecastL3RunReview struct {
	RunID             string             `json:"run_id"`
	TargetType        string             `json:"target_type"`
	TargetKey         string             `json:"target_key"`
	ReviewScore       int                `json:"review_score"`
	ReviewGrade       string             `json:"review_grade"`
	ReviewVerdict     string             `json:"review_verdict"`
	ScenarioHit       bool               `json:"scenario_hit"`
	TriggerHit        bool               `json:"trigger_hit"`
	InvalidationEarly bool               `json:"invalidation_early"`
	BiasLabel         string             `json:"bias_label,omitempty"`
	RoleEffectiveness map[string]float64 `json:"role_effectiveness,omitempty"`
	ReviewNotes       []string           `json:"review_notes,omitempty"`
	ReviewedAt        string             `json:"reviewed_at,omitempty"`
}
```

Do not remove or repurpose `StrategyForecastL3LearningRecord`; this task adds a user-facing projection, not a replacement.

- [ ] **Step 4: Implement minimal scoring projection**

In `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning.go`, add:

```go
func buildStrategyForecastL3RunReview(item model.StrategyForecastL3LearningRecord) model.StrategyForecastL3RunReview
```

Rules:
- `scenario_hit && trigger_hit && !invalidation_early` => `A / 85`
- `scenario_hit && !trigger_hit` => `B / 70`
- `!scenario_hit && !invalidation_early` => `C / 55`
- `invalidation_early` => `D / 35`

Also map `bias_label` into human-readable `review_verdict` and `review_notes`.

- [ ] **Step 5: Re-run the learning tests**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuildStrategyForecastL3RunReviewGradesScenarioAndTriggerHit|TestBuildStrategyForecastL3RunReviewPenalizesEarlyInvalidation'
```

Expected: PASS.

- [ ] **Step 6: Commit Task 1**

```bash
git add backend/internal/growth/model/strategy_forecast_l3.go \
        backend/internal/growth/repo/strategy_forecast_l3_learning.go \
        backend/internal/growth/repo/strategy_forecast_l3_learning_test.go
git commit -m "feat: add deep forecast run review scoring"
```

## Task 2: Add Same-Target Success History And Default Comparison

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo_test.go`

- [ ] **Step 1: Write failing repo tests for success history and default compare**

Add tests in `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo_test.go`:

```go
func TestListStrategyForecastL3HistoryForTargetReturnsOnlySucceededRuns(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	seedHistoryRunsForTarget(t, repo, "600519.SH", model.StrategyForecastL3TargetTypeStock)

	items, err := repo.ListStrategyForecastL3HistoryForTarget(model.StrategyForecastL3TargetTypeStock, "600519.SH", 1, 20)
	if err != nil {
		t.Fatalf("ListStrategyForecastL3HistoryForTarget() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected successful history items")
	}
	for _, item := range items {
		if item.ReviewGrade == "" {
			t.Fatalf("expected review grade in history item, got %+v", item)
		}
	}
}

func TestBuildStrategyForecastL3HistoryCompareDefaultsToLatestVsPrevious(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	seedHistoryRunsForTarget(t, repo, "AU2408", model.StrategyForecastL3TargetTypeFutures)

	compare, err := repo.GetStrategyForecastL3HistoryCompare(model.StrategyForecastL3TargetTypeFutures, "AU2408", "", "")
	if err != nil {
		t.Fatalf("GetStrategyForecastL3HistoryCompare() error = %v", err)
	}
	if compare.LeftRun == nil || compare.RightRun == nil {
		t.Fatalf("expected latest vs previous pair, got %+v", compare)
	}
	if len(compare.VerdictShift) == 0 {
		t.Fatalf("expected verdict shift summary, got %+v", compare)
	}
}
```

- [ ] **Step 2: Run the repo tests to verify they fail**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestListStrategyForecastL3HistoryForTargetReturnsOnlySucceededRuns|TestBuildStrategyForecastL3HistoryCompareDefaultsToLatestVsPrevious'
```

Expected: FAIL because history methods and public history models do not exist yet.

- [ ] **Step 3: Add public history objects**

Extend `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go` with:

```go
type StrategyForecastL3HistoryItem struct { ... }
type StrategyForecastL3EvidenceDiff struct { ... }
type StrategyForecastL3HistoryCompare struct { ... }
```

Keep them small and summary-oriented. Do not embed full report bodies.

- [ ] **Step 4: Extend repo interfaces**

In `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`, add:

```go
ListStrategyForecastL3HistoryForTarget(targetType string, targetKey string, page int, pageSize int) ([]model.StrategyForecastL3HistoryItem, int, error)
GetStrategyForecastL3HistoryCompare(targetType string, targetKey string, leftRunID string, rightRunID string) (model.StrategyForecastL3HistoryCompare, error)
GetStrategyForecastL3RunReview(runID string) (model.StrategyForecastL3RunReview, error)
```

- [ ] **Step 5: Implement same-target history**

In `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`:

- Filter by:
  - `target_type`
  - `target_key`
  - `status = SUCCEEDED`
- Sort by most recent first
- Map runs + report summary + run review into `StrategyForecastL3HistoryItem`

Prefer helpers:

```go
func buildStrategyForecastL3HistoryItem(run model.StrategyForecastL3Run, report *model.StrategyForecastL3Report, review model.StrategyForecastL3RunReview) model.StrategyForecastL3HistoryItem
func buildStrategyForecastL3HistoryItemsForTarget(...)
```

- [ ] **Step 6: Implement default `latest vs previous` compare**

In `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`:

- If `left_run_id/right_run_id` empty:
  - pick `latest` and `previous` from successful history
- If fewer than 2 success runs:
  - return a stable compare object with empty shifts and only available side populated
- Build:
  - `verdict_shift`
  - `review_shift`
  - placeholder `evidence_shift` from dimension evidence summary diff

Do not compare raw markdown bodies.

- [ ] **Step 7: Re-run the history repo tests**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestListStrategyForecastL3HistoryForTargetReturnsOnlySucceededRuns|TestBuildStrategyForecastL3HistoryCompareDefaultsToLatestVsPrevious'
```

Expected: PASS.

- [ ] **Step 8: Commit Task 2**

```bash
git add backend/internal/growth/model/strategy_forecast_l3.go \
        backend/internal/growth/repo/interfaces.go \
        backend/internal/growth/repo/strategy_forecast_l3_repo.go \
        backend/internal/growth/repo/strategy_forecast_l3_repo_test.go
git commit -m "feat: add deep forecast history comparison data"
```

## Task 3: Expose User-Facing History, Compare, And Review APIs

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/user_growth_handler_forecast_l3.go`
- Modify if needed: `/Users/gjhan21/cursor/sercherai/backend/router/user.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/forecast_l3_handler_test.go`

- [ ] **Step 1: Add failing handler tests**

Add tests in `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/forecast_l3_handler_test.go` for:

```go
func TestListForecastL3HistoryForTargetRequiresAuthAndReturnsHistory(t *testing.T) { ... }
func TestGetForecastL3HistoryCompareDefaultsToLatestVsPrevious(t *testing.T) { ... }
func TestGetForecastL3RunReviewReturnsStructuredReview(t *testing.T) { ... }
```

Validate:
- user auth required
- successful response payload has review/history fields
- default compare uses stable object structure

- [ ] **Step 2: Run the handler tests to verify they fail**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler -run 'TestListForecastL3HistoryForTargetRequiresAuthAndReturnsHistory|TestGetForecastL3HistoryCompareDefaultsToLatestVsPrevious|TestGetForecastL3RunReviewReturnsStructuredReview'
```

Expected: FAIL because endpoints do not exist yet.

- [ ] **Step 3: Add user-facing handlers**

In `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/user_growth_handler_forecast_l3.go`, add handlers for:

- `GET /api/v1/forecast/targets/history`
- `GET /api/v1/forecast/targets/history/compare`
- `GET /api/v1/forecast/runs/:id/review`

Rules:
- all require authenticated user
- history endpoints only return successful history summaries
- review endpoint returns summary/review only, not privileged full report text

- [ ] **Step 4: Wire routes**

Register new handlers in the correct user route file. Follow the existing `/api/v1/forecast/*` pattern instead of adding a new route group.

- [ ] **Step 5: Re-run handler tests**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler -run 'TestListForecastL3HistoryForTargetRequiresAuthAndReturnsHistory|TestGetForecastL3HistoryCompareDefaultsToLatestVsPrevious|TestGetForecastL3RunReviewReturnsStructuredReview'
```

Expected: PASS.

- [ ] **Step 6: Commit Task 3**

```bash
git add backend/internal/growth/handler/user_growth_handler_forecast_l3.go \
        backend/router/user.go \
        backend/internal/growth/handler/forecast_l3_handler_test.go
git commit -m "feat: expose deep forecast history review APIs"
```

## Task 4: Add Shared History View-Model And Detail-Page History Comparison

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/api/forecast.js`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-history-view-model.js`
- Create: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-history-view-model.test.mjs`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-localization.js`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs`

- [ ] **Step 1: Add failing shared view-model tests**

Create `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-history-view-model.test.mjs`:

```js
import test from "node:test";
import assert from "node:assert/strict";
import { buildForecastHistoryViewModel } from "./forecast-history-view-model.js";

test("buildForecastHistoryViewModel defaults to latest vs previous", () => {
  const vm = buildForecastHistoryViewModel({
    historyItems: [
      { run_id: "run3", generated_at: "2026-05-24T12:00:00Z", review_grade: "A" },
      { run_id: "run2", generated_at: "2026-05-23T12:00:00Z", review_grade: "B" },
      { run_id: "run1", generated_at: "2026-05-22T12:00:00Z", review_grade: "C" }
    ]
  });

  assert.equal(vm.leftRun.run_id, "run3");
  assert.equal(vm.rightRun.run_id, "run2");
});
```

- [ ] **Step 2: Add failing detail-page source-contract assertions**

Extend both detail-page tests to require:

- `同标的历史`
- `最新一次 vs 上一次`
- `完整复盘评分`
- `查看全部成功 run`
- `结论变化`
- `证据变化`

- [ ] **Step 3: Run the new frontend tests to verify they fail**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai && node --test \
  newclient/src/shared/lib/forecast-history-view-model.test.mjs \
  newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  newclient/src/apps/h5/views/forecast-detail-view.test.mjs
```

Expected: FAIL because APIs and UI sections do not exist yet.

- [ ] **Step 4: Add forecast history API helpers**

In `/Users/gjhan21/cursor/sercherai/newclient/src/api/forecast.js`, add:

```js
export function listForecastHistory(params) { ... }
export function getForecastHistoryCompare(params) { ... }
export function getForecastRunReview(id) { ... }
```

- [ ] **Step 5: Build shared history view-model**

Implement `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-history-view-model.js`:

- sort history
- choose default pair (`latest vs previous`)
- normalize review grade / review score
- normalize verdict shift / evidence shift / review shift
- provide compact time-axis items for PC/H5

Keep the module pure and testable.

- [ ] **Step 6: Update detail views**

In both PC/H5 detail pages:

- fetch history + compare + review alongside existing run detail
- add `同标的历史` section
- render:
  - history switcher
  - latest vs previous summary
  - `结论变化`
  - `证据变化`
  - `完整复盘评分`
  - `查看全部成功 run`

H5 should render a compressed version of the same information, not a separate interaction model.

- [ ] **Step 7: Re-run the frontend tests**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai && node --test \
  newclient/src/shared/lib/forecast-history-view-model.test.mjs \
  newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  newclient/src/apps/h5/views/forecast-detail-view.test.mjs
```

Expected: PASS.

- [ ] **Step 8: Commit Task 4**

```bash
git add newclient/src/api/forecast.js \
        newclient/src/shared/lib/forecast-history-view-model.js \
        newclient/src/shared/lib/forecast-history-view-model.test.mjs \
        newclient/src/shared/lib/forecast-localization.js \
        newclient/src/apps/pc/views/forecast/ForecastDetailView.vue \
        newclient/src/apps/h5/views/forecast/ForecastDetailView.vue \
        newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
        newclient/src/apps/h5/views/forecast-detail-view.test.mjs
git commit -m "feat: add forecast history comparison views"
```

## Task 5: Add Forecast Lab History Overview And Comparison Entry

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastLabView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastLabView.vue`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-lab-view.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-lab-view.test.mjs`

- [ ] **Step 1: Add failing lab test assertions**

Extend both lab source-contract tests to require:

- `历史研究概览`
- `查看同标的历史对比`
- `成功 run 数量`
- `最近复盘评分趋势`
- `当前仅有 1 次成功深推演，暂不能形成历史对比`

- [ ] **Step 2: Run the lab tests to verify they fail**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai && node --test \
  newclient/src/apps/pc/views/forecast-lab-view.test.mjs \
  newclient/src/apps/h5/views/forecast-lab-view.test.mjs
```

Expected: FAIL because phase 3 history overview language and CTA do not exist yet.

- [ ] **Step 3: Extend Forecast Lab with history overview**

In both lab views:

- load same-target history when a focused target exists
- display:
  - successful run count
  - latest review grade
  - latest vs previous verdict shift summary
  - recent review trend
- add CTA:
  - `查看同标的历史对比`

If only one successful run exists, show the locked message instead of compare CTA.

- [ ] **Step 4: Re-run the lab tests**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai && node --test \
  newclient/src/apps/pc/views/forecast-lab-view.test.mjs \
  newclient/src/apps/h5/views/forecast-lab-view.test.mjs
```

Expected: PASS.

- [ ] **Step 5: Commit Task 5**

```bash
git add newclient/src/apps/pc/views/forecast/ForecastLabView.vue \
        newclient/src/apps/h5/views/forecast/ForecastLabView.vue \
        newclient/src/apps/pc/views/forecast-lab-view.test.mjs \
        newclient/src/apps/h5/views/forecast-lab-view.test.mjs
git commit -m "feat: add forecast history entry in research center"
```

## Task 6: Full Regression And Completion Handoff

**Files:**
- No planned source changes; verification and cleanup only

- [ ] **Step 1: Run backend phase 3 tests**

```bash
cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuildStrategyForecastL3RunReview|TestListStrategyForecastL3HistoryForTarget|TestBuildStrategyForecastL3HistoryCompareDefaultsToLatestVsPrevious'
```

Expected: PASS.

- [ ] **Step 2: Run handler phase 3 tests**

```bash
cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler -run 'TestListForecastL3HistoryForTargetRequiresAuthAndReturnsHistory|TestGetForecastL3HistoryCompareDefaultsToLatestVsPrevious|TestGetForecastL3RunReviewReturnsStructuredReview'
```

Expected: PASS.

- [ ] **Step 3: Run frontend phase 3 tests**

```bash
cd /Users/gjhan21/cursor/sercherai && node --test \
  newclient/src/shared/lib/forecast-history-view-model.test.mjs \
  newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  newclient/src/apps/h5/views/forecast-detail-view.test.mjs \
  newclient/src/apps/pc/views/forecast-lab-view.test.mjs \
  newclient/src/apps/h5/views/forecast-lab-view.test.mjs
```

Expected: PASS.

- [ ] **Step 4: Run full newclient build**

```bash
cd /Users/gjhan21/cursor/sercherai/newclient && npm run build
```

Expected: PASS with both `dist/index.html` and `dist-h5/m/index.html` generated.

- [ ] **Step 5: Review branch diff and clean generated noise**

Check:

```bash
cd /Users/gjhan21/cursor/sercherai && git status --short
```

If `dist` or `node_modules` noise appears, restore them before handoff.

- [ ] **Step 6: Final commit if cleanup changed tracked files**

Only if needed:

```bash
git add ...
git commit -m "chore: finalize deep forecast phase3 history review"
```

- [ ] **Step 7: Hand off for finishing-a-development-branch**

After all checks pass, stop implementation work and transition into the finishing skill for merge / PR / cleanup options.
