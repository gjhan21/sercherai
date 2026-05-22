# Technical Analysis Selection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a code-aligned implementation path for a new stock selection system that uses a shared trend-oriented candidate pool, a market-analysis layer, a short-term primary recommendation head, and a swing-trade auxiliary recommendation head.

**Architecture:** Reuse the existing stock-selection runtime, profile/template system, context builder, report persistence, and admin lifecycle. Add a market-analysis stage and split the current single stock-selection decision path into a shared first-layer candidate pool plus two explicit second-layer recommendation heads with distinct contracts.

**Tech Stack:** Go, Gin, MySQL, Python, FastAPI, Pydantic, existing `strategy-engine`, existing stock-selection admin and evaluation pipeline.

---

## Scope

This plan covers one coherent sub-project:

- `大盘分析 + 趋势型待选池 + 超短线主推头 + 短波段辅推头`

It does not include:

- live broker execution
- complex ML ranking
- full frontend redesign

## Current Code Reality

### What we already have

- Go can already build stock daily context and T1-oriented derived fields in:
  - [backend/internal/growth/repo/strategy_engine_context_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go)
  - [backend/internal/growth/model/strategy_engine_context.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_engine_context.go)
- Python can already:
  - normalize stock features in [stock_feature_factory.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/features/stock_feature_factory.py)
  - build a universe in [stock_universe_builder.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/universe/stock_universe_builder.py)
  - create five-bucket seed pools in [stock_seed_miner.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/seeds/stock_seed_miner.py)
  - run seven T1 strategies in [intraday_seed_miner.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/seeds/intraday_seed_miner.py)
  - fuse generic scores in [stock_decision_fusion.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/decision/stock_decision_fusion.py)
  - persist run outputs through [stock_selection_pipeline.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py)
- Go already persists runs, candidates, evidence, evaluations, profiles, templates, and admin review flows.
- Admin already has a full `stock-selection` module with:
  - overview
  - runs
  - rules
  - factors
  - templates
  - profiles
  - candidates / reviews
  - evaluation

### What is missing

1. There is no explicit market-analysis stage.
2. There is no explicit first-layer `趋势型待选池` contract.
3. There is no split between:
   - short-term primary head
   - swing auxiliary head
4. Current reports do not distinguish:
   - market conclusion
   - primary short-term recommendation
   - auxiliary swing recommendation
5. Current template/profile payloads do not encode this new layered decision structure.
6. Current admin pages still describe the old generic candidate/portfolio contract rather than:
   - market analysis
   - shared trend candidate pool
   - short-term primary recommendations
   - swing auxiliary recommendations

## File Structure

### New Python files

- `services/strategy-engine/app/domain/market/market_daily_analyzer.py`
  - analyze index / breadth / sentiment and output market state
- `services/strategy-engine/app/domain/candidates/trend_candidate_pool_builder.py`
  - build the first-layer trend-oriented candidate pool
- `services/strategy-engine/app/domain/heads/short_term_recommendation_head.py`
  - decide short-term primary recommendations
- `services/strategy-engine/app/domain/heads/swing_recommendation_head.py`
  - decide swing auxiliary recommendations
- `services/strategy-engine/tests/test_market_daily_analyzer.py`
- `services/strategy-engine/tests/test_trend_candidate_pool_builder.py`
- `services/strategy-engine/tests/test_short_term_recommendation_head.py`
- `services/strategy-engine/tests/test_swing_recommendation_head.py`

### Existing Python files to modify

- [services/strategy-engine/app/domain/models.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/models.py)
- [services/strategy-engine/app/schemas/stock.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/schemas/stock.py)
- [services/strategy-engine/app/domain/features/stock_feature_factory.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/features/stock_feature_factory.py)
- [services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py)
- [services/strategy-engine/app/domain/reports/stock_report_builder.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/reports/stock_report_builder.py)
- [services/strategy-engine/app/main.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/main.py)

### New Go files

