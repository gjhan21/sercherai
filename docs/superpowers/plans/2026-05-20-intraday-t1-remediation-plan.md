# Intraday T1 Remediation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align the current `INTRADAY_T1` stock recommendation system with the newly approved business contract: overnight watchlist, `09:30-10:00` confirmation, `10:00` entry, combined stop-loss, `T+2` exit, market-state-based position sizing, and replayable recommendation cards.

**Architecture:** Keep the existing `Go backend -> strategy-engine -> run persistence -> admin review -> publish` backbone, but split the current “daily-bar candidate mining” flow into two distinct phases: `T-day overnight watchlist generation` and `T+1 intraday confirmation/execution`. Reuse the existing run/profile/template/admin framework where possible, and add a thin execution-specific layer rather than rewriting the whole stock-selection stack.

**Tech Stack:** Go, Gin, MySQL, Python, FastAPI, Pydantic, existing `strategy-engine`, existing stock selection admin/review pipeline.

---

## Current-State Diagnosis

### What already exists

- `strategy-engine` already supports an `INTRADAY_T1` template and a seven-strategy mining path in [stock_selection_pipeline.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py).
- Intraday-oriented daily features are already built into the Go context payload via [strategy_engine_context.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_engine_context.go) and [strategy_engine_context_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go).
- The Python side already has:
  - strategy mining in [intraday_seed_miner.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/seeds/intraday_seed_miner.py)
  - cross-strategy fusion in [intraday_decision_fusion.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/decision/intraday_decision_fusion.py)
  - portfolio filtering in [portfolio_guard.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/risk/portfolio_guard.py)
  - basic backtesting in [intraday_backtest.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/backtesting/intraday_backtest.py)
- Go already has a mature run/review/admin shell around stock selection in:
  - [stock_selection_run_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_repo.go)
  - [stock_selection_profile_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_profile_repo.go)
  - [stock_selection_admin_handler.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/stock_selection_admin_handler.go)

### Where the current system diverges from the approved design

1. The current `INTRADAY_T1` path is still fundamentally a `T-day end-of-day daily-bar selection engine`, not a `T+1 10:00 confirmation engine`.
2. The strategy set still includes `limit_up_next_day` in [limit_up_next_day.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/strategies/limit_up_next_day.py), which directly conflicts with the approved rule `T日涨停股不进入正式推荐`.
3. The universe defaults are too loose or misaligned:
   - Go default listing-days threshold is `180`, not `60`
   - template minimum turnover is `1e8`, not the approved `5e8`
   - market scope defaults still allow broad `CN_A_ALL`, not first-phase `沪深主板` only
4. The report contract in [stock_report_builder.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/reports/stock_report_builder.py) is built for generic stock recommendations:
   - generic take-profit / stop-loss strings
   - no `10:00` confirmation card
   - no watchlist-to-promotion / rejection state
   - no explicit `T+2` exit contract
5. The current backtest in [intraday_backtest.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/backtesting/intraday_backtest.py) uses:
   - `T日收盘价` as buy price
   - `T+1 收盘价` as sell proxy
   - no `10:00` confirmation
   - no `-3%` hard stop
   - no VWAP / opening-range structure stop
6. The current persistence model stores generic run/candidate/portfolio snapshots, but not:
   - recommendation contract fields
   - promotion / rejection reason states
   - market state `NORMAL / WEAK_RISK_OFF`
   - execution plan fields like `entry_price_ref`, `position_plan`, `stop schema`
7. The current Go context builder provides daily truth features, but not the explicit `09:30-10:00` confirmation inputs needed for production:
   - opening range
   - VWAP relation
   - first-30-minute turnover / price structure
   - market-state intraday summary

## File Map

### Existing files to modify

- Python engine
  - [services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py)
  - [services/strategy-engine/app/domain/seeds/intraday_seed_miner.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/seeds/intraday_seed_miner.py)
  - [services/strategy-engine/app/domain/decision/intraday_decision_fusion.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/decision/intraday_decision_fusion.py)
  - [services/strategy-engine/app/domain/risk/portfolio_guard.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/risk/portfolio_guard.py)
  - [services/strategy-engine/app/domain/reports/stock_report_builder.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/reports/stock_report_builder.py)
  - [services/strategy-engine/app/domain/backtesting/intraday_backtest.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/backtesting/intraday_backtest.py)
  - [services/strategy-engine/app/schemas/stock.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/schemas/stock.py)
- Go backend
  - [backend/internal/growth/model/strategy_engine_context.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_engine_context.go)
  - [backend/internal/growth/repo/strategy_engine_context_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go)
  - [backend/internal/growth/repo/stock_selection_run_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_repo.go)
  - [backend/internal/growth/model/stock_selection_run.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_run.go)
  - [backend/internal/growth/handler/stock_selection_admin_handler.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/stock_selection_admin_handler.go)
- SQL / bootstrap
  - [backend/migrations/20260505_01_intraday_t1_template.sql](/Users/gjhan21/cursor/sercherai/backend/migrations/20260505_01_intraday_t1_template.sql)

### New files recommended

- Python engine
  - `services/strategy-engine/app/domain/confirmation/intraday_watchlist_builder.py`
  - `services/strategy-engine/app/domain/confirmation/intraday_confirmation_engine.py`
  - `services/strategy-engine/app/domain/confirmation/intraday_market_state.py`
  - `services/strategy-engine/app/domain/confirmation/intraday_execution_contract.py`
  - `services/strategy-engine/tests/test_intraday_watchlist_builder.py`
  - `services/strategy-engine/tests/test_intraday_confirmation_engine.py`
  - `services/strategy-engine/tests/test_intraday_execution_contract.py`
- Go backend
  - `backend/migrations/20260520_00_intraday_t1_contract_upgrade.sql`
  - `backend/internal/growth/model/stock_selection_intraday.go`
  - `backend/internal/growth/repo/stock_selection_intraday_repo.go`
  - `backend/internal/growth/repo/stock_selection_intraday_repo_test.go`

## Remediation Strategy

### Phase 1: Stop fighting the approved contract

#### Objective

Remove the parts of the current system that are now explicitly wrong for the approved business direction.

#### Required changes

- Remove `limit_up_next_day` from first-phase formal recommendation flow.
- Re-anchor `INTRADAY_T1` around:
  - `沪深主板`
  - `listing_days >= 60`
  - `avg_turnover20 >= 5e8`
  - `T日涨停股排除`
- Tighten portfolio size from current template `limit=5` to first-phase formal recommendation `1-2`.
- Change watchlist semantics from “generic backup list” to “overnight observation list”.

#### Why first

Until these constraints are corrected, every downstream confirmation / backtest / reporting layer will keep producing behavior that violates the approved business promise.

## Phase 2: Split overnight selection from intraday confirmation

#### Objective

Refactor the current single-phase `INTRADAY_T1` path into a two-phase system.

#### Target design

- Phase A: `T日收盘 -> 观察池`
- Phase B: `T+1 09:30-10:00 -> 正式推荐确认`

#### Required changes

- Add an explicit `watchlist builder` that outputs:
  - `style_tag`
  - `overnight_score`
  - `selected_reason`
  - `invalidations`
  - `10:00_focus`
- Stop treating `intraday_seed_miner` output as immediately publishable candidates.
- Make `stock_selection_pipeline` branch by execution phase:
  - overnight watchlist generation
  - intraday confirmation / execution contract generation

#### Key code implication

The current pipeline in [stock_selection_pipeline.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py) should no longer jump directly from strategy mining to `portfolio_guard -> report_builder`.

It needs a new intermediate step:

- `watchlist_generation`
- `intraday_confirmation`
- `formal_recommendation_selection`