- `backend/internal/growth/model/market_analysis.go`
- `backend/internal/growth/repo/market_analysis_repo.go`
- `backend/internal/growth/repo/market_analysis_repo_test.go`
- `backend/migrations/20260520_01_market_analysis_and_dual_head_selection.sql`

### Existing Go files to modify

- [backend/internal/growth/model/strategy_engine_context.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_engine_context.go)
- [backend/internal/growth/repo/strategy_engine_context_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go)
- [backend/internal/growth/model/stock_selection_run.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_run.go)
- [backend/internal/growth/model/stock_selection_v2.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_v2.go)
- [backend/internal/growth/repo/stock_selection_run_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_repo.go)
- [backend/internal/growth/handler/stock_selection_admin_handler.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/stock_selection_admin_handler.go)

### Existing Admin files to modify

- [admin/src/lib/stock-selection.js](/Users/gjhan21/cursor/sercherai/admin/src/lib/stock-selection.js)
- [admin/src/views/stock-selection/StockSelectionOverviewView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionOverviewView.vue)
- [admin/src/views/stock-selection/StockSelectionProfilesView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionProfilesView.vue)
- [admin/src/views/stock-selection/StockSelectionRulesView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionRulesView.vue)
- [admin/src/views/stock-selection/StockSelectionFactorsView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionFactorsView.vue)
- [admin/src/views/stock-selection/StockSelectionRunsView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionRunsView.vue)
- [admin/src/views/stock-selection/StockSelectionCandidatesView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionCandidatesView.vue)
- [admin/src/views/stock-selection/StockSelectionEvaluationView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionEvaluationView.vue)

## Implementation Stages

### Stage 1: Introduce explicit market-analysis output

**Why first**

The rest of the system needs a stable market-state contract before it can bias candidate selection or recommendation heads.

**Target**

Add a dedicated market analysis object with:

- trend state
- next-day rhythm
- risk posture
- breadth / sentiment summary

### Stage 2: Build a trend-oriented first-layer candidate pool

**Why second**

This is the shared upstream layer for both short-term and swing outputs.

**Target**

Replace the current implicit “generic candidate pool” semantics with an explicit `趋势型待选池`.

**Key design rules**

- use broad filtering plus trend-first ranking
- combine:
  - market bias
  - sector/theme strength
  - technical structure
  - fund-flow assist
  - risk deductions

### Stage 3: Split second-layer decision heads

**Why third**

We need to stop mixing short-term and swing logic inside a single recommendation contract.

**Target**

Add:

- `short_term_recommendation_head`
- `swing_recommendation_head`

Each head must consume the same first-layer pool but produce different outputs.

### Stage 4: Rework the stock report contract

**Why fourth**

The current report model cannot clearly express:

- market conclusion
- primary short-term picks
- auxiliary swing picks

**Target**

Make the report output three explicit sections and preserve replayability.

### Stage 5: Persist new layered outputs

**Why fifth**

Admin, evaluation, and compare flows all depend on stored run artifacts.

**Target**

Persist:

- market-analysis result
- first-layer candidate pool semantics
- primary recommendation head output
- auxiliary recommendation head output

### Stage 6: Update template/profile semantics

**Why sixth**

Current template/profile configs are built around generic stock-selection tuning.

**Target**

Introduce configuration buckets for:

- market analysis
- first-layer candidate pool
- short-term head
- swing head

### Stage 7: Expose layered contracts to admin APIs

**Why seventh**

The admin module already exists, so backend contracts must evolve before admin pages can present the new business model cleanly.

**Target**

Expose layered run/profile/evaluation payloads that the current admin module can consume without inventing a parallel API family.

### Stage 8: Upgrade admin pages around the existing stock-selection module

**Why eighth**

This project is not complete until strategy admins can:

- configure the layered model
- inspect a run
- review outputs
- compare results
- evaluate primary vs auxiliary performance

**Target**

Keep the current admin IA, but remap page responsibilities to the new layered contract.