## Phase 3: Add production intraday confirmation inputs

#### Objective

Feed the Python engine with the data it actually needs for `09:30-10:00` confirmation.

#### Required inputs

- opening range high / low
- first-30-minute return
- first-30-minute turnover
- VWAP relation
- opening trend shape
- board / theme intraday breadth
- intraday market-state summary

#### Current gap

The Go context builder currently computes end-of-day and recent-history features only. It does not expose a first-30-minute confirmation contract.

#### Required changes

- Extend [strategy_engine_context.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_engine_context.go) seed structs with intraday confirmation fields.
- Extend [strategy_engine_context_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go) to read `market_intraday_quotes` and compute:
  - `open_30m_high`
  - `open_30m_low`
  - `open_30m_return`
  - `open_30m_amount`
  - `above_vwap_ratio`
  - `open_30m_structure`
- Add aggregate market-state inputs for:
  - index / turnover
  - short-term sentiment
  - theme continuity

#### Note

This is the most important backend upgrade. Without it, the current T1 system can only simulate “short-cycle daily factor selection”, not true `10:00` confirmation.

## Phase 4: Build the confirmation engine

#### Objective

Implement the `10:00` confirmation layer described in the design.

#### Required changes

- Add a dedicated `intraday_confirmation_engine` responsible for:
  - hard filters
  - score recomputation
  - double-gate selection
  - market-state downshift
  - position sizing plan
- Encode:
  - `+4%` chase cap
  - total score threshold `>= 72`
  - stock strength threshold `>= 32`
  - sector resonance or capital sentiment pass condition
- Allow at most:
  - `2` formal recommendations in `NORMAL`
  - `1` formal recommendation in `WEAK_RISK_OFF`

#### Why separate engine

Trying to cram this into [intraday_decision_fusion.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/decision/intraday_decision_fusion.py) would blur two different concerns:

- overnight candidate fusion
- execution-time confirmation

They should be separate modules.

## Phase 5: Replace generic report output with a recommendation contract

#### Objective

Make the formal recommendation output match the approved product contract instead of generic stock recommendation fields.

#### Required changes

- Extend [stock.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/schemas/stock.py) with fields for:
  - `market_state`
  - `style_tag`
  - `entry_time`
  - `entry_price_ref`
  - `overnight_reason`
  - `confirmation_reason`
  - `score_breakdown`
  - `risk_controls`
  - `position_plan`
  - `recommendation_status`
  - `promotion_state`
- Update [stock_report_builder.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/reports/stock_report_builder.py) to produce:
  - formal recommendation cards
  - watchlist rejection reasons
  - `T+2` exit contract
  - stop-loss explanation

#### Immediate cleanup

The current generic strings like “上涨10%-15%分批止盈” and “回撤5%止损” should not survive in first-phase T1. They contradict the approved rule:

- no early take-profit
- hard stop `-3%`
- structure stop
- default `T+2` exit

## Phase 6: Upgrade persistence for replay and trust

#### Objective

Persist the exact decision contract so admin, replay, and user-facing review can explain what happened.

#### Required changes

- Add new persisted fields for:
  - `market_state`
  - `style_tag`
  - `entry_time`
  - `entry_price_ref`
  - `position_plan`
  - `promotion_state`
  - `confirmation_reason`
  - `invalidations`
  - `risk_controls`
- Extend run snapshots and portfolio entries so they can answer:
  - why A got promoted and B did not
  - whether rejection happened by hard filter or score gate
  - what stop-loss rule was active

#### Current limitation

The existing run schema in [stock_selection_run.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_selection_run.go) is sufficient for generic ranking replay, but not for execution-contract replay.

## Phase 7: Redefine backtest to match reality

#### Objective

Make the evaluation engine reflect the approved trading contract, not a daily close-to-close shortcut.

#### Required changes