### Stage 9: Upgrade evaluations

**Why ninth**

Once outputs are split, evaluation must distinguish between:

- short-term primary recommendation performance
- swing auxiliary recommendation performance

**Target**

Add separate evaluation scopes, not one shared score bucket.

## Task Breakdown

### Task 1: Add market-analysis domain contract

**Files:**
- Create: `services/strategy-engine/app/domain/market/market_daily_analyzer.py`
- Modify: [services/strategy-engine/app/domain/models.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/models.py)
- Test: `services/strategy-engine/tests/test_market_daily_analyzer.py`

- [ ] **Step 1: Write the failing test**

Add tests that verify the analyzer can classify:

- strong uptrend
- neutral range
- risk-off
- next-day rhythm labels

- [ ] **Step 2: Run test to verify it fails**

Run: `pytest services/strategy-engine/tests/test_market_daily_analyzer.py -v`
Expected: FAIL because analyzer module does not exist

- [ ] **Step 3: Write minimal implementation**

Implement:

- market trend state
- next-day rhythm state
- summary payload

- [ ] **Step 4: Run test to verify it passes**

Run: `pytest services/strategy-engine/tests/test_market_daily_analyzer.py -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/strategy-engine/app/domain/market/market_daily_analyzer.py services/strategy-engine/app/domain/models.py services/strategy-engine/tests/test_market_daily_analyzer.py
git commit -m "feat: add market daily analyzer"
```

### Task 2: Build first-layer trend candidate pool

**Files:**
- Create: `services/strategy-engine/app/domain/candidates/trend_candidate_pool_builder.py`
- Modify: [services/strategy-engine/app/domain/features/stock_feature_factory.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/features/stock_feature_factory.py)
- Modify: [services/strategy-engine/app/domain/models.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/models.py)
- Test: `services/strategy-engine/tests/test_trend_candidate_pool_builder.py`

- [ ] **Step 1: Write the failing test**

Cover:

- trend-first scoring
- exclusion of weak-structure candidates
- market-state bias effects
- output size reduction behavior

- [ ] **Step 2: Run test to verify it fails**

Run: `pytest services/strategy-engine/tests/test_trend_candidate_pool_builder.py -v`
Expected: FAIL because builder module does not exist

- [ ] **Step 3: Write minimal implementation**

Implement a builder that consumes features plus market analysis and returns:

- ranked trend candidate pool
- stage summary
- reason tags

- [ ] **Step 4: Run test to verify it passes**

Run: `pytest services/strategy-engine/tests/test_trend_candidate_pool_builder.py -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/strategy-engine/app/domain/candidates/trend_candidate_pool_builder.py services/strategy-engine/app/domain/features/stock_feature_factory.py services/strategy-engine/app/domain/models.py services/strategy-engine/tests/test_trend_candidate_pool_builder.py
git commit -m "feat: add trend candidate pool builder"
```

### Task 3: Add short-term primary recommendation head

**Files:**
- Create: `services/strategy-engine/app/domain/heads/short_term_recommendation_head.py`
- Modify: [services/strategy-engine/app/domain/models.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/models.py)
- Test: `services/strategy-engine/tests/test_short_term_recommendation_head.py`

- [ ] **Step 1: Write the failing test**

Cover:

- primary picks chosen from shared candidate pool
- minute-confirmation-aware input contract
- recommendation count cap
- rejection of weak intraday structure

- [ ] **Step 2: Run test to verify it fails**

Run: `pytest services/strategy-engine/tests/test_short_term_recommendation_head.py -v`
Expected: FAIL because head module does not exist

- [ ] **Step 3: Write minimal implementation**

Implement:

- short-term selection logic
- market-state sensitivity
- recommendation explanation payload

- [ ] **Step 4: Run test to verify it passes**

Run: `pytest services/strategy-engine/tests/test_short_term_recommendation_head.py -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/strategy-engine/app/domain/heads/short_term_recommendation_head.py services/strategy-engine/app/domain/models.py services/strategy-engine/tests/test_short_term_recommendation_head.py
git commit -m "feat: add short-term recommendation head"
```