- Replace current backtest assumptions in [intraday_backtest.py](/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/backtesting/intraday_backtest.py):
  - buy at `T日收盘`
  - sell at `T+1 收盘`
- New evaluation flow:
  - `T日` build watchlist
  - `T+1 10:00` run confirmation
  - use `10:00-10:01` average as entry reference
  - apply:
    - `-3%` hard stop
    - opening-range / VWAP structure stop
    - `T+2` open exit if still alive
- Add metrics for:
  - watchlist promotion rate
  - hard-stop rate
  - structure-stop rate
  - `NORMAL` vs `WEAK_RISK_OFF`
  - A-class vs B-class outcomes

#### Why critical

Without this, the current backtest will systematically overstate or distort first-phase live behavior.

## Phase 8: Keep admin flow stable while changing semantics

#### Objective

Preserve the existing admin run / review / publish shell while changing what a “run” means.

#### Required changes

- Keep current Go run creation flow in [stock_selection_run_repo.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_repo.go).
- Change run semantics so admin can clearly see:
  - overnight watchlist run
  - confirmation run
  - formal recommendation result
- Update overview / compare pages to surface:
  - market state
  - recommendation count limits
  - promotion / rejection statistics
  - weak-market downshift behavior

#### Recommendation

Do not create a brand new admin subsystem first. Reuse the existing `stock_selection_runs` lifecycle and add intraday-specific fields and stage labels.

## Concrete Rectification Order

### Iteration 1: Contract alignment

- Update `INTRADAY_T1` template defaults
- Remove `limit_up_next_day` from first-phase formal path
- Tighten universe thresholds to approved values
- Reduce formal recommendation count to `1-2`
- Replace generic take-profit / stop-loss strings

### Iteration 2: Overnight watchlist layer

- Introduce watchlist builder
- Add watchlist metadata to report schema
- Persist observation-card fields
- Update run stages to explicitly show watchlist generation

### Iteration 3: Intraday context enrichment

- Compute first-30-minute confirmation fields in Go
- Expose them in strategy engine context
- Add market-state aggregate inputs

### Iteration 4: Confirmation engine and market-state switch

- Implement hard filters
- Implement composite scoring
- Implement double-gate selection
- Implement `NORMAL / WEAK_RISK_OFF`
- Implement position sizing plan

### Iteration 5: Replayable recommendation cards

- Update schema
- Update report builder
- Update persistence
- Update admin read models

### Iteration 6: Realistic backtest and review loop

- Rebuild backtest contract
- Add stop-loss-aware metrics
- Add watchlist-to-formal promotion analytics
- Add weak-market / normal-market comparison

## What should not be changed yet

- Do not expand to创业板 in first-phase remediation.
- Do not add dynamic take-profit in first phase.
- Do not introduce ML ranking before rule contract is stable.
- Do not rewrite the whole stock selection admin system.
- Do not merge overnight watchlist and intraday confirmation into one opaque score.

## Recommended First Engineering Slice

If we want the smallest high-value first slice, it should be:

1. Fix template/universe/strategy alignment
2. Introduce `watchlist -> confirmation -> formal recommendation` stage separation
3. Add report contract fields for recommendation cards
4. Rework backtest to stop using close-to-close shortcuts

That slice alone will not finish the full product, but it will stop the current system from promising one thing in business language and doing another thing in code.

## Validation Checklist

- [ ] `INTRADAY_T1` no longer recommends `T日涨停` stocks
- [ ] universe is restricted to first-phase main-board rules
- [ ] formal recommendations are capped at `1-2`
- [ ] weak-market state reduces both count and position sizing
- [ ] recommendation cards include entry, stop, exit, invalidation, and confirmation reasoning
- [ ] watchlist items can be replayed with promoted / rejected reasons
- [ ] backtest uses `10:00` entry and `T+2` exit logic
- [ ] stop-loss metrics are visible in evaluation output
- [ ] admin run stages clearly distinguish watchlist from formal recommendation