### Task 4: Add swing auxiliary recommendation head

**Files:**
- Create: `services/strategy-engine/app/domain/heads/swing_recommendation_head.py`
- Modify: [services/strategy-engine/app/domain/models.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/models.py)
- Test: `services/strategy-engine/tests/test_swing_recommendation_head.py`

- [ ] **Step 1: Write the failing test**

Cover:

- swing picks chosen from same shared pool
- fixed auxiliary output size
- trend-continuation preference
- separation from short-term contract

- [ ] **Step 2: Run test to verify it fails**

Run: `pytest services/strategy-engine/tests/test_swing_recommendation_head.py -v`
Expected: FAIL because head module does not exist

- [ ] **Step 3: Write minimal implementation**

Implement:

- swing ranking logic
- support/resistance output
- auxiliary explanation payload

- [ ] **Step 4: Run test to verify it passes**

Run: `pytest services/strategy-engine/tests/test_swing_recommendation_head.py -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/strategy-engine/app/domain/heads/swing_recommendation_head.py services/strategy-engine/app/domain/models.py services/strategy-engine/tests/test_swing_recommendation_head.py
git commit -m "feat: add swing recommendation head"
```

### Task 5: Rewire stock selection pipeline around layered outputs

**Files:**
- Modify: [services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py)
- Modify: [services/strategy-engine/app/main.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/main.py)
- Test: [services/strategy-engine/tests/test_stock_pipeline.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/tests/test_stock_pipeline.py)

- [ ] **Step 1: Write the failing test**

Add assertions that pipeline output now includes:

- market conclusion
- first-layer candidate pool summary
- short-term primary recommendations
- swing auxiliary recommendations

- [ ] **Step 2: Run test to verify it fails**

Run: `pytest services/strategy-engine/tests/test_stock_pipeline.py -v`
Expected: FAIL because current pipeline does not emit layered sections

- [ ] **Step 3: Write minimal implementation**

Refactor the pipeline sequence to:

- load context
- run market analysis
- build candidate pool
- run both heads
- build layered report

- [ ] **Step 4: Run test to verify it passes**

Run: `pytest services/strategy-engine/tests/test_stock_pipeline.py -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py services/strategy-engine/app/main.py services/strategy-engine/tests/test_stock_pipeline.py
git commit -m "feat: rewire stock pipeline for layered recommendations"
```

### Task 6: Extend report schema and report builder

**Files:**
- Modify: [services/strategy-engine/app/schemas/stock.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/schemas/stock.py)
- Modify: [services/strategy-engine/app/domain/reports/stock_report_builder.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/reports/stock_report_builder.py)
- Test: `services/strategy-engine/tests/test_stock_report_builder_layered.py`

- [ ] **Step 1: Write the failing test**

Cover:

- market conclusion section
- short-term recommendation section
- swing auxiliary section
- shared candidate-pool metadata

- [ ] **Step 2: Run test to verify it fails**

Run: `pytest services/strategy-engine/tests/test_stock_report_builder_layered.py -v`
Expected: FAIL because layered report schema does not exist

- [ ] **Step 3: Write minimal implementation**

Extend schema and builder so reports can serialize layered outputs cleanly.

- [ ] **Step 4: Run test to verify it passes**

Run: `pytest services/strategy-engine/tests/test_stock_report_builder_layered.py -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/strategy-engine/app/schemas/stock.py services/strategy-engine/app/domain/reports/stock_report_builder.py services/strategy-engine/tests/test_stock_report_builder_layered.py
git commit -m "feat: add layered stock report contract"
```

### Task 7: Persist layered selection artifacts in Go

**Files:**
- Create: `backend/migrations/20260520_01_market_analysis_and_dual_head_selection.sql`
- Modify: [backend/internal/growth/model/stock_selection_run.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_run.go)
- Modify: [backend/internal/growth/model/stock_selection_v2.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_v2.go)
- Modify: [backend/internal/growth/repo/stock_selection_run_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_repo.go)
- Test: `backend/internal/growth/repo/stock_selection_dual_head_repo_test.go`

- [ ] **Step 1: Write the failing test**

Cover:

- persistence of market-analysis result
- persistence of first-layer candidate pool metadata
- persistence of short-term and swing sections separately

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/internal/growth/repo/... -run DualHead -v`
Expected: FAIL because schema/repo fields do not exist

- [ ] **Step 3: Write minimal implementation**

Add schema fields and repo persistence logic for layered outputs.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./backend/internal/growth/repo/... -run DualHead -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/20260520_01_market_analysis_and_dual_head_selection.sql backend/internal/growth/model/stock_selection_run.go backend/internal/growth/model/stock_selection_v2.go backend/internal/growth/repo/stock_selection_run_repo.go backend/internal/growth/repo/stock_selection_dual_head_repo_test.go
git commit -m "feat: persist layered stock selection outputs"
```

### Task 8: Add configuration support for layered stock selection

**Files:**
- Modify: [backend/internal/growth/model/stock_selection_profile.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_profile.go)
- Modify: [backend/internal/growth/repo/stock_selection_profile_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_profile_repo.go)
- Modify: [backend/internal/growth/repo/stock_selection_run_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_repo.go)
- Test: `backend/internal/growth/repo/stock_selection_profile_layered_config_test.go`

- [ ] **Step 1: Write the failing test**

Cover:

- profile/template payload merge for market-analysis config
- first-layer candidate-pool config
- short-term head config
- swing head config

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/internal/growth/repo/... -run LayeredConfig -v`
Expected: FAIL because config buckets do not exist

- [ ] **Step 3: Write minimal implementation**

Implement config plumbing without breaking existing profiles.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./backend/internal/growth/repo/... -run LayeredConfig -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/growth/model/stock_selection_profile.go backend/internal/growth/repo/stock_selection_profile_repo.go backend/internal/growth/repo/stock_selection_run_repo.go backend/internal/growth/repo/stock_selection_profile_layered_config_test.go
git commit -m "feat: add layered stock selection config"
```

### Task 9: Add separate evaluation scopes for primary and auxiliary outputs

**Files:**
- Modify: [backend/internal/growth/model/stock_selection_v2.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_v2.go)
- Modify: [backend/internal/growth/repo/stock_selection_run_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_repo.go)
- Modify: [services/strategy-engine/app/schemas/stock.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/schemas/stock.py)
- Test: `backend/internal/growth/repo/stock_selection_dual_scope_evaluation_test.go`

- [ ] **Step 1: Write the failing test**

Cover:

- short-term primary evaluation scope
- swing auxiliary evaluation scope
- no mixed metrics pollution

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/internal/growth/repo/... -run DualScopeEvaluation -v`
Expected: FAIL because scope split does not exist

- [ ] **Step 3: Write minimal implementation**

Implement split evaluation scope persistence and read models.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./backend/internal/growth/repo/... -run DualScopeEvaluation -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/growth/model/stock_selection_v2.go backend/internal/growth/repo/stock_selection_run_repo.go services/strategy-engine/app/schemas/stock.py backend/internal/growth/repo/stock_selection_dual_scope_evaluation_test.go
git commit -m "feat: split evaluation scopes for layered stock selection"
```

### Task 10: Extend backend admin contracts for layered stock selection

**Files:**
- Modify: [backend/internal/growth/model/stock_selection_profile.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_profile.go)
- Modify: [backend/internal/growth/model/stock_selection_run.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_run.go)
- Modify: [backend/internal/growth/model/stock_selection_v2.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_v2.go)
- Modify: [backend/internal/growth/handler/stock_selection_admin_handler.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/stock_selection_admin_handler.go)
- Test: `backend/internal/growth/handler/stock_selection_admin_layered_contract_test.go`

- [ ] **Step 1: Write the failing test**

Cover:

- overview payload includes market-analysis summary and primary/auxiliary counts
- run detail payload includes layered stage summaries
- profile/template payloads include layered config buckets
- evaluation payloads expose separate scopes cleanly

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/internal/growth/handler/... -run LayeredContract -v`
Expected: FAIL because handler payloads do not expose the layered contract yet

- [ ] **Step 3: Write minimal implementation**

Implement additive API contract changes for existing admin endpoints instead of creating a new endpoint family.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./backend/internal/growth/handler/... -run LayeredContract -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/growth/model/stock_selection_profile.go backend/internal/growth/model/stock_selection_run.go backend/internal/growth/model/stock_selection_v2.go backend/internal/growth/handler/stock_selection_admin_handler.go backend/internal/growth/handler/stock_selection_admin_layered_contract_test.go
git commit -m "feat: expose layered stock selection admin contracts"
```

### Task 11: Refactor admin overview and run-center for layered outputs

**Files:**
- Modify: [admin/src/lib/stock-selection.js](/Users/gjhan21/cursor/sercherai/admin/src/lib/stock-selection.js)
- Modify: [admin/src/views/stock-selection/StockSelectionOverviewView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionOverviewView.vue)
- Modify: [admin/src/views/stock-selection/StockSelectionRunsView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionRunsView.vue)
- Test: `admin` view/unit test or snapshot tests if available

- [ ] **Step 1: Write the failing test**

Cover:

- overview renders market conclusion and layered output counts
- runs page renders `MARKET_ANALYSIS`, `TREND_CANDIDATE_POOL`, `SHORT_TERM_PRIMARY`, `SWING_AUXILIARY`
- formatters support new market rhythm / head labels / stage labels

- [ ] **Step 2: Run test to verify it fails**

Run the relevant admin tests or snapshot checks for the stock-selection overview/runs views.
Expected: FAIL because the current pages still assume the old generic contract

- [ ] **Step 3: Write minimal implementation**

Update the pages to treat the existing stock-selection module as:

- a daily market-analysis dashboard
- a layered run center

- [ ] **Step 4: Run test to verify it passes**

Run the relevant admin tests or snapshot checks for the stock-selection overview/runs views.
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add admin/src/lib/stock-selection.js admin/src/views/stock-selection/StockSelectionOverviewView.vue admin/src/views/stock-selection/StockSelectionRunsView.vue
git commit -m "feat: align stock selection overview and runs with layered outputs"
```

### Task 12: Refactor admin config, candidate-review, and evaluation pages

**Files:**
- Modify: [admin/src/views/stock-selection/StockSelectionProfilesView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionProfilesView.vue)
- Modify: [admin/src/views/stock-selection/StockSelectionRulesView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionRulesView.vue)
- Modify: [admin/src/views/stock-selection/StockSelectionFactorsView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionFactorsView.vue)
- Modify: [admin/src/views/stock-selection/StockSelectionCandidatesView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionCandidatesView.vue)
- Modify: [admin/src/views/stock-selection/StockSelectionEvaluationView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionEvaluationView.vue)
- Test: `admin` view/unit test or snapshot tests if available

- [ ] **Step 1: Write the failing test**

Cover:

- profiles/templates expose layered config buckets
- candidates page distinguishes shared candidate pool vs short-term primary vs swing auxiliary
- evaluation page distinguishes `SHORT_TERM_PRIMARY` and `SWING_AUXILIARY`

- [ ] **Step 2: Run test to verify it fails**

Run the relevant admin tests or snapshot checks for stock-selection config/candidates/evaluation views.
Expected: FAIL because the current pages do not understand the layered business model

- [ ] **Step 3: Write minimal implementation**

Implement page-level remapping while preserving the existing admin route structure.

- [ ] **Step 4: Run test to verify it passes**

Run the relevant admin tests or snapshot checks for stock-selection config/candidates/evaluation views.
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add admin/src/views/stock-selection/StockSelectionProfilesView.vue admin/src/views/stock-selection/StockSelectionRulesView.vue admin/src/views/stock-selection/StockSelectionFactorsView.vue admin/src/views/stock-selection/StockSelectionCandidatesView.vue admin/src/views/stock-selection/StockSelectionEvaluationView.vue
git commit -m "feat: align stock selection admin pages with layered business model"
```

## Backend / Admin Contract Checklist

Before frontend delivery, backend must guarantee these additive fields exist:

- `overview.market_analysis`
- `overview.candidate_pool_summary`
- `overview.head_summary`
- `overview.evaluation_split_summary`
- `run.context_meta.market_analysis`
- `run.context_meta.candidate_pool_summary`
- `run.context_meta.head_output_summary`
- `candidate.recommendation_head`
- `candidate.selection_layer`
- `candidate.reason_tags`
- `candidate.veto_tags`
- `candidate.technical_pattern`
- `evaluation.evaluation_scope`
- `evaluation.head_label`

If any of these are missing, admin delivery should pause rather than workaround with implicit frontend guesses.

## Admin Delivery Sequence

To reduce churn in the Vue pages, deliver admin in this order:

1. `lib + enum support`
   - add new stage labels
   - add head labels
   - add market rhythm labels
   - add evaluation-scope labels
2. `overview + runs`
   - these pages validate the layered pipeline exists end-to-end
3. `profiles + rules + factors`
   - these pages validate the layered config model
4. `candidates + reviews`
   - this page validates review and publish flows on top of the new model
5. `evaluation`
   - this page validates performance attribution split by recommendation head

## Admin Acceptance Criteria

The admin portion is only complete when a strategy admin can do all of the following without reading logs or raw JSON:

1. On `总览`:
   - tell whether today is attack / neutral / defense
   - tell how many stocks survived each layer
   - tell how many main vs auxiliary recommendations were generated

2. On `配置方案`:
   - distinguish stock-pool constraints from market-analysis parameters
   - distinguish candidate-pool weights from short-term/swing head weights

3. On `运行中心`:
   - see `MARKET_ANALYSIS -> TREND_CANDIDATE_POOL -> SHORT_TERM_PRIMARY -> SWING_AUXILIARY`
   - inspect each stage output and warning summary

4. On `候选与审核`:
   - distinguish shared pool vs primary picks vs auxiliary picks
   - approve/reject a run without ambiguity about what is being published

5. On `评估复盘`:
   - compare short-term primary and swing auxiliary performance separately
   - filter by market regime and still keep scope separation intact

## Validation Commands

- Python unit tests:
  - `pytest services/strategy-engine/tests/test_market_daily_analyzer.py -v`
  - `pytest services/strategy-engine/tests/test_trend_candidate_pool_builder.py -v`
  - `pytest services/strategy-engine/tests/test_short_term_recommendation_head.py -v`
  - `pytest services/strategy-engine/tests/test_swing_recommendation_head.py -v`
  - `pytest services/strategy-engine/tests/test_stock_pipeline.py -v`
- Go repo tests:
  - `go test ./backend/internal/growth/repo/... -run DualHead -v`
  - `go test ./backend/internal/growth/repo/... -run LayeredConfig -v`
  - `go test ./backend/internal/growth/repo/... -run DualScopeEvaluation -v`
  - `go test ./backend/internal/growth/handler/... -run LayeredContract -v`
- Admin checks:
  - run the available admin unit/snapshot tests for stock-selection views
  - if view tests are unavailable, at minimum run the admin build/check path used by the repo and manually verify the stock-selection pages

## Review Notes

- Keep the existing run/profile/template/admin shell intact.
- Do not fork a separate admin product for this feature; evolve the current `/stock-selection` module.
- Do not let short-term and swing logic collapse back into one generic score.
- Keep first-layer candidate pool shared, but keep second-layer heads separate.
- Prefer additive schema evolution so old runs remain readable.
